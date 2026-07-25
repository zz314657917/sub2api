package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleGrokAccountUpstreamError5xxRespectsPoolMode(t *testing.T) {
	t.Run("pool mode keeps scheduling state", func(t *testing.T) {
		account := &Account{
			ID:       611,
			Platform: PlatformGrok,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"pool_mode": true,
			},
		}
		svc := &OpenAIGatewayService{}

		svc.handleGrokAccountUpstreamError(context.Background(), account, http.StatusBadGateway, nil, nil)

		require.Nil(t, account.TempUnschedulableUntil)
		require.Empty(t, account.TempUnschedulableReason)
	})

	t.Run("non-pool mode keeps two minute cooldown", func(t *testing.T) {
		account := &Account{ID: 612, Platform: PlatformGrok, Type: AccountTypeAPIKey}
		svc := &OpenAIGatewayService{}
		before := time.Now()

		svc.handleGrokAccountUpstreamError(context.Background(), account, http.StatusBadGateway, nil, nil)

		require.NotNil(t, account.TempUnschedulableUntil)
		require.Equal(t, "grok upstream temporary error", account.TempUnschedulableReason)
		require.WithinDuration(t, before.Add(2*time.Minute), *account.TempUnschedulableUntil, time.Second)
	})
}
