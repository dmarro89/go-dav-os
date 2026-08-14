"""Unix stream adapter for the QEMU Agent serial transport."""

import socket
import time

from agent_bridge_protocol import (
    HEADER_SIZE,
    ProtocolError,
    decode_request_body,
    encode_response_frame,
    request_payload_length,
)


def remaining_timeout(deadline):
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise ProtocolError("transport timed out", "transport_timeout")
    return remaining


def read_exact(stream, size, deadline):
    data = bytearray()
    while len(data) < size:
        stream.settimeout(remaining_timeout(deadline))
        try:
            chunk = stream.recv(size - len(data))
        except socket.timeout as error:
            raise ProtocolError("transport timed out", "transport_timeout") from error
        except OSError as error:
            raise ProtocolError("transport read failed", "transport_error") from error
        if not chunk:
            raise ProtocolError("partial frame")
        data.extend(chunk)
    return bytes(data)


class SerialTransport:
    def __init__(self, stream):
        self.stream = stream

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        self.close()

    def close(self):
        self.stream.close()

    def receive(self, deadline):
        header = read_exact(self.stream, HEADER_SIZE, deadline)
        payload_len = request_payload_length(header)
        body = read_exact(self.stream, payload_len + 2, deadline)
        return decode_request_body(header, body)

    def send(self, payload, deadline):
        self.stream.settimeout(remaining_timeout(deadline))
        try:
            self.stream.sendall(encode_response_frame(payload))
        except socket.timeout as error:
            raise ProtocolError("transport timed out", "transport_timeout") from error
        except OSError as error:
            raise ProtocolError("transport write failed", "transport_error") from error


def connect(path, deadline):
    while True:
        stream = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            stream.settimeout(remaining_timeout(deadline))
            stream.connect(path)
            return SerialTransport(stream)
        except ProtocolError:
            stream.close()
            raise
        except socket.timeout as error:
            stream.close()
            raise ProtocolError("transport timed out", "transport_timeout") from error
        except (FileNotFoundError, ConnectionRefusedError):
            stream.close()
            time.sleep(min(0.05, remaining_timeout(deadline)))
        except OSError as error:
            stream.close()
            raise ProtocolError("transport connection failed", "transport_error") from error
