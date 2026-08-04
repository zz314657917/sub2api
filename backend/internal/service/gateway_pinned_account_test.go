package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type pinnedGatewayAccountRepo struct {
	AccountRepository
	accounts map[int64]*Account
}

func (r pinnedGatewayAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	account := r.accounts[id]
	if account == nil {
		return nil, errors.New("account not found")
	}
	return account, nil
}

func TestGatewayService_PinnedAccountRejectsFallbackAndUsesOnlyBoundAccount(t *testing.T) {
	groupID := int64(31)
	pinned := &Account{
		ID:          101,
		Platform:    PlatformAnthropic,
		Status:      StatusActive,
		Schedulable: true,
		GroupIDs:    []int64{groupID},
	}
	svc := &GatewayService{accountRepo: pinnedGatewayAccountRepo{accounts: map[int64]*Account{pinned.ID: pinned}}}
	ctx := context.WithValue(context.Background(), ctxkey.APIKeyPinnedAccountID, pinned.ID)

	account, err := svc.SelectAccountForModelWithExclusions(ctx, &groupID, "sticky-session", "", nil)
	require.NoError(t, err)
	require.Same(t, pinned, account)

	account, err = svc.SelectAccountForModelWithExclusions(ctx, &groupID, "sticky-session", "", map[int64]struct{}{pinned.ID: {}})
	require.Nil(t, account)
	require.ErrorIs(t, err, ErrCafeAccountUnavailable)
}

func TestGeminiMessagesCompatService_PinnedAccountRejectsStickyAndFallback(t *testing.T) {
	groupID := int64(32)
	pinned := &Account{
		ID:          102,
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		GroupIDs:    []int64{groupID},
	}
	other := &Account{
		ID:          103,
		Platform:    PlatformGemini,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		GroupIDs:    []int64{groupID},
	}
	group := &Group{ID: groupID, Platform: PlatformGemini, Status: StatusActive, Hydrated: true}
	ctx := context.WithValue(context.Background(), ctxkey.Group, group)
	ctx = context.WithValue(ctx, ctxkey.APIKeyPinnedAccountID, pinned.ID)
	svc := &GeminiMessagesCompatService{accountRepo: pinnedGatewayAccountRepo{accounts: map[int64]*Account{
		pinned.ID: pinned,
		other.ID:  other,
	}}}

	account, err := svc.SelectAccountForModelWithExclusions(ctx, &groupID, "sticky-session", "", nil)
	require.NoError(t, err)
	require.Same(t, pinned, account)

	account, err = svc.SelectAccountForModelWithExclusions(ctx, &groupID, "sticky-session", "", map[int64]struct{}{pinned.ID: {}})
	require.Nil(t, account)
	require.ErrorIs(t, err, ErrCafeAccountUnavailable)
}

func TestOpenAIGatewayService_PinnedAccountSkipsPreviousResponseSticky(t *testing.T) {
	groupID := int64(33)
	other := Account{
		ID:          104,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		GroupIDs:    []int64{groupID},
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{other}},
		cache:              cache,
		cfg:                newOpenAIWSV2TestConfig(),
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}
	require.NoError(t, store.BindResponseAccount(context.Background(), groupID, "resp_cafe", other.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(context.Background(), &groupID, "resp_cafe", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.NotNil(t, selection, "an ordinary key may use previous_response_id stickiness")
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}

	pinnedCtx := context.WithValue(context.Background(), ctxkey.APIKeyPinnedAccountID, int64(105))
	selection, err = svc.SelectAccountByPreviousResponseID(pinnedCtx, &groupID, "resp_cafe", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.Nil(t, selection, "a Cafe managed key must defer to the pinned scheduler")
}
