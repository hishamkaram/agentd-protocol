package protocol

import (
	"encoding/json"
	"time"
)

// ControlType discriminates control messages in the relay protocol.
type ControlType string

// Control message types for the relay protocol.
const (
	CtrlRegister            ControlType = "register"
	CtrlJoin                ControlType = "join"
	CtrlHeartbeat           ControlType = "heartbeat" // reserved for future application-level keepalive; WebSocket ping/pong handles heartbeats
	CtrlAck                 ControlType = "ack"
	CtrlError               ControlType = "error"
	CtrlSyncPolicies        ControlType = "sync_policies"
	CtrlStatusUpdate        ControlType = "status_update"
	CtrlAuditEntry          ControlType = "audit_entry"
	CtrlDeactivateDeveloper ControlType = "deactivate_developer"
	CtrlClientConnected     ControlType = "client_connected"
	CtrlClientCount         ControlType = "client_count"
)

// ControlMessage is the wire format for relay control protocol messages.
type ControlMessage struct {
	Type    ControlType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RegisterPayload is sent by the daemon to register a session on the relay.
type RegisterPayload struct {
	SessionID   string `json:"sid"`
	KeyHMAC     string `json:"key_hmac"`
	AgentType   string `json:"agent"`
	ProjectName string `json:"project"`
	DisplayName string `json:"display"`
	DeveloperID string `json:"dev_id,omitempty"`
}

// JoinPayload is sent by the mobile client to join a session.
// JWT carries the clientToken from QRPayload.token.
type JoinPayload struct {
	SessionID string `json:"sid"`
	JWT       string `json:"jwt"`
	ClientID  string `json:"client_id,omitempty"`
}

// AckPayload is the relay's acknowledgement of a successful registration or join.
type AckPayload struct {
	SessionID string `json:"sid"`
	ClientID  string `json:"client_id,omitempty"`
}

// ErrorPayload is the relay's error response to a failed control operation.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// StatusUpdatePayload is sent by the daemon to the relay every 30s with
// current session state. The relay intercepts this and updates its in-memory
// state; it is never forwarded to clients.
type StatusUpdatePayload struct {
	SessionID   string  `json:"sid"`
	State       string  `json:"state"`
	CostUSD     float64 `json:"cost_usd,omitempty"`
	Project     string  `json:"project,omitempty"`
	AgentType   string  `json:"agent,omitempty"`
	DeveloperID string  `json:"dev_id,omitempty"`
	CreatedAt   int64   `json:"created_at,omitempty"`
}

// AuditEntryPayload carries a single audit event from daemon to relay.
// InputHash is sha256hex of raw tool_input JSON bytes — NOT plaintext.
type AuditEntryPayload struct {
	Timestamp   time.Time `json:"ts"`
	SessionID   string    `json:"session_id"`
	DeveloperID string    `json:"developer_id"`
	AgentType   string    `json:"agent"`
	EventType   string    `json:"event"`
	Tool        string    `json:"tool,omitempty"`
	InputHash   string    `json:"input_hash,omitempty"`
	Decision    string    `json:"decision,omitempty"`
	PolicyName  string    `json:"policy,omitempty"`
	CostUSD     float64   `json:"cost_usd,omitempty"`
}

// DeactivateDeveloperPayload is sent by the relay to the daemon when a
// developer is deactivated via SCIM.
type DeactivateDeveloperPayload struct {
	DeveloperID string `json:"developer_id"`
}

// ClientConnectedPayload is sent by the relay to the daemon when a PWA client
// connects or reconnects. The daemon uses this to replay message history.
type ClientConnectedPayload struct {
	SessionID string `json:"session_id"`
}

// ClientCountPayload is sent by the relay to the daemon after every client
// join and disconnect. Informs the daemon of the current connected client count
// for replay optimization (first client: full replay, subsequent: session list only).
type ClientCountPayload struct {
	Count     int    `json:"count"`
	SessionID string `json:"session_id"`
}

// SyncPoliciesPayload is sent by relay to daemon after CtrlAck.
// Carries the current org/team policy set.
type SyncPoliciesPayload struct {
	Policies []PolicyJSON `json:"policies"`
}
