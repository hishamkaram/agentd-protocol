package protocol

import "fmt"

const (
	// MsgCommandReceipt is the legacy AgentMessage type carrying a
	// CommandReceiptPayload from daemon to PWA.
	MsgCommandReceipt = "command_receipt"

	CommandReceiptSchemaVersion = "agentd.command_receipt.v1"
)

type CommandReceiptStage string

const (
	CommandReceiptStageDaemonReceived   CommandReceiptStage = "daemon_received"
	CommandReceiptStageDaemonAccepted   CommandReceiptStage = "daemon_accepted"
	CommandReceiptStageQueued           CommandReceiptStage = "queued"
	CommandReceiptStageDispatched       CommandReceiptStage = "dispatched"
	CommandReceiptStageProviderAccepted CommandReceiptStage = "provider_accepted"
	CommandReceiptStageResponseStarted  CommandReceiptStage = "response_started"
	CommandReceiptStageCompleted        CommandReceiptStage = "completed"
	CommandReceiptStageFailed           CommandReceiptStage = "failed"
)

type CommandReceiptReasonCode string

const (
	CommandReceiptReasonTransportSendFailed    CommandReceiptReasonCode = "transport_send_failed"
	CommandReceiptReasonRelayRouteFailed       CommandReceiptReasonCode = "relay_route_failed"
	CommandReceiptReasonInvalidPayload         CommandReceiptReasonCode = "invalid_payload"
	CommandReceiptReasonSessionNotFound        CommandReceiptReasonCode = "session_not_found"
	CommandReceiptReasonSessionNotActive       CommandReceiptReasonCode = "session_not_active"
	CommandReceiptReasonQueueFull              CommandReceiptReasonCode = "queue_full"
	CommandReceiptReasonProviderNotReady       CommandReceiptReasonCode = "provider_not_ready"
	CommandReceiptReasonProviderSendFailed     CommandReceiptReasonCode = "provider_send_failed"
	CommandReceiptReasonHandlerFailed          CommandReceiptReasonCode = "handler_failed"
	CommandReceiptReasonTimeoutNoDaemonReceipt CommandReceiptReasonCode = "timeout_no_daemon_receipt"
)

// CommandReceiptPayload is the redacted daemon-to-PWA receipt for a
// PWA-originated command. It intentionally contains only stable identifiers,
// low-cardinality state, and redacted failure classification.
type CommandReceiptPayload struct {
	SchemaVersion    string                   `json:"schema_version"`
	ReceiptID        string                   `json:"receipt_id"`
	CommandID        string                   `json:"command_id"`
	CommandType      string                   `json:"command_type"`
	SessionID        string                   `json:"session_id,omitempty"`
	TraceID          string                   `json:"trace_id,omitempty"`
	Stage            CommandReceiptStage      `json:"stage"`
	Terminal         bool                     `json:"terminal"`
	Retryable        bool                     `json:"retryable"`
	ReasonCode       CommandReceiptReasonCode `json:"reason_code,omitempty"`
	ObservedAtUnixMS int64                    `json:"observed_at_unix_ms"`
}

// ValidateCommandReceipt verifies the non-sensitive receipt invariants shared
// across daemon, PWA fixtures, and production synthetic checks.
func ValidateCommandReceipt(r CommandReceiptPayload) error {
	if r.SchemaVersion != CommandReceiptSchemaVersion {
		return fmt.Errorf("protocol.ValidateCommandReceipt: schema_version = %q, want %q", r.SchemaVersion, CommandReceiptSchemaVersion)
	}
	if r.ReceiptID == "" {
		return fmt.Errorf("protocol.ValidateCommandReceipt: receipt_id is required")
	}
	if r.CommandID == "" {
		return fmt.Errorf("protocol.ValidateCommandReceipt: command_id is required")
	}
	if r.CommandType == "" {
		return fmt.Errorf("protocol.ValidateCommandReceipt: command_type is required")
	}
	if !IsKnownCommandReceiptStage(r.Stage) {
		return fmt.Errorf("protocol.ValidateCommandReceipt: unknown stage %q", r.Stage)
	}
	if r.ReasonCode != "" && !IsKnownCommandReceiptReasonCode(r.ReasonCode) {
		return fmt.Errorf("protocol.ValidateCommandReceipt: unknown reason_code %q", r.ReasonCode)
	}
	if r.Stage == CommandReceiptStageFailed && !r.Terminal {
		return fmt.Errorf("protocol.ValidateCommandReceipt: failed receipts must be terminal")
	}
	if r.Terminal && r.Stage != CommandReceiptStageCompleted && r.Stage != CommandReceiptStageFailed {
		return fmt.Errorf("protocol.ValidateCommandReceipt: terminal stage %q must be completed or failed", r.Stage)
	}
	if r.ObservedAtUnixMS <= 0 {
		return fmt.Errorf("protocol.ValidateCommandReceipt: observed_at_unix_ms must be positive")
	}
	return nil
}

func IsKnownCommandReceiptStage(stage CommandReceiptStage) bool {
	switch stage {
	case CommandReceiptStageDaemonReceived,
		CommandReceiptStageDaemonAccepted,
		CommandReceiptStageQueued,
		CommandReceiptStageDispatched,
		CommandReceiptStageProviderAccepted,
		CommandReceiptStageResponseStarted,
		CommandReceiptStageCompleted,
		CommandReceiptStageFailed:
		return true
	default:
		return false
	}
}

func IsKnownCommandReceiptReasonCode(code CommandReceiptReasonCode) bool {
	switch code {
	case CommandReceiptReasonTransportSendFailed,
		CommandReceiptReasonRelayRouteFailed,
		CommandReceiptReasonInvalidPayload,
		CommandReceiptReasonSessionNotFound,
		CommandReceiptReasonSessionNotActive,
		CommandReceiptReasonQueueFull,
		CommandReceiptReasonProviderNotReady,
		CommandReceiptReasonProviderSendFailed,
		CommandReceiptReasonHandlerFailed,
		CommandReceiptReasonTimeoutNoDaemonReceipt:
		return true
	default:
		return false
	}
}
