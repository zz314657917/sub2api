package service

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type sparkShadowAdminRepo struct {
	AccountRepository
	mu              sync.Mutex
	accounts        map[int64]*Account
	shadows         map[int64][]*Account
	bulkUpdateIDs   []int64
	shadowLookupIDs []int64
	atomicShadowIDs []int64
	updatedIDs      []int64
	createAttempts  int
	createFailure   error
	refreshCalls    int
	deletedIDs      []int64
}

func (r *sparkShadowAdminRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.accounts[id], nil
}

func (r *sparkShadowAdminRepo) GetByIDs(_ context.Context, ids []int64) ([]*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*Account, 0, len(ids))
	for _, id := range ids {
		if account := r.accounts[id]; account != nil {
			out = append(out, account)
		}
	}
	return out, nil
}

func (r *sparkShadowAdminRepo) ListShadowsByParent(_ context.Context, parentID int64) ([]*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shadowLookupIDs = append(r.shadowLookupIDs, parentID)
	return r.shadows[parentID], nil
}

func (r *sparkShadowAdminRepo) Create(_ context.Context, account *Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.createAttempts++
	if r.createFailure != nil {
		return r.createFailure
	}
	if r.createAttempts > 1 {
		return errors.New("duplicate key value violates unique constraint account_spark_shadow_parent_uq")
	}
	account.ID = 900
	return nil
}

func (r *sparkShadowAdminRepo) BulkUpdate(_ context.Context, ids []int64, _ AccountBulkUpdate) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bulkUpdateIDs = append([]int64(nil), ids...)
	return int64(len(ids)), nil
}

func (r *sparkShadowAdminRepo) Update(_ context.Context, account *Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatedIDs = append(r.updatedIDs, account.ID)
	return nil
}

func (r *sparkShadowAdminRepo) UpdateWithAccountBillingSettingsAndShadowProxy(
	_ context.Context,
	account *Account,
	_ *bool,
	_ *bool,
	_ *float64,
	shadowIDs []int64,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatedIDs = append(r.updatedIDs, account.ID)
	r.atomicShadowIDs = append([]int64(nil), shadowIDs...)
	return nil
}

func (r *sparkShadowAdminRepo) RefreshQuotaWindows(_ context.Context, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshCalls++
	return nil
}

func (r *sparkShadowAdminRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deletedIDs = append(r.deletedIDs, id)
	return nil
}

func TestSparkShadowBulkProxySyncAndChildCredentialGuard(t *testing.T) {
	parentID := int64(1)
	childID := int64(2)
	proxyID := int64(7)
	child := &Account{ID: childID, ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark}
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{
			parentID: {ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive},
			childID:  child,
		},
		shadows: map[int64][]*Account{parentID: {child}},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	_, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{AccountIDs: []int64{parentID}, ProxyID: &proxyID})
	require.NoError(t, err)
	require.Equal(t, []int64{parentID, childID}, repo.bulkUpdateIDs)
	require.Equal(t, []int64{parentID}, repo.shadowLookupIDs)

	_, err = svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{AccountIDs: []int64{childID}, Credentials: map[string]any{"access_token": "forbidden"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "SPARK_SHADOW_NO_CREDENTIALS")
}

func TestSparkShadowConcurrentCreateReturnsStructuredConflict(t *testing.T) {
	parentID := int64(8)
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{
			parentID: {ID: parentID, Name: "parent", Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive},
		},
	}
	svc := &adminServiceImpl{accountRepo: repo}
	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := svc.CreateShadow(context.Background(), parentID, "", 0, 0, nil)
			errs <- err
		}()
	}
	close(start)

	successes := 0
	conflicts := 0
	for range 2 {
		err := <-errs
		if err == nil {
			successes++
			continue
		}
		require.Equal(t, http.StatusConflict, infraerrors.Code(err))
		require.Equal(t, "SPARK_SHADOW_ALREADY_EXISTS", infraerrors.Reason(err))
		conflicts++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	require.Equal(t, 2, repo.createAttempts)
}

func TestSparkShadowCreatePreservesNonUniqueFailure(t *testing.T) {
	parentID := int64(9)
	createFailure := errors.New("insert violates foreign key constraint")
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{
			parentID: {ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive},
		},
		createFailure: createFailure,
	}

	_, err := (&adminServiceImpl{accountRepo: repo}).CreateShadow(context.Background(), parentID, "", 0, 0, nil)
	require.ErrorIs(t, err, createFailure)
	require.NotEqual(t, "SPARK_SHADOW_ALREADY_EXISTS", infraerrors.Reason(err))
}

func TestSparkShadowBulkProxySkipsNonOAuthParents(t *testing.T) {
	accountID := int64(21)
	proxyID := int64(7)
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{
			accountID: {ID: accountID, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive},
		},
	}

	_, err := (&adminServiceImpl{accountRepo: repo}).BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs: []int64{accountID},
		ProxyID:    &proxyID,
	})

	require.NoError(t, err)
	require.Equal(t, []int64{accountID}, repo.bulkUpdateIDs)
	require.Empty(t, repo.shadowLookupIDs)
}

func TestSparkShadowSingleProxySyncIsAtomicAndSkipsNonOAuthAccounts(t *testing.T) {
	parentID := int64(31)
	childID := int64(32)
	proxyID := int64(7)
	child := &Account{ID: childID, ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark}
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{
			parentID: {ID: parentID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive},
			childID:  child,
		},
		shadows: map[int64][]*Account{parentID: {child}},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	updated, err := svc.UpdateAccount(context.Background(), parentID, &UpdateAccountInput{ProxyID: &proxyID})
	require.NoError(t, err)
	require.Equal(t, &proxyID, updated.ProxyID)
	require.Equal(t, []int64{parentID}, repo.shadowLookupIDs)
	require.Equal(t, []int64{parentID}, repo.updatedIDs)
	require.Equal(t, []int64{childID}, repo.atomicShadowIDs)

	apiKeyID := int64(33)
	repo.accounts[apiKeyID] = &Account{ID: apiKeyID, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive}
	_, err = svc.UpdateAccount(context.Background(), apiKeyID, &UpdateAccountInput{ProxyID: &proxyID})
	require.NoError(t, err)
	require.Equal(t, []int64{parentID}, repo.shadowLookupIDs)
	require.Equal(t, []int64{parentID, apiKeyID}, repo.updatedIDs)
}

func TestSparkShadowRefreshQuotaAndParentDeleteGuards(t *testing.T) {
	parentID := int64(11)
	childID := int64(12)
	child := &Account{ID: childID, ParentAccountID: &parentID, QuotaDimension: QuotaDimensionSpark}
	repo := &sparkShadowAdminRepo{
		accounts: map[int64]*Account{parentID: {ID: parentID}, childID: child},
		shadows:  map[int64][]*Account{parentID: {child}},
	}
	svc := &adminServiceImpl{accountRepo: repo}

	_, err := svc.RefreshAccountQuota(context.Background(), childID)
	require.Error(t, err)
	require.Zero(t, repo.refreshCalls)
	require.NoError(t, svc.DeleteAccount(context.Background(), parentID))
	require.Equal(t, []int64{childID, parentID}, repo.deletedIDs)
}
