package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRelayEnvelopeSchemaAcceptsOpaqueRecoveryFields(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("schemas", "fixtures", "relay-envelope-opaque-recovery.json"))
	if err != nil {
		t.Fatalf("read recovery fixture: %v", err)
	}
	var envelope RelayEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("unmarshal recovery fixture: %v", err)
	}
	if !envelope.Reliable || envelope.EnvelopeID == "" || len(envelope.OpaqueControl) == 0 {
		t.Fatalf("recovery fixture is missing reliable correlation fields: %+v", envelope)
	}

	var fixtureFields map[string]any
	if err := json.Unmarshal(raw, &fixtureFields); err != nil {
		t.Fatalf("unmarshal recovery fixture fields: %v", err)
	}
	schema := readV2Schema(t, "relay-envelope.schema.json")
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("relay envelope schema properties missing")
	}
	for name := range fixtureFields {
		if _, declared := properties[name]; !declared {
			t.Errorf("relay envelope schema rejects recovery field %q", name)
		}
	}
}
