package service

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func newModelBilling20261022Sources(t *testing.T) map[string]*BillingService {
	t.Helper()
	cfg := &config.Config{}
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	ps := &PricingService{}
	data, err := ps.parsePricingData(body)
	require.NoError(t, err)
	return map[string]*BillingService{
		"hardcoded": NewBillingService(cfg, nil),
		"bundle":    NewBillingService(cfg, &PricingService{pricingData: data}),
		"stale": NewBillingService(cfg, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
			"claude-sonnet-4": {InputCostPerToken: 99e-6},
			"claude-opus-5-5": {InputCostPerToken: 4e-6, OutputCostPerToken: 20e-6},
			"grok-4.7":        {InputCostPerToken: 2e-6, OutputCostPerToken: 6e-6},
			"deepseek-flash":  {InputCostPerToken: 99, OutputCostPerToken: 99},
		}}),
	}
}

func newModelBilling20261022Time(value string) time.Time {
	at, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return at
}

func TestNewModelBilling20261022Cards(t *testing.T) {
	for _, dynamic := range []bool{false, true} {
		bs := NewBillingService(&config.Config{}, nil)
		if dynamic {
			bs.pricingService = &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				"claude-sonnet-4": {InputCostPerToken: 3e-6, OutputCostPerToken: 15e-6},
				"claude-opus-5":   {InputCostPerToken: 5e-6, OutputCostPerToken: 25e-6},
			}}
		}
		for _, tc := range []struct {
			model         string
			in, out, read float64
		}{
			{"grok-4.7-fast", 4, 12, 1},
			{"x-ai/grok-4.7-build-fast", 4, 12, 1},
			{"gemini-3.6-flash", .75, 3.75, .075},
			{"gemini-3.7-flash-high", .75, 3.75, .075},
			{"gemini-3.8-flash-medium", .75, 3.75, .075},
			{"gemini-4-argon", 2, 10, .1},
			{"gemini-4-argon-cyber-permissive-low", 4, 20, .2},
			{"claude-sonnet-5.5-high", 2, 10, .2},
			{"claude-haiku-5-5-low", .1, .5, .01},
			{"claude-opus-5.5-medium", 4, 20, .2},
		} {
			t.Run(tc.model, func(t *testing.T) {
				p, err := bs.GetModelPricing(tc.model)
				require.NoError(t, err)
				require.InDelta(t, tc.in*1e-6, p.InputPricePerToken, 1e-15)
				require.InDelta(t, tc.out*1e-6, p.OutputPricePerToken, 1e-15)
				require.InDelta(t, tc.read*1e-6, p.CacheReadPricePerToken, 1e-15)
			})
		}
	}
}

func TestNewModelBilling20261022DeepSeekBoundaries(t *testing.T) {
	for source, bs := range newModelBilling20261022Sources(t) {
		t.Run(source, func(t *testing.T) {
			for _, tc := range []struct {
				at     string
				factor float64
			}{
				{"2026-10-22T00:59:59.999999999Z", 1},
				{"2026-10-22T01:00:00Z", 2},
				{"2026-10-22T03:59:59.999999999Z", 2},
				{"2026-10-22T04:00:00Z", 1},
				{"2026-10-22T05:59:59.999999999Z", 1},
				{"2026-10-22T06:00:00Z", 2},
				{"2026-10-22T09:59:59.999999999Z", 2},
				{"2026-10-22T10:00:00Z", 1},
				{"2026-10-23T15:59:59Z", 1},
				{"2026-10-23T16:00:00Z", 1},
				{"2026-10-24T01:00:00Z", 1},
				{"2026-10-25T01:00:00Z", 1},
				{"2026-10-25T15:59:59Z", 1},
				{"2026-10-25T16:00:00Z", 1},
				{"2026-10-26T01:00:00Z", 2},
			} {
				at := newModelBilling20261022Time(tc.at)
				require.Equal(t, tc.factor, deepseekPeakMultiplierAt(at))
				for _, model := range []string{"deepseek-flash", "deepseek-v4-flash", "deepseek-v4.1-flash", "deepseek-v4-flash-vision-exp", "deepseek-chat", "deepseek-reasoner", "deepseek-v4-pro", "deepseek-v4-pro-0813", "deepseek-v4-flash-20261001"} {
					for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
						bd, err := bs.CalculateCostUnified(CostInput{
							Ctx: context.Background(), PricingAt: at, Model: model, Resolver: resolver,
							Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000}, RateMultiplier: 3,
						})
						require.NoError(t, err)
						in, out, read := .00015, .00060, .000003
						if strings.HasPrefix(model, "deepseek-v4-pro") {
							in, out, read = .0045, .0135, .00015
						}
						require.InDelta(t, in*tc.factor, bd.InputCost, 1e-15)
						require.InDelta(t, out*tc.factor, bd.OutputCost, 1e-15)
						require.InDelta(t, read*tc.factor, bd.CacheReadCost, 1e-15)
						require.InDelta(t, (in+out+read)*tc.factor*3, bd.ActualCost, 1e-15)
					}
				}
			}
		})
	}
}

func TestNewModelBilling20261022ProCardAndLegacy(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	tokens := UsageTokens{InputTokens: 100000, OutputTokens: 1000, CacheReadTokens: 1000}
	for _, tc := range []struct {
		at            string
		in, out, read float64
	}{
		{"2026-09-14T03:59:59.999999999Z", 4.50 * 2, 13.50 * 2, .15 * 2},
		{"2026-09-14T04:00:00Z", 4.50, 13.50, .15},
		{"2026-09-14T04:00:00.000000001Z", 4.50, 13.50, .15},
		{"2026-10-24T01:00:00Z", 4.50, 13.50, .15},
		{"2026-10-26T01:00:00Z", 4.50 * 2, 13.50 * 2, .15 * 2},
	} {
		at := newModelBilling20261022Time(tc.at)
		want := (100000*tc.in + 1000*tc.out + 1000*tc.read) * 1e-6
		ctx := bs.captureBillingPricingTime(context.Background(), at)
		for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
			bd, err := bs.CalculateTokenCostForRequest(TokenCostRequest{Ctx: ctx, Model: "deepseek-v4-pro",
				Tokens: tokens, Resolver: resolver, RateMultiplier: 1,
				LegacyLongContext: &LegacyLongContextRule{Threshold: 10, Multiplier: 9}})
			require.NoError(t, err)
			require.InDelta(t, want, bd.TotalCost, 1e-15)
		}
		bd, err := bs.CalculateCostWithLongContext("deepseek-v4-pro", tokens, 1, 10, 9, at)
		require.NoError(t, err)
		require.InDelta(t, want, bd.TotalCost, 1e-15)
		legacy, err := bs.calculateCostInternalWithPolicy("deepseek-v4-pro", tokens, 1, "", nil, true, at)
		require.NoError(t, err)
		require.InDelta(t, want, legacy.TotalCost, 1e-15)
	}
	_, err := bs.GetModelPricing("deepseek-future-unknown")
	require.Error(t, err)
	_, err = bs.GetModelPricing("deepseek-v4-pro-unknown")
	require.Error(t, err)
	// Existing explicitly catalogued older identities remain supported.
	ps := &PricingService{pricingData: map[string]*LiteLLMModelPricing{"deepseek-v3-2-251201": {InputCostPerToken: 3e-7}}}
	p, err := NewBillingService(&config.Config{}, ps).GetModelPricing("deepseek-v3-2-251201")
	require.NoError(t, err)
	require.Equal(t, 3e-7, p.InputPricePerToken)
}

func TestNewModelBilling20261022ThresholdsAndCacheTTL(t *testing.T) {
	for source, bs := range newModelBilling20261022Sources(t) {
		for _, tc := range []struct {
			model                                 string
			threshold                             int
			in, out, read, write5, write1, factor float64
			inclusive                             bool
		}{
			{"grok-4.7-fast", 200000, 4, 12, 1, 0, 0, 1.5, true},
			{"xai/grok-4.7-build-fast", 200000, 4, 12, 1, 0, 0, 1.5, true},
			{"claude-haiku-5.5-high", 100000, .1, .5, .01, .125, .2, 5, false},
		} {
			t.Run(source+"/"+tc.model, func(t *testing.T) {
				for _, total := range []int{tc.threshold - 1, tc.threshold, tc.threshold + 1} {
					tokens := UsageTokens{InputTokens: total - 2000, OutputTokens: 1000, CacheReadTokens: 1000,
						CacheCreationTokens: 1000, CacheCreation5mTokens: 600, CacheCreation1hTokens: 400}
					factor := 1.0
					if total > tc.threshold || (total == tc.threshold && tc.inclusive) {
						factor = tc.factor
					}
					for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
						bd, err := bs.CalculateCostUnified(CostInput{Model: tc.model, Tokens: tokens, RateMultiplier: 1, Resolver: resolver})
						require.NoError(t, err)
						require.InDelta(t, float64(total-2000)*tc.in*1e-6*factor, bd.InputCost, 1e-12)
						require.InDelta(t, .001*tc.out*factor, bd.OutputCost, 1e-12)
						require.InDelta(t, .001*tc.read*factor, bd.CacheReadCost, 1e-12)
						require.InDelta(t, (600*tc.write5+400*tc.write1)*1e-6*factor, bd.CacheCreationCost, 1e-12)
					}
				}
			})
		}
		for _, model := range []string{"claude-sonnet-5-5", "claude-opus-5.5-low"} {
			bd, err := bs.CalculateCost(model, UsageTokens{CacheCreationTokens: 1000, CacheCreation5mTokens: 600, CacheCreation1hTokens: 400}, 1)
			require.NoError(t, err)
			want := .0031
			if model == "claude-opus-5.5-low" {
				want = .0062
			}
			require.InDelta(t, want, bd.CacheCreationCost, 1e-15)
		}
	}
}

func TestNewModelBilling20261022GeminiPriorityAndAliases(t *testing.T) {
	for source, bs := range newModelBilling20261022Sources(t) {
		for _, tc := range []struct {
			model         string
			in, out, read float64
		}{
			{"gemini-3.6-flash", 1.35, 6.75, .135},
			{"antigravity/gemini-3.6-flash-high", 1.35, 6.75, .135},
			{"google/gemini-4-argon-low", 3.6, 18, .18},
			{"models/gemini-4-argon-cyber-permissive-medium", 7.2, 36, .36},
			{"gemini-3.7-flash-tiered", 1.5, 7.5, .15}, // generic 2x; no invented 1.8x card
			{"gemini-3.8-flash-high", 1.5, 7.5, .15},
			{"claude-opus-5.5-high", 8, 40, .4},
		} {
			t.Run(source+"/"+tc.model, func(t *testing.T) {
				tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000, CacheReadTokens: 1000}
				for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
					bd, err := bs.CalculateCostUnified(CostInput{Model: tc.model, Tokens: tokens, RateMultiplier: 1, ServiceTier: "priority", Resolver: resolver})
					require.NoError(t, err)
					require.InDelta(t, .3*tc.in+.001*tc.out+.001*tc.read, bd.TotalCost, 1e-12)
					require.Zero(t, bd.CacheCreationCost)
				}
				if tc.model[0:6] == "gemini" {
					bd, err := bs.CalculateTokenCostForRequest(TokenCostRequest{Ctx: context.Background(), Model: tc.model, Tokens: tokens,
						RateMultiplier: 1, LegacyLongContext: bs.LegacyLongContextRule(PlatformGemini)})
					require.NoError(t, err)
					standard, err := bs.CalculateCost(tc.model, tokens, 1)
					require.NoError(t, err)
					require.InDelta(t, standard.TotalCost, bd.TotalCost, 1e-12)
				}
			})
		}
	}
}

func TestNewModelBilling20261022ExplicitAliasesZeroAndNonmutation(t *testing.T) {
	ps := &PricingService{}
	data, err := ps.parsePricingData([]byte(`{
		"gemini-3.6-flash-high":{"input_cost_per_token":0,"output_cost_per_token":0,"cache_read_input_token_cost":0},
		"gemini-4-argon-low":{"input_cost_per_token":9e-6,"output_cost_per_token":8e-6,"cache_read_input_token_cost":7e-6,"input_cost_per_token_priority":0,"output_cost_per_token_priority":0,"cache_read_input_token_cost_priority":0},
		"claude-opus-5.5-high":{"input_cost_per_token":0,"output_cost_per_token":0,"cache_read_input_token_cost":0,"cache_creation_input_token_cost":0,"cache_creation_input_token_cost_above_1hr":0},
		"claude-sonnet-5-5-low":{"input_cost_per_token":7e-6,"output_cost_per_token":9e-6,"cache_read_input_token_cost":2e-6,"cache_creation_input_token_cost":3e-6,"cache_creation_input_token_cost_above_1hr":4e-6},
		"grok-4.7-fast":{"input_cost_per_token":0,"output_cost_per_token":0,"cache_read_input_token_cost":0}
	}`))
	require.NoError(t, err)
	ps.pricingData = data
	before, err := json.Marshal(data)
	require.NoError(t, err)
	bs := NewBillingService(&config.Config{}, ps)
	tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000, CacheReadTokens: 1000}
	for _, model := range []string{"gemini-3.6-flash-high", "claude-opus-5.5-high", "grok-4.7-fast"} {
		for _, tier := range []string{"", "priority"} {
			bd, err := bs.CalculateCostWithServiceTier(model, tokens, 1, tier)
			require.NoError(t, err)
			require.Zero(t, bd.TotalCost, model+"/"+tier)
		}
	}
	bd, err := bs.CalculateCostWithServiceTier("gemini-4-argon-low", tokens, 1, "priority")
	require.NoError(t, err)
	require.Zero(t, bd.TotalCost)
	bd, err = bs.CalculateCost("claude-sonnet-5-5-low", tokens, 1)
	require.NoError(t, err)
	require.InDelta(t, 2.111, bd.TotalCost, 1e-12)
	after, err := json.Marshal(data)
	require.NoError(t, err)
	require.Equal(t, string(before), string(after))
	for _, model := range []string{"claude-sonnet-5-5-unknown", "claude-haiku-5.5-unknown", "claude-opus-5-5-unknown", "gemini-4-argon-cyber-unknown", "gemini-3.6-flash-unknown", "grok-4.7-fast-unknown"} {
		_, err := bs.GetModelPricing(model)
		require.Error(t, err, model)
	}
}

func TestNewModelBilling20261022AdminAndSharedPrice(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	shared := &ModelPricing{InputPricePerToken: 7e-6, OutputPricePerToken: 8e-6, CacheReadPricePerToken: 9e-6}
	original := *shared
	resolver := NewModelPricingResolver(nil, bs)
	for _, source := range []string{PricingSourceGroup, PricingSourceChannel} {
		for _, free := range []bool{false, true} {
			price := shared
			if free {
				price = &ModelPricing{}
			}
			resolved := &ResolvedPricing{Mode: BillingModeToken, BasePricing: price, Source: source}
			cost, err := bs.CalculateCostUnified(CostInput{Model: "deepseek-v4-pro", PricingAt: newModelBilling20261022Time("2026-10-22T01:00:00Z"),
				Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000}, RateMultiplier: 3,
				Resolver: resolver, Resolved: resolved})
			require.NoError(t, err)
			want := .072
			if free {
				want = 0
			}
			require.InDelta(t, want, cost.ActualCost, 1e-15)
		}
	}
	require.True(t, reflect.DeepEqual(original, *shared))
	p, err := bs.GetModelPricing("deepseek-v4-pro")
	require.NoError(t, err)
	copy := *p
	_, err = bs.CalculateCostUnified(CostInput{Model: "deepseek-v4-pro", PricingAt: newModelBilling20261022Time("2026-09-14T03:59:59Z"),
		Tokens: UsageTokens{InputTokens: 1000}, RateMultiplier: 1, Resolved: &ResolvedPricing{BasePricing: p, Source: PricingSourceLiteLLM}})
	require.NoError(t, err)
	require.Equal(t, copy, *p)
	inputPrice, outputPrice, readPrice := 11e-6, 12e-6, 13e-6
	cost, err := bs.CalculateCostUnified(CostInput{Model: "deepseek-flash", PricingAt: newModelBilling20261022Time("2026-10-22T01:00:00Z"),
		Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000}, RateMultiplier: 1,
		Resolved: &ResolvedPricing{Source: PricingSourceChannel, Mode: BillingModeToken, BasePricing: shared,
			Intervals: []PricingInterval{{InputPrice: &inputPrice, OutputPrice: &outputPrice, CacheReadPrice: &readPrice}}}})
	require.NoError(t, err)
	require.InDelta(t, .036, cost.TotalCost, 1e-15)
}

func newModelBilling20261022OpusPriorityCard(t *testing.T, variant string) *BillingService {
	t.Helper()
	card := map[string]float64{
		"input_cost_per_token": 4e-6, "output_cost_per_token": 20e-6,
		"cache_read_input_token_cost":     .2e-6,
		"cache_creation_input_token_cost": 5e-6, "cache_creation_input_token_cost_above_1hr": 8e-6,
		"input_cost_per_token_priority": 0, "output_cost_per_token_priority": 0,
		"cache_read_input_token_cost_priority": 0, "cache_creation_input_token_cost_priority": 0,
	}
	switch variant {
	case "nonzero":
		card["input_cost_per_token_priority"], card["output_cost_per_token_priority"] = 8e-6, 40e-6
		card["cache_read_input_token_cost_priority"], card["cache_creation_input_token_cost_priority"] = .4e-6, 10e-6
	case "partial-input":
		delete(card, "input_cost_per_token_priority")
	case "partial-write":
		delete(card, "cache_creation_input_token_cost_priority")
	case "mixed-input":
		card["input_cost_per_token_priority"] = 8e-6
	}
	body, err := json.Marshal(map[string]any{"claude-opus-5-5-high": card})
	require.NoError(t, err)
	ps := &PricingService{}
	ps.pricingData, err = ps.parsePricingData(body)
	require.NoError(t, err)
	return NewBillingService(&config.Config{}, ps)
}

func TestNewModelBilling20261022OpusPriorityZero5m(t *testing.T) {
	for _, tc := range []struct {
		name, card                                 string
		aggregate, write5, write1                  int
		wantInput, wantOutput, wantRead, wantWrite float64
	}{
		{name: "zero-no-write", card: "zero"},
		{name: "zero-5m", card: "zero", aggregate: 8000, write5: 8000},
		{name: "zero-implicit-5m", card: "zero", aggregate: 8000},
		{name: "zero-1h-not-free", card: "zero", aggregate: 8000, write1: 8000, wantInput: .56, wantOutput: .4, wantRead: .008, wantWrite: .128},
		{name: "zero-mixed-ttl", card: "zero", aggregate: 8000, write5: 4000, write1: 4000, wantInput: .56, wantOutput: .4, wantRead: .008, wantWrite: .064},
		{name: "nonzero-5m-unchanged", card: "nonzero", aggregate: 8000, write5: 8000, wantInput: .56, wantOutput: .4, wantRead: .008, wantWrite: .08},
		{name: "nonzero-1h-unchanged", card: "nonzero", aggregate: 8000, write1: 8000, wantInput: .56, wantOutput: .4, wantRead: .008, wantWrite: .128},
		{name: "missing-input-retains-strict-fallback", card: "partial-input", aggregate: 8000, write5: 8000, wantInput: .56, wantOutput: .4, wantRead: .008},
		{name: "missing-write-retains-strict-fallback", card: "partial-write", aggregate: 8000, write5: 8000, wantInput: .56, wantOutput: .4, wantRead: .008, wantWrite: .08},
		{name: "mixed-nonzero-retains-strict-fallback", card: "mixed-input", aggregate: 8000, write5: 8000, wantInput: .56, wantOutput: .4, wantRead: .008},
	} {
		for _, withResolver := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/legacy", true: "/resolver"}[withResolver], func(t *testing.T) {
				bs := newModelBilling20261022OpusPriorityCard(t, tc.card)
				before, err := json.Marshal(bs.pricingService.pricingData)
				require.NoError(t, err)
				var resolver *ModelPricingResolver
				if withResolver {
					resolver = NewModelPricingResolver(nil, bs)
				}
				bd, err := bs.CalculateCostUnified(CostInput{
					Model: "claude-opus-5-5-high", ServiceTier: "priority", RateMultiplier: 1, Resolver: resolver,
					Tokens: UsageTokens{InputTokens: 70000, OutputTokens: 10000, CacheReadTokens: 20000,
						CacheCreationTokens: tc.aggregate, CacheCreation5mTokens: tc.write5, CacheCreation1hTokens: tc.write1},
				})
				require.NoError(t, err)
				require.InDelta(t, tc.wantInput, bd.InputCost, 1e-12)
				require.InDelta(t, tc.wantOutput, bd.OutputCost, 1e-12)
				require.InDelta(t, tc.wantRead, bd.CacheReadCost, 1e-12)
				require.InDelta(t, tc.wantWrite, bd.CacheCreationCost, 1e-12)
				require.InDelta(t, tc.wantInput+tc.wantOutput+tc.wantRead+tc.wantWrite, bd.TotalCost, 1e-12)
				require.Equal(t, bd.TotalCost, bd.ActualCost)
				after, err := json.Marshal(bs.pricingService.pricingData)
				require.NoError(t, err)
				require.Equal(t, string(before), string(after))
			})
		}
	}
}
