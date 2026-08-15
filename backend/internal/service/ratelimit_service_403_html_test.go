package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const openAI403HTMLBody = "<!DOCTYPE html><html><body>403 Forbidden</body></html>"

type html403AccountRepoStub struct {
	AccountRepository
	setErrorCalls int
	tempCalls     int
}

func (s *html403AccountRepoStub) SetError(context.Context, int64, string) error {
	s.setErrorCalls++
	return nil
}

func (s *html403AccountRepoStub) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	s.tempCalls++
	return nil
}

type html403CounterStub struct {
	increments int
	count      int64
}

func (s *html403CounterStub) IncrementOpenAI403Count(context.Context, int64, int) (int64, error) {
	s.increments++
	return s.count, nil
}

func (s *html403CounterStub) ResetOpenAI403Count(context.Context, int64) error { return nil }

func newHTML403Service(count int64) (*RateLimitService, *html403AccountRepoStub, *html403CounterStub) {
	repo := &html403AccountRepoStub{}
	counter := &html403CounterStub{count: count}
	svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	svc.SetOpenAI403CounterCache(counter)
	return svc, repo, counter
}

func TestHandleUpstreamError_OpenAIHTML403(t *testing.T) {
	svc, repo, counter := newHTML403Service(1)
	account := &Account{ID: 501, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(openAI403HTMLBody)))
	require.Zero(t, counter.increments)
	require.Zero(t, repo.tempCalls)
	require.Zero(t, repo.setErrorCalls)
}

func TestHandleUpstreamError_OpenAIStructured403(t *testing.T) {
	t.Run("structured_json", func(t *testing.T) {
		svc, repo, counter := newHTML403Service(1)
		account := &Account{ID: 502, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
		require.True(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(`{"error":{"message":"forbidden"}}`)))
		require.Equal(t, 1, counter.increments)
		require.Equal(t, 1, repo.tempCalls)
		require.Zero(t, repo.setErrorCalls)
	})
	t.Run("plain_text", func(t *testing.T) {
		svc, repo, counter := newHTML403Service(openAI403DisableThreshold)
		account := &Account{ID: 503, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
		require.True(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte("Forbidden")))
		require.Equal(t, 1, counter.increments)
		require.Zero(t, repo.tempCalls)
		require.Equal(t, 1, repo.setErrorCalls)
	})
}

func TestHandleUpstreamError_HTML403OnOtherPlatformsUnchanged(t *testing.T) {
	svc, repo, counter := newHTML403Service(1)
	account := &Account{ID: 504, Platform: PlatformAnthropic, Type: AccountTypeAPIKey}
	require.True(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(openAI403HTMLBody)))
	require.Zero(t, counter.increments)
	require.Zero(t, repo.tempCalls)
	require.Equal(t, 1, repo.setErrorCalls)
}

func TestIsHTMLResponse(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{"<!doctype html><html></html>", true},
		{" \n<HTML lang=\"en\">", true},
		{`{"error":{"message":"forbidden"}}`, false},
		{"Forbidden", false},
	} {
		require.Equal(t, tc.want, isHTMLResponse([]byte(tc.body)))
	}
}
