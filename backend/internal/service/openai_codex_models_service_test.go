package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type codexManifestHTTPStub struct {
	body    string
	status  int
	etag    string
	lastReq *http.Request
}

func (s *codexManifestHTTPStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.lastReq = req
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	header := make(http.Header)
	if s.etag != "" {
		header.Set("ETag", s.etag)
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func TestFetchCodexModelsManifestAPIKeyFallsBackForStaleHeaderVersion(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6-sol"}]}`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 4, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}

	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.125.0", "")

	require.NoError(t, err)
	require.NotNil(t, manifest)
	require.Equal(t, "0.125.0", upstream.lastReq.URL.Query().Get("client_version"))
	require.Equal(t, codexCLIVersion, upstream.lastReq.Header.Get("Version"))
}

func (s *codexManifestHTTPStub) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxyURL, accountID, accountConcurrency)
}

func TestFetchCodexModelsManifestOAuthPassesThroughVerbatim(t *testing.T) {
	manifestBody := `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer oauth-token", r.Header.Get("Authorization"))
		require.Equal(t, openai.CodexDefaultOriginator, r.Header.Get("Originator"))
		require.Equal(t, codexCLIUserAgent, r.Header.Get("User-Agent"))
		require.Equal(t, "0.144.1", r.URL.Query().Get("client_version"))
		w.Header().Set("ETag", `W/"oauth"`)
		_, _ = w.Write([]byte(manifestBody))
	}))
	defer server.Close()
	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "oauth-token"}}
	manifest, err := (&OpenAIGatewayService{}).FetchCodexModelsManifest(context.Background(), account, "0.144.1", "")
	require.NoError(t, err)
	require.Equal(t, manifestBody, string(manifest.Body))
	require.Equal(t, `W/"oauth"`, manifest.ETag)
}

func TestFetchCodexModelsManifestAPIKeyAdjustsOnlyTargetedModels(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true},{"slug":"gpt-5.5","use_responses_lite":true}]}`, etag: `"upstream"`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.1", "")
	require.NoError(t, err)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-5.6-sol"`)
	require.Contains(t, string(manifest.Body), `"use_responses_lite":false`)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-5.5","use_responses_lite":true`)
	require.NotEqual(t, `"upstream"`, manifest.ETag)
}

func TestFetchCodexModelsManifestAPIKeyConvertsOpenAIModelList(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"object":"list","data":[{"id":"gpt-5.6-sol"}]}`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"models":[{"slug":"gpt-5.6-sol"}]}`, string(manifest.Body))
}

func TestCodexModelsManifestETagMatchesWeakAndMultipleValues(t *testing.T) {
	require.True(t, codexModelsManifestETagMatches(`"other", W/"abc"`, `"abc"`))
	require.True(t, codexModelsManifestETagMatches("*", `"abc"`))
	require.False(t, codexModelsManifestETagMatches(`"other"`, `"abc"`))
}
