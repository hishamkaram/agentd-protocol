package protocol_test

import (
	"encoding/json"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestCostPayloadBackwardCompatibleRoundTrip(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"input_tokens":100,"output_tokens":50,"turn_cost_usd":0.01,"session_total_usd":0.02,"budget_limit_usd":5,"budget_remain_usd":4.98}`)
	var got protocol.CostPayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal old payload: %v", err)
	}
	if got.InputTokens != 100 || got.OutputTokens != 50 || got.TurnCostUSD != 0.01 ||
		got.SessionTotal != 0.02 || got.BudgetLimit != 5 || got.BudgetRemain != 4.98 {
		t.Fatalf("old fields mismatch: %+v", got)
	}
	if got.CachedInputTokens != 0 || got.PricingAvailable || got.TurnCredits != 0 {
		t.Fatalf("new fields should default to zero values: %+v", got)
	}
}

func TestCostPayloadAdditiveFieldsRoundTrip(t *testing.T) {
	t.Parallel()

	in := protocol.CostPayload{
		InputTokens:           1000,
		CachedInputTokens:     250,
		OutputTokens:          100,
		ReasoningOutputTokens: 40,
		TurnCostUSD:           0.1234,
		SessionTotal:          0.5678,
		BudgetLimit:           5,
		BudgetRemain:          4.4322,
		PricingModel:          "gpt-5.5",
		PricingAvailable:      true,
		TurnCredits:           1.25,
		SessionTotalCredits:   2.5,
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got protocol.CostPayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != in {
		t.Fatalf("round-trip mismatch\nwant: %+v\n got: %+v\n raw: %s", in, got, raw)
	}
}
