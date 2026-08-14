#!/usr/bin/env python3
"""Serve one framed transport probe over the QEMU Agent serial socket."""

import argparse
import time

from agent_bridge_protocol import ProtocolError
from agent_bridge_transport import connect


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", default="/tmp/davos-agent.sock")
    parser.add_argument("--timeout", type=float, default=5.0)
    args = parser.parse_args()

    deadline = time.monotonic() + args.timeout
    with connect(args.socket, deadline) as transport:
        request = transport.receive(deadline)
        if request != b"ping":
            raise ProtocolError("unexpected probe payload")
        transport.send(b"pong", deadline)


if __name__ == "__main__":
    main()
