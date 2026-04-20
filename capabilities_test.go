package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

// TestAgentCapabilityFields asserts the struct has exactly the six required
// bool fields per specs/185-codex-parity-gaps/contracts/agent-capability.md
// (original 4 fields) and specs/186-codex-remaining-gaps/contracts/
// session-scoped-approval-capability.md + answer-question-free-text-capability.md
// (2 new fields added by feature 186).
func TestAgentCapabilityFields(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.AgentCapability{})
	if typ.Kind() != reflect.Struct {
		t.Fatalf("AgentCapability must be a struct, got %v", typ.Kind())
	}

	want := map[string]string{
		"AnswerQuestion":         "answer_question",
		"SendToolResult":         "send_tool_result",
		"RewindFiles":            "rewind_files",
		"MCPHotApply":            "mcp_hot_apply",
		"SessionScopedApproval":  "session_scoped_approval",
		"AnswerQuestionFreeText": "answer_question_free_text",
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
	forbidden := []string{"AnswerQuestion", "SendToolResult", "RewindFiles", "MCPHotApply", "SessionScopedApproval", "AnswerQuestionFreeText"}
	for _, k := range forbidden {
		if strings.Contains(payload, k) {
			t.Errorf("unexpected Go field name %q in JSON payload: %s", k, payload)
		}
	}
}

// Feature 186 — T002-T: new capability fields `SessionScopedApproval` and
// `AnswerQuestionFreeText`. These tests verify the new fields are bool, carry
// canonical JSON tags, have the zero-value default (false), and survive a
// marshal → unmarshal round-trip independently of the existing 4 fields.

// TestAgentCapabilitySessionScopedApprovalField pins the new field's type,
// JSON tag, and zero-value semantics.
func TestAgentCapabilitySessionScopedApprovalField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.AgentCapability{})
	field, ok := typ.FieldByName("SessionScopedApproval")
	if !ok {
		t.Fatalf("AgentCapability missing field SessionScopedApproval")
	}
	if field.Type.Kind() != reflect.Bool {
		t.Fatalf("SessionScopedApproval: want bool, got %s", field.Type.Kind())
	}
	if got := field.Tag.Get("json"); got != "session_scoped_approval" {
		t.Fatalf("SessionScopedApproval: json tag want %q, got %q", "session_scoped_approval", got)
	}

	var zero protocol.AgentCapability
	if zero.SessionScopedApproval != false {
		t.Fatalf("zero-value SessionScopedApproval: want false, got %v", zero.SessionScopedApproval)
	}
}

// TestAgentCapabilityAnswerQuestionFreeTextField pins the new field's type,
// JSON tag, and zero-value semantics.
func TestAgentCapabilityAnswerQuestionFreeTextField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.AgentCapability{})
	field, ok := typ.FieldByName("AnswerQuestionFreeText")
	if !ok {
		t.Fatalf("AgentCapability missing field AnswerQuestionFreeText")
	}
	if field.Type.Kind() != reflect.Bool {
		t.Fatalf("AnswerQuestionFreeText: want bool, got %s", field.Type.Kind())
	}
	if got := field.Tag.Get("json"); got != "answer_question_free_text" {
		t.Fatalf("AnswerQuestionFreeText: json tag want %q, got %q", "answer_question_free_text", got)
	}

	var zero protocol.AgentCapability
	if zero.AnswerQuestionFreeText != false {
		t.Fatalf("zero-value AnswerQuestionFreeText: want false, got %v", zero.AnswerQuestionFreeText)
	}
}

// TestAgentCapability186RoundTrip verifies the new fields marshal and
// unmarshal with the other 4 correctly across the per-agent shapes called out
// in data-model.md (Claude: AnswerQuestionFreeText=true, SSA=false; Codex:
// AnswerQuestionFreeText=false, SSA=true).
func TestAgentCapability186RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.AgentCapability
	}{
		{
			name: "claude 186 shape",
			in: protocol.AgentCapability{
				AnswerQuestion:         true,
				SendToolResult:         true,
				RewindFiles:            true,
				MCPHotApply:            true,
				SessionScopedApproval:  false,
				AnswerQuestionFreeText: true,
			},
		},
		{
			name: "codex 186 shape",
			in: protocol.AgentCapability{
				AnswerQuestion:         false,
				SendToolResult:         false,
				RewindFiles:            false,
				MCPHotApply:            true,
				SessionScopedApproval:  true,
				AnswerQuestionFreeText: false,
			},
		},
		{
			name: "ssa true only",
			in: protocol.AgentCapability{
				SessionScopedApproval: true,
			},
		},
		{
			name: "free-text true only",
			in: protocol.AgentCapability{
				AnswerQuestionFreeText: true,
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

// TestAgentCapability186JSONKeys verifies the new snake_case keys appear in
// the marshaled payload exactly per the contract.
func TestAgentCapability186JSONKeys(t *testing.T) {
	t.Parallel()

	cap := protocol.AgentCapability{
		SessionScopedApproval:  true,
		AnswerQuestionFreeText: false,
	}
	raw, err := json.Marshal(cap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(raw)

	wantKeys := []string{
		`"session_scoped_approval":true`,
		`"answer_question_free_text":false`,
	}
	for _, k := range wantKeys {
		if !strings.Contains(payload, k) {
			t.Errorf("missing key %q in JSON payload: %s", k, payload)
		}
	}
}

// TestAgentCapabilityAnswerQuestionDeprecationCoexistence pins the
// feature-186 FR-017 deprecation pattern: `AnswerQuestion` (deprecated
// in 186, retained for one release) MUST emit alongside the new
// `AnswerQuestionFreeText` field, with matching per-agent values.
//
// Claude: both fields true  (free-text reply row rendered)
// Codex:  both fields false (structured elicitation only)
//
// PWA consumers read `answer_question_free_text ?? answer_question ??
// true` — the new field wins, the legacy field is the fallback, the
// default favors claude-style rendering for old daemons with neither
// field populated.
func TestAgentCapabilityAnswerQuestionDeprecationCoexistence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                       string
		cap                        protocol.AgentCapability
		wantAQ, wantAQFT           bool
		wantKeepsBackCompatEmitted bool
	}{
		{
			name: "claude both true",
			cap: protocol.AgentCapability{
				AnswerQuestion:         true,
				AnswerQuestionFreeText: true,
			},
			wantAQ:                     true,
			wantAQFT:                   true,
			wantKeepsBackCompatEmitted: true,
		},
		{
			name: "codex both false (zero-value)",
			cap: protocol.AgentCapability{
				AnswerQuestion:         false,
				AnswerQuestionFreeText: false,
			},
			wantAQ:                     false,
			wantAQFT:                   false,
			wantKeepsBackCompatEmitted: true, // key still appears in JSON (no omitempty)
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.cap)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			payload := string(raw)

			// Both snake-case keys MUST appear in the same payload.
			if !strings.Contains(payload, `"answer_question":`) {
				t.Errorf("deprecated `answer_question` key missing from payload; back-compat broken: %s", payload)
			}
			if !strings.Contains(payload, `"answer_question_free_text":`) {
				t.Errorf("new `answer_question_free_text` key missing from payload: %s", payload)
			}

			// Round-trip preserves both fields independently.
			var got protocol.AgentCapability
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.AnswerQuestion != tt.wantAQ {
				t.Errorf("AnswerQuestion after round-trip = %v, want %v", got.AnswerQuestion, tt.wantAQ)
			}
			if got.AnswerQuestionFreeText != tt.wantAQFT {
				t.Errorf("AnswerQuestionFreeText after round-trip = %v, want %v", got.AnswerQuestionFreeText, tt.wantAQFT)
			}
		})
	}
}
