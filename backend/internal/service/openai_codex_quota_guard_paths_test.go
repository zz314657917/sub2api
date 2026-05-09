package service

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func exhaustedOpenAICodex7dHeadersForTest() http.Header {
	headers := http.Header{}
	headers.Set("x-codex-primary-used-percent", "100")
	headers.Set("x-codex-primary-reset-after-seconds", "3600")
	headers.Set("x-codex-primary-window-minutes", "10080")
	headers.Set("x-codex-secondary-used-percent", "12")
	headers.Set("x-codex-secondary-reset-after-seconds", "1200")
	headers.Set("x-codex-secondary-window-minutes", "300")
	return headers
}

func TestRateLimitService_Handle429_OpenAIOAuth7dExhaustedTempBlocksWithoutRateLimit(t *testing.T) {
	account := Account{ID: 901, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	repo := &openAICodexSnapshotAsyncRepo{
		stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{account}},
		updateExtraCh:         make(chan map[string]any, 1),
		rateLimitCh:           make(chan time.Time, 1),
		tempUnschedCh:         make(chan codexTempUnschedCall, 1),
	}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)

	svc.handle429(context.Background(), &account, exhaustedOpenAICodex7dHeadersForTest(), nil)

	select {
	case updates := <-repo.updateExtraCh:
		require.Equal(t, 100.0, updates["codex_7d_used_percent"])
	case <-time.After(2 * time.Second):
		t.Fatal("expected codex snapshot to be persisted")
	}

	select {
	case call := <-repo.tempUnschedCh:
		require.WithinDuration(t, time.Now().Add(time.Hour), call.until, 3*time.Second)
		require.Contains(t, call.reason, "OpenAI Codex 7d usage reached 100%")
	case <-time.After(2 * time.Second):
		t.Fatal("expected OpenAI OAuth 7d exhaustion to write temp_unschedulable_until")
	}

	select {
	case resetAt := <-repo.rateLimitCh:
		t.Fatalf("OpenAI OAuth 7d exhaustion should not write rate_limit_reset_at: %v", resetAt)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestRateLimitService_Handle429_OpenAIAPIKey7dExhaustedDoesNotTempBlock(t *testing.T) {
	account := Account{ID: 902, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	repo := &openAICodexSnapshotAsyncRepo{
		stubOpenAIAccountRepo: stubOpenAIAccountRepo{accounts: []Account{account}},
		updateExtraCh:         make(chan map[string]any, 1),
		rateLimitCh:           make(chan time.Time, 1),
		tempUnschedCh:         make(chan codexTempUnschedCall, 1),
	}
	svc := NewRateLimitService(repo, nil, nil, nil, nil)

	svc.handle429(context.Background(), &account, exhaustedOpenAICodex7dHeadersForTest(), nil)

	select {
	case <-repo.updateExtraCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected codex snapshot to be persisted")
	}

	select {
	case <-repo.rateLimitCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected OpenAI API key 429 to keep normal rate-limit behavior")
	}

	select {
	case call := <-repo.tempUnschedCh:
		t.Fatalf("OpenAI API key should not write temp_unschedulable_until: %+v", call)
	case <-time.After(200 * time.Millisecond):
	}
}

func requireCodexTempBlock(t *testing.T, ch <-chan codexTempUnschedCall) {
	t.Helper()

	select {
	case call := <-ch:
		require.WithinDuration(t, time.Now().Add(time.Hour), call.until, 3*time.Second)
		if !strings.Contains(call.reason, "OpenAI Codex 7d usage reached 100%") {
			t.Fatalf("unexpected temp reason: %s", call.reason)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected OpenAI OAuth 7d exhaustion to write temp_unschedulable_until")
	}
}
