package protocol

// GitFileStatus describes one dirty file reported by `git status --porcelain=v2`.
// Shipped inside GitStatusPayload.Files. Paths are repo-root-relative and never
// absolute.
type GitFileStatus struct {
	Path        string `json:"path"`
	OrigPath    string `json:"orig_path,omitempty"`
	XY          string `json:"xy"`
	IsStaged    bool   `json:"is_staged"`
	IsUntracked bool   `json:"is_untracked"`
	IsConflict  bool   `json:"is_conflict"`
	IsRename    bool   `json:"is_rename"`
	Insertions  int    `json:"insertions"`
	Deletions   int    `json:"deletions"`
	IsBinary    bool   `json:"is_binary"`
	IsRedacted  bool   `json:"is_redacted"`
	SizeBytes   int64  `json:"size_bytes"`
}

// GitStatusPayload is the full per-session status snapshot returned by
// GetGitStatus and carried in MsgGitStatusUpdate pushes. Files is sorted
// ascending by path and capped at 500 entries.
type GitStatusPayload struct {
	RepoRoot        string          `json:"repo_root"`
	Files           []GitFileStatus `json:"files"`
	TotalInsertions int             `json:"total_insertions"`
	TotalDeletions  int             `json:"total_deletions"`
	GeneratedAt     int64           `json:"generated_at"`
}

// GitStatusRequest is the PWA → daemon request to fetch the current status
// snapshot for a session. The response body is a GitStatusPayload.
type GitStatusRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

// GitDiffRequest is the PWA → daemon request for one file's unified diff.
// Path must be repo-root-relative and must not escape the repo root.
type GitDiffRequest struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
}

// GitDiffResponse is the daemon → PWA response carrying one file's diff
// (or a safe placeholder state). DiffText is raw unified diff and is empty
// when IsBinary, IsRedacted, or Status is "removed".
type GitDiffResponse struct {
	RequestID   string `json:"request_id"`
	Path        string `json:"path"`
	Status      string `json:"status"`
	IsBinary    bool   `json:"is_binary"`
	IsRedacted  bool   `json:"is_redacted"`
	IsTruncated bool   `json:"is_truncated"`
	DiffText    string `json:"diff_text"`
	SizeBytes   int64  `json:"size_bytes"`
	Insertions  int    `json:"insertions"`
	Deletions   int    `json:"deletions"`
}

// GitNotAvailablePayload is a one-shot notice emitted when gitwatch cannot
// observe a session (WorkDir is not a git repo, or git binary missing).
type GitNotAvailablePayload struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

// GitDiffStatus enum values for GitDiffResponse.Status.
const (
	GitDiffStatusModified string = "modified"
	GitDiffStatusAdded    string = "added"
	GitDiffStatusDeleted  string = "deleted"
	GitDiffStatusRenamed  string = "renamed"
	GitDiffStatusConflict string = "conflict"
	GitDiffStatusRemoved  string = "removed"
)

// GitNotAvailableReason enum values for GitNotAvailablePayload.Reason.
const (
	GitNotAvailableReasonNotARepo       string = "not_a_git_repo"
	GitNotAvailableReasonGitMissing     string = "git_binary_missing"
	GitNotAvailableReasonWorkDirInvalid string = "work_dir_invalid"
)
