package service

import (
	"context"
	"testing"
)

func TestAccountCanBeUsedByUser_SystemAccountAllowed(t *testing.T) {
	account := &Account{ID: 1}

	if !account.CanBeUsedByUser(42) {
		t.Fatal("system account should be usable by any user")
	}
}

func TestAccountCanBeUsedByUser_PrivateOwnerOnly(t *testing.T) {
	ownerID := int64(10)
	account := &Account{
		ID:          1,
		OwnerUserID: &ownerID,
		ShareMode:   AccountShareModePrivate,
		ShareStatus: AccountShareStatusNotShared,
	}

	if !account.CanBeUsedByUser(ownerID) {
		t.Fatal("owner should be able to use private account")
	}
	if account.CanBeUsedByUser(11) {
		t.Fatal("other users must not use private account")
	}
}

func TestAccountCanBeUsedByUser_PublicRequiresActiveReview(t *testing.T) {
	ownerID := int64(10)
	account := &Account{
		ID:          1,
		OwnerUserID: &ownerID,
		ShareMode:   AccountShareModePublic,
		ShareStatus: AccountShareStatusPendingReview,
	}

	if !account.CanBeUsedByUser(ownerID) {
		t.Fatal("owner should keep access while public account is pending review")
	}
	if account.CanBeUsedByUser(11) {
		t.Fatal("other users must not use public account before review approval")
	}

	account.ShareStatus = AccountShareStatusActive
	if !account.CanBeUsedByUser(11) {
		t.Fatal("other users should use active public account")
	}
}

func TestShouldSkipBillingForSelfOwnedPrivateAccount(t *testing.T) {
	ownerID := int64(10)
	privateAccount := &Account{
		ID:          1,
		OwnerUserID: &ownerID,
		ShareMode:   AccountShareModePrivate,
		ShareStatus: AccountShareStatusNotShared,
	}
	publicAccount := &Account{
		ID:          2,
		OwnerUserID: &ownerID,
		ShareMode:   AccountShareModePublic,
		ShareStatus: AccountShareStatusActive,
	}

	if !ShouldSkipBillingForSelfOwnedPrivateAccount(ownerID, privateAccount) {
		t.Fatal("self-owned private account usage should skip platform billing")
	}
	if ShouldSkipBillingForSelfOwnedPrivateAccount(11, privateAccount) {
		t.Fatal("other users must not skip billing for someone else's private account")
	}
	if ShouldSkipBillingForSelfOwnedPrivateAccount(ownerID, publicAccount) {
		t.Fatal("public shared account usage should stay billable")
	}
}

func TestBuildUsageBillingCommand_SelfOwnedPrivateSkipsUserAndAPIKeyBilling(t *testing.T) {
	ownerID := int64(10)
	accountRate := 1.25
	cmd := buildUsageBillingCommand("req-private", &UsageLog{
		Model:       "claude-3-5-sonnet",
		BillingType: BillingTypeBalance,
	}, &postUsageBillingParams{
		Cost: &CostBreakdown{
			TotalCost:  2,
			ActualCost: 3,
		},
		User: &User{ID: ownerID},
		APIKey: &APIKey{
			ID:          100,
			Quota:       10,
			RateLimit5h: 10,
		},
		Account: &Account{
			ID:          200,
			Type:        AccountTypeAPIKey,
			OwnerUserID: &ownerID,
			ShareMode:   AccountShareModePrivate,
			ShareStatus: AccountShareStatusNotShared,
			Extra: map[string]any{
				"quota_limit": 100,
			},
		},
		AccountRateMultiplier: accountRate,
		APIKeyService:         noopAPIKeyQuotaUpdater{},
	})

	if cmd == nil {
		t.Fatal("expected billing command")
	}
	if cmd.BalanceCost != 0 {
		t.Fatalf("self-owned private account should not deduct balance, got %v", cmd.BalanceCost)
	}
	if cmd.SubscriptionCost != 0 {
		t.Fatalf("self-owned private account should not deduct subscription, got %v", cmd.SubscriptionCost)
	}
	if cmd.APIKeyQuotaCost != 0 {
		t.Fatalf("self-owned private account should not deduct api key quota, got %v", cmd.APIKeyQuotaCost)
	}
	if cmd.APIKeyRateLimitCost != 0 {
		t.Fatalf("self-owned private account should not update api key rate limits, got %v", cmd.APIKeyRateLimitCost)
	}
	if cmd.AccountQuotaCost != 2*accountRate {
		t.Fatalf("account quota stats should still be updated, got %v", cmd.AccountQuotaCost)
	}
}

func TestBuildUsageBillingCommand_PublicSharedAccountAddsOwnerShare(t *testing.T) {
	ownerID := int64(10)
	cmd := buildUsageBillingCommand("req-public", &UsageLog{
		Model:       "gpt-5.2",
		BillingType: BillingTypeBalance,
	}, &postUsageBillingParams{
		Cost: &CostBreakdown{
			TotalCost:  1,
			ActualCost: 1.5,
		},
		User:   &User{ID: 11},
		APIKey: &APIKey{ID: 100},
		Account: &Account{
			ID:          200,
			Type:        AccountTypeOAuth,
			OwnerUserID: &ownerID,
			ShareMode:   AccountShareModePublic,
			ShareStatus: AccountShareStatusActive,
		},
		AccountShareEnabled:          true,
		AccountShareOwnerRatePercent: 80,
		AccountShareFreezeHours:      72,
	})

	if cmd == nil {
		t.Fatal("expected billing command")
	}
	if cmd.AccountShareOwnerUserID != ownerID {
		t.Fatalf("expected share owner %d, got %d", ownerID, cmd.AccountShareOwnerUserID)
	}
	if cmd.AccountShareOwnerRatePercent != 80 {
		t.Fatalf("expected owner rate 80, got %v", cmd.AccountShareOwnerRatePercent)
	}
	if cmd.AccountShareFreezeHours != 72 {
		t.Fatalf("expected freeze hours 72, got %d", cmd.AccountShareFreezeHours)
	}
	if cmd.BalanceCost != 1.5 {
		t.Fatalf("public shared usage by another user remains billable, got balance cost %v", cmd.BalanceCost)
	}
}

type noopAPIKeyQuotaUpdater struct{}

func (noopAPIKeyQuotaUpdater) UpdateQuotaUsed(context.Context, int64, float64) error {
	return nil
}

func (noopAPIKeyQuotaUpdater) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return nil
}
