// Package protocol — bidirectional Claude↔Codex governed delegation wire
// protocol (delegation feature, Phase 1).
//
// This file defines the delegation MessageType payloads, the engine and
// lifecycle string constants, the additive approval-attribution fields, and the
// durable delegation-link schema replayed across daemon restarts.
//
// These wire types are MessageType frames (the outer AgentMessage.Type),
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

import "encoding/json"

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

	// MsgStartDelegation is sent by the PWA to the daemon to USER-initiate a
	// delegation (the "hand off" trigger): the daemon parks the named source
	// session (when Await) and spawns a delegate of ToEngine in the source's
	// working directory. This is the user-initiated counterpart to the
	// agent-autonomous MCP `delegate` tool path; both converge on the same
	// daemon spawn machinery. PWA→daemon command frame, forwarded opaquely by
	// the relay, discriminated by the outer AgentMessage.Type — same convention
	// as the other four delegation frames.
	MsgStartDelegation = "start_delegation"
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

	// DelegationStateRejected is a DelegationStatusPayload-only state (it is
	// NEVER persisted into PersistedDelegationLink.State — no link is ever
	// created for a rejected start). It reports that a user-initiated
	// start_delegation was refused by the daemon BEFORE any delegate session was
	// spawned (delegation disabled / killswitch off, admission-depth-cycle-writer
	// reject, unknown engine, source-not-found, quiescence failure). The frame is
	// keyed by SourceSID (the would-be source session, NOT a delegate — no
	// DelegateSID exists yet) and carries a redacted Error string the PWA
	// reconciles onto its per-source delegationStartErrors cell. This closes the
	// silent-failure gap where a rejected hand-off produced no negative frame and
	// the StartDelegationSheet had already closed optimistically.
	DelegationStateRejected = "rejected"
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
//
// Two usage shapes share this struct:
//   - Live link status (the original shape): keyed by DelegateSID, State is one
//     of pending | active | completed | cancelled.
//   - Daemon rejection of a user-initiated start_delegation (additive, backward
//     compatible): State=DelegationStateRejected, SourceSID identifies the
//     would-be source session, Error carries a redacted human-readable reason,
//     and DelegateSID is EMPTY (no delegate was ever spawned). SourceSID and
//     Error are both omitempty, so every pre-existing live-status emit marshals
//     to exactly the bytes it did before this field pair was added — old daemons
//     and old PWAs stay byte-compatible.
type DelegationStatusPayload struct {
	DelegateSID string `json:"delegate_sid"`           // delegate agent session ID (EMPTY on a rejection frame)
	State       string `json:"state"`                  // DelegationState* — pending | active | completed | cancelled | rejected
	DiffSummary string `json:"diff_summary,omitempty"` // optional human-readable summary of changes so far
	UpdatedAt   int64  `json:"updated_at,omitempty"`   // unix milliseconds
	// SourceSID is populated ONLY on a DelegationStateRejected frame to key the
	// rejection to the would-be source session (a normal live-status frame leaves
	// it empty; omitempty drops it so the wire bytes are unchanged for the live
	// path). The PWA reconciles a rejection onto delegationStartErrors[source_sid].
	SourceSID string `json:"source_sid,omitempty"`
	// Error is the redacted, human-readable rejection reason surfaced on the
	// source SessionCard. Populated ONLY on a DelegationStateRejected frame;
	// omitempty drops it for the live-status path. It MUST be a safe,
	// client-appropriate string (no internal paths / state) per the
	// redact-before-PWA error-handling rule.
	Error string `json:"error,omitempty"`
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

// StartDelegationPayload is the PWA→daemon command that USER-initiates a
// delegation. Carried inside an AgentMessage envelope with
// Type=MsgStartDelegation. The daemon resolves the source session, parks it
// (when Await), and spawns a delegate of ToEngine in the source's working
// directory with Prompt as the delegate's first task.
//
// Await default note: the JSON bool zero-value is false, but the governed
// product default is await=TRUE when the field is ABSENT. omitempty drops an
// explicit false, so an absent field and an explicit false are
// indistinguishable on the wire. The PWA start-delegation UI ALWAYS sends await
// explicitly (Completion 2), so the daemon's start_delegation decoder treats the
// PWA as always-explicit. The daemon-side handler resolves the absent⇒true
// default by decoding into a *bool / inspecting the raw JSON, NOT by reading
// this struct's Go zero value. If a future non-PWA producer needs absent⇒true,
// switch this field to *bool then; keep omitempty so old consumers stay
// byte-compatible.
type StartDelegationPayload struct {
	SourceSID string `json:"source_sid"`      // source session to park + delegate from
	ToEngine  string `json:"to_engine"`       // EngineClaude | EngineCodex — the delegate engine
	Prompt    string `json:"prompt"`          // task prompt delivered to the delegate as its first message
	Await     bool   `json:"await,omitempty"` // park source + return result to source agent; daemon defaults absent⇒true
}

// StartDelegationAwaitOrDefault resolves the governed await semantics for a
// start_delegation frame from the RAW JSON bytes, applying the absent⇒true rule
// that the plain-bool StartDelegationPayload.Await field cannot express on its
// own (omitempty drops an explicit false, so an absent field and an explicit
// false are indistinguishable after decoding into the struct).
//
// This is the start_delegation analog of delegation.DelegateInput.AwaitOrDefault
// for the MCP path: it is the SINGLE place the absent⇒true rule is applied for
// the user-initiated trigger, so both daemon dispatch handlers (local WS and
// relay) resolve await identically (pattern-consistency dual-path) and a future
// non-PWA producer that omits await still gets the safe parked default rather
// than the dangerous fire-and-forget false.
//
// Resolution:
//   - the "await" key is ABSENT, JSON null, or the bytes do not decode      ⇒ true
//   - the "await" key is present with an explicit boolean (true|false)      ⇒ that value
//
// A malformed frame is treated as absent (⇒ true): the safe default is to park,
// and the caller has already validated the frame's required fields separately.
func StartDelegationAwaitOrDefault(raw json.RawMessage) bool {
	var probe struct {
		Await *bool `json:"await"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || probe.Await == nil {
		return true
	}
	return *probe.Await
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
	// Await is the governed delegation lifecycle flag persisted so restart recovery
	// can distinguish an await=true (parked source, returns result) delegation from
	// an await=false (fire-and-forget, source never parked) one. It is a *bool with
	// the same absent⇒true idiom as StartDelegationPayload.Await /
	// StartDelegationAwaitOrDefault: a NIL (absent) value resolves to true on
	// recovery, so a journal written before this field existed recovers as the safe
	// parked default. omitempty drops a nil pointer entirely, so old daemons and old
	// journals stay byte-compatible. Without this field, recovery hardcoded
	// await=true and re-parked a fire-and-forget source — its turns would queue, the
	// quiescence clamp would fire, and an unwanted synthetic result would be injected.
	Await *bool `json:"await,omitempty"`
}
