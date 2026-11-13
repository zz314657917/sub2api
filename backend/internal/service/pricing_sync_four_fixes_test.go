package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// Embedding preserves unrelated remote-client capabilities. Each hash entry
// point delegates to the same deterministic fixture; no network is used.
type fourFixesPricingRemote struct {
	PricingRemoteClient
	body        []byte
	hash        string
	bodyFetches int
	hashFetches int
}

func (r *fourFixesPricingRemote) FetchPricingJSON(context.Context, string) ([]byte, error) {
	r.bodyFetches++
	return append([]byte(nil), r.body...), nil
}

func (r *fourFixesPricingRemote) FetchHash(context.Context, string) (string, error) {
	r.hashFetches++
	return r.hash, nil
}

func (r *fourFixesPricingRemote) FetchPricingHash(ctx context.Context, url string) (string, error) {
	return r.FetchHash(ctx, url)
}

func (r *fourFixesPricingRemote) FetchHashText(ctx context.Context, url string) (string, error) {
	return r.FetchHash(ctx, url)
}

func fourFixesBodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func newFourFixesPricingSync(t *testing.T, remote *fourFixesPricingRemote) *PricingService {
	t.Helper()
	cfg := &config.Config{}
	cfg.Pricing.DataDir = t.TempDir()
	cfg.Pricing.RemoteURL = "https://example.com/pricing.json"
	cfg.Pricing.HashURL = "https://example.com/pricing.sha256"
	return NewPricingService(cfg, remote)
}

func requireFourFixesPricingAnchor(t *testing.T, svc *PricingService, want string) {
	t.Helper()
	svc.mu.RLock()
	got := svc.localHash
	svc.mu.RUnlock()
	require.Equal(t, want, got)
	persisted, err := os.ReadFile(svc.getHashFilePath())
	require.NoError(t, err)
	require.Equal(t, want+"\n", string(persisted))
}

func TestPricingSyncFourFixesMismatchRetriesAndMatchingHashSuppressesDownload(t *testing.T) {
	oldBody := []byte(`{"claude-sonnet-5-5":{"input_cost_per_token":0.000002,"output_cost_per_token":0.000010,"cache_read_input_token_cost":0.0000001}}`)
	newBody := []byte(`{"claude-sonnet-5-5":{"input_cost_per_token":0.000003,"output_cost_per_token":0.000010,"cache_read_input_token_cost":0.0000001}}`)
	remote := &fourFixesPricingRemote{body: oldBody, hash: fourFixesBodyHash(newBody)}
	svc := newFourFixesPricingSync(t, remote)
	bodyHash := fourFixesBodyHash(oldBody)
	require.NotEqual(t, remote.hash, bodyHash)
	require.NoError(t, svc.syncWithRemote())
	require.Equal(t, 1, remote.bodyFetches)
	requireFourFixesPricingAnchor(t, svc, bodyHash)
	persisted, err := os.ReadFile(svc.getPricingFilePath())
	require.NoError(t, err)
	require.Equal(t, oldBody, persisted)
	// A still-mismatched publication must be downloaded again.
	require.NoError(t, svc.syncWithRemote())
	require.Equal(t, 2, remote.bodyFetches)
	requireFourFixesPricingAnchor(t, svc, bodyHash)
	// Once the body catches up, acknowledge that version and suppress retries.
	remote.body = newBody
	require.NoError(t, svc.syncWithRemote())
	require.Equal(t, 3, remote.bodyFetches)
	requireFourFixesPricingAnchor(t, svc, remote.hash)
	require.NoError(t, svc.syncWithRemote())
	require.Equal(t, 3, remote.bodyFetches)
	require.Greater(t, remote.hashFetches, 0)
	price := svc.GetModelPricing("claude-sonnet-5-5")
	require.NotNil(t, price)
	require.InDelta(t, 3e-6, price.InputCostPerToken, 1e-15)
}

func TestPricingSyncFourFixesMatchingHashCaseAndNoHash(t *testing.T) {
	body := []byte(`{"claude-sonnet-5-5":{"input_cost_per_token":0.000002,"output_cost_per_token":0.000010}}`)
	t.Run("matching uppercase", func(t *testing.T) {
		remote := &fourFixesPricingRemote{body: body, hash: strings.ToUpper(fourFixesBodyHash(body))}
		svc := newFourFixesPricingSync(t, remote)
		require.NoError(t, svc.downloadPricingData())
		// fetchRemoteHash may normalize hexadecimal case.
		svc.mu.RLock()
		anchor := svc.localHash
		svc.mu.RUnlock()
		require.True(t, strings.EqualFold(remote.hash, anchor))
		requireFourFixesPricingAnchor(t, svc, anchor)
		require.NoError(t, svc.syncWithRemote())
		require.Equal(t, 1, remote.bodyFetches)
	})
	t.Run("no hash URL", func(t *testing.T) {
		remote := &fourFixesPricingRemote{body: body}
		svc := newFourFixesPricingSync(t, remote)
		svc.cfg.Pricing.HashURL = ""
		require.NoError(t, svc.downloadPricingData())
		requireFourFixesPricingAnchor(t, svc, fourFixesBodyHash(body))
		require.Zero(t, remote.hashFetches)
	})
}

func TestSonnet55FourFixesFallbackCatalogAndBilling(t *testing.T) {
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &raw))
	var catalogCard struct {
		Read   float64 `json:"cache_read_input_token_cost"`
		Write5 float64 `json:"cache_creation_input_token_cost"`
		Write1 float64 `json:"cache_creation_input_token_cost_above_1hr"`
		Input  float64 `json:"input_cost_per_token"`
		Output float64 `json:"output_cost_per_token"`
	}
	require.NoError(t, json.Unmarshal(raw["claude-sonnet-5-5"], &catalogCard))
	require.InDelta(t, .1e-6, catalogCard.Read, 1e-15)
	require.InDelta(t, 2.5e-6, catalogCard.Write5, 1e-15)
	require.InDelta(t, 4e-6, catalogCard.Write1, 1e-15)
	require.InDelta(t, 2e-6, catalogCard.Input, 1e-15)
	require.InDelta(t, 10e-6, catalogCard.Output, 1e-15)
	ps := &PricingService{}
	data, err := ps.parsePricingData(body)
	require.NoError(t, err)
	for source, bs := range map[string]*BillingService{
		"fallback": NewBillingService(&config.Config{}, nil),
		"catalog":  NewBillingService(&config.Config{}, &PricingService{pricingData: data}),
	} {
		t.Run(source, func(t *testing.T) {
			for _, model := range []string{"claude-sonnet-5-5", "claude-sonnet-5.5-high"} {
				price, err := bs.GetModelPricing(model)
				require.NoError(t, err)
				require.InDelta(t, .1e-6, price.CacheReadPricePerToken, 1e-15)
				require.InDelta(t, 2e-6, price.InputPricePerToken, 1e-15)
				require.InDelta(t, 10e-6, price.OutputPricePerToken, 1e-15)
				for _, input := range []int{1000, 300000} {
					for _, withResolver := range []bool{false, true} {
						var resolver *ModelPricingResolver
						if withResolver {
							resolver = NewModelPricingResolver(nil, bs)
						}
						cost, err := bs.CalculateCostUnified(CostInput{
							Ctx: context.Background(), Model: model, Resolver: resolver, RateMultiplier: 1,
							Tokens: UsageTokens{InputTokens: input, OutputTokens: 1000, CacheReadTokens: 1000,
								CacheCreationTokens: 1000, CacheCreation5mTokens: 600, CacheCreation1hTokens: 400},
						})
						require.NoError(t, err)
						require.InDelta(t, float64(input)*2e-6, cost.InputCost, 1e-12)
						require.InDelta(t, .010, cost.OutputCost, 1e-15)
						require.InDelta(t, .0001, cost.CacheReadCost, 1e-15)
						require.InDelta(t, .0031, cost.CacheCreationCost, 1e-15)
						require.InDelta(t, float64(input)*2e-6+.010+.0001+.0031, cost.TotalCost, 1e-12)
					}
				}
			}
		})
	}
}
