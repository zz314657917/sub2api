package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAILongContextBillingExtraNormalizesAndFailsClosed(t *testing.T) {
	created, err := normalizeOpenAILongContextBillingExtra(PlatformOpenAI, nil)
	require.NoError(t, err)
	require.Equal(t, false, created[openAILongContextBillingEnabledKey])

	_, err = normalizeOpenAILongContextBillingExtra(PlatformOpenAI, map[string]any{
		openAILongContextBillingEnabledKey: "true",
	})
	require.Error(t, err)

	account := &Account{Platform: PlatformOpenAI, Extra: map[string]any{
		openAILongContextBillingEnabledKey: true,
	}}
	updated, err := normalizeOpenAILongContextBillingUpdateExtra(account, map[string]any{"other": "value"})
	require.NoError(t, err)
	require.Equal(t, true, updated[openAILongContextBillingEnabledKey])

	require.True(t, account.IsOpenAILongContextBillingEnabled())
	require.False(t, (&Account{Platform: PlatformOpenAI}).IsOpenAILongContextBillingEnabled())
	require.False(t, (&Account{Platform: PlatformOpenAI, Extra: map[string]any{
		openAILongContextBillingEnabledKey: "true",
	}}).IsOpenAILongContextBillingEnabled())
}

func TestAdminServiceCreateOpenAILongContextBillingDefaultsOffAndRejectsMalformedValue(t *testing.T) {
	repo := &upstreamBillingProbeAccountRepo{}
	svc := &adminServiceImpl{accountRepo: repo}

	created, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "openai-default",
		Platform:             PlatformOpenAI,
		Type:                 AccountTypeAPIKey,
		Credentials:          map[string]any{"api_key": "sk-test"},
		SkipDefaultGroupBind: true,
	})
	require.NoError(t, err)
	require.Equal(t, false, created.Extra[openAILongContextBillingEnabledKey])

	_, err = svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "openai-malformed",
		Platform:             PlatformOpenAI,
		Type:                 AccountTypeAPIKey,
		Credentials:          map[string]any{"api_key": "sk-test"},
		Extra:                map[string]any{openAILongContextBillingEnabledKey: "true"},
		SkipDefaultGroupBind: true,
	})
	require.Error(t, err)
	require.Len(t, repo.accounts, 1)
}

func TestAdminServiceUpdateOpenAILongContextBillingPreservesExistingValueAndRejectsMalformedValue(t *testing.T) {
	const accountID = int64(2201)
	repo := &upstreamBillingProbeAccountRepo{accounts: map[int64]*Account{
		accountID: {
			ID:       accountID,
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Status:   StatusActive,
			Extra:    map[string]any{openAILongContextBillingEnabledKey: true},
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}

	updated, err := svc.UpdateAccount(context.Background(), accountID, &UpdateAccountInput{
		Extra: map[string]any{"custom": "value"},
	})
	require.NoError(t, err)
	require.Equal(t, true, updated.Extra[openAILongContextBillingEnabledKey])
	require.Equal(t, "value", updated.Extra["custom"])

	_, err = svc.UpdateAccount(context.Background(), accountID, &UpdateAccountInput{
		Extra: map[string]any{openAILongContextBillingEnabledKey: "true"},
	})
	require.Error(t, err)
	require.Equal(t, true, repo.accounts[accountID].Extra[openAILongContextBillingEnabledKey])
}

func TestAdminServiceOpenAILongContextBillingExtraAndBulkUpdateRejectMalformedValueBeforeWrite(t *testing.T) {
	const accountID = int64(2202)
	repo := &upstreamBillingProbeAccountRepo{accounts: map[int64]*Account{
		accountID: {ID: accountID, Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
	}}
	svc := &adminServiceImpl{accountRepo: repo}

	err := svc.UpdateAccountExtra(context.Background(), accountID, map[string]any{
		openAILongContextBillingEnabledKey: "false",
	})
	require.Error(t, err)
	require.Empty(t, repo.accounts[accountID].Extra)

	_, err = svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{accountID},
		Extra:      map[string]any{openAILongContextBillingEnabledKey: "true"},
	})
	require.Error(t, err)
	require.Empty(t, repo.bulkUpdates)
}
