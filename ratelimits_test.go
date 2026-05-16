package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

// Feature 186 — T003-T: `RateLimitsUpdatedPayload` + `RateLimitBucket` wire
// types. These tests pin the struct shapes, JSON tags (including omitempty on
// optional fields), the `MsgRateLimitsUpdated` MessageType constant value, the
// marshal/unmarshal round-trip, and the malformed-JSON safety contract.
//
// Source of truth: specs/186-codex-remaining-gaps/contracts/
// rate-limits-updated-payload.md and data-model.md → RateLimitsUpdatedPayload.

// TestRateLimitBucketFields pins the 5 fields + types + JSON tags on
// RateLimitBucket per data-model.md.
func TestRateLimitBucketFields(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.RateLimitBucket{})
	if typ.Kind() != reflect.Struct {
		t.Fatalf("RateLimitBucket must be a struct, got %v", typ.Kind())
	}

	type want struct {
		goKind  reflect.Kind
		jsonTag string
	}
	cases := map[string]want{
		"UsedPercent":     {reflect.Float64, "used_percent"},
		"WindowStart":     {reflect.String, "window_start,omitempty"},
		"WindowEnd":       {reflect.String, "window_end,omitempty"},
		"ResetsInSeconds": {reflect.Int64, "resets_in_seconds,omitempty"},
		"LabelID":         {reflect.String, "label_id,omitempty"},
	}

	if got := typ.NumField(); got != len(cases) {
		t.Fatalf("RateLimitBucket field count: want %d, got %d", len(cases), got)
	}

	for fieldName, w := range cases {
		fieldName, w := fieldName, w
		t.Run(fieldName, func(t *testing.T) {
			t.Parallel()

			f, ok := typ.FieldByName(fieldName)
			if !ok {
				t.Fatalf("RateLimitBucket missing field %s", fieldName)
			}
			if f.Type.Kind() != w.goKind {
				t.Fatalf("RateLimitBucket.%s: want %s, got %s", fieldName, w.goKind, f.Type.Kind())
			}
			if got := f.Tag.Get("json"); got != w.jsonTag {
				t.Fatalf("RateLimitBucket.%s: json tag want %q, got %q", fieldName, w.jsonTag, got)
			}
		})
	}
}

// TestRateLimitsUpdatedPayloadFields pins the 5 top-level fields on
// RateLimitsUpdatedPayload per data-model.md.
func TestRateLimitsUpdatedPayloadFields(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(protocol.RateLimitsUpdatedPayload{})
	if typ.Kind() != reflect.Struct {
		t.Fatalf("RateLimitsUpdatedPayload must be a struct, got %v", typ.Kind())
	}

	type want struct {
		// goTypeStr is a substring check against field.Type.String() so we
		// allow "*protocol.RateLimitBucket" and "map[string]*protocol.RateLimitBucket"
		// to be matched without coupling to the package-path prefix.
		goTypeStr string
		jsonTag   string
	}
	cases := map[string]want{
		"PrimaryWindow":     {"*", "primary_window,omitempty"},
		"SecondaryWindow":   {"*", "secondary_window,omitempty"},
		"ByLimitID":         {"map[string]", "by_limit_id,omitempty"},
		"SessionID":         {"string", "session_id"},
		"ReceivedAtMillis":  {"int64", "received_at_ms"},
		"PlanType":          {"string", "plan_type,omitempty"},
		"CreditsBalance":    {"string", "credits_balance,omitempty"},
		"CreditsHasCredits": {"*", "credits_has_credits,omitempty"},
		"CreditsUnlimited":  {"*", "credits_unlimited,omitempty"},
	}

	if got := typ.NumField(); got != len(cases) {
		t.Fatalf("RateLimitsUpdatedPayload field count: want %d, got %d", len(cases), got)
	}

	for fieldName, w := range cases {
		fieldName, w := fieldName, w
		t.Run(fieldName, func(t *testing.T) {
			t.Parallel()

			f, ok := typ.FieldByName(fieldName)
			if !ok {
				t.Fatalf("RateLimitsUpdatedPayload missing field %s", fieldName)
			}
			if !strings.Contains(f.Type.String(), w.goTypeStr) {
				t.Fatalf("RateLimitsUpdatedPayload.%s: Go type %q must contain %q", fieldName, f.Type.String(), w.goTypeStr)
			}
			if got := f.Tag.Get("json"); got != w.jsonTag {
				t.Fatalf("RateLimitsUpdatedPayload.%s: json tag want %q, got %q", fieldName, w.jsonTag, got)
			}
		})
	}
}

// TestMsgRateLimitsUpdatedConstant pins the wire-type string used for the
// MessageType discriminator.
func TestMsgRateLimitsUpdatedConstant(t *testing.T) {
	t.Parallel()

	// Accept either a typed MessageType-ish constant or a bare string alias;
	// the contract only requires the value "rate_limits_updated". Use an
	// untyped comparison via fmt.Sprintf to keep this test robust against
	// implementer choice (MessageType vs string alias).
	got := string(protocol.MsgRateLimitsUpdated)
	if got != "rate_limits_updated" {
		t.Fatalf("MsgRateLimitsUpdated: want %q, got %q", "rate_limits_updated", got)
	}
}

// TestRateLimitsUpdatedPayloadRoundTrip verifies a representative payload
// round-trips through JSON without loss.
func TestRateLimitsUpdatedPayloadRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.RateLimitsUpdatedPayload
	}{
		{
			name: "full payload with both windows and by_limit_id",
			in: protocol.RateLimitsUpdatedPayload{
				PrimaryWindow: &protocol.RateLimitBucket{
					UsedPercent:     42.5,
					WindowStart:     "2026-04-20T00:00:00Z",
					WindowEnd:       "2026-04-20T01:00:00Z",
					ResetsInSeconds: 1800,
					LabelID:         "rpm-5h",
				},
				SecondaryWindow: &protocol.RateLimitBucket{
					UsedPercent:     12.0,
					WindowStart:     "2026-04-20T00:00:00Z",
					WindowEnd:       "2026-04-27T00:00:00Z",
					ResetsInSeconds: 604800,
					LabelID:         "weekly",
				},
				ByLimitID: map[string]*protocol.RateLimitBucket{
					"rpm-5h": {UsedPercent: 42.5, LabelID: "rpm-5h"},
					"weekly": {UsedPercent: 12.0, LabelID: "weekly"},
				},
				SessionID:         "sess-abc",
				ReceivedAtMillis:  1713600000000,
				PlanType:          "plus",
				CreditsBalance:    "1200",
				CreditsHasCredits: boolPtr(true),
				CreditsUnlimited:  boolPtr(false),
			},
		},
		{
			name: "minimal payload — required fields only",
			in: protocol.RateLimitsUpdatedPayload{
				SessionID:        "sess-min",
				ReceivedAtMillis: 1,
			},
		},
		{
			name: "primary only",
			in: protocol.RateLimitsUpdatedPayload{
				PrimaryWindow:    &protocol.RateLimitBucket{UsedPercent: 0.0},
				SessionID:        "sess-p",
				ReceivedAtMillis: 42,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got protocol.RateLimitsUpdatedPayload
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if !reflect.DeepEqual(got, tt.in) {
				t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", tt.in, got, raw)
			}
		})
	}
}

// TestRateLimitsUpdatedPayloadJSONKeys asserts the exact snake_case keys in
// the marshaled payload (including omitempty drops).
func TestRateLimitsUpdatedPayloadJSONKeys(t *testing.T) {
	t.Parallel()

	p := protocol.RateLimitsUpdatedPayload{
		PrimaryWindow: &protocol.RateLimitBucket{
			UsedPercent:     50.0,
			WindowStart:     "start",
			WindowEnd:       "end",
			ResetsInSeconds: 60,
			LabelID:         "lbl",
		},
		SessionID:        "sid",
		ReceivedAtMillis: 123,
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(raw)

	wantKeys := []string{
		`"primary_window"`,
		`"used_percent":50`,
		`"window_start":"start"`,
		`"window_end":"end"`,
		`"resets_in_seconds":60`,
		`"label_id":"lbl"`,
		`"session_id":"sid"`,
		`"received_at_ms":123`,
	}
	for _, k := range wantKeys {
		if !strings.Contains(payload, k) {
			t.Errorf("missing key %q in JSON payload: %s", k, payload)
		}
	}

	// omitempty: SecondaryWindow and ByLimitID must be absent when unset.
	forbidden := []string{`"secondary_window"`, `"by_limit_id"`}
	for _, k := range forbidden {
		if strings.Contains(payload, k) {
			t.Errorf("unexpected key %q present (should be omitted): %s", k, payload)
		}
	}
}

func boolPtr(v bool) *bool {
	return &v
}

// TestRateLimitsUpdatedPayloadMalformedJSON verifies malformed input returns a
// Go unmarshal error rather than panicking. Per data-model.md the daemon-side
// parser falls back to the MsgOutput text path on error.
func TestRateLimitsUpdatedPayloadMalformedJSON(t *testing.T) {
	t.Parallel()

	badInputs := []string{
		`{`,                                   // truncated
		`{"primary_window": "not-an-object"}`, // type mismatch
		`{"received_at_ms": "not-a-number"}`,
		`null-garbage`,
	}

	for _, raw := range badInputs {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			var p protocol.RateLimitsUpdatedPayload
			err := json.Unmarshal([]byte(raw), &p)
			if err == nil {
				t.Fatalf("expected unmarshal error for malformed input %q", raw)
			}
		})
	}
}
