import io
import json
import sys
import time
import unittest
from pathlib import Path
from unittest import mock


SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))

import agent_bridge
import agent_bridge_protocol
import agent_bridge_provider
import agent_bridge_transport


DEFAULT_REQUEST = {
    "input": "show me the files",
    "context": {},
    "allowedActions": ["list_files", "read_file"],
}


class MemoryTransport:
    def __init__(self, payload):
        self.payload = payload
        self.response = None
        self.receive_deadline = None
        self.send_deadline = None

    def receive(self, deadline):
        self.receive_deadline = deadline
        return self.payload

    def send(self, payload, deadline):
        self.response = payload
        self.send_deadline = deadline


class FakeStream:
    def __init__(self, incoming):
        self.incoming = bytearray(incoming)
        self.sent = bytearray()
        self.timeout = None
        self.closed = False

    def settimeout(self, timeout):
        self.timeout = timeout

    def recv(self, size):
        chunk = self.incoming[:size]
        del self.incoming[:size]
        return bytes(chunk)

    def sendall(self, data):
        self.sent.extend(data)

    def close(self):
        self.closed = True


class FakeHTTPResponse:
    def __init__(self, payload):
        self.payload = payload

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        return None

    def read(self, _size):
        return self.payload


def request_frame(payload):
    header = (
        agent_bridge_protocol.MAGIC
        + bytes((agent_bridge_protocol.VERSION, agent_bridge_protocol.REQUEST))
        + len(payload).to_bytes(2, "big")
    )
    crc = agent_bridge_protocol.checksum(header[2:] + payload).to_bytes(2, "big")
    return header + payload + crc


class AgentBridgeTest(unittest.TestCase):
    def test_fake_provider_serves_one_structured_request(self):
        transport = MemoryTransport(json.dumps(DEFAULT_REQUEST).encode("utf-8"))

        agent_bridge.serve(transport, agent_bridge_provider.FakeProvider(), 1.0, once=True)

        response = json.loads(transport.response)
        self.assertEqual(response["intent"], "list_files")
        self.assertEqual(response["action"], "list_files")
        self.assertNotIn("error", response)

    def test_provider_failure_returns_bounded_structured_error(self):
        class FailingProvider:
            def plan(self, _request):
                raise agent_bridge_provider.ProviderError("provider request failed")

        payload = agent_bridge.response_for_payload(
            json.dumps(DEFAULT_REQUEST).encode("utf-8"), FailingProvider()
        )

        response = json.loads(payload)
        self.assertEqual(response["error"]["code"], "provider_error")
        self.assertEqual(response["error"]["message"], "provider request failed")
        self.assertLessEqual(len(payload), agent_bridge_protocol.MAX_RESPONSE_PAYLOAD)

    def test_non_ascii_provider_failure_is_bounded(self):
        class FailingProvider:
            def plan(self, _request):
                raise agent_bridge_provider.ProviderError("🔥" * 1000)

        payload = agent_bridge.response_for_payload(
            json.dumps(DEFAULT_REQUEST).encode("utf-8"), FailingProvider()
        )

        self.assertEqual(json.loads(payload)["error"]["code"], "provider_error")
        self.assertLessEqual(len(payload), agent_bridge_protocol.MAX_RESPONSE_PAYLOAD)

    def test_invalid_request_returns_structured_error(self):
        payload = agent_bridge.response_for_payload(
            b"not-json", agent_bridge_provider.FakeProvider()
        )

        response = json.loads(payload)
        self.assertEqual(response["error"]["code"], "invalid_request")
        self.assertLessEqual(len(payload), agent_bridge_protocol.MAX_RESPONSE_PAYLOAD)

    def test_recursive_json_request_is_invalid(self):
        with mock.patch.object(
            agent_bridge_protocol.json, "loads", side_effect=RecursionError
        ):
            with self.assertRaises(agent_bridge_protocol.ProtocolError) as raised:
                agent_bridge_protocol.decode_agent_request(b"{}")

        self.assertEqual(raised.exception.code, "invalid_request")
        self.assertIsInstance(raised.exception.__cause__, RecursionError)

    def test_provider_processing_gets_a_separate_send_window(self):
        transport = MemoryTransport(json.dumps(DEFAULT_REQUEST).encode("utf-8"))

        with mock.patch.object(agent_bridge.time, "monotonic", side_effect=(10.0, 20.0)):
            agent_bridge.serve(
                transport, agent_bridge_provider.FakeProvider(), 1.0, once=True
            )

        self.assertEqual(transport.receive_deadline, 11.0)
        self.assertEqual(transport.send_deadline, 21.0)

    def test_serial_transport_reads_and_writes_dv1_frames(self):
        payload = json.dumps(DEFAULT_REQUEST).encode("utf-8")
        stream = FakeStream(request_frame(payload))
        transport = agent_bridge_transport.SerialTransport(stream)
        deadline = time.monotonic() + 1.0

        self.assertEqual(transport.receive(deadline), payload)
        transport.send(b"{}", deadline)

        self.assertEqual(
            bytes(stream.sent), agent_bridge_protocol.encode_response_frame(b"{}")
        )

    def test_transport_failure_is_structured_and_bounded(self):
        payload = json.dumps(DEFAULT_REQUEST).encode("utf-8")
        invalid_frame = bytearray(request_frame(payload))
        invalid_frame[-1] ^= 1
        stream = FakeStream(bytes(invalid_frame) + request_frame(payload))
        transport = agent_bridge_transport.SerialTransport(stream)

        agent_bridge.serve(
            transport, agent_bridge_provider.FakeProvider(), 1.0, once=True
        )

        first_payload_len = int.from_bytes(stream.sent[4:6], "big")
        first_payload = bytes(stream.sent[6 : 6 + first_payload_len])
        second_frame_offset = 6 + first_payload_len + 2
        second_payload_len = int.from_bytes(
            stream.sent[second_frame_offset + 4 : second_frame_offset + 6], "big"
        )
        second_payload = bytes(
            stream.sent[
                second_frame_offset + 6 : second_frame_offset + 6 + second_payload_len
            ]
        )

        self.assertEqual(json.loads(first_payload)["error"]["code"], "protocol_error")
        self.assertLessEqual(
            len(first_payload), agent_bridge_protocol.MAX_RESPONSE_PAYLOAD
        )
        self.assertEqual(json.loads(second_payload)["action"], "list_files")

    def test_partial_frame_terminates_service(self):
        payload = json.dumps(DEFAULT_REQUEST).encode("utf-8")
        stream = FakeStream(request_frame(payload)[:-1])
        transport = agent_bridge_transport.SerialTransport(stream)

        with self.assertRaises(agent_bridge_protocol.ProtocolError):
            agent_bridge.serve(
                transport, agent_bridge_provider.FakeProvider(), 1.0, once=True
            )

        self.assertEqual(stream.sent, b"")

    def test_send_failure_terminates_service(self):
        class FailingSendTransport(MemoryTransport):
            def send(self, _payload, _deadline):
                raise agent_bridge_protocol.ProtocolError(
                    "transport write failed", "transport_error"
                )

        transport = FailingSendTransport(json.dumps(DEFAULT_REQUEST).encode("utf-8"))

        with self.assertRaises(agent_bridge_protocol.ProtocolError) as raised:
            agent_bridge.serve(
                transport, agent_bridge_provider.FakeProvider(), 1.0, once=True
            )

        self.assertEqual(raised.exception.code, "transport_error")

    def test_provider_selection_is_host_environment_only(self):
        self.assertIsInstance(
            agent_bridge_provider.provider_from_environment({}),
            agent_bridge_provider.FakeProvider,
        )
        self.assertIsInstance(
            agent_bridge_provider.provider_from_environment(
                {"DAVOS_AGENT_PROVIDER": "   "}
            ),
            agent_bridge_provider.FakeProvider,
        )
        with self.assertRaises(agent_bridge_provider.ProviderError) as raised:
            agent_bridge_provider.provider_from_environment(
                {"DAVOS_AGENT_PROVIDER": "unsupported"}
            )
        self.assertEqual(raised.exception.code, "provider_config")
        provider = agent_bridge_provider.provider_from_environment(
            {
                "DAVOS_AGENT_PROVIDER": "openai",
                "DAVOS_AGENT_API_KEY": "secret",
                "DAVOS_AGENT_MODEL": "test-model",
                "DAVOS_AGENT_BASE_URL": "https://provider.example/v1",
            }
        )
        self.assertIsInstance(provider, agent_bridge_provider.OpenAICompatibleProvider)
        self.assertEqual(provider.model, "test-model")
        self.assertEqual(provider.endpoint, "https://provider.example/v1/chat/completions")

    def test_openai_compatible_provider_rejects_non_http_endpoint(self):
        with self.assertRaises(agent_bridge_provider.ProviderError) as raised:
            agent_bridge_provider.OpenAICompatibleProvider(
                "secret", "test-model", "file:///tmp/provider"
            )

        self.assertEqual(raised.exception.code, "provider_config")

    def test_openai_compatible_provider_returns_json_plan_without_network(self):
        provider = agent_bridge_provider.OpenAICompatibleProvider(
            "secret", "test-model", "https://provider.example/v1"
        )
        envelope = {
            "choices": [
                {
                    "message": {
                        "content": json.dumps(
                            {
                                "intent": "list_files",
                                "action": "list_files",
                                "args": [],
                                "risk": "safe",
                            }
                        )
                    }
                }
            ]
        }

        def fake_urlopen(request, timeout):
            self.assertEqual(request.get_header("Authorization"), "Bearer secret")
            self.assertEqual(request.full_url, provider.endpoint)
            self.assertEqual(timeout, 10.0)
            return FakeHTTPResponse(json.dumps(envelope).encode("utf-8"))

        with mock.patch.object(agent_bridge_provider.urllib.request, "urlopen", fake_urlopen):
            response = provider.plan(DEFAULT_REQUEST)

        self.assertEqual(response["action"], "list_files")
        self.assertNotIn("secret", json.dumps(response))

    def test_recursive_provider_response_is_structured_and_bounded(self):
        class RecursiveProvider(agent_bridge_provider.OpenAICompatibleProvider):
            def plan(self, request):
                with mock.patch.object(
                    agent_bridge_provider.json, "loads", side_effect=RecursionError
                ):
                    return super().plan(request)

        provider = RecursiveProvider(
            "secret", "test-model", "https://provider.example/v1"
        )
        with mock.patch.object(
            agent_bridge_provider.urllib.request,
            "urlopen",
            return_value=FakeHTTPResponse(b"{}"),
        ):
            payload = agent_bridge.response_for_payload(
                json.dumps(DEFAULT_REQUEST).encode("utf-8"), provider
            )

        self.assertEqual(json.loads(payload)["error"]["code"], "provider_error")
        self.assertLessEqual(len(payload), agent_bridge_protocol.MAX_RESPONSE_PAYLOAD)

    def test_main_reports_socket_failures_as_transport_errors(self):
        stderr = io.StringIO()
        with (
            mock.patch.object(agent_bridge, "connect", side_effect=OSError("socket failed")),
            mock.patch.object(agent_bridge.sys, "argv", ["agent_bridge.py"]),
            mock.patch.object(agent_bridge.sys, "stderr", stderr),
        ):
            status = agent_bridge.main()

        self.assertEqual(status, 1)
        self.assertEqual(json.loads(stderr.getvalue())["error"]["code"], "transport_error")

    def test_incomplete_provider_configuration_does_not_echo_secrets(self):
        environment = {
            "DAVOS_AGENT_PROVIDER": "openai",
            "DAVOS_AGENT_API_KEY": "do-not-log",
        }

        with self.assertRaises(agent_bridge_provider.ProviderError) as raised:
            agent_bridge_provider.provider_from_environment(environment)

        self.assertNotIn("do-not-log", str(raised.exception))


if __name__ == "__main__":
    unittest.main()
