// Package protocol - durable session replay recovery wire protocol.
package protocol

import "encoding/json"

// Message-type constants for daemon-owned control-journal recovery.
const (
	// MsgReplayRequest is a PWA->daemon request for daemon messages whose
	// inner AgentMessage.seq is greater than AfterSeq for the named session.
	MsgReplayRequest = "replay_request"

	// MsgReplayComplete is the daemon->PWA sentinel emitted after a requested
	// replay range has been drained.
	MsgReplayComplete = "replay_complete"

	// MsgSessionSnapshotRequest is a PWA->daemon request for the daemon-owned
	// read model of a route-bound active session.
	MsgSessionSnapshotRequest = "session_snapshot_request"

	// MsgSessionSnapshot is the daemon->PWA response containing active-session
	// transcript and pending state. Historical pending entries are state only.
	MsgSessionSnapshot = "session_snapshot"

	// MsgSessionHead is a daemon->PWA announcement of the current durable
	// per-session head sequence.
	MsgSessionHead = "session_head"

	// MsgHistoryPageRequest is a PWA->daemon request for older durable
	// transcript messages before a sequence boundary.
	MsgHistoryPageRequest = "history_page_request"

	// MsgHistoryPage is the daemon->PWA response containing a backward history
	// page from durable storage.
	MsgHistoryPage = "history_page"
)

// ReplayRequest is sent by the PWA when it detects a gap in the inner
// per-session AgentMessage sequence.
type ReplayRequest struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	AfterSeq  uint64 `json:"after_seq"`
}

// SessionHeadPayload is carried in an AgentMessage with Type MsgSessionHead.
type SessionHeadPayload struct {
	SessionID         string `json:"session_id"`
	HeadSeq           uint64 `json:"head_seq"`
	RetainedOldestSeq uint64 `json:"retained_oldest_seq,omitempty"`
}

// HistoryPageRequest is sent by the PWA when the user scrolls back before the
// currently loaded transcript window.
type HistoryPageRequest struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	BeforeSeq uint64 `json:"before_seq"`
	Limit     int    `json:"limit,omitempty"`
	RequestID string `json:"request_id"`
}

// HistoryPagePayload is carried in an AgentMessage with Type MsgHistoryPage.
type HistoryPagePayload struct {
	SessionID         string            `json:"session_id"`
	Messages          []json.RawMessage `json:"messages"`
	OldestSeq         uint64            `json:"oldest_seq"`
	HasMore           bool              `json:"has_more"`
	RequestID         string            `json:"request_id"`
	Error             string            `json:"error,omitempty"`
	Trimmed           bool              `json:"trimmed,omitempty"`
	RetainedOldestSeq uint64            `json:"retained_oldest_seq,omitempty"`
}

// ReplayCompletePayload is carried in an AgentMessage with Type
// MsgReplayComplete after the daemon finishes a requested replay.
type ReplayCompletePayload struct {
	SessionID         string `json:"session_id"`
	FromSeq           uint64 `json:"from_seq"`
	ToSeq             uint64 `json:"to_seq"`
	HeadSeq           uint64 `json:"head_seq,omitempty"`
	Count             int    `json:"count"`
	Error             string `json:"error,omitempty"`
	ErrorCode         string `json:"error_code,omitempty"`
	RetryAfterMS      int    `json:"retry_after_ms,omitempty"`
	Trimmed           bool   `json:"trimmed,omitempty"`
	TrimFloor         uint64 `json:"trim_floor,omitempty"`
	RetainedOldestSeq uint64 `json:"retained_oldest_seq,omitempty"`
}

// SessionSnapshotRequest is sent by the PWA when a route-bound active session
// should be hydrated from daemon-authoritative state before broad replay.
type SessionSnapshotRequest struct {
	Type        string `json:"type"`
	SessionID   string `json:"session_id"`
	Reason      string `json:"reason,omitempty"`
	AfterCursor string `json:"after_cursor,omitempty"`
}

// SessionSnapshotPayload is carried in a daemon->PWA AgentMessage with Type
// MsgSessionSnapshot. Message, session, approval, and prompt entries are raw
// JSON here because their full structs are daemon/PWA-owned; this package pins
// the cross-repo envelope and state semantics.
type SessionSnapshotPayload struct {
	SessionID         string            `json:"session_id"`
	Session           json.RawMessage   `json:"session,omitempty"`
	Messages          []json.RawMessage `json:"messages"`
	PendingApprovals  []json.RawMessage `json:"pending_approvals"`
	PendingPrompts    []json.RawMessage `json:"pending_prompts"`
	HeadSeq           uint64            `json:"head_seq,omitempty"`
	RetainedOldestSeq uint64            `json:"retained_oldest_seq,omitempty"`
	Cursor            string            `json:"cursor,omitempty"`
	Sequence          uint64            `json:"sequence,omitempty"`
	Complete          bool              `json:"complete"`
	Error             string            `json:"error,omitempty"`
}
