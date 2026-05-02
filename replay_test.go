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
