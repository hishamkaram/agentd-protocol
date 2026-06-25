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
				SourceSID:             "sess-123-source",
				SourceEngine:          tt.sourceEngine,
				DelegateSID:           "sess-456-delegate",
				DelegateEngine:        tt.delegateEngine,
				WorkDir:               "/tmp/workspace",
				TriggeredBy:           tt.triggeredBy,
				CreatedAt:             1745692800000,
				InheritedStateSummary: "branch feat/x @ abc1234, claude/high",
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			rawStr := string(raw)
			for _, want := range []string{`"source_sid"`, `"source_engine"`, `"delegate_sid"`, `"delegate_engine"`, `"work_dir"`, `"triggered_by"`, `"created_at"`, `"inherited_state_summary"`} {
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
	for _, absent := range []string{`"work_dir"`, `"triggered_by"`, `"parked"`, `"inherited_state_summary"`} {
		if strings.Contains(rawStr, absent) {
			t.Errorf("expected %s to be omitted when empty; got %s", absent, rawStr)
		}
	}
}

// TestDelegationLinkPayload_ParkedDiscriminator locks the Finding #1 fix: the
// three-state *bool Parked discriminator must be byte-distinguishable on the wire
// for all three states. An await=true link sends `"parked":true`; an await=false
// fire-and-forget link sends an EXPLICIT `"parked":false` (NOT omitted — that is
// the bug this test guards against: a plain bool+omitempty would drop false,
// making it indistinguishable from an old-daemon absent value and causing the PWA
// to mislabel a concurrent delegation as parked); a legacy nil link omits the key.
func TestDelegationLinkPayload_ParkedDiscriminator(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name       string
		parked     *bool
		wantOnWire bool   // whether `"parked"` appears on the wire at all
		wantValue  string // the exact `"parked":<v>` substring when on the wire
	}{
		{name: "await_true_parked", parked: boolPtr(true), wantOnWire: true, wantValue: `"parked":true`},
		{name: "await_false_fire_and_forget", parked: boolPtr(false), wantOnWire: true, wantValue: `"parked":false`},
		{name: "legacy_nil_omitted", parked: nil, wantOnWire: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationLinkPayload{
				SourceSID:      "sess-123-source",
				SourceEngine:   EngineClaude,
				DelegateSID:    "sess-456-delegate",
				DelegateEngine: EngineCodex,
				CreatedAt:      1745692800000,
				Parked:         tt.parked,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			rawStr := string(raw)

			if got := strings.Contains(rawStr, `"parked"`); got != tt.wantOnWire {
				t.Errorf("parked-on-wire = %v, want %v; json=%s", got, tt.wantOnWire, rawStr)
			}
			if tt.wantOnWire && !strings.Contains(rawStr, tt.wantValue) {
				t.Errorf("expected %s on the wire; got %s", tt.wantValue, rawStr)
			}

			var decoded DelegationLinkPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			switch {
			case tt.parked == nil && decoded.Parked != nil:
				t.Errorf("nil parked round-tripped to non-nil %v", *decoded.Parked)
			case tt.parked != nil && decoded.Parked == nil:
				t.Errorf("explicit parked %v round-tripped to nil", *tt.parked)
			case tt.parked != nil && *decoded.Parked != *tt.parked:
				t.Errorf("parked round-trip mismatch: want %v, got %v", *tt.parked, *decoded.Parked)
			}
		})
	}
}

// TestDelegationLinkPayload_ParkedWireDistinguishability is the end-to-end proof
// that await=false survives the wire as a value the consumer can tell apart from
// both await=true and a legacy-absent link. It marshals via the REAL Go struct
// (not hand-injected JSON) and then applies the PWA's `parked !== false` read rule
// to assert the resolved park state per case — exactly the chain that was broken
// when Parked was a plain bool (await=false dropped the key ⇒ PWA read it as parked).
func TestDelegationLinkPayload_ParkedWireDistinguishability(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name           string
		parked         *bool
		wantResolvedPK bool // PWA's `parked !== false` over the decoded value
	}{
		{name: "await_true_resolves_parked", parked: boolPtr(true), wantResolvedPK: true},
		{name: "await_false_resolves_unparked", parked: boolPtr(false), wantResolvedPK: false},
		{name: "legacy_nil_resolves_parked", parked: nil, wantResolvedPK: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Producer side: marshal the actual struct the daemon emits.
			raw, err := json.Marshal(&DelegationLinkPayload{
				SourceSID:      "s",
				SourceEngine:   EngineClaude,
				DelegateSID:    "d",
				DelegateEngine: EngineCodex,
				CreatedAt:      1,
				Parked:         tt.parked,
			})
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			// Consumer side: decode into a *bool-shaped view (the TS `parked?:
			// boolean` mirror) and apply the PWA's `parked !== false` rule.
			var decoded struct {
				Parked *bool `json:"parked"`
			}
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			// `parked !== false`: true and undefined ⇒ parked; explicit false ⇒ not.
			resolvedPK := decoded.Parked == nil || *decoded.Parked != false
			if resolvedPK != tt.wantResolvedPK {
				t.Errorf("resolved park state = %v, want %v; wire=%s", resolvedPK, tt.wantResolvedPK, string(raw))
			}
		})
	}
}

// TestIsKnownEngine locks the shared cross-module engine allow-list (Finding #3).
// internal/delegation and the daemon's session.IsKnownEngine both resolve
// through this one function so the allow-list cannot drift across the 3 modules.
func TestIsKnownEngine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		engine string
		want   bool
	}{
		{name: "claude", engine: EngineClaude, want: true},
		{name: "codex", engine: EngineCodex, want: true},
		{name: "empty", engine: "", want: false},
		{name: "unknown", engine: "gemini", want: false},
		{name: "typo", engine: "Claude", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsKnownEngine(tt.engine); got != tt.want {
				t.Errorf("IsKnownEngine(%q) = %v, want %v", tt.engine, got, tt.want)
			}
		})
	}
}

// TestKnownEngines verifies KnownEngines returns the canonical allow-list and
// that each call returns a FRESH copy a caller cannot use to mutate the shared
// source of truth.
func TestKnownEngines(t *testing.T) {
	t.Parallel()

	got := KnownEngines()
	want := []string{EngineClaude, EngineCodex}
	if len(got) != len(want) {
		t.Fatalf("KnownEngines() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("KnownEngines()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	// Every returned engine must satisfy IsKnownEngine (internal consistency).
	for _, e := range got {
		if !IsKnownEngine(e) {
			t.Errorf("KnownEngines() returned %q but IsKnownEngine(%q) = false", e, e)
		}
	}
	// Mutating the returned slice must not affect a subsequent call.
	got[0] = "mutated"
	second := KnownEngines()
	if second[0] != EngineClaude {
		t.Errorf("KnownEngines() not isolated: mutation leaked, got[0] = %q", second[0])
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
		SourceSID:             "sess-123-source",
		SourceEngine:          EngineClaude,
		DelegateSID:           "sess-456-delegate",
		DelegateEngine:        EngineCodex,
		WorkDir:               "/tmp/workspace",
		TriggeredBy:           DelegationTriggeredByUser,
		State:                 DelegationStateActive,
		CreatedAt:             1745692800000,
		UpdatedAt:             1745692860000,
		InheritedApprovalMode: "bypass",
		InheritedSandboxMode:  "danger-full-access",
		DeliveryState:         DelegationDeliveryStateDelivered,
		ValidationState:       DelegationValidationStatePassed,
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{`"source_sid"`, `"source_engine"`, `"delegate_sid"`, `"delegate_engine"`, `"created_at"`, `"inherited_approval_mode"`, `"inherited_sandbox_mode"`, `"delivery_state"`, `"validation_state"`} {
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
	for _, absent := range []string{`"work_dir"`, `"triggered_by"`, `"state"`, `"updated_at"`, `"await"`, `"inherited_approval_mode"`, `"inherited_sandbox_mode"`, `"delivery_state"`, `"validation_state"`} {
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
		SourceSID:             "sess-123-source",
		SourceEngine:          EngineClaude,
		InheritedApprovalMode: "bypass",
		InheritedSandboxMode:  "danger-full-access",
	}

	raw, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	rawStr := string(raw)
	for _, want := range []string{`"source_sid"`, `"source_engine"`, `"inherited_approval_mode"`, `"inherited_sandbox_mode"`} {
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

// ─── Delegation Context Transfer (task #58) Field Tests ──────────────────────────

// TestDelegationLinkPayload_InheritedStateSummaryRoundtrip validates the additive
// disclosure-only inherited_state_summary field: a non-empty summary appears on the
// wire and survives round-trip, while an empty summary is omitted (omitempty) so a
// pre-feature producer marshals to byte-identical JSON (backward compatible).
func TestDelegationLinkPayload_InheritedStateSummaryRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		summary    string
		wantOnWire bool
	}{
		{name: "with_summary", summary: "branch feat/x @ abc1234, claude/high", wantOnWire: true},
		{name: "empty_summary_omitted", summary: "", wantOnWire: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationLinkPayload{
				SourceSID:             "sess-123-source",
				SourceEngine:          EngineClaude,
				DelegateSID:           "sess-456-delegate",
				DelegateEngine:        EngineCodex,
				CreatedAt:             1745692800000,
				InheritedStateSummary: tt.summary,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			gotOnWire := strings.Contains(string(raw), `"inherited_state_summary"`)
			if gotOnWire != tt.wantOnWire {
				t.Errorf("inherited_state_summary-on-wire = %v, want %v; raw=%s", gotOnWire, tt.wantOnWire, string(raw))
			}

			var decoded DelegationLinkPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded.InheritedStateSummary != tt.summary {
				t.Errorf("inherited_state_summary round-trip mismatch: want %q, got %q", tt.summary, decoded.InheritedStateSummary)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestPersistedDelegationLink_DeliveryStateRoundtrip validates the additive scalar
// delivery_state audit field across its three canonical states plus the legacy
// empty (not-tracked) case. An empty value is omitted (omitempty) so journals
// written before this field existed stay byte-identical and recover as not-tracked.
func TestPersistedDelegationLink_DeliveryStateRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		state      string
		wantOnWire bool
	}{
		{name: "pending", state: DelegationDeliveryStatePending, wantOnWire: true},
		{name: "delivered", state: DelegationDeliveryStateDelivered, wantOnWire: true},
		{name: "failed", state: DelegationDeliveryStateFailed, wantOnWire: true},
		{name: "legacy_empty_omitted", state: "", wantOnWire: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := PersistedDelegationLink{
				SourceSID:      "sess-123-source",
				SourceEngine:   EngineClaude,
				DelegateSID:    "sess-456-delegate",
				DelegateEngine: EngineCodex,
				CreatedAt:      1745692800000,
				DeliveryState:  tt.state,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			gotOnWire := strings.Contains(string(raw), `"delivery_state"`)
			if gotOnWire != tt.wantOnWire {
				t.Errorf("delivery_state-on-wire = %v, want %v; raw=%s", gotOnWire, tt.wantOnWire, string(raw))
			}

			var decoded PersistedDelegationLink
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded.DeliveryState != tt.state {
				t.Errorf("delivery_state round-trip mismatch: want %q, got %q", tt.state, decoded.DeliveryState)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationDeliveryStateConstants pins the three durable handoff delivery-state
// wire strings. These are stable wire values audited across daemon restarts;
// changing any of them is a breaking change for any consumer that reads the
// persisted delivery_state audit field.
func TestDelegationDeliveryStateConstants(t *testing.T) {
	t.Parallel()

	// Pin each constant against a hardcoded string literal directly so a value
	// rename (e.g. "pending" -> "queued") fails the test. Do NOT compare a
	// constant against itself via a map — that weakens the pin.
	cases := []struct {
		name   string
		got    string
		expect string
	}{
		{"pending", DelegationDeliveryStatePending, "pending"},
		{"delivered", DelegationDeliveryStateDelivered, "delivered"},
		{"failed", DelegationDeliveryStateFailed, "failed"},
	}
	for _, tc := range cases {
		if tc.got != tc.expect {
			t.Errorf("delivery-state constant %s = %q, want %q", tc.name, tc.got, tc.expect)
		}
	}
}

// TestDelegationValidationStateConstants pins the three durable expected_output
// validation-state wire strings. delegation.go documents these as stable wire
// values persisted in the validation_state audit field; changing any of them is a
// breaking change for any consumer that reads it across daemon restarts.
func TestDelegationValidationStateConstants(t *testing.T) {
	t.Parallel()

	// Pin each constant against a hardcoded string literal directly so a value
	// rename (e.g. "awaiting_validation" -> "pending_validation") fails the test.
	// Do NOT compare a constant against itself via a map — that weakens the pin.
	cases := []struct {
		name   string
		got    string
		expect string
	}{
		{"awaiting", DelegationValidationStateAwaiting, "awaiting_validation"},
		{"passed", DelegationValidationStatePassed, "passed"},
		{"failed", DelegationValidationStateFailed, "failed"},
	}
	for _, tc := range cases {
		if tc.got != tc.expect {
			t.Errorf("validation-state constant %s = %q, want %q", tc.name, tc.got, tc.expect)
		}
	}
}

// TestPersistedDelegationLink_ValidationStateRoundtrip validates the additive scalar
// validation_state audit field across its three canonical states plus the legacy
// empty (validation-never-engaged) case. An empty value is omitted (omitempty) so
// journals written before this field existed stay byte-identical and recover as
// not-tracked. The "awaiting_validation" crash-recovery state (delegation.go) is the
// state most worth a dedicated round-trip pin.
func TestPersistedDelegationLink_ValidationStateRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		state      string
		wantOnWire bool
	}{
		{name: "awaiting_validation", state: DelegationValidationStateAwaiting, wantOnWire: true},
		{name: "passed", state: DelegationValidationStatePassed, wantOnWire: true},
		{name: "failed", state: DelegationValidationStateFailed, wantOnWire: true},
		{name: "legacy_empty_omitted", state: "", wantOnWire: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := PersistedDelegationLink{
				SourceSID:       "sess-123-source",
				SourceEngine:    EngineClaude,
				DelegateSID:     "sess-456-delegate",
				DelegateEngine:  EngineCodex,
				CreatedAt:       1745692800000,
				ValidationState: tt.state,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			gotOnWire := strings.Contains(string(raw), `"validation_state"`)
			if gotOnWire != tt.wantOnWire {
				t.Errorf("validation_state-on-wire = %v, want %v; raw=%s", gotOnWire, tt.wantOnWire, string(raw))
			}

			var decoded PersistedDelegationLink
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded.ValidationState != tt.state {
				t.Errorf("validation_state round-trip mismatch: want %q, got %q", tt.state, decoded.ValidationState)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}
