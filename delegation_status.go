package protocol

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
//     to exactly the bytes it did before this field pair was added - old daemons
//     and old PWAs stay byte-compatible.
type DelegationStatusPayload struct {
	DelegateSID string `json:"delegate_sid"`           // delegate agent session ID (EMPTY on a rejection frame)
	State       string `json:"state"`                  // DelegationState* - pending | active | completed | cancelled | rejected
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
	Status      string `json:"status"`                 // DelegationResult* - completed | error | cancelled
	Summary     string `json:"summary"`                // human-readable result summary
	DiffSummary string `json:"diff_summary,omitempty"` // final changeset summary
	PassFail    string `json:"pass_fail,omitempty"`    // DelegationPassFail* - pass | fail completion classification
}
