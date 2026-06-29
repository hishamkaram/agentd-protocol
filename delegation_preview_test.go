package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDelegationPreviewPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	original := DelegationPreviewPayload{
		PreviewID:             "preview-123",
		SourceSID:             "source-1",
		SourceEngine:          EngineClaude,
		TargetEngine:          EngineCodex,
		Await:                 true,
		TriggeredBy:           DelegationTriggeredByUser,
		Prompt:                "implement the approved plan",
		Context:               "source context",
		ExpectedOutput:        "patch and summary",
		ApprovedPlan:          "1. do the thing",
		GitContext:            "branch feature/delegation",
		InheritedStateSummary: "branch feature/delegation",
		ByteStatus: DelegationPreviewByteStatus{
			AssembledBytes: 2048,
			MaxBytes:       4096,
			OverLimit:      false,
			Truncated:      false,
		},
		TimeoutAt:          1745693100000,
		TimeoutRemainingMS: 300000,
		CreatedAt:          1745692800000,
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{
		`"preview_id"`,
		`"source_sid"`,
		`"source_engine"`,
		`"target_engine"`,
		`"await"`,
		`"triggered_by"`,
		`"prompt"`,
		`"context"`,
		`"expected_output"`,
		`"approved_plan"`,
		`"git_context"`,
		`"inherited_state_summary"`,
		`"byte_status"`,
		`"timeout_at"`,
		`"timeout_remaining_ms"`,
		`"created_at"`,
	} {
		if !strings.Contains(rawStr, want) {
			t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
		}
	}

	var decoded DelegationPreviewPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
	}
}

func TestDelegationPreviewDecisionPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		decision string
	}{
		{name: "approve", decision: DelegationPreviewDecisionApprove},
		{name: "deny", decision: DelegationPreviewDecisionDeny},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationPreviewDecisionPayload{
				PreviewID:         "preview-123",
				Decision:          tt.decision,
				Prompt:            "edited prompt",
				Context:           "edited context",
				ExpectedOutput:    "edited output",
				Notes:             "human note",
				DenyReason:        "not now",
				ContextSet:        true,
				ExpectedOutputSet: true,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			var decoded DelegationPreviewDecisionPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

func TestDelegationPreviewDecisionPayload_FieldPresence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		raw                   string
		wantContext           string
		wantContextSet        bool
		wantExpectedOutput    string
		wantExpectedOutputSet bool
	}{
		{
			name:                  "omitted fields are not marked set",
			raw:                   `{"preview_id":"preview-1","decision":"approve","prompt":"edited"}`,
			wantContextSet:        false,
			wantExpectedOutputSet: false,
		},
		{
			name:                  "explicit empty fields are marked set",
			raw:                   `{"preview_id":"preview-1","decision":"approve","context":"","expected_output":""}`,
			wantContext:           "",
			wantContextSet:        true,
			wantExpectedOutput:    "",
			wantExpectedOutputSet: true,
		},
		{
			name:                  "explicit non-empty fields are marked set",
			raw:                   `{"preview_id":"preview-1","decision":"approve","context":"ctx","expected_output":"out"}`,
			wantContext:           "ctx",
			wantContextSet:        true,
			wantExpectedOutput:    "out",
			wantExpectedOutputSet: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var decoded DelegationPreviewDecisionPayload
			if err := json.Unmarshal([]byte(tt.raw), &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded.Context != tt.wantContext || decoded.ContextSet != tt.wantContextSet {
				t.Fatalf("context = %q/set %v, want %q/set %v", decoded.Context, decoded.ContextSet, tt.wantContext, tt.wantContextSet)
			}
			if decoded.ExpectedOutput != tt.wantExpectedOutput || decoded.ExpectedOutputSet != tt.wantExpectedOutputSet {
				t.Fatalf("expected_output = %q/set %v, want %q/set %v", decoded.ExpectedOutput, decoded.ExpectedOutputSet, tt.wantExpectedOutput, tt.wantExpectedOutputSet)
			}
		})
	}
}

func TestDelegationPreviewDecisionPayload_MarshalExplicitEmptyFields(t *testing.T) {
	t.Parallel()

	payload := DelegationPreviewDecisionPayload{
		PreviewID:         "preview-1",
		Decision:          DelegationPreviewDecisionApprove,
		Prompt:            "edited",
		Context:           "",
		ExpectedOutput:    "",
		ContextSet:        true,
		ExpectedOutputSet: true,
	}
	raw, err := json.Marshal(&payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	rawStr := string(raw)
	for _, want := range []string{`"context":""`, `"expected_output":""`} {
		if !strings.Contains(rawStr, want) {
			t.Fatalf("marshaled decision missing explicit clear %s; got %s", want, rawStr)
		}
	}

	var decoded DelegationPreviewDecisionPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !decoded.ContextSet || !decoded.ExpectedOutputSet {
		t.Fatalf("explicit clear presence lost after roundtrip: %+v", decoded)
	}
}

func TestDelegationPreviewWireConstants(t *testing.T) {
	t.Parallel()

	if MsgDelegationPreview != "delegation_preview" {
		t.Fatalf("MsgDelegationPreview = %q, want delegation_preview", MsgDelegationPreview)
	}
	if MsgDelegationPreviewDecision != "delegation_preview_decision" {
		t.Fatalf("MsgDelegationPreviewDecision = %q, want delegation_preview_decision", MsgDelegationPreviewDecision)
	}
	if DelegationPreviewDecisionApprove != "approve" {
		t.Fatalf("DelegationPreviewDecisionApprove = %q, want approve", DelegationPreviewDecisionApprove)
	}
	if DelegationPreviewDecisionDeny != "deny" {
		t.Fatalf("DelegationPreviewDecisionDeny = %q, want deny", DelegationPreviewDecisionDeny)
	}
}
