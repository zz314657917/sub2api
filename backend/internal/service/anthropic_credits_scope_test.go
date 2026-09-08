package service

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type creditsScopeRepo struct {
	AccountRepository
	modelCalls, accountCalls int
	scope                    string
	until                    time.Time
}

func (r *creditsScopeRepo) SetModelRateLimit(_ context.Context, _ int64, scope string, until time.Time) error {
	r.modelCalls++
	r.scope, r.until = scope, until
	return nil
}
func (r *creditsScopeRepo) SetRateLimited(_ context.Context, _ int64, _ time.Time) error {
	r.accountCalls++
	return nil
}
func (r *creditsScopeRepo) UpdateSessionWindow(context.Context, int64, *time.Time, *time.Time, string) error {
	return nil
}

func TestFableCreditsScopeRequestedModelAndSharedWindow(t *testing.T) {
	for _, shared := range []bool{false, true} {
		t.Run(strconv.FormatBool(shared), func(t *testing.T) {
			repo := &creditsScopeRepo{}
			svc := NewRateLimitService(repo, nil, nil, nil, nil)
			headers := http.Header{}
			reset := time.Now().Add(time.Hour).Truncate(time.Second)
			if shared {
				headers.Set("anthropic-ratelimit-unified-5h-status", "rejected")
				headers.Set("anthropic-ratelimit-unified-5h-reset", strconv.FormatInt(reset.Unix(), 10))
			}
			started := time.Now()
			body := []byte(`{"error":{"details":{"error_code":"credits_required"}}}`)
			account := &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth}
			require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, headers, body, "claude-fable-5-1[1m]"))
			require.Equal(t, 1, repo.modelCalls)
			require.Equal(t, anthropicFableRateLimitKey, repo.scope)
			require.True(t, repo.until.After(started))
			if shared {
				require.Equal(t, 1, repo.accountCalls)
			} else {
				require.Zero(t, repo.accountCalls)
			}
		})
	}
}
