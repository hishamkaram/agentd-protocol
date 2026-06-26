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

// knownEngines is the canonical, ordered allow-list of recognized delegation
// engine wire strings. It is the SINGLE source of truth shared by every module
// that must validate an engine string: internal/delegation gates on it via
// IsKnownEngine (the delegation package CANNOT import the daemon's
// internal/session — a depguard boundary forbids delegation→session), and
// internal/session's engineAgentPairs bijection is drift-guarded against
// KnownEngines() so the allow-list and the engine↔AgentType mapping never
// silently diverge across the 3 modules.
//
// It is a package-level slice literal (not a mutable var the callers handle
// directly): IsKnownEngine reads it for membership and KnownEngines returns a
// fresh copy on each call, so no caller can mutate the canonical set. WHEN
// ADDING A NEW ENGINE, add its Engine* const here AND its {engine, AgentType}
// pair to session.engineAgentPairs — the drift-guard test fails until both
// land.
var knownEngines = []string{EngineClaude, EngineCodex}

// IsKnownEngine reports whether engine is one of the recognized delegation
// engine wire strings (EngineClaude | EngineCodex). It is the shared,
// cross-module allow-list check: internal/delegation's validateInput and the
// daemon's session.IsKnownEngine both resolve through this one function so a
// typo'd or hostile to_engine is rejected identically on every path, closing
// the silent-drift risk of three independent hardcoded {claude,codex} switches.
func IsKnownEngine(engine string) bool {
	for _, e := range knownEngines {
		if e == engine {
			return true
		}
	}
	return false
}

// KnownEngines returns a freshly allocated copy of the canonical engine
// allow-list. Callers receive their own slice so the package-level source of
// truth cannot be mutated. The daemon's session package drift-guards its
// engineAgentPairs bijection against this set.
func KnownEngines() []string {
	out := make([]string, len(knownEngines))
	copy(out, knownEngines)
	return out
}

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

	// MsgDelegationCancelAck is sent by the daemon to the PWA acknowledging
	// receipt of a MsgDelegationCancel: the cancel was accepted and graceful
	// teardown has begun (the source delegation moves to DelegationStateCancelling).
	// Modeled on the GitSyncCancelResponse ack→terminal split (gitsync.go): this
	// ACK confirms receipt; the canonical terminal
	// delegation_result{status:"cancelled"} follows when teardown completes. It
	// carries ONLY session IDs + the request ID (+ a redacted reason on reject) —
	// never any delegate output or inherited content.
	MsgDelegationCancelAck = "delegation_cancel_ack"

	// MsgDelegationForceAbort is sent by the PWA to the daemon to ESCALATE a
	// stalled cancel to a hard stop — the named UX escape after the daemon
	// reports DelegationStateCancelStalled. The daemon hard-stops the delegate
	// (DelegationStateForceKilled) and still emits the canonical terminal
	// delegation_result{status:"cancelled"} so old PWAs self-heal. Carries ONLY
	// the delegate session ID + the request ID.
	MsgDelegationForceAbort = "delegation_force_abort"

	// MsgDelegationPreview is sent by the daemon to the PWA before any delegate is
	// spawned. It carries the human-reviewable handoff draft for an explicit
	// approve/deny decision. No delegation_link/status/result payload carries this
	// plaintext body.
	MsgDelegationPreview = "delegation_preview"

	// MsgDelegationPreviewDecision is sent by the PWA to approve or deny a pending
	// preview. Approval may include edited handoff fields; denial never spawns a
	// delegate.
	MsgDelegationPreviewDecision = "delegation_preview_decision"

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

	// ── Intermediate cancel-teardown states (cancel-teardown feature) ──
	//
	// These three states are carried in DelegationStatusPayload.State (and may
	// transiently appear in PersistedDelegationLink.State during teardown) to
	// drive the TRANSIENT cancel UI. They are NEVER the terminal frame: the
	// canonical terminal signal stays delegation_result{status:"cancelled"}, so
	// an old PWA that only resolves status==="cancelled" still self-heals on the
	// result frame and never sees a non-terminal dead-end. Changing any value is
	// a breaking change across all 3 repos.

	// DelegationStateCancelling reports that a cancel request was received and
	// graceful teardown is in progress (emitted alongside MsgDelegationCancelAck).
	DelegationStateCancelling = "cancelling"

	// DelegationStateCancelStalled reports that graceful teardown exceeded the
	// bounded grace window; the PWA surfaces the force-abort escape hatch.
	DelegationStateCancelStalled = "cancel_stalled"

	// DelegationStateForceKilled reports that a force-abort hard-stopped the
	// delegate. The terminal delegation_result{status:"cancelled"} still follows.
	DelegationStateForceKilled = "force_killed"
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

// Preview decisions carried in DelegationPreviewDecisionPayload.Decision.
const (
	DelegationPreviewDecisionApprove = "approve"
	DelegationPreviewDecisionDeny    = "deny"
)

// Durable handoff first-message delivery states carried in the scalar audit field
// PersistedDelegationLink.DeliveryState. Empty (absent) means legacy/not-tracked.
// These are stable wire strings — changing any of them is a breaking change for any
// consumer that reads the persisted delivery_state audit field.
const (
	DelegationDeliveryStatePending   = "pending"
	DelegationDeliveryStateDelivered = "delivered"
	DelegationDeliveryStateFailed    = "failed"
)

// Durable expected_output validation states carried in the scalar audit field
// PersistedDelegationLink.ValidationState (task #58 Phase C). Empty (absent) means
// validation was never engaged (the advisory default, max_validation_retries == 0).
// These are stable wire strings — changing any of them is a breaking change for any
// consumer that reads the persisted validation_state audit field.
//
//   - "awaiting_validation": a turn-end result did not match expected_output and the
//     daemon issued a bounded re-prompt; a crash in this window recovers as a
//     status-only result (the captured partial rides the encrypted blob, never here).
//   - "passed": the (re-prompted) result matched the requested shape.
//   - "failed": the retry budget was exhausted with the result still mismatching;
//     the last captured result is returned advisory-style.
const (
	DelegationValidationStateAwaiting = "awaiting_validation"
	DelegationValidationStatePassed   = "passed"
	DelegationValidationStateFailed   = "failed"
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
	// Parked discriminates an await=true delegation (the source agent is PARKED —
	// entry.delegating=true — and receives the delegate result as a synthetic
	// follow-up turn) from an await=false fire-and-forget delegation (the source is
	// NEVER parked and runs concurrently). The PWA derives isParkedSource /
	// deleteSuppressed / delegate-list-hiding from this discriminator.
	//
	// THREE-STATE *bool (Finding #1): a plain bool+omitempty cannot satisfy the two
	// back-compat goals simultaneously — Go's json.Marshal drops a false bool, so an
	// await=false link and an OLD daemon that never set the field would BOTH serialize
	// without a "parked" key (false and absent collapse to identical wire bytes), and
	// the PWA's `parked !== false` rule would mislabel a concurrent await=false
	// delegation as parked. The field is therefore a *bool (mirroring
	// PersistedDelegationLink.Await / StartDelegationPayload.Await):
	//   - &true  (await=true)  ⇒ `"parked":true`  — source is parked.
	//   - &false (await=false) ⇒ `"parked":false` — fire-and-forget, NOT parked.
	//   - nil    (legacy/old daemon) ⇒ key omitted — PWA's `parked !== false` reads
	//     absent⇒undefined⇒true, preserving the pre-field assume-parked default.
	// A daemon that DOES set the field always sends the truthful await value, so the
	// PWA can distinguish all three states.
	Parked *bool `json:"parked,omitempty"`
	// InheritedStateSummary is a disclosure-only, compact, human-readable summary of
	// the benign source state inherited by the delegate (e.g. "branch X @ <head>,
	// model/effort"). It is NOT machine-parsed and carries no inherited context
	// content — only a glanceable description for the human and for link history.
	// Empty when no inheritance occurred (the default); omitempty keeps the wire
	// byte-identical to a pre-feature producer. This follows the same disclosure
	// pattern as ApprovalAttribution.InheritedApprovalMode / InheritedSandboxMode and
	// the "all optional fields use omitempty" contract at the top of this file.
	InheritedStateSummary string `json:"inherited_state_summary,omitempty"`
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
	// RequestID is an at-least-once idempotency key. The daemon de-dupes repeat
	// cancels by (delegate_sid, request_id) so a re-sent cancel after a lost ack
	// converges rather than double-tearing-down, and echoes it back on the
	// MsgDelegationCancelAck so the PWA can reconcile the ack to its in-flight
	// request. omitempty keeps a pre-feature cancel frame byte-identical; an
	// absent request_id is treated by the daemon as a fresh (non-idempotent) cancel.
	RequestID string `json:"request_id,omitempty"`
}

// DelegationCancelAckPayload acknowledges a MsgDelegationCancel. Carried inside an
// AgentMessage envelope with Type=MsgDelegationCancelAck.
//
// This is the ack half of the ack→terminal split modeled on GitSyncCancelResponse
// (gitsync.go): Accepted=true means the daemon found the link and began graceful
// teardown (the source delegation moves to DelegationStateCancelling); the
// canonical terminal delegation_result{status:"cancelled"} follows shortly.
// Accepted=false with a redacted Error means the cancel could not be started.
// An already-terminal link is acknowledged Accepted=true (idempotent no-op) so a
// duplicate or late cancel converges rather than erroring.
//
// SECURITY (no-content-on-wire): carries ONLY session IDs + the request ID +
// accepted + a redacted reason — never delegate output, inherited context, diffs,
// or any other content.
type DelegationCancelAckPayload struct {
	DelegateSID string `json:"delegate_sid"`         // delegate agent session ID being cancelled
	RequestID   string `json:"request_id,omitempty"` // echoes DelegationCancelPayload.RequestID for idempotent client reconcile
	Accepted    bool   `json:"accepted"`             // true = teardown started OR already terminal (idempotent); false = could not start
	Error       string `json:"error,omitempty"`      // redacted, client-safe reason when Accepted=false
}

// DelegationForceAbortPayload escalates a stalled cancel to a hard stop. Carried
// inside an AgentMessage envelope with Type=MsgDelegationForceAbort.
//
// The PWA sends this only after a DelegationStateCancelStalled state (the named
// escape hatch); the daemon hard-stops the delegate (DelegationStateForceKilled)
// and emits the canonical terminal delegation_result{status:"cancelled"}.
// Idempotent: a force-abort on an already-terminal delegate is a benign no-op.
//
// SECURITY (no-content-on-wire): carries ONLY the delegate session ID + the
// request ID — no content.
type DelegationForceAbortPayload struct {
	DelegateSID string `json:"delegate_sid"`         // delegate agent session ID to hard-stop
	RequestID   string `json:"request_id,omitempty"` // idempotency key, paired with the originating cancel
}

// DelegationPreviewPayload is the human-reviewable handoff draft emitted before a
// delegate is spawned. It is intentionally richer than DelegationLinkPayload:
// this is the only live plaintext preview surface, while link/status/result and
// persisted-link payloads remain content-free.
type DelegationPreviewPayload struct {
	PreviewID             string                      `json:"preview_id"`
	SourceSID             string                      `json:"source_sid"`
	SourceEngine          string                      `json:"source_engine,omitempty"`
	TargetEngine          string                      `json:"target_engine"`
	Await                 bool                        `json:"await"`
	TriggeredBy           string                      `json:"triggered_by,omitempty"`
	Prompt                string                      `json:"prompt"`
	Context               string                      `json:"context,omitempty"`
	ExpectedOutput        string                      `json:"expected_output,omitempty"`
	ApprovedPlan          string                      `json:"approved_plan,omitempty"`
	GitContext            string                      `json:"git_context,omitempty"`
	InheritedStateSummary string                      `json:"inherited_state_summary,omitempty"`
	ByteStatus            DelegationPreviewByteStatus `json:"byte_status"`
	TimeoutAt             int64                       `json:"timeout_at,omitempty"`
	TimeoutRemainingMS    int64                       `json:"timeout_remaining_ms"`
	CreatedAt             int64                       `json:"created_at"`
}

// DelegationPreviewByteStatus gives the PWA a bounded-size readout for the draft
// handoff. It carries only sizes and status bits, never hidden content.
type DelegationPreviewByteStatus struct {
	AssembledBytes int  `json:"assembled_bytes"`
	MaxBytes       int  `json:"max_bytes"`
	OverLimit      bool `json:"over_limit"`
	Truncated      bool `json:"truncated"`
}

// DelegationPreviewDecisionPayload answers a pending preview. Approve may carry
// edited handoff fields; deny may carry a redacted human reason.
type DelegationPreviewDecisionPayload struct {
	PreviewID      string `json:"preview_id"`
	Decision       string `json:"decision"` // DelegationPreviewDecisionApprove | DelegationPreviewDecisionDeny
	Prompt         string `json:"prompt,omitempty"`
	Context        string `json:"context,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	Notes          string `json:"notes,omitempty"`
	DenyReason     string `json:"deny_reason,omitempty"`

	ContextSet        bool `json:"-"`
	ExpectedOutputSet bool `json:"-"`
}

// MarshalJSON preserves explicit clears for editable fields. Non-empty context
// and expected_output are emitted for ordinary producers; ContextSet and
// ExpectedOutputSet force emission even when the intended edited value is "".
func (p DelegationPreviewDecisionPayload) MarshalJSON() ([]byte, error) {
	out := map[string]any{
		"preview_id": p.PreviewID,
		"decision":   p.Decision,
	}
	if p.Prompt != "" {
		out["prompt"] = p.Prompt
	}
	if p.ContextSet || p.Context != "" {
		out["context"] = p.Context
	}
	if p.ExpectedOutputSet || p.ExpectedOutput != "" {
		out["expected_output"] = p.ExpectedOutput
	}
	if p.Notes != "" {
		out["notes"] = p.Notes
	}
	if p.DenyReason != "" {
		out["deny_reason"] = p.DenyReason
	}
	return json.Marshal(out)
}

// UnmarshalJSON records presence for editable fields where an omitted value
// means "leave the preview as-is" but an explicit empty string means "clear it".
func (p *DelegationPreviewDecisionPayload) UnmarshalJSON(data []byte) error {
	type alias DelegationPreviewDecisionPayload
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = DelegationPreviewDecisionPayload(decoded)
	_, p.ContextSet = raw["context"]
	_, p.ExpectedOutputSet = raw["expected_output"]
	return nil
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
	// InheritedApprovalMode / InheritedSandboxMode disclose, on a delegate-SPAWN
	// approval card, the elevated posture a delegate WILL inherit from its source
	// when the operator-enabled, per-call opt-in fires (agentd config
	// delegation.inherit_source_approval_mode + the delegate tool's inherit_approval
	// arg). They let the human approving the spawn see that the delegate will run
	// with permission prompts disabled before granting it. Both omitempty: a spawn
	// approval that does not inherit marshals to exactly the bytes it did before
	// this feature shipped. Populated by the daemon on the source-session spawn
	// approval only — distinct from SourceSID/SourceEngine above, which the daemon
	// stamps on approvals raised BY a delegate. Values are agentd approval-mode /
	// Codex sandbox-mode strings; this module stays dependency-free and does not
	// validate them.
	InheritedApprovalMode string `json:"inherited_approval_mode,omitempty"`
	InheritedSandboxMode  string `json:"inherited_sandbox_mode,omitempty"`
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
	// InheritedApprovalMode / InheritedSandboxMode record the elevated posture a
	// delegate inherited from its source at spawn (empty when no inheritance
	// occurred — the default), persisted so an operator can audit which delegates
	// ran with an inherited, permission-prompt-disabled posture after the fact.
	// Companion to the daemon's INFO spawn log. omitempty keeps pre-feature
	// journals byte-identical; an absent value recovers as "no inheritance".
	InheritedApprovalMode string `json:"inherited_approval_mode,omitempty"`
	InheritedSandboxMode  string `json:"inherited_sandbox_mode,omitempty"`
	// DeliveryState records the durable handoff first-message delivery state, one of
	// DelegationDeliveryState* — "pending" | "delivered" | "failed". Empty (absent)
	// means legacy/not-tracked. omitempty keeps pre-feature journals byte-identical.
	// This is a SCALAR audit field ONLY — it records WHETHER the inherited handoff
	// first message reached the delegate, NEVER the inherited context content itself.
	DeliveryState string `json:"delivery_state,omitempty"`
	// ValidationState records the durable expected_output enforcement state, one of
	// DelegationValidationState* — "awaiting_validation" | "passed" | "failed". Empty
	// (absent) means validation was never engaged (max_validation_retries == 0, the
	// advisory default). omitempty keeps pre-feature journals byte-identical. Like
	// DeliveryState this is a SCALAR audit field ONLY — it records the validation
	// PHASE, NEVER the result content or the requested shape (the captured partial
	// rides the separate ENCRYPTED handoff blob store, never this plaintext link).
	ValidationState string `json:"validation_state,omitempty"`
}
