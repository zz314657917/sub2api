//go:build unit

package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentMethodFingerprint_StableAndOneWay(t *testing.T) {
	t.Parallel()

	got := paymentMethodFingerprint("alipay", map[string]string{"buyer_id": "2088123412341234"})
	again := paymentMethodFingerprint("alipay", map[string]string{"buyer_id": " 2088123412341234 "})

	require.NotEmpty(t, got)
	require.Equal(t, got, again)
	require.True(t, strings.HasPrefix(got, paymentMethodFingerprintPrefix))
	require.NotContains(t, got, "2088123412341234")
}

func TestPaymentMethodFingerprint_IgnoresNonPayerMetadata(t *testing.T) {
	t.Parallel()

	require.Empty(t, paymentMethodFingerprint("wxpay", map[string]string{
		"appid":       "wx-app",
		"mchid":       "merchant",
		"trade_state": "SUCCESS",
	}))
}

func TestPaymentMethodFingerprint_DistinguishesProviderScopes(t *testing.T) {
	t.Parallel()

	alipay := paymentMethodFingerprint("alipay", map[string]string{"buyer_id": "same-id"})
	wxpay := paymentMethodFingerprint("wxpay", map[string]string{"payer_openid": "same-id"})

	require.NotEmpty(t, alipay)
	require.NotEmpty(t, wxpay)
	require.NotEqual(t, alipay, wxpay)
}
