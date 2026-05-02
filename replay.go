// Package protocol - durable session replay recovery wire protocol.
package protocol

// Message-type constants for daemon-owned control-journal recovery.
const (
	// MsgReplayRequest is a PWA->daemon request for daemon messages whose
	// inner AgentMessage.seq is greater than AfterSeq for the named session.
	MsgReplayRequest = "replay_request"

	// MsgReplayComplete is the daemon->PWA sentinel emitted after a requested
	// replay range has been drained.
	MsgReplayComplete = "replay_complete"
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
