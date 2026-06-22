package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestDelegationLinkPayload_Roundtrip validates marshal/unmarshal for the
// DelegationLinkPayload struct. Mirrors the approval_test.go roundtrip pattern
// per CLAUDE.md "Roundtrip Tested" requirement.
func TestDelegationLinkPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sourceEngine   string
		delegateEngine string
		triggeredBy    string
	}{
		{name: "claude_to_codex", sourceEngine: EngineClaude, delegateEngine: EngineCodex, triggeredBy: DelegationTriggeredByUser},
		{name: "codex_to_claude", sourceEngine: EngineCodex, delegateEngine: EngineClaude, triggeredBy: DelegationTriggeredByAuto},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationLinkPayload{
				SourceSID:      "sess-123-source",
				SourceEngine:   tt.sourceEngine,
				DelegateSID:    "sess-456-delegate",
				DelegateEngine: tt.delegateEngine,
				WorkDir:        "/tmp/workspace",
				TriggeredBy:    tt.triggeredBy,
				CreatedAt:      1745692800000,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			rawStr := string(raw)
			for _, want := range []string{`"source_sid"`, `"source_engine"`, `"delegate_sid"`, `"delegate_engine"`, `"work_dir"`, `"triggered_by"`, `"created_at"`} {
				if !strings.Contains(rawStr, want) {
					t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
				}
			}

			var decoded DelegationLinkPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationLinkPayload_OmitemptyOptional verifies that the optional fields
// (work_dir, triggered_by) are omitted from the wire when unset, so older
// producers/consumers remain byte-compatible (backward compatibility).
func TestDelegationLinkPayload_OmitemptyOptional(t *testing.T) {
	t.Parallel()

	minimal := DelegationLinkPayload{
		SourceSID:      "sess-123-source",
		SourceEngine:   EngineClaude,
		DelegateSID:    "sess-456-delegate",
		DelegateEngine: EngineCodex,
		CreatedAt:      1745692800000,
	}

	raw, err := json.Marshal(&minimal)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, absent := range []string{`"work_dir"`, `"triggered_by"`} {
		if strings.Contains(rawStr, absent) {
			t.Errorf("expected %s to be omitted when empty; got %s", absent, rawStr)
		}
	}
}

// TestDelegationStatusPayload_Roundtrip validates marshal/unmarshal across the
// canonical link lifecycle states.
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

// TestDelegationCancelPayload_Roundtrip validates marshal/unmarshal for the
// cancellation request payload.
func TestDelegationCancelPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reason string
	}{
		{name: "with_reason", reason: "User requested cancellation"},
		{name: "without_reason", reason: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationCancelPayload{
				DelegateSID: "sess-456-delegate",
				Reason:      tt.reason,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			if !strings.Contains(string(raw), `"delegate_sid"`) {
				t.Errorf("marshaled JSON missing %q; got %s", "delegate_sid", string(raw))
			}
			if tt.reason == "" && strings.Contains(string(raw), `"reason"`) {
				t.Errorf("expected reason to be omitted when empty; got %s", string(raw))
			}

			var decoded DelegationCancelPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestStartDelegationPayload_Roundtrip validates marshal/unmarshal for the
// user-initiated start-delegation command (the PWA "hand off" trigger).
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

// TestPersistedDelegationLink_Roundtrip validates marshal/unmarshal for the
// durable link schema replayed across daemon restarts (Phase 3 recovery, C3)
// and rendered in PWA link history (Phase 4, C7).
func TestPersistedDelegationLink_Roundtrip(t *testing.T) {
	t.Parallel()

	original := PersistedDelegationLink{
		SourceSID:      "sess-123-source",
		SourceEngine:   EngineClaude,
		DelegateSID:    "sess-456-delegate",
		DelegateEngine: EngineCodex,
		WorkDir:        "/tmp/workspace",
		TriggeredBy:    DelegationTriggeredByUser,
		State:          DelegationStateActive,
		CreatedAt:      1745692800000,
		UpdatedAt:      1745692860000,
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{`"source_sid"`, `"source_engine"`, `"delegate_sid"`, `"delegate_engine"`, `"created_at"`} {
		if !strings.Contains(rawStr, want) {
			t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
		}
	}

	var decoded PersistedDelegationLink
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded != original {
		t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
	}
}

// TestPersistedDelegationLink_OmitemptyOptional verifies the optional fields
// drop off the wire when unset so the durable record stays forward/backward
// compatible across daemon restarts.
func TestPersistedDelegationLink_OmitemptyOptional(t *testing.T) {
	t.Parallel()

	minimal := PersistedDelegationLink{
		SourceSID:      "sess-123-source",
		SourceEngine:   EngineClaude,
		DelegateSID:    "sess-456-delegate",
		DelegateEngine: EngineCodex,
		CreatedAt:      1745692800000,
	}

	raw, err := json.Marshal(&minimal)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, absent := range []string{`"work_dir"`, `"triggered_by"`, `"state"`, `"updated_at"`, `"await"`} {
		if strings.Contains(rawStr, absent) {
			t.Errorf("expected %s to be omitted when empty; got %s", absent, rawStr)
		}
	}
}

// TestPersistedDelegationLink_AwaitRoundtrip validates the Await *bool tri-state
// (Finding C): a NIL Await is omitted from the wire (so old journals stay
// byte-compatible and recovery resolves absent⇒true), while an explicit true/false
// round-trips through marshal/unmarshal preserving the fire-and-forget distinction.
func TestPersistedDelegationLink_AwaitRoundtrip(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name       string
		await      *bool
		wantInJSON bool // whether `"await"` appears on the wire
	}{
		{name: "nil await omitted", await: nil, wantInJSON: false},
		{name: "explicit false persists", await: boolPtr(false), wantInJSON: true},
		{name: "explicit true persists", await: boolPtr(true), wantInJSON: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := PersistedDelegationLink{
				SourceSID:      "sess-src",
				SourceEngine:   EngineClaude,
				DelegateSID:    "sess-del",
				DelegateEngine: EngineCodex,
				CreatedAt:      1745692800000,
				Await:          tt.await,
			}
			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			gotInJSON := strings.Contains(string(raw), `"await"`)
			if gotInJSON != tt.wantInJSON {
				t.Errorf("await-in-JSON = %v, want %v; raw=%s", gotInJSON, tt.wantInJSON, string(raw))
			}

			var decoded PersistedDelegationLink
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			switch {
			case tt.await == nil && decoded.Await != nil:
				t.Errorf("nil await round-tripped to non-nil %v", *decoded.Await)
			case tt.await != nil && decoded.Await == nil:
				t.Errorf("explicit await %v round-tripped to nil", *tt.await)
			case tt.await != nil && *decoded.Await != *tt.await:
				t.Errorf("await round-trip mismatch: want %v, got %v", *tt.await, *decoded.Await)
			}
		})
	}
}

// TestApprovalAttribution_Roundtrip validates the additive attribution fields
// (source_sid, source_engine) that the daemon embeds into its ApprovalPayload.
// Both fields use omitempty so old daemons that never set them produce
// byte-identical JSON (backward compatible — old PWAs ignore unknown fields).
func TestApprovalAttribution_Roundtrip(t *testing.T) {
	t.Parallel()

	original := ApprovalAttribution{
		SourceSID:    "sess-123-source",
		SourceEngine: EngineClaude,
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{`"source_sid"`, `"source_engine"`} {
		if !strings.Contains(rawStr, want) {
			t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
		}
	}

	var decoded ApprovalAttribution
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded != original {
		t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
	}
}

// TestApprovalAttribution_OmitemptyBackwardCompat verifies both attribution
// fields are omitted from the wire when unset. This is the backward-compat
// guarantee: an approval emitted by a daemon that does not set attribution
// marshals to exactly the same bytes it did before this feature shipped.
func TestApprovalAttribution_OmitemptyBackwardCompat(t *testing.T) {
	t.Parallel()

	var empty ApprovalAttribution

	raw, err := json.Marshal(&empty)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if string(raw) != "{}" {
		t.Errorf("expected empty attribution to marshal to {}; got %s", string(raw))
	}
}

// ─── Constant Pinning Tests ────────────────────────────────────────────────────

// TestDelegationMessageTypeConstants pins the four wire-type strings. Any change
// here is a breaking change across all 3 repos (daemon, relay, PWA).
func TestDelegationMessageTypeConstants(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"delegation_link":   MsgDelegationLink,
		"delegation_status": MsgDelegationStatus,
		"delegation_result": MsgDelegationResult,
		"delegation_cancel": MsgDelegationCancel,
	}

	for expected, actual := range want {
		if expected != actual {
			t.Errorf("delegation MessageType constant for %q = %q, want %q", expected, actual, expected)
		}
	}
}

// TestEngineConstants pins the canonical engine wire strings. Engine values are
// "claude" | "codex"; changing either is a breaking change across all repos.
func TestEngineConstants(t *testing.T) {
	t.Parallel()

	if EngineClaude != "claude" {
		t.Errorf("EngineClaude = %q, want %q", EngineClaude, "claude")
	}
	if EngineCodex != "codex" {
		t.Errorf("EngineCodex = %q, want %q", EngineCodex, "codex")
	}
}
