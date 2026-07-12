package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAPIKeyAuthForwardsUserScopedOpenAIFastPolicyToManagedAndPassthroughHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name        string
		accountType string
		passthrough bool
	}{
		{name: "managed API key", accountType: service.AccountTypeAPIKey},
		{name: "API key passthrough", accountType: service.AccountTypeAPIKey, passthrough: true},
		{name: "OAuth passthrough", accountType: service.AccountTypeOAuth, passthrough: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			captures := make(chan openAIFastPolicyForwardingCapture, 2)
			upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "read request body", http.StatusInternalServerError)
					return
				}
				captures <- openAIFastPolicyForwardingCapture{
					body:   append([]byte(nil), body...),
					header: r.Header.Clone(),
				}
				if gjson.GetBytes(body, "stream").Bool() {
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_user_scope\",\"model\":\"gpt-5.5\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"input_tokens_details\":{\"cached_tokens\":0}}}}\n\ndata: [DONE]\n\n")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"id":"resp_user_scope","object":"response","model":"gpt-5.5","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2,"input_tokens_details":{"cached_tokens":0}}}`)
			}))
			defer upstreamServer.Close()

			settings := &service.OpenAIFastPolicySettings{
				Rules: []service.OpenAIFastPolicyRule{
					{
						ServiceTier: service.OpenAIFastTierPriority,
						Action:      service.BetaPolicyActionFilter,
						Scope:       service.BetaPolicyScopeAll,
					},
					{
						ServiceTier: service.OpenAIFastTierPriority,
						Action:      service.BetaPolicyActionPass,
						Scope:       service.BetaPolicyScopeAll,
						UserIDs:     []int64{42},
					},
				},
			}
			settingsJSON, err := json.Marshal(settings)
			require.NoError(t, err)

			cfg := &config.Config{RunMode: config.RunModeSimple}
			cfg.Security.URLAllowlist.Enabled = false
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			settingService := service.NewSettingService(&openAIFastPolicyForwardingSettingRepo{value: string(settingsJSON)}, cfg)
			captureUpstream := newOpenAIFastPolicyForwardingHTTPUpstream(t, upstreamServer)
			gatewayService := service.NewOpenAIGatewayService(
				nil, nil, nil, nil, nil, nil, nil, cfg,
				nil, nil, nil, nil, nil, captureUpstream,
				nil, nil, nil, nil, nil, nil, settingService, nil,
			)

			groupID := int64(101)
			group := &service.Group{
				ID:       groupID,
				Name:     "openai",
				Status:   service.StatusActive,
				Platform: service.PlatformOpenAI,
				Hydrated: true,
			}
			apiKeys := map[string]*service.APIKey{
				"key-user-42": newOpenAIFastPolicyForwardingAPIKey(1, "key-user-42", 42, groupID, group),
				"key-user-43": newOpenAIFastPolicyForwardingAPIKey(2, "key-user-43", 43, groupID, group),
			}
			apiKeyService := service.NewAPIKeyService(&openAIFastPolicyForwardingAPIKeyRepo{apiKeys: apiKeys}, nil, nil, nil, nil, nil, cfg)

			account := &service.Account{
				ID:          900,
				Name:        tc.name,
				Platform:    service.PlatformOpenAI,
				Type:        tc.accountType,
				Status:      service.StatusActive,
				Schedulable: true,
				Concurrency: 1,
				Extra: map[string]any{
					"use_responses_api":  true,
					"openai_passthrough": tc.passthrough,
				},
			}
			if tc.accountType == service.AccountTypeOAuth {
				account.Credentials = map[string]any{
					"access_token":       "oauth-token",
					"chatgpt_account_id": "chatgpt-account",
				}
			} else {
				account.Credentials = map[string]any{
					"api_key":  "sk-test",
					"base_url": upstreamServer.URL,
				}
			}

			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(apiKeyService, nil, cfg)))
			router.POST("/v1/responses", func(c *gin.Context) {
				body, readErr := io.ReadAll(c.Request.Body)
				if readErr != nil {
					c.Status(http.StatusBadRequest)
					return
				}
				service.SetOpenAIClientTransport(c, service.OpenAIClientTransportHTTP)
				if _, forwardErr := gatewayService.Forward(c.Request.Context(), c, account, body); forwardErr != nil {
					c.Status(http.StatusBadGateway)
					return
				}
				c.Status(http.StatusOK)
			})

			send := func(apiKey string) {
				request := httptest.NewRequest(
					http.MethodPost,
					"/v1/responses",
					bytes.NewBufferString(`{"model":"gpt-5.5","stream":false,"service_tier":"priority","instructions":"test","user_id":42,"metadata":{"user_id":42},"input":"hi"}`),
				)
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("x-api-key", apiKey)
				request.Header.Set("x-sub2api-user-id", "42")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			}

			send("key-user-42")
			send("key-user-43")

			matchingUserCapture := <-captures
			differentUserCapture := <-captures
			require.Equal(t, int64(42), gjson.GetBytes(matchingUserCapture.body, "user_id").Int())
			require.Equal(t, int64(42), gjson.GetBytes(differentUserCapture.body, "user_id").Int())
			require.Empty(t, matchingUserCapture.header.Get("x-sub2api-user-id"))
			require.Empty(t, differentUserCapture.header.Get("x-sub2api-user-id"))
			require.Equal(t, service.OpenAIFastTierPriority, gjson.GetBytes(matchingUserCapture.body, "service_tier").String())
			require.False(t, gjson.GetBytes(differentUserCapture.body, "service_tier").Exists(), "spoofed body/header identity must not bypass the trusted API key owner")
		})
	}
}

type openAIFastPolicyForwardingCapture struct {
	body   []byte
	header http.Header
}

func newOpenAIFastPolicyForwardingAPIKey(id int64, key string, userID, groupID int64, group *service.Group) *service.APIKey {
	return &service.APIKey{
		ID:      id,
		UserID:  userID,
		Key:     key,
		Status:  service.StatusActive,
		GroupID: &groupID,
		User: &service.User{
			ID:          userID,
			Role:        service.RoleUser,
			Status:      service.StatusActive,
			Balance:     10,
			Concurrency: 1,
		},
		Group: group,
	}
}

type openAIFastPolicyForwardingAPIKeyRepo struct {
	service.APIKeyRepository
	apiKeys map[string]*service.APIKey
}

func (r *openAIFastPolicyForwardingAPIKeyRepo) GetByKeyForAuth(_ context.Context, key string) (*service.APIKey, error) {
	apiKey, ok := r.apiKeys[key]
	if !ok {
		return nil, service.ErrAPIKeyNotFound
	}
	clone := *apiKey
	return &clone, nil
}

func (r *openAIFastPolicyForwardingAPIKeyRepo) UpdateLastUsed(context.Context, int64, time.Time) error {
	return nil
}

type openAIFastPolicyForwardingSettingRepo struct {
	service.SettingRepository
	value string
}

func (r *openAIFastPolicyForwardingSettingRepo) GetValue(context.Context, string) (string, error) {
	return r.value, nil
}

type openAIFastPolicyForwardingHTTPUpstream struct {
	client *http.Client
	target *url.URL
}

func newOpenAIFastPolicyForwardingHTTPUpstream(t *testing.T, server *httptest.Server) *openAIFastPolicyForwardingHTTPUpstream {
	t.Helper()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	return &openAIFastPolicyForwardingHTTPUpstream{client: server.Client(), target: target}
}

func (u *openAIFastPolicyForwardingHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	forwarded := req.Clone(req.Context())
	forwarded.URL.Scheme = u.target.Scheme
	forwarded.URL.Host = u.target.Host
	forwarded.Host = ""
	return u.client.Do(forwarded)
}

func (u *openAIFastPolicyForwardingHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}
