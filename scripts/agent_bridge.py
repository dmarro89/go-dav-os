#!/usr/bin/env python3
"""Serve Agent planning requests over the QEMU serial transport."""

import argparse
import json
import os
import sys
import time

from agent_bridge_protocol import (
    ProtocolError,
    decode_agent_request,
    encode_agent_response,
    error_response,
)
from agent_bridge_planner import PlannerError, validate_provider_plan
from agent_bridge_provider import ProviderError, provider_from_environment
from agent_bridge_transport import connect


def response_for_payload(payload, provider):
    try:
        request = decode_agent_request(payload)
        response = provider.plan(request)
        return encode_agent_response(validate_provider_plan(request, response))
    except (PlannerError, ProtocolError, ProviderError) as error:
        safe_error = error_response(error.code, str(error))
        return encode_agent_response(safe_error)


def serve(transport, provider, timeout, once=False):
    while True:
        try:
            payload = transport.receive(time.monotonic() + timeout)
        except ProtocolError as error:
            if not error.recoverable:
                raise
            safe_error = error_response(error.code, str(error))
            transport.send(
                encode_agent_response(safe_error), time.monotonic() + timeout
            )
            continue
        response = response_for_payload(payload, provider)
        transport.send(response, time.monotonic() + timeout)
        if once:
            return


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", default="/tmp/davos-agent.sock")
    parser.add_argument("--timeout", type=float, default=10.0)
    parser.add_argument("--once", action="store_true")
    args = parser.parse_args()

    try:
        provider = provider_from_environment(os.environ)
        connect_deadline = time.monotonic() + args.timeout
        with connect(args.socket, connect_deadline) as transport:
            serve(transport, provider, args.timeout, args.once)
    except (ProtocolError, ProviderError, OSError) as error:
        code = getattr(error, "code", "transport_error")
        json.dump(error_response(code, str(error)), sys.stderr, separators=(",", ":"))
        sys.stderr.write("\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
