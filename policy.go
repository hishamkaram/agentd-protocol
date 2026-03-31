package protocol

// PolicyJSON represents a single policy rule in the relay wire format.
type PolicyJSON struct {
	Name      string          `json:"name"`
	Match     PolicyMatchJSON `json:"match"`
	Action    string          `json:"action"`
	Approvers []string        `json:"approvers,omitempty"`
	TimeoutS  int             `json:"timeout_s,omitempty"`
	Reason    string          `json:"reason,omitempty"`
}

// PolicyMatchJSON represents the match criteria for a policy rule.
type PolicyMatchJSON struct {
	Tool        []string `json:"tool,omitempty"`
	FilePattern string   `json:"file_pattern,omitempty"`
	Branch      string   `json:"branch,omitempty"`
	Command     string   `json:"command,omitempty"`
	RiskLevel   []string `json:"risk_level,omitempty"`
}
