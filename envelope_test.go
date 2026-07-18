package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRelayEnvelopeReliableJSONShape(t *testing.T) {
	t.Parallel()

	in := RelayEnvelope{
		SessionID:      "session-123",
		Seq:            7,
		Encrypted:      []byte("ciphertext"),
		KeyEpoch:       3,
		TraceID:        "trace-123",
		EnvelopeID:     "env-123",
		TargetClientID: "client-123",
		Reliable:       true,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal RelayEnvelope: %v", err)
	}
	const wantJSON = `{"sid":"session-123","seq":7,"enc":"Y2lwaGVydGV4dA==","key_epoch":3,"tid":"trace-123","eid":"env-123","client_id":"client-123","rel":true}`
	if string(raw) != wantJSON {
		t.Fatalf("RelayEnvelope JSON = %s, want %s", raw, wantJSON)
	}

	var out RelayEnvelope
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal RelayEnvelope: %v", err)
	}
	if out.SessionID != in.SessionID || out.Seq != in.Seq || out.KeyEpoch != in.KeyEpoch || out.TraceID != in.TraceID || out.EnvelopeID != in.EnvelopeID || out.TargetClientID != in.TargetClientID || out.Reliable != in.Reliable {
		t.Fatalf("RelayEnvelope scalar roundtrip = %+v, want %+v", out, in)
	}
	if string(out.Encrypted) != string(in.Encrypted) {
		t.Fatalf("RelayEnvelope encrypted roundtrip = %q, want %q", out.Encrypted, in.Encrypted)
	}
}

func TestRelayEnvelopeEmitsZeroKeyEpochAndOmitsFalseReliable(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RelayEnvelope{
		SessionID: "session-123",
		Seq:       7,
		Encrypted: []byte("ciphertext"),
	})
	if err != nil {
		t.Fatalf("marshal RelayEnvelope: %v", err)
	}
	const wantJSON = `{"sid":"session-123","seq":7,"enc":"Y2lwaGVydGV4dA==","key_epoch":0}`
	if string(raw) != wantJSON {
		t.Fatalf("RelayEnvelope JSON = %s, want %s", raw, wantJSON)
	}
}

func TestRelayEnvelopeOpaqueControlRoundTrip(t *testing.T) {
	t.Parallel()
	in := RelayEnvelope{
		SessionID:      "session-1",
		Encrypted:      []byte{},
		KeyEpoch:       7,
		TargetClientID: "client-1",
		Reliable:       true,
		OpaqueControl:  []byte(`{"type":"client_key_sync_response"}`),
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal RelayEnvelope: %v", err)
	}
	if !strings.Contains(string(raw), `"opaque_ctrl":"eyJ0eXBlIjoiY2xpZW50X2tleV9zeW5jX3Jlc3BvbnNlIn0="`) {
		t.Fatalf("RelayEnvelope JSON = %s, want base64 opaque control", raw)
	}
	var out RelayEnvelope
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal RelayEnvelope: %v", err)
	}
	if string(out.OpaqueControl) != string(in.OpaqueControl) {
		t.Fatalf("OpaqueControl = %q, want %q", out.OpaqueControl, in.OpaqueControl)
	}
}
