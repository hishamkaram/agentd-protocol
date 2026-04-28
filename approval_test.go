package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestApprovalResolvedPayload_Roundtrip validates marshal/unmarshal for the
// ApprovalResolvedPayload struct across all six canonical decision values.
// Mirrors the protocol_test.go roundtrip pattern used for every other wire
// type per CLAUDE.md "Roundtrip Tested" requirement.
func TestApprovalResolvedPayload_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		decision string
	}{
		{name: "allow", decision: ApprovalDecisionAllow},
		{name: "allow_session", decision: ApprovalDecisionAllowSession},
		{name: "deny", decision: ApprovalDecisionDeny},
		{name: "timeout", decision: ApprovalDecisionTimeout},
		{name: "canceled", decision: ApprovalDecisionCanceled},
		{name: "superseded", decision: ApprovalDecisionSuperseded},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := ApprovalResolvedPayload{
				ApprovalID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
				SessionID:  "session-001",
				Decision:   tt.decision,
				ResolvedAt: 1745692800000,
			}

			raw, err := json.Marshal(&original)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			// JSON tags must match the contract documented in
			// contracts/msg-approval-resolved.md. Verify each key appears.
			rawStr := string(raw)
			for _, want := range []string{`"approval_id"`, `"session_id"`, `"decision"`, `"resolved_at"`} {
				if !strings.Contains(rawStr, want) {
					t.Errorf("marshaled JSON missing %s; got %s", want, rawStr)
				}
			}

			var decoded ApprovalResolvedPayload
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if decoded != original {
				t.Errorf("roundtrip mismatch:\n  want=%+v\n  got =%+v", original, decoded)
			}
		})
	}
}

// TestMsgApprovalResolvedConstant pins the wire-type string. Any change
// here is a breaking change across all 3 repos (daemon, relay, PWA).
func TestMsgApprovalResolvedConstant(t *testing.T) {
	t.Parallel()
	if MsgApprovalResolved != "approval_resolved" {
		t.Errorf("MsgApprovalResolved = %q, want %q", MsgApprovalResolved, "approval_resolved")
	}
}

// TestMsgPendingApprovalReplayConstant pins the wire-type string for the
// reconnect-replay signal. Distinct from MsgApproval to allow the PWA
// handler to early-return BEFORE addMessage (so chat history is not
// duplicated on each reload).
func TestMsgPendingApprovalReplayConstant(t *testing.T) {
	t.Parallel()
	if MsgPendingApprovalReplay != "pending_approval_replay" {
		t.Errorf("MsgPendingApprovalReplay = %q, want %q", MsgPendingApprovalReplay, "pending_approval_replay")
	}
	if MsgPendingApprovalReplay == MsgApprovalResolved {
		t.Errorf("MsgPendingApprovalReplay must differ from MsgApprovalResolved")
	}
}

// TestApprovalDecisionConstants pins the canonical decision string values.
// These appear on the wire inside ApprovalResolvedPayload.Decision and the
// PWA branches on them. Changing any of these is a breaking change.
func TestApprovalDecisionConstants(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"allow":         ApprovalDecisionAllow,
		"allow_session": ApprovalDecisionAllowSession,
		"deny":          ApprovalDecisionDeny,
		"timeout":       ApprovalDecisionTimeout,
		"canceled":      ApprovalDecisionCanceled,
		"superseded":    ApprovalDecisionSuperseded,
	}

	for expected, actual := range want {
		if expected != actual {
			t.Errorf("ApprovalDecision constant for %q = %q, want %q", expected, actual, expected)
		}
	}
}
