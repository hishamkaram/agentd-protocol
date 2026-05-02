# Changelog

All notable changes to `agentd-protocol` (the shared wire-protocol module
for AgentD) are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Versioning convention (see [README.md](./README.md) for full policy):

- **MAJOR** (`v0.x → v1.x`, then `v1 → v2` requires `/v2` import path) — removed field, renamed JSON tag, removed message type, type-shape change.
- **MINOR** (`v0.1.0 → v0.2.0`) — wire-additive: new message types, new fields on existing structs, new capability flags. Pre-v1 minor bumps may also include localized breaking changes per the Go "anything goes pre-v1" convention.
- **PATCH** (`v0.1.0 → v0.1.1`) — no wire-format change: doc fixes, test improvements, dependency bumps.

## [Unreleased]

### Wire types

- `MsgReplayRequest`, `MsgReplayComplete` constants for durable journal recovery
- `ReplayRequest`, `ReplayCompletePayload`

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

- `MsgApprovalResolved`, `MsgPendingApprovalReplay` constants
- `ApprovalResolvedPayload`, `PendingApprovalReplayPayload`

#### History replay (feature 192)

- `MsgHistoryReplayComplete` constant
- `HistoryReplayCompletePayload`

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
  when re-pinned to `v0.1.0`.

[Unreleased]: https://github.com/hishamkaram/agentd-protocol/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/hishamkaram/agentd-protocol/releases/tag/v0.1.0
