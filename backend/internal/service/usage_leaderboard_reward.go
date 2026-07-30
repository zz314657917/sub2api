package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/robfig/cron/v3"
)

const (
	LeaderboardDailyRewardModeDisabled  = "disabled"
	LeaderboardDailyRewardModeRedPacket = "red_packet"
	LeaderboardDailyRewardModeLottery   = "lottery"

	LeaderboardRewardModeDisabled  = LeaderboardDailyRewardModeDisabled
	LeaderboardRewardModeRedPacket = LeaderboardDailyRewardModeRedPacket
	LeaderboardRewardModeLottery   = LeaderboardDailyRewardModeLottery

	defaultLeaderboardDailyRewardLotteryCron = "0 12 * * 4"
	defaultLeaderboardLotteryCron            = defaultLeaderboardDailyRewardLotteryCron

	leaderboardRewardReasonEligible          = "eligible"
	leaderboardRewardReasonDisabled          = "disabled"
	leaderboardRewardReasonSettling          = "settling"
	leaderboardRewardReasonThresholdNotMet   = "threshold_not_met"
	leaderboardRewardReasonNotRanked         = "not_ranked"
	leaderboardRewardReasonNotTopThree       = "not_top_three"
	leaderboardRewardReasonNotTopTen         = "not_top_ten"
	leaderboardRewardReasonZeroReward        = "zero_reward"
	leaderboardRewardReasonAlreadyClaimed    = "already_claimed"
	leaderboardRewardReasonLotteryNotDrawn   = "lottery_not_drawn"
	leaderboardRewardReasonLotteryNotWinner  = "lottery_not_winner"
	leaderboardRewardReasonPacketUnavailable = "red_packet_unavailable"

	leaderboardRewardSettlementDelay = 30 * time.Minute
	leaderboardRewardTopLimit        = 10
	leaderboardRedPacketCount        = 10
)

var (
	ErrLeaderboardDailyRewardClaimNotFound    = infraerrors.NotFound("LEADERBOARD_DAILY_REWARD_CLAIM_NOT_FOUND", "leaderboard daily reward claim not found")
	ErrLeaderboardDailyRewardAlreadyClaimed   = infraerrors.Conflict("LEADERBOARD_DAILY_REWARD_ALREADY_CLAIMED", "leaderboard daily reward already claimed")
	ErrLeaderboardDailyRewardNotEligible      = infraerrors.Forbidden("LEADERBOARD_DAILY_REWARD_NOT_ELIGIBLE", "leaderboard daily reward is not claimable")
	ErrLeaderboardDailyRewardUnavailable      = infraerrors.ServiceUnavailable("LEADERBOARD_DAILY_REWARD_UNAVAILABLE", "leaderboard daily reward service is unavailable")
	ErrLeaderboardRedPacketUnavailable        = infraerrors.Conflict("LEADERBOARD_RED_PACKET_UNAVAILABLE", "leaderboard red packet is unavailable")
	ErrLeaderboardLotteryRewardNotFound       = infraerrors.NotFound("LEADERBOARD_LOTTERY_REWARD_NOT_FOUND", "leaderboard lottery reward not found")
	ErrLeaderboardLotteryRewardAlreadyDrawn   = infraerrors.Conflict("LEADERBOARD_LOTTERY_REWARD_ALREADY_DRAWN", "leaderboard lottery reward already drawn")
	ErrLeaderboardLotteryRewardAlreadySettled = infraerrors.Conflict("LEADERBOARD_LOTTERY_REWARD_ALREADY_SETTLED", "leaderboard lottery reward already settled")
)

type leaderboardDailyRewardSettings struct {
	Mode               string
	Enabled            bool
	MinTotalActualCost float64
	RankAmounts        map[int]float64
	RedPacketTotal     float64
	RedPacketMin       float64
	RedPacketMax       float64
	LotteryAmount      float64
	LotteryCron        string
}

// LeaderboardDailyRewardClaim records one user's settlement claim for a reward period.
type LeaderboardDailyRewardClaim struct {
	ID              int64
	RewardDate      string
	RewardMode      string
	UserID          int64
	Rank            int
	Amount          float64
	TotalActualCost float64
	RedeemCodeID    *int64
	PacketID        *int64
	LotteryRunID    *int64
	CreatedAt       time.Time
}

// LeaderboardRedPacketReward records one pre-split red packet for a reward period.
type LeaderboardRedPacketReward struct {
	ID          int64
	RewardDate  string
	PacketIndex int
	Amount      float64
	ClaimedBy   *int64
	ClaimID     *int64
	ClaimedAt   *time.Time
	CreatedAt   time.Time
}

type LeaderboardRedPacket struct {
	ID         int64
	RewardDate string
	PacketNo   int
	Amount     float64
	ClaimedBy  *int64
	ClaimID    *int64
	ClaimedAt  *time.Time
	CreatedAt  time.Time
}

// LeaderboardRedPacketSummary summarizes packet state for the current viewer.
type LeaderboardRedPacketSummary struct {
	PacketCount        int
	ClaimedCount       int
	CurrentUserClaimed bool
	CurrentUserAmount  float64
	ClaimsByUser       map[int64]float64
}

// LeaderboardLotteryReward records the weekly lottery draw and optional settlement link.
type LeaderboardLotteryReward struct {
	ID              int64
	RewardDate      string
	WinnerUserID    int64
	WinnerRank      int
	Amount          float64
	TotalActualCost float64
	ClaimID         *int64
	RedeemCodeID    *int64
	DrawnAt         time.Time
	CreatedAt       time.Time
}

type LeaderboardLotteryRun struct {
	ID              int64
	RewardDate      string
	WinnerUserID    int64
	WinnerRank      int
	Amount          float64
	TotalActualCost float64
	RedeemCodeID    *int64
	CreatedAt       time.Time
}

// LeaderboardDailyRewardClaimStore is implemented by repositories that persist reward claims.
type LeaderboardDailyRewardClaimStore interface {
	GetLeaderboardDailyRewardClaimByMode(ctx context.Context, rewardDate, rewardMode string, userID int64) (*LeaderboardDailyRewardClaim, error)
	CreateLeaderboardDailyRewardClaim(ctx context.Context, claim *LeaderboardDailyRewardClaim) error
	AttachLeaderboardDailyRewardClaimRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error
}

type LeaderboardRedPacketStore interface {
	EnsureLeaderboardRedPackets(ctx context.Context, rewardDate string, amounts []float64) error
	GetLeaderboardRedPackets(ctx context.Context, rewardDate string) ([]LeaderboardRedPacket, error)
	ClaimRandomLeaderboardRedPacket(ctx context.Context, rewardDate string, userID int64, claimID int64) (*LeaderboardRedPacket, error)
}

type LeaderboardLotteryRewardStore interface {
	GetLeaderboardLotteryRun(ctx context.Context, rewardDate string) (*LeaderboardLotteryRun, error)
	CreateLeaderboardLotteryRun(ctx context.Context, run *LeaderboardLotteryRun) error
	AttachLeaderboardLotteryRunRedeemCode(ctx context.Context, runID, redeemCodeID int64) error
}

// LeaderboardDailyRewardClaimResult represents a successful reward claim.
type LeaderboardDailyRewardClaimResult struct {
	DailyRewards    *usagestats.LeaderboardDailyRewards `json:"daily_rewards"`
	ClaimedAmount   float64                             `json:"claimed_amount"`
	RedPacketAmount float64                             `json:"red_packet_amount,omitempty"`
	LotteryAmount   float64                             `json:"lottery_amount,omitempty"`
}

type leaderboardRewardBalanceGrantRepository interface {
	AddBalance(ctx context.Context, id int64, amount float64) error
}

// GetLeaderboardDailyRewards returns last week's leaderboard reward status.
// The reward settlement window is intentionally based on the server timezone,
// so clients cannot change reward dates or ranking windows by changing timezone.
func (s *UsageService) GetLeaderboardDailyRewards(ctx context.Context, userID int64, _ string) (*usagestats.LeaderboardDailyRewards, error) {
	return s.getLeaderboardDailyRewards(ctx, userID, timezone.Now())
}

func (s *UsageService) getLeaderboardDailyRewards(ctx context.Context, userID int64, now time.Time) (*usagestats.LeaderboardDailyRewards, error) {
	start, end, rewardDate, settlementTZ, defaultAvailableAt := leaderboardRewardWindow(now)
	leaderboard, err := s.GetUserLeaderboard(ctx, start, end, leaderboardRewardTopLimit, userID)
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

	claimAvailableAt := defaultAvailableAt
	if settings.Mode == LeaderboardDailyRewardModeLottery {
		claimAvailableAt = leaderboardLotteryDrawAt(end, settings.LotteryCron)
	}

	topUsers := leaderboardDailyRewardTopUsers(leaderboard.Ranking)
	status := &usagestats.LeaderboardDailyRewards{
		RewardDate:               rewardDate,
		RewardMode:               settings.Mode,
		SettlementTimezone:       settlementTZ,
		SettlementReady:          !now.Before(claimAvailableAt),
		ClaimAvailableAt:         claimAvailableAt.Format(time.RFC3339),
		Enabled:                  settings.Enabled,
		MinTotalActualCost:       settings.MinTotalActualCost,
		YesterdayTotalActualCost: leaderboard.TotalActualCost,
		ThresholdMet:             leaderboard.TotalActualCost > settings.MinTotalActualCost,
		Rewards:                  leaderboardDailyRewardTiers(settings),
		TopUsers:                 topUsers,
		RedPacketPoolAmount:      settings.RedPacketTotal,
		RedPacketMinAmount:       settings.RedPacketMin,
		RedPacketMaxAmount:       settings.RedPacketMax,
		LotteryAmount:            settings.LotteryAmount,
		LotteryCron:              settings.LotteryCron,
		Reason:                   leaderboardRewardReasonDisabled,
	}
	if settings.Mode == LeaderboardDailyRewardModeLottery {
		status.LotteryDrawAt = claimAvailableAt.Format(time.RFC3339)
	}

	if leaderboard.CurrentUserEntry != nil {
		status.CurrentUserRank = leaderboard.CurrentUserEntry.Rank
		if amount, ok := settings.RankAmounts[int(leaderboard.CurrentUserEntry.Rank)]; ok {
			status.CurrentUserRewardAmount = amount
		}
	}

	if settings.Mode == LeaderboardDailyRewardModeRedPacket && status.SettlementReady && status.ThresholdMet && settings.RedPacketTotal > 0 {
		_ = s.ensureLeaderboardRedPackets(ctx, rewardDate, settings)
	}
	if settings.Mode == LeaderboardDailyRewardModeLottery && status.SettlementReady && status.ThresholdMet && settings.LotteryAmount > 0 {
		if _, settleErr := s.settleLeaderboardLotteryRewardFromTopUsers(ctx, rewardDate, topUsers, leaderboard.TotalActualCost, settings, now); settleErr != nil && !errors.Is(settleErr, ErrLeaderboardDailyRewardUnavailable) {
			return nil, settleErr
		}
	}

	if store := s.leaderboardDailyRewardClaimStore(); store != nil {
		claim, err := store.GetLeaderboardDailyRewardClaimByMode(ctx, rewardDate, settings.Mode, userID)
		if err != nil && !errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
			return nil, err
		}
		if claim != nil {
			status.Claimed = true
			status.CurrentUserRewardAmount = claim.Amount
		}
	}

	if settings.Mode == LeaderboardDailyRewardModeRedPacket {
		if summary, err := s.getLeaderboardRedPacketSummary(ctx, rewardDate, userID); err != nil {
			return nil, err
		} else if summary != nil {
			status.RedPacketCount = summary.PacketCount
			status.RedPacketClaimedCount = summary.ClaimedCount
			applyLeaderboardRedPacketClaims(status.TopUsers, summary.ClaimsByUser)
			if summary.CurrentUserClaimed {
				status.Claimed = true
				status.CurrentUserRewardAmount = summary.CurrentUserAmount
			}
		}
	}

	if settings.Mode == LeaderboardDailyRewardModeLottery {
		if lottery, err := s.getLeaderboardLotteryReward(ctx, rewardDate); err != nil {
			return nil, err
		} else if lottery != nil {
			status.LotteryWinnerRank = leaderboardInt64Ptr(int64(lottery.WinnerRank))
			status.LotteryWinnerUserID = leaderboardInt64Ptr(lottery.WinnerUserID)
			markLeaderboardLotteryWinner(status.TopUsers, lottery)
			if winner := leaderboardTopUserByUserID(status.TopUsers, lottery.WinnerUserID); winner != nil {
				status.LotteryWinnerDisplayName = stringPtr(firstNonEmptyLeaderboardReward(winner.DisplayName, winner.Username, winner.EmailMasked))
				status.LotteryWinnerEmailMasked = stringPtr(winner.EmailMasked)
			}
			if lottery.WinnerUserID == userID {
				status.CurrentUserRewardAmount = lottery.Amount
				if lottery.RedeemCodeID != nil {
					status.Claimed = true
				}
			}
		}
	}

	status.CanClaim, status.Reason = resolveLeaderboardDailyRewardClaimState(status)
	return status, nil
}

// ClaimLeaderboardDailyReward grants last week's leaderboard reward to the user.
func (s *UsageService) ClaimLeaderboardDailyReward(ctx context.Context, userID int64, _ string) (*LeaderboardDailyRewardClaimResult, error) {
	if err := s.EnsureLeaderboardAccess(ctx, userID); err != nil {
		return nil, err
	}
	return s.claimLeaderboardDailyReward(ctx, userID, timezone.Now())
}

// SettleDueLeaderboardLotteryRewards settles the last completed leaderboard period
// when the configured lottery draw time has arrived. It is safe to call repeatedly.
func (s *UsageService) SettleDueLeaderboardLotteryRewards(ctx context.Context, now time.Time) error {
	start, end, rewardDate, _, _ := leaderboardRewardWindow(now)
	settings, err := s.getLeaderboardDailyRewardSettings(ctx)
	if err != nil {
		return err
	}
	if settings.Mode != LeaderboardDailyRewardModeLottery || settings.LotteryAmount <= 0 {
		return nil
	}
	if now.Before(leaderboardLotteryDrawAt(end, settings.LotteryCron)) {
		return nil
	}
	leaderboard, err := s.GetUserLeaderboard(ctx, start, end, leaderboardRewardTopLimit, 0)
	if err != nil {
		return err
	}
	if leaderboard == nil || leaderboard.TotalActualCost <= settings.MinTotalActualCost {
		return nil
	}
	topUsers := leaderboardDailyRewardTopUsers(leaderboard.Ranking)
	_, err = s.settleLeaderboardLotteryRewardFromTopUsers(ctx, rewardDate, topUsers, leaderboard.TotalActualCost, settings, now)
	if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) || errors.Is(err, ErrLeaderboardLotteryRewardAlreadySettled) {
		return nil
	}
	return err
}

func (s *UsageService) claimLeaderboardDailyReward(ctx context.Context, userID int64, now time.Time) (*LeaderboardDailyRewardClaimResult, error) {
	status, settings, topUsers, err := s.prepareLeaderboardRewardClaim(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if !status.CanClaim {
		if settings.Mode == LeaderboardDailyRewardModeLottery && status.Reason == leaderboardRewardReasonLotteryNotDrawn {
			lottery, err := s.settleLeaderboardLotteryRewardFromTopUsers(ctx, status.RewardDate, topUsers, status.YesterdayTotalActualCost, settings, now)
			if err != nil {
				if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) || errors.Is(err, ErrLeaderboardLotteryRewardAlreadySettled) {
					return nil, ErrLeaderboardDailyRewardAlreadyClaimed
				}
				return nil, err
			}
			if lottery != nil && lottery.WinnerUserID == userID {
				updated, err := s.getLeaderboardDailyRewards(ctx, userID, now)
				if err != nil {
					return nil, err
				}
				return &LeaderboardDailyRewardClaimResult{DailyRewards: updated, ClaimedAmount: lottery.Amount, LotteryAmount: lottery.Amount}, nil
			}
			return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": leaderboardRewardReasonLotteryNotWinner})
		}
		if status.Reason == leaderboardRewardReasonAlreadyClaimed {
			return nil, ErrLeaderboardDailyRewardAlreadyClaimed
		}
		return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": status.Reason})
	}

	switch settings.Mode {
	case LeaderboardDailyRewardModeRedPacket:
		return s.claimLeaderboardRedPacketReward(ctx, userID, status, settings, now)
	case LeaderboardDailyRewardModeLottery:
		lottery, err := s.settleLeaderboardLotteryRewardFromTopUsers(ctx, status.RewardDate, topUsers, status.YesterdayTotalActualCost, settings, now)
		if err != nil {
			if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) || errors.Is(err, ErrLeaderboardLotteryRewardAlreadySettled) {
				return nil, ErrLeaderboardDailyRewardAlreadyClaimed
			}
			return nil, err
		}
		if lottery == nil || lottery.WinnerUserID != userID {
			return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": leaderboardRewardReasonLotteryNotWinner})
		}
		updated, err := s.getLeaderboardDailyRewards(ctx, userID, now)
		if err != nil {
			return nil, err
		}
		return &LeaderboardDailyRewardClaimResult{DailyRewards: updated, ClaimedAmount: lottery.Amount, LotteryAmount: lottery.Amount}, nil
	default:
		return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": status.Reason})
	}
}

func (s *UsageService) prepareLeaderboardRewardClaim(ctx context.Context, userID int64, now time.Time) (*usagestats.LeaderboardDailyRewards, leaderboardDailyRewardSettings, []usagestats.LeaderboardDailyRewardTopUser, error) {
	start, end, rewardDate, _, defaultAvailableAt := leaderboardRewardWindow(now)
	leaderboard, err := s.GetUserLeaderboard(ctx, start, end, leaderboardRewardTopLimit, userID)
	if err != nil {
		return nil, leaderboardDailyRewardSettings{}, nil, err
	}
	if leaderboard == nil {
		leaderboard = &usagestats.UserLeaderboardResponse{}
	}
	settings, err := s.getLeaderboardDailyRewardSettings(ctx)
	if err != nil {
		return nil, leaderboardDailyRewardSettings{}, nil, err
	}
	availableAt := defaultAvailableAt
	if settings.Mode == LeaderboardDailyRewardModeLottery {
		availableAt = leaderboardLotteryDrawAt(end, settings.LotteryCron)
	}
	topUsers := leaderboardDailyRewardTopUsers(leaderboard.Ranking)
	status := &usagestats.LeaderboardDailyRewards{
		RewardDate:               rewardDate,
		RewardMode:               settings.Mode,
		SettlementTimezone:       timezone.Name(),
		SettlementReady:          !now.Before(availableAt),
		ClaimAvailableAt:         availableAt.Format(time.RFC3339),
		Enabled:                  settings.Enabled,
		MinTotalActualCost:       settings.MinTotalActualCost,
		YesterdayTotalActualCost: leaderboard.TotalActualCost,
		ThresholdMet:             leaderboard.TotalActualCost > settings.MinTotalActualCost,
		Rewards:                  leaderboardDailyRewardTiers(settings),
		TopUsers:                 topUsers,
		RedPacketPoolAmount:      settings.RedPacketTotal,
		RedPacketMinAmount:       settings.RedPacketMin,
		RedPacketMaxAmount:       settings.RedPacketMax,
		LotteryAmount:            settings.LotteryAmount,
		LotteryCron:              settings.LotteryCron,
	}
	if leaderboard.CurrentUserEntry != nil {
		status.CurrentUserRank = leaderboard.CurrentUserEntry.Rank
	}
	if store := s.leaderboardDailyRewardClaimStore(); store != nil {
		claim, err := store.GetLeaderboardDailyRewardClaimByMode(ctx, rewardDate, settings.Mode, userID)
		if err != nil && !errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
			return nil, leaderboardDailyRewardSettings{}, nil, err
		}
		if claim != nil {
			status.Claimed = true
			status.CurrentUserRewardAmount = claim.Amount
		}
	}
	if settings.Mode == LeaderboardDailyRewardModeLottery {
		if lottery, err := s.getLeaderboardLotteryReward(ctx, rewardDate); err != nil {
			return nil, leaderboardDailyRewardSettings{}, nil, err
		} else if lottery != nil {
			status.LotteryWinnerRank = leaderboardInt64Ptr(int64(lottery.WinnerRank))
			status.LotteryWinnerUserID = leaderboardInt64Ptr(lottery.WinnerUserID)
			markLeaderboardLotteryWinner(status.TopUsers, lottery)
			if winner := leaderboardTopUserByUserID(status.TopUsers, lottery.WinnerUserID); winner != nil {
				status.LotteryWinnerDisplayName = stringPtr(firstNonEmptyLeaderboardReward(winner.DisplayName, winner.Username, winner.EmailMasked))
				status.LotteryWinnerEmailMasked = stringPtr(winner.EmailMasked)
			}
			if lottery.WinnerUserID == userID {
				status.CurrentUserRewardAmount = lottery.Amount
				if lottery.RedeemCodeID != nil {
					status.Claimed = true
				}
			}
		}
	}
	status.CanClaim, status.Reason = resolveLeaderboardDailyRewardClaimState(status)
	return status, settings, topUsers, nil
}

func (s *UsageService) claimLeaderboardRedPacketReward(ctx context.Context, userID int64, status *usagestats.LeaderboardDailyRewards, settings leaderboardDailyRewardSettings, now time.Time) (*LeaderboardDailyRewardClaimResult, error) {
	if s.userRepo == nil || s.redeemRepo == nil || s.leaderboardDailyRewardClaimStore() == nil || s.leaderboardRedPacketStore() == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}

	claimCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin leaderboard red packet transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		claimCtx = dbent.NewTxContext(ctx, tx)
	}

	if err := s.ensureLeaderboardRedPackets(claimCtx, status.RewardDate, settings); err != nil {
		return nil, err
	}
	claim := &LeaderboardDailyRewardClaim{
		RewardDate:      status.RewardDate,
		RewardMode:      LeaderboardDailyRewardModeRedPacket,
		UserID:          userID,
		Rank:            int(status.CurrentUserRank),
		Amount:          0,
		TotalActualCost: status.YesterdayTotalActualCost,
	}
	store := s.leaderboardDailyRewardClaimStore()
	if err := store.CreateLeaderboardDailyRewardClaim(claimCtx, claim); err != nil {
		if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) {
			return nil, ErrLeaderboardDailyRewardAlreadyClaimed
		}
		return nil, err
	}

	packet, err := s.leaderboardRedPacketStore().ClaimRandomLeaderboardRedPacket(claimCtx, status.RewardDate, userID, claim.ID)
	if err != nil {
		if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) {
			return nil, ErrLeaderboardDailyRewardAlreadyClaimed
		}
		if errors.Is(err, ErrLeaderboardRedPacketUnavailable) {
			return nil, ErrLeaderboardDailyRewardNotEligible.WithMetadata(map[string]string{"reason": leaderboardRewardReasonPacketUnavailable})
		}
		return nil, err
	}
	claim.Amount = packet.Amount
	if packet.ClaimID != nil {
		claim.PacketID = &packet.ID
	}

	redeemCode, err := s.createLeaderboardRewardAuditAndGrant(claimCtx, status.RewardDate, LeaderboardDailyRewardModeRedPacket, userID, int(status.CurrentUserRank), packet.Amount, now)
	if err != nil {
		return nil, err
	}
	if err := store.AttachLeaderboardDailyRewardClaimRedeemCode(claimCtx, claim.ID, redeemCode.ID); err != nil {
		return nil, err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit leaderboard red packet transaction: %w", err)
		}
	}

	status.Claimed = true
	status.CanClaim = false
	status.Reason = leaderboardRewardReasonAlreadyClaimed
	status.CurrentUserRewardAmount = packet.Amount
	s.invalidateUsageCaches(ctx, userID, true)

	return &LeaderboardDailyRewardClaimResult{
		DailyRewards:    status,
		ClaimedAmount:   packet.Amount,
		RedPacketAmount: packet.Amount,
	}, nil
}

func (s *UsageService) settleLeaderboardLotteryRewardFromTopUsers(ctx context.Context, rewardDate string, topUsers []usagestats.LeaderboardDailyRewardTopUser, totalActualCost float64, settings leaderboardDailyRewardSettings, now time.Time) (*LeaderboardLotteryReward, error) {
	if settings.Mode != LeaderboardDailyRewardModeLottery || settings.LotteryAmount <= 0 || len(topUsers) == 0 {
		return nil, nil
	}
	if s.userRepo == nil || s.redeemRepo == nil || s.leaderboardDailyRewardClaimStore() == nil || s.leaderboardLotteryRewardStore() == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}
	if s.entClient == nil && dbent.TxFromContext(ctx) == nil {
		return nil, ErrLeaderboardDailyRewardUnavailable
	}

	winner := selectLeaderboardLotteryWinner(rewardDate, topUsers)
	if winner.UserID <= 0 {
		return nil, nil
	}

	claimCtx := ctx
	var tx *dbent.Tx
	if s.entClient != nil && dbent.TxFromContext(ctx) == nil {
		txCandidate, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin leaderboard lottery transaction: %w", txErr)
		}
		tx = txCandidate
		defer func() { _ = tx.Rollback() }()
		claimCtx = dbent.NewTxContext(ctx, tx)
	}

	lotteryStore := s.leaderboardLotteryRewardStore()
	run, err := lotteryStore.GetLeaderboardLotteryRun(claimCtx, rewardDate)
	if errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
		run = &LeaderboardLotteryRun{
			RewardDate:      rewardDate,
			WinnerUserID:    winner.UserID,
			WinnerRank:      int(winner.Rank),
			Amount:          settings.LotteryAmount,
			TotalActualCost: totalActualCost,
		}
		if createErr := lotteryStore.CreateLeaderboardLotteryRun(claimCtx, run); createErr != nil {
			if !errors.Is(createErr, ErrLeaderboardDailyRewardAlreadyClaimed) {
				return nil, createErr
			}
			run, err = lotteryStore.GetLeaderboardLotteryRun(claimCtx, rewardDate)
			if err != nil {
				return nil, err
			}
		}
	} else if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, nil
	}
	lottery := leaderboardLotteryRewardFromRun(run)
	if run.RedeemCodeID != nil {
		return lottery, nil
	}

	claim := &LeaderboardDailyRewardClaim{
		RewardDate:      rewardDate,
		RewardMode:      LeaderboardDailyRewardModeLottery,
		UserID:          run.WinnerUserID,
		Rank:            run.WinnerRank,
		Amount:          run.Amount,
		TotalActualCost: run.TotalActualCost,
		LotteryRunID:    &run.ID,
	}
	store := s.leaderboardDailyRewardClaimStore()
	if err := store.CreateLeaderboardDailyRewardClaim(claimCtx, claim); err != nil {
		if errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed) {
			existing, getErr := store.GetLeaderboardDailyRewardClaimByMode(claimCtx, rewardDate, LeaderboardDailyRewardModeLottery, run.WinnerUserID)
			if getErr != nil {
				return nil, getErr
			}
			if existing.RedeemCodeID != nil {
				if attachErr := lotteryStore.AttachLeaderboardLotteryRunRedeemCode(claimCtx, run.ID, *existing.RedeemCodeID); attachErr != nil && !errors.Is(attachErr, ErrLeaderboardDailyRewardClaimNotFound) {
					return nil, attachErr
				}
				lottery.ClaimID = &existing.ID
				lottery.RedeemCodeID = existing.RedeemCodeID
				return lottery, nil
			}
		}
		return nil, err
	}

	redeemCode, err := s.createLeaderboardRewardAuditAndGrant(claimCtx, rewardDate, LeaderboardDailyRewardModeLottery, run.WinnerUserID, run.WinnerRank, run.Amount, now)
	if err != nil {
		return nil, err
	}
	if err := store.AttachLeaderboardDailyRewardClaimRedeemCode(claimCtx, claim.ID, redeemCode.ID); err != nil {
		return nil, err
	}
	if err := lotteryStore.AttachLeaderboardLotteryRunRedeemCode(claimCtx, run.ID, redeemCode.ID); err != nil {
		if !errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
			return nil, err
		}
		return nil, ErrLeaderboardDailyRewardAlreadyClaimed
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit leaderboard lottery transaction: %w", err)
		}
	}

	lottery.ClaimID = &claim.ID
	lottery.RedeemCodeID = &redeemCode.ID
	s.invalidateUsageCaches(ctx, run.WinnerUserID, true)
	return lottery, nil
}

func (s *UsageService) createLeaderboardRewardAuditAndGrant(ctx context.Context, rewardDate, rewardMode string, userID int64, rank int, amount float64, now time.Time) (*RedeemCode, error) {
	usedAt := now.UTC()
	redeemCode := &RedeemCode{
		Code:   leaderboardRewardRedeemCode(rewardDate, rewardMode, userID),
		Type:   RedeemTypeLeaderboardReward,
		Value:  amount,
		Status: StatusUsed,
		UsedBy: &userID,
		UsedAt: &usedAt,
		Notes:  fmt.Sprintf("leaderboard weekly %s reward %s rank %d", rewardMode, rewardDate, rank),
	}
	if err := s.redeemRepo.Create(ctx, redeemCode); err != nil {
		return nil, fmt.Errorf("create leaderboard reward audit record: %w", err)
	}
	if err := grantLeaderboardRewardBalance(ctx, s.userRepo, userID, amount); err != nil {
		return nil, fmt.Errorf("update leaderboard reward balance: %w", err)
	}
	return redeemCode, nil
}

func (s *UsageService) getLeaderboardDailyRewardSettings(ctx context.Context) (leaderboardDailyRewardSettings, error) {
	result := leaderboardDailyRewardSettings{
		Mode:        LeaderboardDailyRewardModeDisabled,
		RankAmounts: map[int]float64{1: 0, 2: 0, 3: 0},
		LotteryCron: defaultLeaderboardDailyRewardLotteryCron,
	}
	if s.settingRepo == nil {
		return result, nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingKeyLeaderboardRewardMode,
		SettingKeyLeaderboardDailyRewardEnabled,
		SettingKeyLeaderboardDailyRewardMinTotalActualCost,
		SettingKeyLeaderboardDailyRewardRank1Amount,
		SettingKeyLeaderboardDailyRewardRank2Amount,
		SettingKeyLeaderboardDailyRewardRank3Amount,
		SettingKeyLeaderboardRedPacketPoolAmount,
		SettingKeyLeaderboardRedPacketMinAmount,
		SettingKeyLeaderboardRedPacketMaxAmount,
		SettingKeyLeaderboardLotteryAmount,
		SettingKeyLeaderboardLotteryCron,
	})
	if err != nil {
		return result, fmt.Errorf("get leaderboard daily reward settings: %w", err)
	}
	result.Mode = NormalizeLeaderboardRewardMode(values[SettingKeyLeaderboardRewardMode], values[SettingKeyLeaderboardDailyRewardEnabled] == "true")
	result.Enabled = result.Mode != LeaderboardDailyRewardModeDisabled
	result.MinTotalActualCost = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardMinTotalActualCost], 0)
	result.RankAmounts[1] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank1Amount], 0)
	result.RankAmounts[2] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank2Amount], 0)
	result.RankAmounts[3] = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardDailyRewardRank3Amount], 0)
	result.RedPacketTotal = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardRedPacketPoolAmount], 0)
	if result.RedPacketTotal <= 0 && result.Mode == LeaderboardDailyRewardModeRedPacket {
		result.RedPacketTotal = result.RankAmounts[1] + result.RankAmounts[2] + result.RankAmounts[3]
	}
	result.RedPacketMin = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardRedPacketMinAmount], 0)
	result.RedPacketMax = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardRedPacketMaxAmount], 0)
	if result.RedPacketMax > 0 && result.RedPacketMax < result.RedPacketMin {
		result.RedPacketMax = result.RedPacketMin
	}
	result.LotteryAmount = parseNonNegativeFloatSetting(values[SettingKeyLeaderboardLotteryAmount], 0)
	result.LotteryCron = normalizeLeaderboardLotteryCron(values[SettingKeyLeaderboardLotteryCron])
	return result, nil
}

func NormalizeLeaderboardDailyRewardMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case LeaderboardDailyRewardModeRedPacket:
		return LeaderboardDailyRewardModeRedPacket
	case LeaderboardDailyRewardModeLottery:
		return LeaderboardDailyRewardModeLottery
	default:
		return LeaderboardDailyRewardModeDisabled
	}
}

func NormalizeLeaderboardRewardMode(raw string, legacyEnabled bool) string {
	normalized := NormalizeLeaderboardDailyRewardMode(raw)
	if normalized == LeaderboardDailyRewardModeDisabled && strings.TrimSpace(raw) == "" && legacyEnabled {
		return LeaderboardDailyRewardModeRedPacket
	}
	return normalized
}

func normalizeLeaderboardLotteryCron(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || !isValidLeaderboardLotteryCron(value) {
		return defaultLeaderboardDailyRewardLotteryCron
	}
	return value
}

func (s *UsageService) leaderboardDailyRewardClaimStore() LeaderboardDailyRewardClaimStore {
	store, _ := s.usageRepo.(LeaderboardDailyRewardClaimStore)
	return store
}

func (s *UsageService) leaderboardRedPacketStore() LeaderboardRedPacketStore {
	store, _ := s.usageRepo.(LeaderboardRedPacketStore)
	return store
}

func (s *UsageService) leaderboardLotteryRewardStore() LeaderboardLotteryRewardStore {
	store, _ := s.usageRepo.(LeaderboardLotteryRewardStore)
	return store
}

func (s *UsageService) ensureLeaderboardRedPackets(ctx context.Context, rewardDate string, settings leaderboardDailyRewardSettings) error {
	store := s.leaderboardRedPacketStore()
	if store == nil {
		return nil
	}
	amounts := splitLeaderboardRedPacketAmounts(rewardDate, settings.RedPacketTotal, leaderboardRedPacketCount, settings.RedPacketMin, settings.RedPacketMax)
	return store.EnsureLeaderboardRedPackets(ctx, rewardDate, amounts)
}

func (s *UsageService) getLeaderboardRedPacketSummary(ctx context.Context, rewardDate string, userID int64) (*LeaderboardRedPacketSummary, error) {
	store := s.leaderboardRedPacketStore()
	if store == nil {
		return nil, nil
	}
	packets, err := store.GetLeaderboardRedPackets(ctx, rewardDate)
	if err != nil {
		if errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
			return &LeaderboardRedPacketSummary{}, nil
		}
		return nil, err
	}
	summary := &LeaderboardRedPacketSummary{
		PacketCount:  len(packets),
		ClaimsByUser: make(map[int64]float64),
	}
	for _, packet := range packets {
		if packet.ClaimedBy != nil {
			summary.ClaimedCount++
			summary.ClaimsByUser[*packet.ClaimedBy] = packet.Amount
			if *packet.ClaimedBy == userID {
				summary.CurrentUserClaimed = true
				summary.CurrentUserAmount = packet.Amount
			}
		}
	}
	return summary, nil
}

func applyLeaderboardRedPacketClaims(topUsers []usagestats.LeaderboardDailyRewardTopUser, claimsByUser map[int64]float64) {
	if len(claimsByUser) == 0 {
		return
	}
	for i := range topUsers {
		amount, ok := claimsByUser[topUsers[i].UserID]
		if !ok {
			continue
		}
		topUsers[i].Claimed = true
		topUsers[i].ClaimedAmount = &amount
	}
}

func markLeaderboardLotteryWinner(topUsers []usagestats.LeaderboardDailyRewardTopUser, lottery *LeaderboardLotteryReward) {
	if lottery == nil {
		return
	}
	for i := range topUsers {
		if topUsers[i].UserID != lottery.WinnerUserID {
			continue
		}
		amount := lottery.Amount
		topUsers[i].LotteryWinner = true
		topUsers[i].Claimed = lottery.RedeemCodeID != nil || lottery.ClaimID != nil
		topUsers[i].ClaimedAmount = &amount
		return
	}
}

func leaderboardTopUserByUserID(topUsers []usagestats.LeaderboardDailyRewardTopUser, userID int64) *usagestats.LeaderboardDailyRewardTopUser {
	for i := range topUsers {
		if topUsers[i].UserID == userID {
			return &topUsers[i]
		}
	}
	return nil
}

func (s *UsageService) getLeaderboardLotteryReward(ctx context.Context, rewardDate string) (*LeaderboardLotteryReward, error) {
	store := s.leaderboardLotteryRewardStore()
	if store == nil {
		return nil, nil
	}
	run, err := store.GetLeaderboardLotteryRun(ctx, rewardDate)
	if errors.Is(err, ErrLeaderboardDailyRewardClaimNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return leaderboardLotteryRewardFromRun(run), nil
}

func leaderboardLotteryRewardFromRun(run *LeaderboardLotteryRun) *LeaderboardLotteryReward {
	if run == nil {
		return nil
	}
	return &LeaderboardLotteryReward{
		ID:              run.ID,
		RewardDate:      run.RewardDate,
		WinnerUserID:    run.WinnerUserID,
		WinnerRank:      run.WinnerRank,
		Amount:          run.Amount,
		TotalActualCost: run.TotalActualCost,
		RedeemCodeID:    run.RedeemCodeID,
		DrawnAt:         run.CreatedAt,
		CreatedAt:       run.CreatedAt,
	}
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
	localNow := now.In(loc)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
	weekdayOffset := (int(today.Weekday()) + 6) % 7
	thisWeekStart := today.AddDate(0, 0, -weekdayOffset)
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	rewardDate := fmt.Sprintf("%s~%s", lastWeekStart.Format("2006-01-02"), thisWeekStart.AddDate(0, 0, -1).Format("2006-01-02"))
	return lastWeekStart, thisWeekStart, rewardDate, settlementTZ, thisWeekStart.Add(leaderboardRewardSettlementDelay)
}

func leaderboardDailyRewardTiers(settings leaderboardDailyRewardSettings) []usagestats.LeaderboardDailyRewardTier {
	return []usagestats.LeaderboardDailyRewardTier{
		{Rank: 1, Amount: settings.RankAmounts[1]},
		{Rank: 2, Amount: settings.RankAmounts[2]},
		{Rank: 3, Amount: settings.RankAmounts[3]},
	}
}

func leaderboardDailyRewardTopUsers(items []usagestats.UserLeaderboardItem) []usagestats.LeaderboardDailyRewardTopUser {
	topUsers := make([]usagestats.LeaderboardDailyRewardTopUser, 0, leaderboardRewardTopLimit)
	for _, item := range items {
		if item.Rank <= 0 || item.Rank > leaderboardRewardTopLimit {
			continue
		}
		topUsers = append(topUsers, usagestats.LeaderboardDailyRewardTopUser{
			Rank:        item.Rank,
			UserID:      item.UserID,
			DisplayName: item.DisplayName,
			EmailMasked: item.EmailMasked,
			AvatarURL:   item.AvatarURL,
			Username:    item.Username,
			Email:       item.Email,
			Tokens:      item.Tokens,
			ActualCost:  item.ActualCost,
		})
	}
	return topUsers
}

func resolveLeaderboardDailyRewardClaimState(status *usagestats.LeaderboardDailyRewards) (bool, string) {
	if status == nil {
		return false, leaderboardRewardReasonDisabled
	}
	if status.Claimed {
		return false, leaderboardRewardReasonAlreadyClaimed
	}
	if !status.Enabled || status.RewardMode == LeaderboardDailyRewardModeDisabled {
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
	if status.CurrentUserRank > leaderboardRewardTopLimit {
		return false, leaderboardRewardReasonNotTopTen
	}

	switch status.RewardMode {
	case LeaderboardDailyRewardModeRedPacket:
		if status.RedPacketPoolAmount <= 0 {
			return false, leaderboardRewardReasonZeroReward
		}
	case LeaderboardDailyRewardModeLottery:
		if status.LotteryAmount <= 0 {
			return false, leaderboardRewardReasonZeroReward
		}
		if status.LotteryWinnerRank == nil {
			return false, leaderboardRewardReasonLotteryNotDrawn
		}
		if *status.LotteryWinnerRank <= 0 || status.CurrentUserRank != *status.LotteryWinnerRank {
			return false, leaderboardRewardReasonLotteryNotWinner
		}
	default:
		return false, leaderboardRewardReasonDisabled
	}
	return true, leaderboardRewardReasonEligible
}

func leaderboardRewardRedeemCode(rewardDate, rewardMode string, userID int64) string {
	normalizedDate := strings.NewReplacer("-", "", "~", "").Replace(rewardDate)
	modePart := "LDR"
	switch rewardMode {
	case LeaderboardDailyRewardModeLottery:
		modePart = "LDL"
	case LeaderboardDailyRewardModeRedPacket:
		modePart = "LDP"
	}
	return modePart + normalizedDate + "U" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func splitLeaderboardRedPacketAmounts(rewardDate string, total float64, count int, minAmount float64, maxAmount float64) []float64 {
	if count <= 0 {
		return nil
	}
	amounts := make([]float64, count)
	if total <= 0 {
		return amounts
	}
	minAmount = roundLeaderboardRewardAmount(math.Max(0, minAmount))
	maxAmount = roundLeaderboardRewardAmount(math.Max(0, maxAmount))
	if minAmount > 0 && minAmount*float64(count) > total {
		minAmount = roundLeaderboardRewardAmount(total / float64(count))
	}
	if maxAmount > 0 && maxAmount*float64(count) < total {
		maxAmount = roundLeaderboardRewardAmount(total / float64(count))
	}
	if maxAmount > 0 && maxAmount < minAmount {
		maxAmount = minAmount
	}
	remainingTotal := total
	if minAmount > 0 && total >= minAmount*float64(count) {
		for i := range amounts {
			amounts[i] = minAmount
		}
		remainingTotal = roundLeaderboardRewardAmount(total - minAmount*float64(count))
	}
	weights := make([]uint64, count)
	var totalWeight uint64
	for i := 0; i < count; i++ {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", rewardDate, i+1)))
		weight := binary.BigEndian.Uint64(sum[:8])%10000 + 1
		weights[i] = weight
		totalWeight += weight
	}
	var allocated float64
	for i := 0; i < count-1; i++ {
		amount := roundLeaderboardRewardAmount(remainingTotal * float64(weights[i]) / float64(totalWeight))
		if maxAmount > 0 {
			amount = math.Min(amount, math.Max(0, maxAmount-amounts[i]))
		}
		amount = roundLeaderboardRewardAmount(amount)
		amounts[i] = roundLeaderboardRewardAmount(amounts[i] + amount)
		allocated = roundLeaderboardRewardAmount(allocated + amount)
	}
	lastExtra := roundLeaderboardRewardAmount(math.Max(0, remainingTotal-allocated))
	if maxAmount > 0 {
		lastExtra = math.Min(lastExtra, math.Max(0, maxAmount-amounts[count-1]))
	}
	amounts[count-1] = roundLeaderboardRewardAmount(amounts[count-1] + lastExtra)
	allocated = roundLeaderboardRewardAmount(allocated + lastExtra)
	leftover := roundLeaderboardRewardAmount(remainingTotal - allocated)
	for i := 0; i < count && leftover > 0; i++ {
		room := leftover
		if maxAmount > 0 {
			room = math.Max(0, maxAmount-amounts[i])
		}
		if room <= 0 {
			continue
		}
		add := roundLeaderboardRewardAmount(math.Min(room, leftover))
		amounts[i] = roundLeaderboardRewardAmount(amounts[i] + add)
		leftover = roundLeaderboardRewardAmount(leftover - add)
	}
	return amounts
}

func roundLeaderboardRewardAmount(value float64) float64 {
	return math.Round(value*1e8) / 1e8
}

func leaderboardInt64Ptr(value int64) *int64 {
	return &value
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func firstNonEmptyLeaderboardReward(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func selectLeaderboardLotteryWinner(rewardDate string, topUsers []usagestats.LeaderboardDailyRewardTopUser) usagestats.LeaderboardDailyRewardTopUser {
	candidates := make([]usagestats.LeaderboardDailyRewardTopUser, 0, len(topUsers))
	for _, user := range topUsers {
		if user.UserID > 0 && user.Rank > 0 && user.Rank <= leaderboardRewardTopLimit {
			candidates = append(candidates, user)
		}
	}
	if len(candidates) == 0 {
		return usagestats.LeaderboardDailyRewardTopUser{}
	}
	sum := sha256.Sum256([]byte("leaderboard-lottery:" + rewardDate))
	index := int(binary.BigEndian.Uint64(sum[:8]) % uint64(len(candidates)))
	return candidates[index]
}

func leaderboardLotteryDrawAt(periodEnd time.Time, cronExpr string) time.Time {
	loc := timezone.Location()
	periodEnd = periodEnd.In(loc)
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(strings.TrimSpace(cronExpr))
	if err != nil {
		return periodEnd.Add(leaderboardRewardSettlementDelay)
	}
	return schedule.Next(periodEnd.Add(-time.Nanosecond))
}

func isValidLeaderboardLotteryCron(cronExpr string) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(strings.TrimSpace(cronExpr))
	return err == nil
}
