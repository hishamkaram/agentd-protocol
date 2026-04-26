package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestHistoryReplayCompleteConstant verifies the exact wire-type literal.
// The byte-for-byte string equality is the entire reason this constant lives
// in agentd-protocol and is mirrored in agentd/internal/session/ with an
// init() panic cross-check (per the feature 169→170 wire-type-drift incident).
func TestHistoryReplayCompleteConstant(t *testing.T) {
	t.Parallel()

	const want = "history_replay_complete"
	if MsgHistoryReplayComplete != want {
		t.Errorf("MsgHistoryReplayComplete = %q, want %q", MsgHistoryReplayComplete, want)
	}
	if MsgHistoryReplayComplete == "" {
		t.Errorf("MsgHistoryReplayComplete is empty")
	}
}

// TestHistoryReplayCompletePayloadRoundtrip verifies JSON marshal/unmarshal
// round-tripping for the daemon→PWA `history_replay_complete` payload.
// Per the contract (specs/192-history-replay-sync/contracts/history-replay-complete.md):
//   - SessionID maps to JSON tag "session_id"
//   - SessionID is required and non-empty in production but the payload type
//     itself must round-trip even an empty value (defensive — the consumer's
//     handler is responsible for validation, not the wire type).
func TestHistoryReplayCompletePayloadRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   HistoryReplayCompletePayload
	}{
		{
			name: "typical session id",
			in:   HistoryReplayCompletePayload{SessionID: "test-sid-123"},
		},
		{
			name: "uuid shape",
			in:   HistoryReplayCompletePayload{SessionID: "550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name: "empty session id (edge case — preserved through roundtrip)",
			in:   HistoryReplayCompletePayload{SessionID: ""},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var out HistoryReplayCompletePayload
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if tt.in != out {
				t.Errorf("roundtrip mismatch:\n in:  %+v\n out: %+v", tt.in, out)
			}
		})
	}
}

// TestHistoryReplayCompletePayloadJSONShape verifies the on-wire JSON
// matches the exact shape from the contract (the `session_id` field name).
// This is the producer-side guard against accidental json-tag drift like
// SessionId / sessionID / sessionId variants.
func TestHistoryReplayCompletePayloadJSONShape(t *testing.T) {
	t.Parallel()

	in := HistoryReplayCompletePayload{SessionID: "test-sid-123"}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := string(data)
	if !strings.Contains(got, `"session_id":"test-sid-123"`) {
		t.Errorf("expected JSON to contain `\"session_id\":\"test-sid-123\"`, got: %s", got)
	}
	for _, banned := range []string{`"sessionID"`, `"sessionId"`, `"SessionID"`, `"session-id"`} {
		if strings.Contains(got, banned) {
			t.Errorf("JSON must not contain %s, got: %s", banned, got)
		}
	}
}
