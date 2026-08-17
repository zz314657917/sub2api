package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func gjsonResult(raw string) gjson.Result { return gjson.Parse(raw) }

func TestCNProviderBalanceCheckRunOnceProbesCodingPlanQuota(t *testing.T) {
	account := cnAccount(11, PlatformKimi, AccountModeCoding, "https://api.kimi.com/coding")
	repo := &cnProviderRepoStub{platforms: map[string][]Account{PlatformKimi: {*account}}}
	quota := &cnQuotaStub{}
	checker := NewCNProviderBalanceCheckService(repo, nil, nil, &config.Config{}, 0)
	checker.quotaService = quota
	checker.runOnce()
	require.Equal(t, 1, quota.calls)
}

func TestCNProviderBalanceCheckRunOnceWithoutQuotaService(t *testing.T) {
	account := cnAccount(12, PlatformZhipu, AccountModeCoding, "https://open.bigmodel.cn")
	repo := &cnProviderRepoStub{platforms: map[string][]Account{PlatformZhipu: {*account}}}
	checker := NewCNProviderBalanceCheckService(repo, nil, nil, &config.Config{}, 0)
	require.NotPanics(t, checker.runOnce)
}

func TestCNProviderBalanceCheckClearsOnlyOwnedPause(t *testing.T) {
	account := cnAccount(13, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
	account.Schedulable = false
	until := time.Now().Add(time.Minute)
	account.TempUnschedulableUntil = &until
	account.TempUnschedulableReason = cnBalanceLowReason("prior")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{13: account}}
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusOK, `{"is_available":true,"balance_infos":[{"currency":"USD","total_balance":"5"}]}`)}
	checker := NewCNProviderBalanceCheckService(repo, NewCNProviderBalanceService(repo, nil, upstream, nil), nil, &config.Config{}, 0)
	checker.checkOne(context.Background(), account, 1)
	require.Equal(t, 1, repo.clearCalls)
	account.TempUnschedulableReason = "other-system: pause"
	checker.checkOne(context.Background(), account, 1)
	require.Equal(t, 1, repo.clearCalls)
}
