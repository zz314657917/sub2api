package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type cnProviderRepoStub struct {
	AccountRepository
	accounts   map[int64]*Account
	platforms  map[string][]Account
	updates    []map[string]any
	setCalls   int
	clearCalls int
}

func (r *cnProviderRepoStub) GetByID(_ context.Context, id int64) (*Account, error) {
	return r.accounts[id], nil
}
func (r *cnProviderRepoStub) ListByPlatform(_ context.Context, platform string) ([]Account, error) {
	return r.platforms[platform], nil
}
func (r *cnProviderRepoStub) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.updates = append(r.updates, updates)
	return nil
}
func (r *cnProviderRepoStub) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, _ string) error {
	r.setCalls++
	return nil
}
func (r *cnProviderRepoStub) ClearTempUnschedulable(_ context.Context, _ int64) error {
	r.clearCalls++
	return nil
}

type cnCountingUpstream struct {
	calls    int
	proxy    string
	response *http.Response
	err      error
}

func (u *cnCountingUpstream) Do(_ *http.Request, proxy string, _ int64, _ int) (*http.Response, error) {
	u.calls++
	u.proxy = proxy
	return u.response, u.err
}
func (u *cnCountingUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

type cnQuotaStub struct{ calls int }

func (s *cnQuotaStub) QueryUsage(_ context.Context, _ int64) (*CNProviderQuotaProbeResult, error) {
	s.calls++
	return &CNProviderQuotaProbeResult{Success: true}, nil
}

func cnResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

type cnFailingReadCloser struct{}

func (cnFailingReadCloser) Read([]byte) (int, error) { return 0, errors.New("injected read failure") }
func (cnFailingReadCloser) Close() error             { return nil }
func cnAccount(id int64, platform, mode, base string) *Account {
	return &Account{ID: id, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "secret", "account_mode": mode, "base_url": base}}
}
func cnAllowlist(hosts ...string) *config.Config {
	return &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: true, UpstreamHosts: hosts}}}
}

func TestCNProviderQuotaService_RejectsURLBlockedByPolicy(t *testing.T) {
	account := cnAccount(1, PlatformKimi, AccountModeCoding, "https://blocked.example")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{1: account}}
	upstream := &cnCountingUpstream{}
	_, err := NewCNProviderQuotaService(repo, nil, upstream, cnAllowlist("api.kimi.com")).QueryUsage(context.Background(), 1)
	require.Error(t, err)
	require.Zero(t, upstream.calls, "rejected final URL must cause zero outbound requests")
}

func TestCNProviderBalanceService_RejectsURLBlockedByPolicy(t *testing.T) {
	account := cnAccount(2, PlatformDeepseek, AccountModePayG, "https://blocked.example")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{2: account}}
	upstream := &cnCountingUpstream{}
	_, err := NewCNProviderBalanceService(repo, nil, upstream, cnAllowlist("api.deepseek.com")).QueryBalance(context.Background(), 2)
	require.Error(t, err)
	require.Zero(t, upstream.calls, "rejected final URL must cause zero outbound requests")
}

func TestCNProviderBalanceService_OfficialHostPassesValidation(t *testing.T) {
	account := cnAccount(3, PlatformKimi, AccountModePayG, "https://ignored.example")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{3: account}}
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusOK, `{"data":{"available_balance":"12.5"}}`)}
	result, err := NewCNProviderBalanceService(repo, nil, upstream, cnAllowlist("api.moonshot.cn")).QueryBalance(context.Background(), 3)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, 1, upstream.calls)
}

func TestCNProviderBalanceService_DeepSeekKeepsAccountWhenAnyCurrencyHealthy(t *testing.T) {
	account := cnAccount(4, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{4: account}}
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusOK, `{"is_available":true,"balance_infos":[{"currency":"CNY","total_balance":"0"},{"currency":"USD","total_balance":"1.25"}]}`)}
	result, err := NewCNProviderBalanceService(repo, nil, upstream, cnAllowlist("api.deepseek.com")).QueryBalance(context.Background(), 4)
	require.NoError(t, err)
	require.False(t, allCNBalancesBelowThreshold(result, 1))
	require.Len(t, result.Balances, 2)
}

func TestCNProviderQuotaService_ProbeFailurePreservesSnapshot(t *testing.T) {
	account := cnAccount(5, PlatformKimi, AccountModeCoding, "https://api.kimi.com/coding")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{5: account}}
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusBadGateway, "failure")}
	result, err := NewCNProviderQuotaService(repo, nil, upstream, nil).QueryUsage(context.Background(), 5)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Empty(t, repo.updates)
}

func TestCNProviderQuotaService_ReadFailurePreservesSnapshot(t *testing.T) {
	account := cnAccount(6, PlatformKimi, AccountModeCoding, "https://api.kimi.com/coding")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{6: account}}
	upstream := &cnCountingUpstream{response: &http.Response{StatusCode: http.StatusOK, Body: cnFailingReadCloser{}}}
	result, err := NewCNProviderQuotaService(repo, nil, upstream, nil).QueryUsage(context.Background(), 6)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Equal(t, "response_read_failed", result.Error)
	require.Empty(t, repo.updates)
}

func TestCNProviderBalanceService_ReadFailurePreservesSnapshotAndDoesNotPause(t *testing.T) {
	account := cnAccount(7, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{7: account}}
	upstream := &cnCountingUpstream{response: &http.Response{StatusCode: http.StatusOK, Body: cnFailingReadCloser{}}}
	service := NewCNProviderBalanceService(repo, nil, upstream, nil)
	result, err := service.QueryBalance(context.Background(), 7)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Equal(t, "response_read_failed", result.Error)
	require.Empty(t, repo.updates)

	checkerUpstream := &cnCountingUpstream{response: &http.Response{StatusCode: http.StatusOK, Body: cnFailingReadCloser{}}}
	checker := NewCNProviderBalanceCheckService(repo, NewCNProviderBalanceService(repo, nil, checkerUpstream, nil), nil, &config.Config{}, 0)
	checker.checkOne(context.Background(), account, 1)
	require.Zero(t, repo.setCalls)
	require.Empty(t, repo.updates)
}

func TestCNProviderProbeInvalidPayloadPreservesSnapshot(t *testing.T) {
	t.Run("quota", func(t *testing.T) {
		account := cnAccount(8, PlatformKimi, AccountModeCoding, "https://api.kimi.com/coding")
		repo := &cnProviderRepoStub{accounts: map[int64]*Account{8: account}}
		result, err := NewCNProviderQuotaService(repo, nil, &cnCountingUpstream{response: cnResponse(http.StatusOK, `{}`)}, nil).QueryUsage(context.Background(), 8)
		require.NoError(t, err)
		require.False(t, result.Success)
		require.Equal(t, "invalid_response", result.Error)
		require.Empty(t, repo.updates)
	})
	t.Run("balance", func(t *testing.T) {
		account := cnAccount(9, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
		repo := &cnProviderRepoStub{accounts: map[int64]*Account{9: account}}
		result, err := NewCNProviderBalanceService(repo, nil, &cnCountingUpstream{response: cnResponse(http.StatusOK, `{}`)}, nil).QueryBalance(context.Background(), 9)
		require.NoError(t, err)
		require.False(t, result.Success)
		require.Equal(t, "invalid_response", result.Error)
		require.Empty(t, repo.updates)
	})
}

func TestAllCNBalancesBelowThreshold(t *testing.T) {
	require.True(t, allCNBalancesBelowThreshold(&CNProviderBalanceResult{Balances: []CNProviderBalanceEntry{{Balance: .2}, {Balance: .9}}}, 1))
	require.False(t, allCNBalancesBelowThreshold(&CNProviderBalanceResult{Balances: []CNProviderBalanceEntry{{Balance: .2}, {Balance: 1}}}, 1))
}

func TestParseKimiUsageTiers(t *testing.T) {
	tiers := parseKimiUsageTiers([]byte(`{"limits":[{"detail":{"limit":100,"remaining":25,"resetTime":"2026-01-01T00:00:00Z"}}],"usage":{"limit":"200","remaining":"50"}}`))
	require.Len(t, tiers, 2)
	require.Equal(t, "5h", tiers[0].Window)
	require.Equal(t, 75.0, tiers[0].UsedPercent)
	require.Equal(t, "weekly", tiers[1].Window)
}

func TestParseZhipuTokenTiers_UnitClassification(t *testing.T) {
	tiers := parseZhipuTokenTiers(gjsonResult(`{"limits":[{"type":"TOKENS_LIMIT","unit":6,"percentage":80,"nextResetTime":2000000000000},{"type":"TOKENS_LIMIT","unit":3,"percentage":20,"nextResetTime":1000000000000}]}`))
	require.Len(t, tiers, 2)
	require.Equal(t, "5h", tiers[0].Window)
	require.Equal(t, 20.0, tiers[0].UsedPercent)
	require.Equal(t, "weekly", tiers[1].Window)
}

func TestParseZhipuTokenTiers_IgnoresNonTokenEntries(t *testing.T) {
	tiers := parseZhipuTokenTiers(gjsonResult(`{"limits":[{"type":"CREDIT_LIMIT","unit":3,"percentage":99},{"type":"TOKENS_LIMIT","unit":6,"percentage":40}]}`))
	require.Len(t, tiers, 1)
	require.Equal(t, "weekly", tiers[0].Window)
	require.Equal(t, 40.0, tiers[0].UsedPercent)
}

func TestCNQuotaExtraUpdates(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updates := cnQuotaExtraUpdates(PlatformKimi, []CNQuotaTier{{Window: "5h", UsedPercent: 10, ResetAt: "2026-01-01T01:00:00Z"}, {Window: "weekly", UsedPercent: 20, ResetAt: "2026-01-08T00:00:00Z"}}, now)
	require.Equal(t, 10.0, updates["kimi_5h_used_percent"])
	require.Equal(t, "2026-01-01T01:00:00Z", updates["kimi_5h_reset_at"])
	require.Equal(t, 20.0, updates["kimi_weekly_used_percent"])
}

func TestCNBalanceURL(t *testing.T) {
	require.Equal(t, "https://api.moonshot.cn/v1/users/me/balance", cnBalanceURL(cnAccount(1, PlatformKimi, AccountModePayG, "https://x")))
	require.Equal(t, "https://api.deepseek.com/user/balance", cnBalanceURL(cnAccount(1, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")))
}
func TestKimiQuotaURL(t *testing.T) {
	require.Equal(t, "https://api.kimi.com/coding/v1/usages", kimiQuotaURL("https://api.kimi.com/coding/v1"))
}
func TestZhipuQuotaHost(t *testing.T) {
	require.Equal(t, "https://open.bigmodel.cn", zhipuQuotaHost("https://open.bigmodel.cn/api/paas/v4"))
	require.Equal(t, "https://api.z.ai", zhipuQuotaHost("https://api.z.ai/api/paas/v4"))
}
