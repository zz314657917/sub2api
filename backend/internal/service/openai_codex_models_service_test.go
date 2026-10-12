package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	manifestBody := `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true,"service_tiers":[{"id":"priority","name":"Fast"},{"id":"ultrafast","name":"Ultrafast"}]}]}`
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
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-6-astra","use_responses_lite":true},{"slug":"gpt-5.6-sol","use_responses_lite":true},{"slug":"gpt-5.5","use_responses_lite":true}]}`, etag: `"upstream"`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.1", "")
	require.NoError(t, err)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-5.6-sol"`)
	require.Contains(t, string(manifest.Body), `"use_responses_lite":false`)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-6-astra"`)
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

func TestFetchCodexModelsManifestUsesConfiguredBodyLimit(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6"}]}`}
	service := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 8}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, httpUpstream: upstream}
	account := &Account{ID: 5, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	_, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.0", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "response exceeds 8 bytes")
	require.True(t, IsRetryableCodexModelsManifestError(err))
}

func TestFetchCodexModelsManifestAcceptsConfiguredLimitAboveLegacyBoundary(t *testing.T) {
	body := `{"models":[{"slug":"gpt-5.6","display_name":"` + strings.Repeat("x", (8<<20)+1024) + `"}]}`
	upstream := &codexManifestHTTPStub{body: body}
	service := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 16 << 20}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, httpUpstream: upstream}
	account := &Account{ID: 6, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.0", "")
	require.NoError(t, err)
	require.Equal(t, body, string(manifest.Body))
}

func TestCodexModelsManifestETagMatchesWeakAndMultipleValues(t *testing.T) {
	require.True(t, codexModelsManifestETagMatches(`"other", W/"abc"`, `"abc"`))
	require.True(t, codexModelsManifestETagMatches("*", `"abc"`))
	require.False(t, codexModelsManifestETagMatches(`"other"`, `"abc"`))
}

func TestFetchCodexModelsManifestGPT61ContextDefault(t *testing.T) {
	for _, openAIList := range []bool{false, true} {
		t.Run(fmt.Sprintf("openai_list=%t", openAIList), func(t *testing.T) {
			entries := `[
				{"slug":"gpt-6.1-sol","max_context_window":1000000,"auto_compact_token_limit":250000,"extra":{"keep":true}},
				{"slug":"openai/gpt_6.1_sol_high","context_window":null,"use_responses_lite":true},
				{"slug":"gpt-6.1-sol-max","context_window":272000,"max_context_window":272000},
				{"slug":"gpt-6.1-sol-low","context_window":900000},
				{"slug":"gpt-6.1-sol-medium","context_window":0},
				{"slug":"gpt-6.1-sol-unknown","context_window":null},
				{"slug":"gpt-6-sol"},
				{"slug":"gpt-5.6-sol","use_responses_lite":true}]`
			body := `{"models":` + entries + `,"extra":"keep"}`
			if openAIList {
				body = `{"object":"list","data":` + strings.ReplaceAll(entries, `"slug"`, `"id"`) + `}`
			}
			upstream := &codexManifestHTTPStub{body: body, etag: `W/"context-source"`}
			svc := &OpenAIGatewayService{
				cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
				httpUpstream: upstream,
			}
			account := &Account{ID: 61, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test-key", "base_url": "https://provider.example/v1"}}
			manifest, err := svc.FetchCodexModelsManifest(context.Background(), account, "", "")
			require.NoError(t, err)
			var got struct {
				Models []map[string]json.RawMessage `json:"models"`
				Extra  string                       `json:"extra"`
			}
			require.NoError(t, json.Unmarshal(manifest.Body, &got))
			require.Len(t, got.Models, 8)
			require.JSONEq(t, `500000`, string(got.Models[0]["context_window"]))
			require.JSONEq(t, `1000000`, string(got.Models[0]["max_context_window"]))
			require.JSONEq(t, `250000`, string(got.Models[0]["auto_compact_token_limit"]))
			require.JSONEq(t, `{"keep":true}`, string(got.Models[0]["extra"]))
			require.JSONEq(t, `500000`, string(got.Models[1]["context_window"]))
			require.JSONEq(t, `true`, string(got.Models[1]["use_responses_lite"]))
			require.JSONEq(t, `272000`, string(got.Models[2]["context_window"]))
			require.JSONEq(t, `272000`, string(got.Models[2]["max_context_window"]))
			require.JSONEq(t, `900000`, string(got.Models[3]["context_window"]))
			require.JSONEq(t, `0`, string(got.Models[4]["context_window"]))
			require.JSONEq(t, `null`, string(got.Models[5]["context_window"]))
			require.NotContains(t, got.Models[6], "context_window")
			require.NotContains(t, got.Models[7], "context_window")
			require.JSONEq(t, `false`, string(got.Models[7]["use_responses_lite"]))
			if !openAIList {
				require.Equal(t, "keep", got.Extra)
			}
			require.Equal(t, fmt.Sprintf(`"%x"`, sha256.Sum256(manifest.Body)), manifest.ETag)
			require.NotEqual(t, upstream.etag, manifest.ETag)

			conditional, err := svc.FetchCodexModelsManifest(context.Background(), account, "", manifest.ETag)
			require.NoError(t, err)
			require.True(t, conditional.NotModified)
			require.Empty(t, conditional.Body)

			// Expire the cache, then prove refresh sends the original upstream
			// validator and preserves the adapted body/hash after upstream 304.
			svc.codexModelsManifestCache.mu.Lock()
			var cacheKey string
			for key, entry := range svc.codexModelsManifestCache.entries {
				cacheKey = key
				entry.expiresAt = time.Now().Add(-time.Minute)
				entry.staleUntil = time.Now().Add(time.Minute)
				svc.codexModelsManifestCache.entries[key] = entry
			}
			svc.codexModelsManifestCache.mu.Unlock()
			require.NotEmpty(t, cacheKey)
			req := codexModelsManifestRequest{url: upstream.lastReq.URL.String(), headers: upstream.lastReq.Header.Clone(),
				credentialAccount: account, accountID: account.ID, useAPIKeyUpstream: true}
			upstream.status = http.StatusNotModified
			refreshed := <-svc.refreshCachedAPIKeyCodexModelsManifest(cacheKey, req)
			require.NoError(t, refreshed.Err)
			require.Equal(t, `W/"context-source"`, upstream.lastReq.Header.Get("If-None-Match"))
			require.Equal(t, manifest, refreshed.Val)
			again, err := svc.FetchCodexModelsManifest(context.Background(), account, "", "")
			require.NoError(t, err)
			require.Equal(t, manifest.Body, again.Body)
			require.Equal(t, manifest.ETag, again.ETag)
		})
	}
}

func TestFetchCodexModelsManifestGPT61OAuthContextVerbatim(t *testing.T) {
	for _, field := range []string{"", `,"context_window":null`, `,"context_window":272000`, `,"context_window":900000`} {
		t.Run(field, func(t *testing.T) {
			body := "{ \"models\": [{\"slug\":\"gpt-6.1-sol\"" + field + `,"max_context_window":1000000,"use_responses_lite":true}] }`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("If-None-Match") == `W/"oauth-context"` {
					w.Header().Set("ETag", `W/"oauth-context"`)
					w.WriteHeader(http.StatusNotModified)
					return
				}
				w.Header().Set("ETag", `W/"oauth-context"`)
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			original := chatgptCodexModelsURL
			chatgptCodexModelsURL = server.URL
			defer func() { chatgptCodexModelsURL = original }()
			account := &Account{ID: 62, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
				Credentials: map[string]any{"access_token": "test-token"}}
			svc := &OpenAIGatewayService{}
			manifest, err := svc.FetchCodexModelsManifest(context.Background(), account, "", "")
			require.NoError(t, err)
			require.Equal(t, body, string(manifest.Body))
			require.Equal(t, `W/"oauth-context"`, manifest.ETag)
			conditional, err := svc.FetchCodexModelsManifest(context.Background(), account, "", manifest.ETag)
			require.NoError(t, err)
			require.True(t, conditional.NotModified)
		})
	}
}

func TestFetchCodexModelsManifestGPT61ExplicitAPIKeyRepresentationUnchanged(t *testing.T) {
	body := `{ "extra": true, "models": [
		{"slug":"gpt-6.1-sol","context_window":272000,"max_context_window":1000000,"auto_compact_token_limit":250000},
		{"slug":"gpt-6.1-sol-max","context_window":900000},
		{"slug":"gpt-6.1-sol-low","context_window":0},
		{"slug":"gpt-6.1-sol-unknown","context_window":null},
		{"slug":"gpt-6-sol"}] }`
	upstream := &codexManifestHTTPStub{body: body, etag: `"explicit-context"`}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 63, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "test-key", "base_url": "https://provider.example/v1"}}
	manifest, err := svc.FetchCodexModelsManifest(context.Background(), account, "", "")
	require.NoError(t, err)
	require.Equal(t, body, string(manifest.Body))
	require.Equal(t, upstream.etag, manifest.ETag)
}
