package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type v027PausedRefreshRepo struct {
	schedulerTestOpenAIAccountRepo
	activeAccounts []Account
	updated        map[int64]map[string]any
	listErr        error
	setSchedulable int
}

func (r *v027PausedRefreshRepo) ListActive(context.Context) ([]Account, error) {
	return r.activeAccounts, r.listErr
}

func (r *v027PausedRefreshRepo) UpdateCredentials(_ context.Context, id int64, credentials map[string]any) error {
	if r.updated == nil {
		r.updated = make(map[int64]map[string]any)
	}
	r.updated[id] = cloneCredentials(credentials)
	return nil
}

func (r *v027PausedRefreshRepo) SetSchedulable(context.Context, int64, bool) error {
	r.setSchedulable++
	return nil
}

type v027PausedRefreshRefresher struct{}

func (v027PausedRefreshRefresher) CanRefresh(account *Account) bool {
	return account.Type == AccountTypeOAuth
}
func (v027PausedRefreshRefresher) NeedsRefresh(*Account, time.Duration) bool { return true }
func (v027PausedRefreshRefresher) Refresh(context.Context, *Account) (map[string]any, error) {
	return map[string]any{"access_token": "renewed"}, nil
}

func TestV027PausedRefresh_KeepsAdministratorPauseDuringRefresh(t *testing.T) {
	paused := Account{ID: 27, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Schedulable: false}
	enabled := Account{ID: 28, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Schedulable: true}
	nonOAuth := Account{ID: 29, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Schedulable: true}
	repo := &v027PausedRefreshRepo{activeAccounts: []Account{paused, enabled, nonOAuth}}
	service := &TokenRefreshService{
		accountRepo: repo,
		refreshers:  []TokenRefresher{v027PausedRefreshRefresher{}},
		cfg:         &config.TokenRefreshConfig{RefreshBeforeExpiryHours: 1, MaxRetries: 1},
	}

	service.processRefresh()

	require.Equal(t, "renewed", repo.updated[paused.ID]["access_token"])
	require.Equal(t, "renewed", repo.updated[enabled.ID]["access_token"])
	require.NotContains(t, repo.updated, nonOAuth.ID, "non-OAuth accounts are not refresh candidates")
	require.False(t, repo.activeAccounts[0].Schedulable, "token refresh must not reverse the administrator pause")
	require.Zero(t, repo.setSchedulable, "refresh must not call SetSchedulable")
}

func TestV027PausedRefresh_RepositoryErrorStopsCycle(t *testing.T) {
	repo := &v027PausedRefreshRepo{listErr: context.DeadlineExceeded}
	service := &TokenRefreshService{accountRepo: repo, refreshers: []TokenRefresher{v027PausedRefreshRefresher{}}, cfg: &config.TokenRefreshConfig{MaxRetries: 1}}

	service.processRefresh()

	require.Empty(t, repo.updated)
	require.Zero(t, repo.setSchedulable)
}
