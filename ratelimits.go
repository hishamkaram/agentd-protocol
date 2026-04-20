// Package protocol — rate-limits wire type.
//
// This file defines RateLimitsUpdatedPayload and its nested RateLimitBucket,
// the structured wire payload emitted by the daemon when the Codex SDK
// reports an AccountRateLimitsUpdated event. The payload is carried on a new
// MsgRateLimitsUpdated wire type ("rate_limits_updated").
//
// Added in feature 186 (codex-remaining-gaps). See
// specs/186-codex-remaining-gaps/contracts/rate-limits-updated-payload.md for
// the canonical field mapping + back-compat notes, and
// specs/186-codex-remaining-gaps/data-model.md for the complete entity table.
//
// Both agentd and agentd-relay consume this type via type aliases in their
// respective session/relay type files. The TypeScript mirror lives in
// agentd-web/src/types/index.ts and must be kept in lockstep with the
// JSON tags below.

package protocol

// MsgRateLimitsUpdated is the wire-type discriminator for
// RateLimitsUpdatedPayload. Emitted by the daemon on each
// AccountRateLimitsUpdated SDK event. Old PWA clients that do not handle the
// type drop the message silently — graceful degradation.
const MsgRateLimitsUpdated = "rate_limits_updated"

// RateLimitBucket is one window of rate-limit usage. The underlying values
// come from the Codex SDK's typed RateLimits struct
// (codex-agent-sdk-go/types/account.go). UsedPercent is clamped to [0, 100]
// defensively before emission.
type RateLimitBucket struct {
	UsedPercent     float64 `json:"used_percent"`
	WindowStart     string  `json:"window_start,omitempty"`
	WindowEnd       string  `json:"window_end,omitempty"`
	ResetsInSeconds int64   `json:"resets_in_seconds,omitempty"`
	LabelID         string  `json:"label_id,omitempty"`
}

// RateLimitsUpdatedPayload is the structured snapshot of the account's
// rate-limit state at the time of emission. PrimaryWindow / SecondaryWindow
// cover the SDK's two-window model; ByLimitID exposes the richer per-limit-ID
// view when the SDK provides it. SessionID + ReceivedAtMillis are required so
// the PWA can correlate and order snapshots across reconnects.
//
// On SDK parse failure, the daemon falls back to emitting only the back-compat
// MsgOutput text "rate limits updated" (preserves feature 185 behavior) and
// logs `codex_rate_limits_parse_failed` at warn level.
type RateLimitsUpdatedPayload struct {
	PrimaryWindow    *RateLimitBucket            `json:"primary_window,omitempty"`
	SecondaryWindow  *RateLimitBucket            `json:"secondary_window,omitempty"`
	ByLimitID        map[string]*RateLimitBucket `json:"by_limit_id,omitempty"`
	SessionID        string                      `json:"session_id"`
	ReceivedAtMillis int64                       `json:"received_at_ms"`
}
