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

    if contains_token(text, "memorymap") or (contains_token(text, "memory") and contains_token(text, "map")):
        return (
            "show_memory_map",
            "show_memory_map",
            [],
            "safe",
            "Matched request to show the memory map.",
        )
    if contains_token(text, "version"):
        return "show_version", "show_version", [], "safe", "Matched request to show the OS version."
    if contains_token(text, "ticks"):
        return "show_ticks", "show_ticks", [], "safe", "Matched request to show PIT ticks."
    if contains_token(text, "history"):
        return "show_history", "show_history", [], "safe", "Matched request to show Agent history."
    if starts_with_list(text):
        return "list_files", "list_files", [], "safe", "Matched request to list files."
    if contains_token(text, "delete") or contains_token(text, "remove"):
        target = last_token(input_text)
        if target:
            return (
                "delete_file",
                "delete_file",
                [target],
                "risky",
                "Matched delete request for file " + target + ".",
            )
    if contains_token(text, "read") or contains_token(text, "cat"):
        target = last_token(input_text)
        if target:
            return "read_file", "read_file", [target], "safe", "Matched read request for file " + target + "."
    if contains_token(text, "stat") or contains_token(text, "status"):
        target = last_token(input_text)
        if target:
            return "stat_file", "stat_file", [target], "safe", "Matched stat request for file " + target + "."
    if contains_token(text, "show"):
        # "show files" asks for a listing while "show files.txt" names a file.
        if show_asks_for_list(text):
            return "list_files", "list_files", [], "safe", "Matched request to list files."
        target = last_token(input_text)
        if target and lower_ascii(target) != "show":
            return "read_file", "read_file", [target], "safe", "Matched show request for file " + target + "."

    if contains_token(text, "files") or contains_token(text, "file") or contains_token(text, "list") or contains_token(text, "file list") or contains_token(text, "list files") or contains_token(text, "ls"):
        return "list_files", "list_files", [], "safe", "Matched request to list files."

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


def contains_token(text, token):
    # Keyword matching must not fire inside longer words, otherwise file names
    # such as "tools" or "profile" would be misread as list requests.
    start = 0
    while True:
        idx = text.find(token, start)
        if idx < 0:
            return False
        before_ok = idx == 0 or not _is_word_char(text[idx - 1])
        after = idx + len(token)
        after_ok = after == len(text) or not _is_word_char(text[after])
        if before_ok and after_ok:
            return True
        start = idx + 1


def _is_word_char(char):
    return "a" <= char <= "z" or "A" <= char <= "Z" or "0" <= char <= "9"


def last_token(text):
    tokens = [token for token in text.split() if token]
    if not tokens:
        return ""
    token = tokens[-1]
    if lower_ascii(token) in {"read", "cat", "delete", "remove", "stat", "status", "show"}:
        return ""
    return token


def show_asks_for_list(text):
    # Whitespace-separated tokens keep "files" distinct from "files.txt";
    # trailing sentence punctuation is dropped so "files?" still lists.
    seen_show = False
    for part in text.split():
        while part and not part[-1].isalnum():
            part = part[:-1]
        if seen_show and part in {"ls", "list", "file", "files"}:
            return True
        if part == "show":
            seen_show = True
    return False


def starts_with_list(text):
    # A leading list command wins over action keywords inside file names.
    parts = text.split()
    return bool(parts) and parts[0] in {"list", "ls"}


if __name__ == "__main__":
    main()
