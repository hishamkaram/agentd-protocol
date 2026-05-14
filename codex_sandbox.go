package protocol

// CodexSandboxMode is the cross-repo wire enum for Codex per-session sandbox
// startup controls. Values intentionally match the Codex SDK thread/start
// sandbox literals.
type CodexSandboxMode string

const (
	CodexSandboxReadOnly         CodexSandboxMode = "read-only"
	CodexSandboxWorkspaceWrite   CodexSandboxMode = "workspace-write"
	CodexSandboxDangerFullAccess CodexSandboxMode = "danger-full-access"
)

// KnownCodexSandboxModes returns the daemon/PWA-supported sandbox modes in UI
// order from safest to most permissive.
func KnownCodexSandboxModes() []CodexSandboxMode {
	return []CodexSandboxMode{
		CodexSandboxReadOnly,
		CodexSandboxWorkspaceWrite,
		CodexSandboxDangerFullAccess,
	}
}

// IsKnownCodexSandboxMode reports whether mode is a valid wire literal.
func IsKnownCodexSandboxMode(mode CodexSandboxMode) bool {
	switch mode {
	case CodexSandboxReadOnly, CodexSandboxWorkspaceWrite, CodexSandboxDangerFullAccess:
		return true
	default:
		return false
	}
}
