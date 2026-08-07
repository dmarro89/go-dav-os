# v0.6.0 QEMU Agent Transport

Issue #184 adds the first bounded byte transport between the QEMU guest and a
host process. It carries one Agent request and one response at a time. It does
not parse JSON, call a provider, execute a shell command, or move credentials
into the guest.

## QEMU endpoint

The guest uses the first emulated serial port (COM1, I/O base `0x3F8`) in
polled 38400 8N1 mode. QEMU exposes that port as a Unix stream socket on the
host:

```sh
make run-agent-transport
```

The default socket is `/tmp/davos-agent.sock`. Override it when needed:

```sh
make run-agent-transport AGENT_SERIAL_SOCKET=/tmp/my-davos-agent.sock
```

After QEMU starts, attach the one-shot host probe in another terminal:

```sh
python3 scripts/agent_transport_probe.py --socket /tmp/davos-agent.sock
```

Then run this command in DavOS:

```text
agent transport ping
```

The guest sends a framed `ping`; the host validates it and returns a framed
`pong`. A successful round trip prints `Agent transport: connected`. The probe
contains no provider or networking logic and exists only to verify the channel.

## Wire frame

Every integer uses network byte order (big-endian).

| Offset | Size | Field | Value |
|---:|---:|---|---|
| 0 | 2 | Magic | ASCII `DV` |
| 2 | 1 | Version | `1` |
| 3 | 1 | Kind | `1` request, `2` response |
| 4 | 2 | Payload length | Number of payload bytes |
| 6 | N | Payload | Opaque bounded bytes |
| 6 + N | 2 | Checksum | CRC-16/CCITT-FALSE |

The checksum covers version, kind, payload length and payload. Magic bytes are
not included.

Limits:

- request payload: 2048 bytes maximum;
- response payload: 1024 bytes maximum;
- one in-flight request;
- no multiplexing or concurrent exchanges.

These limits are checked before payload bytes are copied. The guest uses only
fixed-size storage.

## Failure behavior

Bad magic, version, kind or checksum is rejected as malformed framing. A length
over the limit is rejected immediately. If a frame stops partway through, the
fixed serial poll budget expires and the exchange returns a partial-frame
error. A later exchange starts with a reset decoder and drains buffered input.

All read and write loops have a fixed upper bound, so an unavailable or broken
host cannot hold the shell indefinitely. `agent transport ping` reports the
transport error and returns to the prompt. Higher-level request timeouts,
bridge health and planner fallback remain part of issue #188.

## Trust boundary

The serial payload is untrusted transport data. Future bridge responses must
still pass the existing Agent plan validation, safety gate and constrained
executor. Provider API keys, HTTP calls and JSON parsing remain host-only.
