# v0.5.0 Agent Runtime Architecture

Issue #151 defines the foundation for the v0.5.0 Minimum Working Agent. The
runtime is intentionally split into small components so natural-language input
never flows directly into shell execution.

## Package Boundary

The `agent` package owns the shared pipeline and data model:

```txt
Planner
  -> Validator
  -> Safety Gate
  -> Executor
  -> Formatter
  -> Context update
```

The package exposes typed `Plan` and `Action` values. There is no raw shell
command field in an action. Shell integration must map known action kinds such
as `list_files`, `read_file`, or `delete_file` onto explicit internal APIs.

## Components

- `Planner`: converts user input and lightweight context into a typed plan.
- `DeterministicPlanner`: local rule-based planner that is always available.
- `LLMPlanner`: planner facade backed by a host-side `BridgeClient`.
- `BridgeClient`: boundary for the external LLM bridge process.
- `Validator`: rejects malformed plans and unsupported action kinds.
- `SafetyGate`: allows safe actions and stops risky actions for confirmation.
- `Executor`: constrained action surface; it executes typed actions only.
- `Formatter`: turns action results into the structured agent response.
- `Context`: records minimal recent inputs, outputs, and the last intent.

## Security Model

The LLM can propose plans, but it cannot execute commands. The runtime accepts
only known `ActionKind` values, validates bounds on action fields, and fails
closed when a handler is not wired. Risky actions such as deletion are marked as
`confirmation_required` by the safety gate before any executor is called.

The LLM planner does not silently fall back to deterministic mode. If the bridge
is not configured or returns an error, planning fails and the runtime returns
that error before validation, safety evaluation, or execution.

The host bridge JSON contract is documented in
[`llm_bridge_protocol.md`](./llm_bridge_protocol.md). That protocol carries
structured requests, lightweight context, an explicit action allow list, and
typed plan responses. It does not permit raw shell command execution.

## Initial Integration Direction

The first shell-facing integration should wire a narrow `AllowedActionExecutor`
to existing filesystem helpers. The deterministic planner can prove the flow in
QEMU, while the LLM planner can later delegate planning to a host-side bridge
that returns the same typed plan structure.
