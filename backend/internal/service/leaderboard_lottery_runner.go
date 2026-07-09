package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const leaderboardLotteryRunnerInterval = time.Minute

// LeaderboardLotteryRunner periodically settles due weekly leaderboard lotteries.
// Settlement remains idempotent through the repository unique constraints.
type LeaderboardLotteryRunner struct {
	usageService *UsageService
	interval     time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
}

func NewLeaderboardLotteryRunner(usageService *UsageService, interval time.Duration) *LeaderboardLotteryRunner {
	if interval <= 0 {
		interval = leaderboardLotteryRunnerInterval
	}
	return &LeaderboardLotteryRunner{
		usageService: usageService,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

func (r *LeaderboardLotteryRunner) Start() {
	if r == nil || r.usageService == nil || r.interval <= 0 {
		return
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		r.runOnce()
		for {
			select {
			case <-ticker.C:
				r.runOnce()
			case <-r.stopCh:
				return
			}
		}
	}()
}

func (r *LeaderboardLotteryRunner) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	r.wg.Wait()
}

func (r *LeaderboardLotteryRunner) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := r.usageService.SettleDueLeaderboardLotteryRewards(ctx, timezone.Now()); err != nil &&
		!errors.Is(err, ErrLeaderboardDailyRewardUnavailable) {
		log.Printf("[LeaderboardLottery] settle due rewards failed: %v", err)
	}
}

func ProvideLeaderboardLotteryRunner(usageService *UsageService) *LeaderboardLotteryRunner {
	runner := NewLeaderboardLotteryRunner(usageService, leaderboardLotteryRunnerInterval)
	runner.Start()
	return runner
}
