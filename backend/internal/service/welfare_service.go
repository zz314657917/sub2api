package service

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const (
	welfareRewardScale          = int64(100000000)
	welfareDailyRewardStepScale = int64(10)

	welfareReasonAvailable          = "available"
	welfareReasonDisabled           = "disabled"
	welfareReasonAlreadyClaimed     = "already_claimed"
	welfareReasonNotReached         = "not_reached"
	welfareReasonZeroReward         = "zero_reward"
	welfareReasonAlreadyChecked     = "already_checked"
	welfareReasonNotConfigured      = "not_configured"
	welfareReasonExpired            = "expired"
	welfareReasonInProgress         = "in_progress"
	welfareReasonExhausted          = "exhausted"
	welfareReasonDailyLimit         = "daily_limit"
	welfareReasonRegistrationTooNew = "registration_too_new"
	welfareMilestoneDay7            = 7
	welfareMilestoneDay14           = 14
	welfareMilestoneDay21           = 21
	welfareMilestoneDay28           = 28
	welfareFirstRechargeBonusAudit  = "FIRST_RECHARGE_BONUS_GRANTED"

	defaultNewUserTrialQuotaAmount            = 0.1
	defaultNewUserTrialDailySiteQuotaAmount   = 5.0
	defaultNewUserTrialDailyIPActivationLimit = 3
	defaultDailyCheckinMinAccountAgeHours     = 24
	defaultFirstRechargeBonusAmount           = 5.0
)

var (
	ErrWelfareDisabled                       = infraerrors.Forbidden("WELFARE_DISABLED", "welfare system is disabled")
	ErrWelfareDailyCheckinDisabled           = infraerrors.Forbidden("WELFARE_DAILY_CHECKIN_DISABLED", "daily check-in welfare is disabled")
	ErrWelfareDailyCheckinNotClaimable       = infraerrors.Forbidden("WELFARE_DAILY_CHECKIN_NOT_CLAIMABLE", "daily check-in reward is not claimable")
	ErrWelfareDailyCheckinNotFound           = infraerrors.NotFound("WELFARE_DAILY_CHECKIN_NOT_FOUND", "daily check-in record not found")
	ErrWelfareDailyCheckinAlreadyClaimed     = infraerrors.Conflict("WELFARE_DAILY_CHECKIN_ALREADY_CLAIMED", "daily check-in reward already claimed")
	ErrWelfareCheckinMilestoneNotFound       = infraerrors.NotFound("WELFARE_CHECKIN_MILESTONE_NOT_FOUND", "daily check-in milestone claim not found")
	ErrWelfareCheckinMilestoneAlreadyClaimed = infraerrors.Conflict("WELFARE_CHECKIN_MILESTONE_ALREADY_CLAIMED", "daily check-in milestone already claimed")
	ErrWelfareCheckinMilestoneNotClaimable   = infraerrors.Forbidden("WELFARE_CHECKIN_MILESTONE_NOT_CLAIMABLE", "daily check-in milestone is not claimable")
	ErrWelfareDailyCheckinUnavailable        = infraerrors.ServiceUnavailable("WELFARE_DAILY_CHECKIN_UNAVAILABLE", "daily check-in welfare service is unavailable")
	ErrWelfareNewUserTrialDisabled           = infraerrors.Forbidden("WELFARE_NEW_USER_TRIAL_DISABLED", "new user trial is disabled")
	ErrWelfareNewUserTrialUnavailable        = infraerrors.ServiceUnavailable("WELFARE_NEW_USER_TRIAL_UNAVAILABLE", "new user trial service is unavailable")
	ErrWelfareNewUserTrialNotAvailable       = infraerrors.Forbidden("WELFARE_NEW_USER_TRIAL_NOT_AVAILABLE", "new user trial is not available")
	ErrWelfareNewUserTrialAlreadyInProgress  = infraerrors.Conflict("WELFARE_NEW_USER_TRIAL_IN_PROGRESS", "new user trial request is already in progress")
	ErrWelfareNewUserTrialExhausted          = infraerrors.Forbidden("WELFARE_NEW_USER_TRIAL_EXHAUSTED", "new user trial quota is exhausted")
	ErrWelfareNewUserTrialDailyLimitExceeded = infraerrors.TooManyRequests("WELFARE_NEW_USER_TRIAL_DAILY_LIMIT", "new user trial daily limit exceeded")
	ErrWelfareNewUserTrialNotFound           = infraerrors.NotFound("WELFARE_NEW_USER_TRIAL_NOT_FOUND", "new user trial not found")
	ErrWelfareNewUserTrialRewardClaimed      = infraerrors.Conflict("WELFARE_NEW_USER_TRIAL_REWARD_CLAIMED", "new user trial reward already claimed")
	ErrFirstRechargeBonusRefundUnsupported   = infraerrors.Forbidden("FIRST_RECHARGE_BONUS_REFUND_UNSUPPORTED", "orders with first recharge bonus cannot be refunded automatically")
)

type WelfareService struct {
	repo                    WelfareRepository
	userRepo                UserRepository
	redeemRepo              RedeemCodeRepository
	settingRepo             SettingRepository
	entClient               *dbent.Client
	authCacheInvalidator    APIKeyAuthCacheInvalidator
	billingCacheInvalidator welfareBalanceCacheInvalidator
	now                     func() time.Time
	systemTicketSvc         *SystemTicketService
}

type welfareBalanceCacheInvalidator interface {
	InvalidateUserBalance(ctx context.Context, userID int64) error
}

type welfareBalanceGrantRepository interface {
	AddBalance(ctx context.Context, id int64, amount float64) error
}

type welfareRedeemCodeLookupRepository interface {
	GetByCode(ctx context.Context, code string) (*RedeemCode, error)
}

type WelfareRepository interface {
	GetDailyCheckin(ctx context.Context, checkinDate string, userID int64) (*WelfareDailyCheckinRecord, error)
	ListDailyCheckins(ctx context.Context, userID int64, month string) ([]WelfareDailyCheckinRecord, error)
	CreateDailyCheckin(ctx context.Context, record *WelfareDailyCheckinRecord) error
	AttachDailyCheckinRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error
	GetDailyCheckinMilestoneClaim(ctx context.Context, month string, milestoneDay int, userID int64) (*WelfareDailyCheckinMilestoneClaim, error)
	ListDailyCheckinMilestoneClaims(ctx context.Context, month string, userID int64) ([]WelfareDailyCheckinMilestoneClaim, error)
	CreateDailyCheckinMilestoneClaim(ctx context.Context, claim *WelfareDailyCheckinMilestoneClaim) error
	AttachDailyCheckinMilestoneRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error
	GetNewUserTrial(ctx context.Context, userID int64) (*WelfareNewUserTrial, error)
	BeginNewUserTrial(ctx context.Context, userID int64, quotaAmount float64, clientIP, requestID string, dayStart time.Time, ipActivationLimit int) (*WelfareNewUserTrial, error)
	CancelNewUserTrial(ctx context.Context, trialID int64, requestID string) error
	ConsumeNewUserTrial(ctx context.Context, input WelfareNewUserTrialConsumeInput) (*WelfareNewUserTrial, bool, error)
	SumNewUserTrialUsageSince(ctx context.Context, since time.Time) (float64, error)
	CountNewUserTrialActivationsByIPSince(ctx context.Context, clientIP string, since time.Time) (int, error)
	FirstSuccessfulUsageAt(ctx context.Context, userID int64) (*time.Time, error)
}

type WelfareDailyCheckinRecord struct {
	ID           int64     `json:"id"`
	CheckinDate  string    `json:"checkin_date"`
	RewardMonth  string    `json:"reward_month"`
	UserID       int64     `json:"user_id"`
	Amount       float64   `json:"amount"`
	RedeemCodeID *int64    `json:"redeem_code_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type WelfareDailyCheckinMilestoneClaim struct {
	ID           int64     `json:"id"`
	RewardMonth  string    `json:"reward_month"`
	MilestoneDay int       `json:"milestone_day"`
	UserID       int64     `json:"user_id"`
	Amount       float64   `json:"amount"`
	RedeemCodeID *int64    `json:"redeem_code_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type WelfareNewUserTrial struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	QuotaAmount    float64    `json:"quota_amount"`
	QuotaUsed      float64    `json:"quota_used"`
	Status         string     `json:"status"`
	ActivatedIP    string     `json:"activated_ip,omitempty"`
	FirstStartedAt *time.Time `json:"first_started_at,omitempty"`
	FirstSuccessAt *time.Time `json:"first_success_at,omitempty"`
	LastRequestID  string     `json:"last_request_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type WelfareNewUserTrialConsumeInput struct {
	TrialID        int64
	UserID         int64
	TrialRequestID string
	RequestID      string
	Amount         float64
	Model          string
	APIKeyID       int64
}

type WelfareOverview struct {
	Enabled      bool                     `json:"enabled"`
	Modules      WelfareModules           `json:"modules"`
	DailyCheckin *WelfareDailyCheckinView `json:"daily_checkin,omitempty"`
	NewUserTrial *WelfareNewUserTrialView `json:"new_user_trial,omitempty"`
	Recharge     *WelfareRechargeView     `json:"recharge,omitempty"`
}

type WelfareModules struct {
	DailyCheckin bool `json:"daily_checkin"`
	NewUserTrial bool `json:"new_user_trial"`
	Recharge     bool `json:"recharge"`
	VIP          bool `json:"vip"`
}

type WelfareDailyCheckinView struct {
	Enabled            bool                           `json:"enabled"`
	Today              string                         `json:"today"`
	RewardMonth        string                         `json:"reward_month"`
	CheckedToday       bool                           `json:"checked_today"`
	TodayRewardAmount  float64                        `json:"today_reward_amount"`
	RewardMin          float64                        `json:"reward_min"`
	RewardMax          float64                        `json:"reward_max"`
	CurrentStreakDays  int                            `json:"current_streak_days"`
	MonthCheckinDays   int                            `json:"month_checkin_days"`
	CheckinDates       []string                       `json:"checkin_dates"`
	MilestoneEnabled   bool                           `json:"milestone_enabled"`
	Milestones         []WelfareDailyCheckinMilestone `json:"milestones"`
	CanClaimToday      bool                           `json:"can_claim_today"`
	Reason             string                         `json:"reason"`
	CanClaimAfter      string                         `json:"can_claim_after,omitempty"`
	SettlementTimezone string                         `json:"settlement_timezone"`
}

type WelfareDailyCheckinMilestone struct {
	Day          int     `json:"day"`
	Amount       float64 `json:"amount"`
	Claimed      bool    `json:"claimed"`
	Claimable    bool    `json:"claimable"`
	Reason       string  `json:"reason"`
	ClaimedAt    string  `json:"claimed_at,omitempty"`
	RedeemCodeID *int64  `json:"redeem_code_id,omitempty"`
}

type WelfareNewUserTrialView struct {
	Enabled                bool    `json:"enabled"`
	QuotaAmount            float64 `json:"quota_amount"`
	QuotaUsed              float64 `json:"quota_used"`
	RemainingQuota         float64 `json:"remaining_quota"`
	SuccessRewardAmount    float64 `json:"success_reward_amount"`
	SuccessRewardClaimable bool    `json:"success_reward_claimable"`
	SuccessRewardClaimed   bool    `json:"success_reward_claimed"`
	SuccessRewardReason    string  `json:"success_reward_reason"`
	Status                 string  `json:"status"`
	CanUse                 bool    `json:"can_use"`
	Reason                 string  `json:"reason"`
	FirstStartedAt         string  `json:"first_started_at,omitempty"`
	FirstSuccessAt         string  `json:"first_success_at,omitempty"`
	SuccessRewardClaimedAt string  `json:"success_reward_claimed_at,omitempty"`
}

type WelfareRechargeView struct {
	Enabled                bool    `json:"enabled"`
	FirstBonusAmount       float64 `json:"first_bonus_amount"`
	FirstBonusClaimed      bool    `json:"first_bonus_claimed"`
	FirstBonusClaimedAt    string  `json:"first_bonus_claimed_at,omitempty"`
	ValidDays              int     `json:"valid_days"`
	ExpiresAt              string  `json:"expires_at,omitempty"`
	CanStackMonthlyBonus   bool    `json:"can_stack_monthly_bonus"`
	MonthlyBonusMayBlock   bool    `json:"monthly_bonus_may_block"`
	FirstRechargeCompleted bool    `json:"first_recharge_completed"`
	Reason                 string  `json:"reason"`
}

type WelfareDailyCheckinClaimResult struct {
	DailyCheckin  *WelfareDailyCheckinView `json:"daily_checkin"`
	ClaimedAmount float64                  `json:"claimed_amount"`
}

type WelfareNewUserTrialRewardClaimResult struct {
	NewUserTrial  *WelfareNewUserTrialView `json:"new_user_trial"`
	ClaimedAmount float64                  `json:"claimed_amount"`
}

type WelfareDailyCheckinMilestoneClaimResult struct {
	DailyCheckin  *WelfareDailyCheckinView     `json:"daily_checkin"`
	Milestone     WelfareDailyCheckinMilestone `json:"milestone"`
	ClaimedAmount float64                      `json:"claimed_amount"`
}

type welfareSettings struct {
	Enabled                            bool
	DailyCheckinEnabled                bool
	NewUserTrialEnabled                bool
	RechargeEnabled                    bool
	VIPEnabled                         bool
	RewardMin                          float64
	RewardMax                          float64
	DailyCheckinMinAccountAgeHours     int
	MilestoneEnabled                   bool
	MilestoneAmounts                   map[int]float64
	NewUserTrialQuotaAmount            float64
	NewUserTrialSuccessRewardAmount    float64
	NewUserTrialSuccessRewardEnabledAt *time.Time
	NewUserTrialDailySiteQuotaAmount   float64
	NewUserTrialDailyIPActivationLimit int
	FirstRechargeBonusAmount           float64
	FirstRechargeBonusValidDays        int
	FirstRechargeBonusStackMonthly     bool
}

func NewWelfareService(
	repo WelfareRepository,
	userRepo UserRepository,
	redeemRepo RedeemCodeRepository,
	settingRepo SettingRepository,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	billingCacheService *BillingCacheService,
) *WelfareService {
	var billingCacheInvalidator welfareBalanceCacheInvalidator
	if billingCacheService != nil {
		billingCacheInvalidator = billingCacheService
	}
	return &WelfareService{
		repo:                    repo,
		userRepo:                userRepo,
		redeemRepo:              redeemRepo,
		settingRepo:             settingRepo,
		entClient:               entClient,
		authCacheInvalidator:    authCacheInvalidator,
		billingCacheInvalidator: billingCacheInvalidator,
		now:                     timezone.Now,
	}
}

func (s *WelfareService) SetSystemTicketService(systemTicketSvc *SystemTicketService) {
	if s != nil {
		s.systemTicketSvc = systemTicketSvc
	}
}

func (s *WelfareService) GetOverview(ctx context.Context, userID int64) (*WelfareOverview, error) {
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	overview := &WelfareOverview{
		Enabled: settings.Enabled,
		Modules: WelfareModules{
			DailyCheckin: settings.Enabled && settings.DailyCheckinEnabled,
			NewUserTrial: settings.Enabled && settings.NewUserTrialEnabled,
			Recharge:     settings.Enabled && settings.RechargeEnabled,
			VIP:          settings.Enabled && settings.VIPEnabled,
		},
	}
	daily, err := s.buildDailyCheckinView(ctx, userID, s.nowTime(), settings)
	if err != nil {
		return nil, err
	}
	overview.DailyCheckin = daily
	trial, err := s.buildNewUserTrialView(ctx, userID, settings)
	if err != nil {
		return nil, err
	}
	overview.NewUserTrial = trial
	recharge, err := s.buildRechargeView(ctx, userID, settings)
	if err != nil {
		return nil, err
	}
	overview.Recharge = recharge
	return overview, nil
}

func (s *WelfareService) GetDailyCheckin(ctx context.Context, userID int64) (*WelfareDailyCheckinView, error) {
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildDailyCheckinView(ctx, userID, s.nowTime(), settings)
}

func (s *WelfareService) BeginNewUserTrial(ctx context.Context, userID int64, clientIP string) (*NewUserTrialSession, error) {
	if s == nil || s.repo == nil {
		return nil, ErrWelfareNewUserTrialUnavailable
	}
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled || !settings.NewUserTrialEnabled {
		return nil, ErrWelfareNewUserTrialDisabled.WithMetadata(map[string]string{"reason": welfareReasonDisabled})
	}
	if settings.NewUserTrialQuotaAmount <= 0 {
		return nil, ErrWelfareNewUserTrialNotAvailable.WithMetadata(map[string]string{"reason": welfareReasonZeroReward})
	}

	dayStart := timezone.StartOfDay(s.nowTime())
	if settings.NewUserTrialDailySiteQuotaAmount > 0 {
		used, err := s.repo.SumNewUserTrialUsageSince(ctx, dayStart)
		if err != nil {
			return nil, err
		}
		if used >= settings.NewUserTrialDailySiteQuotaAmount {
			return nil, ErrWelfareNewUserTrialDailyLimitExceeded.WithMetadata(map[string]string{"reason": welfareReasonDailyLimit})
		}
	}
	if settings.NewUserTrialDailyIPActivationLimit > 0 && strings.TrimSpace(clientIP) != "" {
		count, err := s.repo.CountNewUserTrialActivationsByIPSince(ctx, clientIP, dayStart)
		if err != nil {
			return nil, err
		}
		existing, existingErr := s.repo.GetNewUserTrial(ctx, userID)
		if existingErr != nil && !errors.Is(existingErr, ErrWelfareNewUserTrialNotFound) {
			return nil, existingErr
		}
		if count >= settings.NewUserTrialDailyIPActivationLimit && existing == nil {
			return nil, ErrWelfareNewUserTrialDailyLimitExceeded.WithMetadata(map[string]string{"reason": welfareReasonDailyLimit})
		}
	}

	requestID := "trial:" + generateRequestID()
	trial, err := s.repo.BeginNewUserTrial(ctx, userID, settings.NewUserTrialQuotaAmount, clientIP, requestID, dayStart, settings.NewUserTrialDailyIPActivationLimit)
	if err != nil {
		if errors.Is(err, ErrWelfareNewUserTrialDailyLimitExceeded) {
			return nil, ErrWelfareNewUserTrialDailyLimitExceeded.WithMetadata(map[string]string{"reason": welfareReasonDailyLimit})
		}
		return nil, err
	}
	remaining := normalizeTrialRemaining(trial.QuotaAmount, trial.QuotaUsed)
	if remaining <= 0 {
		return nil, ErrWelfareNewUserTrialExhausted.WithMetadata(map[string]string{"reason": welfareReasonExhausted})
	}
	return &NewUserTrialSession{
		TrialID:   trial.ID,
		UserID:    userID,
		RequestID: requestID,
		QuotaLeft: remaining,
	}, nil
}

func (s *WelfareService) CancelNewUserTrial(ctx context.Context, session *NewUserTrialSession) {
	if s == nil || s.repo == nil || session == nil || session.TrialID <= 0 || session.RequestID == "" {
		return
	}
	if err := s.repo.CancelNewUserTrial(ctx, session.TrialID, session.RequestID); err != nil {
		logger.LegacyPrintf("service.welfare", "cancel new user trial failed: trial=%d request_id=%s err=%v", session.TrialID, session.RequestID, err)
	}
}

func (s *WelfareService) ConsumeNewUserTrial(ctx context.Context, session *NewUserTrialSession, requestID string, amount float64, model string, apiKeyID int64) error {
	if s == nil || s.repo == nil || session == nil {
		return nil
	}
	if session.TrialID <= 0 || session.UserID <= 0 {
		return nil
	}
	if requestID = strings.TrimSpace(requestID); requestID == "" {
		requestID = session.RequestID
	}
	if requestID == "" || amount <= 0 {
		return nil
	}
	trial, applied, err := s.repo.ConsumeNewUserTrial(ctx, WelfareNewUserTrialConsumeInput{
		TrialID:        session.TrialID,
		UserID:         session.UserID,
		TrialRequestID: session.RequestID,
		RequestID:      requestID,
		Amount:         amount,
		Model:          model,
		APIKeyID:       apiKeyID,
	})
	if err == nil && applied {
		s.notifyNewUserTrialSuccessRewardBestEffort(ctx, trial, model, apiKeyID)
	}
	return err
}

func (s *WelfareService) notifyNewUserTrialSuccessRewardBestEffort(ctx context.Context, trial *WelfareNewUserTrial, model string, apiKeyID int64) {
	if s == nil || s.systemTicketSvc == nil || trial == nil || trial.UserID <= 0 || trial.FirstSuccessAt == nil {
		return
	}
	settings, err := s.getSettings(ctx)
	if err != nil {
		logger.LegacyPrintf("service.welfare", "load settings for system ticket notification failed: user_id=%d err=%v", trial.UserID, err)
		return
	}
	if !settings.Enabled || !settings.NewUserTrialEnabled || settings.NewUserTrialSuccessRewardAmount <= 0 {
		return
	}
	qualified, err := s.hasNewUserTrialRewardQualifyingUsage(ctx, trial.UserID, trial, settings.NewUserTrialSuccessRewardEnabledAt)
	if err != nil || !qualified {
		if err != nil {
			logger.LegacyPrintf("service.welfare", "check trial reward notification eligibility failed: user_id=%d err=%v", trial.UserID, err)
		}
		return
	}
	event := NewWelfareFirstAPIUnclaimedSystemTicketNotification(trial.UserID, trial.ID, settings.NewUserTrialSuccessRewardAmount, model, apiKeyID, *trial.FirstSuccessAt)
	s.systemTicketSvc.NotifyEventBestEffort(ctx, "service.welfare", trial.UserID, event)
}

func (s *WelfareService) ClaimNewUserTrialSuccessReward(ctx context.Context, userID int64) (*WelfareNewUserTrialRewardClaimResult, error) {
	if s == nil || s.repo == nil || s.userRepo == nil || s.redeemRepo == nil {
		return nil, ErrWelfareNewUserTrialUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrWelfareNewUserTrialUnavailable
	}
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled || !settings.NewUserTrialEnabled {
		return nil, ErrWelfareNewUserTrialDisabled.WithMetadata(map[string]string{"reason": welfareReasonDisabled})
	}
	if settings.NewUserTrialSuccessRewardAmount <= 0 {
		return nil, ErrWelfareNewUserTrialNotAvailable.WithMetadata(map[string]string{"reason": welfareReasonZeroReward})
	}
	trial, err := s.repo.GetNewUserTrial(ctx, userID)
	if err != nil {
		if !errors.Is(err, ErrWelfareNewUserTrialNotFound) {
			return nil, err
		}
		trial = nil
	}
	qualified, err := s.hasNewUserTrialRewardQualifyingUsage(ctx, userID, trial, settings.NewUserTrialSuccessRewardEnabledAt)
	if err != nil {
		return nil, err
	}
	if !qualified {
		return nil, ErrWelfareNewUserTrialNotAvailable.WithMetadata(map[string]string{"reason": welfareReasonNotReached})
	}
	if reward, err := s.getNewUserTrialSuccessReward(ctx, userID); err != nil {
		return nil, err
	} else if reward != nil {
		return nil, ErrWelfareNewUserTrialRewardClaimed.WithMetadata(map[string]string{"reason": welfareReasonAlreadyClaimed})
	}

	if trial == nil {
		trial = &WelfareNewUserTrial{UserID: userID}
	}
	if err := s.grantNewUserTrialSuccessReward(ctx, trial, settings.NewUserTrialSuccessRewardAmount); err != nil {
		return nil, err
	}
	view, err := s.buildNewUserTrialView(ctx, userID, settings)
	if err != nil {
		return nil, err
	}
	return &WelfareNewUserTrialRewardClaimResult{
		NewUserTrial:  view,
		ClaimedAmount: settings.NewUserTrialSuccessRewardAmount,
	}, nil
}

func (s *WelfareService) grantNewUserTrialSuccessReward(ctx context.Context, trial *WelfareNewUserTrial, amount float64) error {
	amount = normalizeNonNegativeFloat(amount)
	if amount <= 0 || trial == nil || trial.UserID <= 0 {
		return nil
	}
	if s.userRepo == nil || s.redeemRepo == nil {
		return ErrWelfareNewUserTrialUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return ErrWelfareNewUserTrialUnavailable
	}
	code := newUserTrialRewardRedeemCode(trial.UserID)
	if lookup, ok := s.redeemRepo.(welfareRedeemCodeLookupRepository); ok {
		existing, err := lookup.GetByCode(ctx, code)
		if err == nil && existing != nil {
			return nil
		}
		if err != nil && !errors.Is(err, ErrRedeemCodeNotFound) {
			return err
		}
	}

	now := s.nowTime().UTC()
	redeemCode := &RedeemCode{
		Code:   code,
		Type:   RedeemTypeNewUserReward,
		Value:  amount,
		Status: StatusUsed,
		UsedBy: &trial.UserID,
		UsedAt: &now,
		Notes:  "new user trial success reward",
	}
	if trial.ID > 0 {
		redeemCode.Notes = fmt.Sprintf("new user trial success reward trial %d", trial.ID)
	}

	grantCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return fmt.Errorf("begin new user trial reward transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		grantCtx = dbent.NewTxContext(ctx, tx)
	}

	if err := s.redeemRepo.Create(grantCtx, redeemCode); err != nil {
		if lookup, ok := s.redeemRepo.(welfareRedeemCodeLookupRepository); ok {
			if existing, getErr := lookup.GetByCode(ctx, code); getErr == nil && existing != nil {
				return nil
			}
		}
		if isRedeemCodeConflict(err) {
			return nil
		}
		return fmt.Errorf("create new user trial reward audit record: %w", err)
	}
	if err := grantWelfareBalance(grantCtx, s.userRepo, trial.UserID, amount); err != nil {
		return fmt.Errorf("update new user trial reward balance: %w", err)
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit new user trial reward transaction: %w", err)
		}
	}
	s.invalidateBalanceCaches(ctx, trial.UserID)
	return nil
}

func (s *WelfareService) ClaimDailyCheckin(ctx context.Context, userID int64) (*WelfareDailyCheckinClaimResult, error) {
	if s.repo == nil || s.userRepo == nil || s.redeemRepo == nil {
		return nil, ErrWelfareDailyCheckinUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrWelfareDailyCheckinUnavailable
	}

	now := s.nowTime()
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, ErrWelfareDisabled
	}
	if !settings.DailyCheckinEnabled {
		return nil, ErrWelfareDailyCheckinDisabled
	}
	if settings.RewardMax <= 0 {
		return nil, ErrWelfareDailyCheckinNotClaimable.WithMetadata(map[string]string{"reason": welfareReasonZeroReward})
	}
	canClaim, eligibleAt, err := s.canClaimDailyCheckinByAccountAge(ctx, userID, now, settings.DailyCheckinMinAccountAgeHours)
	if err != nil {
		return nil, err
	}
	if !canClaim {
		return nil, ErrWelfareDailyCheckinNotClaimable.WithMetadata(map[string]string{
			"reason":          welfareReasonRegistrationTooNew,
			"can_claim_after": eligibleAt.UTC().Format(time.RFC3339),
		})
	}

	amount, err := randomRewardAmount(settings.RewardMin, settings.RewardMax)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, ErrWelfareDailyCheckinNotClaimable.WithMetadata(map[string]string{"reason": welfareReasonZeroReward})
	}

	today := timezone.StartOfDay(now)
	claim := &WelfareDailyCheckinRecord{
		CheckinDate: today.Format("2006-01-02"),
		RewardMonth: today.Format("2006-01"),
		UserID:      userID,
		Amount:      amount,
	}

	claimCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin daily check-in transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		claimCtx = dbent.NewTxContext(ctx, tx)
	}

	if err := s.repo.CreateDailyCheckin(claimCtx, claim); err != nil {
		if errors.Is(err, ErrWelfareDailyCheckinAlreadyClaimed) {
			return nil, ErrWelfareDailyCheckinAlreadyClaimed
		}
		return nil, err
	}

	usedAt := now.UTC()
	redeemCode := &RedeemCode{
		Code:   dailyCheckinRedeemCode(claim.CheckinDate, userID),
		Type:   RedeemTypeDailyCheckin,
		Value:  amount,
		Status: StatusUsed,
		UsedBy: &userID,
		UsedAt: &usedAt,
		Notes:  fmt.Sprintf("daily check-in reward %s", claim.CheckinDate),
	}
	if err := s.redeemRepo.Create(claimCtx, redeemCode); err != nil {
		return nil, fmt.Errorf("create daily check-in audit record: %w", err)
	}
	if err := grantWelfareBalance(claimCtx, s.userRepo, userID, amount); err != nil {
		return nil, fmt.Errorf("update daily check-in balance: %w", err)
	}
	if err := s.repo.AttachDailyCheckinRedeemCode(claimCtx, claim.ID, redeemCode.ID); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit daily check-in transaction: %w", err)
		}
	}

	s.invalidateBalanceCaches(ctx, userID)
	daily, err := s.buildDailyCheckinView(ctx, userID, now, settings)
	if err != nil {
		return nil, err
	}
	return &WelfareDailyCheckinClaimResult{
		DailyCheckin:  daily,
		ClaimedAmount: amount,
	}, nil
}

func (s *WelfareService) ClaimDailyCheckinMilestone(ctx context.Context, userID int64, day int) (*WelfareDailyCheckinMilestoneClaimResult, error) {
	if s.repo == nil || s.userRepo == nil || s.redeemRepo == nil {
		return nil, ErrWelfareDailyCheckinUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrWelfareDailyCheckinUnavailable
	}
	if !validWelfareMilestoneDay(day) {
		return nil, ErrWelfareCheckinMilestoneNotClaimable.WithMetadata(map[string]string{"reason": welfareReasonNotConfigured})
	}

	now := s.nowTime()
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, ErrWelfareDisabled
	}
	if !settings.DailyCheckinEnabled {
		return nil, ErrWelfareDailyCheckinDisabled
	}
	if !settings.MilestoneEnabled {
		return nil, ErrWelfareCheckinMilestoneNotClaimable.WithMetadata(map[string]string{"reason": welfareReasonDisabled})
	}

	status, err := s.buildDailyCheckinView(ctx, userID, now, settings)
	if err != nil {
		return nil, err
	}
	var milestone WelfareDailyCheckinMilestone
	found := false
	for _, item := range status.Milestones {
		if item.Day == day {
			milestone = item
			found = true
			break
		}
	}
	if !found || !milestone.Claimable {
		reason := welfareReasonNotReached
		if found {
			reason = milestone.Reason
		}
		if reason == welfareReasonAlreadyClaimed {
			return nil, ErrWelfareCheckinMilestoneAlreadyClaimed
		}
		return nil, ErrWelfareCheckinMilestoneNotClaimable.WithMetadata(map[string]string{"reason": reason})
	}

	claim := &WelfareDailyCheckinMilestoneClaim{
		RewardMonth:  status.RewardMonth,
		MilestoneDay: day,
		UserID:       userID,
		Amount:       milestone.Amount,
	}

	claimCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin daily check-in milestone transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		claimCtx = dbent.NewTxContext(ctx, tx)
	}

	if err := s.repo.CreateDailyCheckinMilestoneClaim(claimCtx, claim); err != nil {
		if errors.Is(err, ErrWelfareCheckinMilestoneAlreadyClaimed) {
			return nil, ErrWelfareCheckinMilestoneAlreadyClaimed
		}
		return nil, err
	}

	usedAt := now.UTC()
	redeemCode := &RedeemCode{
		Code:   dailyCheckinMilestoneRedeemCode(status.RewardMonth, day, userID),
		Type:   RedeemTypeCheckinMilestone,
		Value:  milestone.Amount,
		Status: StatusUsed,
		UsedBy: &userID,
		UsedAt: &usedAt,
		Notes:  fmt.Sprintf("daily check-in milestone %s day %d", status.RewardMonth, day),
	}
	if err := s.redeemRepo.Create(claimCtx, redeemCode); err != nil {
		return nil, fmt.Errorf("create daily check-in milestone audit record: %w", err)
	}
	if err := grantWelfareBalance(claimCtx, s.userRepo, userID, milestone.Amount); err != nil {
		return nil, fmt.Errorf("update daily check-in milestone balance: %w", err)
	}
	if err := s.repo.AttachDailyCheckinMilestoneRedeemCode(claimCtx, claim.ID, redeemCode.ID); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit daily check-in milestone transaction: %w", err)
		}
	}

	s.invalidateBalanceCaches(ctx, userID)
	daily, err := s.buildDailyCheckinView(ctx, userID, now, settings)
	if err != nil {
		return nil, err
	}
	for _, item := range daily.Milestones {
		if item.Day == day {
			milestone = item
			break
		}
	}
	return &WelfareDailyCheckinMilestoneClaimResult{
		DailyCheckin:  daily,
		Milestone:     milestone,
		ClaimedAmount: claim.Amount,
	}, nil
}

func (s *WelfareService) buildDailyCheckinView(ctx context.Context, userID int64, now time.Time, settings welfareSettings) (*WelfareDailyCheckinView, error) {
	today := timezone.StartOfDay(now)
	month := today.Format("2006-01")
	view := &WelfareDailyCheckinView{
		Enabled:            settings.Enabled && settings.DailyCheckinEnabled,
		Today:              today.Format("2006-01-02"),
		RewardMonth:        month,
		RewardMin:          settings.RewardMin,
		RewardMax:          settings.RewardMax,
		MilestoneEnabled:   settings.Enabled && settings.DailyCheckinEnabled && settings.MilestoneEnabled,
		Milestones:         make([]WelfareDailyCheckinMilestone, 0, 4),
		Reason:             welfareReasonAvailable,
		SettlementTimezone: timezone.Name(),
	}
	if !settings.Enabled {
		view.Reason = welfareReasonDisabled
	}
	if settings.Enabled && !settings.DailyCheckinEnabled {
		view.Reason = welfareReasonDisabled
	}
	if !view.Enabled || s.repo == nil {
		return view, nil
	}

	checkins, err := s.repo.ListDailyCheckins(ctx, userID, month)
	if err != nil {
		return nil, err
	}
	checkedDates := make(map[string]WelfareDailyCheckinRecord, len(checkins))
	for _, item := range checkins {
		checkedDates[item.CheckinDate] = item
		view.CheckinDates = append(view.CheckinDates, item.CheckinDate)
	}
	view.MonthCheckinDays = len(view.CheckinDates)
	if todayRecord, ok := checkedDates[view.Today]; ok {
		view.CheckedToday = true
		view.TodayRewardAmount = todayRecord.Amount
	}
	view.CurrentStreakDays = monthlyStreakDays(today, checkedDates)
	view.CanClaimToday = view.Enabled && !view.CheckedToday && settings.RewardMax > 0
	if !view.Enabled {
		view.CanClaimToday = false
		view.Reason = welfareReasonDisabled
	} else if view.CheckedToday {
		view.Reason = welfareReasonAlreadyChecked
	} else if settings.RewardMax <= 0 {
		view.Reason = welfareReasonZeroReward
	}
	if view.CanClaimToday {
		canClaim, eligibleAt, err := s.canClaimDailyCheckinByAccountAge(ctx, userID, now, settings.DailyCheckinMinAccountAgeHours)
		if err != nil {
			return nil, err
		}
		if !canClaim {
			view.CanClaimToday = false
			view.Reason = welfareReasonRegistrationTooNew
			view.CanClaimAfter = eligibleAt.UTC().Format(time.RFC3339)
		}
	}

	if !view.MilestoneEnabled {
		return view, nil
	}

	claims, err := s.repo.ListDailyCheckinMilestoneClaims(ctx, month, userID)
	if err != nil {
		return nil, err
	}
	claimMap := make(map[int]WelfareDailyCheckinMilestoneClaim, len(claims))
	for _, claim := range claims {
		claimMap[claim.MilestoneDay] = claim
	}
	for _, day := range []int{welfareMilestoneDay7, welfareMilestoneDay14, welfareMilestoneDay21, welfareMilestoneDay28} {
		amount := settings.MilestoneAmounts[day]
		item := WelfareDailyCheckinMilestone{
			Day:    day,
			Amount: amount,
			Reason: welfareReasonNotReached,
		}
		if claim, ok := claimMap[day]; ok {
			item.Claimed = true
			item.Claimable = false
			item.Reason = welfareReasonAlreadyClaimed
			item.ClaimedAt = claim.CreatedAt.UTC().Format(time.RFC3339)
			item.RedeemCodeID = claim.RedeemCodeID
		} else if !view.Enabled {
			item.Reason = welfareReasonDisabled
		} else if amount <= 0 {
			item.Reason = welfareReasonZeroReward
		} else if view.CurrentStreakDays >= day {
			item.Claimable = true
			item.Reason = welfareReasonAvailable
		}
		view.Milestones = append(view.Milestones, item)
	}
	return view, nil
}

func (s *WelfareService) buildNewUserTrialView(ctx context.Context, userID int64, settings welfareSettings) (*WelfareNewUserTrialView, error) {
	quota := settings.NewUserTrialQuotaAmount
	if quota <= 0 {
		quota = defaultNewUserTrialQuotaAmount
	}
	view := &WelfareNewUserTrialView{
		Enabled:             settings.Enabled && settings.NewUserTrialEnabled,
		QuotaAmount:         quota,
		SuccessRewardAmount: settings.NewUserTrialSuccessRewardAmount,
		SuccessRewardReason: welfareReasonNotReached,
		Reason:              welfareReasonAvailable,
		Status:              "available",
		CanUse:              settings.Enabled && settings.NewUserTrialEnabled,
	}
	if !settings.Enabled || !settings.NewUserTrialEnabled {
		view.CanUse = false
		view.Reason = welfareReasonDisabled
	}
	if s.repo == nil || !view.Enabled {
		s.applyNewUserTrialSuccessRewardState(ctx, userID, view, nil, settings.NewUserTrialSuccessRewardEnabledAt)
		view.RemainingQuota = normalizeTrialRemaining(view.QuotaAmount, view.QuotaUsed)
		return view, nil
	}
	trial, err := s.repo.GetNewUserTrial(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrWelfareNewUserTrialNotFound) {
			s.applyNewUserTrialSuccessRewardState(ctx, userID, view, nil, settings.NewUserTrialSuccessRewardEnabledAt)
			view.RemainingQuota = normalizeTrialRemaining(view.QuotaAmount, view.QuotaUsed)
			return view, nil
		}
		return nil, err
	}
	view.QuotaAmount = trial.QuotaAmount
	if view.QuotaAmount <= 0 {
		view.QuotaAmount = quota
	}
	view.QuotaUsed = trial.QuotaUsed
	view.RemainingQuota = normalizeTrialRemaining(view.QuotaAmount, trial.QuotaUsed)
	view.Status = normalizeNewUserTrialStatus(trial.Status, view.RemainingQuota)
	view.CanUse = view.Enabled && trialCanBegin(view.Status) && view.RemainingQuota > 0
	view.Reason = trialViewReason(view.Status, view.CanUse)
	if trial.FirstStartedAt != nil {
		view.FirstStartedAt = trial.FirstStartedAt.UTC().Format(time.RFC3339)
	}
	if trial.FirstSuccessAt != nil {
		view.FirstSuccessAt = trial.FirstSuccessAt.UTC().Format(time.RFC3339)
	}
	s.applyNewUserTrialSuccessRewardState(ctx, userID, view, trial, settings.NewUserTrialSuccessRewardEnabledAt)
	return view, nil
}

func (s *WelfareService) buildRechargeView(ctx context.Context, userID int64, settings welfareSettings) (*WelfareRechargeView, error) {
	view := &WelfareRechargeView{
		Enabled:              settings.Enabled && settings.RechargeEnabled && settings.FirstRechargeBonusAmount > 0,
		FirstBonusAmount:     settings.FirstRechargeBonusAmount,
		ValidDays:            settings.FirstRechargeBonusValidDays,
		CanStackMonthlyBonus: settings.FirstRechargeBonusStackMonthly,
		MonthlyBonusMayBlock: !settings.FirstRechargeBonusStackMonthly,
		Reason:               welfareReasonAvailable,
	}
	if !settings.Enabled || !settings.RechargeEnabled {
		view.Reason = welfareReasonDisabled
	}
	if settings.FirstRechargeBonusAmount <= 0 {
		view.Reason = welfareReasonZeroReward
	}
	if settings.FirstRechargeBonusValidDays > 0 {
		user, err := s.firstRechargeBonusUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if expiresAt := firstRechargeBonusExpiresAt(user, settings.FirstRechargeBonusValidDays); !expiresAt.IsZero() {
			view.ExpiresAt = expiresAt.UTC().Format(time.RFC3339)
			if view.Enabled && !s.nowTime().Before(expiresAt) {
				view.Enabled = false
				view.Reason = welfareReasonExpired
			}
		}
	}
	reward, err := s.getFirstRechargeBonus(ctx, userID)
	if err != nil {
		return nil, err
	}
	if reward != nil {
		view.FirstBonusClaimed = true
		view.Reason = welfareReasonAlreadyClaimed
		if reward.UsedAt != nil {
			view.FirstBonusClaimedAt = reward.UsedAt.UTC().Format(time.RFC3339)
		}
	}
	if !view.FirstBonusClaimed {
		firstRechargeCompleted, err := s.hasFirstRechargeBlockingOrder(ctx, userID)
		if err != nil {
			return nil, err
		}
		view.FirstRechargeCompleted = firstRechargeCompleted
		if firstRechargeCompleted && view.Reason == welfareReasonAvailable {
			view.Enabled = false
			view.Reason = welfareReasonNotReached
		}
	}
	return view, nil
}

func (s *WelfareService) applyNewUserTrialSuccessRewardState(ctx context.Context, userID int64, view *WelfareNewUserTrialView, trial *WelfareNewUserTrial, enabledAt *time.Time) {
	if view == nil {
		return
	}
	view.SuccessRewardClaimable = false
	view.SuccessRewardClaimed = false
	if view.SuccessRewardAmount <= 0 {
		view.SuccessRewardReason = welfareReasonZeroReward
		return
	}
	if !view.Enabled {
		view.SuccessRewardReason = welfareReasonDisabled
		return
	}
	qualified, err := s.hasNewUserTrialRewardQualifyingUsage(ctx, userID, trial, enabledAt)
	if err != nil {
		view.SuccessRewardReason = welfareReasonNotConfigured
		return
	}
	if !qualified {
		view.SuccessRewardReason = welfareReasonNotReached
		return
	}
	reward, err := s.getNewUserTrialSuccessReward(ctx, userID)
	if err != nil {
		view.SuccessRewardReason = welfareReasonNotConfigured
		return
	}
	if reward != nil {
		view.SuccessRewardClaimed = true
		view.SuccessRewardReason = welfareReasonAlreadyClaimed
		if reward.UsedAt != nil {
			view.SuccessRewardClaimedAt = reward.UsedAt.UTC().Format(time.RFC3339)
		}
		return
	}
	view.SuccessRewardClaimable = true
	view.SuccessRewardReason = welfareReasonAvailable
}

func (s *WelfareService) hasNewUserTrialRewardQualifyingUsage(ctx context.Context, userID int64, trial *WelfareNewUserTrial, enabledAt *time.Time) (bool, error) {
	if s == nil || s.repo == nil || s.userRepo == nil || userID <= 0 || enabledAt == nil || enabledAt.IsZero() {
		return false, nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil || user.CreatedAt.IsZero() || !user.CreatedAt.After(*enabledAt) {
		return false, nil
	}
	var firstSuccessAt *time.Time
	if trial != nil && trial.FirstSuccessAt != nil {
		firstSuccessAt = trial.FirstSuccessAt
	} else {
		firstSuccessAt, err = s.repo.FirstSuccessfulUsageAt(ctx, userID)
		if err != nil {
			return false, err
		}
	}
	return firstSuccessAt != nil && firstSuccessAt.After(*enabledAt), nil
}

func (s *WelfareService) getNewUserTrialSuccessReward(ctx context.Context, userID int64) (*RedeemCode, error) {
	if s == nil || s.redeemRepo == nil || userID <= 0 {
		return nil, nil
	}
	lookup, ok := s.redeemRepo.(welfareRedeemCodeLookupRepository)
	if !ok {
		return nil, nil
	}
	reward, err := lookup.GetByCode(ctx, newUserTrialRewardRedeemCode(userID))
	if err != nil {
		if errors.Is(err, ErrRedeemCodeNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return reward, nil
}

func (s *WelfareService) getFirstRechargeBonus(ctx context.Context, userID int64) (*RedeemCode, error) {
	if s == nil || s.redeemRepo == nil || userID <= 0 {
		return nil, nil
	}
	lookup, ok := s.redeemRepo.(welfareRedeemCodeLookupRepository)
	if !ok {
		return nil, nil
	}
	reward, err := lookup.GetByCode(ctx, firstRechargeBonusRedeemCode(userID))
	if err != nil {
		if errors.Is(err, ErrRedeemCodeNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return reward, nil
}

func (s *WelfareService) GrantFirstRechargeBonusForOrder(ctx context.Context, orderID int64) (bool, error) {
	if s == nil || s.entClient == nil || s.userRepo == nil || s.redeemRepo == nil || orderID <= 0 {
		return false, nil
	}
	settings, err := s.getSettings(ctx)
	if err != nil {
		return false, err
	}
	amount := normalizeNonNegativeFloat(settings.FirstRechargeBonusAmount)
	if !settings.Enabled || !settings.RechargeEnabled || amount <= 0 {
		return false, nil
	}

	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if order.UserID <= 0 || order.OrderType != payment.OrderTypeBalance || order.Amount <= 0 || !isFirstRechargeBonusEligibleOrderStatus(order.Status) {
		return false, nil
	}
	if order.Status == OrderStatusFailed && !paymentOrderHasPaymentEvidence(order) {
		return false, nil
	}

	if reward, err := s.getFirstRechargeBonus(ctx, order.UserID); err != nil {
		return false, err
	} else if reward != nil {
		return false, nil
	}

	first, err := s.isFirstRechargeOrder(ctx, order)
	if err != nil || !first {
		return false, err
	}
	if !settings.FirstRechargeBonusStackMonthly && monthlyRechargeDecisionFromOrder(order).BonusAmount > 0 && order.Amount > order.PayAmount {
		return false, nil
	}
	if settings.FirstRechargeBonusValidDays > 0 {
		user, err := s.firstRechargeBonusUser(ctx, order.UserID)
		if err != nil {
			return false, err
		}
		if firstRechargeBonusExpired(user, settings.FirstRechargeBonusValidDays, s.nowTime()) {
			return false, nil
		}
	}

	code := firstRechargeBonusRedeemCode(order.UserID)
	now := s.nowTime().UTC()
	redeemCode := &RedeemCode{
		Code:   code,
		Type:   RedeemTypeFirstRechargeBonus,
		Value:  amount,
		Status: StatusUsed,
		UsedBy: &order.UserID,
		UsedAt: &now,
		Notes:  fmt.Sprintf("first recharge bonus order %d", order.ID),
	}

	grantCtx := ctx
	var tx *dbent.Tx
	if dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return false, fmt.Errorf("begin first recharge bonus transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		grantCtx = dbent.NewTxContext(ctx, tx)
	}

	if err := s.redeemRepo.Create(grantCtx, redeemCode); err != nil {
		if isRedeemCodeConflict(err) {
			return false, nil
		}
		return false, fmt.Errorf("create first recharge bonus audit record: %w", err)
	}
	if err := grantWelfareBalance(grantCtx, s.userRepo, order.UserID, amount); err != nil {
		return false, fmt.Errorf("update first recharge bonus balance: %w", err)
	}
	if err := s.writeFirstRechargeBonusAudit(grantCtx, order.ID, amount, code); err != nil {
		return false, err
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit first recharge bonus transaction: %w", err)
		}
	}
	s.invalidateBalanceCaches(ctx, order.UserID)
	return true, nil
}

func (s *WelfareService) isFirstRechargeOrder(ctx context.Context, order *dbent.PaymentOrder) (bool, error) {
	if s == nil || s.entClient == nil || order == nil || order.ID <= 0 || order.UserID <= 0 {
		return false, nil
	}
	count, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(order.UserID),
			paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
			paymentorder.AmountGT(0),
			paymentorder.IDLT(order.ID),
			firstRechargeBlockingOrderPredicate(),
		).
		Limit(1).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *WelfareService) hasFirstRechargeBlockingOrder(ctx context.Context, userID int64) (bool, error) {
	if s == nil || s.entClient == nil || userID <= 0 {
		return false, nil
	}
	return s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(userID),
			paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
			paymentorder.AmountGT(0),
			firstRechargeBlockingOrderPredicate(),
		).
		Limit(1).
		Exist(ctx)
}

func (s *WelfareService) firstRechargeBonusUser(ctx context.Context, userID int64) (*User, error) {
	if s == nil || s.userRepo == nil || userID <= 0 {
		return nil, nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func firstRechargeBonusExpiresAt(user *User, validDays int) time.Time {
	if user == nil || user.CreatedAt.IsZero() || validDays <= 0 {
		return time.Time{}
	}
	return user.CreatedAt.AddDate(0, 0, validDays)
}

func firstRechargeBonusExpired(user *User, validDays int, now time.Time) bool {
	expiresAt := firstRechargeBonusExpiresAt(user, validDays)
	return !expiresAt.IsZero() && !now.Before(expiresAt)
}

func (s *WelfareService) writeFirstRechargeBonusAudit(ctx context.Context, orderID int64, amount float64, code string) error {
	client := s.entClientFromContext(ctx)
	if client == nil || orderID <= 0 {
		return nil
	}
	detail, _ := json.Marshal(map[string]any{
		"amount":     amount,
		"redeemCode": code,
	})
	if _, err := client.PaymentAuditLog.Create().
		SetOrderID(strconv.FormatInt(orderID, 10)).
		SetAction(welfareFirstRechargeBonusAudit).
		SetDetail(string(detail)).
		SetOperator("system").
		Save(ctx); err != nil {
		return fmt.Errorf("write first recharge bonus audit: %w", err)
	}
	return nil
}

func (s *WelfareService) OrderHasFirstRechargeBonus(ctx context.Context, orderID int64) (bool, error) {
	client := s.entClientFromContext(ctx)
	if client == nil || orderID <= 0 {
		return false, nil
	}
	logs, err := client.PaymentAuditLog.Query().
		Where(
			paymentauditlog.OrderIDEQ(strconv.FormatInt(orderID, 10)),
			paymentauditlog.ActionEQ(welfareFirstRechargeBonusAudit),
		).
		All(ctx)
	if err != nil {
		return false, err
	}
	for _, log := range logs {
		if !firstRechargeBonusAuditLooksLikeMonthlyRecharge(log.Detail) {
			return true, nil
		}
	}
	return false, nil
}

func firstRechargeBonusAuditLooksLikeMonthlyRecharge(detail string) bool {
	var payload map[string]any
	if err := json.Unmarshal([]byte(detail), &payload); err != nil {
		return false
	}
	if code := strings.ToUpper(stringSnapshotValue(payload, "redeemCode")); code != "" {
		return strings.HasPrefix(code, "MRB")
	}
	return strings.TrimSpace(stringSnapshotValue(payload, "period")) != "" &&
		strings.TrimSpace(stringSnapshotValue(payload, "package_id")) != ""
}

func (s *WelfareService) entClientFromContext(ctx context.Context) *dbent.Client {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	if s == nil {
		return nil
	}
	return s.entClient
}

func (s *WelfareService) canClaimDailyCheckinByAccountAge(ctx context.Context, userID int64, now time.Time, minAgeHours int) (bool, time.Time, error) {
	if s == nil || s.userRepo == nil || userID <= 0 || minAgeHours <= 0 {
		return true, time.Time{}, nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, time.Time{}, err
	}
	if user == nil || user.CreatedAt.IsZero() {
		return true, time.Time{}, nil
	}
	eligibleAt := user.CreatedAt.Add(time.Duration(minAgeHours) * time.Hour)
	return !now.Before(eligibleAt), eligibleAt, nil
}

func normalizeTrialRemaining(quotaAmount, quotaUsed float64) float64 {
	remaining := normalizeNonNegativeFloat(quotaAmount) - normalizeNonNegativeFloat(quotaUsed)
	if remaining < 0 {
		return 0
	}
	return math.Round(remaining*float64(welfareRewardScale)) / float64(welfareRewardScale)
}

func normalizeNewUserTrialStatus(status string, remaining float64) string {
	switch strings.TrimSpace(status) {
	case "in_progress":
		return "in_progress"
	case "active":
		if remaining <= 0 {
			return "exhausted"
		}
		return "active"
	case "exhausted":
		return "exhausted"
	default:
		if remaining <= 0 {
			return "exhausted"
		}
		return "available"
	}
}

func trialCanBegin(status string) bool {
	switch normalizeNewUserTrialStatus(status, 1) {
	case "available", "active":
		return true
	default:
		return false
	}
}

func trialViewReason(status string, canUse bool) string {
	if canUse {
		return welfareReasonAvailable
	}
	switch status {
	case "in_progress":
		return welfareReasonInProgress
	case "exhausted":
		return welfareReasonExhausted
	default:
		return welfareReasonDisabled
	}
}

func (s *WelfareService) getSettings(ctx context.Context) (welfareSettings, error) {
	result := welfareSettings{
		MilestoneAmounts: map[int]float64{
			welfareMilestoneDay7:  0,
			welfareMilestoneDay14: 0,
			welfareMilestoneDay21: 0,
			welfareMilestoneDay28: 0,
		},
	}
	if s.settingRepo == nil {
		return result, nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyWelfareEnabled,
		SettingKeyWelfareDailyCheckinEnabled,
		SettingKeyWelfareRechargeEnabled,
		SettingKeyWelfareVIPEnabled,
		SettingKeyWelfareDailyCheckinRewardMin,
		SettingKeyWelfareDailyCheckinRewardMax,
		SettingKeyWelfareDailyCheckinMinAccountAgeHours,
		SettingKeyWelfareDailyCheckinMilestoneEnabled,
		SettingKeyWelfareDailyCheckinMilestone7Amount,
		SettingKeyWelfareDailyCheckinMilestone14Amount,
		SettingKeyWelfareDailyCheckinMilestone21Amount,
		SettingKeyWelfareDailyCheckinMilestone28Amount,
		SettingKeyWelfareNewUserTrialEnabled,
		SettingKeyWelfareNewUserTrialQuotaAmount,
		SettingKeyWelfareNewUserTrialSuccessRewardAmount,
		SettingKeyWelfareNewUserTrialSuccessRewardEnabledAt,
		SettingKeyWelfareNewUserTrialDailySiteQuotaAmount,
		SettingKeyWelfareNewUserTrialDailyIPActivationLimit,
		SettingKeyWelfareFirstRechargeBonusAmount,
		SettingKeyWelfareFirstRechargeBonusValidDays,
		SettingKeyWelfareFirstRechargeBonusStackMonthly,
	})
	if err != nil {
		return result, fmt.Errorf("get welfare settings: %w", err)
	}
	result.Enabled = values[SettingKeyWelfareEnabled] == "true"
	result.DailyCheckinEnabled = values[SettingKeyWelfareDailyCheckinEnabled] == "true"
	result.NewUserTrialEnabled = values[SettingKeyWelfareNewUserTrialEnabled] == "true"
	result.RechargeEnabled = values[SettingKeyWelfareRechargeEnabled] == "true"
	result.VIPEnabled = values[SettingKeyWelfareVIPEnabled] == "true"
	result.RewardMin = normalizeDailyRewardAmount(parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinRewardMin], 0))
	result.RewardMax = normalizeDailyRewardAmount(parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinRewardMax], result.RewardMin))
	if result.RewardMax < result.RewardMin {
		result.RewardMax = result.RewardMin
	}
	result.DailyCheckinMinAccountAgeHours = parseNonNegativeIntSetting(values[SettingKeyWelfareDailyCheckinMinAccountAgeHours], defaultDailyCheckinMinAccountAgeHours)
	result.MilestoneEnabled = !isFalseSettingValue(values[SettingKeyWelfareDailyCheckinMilestoneEnabled])
	result.MilestoneAmounts[welfareMilestoneDay7] = parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinMilestone7Amount], 0)
	result.MilestoneAmounts[welfareMilestoneDay14] = parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinMilestone14Amount], 0)
	result.MilestoneAmounts[welfareMilestoneDay21] = parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinMilestone21Amount], 0)
	result.MilestoneAmounts[welfareMilestoneDay28] = parseNonNegativeFloatSetting(values[SettingKeyWelfareDailyCheckinMilestone28Amount], 0)
	result.NewUserTrialQuotaAmount = parseNonNegativeFloatSetting(values[SettingKeyWelfareNewUserTrialQuotaAmount], defaultNewUserTrialQuotaAmount)
	if result.NewUserTrialQuotaAmount <= 0 {
		result.NewUserTrialQuotaAmount = defaultNewUserTrialQuotaAmount
	}
	result.NewUserTrialSuccessRewardAmount = parseNonNegativeFloatSetting(values[SettingKeyWelfareNewUserTrialSuccessRewardAmount], 0)
	result.NewUserTrialSuccessRewardEnabledAt = parseOptionalRFC3339Setting(values[SettingKeyWelfareNewUserTrialSuccessRewardEnabledAt])
	result.NewUserTrialDailySiteQuotaAmount = parseNonNegativeFloatSetting(values[SettingKeyWelfareNewUserTrialDailySiteQuotaAmount], defaultNewUserTrialDailySiteQuotaAmount)
	result.FirstRechargeBonusAmount = parseNonNegativeFloatSetting(values[SettingKeyWelfareFirstRechargeBonusAmount], defaultFirstRechargeBonusAmount)
	result.FirstRechargeBonusValidDays = parseNonNegativeIntSetting(values[SettingKeyWelfareFirstRechargeBonusValidDays], 0)
	result.FirstRechargeBonusStackMonthly = values[SettingKeyWelfareFirstRechargeBonusStackMonthly] == "true"
	if ipLimit, err := strconv.Atoi(strings.TrimSpace(values[SettingKeyWelfareNewUserTrialDailyIPActivationLimit])); err == nil && ipLimit >= 0 {
		result.NewUserTrialDailyIPActivationLimit = ipLimit
	} else {
		result.NewUserTrialDailyIPActivationLimit = defaultNewUserTrialDailyIPActivationLimit
	}
	return result, nil
}

func (s *WelfareService) nowTime() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return timezone.Now()
}

func monthlyStreakDays(today time.Time, checkedDates map[string]WelfareDailyCheckinRecord) int {
	current := today
	if _, ok := checkedDates[current.Format("2006-01-02")]; !ok {
		current = current.AddDate(0, 0, -1)
	}
	month := today.Format("2006-01")
	streak := 0
	for current.Format("2006-01") == month {
		if _, ok := checkedDates[current.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		current = current.AddDate(0, 0, -1)
	}
	return streak
}

func randomRewardAmount(minValue, maxValue float64) (float64, error) {
	minUnits := int64(math.Round(normalizeNonNegativeFloat(minValue) * float64(welfareDailyRewardStepScale)))
	maxUnits := int64(math.Round(normalizeNonNegativeFloat(maxValue) * float64(welfareDailyRewardStepScale)))
	if maxUnits <= 0 {
		return 0, nil
	}
	if minUnits <= 0 {
		minUnits = 1
	}
	if maxUnits < minUnits {
		maxUnits = minUnits
	}
	if maxUnits == minUnits {
		return float64(minUnits) / float64(welfareDailyRewardStepScale), nil
	}
	n, err := crand.Int(crand.Reader, big.NewInt(maxUnits-minUnits+1))
	if err != nil {
		return 0, fmt.Errorf("generate daily check-in reward: %w", err)
	}
	return float64(minUnits+n.Int64()) / float64(welfareDailyRewardStepScale), nil
}

func normalizeDailyRewardAmount(value float64) float64 {
	return math.Round(normalizeNonNegativeFloat(value)*float64(welfareDailyRewardStepScale)) / float64(welfareDailyRewardStepScale)
}

func validWelfareMilestoneDay(day int) bool {
	switch day {
	case welfareMilestoneDay7, welfareMilestoneDay14, welfareMilestoneDay21, welfareMilestoneDay28:
		return true
	default:
		return false
	}
}

func grantWelfareBalance(ctx context.Context, userRepo UserRepository, userID int64, amount float64) error {
	if grantRepo, ok := userRepo.(welfareBalanceGrantRepository); ok {
		return grantRepo.AddBalance(ctx, userID, amount)
	}
	return userRepo.UpdateBalance(ctx, userID, amount)
}

func (s *WelfareService) invalidateBalanceCaches(ctx context.Context, userID int64) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheInvalidator != nil {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.billingCacheInvalidator.InvalidateUserBalance(cacheCtx, userID)
	}
}

func dailyCheckinRedeemCode(checkinDate string, userID int64) string {
	return "DCK" + strings.ReplaceAll(checkinDate, "-", "") + "U" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func dailyCheckinMilestoneRedeemCode(month string, day int, userID int64) string {
	return "DCM" + strings.ReplaceAll(month, "-", "") + "D" + strconv.Itoa(day) + "U" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func newUserTrialRewardRedeemCode(userID int64) string {
	return "NUTR" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func firstRechargeBonusRedeemCode(userID int64) string {
	return "FRB" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func firstRechargeBlockingOrderStatuses() []string {
	return []string{
		OrderStatusPaid,
		OrderStatusRecharging,
		OrderStatusCompleted,
		OrderStatusRefundRequested,
		OrderStatusRefunding,
		OrderStatusPartiallyRefunded,
		OrderStatusRefunded,
		OrderStatusRefundFailed,
	}
}

func firstRechargeBlockingOrderPredicate() predicate.PaymentOrder {
	return paymentorder.Or(
		paymentorder.StatusIn(firstRechargeBlockingOrderStatuses()...),
		paymentorder.And(
			paymentorder.StatusEQ(OrderStatusFailed),
			paymentorder.Or(
				paymentorder.PaidAtNotNil(),
				paymentorder.PaymentTradeNoNEQ(""),
			),
		),
	)
}

func paymentOrderHasPaymentEvidence(order *dbent.PaymentOrder) bool {
	return order != nil && (order.PaidAt != nil || strings.TrimSpace(order.PaymentTradeNo) != "")
}

func isFirstRechargeBonusEligibleOrderStatus(status string) bool {
	for _, item := range firstRechargeBlockingOrderStatuses() {
		if status == item {
			return true
		}
	}
	return status == OrderStatusFailed
}

func isRedeemCodeConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRedeemCodeUsed) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique") || strings.Contains(message, "duplicate")
}
