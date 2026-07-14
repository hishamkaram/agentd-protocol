package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPendingTurnCapabilityConstant(t *testing.T) {
	if CapabilityPendingTurnState != "pending_turn.state" {
		t.Fatalf("CapabilityPendingTurnState = %q, want pending_turn.state", CapabilityPendingTurnState)
	}
}

func TestPendingTurnStateJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := PendingTurnStatePayload{
		SchemaVersion: PendingTurnStateSchemaVersion,
		SessionID:     "session-123",
		Revision:      7,
		Turns: []PendingTurnItem{{
			ClientMessageID: "client-1",
			QueuePosition:   1,
			QueuedAtUnixMS:  101,
			UpdatedAtUnixMS: 202,
			Stage:           PendingTurnStageQueued,
			Text:            "queued prompt",
			SkillID:         "skill-1",
			Image:           &PendingTurnImageDisplay{MediaType: "image/png", SizeBytes: 128},
		}},
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal PendingTurnStatePayload: %v", err)
	}
	const wantJSON = `{"schema_version":1,"session_id":"session-123","revision":7,"turns":[{"client_message_id":"client-1","queue_position":1,"queued_at_unix_ms":101,"updated_at_unix_ms":202,"stage":"queued","text":"queued prompt","skill_id":"skill-1","image":{"media_type":"image/png","size_bytes":128}}]}`
	if string(raw) != wantJSON {
		t.Fatalf("PendingTurnStatePayload JSON = %s, want %s", raw, wantJSON)
	}

	var out PendingTurnStatePayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal PendingTurnStatePayload: %v", err)
	}
	if err := ValidatePendingTurnState(out); err != nil {
		t.Fatalf("ValidatePendingTurnState: %v", err)
	}
	if out.Turns[0].Text != in.Turns[0].Text || out.Turns[0].Image.SizeBytes != 128 {
		t.Fatalf("roundtrip = %+v, want %+v", out, in)
	}
}

func TestPendingTurnStateEmptyTurnsMarshalAsArray(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(PendingTurnStatePayload{
		SchemaVersion: PendingTurnStateSchemaVersion,
		SessionID:     "session-123",
		Revision:      1,
	})
	if err != nil {
		t.Fatalf("marshal empty PendingTurnStatePayload: %v", err)
	}
	const wantJSON = `{"schema_version":1,"session_id":"session-123","revision":1,"turns":[]}`
	if string(raw) != wantJSON {
		t.Fatalf("PendingTurnStatePayload JSON = %s, want %s", raw, wantJSON)
	}
}

func TestPendingTurnImageDisplayPreservesZeroSize(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(PendingTurnImageDisplay{MediaType: "image/png", SizeBytes: 0})
	if err != nil {
		t.Fatalf("marshal zero-byte PendingTurnImageDisplay: %v", err)
	}
	const wantJSON = `{"media_type":"image/png","size_bytes":0}`
	if string(raw) != wantJSON {
		t.Fatalf("PendingTurnImageDisplay JSON = %s, want %s", raw, wantJSON)
	}
}

func TestPendingTurnImageDisplayKeepsRequiredMediaType(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(PendingTurnImageDisplay{})
	if err != nil {
		t.Fatalf("marshal empty PendingTurnImageDisplay: %v", err)
	}
	const wantJSON = `{"media_type":"","size_bytes":0}`
	if string(raw) != wantJSON {
		t.Fatalf("PendingTurnImageDisplay JSON = %s, want %s", raw, wantJSON)
	}
}

func TestPendingTurnStateValidation(t *testing.T) {
	t.Parallel()

	valid := PendingTurnStatePayload{
		SchemaVersion: PendingTurnStateSchemaVersion,
		SessionID:     "session-123",
		Revision:      1,
		Turns: []PendingTurnItem{{
			ClientMessageID: "client-1",
			QueuePosition:   1,
			QueuedAtUnixMS:  101,
			UpdatedAtUnixMS: 202,
			Stage:           PendingTurnStageDeliveryUnknown,
			Text:            "prompt",
		}},
	}
	if err := ValidatePendingTurnState(valid); err != nil {
		t.Fatalf("ValidatePendingTurnState(valid): %v", err)
	}

	cases := []struct {
		name string
		edit func(*PendingTurnStatePayload)
		want string
	}{
		{name: "bad schema", edit: func(p *PendingTurnStatePayload) { p.SchemaVersion = 2 }, want: "schema_version"},
		{name: "missing session", edit: func(p *PendingTurnStatePayload) { p.SessionID = "" }, want: "session_id"},
		{name: "zero revision", edit: func(p *PendingTurnStatePayload) { p.Revision = 0 }, want: "revision"},
		{name: "duplicate id", edit: func(p *PendingTurnStatePayload) { p.Turns = append(p.Turns, p.Turns[0]); p.Turns[1].QueuePosition = 2 }, want: "duplicate"},
		{name: "bad position", edit: func(p *PendingTurnStatePayload) { p.Turns[0].QueuePosition = 2 }, want: "queue_position"},
		{name: "bad stage", edit: func(p *PendingTurnStatePayload) { p.Turns[0].Stage = "lost" }, want: "stage"},
		{name: "missing image media type", edit: func(p *PendingTurnStatePayload) { p.Turns[0].Image = &PendingTurnImageDisplay{} }, want: "media_type"},
		{name: "negative image size", edit: func(p *PendingTurnStatePayload) {
			p.Turns[0].Image = &PendingTurnImageDisplay{MediaType: "image/png", SizeBytes: -1}
		}, want: "size_bytes"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := valid
			p.Turns = append([]PendingTurnItem(nil), valid.Turns...)
			tc.edit(&p)
			err := ValidatePendingTurnState(p)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidatePendingTurnState error = %v, want containing %q", err, tc.want)
			}
		})
	}
}
