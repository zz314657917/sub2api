package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

const (
	leaderboardRewardReasonEligible        = "eligible"
	leaderboardRewardReasonDisabled        = "disabled"
	leaderboardRewardReasonSettling        = "settling"
	leaderboardRewardReasonThresholdNotMet = "threshold_not_met"
	leaderboardRewardReasonNotRanked       = "not_ranked"
	leaderboardRewardReasonNotTopThree     = "not_top_three"
	leaderboardRewardReasonZeroReward      = "zero_reward"
	leaderboardRewardReasonAlreadyClaimed  = "already_claimed"

	leaderboardRewardSettlementDelay = 30 * time.Minute
)

var (
	ErrLeaderboardDailyRewardClaimNotFound  = infraerrors.NotFound("LEADERBOARD_DAILY_REWARD_CLAIM_NOT_FOUND", "leaderboard daily reward claim not found")
	ErrLeaderboardDailyRewardAlreadyClaimed = infraerrors.Conflict("LEADERBOARD_DAILY_REWARD_ALREADY_CLAIMED", "leaderboard daily reward already claimed")
	ErrLeaderboardDailyRewardNotEligible    = infraerrors.Forbidden("LEADERBOARD_DAILY_REWARD_NOT_ELIGIBLE", "leaderboard daily reward is not claimable")
	ErrLeaderboardDailyRewardUnavailable    = infraerrors.ServiceUnavailable("LEADERBOARD_DAILY_REWARD_UNAVAILABLE", "leaderboard daily reward service is unavailable")
)

type leaderboardDailyRewardSettings struct {
	Enabled            bool
	MinTotalActualCost float64
	RankAmounts        map[int]float64
}

// LeaderboardDailyRewardClaim records one user's settlement claim for a reward date.
type LeaderboardDailyRewardClaim struct {
	ID              int64
	RewardDate      string
	UserID          int64
	Rank            int
	Amount          float64
	TotalActualCost float64
	RedeemCodeID    *int64
	CreatedAt       time.Time
}

// LeaderboardDailyRewardClaimStore is implemented by repositories that persist reward claims.
type LeaderboardDailyRewardClaimStore interface {
	GetLeaderboardDailyRewardClaim(ctx context.Context, rewardDate string, userID int64) (*LeaderboardDailyRewardClaim, error)
	CreateLeaderboardDailyRewardClaim(ctx context.Context, claim *LeaderboardDailyRewardClaim) error
	AttachLeaderboardDailyRewardClaimRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error
}

// LeaderboardDailyRewardClaimResult represents a successful reward claim.
type LeaderboardDailyRewardClaimResult struct {
	DailyRewards  *usagestats.LeaderboardDailyRewards `json:"daily_rewards"`
	ClaimedAmount float64                             `json:"claimed_amount"`
}

type leaderboardRewardBalanceGrantRepository interface {
	AddBalance(ctx context.Context, id int64, amount float64) error
}

// GetLeaderboardDailyRewards returns yesterday's leaderboard reward status.
// The reward settlement window is intentionally based on the server timezone,
// so clients cannot change reward dates or ranking windows by changing timezone.
func (s *UsageService) GetLeaderboardDailyRewards(ctx context.Context, userID int64, _ string) (*usagestats.LeaderboardDailyRewards, error) {
	return s.getLeaderboardDailyRewards(ctx, userID, timezone.Now())
}

func (s *UsageService) getLeaderboardDailyRewards(ctx context.Context, userID int64, now time.Time) (*usagestats.LeaderboardDailyRewards, error) {
	start, end, rewardDate, settlementTZ, claimAvailableAt := leaderboardRewardWindow(now)
	leaderboard, err := s.GetUserLeaderboard(ctx, start, end, 3, userID)
	if err != nil {
		return nil, err
	}
	if leaderboard == nil {
		leaderboard = &usagestats.UserLeaderboardResponse{}
	}

	settings, err := s.getLeaderboardDailyRewardSettings(ctx)
	if err != nil {
		return nil, err
	}

	status := &usagestats.LeaderboardDailyRewards{
		RewardDate:               rewardDate,
		SettlementTimezone:       settlementTZ,
		SettlementReady:          !now.Before(claimAvailableAt),
		ClaimAvailableAt:         claimAvailableAt.Format(time.RFC3339),
		Enabled:                  settings.Enabled,
		MinTotalActualCost:       settings.MinTotalActualCost,
		YesterdayTotalActualCost: leaderboard.TotalActualCost,
		ThresholdMet:             leaderboard.TotalActualCost > settings.MinTotalActualCost,
		Rewards:                  leaderboardDailyRewardTiers(settings),
		Reason:                   leaderboardRewardReasonDisabled,
	}

	if leaderboard.CurrentUserEntry != nil {
		status.CurrentUserRank = leaderboard.CurrentUserEntry.Rank
		if amount, ok := settings.RankAmounts[int(leaderboard.CurrentUserEntry.Rank)]; ok {
			status.CurrentUserRewardAmount = amount
		}
	}

	if store := s.leaderboardDailyRewardClaimStore(); store != nil {
		claim, err := store.GetLeaderboardDailyRewardClaim(ctx, rewardDate, userID)
		if err != nil && !errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
			return nil, err
		}
		status.Claimed = claim != nil
	}

	status.CanClaim, status.Reason = resolveLeaderboardDailyRewardClaimState(status)
	return status, nil
}

// ClaimLeaderboardDailyReward grants yesterday's top-3 balance reward to the user.
func (s *UsageService) ClaimLeaderboardDailyReward(ctx context.Context, userID int64, _ string) (*LeaderboardDailyRewardClaimResult, error) {
	return s.claimLeaderboardDailyReward(ctx, userID, timezone.Now())
}

func (s *UsageService) claimLeaderboardDailyReward(ctx context.Context, userID int64, now time.Time) (*LeaderboardDailyRewardClaimResult, error) {
	if s.userRepo == nil || s.redeemRepo == nil || s.leaderboardDailyRewardClaimStore() == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}

	status, err := s.getLeaderboardDailyRewards(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if !status.CanClaim {
		if status.Reason == leaderboardRewardReasonAlreadyClaimed {
			return nil, ErrLeaderboardDailyRewardAlreadyClaimed
		}
		return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": status.Reason})
	}

	claim := &LeaderboardDailyRewardClaim{
		RewardDate:      status.RewardDate,
		UserID:          userID,
		Rank:            int(status.CurrentUserRank),
		Amount:          status.CurrentUserRewardAmount,
		TotalActualCost: status.YesterdayTotalActualCost,
	}

	claimCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin leaderboard daily reward transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		claimCtx = dbent.NewTxContext(ctx, tx)
	}

	store := s.leaderboardDailyRewardClaimStore()
	if err := store.CreateLeaderboardDailyRewardClaim(claimCtx, claim); err != nil {
		if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) {
			return nil, ErrLeaderboardDailyRewardAlreadyClaimed
		}
		return nil, err
	}

	usedAt := now.UTC()
	redeemCode := &RedeemCode{
		Code:   leaderboardRewardRedeemCode(status.RewardDate, userID),
		Type:   RedeemTypeLeaderboardReward,
		Value:  status.CurrentUserRewardAmount,
		Status: StatusUsed,
		UsedBy: &userID,
		UsedAt: &usedAt,
		Notes:  fmt.Sprintf("leaderboard daily reward %s rank %d", status.RewardDate, status.CurrentUserRank),
	}
	if err := s.redeemRepo.Create(claimCtx, redeemCode); err != nil {
		return nil, fmt.Errorf("create leaderboard reward audit record: %w", err)
	}
	if err := grantLeaderboardRewardBalance(claimCtx, s.userRepo, userID, status.CurrentUserRewardAmount); err != nil {
		return nil, fmt.Errorf("update leaderboard reward balance: %w", err)
	}
	if err := store.AttachLeaderboardDailyRewardClaimRedeemCode(claimCtx, claim.ID, redeemCode.ID); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit leaderboard daily reward transaction: %w", err)
		}
	}

	status.Claimed = true
	status.CanClaim = false
	status.Reason = leaderboardRewardReasonAlreadyClaimed
	s.invalidateUsageCaches(ctx, userID, true)

	return &LeaderboardDailyRewardClaimResult{
		DailyRewards:  status,
		ClaimedAmount: status.CurrentUserRewardAmount,
	}, nil
}

func (s *UsageService) getLeaderboardDailyRewardSettings(ctx context.Context) (leaderboardDailyRewardSettings, error) {
	result := leaderboardDailyRewardSettings{
		RankAmounts: map[int]float64{1: 0, 2: 0, 3: 0},
	}
	if s.settingRepo == nil {
		return result, nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyLeaderboardDailyRewardEnabled,
		SettingKeyLeaderboardDailyRewardMinTotalActualCost,
		SettingKeyLeaderboardDailyRewardRank1Amount,
		SettingKeyLeaderboardDailyRewardRank2Amount,
		SettingKeyLeaderboardDailyRewardRank3Amount,
	})
	if err != nil {
		return result, fmt.Errorf("get leaderboard daily reward settings: %w", err)
	}
	result.Enabled = values[SettingKeyLeaderboardDailyRewardEnabled] == "true"
	result.MinTotalActualCost = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardMinTotalActualCost], 0)
	result.RankAmounts[1] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank1Amount], 0)
	result.RankAmounts[2] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank2Amount], 0)
	result.RankAmounts[3] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank3Amount], 0)
	return result, nil
}

func (s *UsageService) leaderboardDailyRewardClaimStore() LeaderboardDailyRewardClaimStore {
	store, _ := s.usageRepo.(LeaderboardDailyRewardClaimStore)
	return store
}

func grantLeaderboardRewardBalance(ctx context.Context, userRepo UserRepository, userID int64, amount float64) error {
	if grantRepo, ok := userRepo.(leaderboardRewardBalanceGrantRepository); ok {
		return grantRepo.AddBalance(ctx, userID, amount)
	}
	return userRepo.UpdateBalance(ctx, userID, amount)
}

func leaderboardRewardWindow(now time.Time) (time.Time, time.Time, string, string, time.Time) {
	loc := timezone.Location()
	settlementTZ := timezone.Name()
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	start := today.AddDate(0, 0, -1)
	return start, today, start.Format("2006-01-02"), settlementTZ, today.Add(leaderboardRewardSettlementDelay)
}

func leaderboardDailyRewardTiers(settings leaderboardDailyRewardSettings) []usagestats.LeaderboardDailyRewardTier {
	return []usagestats.LeaderboardDailyRewardTier{
		{Rank: 1, Amount: settings.RankAmounts[1]},
		{Rank: 2, Amount: settings.RankAmounts[2]},
		{Rank: 3, Amount: settings.RankAmounts[3]},
	}
}

func resolveLeaderboardDailyRewardClaimState(status *usagestats.LeaderboardDailyRewards) (bool, string) {
	if status == nil {
		return false, leaderboardRewardReasonDisabled
	}
	if status.Claimed {
		return false, leaderboardRewardReasonAlreadyClaimed
	}
	if !status.Enabled {
		return false, leaderboardRewardReasonDisabled
	}
	if !status.SettlementReady {
		return false, leaderboardRewardReasonSettling
	}
	if !status.ThresholdMet {
		return false, leaderboardRewardReasonThresholdNotMet
	}
	if status.CurrentUserRank <= 0 {
		return false, leaderboardRewardReasonNotRanked
	}
	if status.CurrentUserRank > 3 {
		return false, leaderboardRewardReasonNotTopThree
	}
	if status.CurrentUserRewardAmount <= 0 {
		return false, leaderboardRewardReasonZeroReward
	}
	return true, leaderboardRewardReasonEligible
}

func leaderboardRewardRedeemCode(rewardDate string, userID int64) string {
	return "LDR" + strings.ReplaceAll(rewardDate, "-", "") + "U" + strings.ToUpper(strconv.FormatInt(userID, 36))
}
