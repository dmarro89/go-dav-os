# v0.5.0 Fake Host LLM Bridge

Issue #156 adds a deterministic host-side bridge for local testing. It lets the
Agent pipeline exercise the LLM bridge contract without API keys, network
access, or a real provider.

The fake bridge is implemented by `scripts/fake_llm_bridge.py`. It reads one
JSON Agent request from standard input and writes one structured JSON plan
response to standard output.

## Run It

```sh
printf '%s\n' '{
  "input": "show me the files",
  "context": {},
  "allowedActions": [
    "list_files",
    "read_file",
    "stat_file",
    "delete_file",
    "show_history",
    "show_version",
    "show_ticks",
    "show_memory_map"
  ]
}' | python3 scripts/fake_llm_bridge.py
```

Example response:

```json
{
  "action": "list_files",
  "args": [],
  "explanation": "Matched request to list files.",
  "intent": "list_files",
  "risk": "safe"
}
```

## Mappings

The fake bridge recognizes a small deterministic phrase set:

```txt
show me the files -> list_files
show files        -> list_files
list files        -> list_files
read notes        -> read_file notes
show notes        -> read_file notes
stat notes        -> stat_file notes
delete notes      -> delete_file notes
remove notes      -> delete_file notes
show history      -> show_history
show version      -> show_version
show ticks        -> show_ticks
show memory map   -> show_memory_map
show memorymap    -> show_memory_map
```

`delete_file` responses are marked `risky`. All other built-in fake responses
are marked `safe`.

## Allow List Behavior

The bridge respects the request's `allowedActions` field. If a request maps to
an action that is not allowed, the fake bridge returns:

```json
{
  "intent": "unknown",
  "action": "unknown",
  "args": [],
  "risk": "safe",
  "explanation": "No allowed action matched the request."
}
```

`unknown` is non-executable and exists only as a safe fallback plan.

## Local Testing

Run the bridge tests with:

```sh
python3 scripts/test_fake_llm_bridge.py
```

The script is host-side only. It is suitable for local harnesses and QEMU demos
that can pass the protocol JSON over stdin/stdout, serial, debug console, or a
future transport adapter. It does not add any kernel dependency on Python.
