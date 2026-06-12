package protocol

// SessionRecoveryReason is the daemon-classified reason a session can be
// repaired without losing its AgentD session identity or transcript.
type SessionRecoveryReason string

const (
	SessionRecoveryReasonDaemonRestarted     SessionRecoveryReason = "daemon_restarted"
	SessionRecoveryReasonProviderExited      SessionRecoveryReason = "provider_exited"
	SessionRecoveryReasonAuthRequired        SessionRecoveryReason = "auth_required"
	SessionRecoveryReasonProviderStartFailed SessionRecoveryReason = "provider_start_failed"
	SessionRecoveryReasonWorkdirMissing      SessionRecoveryReason = "workdir_missing"
)

// SessionRecoveryAction is the user-visible repair action for a recoverable
// paused session.
type SessionRecoveryAction string

const (
	SessionRecoveryActionResume         SessionRecoveryAction = "resume"
	SessionRecoveryActionRetry          SessionRecoveryAction = "retry"
	SessionRecoveryActionLoginThenRetry SessionRecoveryAction = "login_then_retry"
	SessionRecoveryActionRestorePath    SessionRecoveryAction = "restore_path"
)

// SessionRecoveryInfo is the structured recovery metadata on SessionInfo and
// StatusPayload.
type SessionRecoveryInfo struct {
	Recoverable bool                  `json:"recoverable"`
	Reason      SessionRecoveryReason `json:"reason,omitempty"`
	Action      SessionRecoveryAction `json:"action,omitempty"`
	UserMessage string                `json:"user_message,omitempty"`
	Provider    string                `json:"provider,omitempty"`
}

// StatusPayload pins the cross-repo additive recovery field for daemon status
// updates. Repos with richer status payloads embed the same JSON field in their
// local StatusPayload type.
type StatusPayload struct {
	Recovery *SessionRecoveryInfo `json:"recovery,omitempty"`
}
