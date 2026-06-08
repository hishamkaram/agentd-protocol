package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAgentExtensionDefJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := AgentExtensionDef{
		ID:               "ext-1234",
		RefID:            "codex-skill-abcd",
		Kind:             AgentExtensionKindSkill,
		Invocation:       AgentExtensionInvocationSkillID,
		Scope:            AgentExtensionScopeProject,
		Name:             "openai-docs",
		DisplayName:      "OpenAI Docs",
		Description:      "Use official docs",
		ShortDescription: "Official docs",
		DefaultPrompt:    "Use $openai-docs",
		DependencyCount:  2,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal AgentExtensionDef: %v", err)
	}
	gotJSON := string(raw)
	if strings.Contains(gotJSON, "/home/") {
		t.Fatalf("extension JSON leaked a local path: %s", gotJSON)
	}

	var out AgentExtensionDef
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal AgentExtensionDef: %v", err)
	}
	if out != in {
		t.Fatalf("AgentExtensionDef roundtrip = %+v, want %+v", out, in)
	}
}
