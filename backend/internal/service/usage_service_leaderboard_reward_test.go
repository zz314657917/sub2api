package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	apptimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type leaderboardRewardUsageRepo struct {
	UsageLogRepository
	response         *usagestats.UserLeaderboardResponse
	claim            *LeaderboardDailyRewardClaim
	createClaimErr   error
	createdClaims    []LeaderboardDailyRewardClaim
	attachedClaimID  int64
	attachedRedeemID int64
	leaderboardCalls []leaderboardRewardCall
}

type leaderboardRewardCall struct {
	start time.Time
	end   time.Time
	limit int
	user  int64
}

func (r *leaderboardRewardUsageRepo) GetUserLeaderboard(_ context.Context, start time.Time, end time.Time, limit int, userID int64) (*usagestats.UserLeaderboardResponse, error) {
	r.leaderboardCalls = append(r.leaderboardCalls, leaderboardRewardCall{
		start: start,
		end:   end,
		limit: limit,
		user:  userID,
	})
	if r.response == nil {
		return &usagestats.UserLeaderboardResponse{}, nil
	}
	return r.response, nil
}

func (r *leaderboardRewardUsageRepo) GetLeaderboardDailyRewardClaim(context.Context, string, int64) (*LeaderboardDailyRewardClaim, error) {
	if r.claim == nil {
		return nil, ErrLeaderboardDailyRewardClaimNotFound
	}
	claim := *r.claim
	return &claim, nil
}

func (r *leaderboardRewardUsageRepo) CreateLeaderboardDailyRewardClaim(_ context.Context, claim *LeaderboardDailyRewardClaim) error {
	if r.createClaimErr != nil {
		return r.createClaimErr
	}
	claim.ID = int64(len(r.createdClaims) + 100)
	claim.CreatedAt = time.Date(2026, 5, 9, 1, 2, 3, 0, time.UTC)
	r.createdClaims = append(r.createdClaims, *claim)
	return nil
}

func (r *leaderboardRewardUsageRepo) AttachLeaderboardDailyRewardClaimRedeemCode(_ context.Context, claimID, redeemCodeID int64) error {
	r.attachedClaimID = claimID
	r.attachedRedeemID = redeemCodeID
	return nil
}

type leaderboardRewardSettingRepo struct {
	SettingRepository
	values map[string]string
}

func (r *leaderboardRewardSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r.values[key]
	}
	return out, nil
}

type leaderboardRewardUserRepo struct {
	UserRepository
	updates []float64
	grants  []float64
}

func (r *leaderboardRewardUserRepo) UpdateBalance(_ context.Context, _ int64, amount float64) error {
	r.updates = append(r.updates, amount)
	return nil
}

func (r *leaderboardRewardUserRepo) AddBalance(_ context.Context, _ int64, amount float64) error {
	r.grants = append(r.grants, amount)
	return nil
}

type leaderboardRewardRedeemRepo struct {
	RedeemCodeRepository
	created []RedeemCode
}

func (r *leaderboardRewardRedeemRepo) Create(_ context.Context, code *RedeemCode) error {
	code.ID = int64(len(r.created) + 700)
	r.created = append(r.created, *code)
	return nil
}

func leaderboardRewardTestNow(t *testing.T) time.Time {
	t.Helper()
	require.NoError(t, apptimezone.Init("Asia/Shanghai"))
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	return time.Date(2026, 5, 9, 10, 0, 0, 0, loc)
}

func leaderboardRewardTxContext() context.Context {
	return dbent.NewTxContext(context.Background(), &dbent.Tx{})
}

func leaderboardRewardResponse(rank int64, total float64) *usagestats.UserLeaderboardResponse {
	entry := &usagestats.UserLeaderboardItem{
		Rank:          rank,
		UserID:        42,
		ActualCost:    12,
		Requests:      3,
		Tokens:        300,
		IsCurrentUser: true,
	}
	return &usagestats.UserLeaderboardResponse{
		TotalActualCost:  total,
		Ranking:          []usagestats.UserLeaderboardItem{*entry},
		CurrentUserEntry: entry,
	}
}

func leaderboardRewardSettings(enabled bool, minTotal, rank1, rank2, rank3 float64) *leaderboardRewardSettingRepo {
	return &leaderboardRewardSettingRepo{values: map[string]string{
		SettingKeyLeaderboardDailyRewardEnabled:            strconvBool(enabled),
		SettingKeyLeaderboardDailyRewardMinTotalActualCost: strconvFloat(minTotal),
		SettingKeyLeaderboardDailyRewardRank1Amount:        strconvFloat(rank1),
		SettingKeyLeaderboardDailyRewardRank2Amount:        strconvFloat(rank2),
		SettingKeyLeaderboardDailyRewardRank3Amount:        strconvFloat(rank3),
	}}
}

func strconvBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func strconvFloat(v float64) string {
	return fmt.Sprintf("%.8f", v)
}

func TestLeaderboardDailyRewardsDefaultDisabled(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(1, 20)}
	svc := NewUsageService(usageRepo, nil, nil, nil)

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.Equal(t, "2026-04-27~2026-05-03", got.RewardDate)
	require.False(t, got.Enabled)
	require.False(t, got.CanClaim)
	require.Equal(t, leaderboardRewardReasonDisabled, got.Reason)
}

func TestLeaderboardDailyRewardWindowUsesServerTimezone(t *testing.T) {
	require.NoError(t, apptimezone.Init("Asia/Shanghai"))
	now := time.Date(2026, 5, 10, 16, 30, 0, 0, time.UTC)

	start, end, rewardDate, settlementTZ, claimAvailableAt := leaderboardRewardWindow(now)

	require.Equal(t, "Asia/Shanghai", settlementTZ)
	require.Equal(t, "2026-05-04~2026-05-10", rewardDate)
	require.Equal(t, time.Date(2026, 5, 4, 0, 0, 0, 0, apptimezone.Location()), start)
	require.Equal(t, time.Date(2026, 5, 11, 0, 0, 0, 0, apptimezone.Location()), end)
	require.Equal(t, time.Date(2026, 5, 11, 0, 30, 0, 0, apptimezone.Location()), claimAvailableAt)
}

func TestLeaderboardDailyRewardsWaitsForSettlementDelay(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(1, 101)}
	svc := NewUsageService(usageRepo, nil, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), nil)
	now := time.Date(2026, 5, 4, 0, 10, 0, 0, apptimezone.Location())

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, now)

	require.NoError(t, err)
	require.False(t, got.SettlementReady)
	require.False(t, got.CanClaim)
	require.Equal(t, leaderboardRewardReasonSettling, got.Reason)
	require.Equal(t, "2026-05-04T00:30:00+08:00", got.ClaimAvailableAt)
}

func TestLeaderboardDailyRewardsRequiresStrictlyGreaterThanThreshold(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(1, 100)}
	svc := NewUsageService(usageRepo, nil, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), nil)

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.False(t, got.ThresholdMet)
	require.False(t, got.CanClaim)
	require.Equal(t, leaderboardRewardReasonThresholdNotMet, got.Reason)
}

func TestLeaderboardDailyRewardsTopThreeCanClaimWhenThresholdMet(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(2, 101)}
	svc := NewUsageService(usageRepo, nil, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), nil)

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.True(t, got.ThresholdMet)
	require.True(t, got.CanClaim)
	require.Equal(t, int64(2), got.CurrentUserRank)
	require.Equal(t, 3.0, got.CurrentUserRewardAmount)
	require.Equal(t, leaderboardRewardReasonEligible, got.Reason)
}

func TestLeaderboardDailyRewardsIncludesLastWeekTopUsers(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			TotalActualCost: 101,
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 11, Username: "Alpha Winner", Email: "alpha@example.com", Tokens: 300},
				{Rank: 2, UserID: 42, Username: "Beta Winner", Email: "beta@example.com", Tokens: 200, IsCurrentUser: true},
				{Rank: 3, UserID: 33, Username: "Gamma Winner", Email: "gamma@example.com", Tokens: 100},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 2, UserID: 42, Username: "Beta Winner", Email: "beta@example.com", Tokens: 200, IsCurrentUser: true},
		},
	}
	svc := NewUsageService(usageRepo, nil, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), nil)

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.Equal(t, []usagestats.LeaderboardDailyRewardTopUser{
		{Rank: 1, UserID: 11, Username: "Alpha Winner", Email: "alpha@example.com"},
		{Rank: 2, UserID: 42, Username: "Beta Winner", Email: "beta@example.com"},
		{Rank: 3, UserID: 33, Username: "Gamma Winner", Email: "gamma@example.com"},
	}, got.TopUsers)
}

func TestLeaderboardDailyRewardsNonTopThreeCannotClaim(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(4, 101)}
	svc := NewUsageService(usageRepo, nil, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), nil)

	got, err := svc.getLeaderboardDailyRewards(context.Background(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.False(t, got.CanClaim)
	require.Equal(t, leaderboardRewardReasonNotTopThree, got.Reason)
}

func TestClaimLeaderboardDailyRewardAddsBalanceAndAuditRecord(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(1, 101)}
	userRepo := &leaderboardRewardUserRepo{}
	redeemRepo := &leaderboardRewardRedeemRepo{}
	svc := NewUsageService(usageRepo, userRepo, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), redeemRepo)

	got, err := svc.claimLeaderboardDailyReward(leaderboardRewardTxContext(), 42, leaderboardRewardTestNow(t))

	require.NoError(t, err)
	require.Equal(t, 5.0, got.ClaimedAmount)
	require.Empty(t, userRepo.updates)
	require.Equal(t, []float64{5}, userRepo.grants)
	require.Len(t, redeemRepo.created, 1)
	require.Equal(t, RedeemTypeLeaderboardReward, redeemRepo.created[0].Type)
	require.Equal(t, StatusUsed, redeemRepo.created[0].Status)
	require.Equal(t, int64(42), *redeemRepo.created[0].UsedBy)
	require.Len(t, usageRepo.createdClaims, 1)
	require.Equal(t, int64(100), usageRepo.attachedClaimID)
	require.Equal(t, int64(700), usageRepo.attachedRedeemID)
	require.True(t, got.DailyRewards.Claimed)
}

func TestClaimLeaderboardDailyRewardDuplicateClaimConflictsBeforeSideEffects(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{
		response:       leaderboardRewardResponse(1, 101),
		createClaimErr: ErrLeaderboardDailyRewardAlreadyClaimed,
	}
	userRepo := &leaderboardRewardUserRepo{}
	redeemRepo := &leaderboardRewardRedeemRepo{}
	svc := NewUsageService(usageRepo, userRepo, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), redeemRepo)

	_, err := svc.claimLeaderboardDailyReward(leaderboardRewardTxContext(), 42, leaderboardRewardTestNow(t))

	require.True(t, errors.Is(err, ErrLeaderboardDailyRewardAlreadyClaimed))
	require.Empty(t, userRepo.updates)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)
}

func TestClaimLeaderboardDailyRewardRequiresTransactionProtection(t *testing.T) {
	usageRepo := &leaderboardRewardUsageRepo{response: leaderboardRewardResponse(1, 101)}
	userRepo := &leaderboardRewardUserRepo{}
	redeemRepo := &leaderboardRewardRedeemRepo{}
	svc := NewUsageService(usageRepo, userRepo, nil, nil)
	svc.SetLeaderboardRewardDependencies(leaderboardRewardSettings(true, 100, 5, 3, 1), redeemRepo)

	_, err := svc.claimLeaderboardDailyReward(context.Background(), 42, leaderboardRewardTestNow(t))

	require.True(t, errors.Is(err, ErrLeaderboardDailyRewardUnavailable))
	require.Empty(t, usageRepo.createdClaims)
	require.Empty(t, userRepo.updates)
	require.Empty(t, userRepo.grants)
	require.Empty(t, redeemRepo.created)
}
