package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiPoolModeSkippedFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &GeminiMessagesCompatService{}

	poolAccount := func(extra map[string]any) *Account {
		credentials := map[string]any{"pool_mode": true}
		for key, value := range extra {
			credentials[key] = value
		}
		return &Account{ID: 300, Type: AccountTypeAPIKey, Platform: PlatformGemini, Credentials: credentials}
	}

	tests := []struct {
		name              string
		account           *Account
		statusCode        int
		expectFailover    bool
		expectSameAccount bool
	}{
		{"pool_429_same_account_retry", poolAccount(nil), http.StatusTooManyRequests, true, true},
		{"pool_500_failover_only", poolAccount(nil), http.StatusInternalServerError, true, false},
		{"pool_custom_500_same_account_retry", poolAccount(map[string]any{
			"pool_mode_retry_status_codes": []any{float64(http.StatusInternalServerError)},
		}), http.StatusInternalServerError, true, true},
		{"pool_400_no_failover", poolAccount(nil), http.StatusBadRequest, false, false},
		{"non_pool_no_failover", &Account{ID: 301, Type: AccountTypeAPIKey, Platform: PlatformGemini}, http.StatusTooManyRequests, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(writer)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

			failoverErr := svc.poolModeSkippedFailoverError(ctx, tt.account, tt.statusCode, []byte(`{"error":{"message":"upstream"}}`), "req-1")
			if !tt.expectFailover {
				require.Nil(t, failoverErr)
				return
			}
			require.NotNil(t, failoverErr)
			require.Equal(t, tt.statusCode, failoverErr.StatusCode)
			require.Equal(t, tt.expectSameAccount, failoverErr.RetryableOnSameAccount)
		})
	}
}

func TestGeminiPoolModeRetryableOnSameAccount(t *testing.T) {
	poolAccount := &Account{
		Type:        AccountTypeAPIKey,
		Platform:    PlatformGemini,
		Credentials: map[string]any{"pool_mode": true},
	}

	require.True(t, geminiPoolModeRetryableOnSameAccount(poolAccount, http.StatusTooManyRequests))
	require.False(t, geminiPoolModeRetryableOnSameAccount(poolAccount, http.StatusInternalServerError))
	require.False(t, geminiPoolModeRetryableOnSameAccount(nil, http.StatusTooManyRequests))
}
