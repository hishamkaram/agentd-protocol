package protocol

// MsgInteractivePromptResolved is the daemon->PWA tombstone broadcast after
// an interactive prompt answer is accepted by the daemon-side agent adapter.
const MsgInteractivePromptResolved = "interactive_prompt_resolved"

// InteractivePromptResolvedPayload is carried by MsgInteractivePromptResolved.
// The PWA records QuestionID as resolved for SessionID and clears only matching
// pending/in-flight prompt UI state. Reason is a stable, human-readable machine
// string such as "answered".
type InteractivePromptResolvedPayload struct {
	QuestionID string `json:"question_id"`
	SessionID  string `json:"session_id"`
	Reason     string `json:"reason"`
	ResolvedAt int64  `json:"resolved_at"`
}
