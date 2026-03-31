// Package protocol defines the shared wire protocol types for the AgentD relay
// system. Both the daemon (agentd) and relay server (agentd-relay) import this
// module to guarantee type identity across the WebSocket JSON contract.
//
// This module has zero external dependencies — only encoding/json, time,
// crypto/rand, and encoding/hex.
package protocol

// RelayEnvelope is the wire format for encrypted session messages routed
// between daemon and PWA via the relay. The relay routes by SessionID and
// NEVER inspects the Encrypted payload.
type RelayEnvelope struct {
	SessionID string `json:"sid"`
	Seq       uint64 `json:"seq"`
	Encrypted []byte `json:"enc"`
	TraceID   string `json:"tid,omitempty"` // W3C trace-id (32 hex chars); optional for backward compat
}
