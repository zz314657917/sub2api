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
	if len(pools.Shared.Groups) != 0 {
		t.Fatalf("shared pool must not expose raw internal groups, got %#v", pools.Shared.Groups)
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

func TestUserAccountService_GetCapacityPoolsIncludesOwnApprovedSharedAccounts(t *testing.T) {
	ownerID := int64(10)
	groupID := int64(77)
	now := time.Now()
	rateLimitResetAt := now.Add(30 * time.Minute)
	repo := &capacityPoolAccountRepoStub{
		owned: []Account{
			{
				ID:          1,
				Name:        "my-approved-shared-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &ownerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Groups: []*Group{{
					ID:       groupID,
					Name:     "Owner Shared Pool",
					Platform: PlatformOpenAI,
				}},
				Extra: map[string]any{
					"codex_5h_used_percent": 25,
					"share_display_tier":    "plus",
				},
			},
			{
				ID:               3,
				Name:             "my-approved-limited-openai",
				Platform:         PlatformOpenAI,
				Type:             AccountTypeAPIKey,
				OwnerUserID:      &ownerID,
				ShareMode:        AccountShareModePublic,
				ShareStatus:      AccountShareStatusActive,
				Status:           StatusActive,
				Schedulable:      true,
				RateLimitResetAt: &rateLimitResetAt,
				Groups: []*Group{{
					ID:       groupID,
					Name:     "Owner Shared Pool",
					Platform: PlatformOpenAI,
				}},
				Credentials: map[string]any{
					"plan_type": "plus",
				},
			},
			{
				ID:          2,
				Name:        "my-pending-shared-openai",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &ownerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusPendingReview,
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
	if pools.Mine.TotalAccounts != 3 {
		t.Fatalf("expected all owned accounts in mine pool, got %d", pools.Mine.TotalAccounts)
	}
	if pools.Shared.TotalAccounts != 2 {
		t.Fatalf("expected own approved shared account in shared pool, got %d", pools.Shared.TotalAccounts)
	}
	if pools.Shared.OwnContributedAccounts != 2 {
		t.Fatalf("expected own contributed shared account count 2, got %d", pools.Shared.OwnContributedAccounts)
	}
	sharedGroup := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Plus")
	if sharedGroup == nil {
		t.Fatalf("expected owner shared group, got %#v", pools.Shared.Groups)
	}
	if sharedGroup.TotalAccounts != 2 || sharedGroup.OwnContributedAccounts != 2 || sharedGroup.Windows["5h"].UsedPercent != 25 {
		t.Fatalf("unexpected owner shared group summary: %#v", sharedGroup)
	}
	if sharedGroup.UnavailableReasons["rate_limited"] != 1 {
		t.Fatalf("expected rate-limited reason summary, got %#v", sharedGroup.UnavailableReasons)
	}
}

func TestUserAccountService_GetCapacityPoolsIncludesOnlyPublicOrWrappedSystemSharedAccounts(t *testing.T) {
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
			{
				ID:          3,
				Name:        "hosted-pro-key",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"share_display_tier":         "pro",
					"share_display_percent_only": true,
				},
			},
			{
				ID:          4,
				Name:        "hosted-plus-oauth",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"codex_5h_used_percent":       18,
					"share_display_tier":          "plus",
					"share_display_percent_only":  true,
					"share_display_account_count": 2,
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Shared.TotalAccounts != 4 {
		t.Fatalf("expected public and wrapped system accounts in shared pool, got %d", pools.Shared.TotalAccounts)
	}
	sharedOpenAI := findCapacityPoolSection(pools.Shared.Sections, PlatformOpenAI, AccountTypeOAuth)
	if sharedOpenAI == nil || sharedOpenAI.TotalAccounts != 3 || sharedOpenAI.Windows["5h"].UsedPercent != 42 {
		t.Fatalf("unexpected marked system shared section: %#v", sharedOpenAI)
	}
	wrappedPlus := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Plus")
	if wrappedPlus == nil || wrappedPlus.TotalAccounts != 2 || !wrappedPlus.PercentOnlyQuota {
		t.Fatalf("unexpected wrapped system oauth group: %#v", pools.Shared.Groups)
	}
	wrappedPro := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if wrappedPro == nil || wrappedPro.TotalAccounts != 1 || !wrappedPro.PercentOnlyQuota {
		t.Fatalf("unexpected wrapped system shared group: %#v", pools.Shared.Groups)
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
			Credentials: map[string]any{
				"plan_type": "plus",
			},
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
	sharedGroup := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Plus")
	if sharedGroup == nil || sharedGroup.TotalAccounts != len(shared) {
		t.Fatalf("unexpected shared group pagination result: %#v", pools.Shared.Groups)
	}
}

func TestUserAccountService_GetCapacityPoolsAggregatesQuotaWindowPercent(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	groupID := int64(88)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "shared-pro-a",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Groups: []*Group{{
					ID:       groupID,
					Name:     "Pro APIKey Pool",
					Platform: PlatformOpenAI,
				}},
				Extra: map[string]any{
					"quota_daily_limit":          100.0,
					"quota_daily_used":           10.0,
					"quota_daily_start":          time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					"share_display_percent_only": true,
					"share_display_tier":         "pro",
				},
			},
			{
				ID:          2,
				Name:        "shared-pro-b",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Groups: []*Group{{
					ID:       groupID,
					Name:     "Pro APIKey Pool",
					Platform: PlatformOpenAI,
				}},
				Extra: map[string]any{
					"quota_daily_limit":  200.0,
					"quota_daily_used":   80.0,
					"quota_daily_start":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
					"share_display_tier": "pro",
				},
			},
			{
				ID:          3,
				Name:        "shared-pro-limited",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: false,
				Groups: []*Group{{
					ID:       groupID,
					Name:     "Pro APIKey Pool",
					Platform: PlatformOpenAI,
				}},
				Extra: map[string]any{
					"quota_daily_limit":  100.0,
					"quota_daily_used":   100.0,
					"quota_daily_start":  time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
					"share_display_tier": "pro",
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if group == nil {
		t.Fatalf("expected shared group, got %#v", pools.Shared.Groups)
	}
	window := group.Windows["5h"]
	if window.UsedPercent != 30 {
		t.Fatalf("expected aggregate 5h display percent 30, got %#v", window)
	}
	if window.RemainingUnits != 210 {
		t.Fatalf("expected aggregate remaining quota 210, got %#v", window)
	}
	if _, ok := group.Windows["1d"]; ok {
		t.Fatalf("shared OpenAI plan group must not expose raw quota daily window: %#v", group.Windows)
	}
	if !group.PercentOnlyQuota || !pools.Shared.PercentOnlyQuota {
		t.Fatalf("expected percent-only marker to propagate: pool=%v group=%v", pools.Shared.PercentOnlyQuota, group.PercentOnlyQuota)
	}
}

func TestUserAccountService_GetCapacityPoolsHidesRawSharedGroupsWithoutPlan(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "image2-shared",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Groups: []*Group{{
					ID:       88,
					Name:     "chatgpt-image2图片生成专用",
					Platform: PlatformOpenAI,
				}},
			},
			{
				ID:          2,
				Name:        "j92wqgddr0@kairo.edu.kg-plus",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"plan_type": "plus",
				},
				Groups: []*Group{{
					ID:       89,
					Name:     "GPT-低价号池",
					Platform: PlatformOpenAI,
				}},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Shared.TotalAccounts != 2 {
		t.Fatalf("expected both shared accounts counted in pool summary, got %#v", pools.Shared)
	}
	if len(pools.Shared.Groups) != 1 {
		t.Fatalf("expected only one public plan display group, got %#v", pools.Shared.Groups)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Plus")
	if group == nil || group.TotalAccounts != 1 {
		t.Fatalf("expected only plus account in display group, got %#v", pools.Shared.Groups)
	}
	if findCapacityPoolGroup(pools.Shared.Groups, 88, "chatgpt-image2图片生成专用") != nil ||
		findCapacityPoolGroup(pools.Shared.Groups, 89, "GPT-低价号池") != nil {
		t.Fatalf("shared pool leaked raw internal groups: %#v", pools.Shared.Groups)
	}
}

func TestUserAccountService_GetCapacityPoolsMasksShareDisplayAPIKeyQuotaAsCodexWindows(t *testing.T) {
	ownerID := int64(10)
	now := time.Now()
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "hosted-pro-key",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"quota_daily_limit":          100.0,
					"quota_daily_used":           25.0,
					"quota_daily_start":          now.Add(-1 * time.Hour).Format(time.RFC3339),
					"quota_weekly_limit":         500.0,
					"quota_weekly_used":          300.0,
					"quota_weekly_start":         now.Add(-24 * time.Hour).Format(time.RFC3339),
					"share_display_tier":         "pro",
					"share_display_percent_only": true,
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if group == nil {
		t.Fatalf("expected OpenAI Pro display group, got %#v", pools.Shared.Groups)
	}
	if _, ok := group.Windows["1d"]; ok {
		t.Fatalf("share-display API key must not expose quota daily window: %#v", group.Windows)
	}
	if _, ok := group.Windows["7d_quota"]; ok {
		t.Fatalf("share-display API key must not expose quota weekly window: %#v", group.Windows)
	}
	if window := group.Windows["5h"]; window.UsedPercent != 25 || window.WindowMinutes != 300 {
		t.Fatalf("expected daily quota to be masked as 5h window, got %#v", window)
	}
	if window := group.Windows["7d"]; window.UsedPercent != 60 || window.WindowMinutes != 10080 {
		t.Fatalf("expected weekly quota to be masked as 7d window, got %#v", window)
	}
}

func TestUserAccountService_GetCapacityPoolsUsesShareDisplayAccountCount(t *testing.T) {
	ownerID := int64(10)
	now := time.Now()
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "hosted-pro-key",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"quota_weekly_limit":          100.0,
					"quota_weekly_used":           20.0,
					"quota_weekly_start":          now.Add(-24 * time.Hour).Format(time.RFC3339),
					"share_display_tier":          "pro",
					"share_display_percent_only":  true,
					"share_display_account_count": 8,
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if pools.Shared.TotalAccounts != 8 || pools.Shared.ActiveAccounts != 8 || pools.Shared.SchedulableAccounts != 8 {
		t.Fatalf("expected weighted shared pool counters, got %#v", pools.Shared)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if group == nil {
		t.Fatalf("expected OpenAI Pro display group, got %#v", pools.Shared.Groups)
	}
	if group.TotalAccounts != 8 || group.ActiveAccounts != 8 || group.SchedulableAccounts != 8 {
		t.Fatalf("expected weighted group counters, got %#v", group)
	}
	if window := group.Windows["7d"]; window.SnapshotAccounts != 8 || window.SchedulableSnapshotAccounts != 8 {
		t.Fatalf("expected weighted window snapshot counters, got %#v", window)
	}
}

func TestUserAccountService_GetCapacityPoolsUsesShareDisplayGroupName(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "internal-key-name",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Groups: []*Group{{
					ID:       88,
					Name:     "Internal Pool",
					Platform: PlatformOpenAI,
				}},
				Extra: map[string]any{
					"share_display_tier":         "pro",
					"share_display_percent_only": true,
				},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if group == nil {
		t.Fatalf("expected OpenAI Pro display group, got %#v", pools.Shared.Groups)
	}
	if group.GroupID != nil {
		t.Fatalf("share display group must not expose internal group id, got %v", *group.GroupID)
	}
	if !group.PercentOnlyQuota {
		t.Fatalf("expected percent-only marker on display group")
	}
}

func TestUserAccountService_GetCapacityPoolsGroupsSharedOpenAIAPIKeysByPlan(t *testing.T) {
	ownerID := int64(10)
	otherOwnerID := int64(11)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "plus-by-plan",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Credentials: map[string]any{
					"plan_type": "plus",
				},
				Groups: []*Group{{
					ID:       88,
					Name:     "GPT-低价号池",
					Platform: PlatformOpenAI,
				}},
			},
			{
				ID:          2,
				Name:        "plus-by-display",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				OwnerUserID: &otherOwnerID,
				ShareMode:   AccountShareModePublic,
				ShareStatus: AccountShareStatusActive,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"share_display_tier": "plus",
				},
				Groups: []*Group{{
					ID:       89,
					Name:     "OpenAI Plus",
					Platform: PlatformOpenAI,
				}},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if len(pools.Shared.Groups) != 1 {
		t.Fatalf("expected plan-based grouping, got %#v", pools.Shared.Groups)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Plus")
	if group == nil {
		t.Fatalf("expected OpenAI Plus group, got %#v", pools.Shared.Groups)
	}
	if group.TotalAccounts != 2 || group.SchedulableAccounts != 2 {
		t.Fatalf("unexpected OpenAI Plus totals: %#v", group)
	}
	if group.GroupID != nil {
		t.Fatalf("plan display group must not expose internal group id, got %v", *group.GroupID)
	}
}

func TestUserAccountService_GetCapacityPoolsGroupsSystemOpenAIAPIKeysByPlan(t *testing.T) {
	ownerID := int64(10)
	repo := &capacityPoolAccountRepoStub{
		schedulable: []Account{
			{
				ID:          1,
				Name:        "system-pro-a",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"share_display_tier": "pro",
					"share_display_name": "8000",
				},
				Groups: []*Group{{
					ID:       88,
					Name:     "codex",
					Platform: PlatformOpenAI,
				}},
			},
			{
				ID:          2,
				Name:        "system-pro-b",
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				ShareMode:   AccountShareModePrivate,
				ShareStatus: AccountShareStatusNotShared,
				Status:      StatusActive,
				Schedulable: true,
				Extra: map[string]any{
					"share_display_tier": "pro",
					"share_display_name": "1",
				},
				Groups: []*Group{{
					ID:       89,
					Name:     "8000",
					Platform: PlatformOpenAI,
				}},
			},
		},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})

	pools, err := svc.GetCapacityPools(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("GetCapacityPools returned error: %v", err)
	}
	if len(pools.Shared.Groups) != 1 {
		t.Fatalf("expected system plan grouping, got %#v", pools.Shared.Groups)
	}
	group := findCapacityPoolGroup(pools.Shared.Groups, 0, "OpenAI Pro")
	if group == nil {
		t.Fatalf("expected OpenAI Pro group, got %#v", pools.Shared.Groups)
	}
	if group.TotalAccounts != 2 || group.SchedulableAccounts != 2 {
		t.Fatalf("unexpected OpenAI Pro totals: %#v", group)
	}
	if group.GroupID != nil {
		t.Fatalf("system plan display group must not expose internal group id, got %v", *group.GroupID)
	}
}

func TestUserAccountService_CreateBlocksAPIKeyUpload(t *testing.T) {
	svc := NewUserAccountService(&capacityPoolAccountRepoStub{}, accountShareSettingsStub{enabled: true})

	for _, tc := range []struct {
		name        string
		accountType string
		credentials map[string]any
	}{
		{name: "apikey", accountType: AccountTypeAPIKey, credentials: map[string]any{"api_key": "sk-test"}},
		{name: "upstream", accountType: AccountTypeUpstream, credentials: map[string]any{"api_key": "sk-test"}},
		{name: "oauth api key credential", accountType: AccountTypeOAuth, credentials: map[string]any{"api_key": "sk-test"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), 10, CreateAccountRequest{
				Name:        "user-openai-key",
				Platform:    PlatformOpenAI,
				Type:        tc.accountType,
				Credentials: tc.credentials,
			})
			if !errors.Is(err, ErrUserAccountAPIKeyBlocked) {
				t.Fatalf("expected ErrUserAccountAPIKeyBlocked, got %v", err)
			}
		})
	}
}

func TestUserAccountService_UpdateBlocksAPIKeyCredentials(t *testing.T) {
	ownerID := int64(10)
	repo := &capacityPoolAccountRepoStub{
		owned: []Account{{
			ID:          1,
			Name:        "user-openai-key",
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			OwnerUserID: &ownerID,
			Status:      StatusActive,
			Schedulable: true,
		}},
	}
	svc := NewUserAccountService(repo, accountShareSettingsStub{enabled: true})
	credentials := map[string]any{"api_key": "sk-new"}

	_, err := svc.Update(context.Background(), ownerID, 1, UpdateAccountRequest{
		Credentials: &credentials,
	})
	if !errors.Is(err, ErrUserAccountAPIKeyBlocked) {
		t.Fatalf("expected ErrUserAccountAPIKeyBlocked, got %v", err)
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

func (s *capacityPoolAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	all := s.all
	if all == nil {
		all = append(append([]Account(nil), s.owned...), s.schedulable...)
	}
	for i := range all {
		if all[i].ID == id {
			account := all[i]
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account %d not found", id)
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
