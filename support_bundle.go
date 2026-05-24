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
	EventLimit          *int  `json:"event_limit,omitempty"`
	IncludeRecentEvents *bool `json:"include_recent_events,omitempty"`
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
	TransportKind    TransportKind
	Connected        bool
	Stale            bool
	SelectedProtocol string
	Capabilities     []string
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
	ActiveMode            string                  `json:"active_mode"`
	Connected             bool                    `json:"connected"`
	Stale                 bool                    `json:"stale"`
	SelectedProtocol      string                  `json:"selected_protocol,omitempty"`
	RelayURLCategory      SupportRelayURLCategory `json:"relay_url_category,omitempty"`
	LastInboundAgeMS      *int64                  `json:"last_inbound_age_ms,omitempty"`
	LastPongAgeMS         *int64                  `json:"last_pong_age_ms,omitempty"`
	PendingJSONRPCRequest int                     `json:"pending_jsonrpc_request_count,omitempty"`
}

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
