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
// supports. Older daemons that do not populate this struct will emit it as the
// zero value (all false) when embedded by pointer with omitempty — PWA falls
// back to capability-specific defaults when the field is absent.
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
//
// Feature 188 (MCP parity) adds two more fields:
//   - MCPReconnect — whether the agent exposes a real reconnect operation for
//     a single MCP server. Claude: true; Codex: false.
//   - MCPLiveStatusLimited — whether MCP status is inventory-derived / limited
//     rather than a full live-health report. Claude: false; Codex: true.
//
// Feature 195 (bypass-mode-optin) adds one field:
//   - SupportsBypassPermissions — whether the agent supports runtime switching
//     to "bypass" approval mode. For Claude Code, true iff the daemon was
//     configured with `agents.claude_code.allow_bypass_permissions: true`,
//     which causes the CLI subprocess to launch with
//     `--allow-dangerously-skip-permissions`. Codex: false unconditionally
//     (Codex has its own approval semantics). The PWA reads this flag to
//     decide whether to include "bypass" in the approval-mode cycle. See
//     specs/195-bypass-mode-optin/contracts/agent-capability-bypass.md for
//     the canonical contract.
//
// Feature 229 (codex-runtime-full-access) adds one field:
//   - SupportsRuntimeFullAccess — whether the live session supports the
//     Codex-only runtime transition to approval_policy=never plus
//     sandbox=danger-full-access. This is true only when the daemon was
//     configured with `agents.codex.allow_runtime_full_access_mode: true`.
//     Claude remains false.
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
	AnswerQuestion            bool `json:"answer_question"`
	SendToolResult            bool `json:"send_tool_result"`
	RewindFiles               bool `json:"rewind_files"`
	MCPHotApply               bool `json:"mcp_hot_apply"`
	MCPReconnect              bool `json:"mcp_reconnect"`
	MCPLiveStatusLimited      bool `json:"mcp_live_status_limited"`
	SessionScopedApproval     bool `json:"session_scoped_approval"`
	AnswerQuestionFreeText    bool `json:"answer_question_free_text"`
	SupportsBypassPermissions bool `json:"supports_bypass_permissions"`
	SupportsRuntimeFullAccess bool `json:"supports_runtime_full_access"`
	// SupportsDelegation declares whether the daemon will honor a user-initiated
	// cross-agent hand-off (start_delegation) for this session. It is a
	// DAEMON-WIDE capability (true iff the daemon was started with
	// delegation.enabled=true), NOT a per-agent CLI feature, so the Manager
	// stamps it from the wired delegation coordinator rather than the per-agent
	// adapter's Capabilities() method. The PWA gates the hand-off START affordance
	// (SessionCard canStartHandoff) on this flag so a PWA pointed at an
	// older / killswitch-off daemon does not render a button that does nothing
	// over the relay (Finding #18). Older daemons that predate this field emit it
	// absent → the zero value false → the PWA hides the affordance (fail-safe:
	// never advertise a hand-off the daemon will silently drop). Default-false is
	// the deliberate fallback for an unset capability.
	SupportsDelegation bool `json:"supports_delegation"`
}
