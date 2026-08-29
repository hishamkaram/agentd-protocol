package protocol

import (
	"fmt"
	"strings"
)

const (
	// MsgBackgroundControl carries task/terminal-scoped background controls.
	// Session stop and turn interrupt remain separate message types.
	MsgBackgroundControl       = "background_control"
	MsgBackgroundControlResult = "background_control_result"
	MsgBackgroundState         = "background_state"
)

const (
	BackgroundControlActionStopTask         = "stop_task"
	BackgroundControlActionStopAllTerminals = "stop_all_terminals"
)

const (
	BackgroundControlStatusAccepted    = "accepted"
	BackgroundControlStatusUnsupported = "unsupported"
)

// BackgroundControlRequest asks the daemon to invoke a narrow provider
// background-control capability. A successful result only means the provider
// accepted the request; task and terminal lifecycle remains provider-authored.
type BackgroundControlRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
	TaskID    string `json:"task_id,omitempty"`
	Action    string `json:"action"`
}

// Validate enforces the action/target contract for background controls while
// preserving normal JSON backward compatibility for optional future fields.
func (r BackgroundControlRequest) Validate() error {
	if strings.TrimSpace(r.Type) != "" && r.Type != MsgBackgroundControl {
		return fmt.Errorf("background control type %q must be %q", r.Type, MsgBackgroundControl)
	}
	if strings.TrimSpace(r.RequestID) == "" {
		return fmt.Errorf("background control request_id is required")
	}
	if strings.TrimSpace(r.SessionID) == "" {
		return fmt.Errorf("background control session_id is required")
	}
	switch r.Action {
	case BackgroundControlActionStopTask:
		if strings.TrimSpace(r.TaskID) == "" {
			return fmt.Errorf("background control task_id is required for %s", r.Action)
		}
	case BackgroundControlActionStopAllTerminals:
		if strings.TrimSpace(r.TaskID) != "" {
			return fmt.Errorf("background control task_id is incompatible with %s", r.Action)
		}
	default:
		return fmt.Errorf("background control action %q is unsupported", r.Action)
	}
	return nil
}

// BackgroundControlResult is the sanitized, request-correlated response to a
// background control request. Error text is safe for clients.
type BackgroundControlResult struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id,omitempty"`
	SessionID string `json:"session_id"`
	TaskID    string `json:"task_id,omitempty"`
	Action    string `json:"action,omitempty"`
	Success   bool   `json:"success"`
	Status    string `json:"status,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

// BackgroundStatePayload carries provider-authored background lifecycle state.
// Terminal handles are session-scoped and never stable across sessions.
type BackgroundStatePayload struct {
	Type                  string   `json:"type"`
	SessionID             string   `json:"session_id"`
	ActiveTaskIDs         []string `json:"active_task_ids,omitempty"`
	ActiveTerminalHandles []string `json:"active_terminal_handles,omitempty"`
	ActiveTerminalCount   int      `json:"active_terminal_count,omitempty"`
	UpdatedAtUnixMS       int64    `json:"updated_at_unix_ms,omitempty"`
}

// Validate checks a background state update before local relay to clients.
func (p BackgroundStatePayload) Validate() error {
	if strings.TrimSpace(p.Type) != "" && p.Type != MsgBackgroundState {
		return fmt.Errorf("background state type %q must be %q", p.Type, MsgBackgroundState)
	}
	if strings.TrimSpace(p.SessionID) == "" {
		return fmt.Errorf("background state session_id is required")
	}
	if p.ActiveTerminalCount < 0 {
		return fmt.Errorf("background state active_terminal_count must be non-negative")
	}
	if p.ActiveTerminalHandles != nil && p.ActiveTerminalCount != 0 && p.ActiveTerminalCount != len(p.ActiveTerminalHandles) {
		return fmt.Errorf("background state active_terminal_count must match active_terminal_handles length")
	}
	if p.UpdatedAtUnixMS < 0 {
		return fmt.Errorf("background state updated_at_unix_ms must be non-negative")
	}
	return nil
}
