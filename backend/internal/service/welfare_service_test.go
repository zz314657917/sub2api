package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	apptimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

type welfareRepoStub struct {
	daily      []WelfareDailyCheckinRecord
	milestones []WelfareDailyCheckinMilestoneClaim
	trials     []WelfareNewUserTrial
	trialUsage map[string]WelfareNewUserTrialConsumeInput

	createDailyErr     error
	createMilestoneErr error

	createdDaily      []WelfareDailyCheckinRecord
	createdMilestones []WelfareDailyCheckinMilestoneClaim
	grantedVouchers   []WelfareVoucherGrantInput
	voucherSummary    *WelfareVoucherBalanceSummary

	attachedDailyClaimID      int64
	attachedDailyRedeemID     int64
	attachedMilestoneClaimID  int64
	attachedMilestoneRedeemID int64

	siteUsageSince float64
	firstUsageAt   *time.Time
	ipActivations  map[string]int
}

func (r *welfareRepoStub) GetDailyCheckin(_ context.Context, checkinDate string, userID int64) (*WelfareDailyCheckinRecord, error) {
	for _, item := range r.daily {
		if item.CheckinDate == checkinDate && item.UserID == userID {
			clone := item
			return &clone, nil
		}
	}
	return nil, ErrWelfareDailyCheckinNotFound
}

func (r *welfareRepoStub) ListDailyCheckins(_ context.Context, userID int64, month string) ([]WelfareDailyCheckinRecord, error) {
	out := make([]WelfareDailyCheckinRecord, 0)
	for _, item := range r.daily {
		if item.UserID == userID && item.RewardMonth == month {
			out = append(out, item)
		}
	}
	for _, item := range r.createdDaily {
		if item.UserID == userID && item.RewardMonth == month {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *welfareRepoStub) CreateDailyCheckin(_ context.Context, record *WelfareDailyCheckinRecord) error {
	if r.createDailyErr != nil {
		return r.createDailyErr
	}
	record.ID = int64(len(r.createdDaily) + 100)
	record.CreatedAt = time.Date(2026, 5, 13, 1, 2, 3, 0, time.UTC)
	r.createdDaily = append(r.createdDaily, *record)
	return nil
}

func (r *welfareRepoStub) AttachDailyCheckinRedeemCode(_ context.Context, claimID, redeemCodeID int64) error {
	r.attachedDailyClaimID = claimID
	r.attachedDailyRedeemID = redeemCodeID
	return nil
}

func (r *welfareRepoStub) GetDailyCheckinMilestoneClaim(_ context.Context, month string, milestoneDay int, userID int64) (*WelfareDailyCheckinMilestoneClaim, error) {
	for _, item := range r.milestones {
		if item.RewardMonth == month && item.MilestoneDay == milestoneDay && item.UserID == userID {
			clone := item
			return &clone, nil
		}
	}
	return nil, ErrWelfareCheckinMilestoneNotFound
}

func (r *welfareRepoStub) ListDailyCheckinMilestoneClaims(_ context.Context, month string, userID int64) ([]WelfareDailyCheckinMilestoneClaim, error) {
	out := make([]WelfareDailyCheckinMilestoneClaim, 0)
	for _, item := range r.milestones {
		if item.UserID == userID && item.RewardMonth == month {
			out = append(out, item)
		}
	}
	for _, item := range r.createdMilestones {
		if item.UserID == userID && item.RewardMonth == month {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *welfareRepoStub) CreateDailyCheckinMilestoneClaim(_ context.Context, claim *WelfareDailyCheckinMilestoneClaim) error {
	if r.createMilestoneErr != nil {
		return r.createMilestoneErr
	}
	claim.ID = int64(len(r.createdMilestones) + 200)
	claim.CreatedAt = time.Date(2026, 5, 13, 1, 2, 3, 0, time.UTC)
	r.createdMilestones = append(r.createdMilestones, *claim)
	return nil
}

func (r *welfareRepoStub) AttachDailyCheckinMilestoneRedeemCode(_ context.Context, claimID, redeemCodeID int64) error {
	r.attachedMilestoneClaimID = claimID
	r.attachedMilestoneRedeemID = redeemCodeID
	return nil
}

func (r *welfareRepoStub) GrantVoucher(_ context.Context, input WelfareVoucherGrantInput) error {
	r.grantedVouchers = append(r.grantedVouchers, input)
	return nil
}

func (r *welfareRepoStub) GetVoucherBalanceSummary(_ context.Context, userID int64) (*WelfareVoucherBalanceSummary, error) {
	if r.voucherSummary != nil {
		clone := *r.voucherSummary
		return &clone, nil
	}
	return &WelfareVoucherBalanceSummary{TotalAvailable: 0}, nil
}

func (r *welfareRepoStub) GetNewUserTrial(_ context.Context, userID int64) (*WelfareNewUserTrial, error) {
	for _, item := range r.trials {
		if item.UserID == userID {
			clone := item
			return &clone, nil
		}
	}
	return nil, ErrWelfareNewUserTrialNotFound
}

func (r *welfareRepoStub) BeginNewUserTrial(_ context.Context, userID int64, quotaAmount float64, clientIP, requestID string, dayStart time.Time, ipActivationLimit int) (*WelfareNewUserTrial, error) {
	now := time.Date(2026, 5, 13, 1, 2, 3, 0, time.UTC)
	for i := range r.trials {
		if r.trials[i].UserID != userID {
			continue
		}
		if r.trials[i].Status == "in_progress" {
			return nil, ErrWelfareNewUserTrialAlreadyInProgress
		}
		if r.trials[i].QuotaUsed >= r.trials[i].QuotaAmount || r.trials[i].Status == "exhausted" {
			return nil, ErrWelfareNewUserTrialExhausted
		}
		r.trials[i].Status = "in_progress"
		r.trials[i].LastRequestID = requestID
		if r.trials[i].FirstStartedAt == nil {
			r.trials[i].FirstStartedAt = &now
		}
		if r.trials[i].ActivatedIP == "" {
			r.trials[i].ActivatedIP = clientIP
		}
		clone := r.trials[i]
		return &clone, nil
	}
	if ipActivationLimit > 0 && clientIP != "" {
		count, err := r.CountNewUserTrialActivationsByIPSince(context.Background(), clientIP, dayStart)
		if err != nil {
			return nil, err
		}
		if count >= ipActivationLimit {
			return nil, ErrWelfareNewUserTrialDailyLimitExceeded
		}
	}
	trial := WelfareNewUserTrial{
		ID:             int64(len(r.trials) + 300),
		UserID:         userID,
		QuotaAmount:    quotaAmount,
		Status:         "in_progress",
		ActivatedIP:    clientIP,
		FirstStartedAt: &now,
		LastRequestID:  requestID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.trials = append(r.trials, trial)
	return &trial, nil
}

func (r *welfareRepoStub) CancelNewUserTrial(_ context.Context, trialID int64, requestID string) error {
	for i := range r.trials {
		if r.trials[i].ID == trialID && r.trials[i].LastRequestID == requestID && r.trials[i].Status == "in_progress" {
			if r.trials[i].QuotaUsed >= r.trials[i].QuotaAmount {
				r.trials[i].Status = "exhausted"
			} else if r.trials[i].FirstSuccessAt != nil {
				r.trials[i].Status = "active"
			} else {
				r.trials[i].Status = "available"
			}
		}
	}
	return nil
}

func (r *welfareRepoStub) ConsumeNewUserTrial(_ context.Context, input WelfareNewUserTrialConsumeInput) (*WelfareNewUserTrial, bool, error) {
	if input.Amount <= 0 {
		trial, err := r.GetNewUserTrial(context.Background(), input.UserID)
		return trial, false, err
	}
	if r.trialUsage == nil {
		r.trialUsage = make(map[string]WelfareNewUserTrialConsumeInput)
	}
	if existing, ok := r.trialUsage[input.RequestID]; ok {
		trial, err := r.GetNewUserTrial(context.Background(), existing.UserID)
		return trial, false, err
	}
	now := welfareTestRewardEnabledAt().Add(25 * time.Hour)
	for i := range r.trials {
		if r.trials[i].ID == input.TrialID && r.trials[i].UserID == input.UserID {
			r.trials[i].QuotaUsed = minFloat64(r.trials[i].QuotaAmount, r.trials[i].QuotaUsed+input.Amount)
			r.trials[i].Status = "active"
			r.trials[i].FirstSuccessAt = &now
			if r.trials[i].QuotaUsed >= r.trials[i].QuotaAmount {
				r.trials[i].Status = "exhausted"
			}
			r.trialUsage[input.RequestID] = input
			r.siteUsageSince += input.Amount
			clone := r.trials[i]
			return &clone, true, nil
		}
	}
	return nil, false, ErrWelfareNewUserTrialNotFound
}

func (r *welfareRepoStub) SumNewUserTrialUsageSince(_ context.Context, _ time.Time) (float64, error) {
	return r.siteUsageSince, nil
}

func (r *welfareRepoStub) CountNewUserTrialActivationsByIPSince(_ context.Context, clientIP string, _ time.Time) (int, error) {
	if r.ipActivations != nil {
		return r.ipActivations[clientIP], nil
	}
	count := 0
	for _, item := range r.trials {
		if item.ActivatedIP == clientIP && item.FirstStartedAt != nil {
			count++
		}
	}
	return count, nil
}

func (r *welfareRepoStub) FirstSuccessfulUsageAt(_ context.Context, userID int64) (*time.Time, error) {
	if r.firstUsageAt != nil {
		firstUsageAt := *r.firstUsageAt
		return &firstUsageAt, nil
	}
	for _, input := range r.trialUsage {
		if input.UserID == userID {
			return nil, nil
		}
	}
	return nil, nil
}

type welfareSettingRepoStub struct {
	SettingRepository
	values map[string]string
}

func (r *welfareSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r.values[key]
	}
	return out, nil
}

type welfareUserRepoStub struct {
	UserRepository
	createdAt time.Time
	updates   []float64
	grants    []float64
}

func (r *welfareUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	createdAt := r.createdAt
	if createdAt.IsZero() {
		createdAt = time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	}
	return &User{ID: id, CreatedAt: createdAt}, nil
}

func (r *welfareUserRepoStub) UpdateBalance(_ context.Context, _ int64, amount float64) error {
	r.updates = append(r.updates, amount)
	return nil
}

func (r *welfareUserRepoStub) AddBalance(_ context.Context, _ int64, amount float64) error {
	r.grants = append(r.grants, amount)
	return nil
}

type welfareRedeemRepoStub struct {
	RedeemCodeRepository
	created []RedeemCode
}

func (r *welfareRedeemRepoStub) Create(_ context.Context, code *RedeemCode) error {
	code.ID = int64(len(r.created) + 700)
	r.created = append(r.created, *code)
	return nil
}

func (r *welfareRedeemRepoStub) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	for i := range r.created {
		if r.created[i].Code == code {
			return &r.created[i], nil
		}
	}
	return nil, ErrRedeemCodeNotFound
}

type welfareAuthInvalidatorStub struct {
	userIDs []int64
}

func (s *welfareAuthInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string)    {}
func (s *welfareAuthInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}
func (s *welfareAuthInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
}

type welfareBalanceInvalidatorStub struct {
	userIDs []int64
}

func (s *welfareBalanceInvalidatorStub) InvalidateUserBalance(_ context.Context, userID int64) error {
	s.userIDs = append(s.userIDs, userID)
	return nil
}

func welfareTestNow(t *testing.T) time.Time {
	t.Helper()
	require.NoError(t, apptimezone.Init("Asia/Shanghai"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	return time.Date(2026, 5, 13, 10, 0, 0, 0, loc)
}

func welfareTxContext() context.Context {
	return dbent.NewTxContext(context.Background(), &dbent.Tx{})
}

func welfareSettingRepo(enabled, daily bool, min, max float64) *welfareSettingRepoStub {
	return &welfareSettingRepoStub{values: map[string]string{
		SettingKeyWelfareEnabled:                            welfareBool(enabled),
		SettingKeyWelfareDailyCheckinEnabled:                welfareBool(daily),
		SettingKeyWelfareRechargeEnabled:                    "true",
		SettingKeyWelfareVIPEnabled:                         "true",
		SettingKeyWelfareFirstRechargeBonusAmount:           "5.00000000",
		SettingKeyWelfareFirstRechargeBonusValidDays:        "0",
		SettingKeyWelfareFirstRechargeBonusStackMonthly:     "false",
		SettingKeyWelfareDailyCheckinRewardMin:              welfareFloat(min),
		SettingKeyWelfareDailyCheckinRewardMax:              welfareFloat(max),
		SettingKeyWelfareDailyCheckinMinAccountAgeHours:     strconv.Itoa(defaultDailyCheckinMinAccountAgeHours),
		SettingKeyWelfareVoucherValidDays:                   "0",
		SettingKeyWelfareDailyCheckinMilestoneEnabled:       "true",
		SettingKeyWelfareDailyCheckinMilestone7Amount:       "7.00000000",
		SettingKeyWelfareDailyCheckinMilestone14Amount:      "14.00000000",
		SettingKeyWelfareDailyCheckinMilestone21Amount:      "21.00000000",
		SettingKeyWelfareDailyCheckinMilestone28Amount:      "28.00000000",
		SettingKeyWelfareNewUserTrialEnabled:                "false",
		SettingKeyWelfareNewUserTrialQuotaAmount:            "0.10000000",
		SettingKeyWelfareNewUserTrialSuccessRewardAmount:    "0.00000000",
		SettingKeyWelfareNewUserTrialSuccessRewardEnabledAt: "",
		SettingKeyWelfareNewUserTrialDailySiteQuotaAmount:   "5.00000000",
		SettingKeyWelfareNewUserTrialDailyIPActivationLimit: "3",
	}}
}

func welfareTrialSettingRepo(enabled bool, quota, siteLimit float64, ipLimit int) *welfareSettingRepoStub {
	repo := welfareSettingRepo(true, true, 1, 1)
	repo.values[SettingKeyWelfareNewUserTrialEnabled] = welfareBool(enabled)
	repo.values[SettingKeyWelfareNewUserTrialQuotaAmount] = welfareFloat(quota)
	repo.values[SettingKeyWelfareNewUserTrialDailySiteQuotaAmount] = welfareFloat(siteLimit)
	repo.values[SettingKeyWelfareNewUserTrialDailyIPActivationLimit] = strconv.Itoa(ipLimit)
	return repo
}

func welfareTrialSettingRepoWithReward(enabled bool, quota, reward, siteLimit float64, ipLimit int) *welfareSettingRepoStub {
	repo := welfareTrialSettingRepo(enabled, quota, siteLimit, ipLimit)
	repo.values[SettingKeyWelfareNewUserTrialSuccessRewardAmount] = welfareFloat(reward)
	if reward > 0 {
		repo.values[SettingKeyWelfareNewUserTrialSuccessRewardEnabledAt] = welfareTestRewardEnabledAt().Format(time.RFC3339)
	}
	return repo
}

func welfareTestRewardEnabledAt() time.Time {
	return time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
}

func welfareSettingRepoWithDailyMinAccountAgeHours(enabled, daily bool, min, max float64, hours int) *welfareSettingRepoStub {
	repo := welfareSettingRepo(enabled, daily, min, max)
	repo.values[SettingKeyWelfareDailyCheckinMinAccountAgeHours] = strconv.Itoa(hours)
	return repo
}

func welfareBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func welfareFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 8, 64)
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func welfareDaily(userID int64, date string, amount float64) WelfareDailyCheckinRecord {
	return WelfareDailyCheckinRecord{
		CheckinDate: date,
		RewardMonth: date[:7],
		UserID:      userID,
		Amount:      amount,
		CreatedAt:   time.Date(2026, 5, 13, 1, 2, 3, 0, time.UTC),
	}
}

func TestWelfareOverviewDisabledDoesNotLoadDailyHistory(t *testing.T) {
	repo := &welfareRepoStub{daily: []WelfareDailyCheckinRecord{welfareDaily(42, "2026-05-13", 1)}}
	svc := NewWelfareService(repo, nil, nil, welfareSettingRepo(false, true, 1, 2), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	got, err := svc.GetOverview(context.Background(), 42)

	require.NoError(t, err)
	require.False(t, got.Enabled)
	require.False(t, got.Modules.DailyCheckin)
	require.NotNil(t, got.DailyCheckin)
	require.False(t, got.DailyCheckin.Enabled)
	require.False(t, got.DailyCheckin.CheckedToday)
	require.Empty(t, got.DailyCheckin.CheckinDates)
}

func TestWelfareDailyCheckinMonthlyStreakStartsFromYesterdayWhenTodayUnchecked(t *testing.T) {
	repo := &welfareRepoStub{daily: []WelfareDailyCheckinRecord{
		welfareDaily(42, "2026-05-10", 1),
		welfareDaily(42, "2026-05-11", 1),
		welfareDaily(42, "2026-05-12", 1),
	}}
	svc := NewWelfareService(repo, nil, nil, welfareSettingRepo(true, true, 1, 2), nil, nil, nil)

	got, err := svc.buildDailyCheckinView(context.Background(), 42, welfareTestNow(t), welfareSettingsStruct(t, true, true, 1, 2))

	require.NoError(t, err)
	require.False(t, got.CheckedToday)
	require.True(t, got.CanClaimToday)
	require.Equal(t, 3, got.CurrentStreakDays)
	require.Equal(t, 3, got.MonthCheckinDays)
	require.Equal(t, welfareReasonNotReached, got.Milestones[0].Reason)
}

func TestWelfareDailyCheckinMonthlyStreakResetsAtMonthBoundary(t *testing.T) {
	repo := &welfareRepoStub{daily: []WelfareDailyCheckinRecord{
		welfareDaily(42, "2026-04-29", 1),
		welfareDaily(42, "2026-04-30", 1),
	}}
	svc := NewWelfareService(repo, nil, nil, welfareSettingRepo(true, true, 1, 2), nil, nil, nil)
	require.NoError(t, apptimezone.Init("Asia/Shanghai"))
	now := time.Date(2026, 5, 1, 9, 0, 0, 0, apptimezone.Location())

	got, err := svc.buildDailyCheckinView(context.Background(), 42, now, welfareSettingsStruct(t, true, true, 1, 2))

	require.NoError(t, err)
	require.Equal(t, 0, got.CurrentStreakDays)
	require.Empty(t, got.CheckinDates)
}

func TestClaimWelfareDailyCheckinGrantsVoucherAndAuditRecord(t *testing.T) {
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{}
	redeemRepo := &welfareRedeemRepoStub{}
	authInvalidator := &welfareAuthInvalidatorStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepo(true, true, 1.25, 1.25), nil, authInvalidator, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	balanceInvalidator := &welfareBalanceInvalidatorStub{}
	svc.billingCacheInvalidator = balanceInvalidator

	got, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.NoError(t, err)
	require.Equal(t, 1.3, got.ClaimedAmount)
	require.Len(t, repo.createdDaily, 1)
	require.Equal(t, "2026-05-13", repo.createdDaily[0].CheckinDate)
	require.Empty(t, userRepo.grants)
	require.Empty(t, userRepo.updates)
	require.Len(t, repo.grantedVouchers, 1)
	require.Equal(t, RedeemTypeDailyCheckin, repo.grantedVouchers[0].SourceType)
	require.Equal(t, int64(100), repo.grantedVouchers[0].SourceID)
	require.Equal(t, 1.3, repo.grantedVouchers[0].Amount)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, RedeemTypeDailyCheckin, redeemRepo.created[0].Type)
	require.Equal(t, StatusUsed, redeemRepo.created[0].Status)
	require.Equal(t, int64(42), *redeemRepo.created[0].UsedBy)
	require.Equal(t, int64(100), repo.attachedDailyClaimID)
	require.Equal(t, int64(700), repo.attachedDailyRedeemID)
	require.Equal(t, []int64{42}, authInvalidator.userIDs)
	require.Equal(t, []int64{42}, balanceInvalidator.userIDs)
}

func TestClaimWelfareDailyCheckinDuplicateReturnsConflict(t *testing.T) {
	repo := &welfareRepoStub{createDailyErr: ErrWelfareDailyCheckinAlreadyClaimed}
	svc := NewWelfareService(repo, &welfareUserRepoStub{}, &welfareRedeemRepoStub{}, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareDailyCheckinAlreadyClaimed)
}

func TestWelfareDailyCheckinRequiresAccountAgeInStatus(t *testing.T) {
	now := welfareTestNow(t)
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: now.Add(-23 * time.Hour)}
	svc := NewWelfareService(repo, userRepo, nil, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)

	got, err := svc.buildDailyCheckinView(context.Background(), 42, now, welfareSettingsStruct(t, true, true, 1, 1))

	require.NoError(t, err)
	require.False(t, got.CanClaimToday)
	require.Equal(t, welfareReasonRegistrationTooNew, got.Reason)
	require.Equal(t, now.Add(time.Hour).UTC().Format(time.RFC3339), got.CanClaimAfter)
	require.Empty(t, repo.createdDaily)
}

func TestClaimWelfareDailyCheckinRequiresAccountAge(t *testing.T) {
	now := welfareTestNow(t)
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: now.Add(-23 * time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareDailyCheckinNotClaimable)
	require.Contains(t, err.Error(), welfareReasonRegistrationTooNew)
	require.Empty(t, repo.createdDaily)
	require.Empty(t, userRepo.grants)
	require.Empty(t, repo.grantedVouchers)
	require.Empty(t, redeemRepo.created)
}

func TestClaimWelfareDailyCheckinAllowsAccountAt24Hours(t *testing.T) {
	now := welfareTestNow(t)
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: now.Add(-24 * time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return now }

	got, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.NoError(t, err)
	require.Equal(t, 1.0, got.ClaimedAmount)
	require.Len(t, repo.createdDaily, 1)
	require.Empty(t, userRepo.grants)
	require.Len(t, repo.grantedVouchers, 1)
}

func TestClaimWelfareDailyCheckinUsesConfiguredMinimumAccountAgeHours(t *testing.T) {
	now := welfareTestNow(t)
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: now.Add(-5 * time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepoWithDailyMinAccountAgeHours(true, true, 1, 1, 6), nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareDailyCheckinNotClaimable)
	require.Empty(t, repo.createdDaily)

	repo = &welfareRepoStub{}
	userRepo = &welfareUserRepoStub{createdAt: now.Add(-5 * time.Hour)}
	redeemRepo = &welfareRedeemRepoStub{}
	svc = NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepoWithDailyMinAccountAgeHours(true, true, 1, 1, 0), nil, nil, nil)
	svc.now = func() time.Time { return now }

	got, err := svc.ClaimDailyCheckin(welfareTxContext(), 42)

	require.NoError(t, err)
	require.Equal(t, 1.0, got.ClaimedAmount)
	require.Len(t, repo.createdDaily, 1)
}

func TestClaimWelfareMilestoneRequiresReachedStreak(t *testing.T) {
	repo := &welfareRepoStub{daily: []WelfareDailyCheckinRecord{
		welfareDaily(42, "2026-05-11", 1),
		welfareDaily(42, "2026-05-12", 1),
		welfareDaily(42, "2026-05-13", 1),
	}}
	svc := NewWelfareService(repo, &welfareUserRepoStub{}, &welfareRedeemRepoStub{}, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.ClaimDailyCheckinMilestone(welfareTxContext(), 42, 7)

	require.ErrorIs(t, err, ErrWelfareCheckinMilestoneNotClaimable)
}

func TestClaimWelfareMilestoneGrantsVoucherAndAuditRecord(t *testing.T) {
	repo := &welfareRepoStub{}
	for day := 7; day >= 1; day-- {
		repo.daily = append(repo.daily, welfareDaily(42, fmt.Sprintf("2026-05-%02d", 14-day), 1))
	}
	userRepo := &welfareUserRepoStub{}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	got, err := svc.ClaimDailyCheckinMilestone(welfareTxContext(), 42, 7)

	require.NoError(t, err)
	require.Equal(t, 7.0, got.ClaimedAmount)
	require.Equal(t, 7, got.Milestone.Day)
	require.True(t, got.Milestone.Claimed)
	require.Empty(t, userRepo.grants)
	require.Len(t, repo.grantedVouchers, 1)
	require.Equal(t, RedeemTypeCheckinMilestone, repo.grantedVouchers[0].SourceType)
	require.Equal(t, int64(200), repo.grantedVouchers[0].SourceID)
	require.Equal(t, 7.0, repo.grantedVouchers[0].Amount)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, RedeemTypeCheckinMilestone, redeemRepo.created[0].Type)
	require.Equal(t, int64(200), repo.attachedMilestoneClaimID)
	require.Equal(t, int64(700), repo.attachedMilestoneRedeemID)
}

func TestWelfareMilestoneDisabledHidesMilestonesAndRejectsClaim(t *testing.T) {
	repo := &welfareRepoStub{}
	for day := 7; day >= 1; day-- {
		repo.daily = append(repo.daily, welfareDaily(42, fmt.Sprintf("2026-05-%02d", 14-day), 1))
	}
	settings := welfareSettingRepo(true, true, 1, 1)
	settings.values[SettingKeyWelfareDailyCheckinMilestoneEnabled] = "false"
	svc := NewWelfareService(repo, &welfareUserRepoStub{}, &welfareRedeemRepoStub{}, settings, nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	status, err := svc.GetDailyCheckin(context.Background(), 42)
	require.NoError(t, err)
	require.False(t, status.MilestoneEnabled)
	require.Empty(t, status.Milestones)

	_, err = svc.ClaimDailyCheckinMilestone(welfareTxContext(), 42, 7)
	require.ErrorIs(t, err, ErrWelfareCheckinMilestoneNotClaimable)
}

func TestClaimWelfareMilestoneDuplicateReturnsConflict(t *testing.T) {
	repo := &welfareRepoStub{createMilestoneErr: ErrWelfareCheckinMilestoneAlreadyClaimed}
	for day := 7; day >= 1; day-- {
		repo.daily = append(repo.daily, welfareDaily(42, fmt.Sprintf("2026-05-%02d", 14-day), 1))
	}
	svc := NewWelfareService(repo, &welfareUserRepoStub{}, &welfareRedeemRepoStub{}, welfareSettingRepo(true, true, 1, 1), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.ClaimDailyCheckinMilestone(welfareTxContext(), 42, 7)

	require.ErrorIs(t, err, ErrWelfareCheckinMilestoneAlreadyClaimed)
}

func TestRandomRewardAmountUsesOneDecimalStep(t *testing.T) {
	amount, err := randomRewardAmount(1.25, 1.25)

	require.NoError(t, err)
	require.Equal(t, 1.3, amount)
}

func TestRandomRewardAmountAvoidsZeroWhenConfiguredMaxPositive(t *testing.T) {
	amount, err := randomRewardAmount(0, 0.5)

	require.NoError(t, err)
	require.True(t, amount >= 0.1 && amount <= 0.5)
	require.Equal(t, math.Round(amount*10), amount*10)
}

func TestWelfareSettingsNormalizeMaxBelowMin(t *testing.T) {
	svc := NewWelfareService(nil, nil, nil, welfareSettingRepo(true, true, 5, 3), nil, nil, nil)

	got, err := svc.getSettings(context.Background())

	require.NoError(t, err)
	require.Equal(t, 5.0, got.RewardMin)
	require.Equal(t, 5.0, got.RewardMax)
}

func TestBeginNewUserTrialDisabled(t *testing.T) {
	repo := &welfareRepoStub{}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(false, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")

	require.ErrorIs(t, err, ErrWelfareNewUserTrialDisabled)
	require.Empty(t, repo.trials)
}

func TestBeginNewUserTrialCreatesSingleUsePool(t *testing.T) {
	repo := &welfareRepoStub{}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	session, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")

	require.NoError(t, err)
	require.Equal(t, int64(42), session.UserID)
	require.Equal(t, 0.1, session.QuotaLeft)
	require.Len(t, repo.trials, 1)
	require.Equal(t, "in_progress", repo.trials[0].Status)
}

func TestConsumeNewUserTrialDeductsPoolOnly(t *testing.T) {
	repo := &welfareRepoStub{}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	session, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	err = svc.ConsumeNewUserTrial(context.Background(), session, "usage-1", 0.04, "claude-sonnet", 9)

	require.NoError(t, err)
	require.Len(t, repo.trials, 1)
	require.Equal(t, 0.04, repo.trials[0].QuotaUsed)
	require.Equal(t, "active", repo.trials[0].Status)
	require.Equal(t, 0.04, repo.siteUsageSince)
}

func TestConsumeNewUserTrialNotifiesUnclaimedSuccessReward(t *testing.T) {
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	systemRepo := newFakeTicketRepo()
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, nil, nil)
	svc.SetSystemTicketService(NewSystemTicketService(systemRepo))
	svc.now = func() time.Time { return welfareTestNow(t) }
	session, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	err = svc.ConsumeNewUserTrial(context.Background(), session, "usage-1", 0.04, "claude-sonnet", 9)

	require.NoError(t, err)
	notification := requireSystemTicketNotification(t, systemRepo, 42, SystemTicketEventWelfareFirstAPIUnclaimed, "welfare_first_api_unclaimed:42")
	require.Equal(t, float64(42), notification.Metadata["user_id"])
	require.Equal(t, float64(session.TrialID), notification.Metadata["trial_id"])
	require.Equal(t, 2.5, notification.Metadata["reward_amount"])
	require.Equal(t, "claude-sonnet", notification.Metadata["model"])
	require.Equal(t, float64(9), notification.Metadata["api_key_id"])
}

func TestClaimNewUserTrialSuccessRewardGrantsConfiguredRewardOnce(t *testing.T) {
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	balanceInvalidator := &welfareBalanceInvalidatorStub{}
	authInvalidator := &welfareAuthInvalidatorStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, authInvalidator, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	svc.billingCacheInvalidator = balanceInvalidator
	session, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	err = svc.ConsumeNewUserTrial(welfareTxContext(), session, "usage-1", 0.04, "claude-sonnet", 9)
	require.NoError(t, err)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)

	result, err := svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)
	require.NoError(t, err)
	require.Equal(t, 2.5, result.ClaimedAmount)
	require.NotNil(t, result.NewUserTrial)
	require.True(t, result.NewUserTrial.SuccessRewardClaimed)
	require.False(t, result.NewUserTrial.SuccessRewardClaimable)
	require.Equal(t, []float64{2.5}, userRepo.grants)
	require.Empty(t, userRepo.updates)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, RedeemTypeNewUserReward, redeemRepo.created[0].Type)
	require.Equal(t, "NUTR16", redeemRepo.created[0].Code)
	require.Equal(t, StatusUsed, redeemRepo.created[0].Status)
	require.Equal(t, int64(42), *redeemRepo.created[0].UsedBy)
	require.Equal(t, []int64{42}, authInvalidator.userIDs)
	require.Equal(t, []int64{42}, balanceInvalidator.userIDs)

	_, err = svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)
	require.ErrorIs(t, err, ErrWelfareNewUserTrialRewardClaimed)
	require.Equal(t, []float64{2.5}, userRepo.grants)
	require.Len(t, redeemRepo.created, 1)
}

func TestClaimNewUserTrialSuccessRewardRequiresSuccessfulTrialCall(t *testing.T) {
	repo := &welfareRepoStub{}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	_, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	_, err = svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareNewUserTrialNotAvailable)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)
}

func TestClaimNewUserTrialSuccessRewardAllowsExistingSuccessfulUsage(t *testing.T) {
	firstUsageAt := welfareTestRewardEnabledAt().Add(2 * time.Hour)
	repo := &welfareRepoStub{firstUsageAt: &firstUsageAt}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	result, err := svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)

	require.NoError(t, err)
	require.Equal(t, 2.5, result.ClaimedAmount)
	require.NotNil(t, result.NewUserTrial)
	require.True(t, result.NewUserTrial.SuccessRewardClaimed)
	require.Equal(t, []float64{2.5}, userRepo.grants)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, "NUTR16", redeemRepo.created[0].Code)
	require.Equal(t, "new user trial success reward", redeemRepo.created[0].Notes)
}

func TestClaimNewUserTrialSuccessRewardRejectsUserRegisteredBeforeRewardEnabled(t *testing.T) {
	firstUsageAt := welfareTestRewardEnabledAt().Add(2 * time.Hour)
	repo := &welfareRepoStub{firstUsageAt: &firstUsageAt}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(-time.Minute)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareNewUserTrialNotAvailable)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)
}

func TestClaimNewUserTrialSuccessRewardRejectsFirstUsageBeforeRewardEnabled(t *testing.T) {
	firstUsageAt := welfareTestRewardEnabledAt().Add(-time.Minute)
	repo := &welfareRepoStub{firstUsageAt: &firstUsageAt}
	userRepo := &welfareUserRepoStub{createdAt: welfareTestRewardEnabledAt().Add(time.Hour)}
	redeemRepo := &welfareRedeemRepoStub{}
	svc := NewWelfareService(repo, userRepo, redeemRepo, welfareTrialSettingRepoWithReward(true, 0.1, 2.5, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.ClaimNewUserTrialSuccessReward(welfareTxContext(), 42)

	require.ErrorIs(t, err, ErrWelfareNewUserTrialNotAvailable)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)
}

func TestConsumeNewUserTrialCapsCostAboveRemainingQuota(t *testing.T) {
	repo := &welfareRepoStub{}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	session, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	err = svc.ConsumeNewUserTrial(context.Background(), session, "usage-1", 0.11, "claude-sonnet", 9)

	require.NoError(t, err)
	require.Equal(t, 0.1, repo.trials[0].QuotaUsed)
	require.Equal(t, "exhausted", repo.trials[0].Status)
}

func TestBeginNewUserTrialRejectsConcurrentRequest(t *testing.T) {
	repo := &welfareRepoStub{}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }
	_, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")
	require.NoError(t, err)

	_, err = svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")

	require.ErrorIs(t, err, ErrWelfareNewUserTrialAlreadyInProgress)
}

func TestBeginNewUserTrialRejectsDailyLimits(t *testing.T) {
	repo := &welfareRepoStub{siteUsageSince: 5}
	svc := NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err := svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")

	require.ErrorIs(t, err, ErrWelfareNewUserTrialDailyLimitExceeded)

	repo = &welfareRepoStub{ipActivations: map[string]int{"203.0.113.8": 3}}
	svc = NewWelfareService(repo, nil, nil, welfareTrialSettingRepo(true, 0.1, 5, 3), nil, nil, nil)
	svc.now = func() time.Time { return welfareTestNow(t) }

	_, err = svc.BeginNewUserTrial(context.Background(), 42, "203.0.113.8")

	require.ErrorIs(t, err, ErrWelfareNewUserTrialDailyLimitExceeded)
}

func welfareSettingsStruct(t *testing.T, enabled, daily bool, min, max float64) welfareSettings {
	t.Helper()
	got, err := NewWelfareService(nil, nil, nil, welfareSettingRepo(enabled, daily, min, max), nil, nil, nil).getSettings(context.Background())
	require.NoError(t, err)
	return got
}
