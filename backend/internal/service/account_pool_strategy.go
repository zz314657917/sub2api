package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

func pinnedAccountIDFromContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	id, ok := ctx.Value(ctxkey.APIKeyPinnedAccountID).(int64)
	return id, ok && id > 0
}

func accountPoolStrategyFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return AccountPoolStrategySharedOnly, false
	}
	value, ok := ctx.Value(ctxkey.APIKeyAccountPoolStrategy).(string)
	if !ok {
		return AccountPoolStrategySharedOnly, false
	}
	return NormalizeAccountPoolStrategy(value), true
}

func withAccountPoolStrategy(ctx context.Context, strategy string) context.Context {
	return context.WithValue(ctx, ctxkey.APIKeyAccountPoolStrategy, NormalizeAccountPoolStrategy(strategy))
}

func withAccountPoolUserID(ctx context.Context, userID int64) context.Context {
	if userID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.APIKeyUserID, userID)
}

func accountPoolUserIDFromContext(ctx context.Context, fallbackUserID int64) int64 {
	if fallbackUserID > 0 {
		return fallbackUserID
	}
	if ctx == nil {
		return 0
	}
	if userID, ok := ctx.Value(ctxkey.APIKeyUserID).(int64); ok {
		return userID
	}
	return 0
}

func accountPoolStrategyIsPrivateFirst(ctx context.Context) bool {
	strategy, configured := accountPoolStrategyFromContext(ctx)
	return configured && strategy == AccountPoolStrategyPrivateFirst
}

func accountPoolStrategyUsesPrivateOnly(ctx context.Context) bool {
	strategy, configured := accountPoolStrategyFromContext(ctx)
	return configured && strategy == AccountPoolStrategyPrivateOnly
}

func filterAccountsForAPIKeyPoolStrategy(ctx context.Context, accounts []Account, fallbackUserID int64) []Account {
	if len(accounts) == 0 {
		return accounts
	}
	strategy, configured := accountPoolStrategyFromContext(ctx)
	userID := accountPoolUserIDFromContext(ctx, fallbackUserID)
	if !configured {
		return filterAccountsForUser(accounts, userID)
	}

	filtered := accounts[:0]
	for _, account := range accounts {
		switch strategy {
		case AccountPoolStrategyPrivateOnly:
			if accountIsOwnedPrivatePoolAccount(&account, userID) {
				filtered = append(filtered, account)
			}
		default:
			if accountIsSharedPoolAccount(&account) {
				filtered = append(filtered, account)
			}
		}
	}
	return filtered
}

func accountAllowedByAPIKeyPoolStrategy(ctx context.Context, account *Account, fallbackUserID int64) bool {
	if account == nil {
		return false
	}
	strategy, configured := accountPoolStrategyFromContext(ctx)
	userID := accountPoolUserIDFromContext(ctx, fallbackUserID)
	if !configured {
		return account.CanBeUsedByUser(userID)
	}
	switch strategy {
	case AccountPoolStrategyPrivateOnly:
		return accountIsOwnedPrivatePoolAccount(account, userID)
	default:
		return accountIsSharedPoolAccount(account)
	}
}

func accountIsOwnedPrivatePoolAccount(account *Account, userID int64) bool {
	if account == nil || account.OwnerUserID == nil || userID <= 0 {
		return false
	}
	return *account.OwnerUserID == userID && account.ShareMode != AccountShareModePublic
}

func accountIsSharedPoolAccount(account *Account) bool {
	if account == nil {
		return false
	}
	if account.OwnerUserID == nil {
		return true
	}
	return account.ShareMode == AccountShareModePublic && account.ShareStatus == AccountShareStatusActive
}

func listPrivatePoolSchedulableAccounts(ctx context.Context, repo AccountRepository, userID int64, platforms []string) ([]Account, error) {
	if repo == nil || userID <= 0 || len(platforms) == 0 {
		return []Account{}, nil
	}
	var (
		accounts []Account
		err      error
	)
	if len(platforms) == 1 {
		accounts, err = repo.ListSchedulableByPlatform(ctx, platforms[0])
	} else {
		accounts, err = repo.ListSchedulableByPlatforms(ctx, platforms)
	}
	if err != nil {
		return nil, err
	}
	filtered := accounts[:0]
	for _, account := range accounts {
		if accountIsOwnedPrivatePoolAccount(&account, userID) {
			filtered = append(filtered, account)
		}
	}
	return filtered, nil
}

func isAccountPoolNoAvailableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNoAvailableAccounts) || errors.Is(err, ErrNoAvailableCompactAccounts) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no available accounts") ||
		strings.Contains(msg, "no available account") ||
		strings.Contains(msg, "no available openai accounts") ||
		strings.Contains(msg, "no available gemini accounts")
}
