package handler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type v0215WSKeyLookup struct {
	key   *service.APIKey
	err   error
	calls int
}

func (l *v0215WSKeyLookup) GetByKey(_ context.Context, _ string) (*service.APIKey, error) {
	l.calls++
	return l.key, l.err
}

func v0215WSKey() *service.APIKey {
	id := int64(7)
	price := 0.8
	return &service.APIKey{
		ID: 3, Key: "fake-local-key", GroupID: &id, User: &service.User{ID: 9},
		Quota: 100, QuotaUsed: 12, PinnedAccountID: 11, AccountPoolStrategy: "local",
		UnavailableRouteGroupIDs: map[int64]struct{}{99: {}},
		Group: &service.Group{
			ID: 7, Platform: service.PlatformOpenAI, SubscriptionType: "subscription",
			RateMultiplier: 1, ImagePrice1K: &price, RoutingScope: "local-scope",
			ModelRouting: map[string][]int64{"local-model": {11}},
			ModelPricing: []service.ChannelModelPricing{{
				Models: []string{"image"}, ImageInputPrice: &price,
				Intervals: []service.PricingInterval{{InputPrice: &price}},
			}},
		},
	}
}

func TestV0215WSBillingRefreshIdentity(t *testing.T) {
	for _, name := range []string{
		"same", "lookup-error", "nil-key", "key-id", "nil-group-id",
		"group-id", "nil-group", "group-object-id", "platform", "subscription",
	} {
		t.Run(name, func(t *testing.T) {
			conn, latest := v0215WSKey(), v0215WSKey()
			latest.Quota = 999
			latest.User = &service.User{ID: 99}
			latest.Group.RateMultiplier = 4
			latest.Group.PeakRateEnabled = true
			latest.Group.PeakStart, latest.Group.PeakEnd, latest.Group.PeakRateMultiplier = "12:00", "18:00", 2
			latest.Group.LongContextPricingEnabled = true
			latest.Group.ImageRateIndependent, latest.Group.ImageRateMultiplier = true, 3
			latest.Group.RoutingScope = "must-not-replace-routing"
			lookup := &v0215WSKeyLookup{key: latest}
			switch name {
			case "lookup-error":
				lookup.err = errors.New("fake auth-cache unavailable")
			case "nil-key":
				lookup.key = nil
			case "key-id":
				latest.ID++
			case "nil-group-id":
				latest.GroupID = nil
			case "group-id":
				*latest.GroupID++
			case "nil-group":
				latest.Group = nil
			case "group-object-id":
				latest.Group.ID++
			case "platform":
				latest.Group.Platform = "other"
			case "subscription":
				latest.Group.SubscriptionType = "standard"
			}
			var snapshots openAIWSTurnBillingAPIKeys
			snapshots.begin(context.Background(), lookup, 2, conn)
			key := snapshots.take(2, conn)
			if name == "same" {
				require.Equal(t, 4.0, key.Group.RateMultiplier)
				require.True(t, key.Group.PeakRateEnabled)
				require.Equal(t, "12:00", key.Group.PeakStart)
				require.Equal(t, 2.0, key.Group.PeakRateMultiplier)
				require.True(t, key.Group.LongContextPricingEnabled)
				require.True(t, key.Group.ImageRateIndependent)
				require.Equal(t, 3.0, key.Group.ImageRateMultiplier)
			} else {
				require.Equal(t, 1.0, key.Group.RateMultiplier)
			}
			require.Same(t, conn.User, key.User)
			require.Equal(t, conn.Quota, key.Quota)
			require.Equal(t, conn.QuotaUsed, key.QuotaUsed)
			require.Equal(t, conn.PinnedAccountID, key.PinnedAccountID)
			require.Equal(t, conn.Group.RoutingScope, key.Group.RoutingScope)
			require.Equal(t, conn.Group.ModelRouting, key.Group.ModelRouting)
			require.Equal(t, conn.UnavailableRouteGroupIDs, key.UnavailableRouteGroupIDs)
			require.Equal(t, 1, lookup.calls)
			require.Equal(t, 1.0, conn.Group.RateMultiplier)
		})
	}
}

func TestV0215WSBillingActualDelayedWorkerContext(t *testing.T) {
	pool := service.NewUsageRecordWorkerPoolWithOptions(service.UsageRecordWorkerPoolOptions{
		WorkerCount: 1, QueueSize: 8, TaskTimeout: 3 * time.Second,
		OverflowPolicy: "drop", AutoScaleEnabled: false,
	})
	defer pool.Stop()
	h := &OpenAIGatewayHandler{usageRecordWorkerPool: pool}
	blocked, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	pool.Submit(func(context.Context) { close(blocked); <-release })
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("worker did not block")
	}
	conn, latest := v0215WSKey(), v0215WSKey()
	lookup := &v0215WSKeyLookup{key: latest}
	var snapshots openAIWSTurnBillingAPIKeys
	latest.Group.RateMultiplier = 2
	snapshots.begin(context.Background(), lookup, 2, conn)
	capturedKey := snapshots.take(2, conn)
	capturedTrial := &service.NewUserTrialSession{TrialID: 1, UserID: conn.User.ID, RequestID: "trial-local", QuotaLeft: 10}
	parent, cancel := context.WithCancel(context.Background())
	parent = context.WithValue(parent, ctxkey.RequestID, "ledger-local")
	parent = context.WithValue(parent, ctxkey.ClientRequestID, "client-local")
	type observed struct {
		multiplier     float64
		ledger, client string
		trial          *service.NewUserTrialSession
		err            error
	}
	got := make(chan observed, 1)
	mode := h.submitOpenAIUsageRecordTask(parent, &service.OpenAIForwardResult{}, func(taskCtx context.Context) {
		// Same captured-trial restoration as the unchanged production AfterTurn.
		taskCtx = service.WithNewUserTrialSession(taskCtx, capturedTrial)
		trial, _ := service.NewUserTrialSessionFromContext(taskCtx)
		got <- observed{
			multiplier: capturedKey.Group.RateMultiplier,
			ledger:     taskCtx.Value(ctxkey.RequestID).(string),
			client:     taskCtx.Value(ctxkey.ClientRequestID).(string),
			trial:      trial, err: taskCtx.Err(),
		}
	})
	require.NotEqual(t, service.UsageRecordSubmitModeDropped, mode)
	latest.Group.RateMultiplier = 9
	snapshots.begin(context.Background(), lookup, 3, conn)
	require.Equal(t, 9.0, snapshots.take(3, conn).Group.RateMultiplier)
	cancel() // A closed client must not cancel delayed settlement.
	unblock()
	select {
	case result := <-got:
		require.Equal(t, 2.0, result.multiplier)
		require.Equal(t, "ledger-local", result.ledger)
		require.Equal(t, "client-local", result.client)
		require.Same(t, capturedTrial, result.trial)
		require.NoError(t, result.err)
	case <-time.After(3 * time.Second):
		t.Fatal("delayed usage worker did not run")
	}
}

func TestV0215WSBillingInterleavedDelayedUsage(t *testing.T) {
	conn, latest := v0215WSKey(), v0215WSKey()
	lookup := &v0215WSKeyLookup{key: latest}
	var snapshots openAIWSTurnBillingAPIKeys
	// First turn's no-BeforeTurn fallback and explicit first hook do not fetch.
	first := snapshots.take(1, conn)
	snapshots.begin(context.Background(), lookup, 1, conn)
	require.Zero(t, lookup.calls)
	latest.Group.RateMultiplier = 2
	snapshots.begin(context.Background(), lookup, 2, conn)
	snapshots.begin(context.Background(), lookup, 2, conn) // BeforeRequest then BeforeTurn: one lookup.
	latest.Group.RateMultiplier = 3
	snapshots.begin(context.Background(), lookup, 3, conn)
	latest.Group.RateMultiplier = 4
	snapshots.begin(context.Background(), lookup, 4, conn)
	// Older AfterTurn must survive multiple newer BeforeTurn hooks.
	two := snapshots.take(2, conn)
	three := snapshots.take(3, conn)
	four := snapshots.take(4, conn)
	require.Equal(t, 1.0, snapshots.take(1, conn).Group.RateMultiplier)
	require.Empty(t, snapshots.keys)

	// An auth snapshot may be reused and updated after task submission. Nested
	// pricing pointers/slices must not bleed into delayed async RecordUsage.
	*latest.Group.ImagePrice1K = 999
	latest.Group.ModelPricing[0].Models[0] = "changed"
	*latest.Group.ModelPricing[0].Intervals[0].InputPrice = 999
	latest.Group.RateMultiplier = 99
	keys := []*service.APIKey{first, two, three, four}
	ready := make(chan struct{})
	results := make(chan float64, len(keys))
	var wg sync.WaitGroup
	for _, key := range keys {
		captured := key
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			results <- captured.Group.RateMultiplier
		}()
	}
	close(ready)
	wg.Wait()
	close(results)
	var multipliers []float64
	for value := range results {
		multipliers = append(multipliers, value)
	}
	require.ElementsMatch(t, []float64{1, 2, 3, 4}, multipliers)
	for _, key := range keys {
		require.Equal(t, 0.8, *key.Group.ImagePrice1K)
		require.Equal(t, "image", key.Group.ModelPricing[0].Models[0])
		require.Equal(t, 0.8, *key.Group.ModelPricing[0].Intervals[0].InputPrice)
	}
	// Unseen turn always falls back to connection pricing, not the last refresh.
	require.Equal(t, 1.0, snapshots.take(20, conn).Group.RateMultiplier)
	require.Equal(t, 3, lookup.calls)
}

func TestV0215WSBillingMissingConnectionMetadata(t *testing.T) {
	for _, name := range []string{"nil", "key", "group-id", "group", "lookup"} {
		t.Run(name, func(t *testing.T) {
			conn := v0215WSKey()
			lookup := &v0215WSKeyLookup{key: v0215WSKey()}
			var source openAIWSTurnAPIKeyLookup = lookup
			switch name {
			case "nil":
				conn = nil
			case "key":
				conn.Key = ""
			case "group-id":
				conn.GroupID = nil
			case "group":
				conn.Group = nil
			case "lookup":
				source = nil
			}
			require.Same(t, conn, refreshOpenAIWSTurnBillingAPIKey(context.Background(), source, conn))
			require.Zero(t, lookup.calls)
		})
	}
}
