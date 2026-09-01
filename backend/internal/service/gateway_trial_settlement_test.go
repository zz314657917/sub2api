package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type gatewayTrialSettlementLogRepo struct {
	UsageLogRepository

	events  *[]string
	payload *UsageBillingSettlementPayload
	applied []int64
	markErr error
	owned   *bool
}

func (r *gatewayTrialSettlementLogRepo) CreatePendingPayload(_ context.Context, log *UsageLog, payload *UsageBillingSettlementPayload) error {
	log.ID = 101
	r.payload = payload
	*r.events = append(*r.events, "persist-pending")
	return nil
}

func (r *gatewayTrialSettlementLogRepo) CreatePendingPayloadWithOwnership(_ context.Context, log *UsageLog, payload *UsageBillingSettlementPayload) (bool, bool, error) {
	log.ID = 101
	r.payload = payload
	*r.events = append(*r.events, "persist-pending")
	if r.owned == nil {
		return true, false, nil
	}
	return *r.owned, false, nil
}

func (r *gatewayTrialSettlementLogRepo) MarkApplied(_ context.Context, usageLogID int64) error {
	if r.markErr != nil {
		return r.markErr
	}
	r.applied = append(r.applied, usageLogID)
	*r.events = append(*r.events, "mark-applied")
	return nil
}

func (r *gatewayTrialSettlementLogRepo) Create(_ context.Context, _ *UsageLog) (bool, error) {
	return true, nil
}

func (r *gatewayTrialSettlementLogRepo) CreatePending(context.Context, *UsageLog, *UsageBillingCommand) error {
	return nil
}

func (r *gatewayTrialSettlementLogRepo) MarkPendingError(context.Context, int64, error) error {
	return nil
}

func (r *gatewayTrialSettlementLogRepo) ClaimPending(context.Context, int, time.Duration) ([]UsageBillingSettlementTask, error) {
	return nil, nil
}

func (r *gatewayTrialSettlementLogRepo) MarkRetry(context.Context, int64, int, error, time.Time, bool) error {
	return nil
}

type gatewayTrialSettlementBillingRepo struct {
	UsageBillingRepository

	events   *[]string
	commands []UsageBillingCommand
}

func (r *gatewayTrialSettlementBillingRepo) Apply(_ context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	r.commands = append(r.commands, *cmd)
	*r.events = append(*r.events, "apply:"+cmd.RequestID)
	return &UsageBillingApplyResult{Applied: true}, nil
}

func newGatewayTrialSettlementService(logRepo UsageLogRepository, billingRepo UsageBillingRepository) *GatewayService {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1
	return NewGatewayService(
		nil, nil, logRepo, billingRepo, nil, nil, nil, nil, cfg, nil, nil,
		NewBillingService(cfg, nil), nil, &BillingCacheService{}, nil, nil, &DeferredService{}, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
}

func newOpenAITrialSettlementService(logRepo UsageLogRepository, billingRepo UsageBillingRepository) *OpenAIGatewayService {
	cfg := &config.Config{}
	cfg.Default.RateMultiplier = 1
	return NewOpenAIGatewayService(
		nil, logRepo, billingRepo, nil, nil, nil, nil, cfg, nil, nil,
		NewBillingService(cfg, nil), nil, &BillingCacheService{}, nil, &DeferredService{}, nil, nil,
		nil, nil, nil, nil, nil,
	)
}

func assertGatewayTrialSettlement(t *testing.T, events []string, logRepo *gatewayTrialSettlementLogRepo, billingRepo *gatewayTrialSettlementBillingRepo, requestID string) {
	t.Helper()
	require.NotNil(t, logRepo.payload)
	require.False(t, logRepo.payload.Primary.FinalizeUsageLog)
	require.Equal(t, requestID, logRepo.payload.Primary.RequestID)
	require.NotNil(t, logRepo.payload.Overage)
	require.Equal(t, requestID+newUserTrialOverageRequestIDSuffix, logRepo.payload.Overage.RequestID)

	require.Len(t, billingRepo.commands, 2)
	require.Equal(t, requestID, billingRepo.commands[0].RequestID)
	require.Equal(t, requestID+newUserTrialOverageRequestIDSuffix, billingRepo.commands[1].RequestID)
	require.Equal(t, []string{
		"persist-pending",
		"apply:" + requestID,
		"apply:" + requestID + newUserTrialOverageRequestIDSuffix,
		"mark-applied",
	}, events)
	require.Equal(t, []int64{101}, logRepo.applied)
}

func TestGatewayTrialSettlement_DefersAppliedUntilPrimaryAndOverageComplete(t *testing.T) {
	events := []string{}
	logRepo := &gatewayTrialSettlementLogRepo{events: &events}
	billingRepo := &gatewayTrialSettlementBillingRepo{events: &events}
	svc := newGatewayTrialSettlementService(logRepo, billingRepo)
	cache := &settlementCacheStub{}
	svc.billingCacheService = &BillingCacheService{cache: cache}
	svc.welfareService = &WelfareService{repo: &welfareRepoStub{trials: []WelfareNewUserTrial{{
		ID: 1, UserID: 2, QuotaAmount: 1, Status: "in_progress", LastRequestID: "trial-session-gateway",
	}}}}
	requestID := "gateway-trial-settlement"

	ctx := WithNewUserTrialSession(context.Background(), &NewUserTrialSession{
		TrialID: 1, UserID: 2, RequestID: "trial-session-gateway", QuotaLeft: 0.01,
	})
	err := svc.RecordUsage(ctx, &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: requestID,
			Model:     "claude-sonnet-4",
			Usage:     ClaudeUsage{InputTokens: 100000, OutputTokens: 100000},
			Duration:  time.Second,
		},
		APIKey:  &APIKey{ID: 3},
		User:    &User{ID: 2},
		Account: &Account{ID: 4, Type: AccountTypeAPIKey},
	})

	require.NoError(t, err)
	assertGatewayTrialSettlement(t, events, logRepo, billingRepo, requestID)
	require.Equal(t, []int64{2}, cache.invalidatedUsers)
}

func TestOpenAIGatewayTrialSettlement_DefersAppliedUntilPrimaryAndOverageComplete(t *testing.T) {
	events := []string{}
	logRepo := &gatewayTrialSettlementLogRepo{events: &events}
	billingRepo := &gatewayTrialSettlementBillingRepo{events: &events}
	svc := newOpenAITrialSettlementService(logRepo, billingRepo)
	svc.welfareService = &WelfareService{repo: &welfareRepoStub{trials: []WelfareNewUserTrial{{
		ID: 5, UserID: 6, QuotaAmount: 1, Status: "in_progress", LastRequestID: "trial-session-openai",
	}}}}
	requestID := "openai-trial-settlement"

	ctx := WithNewUserTrialSession(context.Background(), &NewUserTrialSession{
		TrialID: 5, UserID: 6, RequestID: "trial-session-openai", QuotaLeft: 0.01,
	})
	err := svc.RecordUsage(ctx, &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: requestID,
			Model:     "gpt-5.6",
			Usage:     OpenAIUsage{InputTokens: 100000, OutputTokens: 100000},
			Duration:  time.Second,
		},
		APIKey:  &APIKey{ID: 7},
		User:    &User{ID: 6},
		Account: &Account{ID: 8, Type: AccountTypeAPIKey},
		CostOverride: &CostBreakdown{
			InputCost: 0.2, OutputCost: 0.3, TotalCost: 0.5, ActualCost: 0.5,
		},
	})

	require.NoError(t, err)
	assertGatewayTrialSettlement(t, events, logRepo, billingRepo, requestID)
}

func TestGatewayTrialSettlement_DelaysPostSettlementUntilOutboxFinalized(t *testing.T) {
	events := []string{}
	logRepo := &gatewayTrialSettlementLogRepo{events: &events, markErr: errors.New("outbox finalization failed")}
	billingRepo := &gatewayTrialSettlementBillingRepo{events: &events}
	svc := newGatewayTrialSettlementService(logRepo, billingRepo)
	svc.welfareService = &WelfareService{repo: &welfareRepoStub{trials: []WelfareNewUserTrial{{
		ID: 11, UserID: 12, QuotaAmount: 1, Status: "in_progress", LastRequestID: "trial-session-post-finalize",
	}}}}

	ctx := WithNewUserTrialSession(context.Background(), &NewUserTrialSession{
		TrialID: 11, UserID: 12, RequestID: "trial-session-post-finalize", QuotaLeft: 0.01,
	})
	err := svc.RecordUsage(ctx, &RecordUsageInput{
		Result:  &ForwardResult{RequestID: "gateway-trial-post-finalize", Model: "claude-sonnet-4", Usage: ClaudeUsage{InputTokens: 100000, OutputTokens: 100000}, Duration: time.Second},
		APIKey:  &APIKey{ID: 13},
		User:    &User{ID: 12},
		Account: &Account{ID: 14, Type: AccountTypeAPIKey},
	})

	require.ErrorIs(t, err, logRepo.markErr)
	_, deferred := svc.deferredService.lastUsedUpdates.Load(int64(14))
	require.False(t, deferred, "post-settlement side effects require outbox finalization")
}

func TestGatewayTrialSettlement_ProcessingOutboxSkipsDuplicateDirectSettlement(t *testing.T) {
	events := []string{}
	owned := false
	logRepo := &gatewayTrialSettlementLogRepo{events: &events, owned: &owned}
	billingRepo := &gatewayTrialSettlementBillingRepo{events: &events}
	svc := newGatewayTrialSettlementService(logRepo, billingRepo)
	svc.welfareService = &WelfareService{repo: &welfareRepoStub{trials: []WelfareNewUserTrial{{
		ID: 15, UserID: 16, QuotaAmount: 1, Status: "in_progress", LastRequestID: "trial-session-processing",
	}}}}

	ctx := WithNewUserTrialSession(context.Background(), &NewUserTrialSession{
		TrialID: 15, UserID: 16, RequestID: "trial-session-processing", QuotaLeft: 0.01,
	})
	err := svc.RecordUsage(ctx, &RecordUsageInput{
		Result:  &ForwardResult{RequestID: "gateway-trial-processing", Model: "claude-sonnet-4", Usage: ClaudeUsage{InputTokens: 100000, OutputTokens: 100000}, Duration: time.Second},
		APIKey:  &APIKey{ID: 17},
		User:    &User{ID: 16},
		Account: &Account{ID: 18, Type: AccountTypeAPIKey},
	})

	require.NoError(t, err)
	require.Equal(t, []string{"persist-pending"}, events)
	require.Empty(t, billingRepo.commands)
	require.Empty(t, logRepo.applied)
}

func TestApplyUsageBillingWithNewUserTrialOverage_RetryReportsFinalSettlement(t *testing.T) {
	billing := &settlementBillingRepoFake{results: []*UsageBillingApplyResult{
		{Applied: false}, {Applied: false},
	}}
	welfare := &WelfareService{repo: &welfareRepoStub{trials: []WelfareNewUserTrial{{
		ID: 21, UserID: 22, QuotaAmount: 1, Status: "in_progress",
	}}}}
	deferred := &DeferredService{}
	p := &postUsageBillingParams{
		Cost:                    &CostBreakdown{InputCost: 0.2, OutputCost: 0.2, TotalCost: 0.4, ActualCost: 0.4},
		User:                    &User{ID: 22},
		APIKey:                  &APIKey{ID: 23},
		Account:                 &Account{ID: 24, Type: AccountTypeAPIKey},
		NewUserTrial:            &NewUserTrialSession{TrialID: 21, UserID: 22, RequestID: "trial-direct-retry", QuotaLeft: 0.1},
		DeferSettlementFinalize: true,
	}

	applied, err := applyUsageBillingWithNewUserTrialOverage(
		context.Background(),
		"request-direct-retry",
		&UsageLog{RequestID: "request-direct-retry", Model: "gpt-5.6"},
		p,
		&billingDeps{billingCacheService: &BillingCacheService{}, deferredService: deferred},
		billing,
		welfare,
	)

	require.NoError(t, err)
	require.True(t, applied, "composite completion must drive membership/affiliate after a deduplicated primary retry")
	_, scheduledEarly := deferred.lastUsedUpdates.Load(int64(24))
	require.False(t, scheduledEarly, "the caller owns deferred post-settlement effects until MarkApplied succeeds")
	require.Len(t, billing.calls, 2)
}

func TestFinalizeNewUserTrialOverageCacheRefreshesBalanceAndSubscription(t *testing.T) {
	cache := &settlementCacheStub{}
	deps := &billingDeps{billingCacheService: &BillingCacheService{cache: cache}}
	balanceParams := &postUsageBillingParams{
		Cost:         &CostBreakdown{ActualCost: 0.4},
		User:         &User{ID: 31},
		APIKey:       &APIKey{ID: 32},
		NewUserTrial: &NewUserTrialSession{QuotaLeft: 0.1},
	}
	finalizeNewUserTrialOverageCache(context.Background(), balanceParams, deps)
	require.Equal(t, []int64{31}, cache.invalidatedUsers)

	groupID := int64(34)
	subscriptionParams := &postUsageBillingParams{
		Cost:               &CostBreakdown{ActualCost: 0.4},
		User:               &User{ID: 33},
		APIKey:             &APIKey{ID: 35, GroupID: &groupID},
		NewUserTrial:       &NewUserTrialSession{QuotaLeft: 0.1},
		IsSubscriptionBill: true,
	}
	finalizeNewUserTrialOverageCache(context.Background(), subscriptionParams, deps)
	require.Len(t, cache.subscriptionWrites, 1)
	require.Equal(t, int64(33), cache.subscriptionWrites[0].userID)
	require.Equal(t, int64(34), cache.subscriptionWrites[0].groupID)
	require.InDelta(t, 0.3, cache.subscriptionWrites[0].cost, 0.00000001)
}

func TestGatewayTrialSettlement_FailsClosedWhenWelfareUnavailable(t *testing.T) {
	events := []string{}
	logRepo := &gatewayTrialSettlementLogRepo{events: &events}
	billingRepo := &gatewayTrialSettlementBillingRepo{events: &events}
	svc := newGatewayTrialSettlementService(logRepo, billingRepo)
	requestID := "gateway-trial-no-welfare"

	ctx := WithNewUserTrialSession(context.Background(), &NewUserTrialSession{
		TrialID: 1, UserID: 2, RequestID: "trial-session-no-welfare", QuotaLeft: 0.01,
	})
	err := svc.RecordUsage(ctx, &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: requestID,
			Model:     "claude-sonnet-4",
			Usage:     ClaudeUsage{InputTokens: 100000, OutputTokens: 100000},
			Duration:  time.Second,
		},
		APIKey:  &APIKey{ID: 3},
		User:    &User{ID: 2},
		Account: &Account{ID: 4, Type: AccountTypeAPIKey},
	})

	require.ErrorIs(t, err, ErrWelfareNewUserTrialUnavailable)
	require.Empty(t, logRepo.applied)
	require.Empty(t, billingRepo.commands)
}
