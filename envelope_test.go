package protocol

import (
	"encoding/json"
	"testing"
)

func TestRelayEnvelopeReliableJSONShape(t *testing.T) {
	t.Parallel()

	in := RelayEnvelope{
		SessionID:      "session-123",
		Seq:            7,
		Encrypted:      []byte("ciphertext"),
		TraceID:        "trace-123",
		EnvelopeID:     "env-123",
		TargetClientID: "client-123",
		Reliable:       true,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal RelayEnvelope: %v", err)
	}
	const wantJSON = `{"sid":"session-123","seq":7,"enc":"Y2lwaGVydGV4dA==","tid":"trace-123","eid":"env-123","client_id":"client-123","rel":true}`
	if string(raw) != wantJSON {
		t.Fatalf("RelayEnvelope JSON = %s, want %s", raw, wantJSON)
	}

	var out RelayEnvelope
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal RelayEnvelope: %v", err)
	}
	if out.SessionID != in.SessionID || out.Seq != in.Seq || out.TraceID != in.TraceID || out.EnvelopeID != in.EnvelopeID || out.TargetClientID != in.TargetClientID || out.Reliable != in.Reliable {
		t.Fatalf("RelayEnvelope scalar roundtrip = %+v, want %+v", out, in)
	}
	if string(out.Encrypted) != string(in.Encrypted) {
		t.Fatalf("RelayEnvelope encrypted roundtrip = %q, want %q", out.Encrypted, in.Encrypted)
	}
}

func TestRelayEnvelopeOmitsFalseReliable(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(RelayEnvelope{
		SessionID: "session-123",
		Seq:       7,
		Encrypted: []byte("ciphertext"),
	})
	if err != nil {
		t.Fatalf("marshal RelayEnvelope: %v", err)
	}
	const wantJSON = `{"sid":"session-123","seq":7,"enc":"Y2lwaGVydGV4dA=="}`
	if string(raw) != wantJSON {
		t.Fatalf("RelayEnvelope JSON = %s, want %s", raw, wantJSON)
	}
}
