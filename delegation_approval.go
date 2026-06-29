package protocol

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
// module - the same cross-repo additive-field pattern as StatusPayload /
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
	// approval only - distinct from SourceSID/SourceEngine above, which the daemon
	// stamps on approvals raised BY a delegate. Values are agentd approval-mode /
	// Codex sandbox-mode strings; this module stays dependency-free and does not
	// validate them.
	InheritedApprovalMode string `json:"inherited_approval_mode,omitempty"`
	InheritedSandboxMode  string `json:"inherited_sandbox_mode,omitempty"`
}
