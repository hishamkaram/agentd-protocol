# Changelog

All notable changes to `agentd-protocol` (the shared wire-protocol module
for AgentD) are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Versioning convention (see [README.md](./README.md) for full policy):

- **MAJOR** (`v0.x → v1.x`, then `v1 → v2` requires `/v2` import path) — stable-contract break after v1.0.
- **MINOR** (`v0.1.0 → v0.2.0`) — pre-v1 breaking cleanup, new message types, new fields on existing structs, new capability flags.
- **PATCH** (`v0.1.0 → v0.1.1`) — no wire-format change: doc fixes, test improvements, dependency bumps.

## [Unreleased]

### Wire types

- Add `list_provider_catalogs` / `provider_catalog_list` messages with
  provider-qualified model and canonical runtime-control catalogs, content
  generations, CLI provenance, and sanitized discovery errors. Provider-native
  control spellings remain daemon-side and are never serialized.
- Additive recoverable-session fields and enum values: `budget_exceeded`,
  `provider_limit`, `hosted_capacity`, `wait_then_retry`,
  `increase_budget_then_retry`, plus optional `SessionRecoveryInfo.code` and
  `SessionRecoveryInfo.retry_after_ms`.
- Additive `SessionInfo.authored_agent_count` field for reporting the number of
  provider-accepted authored custom agents available to Claude.

## [v0.1.3] — 2026-06-13

### Documentation

- Preserve historical changelog symbols for tagged pre-v1 wire inventory.

## [v0.1.2] — 2026-06-12

### Wire types

- Breaking pre-v1 cleanup: removed obsolete transcript replay and snapshot wire
  contracts after coordinated migration to selected-session `session_head` +
  `history_page` paging.
- Replaced the historical `MsgPendingApprovalReplay` reconnect marker with
  current control-state `MsgPendingApprovalState`.
- `CtrlPushNotifyResult` control message for relay-to-daemon Web Push delivery results
- `PushNotifyPayload` tracking fields: `notification_id`, `trace_id`, `created_at_unix_ms`, `attempt`
- `PushNotifyResultPayload` for accepted/retryable/permanent Web Push result acknowledgements
- `CostPayload` shared wire type with additive cached-token, pricing-availability, and Codex credit estimate fields
- `RateLimitsUpdatedPayload` additive Codex plan and actual credit snapshot fields
- `AgentCapability.supports_runtime_full_access` additive capability flag for Codex runtime Full access opt-in
- `SupportBundleParams.client_transport` and additive support-bundle transport diagnostics for active session, bootstrap state, relay control errors, inbound/pong age, and pending JSON-RPC count
- `ErrorPayload.retryable`, `ErrorPayload.terminal`, and `ErrorPayload.retry_after_ms` relay control retry metadata for reconnecting, overloaded, and terminal pairing states
- `ClientConnectedPayload.client_id` is now required on current relay-to-daemon `client_connected` controls for targeted bootstrap and history responses
- `SupportBundleTransport.relay_diagnostics` redacted relay pairing, reconnect, history hydration, and pressure counters
- `SessionInfo.recovery`, `StatusPayload.recovery`, `SessionRecoveryInfo`, `SessionRecoveryReason`, and `SessionRecoveryAction` for durable paused session repair metadata

## [v0.1.0] — 2026-04-29

Initial tagged release. Bundles all wire types accumulated since the
module's inception. Captures the protocol surface as of agentd commit
`4818a19` (post-feature-195 ship).

### Wire types

#### Envelope and control plane

- `RelayEnvelope` — outer envelope for relay-routed messages (`envelope.go`)
- `ControlMessage`, `ControlType` enum (`control.go`)
- `RegisterPayload`, `JoinPayload`, `AckPayload`, `ErrorPayload`, `StatusUpdatePayload`, `AuditEntryPayload`, `DeactivateDeveloperPayload`, `ClientConnectedPayload`, `ClientCountPayload`, `KeyRotatePayload`, `SyncPoliciesPayload`

#### Per-agent capabilities (features 185 / 186 / 195)

- `AgentCapability` struct with 9 fields:
  - `MCPHotApply`, `MCPReconnect`, `MCPLiveStatusLimited` (feature 185)
  - `AnswerQuestion`, `SendToolResult`, `RewindFiles` (feature 185)
  - `SessionScopedApproval`, `AnswerQuestionFreeText` (feature 186)
  - `SupportsBypassPermissions` (feature 195)

#### Approval lifecycle (feature 193)

- `MsgApprovalResolved`, historical `MsgPendingApprovalReplay` constants
- `ApprovalResolvedPayload`

#### History hydration (feature 192)

- Historical `MsgHistoryReplayComplete` constant
- Historical `HistoryReplayCompletePayload`

#### Rate limits (feature 184)

- `MsgRateLimitsUpdated` constant
- `RateLimitBucket`, `RateLimitsUpdatedPayload`

#### Git status / changes viewer (feature 160 / 162)

- `GitFileStatus`, `GitStatusPayload`, `GitStatusRequest`
- `GitDiffRequest`, `GitDiffResponse`
- `GitNotAvailablePayload`

#### Git sync (feature 172)

- `Msg{GitBranchList, GitBranchListResponse, GitBranchSwitch, GitBranchSwitchResponse, GitFetch, GitFetchResponse, GitPull, GitPullResponse, GitPush, GitPushResponse}` constants

#### Git worktrees (feature 173)

- `Msg{GitWorktreeList, GitWorktreeListResponse, GitWorktreeAdd, GitWorktreeAddResponse, GitWorktreeAdded, GitWorktreeRemove, GitWorktreeRemoveResponse, GitWorktreeRemoved, GitWorktreeLock, …}` constants

#### MCP server management (feature 169 / 170)

- `MCPServerScope` enum
- `MCPServerConfig`, `MCPServerStatusEntry`
- `MCPListPayload`, `MCPListResponse`
- `MCPMutationPayload`, `MCPMutationResponse`
- `MCPRemovePayload`, `MCPTogglePayload`

#### Policies and tracing

- `PolicyJSON`, `PolicyMatchJSON` (`policy.go`)
- Tracing helpers (`trace.go`)

### Backwards-compatibility

- All capability fields are optional on the consumer side (TS and Go).
  Older PWA bundles ignore unknown fields; older daemons that don't
  serialize a field surface as `false` / zero-value to newer PWA bundles.
- Protocol additions in this release are 100% wire-additive — no
  consumer that compiled against an earlier pseudo-version will break
  when re-pinned to `v0.1.2`.

[Unreleased]: https://github.com/hishamkaram/agentd-protocol/compare/v0.1.3...HEAD
[v0.1.3]: https://github.com/hishamkaram/agentd-protocol/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/hishamkaram/agentd-protocol/compare/v0.1.0...v0.1.2
[v0.1.0]: https://github.com/hishamkaram/agentd-protocol/releases/tag/v0.1.0
