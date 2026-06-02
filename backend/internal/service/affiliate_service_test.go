//go:build unit

package service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestResolveRebateRatePercent_PerUserOverride verifies that per-inviter
// AffRebateRatePercent overrides the global rate, that NULL falls back to the
// global rate, and that out-of-range exclusive rates are clamped silently.
//
// SettingService is left nil here so globalRebateRatePercent returns the
// documented default (AffiliateRebateRateDefault = 20%) — this exercises the
// fallback path without spinning up a settings stub.
func TestResolveRebateRatePercent_PerUserOverride(t *testing.T) {
	t.Parallel()
	svc := &AffiliateService{}

	// nil exclusive rate → falls back to global default (20%)
	require.InDelta(t, AffiliateRebateRateDefault,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{}), 1e-9)

	// exclusive rate set → overrides global
	rate := 50.0
	require.InDelta(t, 50.0,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &rate}), 1e-9)

	// exclusive rate 0 → returns 0 (no rebate, intentional)
	zero := 0.0
	require.InDelta(t, 0.0,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &zero}), 1e-9)

	// exclusive rate above max → clamped to Max
	tooHigh := 250.0
	require.InDelta(t, AffiliateRebateRateMax,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooHigh}), 1e-9)

	// exclusive rate below min → clamped to Min
	tooLow := -5.0
	require.InDelta(t, AffiliateRebateRateMin,
		svc.resolveRebateRatePercent(context.Background(), &AffiliateSummary{AffRebateRatePercent: &tooLow}), 1e-9)
}

// TestIsEnabled_NilSettingServiceReturnsDefault verifies that IsEnabled
// safely handles a nil settingService dependency by returning the default
// (off). This protects callers from nil-pointer crashes in misconfigured
// environments.
func TestIsEnabled_NilSettingServiceReturnsDefault(t *testing.T) {
	t.Parallel()
	svc := &AffiliateService{}
	require.False(t, svc.IsEnabled(context.Background()))
	require.Equal(t, AffiliateEnabledDefault, svc.IsEnabled(context.Background()))
}

func TestClaimInviteeAPICallReward_UsesConfiguredReward(t *testing.T) {
	t.Parallel()

	repo := &affiliateRewardRepoStub{}
	settings := NewSettingService(settingRepoMapStub{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateAPICallRewardAmount: "2.5",
		SettingKeyAffiliateRebateFreezeHours:   "6",
	}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)
	systemRepo := newFakeTicketRepo()
	svc.SetSystemTicketService(NewSystemTicketService(systemRepo))

	amount, err := svc.ClaimInviteeAPICallReward(context.Background(), 100, 200)
	require.NoError(t, err)
	require.Equal(t, 2.5, amount)
	require.Equal(t, int64(100), repo.inviterID)
	require.Equal(t, int64(200), repo.inviteeID)
	require.Equal(t, 2.5, repo.amount)
	require.Equal(t, 6, repo.freezeHours)

	notification := requireSystemTicketNotification(t, systemRepo, 100, SystemTicketEventAffiliateFirstAPIReward, "affiliate_first_api_reward:200")
	require.Equal(t, float64(200), notification.Metadata["invitee_user_id"])
	require.Equal(t, 2.5, notification.Metadata["amount"])
	require.Equal(t, false, notification.Metadata["claimable"])
}

func TestClaimInviteeAPICallReward_DisabledWhenRewardAmountZero(t *testing.T) {
	t.Parallel()

	repo := &affiliateRewardRepoStub{}
	settings := NewSettingService(settingRepoMapStub{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateAPICallRewardAmount: "0",
	}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)

	amount, err := svc.ClaimInviteeAPICallReward(context.Background(), 100, 200)
	require.ErrorIs(t, err, ErrAffiliateAPICallRewardNotEligible)
	require.Zero(t, amount)
	require.False(t, repo.called)
}

func TestClaimInviteeAPICallReward_BlocksRevokedSelfReferral(t *testing.T) {
	t.Parallel()

	repo := &affiliateRewardRepoStub{revokeSelfReferral: true}
	settings := NewSettingService(settingRepoMapStub{
		SettingKeyAffiliateEnabled:             "true",
		SettingKeyAffiliateAPICallRewardAmount: "5",
	}, nil)
	svc := NewAffiliateService(repo, settings, nil, nil)

	amount, err := svc.ClaimInviteeAPICallReward(context.Background(), 100, 200)
	require.ErrorIs(t, err, ErrAffiliateAPICallRewardNotEligible)
	require.Zero(t, amount)
	require.True(t, repo.revokeCalled)
	require.Equal(t, int64(100), repo.revokeInviterID)
	require.Equal(t, int64(200), repo.revokeInviteeUserID)
	require.False(t, repo.called)
}

// TestValidateExclusiveRate_BoundaryAndInvalid covers the validator used by
// admin-facing rate setters: nil is always valid (clear), in-range values
// are accepted, NaN/Inf and out-of-range values produce a typed BadRequest.
func TestValidateExclusiveRate_BoundaryAndInvalid(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateExclusiveRate(nil))

	for _, v := range []float64{0, 0.01, 50, 99.99, 100} {
		v := v
		require.NoError(t, validateExclusiveRate(&v), "value %v should be valid", v)
	}

	for _, v := range []float64{-0.01, 100.01, -100, 200} {
		v := v
		require.Error(t, validateExclusiveRate(&v), "value %v should be rejected", v)
	}

	nan := math.NaN()
	require.Error(t, validateExclusiveRate(&nan))
	posInf := math.Inf(1)
	require.Error(t, validateExclusiveRate(&posInf))
	negInf := math.Inf(-1)
	require.Error(t, validateExclusiveRate(&negInf))
}

func TestMaskEmail(t *testing.T) {
	t.Parallel()
	require.Equal(t, "a***@g***.com", maskEmail("alice@gmail.com"))
	require.Equal(t, "x***@d***", maskEmail("x@domain"))
	require.Equal(t, "", maskEmail(""))
}

func TestIsValidAffiliateCodeFormat(t *testing.T) {
	t.Parallel()

	// 邀请码格式校验同时服务于：
	// 1) 系统自动生成的 12 位随机码（A-Z 去 I/O，2-9 去 0/1）
	// 2) 管理员设置的自定义专属码（如 "VIP2026"、"NEW_USER-1"）
	// 因此校验放宽到 [A-Z0-9_-]{4,32}（要求调用方先 ToUpper）。
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid canonical 12-char", "ABCDEFGHJKLM", true},
		{"valid all digits 2-9", "234567892345", true},
		{"valid mixed", "A2B3C4D5E6F7", true},
		{"valid admin custom short", "VIP1", true},
		{"valid admin custom with hyphen", "NEW-USER", true},
		{"valid admin custom with underscore", "VIP_2026", true},
		{"valid 32-char max", "ABCDEFGHIJKLMNOPQRSTUVWXYZ012345", true},
		// Previously-excluded chars (I/O/0/1) are now allowed since admins may use them.
		{"letter I now allowed", "IBCDEFGHJKLM", true},
		{"letter O now allowed", "OBCDEFGHJKLM", true},
		{"digit 0 now allowed", "0BCDEFGHJKLM", true},
		{"digit 1 now allowed", "1BCDEFGHJKLM", true},
		{"too short (3 chars)", "ABC", false},
		{"too long (33 chars)", "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456", false},
		{"lowercase rejected (caller must ToUpper first)", "abcdefghjklm", false},
		{"empty", "", false},
		{"utf8 non-ascii", "ÄÄÄÄÄÄ", false}, // bytes out of charset
		{"ascii punctuation .", "ABCDEFGHJK.M", false},
		{"whitespace", "ABCDEFGHJK M", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, isValidAffiliateCodeFormat(tc.in))
		})
	}
}

type affiliateRewardRepoStub struct {
	called              bool
	inviterID           int64
	inviteeID           int64
	amount              float64
	freezeHours         int
	revokeSelfReferral  bool
	revokeCalled        bool
	revokeInviterID     int64
	revokeInviteeUserID int64
}

func (r *affiliateRewardRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	return &AffiliateSummary{UserID: userID}, nil
}

func (r *affiliateRewardRepoStub) GetAffiliateByCode(context.Context, string) (*AffiliateSummary, error) {
	panic("unexpected GetAffiliateByCode call")
}

func (r *affiliateRewardRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	panic("unexpected BindInviter call")
}

func (r *affiliateRewardRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	panic("unexpected AccrueQuota call")
}

func (r *affiliateRewardRepoStub) RevokeSelfReferralByPaymentMethod(_ context.Context, inviterID, inviteeUserID int64, _ *int64) (bool, error) {
	r.revokeCalled = true
	r.revokeInviterID = inviterID
	r.revokeInviteeUserID = inviteeUserID
	return r.revokeSelfReferral, nil
}

func (r *affiliateRewardRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	panic("unexpected GetAccruedRebateFromInvitee call")
}

func (r *affiliateRewardRepoStub) ThawFrozenQuota(context.Context, int64) (float64, error) {
	panic("unexpected ThawFrozenQuota call")
}

func (r *affiliateRewardRepoStub) TransferQuotaToBalance(context.Context, int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}

func (r *affiliateRewardRepoStub) ListInvitees(context.Context, int64, int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}

func (r *affiliateRewardRepoStub) ClaimAPICallReward(_ context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int) (bool, error) {
	r.called = true
	r.inviterID = inviterID
	r.inviteeID = inviteeUserID
	r.amount = amount
	r.freezeHours = freezeHours
	return true, nil
}

func (r *affiliateRewardRepoStub) UpdateUserAffCode(context.Context, int64, string) error {
	panic("unexpected UpdateUserAffCode call")
}

func (r *affiliateRewardRepoStub) ResetUserAffCode(context.Context, int64) (string, error) {
	panic("unexpected ResetUserAffCode call")
}

func (r *affiliateRewardRepoStub) SetUserRebateRate(context.Context, int64, *float64) error {
	panic("unexpected SetUserRebateRate call")
}

func (r *affiliateRewardRepoStub) BatchSetUserRebateRate(context.Context, []int64, *float64) error {
	panic("unexpected BatchSetUserRebateRate call")
}

func (r *affiliateRewardRepoStub) ListUsersWithCustomSettings(context.Context, AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	panic("unexpected ListUsersWithCustomSettings call")
}

func (r *affiliateRewardRepoStub) ListAffiliateInviteRecords(context.Context, AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error) {
	panic("unexpected ListAffiliateInviteRecords call")
}

func (r *affiliateRewardRepoStub) ListAffiliateRebateRecords(context.Context, AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error) {
	panic("unexpected ListAffiliateRebateRecords call")
}

func (r *affiliateRewardRepoStub) ListAffiliateTransferRecords(context.Context, AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error) {
	panic("unexpected ListAffiliateTransferRecords call")
}

func (r *affiliateRewardRepoStub) GetAffiliateUserOverview(context.Context, int64) (*AffiliateUserOverview, error) {
	panic("unexpected GetAffiliateUserOverview call")
}

type settingRepoMapStub map[string]string

func (r settingRepoMapStub) Get(ctx context.Context, key string) (*Setting, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Setting{Key: key, Value: value, UpdatedAt: now}, nil
}

func (r settingRepoMapStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r[key]
	if !ok {
		return "", errors.New("setting not found")
	}
	return value, nil
}

func (r settingRepoMapStub) Set(_ context.Context, key, value string) error {
	r[key] = value
	return nil
}

func (r settingRepoMapStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r[key]
	}
	return out, nil
}

func (r settingRepoMapStub) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		r[key] = value
	}
	return nil
}

func (r settingRepoMapStub) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r))
	for key, value := range r {
		out[key] = value
	}
	return out, nil
}

func (r settingRepoMapStub) Delete(_ context.Context, key string) error {
	delete(r, key)
	return nil
}
