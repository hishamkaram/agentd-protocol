package protocol

import (
	"fmt"
	"strings"
)

const (
	// MsgBackgroundControl carries provider-owned background controls.
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

// TerminalObservationState describes whether terminal inventory is safe to
// act on. Empty is reserved for legacy daemons that did not report inventory
// confidence.
type TerminalObservationState string

const (
	TerminalObservationUnknown  TerminalObservationState = "unknown"
	TerminalObservationCurrent  TerminalObservationState = "current"
	TerminalObservationDegraded TerminalObservationState = "degraded"
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
	return validateBackgroundControlTarget(r.Action, r.TaskID)
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

// Validate enforces request correlation and the same exact target contract as
// BackgroundControlRequest. Provider acceptance remains distinct from target
// lifecycle completion.
func (r BackgroundControlResult) Validate() error {
	if strings.TrimSpace(r.Type) != "" && r.Type != MsgBackgroundControlResult {
		return fmt.Errorf("background control result type %q must be %q", r.Type, MsgBackgroundControlResult)
	}
	if strings.TrimSpace(r.RequestID) == "" {
		return fmt.Errorf("background control result request_id is required")
	}
	if strings.TrimSpace(r.SessionID) == "" {
		return fmt.Errorf("background control result session_id is required")
	}
	return validateBackgroundControlTarget(r.Action, r.TaskID)
}

func validateBackgroundControlTarget(action, taskID string) error {
	switch action {
	case BackgroundControlActionStopTask:
		if strings.TrimSpace(taskID) == "" {
			return fmt.Errorf("background control task_id is the only valid target for %s", action)
		}
	case BackgroundControlActionStopAllTerminals:
		if strings.TrimSpace(taskID) != "" {
			return fmt.Errorf("background control targets are incompatible with %s", action)
		}
	default:
		return fmt.Errorf("background control action %q is unsupported", action)
	}
	return nil
}

// BackgroundTerminalDescriptor is a session-scoped terminal projection.
// TerminalID is daemon-generated and does not expose provider process or
// thread identifiers.
type BackgroundTerminalDescriptor struct {
	TerminalID string `json:"terminal_id"`
	RunID      string `json:"run_id,omitempty"`
	Command    string `json:"command,omitempty"`
	Cwd        string `json:"cwd,omitempty"`
}

// BackgroundStatePayload carries provider-authored background lifecycle state.
// Terminal handles are session-scoped and never stable across sessions.
type BackgroundStatePayload struct {
	Type                     string                         `json:"type"`
	SessionID                string                         `json:"session_id"`
	ActiveRunIDs             []string                       `json:"active_run_ids,omitempty"`
	ActiveTaskIDs            []string                       `json:"active_task_ids,omitempty"`
	ActiveTerminalHandles    []string                       `json:"active_terminal_handles,omitempty"`
	ActiveTerminals          []BackgroundTerminalDescriptor `json:"active_terminals,omitempty"`
	ActiveTerminalCount      int                            `json:"active_terminal_count,omitempty"`
	TerminalObservation      TerminalObservationState       `json:"terminal_observation,omitempty"`
	TerminalObservedAtUnixMS int64                          `json:"terminal_observed_at_unix_ms,omitempty"`
	UpdatedAtUnixMS          int64                          `json:"updated_at_unix_ms,omitempty"`
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
	if err := p.validateTerminalInventory(); err != nil {
		return err
	}
	switch p.TerminalObservation {
	case "", TerminalObservationUnknown, TerminalObservationCurrent, TerminalObservationDegraded:
	default:
		return fmt.Errorf("background state terminal_observation %q is unsupported", p.TerminalObservation)
	}
	if p.TerminalObservedAtUnixMS < 0 {
		return fmt.Errorf("background state terminal_observed_at_unix_ms must be non-negative")
	}
	if p.UpdatedAtUnixMS < 0 {
		return fmt.Errorf("background state updated_at_unix_ms must be non-negative")
	}
	return nil
}

func (p BackgroundStatePayload) validateTerminalInventory() error {
	current := p.TerminalObservation == TerminalObservationCurrent
	seenHandles, err := validateActiveTerminalHandles(p.ActiveTerminalHandles, p.ActiveTerminalCount, current)
	if err != nil {
		return err
	}
	return validateActiveTerminalDescriptors(p.ActiveTerminals, p.ActiveTerminalCount, current, seenHandles)
}

func validateActiveTerminalHandles(handles []string, count int, current bool) (map[string]struct{}, error) {
	if handles != nil && (current || count != 0) && count != len(handles) {
		return nil, fmt.Errorf("background state active_terminal_count must match active_terminal_handles length")
	}
	if current && count > 0 && len(handles) == 0 {
		return nil, fmt.Errorf("background state current terminal inventory requires active_terminal_handles")
	}
	seenHandles := make(map[string]struct{}, len(handles))
	for i, handle := range handles {
		handle = strings.TrimSpace(handle)
		if handle == "" {
			return nil, fmt.Errorf("background state active_terminal_handles[%d] is required", i)
		}
		if _, exists := seenHandles[handle]; exists {
			return nil, fmt.Errorf("background state active_terminal_handles[%d] is duplicated", i)
		}
		seenHandles[handle] = struct{}{}
	}
	return seenHandles, nil
}

func validateActiveTerminalDescriptors(terminals []BackgroundTerminalDescriptor, count int, current bool, activeHandles map[string]struct{}) error {
	if terminals != nil && (current || count != 0) && count != len(terminals) {
		return fmt.Errorf("background state active_terminal_count must match active_terminals length")
	}
	seenTerminalIDs := make(map[string]struct{}, len(terminals))
	for i, terminal := range terminals {
		terminalID := strings.TrimSpace(terminal.TerminalID)
		if terminalID == "" {
			return fmt.Errorf("background state active_terminals[%d].terminal_id is required", i)
		}
		if _, exists := seenTerminalIDs[terminalID]; exists {
			return fmt.Errorf("background state active_terminals[%d].terminal_id is duplicated", i)
		}
		seenTerminalIDs[terminalID] = struct{}{}
		if current {
			if _, exists := activeHandles[terminalID]; !exists {
				return fmt.Errorf("background state active_terminals[%d].terminal_id is not an active_terminal_handle", i)
			}
		}
	}
	return nil
}
