package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type gatewayModelsAccountRepoStub struct {
	service.AccountRepository

	byGroup map[int64][]service.Account
}

type gatewayModelsResponseForTest struct {
	Object string                    `json:"object"`
	Data   []gatewayModelItemForTest `json:"data"`
}

type gatewayModelItemForTest struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Created   int64  `json:"created"`
	OwnedBy   string `json:"owned_by"`
	CreatedAt string `json:"created_at"`
}

type gatewayModelCatalogResponseForTest struct {
	Object      string                           `json:"object"`
	Items       []gatewayModelCatalogItemForTest `json:"items"`
	ChatModels  []string                         `json:"chat_models"`
	ImageModels []string                         `json:"image_models"`
	VideoModels []string                         `json:"video_models"`
}

type gatewayModelCatalogItemForTest struct {
	ID           string   `json:"id"`
	Capabilities []string `json:"capabilities"`
	Enabled      bool     `json:"enabled"`
}

func (s *gatewayModelsAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]service.Account, error) {
	accounts, ok := s.byGroup[groupID]
	if !ok {
		return nil, nil
	}
	out := make([]service.Account, len(accounts))
	copy(out, accounts)
	return out, nil
}

func newGatewayModelsHandlerForTest(repo service.AccountRepository) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: service.NewGatewayService(
			repo,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		),
	}
}

func TestGatewayModels_GeminiGroupFallsBackToGeminiModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(20)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{ID: 1, Platform: service.PlatformGemini},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: groupID, Platform: service.PlatformGemini},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, "list", got.Object)
	require.Contains(t, modelIDsForTest(got.Data), "gemini-2.5-flash")
	require.NotContains(t, modelIDsForTest(got.Data), "claude-sonnet-4-6")
}

func TestGatewayModels_GeminiGroupFiltersMappedModelsByPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(21)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformAnthropic,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"claude-sonnet-4-6": "claude-sonnet-4-6",
							},
						},
					},
					{
						ID:       2,
						Platform: service.PlatformGemini,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gemini-2.5-flash": "gemini-2.5-flash",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: groupID, Platform: service.PlatformGemini},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gemini-2.5-flash"}, modelIDsForTest(got.Data))
}

func TestGatewayModelCatalog_GroupsMappedModelsByCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(22)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4":     "gpt-5.4",
								"gpt-image-2": "gpt-image-2",
								"sora-2":      "sora-2",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/model-catalog", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: groupID, Platform: service.PlatformOpenAI},
	})

	h.ModelCatalog(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelCatalogResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, "model_catalog", got.Object)
	require.Equal(t, []string{"gpt-5.4"}, got.ChatModels)
	require.Equal(t, []string{"gpt-image-2"}, got.ImageModels)
	require.Equal(t, []string{"sora-2"}, got.VideoModels)

	byID := modelCatalogItemsByIDForTest(got.Items)
	require.Equal(t, []string{service.ModelCapabilityChat}, byID["gpt-5.4"].Capabilities)
	require.Equal(t, []string{service.ModelCapabilityImage}, byID["gpt-image-2"].Capabilities)
	require.True(t, byID["gpt-image-2"].Enabled)
	require.Equal(t, []string{service.ModelCapabilityVideo}, byID["sora-2"].Capabilities)
	require.False(t, byID["sora-2"].Enabled)
}

func TestGatewayModels_MultiGroupRoutesAggregateRoutableModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10031)
	imageGroupID := int64(10032)
	videoGroupID := int64(10033)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				imageGroupID: {
					{
						ID:       2,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
				videoGroupID: {
					{
						ID:       3,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"doubao-seedance-2.0-pro": "doubao-seedance-2.0-pro",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true, ModelPatterns: []string{"gpt-*"}},
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true, ModelPatterns: []string{"gpt-image-*"}},
			{GroupID: videoGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"doubao-seedance-*"}},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: imageGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeImage, ModelMatchPatterns: []string{"gpt-image-*"}, AllowImageGeneration: true, Hydrated: true},
			{ID: videoGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeVideo, ModelMatchPatterns: []string{"doubao-seedance-*"}, Hydrated: true},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.4", "gpt-image-2", "doubao-seedance-2.0-pro"}, modelIDsForTest(got.Data))
	require.Equal(t, "model", got.Data[0].Object)
	require.NotZero(t, got.Data[0].Created)
}

func TestGatewayModels_ForcedPlatformSkipsMultiGroupRouteAggregation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10036)
	imageGroupID := int64(10037)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				imageGroupID: {
					{
						ID:       2,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyForcePlatform), service.PlatformGemini)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true, ModelPatterns: []string{"gpt-*"}},
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true, ModelPatterns: []string{"gpt-image-*"}},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: imageGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeImage, AllowImageGeneration: true, Hydrated: true},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	ids := modelIDsForTest(got.Data)
	require.Contains(t, ids, "gemini-2.5-flash")
	require.NotContains(t, ids, "gpt-5.4")
	require.NotContains(t, ids, "gpt-image-2")
}

func TestGatewayModelCatalog_MultiGroupRoutesAggregateCapabilities(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10041)
	imageGroupID := int64(10042)
	videoGroupID := int64(10043)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				imageGroupID: {
					{
						ID:       2,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
				videoGroupID: {
					{
						ID:       3,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"sora-2": "sora-2",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/model-catalog", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
			{GroupID: videoGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"sora-*"}},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: imageGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeImage, ModelMatchPatterns: []string{"gpt-image-*"}, AllowImageGeneration: true, Hydrated: true},
			{ID: videoGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeVideo, ModelMatchPatterns: []string{"sora-*"}, Hydrated: true},
		},
	})

	h.ModelCatalog(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelCatalogResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.4"}, got.ChatModels)
	require.Equal(t, []string{"gpt-image-2"}, got.ImageModels)
	require.Equal(t, []string{"sora-2"}, got.VideoModels)

	byID := modelCatalogItemsByIDForTest(got.Items)
	require.Equal(t, []string{service.ModelCapabilityChat}, byID["gpt-5.4"].Capabilities)
	require.Equal(t, []string{service.ModelCapabilityImage}, byID["gpt-image-2"].Capabilities)
	require.Equal(t, []string{service.ModelCapabilityVideo}, byID["sora-2"].Capabilities)
	require.False(t, byID["sora-2"].Enabled)
}

func TestGatewayModels_MultiGroupRoutesExpandVideoWildcardFromDefaultCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10051)
	videoGroupID := int64(10052)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				videoGroupID: {
					{ID: 2, Platform: service.PlatformOpenAI},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true, ModelPatterns: []string{"gpt-*"}},
			{GroupID: videoGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"doubao-seedance-*"}},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: videoGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeVideo, ModelMatchPatterns: []string{"doubao-seedance-*"}, Hydrated: true},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	ids := modelIDsForTest(got.Data)
	require.Contains(t, ids, "gpt-5.4")
	require.Contains(t, ids, "doubao-seedance-2-0-fast-480p")
	require.NotContains(t, ids, "doubao-seedance-*")
}

func TestGatewayModels_MultiGroupRoutesDoNotExposeUnsupportedExactPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10053)
	videoGroupID := int64(10054)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				videoGroupID: {
					{
						ID:       2,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"sora-2": "sora-2",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: videoGroupID, Priority: 1, Weight: 1, Enabled: true, ModelPatterns: []string{"veo3.1-fast"}},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: videoGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeVideo, Hydrated: true},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	ids := modelIDsForTest(got.Data)
	require.Contains(t, ids, "gpt-5.4")
	require.NotContains(t, ids, "veo3.1-fast")
	require.NotContains(t, ids, "sora-2")
}

func TestGatewayModels_MultiGroupRoutesDoNotFallbackWhenGroupRulesRejectAllModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(10058)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{"gpt-5.4": "gpt-5.4"},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:                 groupID,
			Platform:           service.PlatformOpenAI,
			Status:             service.StatusActive,
			RoutingScope:       service.GroupRoutingScopeInference,
			ModelMatchPatterns: []string{"claude-*"},
			Hydrated:           true,
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID: groupID, Priority: 1, Weight: 1, Enabled: true,
		}},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Empty(t, got.Data)
}

func TestGatewayModels_MultiGroupRoutesDoNotExposeImageGroupWhenGenerationDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	chatGroupID := int64(10055)
	imageGroupID := int64(10056)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				chatGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
				imageGroupID: {
					{
						ID:       2,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-image-2": "gpt-image-2",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: chatGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeInference, ModelMatchPatterns: []string{"gpt-*"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: chatGroupID, Priority: 1, Weight: 1, Enabled: true, TextOnly: true},
			{GroupID: imageGroupID, Priority: 1, Weight: 1, Enabled: true, ImageOnly: true},
		},
		MultiGroupRouteGroups: []*service.Group{
			{ID: imageGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeImage, AllowImageGeneration: false, Hydrated: true},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	ids := modelIDsForTest(got.Data)
	require.Contains(t, ids, "gpt-5.4")
	require.NotContains(t, ids, "gpt-image-2")
}

func TestGatewayModelCatalog_MultiGroupRoutesClassifyCustomVideoScopeModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	videoGroupID := int64(10057)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				videoGroupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"custom-motion-model": "custom-motion-model",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/model-catalog", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: videoGroupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, RoutingScope: service.GroupRoutingScopeVideo, ModelMatchPatterns: []string{"custom-motion-model"}, Hydrated: true},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: videoGroupID, Priority: 1, Weight: 1, Enabled: true},
		},
	})

	h.ModelCatalog(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelCatalogResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"custom-motion-model"}, got.VideoModels)
	require.Empty(t, got.ChatModels)

	byID := modelCatalogItemsByIDForTest(got.Items)
	require.Equal(t, []string{service.ModelCapabilityVideo}, byID["custom-motion-model"].Capabilities)
	require.False(t, byID["custom-motion-model"].Enabled)
}

func TestGatewayModels_CustomModelsListDisabledKeepsOriginalModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(22)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.5": "gpt-5.5",
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: false,
				Models:  []string{"gpt-5.5"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.4", "gpt-5.5"}, modelIDsForTest(got.Data))
}

func TestGatewayModels_CustomModelsListFiltersAndOrdersMappedModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(23)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4":         "gpt-5.4",
								"gpt-5.5":         "gpt-5.5",
								"legacy-gpt-2024": "legacy-gpt-2024",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.5", "missing-model", "gpt-5.4"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.5", "gpt-5.4"}, modelIDsForTest(got.Data))
}

func TestGatewayModels_CustomModelsListKeepsConcreteModelAllowedByWildcardMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(26)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformAnthropic,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"claude-*": "claude-sonnet-4-6",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformAnthropic,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"claude-sonnet-4-6"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"claude-sonnet-4-6"}, modelIDsForTest(got.Data))
}

func TestGatewayModels_CustomModelsListCanReturnEmptyWhenSelectionsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(24)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{
						ID:       1,
						Platform: service.PlatformOpenAI,
						Credentials: map[string]any{
							"model_mapping": map[string]any{
								"gpt-5.4": "gpt-5.4",
							},
						},
					},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.5"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Empty(t, modelIDsForTest(got.Data))
}

func TestGatewayModels_CustomModelsListFiltersDefaultFallbackModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(25)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{ID: 1, Platform: service.PlatformOpenAI},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.5", "legacy-gpt-2024", "gpt-5.4"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.5", "gpt-5.4"}, modelIDsForTest(got.Data))
}

func TestGatewayModels_OpenAICustomModelsListKeepsOpenAIResponseShapeForDefaultFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	groupID := int64(27)
	h := newGatewayModelsHandlerForTest(
		&gatewayModelsAccountRepoStub{
			byGroup: map[int64][]service.Account{
				groupID: {
					{ID: 1, Platform: service.PlatformOpenAI},
				},
			},
		},
	)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.5", "gpt-5.4"},
			},
		},
	})

	h.Models(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var got gatewayModelsResponseForTest
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, []string{"gpt-5.5", "gpt-5.4"}, modelIDsForTest(got.Data))
	require.Equal(t, "model", got.Data[0].Object)
	require.NotZero(t, got.Data[0].Created)
	require.Equal(t, "openai", got.Data[0].OwnedBy)
	require.Empty(t, got.Data[0].CreatedAt)
}

func modelIDsForTest(models []gatewayModelItemForTest) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	return ids
}

func modelCatalogItemsByIDForTest(models []gatewayModelCatalogItemForTest) map[string]gatewayModelCatalogItemForTest {
	out := make(map[string]gatewayModelCatalogItemForTest, len(models))
	for _, model := range models {
		out[model.ID] = model
	}
	return out
}
