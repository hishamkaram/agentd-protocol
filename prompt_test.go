package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMsgInteractivePromptResolvedConstant(t *testing.T) {
	t.Parallel()

	if MsgInteractivePromptResolved != "interactive_prompt_resolved" {
		t.Fatalf("MsgInteractivePromptResolved = %q, want %q", MsgInteractivePromptResolved, "interactive_prompt_resolved")
	}
}

func TestInteractivePromptResolvedPayloadRoundtrip(t *testing.T) {
	t.Parallel()

	original := InteractivePromptResolvedPayload{
		QuestionID: "q-123",
		SessionID:  "s-123",
		Reason:     "answered",
		ResolvedAt: 1778915600123,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	rawStr := string(raw)
	for _, want := range []string{`"question_id"`, `"session_id"`, `"reason"`, `"resolved_at"`} {
		if !strings.Contains(rawStr, want) {
			t.Fatalf("marshaled JSON missing %s: %s", want, rawStr)
		}
	}

	var decoded InteractivePromptResolvedPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != original {
		t.Fatalf("roundtrip mismatch:\n  got:  %+v\n  want: %+v", decoded, original)
	}
}
