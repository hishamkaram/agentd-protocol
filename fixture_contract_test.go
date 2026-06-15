package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixtureAgentMessage struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id"`
	Agent     string          `json:"agent"`
	Timestamp int64           `json:"ts"`
	Payload   json.RawMessage `json:"payload"`
	Seq       uint64          `json:"seq,omitempty"`
}

type fixtureApprovalPayload struct {
	ApprovalID string `json:"approval_id"`
	SessionID  string `json:"session_id"`
	Tool       string `json:"tool"`
	Command    string `json:"command"`
	Context    string `json:"context"`
	Risk       string `json:"risk"`
	TimeoutS   int    `json:"timeout_s"`
	ExpiresAt  int64  `json:"expires_at"`
}

func TestCrossRepoWireFixturesParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		validate func(t *testing.T, raw []byte)
	}{
		{name: "session-head.json", validate: validateSessionHeadFixture},
		{name: "history-page-request.json", validate: validateHistoryPageRequestFixture},
		{name: "history-page.json", validate: validateHistoryPageFixture},
		{name: "history-page-additive-unknown-fields.json", validate: validateHistoryPageFixture},
		{name: "route-receipt-success.json", validate: validateRouteReceiptFixture},
		{name: "route-receipt-failed.json", validate: validateRouteReceiptFixture},
		{name: "approval-resolved.json", validate: validateApprovalResolvedFixture},
		{name: "pending-approval-state.json", validate: validatePendingApprovalStateFixture},
		{name: "support-bundle-v1-daemon-minimum.json", validate: validateSupportBundleFixture},
		{name: "support-bundle-v2-pwa-merged-minimum.json", validate: validateSupportBundleFixture},
		{name: "support-bundle-v2-additive-unknown-fields.json", validate: validateSupportBundleFixture},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, err := os.ReadFile(filepath.Join("schemas", "fixtures", tt.name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			tt.validate(t, raw)
		})
	}
}

func validateSessionHeadFixture(t *testing.T, raw []byte) {
	t.Helper()

	var msg fixtureAgentMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal agent message: %v", err)
	}
	if msg.Type != MsgSessionHead {
		t.Fatalf("type = %q, want %q", msg.Type, MsgSessionHead)
	}
	var payload SessionHeadPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal SessionHeadPayload: %v", err)
	}
	if payload.SessionID != msg.SessionID || payload.HeadSeq == 0 {
		t.Fatalf("session_head payload = %+v, msg = %+v", payload, msg)
	}
}

func validateHistoryPageRequestFixture(t *testing.T, raw []byte) {
	t.Helper()

	var req HistoryPageRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal HistoryPageRequest: %v", err)
	}
	if req.Type != MsgHistoryPageRequest || req.SessionID == "" || req.BeforeSeq == 0 || req.RequestID == "" {
		t.Fatalf("history_page_request fixture decoded invalid request: %+v", req)
	}
}

func validateHistoryPageFixture(t *testing.T, raw []byte) {
	t.Helper()

	var msg fixtureAgentMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal agent message: %v", err)
	}
	if msg.Type != MsgHistoryPage {
		t.Fatalf("type = %q, want %q", msg.Type, MsgHistoryPage)
	}
	var payload HistoryPagePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal HistoryPagePayload: %v", err)
	}
	if payload.SessionID != msg.SessionID || payload.RequestID == "" {
		t.Fatalf("history_page payload = %+v, msg = %+v", payload, msg)
	}
}

func validateRouteReceiptFixture(t *testing.T, raw []byte) {
	t.Helper()

	var ctrl ControlMessage
	if err := json.Unmarshal(raw, &ctrl); err != nil {
		t.Fatalf("unmarshal ControlMessage: %v", err)
	}
	if ctrl.Type != CtrlRouteReceipt {
		t.Fatalf("control type = %q, want %q", ctrl.Type, CtrlRouteReceipt)
	}
	var payload RouteReceiptPayload
	if err := json.Unmarshal(ctrl.Payload, &payload); err != nil {
		t.Fatalf("unmarshal RouteReceiptPayload: %v", err)
	}
	if payload.EnvelopeID == "" || payload.ObservedAtUnixMs <= 0 {
		t.Fatalf("route receipt fixture missing required fields: %+v", payload)
	}
	if payload.Routed && payload.ReasonCode != "" {
		t.Fatalf("successful route receipt has reason_code: %+v", payload)
	}
	if !payload.Routed && payload.ReasonCode != string(CommandReceiptReasonRelayRouteFailed) {
		t.Fatalf("failed route receipt reason_code = %q, want %q", payload.ReasonCode, CommandReceiptReasonRelayRouteFailed)
	}
}

func validateApprovalResolvedFixture(t *testing.T, raw []byte) {
	t.Helper()

	var msg fixtureAgentMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal agent message: %v", err)
	}
	if msg.Type != MsgApprovalResolved {
		t.Fatalf("type = %q, want %q", msg.Type, MsgApprovalResolved)
	}
	var payload ApprovalResolvedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal ApprovalResolvedPayload: %v", err)
	}
	if payload.ApprovalID == "" || payload.SessionID != msg.SessionID || payload.Decision != ApprovalDecisionAllow {
		t.Fatalf("approval_resolved payload = %+v, msg = %+v", payload, msg)
	}
}

func validatePendingApprovalStateFixture(t *testing.T, raw []byte) {
	t.Helper()

	var msg fixtureAgentMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal agent message: %v", err)
	}
	if msg.Type != MsgPendingApprovalState {
		t.Fatalf("type = %q, want %q", msg.Type, MsgPendingApprovalState)
	}
	var payload fixtureApprovalPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal pending approval payload: %v", err)
	}
	if payload.ApprovalID == "" || payload.SessionID != msg.SessionID || payload.Risk == "" {
		t.Fatalf("pending approval payload = %+v, msg = %+v", payload, msg)
	}
}

func validateSupportBundleFixture(t *testing.T, raw []byte) {
	t.Helper()

	var bundle SupportBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatalf("unmarshal SupportBundle: %v", err)
	}
	if bundle.SchemaVersion != SupportBundleSchemaVersion && bundle.SchemaVersion != "agentd.support_bundle.v2" {
		t.Fatalf("schema_version = %q", bundle.SchemaVersion)
	}
	if bundle.BundleID == "" || bundle.GeneratedAtUnixMS <= 0 || !bundle.Redaction.Applied {
		t.Fatalf("support bundle fixture missing required diagnostics: %+v", bundle)
	}
}
