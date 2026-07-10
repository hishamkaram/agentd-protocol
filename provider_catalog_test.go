package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestProviderRuntimeCatalogRoundTrip(t *testing.T) {
	t.Parallel()

	want := ListProviderCatalogsResponse{
		Type:      MsgProviderCatalogList,
		RequestID: "catalog-1",
		Catalogs: []ProviderRuntimeCatalog{
			{
				Agent:        "claude-code",
				Provider:     "bedrock",
				State:        ProviderCatalogReady,
				Generation:   "generation-1",
				Source:       ProviderCatalogSourceSDK,
				CLIVersion:   "2.1.206",
				DiscoveredAt: 1783641600000,
				Models: []ProviderModelInfo{
					{
						Value:                    "provider-model-a",
						ResolvedModel:            "provider-model-a-20260710",
						DisplayName:              "Provider Model A",
						SupportsEffort:           true,
						SupportedEffortLevels:    []string{"low", "high"},
						SupportsAdaptiveThinking: true,
					},
				},
				ApprovalModes: []ProviderControlOption{{Value: "default", ProviderValue: "manual"}},
			},
		},
	}

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if bytes.Contains(raw, []byte("provider_value")) || bytes.Contains(raw, []byte("manual")) {
		t.Fatalf("wire payload exposed provider-owned control value: %s", raw)
	}
	var got ListProviderCatalogsResponse
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.Catalogs) != 1 || got.Catalogs[0].Provider != "bedrock" || got.Catalogs[0].Models[0].ResolvedModel != "provider-model-a-20260710" {
		t.Fatalf("round trip = %+v", got)
	}
}

func TestProviderCatalogUnavailableIsSanitized(t *testing.T) {
	t.Parallel()

	catalog := ProviderRuntimeCatalog{
		Agent:    "claude-code",
		Provider: "vertex",
		State:    ProviderCatalogUnavailable,
		Error: &ProviderCatalogError{
			Code:      ProviderCatalogErrorAuthRequired,
			Message:   "Claude authentication is required",
			Retryable: true,
		},
	}
	raw, err := json.Marshal(catalog)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, forbidden := range []string{"stderr", "path", "token", "credential"} {
		if json.Valid(raw) && containsASCIIInsensitive(raw, forbidden) {
			t.Fatalf("catalog JSON contains forbidden provider detail %q: %s", forbidden, raw)
		}
	}
}

func containsASCIIInsensitive(raw []byte, needle string) bool {
	if len(needle) == 0 || len(raw) < len(needle) {
		return false
	}
	for i := 0; i <= len(raw)-len(needle); i++ {
		matched := true
		for j := range needle {
			b := raw[i+j]
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if b != needle[j] {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
