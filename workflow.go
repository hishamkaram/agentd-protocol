package protocol

const (
	// MsgWorkflowUpdate carries sanitized Claude Workflow lifecycle and progress
	// state from daemon to PWA. It never carries raw workflow scripts, prompts,
	// or local filesystem paths.
	MsgWorkflowUpdate        = "workflow_update"
	MsgWorkflowControl       = "workflow_control"
	MsgWorkflowControlResult = "workflow_control_result"
	MsgListWorkflows         = "list_workflows"
	MsgWorkflowList          = "workflow_list"
)

const (
	WorkflowStatusRunning   = "running"
	WorkflowStatusCompleted = "completed"
	WorkflowStatusFailed    = "failed"
	WorkflowStatusCanceled  = "canceled"
)

const (
	WorkflowControlActionStop = "stop"
)

const (
	WorkflowControlStatusStopped     = "stopped"
	WorkflowControlStatusUnsupported = "unsupported"
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

// WorkflowControlRequest is sent by the PWA to control a single workflow task
// inside one session. Only task-scoped controls belong here; session stop and
// interrupt remain separate messages.
type WorkflowControlRequest struct {
	Type       string `json:"type"`
	RequestID  string `json:"request_id"`
	SessionID  string `json:"session_id"`
	WorkflowID string `json:"workflow_id"`
	TaskID     string `json:"task_id"`
	RunID      string `json:"run_id,omitempty"`
	AgentID    string `json:"agent_id,omitempty"`
	Action     string `json:"action"`
}

// WorkflowControlResult is the sanitized result for a workflow control
// request. Error text is user-facing; ErrorCode is stable for clients.
type WorkflowControlResult struct {
	Type       string `json:"type"`
	RequestID  string `json:"request_id,omitempty"`
	SessionID  string `json:"session_id"`
	WorkflowID string `json:"workflow_id,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	AgentID    string `json:"agent_id,omitempty"`
	Action     string `json:"action,omitempty"`
	Success    bool   `json:"success"`
	Status     string `json:"status,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ListWorkflowsRequest asks the daemon for a sanitized snapshot of saved
// workflow commands for the session's project scope.
type ListWorkflowsRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	Project   string `json:"project,omitempty"`
}

// WorkflowListPayload is a sanitized catalog snapshot. It never carries raw
// paths, script bodies, prompts, transcript dirs, output files, or provider
// internals.
type WorkflowListPayload struct {
	Type         string             `json:"type"`
	RequestID    string             `json:"request_id,omitempty"`
	SessionID    string             `json:"session_id"`
	ProjectLabel string             `json:"project_label,omitempty"`
	Items        []WorkflowListItem `json:"items"`
}

type WorkflowListItem struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Scope           string   `json:"scope"`
	Command         string   `json:"command"`
	PhaseTitles     []string `json:"phase_titles,omitempty"`
	UpdatedAtUnixMS int64    `json:"updated_at_unix_ms,omitempty"`
	SupportsRun     bool     `json:"supports_run"`
}
