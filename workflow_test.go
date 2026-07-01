package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflowPayloadRoundtrip(t *testing.T) {
	t.Parallel()

	original := WorkflowPayload{
		WorkflowID:   "wtismrl5c",
		TaskID:       "wtismrl5c",
		ToolUseID:    "toolu_workflow",
		RunID:        "wf_2c525e4f-988",
		WorkflowName: "minimal-probe",
		Status:       WorkflowStatusCompleted,
		Summary:      "Dynamic workflow completed",
		CurrentStep:  "Probe",
		Phases: []WorkflowPhaseProgress{{
			Index: 1,
			Title: "Probe",
		}},
		Agents: []WorkflowAgentProgress{{
			Index:         1,
			Label:         "probe",
			PhaseIndex:    1,
			PhaseTitle:    "Probe",
			Model:         "claude-opus-4-8",
			State:         "done",
			Tokens:        30574,
			ToolCalls:     0,
			DurationMs:    3283,
			Attempt:       1,
			ResultPreview: "workflow-ok",
		}},
		Usage: &WorkflowUsage{
			TotalTokens: 30574,
			DurationMs:  3470,
		},
		HasArtifact:     true,
		UpdatedAtUnixMS: 1782904297561,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorkflowPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.WorkflowID != original.WorkflowID ||
		decoded.TaskID != original.TaskID ||
		decoded.WorkflowName != original.WorkflowName ||
		decoded.Status != WorkflowStatusCompleted ||
		decoded.Usage == nil ||
		decoded.Usage.TotalTokens != 30574 ||
		!decoded.HasArtifact ||
		len(decoded.Phases) != 1 ||
		len(decoded.Agents) != 1 ||
		decoded.Agents[0].ResultPreview != "workflow-ok" {
		t.Fatalf("decoded workflow payload = %+v", decoded)
	}
}

func TestWorkflowPayloadNoScriptPromptOrPathFields(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(WorkflowPayload{
		WorkflowID:   "wf_1",
		TaskID:       "task_1",
		Status:       WorkflowStatusRunning,
		WorkflowName: "minimal-probe",
		HasArtifact:  true,
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	body := string(raw)
	for _, forbidden := range []string{"script", "scriptPath", "transcriptDir", "prompt", "/tmp/", "/home/"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("workflow payload leaked %q in %s", forbidden, body)
		}
	}
}

func TestWorkflowWireConstants(t *testing.T) {
	t.Parallel()

	if MsgWorkflowUpdate != "workflow_update" {
		t.Fatalf("MsgWorkflowUpdate = %q, want workflow_update", MsgWorkflowUpdate)
	}
	for name, value := range map[string]string{
		"running":   WorkflowStatusRunning,
		"completed": WorkflowStatusCompleted,
		"failed":    WorkflowStatusFailed,
		"canceled":  WorkflowStatusCanceled,
	} {
		if value != name {
			t.Fatalf("%s status = %q", name, value)
		}
	}
}
