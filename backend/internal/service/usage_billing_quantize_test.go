package service

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func usageBillingDecimalPlaces(v float64) int32 {
	return -decimal.NewFromFloat(v).Exponent()
}

func TestUsageBillingCommandQuantizesBalanceAndQuotaIdentically(t *testing.T) {
	const actualCost = 0.000078125

	cmd := &UsageBillingCommand{
		RequestID:          "req-5229",
		UserID:             1,
		APIKeyID:           2,
		AccountID:          3,
		BalanceCost:        actualCost,
		PrepaidBalanceCost: actualCost,
		APIKeyQuotaCost:    actualCost,
	}
	cmd.Normalize()

	require.Equal(t, cmd.BalanceCost, cmd.APIKeyQuotaCost)
	require.Equal(t, cmd.BalanceCost, cmd.PrepaidBalanceCost)
	require.LessOrEqual(t, usageBillingDecimalPlaces(cmd.BalanceCost), int32(UsageBillingMonetaryScale))
}

func TestUsageBillingCommandNormalizePreservesBalanceCheckPolicy(t *testing.T) {
	cmd := &UsageBillingCommand{RequestID: "req-balance-policy", BalanceCost: 0.06}
	cmd.Normalize()
	require.False(t, cmd.RequireBalanceCheck, "ordinary post-response usage must be allowed to record debt")

	cmd.RequireBalanceCheck = true
	cmd.Normalize()
	require.True(t, cmd.RequireBalanceCheck, "known-amount reservations must remain strict")
}

func TestQuantizeUsageBillingAmountBoundaries(t *testing.T) {
	cases := []struct {
		name string
		in   float64
	}{
		{"below_half", 0.000078120},
		{"just_below_half", 0.000078124},
		{"exact_half", 0.000078125},
		{"just_above_half", 0.000078126},
		{"above_half", 0.000078130},
		{"long_tail", 0.0000781234567},
		{"already_quantized", 0.00007813},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := QuantizeUsageBillingAmount(tc.in)
			want, _ := decimal.NewFromFloat(tc.in).Round(UsageBillingMonetaryScale).Float64()

			require.Equal(t, want, got)
			require.LessOrEqual(t, usageBillingDecimalPlaces(got), int32(UsageBillingMonetaryScale))
			require.LessOrEqual(t, math.Abs(got-tc.in), 5e-9)
		})
	}
}

func TestQuantizedAmountsReconcileExactlyOverManyApplications(t *testing.T) {
	const actualCost = 0.000078125

	cmd := &UsageBillingCommand{
		RequestID:           "req-5229-bulk",
		UserID:              1,
		APIKeyID:            2,
		AccountID:           3,
		BalanceCost:         actualCost,
		PrepaidBalanceCost:  actualCost,
		SubscriptionCost:    0,
		APIKeyQuotaCost:     actualCost,
		APIKeyRateLimitCost: actualCost,
	}
	cmd.Normalize()

	unit := decimal.NewFromFloat(cmd.BalanceCost)
	for _, n := range []int64{1, 10, 100, 1000} {
		total := unit.Mul(decimal.NewFromInt(n))
		balance := decimal.NewFromInt(10000).Sub(total)
		quotaUsed := total

		require.True(t, balance.Equal(balance.Round(UsageBillingMonetaryScale)), "n=%d", n)
		require.True(t, quotaUsed.Equal(quotaUsed.Round(UsageBillingMonetaryScale)), "n=%d", n)
		require.True(t, decimal.NewFromInt(10000).Sub(balance).Equal(quotaUsed), "n=%d", n)
	}
}

func TestNormalizeQuantizesEveryMonetaryField(t *testing.T) {
	const raw = 0.0000781234567

	cmd := &UsageBillingCommand{
		RequestID:           "req-5229-fields",
		UserID:              1,
		APIKeyID:            2,
		AccountID:           3,
		BalanceCost:         raw,
		PrepaidBalanceCost:  raw,
		SubscriptionCost:    raw,
		APIKeyQuotaCost:     raw,
		APIKeyRateLimitCost: raw,
		AccountQuotaCost:    raw,
	}
	cmd.Normalize()

	for name, got := range map[string]float64{
		"BalanceCost":         cmd.BalanceCost,
		"PrepaidBalanceCost":  cmd.PrepaidBalanceCost,
		"SubscriptionCost":    cmd.SubscriptionCost,
		"APIKeyQuotaCost":     cmd.APIKeyQuotaCost,
		"APIKeyRateLimitCost": cmd.APIKeyRateLimitCost,
		"AccountQuotaCost":    cmd.AccountQuotaCost,
	} {
		require.LessOrEqual(t, usageBillingDecimalPlaces(got), int32(UsageBillingMonetaryScale), name)
	}
}

func TestNormalizeKeepsFingerprintDerivedFromRawAmounts(t *testing.T) {
	const raw = 0.000078125

	newCommand := func() *UsageBillingCommand {
		return &UsageBillingCommand{
			RequestID:           "req-5229-fp",
			UserID:              1,
			APIKeyID:            2,
			AccountID:           3,
			BalanceCost:         raw,
			PrepaidBalanceCost:  raw,
			SubscriptionCost:    raw,
			APIKeyQuotaCost:     raw,
			APIKeyRateLimitCost: raw,
			AccountQuotaCost:    raw,
		}
	}

	cmd := newCommand()
	expected := buildUsageBillingFingerprint(newCommand())
	cmd.Normalize()

	require.Equal(t, expected, cmd.RequestFingerprint)
	require.NotEqual(t, buildUsageBillingFingerprint(cmd), cmd.RequestFingerprint)
}

func TestNormalizePreservesExplicitFingerprint(t *testing.T) {
	cmd := &UsageBillingCommand{
		RequestID:          "req-5229-explicit",
		RequestFingerprint: "preset-fingerprint",
		BalanceCost:        0.0000781234567,
		PrepaidBalanceCost: 0.0000781234567,
	}
	cmd.Normalize()

	require.Equal(t, "preset-fingerprint", cmd.RequestFingerprint)
	require.LessOrEqual(t, usageBillingDecimalPlaces(cmd.BalanceCost), int32(UsageBillingMonetaryScale))
	require.LessOrEqual(t, usageBillingDecimalPlaces(cmd.PrepaidBalanceCost), int32(UsageBillingMonetaryScale))
}

func TestQuantizeUsageBillingAmountPassesThroughNonFinite(t *testing.T) {
	require.Equal(t, 0.0, QuantizeUsageBillingAmount(0))
	require.True(t, math.IsNaN(QuantizeUsageBillingAmount(math.NaN())))
	require.True(t, math.IsInf(QuantizeUsageBillingAmount(math.Inf(1)), 1))
	require.True(t, math.IsInf(QuantizeUsageBillingAmount(math.Inf(-1)), -1))
}

func TestQuantizeUsageBillingAmountHandlesNegativeAmounts(t *testing.T) {
	got := QuantizeUsageBillingAmount(-0.000078125)
	want, _ := decimal.NewFromFloat(-0.000078125).Round(UsageBillingMonetaryScale).Float64()

	require.Equal(t, want, got)
	require.Equal(t, -QuantizeUsageBillingAmount(0.000078125), got)
}
