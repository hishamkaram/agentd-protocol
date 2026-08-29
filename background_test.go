package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBackgroundWireConstants(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"background_control":        MsgBackgroundControl,
		"background_control_result": MsgBackgroundControlResult,
		"background_state":          MsgBackgroundState,
	} {
		if value != name {
			t.Fatalf("%s message = %q", name, value)
		}
	}
	for name, value := range map[string]string{
		"stop_task":          BackgroundControlActionStopTask,
		"stop_all_terminals": BackgroundControlActionStopAllTerminals,
	} {
		if value != name {
			t.Fatalf("%s action = %q", name, value)
		}
	}
	for name, value := range map[string]string{
		"accepted":    BackgroundControlStatusAccepted,
		"unsupported": BackgroundControlStatusUnsupported,
	} {
		if value != name {
			t.Fatalf("%s status = %q", name, value)
		}
	}
}

func TestBackgroundControlRequestRoundTripAndValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      BackgroundControlRequest
		wantErr bool
	}{
		{
			name: "stop task",
			in: BackgroundControlRequest{
				Type:      MsgBackgroundControl,
				RequestID: "req-1",
				SessionID: "session-1",
				TaskID:    "task-1",
				Action:    BackgroundControlActionStopTask,
			},
		},
		{
			name: "stop all terminals",
			in: BackgroundControlRequest{
				Type:      MsgBackgroundControl,
				RequestID: "req-2",
				SessionID: "session-1",
				Action:    BackgroundControlActionStopAllTerminals,
			},
		},
		{
			name: "stop task requires task id",
			in: BackgroundControlRequest{
				Type:      MsgBackgroundControl,
				RequestID: "req-3",
				SessionID: "session-1",
				Action:    BackgroundControlActionStopTask,
			},
			wantErr: true,
		},
		{
			name: "stop all terminals rejects task id",
			in: BackgroundControlRequest{
				Type:      MsgBackgroundControl,
				RequestID: "req-4",
				SessionID: "session-1",
				TaskID:    "task-foreign",
				Action:    BackgroundControlActionStopAllTerminals,
			},
			wantErr: true,
		},
		{
			name: "requires request id",
			in: BackgroundControlRequest{
				Type:      MsgBackgroundControl,
				SessionID: "session-1",
				TaskID:    "task-1",
				Action:    BackgroundControlActionStopTask,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var decoded BackgroundControlRequest
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if decoded.Type != tt.in.Type ||
				decoded.RequestID != tt.in.RequestID ||
				decoded.SessionID != tt.in.SessionID ||
				decoded.TaskID != tt.in.TaskID ||
				decoded.Action != tt.in.Action {
				t.Fatalf("decoded request = %+v, want %+v", decoded, tt.in)
			}
			if err := decoded.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackgroundControlResultRoundTrip(t *testing.T) {
	t.Parallel()

	original := BackgroundControlResult{
		Type:      MsgBackgroundControlResult,
		RequestID: "req-1",
		SessionID: "session-1",
		TaskID:    "task-1",
		Action:    BackgroundControlActionStopTask,
		Success:   true,
		Status:    BackgroundControlStatusAccepted,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded BackgroundControlResult
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Type != MsgBackgroundControlResult ||
		decoded.RequestID != "req-1" ||
		decoded.SessionID != "session-1" ||
		decoded.TaskID != "task-1" ||
		decoded.Action != BackgroundControlActionStopTask ||
		!decoded.Success ||
		decoded.Status != BackgroundControlStatusAccepted {
		t.Fatalf("decoded result = %+v", decoded)
	}
}

func TestBackgroundStatePayloadRoundTripAndValidate(t *testing.T) {
	t.Parallel()

	original := BackgroundStatePayload{
		Type:                  MsgBackgroundState,
		SessionID:             "session-1",
		ActiveTaskIDs:         []string{"task-1"},
		ActiveTerminalHandles: []string{"proc-1", "proc-2"},
		ActiveTerminalCount:   2,
		UpdatedAtUnixMS:       1782904297561,
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded BackgroundStatePayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	if decoded.SessionID != original.SessionID ||
		decoded.ActiveTerminalCount != 2 ||
		len(decoded.ActiveTerminalHandles) != 2 ||
		decoded.ActiveTerminalHandles[1] != "proc-2" ||
		len(decoded.ActiveTaskIDs) != 1 {
		t.Fatalf("decoded state = %+v", decoded)
	}
}

func TestBackgroundStatePayloadAcceptsHandlesOnlySnapshot(t *testing.T) {
	t.Parallel()

	var decoded BackgroundStatePayload
	raw := []byte(`{"type":"background_state","session_id":"session-1","active_terminal_handles":["proc-1"]}`)
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	if decoded.ActiveTerminalCount != 0 || len(decoded.ActiveTerminalHandles) != 1 {
		t.Fatalf("decoded handles-only state = %+v", decoded)
	}
}

func TestBackgroundMessagesBackwardCompatibleDecode(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`{"type":"background_control_result","request_id":"req-1","session_id":"s1","success":true,"extra":"ignored"}`,
		`{"type":"background_state","session_id":"s1","active_terminal_count":0,"extra":"ignored"}`,
	} {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			var payload map[string]json.RawMessage
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				t.Fatalf("map unmarshal: %v", err)
			}
			if _, ok := payload["extra"]; !ok {
				t.Fatalf("expected extra field in raw payload")
			}
			if strings.Contains(raw, "control_result") {
				var decoded BackgroundControlResult
				if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
					t.Fatalf("result unmarshal: %v", err)
				}
				return
			}
			var decoded BackgroundStatePayload
			if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
				t.Fatalf("state unmarshal: %v", err)
			}
		})
	}
}
