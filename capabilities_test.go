package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

// TestAgentCapabilityFields asserts the struct has exactly the four required
// bool fields per specs/185-codex-parity-gaps/contracts/agent-capability.md.
func TestAgentCapabilityFields(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.AgentCapability{})
	if typ.Kind() != reflect.Struct {
		t.Fatalf("AgentCapability must be a struct, got %v", typ.Kind())
	}

	want := map[string]string{
		"AnswerQuestion": "answer_question",
		"SendToolResult": "send_tool_result",
		"RewindFiles":    "rewind_files",
		"MCPHotApply":    "mcp_hot_apply",
	}

	if got := typ.NumField(); got != len(want) {
		t.Fatalf("AgentCapability field count: want %d, got %d", len(want), got)
	}

	for fieldName, wantTag := range want {
		wantTag := wantTag
		t.Run(fieldName, func(t *testing.T) {
			t.Parallel()

			field, ok := typ.FieldByName(fieldName)
			if !ok {
				t.Fatalf("AgentCapability missing field %s", fieldName)
			}
			if field.Type.Kind() != reflect.Bool {
				t.Fatalf("AgentCapability.%s: want bool, got %s", fieldName, field.Type.Kind())
			}
			gotTag := field.Tag.Get("json")
			// Tag may optionally include options like ,omitempty; contract
			// specifies required fields — no omitempty — but accept exact
			// match as the canonical form.
			if gotTag != wantTag {
				t.Fatalf("AgentCapability.%s: json tag want %q, got %q", fieldName, wantTag, gotTag)
			}
		})
	}
}

// TestAgentCapabilityRoundTrip verifies marshal → unmarshal preserves values
// across all four fields independently and together.
func TestAgentCapabilityRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.AgentCapability
	}{
		{
			name: "claude all true",
			in: protocol.AgentCapability{
				AnswerQuestion: true,
				SendToolResult: true,
				RewindFiles:    true,
				MCPHotApply:    true,
			},
		},
		{
			name: "codex per data-model",
			in: protocol.AgentCapability{
				AnswerQuestion: false,
				SendToolResult: false,
				RewindFiles:    false,
				MCPHotApply:    true,
			},
		},
		{
			name: "all false",
			in:   protocol.AgentCapability{},
		},
		{
			name: "mixed",
			in: protocol.AgentCapability{
				AnswerQuestion: true,
				SendToolResult: false,
				RewindFiles:    true,
				MCPHotApply:    false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got protocol.AgentCapability
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if !reflect.DeepEqual(got, tt.in) {
				t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", tt.in, got, raw)
			}
		})
	}
}

// TestAgentCapabilityJSONKeys verifies the exact snake_case JSON keys appear
// in the marshaled payload per contracts/agent-capability.md field mapping.
func TestAgentCapabilityJSONKeys(t *testing.T) {
	t.Parallel()

	cap := protocol.AgentCapability{
		AnswerQuestion: true,
		SendToolResult: true,
		RewindFiles:    true,
		MCPHotApply:    true,
	}
	raw, err := json.Marshal(cap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(raw)

	wantKeys := []string{
		`"answer_question":true`,
		`"send_tool_result":true`,
		`"rewind_files":true`,
		`"mcp_hot_apply":true`,
	}
	for _, k := range wantKeys {
		if !strings.Contains(payload, k) {
			t.Errorf("missing key %q in JSON payload: %s", k, payload)
		}
	}

	// Assert no CamelCase / PascalCase leakage.
	forbidden := []string{"AnswerQuestion", "SendToolResult", "RewindFiles", "MCPHotApply"}
	for _, k := range forbidden {
		if strings.Contains(payload, k) {
			t.Errorf("unexpected Go field name %q in JSON payload: %s", k, payload)
		}
	}
}
