// Package protocol - selected-session transcript history paging wire protocol.
package protocol

import "encoding/json"

// Message-type constants owned by agentd-protocol for transcript history recovery.
const (
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

type HistoryErrorCode string

const (
	HistoryErrorSessionNotFound HistoryErrorCode = "session_not_found"
	HistoryErrorUnavailable     HistoryErrorCode = "history_unavailable"
	HistoryErrorBusy            HistoryErrorCode = "history_busy"
	HistoryErrorInvalidRequest  HistoryErrorCode = "invalid_request"
)

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
	ErrorCode         HistoryErrorCode  `json:"error_code,omitempty"`
	Trimmed           bool              `json:"trimmed,omitempty"`
	RetainedOldestSeq uint64            `json:"retained_oldest_seq,omitempty"`
}
