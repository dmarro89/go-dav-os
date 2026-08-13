"""Bounded DV/1 framing and Agent JSON protocol helpers."""

import binascii
import json


MAGIC = b"DV"
VERSION = 1
REQUEST = 1
RESPONSE = 2
HEADER_SIZE = 6
MAX_REQUEST_PAYLOAD = 2048
MAX_RESPONSE_PAYLOAD = 1024
MAX_ERROR_MESSAGE = 96


class ProtocolError(Exception):
    def __init__(self, message, code="protocol_error", recoverable=False):
        super().__init__(message)
        self.code = code
        self.recoverable = recoverable


def checksum(data):
    return binascii.crc_hqx(data, 0xFFFF)


def request_payload_length(header):
    if len(header) != HEADER_SIZE:
        raise ProtocolError("partial frame")
    if header[:2] != MAGIC or header[2] != VERSION or header[3] != REQUEST:
        raise ProtocolError("malformed request header")
    payload_len = int.from_bytes(header[4:6], "big")
    if payload_len > MAX_REQUEST_PAYLOAD:
        raise ProtocolError("request payload is too large")
    return payload_len


def decode_request_body(header, body):
    payload_len = request_payload_length(header)
    if len(body) != payload_len + 2:
        raise ProtocolError("partial frame")
    payload = body[:-2]
    received_checksum = int.from_bytes(body[-2:], "big")
    if received_checksum != checksum(header[2:] + payload):
        raise ProtocolError("request checksum mismatch", recoverable=True)
    return payload


def encode_response_frame(payload):
    if len(payload) > MAX_RESPONSE_PAYLOAD:
        raise ProtocolError("response payload is too large")
    header = MAGIC + bytes((VERSION, RESPONSE)) + len(payload).to_bytes(2, "big")
    return header + payload + checksum(header[2:] + payload).to_bytes(2, "big")


def decode_agent_request(payload):
    try:
        request = json.loads(payload.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError, RecursionError) as error:
        raise ProtocolError("request must be valid UTF-8 JSON", "invalid_request") from error
    if not isinstance(request, dict):
        raise ProtocolError("request must be a JSON object", "invalid_request")
    return request


def encode_agent_response(response):
    if not isinstance(response, dict):
        raise ProtocolError("provider response must be a JSON object", "invalid_response")
    try:
        payload = json.dumps(response, separators=(",", ":"), sort_keys=True).encode("utf-8")
    except (TypeError, ValueError) as error:
        raise ProtocolError("provider response must be valid JSON", "invalid_response") from error
    if len(payload) > MAX_RESPONSE_PAYLOAD:
        raise ProtocolError("provider response is too large", "invalid_response")
    return payload


def error_response(code, message):
    response = {
        "error": {
            "code": str(code)[:32],
            "message": str(message)[:MAX_ERROR_MESSAGE],
        }
    }
    while (
        len(json.dumps(response, separators=(",", ":"), sort_keys=True).encode("utf-8"))
        > MAX_RESPONSE_PAYLOAD
    ):
        response["error"]["message"] = response["error"]["message"][:-1]
    return response
