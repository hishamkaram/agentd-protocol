// Package protocol — approval lifecycle wire protocol (feature 193).
//
// This file defines the daemon→PWA wire constants and payloads added in
// feature 193 to fix approval gates re-appearing after approve/reject.
// The fix introduces two new wire types:
//
//   - MsgApprovalResolved: tombstone broadcast on every approval terminal
//     state (allow / allow_session / deny / timeout / canceled / superseded).
//     Drives live multi-client convergence (Slack chat.update / GitHub
//     waiting-deployments / Buildkite block-step pattern). PWA dismisses the
//     ApprovalGate modal upon receipt; ANY client that decided AND any other
//     client showing the same approval converges via this single signal.
//
//   - MsgPendingApprovalState: control-state signal emitted during reconnect
//     metadata bootstrap. Triggers setPendingApproval on the PWA WITHOUT
//     entering chat history. Reuses
//     the existing ApprovalPayload struct shape (defined in
//     agentd/internal/session/types.go); only the outer wire-type constant
//     differs from MsgApproval. The daemon re-stamps `ts` per emission so
//     the PWA's seenMessageKeys dedup key changes — bypasses the existing
//     dedupe naturally without weakening it.
//
// agentd-protocol is the source of truth for the Msg* constants in this file.
// Matching daemon-local constants live under agentd/internal/session/ with
// init() panic cross-checks at daemon startup.
package protocol

// Message-type constants owned by agentd-protocol. Mirrors MUST exist in
// agentd/internal/session/ with init() panic cross-checks that the values agree
// at daemon startup.
const (
	// MsgApprovalResolved is the daemon→PWA tombstone broadcast on every
	// approval terminal state. Carries the approval_id, session_id, the
	// terminal decision string, and the resolution timestamp.
	MsgApprovalResolved = "approval_resolved"

	// MsgPendingApprovalState is the daemon→PWA control-state signal
	// emitted during bootstrap for any approval still pending in the
	// daemon's pendingApprovals map. Reuses the ApprovalPayload struct
	// shape; only the wire-type constant differs from MsgApproval. PWA
	// handler MUST early-return BEFORE addMessage so the state frame does not
	// duplicate chat-history bubbles (see ChatView.tsx filter
	// behavior — approvals render as bubbles by default).
	MsgPendingApprovalState = "pending_approval_state"
)

// ApprovalDecision values used in ApprovalResolvedPayload.Decision.
// agentd-protocol owns these stable wire strings; daemon, relay, and PWA
// consumers must update in lockstep for any breaking value change.
const (
	// ApprovalDecisionAllow — user approved (or auto-allowed by policy).
	ApprovalDecisionAllow = "allow"

	// ApprovalDecisionAllowSession — user approved with session-scoped
	// trust for the exact tool+command tuple.
	ApprovalDecisionAllowSession = "allow_session"

	// ApprovalDecisionDeny — user denied (or auto-denied by policy).
	ApprovalDecisionDeny = "deny"

	// ApprovalDecisionTimeout — TTL elapsed without a user decision.
	// Daemon emits this only for approvals with a positive configured timeout.
	ApprovalDecisionTimeout = "timeout"

	// ApprovalDecisionCanceled — SDK outer context was canceled (e.g.
	// session stopped, agent crashed) while the approval was pending.
	// Without this signal the modal would orphan on the PWA until the
	// next terminal status. See feature 193 spec User Story 5.
	ApprovalDecisionCanceled = "canceled"

	// ApprovalDecisionSuperseded is retained as a stable wire value for
	// backward compatibility with daemons that emitted synthetic denies before
	// same-session approvals were queued independently. Current daemons should
	// leave each approval_id pending until its own user decision, timeout, or
	// cancellation.
	ApprovalDecisionSuperseded = "superseded"
)

// ApprovalResolvedPayload is the daemon→PWA tombstone payload broadcast
// on every approval terminal state. Carried inside an AgentMessage envelope
// with Type=MsgApprovalResolved.
//
// The PWA matches the payload's ApprovalID against its queued
// pendingApprovals[sessionId] entries and clears only that approval_id.
// Idempotent — receiving the same tombstone twice is a no-op on the consumer.
//
// SessionID is required so the PWA can scope the clear to the correct
// session entry without an additional lookup.
//
// Decision is one of the ApprovalDecision* constants above. PWA logs
// unknown values at warn level and does NOT clear the modal (defense in
// depth — better to show stale modal than dismiss on bad signal).
//
// ResolvedAt is a unix-millisecond timestamp captured at the daemon when
// the resolution occurred. PWA does not act on this field beyond
// storage/audit.
type ApprovalResolvedPayload struct {
	ApprovalID string `json:"approval_id"`
	SessionID  string `json:"session_id"`
	Decision   string `json:"decision"`
	ResolvedAt int64  `json:"resolved_at"`
}
