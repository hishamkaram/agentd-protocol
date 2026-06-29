package protocol

// PersistedDelegationLink is the minimal durable representation of a
// source<->delegate link. The daemon persists this and replays it on restart so
// it can reconstruct the parked-source / live-delegate relationship (Phase 3
// recovery, C3); a reloaded PWA renders it in link history (Phase 4, C7).
//
// Timestamps are unix milliseconds (int64) - matching the existing module
// convention (ApprovalResolvedPayload.ResolvedAt, DelegationLinkPayload.CreatedAt)
// and avoiding the time.Time monotonic-clock roundtrip trap. All optional fields
// use omitempty for forward/backward compatibility across daemon restarts.
//
// Phase 3 (daemon repo, NOT here): the daemon's internal persisted struct in
// agentd/internal/session/persistence.go adds daemon-specific fields that are
// NOT part of this shared schema - e.g. CancelledAt, CancelReason, FinalStatus,
// and retry bookkeeping. Only the wire/shared portion lives in agentd-protocol.
type PersistedDelegationLink struct {
	SourceSID      string `json:"source_sid"`             // source agent session ID
	SourceEngine   string `json:"source_engine"`          // EngineClaude | EngineCodex
	DelegateSID    string `json:"delegate_sid"`           // delegate agent session ID
	DelegateEngine string `json:"delegate_engine"`        // EngineClaude | EngineCodex
	WorkDir        string `json:"work_dir,omitempty"`     // working directory scope for the delegation
	TriggeredBy    string `json:"triggered_by,omitempty"` // DelegationTriggeredBy* - user | auto | system
	State          string `json:"state,omitempty"`        // DelegationState* - pending | active | completed | cancelled
	CreatedAt      int64  `json:"created_at"`             // unix milliseconds
	UpdatedAt      int64  `json:"updated_at,omitempty"`   // unix milliseconds
	// Await is the governed delegation lifecycle flag persisted so restart recovery
	// can distinguish an await=true (parked source, returns result) delegation from
	// an await=false (fire-and-forget, source never parked) one. It is a *bool with
	// the same absent=>true idiom as StartDelegationPayload.Await /
	// StartDelegationAwaitOrDefault: a NIL (absent) value resolves to true on
	// recovery, so a journal written before this field existed recovers as the safe
	// parked default. omitempty drops a nil pointer entirely, so old daemons and old
	// journals stay byte-compatible. Without this field, recovery hardcoded
	// await=true and re-parked a fire-and-forget source - its turns would queue, the
	// quiescence clamp would fire, and an unwanted synthetic result would be injected.
	Await *bool `json:"await,omitempty"`
	// InheritedApprovalMode / InheritedSandboxMode record the elevated posture a
	// delegate inherited from its source at spawn (empty when no inheritance
	// occurred - the default), persisted so an operator can audit which delegates
	// ran with an inherited, permission-prompt-disabled posture after the fact.
	// Companion to the daemon's INFO spawn log. omitempty keeps pre-feature
	// journals byte-identical; an absent value recovers as "no inheritance".
	InheritedApprovalMode string `json:"inherited_approval_mode,omitempty"`
	InheritedSandboxMode  string `json:"inherited_sandbox_mode,omitempty"`
	// DeliveryState records the durable handoff first-message delivery state, one of
	// DelegationDeliveryState* - "pending" | "delivered" | "failed". Empty (absent)
	// means legacy/not-tracked. omitempty keeps pre-feature journals byte-identical.
	// This is a SCALAR audit field ONLY - it records WHETHER the inherited handoff
	// first message reached the delegate, NEVER the inherited context content itself.
	DeliveryState string `json:"delivery_state,omitempty"`
	// ValidationState records the durable expected_output enforcement state, one of
	// DelegationValidationState* - "awaiting_validation" | "passed" | "failed". Empty
	// (absent) means validation was never engaged (max_validation_retries == 0, the
	// advisory default). omitempty keeps pre-feature journals byte-identical. Like
	// DeliveryState this is a SCALAR audit field ONLY - it records the validation
	// PHASE, NEVER the result content or the requested shape (the captured partial
	// rides the separate ENCRYPTED handoff blob store, never this plaintext link).
	ValidationState string `json:"validation_state,omitempty"`
}
