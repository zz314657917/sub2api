package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// Videos handles OpenAI-compatible asynchronous video generation.
// POST /v1/videos/generations
func (h *OpenAIGatewayHandler) Videos(c *gin.Context) {
	streamStarted := false
	requestStart := requestStartedAt(c)

	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(
		c,
		"handler.openai_gateway.videos",
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
	)
	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			h.errorResponse(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}
	if !gjson.ValidBytes(body) {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}
	modelResult := gjson.GetBytes(body, "model")
	if !modelResult.Exists() || modelResult.Type != gjson.String || strings.TrimSpace(modelResult.String()) == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	reqModel := modelResult.String()
	reqLog = reqLog.With(zap.String("model", reqModel))
	setOpsRequestContext(c, reqModel, false, body)
	setOpsEndpointContext(c, "", int16(service.RequestTypeSync))

	if resolved, ok := h.resolveAPIKeyForModelRequest(c, apiKey, reqModel, false); !ok {
		return
	} else {
		apiKey = resolved
		reqLog = reqLog.With(zap.Any("resolved_group_id", apiKey.GroupID))
	}
	if apiKey.Group == nil || apiKey.Group.Platform != service.PlatformOpenAI {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Videos API is not supported for this platform")
		return
	}
	if decision := h.checkSecurityAudit(c, reqLog, apiKey, subject, "openai_videos", reqModel, body); decision != nil && !decision.AllowNextStage {
		h.openAISecurityAuditError(c, decision)
		return
	}

	channelMapping, _ := h.gatewayService.ResolveChannelMappingAndRestrict(c.Request.Context(), apiKey.GroupID, reqModel)
	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	service.SetOpsLatencyMs(c, service.OpsAuthLatencyMsKey, time.Since(requestStart).Milliseconds())

	userReleaseFunc, acquired := h.acquireResponsesUserSlot(c, subject.UserID, subject.Concurrency, false, &streamStarted, reqLog)
	if !acquired {
		return
	}
	if userReleaseFunc != nil {
		defer userReleaseFunc()
	}
	if err := h.billingCacheService.CheckBillingEligibility(c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription); err != nil {
		reqLog.Info("openai_videos.billing_check_failed", zap.Error(err))
		status, code, message, retryAfter := billingErrorDetails(err)
		if retryAfter > 0 {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
		}
		h.errorResponse(c, status, code, message)
		return
	}

	failedAccountIDs := make(map[int64]struct{})
	var lastFailoverErr *service.UpstreamFailoverError
	switchCount := 0
	maxAccountSwitches := h.maxAccountSwitches
	if maxAccountSwitches <= 0 {
		maxAccountSwitches = 3
	}
	sessionHash := h.gatewayService.GenerateExplicitSessionHash(c, body)
	routingStart := time.Now()

	for {
		selection, _, err := h.gatewayService.SelectAccountWithSchedulerForMediaCapabilityForUser(
			c.Request.Context(),
			apiKey.GroupID,
			"",
			sessionHash,
			reqModel,
			failedAccountIDs,
			service.OpenAIUpstreamTransportHTTPSSE,
			"",
			service.AccountCapabilityVideo,
			false,
			subject.UserID,
		)
		if err != nil {
			reqLog.Warn("openai_videos.account_select_failed", zap.Error(err), zap.Int("excluded_account_count", len(failedAccountIDs)))
			if len(failedAccountIDs) == 0 {
				markOpsRoutingCapacityLimitedIfNoAvailable(c, err)
				h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Service temporarily unavailable")
				return
			}
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, false)
			} else {
				h.errorResponse(c, http.StatusBadGateway, "api_error", "Upstream request failed")
			}
			return
		}
		if selection == nil || selection.Account == nil {
			markOpsRoutingCapacityLimited(c)
			h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "No available accounts")
			return
		}
		account := selection.Account
		if account.Type != service.AccountTypeAPIKey {
			if selection.ReleaseFunc != nil {
				selection.ReleaseFunc()
			}
			failedAccountIDs[account.ID] = struct{}{}
			continue
		}
		setOpsSelectedAccount(c, account.ID, account.Platform)

		accountReleaseFunc, accountAcquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, sessionHash, selection, false, &streamStarted, reqLog)
		if !accountAcquired {
			return
		}
		service.SetOpsLatencyMs(c, service.OpsRoutingLatencyMsKey, time.Since(routingStart).Milliseconds())
		forwardStart := time.Now()
		forwardBody := body
		if channelMapping.Mapped {
			forwardBody = h.gatewayService.ReplaceModelInBody(body, channelMapping.MappedModel)
		}
		upstreamBillingModel := strings.TrimSpace(gjson.GetBytes(forwardBody, "model").String())
		upstreamModel := service.NormalizeOpenAIModelForUpstreamForHandler(account, upstreamBillingModel)
		usageFields := channelMapping.ToUsageFields(reqModel, upstreamModel)
		requestPayloadHash := service.HashUsageRequestPayload(body)
		placeholderTaskID := ""
		var estimatedCost *service.CostBreakdown
		reservedCost := 0.0
		walletReservation := h.shouldReserveOpenAIVideoWallet(apiKey, subscription)
		if walletReservation {
			placeholderTaskID = "pending:" + requestPayloadHash + ":" + strconv.FormatInt(time.Now().UnixNano(), 36)
			reservedTask, cost, reserveErr := h.gatewayService.ReserveOpenAIVideoTaskBalance(c.Request.Context(), &service.OpenAIVideoTaskReserveInput{
				PlaceholderTaskID:  placeholderTaskID,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				RequestBody:        forwardBody,
				RequestPayloadHash: requestPayloadHash,
				ChannelUsageFields: usageFields,
			})
			if reserveErr != nil {
				reqLog.Info("openai_videos.reserve_balance_failed", zap.Error(reserveErr))
				status, code, message, retryAfter := billingErrorDetails(reserveErr)
				if retryAfter > 0 {
					c.Header("Retry-After", strconv.Itoa(retryAfter))
				}
				if errors.Is(reserveErr, service.ErrOpenAIVideoPricingUnavailable) {
					status = http.StatusForbidden
					code = "billing_error"
					message = "Video model pricing is not configured"
				}
				h.errorResponse(c, status, code, message)
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
				return
			}
			estimatedCost = cost
			if estimatedCost != nil {
				reservedCost = estimatedCost.ActualCost
			}
			if reservedTask != nil && reservedTask.TaskID != "" {
				placeholderTaskID = reservedTask.TaskID
			}
		} else if !h.isSimpleRunMode() {
			cost, _, estimateErr := h.gatewayService.EstimateOpenAIVideoCost(c.Request.Context(), apiKey, apiKey.User, account, forwardBody, usageFields)
			if estimateErr != nil {
				reqLog.Info("openai_videos.estimate_cost_failed", zap.Error(estimateErr))
				status, code, message, retryAfter := billingErrorDetails(estimateErr)
				if retryAfter > 0 {
					c.Header("Retry-After", strconv.Itoa(retryAfter))
				}
				if errors.Is(estimateErr, service.ErrOpenAIVideoPricingUnavailable) {
					status = http.StatusForbidden
					code = "billing_error"
					message = "Video model pricing is not configured"
				}
				h.errorResponse(c, status, code, message)
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
				return
			}
			estimatedCost = cost
		}
		writerSizeBeforeForward := c.Writer.Size()
		result, err := func() (*service.OpenAIForwardResult, error) {
			defer func() {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
			}()
			return h.gatewayService.ForwardVideosWithOptions(c.Request.Context(), c, account, forwardBody, "", service.OpenAIVideoForwardOptions{WriteResponse: false})
		}()

		forwardDurationMs := time.Since(forwardStart).Milliseconds()
		if upstreamLatencyMs, _ := getContextInt64(c, service.OpsUpstreamLatencyMsKey); upstreamLatencyMs > 0 && forwardDurationMs > upstreamLatencyMs {
			service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, forwardDurationMs-upstreamLatencyMs)
		} else {
			service.SetOpsLatencyMs(c, service.OpsResponseLatencyMsKey, forwardDurationMs)
		}
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				h.refundOpenAIVideoReservationIfNeeded(c, walletReservation, placeholderTaskID)
				if c.Writer.Size() != writerSizeBeforeForward {
					h.handleFailoverExhausted(c, failoverErr, true)
					return
				}
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				h.gatewayService.RecordOpenAIAccountSwitch()
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				if switchCount >= maxAccountSwitches {
					h.handleFailoverExhausted(c, failoverErr, false)
					return
				}
				switchCount++
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			h.refundOpenAIVideoReservationIfNeeded(c, walletReservation, placeholderTaskID)
			if c.Writer.Size() == writerSizeBeforeForward {
				h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
			}
			reqLog.Warn("openai_videos.forward_failed", zap.Int64("account_id", account.ID), zap.Error(err))
			return
		}

		h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
		if strings.TrimSpace(result.VideoTaskID) == "" {
			h.refundOpenAIVideoReservationIfNeeded(c, walletReservation, placeholderTaskID)
			h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream response missing task_id")
			return
		}
		recordInput := &service.OpenAIVideoTaskRecordInput{
			Result:             result,
			APIKey:             apiKey,
			User:               apiKey.User,
			Account:            account,
			EstimatedCost:      estimatedCost,
			ReservedCost:       reservedCost,
			RequestPayloadHash: requestPayloadHash,
			ChannelUsageFields: usageFields,
		}
		var recordErr error
		if walletReservation {
			_, recordErr = h.gatewayService.BindReservedOpenAIVideoTask(c.Request.Context(), placeholderTaskID, recordInput)
		} else {
			recordErr = h.gatewayService.RecordOpenAIVideoTaskSubmitted(c.Request.Context(), recordInput)
		}
		if recordErr != nil {
			h.refundOpenAIVideoReservationIfNeeded(c, walletReservation, placeholderTaskID)
			reqLog.Error("openai_videos.bind_task_failed", zap.Error(recordErr), zap.String("task_id", result.VideoTaskID))
			h.errorResponse(c, http.StatusBadGateway, "api_error", "Failed to record video task")
			return
		}
		h.gatewayService.WriteOpenAIVideoForwardResult(c, result)
		return
	}
}

func (h *OpenAIGatewayHandler) shouldReserveOpenAIVideoWallet(apiKey *service.APIKey, subscription *service.UserSubscription) bool {
	if h == nil || h.isSimpleRunMode() || apiKey == nil || apiKey.Group == nil {
		return false
	}
	return !(subscription != nil && apiKey.Group.IsSubscriptionType())
}

func (h *OpenAIGatewayHandler) isSimpleRunMode() bool {
	return h != nil && h.cfg != nil && h.cfg.RunMode == config.RunModeSimple
}

func (h *OpenAIGatewayHandler) refundOpenAIVideoReservationIfNeeded(c *gin.Context, reserved bool, taskID string) {
	if h == nil || h.gatewayService == nil || !reserved || strings.TrimSpace(taskID) == "" {
		return
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	_ = h.gatewayService.RefundOpenAIVideoTaskReservation(ctx, taskID)
}

// VideoTask proxies asynchronous video task status.
// GET /v1/tasks/:task_id
func (h *OpenAIGatewayHandler) VideoTask(c *gin.Context) {
	streamStarted := false
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "User context not found")
		return
	}
	reqLog := requestLogger(c, "handler.openai_gateway.video_task", zap.Int64("user_id", subject.UserID), zap.Int64("api_key_id", apiKey.ID), zap.Any("group_id", apiKey.GroupID))
	if !h.ensureResponsesDependencies(c, reqLog) {
		return
	}
	if apiKey.Group == nil || apiKey.Group.Platform != service.PlatformOpenAI {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Tasks API is not supported for this platform")
		return
	}
	taskID := c.Param("task_id")
	if strings.TrimSpace(taskID) == "" {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}
	if task, found := h.gatewayService.GetOpenAIVideoTask(c.Request.Context(), taskID); found {
		if task.UserID != subject.UserID || task.APIKeyID != apiKey.ID {
			h.errorResponse(c, http.StatusNotFound, "not_found_error", "Task not found")
			return
		}
	}
	if account, found := h.gatewayService.ResolveOpenAIVideoTaskAccount(c.Request.Context(), taskID); found {
		setOpsSelectedAccount(c, account.ID, account.Platform)
		body, err := h.gatewayService.ForwardVideoTask(c.Request.Context(), c, account, taskID, c.DefaultQuery("language", "zh"))
		if err != nil {
			if !c.Writer.Written() {
				h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
			}
			return
		}
		h.settleOpenAIVideoTask(c, taskID, body, apiKey, subject, account)
		return
	}

	failedAccountIDs := make(map[int64]struct{})
	maxAccountSwitches := h.maxAccountSwitches
	if maxAccountSwitches <= 0 {
		maxAccountSwitches = 3
	}
	switchCount := 0
	var lastFailoverErr *service.UpstreamFailoverError
	for {
		selection, _, err := h.gatewayService.SelectAccountWithSchedulerForMediaCapabilityForUser(
			c.Request.Context(),
			apiKey.GroupID,
			"",
			"",
			"",
			failedAccountIDs,
			service.OpenAIUpstreamTransportHTTPSSE,
			"",
			service.AccountCapabilityVideo,
			false,
			subject.UserID,
		)
		if err != nil {
			if len(failedAccountIDs) == 0 {
				markOpsRoutingCapacityLimitedIfNoAvailable(c, err)
				h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "Service temporarily unavailable")
				return
			}
			if lastFailoverErr != nil {
				h.handleFailoverExhausted(c, lastFailoverErr, false)
			} else {
				h.errorResponse(c, http.StatusBadGateway, "api_error", "Upstream request failed")
			}
			return
		}
		if selection == nil || selection.Account == nil {
			markOpsRoutingCapacityLimited(c)
			h.errorResponse(c, http.StatusServiceUnavailable, "api_error", "No available accounts")
			return
		}
		account := selection.Account
		if account.Type != service.AccountTypeAPIKey {
			if selection.ReleaseFunc != nil {
				selection.ReleaseFunc()
			}
			failedAccountIDs[account.ID] = struct{}{}
			continue
		}
		setOpsSelectedAccount(c, account.ID, account.Platform)
		accountReleaseFunc, accountAcquired := h.acquireResponsesAccountSlot(c, apiKey.GroupID, "", selection, false, &streamStarted, reqLog)
		if !accountAcquired {
			return
		}
		writerSizeBeforeForward := c.Writer.Size()
		var statusBody []byte
		err = func() error {
			defer func() {
				if accountReleaseFunc != nil {
					accountReleaseFunc()
				}
			}()
			var forwardErr error
			statusBody, forwardErr = h.gatewayService.ForwardVideoTask(c.Request.Context(), c, account, taskID, c.DefaultQuery("language", "zh"))
			return forwardErr
		}()
		if err != nil {
			var failoverErr *service.UpstreamFailoverError
			if errors.As(err, &failoverErr) {
				if c.Writer.Size() != writerSizeBeforeForward {
					h.handleFailoverExhausted(c, failoverErr, true)
					return
				}
				h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
				h.gatewayService.RecordOpenAIAccountSwitch()
				failedAccountIDs[account.ID] = struct{}{}
				lastFailoverErr = failoverErr
				if switchCount >= maxAccountSwitches {
					h.handleFailoverExhausted(c, failoverErr, false)
					return
				}
				switchCount++
				continue
			}
			h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			if c.Writer.Size() == writerSizeBeforeForward {
				h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
			}
			return
		}
		h.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, nil)
		h.settleOpenAIVideoTask(c, taskID, statusBody, apiKey, subject, account)
		return
	}
}

func (h *OpenAIGatewayHandler) settleOpenAIVideoTask(c *gin.Context, taskID string, statusBody []byte, apiKey *service.APIKey, subject middleware2.AuthSubject, account *service.Account) {
	if h == nil || h.gatewayService == nil || len(statusBody) == 0 || apiKey == nil {
		return
	}
	subscription, _ := middleware2.GetSubscriptionFromContext(c)
	userAgent := c.GetHeader("User-Agent")
	clientIP := ip.GetClientIP(c)
	inboundEndpoint := GetInboundEndpoint(c)
	upstreamEndpoint := GetUpstreamEndpoint(c, service.PlatformOpenAI)
	h.submitMandatoryUsageRecordTask(c.Request.Context(), func(ctx context.Context) {
		if err := h.gatewayService.SettleOpenAIVideoTaskIfTerminal(ctx, &service.OpenAIVideoTaskSettleInput{
			TaskID:           taskID,
			StatusBody:       statusBody,
			APIKey:           apiKey,
			User:             apiKey.User,
			Account:          account,
			Subscription:     subscription,
			InboundEndpoint:  inboundEndpoint,
			UpstreamEndpoint: upstreamEndpoint,
			UserAgent:        userAgent,
			IPAddress:        clientIP,
			APIKeyService:    h.apiKeyService,
		}); err != nil {
			logger.L().With(
				zap.String("component", "handler.openai_gateway.video_task"),
				zap.Int64("user_id", subject.UserID),
				zap.Int64("api_key_id", apiKey.ID),
				zap.Any("group_id", apiKey.GroupID),
				zap.String("task_id", taskID),
			).Error("openai_video_task.settle_failed", zap.Error(err))
		}
	})
}
