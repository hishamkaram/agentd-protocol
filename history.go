// Package protocol — history replay sync wire protocol (feature 192).
//
// This file defines the daemon→PWA `history_replay_complete` sentinel sent
// once per session after the daemon finishes replaying that session's
// `s.history` entries. The PWA uses the sentinel to dismiss the loading
// skeleton, which closes the empty-session hang reported in feature 192.
//
// The DAEMON is the source of truth for the Msg* constant in this file.
// A matching constant lives in agentd/internal/session/wsserver_history.go
// with an init() panic that cross-checks equality at daemon startup — the
// explicit drift prevention for the feature 169→170 wire-type-drift incident
// class. Same pattern as agentd-protocol/gitsync.go ↔
// agentd/internal/session/wsserver_gitsync.go:46-75.
package protocol

// Message-type constant — DAEMON-side source of truth.
// Mirror MUST exist in agentd/internal/session/wsserver_history.go with an
// init() panic cross-check that the two values agree at daemon startup.
const (
	MsgHistoryReplayComplete = "history_replay_complete"
)

// HistoryReplayCompletePayload is the daemon→PWA push emitted exactly once
// per session after the daemon's history-replay loop finishes draining
// `s.history` to that connection. It carries no replay content — only the
// session identifier — so the PWA can flip the per-session
// `historyReplayedAt[sid]` flag and dismiss the loading skeleton.
//
// Ordering invariant (FR-003): the sentinel MUST arrive AFTER all replayed
// `s.history` entries on the same WS connection. The daemon emits inside
// the same `s.mu.Lock()` window as the replay loop to preserve order.
//
// SessionID is required and non-empty in production (the daemon writes from
// the map key in `s.history`, which is non-empty by construction). The
// payload type itself round-trips an empty value defensively — validation
// is the consumer handler's responsibility, not the wire type's.
type HistoryReplayCompletePayload struct {
	SessionID string `json:"session_id"`
}
