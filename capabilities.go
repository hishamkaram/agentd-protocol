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
//
// Feature 186 (codex-remaining-gaps) adds two fields:
//   - SessionScopedApproval — enables 3rd "Allow for session" ApprovalGate
//     button for agents whose runtime supports session-scoped permission
//     grants (Codex: true; Claude: false).
//     See specs/186-codex-remaining-gaps/contracts/session-scoped-approval-capability.md.
//   - AnswerQuestionFreeText — whether the free-text reply box applies to
//     agent-initiated AskUserQuestion prompts. Replaces the semantics of
//     AnswerQuestion over a two-release deprecation window. Claude: true;
//     Codex: false (Codex uses structured ElicitationResponse, not free text).
//     See specs/186-codex-remaining-gaps/contracts/answer-question-free-text-capability.md.
//
// The pre-existing AnswerQuestion field is retained verbatim for back-compat
// per FR-017 deprecation plan.
type AgentCapability struct {
	// Deprecated: As of feature 186 (codex-remaining-gaps), consumers
	// should read AnswerQuestionFreeText instead. This field is retained
	// verbatim for one release so older PWAs that predate 186 continue to
	// render the free-text reply row under the original semantics. The
	// PWA consumer chain is `answer_question_free_text ?? answer_question
	// ?? true` — new field wins, legacy field is the fallback, default
	// favors claude-style rendering. See
	// specs/186-codex-remaining-gaps/contracts/answer-question-free-text-capability.md
	// for the full two-release deprecation plan.
	AnswerQuestion         bool `json:"answer_question"`
	SendToolResult         bool `json:"send_tool_result"`
	RewindFiles            bool `json:"rewind_files"`
	MCPHotApply            bool `json:"mcp_hot_apply"`
	SessionScopedApproval  bool `json:"session_scoped_approval"`
	AnswerQuestionFreeText bool `json:"answer_question_free_text"`
}
