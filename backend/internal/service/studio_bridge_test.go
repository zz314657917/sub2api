package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type studioBridgeSettingRepoStub struct {
	values map[string]string
}

func (r *studioBridgeSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *studioBridgeSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (r *studioBridgeSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (r *studioBridgeSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *studioBridgeSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (r *studioBridgeSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *studioBridgeSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type studioBridgeRepoStub struct {
	mu             sync.Mutex
	balance        float64
	reserveCalls   int
	refundCalls    int
	usageLogCalls  int
	usageLogAmount float64
	usageLogCmd    StudioBridgeChargeCommand
	charges        map[string]studioBridgeRepoCharge
	refunds        map[string]studioBridgeRepoCharge
}

func (r *studioBridgeRepoStub) GetUserSummary(context.Context, int64, string, int) (*StudioBridgeUserSummary, error) {
	return nil, ErrUserNotFound
}

type studioBridgeRepoCharge struct {
	cmd          StudioBridgeChargeCommand
	status       string
	refunded     float64
	balanceAfter float64
}

func (r *studioBridgeRepoStub) ReserveStudioBridgeCharge(_ context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	if existing, ok := r.charges[cmd.ChargeKey]; ok {
		if existing.cmd.Fingerprint() != cmd.Fingerprint() {
			return nil, ErrStudioBridgeConflict
		}
		return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: existing.status, Applied: false, Amount: existing.cmd.Amount, BalanceAfter: existing.balanceAfter}, nil
	}
	if r.balance < cmd.Amount {
		return nil, ErrStudioBridgeInsufficient
	}
	r.balance -= cmd.Amount
	r.reserveCalls++
	r.charges[cmd.ChargeKey] = studioBridgeRepoCharge{cmd: cmd, status: "reserved", balanceAfter: r.balance}
	return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: "reserved", Applied: true, Amount: cmd.Amount, BalanceAfter: r.balance}, nil
}

func (r *studioBridgeRepoStub) CommitStudioBridgeCharge(_ context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	charge, ok := r.charges[cmd.ChargeKey]
	if !ok {
		return nil, ErrStudioBridgeChargeKeyEmpty
	}
	if charge.cmd.Fingerprint() != cmd.Fingerprint() {
		return nil, ErrStudioBridgeConflict
	}
	if charge.status == "committed" || charge.status == "refunded" {
		return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: charge.status, Applied: false, Amount: charge.cmd.Amount, BalanceAfter: charge.balanceAfter}, nil
	}
	usageAmount := charge.cmd.Amount - charge.refunded
	if usageAmount > 0 {
		r.usageLogCalls++
		r.usageLogAmount = usageAmount
		r.usageLogCmd = charge.cmd
	}
	charge.status = "committed"
	r.charges[cmd.ChargeKey] = charge
	return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: "committed", Applied: true, Amount: charge.cmd.Amount, BalanceAfter: charge.balanceAfter}, nil
}

func (r *studioBridgeRepoStub) RefundStudioBridgeCharge(_ context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	if cmd.RefundForChargeKey != "" && cmd.RefundForChargeKey != cmd.ChargeKey {
		if existing, ok := r.refunds[cmd.ChargeKey]; ok {
			if existing.cmd.Fingerprint() != cmd.Fingerprint() {
				return nil, ErrStudioBridgeConflict
			}
			return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: existing.status, Applied: false, Amount: existing.cmd.Amount, BalanceAfter: existing.balanceAfter}, nil
		}
		original, ok := r.charges[cmd.RefundForChargeKey]
		if !ok {
			return nil, ErrStudioBridgeChargeKeyEmpty
		}
		if original.cmd.UserID != cmd.UserID || original.status != "reserved" || original.refunded+cmd.Amount > original.cmd.Amount+1e-9 {
			return nil, ErrStudioBridgeConflict
		}
		r.balance += cmd.Amount
		r.refundCalls++
		original.refunded += cmd.Amount
		if original.refunded+1e-9 >= original.cmd.Amount {
			original.status = "refunded"
		}
		original.balanceAfter = r.balance
		r.charges[cmd.RefundForChargeKey] = original
		r.refunds[cmd.ChargeKey] = studioBridgeRepoCharge{cmd: cmd, status: "refunded", balanceAfter: r.balance}
		return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: "refunded", Applied: true, Amount: cmd.Amount, BalanceAfter: r.balance}, nil
	}
	charge, ok := r.charges[cmd.ChargeKey]
	if !ok {
		return nil, ErrStudioBridgeChargeKeyEmpty
	}
	if charge.cmd.Fingerprint() != cmd.Fingerprint() {
		return nil, ErrStudioBridgeConflict
	}
	if charge.status == "refunded" {
		return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: "refunded", Applied: false, Amount: charge.cmd.Amount, BalanceAfter: charge.balanceAfter}, nil
	}
	if charge.status != "reserved" {
		return nil, ErrStudioBridgeConflict
	}
	refundAmount := charge.cmd.Amount - charge.refunded
	r.balance += refundAmount
	r.refundCalls++
	charge.refunded = charge.cmd.Amount
	charge.status = "refunded"
	charge.balanceAfter = r.balance
	r.charges[cmd.ChargeKey] = charge
	return &StudioBridgeChargeResult{ChargeKey: cmd.ChargeKey, Status: "refunded", Applied: true, Amount: refundAmount, BalanceAfter: r.balance}, nil
}

func (r *studioBridgeRepoStub) ensure() {
	if r.charges == nil {
		r.charges = map[string]studioBridgeRepoCharge{}
	}
	if r.refunds == nil {
		r.refunds = map[string]studioBridgeRepoCharge{}
	}
}

type studioBridgeMemoryStore struct {
	mu    sync.Mutex
	items map[string][]byte
}

func newStudioBridgeMemoryStore() *studioBridgeMemoryStore {
	return &studioBridgeMemoryStore{items: map[string][]byte{}}
}

func (s *studioBridgeMemoryStore) Set(_ context.Context, key string, payload []byte, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = append([]byte(nil), payload...)
	return nil
}

func (s *studioBridgeMemoryStore) GetDel(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, ok := s.items[key]
	if !ok {
		return nil, false, nil
	}
	delete(s.items, key)
	return append([]byte(nil), payload...), true, nil
}

func newStudioBridgeTestService(t *testing.T, repo *studioBridgeRepoStub) *StudioBridgeService {
	t.Helper()
	raw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:           true,
		SiteName:          "落叶创艺",
		LaunchReturnURL:   "http://127.0.0.1:8081/auth/sub2api/launch",
		RechargeReturnURL: "http://127.0.0.1:62080/purchase",
		DefaultChatGroup:  "1",
		DefaultImageGroup: "2",
		DefaultVideoGroup: "3",
		InternalSecret:    "secret",
	})
	require.NoError(t, err)
	settings := NewSettingService(&studioBridgeSettingRepoStub{values: map[string]string{SettingKeyStudioBridgeLuoyeAI: raw}}, &config.Config{})
	return NewStudioBridgeService(settings, repo, newStudioBridgeMemoryStore())
}

func TestStudioBridgeRefundRestoresReservedBalanceOnce(t *testing.T) {
	ctx := context.Background()
	repo := &studioBridgeRepoStub{balance: 10}
	svc := newStudioBridgeTestService(t, repo)

	reserved, err := svc.Reserve(ctx, StudioBridgeChargeCommand{
		AppID:     StudioBridgeAppLuoyeAI,
		UserID:    42,
		ChargeKey: "task:42:image:precharge",
		Amount:    0.05,
	}, "secret")
	require.NoError(t, err)
	require.True(t, reserved.Applied)
	require.InDelta(t, 9.95, reserved.BalanceAfter, 0.000001)

	refunded, err := svc.Refund(ctx, StudioBridgeChargeCommand{
		AppID:              StudioBridgeAppLuoyeAI,
		UserID:             42,
		ChargeKey:          "task:42:image:refund",
		RefundForChargeKey: "task:42:image:precharge",
		Amount:             0.05,
	}, "secret")
	require.NoError(t, err)
	require.True(t, refunded.Applied)
	require.InDelta(t, 10, refunded.BalanceAfter, 0.000001)

	duplicate, err := svc.Refund(ctx, StudioBridgeChargeCommand{
		AppID:              StudioBridgeAppLuoyeAI,
		UserID:             42,
		ChargeKey:          "task:42:image:refund",
		RefundForChargeKey: "task:42:image:precharge",
		Amount:             0.05,
	}, "secret")
	require.NoError(t, err)
	require.False(t, duplicate.Applied)
	require.Equal(t, 1, repo.reserveCalls)
	require.Equal(t, 1, repo.refundCalls)
	require.InDelta(t, 10, repo.balance, 0.000001)
}

func TestStudioBridgeConcurrentReserveIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := &studioBridgeRepoStub{balance: 10}
	svc := newStudioBridgeTestService(t, repo)
	cmd := StudioBridgeChargeCommand{
		AppID:     StudioBridgeAppLuoyeAI,
		UserID:    42,
		ChargeKey: "task:42:image:precharge",
		Amount:    0.05,
	}

	var wg sync.WaitGroup
	results := make(chan *StudioBridgeChargeResult, 12)
	errors := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := svc.Reserve(ctx, cmd, "secret")
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}
	wg.Wait()
	close(results)
	close(errors)

	require.Empty(t, errors)
	applied := 0
	for result := range results {
		if result.Applied {
			applied++
		}
	}
	require.Equal(t, 1, applied)
	require.Equal(t, 1, repo.reserveCalls)
	require.InDelta(t, 9.95, repo.balance, 0.000001)
}

func TestStudioBridgeCommitCreatesUsageLogOnceWithNetAmount(t *testing.T) {
	ctx := context.Background()
	repo := &studioBridgeRepoStub{balance: 10}
	svc := newStudioBridgeTestService(t, repo)
	reserveCmd := StudioBridgeChargeCommand{
		AppID:       StudioBridgeAppLuoyeAI,
		UserID:      42,
		ChargeKey:   "task:42:image:precharge",
		Amount:      0.10,
		TaskID:      "task-1",
		Mode:        "generate",
		Model:       "gpt-image-2",
		ActorUserID: "sub2api:99",
		TeamID:      "team-1",
	}

	reserved, err := svc.Reserve(ctx, reserveCmd, "secret")
	require.NoError(t, err)
	require.True(t, reserved.Applied)

	refunded, err := svc.Refund(ctx, StudioBridgeChargeCommand{
		AppID:              StudioBridgeAppLuoyeAI,
		UserID:             42,
		ChargeKey:          "task:42:image:refund",
		RefundForChargeKey: reserveCmd.ChargeKey,
		Amount:             0.04,
	}, "secret")
	require.NoError(t, err)
	require.True(t, refunded.Applied)

	committed, err := svc.Commit(ctx, reserveCmd, "secret")
	require.NoError(t, err)
	require.True(t, committed.Applied)
	require.Equal(t, 1, repo.usageLogCalls)
	require.InDelta(t, 0.06, repo.usageLogAmount, 0.000001)
	require.Equal(t, "task-1", repo.usageLogCmd.TaskID)
	require.Equal(t, "generate", repo.usageLogCmd.Mode)
	require.Equal(t, "gpt-image-2", repo.usageLogCmd.Model)
	require.Equal(t, "sub2api:99", repo.usageLogCmd.ActorUserID)
	require.Equal(t, "team-1", repo.usageLogCmd.TeamID)

	duplicate, err := svc.Commit(ctx, reserveCmd, "secret")
	require.NoError(t, err)
	require.False(t, duplicate.Applied)
	require.Equal(t, 1, repo.usageLogCalls)
}

func TestStudioBridgeSessionProbeOriginValidation(t *testing.T) {
	ctx := context.Background()
	svc := newStudioBridgeTestService(t, &studioBridgeRepoStub{})

	require.NoError(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "http://127.0.0.1:8081"))
	require.NoError(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "http://127.0.0.1:8081/profile"))
	require.ErrorIs(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "http://evil.example.com"), ErrStudioBridgeInvalidReturn)
	require.ErrorIs(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "javascript:alert(1)"), ErrStudioBridgeInvalidReturn)
}

func TestStudioBridgeSessionProbeOriginAllowsConfiguredReturnDomain(t *testing.T) {
	ctx := context.Background()
	raw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:              true,
		SiteName:             "落叶创艺",
		AllowedReturnDomains: []string{"example.com"},
		LaunchReturnURL:      "http://127.0.0.1:8081/auth/sub2api/launch",
		RechargeReturnURL:    "http://127.0.0.1:62080/purchase",
		DefaultChatGroup:     "1",
		DefaultImageGroup:    "2",
		DefaultVideoGroup:    "3",
		InternalSecret:       "secret",
	})
	require.NoError(t, err)
	settings := NewSettingService(&studioBridgeSettingRepoStub{values: map[string]string{SettingKeyStudioBridgeLuoyeAI: raw}}, &config.Config{})
	svc := NewStudioBridgeService(settings, &studioBridgeRepoStub{}, newStudioBridgeMemoryStore())

	require.NoError(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "https://luoye.example.com"))
	require.ErrorIs(t, svc.ValidateSessionProbeOrigin(ctx, StudioBridgeAppLuoyeAI, "https://example.org"), ErrStudioBridgeInvalidReturn)
}

func TestStudioBridgeEnabledSettingsRequireChatAndImageGroups(t *testing.T) {
	err := validateStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:           true,
		SiteName:          "落叶创艺",
		LaunchReturnURL:   "http://127.0.0.1:8081/auth/sub2api/launch",
		RechargeReturnURL: "http://127.0.0.1:62080/purchase",
		InternalSecret:    "secret",
	})
	require.ErrorIs(t, err, ErrStudioBridgeGroupRequired)

	err = validateStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:           true,
		DefaultChatGroup:  "1",
		DefaultImageGroup: "2",
	})
	require.NoError(t, err)

	err = validateStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:           true,
		DefaultChatGroup:  "1",
		DefaultVideoGroup: "3",
	})
	require.ErrorIs(t, err, ErrStudioBridgeGroupRequired)
}
