package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type settlementBillingRepoFake struct {
	calls         []UsageBillingCommand
	err           error
	result        *UsageBillingApplyResult
	errs          []error
	results       []*UsageBillingApplyResult
	started       chan struct{}
	waitForCancel bool
	startedOnce   sync.Once
}

func (f *settlementBillingRepoFake) Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	if cmd != nil {
		f.calls = append(f.calls, *cmd)
	}
	if f.started != nil {
		f.startedOnce.Do(func() { close(f.started) })
	}
	if f.waitForCancel {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	err := f.err
	if len(f.errs) > 0 {
		err = f.errs[0]
		f.errs = f.errs[1:]
	}
	if err != nil {
		return nil, err
	}
	if len(f.results) > 0 {
		result := f.results[0]
		f.results = f.results[1:]
		return result, nil
	}
	if f.result != nil {
		return f.result, nil
	}
	return &UsageBillingApplyResult{Applied: true}, nil
}

type settlementOutboxFake struct {
	tasks             []UsageBillingSettlementTask
	appliedIDs        []int64
	markAttemptCalls  int
	markAttemptResult *bool
	retries           int
	claimCalls        int
}

type settlementCacheStub struct {
	invalidatedUsers   []int64
	subscriptionWrites []settlementSubscriptionWrite
}

type settlementSubscriptionWrite struct {
	userID  int64
	groupID int64
	cost    float64
}

func (s *settlementCacheStub) GetUserBalance(context.Context, int64) (float64, error) {
	return 0, nil
}
func (s *settlementCacheStub) SetUserBalance(context.Context, int64, float64) error { return nil }
func (s *settlementCacheStub) DeductUserBalance(context.Context, int64, float64) error {
	return nil
}
func (s *settlementCacheStub) InvalidateUserBalance(_ context.Context, userID int64) error {
	s.invalidatedUsers = append(s.invalidatedUsers, userID)
	return nil
}
func (s *settlementCacheStub) GetSubscriptionCache(context.Context, int64, int64) (*SubscriptionCacheData, error) {
	return nil, nil
}
func (s *settlementCacheStub) SetSubscriptionCache(context.Context, int64, int64, *SubscriptionCacheData) error {
	return nil
}
func (s *settlementCacheStub) UpdateSubscriptionUsage(_ context.Context, userID, groupID int64, cost float64) error {
	s.subscriptionWrites = append(s.subscriptionWrites, settlementSubscriptionWrite{userID: userID, groupID: groupID, cost: cost})
	return nil
}
func (s *settlementCacheStub) InvalidateSubscriptionCache(context.Context, int64, int64) error {
	return nil
}
func (s *settlementCacheStub) GetAPIKeyRateLimit(context.Context, int64) (*APIKeyRateLimitCacheData, error) {
	return nil, nil
}
func (s *settlementCacheStub) SetAPIKeyRateLimit(context.Context, int64, *APIKeyRateLimitCacheData) error {
	return nil
}

func (s *settlementCacheStub) UpdateAPIKeyRateLimitUsage(context.Context, int64, float64) error {
	return nil
}
func (s *settlementCacheStub) InvalidateAPIKeyRateLimit(context.Context, int64) error { return nil }

func (f *settlementOutboxFake) CreatePending(context.Context, *UsageLog, *UsageBillingCommand) error {
	return nil
}
func (f *settlementOutboxFake) CreatePendingPayload(context.Context, *UsageLog, *UsageBillingSettlementPayload) error {
	return nil
}
func (f *settlementOutboxFake) MarkPendingError(context.Context, int64, error) error { return nil }
func (f *settlementOutboxFake) MarkApplied(_ context.Context, usageLogID int64) error {
	f.appliedIDs = append(f.appliedIDs, usageLogID)
	return nil
}
func (f *settlementOutboxFake) MarkAppliedAttempt(_ context.Context, _ int64, _ int, usageLogID int64) (bool, error) {
	f.markAttemptCalls++
	if f.markAttemptResult != nil {
		if *f.markAttemptResult {
			f.appliedIDs = append(f.appliedIDs, usageLogID)
		}
		return *f.markAttemptResult, nil
	}
	f.appliedIDs = append(f.appliedIDs, usageLogID)
	return true, nil
}
func (f *settlementOutboxFake) ClaimPending(context.Context, int, time.Duration) ([]UsageBillingSettlementTask, error) {
	f.claimCalls++
	tasks := f.tasks
	f.tasks = nil
	return tasks, nil
}

func TestUsageBillingSettlementService_StopCancelsActiveRunBeforeSettlementWrites(t *testing.T) {
	billing := &settlementBillingRepoFake{started: make(chan struct{}), waitForCancel: true}
	outbox := &settlementOutboxFake{tasks: []UsageBillingSettlementTask{{
		ID: 4, UsageLogID: 45, Attempts: 1,
		Payload: UsageBillingSettlementPayload{Version: 1, Primary: UsageBillingCommand{RequestID: "req-stop"}},
	}}}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox}
	done := make(chan struct{})
	go func() {
		svc.RunOnce()
		close(done)
	}()

	select {
	case <-billing.started:
	case <-time.After(time.Second):
		t.Fatal("settlement apply did not start")
	}
	svc.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("settlement run did not stop after cancellation")
	}

	require.Empty(t, outbox.appliedIDs)
	require.Zero(t, outbox.retries)
	require.Equal(t, 1, outbox.claimCalls)
	svc.RunOnce()
	require.Equal(t, 1, outbox.claimCalls, "stopped service must not restart a callback")
}

func TestUsageBillingSettlementService_StartStopDoesNotRescheduleTimer(t *testing.T) {
	billing := &settlementBillingRepoFake{}
	outbox := &settlementOutboxFake{}
	var scheduled []func()
	cancelled := 0
	svc := &UsageBillingSettlementService{
		billingRepo: billing,
		outbox:      outbox,
		timingWheel: &TimingWheelService{},
		scheduleFn: func(_ string, _ time.Duration, fn func()) {
			scheduled = append(scheduled, fn)
		},
		cancelFn: func(string) { cancelled++ },
	}

	svc.Start()
	require.Len(t, scheduled, 1)
	callback := scheduled[0]
	svc.Stop()
	callback()
	svc.Start()

	require.Len(t, scheduled, 1, "stopped callback and later Start must not schedule a successor")
	require.Equal(t, 1, cancelled)
}
func (f *settlementOutboxFake) MarkRetry(context.Context, int64, int, error, time.Time, bool) error {
	f.retries++
	return nil
}

type settlementAuthCacheInvalidatorStub struct {
	userIDs []int64
}

func (s *settlementAuthCacheInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string) {}
func (s *settlementAuthCacheInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
}
func (s *settlementAuthCacheInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}

type settlementWelfareFake struct {
	calls int
	err   error
	errs  []error
}

func (f *settlementWelfareFake) ConsumeNewUserTrial(context.Context, *NewUserTrialSession, string, float64, string, int64) error {
	f.calls++
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		return err
	}
	return f.err
}

type settlementPersistenceRepoFake struct {
	UsageLogRepository
	calls int
	errs  []error
}

func (f *settlementPersistenceRepoFake) CreatePending(context.Context, *UsageLog, *UsageBillingCommand) error {
	return nil
}
func (f *settlementPersistenceRepoFake) CreatePendingPayload(context.Context, *UsageLog, *UsageBillingSettlementPayload) error {
	f.calls++
	if len(f.errs) == 0 {
		return nil
	}
	err := f.errs[0]
	f.errs = f.errs[1:]
	return err
}
func (f *settlementPersistenceRepoFake) MarkPendingError(context.Context, int64, error) error {
	return nil
}
func (f *settlementPersistenceRepoFake) MarkApplied(context.Context, int64) error { return nil }
func (f *settlementPersistenceRepoFake) ClaimPending(context.Context, int, time.Duration) ([]UsageBillingSettlementTask, error) {
	return nil, nil
}
func (f *settlementPersistenceRepoFake) MarkRetry(context.Context, int64, int, error, time.Time, bool) error {
	return nil
}

func TestBuildUsageBillingSettlementPayloadIncludesTrialOverage(t *testing.T) {
	trial := &NewUserTrialSession{TrialID: 7, UserID: 8, RequestID: "trial-request", QuotaLeft: 0.25}
	p := &postUsageBillingParams{
		Cost:                    &CostBreakdown{ActualCost: 0.4, TotalCost: 0.4},
		User:                    &User{ID: 8},
		APIKey:                  &APIKey{ID: 9},
		Account:                 &Account{ID: 10, Type: AccountTypeAPIKey},
		NewUserTrial:            trial,
		DeferSettlementFinalize: true,
	}
	log := &UsageLog{RequestID: "req-1", Model: "gpt-5.6", ActualCost: 0.4}
	payload, err := buildUsageBillingSettlementPayload("req-1", log, p)
	require.NoError(t, err)
	require.Equal(t, 1, payload.Version)
	require.False(t, payload.Primary.FinalizeUsageLog)
	require.NotNil(t, payload.Overage)
	require.Equal(t, "req-1:trial-overage", payload.Overage.RequestID)
	require.InDelta(t, 0.15, payload.Overage.BalanceCost, 1e-9)
	require.NotNil(t, payload.Trial)
	require.InDelta(t, 0.25, payload.Trial.Amount, 1e-9)
}

func TestPersistUsageLogForBillingRetriesAfterTransientPersistenceError(t *testing.T) {
	repo := &settlementPersistenceRepoFake{errs: []error{errors.New("connection reset"), nil}}
	usageLog := &UsageLog{RequestID: "req-persist-retry", Model: "gpt-5.6"}
	p := &postUsageBillingParams{
		Cost:    &CostBreakdown{ActualCost: 0.2, TotalCost: 0.2},
		User:    &User{ID: 1},
		APIKey:  &APIKey{ID: 2},
		Account: &Account{ID: 3},
	}

	owned, err := persistUsageLogForBilling(context.Background(), repo, usageLog, usageLog.RequestID, p)
	require.NoError(t, err)
	require.True(t, owned)
	require.Equal(t, 2, repo.calls)
}

func TestUsageBillingSettlementService_ReplaysCompositeExactlyOnce(t *testing.T) {
	billing := &settlementBillingRepoFake{}
	outbox := &settlementOutboxFake{tasks: []UsageBillingSettlementTask{{
		ID: 1, UsageLogID: 42, Attempts: 1,
		Payload: UsageBillingSettlementPayload{
			Version: 1,
			Primary: UsageBillingCommand{RequestID: "req-1"},
			Overage: &UsageBillingCommand{RequestID: "req-1:trial-overage", BalanceCost: 0.15},
			Trial:   &UsageBillingTrialPayload{TrialID: 7, UserID: 8, TrialRequestID: "trial-request", RequestID: "req-1", Amount: 0.25, Model: "gpt-5.6", APIKeyID: 9},
		},
	}}}
	welfare := &settlementWelfareFake{}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox, welfare: welfare}

	svc.RunOnce()
	svc.RunOnce()

	require.Len(t, billing.calls, 2)
	require.Equal(t, "req-1", billing.calls[0].RequestID)
	require.Equal(t, "req-1:trial-overage", billing.calls[1].RequestID)
	require.Equal(t, 1, welfare.calls)
	require.Equal(t, []int64{42}, outbox.appliedIDs)
	require.Zero(t, outbox.retries)
}

func TestUsageBillingSettlementService_RetryKeepsPendingOnTransientFailure(t *testing.T) {
	billing := &settlementBillingRepoFake{err: errors.New("temporary db failure")}
	outbox := &settlementOutboxFake{tasks: []UsageBillingSettlementTask{{
		ID: 2, UsageLogID: 43, Attempts: 1,
		Payload: UsageBillingSettlementPayload{Version: 1, Primary: UsageBillingCommand{RequestID: "req-2"}},
	}}}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox}

	svc.RunOnce()

	require.Equal(t, 1, outbox.retries)
	require.Empty(t, outbox.appliedIDs)
}

func TestUsageBillingSettlementService_FinalizesDeferredEffectsAfterTrialRetry(t *testing.T) {
	billing := &settlementBillingRepoFake{results: []*UsageBillingApplyResult{
		{Applied: true}, {Applied: true}, {Applied: false}, {Applied: false},
	}}
	outbox := &settlementOutboxFake{tasks: []UsageBillingSettlementTask{{
		ID: 6, UsageLogID: 47, Attempts: 1,
		Payload: UsageBillingSettlementPayload{
			Version: 1,
			Primary: UsageBillingCommand{RequestID: "req-retry-primary", AccountID: 48, InputTokens: 10},
			Overage: &UsageBillingCommand{RequestID: "req-retry-overage", AccountID: 48},
			Trial:   &UsageBillingTrialPayload{TrialID: 49, UserID: 50, TrialRequestID: "trial-retry", RequestID: "req-retry-primary", Amount: 0.1},
		},
	}}}
	welfare := &settlementWelfareFake{errs: []error{errors.New("trial pool unavailable"), nil}}
	deferred := &DeferredService{}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox, welfare: welfare, deferred: deferred}

	svc.RunOnce()
	_, early := deferred.lastUsedUpdates.Load(int64(48))
	require.False(t, early, "partial composite attempts must not run deferred effects")
	require.Equal(t, 1, outbox.retries)

	outbox.tasks = []UsageBillingSettlementTask{{
		ID: 6, UsageLogID: 47, Attempts: 2,
		Payload: UsageBillingSettlementPayload{
			Version: 1,
			Primary: UsageBillingCommand{RequestID: "req-retry-primary", AccountID: 48, InputTokens: 10},
			Overage: &UsageBillingCommand{RequestID: "req-retry-overage", AccountID: 48},
			Trial:   &UsageBillingTrialPayload{TrialID: 49, UserID: 50, TrialRequestID: "trial-retry", RequestID: "req-retry-primary", Amount: 0.1},
		},
	}}
	svc.RunOnce()

	_, finalized := deferred.lastUsedUpdates.Load(int64(48))
	require.True(t, finalized, "the final claimed attempt must publish deferred effects even when primary was deduplicated")
	require.Equal(t, []int64{47}, outbox.appliedIDs)
	require.Len(t, billing.calls, 4)
	require.Equal(t, 2, welfare.calls)
}

func TestUsageBillingSettlementService_FinalizePostSettlementRefreshesTrialOverageCaches(t *testing.T) {
	cache := &settlementCacheStub{}
	svc := &UsageBillingSettlementService{billingCache: &BillingCacheService{cache: cache}}

	svc.finalizePostSettlement(context.Background(), UsageBillingSettlementPayload{
		Primary: UsageBillingCommand{UserID: 61},
		Trial:   &UsageBillingTrialPayload{UserID: 61},
		Overage: &UsageBillingCommand{UserID: 61, BalanceCost: 0.2},
	}, &UsageBillingApplyResult{BalanceCost: 0.2})
	require.Equal(t, []int64{61}, cache.invalidatedUsers)

	groupID := int64(63)
	svc.finalizePostSettlement(context.Background(), UsageBillingSettlementPayload{
		Primary: UsageBillingCommand{UserID: 62},
		Trial:   &UsageBillingTrialPayload{UserID: 62},
		Overage: &UsageBillingCommand{UserID: 62, GroupID: &groupID, SubscriptionCost: 0.3},
	}, &UsageBillingApplyResult{BalanceCost: 0.2})
	require.Equal(t, []settlementSubscriptionWrite{{userID: 62, groupID: 63, cost: 0.3}}, cache.subscriptionWrites)

	svc.finalizePostSettlement(context.Background(), UsageBillingSettlementPayload{
		Primary: UsageBillingCommand{UserID: 64, GroupID: &groupID, SubscriptionCost: 0.4},
	}, nil)
	require.Equal(t, []settlementSubscriptionWrite{
		{userID: 62, groupID: 63, cost: 0.3},
		{userID: 64, groupID: 63, cost: 0.4},
	}, cache.subscriptionWrites)
}

func TestUsageBillingSettlementService_DoesNotChargeBeforeMissingWelfareIsDetected(t *testing.T) {
	billing := &settlementBillingRepoFake{}
	outbox := &settlementOutboxFake{tasks: []UsageBillingSettlementTask{{
		ID: 5, UsageLogID: 46, Attempts: 1,
		Payload: UsageBillingSettlementPayload{
			Version: 1,
			Primary: UsageBillingCommand{RequestID: "req-trial-no-welfare"},
			Trial:   &UsageBillingTrialPayload{TrialID: 9, UserID: 10, RequestID: "req-trial-no-welfare", Amount: 0.1},
		},
	}}}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox}

	svc.RunOnce()

	require.Empty(t, billing.calls)
	require.Equal(t, 1, outbox.retries)
	require.Empty(t, outbox.appliedIDs)
}

func TestUsageBillingSettlementService_StaleLeaseDoesNotRetryOrFinalizeAgain(t *testing.T) {
	marked := false
	billing := &settlementBillingRepoFake{}
	outbox := &settlementOutboxFake{
		markAttemptResult: &marked,
		tasks: []UsageBillingSettlementTask{{
			ID: 3, UsageLogID: 44, Attempts: 2,
			Payload: UsageBillingSettlementPayload{Version: 1, Primary: UsageBillingCommand{RequestID: "req-stale"}},
		}},
	}
	svc := &UsageBillingSettlementService{billingRepo: billing, outbox: outbox}

	svc.RunOnce()

	require.Len(t, billing.calls, 1, "the claimed local command may finish, but stale finalization must be ignored")
	require.Equal(t, 1, outbox.markAttemptCalls)
	require.Empty(t, outbox.appliedIDs)
	require.Zero(t, outbox.retries)
}

func TestUsageBillingSettlementService_FinalizePostSettlementMergesOverageEffectsAndInvalidatesAuth(t *testing.T) {
	cache := &settlementCacheStub{}
	auth := &settlementAuthCacheInvalidatorStub{}
	svc := &UsageBillingSettlementService{
		billingCache:         &BillingCacheService{cache: cache},
		authCacheInvalidator: auth,
	}
	groupID := int64(12)
	svc.finalizePostSettlement(context.Background(), UsageBillingSettlementPayload{
		Primary: UsageBillingCommand{
			UserID:              10,
			APIKeyID:            11,
			GroupID:             &groupID,
			APIKeyRateLimitCost: 0.4,
			InputTokens:         5,
		},
		Overage: &UsageBillingCommand{
			UserID:              10,
			APIKeyID:            11,
			APIKeyRateLimitCost: 0.2,
		},
	}, &UsageBillingApplyResult{APIKeyQuotaExhausted: true})

	require.Equal(t, []int64{10}, auth.userIDs)
}

func TestMergeUsageBillingApplyResultsPreservesExactlyOneCombinedOutcome(t *testing.T) {
	newBalance := -0.25
	merged := mergeUsageBillingApplyResults(
		&UsageBillingApplyResult{Applied: true, BalanceCost: 0.5, APIKeyQuotaExhausted: true},
		&UsageBillingApplyResult{Applied: false, BalanceCost: 0.25, PrepaidBalanceCost: 0.1, NewBalance: &newBalance},
	)
	require.True(t, merged.Applied)
	require.True(t, merged.APIKeyQuotaExhausted)
	require.InDelta(t, 0.75, merged.BalanceCost, 1e-9)
	require.InDelta(t, 0.1, merged.PrepaidBalanceCost, 1e-9)
	require.NotNil(t, merged.NewBalance)
	require.InDelta(t, -0.25, *merged.NewBalance, 1e-9)
}
