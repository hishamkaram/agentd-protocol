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

## Public Type Surface

The module groups wire contracts by feature area instead of exposing repo-specific aliases:

| Area | Representative types | Purpose |
|------|----------------------|---------|
| Relay transport | `RelayEnvelope`, `ControlMessage`, `ControlType` | Encrypted relay envelopes plus register/join/status/audit/key-rotation/push/termination control frames |
| Route receipts | `RouteReceiptPayload`, `CommandReceiptPayload`, `CommandReceiptStage`, `CommandReceiptReasonCode` | Transport and command acknowledgement metadata without decrypted payload content |
| Replay and recovery | `ReplayRequest`, `SessionHeadPayload`, `HistoryPageRequest`, `HistoryPagePayload`, `ReplayCompletePayload`, `SessionSnapshotRequest`, `SessionSnapshotPayload`, `HistoryReplayCompletePayload` | Durable session journal replay, route reopen snapshots, and history pagination diagnostics |
| Protocol v2 | `JSONRPCRequest`, `JSONRPCResponse`, `JSONRPCError`, `ProtocolHelloParams`, `ProtocolHelloResult`, `TransportPingParams`, `TransportPongResult`, `AgentDEventEnvelope`, `ReplayCursor` | JSON-RPC framing, protocol negotiation, ping/pong, and v2 event envelopes |
| Session features | `AgentCapability`, `SessionFeatureStatusPayload`, `SessionInfo`, `CodexSandboxMode`, `RateLimitsUpdatedPayload`, `RateLimitBucket` | Per-agent capability and feature-health surfaces consumed by the PWA |
| Git and worktrees | `GitStatusPayload`, `GitDiffRequest`, `GitDiffResponse`, git sync request/response types, `GitWorktree*`, `SessionWorktree*` | Git status, diff, branch/stash/fetch/pull/push, and worktree management contracts |
| MCP | `MCPServerConfig`, `MCPServerStatusEntry`, `MCPListPayload`, `MCPMutationPayload`, `MCPServersChangedPayload` | MCP server inventory and mutation wire types |
| Cost, prompt, and approvals | `CostPayload`, `InteractivePromptResolvedPayload`, `ApprovalResolvedPayload` | Runtime cost samples, interactive prompt completion, and approval tombstones |
| Support bundles | `SupportBundleParams`, `SupportBundle`, `SupportBundleClient`, `SupportBundleDaemon`, `SupportRelayDiagnostics`, `SupportHistoryReplayState` | Redacted client/daemon/relay diagnostics for support export flows |
| Policies and tracing | `PolicyJSON`, `PolicyMatchJSON`, `NewTraceID()`, `ValidTraceID()` | Policy sync payloads and W3C-compatible trace IDs |

### Control Types

```
register · join · heartbeat · ack · error · sync_policies
status_update · audit_entry · deactivate_developer · client_connected
client_count · key_rotate · entitlement_update · entitlement_violation
push_notify · push_notify_result
terminate_session · terminate_session_ack
route_receipt
```

### Payload Types

```
RegisterPayload · JoinPayload · AckPayload · ErrorPayload
StatusUpdatePayload · AuditEntryPayload · DeactivateDeveloperPayload
TerminateSessionPayload · TerminateSessionAckPayload
ClientConnectedPayload · ClientCountPayload · KeyRotatePayload
SyncPoliciesPayload · EntitlementUpdatePayload · EntitlementViolationPayload
PushNotifyPayload · PushNotifyResultPayload · RouteReceiptPayload
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

Replay recovery is daemon-owned: the PWA sends `ReplayRequest{after_seq}` when it detects a gap in the inner per-session `AgentMessage.seq`, and the daemon responds with replayed `AgentMessage` frames followed by `replay_complete`. Route-bound reconnects can request `SessionSnapshotRequest` first so the active chat view is rebuilt from daemon state before broad replay. `RelayEnvelope.seq` remains the outer transport sequence for relay delivery and is not the cursor used for UI replay.

Relay joins may also carry optional `nav_session_id` metadata. The relay forwards that value in `ClientConnectedPayload` so the daemon can prioritize the active session snapshot, while legacy clients omit the field and keep the existing replay behavior. New relays also assign a per-connection `client_id`, return it in `AckPayload`, and forward it in `ClientConnectedPayload`; daemon replay can set `RelayEnvelope.client_id` to deliver join-bound history only to that PWA connection.

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
