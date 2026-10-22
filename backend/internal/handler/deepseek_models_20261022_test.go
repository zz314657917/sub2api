package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestDeepSeekCompat20261022DefaultModels(t *testing.T) {
	models := defaultModelIDsForPlatform(service.PlatformDeepseek)
	for _, id := range []string{"deepseek-chat", "deepseek-reasoner", "deepseek-v4-flash", "deepseek-v4-pro", "deepseek-v4.1-flash"} {
		require.Contains(t, models, id)
	}
	for _, id := range models {
		require.NotContains(t, id, "claude")
	}
	require.Empty(t, defaultImageModelIDsForPlatform(service.PlatformDeepseek))
}

func TestDeepSeekCompat20261022ModelsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, custom := range []bool{false, true} {
		groupID := int64(20261022)
		account := service.Account{ID: 20261022, Platform: service.PlatformDeepseek}
		if custom {
			account.Credentials = map[string]any{"model_mapping": map[string]any{"local-custom": "deepseek-v4.1-flash"}}
		}
		h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{groupID: {account}},
		})
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
			Group: &service.Group{ID: groupID, Platform: service.PlatformDeepseek},
		})
		h.Models(c)
		require.Equal(t, http.StatusOK, rec.Code)
		var got gatewayModelsResponseForTest
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		ids := modelIDsForTest(got.Data)
		if custom {
			require.Equal(t, []string{"local-custom"}, ids)
		} else {
			require.Contains(t, ids, "deepseek-v4.1-flash")
			require.NotContains(t, ids, "claude-sonnet-4-6")
		}
	}
}

func TestDeepSeekCompat20261022NativeRequirementInference(t *testing.T) {
	nativeBody := []byte(`{"stream":true,"input":[{"type":"compaction_trigger"}]}`)
	native := newCompactBodySignalTestContext(t, "/v1/responses", nativeBody)
	require.True(t, isBareOpenAIResponsesPath(native) && isOpenAIRemoteCompactionV2Request(nativeBody))
	legacy := newCompactBodySignalTestContext(t, "/v1/responses/compact", []byte(`{}`))
	require.True(t, isOpenAILegacyCompactPath(legacy))
	require.Equal(t, service.OpenAIEndpointCapabilityResponses,
		openAIResponsesRequiredCapabilityForRequest(false, true, service.PlatformDeepseek))
	require.Equal(t, service.OpenAIEndpointCapabilityChatCompletions,
		openAIResponsesRequiredCapabilityForRequest(false, false, service.PlatformDeepseek))
	require.Equal(t, service.OpenAIEndpointCapabilityChatCompletions,
		openAIResponsesRequiredCapabilityForRequest(true, false, service.PlatformDeepseek))
	for _, platform := range []string{service.PlatformKimi, service.PlatformZhipu, service.PlatformGrok} {
		require.Equal(t, service.OpenAIEndpointCapabilityChatCompletions,
			openAIResponsesRequiredCapabilityForRequest(false, true, platform))
	}
	require.Equal(t, service.OpenAIEndpointCapabilityResponses,
		openAIResponsesRequiredCapabilityForRequest(true, false, service.PlatformOpenAI))
}

func TestDeepSeekCompat20261022CNImageGenerationHandlerDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, target := range []struct{ name, platform, protocol string }{
		{"deepseek_chat", service.PlatformDeepseek, service.APIProtocolChatCompletions},
		{"deepseek_responses", service.PlatformDeepseek, service.APIProtocolResponses},
		{"kimi", service.PlatformKimi, service.APIProtocolChatCompletions},
		{"zhipu", service.PlatformZhipu, service.APIProtocolChatCompletions},
	} {
		for _, capability := range []struct {
			name  string
			value any
		}{
			{"missing", nil}, {"empty", []string{}}, {"image", []string{"image"}},
		} {
			t.Run(target.name+"/"+capability.name, func(t *testing.T) {
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"id":"resp_anonymous","object":"response","status":"completed","output":[],"usage":{"input_tokens":1,"output_tokens":1}}`)
				}))
				defer server.Close()
				fixture := newRetryHandlerFixture(t, 1, 0, false)
				fixture.h.imageLimiter = &imageConcurrencyLimiter{}
				account := &fixture.repo.accounts[0]
				account.Platform = target.platform
				account.Credentials["base_url"] = server.URL
				account.Credentials["api_protocol"] = target.protocol
				account.Credentials["model_mapping"] = map[string]any{"deepseek-v4.1-flash": "deepseek-v4.1-flash"}
				account.Extra = nil
				if capability.value != nil {
					account.Extra = map[string]any{"supported_capabilities": capability.value}
				}
				fixture.up.respond = func(req *http.Request, _ int64, _ int) *http.Response {
					resp, err := server.Client().Do(req)
					require.NoError(t, err)
					return resp
				}
				groupID := int64(100)
				key := &service.APIKey{
					ID: 20261022, GroupID: &groupID,
					User: &service.User{ID: 7, Status: service.StatusActive},
					Group: &service.Group{ID: groupID, Status: service.StatusActive,
						Platform: target.platform, AllowImageGeneration: true},
				}
				require.True(t, service.GroupAllowsImageGeneration(key.Group))
				require.Empty(t, key.MultiGroupRouteGroups, "exercise legacy single-group route")
				router := gin.New()
				router.Use(func(c *gin.Context) {
					c.Set(string(middleware2.ContextKeyAPIKey), key)
					c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 7, Concurrency: 1})
					c.Next()
				})
				router.POST("/v1/responses", fixture.h.Responses)
				body := `{"model":"deepseek-v4.1-flash","input":"draw a tree","tools":[{"type":"image_generation"}]}`
				require.True(t, service.IsImageGenerationIntentForPlatform("/v1/responses", "deepseek-v4.1-flash", []byte(body), target.platform))
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body)))
				require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
				require.Equal(t, "invalid_request_error", gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
				require.Contains(t, gjson.GetBytes(rec.Body.Bytes(), "error.message").String(), "Image generation is not supported")
				require.Zero(t, calls.Load())
				require.Empty(t, fixture.up.ids)
				require.Zero(t, fixture.slots.acquired, "reject before account selection/acquisition")
				t.Logf("public Responses legacy single-group platform=%s protocol=%s capabilities=%s allow_image=true status=%d upstream_calls=%d", target.platform, target.protocol, capability.name, rec.Code, calls.Load())
			})
		}
	}
}
