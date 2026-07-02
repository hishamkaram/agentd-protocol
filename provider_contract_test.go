package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestProviderCapabilityContractRoundTrip(t *testing.T) {
	t.Parallel()

	in := protocol.ProviderCapabilityContract{
		SchemaVersion: protocol.ProviderCapabilityContractSchemaVersion,
		Commands: []protocol.ProviderCommandDescriptor{{
			Name:                "plan",
			ArgumentRequirement: protocol.ProviderCommandArgumentOptional,
			Lifecycle:           protocol.ProviderCommandLifecycleProviderTurn,
			StatusAfterDispatch: protocol.ProviderCommandStatusRunning,
			Source:              protocol.ProviderCapabilitySourcePromptComposed,
			Support:             protocol.ProviderCapabilitySupported,
			UserDescription:     "Draft a plan before making changes.",
			Limitations:         []string{"No SDK RPC exists for Codex plan mode."},
		}},
		Approval: []protocol.ProviderFeatureDescriptor{{
			ID:              "approval.session_scoped",
			Support:         protocol.ProviderCapabilitySupported,
			Source:          protocol.ProviderCapabilitySourceDaemonConfig,
			UserLabel:       "Session approvals",
			UserDescription: "Allow this action for the rest of the session.",
		}},
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got protocol.ProviderCapabilityContract
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}
}

func TestProviderCapabilityContractJSONKeys(t *testing.T) {
	t.Parallel()

	in := protocol.ProviderCapabilityContract{
		SchemaVersion: protocol.ProviderCapabilityContractSchemaVersion,
		Commands: []protocol.ProviderCommandDescriptor{{
			Name:                "model",
			ArgumentRequirement: protocol.ProviderCommandArgumentRequired,
			Lifecycle:           protocol.ProviderCommandLifecycleProviderControl,
			StatusAfterDispatch: protocol.ProviderCommandStatusUnchanged,
			Source:              protocol.ProviderCapabilitySourceSDKRPC,
			Support:             protocol.ProviderCapabilitySupported,
			UserDescription:     "Switch the model.",
		}},
		Model: []protocol.ProviderFeatureDescriptor{{
			ID:              "model.switch",
			Support:         protocol.ProviderCapabilitySupported,
			Source:          protocol.ProviderCapabilitySourceSDKRPC,
			UserLabel:       "Model switching",
			UserDescription: "Switch models during a session.",
		}},
		RuntimeSettings: []protocol.ProviderFeatureDescriptor{{
			ID:              "effort.apply",
			Support:         protocol.ProviderCapabilitySupported,
			Source:          protocol.ProviderCapabilitySourceSDKRPC,
			UserLabel:       "Effort",
			UserDescription: "Change reasoning effort during a session.",
		}},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(raw)

	wantKeys := []string{
		`"schema_version":1`,
		`"commands":[`,
		`"argument_requirement":"required"`,
		`"status_after_dispatch":"unchanged"`,
		`"user_description":"Switch the model."`,
		`"model":[`,
		`"user_label":"Model switching"`,
		`"runtime_settings":[`,
		`"user_label":"Effort"`,
	}
	for _, key := range wantKeys {
		if !strings.Contains(payload, key) {
			t.Errorf("missing key %q in JSON payload: %s", key, payload)
		}
	}
	forbidden := []string{"SchemaVersion", "ArgumentRequirement", "StatusAfterDispatch", "UserDescription"}
	for _, key := range forbidden {
		if strings.Contains(payload, key) {
			t.Errorf("unexpected Go field name %q in JSON payload: %s", key, payload)
		}
	}
}

func TestSessionInfoProviderContractField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.SessionInfo{})
	field, ok := typ.FieldByName("ProviderContract")
	if !ok {
		t.Fatalf("SessionInfo missing ProviderContract field")
	}
	if field.Type.Kind() != reflect.Pointer {
		t.Fatalf("SessionInfo.ProviderContract: want pointer, got %s", field.Type.Kind())
	}
	if got := field.Tag.Get("json"); got != "provider_contract,omitempty" {
		t.Fatalf("SessionInfo.ProviderContract json tag: want %q, got %q", "provider_contract,omitempty", got)
	}

	raw, err := json.Marshal(protocol.SessionInfo{})
	if err != nil {
		t.Fatalf("marshal zero: %v", err)
	}
	if strings.Contains(string(raw), "provider_contract") {
		t.Fatalf("zero SessionInfo unexpectedly marshaled provider_contract: %s", raw)
	}
}
