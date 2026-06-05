package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	openAIVideosGenerationsEndpoint = "/v1/videos/generations"
	openAITasksEndpointPrefix       = "/v1/tasks/"
	openAIVideosMaxResponseBytes    = 8 << 20
	openAIVideoTaskAccountTTL       = 24 * time.Hour
)

type openAIVideoTaskAccountRef struct {
	AccountID int64
	ExpiresAt time.Time
}

type OpenAIVideoTaskRecordInput struct {
	Result             *OpenAIForwardResult
	APIKey             *APIKey
	User               *User
	Account            *Account
	EstimatedCost      *CostBreakdown
	ReservedCost       float64
	RequestPayloadHash string
	ChannelUsageFields
}

type OpenAIVideoForwardOptions struct {
	WriteResponse bool
}

type OpenAIVideoTaskReserveInput struct {
	PlaceholderTaskID  string
	APIKey             *APIKey
	User               *User
	Account            *Account
	RequestBody        []byte
	RequestPayloadHash string
	ChannelUsageFields
}

type OpenAIVideoTaskSettleInput struct {
	TaskID           string
	StatusBody       []byte
	APIKey           *APIKey
	User             *User
	Account          *Account
	Subscription     *UserSubscription
	InboundEndpoint  string
	UpstreamEndpoint string
	UserAgent        string
	IPAddress        string
	APIKeyService    APIKeyQuotaUpdater
}

func (s *OpenAIGatewayService) ForwardVideos(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	defaultMappedModel string,
) (*OpenAIForwardResult, error) {
	return s.ForwardVideosWithOptions(ctx, c, account, body, defaultMappedModel, OpenAIVideoForwardOptions{WriteResponse: true})
}

func (s *OpenAIGatewayService) ForwardVideosWithOptions(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	body []byte,
	defaultMappedModel string,
	opts OpenAIVideoForwardOptions,
) (*OpenAIForwardResult, error) {
	startTime := time.Now()
	originalModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if originalModel == "" {
		writeOpenAIVideosError(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return nil, fmt.Errorf("missing model in request")
	}
	if account == nil || account.Type != AccountTypeAPIKey {
		return nil, fmt.Errorf("videos API requires an OpenAI API key account")
	}

	billingModel := resolveOpenAIForwardModel(account, originalModel, defaultMappedModel)
	upstreamModel := normalizeOpenAIModelForUpstream(account, billingModel)
	upstreamBody := body
	if upstreamModel != originalModel {
		upstreamBody = ReplaceModelInBody(body, upstreamModel)
	}
	setOpsUpstreamRequestBody(c, upstreamBody)

	token := account.GetOpenAIApiKey()
	if token == "" {
		return nil, fmt.Errorf("account %d missing api_key", account.ID)
	}
	targetURL, err := s.openAICompatibleEndpointURL(account, openAIVideosGenerationsEndpoint)
	if err != nil {
		return nil, err
	}

	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	req, err := http.NewRequestWithContext(upstreamCtx, http.MethodPost, targetURL, bytes.NewReader(upstreamBody))
	releaseUpstreamCtx()
	if err != nil {
		return nil, fmt.Errorf("build upstream request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, values := range c.Request.Header {
		if openaiCCRawAllowedHeaders[strings.ToLower(key)] {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		req.Header.Set("user-agent", customUA)
	}

	respBody, resp, err := s.doOpenAICompatibleJSONRequest(c, req, account)
	if err != nil {
		return nil, err
	}
	taskID := extractOpenAIVideoTaskID(respBody)
	if taskID != "" {
		s.rememberOpenAIVideoTaskAccountID(taskID, account.ID)
	}
	if opts.WriteResponse {
		writeOpenAICompatibleJSONResponse(c, resp, respBody, s.responseHeaderFilter)
	}

	return &OpenAIForwardResult{
		RequestID:       firstNonEmptyString(resp.Header.Get("x-request-id"), resp.Header.Get("request-id")),
		Model:           originalModel,
		BillingModel:    billingModel,
		UpstreamModel:   upstreamModel,
		Stream:          false,
		Duration:        time.Since(startTime),
		ResponseHeaders: resp.Header.Clone(),
		VideoTaskID:     taskID,
		ResponseBody:    respBody,
	}, nil
}

func (s *OpenAIGatewayService) EstimateOpenAIVideoCost(ctx context.Context, apiKey *APIKey, user *User, account *Account, body []byte, fields ChannelUsageFields) (*CostBreakdown, string, error) {
	if s == nil || s.billingService == nil || apiKey == nil || user == nil {
		return nil, "", errors.New("openai video billing context is incomplete")
	}
	if apiKey.Group == nil || apiKey.GroupID == nil {
		return nil, "", errors.New("openai video group is required for billing")
	}
	originalModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if originalModel == "" {
		return nil, "", errors.New("openai video model is required")
	}
	billingModel := originalModel
	if fields.BillingModelSource == BillingModelSourceChannelMapped && fields.ChannelMappedModel != "" {
		billingModel = fields.ChannelMappedModel
	}
	if fields.BillingModelSource == BillingModelSourceRequested && fields.OriginalModel != "" {
		billingModel = fields.OriginalModel
	}
	if fields.BillingModelSource == BillingModelSourceUpstream && account != nil {
		billingModel = normalizeOpenAIModelForUpstream(account, billingModel)
	}
	multiplier := 1.0
	if s.cfg != nil {
		multiplier = s.cfg.Default.RateMultiplier
	}
	resolver := s.userGroupRateResolver
	if resolver == nil {
		resolver = newUserGroupRateResolver(nil, nil, resolveUserGroupRateCacheTTL(s.cfg), nil, "service.openai_gateway")
	}
	multiplier = resolver.Resolve(ctx, user.ID, *apiKey.GroupID, apiKey.Group.RateMultiplier)
	if s.membershipService != nil {
		multiplier = s.membershipService.ApplyRateMultiplier(ctx, user.ID, multiplier)
	}
	if s.resolver == nil {
		return nil, billingModel, errors.New("openai video channel pricing resolver is unavailable")
	}
	resolved := s.resolveOpenAIChannelPricing(ctx, billingModel, apiKey)
	if resolved == nil || (resolved.Mode != BillingModePerRequest && resolved.Mode != BillingModeImage) {
		return nil, billingModel, fmt.Errorf("%w for model: %s", ErrOpenAIVideoPricingUnavailable, billingModel)
	}
	gid := apiKey.Group.ID
	sizeTier, requestCount := extractOpenAIVideoBillingTierAndCount(body, nil)
	cost, err := s.billingService.CalculateCostUnified(CostInput{
		Ctx:            ctx,
		Model:          billingModel,
		GroupID:        &gid,
		RequestCount:   1,
		SizeTier:       sizeTier,
		RateMultiplier: multiplier,
		Resolver:       s.resolver,
		Resolved:       resolved,
	})
	if err != nil {
		return nil, billingModel, fmt.Errorf("%w: %v", ErrOpenAIVideoPricingUnavailable, err)
	}
	if shouldApplyOpenAIVideoPerSecondPricing(resolved, sizeTier, requestCount) {
		cost, err = s.billingService.CalculateCostUnified(CostInput{
			Ctx:            ctx,
			Model:          billingModel,
			GroupID:        &gid,
			RequestCount:   requestCount,
			SizeTier:       openAIVideoBaseBillingTier(sizeTier),
			RateMultiplier: multiplier,
			Resolver:       s.resolver,
			Resolved:       resolved,
		})
		if err != nil {
			return nil, billingModel, fmt.Errorf("%w: %v", ErrOpenAIVideoPricingUnavailable, err)
		}
	}
	if cost == nil || cost.ActualCost <= 0 {
		return nil, billingModel, fmt.Errorf("%w for model: %s", ErrOpenAIVideoPricingUnavailable, billingModel)
	}
	return cost, billingModel, nil
}

func (s *OpenAIGatewayService) ReserveOpenAIVideoTaskBalance(ctx context.Context, input *OpenAIVideoTaskReserveInput) (*OpenAIVideoTask, *CostBreakdown, error) {
	if s == nil || s.openaiVideoTaskRepo == nil || input == nil || input.APIKey == nil || input.User == nil || input.Account == nil {
		return nil, nil, nil
	}
	placeholderTaskID := strings.TrimSpace(input.PlaceholderTaskID)
	if placeholderTaskID == "" {
		return nil, nil, errors.New("openai video placeholder task_id is required")
	}
	cost, billingModel, err := s.EstimateOpenAIVideoCost(ctx, input.APIKey, input.User, input.Account, input.RequestBody, input.ChannelUsageFields)
	if err != nil {
		return nil, nil, err
	}
	task, err := s.openaiVideoTaskRepo.ReserveBalance(ctx, &OpenAIVideoTaskUpsertInput{
		TaskID:             placeholderTaskID,
		Provider:           OpenAIVideoTaskProviderOpenAI,
		UserID:             input.User.ID,
		APIKeyID:           input.APIKey.ID,
		GroupID:            input.APIKey.GroupID,
		AccountID:          input.Account.ID,
		Model:              firstNonEmptyString(input.ChannelUsageFields.OriginalModel, strings.TrimSpace(gjson.GetBytes(input.RequestBody, "model").String())),
		BillingModel:       billingModel,
		UpstreamModel:      normalizeOpenAIModelForUpstream(input.Account, billingModel),
		ChannelID:          input.ChannelID,
		OriginalModel:      input.OriginalModel,
		ChannelMappedModel: input.ChannelMappedModel,
		BillingModelSource: input.BillingModelSource,
		ModelMappingChain:  input.ModelMappingChain,
		Status:             "reserved",
		EstimatedCost:      cost.ActualCost,
		ReservedCost:       cost.ActualCost,
		RequestPayloadHash: input.RequestPayloadHash,
	})
	if err != nil {
		return nil, nil, err
	}
	s.invalidateOpenAIVideoBalanceCache(ctx, input.User.ID)
	return task, cost, nil
}

func (s *OpenAIGatewayService) RememberOpenAIVideoTaskAccountForTest(taskID string, accountID int64, ttl time.Duration) {
	if s == nil || strings.TrimSpace(taskID) == "" || accountID <= 0 {
		return
	}
	if ttl <= 0 {
		ttl = openAIVideoTaskAccountTTL
	}
	s.openaiVideoTaskAccounts.Store(strings.TrimSpace(taskID), openAIVideoTaskAccountRef{
		AccountID: accountID,
		ExpiresAt: time.Now().Add(ttl),
	})
}

func (s *OpenAIGatewayService) RecordOpenAIVideoTaskSubmitted(ctx context.Context, input *OpenAIVideoTaskRecordInput) error {
	if s == nil || s.openaiVideoTaskRepo == nil || input == nil || input.Result == nil || input.APIKey == nil || input.User == nil || input.Account == nil {
		return nil
	}
	taskID := strings.TrimSpace(input.Result.VideoTaskID)
	if taskID == "" {
		return nil
	}
	estimatedCost := 0.0
	if input.EstimatedCost != nil {
		estimatedCost = input.EstimatedCost.ActualCost
	}
	reservedCost := normalizeNonNegativeFloat(input.ReservedCost)
	task, err := s.openaiVideoTaskRepo.UpsertSubmitted(ctx, &OpenAIVideoTaskUpsertInput{
		TaskID:             taskID,
		Provider:           OpenAIVideoTaskProviderOpenAI,
		UserID:             input.User.ID,
		APIKeyID:           input.APIKey.ID,
		GroupID:            input.APIKey.GroupID,
		AccountID:          input.Account.ID,
		Model:              strings.TrimSpace(input.Result.Model),
		BillingModel:       strings.TrimSpace(input.Result.BillingModel),
		UpstreamModel:      strings.TrimSpace(input.Result.UpstreamModel),
		ChannelID:          input.ChannelID,
		OriginalModel:      input.OriginalModel,
		ChannelMappedModel: input.ChannelMappedModel,
		BillingModelSource: input.BillingModelSource,
		ModelMappingChain:  input.ModelMappingChain,
		Status:             firstNonEmptyString(extractOpenAIVideoTaskStatus(input.Result.ResponseBody), "submitted"),
		EstimatedCost:      estimatedCost,
		ReservedCost:       reservedCost,
		RequestPayloadHash: input.RequestPayloadHash,
		SubmitResponse:     input.Result.ResponseBody,
	})
	if err != nil {
		return err
	}
	if task != nil {
		s.rememberOpenAIVideoTaskAccountID(task.TaskID, task.AccountID)
	}
	return nil
}

func (s *OpenAIGatewayService) BindReservedOpenAIVideoTask(ctx context.Context, placeholderTaskID string, input *OpenAIVideoTaskRecordInput) (*OpenAIVideoTask, error) {
	if s == nil || s.openaiVideoTaskRepo == nil || input == nil || input.Result == nil || input.APIKey == nil || input.User == nil || input.Account == nil {
		return nil, nil
	}
	taskID := strings.TrimSpace(input.Result.VideoTaskID)
	if strings.TrimSpace(placeholderTaskID) == "" || taskID == "" {
		return nil, errors.New("openai video task_id is required")
	}
	estimatedCost := 0.0
	if input.EstimatedCost != nil {
		estimatedCost = input.EstimatedCost.ActualCost
	}
	reservedCost := normalizeNonNegativeFloat(input.ReservedCost)
	if reservedCost <= 0 && input.EstimatedCost != nil {
		reservedCost = input.EstimatedCost.ActualCost
	}
	task, err := s.openaiVideoTaskRepo.BindSubmitted(ctx, placeholderTaskID, &OpenAIVideoTaskUpsertInput{
		TaskID:             taskID,
		Provider:           OpenAIVideoTaskProviderOpenAI,
		UserID:             input.User.ID,
		APIKeyID:           input.APIKey.ID,
		GroupID:            input.APIKey.GroupID,
		AccountID:          input.Account.ID,
		Model:              strings.TrimSpace(input.Result.Model),
		BillingModel:       strings.TrimSpace(input.Result.BillingModel),
		UpstreamModel:      strings.TrimSpace(input.Result.UpstreamModel),
		ChannelID:          input.ChannelID,
		OriginalModel:      input.OriginalModel,
		ChannelMappedModel: input.ChannelMappedModel,
		BillingModelSource: input.BillingModelSource,
		ModelMappingChain:  input.ModelMappingChain,
		Status:             firstNonEmptyString(extractOpenAIVideoTaskStatus(input.Result.ResponseBody), "submitted"),
		EstimatedCost:      estimatedCost,
		ReservedCost:       reservedCost,
		RequestPayloadHash: input.RequestPayloadHash,
		SubmitResponse:     input.Result.ResponseBody,
	})
	if err != nil {
		return nil, err
	}
	if task != nil {
		s.rememberOpenAIVideoTaskAccountID(task.TaskID, task.AccountID)
	}
	return task, nil
}

func (s *OpenAIGatewayService) RefundOpenAIVideoTaskReservation(ctx context.Context, taskID string) error {
	if s == nil || s.openaiVideoTaskRepo == nil || strings.TrimSpace(taskID) == "" {
		return nil
	}
	task, err := s.openaiVideoTaskRepo.RefundReserved(ctx, OpenAIVideoTaskProviderOpenAI, taskID)
	if err == nil && task != nil {
		s.invalidateOpenAIVideoBalanceCache(ctx, task.UserID)
	}
	return err
}

func (s *OpenAIGatewayService) SettleOpenAIVideoTaskIfTerminal(ctx context.Context, input *OpenAIVideoTaskSettleInput) error {
	if s == nil || s.openaiVideoTaskRepo == nil || input == nil {
		return nil
	}
	taskID := strings.TrimSpace(input.TaskID)
	if taskID == "" {
		return nil
	}
	status := extractOpenAIVideoTaskStatus(input.StatusBody)
	if status == "" {
		return nil
	}
	now := time.Now()
	update := &OpenAIVideoTaskStatusUpdate{
		TaskID:             taskID,
		Provider:           OpenAIVideoTaskProviderOpenAI,
		Status:             status,
		LastStatusResponse: input.StatusBody,
	}
	if isOpenAIVideoTaskFailureStatus(status) {
		refundedTask, err := s.openaiVideoTaskRepo.RefundReserved(ctx, OpenAIVideoTaskProviderOpenAI, taskID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if refundedTask != nil && refundedTask.RefundedCost > 0 {
			s.invalidateOpenAIVideoBalanceCache(ctx, refundedTask.UserID)
		}
		update.BillingStatus = OpenAIVideoTaskBillingFailedNoCost
		update.CompletedAt = &now
		_, err = s.openaiVideoTaskRepo.UpdateStatus(ctx, update)
		return err
	}
	if !isOpenAIVideoTaskSuccessStatus(status) {
		_, err := s.openaiVideoTaskRepo.UpdateStatus(ctx, update)
		return err
	}

	task, err := s.openaiVideoTaskRepo.GetByTaskID(ctx, OpenAIVideoTaskProviderOpenAI, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			update.CompletedAt = &now
			_, updateErr := s.openaiVideoTaskRepo.UpdateStatus(ctx, update)
			return updateErr
		}
		return err
	}
	if task == nil {
		return nil
	}
	if task.BillingStatus == OpenAIVideoTaskBillingCharged {
		update.BillingStatus = OpenAIVideoTaskBillingCharged
		update.CompletedAt = &now
		_, err := s.openaiVideoTaskRepo.UpdateStatus(ctx, update)
		return err
	}

	account := input.Account
	if account == nil || account.ID != task.AccountID {
		resolved, resolveErr := s.resolveOpenAIVideoTaskAccountByID(ctx, task.AccountID)
		if resolveErr == nil {
			account = resolved
		}
	}
	apiKey := input.APIKey
	if apiKey == nil || apiKey.ID != task.APIKeyID {
		// Billing requires the authenticated API key object with group/user data.
		apiKey = input.APIKey
	}
	user := input.User
	if user == nil {
		user = apiKey.User
	}
	if account == nil || apiKey == nil || user == nil {
		return nil
	}

	result := &OpenAIForwardResult{
		RequestID:     "video_task:" + taskID,
		Model:         firstNonEmptyString(task.Model, task.OriginalModel),
		BillingModel:  task.BillingModel,
		UpstreamModel: task.UpstreamModel,
		Stream:        false,
		Duration:      time.Since(task.CreatedAt),
	}
	if result.BillingModel == "" {
		result.BillingModel = firstNonEmptyString(task.ChannelMappedModel, task.UpstreamModel, task.Model)
	}
	prepaidCost := normalizeNonNegativeFloat(task.ReservedCost)
	var costOverride *CostBreakdown
	if prepaidCost > 0 {
		result.BillingModel = firstNonEmptyString(task.BillingModel, result.BillingModel)
		costOverride = &CostBreakdown{
			TotalCost:   normalizeNonNegativeFloat(task.EstimatedCost),
			ActualCost:  prepaidCost,
			BillingMode: string(BillingModePerRequest),
		}
		if costOverride.TotalCost <= 0 {
			costOverride.TotalCost = prepaidCost
		}
	}
	fields := ChannelUsageFields{
		ChannelID:          task.ChannelID,
		OriginalModel:      firstNonEmptyString(task.OriginalModel, task.Model),
		ChannelMappedModel: firstNonEmptyString(task.ChannelMappedModel, task.Model),
		BillingModelSource: task.BillingModelSource,
		ModelMappingChain:  task.ModelMappingChain,
	}
	if fields.BillingModelSource == "" {
		fields.BillingModelSource = BillingModelSourceChannelMapped
	}
	videoBillingTier, videoRequestCount := extractOpenAIVideoBillingTierAndCount(task.SubmitResponse, input.StatusBody)

	err = s.RecordUsage(ctx, &OpenAIRecordUsageInput{
		Result:               result,
		APIKey:               apiKey,
		User:                 user,
		Account:              account,
		Subscription:         input.Subscription,
		InboundEndpoint:      input.InboundEndpoint,
		UpstreamEndpoint:     input.UpstreamEndpoint,
		UserAgent:            input.UserAgent,
		IPAddress:            input.IPAddress,
		RequestPayloadHash:   firstNonEmptyString(task.RequestPayloadHash, HashUsageRequestPayload(task.SubmitResponse)),
		RequestIDOverride:    "video_task:" + taskID,
		MediaType:            "video",
		BillingTierOverride:  videoBillingTier,
		RequestCountOverride: videoRequestCount,
		PrepaidBalanceCost:   prepaidCost,
		CostOverride:         costOverride,
		APIKeyService:        input.APIKeyService,
		ChannelUsageFields:   fields,
	})
	if err != nil {
		return err
	}
	update.BillingStatus = OpenAIVideoTaskBillingCharged
	update.CompletedAt = &now
	update.BilledAt = &now
	updated, err := s.openaiVideoTaskRepo.UpdateStatus(ctx, update)
	if err != nil {
		logger.L().With(zap.String("component", "service.openai_gateway")).Warn("openai_video_task.update_charged_failed", zap.String("task_id", taskID), zap.Error(err))
		return err
	}
	if updated != nil && updated.UsageLogID == nil && s.usageLogRepo != nil {
		// usage_logs has its own request_id/api_key_id idempotency; leaving usage_log_id
		// empty is acceptable when the async best-effort writer has not flushed yet.
	}
	return nil
}

func (s *OpenAIGatewayService) ResolveOpenAIVideoTaskAccount(ctx context.Context, taskID string) (*Account, bool) {
	if s == nil {
		return nil, false
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, false
	}
	raw, ok := s.openaiVideoTaskAccounts.Load(taskID)
	if !ok {
		if s.openaiVideoTaskRepo == nil {
			return nil, false
		}
		task, err := s.openaiVideoTaskRepo.GetByTaskID(ctx, OpenAIVideoTaskProviderOpenAI, taskID)
		if err != nil || task == nil || task.AccountID <= 0 {
			return nil, false
		}
		account, err := s.resolveOpenAIVideoTaskAccountByID(ctx, task.AccountID)
		if err != nil || !isOpenAIAccountEligibleForRequest(ctx, account, "", false, OpenAIEndpointCapabilityChatCompletions, AccountCapabilityVideo) || account.Type != AccountTypeAPIKey {
			return nil, false
		}
		s.rememberOpenAIVideoTaskAccountID(taskID, task.AccountID)
		return account, true
	}
	ref, ok := raw.(openAIVideoTaskAccountRef)
	if !ok || ref.AccountID <= 0 {
		s.openaiVideoTaskAccounts.Delete(taskID)
		return nil, false
	}
	if !ref.ExpiresAt.IsZero() && time.Now().After(ref.ExpiresAt) {
		s.openaiVideoTaskAccounts.Delete(taskID)
		return nil, false
	}
	account, err := s.resolveOpenAIVideoTaskAccountByID(ctx, ref.AccountID)
	if err != nil || !isOpenAIAccountEligibleForRequest(ctx, account, "", false, OpenAIEndpointCapabilityChatCompletions, AccountCapabilityVideo) || account.Type != AccountTypeAPIKey {
		return nil, false
	}
	return account, true
}

func (s *OpenAIGatewayService) GetOpenAIVideoTask(ctx context.Context, taskID string) (*OpenAIVideoTask, bool) {
	if s == nil || s.openaiVideoTaskRepo == nil {
		return nil, false
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, false
	}
	task, err := s.openaiVideoTaskRepo.GetByTaskID(ctx, OpenAIVideoTaskProviderOpenAI, taskID)
	if err != nil || task == nil {
		return nil, false
	}
	return task, true
}

func (s *OpenAIGatewayService) resolveOpenAIVideoTaskAccountByID(ctx context.Context, accountID int64) (*Account, error) {
	if s == nil || accountID <= 0 {
		return nil, fmt.Errorf("invalid account id")
	}
	if s.schedulerSnapshot != nil || s.accountRepo != nil {
		return s.getSchedulableAccount(ctx, accountID)
	}
	return nil, fmt.Errorf("account lookup unavailable")
}

func (s *OpenAIGatewayService) rememberOpenAIVideoTaskAccount(body []byte, accountID int64) {
	if s == nil || accountID <= 0 || len(body) == 0 {
		return
	}
	taskID := extractOpenAIVideoTaskID(body)
	if taskID == "" {
		return
	}
	s.rememberOpenAIVideoTaskAccountID(taskID, accountID)
}

func (s *OpenAIGatewayService) rememberOpenAIVideoTaskAccountID(taskID string, accountID int64) {
	if s == nil || accountID <= 0 {
		return
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return
	}
	s.openaiVideoTaskAccounts.Store(taskID, openAIVideoTaskAccountRef{
		AccountID: accountID,
		ExpiresAt: time.Now().Add(openAIVideoTaskAccountTTL),
	})
}

func extractOpenAIVideoTaskID(body []byte) string {
	for _, path := range []string{"task_id", "id", "data.task_id", "data.id", "data.0.task_id", "data.0.id"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func extractOpenAIVideoTaskStatus(body []byte) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	for _, path := range []string{"status", "data.status", "data.0.status", "state", "data.state", "task_status", "data.task_status"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return strings.ToLower(value)
		}
	}
	return ""
}

func isOpenAIVideoTaskSuccessStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "success", "succeeded", "completed", "complete", "done", "finished":
		return true
	default:
		return false
	}
}

func isOpenAIVideoTaskFailureStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "failure", "error", "canceled", "cancelled", "timeout", "expired":
		return true
	default:
		return false
	}
}

func extractOpenAIVideoBillingTier(submitBody []byte, statusBody []byte) string {
	tier, _ := extractOpenAIVideoBillingTierAndCount(submitBody, statusBody)
	return tier
}

func extractOpenAIVideoBillingTierAndCount(submitBody []byte, statusBody []byte) (string, int) {
	resolution := firstNonEmptyString(
		videoJSONString(submitBody, "resolution"),
		videoJSONString(submitBody, "size"),
		videoJSONString(submitBody, "quality"),
		videoJSONString(statusBody, "resolution"),
		videoJSONString(statusBody, "size"),
		videoJSONString(statusBody, "quality"),
	)
	duration := firstNonEmptyString(
		videoJSONNumberString(submitBody, "duration"),
		videoJSONNumberString(submitBody, "duration_seconds"),
		videoJSONNumberString(submitBody, "seconds"),
		videoJSONNumberString(statusBody, "duration"),
		videoJSONNumberString(statusBody, "duration_seconds"),
		videoJSONNumberString(statusBody, "seconds"),
	)
	resolution = normalizeOpenAIVideoTierPart(resolution)
	duration = normalizeOpenAIVideoTierPart(duration)
	switch {
	case resolution != "" && duration != "":
		return resolution + ":" + duration + "s", parseOpenAIVideoDurationCount(duration)
	case resolution != "":
		return resolution, 1
	case duration != "":
		return duration + "s", parseOpenAIVideoDurationCount(duration)
	default:
		return "", 1
	}
}

func openAIVideoBaseBillingTier(tier string) string {
	base, _, ok := strings.Cut(strings.TrimSpace(tier), ":")
	if !ok {
		return ""
	}
	return strings.TrimSpace(base)
}

func parseOpenAIVideoDurationCount(duration string) int {
	duration = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(duration), "s"))
	if duration == "" {
		return 1
	}
	value, err := strconv.ParseFloat(duration, 64)
	if err != nil || value <= 0 {
		return 1
	}
	count := int(value)
	if float64(count) < value {
		count++
	}
	if count <= 0 {
		return 1
	}
	return count
}

func shouldApplyOpenAIVideoPerSecondPricing(resolved *ResolvedPricing, tier string, requestCount int) bool {
	if resolved == nil || requestCount <= 1 {
		return false
	}
	if hasExactRequestTierPrice(resolved, tier) {
		return false
	}
	baseTier := openAIVideoBaseBillingTier(tier)
	if baseTier == "" {
		return false
	}
	if !hasExactRequestTierPrice(resolved, baseTier) {
		return false
	}
	return true
}

func hasExactRequestTierPrice(resolved *ResolvedPricing, tierLabel string) bool {
	tierLabel = strings.TrimSpace(tierLabel)
	if resolved == nil || tierLabel == "" {
		return false
	}
	for _, tier := range resolved.RequestTiers {
		if strings.TrimSpace(tier.TierLabel) == tierLabel && tier.PerRequestPrice != nil && *tier.PerRequestPrice > 0 {
			return true
		}
	}
	return false
}

func videoJSONString(body []byte, path string) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	for _, prefix := range []string{"", "data.", "data.0."} {
		if value := strings.TrimSpace(gjson.GetBytes(body, prefix+path).String()); value != "" {
			return value
		}
	}
	return ""
}

func videoJSONNumberString(body []byte, path string) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	for _, prefix := range []string{"", "data.", "data.0."} {
		value := gjson.GetBytes(body, prefix+path)
		if value.Exists() {
			if value.Type == gjson.Number {
				return strings.TrimRight(strings.TrimRight(value.Raw, "0"), ".")
			}
			if raw := strings.TrimSpace(value.String()); raw != "" {
				raw = strings.TrimSuffix(strings.ToLower(raw), "s")
				return raw
			}
		}
	}
	return ""
}

func normalizeOpenAIVideoTierPart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(strings.ToLower(value), "p")
	if value == "" {
		return ""
	}
	return value
}

func (s *OpenAIGatewayService) ForwardVideoTask(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	taskID string,
	language string,
) ([]byte, error) {
	if account == nil || account.Type != AccountTypeAPIKey {
		return nil, fmt.Errorf("tasks API requires an OpenAI API key account")
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		writeOpenAIVideosError(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return nil, fmt.Errorf("missing task_id")
	}
	token := account.GetOpenAIApiKey()
	if token == "" {
		return nil, fmt.Errorf("account %d missing api_key", account.ID)
	}
	targetURL, err := s.openAICompatibleEndpointURL(account, openAITasksEndpointPrefix+url.PathEscape(taskID))
	if err != nil {
		return nil, err
	}
	if language = strings.TrimSpace(language); language != "" {
		targetURL += "?language=" + url.QueryEscape(language)
	}

	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	req, err := http.NewRequestWithContext(upstreamCtx, http.MethodGet, targetURL, nil)
	releaseUpstreamCtx()
	if err != nil {
		return nil, fmt.Errorf("build upstream request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	for key, values := range c.Request.Header {
		if openaiCCRawAllowedHeaders[strings.ToLower(key)] {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	if customUA := account.GetOpenAIUserAgent(); customUA != "" {
		req.Header.Set("user-agent", customUA)
	}

	respBody, resp, err := s.doOpenAICompatibleJSONRequest(c, req, account)
	if err != nil {
		return respBody, err
	}
	writeOpenAICompatibleJSONResponse(c, resp, respBody, s.responseHeaderFilter)
	return respBody, nil
}

func (s *OpenAIGatewayService) openAICompatibleEndpointURL(account *Account, endpoint string) (string, error) {
	baseURL := "https://api.openai.com"
	if account != nil && strings.TrimSpace(account.GetOpenAIBaseURL()) != "" {
		baseURL = account.GetOpenAIBaseURL()
	}
	validatedURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return buildOpenAIEndpointURL(validatedURL, endpoint), nil
}

func (s *OpenAIGatewayService) doOpenAICompatibleJSONRequest(c *gin.Context, req *http.Request, account *Account) ([]byte, *http.Response, error) {
	proxyURL := ""
	if account != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	upstreamStart := time.Now()
	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(upstreamStart).Milliseconds())
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			UpstreamURL:        safeUpstreamURL(req.URL.String()),
			Kind:               "request_error",
			Message:            safeErr,
		})
		writeOpenAIVideosError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
		return nil, nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	if resp == nil {
		writeOpenAIVideosError(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
		return nil, nil, fmt.Errorf("upstream request failed: empty response")
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, openAIVideosMaxResponseBytes))
	if readErr != nil {
		writeOpenAIVideosError(c, http.StatusBadGateway, "api_error", "Failed to read upstream response")
		return nil, resp, fmt.Errorf("read upstream body: %w", readErr)
	}

	if resp.StatusCode >= 400 {
		upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
		upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
		if s.shouldFailoverOpenAIUpstreamResponse(resp.StatusCode, upstreamMsg, respBody) {
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				UpstreamURL:        safeUpstreamURL(req.URL.String()),
				Kind:               "failover",
				Message:            upstreamMsg,
			})
			if s.rateLimitService != nil {
				s.rateLimitService.HandleUpstreamError(req.Context(), account, resp.StatusCode, resp.Header, respBody)
			}
			return nil, resp, &UpstreamFailoverError{
				StatusCode:             resp.StatusCode,
				ResponseBody:           respBody,
				RetryableOnSameAccount: account.IsPoolMode() && isPoolModeRetryableStatus(resp.StatusCode),
			}
		}
		writeOpenAICompatibleJSONResponse(c, resp, respBody, s.responseHeaderFilter)
		return nil, resp, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	if code := gjson.GetBytes(respBody, "code"); code.Exists() && code.Int() != 0 && code.Int() != 200 {
		message := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(respBody))
		if message == "" {
			message = fmt.Sprintf("code %d", code.Int())
		}
		setOpsUpstreamError(c, resp.StatusCode, message, "")
		writeOpenAICompatibleJSONResponse(c, resp, respBody, s.responseHeaderFilter)
		return respBody, resp, fmt.Errorf("upstream returned code %d", code.Int())
	}

	return respBody, resp, nil
}

func writeOpenAICompatibleJSONResponse(c *gin.Context, resp *http.Response, body []byte, filter *responseheaders.CompiledHeaderFilter) {
	if c == nil || resp == nil || c.Writer.Written() {
		return
	}
	if resp.Header != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, filter)
	}
	if ct := strings.TrimSpace(resp.Header.Get("Content-Type")); ct != "" {
		c.Writer.Header().Set("Content-Type", ct)
	} else {
		c.Writer.Header().Set("Content-Type", "application/json")
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(body)
}

func (s *OpenAIGatewayService) WriteOpenAIVideoForwardResult(c *gin.Context, result *OpenAIForwardResult) {
	if s == nil || c == nil || result == nil {
		return
	}
	resp := &http.Response{StatusCode: http.StatusOK, Header: result.ResponseHeaders}
	writeOpenAICompatibleJSONResponse(c, resp, result.ResponseBody, s.responseHeaderFilter)
}

func NormalizeOpenAIModelForUpstreamForHandler(account *Account, model string) string {
	return normalizeOpenAIModelForUpstream(account, model)
}

func (s *OpenAIGatewayService) invalidateOpenAIVideoBalanceCache(ctx context.Context, userID int64) {
	if s == nil || s.billingCacheService == nil || userID <= 0 {
		return
	}
	if err := s.billingCacheService.InvalidateUserBalance(ctx, userID); err != nil {
		logger.L().With(zap.String("component", "service.openai_gateway")).Warn("openai_video_task.balance_cache_invalidate_failed", zap.Int64("user_id", userID), zap.Error(err))
	}
}

func writeOpenAIVideosError(c *gin.Context, statusCode int, errType, message string) {
	if c == nil || c.Writer.Written() {
		return
	}
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func IsVideoGenerationIntent(endpoint, requestedModel string, body []byte) bool {
	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	if strings.Contains(endpoint, "/videos/") {
		return true
	}
	model := strings.ToLower(strings.TrimSpace(requestedModel))
	if model == "" && len(body) > 0 && gjson.ValidBytes(body) {
		model = strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "model").String()))
	}
	if model == "" {
		return false
	}
	videoModelHints := []string{
		"kling",
		"wan",
		"veo",
		"seedance",
		"sora",
		"video",
	}
	for _, hint := range videoModelHints {
		if strings.Contains(model, hint) {
			return true
		}
	}
	return false
}

func IsMediaGenerationIntent(endpoint, requestedModel string, body []byte) bool {
	return IsImageGenerationIntent(endpoint, requestedModel, body) || IsVideoGenerationIntent(endpoint, requestedModel, body)
}
