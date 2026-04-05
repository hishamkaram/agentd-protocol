package protocol_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestRelayEnvelopeRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envelope protocol.RelayEnvelope
	}{
		{
			name: "with trace ID",
			envelope: protocol.RelayEnvelope{
				SessionID: "sess-123",
				Seq:       42,
				Encrypted: []byte("ciphertext-bytes"),
				TraceID:   "4bf92f3577b34da6a3ce929d0e0e4736",
			},
		},
		{
			name: "without trace ID (backward compat)",
			envelope: protocol.RelayEnvelope{
				SessionID: "sess-456",
				Seq:       1,
				Encrypted: []byte("data"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.envelope)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			// Verify tid is omitted when empty
			if tt.envelope.TraceID == "" {
				if bytes.Contains(data, []byte(`"tid"`)) {
					t.Errorf("empty TraceID should be omitted from JSON, got: %s", data)
				}
			}
			var decoded protocol.RelayEnvelope
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tt.envelope, decoded) {
				t.Errorf("roundtrip mismatch:\n  got:  %+v\n  want: %+v", decoded, tt.envelope)
			}
		})
	}
}

func TestControlMessageRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.ControlMessage{
		Type:    protocol.CtrlRegister,
		Payload: json.RawMessage(`{"sid":"abc"}`),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded protocol.ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("roundtrip mismatch:\n  got:  %+v\n  want: %+v", decoded, original)
	}
}

func TestRegisterPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.RegisterPayload{
		SessionID:   "sess-123",
		KeyHMAC:     "hmac-abc",
		AgentType:   "claude-code",
		ProjectName: "my-project",
		DisplayName: "My Session",
		DeveloperID: "dev-1",
	}
	assertRoundtrip(t, original)
}

func TestJoinPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.JoinPayload{
		SessionID: "sess-123",
		JWT:       "eyJ...",
	}
	assertRoundtrip(t, original)
}

func TestAckPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	assertRoundtrip(t, protocol.AckPayload{SessionID: "sess-123"})
}

func TestErrorPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	assertRoundtrip(t, protocol.ErrorPayload{Code: "invalid_session", Message: "not found"})
}

func TestStatusUpdatePayloadRoundtrip(t *testing.T) {
	t.Parallel()
	assertRoundtrip(t, protocol.StatusUpdatePayload{
		SessionID:   "sess-123",
		State:       "running",
		CostUSD:     1.23,
		AgentType:   "claude-code",
		DeveloperID: "dev-1",
		CreatedAt:   1712345678000,
	})
}

func TestStatusUpdatePayloadBackwardCompat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload protocol.StatusUpdatePayload
	}{
		{
			name: "omitempty omits new fields when zero",
			payload: protocol.StatusUpdatePayload{
				SessionID: "sess-compat",
				State:     "idle",
			},
		},
		{
			name: "partial new fields",
			payload: protocol.StatusUpdatePayload{
				SessionID: "sess-partial",
				State:     "running",
				AgentType: "aider",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, tt.payload)
		})
	}
}

func TestAuditEntryPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.AuditEntryPayload{
		Timestamp:   time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
		SessionID:   "sess-123",
		DeveloperID: "dev-1",
		AgentType:   "claude-code",
		EventType:   "tool_use",
		Tool:        "Bash",
		InputHash:   "abc123",
		Decision:    "allow",
		PolicyName:  "default",
		CostUSD:     0.05,
	}
	assertRoundtrip(t, original)
}

func TestDeactivateDeveloperPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	assertRoundtrip(t, protocol.DeactivateDeveloperPayload{DeveloperID: "dev-1"})
}

func TestClientConnectedPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	assertRoundtrip(t, protocol.ClientConnectedPayload{SessionID: "sess-123"})
}

func TestSyncPoliciesPayloadRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.SyncPoliciesPayload{
		Policies: []protocol.PolicyJSON{
			{
				Name: "deny-rm",
				Match: protocol.PolicyMatchJSON{
					Tool:    []string{"Bash"},
					Command: "rm -rf",
				},
				Action:   "deny",
				Reason:   "dangerous",
				TimeoutS: 30,
			},
		},
	}
	assertRoundtrip(t, original)
}

func TestPolicyJSONRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.PolicyJSON{
		Name: "allow-read",
		Match: protocol.PolicyMatchJSON{
			Tool:        []string{"Read"},
			FilePattern: "*.go",
			RiskLevel:   []string{"low"},
		},
		Action:    "allow",
		Approvers: []string{"admin@example.com"},
	}
	assertRoundtrip(t, original)
}

func TestControlTypeConstants(t *testing.T) {
	t.Parallel()
	expected := map[protocol.ControlType]string{
		protocol.CtrlRegister:            "register",
		protocol.CtrlJoin:                "join",
		protocol.CtrlHeartbeat:           "heartbeat",
		protocol.CtrlAck:                 "ack",
		protocol.CtrlError:               "error",
		protocol.CtrlSyncPolicies:        "sync_policies",
		protocol.CtrlStatusUpdate:        "status_update",
		protocol.CtrlAuditEntry:          "audit_entry",
		protocol.CtrlDeactivateDeveloper: "deactivate_developer",
		protocol.CtrlClientConnected:     "client_connected",
	}
	for ct, want := range expected {
		if string(ct) != want {
			t.Errorf("ControlType %q != %q", ct, want)
		}
	}
}

func TestRelayEnvelopeEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envelope protocol.RelayEnvelope
	}{
		{
			name:     "zero-value struct",
			envelope: protocol.RelayEnvelope{},
		},
		{
			name: "nil Encrypted field",
			envelope: protocol.RelayEnvelope{
				SessionID: "sess-nil-enc",
				Seq:       7,
				Encrypted: nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, tt.envelope)
		})
	}
}

func TestControlMessageEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("zero-value struct marshal unmarshal", func(t *testing.T) {
		t.Parallel()
		original := protocol.ControlMessage{}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded protocol.ControlMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Type != "" {
			t.Errorf("expected empty Type, got %q", decoded.Type)
		}
		// nil json.RawMessage marshals as "null", which unmarshals as json.RawMessage("null")
		// This is expected JSON behavior — verify it explicitly.
		if decoded.Payload == nil {
			t.Error("expected Payload to be non-nil after roundtrip of zero-value (JSON null)")
		}
	})

	roundtripTests := []struct {
		name string
		msg  protocol.ControlMessage
	}{
		{
			name: "unknown control type",
			msg: protocol.ControlMessage{
				Type:    protocol.ControlType("unknown_type_xyz"),
				Payload: json.RawMessage(`{"foo":"bar"}`),
			},
		},
		{
			name: "null payload roundtrips",
			msg: protocol.ControlMessage{
				Type:    protocol.CtrlHeartbeat,
				Payload: json.RawMessage("null"),
			},
		},
		{
			name: "empty JSON object payload",
			msg: protocol.ControlMessage{
				Type:    protocol.CtrlAck,
				Payload: json.RawMessage(`{}`),
			},
		},
	}
	for _, tt := range roundtripTests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, tt.msg)
		})
	}
}

func TestPolicyJSONEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("zero-value struct", func(t *testing.T) {
		t.Parallel()
		assertRoundtrip(t, protocol.PolicyJSON{})
	})

	t.Run("nil slices in PolicyMatchJSON", func(t *testing.T) {
		t.Parallel()
		assertRoundtrip(t, protocol.PolicyJSON{
			Name:   "nil-slices",
			Match:  protocol.PolicyMatchJSON{},
			Action: "deny",
		})
	})

	t.Run("empty slices become nil after roundtrip due to omitempty", func(t *testing.T) {
		t.Parallel()
		original := protocol.PolicyJSON{
			Name: "empty-match",
			Match: protocol.PolicyMatchJSON{
				Tool:      []string{},
				RiskLevel: []string{},
			},
			Action: "allow",
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded protocol.PolicyJSON
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		// omitempty causes empty slices to be omitted from JSON, resulting in nil after unmarshal.
		if decoded.Name != "empty-match" {
			t.Errorf("Name: got %q, want %q", decoded.Name, "empty-match")
		}
		if decoded.Action != "allow" {
			t.Errorf("Action: got %q, want %q", decoded.Action, "allow")
		}
		if decoded.Match.Tool != nil {
			t.Errorf("Tool: expected nil after roundtrip of empty slice, got %v", decoded.Match.Tool)
		}
		if decoded.Match.RiskLevel != nil {
			t.Errorf("RiskLevel: expected nil after roundtrip of empty slice, got %v", decoded.Match.RiskLevel)
		}
	})
}

func TestSyncPoliciesPayloadEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload protocol.SyncPoliciesPayload
	}{
		{
			name:    "zero-value struct",
			payload: protocol.SyncPoliciesPayload{},
		},
		{
			name:    "empty policies slice",
			payload: protocol.SyncPoliciesPayload{Policies: []protocol.PolicyJSON{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, tt.payload)
		})
	}
}

func TestInvalidJSONDecode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "invalid JSON for ControlMessage",
			input:   `{invalid}`,
			target:  &protocol.ControlMessage{},
			wantErr: true,
		},
		{
			name:    "invalid JSON for RelayEnvelope",
			input:   `{invalid}`,
			target:  &protocol.RelayEnvelope{},
			wantErr: true,
		},
		{
			name:    "truncated JSON for ControlMessage",
			input:   `{"type":"register","pay`,
			target:  &protocol.ControlMessage{},
			wantErr: true,
		},
		{
			name:    "truncated JSON for RelayEnvelope",
			input:   `{"sid":"abc","seq":1,"en`,
			target:  &protocol.RelayEnvelope{},
			wantErr: true,
		},
		{
			name:    "wrong type for seq field",
			input:   `{"sid":"abc","seq":"not-a-number","enc":"AA=="}`,
			target:  &protocol.RelayEnvelope{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := json.Unmarshal([]byte(tt.input), tt.target)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}

// assertRoundtrip marshals v to JSON and unmarshals back, asserting equality.
func assertRoundtrip[T any](t *testing.T, original T) {
	t.Helper()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("roundtrip mismatch:\n  got:  %+v\n  want: %+v", decoded, original)
	}
}
