// Package protocol — bidirectional Claude↔Codex governed delegation wire
// protocol (delegation feature, Phase 1).
//
// This file defines the four delegation MessageType payloads, the engine and
// lifecycle string constants, the additive approval-attribution fields, and the
// durable delegation-link schema replayed across daemon restarts.
//
// These four wire types are MessageType frames (the outer AgentMessage.Type),
// NOT ControlType frames. The relay forwards them opaquely; the daemon and PWA
// discriminate by the outer AgentMessage.Type. They follow the same
// daemon-is-source-of-truth convention as the approval wire types in
// approval.go: matching constants live in the daemon's
// agentd/internal/session/ with init() panic cross-checks at startup (Phase 3).
//
// Engine values are the stable wire strings "claude" | "codex". All optional
// fields use json:",omitempty" so older producers/consumers stay byte
// compatible (backward compatible per CLAUDE.md "Backward Compatible").
package protocol

// ─── Engine Constants ──────────────────────────────────────────────────────────

// Engine values identify which underlying AI agent owns a delegation endpoint.
// These are stable wire strings shared by every delegation payload's
// *Engine field; changing either is a breaking change across all 3 repos.
const (
	// EngineClaude identifies a Claude Code agent endpoint.
	EngineClaude = "claude"

	// EngineCodex identifies an OpenAI Codex agent endpoint.
	EngineCodex = "codex"
)

// ─── MessageType Constants ─────────────────────────────────────────────────────
//
// The DAEMON is the source of truth for these constants. Matching constants
// (with init() panic cross-checks at daemon startup) land in
// agentd/internal/session/ in Phase 3 — the same drift-prevention pattern as
// approval.go ↔ wsserver_approval.go.

const (
	// MsgDelegationLink is sent by the source to the delegate over the relay
	// when a delegation begins. Carries both session identities and engines.
	MsgDelegationLink = "delegation_link"

	// MsgDelegationStatus is sent by the delegate to the source over the relay
	// to report link state changes (pending → active → completed → cancelled).
	MsgDelegationStatus = "delegation_status"

	// MsgDelegationResult is sent by the delegate to the source when the
	// delegation reaches a terminal status. Carries the final status, a
	// human-readable summary, and an optional pass/fail classification.
	MsgDelegationResult = "delegation_result"

	// MsgDelegationCancel is sent by the source to the delegate to terminate
	// an active link.
	MsgDelegationCancel = "delegation_cancel"
)

// ─── Delegation Lifecycle Constants ─────────────────────────────────────────────
//
// These document the canonical string values that appear inside the delegation
// payloads. The daemon and PWA serialize/deserialize from these sets. Values
// are stable wire strings — changing any of them is a breaking change.

// Link lifecycle states carried in DelegationStatusPayload.State and the
// durable PersistedDelegationLink.State.
const (
	DelegationStatePending   = "pending"
	DelegationStateActive    = "active"
	DelegationStateCompleted = "completed"
	DelegationStateCancelled = "cancelled"
)

// Terminal statuses carried in DelegationResultPayload.Status.
const (
	DelegationResultCompleted = "completed"
	DelegationResultError     = "error"
	DelegationResultCancelled = "cancelled"
)

// Pass/fail classification carried in DelegationResultPayload.PassFail.
const (
	DelegationPassFailPass = "pass"
	DelegationPassFailFail = "fail"
)

// Triggers carried in DelegationLinkPayload.TriggeredBy and
// PersistedDelegationLink.TriggeredBy — who initiated the delegation.
const (
	DelegationTriggeredByUser   = "user"
	DelegationTriggeredByAuto   = "auto"
	DelegationTriggeredBySystem = "system"
)

// ─── Wire Payloads ───────────────────────────────────────────────────────────

// DelegationLinkPayload initiates a delegation from the source agent to the
// delegate agent. Carried inside an AgentMessage envelope with
// Type=MsgDelegationLink.
type DelegationLinkPayload struct {
	SourceSID      string `json:"source_sid"`             // source agent session ID
	SourceEngine   string `json:"source_engine"`          // EngineClaude | EngineCodex
	DelegateSID    string `json:"delegate_sid"`           // delegate agent session ID
	DelegateEngine string `json:"delegate_engine"`        // EngineClaude | EngineCodex
	WorkDir        string `json:"work_dir,omitempty"`     // working directory scope for the delegation
	TriggeredBy    string `json:"triggered_by,omitempty"` // DelegationTriggeredBy* — user | auto | system
	CreatedAt      int64  `json:"created_at"`             // unix milliseconds
}

// DelegationStatusPayload reports a link state change from the delegate back to
// the source. Carried inside an AgentMessage envelope with
// Type=MsgDelegationStatus.
type DelegationStatusPayload struct {
	DelegateSID string `json:"delegate_sid"`           // delegate agent session ID
	State       string `json:"state"`                  // DelegationState* — pending | active | completed | cancelled
	DiffSummary string `json:"diff_summary,omitempty"` // optional human-readable summary of changes so far
	UpdatedAt   int64  `json:"updated_at,omitempty"`   // unix milliseconds
}

// DelegationResultPayload terminates a delegation with its final status.
// Carried inside an AgentMessage envelope with Type=MsgDelegationResult.
type DelegationResultPayload struct {
	DelegateSID string `json:"delegate_sid"`           // delegate agent session ID
	Status      string `json:"status"`                 // DelegationResult* — completed | error | cancelled
	Summary     string `json:"summary"`                // human-readable result summary
	DiffSummary string `json:"diff_summary,omitempty"` // final changeset summary
	PassFail    string `json:"pass_fail,omitempty"`    // DelegationPassFail* — pass | fail completion classification
}

// DelegationCancelPayload requests termination of an active link. Carried inside
// an AgentMessage envelope with Type=MsgDelegationCancel.
type DelegationCancelPayload struct {
	DelegateSID string `json:"delegate_sid"`     // delegate agent session ID
	Reason      string `json:"reason,omitempty"` // optional cancellation reason
}

// ─── Approval Attribution (additive wire fields) ────────────────────────────────

// ApprovalAttribution pins the additive attribution fields that the daemon
// embeds into its ApprovalPayload (defined in agentd/internal/session/types.go)
// so a delegated approval can be traced back to the source agent that triggered
// it. Both fields use omitempty: an approval emitted without attribution
// marshals to exactly the bytes it did before this feature shipped, so old PWAs
// and old daemons remain fully compatible.
//
// The daemon does NOT use this struct by value on the wire; it embeds the same
// two JSON fields directly into its richer ApprovalPayload (Phase 3). This pin
// keeps the field names and omitempty contract authoritative in the shared
// module — the same cross-repo additive-field pattern as StatusPayload /
// SessionRecoveryInfo in session_recovery.go.
type ApprovalAttribution struct {
	SourceSID    string `json:"source_sid,omitempty"`    // session ID of the source agent that triggered the approval
	SourceEngine string `json:"source_engine,omitempty"` // EngineClaude | EngineCodex
}

// ─── Durable Persistence Schema (Phase 1 wire portion) ───────────────────────────

// PersistedDelegationLink is the minimal durable representation of a
// source↔delegate link. The daemon persists this and replays it on restart so
// it can reconstruct the parked-source / live-delegate relationship (Phase 3
// recovery, C3); a reloaded PWA renders it in link history (Phase 4, C7).
//
// Timestamps are unix milliseconds (int64) — matching the existing module
// convention (ApprovalResolvedPayload.ResolvedAt, DelegationLinkPayload.CreatedAt)
// and avoiding the time.Time monotonic-clock roundtrip trap. All optional fields
// use omitempty for forward/backward compatibility across daemon restarts.
//
// Phase 3 (daemon repo, NOT here): the daemon's internal persisted struct in
// agentd/internal/session/persistence.go adds daemon-specific fields that are
// NOT part of this shared schema — e.g. CancelledAt, CancelReason, FinalStatus,
// and retry bookkeeping. Only the wire/shared portion lives in agentd-protocol.
type PersistedDelegationLink struct {
	SourceSID      string `json:"source_sid"`             // source agent session ID
	SourceEngine   string `json:"source_engine"`          // EngineClaude | EngineCodex
	DelegateSID    string `json:"delegate_sid"`           // delegate agent session ID
	DelegateEngine string `json:"delegate_engine"`        // EngineClaude | EngineCodex
	WorkDir        string `json:"work_dir,omitempty"`     // working directory scope for the delegation
	TriggeredBy    string `json:"triggered_by,omitempty"` // DelegationTriggeredBy* — user | auto | system
	State          string `json:"state,omitempty"`        // DelegationState* — pending | active | completed | cancelled
	CreatedAt      int64  `json:"created_at"`             // unix milliseconds
	UpdatedAt      int64  `json:"updated_at,omitempty"`   // unix milliseconds
}
