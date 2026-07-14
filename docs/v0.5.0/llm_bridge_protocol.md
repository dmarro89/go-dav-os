# v0.5.0 Host LLM Bridge Protocol

Issue #155 defines the protocol between `go-dav-os` and a host-side LLM bridge.
The bridge runs outside the kernel and returns typed Agent plans. It must never
return raw shell commands.

## Scope

The v0.5.0 bridge protocol is a JSON request/response contract. Transport is
intentionally out of scope: the same payload can be carried by a serial stream,
debug console helper, local process, or test harness.

The kernel-side Agent runtime remains the authority for validation, safety, and
execution. The bridge can only propose one typed action from the provided allow
list.

## Request

```json
{
  "input": "show me the files",
  "context": {
    "lastIntent": "read_file",
    "lastAction": "read_file",
    "lastSummary": "Read file notes",
    "requestCount": 3
  },
  "allowedActions": [
    "list_files",
    "read_file",
    "write_file",
    "stat_file",
    "delete_file",
    "show_history",
    "show_version",
    "show_ticks",
    "show_memory_map"
  ]
}
```

Fields:

- `input`: required string containing the user request.
- `context`: optional lightweight state from recent Agent activity.
- `context.lastIntent`: optional string using the same names as `intent`.
- `context.lastAction`: optional string using the same names as `action`.
- `context.lastSummary`: optional short human-readable summary of the last result.
- `context.requestCount`: optional monotonic count of Agent requests in this session.
- `allowedActions`: required list of executable action names the bridge may return.

The request does not contain shell command text, shell history replay commands,
or an unconstrained tool list.

## Response

```json
{
  "intent": "list_files",
  "action": "list_files",
  "args": [],
  "risk": "safe",
  "explanation": "The user wants to see the files available in the OS."
}
```

Fields:

- `intent`: required intent name.
- `action`: required action name. It must appear in `allowedActions`, except for
  the non-executable fallback action `unknown`.
- `args`: optional list of strings. v0.5.0 uses at most one argument as a target
  name for file and mode actions.
- `risk`: required risk level, either `safe` or `risky`.
- `explanation`: optional short reason for the chosen plan.

The response represents a typed plan, not a command line. Fields such as
`command`, `shell`, `argv`, `script`, or `exec` are invalid.

## Names

Supported intents:

```txt
unknown
list_files
read_file
write_file
delete_file
stat_file
show_help
show_history
show_version
show_ticks
show_memory_map
set_mode
```

Supported bridge actions for v0.5.0:

```txt
list_files
read_file
write_file
stat_file
delete_file
show_history
show_version
show_ticks
show_memory_map
```

The kernel may provide a smaller `allowedActions` list for a specific request.
The bridge must choose from that list, or return `intent: "unknown"` and
`action: "unknown"` when no allowed action matches the request.

Risk levels:

```txt
safe
risky
```

`delete_file` is `risky` unless the local runtime has already converted a
confirmed user action into a safe internal action.

## Invalid Responses

The host integration must reject a bridge response safely when any of these are
true:

- required fields are missing or have the wrong JSON type
- `action` is not in `allowedActions` and is not `unknown`
- `action` is not a known Agent action
- `intent` is not a known Agent intent
- `risk` is not `safe` or `risky`
- `args` contains more than the v0.5.0 action expects
- an argument exceeds the Agent target length limit
- the response contains raw execution fields such as `command`, `shell`,
  `argv`, `script`, or `exec`

Rejected responses must become planner failures or typed `unknown` plans. They
must not be executed and must not be translated into shell input.

## Security Invariant

The bridge is a planner only. It cannot execute, request arbitrary execution, or
expand the action surface. The runtime validates the returned typed plan, runs
the safety gate, and dispatches only through `AllowedActionExecutor`.

## Fake Bridge

For local testing without an LLM provider, use the deterministic fake bridge
documented in [`fake_llm_bridge.md`](./fake_llm_bridge.md).
