package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestMonitorQuotaRawValueErrorsRedactActualCredentials(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformDeepseek, PlatformOpenAI} {
		t.Run(platform, func(t *testing.T) {
			f, usage, quota, balance, accounts := newQuotaFetcherTestSetup(t)
			secret := "arbitrary-account-secret-which-has-no-known-prefix"
			account := &Account{ID: 101, Platform: platform, Credentials: map[string]any{"api_key": secret}}
			if platform == PlatformKimi {
				account.Credentials["account_mode"] = AccountModeCoding
			}
			accounts.accounts[101] = account
			quota.result = &CNProviderQuotaProbeResult{Error: "raw upstream says " + secret, StatusCode: 401}
			balance.result = &CNProviderBalanceResult{Error: "raw upstream says " + secret, StatusCode: 403}
			usage.usage = &UsageInfo{Error: "raw upstream says " + secret}
			snapshot := f.Fetch(context.Background(), 101)
			require.False(t, snapshot.Success)
			encoded, err := json.Marshal(snapshot)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), secret)
			require.Contains(t, snapshot.Error, "[REDACTED]")
			require.NotContains(t, deriveQuotaCheckResult(snapshot, "quota", time.Now()).Message, secret)
		})
	}
}

func TestMonitorNineProviderCapabilityAndLegacyDefaults(t *testing.T) {
	require.Len(t, monitorProviders, 9)
	for provider := range monitorProviders {
		wantProbe := provider != MonitorProviderAntigravity
		require.Equal(t, wantProbe, providerSupportsProbe(provider), provider)
		require.Equal(t, provider == MonitorProviderAntigravity, validateCheckMode(provider, "") != nil, provider)
		require.Equal(t, provider == MonitorProviderMiniMax, validateCheckMode(provider, MonitorCheckModeQuota) != nil, provider)
		require.Equal(t, provider != MonitorProviderOpenAI, validateAPIMode(provider, MonitorAPIModeResponses) != nil, provider)
	}
	require.Error(t, validateProvider("opencode_go"))
	require.Equal(t, MonitorCheckModeProbe, defaultCheckMode(""))
	require.Equal(t, MonitorAPIModeChatCompletions, defaultAPIMode(""))
	require.NoError(t, validateJitter(0, 15))
	require.Error(t, validateJitter(1, 15))
	require.Error(t, validateJitter(-1, 60))
}

func TestMonitorSnapshotPrimarySummaryRetainsUserProjection(t *testing.T) {
	snapshot := &domain.MonitorQuotaSnapshot{Success: true, Balance: new(float64)}
	summary := buildStatusSummary(map[string]*ChannelMonitorLatest{"quota": {Model: "quota", Quota: snapshot}},
		nil, "quota", nil)
	require.Same(t, snapshot, summary.LatestQuota)
	view := buildUserViewFromSummary(&ChannelMonitor{ID: 1, PrimaryModel: "quota"}, summary, nil, nil)
	encoded, err := json.Marshal(view)
	require.NoError(t, err)
	require.False(t, strings.Contains(strings.ToLower(string(encoded)), "quota\":{"))
	require.NotContains(t, string(encoded), "Balance")
	require.NotContains(t, string(encoded), "AccountID")
}

type monitorCRUDRepo struct {
	quotaModeRepoStub
	created *ChannelMonitor
}

func (r *monitorCRUDRepo) Create(_ context.Context, m *ChannelMonitor) error {
	copy := *m
	copy.ID = 88
	r.created = &copy
	m.ID = copy.ID
	r.monitor = &copy
	return nil
}

func TestMonitorCRUDQuotaBindingAndExplicitUnbind(t *testing.T) {
	repo := &monitorCRUDRepo{}
	svc := NewChannelMonitorService(repo, &monitorQuotaTestEncryptor{})
	accounts := map[int64]*Account{12: {ID: 12, Platform: PlatformOpenAI, Type: AccountTypeOAuth}}
	svc.SetQuotaFetcher(newQuotaModeFetcher(accounts, nil))
	created, err := svc.Create(context.Background(), ChannelMonitorCreateParams{
		Name: "quota", Provider: MonitorProviderOpenAI, CheckMode: MonitorCheckModeQuota,
		AccountID: monitorQuotaInt64Ptr(12), Enabled: true, IntervalSeconds: 60, JitterSeconds: 4,
	})
	require.NoError(t, err)
	require.Empty(t, created.APIKey)
	require.Equal(t, "NEW:", repo.created.APIKey, "empty quota key is still encrypted at rest")
	require.Equal(t, "quota", created.PrimaryModel)
	require.Equal(t, MonitorAPIModeChatCompletions, created.APIMode)
	require.Equal(t, int64(12), *created.AccountID)
	_, err = svc.Update(context.Background(), 88, ChannelMonitorUpdateParams{Name: monitorQuotaStrPtr("renamed")})
	require.NoError(t, err)
	require.Equal(t, int64(12), *repo.updated[0].AccountID)
	_, err = svc.Update(context.Background(), 88, ChannelMonitorUpdateParams{
		CheckMode: monitorQuotaStrPtr(MonitorCheckModeProbe), AccountID: monitorQuotaInt64Ptr(0),
		Endpoint: monitorQuotaStrPtr("https://8.8.8.8"), APIKey: monitorQuotaStrPtr("fresh-local-secret"),
		PrimaryModel: monitorQuotaStrPtr("gpt-test"),
	})
	require.NoError(t, err)
	require.Nil(t, repo.updated[1].AccountID)
	require.Equal(t, MonitorCheckModeProbe, repo.updated[1].CheckMode)
	require.Equal(t, "NEW:fresh-local-secret", repo.updated[1].APIKey)
}

func TestMonitorProbeRedactionPrecedesTruncation(t *testing.T) {
	secret := "unusual-key-value-that-is-not-a-known-provider-pattern"
	message := monitorProbeMessage(strings.Repeat("x", 490)+secret, secret)
	require.NotContains(t, message, secret[:10])
}

func TestMonitorLocalQuotaAccountCapabilityDoesNotOverFilter(t *testing.T) {
	for _, platform := range []string{PlatformGrok, PlatformGemini, PlatformAntigravity} {
		require.NoError(t, monitorAccountQuotaCapability(&Account{Platform: platform, Type: AccountTypeAPIKey}), platform)
	}
	for _, platform := range []string{PlatformKimi, PlatformDeepseek} {
		require.NoError(t, monitorAccountQuotaCapability(&Account{Platform: platform}), "absent mode routes to payg balance")
	}
}

func TestMonitorCodingHostCapabilityMatchesLocalFrontendMatrix(t *testing.T) {
	for _, tc := range []struct {
		provider string
		base     string
		valid    bool
	}{
		{PlatformKimi, "", true},
		{PlatformKimi, "https://API.MOONSHOT.CN/v1", true},
		{PlatformKimi, "https://coding.kimi.com", true},
		{PlatformKimi, "https://api.moonshot.ai", true},
		{PlatformKimi, "https://moonshot.cn.evil.example", false},
		{PlatformKimi, "https://relay.example", false},
		{PlatformZhipu, "", true},
		{PlatformZhipu, "https://api.z.ai/api/coding/paas/v4", true},
		{PlatformZhipu, "https://OPEN.BIGMODEL.CN", true},
		{PlatformZhipu, "https://bigmodel.cn.evil.example", false},
		{PlatformZhipu, "malformed endpoint", false},
		{PlatformDeepseek, "", false},
	} {
		account := &Account{Platform: tc.provider, Credentials: map[string]any{
			"account_mode": " coding ", "base_url": tc.base,
		}}
		require.Equal(t, tc.valid, monitorAccountQuotaCapability(account) == nil, "%s %s", tc.provider, tc.base)
	}
}

func TestRunCheck_QuotaRejectsChangedAccountPlatform(t *testing.T) {
	repo := &quotaModeRepoStub{monitor: &ChannelMonitor{ID: 1, Provider: MonitorProviderOpenAI,
		CheckMode: MonitorCheckModeQuota, PrimaryModel: "quota", AccountID: monitorQuotaInt64Ptr(12)}}
	svc := newQuotaModeService(repo)
	fetcher := newQuotaModeFetcher(map[int64]*Account{12: {ID: 12, Platform: PlatformKimi}}, nil)
	svc.SetQuotaFetcher(fetcher)
	results, err := svc.RunCheck(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, MonitorStatusError, results[0].Status)
	require.False(t, results[0].Quota.Success)
	require.Contains(t, results[0].Quota.Error, "does not match")
	require.Len(t, repo.history, 1)
}
