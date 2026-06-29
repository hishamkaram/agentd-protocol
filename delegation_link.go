package protocol

// DelegationLinkPayload initiates a delegation from the source agent to the
// delegate agent. Carried inside an AgentMessage envelope with
// Type=MsgDelegationLink.
type DelegationLinkPayload struct {
	SourceSID      string `json:"source_sid"`             // source agent session ID
	SourceEngine   string `json:"source_engine"`          // EngineClaude | EngineCodex
	DelegateSID    string `json:"delegate_sid"`           // delegate agent session ID
	DelegateEngine string `json:"delegate_engine"`        // EngineClaude | EngineCodex
	WorkDir        string `json:"work_dir,omitempty"`     // working directory scope for the delegation
	TriggeredBy    string `json:"triggered_by,omitempty"` // DelegationTriggeredBy* - user | auto | system
	CreatedAt      int64  `json:"created_at"`             // unix milliseconds
	// Parked discriminates an await=true delegation (the source agent is PARKED -
	// entry.delegating=true - and receives the delegate result as a synthetic
	// follow-up turn) from an await=false fire-and-forget delegation (the source is
	// NEVER parked and runs concurrently). The PWA derives isParkedSource /
	// deleteSuppressed / delegate-list-hiding from this discriminator.
	//
	// THREE-STATE *bool (Finding #1): a plain bool+omitempty cannot satisfy the two
	// back-compat goals simultaneously - Go's json.Marshal drops a false bool, so an
	// await=false link and an OLD daemon that never set the field would BOTH serialize
	// without a "parked" key (false and absent collapse to identical wire bytes), and
	// the PWA's `parked !== false` rule would mislabel a concurrent await=false
	// delegation as parked. The field is therefore a *bool (mirroring
	// PersistedDelegationLink.Await / StartDelegationPayload.Await):
	//   - &true  (await=true)  => `"parked":true`  - source is parked.
	//   - &false (await=false) => `"parked":false` - fire-and-forget, NOT parked.
	//   - nil    (legacy/old daemon) => key omitted - PWA's `parked !== false` reads
	//     absent=>undefined=>true, preserving the pre-field assume-parked default.
	// A daemon that DOES set the field always sends the truthful await value, so the
	// PWA can distinguish all three states.
	Parked *bool `json:"parked,omitempty"`
	// InheritedStateSummary is a disclosure-only, compact, human-readable summary of
	// the benign source state inherited by the delegate (e.g. "branch X @ <head>,
	// model/effort"). It is NOT machine-parsed and carries no inherited context
	// content - only a glanceable description for the human and for link history.
	// Empty when no inheritance occurred (the default); omitempty keeps the wire
	// byte-identical to a pre-feature producer. This follows the same disclosure
	// pattern as ApprovalAttribution.InheritedApprovalMode / InheritedSandboxMode and
	// the "all optional fields use omitempty" contract at the top of delegation.go.
	InheritedStateSummary string `json:"inherited_state_summary,omitempty"`
}
