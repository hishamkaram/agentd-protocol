package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	SupportBundleSchemaVersion = "agentd.support_bundle.v1"
	SupportBundleEventLimit    = 50
	SupportBundleMaxEventLimit = 200
)

type SupportBundleParams struct {
	EventLimit          *int                    `json:"event_limit,omitempty"`
	IncludeRecentEvents *bool                   `json:"include_recent_events,omitempty"`
	ClientTransport     *SupportBundleTransport `json:"client_transport,omitempty"`
}

type NormalizedSupportBundleParams struct {
	EventLimit          int
	IncludeRecentEvents bool
}

func NormalizeSupportBundleParams(params SupportBundleParams) (NormalizedSupportBundleParams, error) {
	normalized := NormalizedSupportBundleParams{
		EventLimit:          SupportBundleEventLimit,
		IncludeRecentEvents: true,
	}
	if params.EventLimit != nil {
		if *params.EventLimit < 0 {
			return NormalizedSupportBundleParams{}, errors.New("event_limit must be >= 0")
		}
		normalized.EventLimit = *params.EventLimit
		if normalized.EventLimit > SupportBundleMaxEventLimit {
			normalized.EventLimit = SupportBundleMaxEventLimit
		}
	}
	if params.IncludeRecentEvents != nil {
		normalized.IncludeRecentEvents = *params.IncludeRecentEvents
	}
	return normalized, nil
}

type SupportBundleContext struct {
	TransportKind              TransportKind
	Connected                  bool
	Stale                      bool
	SelectedProtocol           string
	Capabilities               []string
	ActiveSessionID            string
	ActiveSessionState         string
	ActiveSessionStartupPhase  string
	HistoryHydrationState      SupportHistoryHydrationState
	LastRelayControlError      string
	LastInboundAgeMS           *int64
	LastPongAgeMS              *int64
	PendingJSONRPCRequestCount int
	ClientDiagnostics          *SupportClientDiagnostics
}

type SupportBundle struct {
	SchemaVersion     string                 `json:"schema_version"`
	GeneratedAtUnixMS int64                  `json:"generated_at_unix_ms"`
	BundleID          string                 `json:"bundle_id"`
	Client            *SupportBundleClient   `json:"client,omitempty"`
	Daemon            *SupportBundleDaemon   `json:"daemon,omitempty"`
	Transport         SupportBundleTransport `json:"transport"`
	RecentEvents      []SupportEventEntry    `json:"recent_events"`
	LastErrors        []SupportErrorRef      `json:"last_errors"`
	Redaction         SupportRedaction       `json:"redaction"`
	Partial           bool                   `json:"partial"`
	PartialReasons    []SupportPartialReason `json:"partial_reasons,omitempty"`
}

func (b SupportBundle) MarshalJSON() ([]byte, error) {
	type wire SupportBundle
	if b.RecentEvents == nil {
		b.RecentEvents = []SupportEventEntry{}
	}
	if b.LastErrors == nil {
		b.LastErrors = []SupportErrorRef{}
	}
	return json.Marshal(wire(b))
}

type SupportBundleClient struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SupportBundleDaemon struct {
	Version                string                     `json:"version"`
	ProtocolVersion        string                     `json:"protocol_version"`
	AdvertisedCapabilities []string                   `json:"advertised_capabilities"`
	ConfigSource           SupportConfigSourceSummary `json:"config_source"`
	RelayURLCategory       SupportRelayURLCategory    `json:"relay_url_category"`
}

type SupportConfigSourceKind string

const (
	SupportConfigSourceFile        SupportConfigSourceKind = "file"
	SupportConfigSourceDefaultsEnv SupportConfigSourceKind = "defaults_env"
	SupportConfigSourceUnknown     SupportConfigSourceKind = "unknown"
)

type SupportConfigSourceSummary struct {
	Kind            SupportConfigSourceKind `json:"kind"`
	Basename        string                  `json:"basename,omitempty"`
	PathHash        string                  `json:"path_hash,omitempty"`
	UnknownKeyCount int                     `json:"unknown_key_count"`
	UnknownKeys     []string                `json:"unknown_keys,omitempty"`
}

type SupportRelayURLCategory string

const (
	SupportRelayURLNone           SupportRelayURLCategory = "none"
	SupportRelayURLLocalhost      SupportRelayURLCategory = "localhost"
	SupportRelayURLPrivateNetwork SupportRelayURLCategory = "private_network"
	SupportRelayURLHostedAgentD   SupportRelayURLCategory = "hosted_agentd"
	SupportRelayURLCustomDomain   SupportRelayURLCategory = "custom_domain"
	SupportRelayURLInvalid        SupportRelayURLCategory = "invalid"
)

type SupportBundleTransport struct {
	ActiveMode                string                       `json:"active_mode"`
	Connected                 bool                         `json:"connected"`
	Stale                     bool                         `json:"stale"`
	SelectedProtocol          string                       `json:"selected_protocol,omitempty"`
	RelayURLCategory          SupportRelayURLCategory      `json:"relay_url_category,omitempty"`
	ActiveSessionID           string                       `json:"active_session_id,omitempty"`
	ActiveSessionState        string                       `json:"active_session_state,omitempty"`
	ActiveSessionStartupPhase string                       `json:"active_session_startup_phase,omitempty"`
	HistoryHydrationState     SupportHistoryHydrationState `json:"history_hydration_state,omitempty"`
	LastRelayControlError     string                       `json:"last_relay_control_error,omitempty"`
	RelayDiagnostics          *SupportRelayDiagnostics     `json:"relay_diagnostics,omitempty"`
	ClientDiagnostics         *SupportClientDiagnostics    `json:"client_diagnostics,omitempty"`
	LastInboundAgeMS          *int64                       `json:"last_inbound_age_ms,omitempty"`
	LastPongAgeMS             *int64                       `json:"last_pong_age_ms,omitempty"`
	PendingJSONRPCRequest     int                          `json:"pending_jsonrpc_request_count,omitempty"`
}

type SupportClientDiagnostics struct {
	OutboundCommandCounts       map[string]uint64 `json:"outbound_command_counts,omitempty"`
	ChatViewRenderCount         uint64            `json:"chat_view_render_count,omitempty"`
	TranscriptScrollLoadCount   uint64            `json:"transcript_scroll_load_count,omitempty"`
	HistoryPageRequestsInFlight uint64            `json:"history_page_requests_in_flight,omitempty"`
	WebSocketBufferedAmount     uint64            `json:"websocket_buffered_amount,omitempty"`
}

type SupportRelayDiagnostics struct {
	RelaySessionIDFingerprint string `json:"relay_session_id_fingerprint,omitempty"`
	RegistrationGeneration    uint64 `json:"registration_generation,omitempty"`
	ReconnectCount            uint64 `json:"reconnect_count,omitempty"`
	LastControlCode           string `json:"last_control_code,omitempty"`
	ActiveBootstrapDrains        uint64 `json:"active_bootstrap_drains,omitempty"`
	RelayClientCount          uint64 `json:"relay_client_count,omitempty"`
	QueueRejections           uint64 `json:"queue_rejections,omitempty"`
	SerializerDrops           uint64 `json:"serializer_drops,omitempty"`
}

type SupportHistoryHydrationState string

const (
	SupportHistoryHydrationWaitingTransport SupportHistoryHydrationState = "waiting_transport"
	SupportHistoryHydrationWaitingHistory   SupportHistoryHydrationState = "waiting_history"
	SupportHistoryHydrationRestored         SupportHistoryHydrationState = "restored"
	SupportHistoryHydrationIncomplete       SupportHistoryHydrationState = "incomplete_retryable"
	SupportHistoryHydrationFailed           SupportHistoryHydrationState = "failed"
)

type SupportEventEntry struct {
	TS        int64  `json:"ts"`
	Source    string `json:"source"`
	Event     string `json:"event"`
	Code      string `json:"code,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
}

type SupportErrorRef struct {
	Source    string `json:"source"`
	Code      string `json:"code"`
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	TS        int64  `json:"ts"`
	Retryable bool   `json:"retryable,omitempty"`
}

type SupportRedaction struct {
	Applied      bool `json:"applied"`
	RulesVersion int  `json:"rules_version"`
}

type SupportPartialReason struct {
	Source    string `json:"source"`
	Code      string `json:"code"`
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	TS        int64  `json:"ts"`
	Retryable bool   `json:"retryable,omitempty"`
}

func NewSupportBundleID(unixMS int64, counter uint64) string {
	return fmt.Sprintf("sb_%d_%06d", unixMS, counter%1000000)
}
