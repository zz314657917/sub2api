package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiV027AccountRepo struct {
	service.AccountRepository
	byGroup map[int64][]service.Account
}

type geminiV027ModelsUpstream struct {
	service.HTTPUpstream
	status int
	body   string
}

func (u *geminiV027ModelsUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return &http.Response{StatusCode: u.status, Header: http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"native-id"}}, Body: io.NopCloser(strings.NewReader(u.body))}, nil
}

func (s *geminiV027AccountRepo) ListSchedulableByGroupIDAndPlatforms(_ context.Context, groupID int64, platforms []string) ([]service.Account, error) {
	allowed := make(map[string]struct{}, len(platforms))
	for _, platform := range platforms {
		allowed[platform] = struct{}{}
	}
	var accounts []service.Account
	for _, account := range s.byGroup[groupID] {
		if _, ok := allowed[account.Platform]; ok {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func TestV027GeminiModelListMixedOptOutAndGroupIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID, otherGroupID := int64(502), int64(503)
	repo := &geminiV027AccountRepo{byGroup: map[int64][]service.Account{
		groupID:      {{ID: 1, Platform: service.PlatformAntigravity, Status: service.StatusActive, Schedulable: true, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-opted-out": "upstream"}}}},
		otherGroupID: {{ID: 2, Platform: service.PlatformAntigravity, Status: service.StatusActive, Schedulable: true, Extra: map[string]any{"mixed_scheduling": true}, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-other-group": "upstream"}}}},
	}}
	compat := service.NewGeminiMessagesCompatService(repo, nil, nil, nil, nil, nil, nil, nil, &config.Config{})
	models, err := compat.AntigravityGeminiModelIDs(context.Background(), &groupID, true)
	require.NoError(t, err)
	require.Empty(t, models)
	models, err = compat.AntigravityGeminiModelIDs(context.Background(), &otherGroupID, true)
	require.NoError(t, err)
	require.Contains(t, models, "gemini-other-group")
}

func TestV027GeminiModelListPreservesNativeResponseAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(504)
	repo := &geminiV027AccountRepo{byGroup: map[int64][]service.Account{groupID: {{
		ID: 1, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true,
		Credentials: map[string]any{"api_key": "fake", "base_url": "https://example.test"},
	}}}}
	for _, tt := range []struct {
		status int
		body   string
	}{
		{http.StatusOK, `{"models":[{"name":"models/gemini-native","inputTokenLimit":123}],"nextPageToken":"next"}`},
		{http.StatusTooManyRequests, `{"error":"rate limited"}`},
	} {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			upstream := &geminiV027ModelsUpstream{status: tt.status, body: tt.body}
			cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}
			h := &GatewayHandler{geminiCompatService: service.NewGeminiMessagesCompatService(repo, nil, nil, nil, nil, nil, upstream, nil, cfg)}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models", nil)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{GroupID: &groupID, Group: &service.Group{ID: groupID, Platform: service.PlatformGemini}})
			h.GeminiV1BetaListModels(c)
			require.Equal(t, tt.status, rec.Code)
			if tt.status == http.StatusOK {
				require.Equal(t, "native-id", rec.Header().Get("X-Request-Id"))
				require.Contains(t, rec.Body.String(), `"inputTokenLimit":123`)
				require.Contains(t, rec.Body.String(), `"nextPageToken":"next"`)
			} else {
				require.JSONEq(t, tt.body, rec.Body.String())
			}
		})
	}
}

func TestV027GeminiModelListUsesMappedAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name          string
		forced, mixed bool
	}{
		{"mixed", false, true},
		{"forced", true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			groupID := int64(501)
			repo := &geminiV027AccountRepo{byGroup: map[int64][]service.Account{groupID: {{
				ID: 1, Platform: service.PlatformAntigravity, Status: service.StatusActive, Schedulable: true, Extra: map[string]any{"mixed_scheduling": tt.mixed},
				Credentials: map[string]any{"model_mapping": map[string]any{"gemini-local-custom": "gemini-upstream", "claude-hidden": "claude-upstream"}},
			}}}}
			compat := service.NewGeminiMessagesCompatService(repo, nil, nil, nil, nil, nil, nil, nil, &config.Config{})
			models, err := compat.AntigravityGeminiModelIDs(context.Background(), &groupID, !tt.forced)
			require.NoError(t, err)
			require.Equal(t, []string{"gemini-3-flash", "gemini-3.1-pro-high", "gemini-3.1-pro-low", "gemini-local-custom"}, models)
			h := &GatewayHandler{geminiCompatService: compat}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models", nil)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{GroupID: &groupID, Group: &service.Group{ID: groupID, Platform: service.PlatformGemini}})
			if tt.forced {
				c.Set(string(middleware.ContextKeyForcePlatform), service.PlatformAntigravity)
			}
			h.GeminiV1BetaListModels(c)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var got gemini.ModelsListResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			names := make([]string, 0, len(got.Models))
			for _, model := range got.Models {
				names = append(names, model.Name)
			}
			require.Contains(t, names, "models/gemini-local-custom")
			require.NotContains(t, names, "models/claude-hidden")
		})
	}
}

func TestV027GeminiModelListPreservesUpstreamMetadata(t *testing.T) {
	body := []byte(`{"models":[{"name":"models/gemini-native","inputTokenLimit":123,"custom":{"x":true}}],"nextPageToken":"next"}`)
	merged, ok := appendUpstreamGeminiModels(body, []gemini.Model{gemini.FallbackModel("gemini-native"), gemini.FallbackModel("gemini-local-custom")})
	require.True(t, ok)
	require.Contains(t, string(merged), `"inputTokenLimit":123`)
	require.Contains(t, string(merged), `"nextPageToken":"next"`)
	require.Contains(t, string(merged), "gemini-local-custom")
	for _, invalid := range [][]byte{[]byte(`null`), []byte(`{"models":{}}`)} {
		got, ok := appendUpstreamGeminiModels(invalid, nil)
		require.False(t, ok)
		require.Equal(t, invalid, got)
	}
}
