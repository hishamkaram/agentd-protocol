package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	// MsgSessionFeatureStatus is the AgentMessage type for per-session optional
	// feature lifecycle updates.
	MsgSessionFeatureStatus = "session_feature_status"

	SessionFeatureStatusSchemaVersion = 1
)

type SessionFeature string

const (
	SessionFeatureGitWatch         SessionFeature = "git.watch"
	SessionFeatureMCPSync          SessionFeature = "mcp.sync"
	SessionFeatureBudgetVirtualKey SessionFeature = "budget.virtual_key"
	SessionFeatureCodexModels      SessionFeature = "codex.models"
	SessionFeatureCodexSkills      SessionFeature = "codex.skills"
)

type SessionFeatureState string

const (
	SessionFeatureStatePending  SessionFeatureState = "pending"
	SessionFeatureStateStarting SessionFeatureState = "starting"
	SessionFeatureStateReady    SessionFeatureState = "ready"
	SessionFeatureStateDegraded SessionFeatureState = "degraded"
	SessionFeatureStateFailed   SessionFeatureState = "failed"
	SessionFeatureStateCanceled SessionFeatureState = "canceled"
)

type SessionFeatureReasonCode string

const (
	SessionFeatureReasonQueueFull        SessionFeatureReasonCode = "queue_full"
	SessionFeatureReasonTimeout          SessionFeatureReasonCode = "timeout"
	SessionFeatureReasonHandlerFailed    SessionFeatureReasonCode = "handler_failed"
	SessionFeatureReasonCanceled         SessionFeatureReasonCode = "canceled"
	SessionFeatureReasonNotAGitRepo      SessionFeatureReasonCode = "not_a_git_repo"
	SessionFeatureReasonGitBinaryMissing SessionFeatureReasonCode = "git_binary_missing"
	SessionFeatureReasonWorkDirInvalid   SessionFeatureReasonCode = "work_dir_invalid"
	SessionFeatureReasonMCPStoreFailed   SessionFeatureReasonCode = "mcp_store_failed"
	SessionFeatureReasonMCPApplyFailed   SessionFeatureReasonCode = "mcp_apply_failed"
	SessionFeatureReasonBudgetKeyFailed  SessionFeatureReasonCode = "budget_key_failed"
)

// SessionFeatureStatusPayload is carried by MsgSessionFeatureStatus and
// EventSessionFeatureStatus. agentd-protocol owns this wire shape; daemon
// producers use it to report optional startup/runtime feature lifecycle
// independently from provider startup.
type SessionFeatureStatusPayload struct {
	SchemaVersion    int                      `json:"schema_version"`
	SessionID        string                   `json:"session_id"`
	Feature          SessionFeature           `json:"feature"`
	State            SessionFeatureState      `json:"state"`
	ReasonCode       SessionFeatureReasonCode `json:"reason_code,omitempty"`
	Message          string                   `json:"message,omitempty"`
	ObservedAtUnixMS int64                    `json:"observed_at_unix_ms"`
	Attempt          int                      `json:"attempt,omitempty"`
}

func (p SessionFeatureStatusPayload) Validate() error {
	if p.SchemaVersion != SessionFeatureStatusSchemaVersion {
		return fmt.Errorf("protocol.SessionFeatureStatusPayload: schema_version must be %d", SessionFeatureStatusSchemaVersion)
	}
	if p.SessionID == "" {
		return errors.New("protocol.SessionFeatureStatusPayload: session_id is required")
	}
	if p.Feature == "" {
		return errors.New("protocol.SessionFeatureStatusPayload: feature is required")
	}
	if p.State == "" {
		return errors.New("protocol.SessionFeatureStatusPayload: state is required")
	}
	if !IsKnownSessionFeatureState(p.State) {
		return fmt.Errorf("protocol.SessionFeatureStatusPayload: unknown state %q", p.State)
	}
	if p.ObservedAtUnixMS <= 0 {
		return errors.New("protocol.SessionFeatureStatusPayload: observed_at_unix_ms is required")
	}
	if p.ReasonCode != "" && !IsKnownSessionFeatureReasonCode(p.ReasonCode) {
		return fmt.Errorf("protocol.SessionFeatureStatusPayload: unknown reason_code %q", p.ReasonCode)
	}
	return nil
}

func (p *SessionFeatureStatusPayload) UnmarshalJSON(data []byte) error {
	type alias SessionFeatureStatusPayload
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	payload := SessionFeatureStatusPayload(out)
	if err := payload.Validate(); err != nil {
		return err
	}
	*p = payload
	return nil
}

func IsKnownSessionFeatureState(state SessionFeatureState) bool {
	switch state {
	case SessionFeatureStatePending,
		SessionFeatureStateStarting,
		SessionFeatureStateReady,
		SessionFeatureStateDegraded,
		SessionFeatureStateFailed,
		SessionFeatureStateCanceled:
		return true
	default:
		return false
	}
}

func IsKnownSessionFeatureReasonCode(code SessionFeatureReasonCode) bool {
	switch code {
	case SessionFeatureReasonQueueFull,
		SessionFeatureReasonTimeout,
		SessionFeatureReasonHandlerFailed,
		SessionFeatureReasonCanceled,
		SessionFeatureReasonNotAGitRepo,
		SessionFeatureReasonGitBinaryMissing,
		SessionFeatureReasonWorkDirInvalid,
		SessionFeatureReasonMCPStoreFailed,
		SessionFeatureReasonMCPApplyFailed,
		SessionFeatureReasonBudgetKeyFailed:
		return true
	default:
		return false
	}
}

// SessionInfo pins the cross-repo feature_statuses field. Repos with richer
// session summaries embed the same JSON field in their local SessionInfo type.
type SessionInfo struct {
	FeatureStatuses  []SessionFeatureStatusPayload `json:"feature_statuses,omitempty"`
	ProviderContract *ProviderCapabilityContract   `json:"provider_contract,omitempty"`
	Recovery         *SessionRecoveryInfo          `json:"recovery,omitempty"`
}
