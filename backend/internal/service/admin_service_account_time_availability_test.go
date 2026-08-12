package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type accountAvailabilityAdminRepo struct {
	AccountRepository
	account     *Account
	created     *Account
	updated     *Account
	extraUpdate map[string]any
}

func (r *accountAvailabilityAdminRepo) Create(_ context.Context, account *Account) error {
	r.created = account
	return nil
}

func (r *accountAvailabilityAdminRepo) GetByID(_ context.Context, _ int64) (*Account, error) {
	return r.account, nil
}

func (r *accountAvailabilityAdminRepo) Update(_ context.Context, account *Account) error {
	r.updated = account
	r.account = account
	return nil
}

func (r *accountAvailabilityAdminRepo) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.extraUpdate = updates
	return nil
}

func TestAdminServiceCreateAccountRejectsInvalidAccountAvailability(t *testing.T) {
	repo := &accountAvailabilityAdminRepo{}
	svc := &adminServiceImpl{accountRepo: repo}

	_, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "invalid-window",
		Platform:             PlatformOpenAI,
		SkipDefaultGroupBind: true,
		Extra: map[string]any{
			accountAvailabilityEnabledExtraKey: true,
		},
	})

	require.Error(t, err)
	require.Nil(t, repo.created)
}

func TestAdminServiceUpdateAccountExtraValidatesMergedAvailability(t *testing.T) {
	repo := &accountAvailabilityAdminRepo{account: &Account{
		ID: 1,
		Extra: map[string]any{
			accountAvailabilityStartExtraKey: "18:00",
			accountAvailabilityEndExtraKey:   "22:00",
		},
	}}
	svc := &adminServiceImpl{accountRepo: repo}

	require.NoError(t, svc.UpdateAccountExtra(context.Background(), 1, map[string]any{
		accountAvailabilityEnabledExtraKey: true,
	}))
	require.Equal(t, true, repo.extraUpdate[accountAvailabilityEnabledExtraKey])

	repo.account.Extra = nil
	repo.extraUpdate = nil
	err := svc.UpdateAccountExtra(context.Background(), 1, map[string]any{
		accountAvailabilityEnabledExtraKey: true,
	})
	require.Error(t, err)
	require.Nil(t, repo.extraUpdate)
}
