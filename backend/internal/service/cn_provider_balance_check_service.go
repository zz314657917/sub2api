package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const (
	cnBalanceLowReasonPrefix = "cn-provider-balance:"
	cnQuotaProbeConcurrency  = 4
)

type cnQuotaProber interface {
	QueryUsage(context.Context, int64) (*CNProviderQuotaProbeResult, error)
}
type CNProviderBalanceCheckService struct {
	accountRepo    AccountRepository
	balanceService *CNProviderBalanceService
	quotaService   cnQuotaProber
	cfg            *config.Config
	interval       time.Duration
	stopCh         chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup
}

func NewCNProviderBalanceCheckService(accountRepo AccountRepository, balance *CNProviderBalanceService, quota *CNProviderQuotaService, cfg *config.Config, interval time.Duration) *CNProviderBalanceCheckService {
	return &CNProviderBalanceCheckService{accountRepo: accountRepo, balanceService: balance, quotaService: quota, cfg: cfg, interval: interval, stopCh: make(chan struct{})}
}
func ProvideCNProviderBalanceCheckService(accountRepo AccountRepository, balance *CNProviderBalanceService, quota *CNProviderQuotaService, cfg *config.Config) *CNProviderBalanceCheckService {
	interval := time.Duration(0)
	if cfg != nil && cfg.Gateway.CNProviders.BalanceCheckIntervalMinutes > 0 {
		interval = time.Duration(cfg.Gateway.CNProviders.BalanceCheckIntervalMinutes) * time.Minute
	}
	service := NewCNProviderBalanceCheckService(accountRepo, balance, quota, cfg, interval)
	service.Start()
	return service
}
func (s *CNProviderBalanceCheckService) Start() {
	if s == nil || s.accountRepo == nil || s.balanceService == nil || s.cfg == nil || !s.cfg.Gateway.CNProviders.BalanceCheckEnabled || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}
func (s *CNProviderBalanceCheckService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}
func (s *CNProviderBalanceCheckService) runOnce() {
	if s == nil || s.accountRepo == nil {
		return
	}
	listCtx, listCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer listCancel()
	var quotas []int64
	var payg []*Account
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		accounts, err := s.accountRepo.ListByPlatform(listCtx, platform)
		if err != nil {
			continue
		}
		for index := range accounts {
			account := &accounts[index]
			if !account.IsActive() {
				continue
			}
			if cnCodingPlanProvider(account) != "" {
				quotas = append(quotas, account.ID)
				continue
			}
			if (platform == PlatformKimi || platform == PlatformDeepseek) && account.Schedulable {
				payg = append(payg, account)
			}
		}
	}
	batches := (len(quotas) + cnQuotaProbeConcurrency - 1) / cnQuotaProbeConcurrency
	timeout := 30*time.Second + time.Duration(batches)*cnProbeUpstreamTimeout + time.Duration(len(payg))*5*time.Second
	if timeout > 5*time.Minute {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if s.quotaService != nil {
		sem := make(chan struct{}, cnQuotaProbeConcurrency)
		var wg sync.WaitGroup
		for _, id := range quotas {
			wg.Add(1)
			go func(accountID int64) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				_, _ = s.quotaService.QueryUsage(ctx, accountID)
			}(id)
		}
		wg.Wait()
	}
	threshold := 0.0
	if s.cfg != nil {
		threshold = s.cfg.Gateway.CNProviders.BalanceThreshold
	}
	for _, account := range payg {
		s.checkOne(ctx, account, threshold)
	}
}
func (s *CNProviderBalanceCheckService) checkOne(ctx context.Context, account *Account, threshold float64) {
	if s == nil || s.balanceService == nil || account == nil {
		return
	}
	result, err := s.balanceService.QueryBalance(ctx, account.ID)
	if err != nil || result == nil || !result.Success {
		return
	}
	if !result.Available || allCNBalancesBelowThreshold(result, threshold) {
		if account.IsSchedulable() {
			_ = s.accountRepo.SetTempUnschedulable(ctx, account.ID, time.Now().Add(s.cooldown()), cnBalanceLowReason(fmt.Sprintf("balance below %.2f", threshold)))
		}
		return
	}
	if account.TempUnschedulableUntil != nil && strings.HasPrefix(account.TempUnschedulableReason, cnBalanceLowReasonPrefix) {
		_ = s.accountRepo.ClearTempUnschedulable(ctx, account.ID)
	}
}
func cnBalanceLowReason(detail string) string { return cnBalanceLowReasonPrefix + detail }
func (s *CNProviderBalanceCheckService) cooldown() time.Duration {
	interval := 10 * time.Minute
	if s != nil && s.cfg != nil && s.cfg.Gateway.CNProviders.BalanceCheckIntervalMinutes > 0 {
		interval = time.Duration(s.cfg.Gateway.CNProviders.BalanceCheckIntervalMinutes) * time.Minute
	}
	return interval * 2
}
