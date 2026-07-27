package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type leaderboardAccessUserRepo struct {
	UserRepository
	user  *User
	err   error
	calls int
}

type leaderboardAccessSettingRepo struct {
	SettingRepository
	value string
	err   error
}

func (r *leaderboardAccessSettingRepo) GetValue(context.Context, string) (string, error) {
	return r.value, r.err
}

func (r *leaderboardAccessSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return map[string]string{SettingKeyLeaderboardMinAccountAgeDays: r.value}, nil
}

func (r *leaderboardAccessUserRepo) GetByID(context.Context, int64) (*User, error) {
	r.calls++
	return r.user, r.err
}

func TestLeaderboardAccessAccountAgeBoundary(t *testing.T) {
	now := time.Date(2026, 7, 27, 21, 0, 0, 0, time.UTC)

	require.False(t, hasLeaderboardAccountAge(time.Time{}, now))
	require.False(t, hasLeaderboardAccountAge(now.Add(-leaderboardMinimumAccountAge+time.Nanosecond), now))
	require.True(t, hasLeaderboardAccountAge(now.Add(-leaderboardMinimumAccountAge), now))
	require.True(t, hasLeaderboardAccountAge(now.Add(-leaderboardMinimumAccountAge-time.Nanosecond), now))
}

func TestUsageServiceEnsureLeaderboardAccess(t *testing.T) {
	t.Run("rejects account younger than seven days", func(t *testing.T) {
		repo := &leaderboardAccessUserRepo{user: &User{ID: 42, CreatedAt: time.Now().Add(-6 * 24 * time.Hour)}}
		svc := NewUsageService(nil, repo, nil, nil)

		err := svc.EnsureLeaderboardAccess(context.Background(), 42)

		require.ErrorIs(t, err, ErrLeaderboardAccountTooNew)
		require.Equal(t, 1, repo.calls)
	})

	t.Run("allows account at least seven days old", func(t *testing.T) {
		repo := &leaderboardAccessUserRepo{user: &User{ID: 42, CreatedAt: time.Now().Add(-leaderboardMinimumAccountAge)}}
		svc := NewUsageService(nil, repo, nil, nil)

		require.NoError(t, svc.EnsureLeaderboardAccess(context.Background(), 42))
		require.Equal(t, 1, repo.calls)
	})

	t.Run("fails closed when created at is missing", func(t *testing.T) {
		repo := &leaderboardAccessUserRepo{user: &User{ID: 42}}
		svc := NewUsageService(nil, repo, nil, nil)

		require.ErrorIs(t, svc.EnsureLeaderboardAccess(context.Background(), 42), ErrLeaderboardAccountTooNew)
	})

	t.Run("propagates user repository errors", func(t *testing.T) {
		repoErr := errors.New("user lookup failed")
		repo := &leaderboardAccessUserRepo{err: repoErr}
		svc := NewUsageService(nil, repo, nil, nil)

		require.ErrorIs(t, svc.EnsureLeaderboardAccess(context.Background(), 42), repoErr)
	})
}

func TestUsageServiceEnsureLeaderboardAccessUsesConfiguredDays(t *testing.T) {
	t.Run("uses configured minimum days", func(t *testing.T) {
		userRepo := &leaderboardAccessUserRepo{user: &User{ID: 42, CreatedAt: time.Now().Add(-2 * 24 * time.Hour)}}
		settingRepo := &leaderboardAccessSettingRepo{value: "2"}
		svc := NewUsageService(nil, userRepo, nil, nil)
		svc.SetLeaderboardRewardDependencies(settingRepo, nil)

		require.NoError(t, svc.EnsureLeaderboardAccess(context.Background(), 42))
	})

	t.Run("allows zero configured days", func(t *testing.T) {
		userRepo := &leaderboardAccessUserRepo{user: &User{ID: 42, CreatedAt: time.Now()}}
		settingRepo := &leaderboardAccessSettingRepo{value: "0"}
		svc := NewUsageService(nil, userRepo, nil, nil)
		svc.SetLeaderboardRewardDependencies(settingRepo, nil)

		require.NoError(t, svc.EnsureLeaderboardAccess(context.Background(), 42))
	})

	t.Run("invalid setting falls back to seven days", func(t *testing.T) {
		userRepo := &leaderboardAccessUserRepo{user: &User{ID: 42, CreatedAt: time.Now().Add(-6 * 24 * time.Hour)}}
		settingRepo := &leaderboardAccessSettingRepo{value: "invalid"}
		svc := NewUsageService(nil, userRepo, nil, nil)
		svc.SetLeaderboardRewardDependencies(settingRepo, nil)

		require.ErrorIs(t, svc.EnsureLeaderboardAccess(context.Background(), 42), ErrLeaderboardAccountTooNew)
	})
}
