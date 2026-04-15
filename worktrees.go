// Package protocol — git worktrees wire protocol (feature 173).
//
// This file defines the request/response types, streaming progress events,
// broadcast payloads, and classified error codes for the git worktree
// management family (list, add, remove, lock/unlock, prune, session switch)
// introduced in feature 173.
//
// The DAEMON is the source of truth for every Msg* constant in this file.
// A matching constant block lives in agentd/internal/session/wsserver_worktrees.go
// (added in a later phase) with an init() panic that cross-checks equality
// at daemon startup — the explicit drift prevention for the feature 170
// incident class (see specs/170-fix-mcp-wire-drift/spec.md).
package protocol

// Message-type constants — DAEMON-side source of truth. Feature 173.
//
// Mirror MUST exist in agentd/internal/session/wsserver_worktrees.go with an
// init() panic cross-check that the two blocks agree at daemon startup.
const (
	MsgGitWorktreeList               = "git_worktree_list"
	MsgGitWorktreeListResponse       = "git_worktree_list_response"
	MsgGitWorktreeAdd                = "git_worktree_add"
	MsgGitWorktreeAddResponse        = "git_worktree_add_response"
	MsgGitWorktreeAdded              = "git_worktree_added"
	MsgGitWorktreeRemove             = "git_worktree_remove"
	MsgGitWorktreeRemoveResponse     = "git_worktree_remove_response"
	MsgGitWorktreeRemoved            = "git_worktree_removed"
	MsgGitWorktreeLock               = "git_worktree_lock"
	MsgGitWorktreeLockResponse       = "git_worktree_lock_response"
	MsgGitWorktreeUnlock             = "git_worktree_unlock"
	MsgGitWorktreeUnlockResponse     = "git_worktree_unlock_response"
	MsgGitWorktreeUpdated            = "git_worktree_updated"
	MsgGitWorktreePrune              = "git_worktree_prune"
	MsgGitWorktreePruneResponse      = "git_worktree_prune_response"
	MsgGitWorktreeProgress           = "git_worktree_progress"
	MsgSessionWorktreeSwitch         = "session_worktree_switch"
	MsgSessionWorktreeSwitchResponse = "session_worktree_switch_response"
	MsgSessionWorktreeChanged        = "session_worktree_changed"
)

// Classified error codes for worktree operations. The PWA maps each code
// to a user-facing ErrorBanner headline + suggested-action chips. Feature
// 173. These codes extend the GitSyncErr* family from gitsync.go but stay
// scoped to worktree-specific failure modes so the PWA can render tailored
// messages (e.g., "This branch is already checked out in another worktree").
const (
	// GitWorktreeErrPathExists — target directory already exists.
	GitWorktreeErrPathExists = "worktree_path_exists"
	// GitWorktreeErrPathInvalid — validateWorktreePath rejected the path
	// (contains forbidden chars, escapes the allowlist, not absolute, etc).
	GitWorktreeErrPathInvalid = "worktree_path_invalid"
	// GitWorktreeErrBranchCheckedOut — requested base branch is already
	// checked out in another worktree.
	GitWorktreeErrBranchCheckedOut = "worktree_branch_checked_out"
	// GitWorktreeErrInUse — worktree path is the active workDir for one
	// or more live sessions; remove is refused to prevent orphaning them.
	GitWorktreeErrInUse = "worktree_in_use"
	// GitWorktreeErrPrimary — attempted to remove/switch-to the primary
	// worktree; primary cannot be removed and cannot be switched away
	// from via the same code path.
	GitWorktreeErrPrimary = "worktree_primary"
	// GitWorktreeErrDirty — worktree has uncommitted changes and the
	// request did not opt into --force.
	GitWorktreeErrDirty = "worktree_dirty"
	// GitWorktreeErrMissing — requested worktree path does not exist in
	// `git worktree list` output.
	GitWorktreeErrMissing = "worktree_missing"
	// GitWorktreeErrBranchInvalid — the requested new-branch name failed
	// gitops's branch-name validation (contains forbidden chars, starts
	// with `-`, too long, etc).
	GitWorktreeErrBranchInvalid = "worktree_branch_invalid"
	// GitWorktreesDisabled — operator disabled the worktree feature via
	// gitops.worktrees_enabled=false in daemon config.
	GitWorktreesDisabled = "worktrees_disabled"
)

// Worktree is a single entry in `git worktree list --porcelain -z` output.
// Transient per request — the PWA renders a list of these in the
// WorktreePanel. IsPrimary=true marks the main working tree (cannot be
// removed). Feature 173.
type Worktree struct {
	Path           string `json:"path"`
	Branch         string `json:"branch,omitempty"`
	HeadSHA        string `json:"head_sha,omitempty"`
	Locked         bool   `json:"locked,omitempty"`
	LockedReason   string `json:"locked_reason,omitempty"`
	Prunable       bool   `json:"prunable,omitempty"`
	PrunableReason string `json:"prunable_reason,omitempty"`
	IsPrimary      bool   `json:"is_primary,omitempty"`
	Ahead          int    `json:"ahead,omitempty"`
	Behind         int    `json:"behind,omitempty"`
	Upstream       string `json:"upstream,omitempty"`
}

// GitWorktreeListRequest — PWA → daemon. Asks for the worktree list of
// the repo containing the session's workDir.
type GitWorktreeListRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

// GitWorktreeListResponse — daemon → PWA. Echoes RequestID.
type GitWorktreeListResponse struct {
	RequestID string     `json:"request_id"`
	OK        bool       `json:"ok"`
	Error     string     `json:"error,omitempty"`
	ErrorCode string     `json:"error_code,omitempty"`
	Worktrees []Worktree `json:"worktrees,omitempty"`
}

// GitWorktreeAddRequest — PWA → daemon. NewBranchName non-empty triggers
// `-b <new>` so git creates a new branch at the target path. LockImmediately
// triggers a follow-up `git worktree lock` after the add succeeds.
type GitWorktreeAddRequest struct {
	SessionID       string `json:"session_id"`
	RequestID       string `json:"request_id"`
	Path            string `json:"path"`
	BaseRef         string `json:"base_ref"`
	NewBranchName   string `json:"new_branch_name,omitempty"`
	LockImmediately bool   `json:"lock_immediately,omitempty"`
}

// GitWorktreeAddResponse — daemon → PWA (requestor only). On success,
// Worktree is populated with the created row.
type GitWorktreeAddResponse struct {
	RequestID string    `json:"request_id"`
	OK        bool      `json:"ok"`
	Error     string    `json:"error,omitempty"`
	ErrorCode string    `json:"error_code,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	Worktree  *Worktree `json:"worktree,omitempty"`
}

// GitWorktreeAddedPayload — broadcast to all clients attached to the
// session on successful add (dual-fanout local WS + relay).
type GitWorktreeAddedPayload struct {
	SessionID string   `json:"session_id"`
	Worktree  Worktree `json:"worktree"`
}

// GitWorktreeRemoveRequest — PWA → daemon. Force maps to git's --force
// flag; gated by advanced-disclosure ConfirmRail in the UI.
type GitWorktreeRemoveRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
	Force     bool   `json:"force,omitempty"`
}

// GitWorktreeRemoveResponse — daemon → PWA. On worktree_in_use, Users
// lists the session IDs currently attached to the path so the PWA can
// surface "session-abc is attached — detach first".
type GitWorktreeRemoveResponse struct {
	RequestID string   `json:"request_id"`
	OK        bool     `json:"ok"`
	Error     string   `json:"error,omitempty"`
	ErrorCode string   `json:"error_code,omitempty"`
	Stderr    string   `json:"stderr,omitempty"`
	Users     []string `json:"users,omitempty"`
}

// GitWorktreeRemovedPayload — broadcast on successful removal.
type GitWorktreeRemovedPayload struct {
	SessionID string `json:"session_id"`
	Path      string `json:"path"`
}

// GitWorktreeLockRequest — PWA → daemon. Reason is redacted server-side
// before logging and truncated to 512 bytes before being passed to git.
type GitWorktreeLockRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
	Reason    string `json:"reason,omitempty"`
}

// GitWorktreeLockResponse — daemon → PWA.
type GitWorktreeLockResponse struct {
	RequestID string    `json:"request_id"`
	OK        bool      `json:"ok"`
	Error     string    `json:"error,omitempty"`
	ErrorCode string    `json:"error_code,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	Worktree  *Worktree `json:"worktree,omitempty"`
}

// GitWorktreeUnlockRequest — PWA → daemon.
type GitWorktreeUnlockRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
}

// GitWorktreeUnlockResponse — daemon → PWA.
type GitWorktreeUnlockResponse struct {
	RequestID string    `json:"request_id"`
	OK        bool      `json:"ok"`
	Error     string    `json:"error,omitempty"`
	ErrorCode string    `json:"error_code,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	Worktree  *Worktree `json:"worktree,omitempty"`
}

// GitWorktreeUpdatedPayload — broadcast on either lock or unlock success
// (same payload used for any non-structural worktree row update).
type GitWorktreeUpdatedPayload struct {
	SessionID string   `json:"session_id"`
	Worktree  Worktree `json:"worktree"`
}

// GitWorktreePruneRequest — PWA → daemon. DryRun=true runs
// `git worktree prune --dry-run -v` for a preview; the PWA then renders
// a ConfirmRail before sending the non-dry-run follow-up.
type GitWorktreePruneRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

// GitWorktreePruneResponse — daemon → PWA. RemovedPaths lists the
// .git/worktrees/<name> entries that were (or would be) removed.
type GitWorktreePruneResponse struct {
	RequestID    string   `json:"request_id"`
	OK           bool     `json:"ok"`
	Error        string   `json:"error,omitempty"`
	ErrorCode    string   `json:"error_code,omitempty"`
	Stderr       string   `json:"stderr,omitempty"`
	DryRun       bool     `json:"dry_run"`
	RemovedPaths []string `json:"removed_paths,omitempty"`
	RemovedCount int      `json:"removed_count,omitempty"`
}

// GitWorktreeProgressPayload is a streaming daemon → PWA event emitted
// during long-running `git worktree add --progress` operations. Lossy on
// the daemon side (buffered channel with drop-oldest on full). The
// terminal GitWorktreeAddResponse is always delivered.
//
// Op is always "add" in v1 since `git worktree add` is currently the
// only long-running worktree operation; reserved for future extension
// to "switch" or "prune".
type GitWorktreeProgressPayload struct {
	RequestID string `json:"request_id"`
	Op        string `json:"op"`
	Stage     string `json:"stage"`
	Percent   int    `json:"percent,omitempty"`
	Line      string `json:"line,omitempty"`
}

// SessionWorktreeSwitchRequest — PWA → daemon. Path is the target
// worktree's absolute path (must be one returned by GitWorktreeList for
// the session's current repo).
type SessionWorktreeSwitchRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
}

// SessionWorktreeSwitchResponse — daemon → PWA (requestor only). On
// success, NewWorkDir echoes the target path, NewBranch is the branch
// at the target (may be empty for detached HEAD), and RestartedAt is
// the Unix-millis timestamp of the completed subprocess restart.
type SessionWorktreeSwitchResponse struct {
	RequestID   string `json:"request_id"`
	OK          bool   `json:"ok"`
	Error       string `json:"error,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	NewWorkDir  string `json:"new_work_dir,omitempty"`
	NewBranch   string `json:"new_branch,omitempty"`
	RestartedAt int64  `json:"restarted_at,omitempty"`
}

// SessionWorktreeChangedPayload — server-push event broadcast to all
// clients attached to the session on a successful worktree switch
// (dual-fanout local WS + relay).
type SessionWorktreeChangedPayload struct {
	SessionID   string `json:"session_id"`
	OldWorkDir  string `json:"old_work_dir"`
	NewWorkDir  string `json:"new_work_dir"`
	NewBranch   string `json:"new_branch,omitempty"`
	RestartedAt int64  `json:"restarted_at"`
}
