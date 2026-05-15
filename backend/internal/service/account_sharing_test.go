package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
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

	if account.CanBeUsedByUser(ownerID) {
		t.Fatal("owner must not use public account before review approval")
	}
	if account.CanBeUsedByUser(11) {
		t.Fatal("other users must not use public account before review approval")
	}

	account.ShareStatus = AccountShareStatusActive
	if !account.CanBeUsedByUser(ownerID) {
		t.Fatal("owner should use active public account through the shared pool")
	}
	if !account.CanBeUsedByUser(11) {
		t.Fatal("other users should use active public account")
	}

	account.ShareStatus = AccountShareStatusSuspended
	if account.CanBeUsedByUser(ownerID) {
		t.Fatal("owner must not self-use suspended public account")
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

func TestUserAccountService_ListDisabledWhenAccountShareOff(t *testing.T) {
	svc := NewUserAccountService(nil, accountShareSettingsStub{enabled: false})

	_, _, err := svc.List(context.Background(), 10, pagination.PaginationParams{})
	if !errors.Is(err, ErrUserAccountShareDisabled) {
		t.Fatalf("expected ErrUserAccountShareDisabled, got %v", err)
	}
}

func TestUserAccountService_GetCapacityPoolsDisabledWhenAccountShareOff(t *testing.T) {
	svc := NewUserAccountService(nil, accountShareSettingsStub{enabled: false})

	_, err := svc.GetCapacityPools(context.Background(), 10)
	if !errors.Is(err, ErrUserAccountShareDisabled) {
		t.Fatalf("expected ErrUserAccountShareDisabled, got %v", err)
	}
}

func TestUserAccountService_GetCapacityPoolsSummarizesMineAndSharedPools(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	resetAt := time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC)
	repo := &capacityPoolAccountRepoStub{
		owned: []Account{
			{
				ID:          1,
				Name:        "my-openai-oauth",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &ownerID,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"codex_5h_used_percent":        35.5,
					"codex_5h_reset_after_seconds": 1800,
					"codex_5h_reset_at":            resetAt.Format(time.RFC3339),
					"codex_5h_window_minutes":      300,
					"codex_7d_used_percent":        90,
					"codex_7d_reset_after_seconds": 7200,
					"codex_7d_reset_at":            resetAt.Add(24 * time.Hour).Format(time.RFC3339),
					"codex_7d_window_minutes":      10080,
				},
			},
			{
				ID:          2,
				Name:        "my-apikey",
				Platform:    PlatformAnthropic,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &ownerID,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusDisabled,
				Schedulable: true,
				Extra: map[string]any{
					"quota_limit": 100,
					"quota_used":  25,
				},
			},
		},
		schedulable: []Account{
			{
				ID:          3,
				Name:        "system-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"codex_5h_used_percent": 12,
				},
			},
			{
				ID:          4,
				Name:        "other-public-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"codex_7d_used_percent": 60,
				},
			},
			{
				ID:          5,
				Name:        "other-private-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Mine.TotalAccounts != 2 {
		t.Fatalf("expected 2 owned accounts, got %d", pools.Mine.TotalAccounts)
	}
	if pools.Mine.ActiveAccounts != 1 {
		t.Fatalf("expected 1 active owned account, got %d", pools.Mine.ActiveAccounts)
	}
	if pools.Mine.SchedulableAccounts != 1 {
		t.Fatalf("expected 1 owned schedulable account, got %d", pools.Mine.SchedulableAccounts)
	}
	if pools.Mine.AbnormalAccounts != 1 || pools.Mine.DisabledAccounts != 1 {
		t.Fatalf("unexpected owned abnormal counters: abnormal=%d disabled=%d", pools.Mine.AbnormalAccounts, pools.Mine.DisabledAccounts)
	}
	if pools.Mine.ConfiguredQuota != 100 || pools.Mine.RemainingQuota != 75 {
		t.Fatalf("unexpected owned quota totals: configured=%v remaining=%v", pools.Mine.ConfiguredQuota, pools.Mine.RemainingQuota)
	}
	if len(pools.Mine.Sections) != 2 {
		t.Fatalf("expected 2 owned sections, got %d", len(pools.Mine.Sections))
	}
	openAIMine := findCapacityPoolSection(pools.Mine.Sections, PlatformOpenAI, AccountTypeOAuth)
	if openAIMine == nil || openAIMine.Windows["5h"].UsedPercent != 35.5 || openAIMine.Windows["7d"].ResetAfterSeconds != 7200 {
		t.Fatalf("unexpected owned OpenAI OAuth section: %#v", openAIMine)
	}

	if pools.Shared.TotalAccounts != 1 {
		t.Fatalf("expected 1 shared account, got %d", pools.Shared.TotalAccounts)
	}
	if pools.Shared.SchedulableAccounts != 1 {
		t.Fatalf("expected 1 shared schedulable account, got %d", pools.Shared.SchedulableAccounts)
	}
	sharedGroup := findCapacityPoolGroup(pools.Shared.Groups, 0, "未分组共享池")
	if sharedGroup == nil {
		t.Fatalf("expected ungrouped shared capacity group, got %#v", pools.Shared.Groups)
	}
	if sharedGroup.TotalAccounts != 1 || sharedGroup.SchedulableAccounts != 1 || sharedGroup.Status != "healthy" {
		t.Fatalf("unexpected shared group summary: %#v", sharedGroup)
	}
	sharedOpenAI := findCapacityPoolSection(pools.Shared.Sections, PlatformOpenAI, AccountTypeOAuth)
	if sharedOpenAI == nil {
		t.Fatal("expected shared OpenAI OAuth section")
	}
	if sharedOpenAI.TotalAccounts != 1 || sharedOpenAI.SchedulableAccounts != 1 {
		t.Fatalf("unexpected shared OpenAI totals: %#v", sharedOpenAI)
	}
	if _, ok := sharedOpenAI.Windows["5h"]; ok {
		t.Fatalf("expected system account 5h window to be excluded, got %#v", sharedOpenAI.Windows)
	}
	if sharedOpenAI.Windows["7d"].UsedPercent != 60 {
		t.Fatalf("unexpected shared OpenAI windows: %#v", sharedOpenAI.Windows)
	}
}

func TestUserAccountService_GetCapacityPoolsIncludesOnlyMarkedSystemSharedAccounts(t *testing.T) {
	ownerID := int64(10)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "system-default-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Schedulable: true,
			},
			{
				ID:          2,
				Name:        "system-shared-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"codex_5h_used_percent": 42,
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Shared.TotalAccounts != 1 {
		t.Fatalf("expected only marked system account in shared pool, got %d", pools.Shared.TotalAccounts)
	}
	sharedOpenAI := findCapacityPoolSection(pools.Shared.Sections, PlatformOpenAI, AccountTypeOAuth)
	if sharedOpenAI == nil || sharedOpenAI.TotalAccounts != 1 || sharedOpenAI.Windows["5h"].UsedPercent != 42 {
		t.Fatalf("unexpected marked system shared section: %#v", sharedOpenAI)
	}
}

func TestUserAccountService_GetCapacityPoolsPaginatesAllAccounts(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	owned := make([]Account, 0, 1002)
	for i := 0; i < 1002; i++ {
		owned = append(owned, Account{
			ID:          int64(i + 1),
			Name:        fmt.Sprintf("owned-%d", i),
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			OwnerUserID: &ownerID,
			Status:      StatusActive,
			Schedulable: true,
		})
	}
	shared := make([]Account, 0, 1003)
	for i := 0; i < 1003; i++ {
		shared = append(shared, Account{
			ID:          int64(2000 + i),
			Name:        fmt.Sprintf("shared-%d", i),
			Platform:    PlatformOpenAI,
			Type:        AccountTypeOAuth,
			OwnerUserID: &otherOwnerID,
			ShareMode:   AccountShareModePublic,
			ShareStatus: AccountShareStatusActive,
			Status:      StatusActive,
			Schedulable: true,
			Groups: []*Group{{
				ID:       99,
				Name:     "PLUS共享号池",
				Platform: PlatformOpenAI,
			}},
		})
	}
	repo := &capacityPoolAccountRepoStub{
		owned:       owned,
		schedulable: shared,
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Mine.TotalAccounts != len(owned) {
		t.Fatalf("expected all owned accounts, got %d want %d", pools.Mine.TotalAccounts, len(owned))
	}
	if pools.Shared.TotalAccounts != len(shared) {
		t.Fatalf("expected all shared accounts, got %d want %d", pools.Shared.TotalAccounts, len(shared))
	}
	sharedGroup := findCapacityPoolGroup(pools.Shared.Groups, 99, "PLUS共享号池")
	if sharedGroup == nil || sharedGroup.TotalAccounts != len(shared) {
		t.Fatalf("unexpected shared group pagination result: %#v", pools.Shared.Groups)
	}
}

func findCapacityPoolSection(sections []UserAccountCapacityPoolSection, platform, accountType string) *UserAccountCapacityPoolSection {
	for i := range sections {
		if sections[i].Platform == platform && sections[i].Type == accountType {
			return &sections[i]
		}
	}
	return nil
}

func findCapacityPoolGroup(groups []UserAccountCapacityPoolGroup, groupID int64, groupName string) *UserAccountCapacityPoolGroup {
	for i := range groups {
		if groupID > 0 {
			if groups[i].GroupID != nil && *groups[i].GroupID == groupID {
				return &groups[i]
			}
			continue
		}
		if groups[i].GroupName == groupName {
			return &groups[i]
		}
	}
	return nil
}

type accountShareSettingsStub struct {
	enabled bool
}

func (s accountShareSettingsStub) IsAccountShareEnabled(context.Context) bool {
	return s.enabled
}

func (s accountShareSettingsStub) IsAccountShareAutoReview(context.Context) bool {
	return false
}

func (s accountShareSettingsStub) GetAccountShareUserAccountLimit(context.Context) int {
	return 0
}

type noopAPIKeyQuotaUpdater struct{}

func (noopAPIKeyQuotaUpdater) UpdateQuotaUsed(context.Context, int64, float64) error {
	return nil
}

func (noopAPIKeyQuotaUpdater) UpdateRateLimitUsage(context.Context, int64, float64) error {
	return nil
}

type capacityPoolAccountRepoStub struct {
	AccountRepository
	all         []Account
	owned       []Account
	schedulable []Account
}

func (s *capacityPoolAccountRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	all := s.all
	if all == nil {
		all = append(append([]Account(nil), s.owned...), s.schedulable...)
	}
	pageItems, pages := paginateCapacityPoolStubItems(all, params)
	return pageItems, &pagination.PaginationResult{
		Total:    int64(len(all)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func (s *capacityPoolAccountRepoStub) ListUserOwned(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	pageItems, pages := paginateCapacityPoolStubItems(s.owned, params)
	return pageItems, &pagination.PaginationResult{
		Total:    int64(len(s.owned)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func paginateCapacityPoolStubItems(items []Account, params pagination.PaginationParams) ([]Account, int) {
	limit := params.Limit()
	if limit <= 0 {
		limit = 20
	}
	pages := 0
	if len(items) > 0 {
		pages = (len(items) + limit - 1) / limit
	}
	start := params.Offset()
	if start >= len(items) {
		return []Account{}, pages
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	return append([]Account(nil), items[start:end]...), pages
}

func (s *capacityPoolAccountRepoStub) CountUserOwned(ctx context.Context, userID int64) (int64, error) {
	return int64(len(s.owned)), nil
}

func (s *capacityPoolAccountRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	return append([]Account(nil), s.schedulable...), nil
}
