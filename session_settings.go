package protocol

const SessionRuntimeSettingsCapabilitySchemaVersion = 1

const (
	MsgSetSessionSettings       = "set_session_settings"
	MsgSetSessionSettingsResult = "set_session_settings_result"
)

type SessionRuntimeSettingID string

const (
	SessionRuntimeSettingModel              SessionRuntimeSettingID = "model"
	SessionRuntimeSettingEffort             SessionRuntimeSettingID = "effort"
	SessionRuntimeSettingApprovalMode       SessionRuntimeSettingID = "approval_mode"
	SessionRuntimeSettingSandboxMode        SessionRuntimeSettingID = "sandbox_mode"
	SessionRuntimeSettingWorkflowAutomation SessionRuntimeSettingID = "workflow_automation"
)

type SessionRuntimeSettingSource string

const (
	SessionRuntimeSettingSourceProviderReadback SessionRuntimeSettingSource = "provider_readback"
	SessionRuntimeSettingSourceDaemonRegistry   SessionRuntimeSettingSource = "daemon_registry"
	SessionRuntimeSettingSourceModelMetadata    SessionRuntimeSettingSource = "model_metadata"
	SessionRuntimeSettingSourceUnsupported      SessionRuntimeSettingSource = "unsupported"
)

// SessionSettingsPayload carries applied scalar runtime settings. It
// intentionally excludes raw provider settings, workflow scripts, local paths,
// and prompt text.
type SessionSettingsPayload struct {
	Model              string                      `json:"model,omitempty"`
	Effort             string                      `json:"effort,omitempty"`
	ApprovalMode       string                      `json:"approval_mode,omitempty"`
	SandboxMode        string                      `json:"sandbox_mode,omitempty"`
	WorkflowAutomation *bool                       `json:"workflow_automation,omitempty"`
	Source             SessionRuntimeSettingSource `json:"source,omitempty"`
}

// SessionRuntimeSettingCapability describes one runtime setting in a
// provider-neutral shape for UI enablement.
type SessionRuntimeSettingCapability struct {
	ID              SessionRuntimeSettingID     `json:"id"`
	Supported       bool                        `json:"supported"`
	SupportedValues []string                    `json:"supported_values,omitempty"`
	CurrentValue    string                      `json:"current_value,omitempty"`
	CurrentBool     *bool                       `json:"current_bool,omitempty"`
	Source          SessionRuntimeSettingSource `json:"source,omitempty"`
	DisabledReason  string                      `json:"disabled_reason,omitempty"`
}

type SessionRuntimeSettingsCapability struct {
	SchemaVersion int                               `json:"schema_version"`
	Settings      []SessionRuntimeSettingCapability `json:"settings"`
}

// SetSessionSettingsRequest is sent by a client to change live session
// runtime settings. Omitted fields are left unchanged.
type SetSessionSettingsRequest struct {
	Type               string `json:"type"`
	RequestID          string `json:"request_id"`
	SessionID          string `json:"session_id"`
	Model              string `json:"model,omitempty"`
	Effort             string `json:"effort,omitempty"`
	ApprovalMode       string `json:"approval_mode,omitempty"`
	SandboxMode        string `json:"sandbox_mode,omitempty"`
	WorkflowAutomation *bool  `json:"workflow_automation,omitempty"`
}

type SetSessionSettingsResult struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id"`
	SessionID string                 `json:"session_id"`
	Success   bool                   `json:"success"`
	Settings  SessionSettingsPayload `json:"settings,omitempty"`
	Error     string                 `json:"error,omitempty"`
}
