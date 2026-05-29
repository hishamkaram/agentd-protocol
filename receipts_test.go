package protocol

import (
	"encoding/json"
	"testing"
)

func TestCommandReceiptConstants(t *testing.T) {
	if MsgCommandReceipt != "command_receipt" {
		t.Fatalf("MsgCommandReceipt = %q, want command_receipt", MsgCommandReceipt)
	}
	if EventCommandReceipt != "command.receipt" {
		t.Fatalf("EventCommandReceipt = %q, want command.receipt", EventCommandReceipt)
	}
	if CapabilityCommandReceipts != "command.receipts" {
		t.Fatalf("CapabilityCommandReceipts = %q, want command.receipts", CapabilityCommandReceipts)
	}
}

func TestCommandReceiptJSONRoundtrip(t *testing.T) {
	in := CommandReceiptPayload{
		SchemaVersion:    CommandReceiptSchemaVersion,
		ReceiptID:        "cmd_1:provider_accepted:171",
		CommandID:        "cmd_1",
		CommandType:      "user_message",
		SessionID:        "sess_1",
		TraceID:          "0123456789abcdef0123456789abcdef",
		Stage:            CommandReceiptStageProviderAccepted,
		ObservedAtUnixMS: 1770000000000,
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal CommandReceiptPayload: %v", err)
	}
	const wantJSON = `{"schema_version":"agentd.command_receipt.v1","receipt_id":"cmd_1:provider_accepted:171","command_id":"cmd_1","command_type":"user_message","session_id":"sess_1","trace_id":"0123456789abcdef0123456789abcdef","stage":"provider_accepted","terminal":false,"retryable":false,"observed_at_unix_ms":1770000000000}`
	if string(raw) != wantJSON {
		t.Fatalf("CommandReceiptPayload JSON = %s, want %s", raw, wantJSON)
	}
	var out CommandReceiptPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal CommandReceiptPayload: %v", err)
	}
	if out != in {
		t.Fatalf("CommandReceiptPayload roundtrip = %+v, want %+v", out, in)
	}
	if err := ValidateCommandReceipt(out); err != nil {
		t.Fatalf("ValidateCommandReceipt: %v", err)
	}
}

func TestCommandReceiptValidation(t *testing.T) {
	valid := CommandReceiptPayload{
		SchemaVersion:    CommandReceiptSchemaVersion,
		ReceiptID:        "cmd_1:failed:172",
		CommandID:        "cmd_1",
		CommandType:      "user_message",
		Stage:            CommandReceiptStageFailed,
		Terminal:         true,
		Retryable:        true,
		ReasonCode:       CommandReceiptReasonQueueFull,
		ObservedAtUnixMS: 1770000000000,
	}
	if err := ValidateCommandReceipt(valid); err != nil {
		t.Fatalf("valid receipt rejected: %v", err)
	}

	invalid := valid
	invalid.CommandID = ""
	if err := ValidateCommandReceipt(invalid); err == nil {
		t.Fatal("receipt without command_id accepted")
	}

	invalid = valid
	invalid.Stage = "sent"
	if err := ValidateCommandReceipt(invalid); err == nil {
		t.Fatal("receipt with unknown stage accepted")
	}

	invalid = valid
	invalid.Stage = CommandReceiptStageFailed
	invalid.Terminal = false
	if err := ValidateCommandReceipt(invalid); err == nil {
		t.Fatal("failed non-terminal receipt accepted")
	}
}

func TestProtocolHelloAdvertisesCommandReceipts(t *testing.T) {
	result, err := NegotiateProtocolHello(ProtocolHelloParams{})
	if err != nil {
		t.Fatalf("NegotiateProtocolHello returned error: %v", err)
	}
	if !containsStringValue(result.Capabilities, CapabilityCommandReceipts) {
		t.Fatalf("v2 capabilities missing command.receipts: %v", result.Capabilities)
	}
}
