package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

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

// TestDelegationCancelPayload_RequestIDRoundtrip validates the additive
// idempotency key on the cancel frame.
func TestDelegationCancelPayload_RequestIDRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requestID  string
		wantOnWire bool
	}{
		{name: "with_request_id", requestID: "cancel-req-7f3a", wantOnWire: true},
		{name: "legacy_empty_omitted", requestID: "", wantOnWire: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationCancelPayload{
				DelegateSID: "sess-456-delegate",
				Reason:      "user requested",
				RequestID:   tt.requestID,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			gotOnWire := strings.Contains(string(raw), `"request_id"`)
			if gotOnWire != tt.wantOnWire {
				t.Errorf("request_id-on-wire = %v, want %v; raw=%s", gotOnWire, tt.wantOnWire, string(raw))
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

// TestDelegationCancelAckPayload_Roundtrip validates the daemon→PWA cancel
// acknowledgement (the ack half of the ack→terminal split modeled on
// GitSyncCancelResponse). accepted is always on the wire; request_id and error
// are omitempty.

func TestDelegationCancelAckPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requestID string
		accepted  bool
		errStr    string
	}{
		{name: "accepted_with_request_id", requestID: "cancel-req-1", accepted: true, errStr: ""},
		{name: "rejected_with_reason", requestID: "cancel-req-2", accepted: false, errStr: "unknown delegate"},
		{name: "accepted_legacy_no_request_id", requestID: "", accepted: true, errStr: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationCancelAckPayload{
				DelegateSID: "sess-456-delegate",
				RequestID:   tt.requestID,
				Accepted:    tt.accepted,
				Error:       tt.errStr,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			if !strings.Contains(string(raw), `"delegate_sid"`) {
				t.Errorf("marshaled JSON missing %q; got %s", "delegate_sid", string(raw))
			}
			// accepted has no omitempty — it must always be present so the PWA can
			// distinguish a started-teardown ack from an absent field.
			if !strings.Contains(string(raw), `"accepted"`) {
				t.Errorf("expected accepted to always be present; got %s", string(raw))
			}
			if tt.requestID == "" && strings.Contains(string(raw), `"request_id"`) {
				t.Errorf("expected request_id omitted when empty; got %s", string(raw))
			}
			if tt.errStr == "" && strings.Contains(string(raw), `"error"`) {
				t.Errorf("expected error omitted when empty; got %s", string(raw))
			}

			var decoded DelegationCancelAckPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationForceAbortPayload_Roundtrip validates the PWA→daemon force-abort
// escalation frame (the named UX escape after cancel_stalled). It carries only
// the delegate session ID + an idempotency request_id.

func TestDelegationForceAbortPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requestID string
	}{
		{name: "with_request_id", requestID: "force-req-9"},
		{name: "legacy_no_request_id", requestID: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := DelegationForceAbortPayload{
				DelegateSID: "sess-456-delegate",
				RequestID:   tt.requestID,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}
			if !strings.Contains(string(raw), `"delegate_sid"`) {
				t.Errorf("marshaled JSON missing %q; got %s", "delegate_sid", string(raw))
			}
			if tt.requestID == "" && strings.Contains(string(raw), `"request_id"`) {
				t.Errorf("expected request_id omitted when empty; got %s", string(raw))
			}

			var decoded DelegationForceAbortPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestDelegationCancelLifecycleStateConstants pins the new intermediate cancel
// lifecycle state wire strings. These are carried in DelegationStatusPayload.State
// to drive transient UI; the canonical TERMINAL frame stays
// delegation_result{status:"cancelled"}. Changing any value is a breaking change.

func TestDelegationCancelLifecycleStateConstants(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"cancelling":     DelegationStateCancelling,
		"cancel_stalled": DelegationStateCancelStalled,
		"force_killed":   DelegationStateForceKilled,
	}
	for expected, actual := range want {
		if expected != actual {
			t.Errorf("delegation cancel lifecycle state constant = %q, want %q", actual, expected)
		}
	}
}

// TestDelegationCancelControlFramesNoContent is the SECURITY pin (no-content-on-wire):
// the cancel-ack and force-abort control frames MUST carry ONLY session IDs + the
// request ID (+ accepted/redacted-error on the ack), NEVER delegate output, inherited
// context, diffs, summaries, or other content. The allowed key set is locked so a
// future field that would leak content fails this gate.

func TestDelegationCancelControlFramesNoContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		marshal func() ([]byte, error)
		allowed map[string]bool
	}{
		{
			name: "cancel_ack",
			marshal: func() ([]byte, error) {
				return json.Marshal(&DelegationCancelAckPayload{
					DelegateSID: "sess-delegate",
					RequestID:   "req-1",
					Accepted:    true,
					Error:       "redacted reason",
				})
			},
			allowed: map[string]bool{"delegate_sid": true, "request_id": true, "accepted": true, "error": true},
		},
		{
			name: "force_abort",
			marshal: func() ([]byte, error) {
				return json.Marshal(&DelegationForceAbortPayload{
					DelegateSID: "sess-delegate",
					RequestID:   "req-1",
				})
			},
			allowed: map[string]bool{"delegate_sid": true, "request_id": true},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := tt.marshal()
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var keys map[string]json.RawMessage
			if err := json.Unmarshal(raw, &keys); err != nil {
				t.Fatalf("unmarshal to key map: %v", err)
			}
			for k := range keys {
				if !tt.allowed[k] {
					t.Errorf("frame carries disallowed key %q (possible content leak); full=%s", k, string(raw))
				}
			}
		})
	}
}
