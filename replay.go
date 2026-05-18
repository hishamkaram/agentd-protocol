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
)

// ReplayRequest is sent by the PWA when it detects a gap in the inner
// per-session AgentMessage sequence.
type ReplayRequest struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	AfterSeq  uint64 `json:"after_seq"`
}

// ReplayCompletePayload is carried in an AgentMessage with Type
// MsgReplayComplete after the daemon finishes a requested replay.
type ReplayCompletePayload struct {
	SessionID string `json:"session_id"`
	FromSeq   uint64 `json:"from_seq"`
	ToSeq     uint64 `json:"to_seq"`
	Count     int    `json:"count"`
	Error     string `json:"error,omitempty"`
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
	SessionID        string            `json:"session_id"`
	Session          json.RawMessage   `json:"session,omitempty"`
	Messages         []json.RawMessage `json:"messages"`
	PendingApprovals []json.RawMessage `json:"pending_approvals"`
	PendingPrompts   []json.RawMessage `json:"pending_prompts"`
	Cursor           string            `json:"cursor,omitempty"`
	Sequence         uint64            `json:"sequence,omitempty"`
	Complete         bool              `json:"complete"`
	Error            string            `json:"error,omitempty"`
}
