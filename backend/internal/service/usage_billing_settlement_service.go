package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	usageBillingSettlementWorkerName  = "usage_billing_settlement_worker"
	usageBillingSettlementMaxAttempts = 12
	usageBillingSettlementInterval    = 5 * time.Second
)

// UsageBillingSettlementService replays only local, durable billing commands.
// It never contacts an upstream provider.
type UsageBillingSettlementService struct {
	billingRepo          UsageBillingRepository
	outbox               UsageBillingSettlementRepository
	userRepo             UserRepository
	accountRepo          AccountRepository
	welfare              newUserTrialConsumer
	billingCache         *BillingCacheService
	deferred             *DeferredService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	affiliateService     *AffiliateService
	balanceNotifyService *BalanceNotifyService
	timingWheel          *TimingWheelService
	running              int32
	stopped              int32
	activeMu             sync.Mutex
	activeCancel         context.CancelFunc
	scheduleFn           func(string, time.Duration, func())
	cancelFn             func(string)
	startOnce            sync.Once
	stopOnce             sync.Once
}

type newUserTrialConsumer interface {
	ConsumeNewUserTrial(ctx context.Context, session *NewUserTrialSession, requestID string, amount float64, model string, apiKeyID int64) error
}

// usageBillingSettlementAttemptFinalizer lets a worker finalize only the
// exact lease/attempt it claimed. The optional interface keeps existing test
// doubles and external repository adapters source compatible with the base
// settlement repository contract.
type usageBillingSettlementAttemptFinalizer interface {
	MarkAppliedAttempt(ctx context.Context, taskID int64, attempts int, usageLogID int64) (bool, error)
}

func NewUsageBillingSettlementService(billingRepo UsageBillingRepository, usageLogRepo UsageLogRepository, timingWheel *TimingWheelService, welfare newUserTrialConsumer, billingCache *BillingCacheService, deferred *DeferredService) *UsageBillingSettlementService {
	outbox, _ := usageLogRepo.(UsageBillingSettlementRepository)
	return &UsageBillingSettlementService{billingRepo: billingRepo, outbox: outbox, timingWheel: timingWheel, welfare: welfare, billingCache: billingCache, deferred: deferred}
}

func ProvideUsageBillingSettlementService(
	billingRepo UsageBillingRepository,
	usageLogRepo UsageLogRepository,
	timingWheel *TimingWheelService,
	welfare newUserTrialConsumer,
	billingCache *BillingCacheService,
	deferred *DeferredService,
	userRepo UserRepository,
	accountRepo AccountRepository,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	affiliateService *AffiliateService,
	balanceNotifyService *BalanceNotifyService,
) *UsageBillingSettlementService {
	svc := NewUsageBillingSettlementService(billingRepo, usageLogRepo, timingWheel, welfare, billingCache, deferred)
	svc.userRepo = userRepo
	svc.accountRepo = accountRepo
	svc.authCacheInvalidator = authCacheInvalidator
	svc.affiliateService = affiliateService
	svc.balanceNotifyService = balanceNotifyService
	svc.Start()
	return svc
}

func (s *UsageBillingSettlementService) Start() {
	if s == nil || s.outbox == nil || s.billingRepo == nil || s.timingWheel == nil || s.isStopped() {
		logger.LegacyPrintf("service.usage_billing_settlement", "[UsageBillingSettlement] not started (missing durable dependencies)")
		return
	}
	s.startOnce.Do(func() {
		if s.isStopped() {
			return
		}
		s.scheduleNextRun()
		go s.RunOnce()
	})
}

func (s *UsageBillingSettlementService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		atomic.StoreInt32(&s.stopped, 1)
		s.activeMu.Lock()
		if s.activeCancel != nil {
			s.activeCancel()
		}
		s.cancelScheduledRun()
		s.activeMu.Unlock()
	})
}

// scheduleNextRun owns recurrence so Stop can prevent a callback that is
// already running from scheduling a successor after the timer was cancelled.
func (s *UsageBillingSettlementService) scheduleNextRun() {
	if s == nil || s.timingWheel == nil {
		return
	}
	s.activeMu.Lock()
	defer s.activeMu.Unlock()
	if s.isStopped() {
		return
	}
	s.scheduleRun(usageBillingSettlementWorkerName, usageBillingSettlementInterval, func() {
		if s.isStopped() {
			return
		}
		s.RunOnce()
		if s.isStopped() {
			return
		}
		s.scheduleNextRun()
	})
}

func (s *UsageBillingSettlementService) scheduleRun(name string, delay time.Duration, fn func()) {
	if s.scheduleFn != nil {
		s.scheduleFn(name, delay, fn)
		return
	}
	s.timingWheel.Schedule(name, delay, fn)
}

func (s *UsageBillingSettlementService) cancelScheduledRun() {
	if s.cancelFn != nil {
		s.cancelFn(usageBillingSettlementWorkerName)
		return
	}
	if s.timingWheel != nil {
		s.timingWheel.Cancel(usageBillingSettlementWorkerName)
	}
}

func (s *UsageBillingSettlementService) RunOnce() {
	if s == nil || s.outbox == nil || s.billingRepo == nil || s.isStopped() || !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	s.activeMu.Lock()
	if s.isStopped() {
		s.activeMu.Unlock()
		timeoutCancel()
		cancel()
		atomic.StoreInt32(&s.running, 0)
		return
	}
	s.activeCancel = cancel
	s.activeMu.Unlock()
	defer func() {
		s.activeMu.Lock()
		s.activeCancel = nil
		s.activeMu.Unlock()
		timeoutCancel()
		cancel()
		atomic.StoreInt32(&s.running, 0)
	}()
	tasks, err := s.outbox.ClaimPending(ctx, 32, time.Minute)
	if err != nil {
		if s.isStopped() || errors.Is(err, context.Canceled) {
			return
		}
		logger.LegacyPrintf("service.usage_billing_settlement", "claim failed: %v", err)
		return
	}
	for _, task := range tasks {
		if s.isStopped() {
			return
		}
		payload := task.Payload
		if payload.Primary.RequestID == "" {
			payload = UsageBillingSettlementPayload{Version: 1, Primary: task.Command}
		}
		if payload.Trial != nil && s.welfare == nil {
			err := ErrWelfareNewUserTrialUnavailable
			terminal := task.Attempts >= usageBillingSettlementMaxAttempts
			delay := usageBillingSettlementRetryDelay(task.Attempts)
			if markErr := s.outbox.MarkRetry(ctx, task.ID, task.Attempts, err, time.Now().Add(delay), terminal); markErr != nil {
				logger.LegacyPrintf("service.usage_billing_settlement", "mark retry task=%d err=%v: %v", task.ID, err, markErr)
			}
			continue
		}
		primary := payload.Primary
		primary.UsageLogID = task.UsageLogID
		primary.FinalizeUsageLog = false
		primaryResult, err := s.billingRepo.Apply(ctx, &primary)
		combinedResult := primaryResult
		if err == nil && payload.Overage != nil {
			overage := *payload.Overage
			overage.UsageLogID = task.UsageLogID
			overage.FinalizeUsageLog = false
			overageResult, applyErr := s.billingRepo.Apply(ctx, &overage)
			err = applyErr
			combinedResult = mergeUsageBillingApplyResults(primaryResult, overageResult)
		}
		if err == nil && payload.Trial != nil {
			if s.welfare == nil {
				err = fmt.Errorf("welfare trial settlement service unavailable")
			} else {
				t := payload.Trial
				session := &NewUserTrialSession{TrialID: t.TrialID, UserID: t.UserID, RequestID: t.TrialRequestID}
				err = s.welfare.ConsumeNewUserTrial(ctx, session, t.RequestID, t.Amount, t.Model, t.APIKeyID)
			}
		}
		if err == nil && (payload.Overage != nil || payload.Trial != nil) {
			if reconciler, ok := s.billingRepo.(UsageBillingLedgerReconciler); ok {
				err = reconciler.ReconcileUsageBillingEntry(ctx, &payload)
			}
		}
		if s.isStopped() {
			return
		}
		if err == nil {
			marked := true
			if finalizer, ok := s.outbox.(usageBillingSettlementAttemptFinalizer); ok {
				marked, err = finalizer.MarkAppliedAttempt(ctx, task.ID, task.Attempts, task.UsageLogID)
			} else {
				err = s.outbox.MarkApplied(ctx, task.UsageLogID)
			}
			if err == nil {
				// A stale lease may have been completed by another worker. In
				// that case the current worker must not run post-settlement
				// side effects a second time or move the row back to pending.
				if !marked {
					continue
				}
				if !s.isStopped() {
					s.finalizePostSettlement(ctx, payload, combinedResult)
				}
				continue
			}
		}
		if s.isStopped() {
			return
		}
		terminal := task.Attempts >= usageBillingSettlementMaxAttempts ||
			errors.Is(err, ErrUsageBillingRequestConflict) ||
			errors.Is(err, ErrUsageBillingCommandInvalid) ||
			errors.Is(err, ErrUsageBillingRequestIDRequired)
		delay := usageBillingSettlementRetryDelay(task.Attempts)
		if markErr := s.outbox.MarkRetry(ctx, task.ID, task.Attempts, err, time.Now().Add(delay), terminal); markErr != nil {
			logger.LegacyPrintf("service.usage_billing_settlement", "mark retry task=%d err=%v: %v", task.ID, err, markErr)
		}
	}
}

func (s *UsageBillingSettlementService) isStopped() bool {
	return s == nil || atomic.LoadInt32(&s.stopped) != 0
}

func mergeUsageBillingApplyResults(primary, overage *UsageBillingApplyResult) *UsageBillingApplyResult {
	if primary == nil && overage == nil {
		return nil
	}
	merged := &UsageBillingApplyResult{}
	for _, result := range []*UsageBillingApplyResult{primary, overage} {
		if result == nil {
			continue
		}
		merged.Applied = merged.Applied || result.Applied
		merged.APIKeyQuotaExhausted = merged.APIKeyQuotaExhausted || result.APIKeyQuotaExhausted
		merged.VoucherCost += result.VoucherCost
		merged.BalanceCost += result.BalanceCost
		merged.PrepaidBalanceCost += result.PrepaidBalanceCost
		if result.NewBalance != nil {
			balance := *result.NewBalance
			merged.NewBalance = &balance
		}
		if result.QuotaState != nil {
			merged.QuotaState = result.QuotaState
		}
	}
	return merged
}

func (s *UsageBillingSettlementService) finalizePostSettlement(ctx context.Context, payload UsageBillingSettlementPayload, result *UsageBillingApplyResult) {
	if s == nil {
		return
	}
	primary := payload.Primary
	balanceCost := primary.BalanceCost + primary.PrepaidBalanceCost
	if payload.Trial != nil {
		// Trial primary never deducts from the user wallet; only its overage can
		// invalidate wallet or subscription cache state.
		balanceCost = 0
	}
	if payload.Overage != nil {
		balanceCost += payload.Overage.BalanceCost + payload.Overage.PrepaidBalanceCost
	}
	if s.billingCache != nil {
		if balanceCost > 0 {
			if err := s.billingCache.InvalidateUserBalance(ctx, primary.UserID); err != nil {
				logger.LegacyPrintf("service.usage_billing_settlement", "invalidate balance cache failed: %v", err)
			}
		}
		apiKeyRateLimitCost := primary.APIKeyRateLimitCost
		if payload.Overage != nil {
			apiKeyRateLimitCost += payload.Overage.APIKeyRateLimitCost
		}
		if apiKeyRateLimitCost > 0 {
			s.billingCache.QueueUpdateAPIKeyRateLimitUsage(primary.APIKeyID, apiKeyRateLimitCost)
		}
		if primary.SubscriptionCost > 0 && primary.GroupID != nil {
			s.billingCache.QueueUpdateSubscriptionUsage(primary.UserID, *primary.GroupID, primary.SubscriptionCost)
		}
		if tokens := primary.InputTokens + primary.OutputTokens + primary.CacheCreationTokens + primary.CacheReadTokens; tokens > 0 {
			s.billingCache.RecordMembershipTokenUsage(ctx, primary.UserID, tokens)
		}
		if overage := payload.Overage; overage != nil && overage.SubscriptionCost > 0 && overage.GroupID != nil {
			s.billingCache.QueueUpdateSubscriptionUsage(overage.UserID, *overage.GroupID, overage.SubscriptionCost)
		}
	}
	if result != nil && result.APIKeyQuotaExhausted && s.authCacheInvalidator != nil && primary.UserID > 0 {
		// The durable payload intentionally omits the plaintext API key. Invalidating
		// the user's auth snapshots is broader than key-level invalidation but keeps
		// quota/status changes from remaining stale after an asynchronous retry.
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, primary.UserID)
	}
	if s.affiliateService != nil && primary.UserID > 0 {
		s.affiliateService.NotifyInviteeFirstAPIRewardIfEligible(ctx, primary.UserID)
	}
	s.notifyPostSettlement(ctx, payload, result)
	if s.deferred != nil && primary.AccountID > 0 {
		s.deferred.ScheduleLastUsedUpdate(primary.AccountID)
	}
}

func (s *UsageBillingSettlementService) notifyPostSettlement(ctx context.Context, payload UsageBillingSettlementPayload, result *UsageBillingApplyResult) {
	if s == nil || s.balanceNotifyService == nil || s.userRepo == nil || s.accountRepo == nil {
		return
	}
	primary := payload.Primary
	if primary.UserID <= 0 || primary.AccountID <= 0 {
		return
	}
	user, err := s.userRepo.GetByID(ctx, primary.UserID)
	if err != nil || user == nil {
		return
	}
	account, err := s.accountRepo.GetByID(ctx, primary.AccountID)
	if err != nil || account == nil {
		return
	}
	totalCost := primary.BalanceCost + primary.PrepaidBalanceCost + primary.SubscriptionCost
	params := &postUsageBillingParams{
		Cost: &CostBreakdown{
			ActualCost: totalCost,
			TotalCost:  primary.AccountQuotaCost,
		},
		User:                  user,
		Account:               account,
		IsSubscriptionBill:    primary.SubscriptionCost > 0,
		PrepaidBalanceCost:    primary.PrepaidBalanceCost,
		AccountRateMultiplier: 1,
	}
	if payload.Trial != nil {
		params.NewUserTrial = &NewUserTrialSession{RequestID: payload.Trial.TrialRequestID}
	}
	if result != nil && result.BalanceCost > 0 {
		notifyBalanceLow(params, &billingDeps{balanceNotifyService: s.balanceNotifyService}, result)
	}
	if primary.AccountQuotaCost > 0 {
		notifyAccountQuota(params, &billingDeps{balanceNotifyService: s.balanceNotifyService}, result)
	}
}

func usageBillingSettlementRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	seconds := 1 << min(attempt, 10)
	if seconds > 3600 {
		seconds = 3600
	}
	return time.Duration(seconds) * time.Second
}

func (s *UsageBillingSettlementService) String() string {
	return fmt.Sprintf("UsageBillingSettlementService{started=%t}", s != nil && s.outbox != nil)
}
