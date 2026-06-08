package protocol

// AgentExtensionKind classifies one provider-exposed extension surface.
type AgentExtensionKind string

const (
	AgentExtensionKindSlashCommand AgentExtensionKind = "slash_command"
	AgentExtensionKindSkill        AgentExtensionKind = "skill"
)

// AgentExtensionScope is the display scope used by operator consoles.
type AgentExtensionScope string

const (
	AgentExtensionScopeProject  AgentExtensionScope = "project"
	AgentExtensionScopeGlobal   AgentExtensionScope = "global"
	AgentExtensionScopeBuiltIn  AgentExtensionScope = "built_in"
	AgentExtensionScopeInternal AgentExtensionScope = "internal"
)

// AgentExtensionInvocation describes how a selected extension is invoked.
type AgentExtensionInvocation string

const (
	AgentExtensionInvocationSlashText AgentExtensionInvocation = "slash_text"
	AgentExtensionInvocationSkillID   AgentExtensionInvocation = "skill_id"
)

// AgentExtensionDef is additive UI metadata for a session extension. IDs and
// RefID are opaque; local provider paths must never be placed on this wire type.
type AgentExtensionDef struct {
	ID               string                   `json:"id"`
	RefID            string                   `json:"ref_id,omitempty"`
	Kind             AgentExtensionKind       `json:"kind"`
	Invocation       AgentExtensionInvocation `json:"invocation"`
	Scope            AgentExtensionScope      `json:"scope,omitempty"`
	Name             string                   `json:"name"`
	DisplayName      string                   `json:"display_name,omitempty"`
	Description      string                   `json:"description,omitempty"`
	ShortDescription string                   `json:"short_description,omitempty"`
	AcceptsArguments bool                     `json:"accepts_arguments,omitempty"`
	ArgumentHint     string                   `json:"argument_hint,omitempty"`
	DefaultPrompt    string                   `json:"default_prompt,omitempty"`
	DependencyCount  int                      `json:"dependency_count,omitempty"`
}
