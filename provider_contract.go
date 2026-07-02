package protocol

const ProviderCapabilityContractSchemaVersion = 1

type ProviderCapabilitySupport string

const (
	ProviderCapabilitySupported   ProviderCapabilitySupport = "supported"
	ProviderCapabilityLimited     ProviderCapabilitySupport = "limited"
	ProviderCapabilityUnsupported ProviderCapabilitySupport = "unsupported"
)

type ProviderCapabilitySource string

const (
	ProviderCapabilitySourceSDKRPC         ProviderCapabilitySource = "sdk_rpc"
	ProviderCapabilitySourceSDKInit        ProviderCapabilitySource = "sdk_init"
	ProviderCapabilitySourceCLIDocumented  ProviderCapabilitySource = "cli_documented"
	ProviderCapabilitySourceDaemonCurated  ProviderCapabilitySource = "daemon_curated"
	ProviderCapabilitySourceDaemonConfig   ProviderCapabilitySource = "daemon_config"
	ProviderCapabilitySourcePromptComposed ProviderCapabilitySource = "prompt_composed"
	ProviderCapabilitySourcePWAClient      ProviderCapabilitySource = "pwa_client"
)

type ProviderCommandArgumentRequirement string

const (
	ProviderCommandArgumentNone     ProviderCommandArgumentRequirement = "none"
	ProviderCommandArgumentOptional ProviderCommandArgumentRequirement = "optional"
	ProviderCommandArgumentRequired ProviderCommandArgumentRequirement = "required"
)

type ProviderCommandLifecycle string

const (
	ProviderCommandLifecycleProviderTurn    ProviderCommandLifecycle = "provider_turn"
	ProviderCommandLifecycleProviderControl ProviderCommandLifecycle = "provider_control"
	ProviderCommandLifecycleDaemonControl   ProviderCommandLifecycle = "daemon_control"
	ProviderCommandLifecycleClientLocal     ProviderCommandLifecycle = "client_local"
	ProviderCommandLifecycleNoop            ProviderCommandLifecycle = "noop"
)

type ProviderCommandStatusAfterDispatch string

const (
	ProviderCommandStatusRunning   ProviderCommandStatusAfterDispatch = "running"
	ProviderCommandStatusWaiting   ProviderCommandStatusAfterDispatch = "waiting"
	ProviderCommandStatusUnchanged ProviderCommandStatusAfterDispatch = "unchanged"
)

// ProviderCapabilityContract is the additive v1 provider capability contract
// emitted on SessionInfo.provider_contract. It describes provider facts that
// legacy booleans and UI copy can derive from without changing old fields.
type ProviderCapabilityContract struct {
	SchemaVersion   int                         `json:"schema_version"`
	Commands        []ProviderCommandDescriptor `json:"commands,omitempty"`
	Approval        []ProviderFeatureDescriptor `json:"approval,omitempty"`
	Sandbox         []ProviderFeatureDescriptor `json:"sandbox,omitempty"`
	MCP             []ProviderFeatureDescriptor `json:"mcp,omitempty"`
	Interactive     []ProviderFeatureDescriptor `json:"interactive,omitempty"`
	Skills          []ProviderFeatureDescriptor `json:"skills,omitempty"`
	Model           []ProviderFeatureDescriptor `json:"model,omitempty"`
	RuntimeSettings []ProviderFeatureDescriptor `json:"runtime_settings,omitempty"`
}

type ProviderFeatureDescriptor struct {
	ID              string                    `json:"id"`
	Support         ProviderCapabilitySupport `json:"support"`
	Source          ProviderCapabilitySource  `json:"source"`
	UserLabel       string                    `json:"user_label"`
	UserDescription string                    `json:"user_description"`
	Limitations     []string                  `json:"limitations,omitempty"`
}

type ProviderCommandDescriptor struct {
	Name                string                             `json:"name"`
	ArgumentRequirement ProviderCommandArgumentRequirement `json:"argument_requirement"`
	Lifecycle           ProviderCommandLifecycle           `json:"lifecycle"`
	StatusAfterDispatch ProviderCommandStatusAfterDispatch `json:"status_after_dispatch"`
	Source              ProviderCapabilitySource           `json:"source"`
	Support             ProviderCapabilitySupport          `json:"support"`
	UserDescription     string                             `json:"user_description"`
	Limitations         []string                           `json:"limitations,omitempty"`
}
