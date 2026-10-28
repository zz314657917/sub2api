package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamBillingWorkerStoppedIsDistinctFromOverflow(t *testing.T) {
	p := NewUsageRecordWorkerPoolWithOptions(UsageRecordWorkerPoolOptions{
		WorkerCount: 1, QueueSize: 1, TaskTimeout: time.Second, OverflowPolicy: "drop",
	})
	p.Stop()
	var calls atomic.Int32
	mode := p.Submit(func(context.Context) { calls.Add(1) })
	require.Equal(t, UsageRecordSubmitMode("dropped_stopped"), mode)
	require.True(t, mode.Dropped())
	require.Zero(t, calls.Load())
	require.EqualValues(t, 1, p.Stats().DroppedPoolStopped)
	require.Zero(t, p.Stats().DroppedQueueFull)
	for _, m := range []UsageRecordSubmitMode{UsageRecordSubmitModeEnqueued, UsageRecordSubmitModeSync} {
		require.False(t, m.Dropped())
	}
	require.True(t, UsageRecordSubmitModeDropped.Dropped())
}

func TestStreamBillingWorkerShutdownRace(t *testing.T) {
	p := NewUsageRecordWorkerPoolWithOptions(UsageRecordWorkerPoolOptions{
		WorkerCount: 2, QueueSize: 8, TaskTimeout: time.Second, OverflowPolicy: "sync",
	})
	const submissions = 128
	var executed, accepted, rejected atomic.Int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(submissions + 1)
	go func() { defer wg.Done(); <-start; p.Stop() }()
	for range submissions {
		go func() {
			defer wg.Done()
			<-start
			mode := p.Submit(func(ctx context.Context) {
				if ctx.Err() == nil {
					executed.Add(1)
				}
			})
			switch mode {
			case UsageRecordSubmitModeEnqueued, UsageRecordSubmitModeSync:
				accepted.Add(1)
			case UsageRecordSubmitModeDroppedStopped:
				rejected.Add(1)
			default:
				t.Errorf("sync policy cannot explicitly drop: %q", mode)
			}
		}()
	}
	close(start)
	wg.Wait()
	p.Stop()
	require.Equal(t, accepted.Load(), executed.Load(), "every accepted task finishes once")
	require.EqualValues(t, submissions, accepted.Load()+rejected.Load())
	require.EqualValues(t, rejected.Load(), p.Stats().DroppedPoolStopped)
	require.Zero(t, p.Stats().DroppedQueueFull)
}
