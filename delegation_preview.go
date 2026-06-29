package protocol

import "encoding/json"

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
