package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type subscriptionQuotaWindowRepo struct {
	userSubRepoNoop

	sub                *UserSubscription
	activatedAt        time.Time
	dailyResetAt       time.Time
	monthlyExpectedAt  *time.Time
	monthlyResetAt     time.Time
	monthlyResetCalled bool
}

func (r *subscriptionQuotaWindowRepo) GetByID(_ context.Context, id int64) (*UserSubscription, error) {
	if r.sub == nil || r.sub.ID != id {
		return nil, ErrSubscriptionNotFound
	}
	cp := *r.sub
	return &cp, nil
}

func (r *subscriptionQuotaWindowRepo) ActivateWindowsIfUninitialized(_ context.Context, _ int64, start time.Time) error {
	r.activatedAt = start
	return nil
}

func (r *subscriptionQuotaWindowRepo) ResetDailyUsageIfWindowStart(_ context.Context, _ int64, _ *time.Time, newWindowStart time.Time) error {
	r.dailyResetAt = newWindowStart
	if r.sub != nil {
		r.sub.DailyUsageUSD = 0
		r.sub.DailyWindowStart = &newWindowStart
	}
	return nil
}

func (r *subscriptionQuotaWindowRepo) ResetWeeklyUsageIfWindowStart(context.Context, int64, *time.Time, time.Time) error {
	return nil
}

func (r *subscriptionQuotaWindowRepo) ResetMonthlyUsageIfWindowStart(_ context.Context, _ int64, expectedWindowStart *time.Time, newWindowStart time.Time) error {
	r.monthlyResetCalled = true
	r.monthlyExpectedAt = expectedWindowStart
	r.monthlyResetAt = newWindowStart
	return nil
}

func TestSubscriptionQuotaWindows_DelayedActivationUsesExactTime(t *testing.T) {
	startsAt := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	activatedAt := time.Date(2026, 7, 10, 23, 30, 0, 0, time.UTC)
	repo := &subscriptionQuotaWindowRepo{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return activatedAt }
	sub := &UserSubscription{ID: 1, StartsAt: startsAt, ExpiresAt: startsAt.Add(45 * 24 * time.Hour)}

	require.NoError(t, svc.CheckAndActivateWindow(context.Background(), sub))
	require.Equal(t, activatedAt, repo.activatedAt)
}

func TestSubscriptionQuotaWindows_LegacyAnchorHonorsTermEnd(t *testing.T) {
	windowStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	startsAt := windowStart.Add(23*time.Hour + 30*time.Minute)
	repo := &subscriptionQuotaWindowRepo{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return startsAt.Add(30 * 24 * time.Hour) }
	sub := &UserSubscription{
		ID:                 1,
		Status:             SubscriptionStatusActive,
		StartsAt:           startsAt,
		ExpiresAt:          startsAt.Add(30 * 24 * time.Hour),
		MonthlyWindowStart: &windowStart,
		MonthlyUsageUSD:    12,
	}

	resetAt, ok := sub.automaticWindowStartAt(sub.MonthlyWindowStart, 30*24*time.Hour, sub.ExpiresAt)
	require.False(t, ok)
	require.True(t, resetAt.IsZero())
	require.NoError(t, svc.CheckAndResetWindows(context.Background(), sub))
	require.False(t, repo.monthlyResetCalled)

	needsMaintenance, err := svc.ValidateAndCheckLimits(sub, &Group{})
	require.ErrorIs(t, err, ErrSubscriptionExpired)
	require.False(t, needsMaintenance)

	subs := []UserSubscription{*sub}
	normalizeExpiredWindowsAt(subs, sub.ExpiresAt)
	require.Equal(t, 12.0, subs[0].MonthlyUsageUSD)
	require.NotNil(t, subs[0].MonthlyWindowStart)
}

func TestSubscriptionQuotaWindows_PartialFinalPeriodResetsOnTermAnchor(t *testing.T) {
	legacyWindowStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	startsAt := legacyWindowStart.Add(23*time.Hour + 30*time.Minute)
	now := startsAt.Add(30 * 24 * time.Hour)
	repo := &subscriptionQuotaWindowRepo{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }
	sub := &UserSubscription{
		ID:                 2,
		StartsAt:           startsAt,
		ExpiresAt:          startsAt.Add(45 * 24 * time.Hour),
		MonthlyWindowStart: &legacyWindowStart,
		MonthlyUsageUSD:    12,
	}

	require.NoError(t, svc.CheckAndResetWindows(context.Background(), sub))
	require.True(t, repo.monthlyResetCalled)
	require.Equal(t, legacyWindowStart, *repo.monthlyExpectedAt)
	require.Equal(t, startsAt.Add(30*24*time.Hour), repo.monthlyResetAt)
	require.Zero(t, sub.MonthlyUsageUSD)
}

func TestSubscriptionQuotaWindows_ManualResetUsesExactTime(t *testing.T) {
	resetAt := time.Date(2026, 7, 1, 10, 37, 42, 123, time.UTC)
	repo := &subscriptionQuotaWindowRepo{sub: &UserSubscription{ID: 3, UserID: 10, GroupID: 20}}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return resetAt }

	result, err := svc.AdminResetQuota(context.Background(), 3, true, false, false)

	require.NoError(t, err)
	require.Equal(t, resetAt, repo.dailyResetAt)
	require.NotNil(t, result.DailyWindowStart)
	require.Equal(t, resetAt, *result.DailyWindowStart)
}

func TestSubscriptionQuotaWindows_RenewalStartsAtExactTermTime(t *testing.T) {
	startsAt := time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC)
	renewed := renewedSubscriptionTerm(&UserSubscription{}, "", startsAt, startsAt.Add(24*time.Hour))

	require.Equal(t, startsAt, *renewed.DailyWindowStart)
	require.Equal(t, startsAt, *renewed.WeeklyWindowStart)
	require.Equal(t, startsAt, *renewed.MonthlyWindowStart)
}

func TestSubscriptionQuotaWindows_PreservesLaterManualAnchor(t *testing.T) {
	startsAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	manualAnchor := startsAt.Add(5 * 24 * time.Hour)
	sub := &UserSubscription{StartsAt: startsAt, ExpiresAt: startsAt.Add(100 * 24 * time.Hour)}

	resetAt, ok := sub.automaticWindowStartAt(&manualAnchor, 30*24*time.Hour, manualAnchor.Add(65*24*time.Hour))

	require.True(t, ok)
	require.Equal(t, manualAnchor.Add(60*24*time.Hour), resetAt)
}
