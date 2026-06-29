package protocol

import "encoding/json"

// StartDelegationPayload is the PWA->daemon command that USER-initiates a
// delegation. Carried inside an AgentMessage envelope with
// Type=MsgStartDelegation. The daemon resolves the source session, parks it
// (when Await), and spawns a delegate of ToEngine in the source's working
// directory with Prompt as the delegate's first task.
//
// Await default note: the JSON bool zero-value is false, but the governed
// product default is await=TRUE when the field is ABSENT. omitempty drops an
// explicit false, so an absent field and an explicit false are indistinguishable
// on the wire. The PWA start-delegation UI ALWAYS sends await explicitly
// (Completion 2), so the daemon's start_delegation decoder treats the PWA as
// always-explicit. The daemon-side handler resolves the absent=>true default by
// decoding into a *bool / inspecting the raw JSON, NOT by reading this struct's
// Go zero value. If a future non-PWA producer needs absent=>true, switch this
// field to *bool then; keep omitempty so old consumers stay byte-compatible.
type StartDelegationPayload struct {
	SourceSID string `json:"source_sid"`      // source session to park + delegate from
	ToEngine  string `json:"to_engine"`       // EngineClaude | EngineCodex - the delegate engine
	Prompt    string `json:"prompt"`          // task prompt delivered to the delegate as its first message
	Await     bool   `json:"await,omitempty"` // park source + return result to source agent; daemon defaults absent=>true
}

// StartDelegationAwaitOrDefault resolves the governed await semantics for a
// start_delegation frame from the RAW JSON bytes, applying the absent=>true rule
// that the plain-bool StartDelegationPayload.Await field cannot express on its
// own (omitempty drops an explicit false, so an absent field and an explicit
// false are indistinguishable after decoding into the struct).
//
// This is the start_delegation analog of delegation.DelegateInput.AwaitOrDefault
// for the MCP path: it is the SINGLE place the absent=>true rule is applied for
// the user-initiated trigger, so both daemon dispatch handlers (local WS and
// relay) resolve await identically (pattern-consistency dual-path) and a future
// non-PWA producer that omits await still gets the safe parked default rather
// than the dangerous fire-and-forget false.
//
// Resolution:
//   - the "await" key is ABSENT, JSON null, or the bytes do not decode => true
//   - the "await" key is present with an explicit boolean (true|false) => that value
//
// A malformed frame is treated as absent (=> true): the safe default is to park,
// and the caller has already validated the frame's required fields separately.
func StartDelegationAwaitOrDefault(raw json.RawMessage) bool {
	var probe struct {
		Await *bool `json:"await"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil || probe.Await == nil {
		return true
	}
	return *probe.Await
}
