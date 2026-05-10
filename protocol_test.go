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
		RelayAuth:   "opaque-hosted-relay-token",
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

	tests := []struct {
		name    string
		payload protocol.StatusUpdatePayload
	}{
		{
			name: "all 7 fields populated",
			payload: protocol.StatusUpdatePayload{
				SessionID:   "sess-123",
				State:       "running",
				CostUSD:     1.23,
				Project:     "my-project",
				AgentType:   "claude-code",
				DeveloperID: "dev-1",
				CreatedAt:   1712000000000,
			},
		},
		{
			name: "only original fields (backward compat)",
			payload: protocol.StatusUpdatePayload{
				SessionID: "sess-456",
				State:     "idle",
				CostUSD:   0.50,
			},
		},
		{
			name: "zero-value new fields omitted from JSON",
			payload: protocol.StatusUpdatePayload{
				SessionID: "sess-789",
				State:     "stopped",
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

func TestStatusUpdatePayloadBackwardCompat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
		want protocol.StatusUpdatePayload
	}{
		{
			name: "JSON without new fields unmarshals to zero values",
			json: `{"sid":"sess-old","state":"running","cost_usd":0.5}`,
			want: protocol.StatusUpdatePayload{
				SessionID: "sess-old",
				State:     "running",
				CostUSD:   0.5,
			},
		},
		{
			name: "JSON with all fields round-trips correctly",
			json: `{"sid":"sess-full","state":"idle","cost_usd":1.0,"project":"proj","agent":"aider","dev_id":"dev-2","created_at":1712000000000}`,
			want: protocol.StatusUpdatePayload{
				SessionID:   "sess-full",
				State:       "idle",
				CostUSD:     1.0,
				Project:     "proj",
				AgentType:   "aider",
				DeveloperID: "dev-2",
				CreatedAt:   1712000000000,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got protocol.StatusUpdatePayload
			if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mismatch:\n  got:  %+v\n  want: %+v", got, tt.want)
			}
		})
	}
}

func TestStatusUpdatePayloadOmitempty(t *testing.T) {
	t.Parallel()
	// When new fields are zero, they must not appear in JSON output.
	payload := protocol.StatusUpdatePayload{
		SessionID: "sess-omit",
		State:     "running",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{`"agent"`, `"dev_id"`, `"created_at"`} {
		if bytes.Contains(data, []byte(field)) {
			t.Errorf("zero-value field %s should be omitted, got JSON: %s", field, s)
		}
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

func TestEntitlementUpdatePayloadRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload protocol.EntitlementUpdatePayload
	}{
		{
			name: "active solo",
			payload: protocol.EntitlementUpdatePayload{
				Plan:                  "solo",
				BillingStatus:         "active",
				ActiveSessionLimit:    1,
				CurrentActiveSessions: 0,
				BufferTTLSeconds:      86400,
				UpdatedAt:             1778419200000,
			},
		},
		{
			name: "blocked unpaid",
			payload: protocol.EntitlementUpdatePayload{
				Plan:                  "none",
				BillingStatus:         "unpaid",
				ActiveSessionLimit:    0,
				CurrentActiveSessions: 0,
				BufferTTLSeconds:      0,
				BlockedReason:         "billing_inactive",
				UpdatedAt:             1778419200000,
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

func TestEntitlementViolationPayloadRoundtrip(t *testing.T) {
	t.Parallel()

	original := protocol.EntitlementViolationPayload{
		AgentSessionID:        "sess_abc",
		Reason:                "active_session_limit_reached",
		Plan:                  "solo",
		ActiveSessionLimit:    1,
		CurrentActiveSessions: 2,
		Message:               "Hosted session limit reached",
		OccurredAt:            1778419200000,
	}
	assertRoundtrip(t, original)
}

func TestEntitlementControlPayloadsBackwardCompat(t *testing.T) {
	t.Parallel()

	updateJSON := `{"type":"entitlement_update","payload":{"plan":"pro","billing_status":"active","active_session_limit":3,"current_active_sessions":2,"buffer_ttl_seconds":604800,"updated_at":1778419200000}}`
	var msg protocol.ControlMessage
	if err := json.Unmarshal([]byte(updateJSON), &msg); err != nil {
		t.Fatalf("unmarshal update control: %v", err)
	}
	if msg.Type != protocol.CtrlEntitlementUpdate {
		t.Fatalf("Type = %q, want %q", msg.Type, protocol.CtrlEntitlementUpdate)
	}
	var update protocol.EntitlementUpdatePayload
	if err := json.Unmarshal(msg.Payload, &update); err != nil {
		t.Fatalf("unmarshal update payload: %v", err)
	}
	if update.Plan != "pro" || update.ActiveSessionLimit != 3 || update.BufferTTLSeconds != 604800 {
		t.Fatalf("unexpected update payload: %+v", update)
	}

	violationJSON := `{"type":"entitlement_violation","payload":{"agent_session_id":"sess_abc","reason":"active_session_limit_reached","plan":"solo","active_session_limit":1,"current_active_sessions":2,"occurred_at":1778419200000}}`
	if err := json.Unmarshal([]byte(violationJSON), &msg); err != nil {
		t.Fatalf("unmarshal violation control: %v", err)
	}
	if msg.Type != protocol.CtrlEntitlementViolation {
		t.Fatalf("Type = %q, want %q", msg.Type, protocol.CtrlEntitlementViolation)
	}
	var violation protocol.EntitlementViolationPayload
	if err := json.Unmarshal(msg.Payload, &violation); err != nil {
		t.Fatalf("unmarshal violation payload: %v", err)
	}
	if violation.AgentSessionID != "sess_abc" || violation.Reason != "active_session_limit_reached" {
		t.Fatalf("unexpected violation payload: %+v", violation)
	}
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

func TestKeyRotatePayloadRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload protocol.KeyRotatePayload
	}{
		{
			name: "all fields populated",
			payload: protocol.KeyRotatePayload{
				NewKeyHMAC:    "dGVzdC1uZXcta2V5LWhtYWM",
				PrevKeyHMAC:   "dGVzdC1wcmV2LWtleS1obWFj",
				PrevExpiresAt: 1744300060,
				Epoch:         3,
			},
		},
		{
			name: "epoch zero",
			payload: protocol.KeyRotatePayload{
				NewKeyHMAC:    "YWJj",
				PrevKeyHMAC:   "ZGVm",
				PrevExpiresAt: 1744300000,
				Epoch:         0,
			},
		},
		{
			name: "empty HMACs",
			payload: protocol.KeyRotatePayload{
				NewKeyHMAC:    "",
				PrevKeyHMAC:   "",
				PrevExpiresAt: 0,
				Epoch:         0,
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

func TestKeyRotatePayloadOmitemptyNotUsed(t *testing.T) {
	t.Parallel()
	// KeyRotatePayload fields are NOT omitempty — all fields must be present in JSON
	// because the relay needs all four fields to process the rotation.
	payload := protocol.KeyRotatePayload{
		NewKeyHMAC:    "abc",
		PrevKeyHMAC:   "def",
		PrevExpiresAt: 1744300000,
		Epoch:         1,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{`"new_key_hmac"`, `"prev_key_hmac"`, `"prev_expires_at"`, `"epoch"`} {
		if !bytes.Contains(data, []byte(field)) {
			t.Errorf("field %s must be present in JSON, got: %s", field, s)
		}
	}
}

func TestControlTypeConstants(t *testing.T) {
	t.Parallel()
	expected := map[protocol.ControlType]string{
		protocol.CtrlRegister:             "register",
		protocol.CtrlJoin:                 "join",
		protocol.CtrlHeartbeat:            "heartbeat",
		protocol.CtrlAck:                  "ack",
		protocol.CtrlError:                "error",
		protocol.CtrlSyncPolicies:         "sync_policies",
		protocol.CtrlStatusUpdate:         "status_update",
		protocol.CtrlAuditEntry:           "audit_entry",
		protocol.CtrlDeactivateDeveloper:  "deactivate_developer",
		protocol.CtrlClientConnected:      "client_connected",
		protocol.CtrlClientCount:          "client_count",
		protocol.CtrlKeyRotate:            "key_rotate",
		protocol.CtrlEntitlementUpdate:    "entitlement_update",
		protocol.CtrlEntitlementViolation: "entitlement_violation",
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
