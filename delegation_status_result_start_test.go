package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDelegationStatusPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state string
	}{
		{name: "pending", state: DelegationStatePending},
		{name: "active", state: DelegationStateActive},
		{name: "completed", state: DelegationStateCompleted},
		{name: "cancelled", state: DelegationStateCancelled},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationStatusPayload{
				DelegateSID: "sess-456-delegate",
				State:       tt.state,
				DiffSummary: "Modified 3 files",
				UpdatedAt:   1745692800100,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			rawStr := string(raw)
			for _, want := range []string{`"delegate_sid"`, `"state"`} {
				if !strings.Contains(rawStr, want) {
					t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
				}
			}

			var decoded DelegationStatusPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationStatusPayload_OmitemptyOptional verifies the optional fields
// (diff_summary, updated_at) drop off the wire when unset, matching the
// backward-compat guarantee of the other delegation payloads.

func TestDelegationStatusPayload_OmitemptyOptional(t *testing.T) {
	t.Parallel()

	minimal := DelegationStatusPayload{
		DelegateSID: "sess-456-delegate",
		State:       DelegationStateActive,
	}

	raw, err := json.Marshal(&minimal)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, absent := range []string{`"diff_summary"`, `"updated_at"`, `"source_sid"`, `"error"`} {
		if strings.Contains(rawStr, absent) {
			t.Errorf("expected %s to be omitted when empty; got %s", absent, rawStr)
		}
	}
}

// TestDelegationStatusPayload_RejectionRoundtrip verifies the additive rejection
// shape (Finding #6): State=DelegationStateRejected with SourceSID + Error and an
// EMPTY DelegateSID. The frame must round-trip and the rejection-only fields must
// appear on the wire when set.

func TestDelegationStatusPayload_RejectionRoundtrip(t *testing.T) {
	t.Parallel()

	original := DelegationStatusPayload{
		State:     DelegationStateRejected,
		SourceSID: "sess-123-source",
		Error:     "Delegation is disabled on this daemon.",
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{`"state":"rejected"`, `"source_sid"`, `"error"`} {
		if !strings.Contains(rawStr, want) {
			t.Errorf("rejection JSON missing %s; got %s", want, rawStr)
		}
	}
	// A rejection frame carries no delegate; delegate_sid serializes as empty
	// (no omitempty) but must never carry a value here.
	if strings.Contains(rawStr, `"delegate_sid":"sess`) {
		t.Errorf("rejection frame unexpectedly carries a delegate id; got %s", rawStr)
	}

	var decoded DelegationStatusPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
	}
}

// TestDelegationResultPayload_Roundtrip validates marshal/unmarshal across the
// canonical terminal statuses, including the optional pass/fail classification.

func TestDelegationResultPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   string
		passFail string
	}{
		{name: "completed_pass", status: DelegationResultCompleted, passFail: DelegationPassFailPass},
		{name: "error_fail", status: DelegationResultError, passFail: DelegationPassFailFail},
		{name: "cancelled", status: DelegationResultCancelled, passFail: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationResultPayload{
				DelegateSID: "sess-456-delegate",
				Status:      tt.status,
				Summary:     "Delegation completed successfully",
				DiffSummary: "Final changeset: 5 files modified",
				PassFail:    tt.passFail,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			rawStr := string(raw)
			for _, want := range []string{`"delegate_sid"`, `"status"`, `"summary"`} {
				if !strings.Contains(rawStr, want) {
					t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
				}
			}

			var decoded DelegationResultPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationResultPayload_OmitemptyOptional verifies the optional
// diff_summary and pass_fail fields drop off the wire when unset.

func TestDelegationResultPayload_OmitemptyOptional(t *testing.T) {
	t.Parallel()

	minimal := DelegationResultPayload{
		DelegateSID: "sess-456-delegate",
		Status:      DelegationResultCancelled,
		Summary:     "cancelled by source",
	}

	raw, err := json.Marshal(&minimal)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, absent := range []string{`"diff_summary"`, `"pass_fail"`} {
		if strings.Contains(rawStr, absent) {
			t.Errorf("expected %s to be omitted when empty; got %s", absent, rawStr)
		}
	}
}

// TestStartDelegationPayload_Roundtrip validates marshal/unmarshal for the
// user-initiated start-delegation command.
func TestStartDelegationPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		toEngine string
		await    bool
	}{
		{name: "to_codex_await", toEngine: EngineCodex, await: true},
		{name: "to_claude_fire_and_forget", toEngine: EngineClaude, await: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := StartDelegationPayload{
				SourceSID: "sess-123-source",
				ToEngine:  tt.toEngine,
				Prompt:    "implement the plan",
				Await:     tt.await,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			rawStr := string(raw)
			for _, want := range []string{`"source_sid"`, `"to_engine"`, `"prompt"`} {
				if !strings.Contains(rawStr, want) {
					t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
				}
			}

			var decoded StartDelegationPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestStartDelegationPayload_OmitemptyAwait verifies that await=false drops off
// the wire (omitempty), which is exactly why the daemon decoder must resolve the
// governed absent⇒true default from field PRESENCE rather than the Go zero value.

func TestStartDelegationPayload_OmitemptyAwait(t *testing.T) {
	t.Parallel()

	minimal := StartDelegationPayload{
		SourceSID: "sess-123-source",
		ToEngine:  EngineCodex,
		Prompt:    "do the thing",
		Await:     false,
	}

	raw, err := json.Marshal(&minimal)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if strings.Contains(string(raw), `"await"`) {
		t.Errorf("expected await to be omitted when false (omitempty); got %s", string(raw))
	}
}

// TestStartDelegationAwaitOrDefault locks the governed absent⇒true await rule for
// the user-initiated start_delegation frame, resolved from RAW JSON bytes (the
// plain-bool struct field cannot distinguish absent from explicit-false). This is
// the start_delegation analog of delegation.DelegateInput.AwaitOrDefault and is
// the single source of truth both daemon dispatch handlers call.

func TestStartDelegationAwaitOrDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "absent_defaults_true", raw: `{"source_sid":"s1","to_engine":"codex","prompt":"go"}`, want: true},
		{name: "explicit_true", raw: `{"source_sid":"s1","to_engine":"codex","prompt":"go","await":true}`, want: true},
		{name: "explicit_false_is_honored", raw: `{"source_sid":"s1","to_engine":"codex","prompt":"go","await":false}`, want: false},
		{name: "null_defaults_true", raw: `{"source_sid":"s1","await":null}`, want: true},
		{name: "malformed_defaults_true", raw: `{not json`, want: true},
		{name: "empty_object_defaults_true", raw: `{}`, want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StartDelegationAwaitOrDefault(json.RawMessage(tt.raw)); got != tt.want {
				t.Fatalf("StartDelegationAwaitOrDefault(%s) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
