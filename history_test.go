package protocol

import (
	"encoding/json"
	"testing"
)

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
		ErrorCode:         HistoryErrorUnavailable,
		Trimmed:           true,
		RetainedOldestSeq: 97,
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal HistoryPagePayload: %v", err)
	}
	const wantJSON = `{"session_id":"session-123","messages":[{"type":"output","session_id":"session-123","seq":97},{"type":"output","session_id":"session-123","seq":98}],"oldest_seq":97,"has_more":true,"request_id":"request-123","error":"history unavailable","error_code":"history_unavailable","trimmed":true,"retained_oldest_seq":97}`
	if string(raw) != wantJSON {
		t.Fatalf("HistoryPagePayload JSON = %s, want %s", raw, wantJSON)
	}

	var out HistoryPagePayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal HistoryPagePayload: %v", err)
	}
	if out.SessionID != in.SessionID || out.OldestSeq != in.OldestSeq || out.HasMore != in.HasMore || out.RequestID != in.RequestID || out.Error != in.Error || out.ErrorCode != in.ErrorCode || out.Trimmed != in.Trimmed || out.RetainedOldestSeq != in.RetainedOldestSeq {
		t.Fatalf("HistoryPagePayload scalar roundtrip = %+v, want %+v", out, in)
	}
	if len(out.Messages) != len(in.Messages) || string(out.Messages[0]) != string(in.Messages[0]) || string(out.Messages[1]) != string(in.Messages[1]) {
		t.Fatalf("HistoryPagePayload messages = %s, want %s", out.Messages, in.Messages)
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
