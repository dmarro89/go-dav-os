#!/usr/bin/env python3
"""Serve one framed transport probe over the QEMU Agent serial socket."""

import argparse
import binascii
import socket
import time


MAGIC = b"DV"
VERSION = 1
REQUEST = 1
RESPONSE = 2
HEADER_SIZE = 6
MAX_REQUEST_PAYLOAD = 2048
MAX_RESPONSE_PAYLOAD = 1024


class ProtocolError(Exception):
    pass


def checksum(data):
    return binascii.crc_hqx(data, 0xFFFF)


def read_exact(stream, size):
    data = bytearray()
    while len(data) < size:
        chunk = stream.recv(size - len(data))
        if not chunk:
            raise ProtocolError("partial frame")
        data.extend(chunk)
    return bytes(data)


def read_request(stream):
    header = read_exact(stream, HEADER_SIZE)
    if header[:2] != MAGIC or header[2] != VERSION or header[3] != REQUEST:
        raise ProtocolError("malformed request header")
    payload_len = int.from_bytes(header[4:6], "big")
    if payload_len > MAX_REQUEST_PAYLOAD:
        raise ProtocolError("request payload is too large")
    body = read_exact(stream, payload_len + 2)
    payload = body[:-2]
    received_checksum = int.from_bytes(body[-2:], "big")
    if received_checksum != checksum(header[2:] + payload):
        raise ProtocolError("request checksum mismatch")
    return payload


def encode_response(payload):
    if len(payload) > MAX_RESPONSE_PAYLOAD:
        raise ProtocolError("response payload is too large")
    header = MAGIC + bytes((VERSION, RESPONSE)) + len(payload).to_bytes(2, "big")
    return header + payload + checksum(header[2:] + payload).to_bytes(2, "big")


def connect(path, timeout):
    deadline = time.monotonic() + timeout
    while True:
        stream = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        stream.settimeout(timeout)
        try:
            stream.connect(path)
            return stream
        except (FileNotFoundError, ConnectionRefusedError):
            stream.close()
            if time.monotonic() >= deadline:
                raise
            time.sleep(0.05)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", default="/tmp/davos-agent.sock")
    parser.add_argument("--timeout", type=float, default=5.0)
    args = parser.parse_args()

    with connect(args.socket, args.timeout) as stream:
        request = read_request(stream)
        if request != b"ping":
            raise ProtocolError("unexpected probe payload")
        stream.sendall(encode_response(b"pong"))


if __name__ == "__main__":
    main()
