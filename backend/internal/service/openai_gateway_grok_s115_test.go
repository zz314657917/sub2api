package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleGrokAccountUpstreamErrorDefaultCooldownsRespectPoolMode(t *testing.T) {
	testCases := []struct {
		name     string
		status   int
		headers  http.Header
		reason   string
		cooldown time.Duration
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, reason: "grok oauth token unauthorized", cooldown: 10 * time.Minute},
		{name: "payment required", status: http.StatusPaymentRequired, reason: "grok payment required", cooldown: 30 * time.Minute},
		{name: "forbidden", status: http.StatusForbidden, reason: "grok entitlement or subscription tier denied", cooldown: 30 * time.Minute},
		{name: "rate limited", status: http.StatusTooManyRequests, headers: http.Header{"Retry-After": []string{"45"}}, reason: "grok rate limited", cooldown: 45 * time.Second},
		{name: "upstream error", status: http.StatusBadGateway, reason: "grok upstream temporary error", cooldown: 2 * time.Minute},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" pool mode keeps scheduling state", func(t *testing.T) {
			account := &Account{
				ID:       int64(611 + tc.status),
				Platform: PlatformGrok,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"pool_mode": true,
				},
			}

			(&OpenAIGatewayService{}).handleGrokAccountUpstreamError(
				context.Background(), account, tc.status, tc.headers, nil,
			)

			require.Nil(t, account.TempUnschedulableUntil)
			require.Empty(t, account.TempUnschedulableReason)
		})

		t.Run(tc.name+" non-pool mode keeps cooldown", func(t *testing.T) {
			account := &Account{ID: int64(712 + tc.status), Platform: PlatformGrok, Type: AccountTypeAPIKey}
			before := time.Now()

			(&OpenAIGatewayService{}).handleGrokAccountUpstreamError(
				context.Background(), account, tc.status, tc.headers, nil,
			)

			require.NotNil(t, account.TempUnschedulableUntil)
			require.Equal(t, tc.reason, account.TempUnschedulableReason)
			require.WithinDuration(t, before.Add(tc.cooldown), *account.TempUnschedulableUntil, time.Second)
		})
	}
}
