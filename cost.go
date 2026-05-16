package protocol

// CostPayload is carried by AgentD "cost" messages. The first six fields are
// the historical wire contract. Later fields are additive and optional so older
// clients and daemons continue to interoperate.
type CostPayload struct {
	InputTokens           int     `json:"input_tokens"`
	OutputTokens          int     `json:"output_tokens"`
	TurnCostUSD           float64 `json:"turn_cost_usd"`
	SessionTotal          float64 `json:"session_total_usd"`
	BudgetLimit           float64 `json:"budget_limit_usd"`  // 0 = no limit
	BudgetRemain          float64 `json:"budget_remain_usd"` // 0 = no limit
	CachedInputTokens     int     `json:"cached_input_tokens,omitempty"`
	ReasoningOutputTokens int     `json:"reasoning_output_tokens,omitempty"`
	PricingModel          string  `json:"pricing_model,omitempty"`
	PricingAvailable      bool    `json:"pricing_available,omitempty"`
	TurnCredits           float64 `json:"turn_credits,omitempty"`
	SessionTotalCredits   float64 `json:"session_total_credits,omitempty"`
}
