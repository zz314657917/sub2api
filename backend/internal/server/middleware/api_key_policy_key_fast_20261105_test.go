package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type keyFast20261105MiddlewareRepo struct {
	*openAIFastPolicyForwardingAPIKeyRepo
	loads int
}

func (r *keyFast20261105MiddlewareRepo) GetByKeyForAuth(ctx context.Context, key string) (*service.APIKey, error) {
	r.loads++
	return r.openAIFastPolicyForwardingAPIKeyRepo.GetByKeyForAuth(ctx, key)
}

func TestKeyFast20261105MiddlewareRedisContextAndRealHTTPTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"native-apikey", "native-oauth", "passthrough-oauth"} {
		t.Run(mode, func(t *testing.T) {
			captures := make(chan openAIFastPolicyForwardingCapture, 16)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				captures <- openAIFastPolicyForwardingCapture{body: body, header: r.Header.Clone()}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, "event: response.completed\ndata: "+`{"type":"response.completed","response":{"id":"resp_key_fast","model":"gpt-5.5","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}}`+"\n\n")
			}))
			defer upstream.Close()
			cfg := &config.Config{RunMode: config.RunModeSimple,
				APIKeyAuth: config.APIKeyAuthCacheConfig{L2TTLSeconds: 60}}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			settingsJSON, err := json.Marshal(service.DefaultOpenAIFastPolicySettings())
			require.NoError(t, err)
			settings := service.NewSettingService(&openAIFastPolicyForwardingSettingRepo{value: string(settingsJSON)}, cfg)
			gateway := service.NewOpenAIGatewayService(nil, nil, nil, nil, nil, nil, nil, cfg,
				nil, nil, nil, nil, nil, newOpenAIFastPolicyForwardingHTTPUpstream(t, upstream),
				nil, nil, nil, nil, nil, nil, settings, nil)
			group := &service.Group{ID: 101, Status: service.StatusActive, Platform: service.PlatformOpenAI, Hydrated: true}
			keys := map[string]*service.APIKey{}
			for _, policy := range []string{"force", "filter", "passthrough"} {
				key := newOpenAIFastPolicyForwardingAPIKey(7, policy, 42, 101, group)
				key.OpenAIFastPolicy = policy
				keys[policy] = key
			}
			repo := &keyFast20261105MiddlewareRepo{openAIFastPolicyForwardingAPIKeyRepo: &openAIFastPolicyForwardingAPIKeyRepo{apiKeys: keys}}
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			keyService := service.NewAPIKeyService(repo, nil, nil, nil, nil, repository.NewAPIKeyCache(rdb), cfg)
			account := &service.Account{ID: 9, Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth,
				Status: service.StatusActive, Schedulable: true, Concurrency: 1,
				Credentials: map[string]any{"api_key": "test", "access_token": "test", "chatgpt_account_id": "test"},
				Extra:       map[string]any{"openai_passthrough": mode == "passthrough-oauth"}}
			if mode == "native-apikey" {
				account.Type = service.AccountTypeAPIKey
			}
			var finalTier *string
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keyService, nil, cfg)))
			router.POST("/v1/responses", func(c *gin.Context) {
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				authKey, ok := GetAPIKeyFromContext(c)
				require.True(t, ok)
				require.Equal(t, authKey.OpenAIFastPolicy, c.Request.Context().Value(ctxkey.APIKeyOpenAIFastPolicy))
				result, err := gateway.Forward(c.Request.Context(), c, account, body)
				require.NoError(t, err)
				finalTier = result.ServiceTier
			})
			for _, policy := range []string{"force", "filter", "passthrough"} {
				for i := 0; i < 2; i++ {
					// The second request must hydrate the policy from Redis.
					body := `{"model":"gpt-5.5","stream":false,"input":"hi","instructions":"test","service_tier":"priority","openai_fast_policy":"force"}`
					if policy == "force" {
						body = strings.Replace(body, `"service_tier":"priority"`, `"service_tier":"flex"`, 1)
					}
					req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+policy)
					req.Header.Set("OpenAI-Fast-Policy", "force")
					req.Header.Set("X-Sub2api-OpenAI-Fast-Policy", "force")
					req.Header.Set("X-OpenAI-Codex-Routing-Hint", "spoof")
					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, req)
					require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
					got := <-captures
					if policy == "filter" {
						require.False(t, gjson.GetBytes(got.body, "service_tier").Exists(), "body/header policy spoof must not override trusted filter")
						require.Nil(t, finalTier)
					} else {
						require.Equal(t, "priority", gjson.GetBytes(got.body, "service_tier").String())
						require.NotNil(t, finalTier)
						require.Equal(t, "priority", *finalTier)
					}
					require.Empty(t, got.header.Get("X-Sub2api-OpenAI-Fast-Policy"))
				}
			}
			require.Equal(t, 3, repo.loads, "each policy must be re-used from the auth cache")
		})
	}
}

func TestKeyFast20261105ContextRefreshOnEffectiveKeyChange(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	for _, policy := range []string{"force", "filter", "", "passthrough"} {
		SetAPIKeyContext(c, &service.APIKey{ID: 7, UserID: 42, OpenAIFastPolicy: policy})
		require.Equal(t, service.NormalizeAPIKeyOpenAIFastPolicy(policy), c.Request.Context().Value(ctxkey.APIKeyOpenAIFastPolicy))
	}
}
