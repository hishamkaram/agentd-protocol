package protocol

const (
	// MsgWorkflowUpdate carries sanitized Claude Workflow lifecycle and progress
	// state from daemon to PWA. It never carries raw workflow scripts, prompts,
	// or local filesystem paths.
	MsgWorkflowUpdate = "workflow_update"
)

const (
	WorkflowStatusRunning   = "running"
	WorkflowStatusCompleted = "completed"
	WorkflowStatusFailed    = "failed"
	WorkflowStatusCanceled  = "canceled"
)

// WorkflowPayload is carried in an AgentMessage with Type MsgWorkflowUpdate.
// WorkflowID is stable for client-side grouping and is currently the Claude task
// ID; RunID carries the CLI workflow run identifier when it becomes available.
// Optional fields use omitempty for backward-compatible incremental updates.
type WorkflowPayload struct {
	WorkflowID      string                  `json:"workflow_id"`
	TaskID          string                  `json:"task_id"`
	ToolUseID       string                  `json:"tool_use_id,omitempty"`
	RunID           string                  `json:"run_id,omitempty"`
	WorkflowName    string                  `json:"workflow_name,omitempty"`
	Status          string                  `json:"status"`
	Summary         string                  `json:"summary,omitempty"`
	CurrentStep     string                  `json:"current_step,omitempty"`
	Phases          []WorkflowPhaseProgress `json:"phases,omitempty"`
	Agents          []WorkflowAgentProgress `json:"agents,omitempty"`
	Usage           *WorkflowUsage          `json:"usage,omitempty"`
	HasArtifact     bool                    `json:"has_artifact,omitempty"`
	UpdatedAtUnixMS int64                   `json:"updated_at_unix_ms,omitempty"`
}

type WorkflowUsage struct {
	TotalTokens int `json:"total_tokens,omitempty"`
	ToolUses    int `json:"tool_uses,omitempty"`
	DurationMs  int `json:"duration_ms,omitempty"`
}

type WorkflowPhaseProgress struct {
	Index  int    `json:"index"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

type WorkflowAgentProgress struct {
	Index         int    `json:"index"`
	AgentID       string `json:"agent_id,omitempty"`
	Label         string `json:"label,omitempty"`
	PhaseIndex    int    `json:"phase_index,omitempty"`
	PhaseTitle    string `json:"phase_title,omitempty"`
	Model         string `json:"model,omitempty"`
	State         string `json:"state,omitempty"`
	Tokens        int    `json:"tokens,omitempty"`
	ToolCalls     int    `json:"tool_calls,omitempty"`
	DurationMs    int    `json:"duration_ms,omitempty"`
	Attempt       int    `json:"attempt,omitempty"`
	ResultPreview string `json:"result_preview,omitempty"`
}
