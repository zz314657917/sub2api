package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type groupBuyLifecycleOperationsStub struct {
	mu       sync.Mutex
	counts   map[string]int
	firstErr error
}

type blockingGroupBuyLifecycleOperationsStub struct {
	started chan struct{}
}

func (s *blockingGroupBuyLifecycleOperationsStub) ExpireRounds(ctx context.Context) (int, error) {
	close(s.started)
	<-ctx.Done()
	return 0, ctx.Err()
}

func (s *blockingGroupBuyLifecycleOperationsStub) RefreshExpiredEntitlements(context.Context) (int, error) {
	panic("unexpected refresh after lifecycle cancellation")
}

func (s *blockingGroupBuyLifecycleOperationsStub) ReconcilePendingProviderRefunds(context.Context) (int, error) {
	panic("unexpected reconciliation after lifecycle cancellation")
}

func (s *groupBuyLifecycleOperationsStub) call(name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[name]++
	if name == "expire" && s.firstErr != nil {
		return 0, s.firstErr
	}
	return s.counts[name], nil
}

func (s *groupBuyLifecycleOperationsStub) ExpireRounds(context.Context) (int, error) {
	return s.call("expire")
}

func (s *groupBuyLifecycleOperationsStub) RefreshExpiredEntitlements(context.Context) (int, error) {
	return s.call("entitlements")
}

func (s *groupBuyLifecycleOperationsStub) ReconcilePendingProviderRefunds(context.Context) (int, error) {
	return s.call("refunds")
}

func (s *groupBuyLifecycleOperationsStub) snapshot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.counts))
	for key, value := range s.counts {
		out[key] = value
	}
	return out
}

func TestGroupBuyLifecycleServiceRunsImmediatelyAndOnInterval(t *testing.T) {
	ops := &groupBuyLifecycleOperationsStub{counts: map[string]int{}}
	svc := NewGroupBuyLifecycleService(ops, 10*time.Millisecond)
	svc.Start()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		counts := ops.snapshot()
		if counts["expire"] >= 2 && counts["entitlements"] >= 2 && counts["refunds"] >= 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	svc.Stop()

	counts := ops.snapshot()
	require.GreaterOrEqual(t, counts["expire"], 2)
	require.GreaterOrEqual(t, counts["entitlements"], 2)
	require.GreaterOrEqual(t, counts["refunds"], 2)

	stoppedCounts := ops.snapshot()
	time.Sleep(30 * time.Millisecond)
	require.Equal(t, stoppedCounts, ops.snapshot())
}

func TestGroupBuyLifecycleServiceContinuesAfterOperationError(t *testing.T) {
	ops := &groupBuyLifecycleOperationsStub{
		counts:   map[string]int{},
		firstErr: context.Canceled,
	}
	svc := NewGroupBuyLifecycleService(ops, time.Hour)
	svc.runOnce()

	counts := ops.snapshot()
	require.Equal(t, 1, counts["expire"])
	require.Equal(t, 1, counts["entitlements"])
	require.Equal(t, 1, counts["refunds"])
}

func TestGroupBuyLifecycleServiceIgnoresInvalidStartConfiguration(t *testing.T) {
	ops := &groupBuyLifecycleOperationsStub{counts: map[string]int{}}
	for _, interval := range []time.Duration{0, -time.Millisecond} {
		svc := NewGroupBuyLifecycleService(ops, interval)
		svc.Start()
		svc.Stop()
	}
	NewGroupBuyLifecycleService(nil, time.Millisecond).Start()
	require.Empty(t, ops.snapshot())
}

func TestGroupBuyLifecycleServiceStopCancelsRunningOperation(t *testing.T) {
	ops := &blockingGroupBuyLifecycleOperationsStub{started: make(chan struct{})}
	svc := NewGroupBuyLifecycleService(ops, time.Hour)
	svc.Start()

	select {
	case <-ops.started:
	case <-time.After(time.Second):
		t.Fatal("lifecycle operation did not start")
	}

	stopped := make(chan struct{})
	go func() {
		svc.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("lifecycle stop did not cancel the running operation")
	}
}
