package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

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

// TestPersistedDelegationLink_DeliveryStateRoundtrip validates the additive
// scalar delivery_state audit field.
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
