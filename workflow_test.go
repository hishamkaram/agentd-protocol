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

	for name, value := range map[string]string{
		"workflow_update":         MsgWorkflowUpdate,
		"workflow_control":        MsgWorkflowControl,
		"workflow_control_result": MsgWorkflowControlResult,
		"list_workflows":          MsgListWorkflows,
		"workflow_list":           MsgWorkflowList,
	} {
		if value != name {
			t.Fatalf("%s message = %q", name, value)
		}
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
	if WorkflowControlActionStop != "stop" {
		t.Fatalf("WorkflowControlActionStop = %q, want stop", WorkflowControlActionStop)
	}
}

func TestWorkflowControlRequestRoundtrip(t *testing.T) {
	t.Parallel()

	original := WorkflowControlRequest{
		Type:       MsgWorkflowControl,
		RequestID:  "req-1",
		SessionID:  "session-1",
		WorkflowID: "workflow-1",
		TaskID:     "task-1",
		RunID:      "run-1",
		Action:     WorkflowControlActionStop,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorkflowControlRequest
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Type != MsgWorkflowControl ||
		decoded.RequestID != "req-1" ||
		decoded.SessionID != "session-1" ||
		decoded.WorkflowID != "workflow-1" ||
		decoded.TaskID != "task-1" ||
		decoded.RunID != "run-1" ||
		decoded.Action != WorkflowControlActionStop {
		t.Fatalf("decoded workflow control request = %+v", decoded)
	}
}

func TestWorkflowControlResultRoundtrip(t *testing.T) {
	t.Parallel()

	original := WorkflowControlResult{
		Type:       MsgWorkflowControlResult,
		RequestID:  "req-1",
		SessionID:  "session-1",
		WorkflowID: "workflow-1",
		TaskID:     "task-1",
		RunID:      "run-1",
		Action:     WorkflowControlActionStop,
		Success:    true,
		Status:     WorkflowControlStatusStopped,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorkflowControlResult
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Type != MsgWorkflowControlResult ||
		decoded.RequestID != "req-1" ||
		decoded.SessionID != "session-1" ||
		decoded.TaskID != "task-1" ||
		decoded.Action != WorkflowControlActionStop ||
		!decoded.Success ||
		decoded.Status != WorkflowControlStatusStopped {
		t.Fatalf("decoded workflow control result = %+v", decoded)
	}
}

func TestWorkflowListPayloadRoundtripAndRedactionShape(t *testing.T) {
	t.Parallel()

	original := WorkflowListPayload{
		Type:         MsgWorkflowList,
		RequestID:    "req-1",
		SessionID:    "session-1",
		ProjectLabel: "agentd",
		Items: []WorkflowListItem{{
			ID:              "release-probe",
			Name:            "Release probe",
			Description:     "Verify release readiness",
			Scope:           "project",
			Command:         "release-probe",
			PhaseTitles:     []string{"Plan", "Verify"},
			UpdatedAtUnixMS: 1782904297561,
			SupportsRun:     true,
		}},
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorkflowListPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Type != MsgWorkflowList ||
		decoded.RequestID != "req-1" ||
		decoded.SessionID != "session-1" ||
		decoded.ProjectLabel != "agentd" ||
		len(decoded.Items) != 1 ||
		decoded.Items[0].Command != "release-probe" ||
		!decoded.Items[0].SupportsRun {
		t.Fatalf("decoded workflow list = %+v", decoded)
	}
	body := string(raw)
	for _, forbidden := range []string{"script_path", "scriptPath", "prompt", "transcript", "output_file", "/tmp/", "/home/"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("workflow list leaked %q in %s", forbidden, body)
		}
	}
}
