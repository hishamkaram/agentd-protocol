package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSessionFeatureStatusPayloadJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := SessionFeatureStatusPayload{
		SchemaVersion:    SessionFeatureStatusSchemaVersion,
		SessionID:        "sess-1",
		Feature:          SessionFeatureGitWatch,
		State:            SessionFeatureStateDegraded,
		ReasonCode:       SessionFeatureReasonNotAGitRepo,
		Message:          "not a repo",
		ObservedAtUnixMS: 1710000000000,
		Attempt:          2,
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got SessionFeatureStatusPayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("roundtrip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}
}

func TestSessionFeatureStatusPayloadRequiresFields(t *testing.T) {
	t.Parallel()

	base := SessionFeatureStatusPayload{
		SchemaVersion:    SessionFeatureStatusSchemaVersion,
		SessionID:        "sess-1",
		Feature:          SessionFeatureGitWatch,
		State:            SessionFeatureStatePending,
		ObservedAtUnixMS: 1710000000000,
	}
	tests := []struct {
		name   string
		mutate func(*SessionFeatureStatusPayload)
		want   string
	}{
		{name: "schema_version", mutate: func(p *SessionFeatureStatusPayload) { p.SchemaVersion = 0 }, want: "schema_version"},
		{name: "session_id", mutate: func(p *SessionFeatureStatusPayload) { p.SessionID = "" }, want: "session_id"},
		{name: "feature", mutate: func(p *SessionFeatureStatusPayload) { p.Feature = "" }, want: "feature"},
		{name: "state", mutate: func(p *SessionFeatureStatusPayload) { p.State = "" }, want: "state"},
		{name: "observed_at_unix_ms", mutate: func(p *SessionFeatureStatusPayload) { p.ObservedAtUnixMS = 0 }, want: "observed_at_unix_ms"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := base
			tt.mutate(&payload)
			err := payload.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestSessionFeatureStatusPayloadRejectsUnknownStateAndReason(t *testing.T) {
	t.Parallel()

	rawState := []byte(`{"schema_version":1,"session_id":"sess-1","feature":"git.watch","state":"mystery","observed_at_unix_ms":1710000000000}`)
	var payload SessionFeatureStatusPayload
	if err := json.Unmarshal(rawState, &payload); err == nil {
		t.Fatal("unmarshal accepted unknown state")
	}

	rawReason := []byte(`{"schema_version":1,"session_id":"sess-1","feature":"git.watch","state":"degraded","reason_code":"mystery","observed_at_unix_ms":1710000000000}`)
	if err := json.Unmarshal(rawReason, &payload); err == nil {
		t.Fatal("unmarshal accepted unknown reason_code")
	}
}

func TestSessionFeatureStatusConstants(t *testing.T) {
	t.Parallel()

	if MsgSessionFeatureStatus != "session_feature_status" {
		t.Fatalf("MsgSessionFeatureStatus = %q, want session_feature_status", MsgSessionFeatureStatus)
	}
	if EventSessionFeatureStatus != "session.feature_status" {
		t.Fatalf("EventSessionFeatureStatus = %q, want session.feature_status", EventSessionFeatureStatus)
	}
	if CapabilitySessionFeatureStatus != "session.feature_status" {
		t.Fatalf("CapabilitySessionFeatureStatus = %q, want session.feature_status", CapabilitySessionFeatureStatus)
	}
	result, err := NegotiateProtocolHello(ProtocolHelloParams{})
	if err != nil {
		t.Fatalf("NegotiateProtocolHello: %v", err)
	}
	if !containsStringValue(result.Capabilities, CapabilitySessionFeatureStatus) {
		t.Fatalf("capabilities missing %q: %v", CapabilitySessionFeatureStatus, result.Capabilities)
	}
}

func TestSessionInfoFeatureStatusesJSONRoundtrip(t *testing.T) {
	t.Parallel()

	in := SessionInfo{FeatureStatuses: []SessionFeatureStatusPayload{{
		SchemaVersion:    SessionFeatureStatusSchemaVersion,
		SessionID:        "sess-1",
		Feature:          SessionFeatureMCPSync,
		State:            SessionFeatureStateFailed,
		ReasonCode:       SessionFeatureReasonMCPApplyFailed,
		ObservedAtUnixMS: 1710000000000,
	}}}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"feature_statuses"`) {
		t.Fatalf("SessionInfo JSON missing feature_statuses: %s", raw)
	}
	var got SessionInfo
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("roundtrip mismatch\nwant: %+v\n got: %+v", in, got)
	}
}

func TestSessionFeatureStatusFixtures(t *testing.T) {
	t.Parallel()

	fixtures := []string{
		"session-feature-status-git-watch-pending.json",
		"session-feature-status-git-watch-ready.json",
		"session-feature-status-git-watch-not-repo.json",
		"session-feature-status-mcp-sync-failed.json",
		"session-feature-status-budget-key-degraded.json",
	}
	for _, name := range fixtures {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			raw, err := os.ReadFile(filepath.Join("schemas", "fixtures", name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			var payload SessionFeatureStatusPayload
			if err := json.Unmarshal(raw, &payload); err != nil {
				t.Fatalf("unmarshal fixture: %v\n%s", err, raw)
			}
		})
	}
}
