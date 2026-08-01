package service

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIGatewayServiceForwardVideos_APIKeyUsesConfiguredBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"doubao-seedance-2.0","prompt":"make a video","duration":5}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Request-Id": []string{"req_video"},
			},
			Body: io.NopCloser(strings.NewReader(`{"task_id":"task_123","status":"submitted"}`)),
		},
	}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{
		ID:       7,
		Name:     "apimart",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "test-api-key",
			"base_url": "https://api.apimart.ai",
		},
	}

	result, err := svc.ForwardVideos(context.Background(), c, account, body, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "doubao-seedance-2.0", result.Model)
	require.Equal(t, "doubao-seedance-2.0", result.UpstreamModel)
	require.Equal(t, 0, result.ImageCount)
	require.Empty(t, result.ImageSize)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, "https://api.apimart.ai/v1/videos/generations", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer test-api-key", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
	require.Equal(t, "doubao-seedance-2.0", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "task_123", gjson.Get(rec.Body.String(), "task_id").String())
}

func TestOpenAIGatewayServiceForwardVideos_AppliesAccountModelMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"seedance-video","prompt":"make a video"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"task_id":"task_mapped"}`)),
		},
	}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{
		ID:       8,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":       "test-api-key",
			"base_url":      "https://api.apimart.ai/v1",
			"model_mapping": map[string]any{"seedance-video": "doubao-seedance-2.0"},
		},
	}

	result, err := svc.ForwardVideos(context.Background(), c, account, body, "")

	require.NoError(t, err)
	require.Equal(t, "seedance-video", result.Model)
	require.Equal(t, "doubao-seedance-2.0", result.BillingModel)
	require.Equal(t, "doubao-seedance-2.0", result.UpstreamModel)
	require.Equal(t, "https://api.apimart.ai/v1/videos/generations", upstream.lastReq.URL.String())
	require.Equal(t, "doubao-seedance-2.0", gjson.GetBytes(upstream.lastBody, "model").String())
}

func TestOpenAIGatewayServiceForwardVideoTask_ProxiesStatusRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_123?language=zh", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"task_id":"task_123","status":"completed","output":["https://cdn.example/video.mp4"]}`)),
		},
	}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{
		ID:       9,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "test-api-key",
			"base_url": "https://api.apimart.ai",
		},
	}

	body, err := svc.ForwardVideoTask(context.Background(), c, account, "task_123", "zh")

	require.NoError(t, err)
	require.Contains(t, string(body), "task_123")
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Equal(t, "https://api.apimart.ai/v1/tasks/task_123?language=zh", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer test-api-key", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "completed", gjson.Get(rec.Body.String(), "status").String())
}

func TestOpenAIGatewayServiceForwardVideoTask_RejectsUnsafeTaskID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID:       9,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "test-api-key",
			"base_url": "https://api.apimart.ai",
		},
	}

	for _, taskID := range []string{"", " ", ".", "..", " .. ", "task\x00id", "task\rid", "task\nid", "task\r", "task\n"} {
		t.Run(taskID, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/v1/tasks/invalid", nil)
			upstream := &httpUpstreamRecorder{}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}

			body, err := svc.ForwardVideoTask(context.Background(), c, account, taskID, "zh")

			require.Error(t, err)
			require.Nil(t, body)
			require.Nil(t, upstream.lastReq)
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "task_id is invalid")
		})
	}
}

func TestOpenAIGatewayServiceForwardVideoTask_PreservesAPIMartBusinessErrorBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_missing?language=zh", nil)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"code":404,"message":"task not found","task_id":"task_missing"}`)),
		},
	}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	account := &Account{
		ID:       10,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "test-api-key",
			"base_url": "https://api.apimart.ai",
		},
	}

	body, err := svc.ForwardVideoTask(context.Background(), c, account, "task_missing", "zh")

	require.Error(t, err)
	require.Contains(t, string(body), "task_missing")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(404), gjson.Get(rec.Body.String(), "code").Int())
	require.Equal(t, "task not found", gjson.Get(rec.Body.String(), "message").String())
}

func TestOpenAIGatewayServiceForwardVideos_RemembersTaskAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"doubao-seedance-2.0","prompt":"make a video"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/videos/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	account := &Account{
		ID:          12,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: map[string]any{
			"api_key":  "test-api-key",
			"base_url": "https://api.apimart.ai",
		},
	}
	upstream := &httpUpstreamRecorder{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"data":{"task_id":"task_memory"}}`)),
		},
	}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
	}

	_, err := svc.ForwardVideos(context.Background(), c, account, body, "")
	require.NoError(t, err)

	raw, ok := svc.openaiVideoTaskAccounts.Load("task_memory")
	require.True(t, ok)
	ref, ok := raw.(openAIVideoTaskAccountRef)
	require.True(t, ok)
	require.Equal(t, int64(12), ref.AccountID)
}

func TestOpenAIGatewayServiceEstimateOpenAIVideoCost_UsesChannelTier(t *testing.T) {
	groupID := int64(99)
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: "doubao-seedance-2.0"}] = &ChannelModelPricing{
		BillingMode: BillingModePerRequest,
		Intervals: []PricingInterval{
			{TierLabel: "720:5s", PerRequestPrice: ptrFloat64(0.25)},
			{TierLabel: "1080:5s", PerRequestPrice: ptrFloat64(0.5)},
		},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: billingService,
		resolver:       NewModelPricingResolver(channelService, billingService),
	}
	user := &User{ID: 7}
	apiKey := &APIKey{ID: 8, UserID: user.ID, User: user, GroupID: &groupID, Group: &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 2}}
	account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"doubao-seedance-2.0","prompt":"make a video","resolution":"720p","duration":5}`)

	cost, billingModel, err := svc.EstimateOpenAIVideoCost(context.Background(), apiKey, user, account, body, ChannelUsageFields{
		OriginalModel:      "doubao-seedance-2.0",
		ChannelMappedModel: "doubao-seedance-2.0",
		BillingModelSource: BillingModelSourceChannelMapped,
	})

	require.NoError(t, err)
	require.Equal(t, "doubao-seedance-2.0", billingModel)
	require.NotNil(t, cost)
	require.InDelta(t, 0.25, cost.TotalCost, 0.000001)
	require.InDelta(t, 0.5, cost.ActualCost, 0.000001)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
}

func TestOpenAIGatewayServiceEstimateOpenAIVideoCost_UsesBaseTierAsPerSecondPrice(t *testing.T) {
	groupID := int64(101)
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: "kling-v3-omni"}] = &ChannelModelPricing{
		BillingMode: BillingModePerRequest,
		Intervals: []PricingInterval{
			{TierLabel: "720", PerRequestPrice: ptrFloat64(0.08)},
			{TierLabel: "1080", PerRequestPrice: ptrFloat64(0.12)},
		},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: billingService,
		resolver:       NewModelPricingResolver(channelService, billingService),
	}
	user := &User{ID: 7}
	apiKey := &APIKey{ID: 8, UserID: user.ID, User: user, GroupID: &groupID, Group: &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 2}}
	account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"kling-v3-omni","prompt":"make a video","resolution":"720p","duration":7}`)

	cost, billingModel, err := svc.EstimateOpenAIVideoCost(context.Background(), apiKey, user, account, body, ChannelUsageFields{
		OriginalModel:      "kling-v3-omni",
		ChannelMappedModel: "kling-v3-omni",
		BillingModelSource: BillingModelSourceChannelMapped,
	})

	require.NoError(t, err)
	require.Equal(t, "kling-v3-omni", billingModel)
	require.NotNil(t, cost)
	require.InDelta(t, 0.56, cost.TotalCost, 0.000001)
	require.InDelta(t, 1.12, cost.ActualCost, 0.000001)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
}

func TestOpenAIGatewayServiceEstimateOpenAIVideoCost_BaseTierPerSecondWinsOverDefaultPrice(t *testing.T) {
	groupID := int64(103)
	defaultPrice := 0.2
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: "kling-v3-omni"}] = &ChannelModelPricing{
		BillingMode:     BillingModePerRequest,
		PerRequestPrice: &defaultPrice,
		Intervals: []PricingInterval{
			{TierLabel: "720", PerRequestPrice: ptrFloat64(0.08)},
		},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: billingService,
		resolver:       NewModelPricingResolver(channelService, billingService),
	}
	user := &User{ID: 7}
	apiKey := &APIKey{ID: 8, UserID: user.ID, User: user, GroupID: &groupID, Group: &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 1}}
	account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"kling-v3-omni","prompt":"make a video","resolution":"720p","duration":7}`)

	cost, _, err := svc.EstimateOpenAIVideoCost(context.Background(), apiKey, user, account, body, ChannelUsageFields{
		OriginalModel:      "kling-v3-omni",
		ChannelMappedModel: "kling-v3-omni",
		BillingModelSource: BillingModelSourceChannelMapped,
	})

	require.NoError(t, err)
	require.NotNil(t, cost)
	require.InDelta(t, 0.56, cost.TotalCost, 0.000001)
	require.InDelta(t, 0.56, cost.ActualCost, 0.000001)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
}

func TestOpenAIGatewayServiceEstimateOpenAIVideoCost_PrefersExactDurationTierOverPerSecondBaseTier(t *testing.T) {
	groupID := int64(102)
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: "kling-v3-omni"}] = &ChannelModelPricing{
		BillingMode: BillingModePerRequest,
		Intervals: []PricingInterval{
			{TierLabel: "720", PerRequestPrice: ptrFloat64(0.08)},
			{TierLabel: "720:7s", PerRequestPrice: ptrFloat64(0.45)},
		},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	svc := &OpenAIGatewayService{
		cfg:            &config.Config{},
		billingService: billingService,
		resolver:       NewModelPricingResolver(channelService, billingService),
	}
	user := &User{ID: 7}
	apiKey := &APIKey{ID: 8, UserID: user.ID, User: user, GroupID: &groupID, Group: &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 2}}
	account := &Account{ID: 9, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"model":"kling-v3-omni","prompt":"make a video","resolution":"720p","duration":7}`)

	cost, _, err := svc.EstimateOpenAIVideoCost(context.Background(), apiKey, user, account, body, ChannelUsageFields{
		OriginalModel:      "kling-v3-omni",
		ChannelMappedModel: "kling-v3-omni",
		BillingModelSource: BillingModelSourceChannelMapped,
	})

	require.NoError(t, err)
	require.NotNil(t, cost)
	require.InDelta(t, 0.45, cost.TotalCost, 0.000001)
	require.InDelta(t, 0.9, cost.ActualCost, 0.000001)
	require.Equal(t, string(BillingModePerRequest), cost.BillingMode)
}

func TestOpenAIGatewayServiceVideoTaskSettlement_ChargesOnceOnSuccess(t *testing.T) {
	taskRepo := newOpenAIVideoTaskMemoryRepo()
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, userRepo, subRepo, nil)
	svc.openaiVideoTaskRepo = taskRepo
	account := &Account{ID: 55, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	user := &User{ID: 77}
	apiKey := &APIKey{ID: 88, UserID: user.ID, User: user, GroupID: videoPtrInt64(99), Group: &Group{ID: 99, Platform: PlatformOpenAI, RateMultiplier: 1}}

	result := &OpenAIForwardResult{
		VideoTaskID:   "task_success",
		Model:         "doubao-seedance-2.0",
		BillingModel:  "doubao-seedance-2.0",
		UpstreamModel: "doubao-seedance-2.0",
		ResponseBody:  []byte(`{"task_id":"task_success","status":"submitted","resolution":"720p","duration":5}`),
	}
	require.NoError(t, svc.RecordOpenAIVideoTaskSubmitted(context.Background(), &OpenAIVideoTaskRecordInput{
		Result:             result,
		APIKey:             apiKey,
		User:               user,
		Account:            account,
		EstimatedCost:      &CostBreakdown{TotalCost: 0.2, ActualCost: 0.3, BillingMode: string(BillingModePerRequest)},
		ReservedCost:       0.3,
		RequestPayloadHash: "payload-hash",
		ChannelUsageFields: ChannelUsageFields{
			OriginalModel:      "doubao-seedance-2.0",
			ChannelMappedModel: "doubao-seedance-2.0",
			BillingModelSource: BillingModelSourceChannelMapped,
		},
	}))

	statusBody := []byte(`{"task_id":"task_success","status":"completed","output":["https://cdn.example/video.mp4"]}`)
	require.NoError(t, svc.SettleOpenAIVideoTaskIfTerminal(context.Background(), &OpenAIVideoTaskSettleInput{
		TaskID:           "task_success",
		StatusBody:       statusBody,
		APIKey:           apiKey,
		User:             user,
		Account:          account,
		InboundEndpoint:  "/v1/tasks/:task_id",
		UpstreamEndpoint: "/v1/tasks/:task_id",
	}))
	require.Equal(t, 1, billingRepo.calls)
	require.Equal(t, 1, usageRepo.calls)
	require.NotNil(t, usageRepo.lastLog)
	require.Equal(t, "video_task:task_success", usageRepo.lastLog.RequestID)
	require.NotNil(t, usageRepo.lastLog.MediaType)
	require.Equal(t, "video", *usageRepo.lastLog.MediaType)
	require.NotNil(t, billingRepo.lastCmd)
	require.Equal(t, "video", billingRepo.lastCmd.MediaType)
	require.InDelta(t, 0, billingRepo.lastCmd.BalanceCost, 0.000001)
	require.InDelta(t, 0.3, billingRepo.lastCmd.PrepaidBalanceCost, 0.000001)
	require.InDelta(t, 0.3, usageRepo.lastLog.ActualCost, 0.000001)

	require.NoError(t, svc.SettleOpenAIVideoTaskIfTerminal(context.Background(), &OpenAIVideoTaskSettleInput{
		TaskID:     "task_success",
		StatusBody: statusBody,
		APIKey:     apiKey,
		User:       user,
		Account:    account,
	}))
	require.Equal(t, 1, billingRepo.calls)
	require.Equal(t, 1, usageRepo.calls)
}

func TestOpenAIGatewayServiceVideoTaskSettlement_FailedTaskDoesNotCharge(t *testing.T) {
	taskRepo := newOpenAIVideoTaskMemoryRepo()
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	svc.openaiVideoTaskRepo = taskRepo
	account := &Account{ID: 56, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	user := &User{ID: 78}
	apiKey := &APIKey{ID: 89, UserID: user.ID, User: user, GroupID: videoPtrInt64(100), Group: &Group{ID: 100, Platform: PlatformOpenAI, RateMultiplier: 1}}

	require.NoError(t, svc.RecordOpenAIVideoTaskSubmitted(context.Background(), &OpenAIVideoTaskRecordInput{
		Result:        &OpenAIForwardResult{VideoTaskID: "task_failed", Model: "kling-v3-omni", BillingModel: "kling-v3-omni", UpstreamModel: "kling-v3-omni"},
		APIKey:        apiKey,
		User:          user,
		Account:       account,
		EstimatedCost: &CostBreakdown{TotalCost: 0.5, ActualCost: 0.5, BillingMode: string(BillingModePerRequest)},
		ReservedCost:  0.5,
	}))
	taskRepo.tasks["task_failed"].BillingStatus = OpenAIVideoTaskBillingReserved
	require.NoError(t, svc.SettleOpenAIVideoTaskIfTerminal(context.Background(), &OpenAIVideoTaskSettleInput{
		TaskID:     "task_failed",
		StatusBody: []byte(`{"task_id":"task_failed","status":"failed"}`),
		APIKey:     apiKey,
		User:       user,
		Account:    account,
	}))
	require.Equal(t, 0, billingRepo.calls)
	require.Equal(t, 0, usageRepo.calls)
	task, ok := taskRepo.tasks["task_failed"]
	require.True(t, ok)
	require.Equal(t, OpenAIVideoTaskBillingFailedNoCost, task.BillingStatus)
	require.InDelta(t, 0.5, task.RefundedCost, 0.000001)
}

func TestOpenAIGatewayServiceVideoTaskSettlement_FailedTaskInvalidatesBalanceCache(t *testing.T) {
	taskRepo := newOpenAIVideoTaskMemoryRepo()
	cache := &openAIVideoBillingCacheStub{}
	svc := &OpenAIGatewayService{
		openaiVideoTaskRepo: taskRepo,
		billingCacheService: &BillingCacheService{cache: cache},
	}
	taskRepo.tasks["task_refund_cache"] = &OpenAIVideoTask{
		TaskID:        "task_refund_cache",
		UserID:        78,
		BillingStatus: OpenAIVideoTaskBillingReserved,
		ReservedCost:  0.5,
	}

	require.NoError(t, svc.SettleOpenAIVideoTaskIfTerminal(context.Background(), &OpenAIVideoTaskSettleInput{
		TaskID:     "task_refund_cache",
		StatusBody: []byte(`{"task_id":"task_refund_cache","status":"failed"}`),
	}))

	require.Equal(t, []int64{78}, cache.invalidatedUserIDs)
	task := taskRepo.tasks["task_refund_cache"]
	require.Equal(t, OpenAIVideoTaskBillingFailedNoCost, task.BillingStatus)
	require.InDelta(t, 0.5, task.RefundedCost, 0.000001)
}

type openAIVideoTaskMemoryRepo struct {
	tasks map[string]*OpenAIVideoTask
}

func newOpenAIVideoTaskMemoryRepo() *openAIVideoTaskMemoryRepo {
	return &openAIVideoTaskMemoryRepo{tasks: make(map[string]*OpenAIVideoTask)}
}

func (r *openAIVideoTaskMemoryRepo) UpsertSubmitted(_ context.Context, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error) {
	task := &OpenAIVideoTask{
		ID:                 int64(len(r.tasks) + 1),
		TaskID:             input.TaskID,
		Provider:           firstNonEmptyString(input.Provider, OpenAIVideoTaskProviderOpenAI),
		UserID:             input.UserID,
		APIKeyID:           input.APIKeyID,
		GroupID:            input.GroupID,
		AccountID:          input.AccountID,
		Model:              input.Model,
		BillingModel:       input.BillingModel,
		UpstreamModel:      input.UpstreamModel,
		ChannelID:          input.ChannelID,
		OriginalModel:      input.OriginalModel,
		ChannelMappedModel: input.ChannelMappedModel,
		BillingModelSource: input.BillingModelSource,
		ModelMappingChain:  input.ModelMappingChain,
		Status:             input.Status,
		BillingStatus:      OpenAIVideoTaskBillingPending,
		EstimatedCost:      input.EstimatedCost,
		ReservedCost:       input.ReservedCost,
		RequestPayloadHash: input.RequestPayloadHash,
		SubmitResponse:     input.SubmitResponse,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	r.tasks[input.TaskID] = task
	return task, nil
}

func (r *openAIVideoTaskMemoryRepo) ReserveBalance(_ context.Context, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error) {
	task, err := r.UpsertSubmitted(context.Background(), input)
	if err != nil {
		return nil, err
	}
	task.BillingStatus = OpenAIVideoTaskBillingReserved
	task.EstimatedCost = input.EstimatedCost
	task.ReservedCost = input.ReservedCost
	return task, nil
}

func (r *openAIVideoTaskMemoryRepo) BindSubmitted(_ context.Context, placeholderTaskID string, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error) {
	task, ok := r.tasks[placeholderTaskID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	delete(r.tasks, placeholderTaskID)
	task.TaskID = input.TaskID
	task.UserID = input.UserID
	task.APIKeyID = input.APIKeyID
	task.GroupID = input.GroupID
	task.AccountID = input.AccountID
	task.Model = input.Model
	task.BillingModel = input.BillingModel
	task.UpstreamModel = input.UpstreamModel
	task.ChannelID = input.ChannelID
	task.OriginalModel = input.OriginalModel
	task.ChannelMappedModel = input.ChannelMappedModel
	task.BillingModelSource = input.BillingModelSource
	task.ModelMappingChain = input.ModelMappingChain
	task.Status = input.Status
	if input.EstimatedCost > 0 {
		task.EstimatedCost = input.EstimatedCost
	}
	if input.ReservedCost > 0 {
		task.ReservedCost = input.ReservedCost
	}
	task.RequestPayloadHash = input.RequestPayloadHash
	task.SubmitResponse = input.SubmitResponse
	task.UpdatedAt = time.Now()
	r.tasks[input.TaskID] = task
	return task, nil
}

func (r *openAIVideoTaskMemoryRepo) RefundReserved(_ context.Context, _ string, taskID string) (*OpenAIVideoTask, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if task.BillingStatus == OpenAIVideoTaskBillingReserved && task.ReservedCost > 0 {
		task.RefundedCost = task.ReservedCost
		task.BillingStatus = OpenAIVideoTaskBillingRefunded
		task.UpdatedAt = time.Now()
	}
	return task, nil
}

func (r *openAIVideoTaskMemoryRepo) GetByTaskID(_ context.Context, _ string, taskID string) (*OpenAIVideoTask, error) {
	if task, ok := r.tasks[taskID]; ok {
		return task, nil
	}
	return nil, sql.ErrNoRows
}

func (r *openAIVideoTaskMemoryRepo) UpdateStatus(_ context.Context, input *OpenAIVideoTaskStatusUpdate) (*OpenAIVideoTask, error) {
	task, ok := r.tasks[input.TaskID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if input.Status != "" {
		task.Status = input.Status
	}
	if input.BillingStatus != "" {
		task.BillingStatus = input.BillingStatus
	}
	if input.UsageLogID != nil {
		task.UsageLogID = input.UsageLogID
	}
	if len(input.LastStatusResponse) > 0 {
		task.LastStatusResponse = input.LastStatusResponse
	}
	if input.CompletedAt != nil {
		task.CompletedAt = input.CompletedAt
	}
	if input.BilledAt != nil {
		task.BilledAt = input.BilledAt
	}
	task.UpdatedAt = time.Now()
	return task, nil
}

func videoPtrInt64(v int64) *int64 {
	return &v
}

type openAIVideoBillingCacheStub struct {
	invalidatedUserIDs []int64
}

func (s *openAIVideoBillingCacheStub) GetUserBalance(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}

func (s *openAIVideoBillingCacheStub) SetUserBalance(_ context.Context, _ int64, _ float64) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) DeductUserBalance(_ context.Context, _ int64, _ float64) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) InvalidateUserBalance(_ context.Context, userID int64) error {
	s.invalidatedUserIDs = append(s.invalidatedUserIDs, userID)
	return nil
}

func (s *openAIVideoBillingCacheStub) GetSubscriptionCache(_ context.Context, _, _ int64) (*SubscriptionCacheData, error) {
	return nil, nil
}

func (s *openAIVideoBillingCacheStub) SetSubscriptionCache(_ context.Context, _, _ int64, _ *SubscriptionCacheData) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) UpdateSubscriptionUsage(_ context.Context, _, _ int64, _ float64) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) InvalidateSubscriptionCache(_ context.Context, _, _ int64) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) GetAPIKeyRateLimit(_ context.Context, _ int64) (*APIKeyRateLimitCacheData, error) {
	return nil, nil
}

func (s *openAIVideoBillingCacheStub) SetAPIKeyRateLimit(_ context.Context, _ int64, _ *APIKeyRateLimitCacheData) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) UpdateAPIKeyRateLimitUsage(_ context.Context, _ int64, _ float64) error {
	return nil
}

func (s *openAIVideoBillingCacheStub) InvalidateAPIKeyRateLimit(_ context.Context, _ int64) error {
	return nil
}
