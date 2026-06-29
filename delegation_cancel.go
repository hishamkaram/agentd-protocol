package protocol

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

// DelegationCancelAckPayload acknowledges a MsgDelegationCancel. Carried inside
// an AgentMessage envelope with Type=MsgDelegationCancelAck.
//
// This is the ack half of the ack->terminal split modeled on GitSyncCancelResponse
// (gitsync.go): Accepted=true means the daemon found the link and began graceful
// teardown (the source delegation moves to DelegationStateCancelling); the
// canonical terminal delegation_result{status:"cancelled"} follows shortly.
// Accepted=false with a redacted Error means the cancel could not be started.
// An already-terminal link is acknowledged Accepted=true (idempotent no-op) so a
// duplicate or late cancel converges rather than erroring.
//
// SECURITY (no-content-on-wire): carries ONLY session IDs + the request ID +
// accepted + a redacted reason - never delegate output, inherited context, diffs,
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
// request ID - no content.
type DelegationForceAbortPayload struct {
	DelegateSID string `json:"delegate_sid"`         // delegate session ID to hard-stop
	RequestID   string `json:"request_id,omitempty"` // idempotency key, paired with the originating cancel
}
