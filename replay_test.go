package protocol

import (
	"encoding/json"
	"testing"
)

func TestReplayMessageTypeConstants(t *testing.T) {
	if MsgReplayRequest != "replay_request" {
		t.Fatalf("MsgReplayRequest = %q, want replay_request", MsgReplayRequest)
	}
	if MsgReplayComplete != "replay_complete" {
		t.Fatalf("MsgReplayComplete = %q, want replay_complete", MsgReplayComplete)
	}
	if MsgSessionSnapshotRequest != "session_snapshot_request" {
		t.Fatalf("MsgSessionSnapshotRequest = %q, want session_snapshot_request", MsgSessionSnapshotRequest)
	}
	if MsgSessionSnapshot != "session_snapshot" {
		t.Fatalf("MsgSessionSnapshot = %q, want session_snapshot", MsgSessionSnapshot)
	}
	if MsgReplayRequest == MsgReplayComplete {
		t.Fatal("replay request and completion message types must differ")
	}
}

func TestReplayRequestJSONRoundtrip(t *testing.T) {
	in := ReplayRequest{
		Type:      MsgReplayRequest,
		SessionID: "session-123",
		AfterSeq:  42,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal ReplayRequest: %v", err)
	}
	const wantJSON = `{"type":"replay_request","session_id":"session-123","after_seq":42}`
	if string(raw) != wantJSON {
		t.Fatalf("ReplayRequest JSON = %s, want %s", raw, wantJSON)
	}

	var out ReplayRequest
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal ReplayRequest: %v", err)
	}
	if out != in {
		t.Fatalf("ReplayRequest roundtrip = %+v, want %+v", out, in)
	}
}

func TestReplayCompletePayloadJSONRoundtrip(t *testing.T) {
	in := ReplayCompletePayload{
		SessionID: "session-123",
		FromSeq:   43,
		ToSeq:     99,
		Count:     57,
		Error:     "replay unavailable",
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal ReplayCompletePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","from_seq":43,"to_seq":99,"count":57,"error":"replay unavailable"}`
	if string(raw) != wantJSON {
		t.Fatalf("ReplayCompletePayload JSON = %s, want %s", raw, wantJSON)
	}

	var out ReplayCompletePayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal ReplayCompletePayload: %v", err)
	}
	if out != in {
		t.Fatalf("ReplayCompletePayload roundtrip = %+v, want %+v", out, in)
	}
}

func TestReplayCompletePayloadOmitsEmptyError(t *testing.T) {
	raw, err := json.Marshal(ReplayCompletePayload{
		SessionID: "session-123",
		FromSeq:   1,
		ToSeq:     1,
		Count:     1,
	})
	if err != nil {
		t.Fatalf("marshal ReplayCompletePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","from_seq":1,"to_seq":1,"count":1}`
	if string(raw) != wantJSON {
		t.Fatalf("ReplayCompletePayload JSON = %s, want %s", raw, wantJSON)
	}
}

func TestSessionSnapshotRequestJSONRoundtrip(t *testing.T) {
	in := SessionSnapshotRequest{
		Type:        MsgSessionSnapshotRequest,
		SessionID:   "session-123",
		Reason:      "route_reopen",
		AfterCursor: "99",
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal SessionSnapshotRequest: %v", err)
	}
	const wantJSON = `{"type":"session_snapshot_request","session_id":"session-123","reason":"route_reopen","after_cursor":"99"}`
	if string(raw) != wantJSON {
		t.Fatalf("SessionSnapshotRequest JSON = %s, want %s", raw, wantJSON)
	}

	var out SessionSnapshotRequest
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal SessionSnapshotRequest: %v", err)
	}
	if out != in {
		t.Fatalf("SessionSnapshotRequest roundtrip = %+v, want %+v", out, in)
	}
}

func TestSessionSnapshotPayloadJSONRoundtrip(t *testing.T) {
	in := SessionSnapshotPayload{
		SessionID:        "session-123",
		Session:          json.RawMessage(`{"id":"session-123","state":"running"}`),
		Messages:         []json.RawMessage{json.RawMessage(`{"type":"output","session_id":"session-123","seq":1}`)},
		PendingApprovals: []json.RawMessage{json.RawMessage(`{"approval_id":"approval-1","session_id":"session-123"}`)},
		PendingPrompts:   []json.RawMessage{json.RawMessage(`{"question_id":"question-1"}`)},
		Cursor:           "1",
		Sequence:         1,
		Complete:         true,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal SessionSnapshotPayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","session":{"id":"session-123","state":"running"},"messages":[{"type":"output","session_id":"session-123","seq":1}],"pending_approvals":[{"approval_id":"approval-1","session_id":"session-123"}],"pending_prompts":[{"question_id":"question-1"}],"cursor":"1","sequence":1,"complete":true}`
	if string(raw) != wantJSON {
		t.Fatalf("SessionSnapshotPayload JSON = %s, want %s", raw, wantJSON)
	}

	var out SessionSnapshotPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal SessionSnapshotPayload: %v", err)
	}
	if out.SessionID != in.SessionID || out.Cursor != in.Cursor || out.Sequence != in.Sequence || out.Complete != in.Complete {
		t.Fatalf("SessionSnapshotPayload scalar roundtrip = %+v, want %+v", out, in)
	}
	if string(out.Session) != string(in.Session) || len(out.Messages) != 1 || string(out.Messages[0]) != string(in.Messages[0]) {
		t.Fatalf("SessionSnapshotPayload raw state roundtrip = %+v, want %+v", out, in)
	}
	if len(out.PendingApprovals) != 1 || string(out.PendingApprovals[0]) != string(in.PendingApprovals[0]) {
		t.Fatalf("PendingApprovals = %s, want %s", out.PendingApprovals, in.PendingApprovals)
	}
	if len(out.PendingPrompts) != 1 || string(out.PendingPrompts[0]) != string(in.PendingPrompts[0]) {
		t.Fatalf("PendingPrompts = %s, want %s", out.PendingPrompts, in.PendingPrompts)
	}
}

func TestSessionSnapshotPayloadOmitsOptionalFields(t *testing.T) {
	raw, err := json.Marshal(SessionSnapshotPayload{
		SessionID:        "session-123",
		Messages:         []json.RawMessage{},
		PendingApprovals: []json.RawMessage{},
		PendingPrompts:   []json.RawMessage{},
		Complete:         true,
	})
	if err != nil {
		t.Fatalf("marshal SessionSnapshotPayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","messages":[],"pending_approvals":[],"pending_prompts":[],"complete":true}`
	if string(raw) != wantJSON {
		t.Fatalf("SessionSnapshotPayload JSON = %s, want %s", raw, wantJSON)
	}
}
