package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	LeaderboardMinimumAccountAgeDaysDefault = 7
	LeaderboardMinimumAccountAgeDaysMax     = 3650
	leaderboardMinimumAccountAge            = LeaderboardMinimumAccountAgeDaysDefault * 24 * time.Hour
)

var (
	ErrLeaderboardAccountTooNew = infraerrors.Forbidden(
		"LEADERBOARD_ACCOUNT_TOO_NEW",
		"leaderboard account-age requirement is not met",
	)
	errLeaderboardAccessUnavailable = infraerrors.InternalServer(
		"LEADERBOARD_ACCESS_UNAVAILABLE",
		"leaderboard access check is unavailable",
	)
)

// EnsureLeaderboardAccess rejects accounts that have not reached the minimum age.
func (s *UsageService) EnsureLeaderboardAccess(ctx context.Context, userID int64) error {
	if s == nil || s.userRepo == nil {
		return errLeaderboardAccessUnavailable
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	minimumAgeDays := s.leaderboardMinimumAccountAgeDays(ctx)
	if user == nil || !hasLeaderboardAccountAgeForDays(user.CreatedAt, time.Now(), minimumAgeDays) {
		return ErrLeaderboardAccountTooNew
	}
	return nil
}

func hasLeaderboardAccountAge(createdAt, now time.Time) bool {
	return hasLeaderboardAccountAgeForDays(createdAt, now, LeaderboardMinimumAccountAgeDaysDefault)
}

func hasLeaderboardAccountAgeForDays(createdAt, now time.Time, minimumAgeDays int) bool {
	if createdAt.IsZero() {
		return false
	}
	return !now.Before(createdAt.Add(time.Duration(minimumAgeDays) * 24 * time.Hour))
}

func normalizeLeaderboardMinimumAccountAgeDays(days int) int {
	if days < 0 || days > LeaderboardMinimumAccountAgeDaysMax {
		return LeaderboardMinimumAccountAgeDaysDefault
	}
	return days
}

func parseLeaderboardMinimumAccountAgeDays(raw string) int {
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return LeaderboardMinimumAccountAgeDaysDefault
	}
	return normalizeLeaderboardMinimumAccountAgeDays(days)
}

func (s *UsageService) leaderboardMinimumAccountAgeDays(ctx context.Context) int {
	if s == nil || s.settingRepo == nil {
		return LeaderboardMinimumAccountAgeDaysDefault
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{SettingKeyLeaderboardMinAccountAgeDays})
	if err != nil {
		return LeaderboardMinimumAccountAgeDaysDefault
	}
	return parseLeaderboardMinimumAccountAgeDays(values[SettingKeyLeaderboardMinAccountAgeDays])
}
