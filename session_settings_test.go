package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestSetSessionSettingsRoundTrip(t *testing.T) {
	t.Parallel()

	workflow := false
	in := protocol.SetSessionSettingsRequest{
		Type:               protocol.MsgSetSessionSettings,
		RequestID:          "req-1",
		SessionID:          "sid-1",
		Model:              "gpt-5.4",
		Effort:             "xhigh",
		WorkflowAutomation: &workflow,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	var got protocol.SetSessionSettingsRequest
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}

	payload := string(raw)
	for _, want := range []string{
		`"type":"set_session_settings"`,
		`"request_id":"req-1"`,
		`"session_id":"sid-1"`,
		`"model":"gpt-5.4"`,
		`"effort":"xhigh"`,
		`"workflow_automation":false`,
	} {
		if !strings.Contains(payload, want) {
			t.Fatalf("request JSON missing %s: %s", want, payload)
		}
	}
	for _, forbidden := range []string{"workflow_script", "prompt", "local_path", "settings_json"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("request JSON leaked forbidden key %q: %s", forbidden, payload)
		}
	}
}

func TestSetSessionSettingsResultRoundTrip(t *testing.T) {
	t.Parallel()

	workflow := true
	in := protocol.SetSessionSettingsResult{
		Type:      protocol.MsgSetSessionSettingsResult,
		RequestID: "req-1",
		SessionID: "sid-1",
		Success:   true,
		Settings: protocol.SessionSettingsPayload{
			Model:              "sonnet",
			Effort:             "high",
			WorkflowAutomation: &workflow,
			Source:             protocol.SessionRuntimeSettingSourceProviderReadback,
		},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var got protocol.SetSessionSettingsResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}
}

func TestSessionRuntimeSettingsCapabilityRoundTrip(t *testing.T) {
	t.Parallel()

	workflow := true
	in := protocol.SessionRuntimeSettingsCapability{
		SchemaVersion: protocol.SessionRuntimeSettingsCapabilitySchemaVersion,
		Settings: []protocol.SessionRuntimeSettingCapability{
			{
				ID:              protocol.SessionRuntimeSettingEffort,
				Supported:       true,
				SupportedValues: []string{"low", "medium", "high", "xhigh"},
				CurrentValue:    "high",
				Source:          protocol.SessionRuntimeSettingSourceModelMetadata,
			},
			{
				ID:             protocol.SessionRuntimeSettingWorkflowAutomation,
				Supported:      true,
				CurrentBool:    &workflow,
				Source:         protocol.SessionRuntimeSettingSourceProviderReadback,
				DisabledReason: "",
			},
		},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal capability: %v", err)
	}
	var got protocol.SessionRuntimeSettingsCapability
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal capability: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}
}
