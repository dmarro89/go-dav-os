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


def remaining_timeout(deadline):
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise ProtocolError("probe timed out")
    return remaining


def read_exact(stream, size, deadline):
    data = bytearray()
    while len(data) < size:
        stream.settimeout(remaining_timeout(deadline))
        try:
            chunk = stream.recv(size - len(data))
        except socket.timeout as error:
            raise ProtocolError("probe timed out") from error
        if not chunk:
            raise ProtocolError("partial frame")
        data.extend(chunk)
    return bytes(data)


def read_request(stream, deadline):
    header = read_exact(stream, HEADER_SIZE, deadline)
    if header[:2] != MAGIC or header[2] != VERSION or header[3] != REQUEST:
        raise ProtocolError("malformed request header")
    payload_len = int.from_bytes(header[4:6], "big")
    if payload_len > MAX_REQUEST_PAYLOAD:
        raise ProtocolError("request payload is too large")
    body = read_exact(stream, payload_len + 2, deadline)
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


def connect(path, deadline):
    while True:
        stream = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            stream.settimeout(remaining_timeout(deadline))
            stream.connect(path)
            return stream
        except ProtocolError:
            stream.close()
            raise
        except socket.timeout as error:
            stream.close()
            raise ProtocolError("probe timed out") from error
        except (FileNotFoundError, ConnectionRefusedError):
            stream.close()
            time.sleep(min(0.05, remaining_timeout(deadline)))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", default="/tmp/davos-agent.sock")
    parser.add_argument("--timeout", type=float, default=5.0)
    args = parser.parse_args()

    deadline = time.monotonic() + args.timeout
    with connect(args.socket, deadline) as stream:
        request = read_request(stream, deadline)
        if request != b"ping":
            raise ProtocolError("unexpected probe payload")
        stream.settimeout(remaining_timeout(deadline))
        try:
            stream.sendall(encode_response(b"pong"))
        except socket.timeout as error:
            raise ProtocolError("probe timed out") from error


if __name__ == "__main__":
    main()
