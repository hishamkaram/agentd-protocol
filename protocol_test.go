package protocol_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestRelayEnvelopeRoundtrip(t *testing.T) {
	t.Parallel()
	original := protocol.RelayEnvelope{
		SessionID: "sess-123",
		Seq:       42,
		Encrypted: []byte("ciphertext-bytes"),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded protocol.RelayEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("roundtrip mismatch:\n  got:  %+v\n  want: %+v", decoded, original)
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
		SessionID: "sess-123",
		State:     "running",
		CostUSD:   1.23,
	})
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
