package service

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type membershipRepoFake struct {
	monthlyPaid float64
	grants      map[string]*MembershipGrant
	revoked     []int64
	nextID      int64
}

func newMembershipRepoFake() *membershipRepoFake {
	return &membershipRepoFake{
		grants: make(map[string]*MembershipGrant),
		nextID: 1,
	}
}

func (r *membershipRepoFake) GetMonthlyNetPaid(context.Context, int64, time.Time, time.Time) (float64, error) {
	return r.monthlyPaid, nil
}

func (r *membershipRepoFake) UpsertAutoGrant(_ context.Context, grant *MembershipGrant) (*MembershipGrant, bool, error) {
	key := fmt.Sprintf("%d:%s:%s:%s", grant.UserID, grant.Source, grant.PeriodKey, grant.Tier)
	if existing := r.grants[key]; existing != nil {
		existing.QualifiedAmount = grant.QualifiedAmount
		if existing.Status != MembershipGrantStatusActive {
			existing.StartsAt = grant.StartsAt
			existing.ExpiresAt = grant.ExpiresAt
			existing.Status = MembershipGrantStatusActive
			existing.RevokedAt = nil
			existing.RevokeReason = nil
		}
		if existing.SourceOrderID == nil {
			existing.SourceOrderID = grant.SourceOrderID
		}
		if existing.SubscriptionGroupID == nil {
			existing.SubscriptionGroupID = grant.SubscriptionGroupID
		}
		cp := *existing
		return &cp, false, nil
	}
	cp := *grant
	cp.ID = r.nextID
	r.nextID++
	r.grants[key] = &cp
	return &cp, true, nil
}

func (r *membershipRepoFake) UpdateGrantSubscription(_ context.Context, grantID, subscriptionID, groupID int64) error {
	for _, grant := range r.grants {
		if grant.ID == grantID {
			grant.SubscriptionID = &subscriptionID
			grant.SubscriptionGroupID = &groupID
			return nil
		}
	}
	return nil
}

func (r *membershipRepoFake) ListActiveAutoGrants(_ context.Context, userID int64, now time.Time) ([]MembershipGrant, error) {
	var out []MembershipGrant
	for _, grant := range r.grants {
		if grant.UserID == userID && grant.Status == MembershipGrantStatusActive && !grant.StartsAt.After(now) && grant.ExpiresAt.After(now) {
			out = append(out, *grant)
		}
	}
	return out, nil
}

func (r *membershipRepoFake) GetActiveHighestAutoGrant(ctx context.Context, userID int64, now time.Time) (*MembershipGrant, error) {
	grants, err := r.ListActiveAutoGrants(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	var best *MembershipGrant
	for i := range grants {
		grant := grants[i]
		if best == nil || membershipTierRank(grant.Tier) > membershipTierRank(best.Tier) {
			best = &grant
		}
	}
	return best, nil
}

func (r *membershipRepoFake) RevokeGrant(_ context.Context, grantID int64, reason string) error {
	r.revoked = append(r.revoked, grantID)
	now := time.Now()
	for _, grant := range r.grants {
		if grant.ID == grantID {
			grant.Status = MembershipGrantStatusRevoked
			grant.RevokedAt = &now
			grant.RevokeReason = &reason
		}
	}
	return nil
}

type settingRepoFake struct {
	value string
}

func (r *settingRepoFake) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}

func (r *settingRepoFake) GetValue(context.Context, string) (string, error) {
	if r.value == "" {
		return "", ErrSettingNotFound
	}
	return r.value, nil
}

func (r *settingRepoFake) Set(_ context.Context, _ string, value string) error {
	r.value = value
	return nil
}

func (r *settingRepoFake) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *settingRepoFake) SetMultiple(context.Context, map[string]string) error {
	return nil
}

func (r *settingRepoFake) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *settingRepoFake) Delete(context.Context, string) error {
	return nil
}

func membershipTestSettings() MembershipSettings {
	settings := DefaultMembershipSettings()
	settings.Enabled = true
	for i := range settings.Tiers {
		switch settings.Tiers[i].Level {
		case MembershipTierVIP:
			settings.Tiers[i].SubscriptionGroupID = 10
		case MembershipTierSVIP:
			settings.Tiers[i].SubscriptionGroupID = 20
		}
	}
	return settings
}

func TestMembershipDefaultsAreDisabledForUpgradeCompatibility(t *testing.T) {
	settings := DefaultMembershipSettings()
	if settings.Enabled {
		t.Fatalf("DefaultMembershipSettings().Enabled = true, want false")
	}

	normal := settings.tierByLevel(MembershipTierNormal)
	if normal.RPMLimit != 60 || normal.TPMLimit != 60000 || normal.ImageActiveTasks != 2 {
		t.Fatalf("normal default benefits changed unexpectedly: %+v", normal)
	}
}

func TestMembershipDisabledDoesNotApplyRateOrThrottleLimits(t *testing.T) {
	ctx := context.Background()
	svc := NewMembershipService(newMembershipRepoFake(), &settingRepoFake{}, nil, nil, nil)

	benefits, err := svc.GetEffectiveBenefits(ctx, 1)
	if err != nil {
		t.Fatalf("GetEffectiveBenefits() error = %v", err)
	}
	if benefits.RateMultiplier != 0 || benefits.RPMLimit != 0 || benefits.TPMLimit != 0 {
		t.Fatalf("disabled membership benefits = %+v, want no rate/rpm/tpm enforcement", benefits)
	}
	if benefits.ImageActiveTasks != 2 {
		t.Fatalf("disabled membership image active tasks = %d, want default 2", benefits.ImageActiveTasks)
	}
	if got := svc.ApplyRateMultiplier(ctx, 1, 0.75); got != 0.75 {
		t.Fatalf("ApplyRateMultiplier() = %v, want unchanged 0.75", got)
	}
}

func TestMembershipRecalculateGrantsHighestTierAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := newMembershipRepoFake()
	svc := NewMembershipService(repo, &settingRepoFake{}, nil, nil, nil)
	if _, err := svc.UpdateSettings(ctx, membershipTestSettings()); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	repo.monthlyPaid = 120
	if err := svc.RecalculateForUser(ctx, 1, 1001, "payment_completed"); err != nil {
		t.Fatalf("RecalculateForUser() error = %v", err)
	}
	if err := svc.RecalculateForUser(ctx, 1, 1001, "payment_completed"); err != nil {
		t.Fatalf("duplicate RecalculateForUser() error = %v", err)
	}
	if len(repo.grants) != 1 {
		t.Fatalf("grant count = %d, want 1", len(repo.grants))
	}
	status, err := svc.GetStatus(ctx, 1)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.CurrentTier != MembershipTierSVIP {
		t.Fatalf("CurrentTier = %s, want %s", status.CurrentTier, MembershipTierSVIP)
	}
}

func TestMembershipRecalculateDowngradesAfterRefund(t *testing.T) {
	ctx := context.Background()
	repo := newMembershipRepoFake()
	svc := NewMembershipService(repo, &settingRepoFake{}, nil, nil, nil)
	if _, err := svc.UpdateSettings(ctx, membershipTestSettings()); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	repo.monthlyPaid = 120
	if err := svc.RecalculateForUser(ctx, 1, 1001, "payment_completed"); err != nil {
		t.Fatalf("RecalculateForUser() error = %v", err)
	}
	repo.monthlyPaid = 60
	if err := svc.RecalculateForUser(ctx, 1, 1001, "refund_completed"); err != nil {
		t.Fatalf("refund RecalculateForUser() error = %v", err)
	}
	status, err := svc.GetStatus(ctx, 1)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.CurrentTier != MembershipTierVIP {
		t.Fatalf("CurrentTier = %s, want %s", status.CurrentTier, MembershipTierVIP)
	}
	if len(repo.revoked) != 1 {
		t.Fatalf("revoked count = %d, want 1", len(repo.revoked))
	}
}

func TestMembershipRecalculateUsesSpecifiedPeriod(t *testing.T) {
	ctx := context.Background()
	repo := newMembershipRepoFake()
	svc := NewMembershipService(repo, &settingRepoFake{}, nil, nil, nil)
	if _, err := svc.UpdateSettings(ctx, membershipTestSettings()); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	periodAt := time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC)
	repo.monthlyPaid = 60
	if err := svc.RecalculateForUserAt(ctx, 1, 1001, "payment_completed", periodAt); err != nil {
		t.Fatalf("RecalculateForUserAt() error = %v", err)
	}
	status, err := svc.GetStatus(ctx, 1)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Grant == nil || status.Grant.PeriodKey != "2026-04" || status.CurrentTier != MembershipTierVIP {
		t.Fatalf("status after old-period grant = %+v", status)
	}

	repo.monthlyPaid = 0
	if err := svc.RecalculateForUserAt(ctx, 1, 1001, "refund_completed", periodAt); err != nil {
		t.Fatalf("refund RecalculateForUserAt() error = %v", err)
	}
	status, err = svc.GetStatus(ctx, 1)
	if err != nil {
		t.Fatalf("GetStatus() after refund error = %v", err)
	}
	if status.CurrentTier != MembershipTierNormal {
		t.Fatalf("CurrentTier = %s, want %s", status.CurrentTier, MembershipTierNormal)
	}
	if len(repo.revoked) != 1 {
		t.Fatalf("revoked count = %d, want 1", len(repo.revoked))
	}
}
