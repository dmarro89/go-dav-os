# v0.6.0 Host Agent Bridge

Issue #185 turns the transport probe into a small host service. The service
reads one framed JSON request at a time, delegates planning to a host provider
and writes one framed JSON response. The guest still validates every typed plan
before execution.

## Run with the fake provider

Start QEMU with the Agent serial socket:

```sh
make run-agent-transport
```

Then start the bridge in another terminal:

```sh
python3 scripts/agent_bridge.py --socket /tmp/davos-agent.sock
```

The deterministic fake provider is the default. It needs no network access or
credentials and uses the mappings from the
[v0.5.0 fake bridge](../v0.5.0/fake_llm_bridge.md).

## OpenAI-compatible provider

Provider configuration exists only in the host process environment:

```sh
export DAVOS_AGENT_PROVIDER=openai
export DAVOS_AGENT_API_KEY='<provider API key>'
export DAVOS_AGENT_MODEL='<model name>'
export DAVOS_AGENT_BASE_URL='https://provider.example/v1'
python3 scripts/agent_bridge.py --socket /tmp/davos-agent.sock
```

The base URL must be the provider's API root; the bridge appends
`/chat/completions`. The API key is used only in the host HTTP authorization
header. It is never logged, included in a transport frame or sent to the guest.

## Boundaries

- `agent_bridge_transport.py` owns the QEMU Unix serial connection.
- `agent_bridge_protocol.py` owns `DV/1` framing and JSON encoding limits.
- `agent_bridge_provider.py` owns fake and OpenAI-compatible provider calls.
- `agent_bridge.py` processes requests sequentially and coordinates errors.

Requests remain limited to 2048 bytes and responses to 1024 bytes. Provider
and protocol failures become bounded JSON error objects. Fatal transport errors
are emitted as the same bounded JSON shape on stderr and terminate the service.
The service does not expose a public network endpoint and never puts provider
configuration or credentials in the kernel.

The constrained prompt and strict semantic response validation are implemented
separately by issue #186.
