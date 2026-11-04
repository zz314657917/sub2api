package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type upstreamFiveSchedulerRepo struct {
	schedulerTestOpenAIAccountRepo
	reads  []int64
	lookup func(int64, int) (*Account, error)
}

func (r *upstreamFiveSchedulerRepo) GetByID(ctx context.Context, id int64) (*Account, error) {
	r.reads = append(r.reads, id)
	if r.lookup != nil {
		return r.lookup(id, len(r.reads))
	}
	return r.schedulerTestOpenAIAccountRepo.GetByID(ctx, id)
}

type upstreamFiveConcurrencyCache struct {
	schedulerTestConcurrencyCache
	attempts []int64
}

func (c *upstreamFiveConcurrencyCache) AcquireAccountSlot(ctx context.Context, id int64, maxConcurrency int, requestID string) (bool, error) {
	c.attempts = append(c.attempts, id)
	return c.schedulerTestConcurrencyCache.AcquireAccountSlot(ctx, id, maxConcurrency, requestID)
}

func upstreamFiveSchedulerAccount(id int64, priority int, caps map[string]any) Account {
	return Account{ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive,
		Schedulable: true, Concurrency: 1, Priority: priority,
		Credentials: map[string]any{"openai_capabilities": caps}}
}

func TestUpstreamFiveSchedulerCapabilityBeforeTopK(t *testing.T) {
	ctx := context.Background()
	groupID := int64(110)
	denied := upstreamFiveSchedulerAccount(37001, 0, map[string]any{"chat_completions": true, "embeddings": false})
	allowed := upstreamFiveSchedulerAccount(37002, 5, map[string]any{"embeddings": true})
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.LBTopK = 1
	snapshot := &openAISnapshotCacheStub{snapshotAccounts: []*Account{&denied, &allowed},
		accountsByID: map[int64]*Account{denied.ID: &denied, allowed.ID: &allowed}}
	repo := &upstreamFiveSchedulerRepo{schedulerTestOpenAIAccountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{denied, allowed}}}
	concurrency := &upstreamFiveConcurrencyCache{}
	svc := &OpenAIGatewayService{
		accountRepo: repo, cache: &schedulerTestGatewayCache{}, cfg: cfg,
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshot},
		concurrencyService: NewConcurrencyService(concurrency),
	}
	selection, decision, err := svc.SelectAccountWithSchedulerForCapability(ctx, &groupID, "", "", "", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityEmbeddings, false)
	require.NoError(t, err)
	require.NotNil(t, selection)
	t.Cleanup(selection.ReleaseFunc)
	require.Equal(t, allowed.ID, selection.Account.ID)
	require.Equal(t, openAIAccountScheduleLayerLoadBalance, decision.Layer)
	require.Equal(t, 1, decision.TopK)
	require.Equal(t, 1, decision.CandidateCount, "denied metadata must be removed before top-K")
	require.Equal(t, []int64{allowed.ID}, concurrency.attempts)
	require.NotContains(t, repo.reads, denied.ID, "denied metadata should not need hydration")
}

func TestUpstreamFiveSchedulerFullAccountCapabilityRecheck(t *testing.T) {
	for _, wait := range []bool{false, true} {
		name := "acquire"
		if wait {
			name = "wait_plan"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			groupID := int64(110)
			stale := upstreamFiveSchedulerAccount(38001, 0, map[string]any{"embeddings": true})
			latest := upstreamFiveSchedulerAccount(stale.ID, 0, map[string]any{"embeddings": false})
			cfg := &config.Config{}
			cfg.Gateway.OpenAIWS.LBTopK = 1
			snapshot := &openAISnapshotCacheStub{snapshotAccounts: []*Account{&stale},
				accountsByID: map[int64]*Account{stale.ID: &stale}}
			repo := &upstreamFiveSchedulerRepo{lookup: func(id int64, read int) (*Account, error) {
				if wait && read == 1 {
					return &stale, nil
				}
				return &latest, nil
			}}
			concurrency := &upstreamFiveConcurrencyCache{schedulerTestConcurrencyCache: schedulerTestConcurrencyCache{
				acquireResults: map[int64]bool{stale.ID: !wait},
			}}
			svc := &OpenAIGatewayService{
				accountRepo: repo, cache: &schedulerTestGatewayCache{}, cfg: cfg,
				rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
				schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshot},
				concurrencyService: NewConcurrencyService(concurrency),
			}
			selection, decision, err := svc.SelectAccountWithSchedulerForCapability(ctx, &groupID, "", "", "", nil,
				OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityEmbeddings, false)
			require.Error(t, err)
			require.Nil(t, selection, "revoked capability must neither acquire nor return a wait plan")
			require.Equal(t, openAIAccountScheduleLayerLoadBalance, decision.Layer)
			require.Equal(t, 1, decision.CandidateCount)
			require.Contains(t, repo.reads, stale.ID, "full DB account must be rechecked")
			if wait {
				require.Equal(t, []int64{stale.ID, stale.ID}, repo.reads, "wait path must repeat the full-account capability check")
				require.Equal(t, []int64{stale.ID}, concurrency.attempts)
			} else {
				require.Empty(t, concurrency.attempts, "revoked capability must be rejected before slot acquisition")
			}
		})
	}
}
