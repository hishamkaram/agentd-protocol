// Package protocol — agent capability wire type.
//
// This file defines AgentCapability, the per-agent feature-flag struct that
// the daemon emits to the PWA (embedded in SessionInfo on session_started
// and session_list_response). It is the canonical source of truth for which
// interactive controls the PWA may render for a given agent.
//
// Added in feature 185 (codex-parity-gaps). See
// specs/185-codex-parity-gaps/contracts/agent-capability.md for the canonical
// field mapping and per-agent values, and specs/185-codex-parity-gaps/data-model.md
// for the Go definition and TypeScript mirror.
//
// Both agentd and agentd-relay import this type via type aliases in their
// respective session/relay type files. The TypeScript mirror lives in
// agentd-web/src/types/index.ts and must be kept in lockstep with the
// JSON tags below.

package protocol

// AgentCapability declares which interactive PWA controls the current agent
// supports. Fields are required (non-nullable bool) per the contract. Older
// daemons that do not populate this struct will emit it as the zero value
// (all false) when embedded by pointer with omitempty — PWA falls back to
// "all capabilities available" default per FR-013 when the field is absent.
type AgentCapability struct {
	AnswerQuestion bool `json:"answer_question"`
	SendToolResult bool `json:"send_tool_result"`
	RewindFiles    bool `json:"rewind_files"`
	MCPHotApply    bool `json:"mcp_hot_apply"`
}
