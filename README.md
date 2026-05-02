<p align="center">
  <img src="docs/icon.svg" width="80" height="80" alt="AgentD">
</p>

<h1 align="center">agentd-protocol</h1>

<p align="center">
  Shared wire protocol types for the <a href="https://github.com/hishamkaram/agentd">AgentD</a> ecosystem
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/hishamkaram/agentd-protocol"><img src="https://pkg.go.dev/badge/github.com/hishamkaram/agentd-protocol.svg" alt="Go Reference"></a>
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="MIT License">
  <img src="https://img.shields.io/badge/dependencies-zero-brightgreen.svg" alt="Zero Dependencies">
</p>

---

## What This Is

A tiny Go module containing the shared wire protocol types used by both the [AgentD daemon](https://github.com/hishamkaram/agentd) and [AgentD relay server](https://github.com/hishamkaram/agentd-relay). By importing from a single source of truth, wire format drift between repos is impossible — the Go compiler enforces type identity.

## Types

| File | Types | Purpose |
|------|-------|---------|
| `envelope.go` | `RelayEnvelope` | Encrypted message envelope (sid, seq, enc, tid) |
| `control.go` | `ControlMessage`, `ControlType` (12 constants), 11 payload structs | Relay control protocol |
| `policy.go` | `PolicyJSON`, `PolicyMatchJSON` | Policy rule wire format |
| `capabilities.go` | `AgentCapability` | Per-agent feature flags the daemon emits to the PWA (MCP reconnect, session-scoped approvals, free-text replies, etc.) |
| `replay.go` | `ReplayRequest`, `ReplayCompletePayload` | Durable session journal replay recovery using inner per-session `AgentMessage.seq` values |
| `trace.go` | `NewTraceID()`, `ValidTraceID()` | W3C-compatible trace ID generation |

### Control Types

```
register · join · heartbeat · ack · error · sync_policies
status_update · audit_entry · deactivate_developer · client_connected
client_count · key_rotate
```

### Payload Types

```
RegisterPayload · JoinPayload · AckPayload · ErrorPayload
StatusUpdatePayload · AuditEntryPayload · DeactivateDeveloperPayload
ClientConnectedPayload · ClientCountPayload · KeyRotatePayload
SyncPoliciesPayload
```

## Usage

```go
import protocol "github.com/hishamkaram/agentd-protocol"

env := protocol.RelayEnvelope{
    SessionID: "sess-123",
    Seq:       1,
    Encrypted: ciphertext,
    TraceID:   protocol.NewTraceID(),
}
```

Both `agentd` and `agentd-relay` use type aliases to re-export these types under their `internal/relay` package, so existing code using `relay.RelayEnvelope` continues to work unchanged.

Replay recovery is daemon-owned: the PWA sends `ReplayRequest{after_seq}` when it detects a gap in the inner per-session `AgentMessage.seq`, and the daemon responds with replayed `AgentMessage` frames followed by `replay_complete`. `RelayEnvelope.seq` remains the outer transport sequence for relay delivery and is not the cursor used for UI replay.

## Design Principles

- **Zero dependencies** — only `encoding/json`, `time`, `crypto/rand`, `encoding/hex` from stdlib
- **Backward compatible** — all new fields use `omitempty`; old clients produce identical JSON
- **W3C trace IDs** — 32 lowercase hex chars, ready for OpenTelemetry upgrade
- **Roundtrip tested** — every type has a JSON marshal/unmarshal roundtrip test

## Versioning

This module follows [Semantic Versioning](https://semver.org/) with one
nuance: while in `v0.x` (initial development), wire-additive changes
bump the **MINOR** version, not patch — since consumers (`agentd`,
`agentd-relay`) are expected to opt in to each new field.

| Bump | Trigger |
|---|---|
| **MAJOR** (`v0.x → v1.0`, `v1 → v2` requires `/v2` import path) | Removed field, renamed JSON tag, removed message type |
| **MINOR** (`v0.1.0 → v0.2.0`) | Wire-additive change: new message type, new field on existing struct, new capability flag |
| **PATCH** (`v0.1.0 → v0.1.1`) | No wire-format change: doc fix, test improvement, dep bump |

See [`CHANGELOG.md`](./CHANGELOG.md) for the per-release wire-type
inventory. Tags are pushed manually after a wire-additive PR merges to
`main`. Consumers may pin via the tag (`v0.1.0`) or via a Go
pseudo-version (`v0.0.0-<timestamp>-<sha>`) — both are supported.

## License

[MIT](LICENSE)
