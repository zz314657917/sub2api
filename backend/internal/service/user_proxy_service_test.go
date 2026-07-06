package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

func TestUserAccountService_CreateBindsOnlyUserOwnedProxy(t *testing.T) {
	ownerID := int64(10)
	proxyID := int64(20)
	accountRepo := newUserProxyAccountRepoStub()
	proxyRepo := &userProxyRepoStub{
		proxies: map[int64]*Proxy{
			proxyID: {ID: proxyID, OwnerUserID: &ownerID, Status: StatusActive},
		},
	}
	svc := NewUserAccountService(accountRepo, accountShareSettingsStub{enabled: true})
	svc.SetProxyRepository(proxyRepo)

	account, err := svc.Create(context.Background(), ownerID, CreateAccountRequest{
		Name:        "owned",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		ProxyID:     &proxyID,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if account.ProxyID == nil || *account.ProxyID != proxyID {
		t.Fatalf("expected proxy_id %d, got %#v", proxyID, account.ProxyID)
	}
	if len(accountRepo.created) != 1 || accountRepo.created[0].ProxyID == nil || *accountRepo.created[0].ProxyID != proxyID {
		t.Fatalf("repository create did not receive proxy_id: %#v", accountRepo.created)
	}

	otherProxyID := int64(21)
	_, err = svc.Create(context.Background(), ownerID, CreateAccountRequest{
		Name:        "other",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		ProxyID:     &otherProxyID,
	})
	if !errors.Is(err, ErrProxyNotFound) {
		t.Fatalf("expected ErrProxyNotFound for unowned proxy, got %v", err)
	}
}

func TestUserAccountService_UpdateCanClearUserProxy(t *testing.T) {
	ownerID := int64(10)
	proxyID := int64(20)
	accountRepo := newUserProxyAccountRepoStub()
	accountRepo.accounts[1] = &Account{
		ID:          1,
		Name:        "owned",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		OwnerUserID: &ownerID,
		ProxyID:     &proxyID,
		Status:      StatusActive,
	}
	svc := NewUserAccountService(accountRepo, accountShareSettingsStub{enabled: true})
	svc.SetProxyRepository(&userProxyRepoStub{})

	clear := int64(0)
	account, err := svc.Update(context.Background(), ownerID, 1, UpdateAccountRequest{ProxyID: &clear})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if account.ProxyID != nil {
		t.Fatalf("expected proxy_id cleared, got %#v", account.ProxyID)
	}
	if len(accountRepo.updated) != 1 || accountRepo.updated[0].ProxyID != nil {
		t.Fatalf("repository update did not clear proxy_id: %#v", accountRepo.updated)
	}
}

func TestUserAccountService_CreateAndUpdateUserConcurrency(t *testing.T) {
	ownerID := int64(10)
	accountRepo := newUserProxyAccountRepoStub()
	svc := NewUserAccountService(accountRepo, accountShareSettingsStub{enabled: true})

	account, err := svc.Create(context.Background(), ownerID, CreateAccountRequest{
		Name:        "owned",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		Concurrency: 5,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if account.Concurrency != 5 {
		t.Fatalf("expected concurrency 5, got %d", account.Concurrency)
	}

	nextConcurrency := 2
	updated, err := svc.Update(context.Background(), ownerID, account.ID, UpdateAccountRequest{Concurrency: &nextConcurrency})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Concurrency != 2 {
		t.Fatalf("expected concurrency 2, got %d", updated.Concurrency)
	}
	if len(accountRepo.updated) != 1 || accountRepo.updated[0].Concurrency != 2 {
		t.Fatalf("repository update did not receive concurrency: %#v", accountRepo.updated)
	}
}

func TestUserAccountService_DefaultsInvalidUserConcurrency(t *testing.T) {
	ownerID := int64(10)
	accountRepo := newUserProxyAccountRepoStub()
	svc := NewUserAccountService(accountRepo, accountShareSettingsStub{enabled: true})

	account, err := svc.Create(context.Background(), ownerID, CreateAccountRequest{
		Name:        "owned",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		Concurrency: -1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if account.Concurrency != userAccountDefaultConcurrency {
		t.Fatalf("expected default concurrency %d, got %d", userAccountDefaultConcurrency, account.Concurrency)
	}
}

func TestUserAccountService_UpdateCanClearExpiresAt(t *testing.T) {
	ownerID := int64(10)
	expiresAt := time.Now().Add(time.Hour)
	accountRepo := newUserProxyAccountRepoStub()
	accountRepo.accounts[1] = &Account{
		ID:          1,
		Name:        "owned",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"refresh_token": "rt"},
		OwnerUserID: &ownerID,
		ExpiresAt:   &expiresAt,
		Status:      StatusActive,
	}
	svc := NewUserAccountService(accountRepo, accountShareSettingsStub{enabled: true})

	account, err := svc.Update(context.Background(), ownerID, 1, UpdateAccountRequest{ClearExpiresAt: true})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if account.ExpiresAt != nil {
		t.Fatalf("expected expires_at cleared, got %#v", account.ExpiresAt)
	}
	if len(accountRepo.updated) != 1 || accountRepo.updated[0].ExpiresAt != nil {
		t.Fatalf("repository update did not clear expires_at: %#v", accountRepo.updated)
	}
}

func TestUserProxyService_DeleteRejectsInUseProxy(t *testing.T) {
	ownerID := int64(10)
	proxyID := int64(20)
	proxyRepo := &userProxyRepoStub{
		proxies: map[int64]*Proxy{
			proxyID: {ID: proxyID, OwnerUserID: &ownerID, Status: StatusActive},
		},
		userAccountCounts: map[int64]int64{proxyID: 1},
	}
	svc := NewUserProxyService(proxyRepo)

	err := svc.Delete(context.Background(), ownerID, proxyID)
	if !errors.Is(err, ErrProxyInUse) {
		t.Fatalf("expected ErrProxyInUse, got %v", err)
	}
	if len(proxyRepo.deletedIDs) != 0 {
		t.Fatalf("proxy should not be deleted while in use: %v", proxyRepo.deletedIDs)
	}
}

type userProxyAccountRepoStub struct {
	accounts map[int64]*Account
	created  []*Account
	updated  []*Account
}

func newUserProxyAccountRepoStub() *userProxyAccountRepoStub {
	return &userProxyAccountRepoStub{accounts: map[int64]*Account{}}
}

func (s *userProxyAccountRepoStub) Create(ctx context.Context, account *Account) error {
	cp := *account
	if cp.ID == 0 {
		cp.ID = int64(len(s.accounts) + 1)
	}
	s.accounts[cp.ID] = &cp
	s.created = append(s.created, &cp)
	account.ID = cp.ID
	return nil
}

func (s *userProxyAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	account := s.accounts[id]
	if account == nil {
		return nil, ErrAccountNotFound
	}
	cp := *account
	return &cp, nil
}

func (s *userProxyAccountRepoStub) Update(ctx context.Context, account *Account) error {
	cp := *account
	s.accounts[cp.ID] = &cp
	s.updated = append(s.updated, &cp)
	return nil
}

func (s *userProxyAccountRepoStub) ListUserOwned(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{Total: 0, Page: params.Page, PageSize: params.PageSize, Pages: 0}, nil
}

func (s *userProxyAccountRepoStub) CountUserOwned(ctx context.Context, userID int64) (int64, error) {
	return int64(len(s.accounts)), nil
}

func (s *userProxyAccountRepoStub) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	panic("unexpected GetByIDs call")
}
func (s *userProxyAccountRepoStub) ExistsByID(ctx context.Context, id int64) (bool, error) {
	panic("unexpected ExistsByID call")
}
func (s *userProxyAccountRepoStub) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	panic("unexpected GetByCRSAccountID call")
}
func (s *userProxyAccountRepoStub) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	panic("unexpected FindByExtraField call")
}
func (s *userProxyAccountRepoStub) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	panic("unexpected ListCRSAccountIDs call")
}
func (s *userProxyAccountRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}
func (s *userProxyAccountRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *userProxyAccountRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *userProxyAccountRepoStub) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListByGroup call")
}
func (s *userProxyAccountRepoStub) ListAllByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListAllByGroup call")
}
func (s *userProxyAccountRepoStub) UpdateGroupAccountPriorities(ctx context.Context, groupID int64, updates []GroupAccountPriorityUpdate) error {
	panic("unexpected UpdateGroupAccountPriorities call")
}
func (s *userProxyAccountRepoStub) ListActive(ctx context.Context) ([]Account, error) {
	panic("unexpected ListActive call")
}
func (s *userProxyAccountRepoStub) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListByPlatform call")
}
func (s *userProxyAccountRepoStub) UpdateLastUsed(ctx context.Context, id int64) error {
	panic("unexpected UpdateLastUsed call")
}
func (s *userProxyAccountRepoStub) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	panic("unexpected BatchUpdateLastUsed call")
}
func (s *userProxyAccountRepoStub) SetError(ctx context.Context, id int64, errorMsg string) error {
	panic("unexpected SetError call")
}
func (s *userProxyAccountRepoStub) ClearError(ctx context.Context, id int64) error {
	panic("unexpected ClearError call")
}
func (s *userProxyAccountRepoStub) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	panic("unexpected SetSchedulable call")
}
func (s *userProxyAccountRepoStub) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	panic("unexpected AutoPauseExpiredAccounts call")
}
func (s *userProxyAccountRepoStub) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	panic("unexpected BindGroups call")
}
func (s *userProxyAccountRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	panic("unexpected ListSchedulable call")
}
func (s *userProxyAccountRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupID call")
}
func (s *userProxyAccountRepoStub) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatform call")
}
func (s *userProxyAccountRepoStub) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatform call")
}
func (s *userProxyAccountRepoStub) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatforms call")
}
func (s *userProxyAccountRepoStub) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatforms call")
}
func (s *userProxyAccountRepoStub) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatform call")
}
func (s *userProxyAccountRepoStub) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableUngroupedByPlatforms call")
}
func (s *userProxyAccountRepoStub) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	panic("unexpected SetRateLimited call")
}
func (s *userProxyAccountRepoStub) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time) error {
	panic("unexpected SetModelRateLimit call")
}
func (s *userProxyAccountRepoStub) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	panic("unexpected SetOverloaded call")
}
func (s *userProxyAccountRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	panic("unexpected SetTempUnschedulable call")
}
func (s *userProxyAccountRepoStub) ClearTempUnschedulable(ctx context.Context, id int64) error {
	panic("unexpected ClearTempUnschedulable call")
}
func (s *userProxyAccountRepoStub) ClearRateLimit(ctx context.Context, id int64) error {
	panic("unexpected ClearRateLimit call")
}
func (s *userProxyAccountRepoStub) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	panic("unexpected ClearAntigravityQuotaScopes call")
}
func (s *userProxyAccountRepoStub) ClearModelRateLimits(ctx context.Context, id int64) error {
	panic("unexpected ClearModelRateLimits call")
}
func (s *userProxyAccountRepoStub) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	panic("unexpected UpdateSessionWindow call")
}
func (s *userProxyAccountRepoStub) UpdateSessionWindowEnd(ctx context.Context, id int64, end time.Time) error {
	panic("unexpected UpdateSessionWindowEnd call")
}
func (s *userProxyAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	panic("unexpected UpdateExtra call")
}
func (s *userProxyAccountRepoStub) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	panic("unexpected BulkUpdate call")
}
func (s *userProxyAccountRepoStub) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
	panic("unexpected IncrementQuotaUsed call")
}
func (s *userProxyAccountRepoStub) RefreshQuotaWindows(ctx context.Context, id int64) error {
	panic("unexpected RefreshQuotaWindows call")
}
func (s *userProxyAccountRepoStub) ResetQuotaUsed(ctx context.Context, id int64) error {
	panic("unexpected ResetQuotaUsed call")
}

type userProxyRepoStub struct {
	proxies           map[int64]*Proxy
	userAccountCounts map[int64]int64
	deletedIDs        []int64
}

func (s *userProxyRepoStub) Create(ctx context.Context, proxy *Proxy) error {
	if s.proxies == nil {
		s.proxies = map[int64]*Proxy{}
	}
	if proxy.ID == 0 {
		proxy.ID = int64(len(s.proxies) + 1)
	}
	cp := *proxy
	s.proxies[cp.ID] = &cp
	return nil
}

func (s *userProxyRepoStub) GetUserOwnedByID(ctx context.Context, userID, proxyID int64) (*Proxy, error) {
	proxy := s.proxies[proxyID]
	if proxy == nil || proxy.OwnerUserID == nil || *proxy.OwnerUserID != userID {
		return nil, ErrProxyNotFound
	}
	cp := *proxy
	return &cp, nil
}

func (s *userProxyRepoStub) CountUserOwnedAccountsByProxyID(ctx context.Context, userID, proxyID int64) (int64, error) {
	return s.userAccountCounts[proxyID], nil
}

func (s *userProxyRepoStub) Delete(ctx context.Context, id int64) error {
	s.deletedIDs = append(s.deletedIDs, id)
	delete(s.proxies, id)
	return nil
}

func (s *userProxyRepoStub) ListUserOwned(ctx context.Context, userID int64) ([]Proxy, error) {
	out := []Proxy{}
	for _, proxy := range s.proxies {
		if proxy.OwnerUserID != nil && *proxy.OwnerUserID == userID {
			out = append(out, *proxy)
		}
	}
	return out, nil
}

func (s *userProxyRepoStub) Update(ctx context.Context, proxy *Proxy) error {
	if s.proxies == nil {
		s.proxies = map[int64]*Proxy{}
	}
	cp := *proxy
	s.proxies[cp.ID] = &cp
	return nil
}

func (s *userProxyRepoStub) GetByID(ctx context.Context, id int64) (*Proxy, error) {
	panic("unexpected GetByID call")
}
func (s *userProxyRepoStub) ListByIDs(ctx context.Context, ids []int64) ([]Proxy, error) {
	panic("unexpected ListByIDs call")
}
func (s *userProxyRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (s *userProxyRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (s *userProxyRepoStub) ListWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFiltersAndAccountCount call")
}
func (s *userProxyRepoStub) ListGlobalWithFilters(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]Proxy, *pagination.PaginationResult, error) {
	panic("unexpected ListGlobalWithFilters call")
}
func (s *userProxyRepoStub) ListGlobalWithFiltersAndAccountCount(ctx context.Context, params pagination.PaginationParams, protocol, status, search string) ([]ProxyWithAccountCount, *pagination.PaginationResult, error) {
	panic("unexpected ListGlobalWithFiltersAndAccountCount call")
}
func (s *userProxyRepoStub) ListActive(ctx context.Context) ([]Proxy, error) {
	panic("unexpected ListActive call")
}
func (s *userProxyRepoStub) ListActiveWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	panic("unexpected ListActiveWithAccountCount call")
}
func (s *userProxyRepoStub) ListActiveGlobal(ctx context.Context) ([]Proxy, error) {
	panic("unexpected ListActiveGlobal call")
}
func (s *userProxyRepoStub) ListActiveGlobalWithAccountCount(ctx context.Context) ([]ProxyWithAccountCount, error) {
	panic("unexpected ListActiveGlobalWithAccountCount call")
}
func (s *userProxyRepoStub) ListActiveUserOwned(ctx context.Context, userID int64) ([]Proxy, error) {
	panic("unexpected ListActiveUserOwned call")
}
func (s *userProxyRepoStub) ExistsByHostPortAuth(ctx context.Context, host string, port int, username, password string) (bool, error) {
	panic("unexpected ExistsByHostPortAuth call")
}
func (s *userProxyRepoStub) CountAccountsByProxyID(ctx context.Context, proxyID int64) (int64, error) {
	panic("unexpected CountAccountsByProxyID call")
}
func (s *userProxyRepoStub) ListAccountSummariesByProxyID(ctx context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	panic("unexpected ListAccountSummariesByProxyID call")
}
