package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type affiliateRiskFreezeRepoStub struct {
	activeRiskFreeze bool
	claimCalled      bool
	transferCalled   bool
}

func (r *affiliateRiskFreezeRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	return &AffiliateSummary{UserID: userID}, nil
}

func (r *affiliateRiskFreezeRepoStub) GetAffiliateByCode(context.Context, string) (*AffiliateSummary, error) {
	return nil, ErrAffiliateProfileNotFound
}

func (r *affiliateRiskFreezeRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (r *affiliateRiskFreezeRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	return false, nil
}

func (r *affiliateRiskFreezeRepoStub) RevokeSelfReferralByPaymentMethod(context.Context, int64, int64, *int64) (bool, error) {
	return false, nil
}

func (r *affiliateRiskFreezeRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	return 0, nil
}

func (r *affiliateRiskFreezeRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	return 0, nil
}

func (r *affiliateRiskFreezeRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	r.transferCalled = true
	return 1.25, 9.5, nil
}

func (r *affiliateRiskFreezeRepoStub) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	return nil, nil
}

func (r *affiliateRiskFreezeRepoStub) ClaimAPICallReward(context.Context, int64, int64, float64, int) (bool, error) {
	r.claimCalled = true
	return true, nil
}

func (r *affiliateRiskFreezeRepoStub) HasActiveRiskFreeze(context.Context, int64) (bool, error) {
	return r.activeRiskFreeze, nil
}

func (r *affiliateRiskFreezeRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	return nil
}

func (r *affiliateRiskFreezeRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	return "", nil
}

func (r *affiliateRiskFreezeRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	return nil
}

func (r *affiliateRiskFreezeRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	return nil
}

func (r *affiliateRiskFreezeRepoStub) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}

func (r *affiliateRiskFreezeRepoStub) ListAffiliateInviteRecords(context.Context, AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error) {
	return nil, 0, nil
}

func (r *affiliateRiskFreezeRepoStub) ListAffiliateRebateRecords(context.Context, AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error) {
	return nil, 0, nil
}

func (r *affiliateRiskFreezeRepoStub) ListAffiliateTransferRecords(context.Context, AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error) {
	return nil, 0, nil
}

func (r *affiliateRiskFreezeRepoStub) GetAffiliateUserOverview(context.Context, int64) (*AffiliateUserOverview, error) {
	return nil, nil
}

func TestAffiliateRiskFreezeBlocksTransfer(t *testing.T) {
	t.Parallel()

	repo := &affiliateRiskFreezeRepoStub{activeRiskFreeze: true}
	svc := NewAffiliateService(repo, nil, nil, nil)

	transferred, balance, err := svc.TransferAffiliateQuota(context.Background(), 100)
	require.ErrorIs(t, err, ErrAffiliateRiskFrozen)
	require.Zero(t, transferred)
	require.Zero(t, balance)
	require.False(t, repo.transferCalled)
}

func TestAffiliateRiskFreezeBlocksAPICallRewardClaim(t *testing.T) {
	t.Parallel()

	repo := &affiliateRiskFreezeRepoStub{activeRiskFreeze: true}
	settings := NewSettingService(affiliateRiskSettingRepoStub{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateAPICallRewardAmount: "5",
	}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)

	amount, err := svc.ClaimInviteeAPICallReward(context.Background(), 100, 200)
	require.ErrorIs(t, err, ErrAffiliateRiskFrozen)
	require.Zero(t, amount)
	require.False(t, repo.claimCalled)
}
