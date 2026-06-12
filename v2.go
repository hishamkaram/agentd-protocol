package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// ProtocolV1 names the existing AgentD transport contract before the v2
	// JSON-RPC command plane and typed event stream are selected.
	ProtocolV1 = "agentd.v1"

	// ProtocolV2 names the dual-stack protocol negotiated between PWA and daemon.
	ProtocolV2 = "agentd.v2"

	// AgentDEventProtocolV2 identifies typed streaming event frames. These are
	// intentionally not JSON-RPC notifications; replay and session output need a
	// durable event contract with ordering metadata.
	AgentDEventProtocolV2 = "agentd.event.v2"

	JSONRPCVersion = "2.0"
)

const (
	MethodProtocolHello Method = "protocol.hello"
	MethodTransportPing Method = "transport.ping"
	MethodSupportBundle Method = "support.bundle"

	MethodSessionList      Method = "session.list"
	MethodSessionStart     Method = "session.start"
	MethodSessionStop      Method = "session.stop"
	MethodSessionInterrupt Method = "session.interrupt"
	MethodSessionResume    Method = "session.resume"
	MethodSessionDelete    Method = "session.delete"
	MethodSessionRename    Method = "session.rename"
	MethodSessionFork      Method = "session.fork"

	MethodApprovalDecide  Method = "approval.decide"
	MethodApprovalSetMode Method = "approval.setMode"
	MethodPromptAnswer    Method = "prompt.answer"
	MethodBudgetExtend    Method = "budget.extend"

	MethodModelsList    Method = "models.list"
	MethodProvidersList Method = "providers.list"
)

const (
	CapabilityJSONRPCCommands      = "jsonrpc.commands"
	CapabilityTypedEvents          = "typed.events"
	CapabilityTransportPing        = "transport.ping"
	CapabilitySupportBundle        = "support.bundle"
	CapabilitySessionList          = "session.list"
	CapabilitySessionFeatureStatus = "session.feature_status"
	CapabilityRelayOpaque          = "relay.opaque"
	CapabilityCommandReceipts      = "command.receipts"
)

const (
	EventTransportConnected    EventType = "transport.connected"
	EventTransportDisconnected EventType = "transport.disconnected"
	EventTransportPong         EventType = "transport.pong"
	EventSessionListUpdated    EventType = "session.list.updated"
	EventSessionFeatureStatus  EventType = "session.feature_status"
	EventSessionOutputDelta    EventType = "session.output.delta"
	EventThinkingDelta         EventType = "session.thinking.delta"
	EventApprovalRequested     EventType = "approval.requested"
	EventApprovalResolved      EventType = "approval.resolved"
	EventPromptRequested       EventType = "prompt.requested"
	EventCommandReceipt        EventType = "command.receipt"
	EventError                 EventType = "error"
)

const (
	JSONRPCParseError     = -32700
	JSONRPCInvalidRequest = -32600
	JSONRPCMethodNotFound = -32601
	JSONRPCInvalidParams  = -32602
	JSONRPCInternalError  = -32603
	JSONRPCServerBusy     = -32000
)

type Method string
type EventType string

type ConnectionPhase string

const (
	ConnectionPhaseDisconnected ConnectionPhase = "disconnected"
	ConnectionPhaseConnecting   ConnectionPhase = "connecting"
	ConnectionPhaseNegotiating  ConnectionPhase = "negotiating"
	ConnectionPhaseConnected    ConnectionPhase = "connected"
	ConnectionPhaseStale        ConnectionPhase = "stale"
	ConnectionPhaseReconnecting ConnectionPhase = "reconnecting"
)

type TransportKind string

const (
	TransportLocal TransportKind = "local"
	TransportRelay TransportKind = "relay"
)

// JSONRPCID preserves a JSON-RPC request id in its original JSON form so
// responses can echo string, numeric, or null ids exactly as received.
type JSONRPCID []byte

func (id *JSONRPCID) UnmarshalJSON(data []byte) error {
	if !validJSONRPCID(data) {
		return errors.New("protocol.JSONRPCID: id must be string, number, or null")
	}
	*id = append((*id)[:0], data...)
	return nil
}

func (id JSONRPCID) MarshalJSON() ([]byte, error) {
	if id.Missing() {
		return []byte("null"), nil
	}
	return append([]byte(nil), id...), nil
}

func (id JSONRPCID) Missing() bool {
	return len(id) == 0
}

func (id JSONRPCID) String() string {
	return string(id)
}

func NullJSONRPCID() JSONRPCID {
	return JSONRPCID("null")
}

func NewJSONRPCStringID(value string) (JSONRPCID, error) {
	if value == "" {
		return nil, errors.New("protocol.NewJSONRPCStringID: id is required")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("protocol.NewJSONRPCStringID: marshal: %w", err)
	}
	return JSONRPCID(raw), nil
}

func validJSONRPCID(data []byte) bool {
	switch string(data) {
	case "", "true", "false":
		return false
	case "null":
		return true
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return s != ""
	}
	var n json.Number
	return json.Unmarshal(data, &n) == nil && n.String() != ""
}

// JSONRPCRequest is AgentD's command-plane request envelope.
//
// TraceID is an AgentD extension to the JSON-RPC object. It is correlation
// metadata only and never authorizes a request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      JSONRPCID       `json:"id,omitempty"`
	Method  Method          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	TraceID string          `json:"trace_id,omitempty"`
}

// JSONRPCResponse is AgentD's command-plane response envelope. Exactly one of
// Result or Error is populated.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      JSONRPCID       `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	TraceID string          `json:"trace_id,omitempty"`
}

type JSONRPCError struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *AgentDErrorData `json:"data,omitempty"`
}

type AgentDErrorData struct {
	AgentDCode string          `json:"agentd_code,omitempty"`
	Retryable  bool            `json:"retryable"`
	Details    json.RawMessage `json:"details,omitempty"`
}

// ProtocolHelloParams is sent by clients before opting into v2 behavior.
type ProtocolHelloParams struct {
	Client             string   `json:"client,omitempty"`
	SupportedProtocols []string `json:"supported_protocols,omitempty"`
	Capabilities       []string `json:"capabilities,omitempty"`
}

type ProtocolHelloOffer struct {
	SupportedProtocols     []string
	CapabilitiesByProtocol map[string][]string
}

type ProtocolHelloResult struct {
	Protocol           string   `json:"protocol"`
	SelectedProtocol   string   `json:"selected_protocol"`
	SupportedProtocols []string `json:"supported_protocols"`
	Capabilities       []string `json:"capabilities"`
}

type TransportPingParams struct {
	SentAtUnixMS int64 `json:"sent_at_unix_ms,omitempty"`
}

type TransportPongResult struct {
	Type         string `json:"type"`
	ReceivedAt   int64  `json:"received_at_unix_ms"`
	SentAtUnixMS int64  `json:"sent_at_unix_ms,omitempty"`
}

// AgentDEventEnvelope is the typed v2 streaming envelope used for session
// output, approvals, prompts, and transport events.
type AgentDEventEnvelope struct {
	Protocol  string          `json:"protocol"`
	ID        string          `json:"id"`
	Type      EventType       `json:"type"`
	Seq       uint64          `json:"seq,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	Time      time.Time       `json:"time"`
	Payload   json.RawMessage `json:"payload"`
}

func IsJSONRPCRequest(data []byte) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return false
	}
	if _, ok := fields["jsonrpc"]; !ok {
		return false
	}
	_, hasResult := fields["result"]
	_, hasError := fields["error"]
	return !hasResult && !hasError
}

func DecodeJSONRPCRequest(data []byte) (JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return req, fmt.Errorf("protocol.DecodeJSONRPCRequest: %w", err)
	}
	if req.JSONRPC != JSONRPCVersion {
		return req, errors.New("protocol.DecodeJSONRPCRequest: jsonrpc must be 2.0")
	}
	if req.Method == "" {
		return req, errors.New("protocol.DecodeJSONRPCRequest: method is required")
	}
	if req.ID.Missing() {
		return req, errors.New("protocol.DecodeJSONRPCRequest: id is required")
	}
	if req.TraceID != "" && !ValidTraceID(req.TraceID) {
		return req, errors.New("protocol.DecodeJSONRPCRequest: invalid trace_id")
	}
	return req, nil
}

func NewJSONRPCResult(id JSONRPCID, traceID string, result any) (JSONRPCResponse, error) {
	if id.Missing() {
		return JSONRPCResponse{}, errors.New("protocol.NewJSONRPCResult: id is required")
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return JSONRPCResponse{}, fmt.Errorf("protocol.NewJSONRPCResult: marshal result: %w", err)
	}
	return JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  raw,
		TraceID: traceID,
	}, nil
}

func NewJSONRPCError(id JSONRPCID, traceID string, code int, message string, data *AgentDErrorData) JSONRPCResponse {
	if id.Missing() {
		id = NullJSONRPCID()
	}
	return JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		TraceID: traceID,
	}
}

func NewProtocolHelloResult() ProtocolHelloResult {
	result, _ := NegotiateProtocolHello(ProtocolHelloParams{})
	return result
}

func NegotiateProtocolHello(params ProtocolHelloParams) (ProtocolHelloResult, error) {
	return NegotiateProtocolHelloForOffer(params, FullProtocolHelloOffer())
}

func FullProtocolHelloOffer() ProtocolHelloOffer {
	return ProtocolHelloOffer{
		SupportedProtocols: []string{ProtocolV2, ProtocolV1},
		CapabilitiesByProtocol: map[string][]string{
			ProtocolV2: {
				CapabilityJSONRPCCommands,
				CapabilityTypedEvents,
				CapabilityTransportPing,
				CapabilitySupportBundle,
				CapabilitySessionList,
				CapabilitySessionFeatureStatus,
				CapabilityRelayOpaque,
				CapabilityCommandReceipts,
			},
			ProtocolV1: {CapabilityTransportPing},
		},
	}
}

func NegotiateProtocolHelloForOffer(params ProtocolHelloParams, offer ProtocolHelloOffer) (ProtocolHelloResult, error) {
	supported := append([]string(nil), offer.SupportedProtocols...)
	if len(supported) == 0 {
		supported = []string{ProtocolV2, ProtocolV1}
	}

	selected := supported[0]
	if len(params.SupportedProtocols) > 0 {
		selected = ""
		for _, candidate := range supported {
			if containsProtocol(params.SupportedProtocols, candidate) {
				selected = candidate
				break
			}
		}
		if selected == "" {
			return ProtocolHelloResult{}, errors.New("protocol.NegotiateProtocolHelloForOffer: no common protocol")
		}
	}

	capabilities := append([]string{}, offer.CapabilitiesByProtocol[selected]...)

	return ProtocolHelloResult{
		Protocol:           ProtocolV2,
		SelectedProtocol:   selected,
		SupportedProtocols: supported,
		Capabilities:       capabilities,
	}, nil
}

func containsProtocol(protocols []string, want string) bool {
	for _, protocol := range protocols {
		if protocol == want {
			return true
		}
	}
	return false
}

func NewTransportPongResult(sentAtUnixMS int64) TransportPongResult {
	return TransportPongResult{
		Type:         MsgPong,
		ReceivedAt:   time.Now().UnixMilli(),
		SentAtUnixMS: sentAtUnixMS,
	}
}
