package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const groupBuyLifecycleOperationTimeout = 30 * time.Second

type groupBuyLifecycleOperations interface {
	ExpireRounds(ctx context.Context) (int, error)
	RefreshExpiredEntitlements(ctx context.Context) (int, error)
	ReconcilePendingProviderRefunds(ctx context.Context) (int, error)
}

type cafeRoomExpiryOperations interface {
	ExpireCafeRounds(ctx context.Context) (int, error)
}

type cafeRoomLifecycleOperations interface {
	RunCafeLifecycle(ctx context.Context) (int, error)
}

// GroupBuyLifecycleService runs the existing group-buy lifecycle operations.
type GroupBuyLifecycleService struct {
	operations    groupBuyLifecycleOperations
	cafeExpiry    cafeRoomExpiryOperations
	cafeLifecycle cafeRoomLifecycleOperations
	interval      time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	stateMu       sync.Mutex
	startOnce     sync.Once
	stopOnce      sync.Once
	wg            sync.WaitGroup
}

// SetCafeRoomExpiry adds the optional Cafe-only operation without changing the
// existing GroupBuyService lifecycle contract used by older tests and callers.
func (s *GroupBuyLifecycleService) SetCafeRoomExpiry(operations cafeRoomExpiryOperations) {
	if s == nil {
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.cafeExpiry = operations
}

// SetCafeRoomLifecycle replaces the legacy Cafe-expiry-only ticker operation
// with the complete Cafe lifecycle while preserving the old optional hook for
// callers that have not yet wired the new coordinator.
func (s *GroupBuyLifecycleService) SetCafeRoomLifecycle(operations cafeRoomLifecycleOperations) {
	if s == nil {
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.cafeLifecycle = operations
}

func NewGroupBuyLifecycleService(operations groupBuyLifecycleOperations, interval time.Duration) *GroupBuyLifecycleService {
	ctx, cancel := context.WithCancel(context.Background())
	return &GroupBuyLifecycleService{
		operations: operations,
		interval:   interval,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *GroupBuyLifecycleService) Start() {
	if s == nil || s.operations == nil || s.interval <= 0 {
		return
	}
	s.startOnce.Do(func() {
		s.stateMu.Lock()
		defer s.stateMu.Unlock()
		if s.ctx.Err() != nil {
			return
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()

			s.runOnce()
			for {
				select {
				case <-ticker.C:
					s.runOnce()
				case <-s.ctx.Done():
					return
				}
			}
		}()
	})
}

func (s *GroupBuyLifecycleService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.stateMu.Lock()
		defer s.stateMu.Unlock()
		if s.cancel != nil {
			s.cancel()
		}
	})
	s.wg.Wait()
}

func (s *GroupBuyLifecycleService) runOnce() {
	if s == nil || s.operations == nil || s.ctx == nil || s.ctx.Err() != nil {
		return
	}
	s.runOperation("expire rounds", s.operations.ExpireRounds)
	if s.ctx.Err() != nil {
		return
	}
	s.stateMu.Lock()
	cafeExpiry := s.cafeExpiry
	cafeLifecycle := s.cafeLifecycle
	s.stateMu.Unlock()
	if cafeLifecycle != nil {
		s.runOperation("run cafe lifecycle", cafeLifecycle.RunCafeLifecycle)
		if s.ctx.Err() != nil {
			return
		}
	} else if cafeExpiry != nil {
		s.runOperation("expire cafe rounds", cafeExpiry.ExpireCafeRounds)
		if s.ctx.Err() != nil {
			return
		}
	}
	s.runOperation("refresh expired entitlements", s.operations.RefreshExpiredEntitlements)
	if s.ctx.Err() != nil {
		return
	}
	s.runOperation("reconcile provider refunds", s.operations.ReconcilePendingProviderRefunds)
}

func (s *GroupBuyLifecycleService) runOperation(name string, operation func(context.Context) (int, error)) {
	ctx, cancel := context.WithTimeout(s.ctx, groupBuyLifecycleOperationTimeout)
	defer cancel()
	updated, err := operation(ctx)
	if err != nil {
		if s.ctx.Err() != nil {
			return
		}
		slog.Warn("[GroupBuyLifecycle] operation failed", "operation", name, "error", err)
		return
	}
	if updated > 0 {
		slog.Info("[GroupBuyLifecycle] operation completed", "operation", name, "count", updated)
	}
}
