// Package protocol — git sync wire protocol (feature 172).
//
// This file defines the request/response types, streaming progress events,
// and classified error codes for the git sync action family (branch list,
// branch switch, fetch, pull, push, cancel) introduced in feature 172.
//
// The DAEMON is the source of truth for every Msg* constant in this file.
// A matching constant block lives in agentd/internal/session/wsserver_gitsync.go
// with an init() panic that cross-checks equality at daemon startup — the
// explicit drift prevention for the feature 170 incident class.
package protocol

// Message-type constants — DAEMON-side source of truth.
// Mirror MUST exist in agentd/internal/session/wsserver_gitsync.go with an
// init() panic cross-check that the two blocks agree at daemon startup.
const (
	MsgGitBranchList           = "git_branch_list"
	MsgGitBranchListResponse   = "git_branch_list_response"
	MsgGitBranchSwitch         = "git_branch_switch"
	MsgGitBranchSwitchResponse = "git_branch_switch_response"
	MsgGitFetch                = "git_fetch"
	MsgGitFetchResponse        = "git_fetch_response"
	MsgGitPull                 = "git_pull"
	MsgGitPullResponse         = "git_pull_response"
	MsgGitPush                 = "git_push"
	MsgGitPushResponse         = "git_push_response"
	MsgGitSyncProgress         = "git_sync_progress"
	MsgGitSyncCancel           = "git_sync_cancel"
	MsgGitSyncCancelResponse   = "git_sync_cancel_response"
)

// Classified error codes. Every GitSync*Response.ErrorCode field holds one
// of these values (or the empty string when OK is true). The PWA maps each
// code to a user-facing ErrorBanner headline + suggested-action chips.
const (
	GitSyncErrAuthFailed     = "auth_failed"
	GitSyncErrSSHPromptHang  = "ssh_prompt_hang"
	GitSyncErrNonFastForward = "non_fast_forward"
	GitSyncErrMergeConflict  = "merge_conflict"
	GitSyncErrDirtyWorkTree  = "dirty_work_tree"
	GitSyncErrNoUpstream     = "no_upstream"
	GitSyncErrNotAGitRepo    = "not_a_git_repo"
	GitSyncErrNetwork        = "network"
	GitSyncErrCanceled       = "canceled"
	GitSyncErrTimeout        = "timeout"
	GitSyncErrLockedIndex    = "locked_index"
	GitSyncErrInternal       = "internal"
)

// GitBranch is a single row in the branch switcher. Transient per request.
// Sorted by LastCommitAt descending for MRU behavior in the UI.
type GitBranch struct {
	Name          string `json:"name"`
	IsCurrent     bool   `json:"is_current"`
	Upstream      string `json:"upstream,omitempty"`
	Ahead         int    `json:"ahead"`
	Behind        int    `json:"behind"`
	LastCommitAt  int64  `json:"last_commit_at"`
	LastCommitSha string `json:"last_commit_sha,omitempty"`
}

// GitBranchListRequest — PWA → daemon.
type GitBranchListRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

// GitBranchListResponse — daemon → PWA. Echoes RequestID.
type GitBranchListResponse struct {
	RequestID     string      `json:"request_id"`
	OK            bool        `json:"ok"`
	Error         string      `json:"error,omitempty"`
	ErrorCode     string      `json:"error_code,omitempty"`
	CurrentBranch string      `json:"current_branch,omitempty"`
	Branches      []GitBranch `json:"branches,omitempty"`
}

// GitBranchSwitchRequest — PWA → daemon. First request always has
// StashDirty=false; on DirtyRequired=true response, the PWA opens the
// ConfirmRail and retries with StashDirty=true.
type GitBranchSwitchRequest struct {
	SessionID  string `json:"session_id"`
	RequestID  string `json:"request_id"`
	Branch     string `json:"branch"`
	StashDirty bool   `json:"stash_dirty"`
}

// GitBranchSwitchResponse — daemon → PWA.
type GitBranchSwitchResponse struct {
	RequestID     string `json:"request_id"`
	OK            bool   `json:"ok"`
	Error         string `json:"error,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	Stderr        string `json:"stderr,omitempty"`
	NewBranch     string `json:"new_branch,omitempty"`
	StashRef      string `json:"stash_ref,omitempty"`
	DirtyRequired bool   `json:"dirty_required,omitempty"`
}

// GitFetchRequest — PWA → daemon. Empty Remote triggers git-native default
// resolution (upstream-tracking → origin → fail). See FR-003a in spec.md.
type GitFetchRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Remote    string `json:"remote,omitempty"`
	Prune     bool   `json:"prune,omitempty"`
}

// GitFetchResponse — daemon → PWA.
type GitFetchResponse struct {
	RequestID string `json:"request_id"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	FetchedAt int64  `json:"fetched_at"`
}

// GitPullRequest — PWA → daemon. Rebase=false (default) → --ff-only.
type GitPullRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Rebase    bool   `json:"rebase,omitempty"`
}

// GitPullResponse — daemon → PWA.
type GitPullResponse struct {
	RequestID    string `json:"request_id"`
	OK           bool   `json:"ok"`
	Error        string `json:"error,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	Stderr       string `json:"stderr,omitempty"`
	NewHead      string `json:"new_head,omitempty"`
	FilesChanged int    `json:"files_changed,omitempty"`
}

// GitPushRequest — PWA → daemon. Force semantics:
//
//	ForceWithLease=true, Force=false → --force-with-lease (safe)
//	Force=true, ForceWithLease=false → raw --force (unsafe, advanced UI)
//	both false → plain push
//	both true → daemon rejects with GitSyncErrInternal (ambiguous)
type GitPushRequest struct {
	SessionID      string `json:"session_id"`
	RequestID      string `json:"request_id"`
	Remote         string `json:"remote,omitempty"`
	ForceWithLease bool   `json:"force_with_lease,omitempty"`
	Force          bool   `json:"force,omitempty"`
	SetUpstream    bool   `json:"set_upstream,omitempty"`
}

// GitPushResponse — daemon → PWA.
type GitPushResponse struct {
	RequestID string `json:"request_id"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	PushedRef string `json:"pushed_ref,omitempty"`
}

// GitSyncProgressPayload is a streaming daemon → PWA event emitted during
// long-running fetch/pull/push/switch operations. Lossy on the daemon side
// (buffered channel with drop-oldest on full). The terminal response is
// always delivered.
type GitSyncProgressPayload struct {
	RequestID string `json:"request_id"`
	Op        string `json:"op"`
	Stage     string `json:"stage"`
	Percent   int    `json:"percent,omitempty"`
	Line      string `json:"line,omitempty"`
}

// GitSyncCancelRequest — PWA → daemon.
type GitSyncCancelRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	TargetID  string `json:"target_id"`
}

// GitSyncCancelResponse — daemon → PWA. OK=true means the target's
// context.CancelFunc was found and invoked; the target op will emit its
// own terminal response shortly after with ErrorCode=canceled.
type GitSyncCancelResponse struct {
	RequestID string `json:"request_id"`
	TargetID  string `json:"target_id"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}
