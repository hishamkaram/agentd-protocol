package protocol

const (
	MsgListProviderCatalogs = "list_provider_catalogs"
	MsgProviderCatalogList  = "provider_catalog_list"
)

type ProviderCatalogState string

const (
	ProviderCatalogLoading     ProviderCatalogState = "loading"
	ProviderCatalogReady       ProviderCatalogState = "ready"
	ProviderCatalogUnavailable ProviderCatalogState = "unavailable"
)

type ProviderCatalogSource string

const (
	ProviderCatalogSourceSDK ProviderCatalogSource = "sdk"
)

type ProviderCatalogErrorCode string

const (
	ProviderCatalogErrorAuthRequired   ProviderCatalogErrorCode = "auth_required"
	ProviderCatalogErrorCLINotFound    ProviderCatalogErrorCode = "cli_not_found"
	ProviderCatalogErrorUnsupportedCLI ProviderCatalogErrorCode = "unsupported_cli"
	ProviderCatalogErrorTimeout        ProviderCatalogErrorCode = "timeout"
	ProviderCatalogErrorProvider       ProviderCatalogErrorCode = "provider_error"
)

type ProviderCatalogError struct {
	Code      ProviderCatalogErrorCode `json:"code"`
	Message   string                   `json:"message"`
	Retryable bool                     `json:"retryable"`
}

// ProviderControlOption exposes a stable canonical value to clients while
// retaining the provider value only for daemon-side launch translation.
type ProviderControlOption struct {
	Value         string `json:"value,omitempty"`
	ProviderValue string `json:"-"`
	DisplayName   string `json:"display_name,omitempty"`
}

type ProviderModelInfo struct {
	Value                    string   `json:"value"`
	ResolvedModel            string   `json:"resolved_model,omitempty"`
	DisplayName              string   `json:"display_name"`
	Description              string   `json:"description,omitempty"`
	SupportsEffort           bool     `json:"supports_effort,omitempty"`
	SupportedEffortLevels    []string `json:"supported_effort_levels,omitempty"`
	DefaultEffort            string   `json:"default_effort,omitempty"`
	SupportsAdaptiveThinking bool     `json:"supports_adaptive_thinking,omitempty"`
	SupportsFastMode         bool     `json:"supports_fast_mode,omitempty"`
	SupportsAutoMode         bool     `json:"supports_auto_mode,omitempty"`
	InputModalities          []string `json:"input_modalities,omitempty"`
	SupportsPersonality      bool     `json:"supports_personality,omitempty"`
}

type ProviderRuntimeCatalog struct {
	Agent         string                  `json:"agent"`
	Provider      string                  `json:"provider"`
	State         ProviderCatalogState    `json:"state"`
	Generation    string                  `json:"generation,omitempty"`
	Source        ProviderCatalogSource   `json:"source,omitempty"`
	CLIVersion    string                  `json:"cli_version,omitempty"`
	DiscoveredAt  int64                   `json:"discovered_at,omitempty"`
	Models        []ProviderModelInfo     `json:"models,omitempty"`
	ApprovalModes []ProviderControlOption `json:"approval_modes,omitempty"`
	SandboxModes  []ProviderControlOption `json:"sandbox_modes,omitempty"`
	Error         *ProviderCatalogError   `json:"error,omitempty"`
}

type ListProviderCatalogsRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
}

type ListProviderCatalogsResponse struct {
	Type      string                   `json:"type"`
	RequestID string                   `json:"request_id,omitempty"`
	Catalogs  []ProviderRuntimeCatalog `json:"catalogs"`
}
