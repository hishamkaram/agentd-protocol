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
| `control.go` | `ControlMessage`, `ControlType` (10 constants), 9 payload structs | Relay control protocol |
| `policy.go` | `PolicyJSON`, `PolicyMatchJSON` | Policy rule wire format |
| `trace.go` | `NewTraceID()`, `ValidTraceID()` | W3C-compatible trace ID generation |

### Control Types

```
register · join · heartbeat · ack · error · sync_policies
status_update · audit_entry · deactivate_developer · client_connected
```

### Payload Types

```
RegisterPayload · JoinPayload · AckPayload · ErrorPayload
StatusUpdatePayload · AuditEntryPayload · DeactivateDeveloperPayload
ClientConnectedPayload · SyncPoliciesPayload
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

## Design Principles

- **Zero dependencies** — only `encoding/json`, `time`, `crypto/rand`, `encoding/hex` from stdlib
- **Backward compatible** — all new fields use `omitempty`; old clients produce identical JSON
- **W3C trace IDs** — 32 lowercase hex chars, ready for OpenTelemetry upgrade
- **Roundtrip tested** — every type has a JSON marshal/unmarshal roundtrip test

## License

[MIT](LICENSE)
