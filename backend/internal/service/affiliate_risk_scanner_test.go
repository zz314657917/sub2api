package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type affiliateRiskSettingRepoStub map[string]string

func (r affiliateRiskSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	value, err := r.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	return &Setting{Key: key, Value: value, UpdatedAt: time.Now()}, nil
}

func (r affiliateRiskSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r[key]
	if !ok {
		return "", errors.New("setting not found")
	}
	return value, nil
}

func (r affiliateRiskSettingRepoStub) Set(_ context.Context, key, value string) error {
	r[key] = value
	return nil
}

func (r affiliateRiskSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = r[key]
	}
	return out, nil
}

func (r affiliateRiskSettingRepoStub) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		r[key] = value
	}
	return nil
}

func (r affiliateRiskSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r))
	for key, value := range r {
		out[key] = value
	}
	return out, nil
}

func (r affiliateRiskSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(r, key)
	return nil
}

func TestAffiliateRiskScoreP1(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	rewardAt := now.Add(10 * time.Minute)
	revokedAt := now.Add(20 * time.Minute)
	cluster := AffiliateRiskCluster{
		InviterID:          100,
		InviterEmail:       "3388637010@qq.com",
		InviterLastLoginIP: "2409:8962:e1:391d:7d22:7006:9425:c2f8",
		Invitees: []AffiliateRiskInvitee{
			{UserID: 201, Email: "test001@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2409:8962:e1:391d:aaaa::1", CreatedAt: now, APICallRewardAt: &rewardAt},
			{UserID: 202, Email: "test002@gmail.com", RegisterIP: "2.2.2.2", LastLoginIP: "2409:8962:e1:391d:bbbb::1", CreatedAt: now.Add(time.Minute), APICallRewardAt: &rewardAt},
			{UserID: 203, Email: "test003@gmail.com", RegisterIP: "3.3.3.3", LastLoginIP: "2409:8962:e1:391d:cccc::1", CreatedAt: now.Add(2 * time.Minute), APICallRewardAt: &rewardAt, AffiliateRevokedAt: &revokedAt},
		},
	}

	result := scoreAffiliateRiskCluster(cluster, now.Add(-12*time.Hour), now)
	require.Equal(t, "P1", result.Severity)
	require.GreaterOrEqual(t, result.Score, 90)
	require.Contains(t, result.Title, "疑似刷邀请奖励")
	require.Contains(t, result.Title, "3388637010@qq.com")
	require.NotEmpty(t, result.Fingerprint)
}

func TestAffiliateRiskScoreP2NoAutoBanSignal(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	cluster := AffiliateRiskCluster{
		InviterID:          100,
		InviterEmail:       "risk@example.com",
		InviterLastLoginIP: "2409:8962:e1:391d::1",
		Invitees: []AffiliateRiskInvitee{
			{UserID: 201, Email: "ab001@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2409:8962:e1:391d::2", CreatedAt: now},
			{UserID: 202, Email: "ab002@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2409:8962:e1:391d::3", CreatedAt: now},
			{UserID: 203, Email: "ab003@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2409:8962:e1:391d::4", CreatedAt: now},
		},
	}

	result := scoreAffiliateRiskCluster(cluster, now.Add(-12*time.Hour), now)
	require.Equal(t, "P2", result.Severity)
	require.GreaterOrEqual(t, result.Score, 70)
	require.Less(t, result.Score, 90)
}

func TestAffiliateRiskScoreBelowThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	cluster := AffiliateRiskCluster{
		InviterID: 100,
		Invitees: []AffiliateRiskInvitee{
			{UserID: 201, Email: "single@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2.2.2.2", CreatedAt: now},
		},
	}

	result := scoreAffiliateRiskCluster(cluster, now.Add(-12*time.Hour), now)
	require.Empty(t, result.Severity)
	require.Less(t, result.Score, 50)
}

func TestAffiliateRiskScoreUsesFirstUsageIPAggregation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
	cluster := AffiliateRiskCluster{
		InviterID:          100,
		InviterEmail:       "usage-ip@example.com",
		InviterLastLoginIP: "8.8.8.8",
		Invitees: []AffiliateRiskInvitee{
			{UserID: 201, Email: "ua001@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "9.9.9.1", FirstUsageIP: "8.8.8.8", CreatedAt: now},
			{UserID: 202, Email: "ua002@gmail.com", RegisterIP: "2.2.2.2", LastLoginIP: "9.9.9.2", FirstUsageIP: "8.8.8.8", CreatedAt: now},
			{UserID: 203, Email: "ua003@gmail.com", RegisterIP: "3.3.3.3", LastLoginIP: "9.9.9.3", FirstUsageIP: "8.8.8.8", CreatedAt: now},
		},
	}

	result := scoreAffiliateRiskCluster(cluster, now.Add(-12*time.Hour), now)
	require.Equal(t, "P1", result.Severity)
	require.Contains(t, result.Description, "API使用IP")
}

func TestAffiliateRiskScanIntervalFallback(t *testing.T) {
	t.Parallel()

	settings := NewSettingService(affiliateRiskSettingRepoStub{
		SettingKeyAffiliateRiskScanIntervalMinutes: "1",
	}, nil)
	svc := NewAffiliateRiskScannerService(nil, nil, nil, settings, nil, nil, nil)
	require.Equal(t, time.Duration(AffiliateRiskScanIntervalDefaultMin)*time.Minute, svc.getInterval())

	settings = NewSettingService(affiliateRiskSettingRepoStub{
		SettingKeyAffiliateRiskScanIntervalMinutes: "30",
	}, nil)
	svc = NewAffiliateRiskScannerService(nil, nil, nil, settings, nil, nil, nil)
	require.Equal(t, 30*time.Minute, svc.getInterval())
}

func TestSettingAffiliateRiskScanInterval(t *testing.T) {
	t.Parallel()

	repo := affiliateRiskSettingRepoStub{
		SettingKeyAffiliateRiskScanIntervalMinutes: "1441",
	}
	settings := NewSettingService(repo, nil)
	require.Equal(t, AffiliateRiskScanIntervalDefaultMin, settings.GetAffiliateRiskScanIntervalMinutes(context.Background()))

	repo[SettingKeyAffiliateRiskScanIntervalMinutes] = "5"
	require.Equal(t, 5, settings.GetAffiliateRiskScanIntervalMinutes(context.Background()))
}

func TestAffiliateRiskScannerCreatesAlertAndFreezeOnce(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Add(-time.Hour)
	rewardAt := now.Add(10 * time.Minute)
	riskRepo := &affiliateRiskScannerRepoStub{
		clusters: []AffiliateRiskCluster{
			{
				InviterID:          100,
				InviterEmail:       "3388637010@qq.com",
				InviterLastLoginIP: "2409:8962:e1:391d::1",
				Invitees: []AffiliateRiskInvitee{
					{UserID: 201, Email: "risk001@gmail.com", RegisterIP: "1.1.1.1", LastLoginIP: "2409:8962:e1:391d::2", CreatedAt: now, APICallRewardAt: &rewardAt},
					{UserID: 202, Email: "risk002@gmail.com", RegisterIP: "2.2.2.2", LastLoginIP: "2409:8962:e1:391d::3", CreatedAt: now.Add(time.Minute), APICallRewardAt: &rewardAt},
					{UserID: 203, Email: "risk003@gmail.com", RegisterIP: "3.3.3.3", LastLoginIP: "2409:8962:e1:391d::4", CreatedAt: now.Add(2 * time.Minute), APICallRewardAt: &rewardAt},
				},
			},
		},
	}
	activeAlerts := map[string]*OpsAlertEvent{}
	var createdAlerts []*OpsAlertEvent
	var heartbeats []*OpsUpsertJobHeartbeatInput
	opsRepo := &opsRepoMock{
		GetActiveAlertEventByDimFn: func(_ context.Context, kind string, inviterID int64, fingerprint string) (*OpsAlertEvent, error) {
			return activeAlerts[kind+"|"+fingerprint], nil
		},
		CreateAlertEventFn: func(_ context.Context, event *OpsAlertEvent) (*OpsAlertEvent, error) {
			event.ID = int64(len(createdAlerts) + 1)
			createdAlerts = append(createdAlerts, event)
			key := event.Dimensions["kind"].(string) + "|" + event.Dimensions["fingerprint"].(string)
			activeAlerts[key] = event
			return event, nil
		},
		UpsertJobHeartbeatFn: func(_ context.Context, input *OpsUpsertJobHeartbeatInput) error {
			heartbeats = append(heartbeats, input)
			return nil
		},
	}
	opsService := NewOpsService(opsRepo, affiliateRiskSettingRepoStub{}, &config.Config{Ops: config.OpsConfig{Enabled: true}}, nil, nil, nil, nil, nil, nil, nil, nil)
	scanner := NewAffiliateRiskScannerService(riskRepo, opsRepo, opsService, nil, nil, nil, &config.Config{Ops: config.OpsConfig{Enabled: true}})

	scanner.scanOnce(20 * time.Minute)
	scanner.scanOnce(20 * time.Minute)

	require.Len(t, createdAlerts, 1)
	require.Equal(t, "P1", createdAlerts[0].Severity)
	require.Contains(t, createdAlerts[0].Title, "疑似刷邀请奖励")
	require.Len(t, riskRepo.freezes, 2)
	require.Equal(t, int64(100), riskRepo.freezes[0].InviterID)
	require.GreaterOrEqual(t, riskRepo.freezes[0].Score, 90)
	require.Len(t, heartbeats, 2)
	require.NotNil(t, heartbeats[0].LastSuccessAt)
	require.NotNil(t, heartbeats[0].LastResult)
	require.Contains(t, *heartbeats[0].LastResult, "alerts=1")
	require.Contains(t, *heartbeats[1].LastResult, "alerts=0")
}

func TestAffiliateRiskScannerSkipsWhenOpsDisabled(t *testing.T) {
	t.Parallel()

	riskRepo := &affiliateRiskScannerRepoStub{clusters: []AffiliateRiskCluster{{InviterID: 100}}}
	opsRepo := &opsRepoMock{}
	scanner := NewAffiliateRiskScannerService(riskRepo, opsRepo, nil, nil, nil, nil, &config.Config{Ops: config.OpsConfig{Enabled: false}})

	scanner.scanOnce(20 * time.Minute)

	require.Zero(t, riskRepo.listCalls)
	require.Empty(t, riskRepo.freezes)
}

type affiliateRiskScannerRepoStub struct {
	clusters  []AffiliateRiskCluster
	freezes   []AffiliateRiskFreeze
	listCalls int
}

func (r *affiliateRiskScannerRepoStub) ListAffiliateRiskClusters(context.Context, time.Time, time.Time) ([]AffiliateRiskCluster, error) {
	r.listCalls++
	return r.clusters, nil
}

func (r *affiliateRiskScannerRepoStub) UpsertAffiliateRiskFreeze(_ context.Context, freeze AffiliateRiskFreeze) (bool, error) {
	r.freezes = append(r.freezes, freeze)
	return true, nil
}

func (r *affiliateRiskScannerRepoStub) HasActiveRiskFreeze(context.Context, int64) (bool, error) {
	return false, nil
}
