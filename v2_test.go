package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDecodeJSONRPCRequest(t *testing.T) {
	traceID := NewTraceID()
	raw := []byte(`{"jsonrpc":"2.0","id":"req_1","method":"transport.ping","params":{"sent_at_unix_ms":123},"trace_id":"` + traceID + `"}`)

	req, err := DecodeJSONRPCRequest(raw)
	if err != nil {
		t.Fatalf("DecodeJSONRPCRequest returned error: %v", err)
	}
	if req.JSONRPC != JSONRPCVersion {
		t.Fatalf("JSONRPC = %q, want %q", req.JSONRPC, JSONRPCVersion)
	}
	if string(req.ID) != string(mustJSONRPCStringID(t, "req_1")) {
		t.Fatalf("ID = %q, want req_1", req.ID)
	}
	if req.Method != MethodTransportPing {
		t.Fatalf("Method = %q, want %q", req.Method, MethodTransportPing)
	}
	if req.TraceID != traceID {
		t.Fatalf("TraceID = %q, want %q", req.TraceID, traceID)
	}
}

func TestDecodeJSONRPCRequestRejectsInvalidTraceID(t *testing.T) {
	_, err := DecodeJSONRPCRequest([]byte(`{"jsonrpc":"2.0","id":"req_1","method":"transport.ping","trace_id":"bad"}`))
	if err == nil {
		t.Fatal("DecodeJSONRPCRequest returned nil error for invalid trace_id")
	}
}

func TestDecodeJSONRPCRequestRejectsNotifications(t *testing.T) {
	_, err := DecodeJSONRPCRequest([]byte(`{"jsonrpc":"2.0","method":"transport.ping"}`))
	if err == nil {
		t.Fatal("DecodeJSONRPCRequest returned nil error for missing id")
	}
}

func TestDecodeJSONRPCRequestAcceptsNumericID(t *testing.T) {
	req, err := DecodeJSONRPCRequest([]byte(`{"jsonrpc":"2.0","id":42,"method":"transport.ping"}`))
	if err != nil {
		t.Fatalf("DecodeJSONRPCRequest returned error: %v", err)
	}
	if string(req.ID) != "42" {
		t.Fatalf("ID = %s, want 42", req.ID)
	}
	ok, err := NewJSONRPCResult(req.ID, "", map[string]string{"ok": "true"})
	if err != nil {
		t.Fatalf("NewJSONRPCResult returned error: %v", err)
	}
	raw, err := json.Marshal(ok)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if !regexp.MustCompile(`"id":42`).Match(raw) {
		t.Fatalf("numeric id was not echoed as a number: %s", string(raw))
	}
}

func TestIsJSONRPCRequestRoutesMalformedJSONRPCObjects(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`{"jsonrpc":2,"id":"req_1","method":"transport.ping"}`),
		[]byte(`{"jsonrpc":"1.0","id":"req_1","method":"transport.ping"}`),
		[]byte(`{"jsonrpc":"2.0","id":"req_1"}`),
		[]byte(`{"jsonrpc":"2.0","id":"req_1","method":42}`),
		[]byte(`{"jsonrpc":"2.0","id":"req_1","method":""}`),
	} {
		if !IsJSONRPCRequest(raw) {
			t.Fatalf("IsJSONRPCRequest(%s) = false, want true so decoder returns JSON-RPC error", string(raw))
		}
		if _, err := DecodeJSONRPCRequest(raw); err == nil {
			t.Fatalf("DecodeJSONRPCRequest(%s) returned nil error", string(raw))
		}
	}
}

func TestIsJSONRPCRequestExcludesResponses(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":"req_1","result":{}}`),
		[]byte(`{"jsonrpc":"2.0","id":"req_1","error":{"code":-32601,"message":"Method not found"}}`),
	} {
		if IsJSONRPCRequest(raw) {
			t.Fatalf("IsJSONRPCRequest(%s) = true, want false for JSON-RPC response envelopes", string(raw))
		}
	}
}

func TestJSONRPCResultAndErrorShapes(t *testing.T) {
	ok, err := NewJSONRPCResult(mustJSONRPCStringID(t, "req_1"), "", map[string]string{"type": MsgPong})
	if err != nil {
		t.Fatalf("NewJSONRPCResult returned error: %v", err)
	}
	if ok.Error != nil {
		t.Fatalf("success response has error: %+v", ok.Error)
	}
	if len(ok.Result) == 0 {
		t.Fatal("success response result is empty")
	}

	fail := NewJSONRPCError(mustJSONRPCStringID(t, "req_2"), "", JSONRPCInvalidParams, "Invalid params", &AgentDErrorData{
		AgentDCode: "INVALID_CURSOR",
		Retryable:  false,
	})
	if fail.Result != nil {
		t.Fatalf("error response has result: %s", string(fail.Result))
	}
	if fail.Error == nil || fail.Error.Data == nil {
		t.Fatalf("error response missing error data: %+v", fail)
	}
	if fail.Error.Data.AgentDCode != "INVALID_CURSOR" {
		t.Fatalf("AgentDCode = %q, want INVALID_CURSOR", fail.Error.Data.AgentDCode)
	}
	raw, err := json.Marshal(fail)
	if err != nil {
		t.Fatalf("marshal error response: %v", err)
	}
	if !regexp.MustCompile(`"retryable":false`).Match(raw) {
		t.Fatalf("error response omitted explicit retryable=false: %s", string(raw))
	}
}

func TestJSONRPCServerBusyCodeIsServerError(t *testing.T) {
	if JSONRPCServerBusy < -32099 || JSONRPCServerBusy > -32000 {
		t.Fatalf("JSONRPCServerBusy = %d, want reserved server-error range", JSONRPCServerBusy)
	}
}

func TestJSONRPCErrorUnknownIDMarshalsNullID(t *testing.T) {
	fail := NewJSONRPCError(nil, "", JSONRPCInvalidRequest, "Invalid request", nil)
	raw, err := json.Marshal(fail)
	if err != nil {
		t.Fatalf("marshal error response: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if _, ok := fields["id"]; !ok {
		t.Fatalf("response omitted id: %s", string(raw))
	}
	if fields["id"] != nil {
		t.Fatalf("id = %v, want null in %s", fields["id"], string(raw))
	}
}

func TestProtocolHelloNegotiatesDefaultV2(t *testing.T) {
	result, err := NegotiateProtocolHello(ProtocolHelloParams{})
	if err != nil {
		t.Fatalf("NegotiateProtocolHello returned error: %v", err)
	}
	if result.SelectedProtocol != ProtocolV2 {
		t.Fatalf("SelectedProtocol = %q, want %q", result.SelectedProtocol, ProtocolV2)
	}
	if !containsStringValue(result.Capabilities, CapabilityTypedEvents) {
		t.Fatalf("v2 capabilities missing typed.events: %v", result.Capabilities)
	}
	if !containsStringValue(result.Capabilities, CapabilitySupportBundle) {
		t.Fatalf("v2 capabilities missing support.bundle: %v", result.Capabilities)
	}
}

func TestProtocolHelloNegotiatesAgainstExplicitServerOffer(t *testing.T) {
	result, err := NegotiateProtocolHelloForOffer(ProtocolHelloParams{
		SupportedProtocols: []string{ProtocolV2, ProtocolV1},
	}, ProtocolHelloOffer{
		SupportedProtocols: []string{ProtocolV1},
		CapabilitiesByProtocol: map[string][]string{
			ProtocolV1: {CapabilityJSONRPCCommands, CapabilityTransportPing},
		},
	})
	if err != nil {
		t.Fatalf("NegotiateProtocolHelloForOffer returned error: %v", err)
	}
	if result.SelectedProtocol != ProtocolV1 {
		t.Fatalf("SelectedProtocol = %q, want %q", result.SelectedProtocol, ProtocolV1)
	}
	if containsStringValue(result.Capabilities, CapabilityTypedEvents) {
		t.Fatalf("compatibility offer must not advertise typed.events: %v", result.Capabilities)
	}
	if !containsStringValue(result.Capabilities, CapabilityJSONRPCCommands) {
		t.Fatalf("compatibility offer missing JSON-RPC command capability: %v", result.Capabilities)
	}
}

func TestProtocolHelloOfferRejectsV2OnlyClientWhenServerDoesNotOfferV2(t *testing.T) {
	_, err := NegotiateProtocolHelloForOffer(ProtocolHelloParams{
		SupportedProtocols: []string{ProtocolV2},
	}, ProtocolHelloOffer{
		SupportedProtocols: []string{ProtocolV1},
		CapabilitiesByProtocol: map[string][]string{
			ProtocolV1: {CapabilityTransportPing},
		},
	})
	if err == nil {
		t.Fatal("NegotiateProtocolHelloForOffer returned nil error for v2-only client")
	}
}

func TestProtocolHelloOfferMarshalsEmptyCapabilitiesAsArray(t *testing.T) {
	result, err := NegotiateProtocolHelloForOffer(ProtocolHelloParams{}, ProtocolHelloOffer{
		SupportedProtocols: []string{ProtocolV1},
	})
	if err != nil {
		t.Fatalf("NegotiateProtocolHelloForOffer returned error: %v", err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if !regexp.MustCompile(`"capabilities":\[\]`).Match(raw) {
		t.Fatalf("empty capabilities must marshal as an array: %s", string(raw))
	}
}

func TestProtocolHelloNegotiatesV1WhenClientDoesNotSupportV2(t *testing.T) {
	result, err := NegotiateProtocolHello(ProtocolHelloParams{
		SupportedProtocols: []string{ProtocolV1},
	})
	if err != nil {
		t.Fatalf("NegotiateProtocolHello returned error: %v", err)
	}
	if result.Protocol != ProtocolV2 {
		t.Fatalf("Protocol = %q, want %q", result.Protocol, ProtocolV2)
	}
	if result.SelectedProtocol != ProtocolV1 {
		t.Fatalf("SelectedProtocol = %q, want %q", result.SelectedProtocol, ProtocolV1)
	}
	if !containsStringValue(result.SupportedProtocols, ProtocolV2) || !containsStringValue(result.SupportedProtocols, ProtocolV1) {
		t.Fatalf("SupportedProtocols = %v, want both v2 and v1", result.SupportedProtocols)
	}
	if !sameStringSet(result.Capabilities, []string{CapabilityTransportPing}) {
		t.Fatalf("v1 capabilities = %v, want transport.ping only", result.Capabilities)
	}
}

func TestSupportBundleParamsNormalizeDefaultsCapsAndZero(t *testing.T) {
	normalized, err := NormalizeSupportBundleParams(SupportBundleParams{})
	if err != nil {
		t.Fatalf("NormalizeSupportBundleParams defaults error: %v", err)
	}
	if normalized.EventLimit != SupportBundleEventLimit || !normalized.IncludeRecentEvents {
		t.Fatalf("defaults = %+v, want limit %d and include events", normalized, SupportBundleEventLimit)
	}

	zero := 0
	include := false
	normalized, err = NormalizeSupportBundleParams(SupportBundleParams{
		EventLimit:          &zero,
		IncludeRecentEvents: &include,
	})
	if err != nil {
		t.Fatalf("NormalizeSupportBundleParams zero error: %v", err)
	}
	if normalized.EventLimit != 0 || normalized.IncludeRecentEvents {
		t.Fatalf("zero/include=false = %+v", normalized)
	}

	tooHigh := SupportBundleMaxEventLimit + 50
	normalized, err = NormalizeSupportBundleParams(SupportBundleParams{EventLimit: &tooHigh})
	if err != nil {
		t.Fatalf("NormalizeSupportBundleParams cap error: %v", err)
	}
	if normalized.EventLimit != SupportBundleMaxEventLimit {
		t.Fatalf("capped EventLimit = %d, want %d", normalized.EventLimit, SupportBundleMaxEventLimit)
	}
}

func TestSupportBundleParamsRejectNegativeLimit(t *testing.T) {
	negative := -1
	if _, err := NormalizeSupportBundleParams(SupportBundleParams{EventLimit: &negative}); err == nil {
		t.Fatal("NormalizeSupportBundleParams accepted negative event_limit")
	}
}

func TestSupportBundleMarshalKeepsRequiredListsAsArrays(t *testing.T) {
	raw, err := json.Marshal(SupportBundle{
		SchemaVersion:     SupportBundleSchemaVersion,
		GeneratedAtUnixMS: 123,
		BundleID:          "sb_123_1",
		Transport: SupportBundleTransport{
			ActiveMode: "local",
			Connected:  true,
		},
		Redaction: SupportRedaction{
			Applied:      true,
			RulesVersion: 1,
		},
	})
	if err != nil {
		t.Fatalf("Marshal support bundle: %v", err)
	}
	if !strings.Contains(string(raw), `"recent_events":[]`) {
		t.Fatalf("recent_events did not encode as empty array: %s", string(raw))
	}
	if !strings.Contains(string(raw), `"last_errors":[]`) {
		t.Fatalf("last_errors did not encode as empty array: %s", string(raw))
	}
}

func TestProtocolHelloRejectsUnsupportedProtocols(t *testing.T) {
	_, err := NegotiateProtocolHello(ProtocolHelloParams{
		SupportedProtocols: []string{"example.future"},
	})
	if err == nil {
		t.Fatal("NegotiateProtocolHello returned nil error for unsupported protocols")
	}
}

func TestV2FixturesParse(t *testing.T) {
	fixtures := map[string]any{
		"jsonrpc-hello-request.json":  &JSONRPCRequest{},
		"jsonrpc-hello-response.json": &JSONRPCResponse{},
		"agentd-event.json":           &AgentDEventEnvelope{},
		"relay-envelope.json":         &RelayEnvelope{},
	}

	for name, dst := range fixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("schemas", "fixtures", name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if err := json.Unmarshal(raw, dst); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}
		})
	}
}

func TestProtocolHelloFixtureMatchesDefaultOffer(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("schemas", "fixtures", "jsonrpc-hello-response.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	var fixture ProtocolHelloResult
	if err := json.Unmarshal(resp.Result, &fixture); err != nil {
		t.Fatalf("unmarshal fixture result: %v", err)
	}
	defaultResult, err := NegotiateProtocolHello(ProtocolHelloParams{
		SupportedProtocols: []string{ProtocolV2, ProtocolV1},
	})
	if err != nil {
		t.Fatalf("NegotiateProtocolHello returned error: %v", err)
	}
	if !sameStringSet(fixture.Capabilities, defaultResult.Capabilities) {
		t.Fatalf("fixture capabilities = %v, want %v", fixture.Capabilities, defaultResult.Capabilities)
	}
}

func TestV2SchemasMatchTraceAndEventRules(t *testing.T) {
	eventSchema := readV2Schema(t, "agentd-event-envelope.schema.json")
	eventTypePattern := schemaProperty(t, eventSchema, "type")["pattern"].(string)
	if !regexp.MustCompile(eventTypePattern).MatchString(string(EventError)) {
		t.Fatalf("event type pattern %q rejects exported event %q", eventTypePattern, EventError)
	}

	for _, tc := range []struct {
		file string
		prop string
	}{
		{file: "agentd-command-envelope.schema.json", prop: "trace_id"},
		{file: "agentd-command-response.schema.json", prop: "trace_id"},
		{file: "agentd-event-envelope.schema.json", prop: "trace_id"},
		{file: "relay-envelope.schema.json", prop: "tid"},
	} {
		t.Run(tc.file+"/"+tc.prop, func(t *testing.T) {
			schema := readV2Schema(t, tc.file)
			prop := schemaProperty(t, schema, tc.prop)
			not, ok := prop["not"].(map[string]any)
			if !ok {
				t.Fatalf("%s %s is missing not clause", tc.file, tc.prop)
			}
			if got := not["const"]; got != "00000000000000000000000000000000" {
				t.Fatalf("%s %s zero trace const = %v", tc.file, tc.prop, got)
			}
		})
	}
}

func TestRelayEnvelopeSchemaAllowsZeroSequence(t *testing.T) {
	schema := readV2Schema(t, "relay-envelope.schema.json")
	seq := schemaProperty(t, schema, "seq")
	if got := seq["minimum"]; got != float64(0) {
		t.Fatalf("relay envelope seq minimum = %v, want 0", got)
	}
}

func TestV2OpenRPCDocumentsCursorParams(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("schemas", "openrpc", "agentd.commands.json"))
	if err != nil {
		t.Fatalf("read OpenRPC contract: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal OpenRPC contract: %v", err)
	}
	for _, method := range []Method{MethodSessionSnapshot, MethodSessionReplay} {
		t.Run(string(method), func(t *testing.T) {
			methodDoc := openRPCMethod(t, doc, string(method))
			if got := methodDoc["paramStructure"]; got != "by-name" {
				t.Fatalf("%s paramStructure = %v, want by-name", method, got)
			}
			params := openRPCParams(t, methodDoc, string(method))
			sessionID := openRPCParam(t, params, "session_id")
			if required, ok := sessionID["required"].(bool); !ok || !required {
				t.Fatalf("%s session_id param required = %v, want true", method, sessionID["required"])
			}
			sessionIDSchema, ok := sessionID["schema"].(map[string]any)
			if !ok {
				t.Fatalf("%s session_id schema missing", method)
			}
			if got := sessionIDSchema["minLength"]; got != float64(1) {
				t.Fatalf("%s session_id minLength = %v, want 1", method, got)
			}
			afterSeq := openRPCParam(t, params, "after_seq")
			afterSeqSchema, ok := afterSeq["schema"].(map[string]any)
			if !ok {
				t.Fatalf("%s after_seq schema missing", method)
			}
			if got := afterSeqSchema["minimum"]; got != float64(0) {
				t.Fatalf("%s after_seq minimum = %v, want 0", method, got)
			}
			openRPCParam(t, params, "cursor")
		})
	}
}

func TestV2OpenRPCDocumentsProtocolHelloNegotiation(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("schemas", "openrpc", "agentd.commands.json"))
	if err != nil {
		t.Fatalf("read OpenRPC contract: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal OpenRPC contract: %v", err)
	}
	method := openRPCMethod(t, doc, string(MethodProtocolHello))
	if got := method["paramStructure"]; got != "by-name" {
		t.Fatalf("protocol.hello paramStructure = %v, want by-name", got)
	}
	params := openRPCParams(t, method, string(MethodProtocolHello))
	openRPCParam(t, params, "client")
	openRPCParam(t, params, "supported_protocols")
	openRPCParam(t, params, "capabilities")

	resultSchema := openRPCResultSchemaFromMethod(t, method, string(MethodProtocolHello))
	props, ok := resultSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("protocol.hello result schema missing properties")
	}
	selected, ok := props["selected_protocol"].(map[string]any)
	if !ok {
		t.Fatal("protocol.hello result schema missing selected_protocol")
	}
	enum, ok := selected["enum"].([]any)
	if !ok {
		t.Fatal("selected_protocol schema missing enum")
	}
	if !containsString(enum, ProtocolV2) || !containsString(enum, ProtocolV1) {
		t.Fatalf("selected_protocol enum = %v, want v2 and v1", enum)
	}
}

func readV2Schema(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("schemas", "json-schema", name))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	return schema
}

func schemaProperty(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties missing")
	}
	prop, ok := props[name].(map[string]any)
	if !ok {
		t.Fatalf("schema property %q missing", name)
	}
	return prop
}

func openRPCMethod(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()
	methods, ok := doc["methods"].([]any)
	if !ok {
		t.Fatalf("OpenRPC methods missing")
	}
	for _, rawMethod := range methods {
		method, ok := rawMethod.(map[string]any)
		if !ok || method["name"] != name {
			continue
		}
		return method
	}
	t.Fatalf("OpenRPC method %q missing", name)
	return nil
}

func openRPCParams(t *testing.T, method map[string]any, name string) []any {
	t.Helper()
	params, ok := method["params"].([]any)
	if !ok || len(params) == 0 {
		t.Fatalf("%s params missing", name)
	}
	return params
}

func openRPCParam(t *testing.T, params []any, want string) map[string]any {
	t.Helper()
	for _, rawParam := range params {
		param, ok := rawParam.(map[string]any)
		if !ok || param["name"] != want {
			continue
		}
		return param
	}
	t.Fatalf("OpenRPC param %q missing", want)
	return nil
}

func openRPCResultSchemaFromMethod(t *testing.T, method map[string]any, name string) map[string]any {
	t.Helper()
	result, ok := method["result"].(map[string]any)
	if !ok {
		t.Fatalf("%s result missing", name)
	}
	schema, ok := result["schema"].(map[string]any)
	if !ok {
		t.Fatalf("%s result schema missing", name)
	}
	return schema
}

func containsString(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsStringValue(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, value := range a {
		seen[value]++
	}
	for _, value := range b {
		seen[value]--
		if seen[value] < 0 {
			return false
		}
	}
	return true
}

func mustJSONRPCStringID(t *testing.T, value string) JSONRPCID {
	t.Helper()
	id, err := NewJSONRPCStringID(value)
	if err != nil {
		t.Fatalf("NewJSONRPCStringID: %v", err)
	}
	return id
}
