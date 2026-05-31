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

func TestDurableHistoryMessageTypeConstants(t *testing.T) {
	t.Parallel()

	if MsgSessionHead != "session_head" {
		t.Fatalf("MsgSessionHead = %q, want session_head", MsgSessionHead)
	}
	if MsgHistoryPageRequest != "history_page_request" {
		t.Fatalf("MsgHistoryPageRequest = %q, want history_page_request", MsgHistoryPageRequest)
	}
	if MsgHistoryPage != "history_page" {
		t.Fatalf("MsgHistoryPage = %q, want history_page", MsgHistoryPage)
	}
	if MsgSessionHead == MsgReplayRequest || MsgHistoryPageRequest == MsgReplayRequest || MsgHistoryPage == MsgReplayComplete {
		t.Fatal("durable history message types must not alias existing replay message types")
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

func TestSessionHeadPayloadJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := SessionHeadPayload{
		SessionID:         "session-123",
		HeadSeq:           99,
		RetainedOldestSeq: 7,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal SessionHeadPayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","head_seq":99,"retained_oldest_seq":7}`
	if string(raw) != wantJSON {
		t.Fatalf("SessionHeadPayload JSON = %s, want %s", raw, wantJSON)
	}

	var out SessionHeadPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal SessionHeadPayload: %v", err)
	}
	if out != in {
		t.Fatalf("SessionHeadPayload roundtrip = %+v, want %+v", out, in)
	}
}

func TestHistoryPageRequestJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := HistoryPageRequest{
		Type:      MsgHistoryPageRequest,
		SessionID: "session-123",
		BeforeSeq: 99,
		Limit:     250,
		RequestID: "request-123",
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal HistoryPageRequest: %v", err)
	}
	const wantJSON = `{"type":"history_page_request","session_id":"session-123","before_seq":99,"limit":250,"request_id":"request-123"}`
	if string(raw) != wantJSON {
		t.Fatalf("HistoryPageRequest JSON = %s, want %s", raw, wantJSON)
	}

	var out HistoryPageRequest
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal HistoryPageRequest: %v", err)
	}
	if out != in {
		t.Fatalf("HistoryPageRequest roundtrip = %+v, want %+v", out, in)
	}
}

func TestHistoryPagePayloadJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := HistoryPagePayload{
		SessionID: "session-123",
		Messages: []json.RawMessage{
			json.RawMessage(`{"type":"output","session_id":"session-123","seq":97}`),
			json.RawMessage(`{"type":"output","session_id":"session-123","seq":98}`),
		},
		OldestSeq:         97,
		HasMore:           true,
		RequestID:         "request-123",
		Error:             "history unavailable",
		Trimmed:           true,
		RetainedOldestSeq: 97,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal HistoryPagePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","messages":[{"type":"output","session_id":"session-123","seq":97},{"type":"output","session_id":"session-123","seq":98}],"oldest_seq":97,"has_more":true,"request_id":"request-123","error":"history unavailable","trimmed":true,"retained_oldest_seq":97}`
	if string(raw) != wantJSON {
		t.Fatalf("HistoryPagePayload JSON = %s, want %s", raw, wantJSON)
	}

	var out HistoryPagePayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal HistoryPagePayload: %v", err)
	}
	if out.SessionID != in.SessionID || out.OldestSeq != in.OldestSeq || out.HasMore != in.HasMore || out.RequestID != in.RequestID || out.Error != in.Error || out.Trimmed != in.Trimmed || out.RetainedOldestSeq != in.RetainedOldestSeq {
		t.Fatalf("HistoryPagePayload scalar roundtrip = %+v, want %+v", out, in)
	}
	if len(out.Messages) != len(in.Messages) || string(out.Messages[0]) != string(in.Messages[0]) || string(out.Messages[1]) != string(in.Messages[1]) {
		t.Fatalf("HistoryPagePayload messages = %s, want %s", out.Messages, in.Messages)
	}
}

func TestReplayAndSnapshotHeadSeqJSONRoundtrip(t *testing.T) {
	t.Parallel()

	replay := ReplayCompletePayload{
		SessionID:         "session-123",
		FromSeq:           43,
		ToSeq:             99,
		HeadSeq:           101,
		Count:             57,
		Trimmed:           true,
		TrimFloor:         42,
		RetainedOldestSeq: 43,
	}
	replayRaw, err := json.Marshal(replay)
	if err != nil {
		t.Fatalf("marshal ReplayCompletePayload: %v", err)
	}
	const wantReplayJSON = `{"session_id":"session-123","from_seq":43,"to_seq":99,"head_seq":101,"count":57,"trimmed":true,"trim_floor":42,"retained_oldest_seq":43}`
	if string(replayRaw) != wantReplayJSON {
		t.Fatalf("ReplayCompletePayload JSON = %s, want %s", replayRaw, wantReplayJSON)
	}

	var replayOut ReplayCompletePayload
	if err := json.Unmarshal(replayRaw, &replayOut); err != nil {
		t.Fatalf("unmarshal ReplayCompletePayload: %v", err)
	}
	if replayOut != replay {
		t.Fatalf("ReplayCompletePayload roundtrip = %+v, want %+v", replayOut, replay)
	}

	snapshot := SessionSnapshotPayload{
		SessionID:         "session-123",
		Messages:          []json.RawMessage{},
		PendingApprovals:  []json.RawMessage{},
		PendingPrompts:    []json.RawMessage{},
		HeadSeq:           101,
		RetainedOldestSeq: 43,
		Complete:          true,
	}
	snapshotRaw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal SessionSnapshotPayload: %v", err)
	}
	const wantSnapshotJSON = `{"session_id":"session-123","messages":[],"pending_approvals":[],"pending_prompts":[],"head_seq":101,"retained_oldest_seq":43,"complete":true}`
	if string(snapshotRaw) != wantSnapshotJSON {
		t.Fatalf("SessionSnapshotPayload JSON = %s, want %s", snapshotRaw, wantSnapshotJSON)
	}

	var snapshotOut SessionSnapshotPayload
	if err := json.Unmarshal(snapshotRaw, &snapshotOut); err != nil {
		t.Fatalf("unmarshal SessionSnapshotPayload: %v", err)
	}
	if snapshotOut.SessionID != snapshot.SessionID || snapshotOut.HeadSeq != snapshot.HeadSeq || snapshotOut.RetainedOldestSeq != snapshot.RetainedOldestSeq || snapshotOut.Complete != snapshot.Complete {
		t.Fatalf("SessionSnapshotPayload roundtrip = %+v, want %+v", snapshotOut, snapshot)
	}
}

func TestHistoryPagePayloadOmitsOptionalFields(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(HistoryPagePayload{
		SessionID: "session-123",
		Messages:  []json.RawMessage{},
		OldestSeq: 1,
		HasMore:   false,
		RequestID: "request-123",
	})
	if err != nil {
		t.Fatalf("marshal HistoryPagePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","messages":[],"oldest_seq":1,"has_more":false,"request_id":"request-123"}`
	if string(raw) != wantJSON {
		t.Fatalf("HistoryPagePayload JSON = %s, want %s", raw, wantJSON)
	}
}

func TestReplayCompletePayloadJSONRoundtrip(t *testing.T) {
	in := ReplayCompletePayload{
		SessionID:    "session-123",
		FromSeq:      43,
		ToSeq:        99,
		Count:        57,
		Error:        "replay unavailable",
		ErrorCode:    "unavailable",
		RetryAfterMS: 1500,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal ReplayCompletePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","from_seq":43,"to_seq":99,"count":57,"error":"replay unavailable","error_code":"unavailable","retry_after_ms":1500}`
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
