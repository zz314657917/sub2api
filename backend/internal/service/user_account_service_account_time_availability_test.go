package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserAccountServiceCreateRejectsInvalidAvailabilityBeforePersistence(t *testing.T) {
	repo := newUserProxyAccountRepoStub()
	svc := NewUserAccountService(repo, nil)

	_, err := svc.Create(context.Background(), 42, CreateAccountRequest{
		Name:     "invalid-window",
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			accountAvailabilityEnabledExtraKey: true,
			accountAvailabilityStartExtraKey:   "1:30",
			accountAvailabilityEndExtraKey:     "18:00",
		},
	})

	require.ErrorContains(t, err, "must use HH:MM")
	require.Empty(t, repo.created)
}

func TestUserAccountServiceUpdateRejectsInvalidAvailabilityBeforePersistence(t *testing.T) {
	ownerID := int64(42)
	repo := newUserProxyAccountRepoStub()
	repo.accounts[7] = &Account{
		ID:          7,
		Name:        "owned-account",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		OwnerUserID: &ownerID,
	}
	svc := NewUserAccountService(repo, nil)
	extra := map[string]any{
		accountAvailabilityEnabledExtraKey: true,
		accountAvailabilityStartExtraKey:   "22:00",
		accountAvailabilityEndExtraKey:     "18:00",
	}

	_, err := svc.Update(context.Background(), ownerID, 7, UpdateAccountRequest{Extra: &extra})

	require.ErrorContains(t, err, "start < end")
	require.Empty(t, repo.updated)
	require.Nil(t, repo.accounts[7].Extra)
}
