package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleUpstreamError_CNProviderHTML403SkipsAccountPenalty(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		svc, repo, counter := newHTML403Service(1)
		account := &Account{ID: 601, Platform: platform, Type: AccountTypeAPIKey}

		require.False(t, svc.HandleUpstreamError(context.Background(), account,
			http.StatusForbidden, http.Header{}, []byte("<html><body>Access denied</body></html>")), platform)
		require.Zero(t, counter.increments, platform)
		require.Zero(t, repo.tempCalls, platform)
		require.Zero(t, repo.setErrorCalls, platform)
	}
}

func TestHandleUpstreamError_CNProviderStructured403TempUnschedulable(t *testing.T) {
	svc, repo, counter := newHTML403Service(1)
	account := &Account{ID: 602, Platform: PlatformKimi, Type: AccountTypeAPIKey}

	require.True(t, svc.HandleUpstreamError(context.Background(), account,
		http.StatusForbidden, http.Header{}, []byte(`{"error":{"message":"temporary edge rejection"}}`)))
	require.Equal(t, 1, counter.increments)
	require.Equal(t, 1, repo.tempCalls)
	require.Zero(t, repo.setErrorCalls)
}

func TestHandleUpstreamError_CNProviderStructured403ThresholdDisables(t *testing.T) {
	svc, repo, counter := newHTML403Service(openAI403DisableThreshold)
	account := &Account{ID: 603, Platform: PlatformDeepseek, Type: AccountTypeAPIKey}

	require.True(t, svc.HandleUpstreamError(context.Background(), account,
		http.StatusForbidden, http.Header{}, []byte(`{"error":{"message":"provider policy denial"}}`)))
	require.Equal(t, 1, counter.increments)
	require.Zero(t, repo.tempCalls)
	require.Equal(t, 1, repo.setErrorCalls)
}
