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
| Device recovery | `ClientAuthProof`, `ClientAuthEnrollPayload`, `ClientAuthRevokePayload`, `ClientKeySyncRequestPayload`, `ClientKeySyncResponsePayload`, `P256PublicJWK` | Sender-constrained browser joins, public-key enrollment and revocation, and opaque traffic-key recovery after rotation |
| Route receipts | `RouteReceiptPayload`, `CommandReceiptPayload`, `CommandReceiptStage`, `CommandReceiptReasonCode` | Transport and command acknowledgement metadata without decrypted payload content |
| Transcript recovery | `SessionHeadPayload`, `HistoryPageRequest`, `HistoryPagePayload` | Bounded selected-session transcript pagination and daemon-proven journal head metadata |
| Pending turns | `PendingTurnStatePayload`, `PendingTurnItem`, `PendingTurnStage` | Revisioned replacement snapshots for durable provider-bound user turns without receipt metadata leakage |
| Protocol v2 | `JSONRPCRequest`, `JSONRPCResponse`, `JSONRPCError`, `ProtocolHelloParams`, `ProtocolHelloResult`, `TransportPingParams`, `TransportPongResult`, `AgentDEventEnvelope` | JSON-RPC framing, protocol negotiation, ping/pong, and v2 event envelopes |
| Session features | `AgentCapability`, `SessionFeatureStatusPayload`, `SessionInfo`, `SessionRecoveryInfo`, `CodexSandboxMode`, `RateLimitsUpdatedPayload`, `RateLimitBucket` | Per-agent capability, feature-health, and recoverability surfaces consumed by the PWA |
| Provider catalogs | `ProviderRuntimeCatalog`, `ProviderModelInfo`, `ProviderControlOption`, `ListProviderCatalogsRequest`, `ListProviderCatalogsResponse` | Provider-qualified CLI discovery results with sanitized errors and generation-bound startup controls |
| Git and worktrees | `GitStatusPayload`, `GitDiffRequest`, `GitDiffResponse`, git sync request/response types, `GitWorktree*`, `SessionWorktree*` | Git status, diff, branch/stash/fetch/pull/push, and worktree management contracts |
| MCP | `MCPServerConfig`, `MCPServerStatusEntry`, `MCPListPayload`, `MCPMutationPayload`, `MCPServersChangedPayload` | MCP server inventory and mutation wire types |
| Cost, prompt, and approvals | `CostPayload`, `InteractivePromptResolvedPayload`, `ApprovalResolvedPayload` | Runtime cost samples, interactive prompt completion, and approval tombstones |
| Support bundles | `SupportBundleParams`, `SupportBundle`, `SupportBundleClient`, `SupportBundleDaemon`, `SupportRelayDiagnostics`, `SupportHistoryHydrationState` | Redacted client/daemon/relay diagnostics for support export flows |
| Policies and tracing | `PolicyJSON`, `PolicyMatchJSON`, `NewTraceID()`, `ValidTraceID()` | Policy sync payloads and W3C-compatible trace IDs |

### Control Types

```
register · join · heartbeat · ack · error · sync_policies
status_update · audit_entry · deactivate_developer · client_connected
client_count · key_rotate · entitlement_update · entitlement_violation
client_auth_enroll · client_auth_enroll_ack · client_auth_revoke
client_auth_revoke_ack · client_key_sync_request · client_key_sync_response
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
ClientAuthEnrollPayload · ClientAuthEnrollAckPayload
ClientAuthRevokePayload · ClientAuthRevokeAckPayload
ClientKeySyncRequestPayload · ClientKeySyncResponsePayload
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

Transcript recovery is selected-session and bounded. The daemon emits
`session_head` metadata during connection bootstrap, and the PWA requests
`history_page` slices for the active session only. `RelayEnvelope.seq` remains
the outer transport sequence for relay delivery and is not the transcript cursor.

Relay control errors carry structured `retryable`, `terminal`, and
`retry_after_ms` fields so clients can distinguish transient daemon reconnects
from terminal pairing/session failures without inferring from socket state alone.

`client_auth_v1` and `client_key_sync_v1` are additive capabilities. Current
clients can enroll P-256 signing and key-agreement public keys, authenticate
joins with fixed-width ECDSA proofs, and recover a rotated traffic key through
an ECDH/HKDF/AES-GCM response that remains opaque to the relay. Legacy HMAC
joins remain wire-compatible for clients that have not enrolled a device.
`ClientAuthEnrollPayload.client_id` is also additive: the daemon sets it only
from relay-authenticated source metadata, while durable background
reconciliation leaves it absent.
`RelayEnvelope.key_epoch` stamps the base traffic-key epoch; an absent or zero
value is the legacy initial epoch.

Recoverable session pauses carry `SessionRecoveryInfo` on both `SessionInfo` and
status payloads. The additive `code` and `retry_after_ms` fields let clients
distinguish budget, provider-limit, and hosted-capacity pauses while older
clients continue to render `state: "paused"` plus the human message.

## Design Principles

- **Zero third-party dependencies** — protocol validation and cryptographic byte contracts use only the Go standard library
- **Pre-v1 lockstep** — coordinated contract removals and additions are expected while AgentD is still in foundation development
- **W3C trace IDs** — 32 lowercase hex chars, ready for OpenTelemetry upgrade
- **Roundtrip tested** — every type has a JSON marshal/unmarshal roundtrip test

## Versioning

This module follows [Semantic Versioning](https://semver.org/) with one
nuance: while in `v0.x` (initial development), coordinated AgentD wire
contract removals and wire-additive changes both bump the **MINOR** version.
Consumers (`agentd`, `agentd-relay`, and `agentd-web`) are expected to update
in lockstep across pre-v1 breaking cleanups.

| Bump | Trigger |
|---|---|
| **MAJOR** (`v0.x → v1.0`, `v1 → v2` requires `/v2` import path) | Stable-contract break after v1.0 |
| **MINOR** (`v0.1.0 → v0.2.0`) | Pre-v1 breaking cleanup, new message type, new field on existing struct, new capability flag |
| **PATCH** (`v0.1.0 → v0.1.1`) | No wire-format change: doc fix, test improvement, dep bump |

See [`CHANGELOG.md`](./CHANGELOG.md) for the per-release wire-type
inventory. Tags are pushed manually after a wire-additive PR merges to
`main`. Consumers may pin via the tag (`v0.1.0`) or via a Go
pseudo-version (`v0.0.0-<timestamp>-<sha>`) — both are supported.

## License

[MIT](LICENSE)
