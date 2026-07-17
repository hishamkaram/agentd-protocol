package protocol_test

import (
	"bytes"
	"encoding/json"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestKeyRotatePayloadOmitemptyNotUsed(t *testing.T) {
	t.Parallel()
	payload := protocol.KeyRotatePayload{
		NewKeyHMAC: "abc", PrevKeyHMAC: "def", PrevExpiresAt: 1744300000, Epoch: 1,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, field := range []string{`"new_key_hmac"`, `"prev_key_hmac"`, `"prev_expires_at"`, `"epoch"`} {
		if !bytes.Contains(data, []byte(field)) {
			t.Errorf("field %s must be present in JSON, got: %s", field, data)
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
		protocol.CtrlPushNotify:           "push_notify",
		protocol.CtrlPushNotifyResult:     "push_notify_result",
		protocol.CtrlTerminateSession:     "terminate_session",
		protocol.CtrlTerminateSessionAck:  "terminate_session_ack",
		protocol.CtrlRouteReceipt:         "route_receipt",
	}
	for controlType, want := range expected {
		if string(controlType) != want {
			t.Errorf("ControlType %q != %q", controlType, want)
		}
	}
}

func TestRelayEnvelopeEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		envelope protocol.RelayEnvelope
	}{
		{name: "zero-value struct", envelope: protocol.RelayEnvelope{}},
		{name: "nil Encrypted field", envelope: protocol.RelayEnvelope{SessionID: "sess-nil-enc", Seq: 7}},
		{
			name: "opaque route envelope id",
			envelope: protocol.RelayEnvelope{
				SessionID: "sess-route-id", Seq: 8, Encrypted: []byte("ciphertext"),
				TraceID: "0123456789abcdef0123456789abcdef", EnvelopeID: "cmd-123",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, test.envelope)
		})
	}
}

func TestControlMessageEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("zero-value struct marshal unmarshal", func(t *testing.T) {
		t.Parallel()
		data, err := json.Marshal(protocol.ControlMessage{})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded protocol.ControlMessage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Type != "" || decoded.Payload == nil {
			t.Fatalf("unexpected zero-value roundtrip: %+v", decoded)
		}
	})
	tests := []struct {
		name string
		msg  protocol.ControlMessage
	}{
		{name: "unknown control type", msg: protocol.ControlMessage{Type: "unknown_type_xyz", Payload: json.RawMessage(`{"foo":"bar"}`)}},
		{name: "null payload roundtrips", msg: protocol.ControlMessage{Type: protocol.CtrlHeartbeat, Payload: json.RawMessage("null")}},
		{name: "empty JSON object payload", msg: protocol.ControlMessage{Type: protocol.CtrlAck, Payload: json.RawMessage(`{}`)}},
		{
			name: "route receipt payload",
			msg: protocol.ControlMessage{Type: protocol.CtrlRouteReceipt, Payload: mustJSON(t, protocol.RouteReceiptPayload{
				EnvelopeID: "cmd-123", SessionID: "relay-session",
				TraceID: "0123456789abcdef0123456789abcdef", Routed: false,
				ReasonCode: "relay_route_failed", ObservedAtUnixMs: 1710000000000,
			})},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, test.msg)
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
		assertRoundtrip(t, protocol.PolicyJSON{Name: "nil-slices", Match: protocol.PolicyMatchJSON{}, Action: "deny"})
	})
	t.Run("empty slices become nil after roundtrip due to omitempty", func(t *testing.T) {
		t.Parallel()
		original := protocol.PolicyJSON{
			Name: "empty-match", Match: protocol.PolicyMatchJSON{Tool: []string{}, RiskLevel: []string{}}, Action: "allow",
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded protocol.PolicyJSON
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Name != "empty-match" || decoded.Action != "allow" {
			t.Fatalf("unexpected policy identity: %+v", decoded)
		}
		if decoded.Match.Tool != nil || decoded.Match.RiskLevel != nil {
			t.Fatalf("empty slices should decode as nil: %+v", decoded.Match)
		}
	})
}

func TestSyncPoliciesPayloadEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		payload protocol.SyncPoliciesPayload
	}{
		{name: "zero-value struct", payload: protocol.SyncPoliciesPayload{}},
		{name: "empty policies slice", payload: protocol.SyncPoliciesPayload{Policies: []protocol.PolicyJSON{}}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertRoundtrip(t, test.payload)
		})
	}
}

func TestInvalidJSONDecode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		target interface{}
	}{
		{name: "invalid JSON for ControlMessage", input: `{invalid}`, target: &protocol.ControlMessage{}},
		{name: "invalid JSON for RelayEnvelope", input: `{invalid}`, target: &protocol.RelayEnvelope{}},
		{name: "truncated JSON for ControlMessage", input: `{"type":"register","pay`, target: &protocol.ControlMessage{}},
		{name: "truncated JSON for RelayEnvelope", input: `{"sid":"abc","seq":1,"en`, target: &protocol.RelayEnvelope{}},
		{name: "wrong type for seq field", input: `{"sid":"abc","seq":"not-a-number","enc":"AA=="}`, target: &protocol.RelayEnvelope{}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := json.Unmarshal([]byte(test.input), test.target); err == nil {
				t.Errorf("expected error for input %q, got nil", test.input)
			}
		})
	}
}
