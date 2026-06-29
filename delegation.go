// Package protocol defines the bidirectional Claude<->Codex governed
// delegation wire protocol (delegation feature, Phase 1).
//
// The delegation files define the delegation MessageType payloads, the engine
// and lifecycle string constants, additive approval-attribution fields, and the
// durable delegation-link schema replayed across daemon restarts.
//
// These wire types are MessageType frames (the outer AgentMessage.Type), NOT
// ControlType frames. The relay forwards them opaquely; the daemon and PWA
// discriminate by the outer AgentMessage.Type. agentd-protocol is the canonical
// source of truth for these values; daemon-local mirrors under
// agentd/internal/session/ must cross-check against this package at startup.
//
// Engine values are the stable wire strings "claude" | "codex". All optional
// fields use json:",omitempty" so older producers/consumers stay byte
// compatible (backward compatible per CLAUDE.md "Backward Compatible").
package protocol

// Engine values identify which underlying AI agent owns a delegation endpoint.
// agentd-protocol owns these stable wire strings; daemon, relay, and PWA
// consumers must update in lockstep for any breaking value change.
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
// internal/session - a depguard boundary forbids delegation->session), and
// internal/session's engineAgentPairs bijection is drift-guarded against
// KnownEngines() so the allow-list and the engine<->AgentType mapping never
// silently diverge across the 3 modules.
//
// It is a package-level slice literal (not a mutable var the callers handle
// directly): IsKnownEngine reads it for membership and KnownEngines returns a
// fresh copy on each call, so no caller can mutate the canonical set. WHEN
// ADDING A NEW ENGINE, add its Engine* const here AND its {engine, AgentType}
// pair to session.engineAgentPairs - the drift-guard test fails until both
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

// Message-type constants are owned by agentd-protocol. Daemon-local mirrors
// under agentd/internal/session/ must retain init() panic cross-checks against
// these values at startup.
const (
	// MsgDelegationLink is sent by the source to the delegate over the relay
	// when a delegation begins. Carries both session identities and engines.
	MsgDelegationLink = "delegation_link"

	// MsgDelegationStatus is sent by the delegate to the source over the relay
	// to report link state changes (pending -> active -> completed -> cancelled).
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
	// Modeled on the GitSyncCancelResponse ack->terminal split (gitsync.go): this
	// ACK confirms receipt; the canonical terminal
	// delegation_result{status:"cancelled"} follows when teardown completes. It
	// carries ONLY session IDs + the request ID (+ a redacted reason on reject) -
	// never any delegate output or inherited content.
	MsgDelegationCancelAck = "delegation_cancel_ack"

	// MsgDelegationForceAbort is sent by the PWA to the daemon to ESCALATE a
	// stalled cancel to a hard stop - the named UX escape after the daemon
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
	// daemon spawn machinery. PWA->daemon command frame, forwarded opaquely by
	// the relay, discriminated by the outer AgentMessage.Type - same convention
	// as the other four delegation frames.
	MsgStartDelegation = "start_delegation"
)

// Link lifecycle states carried in DelegationStatusPayload.State and the
// durable PersistedDelegationLink.State.
const (
	DelegationStatePending   = "pending"
	DelegationStateActive    = "active"
	DelegationStateCompleted = "completed"
	DelegationStateCancelled = "cancelled"

	// DelegationStateRejected is a DelegationStatusPayload-only state (it is
	// NEVER persisted into PersistedDelegationLink.State - no link is ever
	// created for a rejected start). It reports that a user-initiated
	// start_delegation was refused by the daemon BEFORE any delegate session was
	// spawned (delegation disabled / killswitch off, admission-depth-cycle-writer
	// reject, unknown engine, source-not-found, quiescence failure). The frame is
	// keyed by SourceSID (the would-be source session, NOT a delegate - no
	// DelegateSID exists yet) and carries a redacted Error string the PWA
	// reconciles onto its per-source delegationStartErrors cell. This closes the
	// silent-failure gap where a rejected hand-off produced no negative frame and
	// the StartDelegationSheet had already closed optimistically.
	DelegationStateRejected = "rejected"

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
// PersistedDelegationLink.TriggeredBy - who initiated the delegation.
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
// These are stable wire strings - changing any of them is a breaking change for
// any consumer that reads the persisted delivery_state audit field.
const (
	DelegationDeliveryStatePending   = "pending"
	DelegationDeliveryStateDelivered = "delivered"
	DelegationDeliveryStateFailed    = "failed"
)

// Durable expected_output validation states carried in the scalar audit field
// PersistedDelegationLink.ValidationState (task #58 Phase C). Empty (absent) means
// validation was never engaged (the advisory default, max_validation_retries == 0).
// These are stable wire strings - changing any of them is a breaking change for any
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
