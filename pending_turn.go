package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	// MsgPendingTurnState is a daemon->PWA replacement snapshot of queued user
	// turns for one session.
	MsgPendingTurnState = "pending_turn_state"

	PendingTurnStateSchemaVersion = 1
)

type PendingTurnStage string

const (
	PendingTurnStageQueued          PendingTurnStage = "queued"
	PendingTurnStageDispatching     PendingTurnStage = "dispatching"
	PendingTurnStageDeliveryUnknown PendingTurnStage = "delivery_unknown"
)

// PendingTurnImageDisplay carries bounded image metadata for display only. It
// intentionally omits raw image bytes so reconnect snapshots do not duplicate
// image payload storage in browsers.
type PendingTurnImageDisplay struct {
	MediaType string `json:"media_type"`
	SizeBytes int    `json:"size_bytes"`
}

// PendingTurnItem is one pending outgoing user turn in queue order.
type PendingTurnItem struct {
	ClientMessageID string                   `json:"client_message_id"`
	QueuePosition   int                      `json:"queue_position"`
	QueuedAtUnixMS  int64                    `json:"queued_at_unix_ms"`
	UpdatedAtUnixMS int64                    `json:"updated_at_unix_ms"`
	Stage           PendingTurnStage         `json:"stage"`
	Text            string                   `json:"text"`
	SkillID         string                   `json:"skill_id,omitempty"`
	Image           *PendingTurnImageDisplay `json:"image,omitempty"`
}

// PendingTurnStatePayload is the complete pending-turn projection for one
// session at a monotonic revision.
type PendingTurnStatePayload struct {
	SchemaVersion int               `json:"schema_version"`
	SessionID     string            `json:"session_id"`
	Revision      uint64            `json:"revision"`
	Turns         []PendingTurnItem `json:"turns"`
}

func (p PendingTurnStatePayload) MarshalJSON() ([]byte, error) {
	type wire PendingTurnStatePayload
	if p.Turns == nil {
		p.Turns = []PendingTurnItem{}
	}
	return json.Marshal(wire(p))
}

func ValidatePendingTurnState(p PendingTurnStatePayload) error {
	if p.SchemaVersion != PendingTurnStateSchemaVersion {
		return fmt.Errorf("protocol.ValidatePendingTurnState: schema_version = %d, want %d", p.SchemaVersion, PendingTurnStateSchemaVersion)
	}
	if p.SessionID == "" {
		return fmt.Errorf("protocol.ValidatePendingTurnState: session_id is required")
	}
	if p.Revision == 0 {
		return fmt.Errorf("protocol.ValidatePendingTurnState: revision must be positive")
	}
	seen := make(map[string]struct{}, len(p.Turns))
	for i, turn := range p.Turns {
		if turn.ClientMessageID == "" {
			return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].client_message_id is required", i)
		}
		if _, ok := seen[turn.ClientMessageID]; ok {
			return fmt.Errorf("protocol.ValidatePendingTurnState: duplicate client_message_id %q", turn.ClientMessageID)
		}
		seen[turn.ClientMessageID] = struct{}{}
		if turn.QueuePosition != i+1 {
			return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].queue_position = %d, want %d", i, turn.QueuePosition, i+1)
		}
		if !IsKnownPendingTurnStage(turn.Stage) {
			return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].stage is unknown: %q", i, turn.Stage)
		}
		if turn.QueuedAtUnixMS <= 0 {
			return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].queued_at_unix_ms must be positive", i)
		}
		if turn.UpdatedAtUnixMS <= 0 {
			return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].updated_at_unix_ms must be positive", i)
		}
		if err := validatePendingTurnImage(i, turn.Image); err != nil {
			return err
		}
	}
	return nil
}

func validatePendingTurnImage(index int, image *PendingTurnImageDisplay) error {
	if image == nil {
		return nil
	}
	if image.MediaType == "" {
		return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].image.media_type is required", index)
	}
	if image.SizeBytes < 0 {
		return fmt.Errorf("protocol.ValidatePendingTurnState: turns[%d].image.size_bytes must be non-negative", index)
	}
	return nil
}

func IsKnownPendingTurnStage(stage PendingTurnStage) bool {
	switch stage {
	case PendingTurnStageQueued,
		PendingTurnStageDispatching,
		PendingTurnStageDeliveryUnknown:
		return true
	default:
		return false
	}
}
