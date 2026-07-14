#!/usr/bin/env python3
"""Deterministic host-side Agent bridge for local testing.

The fake bridge reads one JSON request from stdin and writes one JSON response to
stdout. It follows docs/v0.5.0/llm_bridge_protocol.md and does not call any
external provider.
"""

import json
import sys


MAX_ARG_LEN = 16
RAW_EXECUTION_FIELDS = {"command", "shell", "argv", "script", "exec"}


def main():
    response = plan_from_stdin(sys.stdin.read())
    json.dump(response, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")


def plan_from_stdin(text):
    try:
        request = json.loads(text)
    except json.JSONDecodeError:
        return unknown_response("Invalid JSON request.")

    if not isinstance(request, dict):
        return unknown_response("Request must be a JSON object.")
    return plan_for_request(request)


def plan_for_request(request):
    if contains_raw_execution_field(request):
        return unknown_response("Request contains raw execution fields.")

    input_text = request.get("input", "")
    if not isinstance(input_text, str):
        return unknown_response("Request input must be a string.")

    allowed = request.get("allowedActions", [])
    if not isinstance(allowed, list):
        return unknown_response("allowedActions must be a list.")
    allowed_actions = {action for action in allowed if isinstance(action, str)}

    intent, action, args, risk, explanation = map_input(input_text)
    if action == "unknown" or action not in allowed_actions:
        return unknown_response("No allowed action matched the request.")

    return response(intent, action, args, risk, explanation)


def contains_raw_execution_field(value):
    if isinstance(value, dict):
        for key, nested in value.items():
            if key in RAW_EXECUTION_FIELDS:
                return True
            if contains_raw_execution_field(nested):
                return True
    elif isinstance(value, list):
        for nested in value:
            if contains_raw_execution_field(nested):
                return True
    return False


def map_input(input_text):
    text = lower_ascii(input_text)

    if contains(text, "memorymap") or (contains(text, "memory") and contains(text, "map")):
        return (
            "show_memory_map",
            "show_memory_map",
            [],
            "safe",
            "Matched request to show the memory map.",
        )
    if contains(text, "version"):
        return "show_version", "show_version", [], "safe", "Matched request to show the OS version."
    if contains(text, "ticks"):
        return "show_ticks", "show_ticks", [], "safe", "Matched request to show PIT ticks."
    if contains(text, "history"):
        return "show_history", "show_history", [], "safe", "Matched request to show Agent history."
    if contains(text, "files") or contains(text, "file list") or contains(text, "list files") or contains(text, "ls"):
        return "list_files", "list_files", [], "safe", "Matched request to list files."

    if contains(text, "delete") or contains(text, "remove"):
        target = last_token(text)
        if target:
            return (
                "delete_file",
                "delete_file",
                [target],
                "risky",
                "Matched delete request for file " + target + ".",
            )
    if contains(text, "read") or contains(text, "cat"):
        target = last_token(text)
        if target:
            return "read_file", "read_file", [target], "safe", "Matched read request for file " + target + "."
    if contains(text, "stat") or contains(text, "status"):
        target = last_token(text)
        if target:
            return "stat_file", "stat_file", [target], "safe", "Matched stat request for file " + target + "."
    if contains(text, "show"):
        target = last_token(text)
        if target and target != "show":
            return "read_file", "read_file", [target], "safe", "Matched show request for file " + target + "."

    return "unknown", "unknown", [], "safe", "No supported fake bridge mapping matched the request."


def unknown_response(explanation):
    return response("unknown", "unknown", [], "safe", explanation)


def response(intent, action, args, risk, explanation):
    clean_args = []
    for arg in args:
        clean_args.append(arg[:MAX_ARG_LEN])

    return {
        "intent": intent,
        "action": action,
        "args": clean_args,
        "risk": risk,
        "explanation": explanation,
    }


def lower_ascii(text):
    result = []
    for char in text:
        if "A" <= char <= "Z":
            result.append(chr(ord(char) + 32))
        else:
            result.append(char)
    return "".join(result)


def contains(text, token):
    return token in text


def last_token(text):
    tokens = [token for token in text.split() if token]
    if not tokens:
        return ""
    token = tokens[-1]
    if token in {"read", "cat", "delete", "remove", "stat", "status", "show"}:
        return ""
    return token


if __name__ == "__main__":
    main()
