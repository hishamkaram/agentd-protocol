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
	SessionID      string `json:"sid"`
	Seq            uint64 `json:"seq"`
	Encrypted      []byte `json:"enc"`
	KeyEpoch       uint64 `json:"key_epoch"`           // base-key epoch used to encrypt this envelope; zero means legacy/initial epoch
	TraceID        string `json:"tid,omitempty"`       // W3C trace-id (32 hex chars); optional for backward compat
	EnvelopeID     string `json:"eid,omitempty"`       // opaque sender-selected id for route acknowledgement; relay never inspects Encrypted
	TargetClientID string `json:"client_id,omitempty"` // daemon->client only; empty means fan out to all clients
	Reliable       bool   `json:"rel,omitempty"`       // targeted bootstrap/catch-up delivery must commit only after the client write succeeds
	// OpaqueControl carries a targeted recovery control whose payload is
	// encrypted for an enrolled browser device. Relays route and persist these
	// bytes without decoding their contents.
	OpaqueControl []byte `json:"opaque_ctrl,omitempty"`
}
