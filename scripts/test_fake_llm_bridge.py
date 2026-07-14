import importlib.util
import json
import subprocess
import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BRIDGE_PATH = ROOT / "scripts" / "fake_llm_bridge.py"


spec = importlib.util.spec_from_file_location("fake_llm_bridge", BRIDGE_PATH)
fake_llm_bridge = importlib.util.module_from_spec(spec)
spec.loader.exec_module(fake_llm_bridge)


DEFAULT_ALLOWED = [
    "list_files",
    "read_file",
    "stat_file",
    "delete_file",
    "show_history",
    "show_version",
    "show_ticks",
    "show_memory_map",
]


class FakeLLMBridgeTest(unittest.TestCase):
    def plan(self, input_text, allowed=None):
        return fake_llm_bridge.plan_for_request(
            {
                "input": input_text,
                "context": {
                    "lastIntent": "read_file",
                    "lastAction": "read_file",
                    "lastSummary": "Read file notes",
                    "requestCount": 3,
                },
                "allowedActions": DEFAULT_ALLOWED if allowed is None else allowed,
            }
        )

    def test_cli_reads_agent_request_and_writes_structured_json(self):
        request = {
            "input": "show me the files",
            "context": {},
            "allowedActions": DEFAULT_ALLOWED,
        }

        result = subprocess.run(
            [sys.executable, str(BRIDGE_PATH)],
            input=json.dumps(request),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=True,
        )

        response = json.loads(result.stdout)
        self.assertEqual(response["intent"], "list_files")
        self.assertEqual(response["action"], "list_files")
        self.assertEqual(response["args"], [])
        self.assertEqual(response["risk"], "safe")

    def test_maps_read_request_with_target_arg(self):
        response = self.plan("read notes")

        self.assertEqual(response["intent"], "read_file")
        self.assertEqual(response["action"], "read_file")
        self.assertEqual(response["args"], ["notes"])
        self.assertEqual(response["risk"], "safe")

    def test_maps_delete_request_as_risky(self):
        response = self.plan("delete notes")

        self.assertEqual(response["intent"], "delete_file")
        self.assertEqual(response["action"], "delete_file")
        self.assertEqual(response["args"], ["notes"])
        self.assertEqual(response["risk"], "risky")

    def test_action_not_in_allow_list_returns_unknown(self):
        response = self.plan("delete notes", allowed=["list_files", "read_file"])

        self.assertEqual(response["intent"], "unknown")
        self.assertEqual(response["action"], "unknown")
        self.assertEqual(response["args"], [])
        self.assertEqual(response["risk"], "safe")

    def test_response_does_not_expose_raw_execution_fields(self):
        response = self.plan("show version")

        forbidden = {"command", "shell", "argv", "script", "exec"}
        self.assertTrue(forbidden.isdisjoint(response.keys()))

    def test_rejects_request_with_raw_execution_fields_safely(self):
        response = fake_llm_bridge.plan_for_request(
            {
                "input": "show me the files",
                "context": {"shell": "rm notes"},
                "allowedActions": DEFAULT_ALLOWED,
            }
        )

        self.assertEqual(response["intent"], "unknown")
        self.assertEqual(response["action"], "unknown")


if __name__ == "__main__":
    unittest.main()
