package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

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

// TestDelegationMessageTypeConstants pins the delegation message wire strings.
func TestDelegationMessageTypeConstants(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"delegation_link":        MsgDelegationLink,
		"delegation_status":      MsgDelegationStatus,
		"delegation_result":      MsgDelegationResult,
		"delegation_cancel":      MsgDelegationCancel,
		"delegation_cancel_ack":  MsgDelegationCancelAck,
		"delegation_force_abort": MsgDelegationForceAbort,
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
