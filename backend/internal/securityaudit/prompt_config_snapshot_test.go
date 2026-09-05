package securityaudit

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type delayedSnapshotSettingsRepository struct {
	staticSettingRepository
	readStarted chan struct{}
	release     chan struct{}
	once        sync.Once
}

func (r *delayedSnapshotSettingsRepository) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	values, err := r.staticSettingRepository.GetMultiple(ctx, keys)
	r.once.Do(func() { close(r.readStarted) })
	select {
	case <-r.release:
		return values, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func snapshotConfig(version int64, riskControlEnabled, enabled, blocking bool) (storageConfig, ActiveConfig) {
	storage := DefaultStorageConfig()
	storage.ConfigVersion = version
	storage.Enabled = enabled
	storage.BlockingEnabled = blocking
	active := ActiveConfig{
		ConfigVersion: version, RiskControlEnabled: riskControlEnabled,
		Enabled: enabled, BlockingEnabled: blocking,
	}
	return storage, active
}

func TestConfigManagerSnapshotInstallationRejectsOutOfOrderVersions(t *testing.T) {
	manager := &ConfigManager{clock: fixedClock{now: time.Unix(1, 0)}}
	highStorage, highActive := snapshotConfig(8, true, true, true)
	_, installed := manager.installConfigSnapshot(highStorage, highActive)
	require.True(t, installed)

	lowStorage, lowActive := snapshotConfig(7, true, true, false)
	_, installed = manager.installConfigSnapshot(lowStorage, lowActive)
	require.False(t, installed)

	active, ok := manager.Active()
	require.True(t, ok)
	require.Equal(t, int64(8), active.ConfigVersion)
	require.Equal(t, ModeBlocking, active.EffectiveMode())
	require.Equal(t, int64(8), manager.expected.Load())
	require.True(t, manager.expectedBlocking.Load())
}

func TestConfigManagerSnapshotInstallationRefreshesSameVersionRiskGate(t *testing.T) {
	manager := &ConfigManager{clock: fixedClock{now: time.Unix(1, 0)}}
	storage, enabled := snapshotConfig(8, true, true, true)
	_, installed := manager.installConfigSnapshot(storage, enabled)
	require.True(t, installed)

	_, disabled := snapshotConfig(8, false, true, true)
	_, installed = manager.installConfigSnapshot(storage, disabled)
	require.True(t, installed)

	active, ok := manager.Active()
	require.True(t, ok)
	require.False(t, active.RiskControlEnabled)
	require.False(t, manager.expectedBlocking.Load())
	require.Equal(t, ModeOff, manager.EffectiveMode())
}

func TestConfigManagerObserveExpectedStateDoesNotRegressVersionOrBlockingIntent(t *testing.T) {
	manager := &ConfigManager{}
	manager.observeExpectedState(`{"enabled":true,"blocking_enabled":true,"config_version":8}`, true)
	manager.observeExpectedState(`{"enabled":true,"blocking_enabled":false,"config_version":7}`, true)

	require.Equal(t, int64(8), manager.expected.Load())
	require.True(t, manager.expectedBlocking.Load())

	manager.observeExpectedState(`{"enabled":true`, true)
	require.Equal(t, int64(8), manager.expected.Load())
	require.True(t, manager.expectedBlocking.Load())
}

func TestConfigManagerReloadDoesNotInstallOlderSnapshotAfterNewerWrite(t *testing.T) {
	oldStorage, _ := snapshotConfig(7, true, false, false)
	raw, err := json.Marshal(oldStorage)
	require.NoError(t, err)
	repository := &delayedSnapshotSettingsRepository{
		staticSettingRepository: staticSettingRepository{values: map[string]string{
			SettingKeyPromptAuditConfig: string(raw),
			SettingKeyRiskControl:       "true",
		}},
		readStarted: make(chan struct{}),
		release:     make(chan struct{}),
	}
	manager := NewConfigManager(nil, repository, nil, prefixEncryptor{}, testTotpKeyConfig())
	manager.clock = fixedClock{now: time.Unix(1, 0)}

	reloadDone := make(chan error, 1)
	go func() { reloadDone <- manager.Reload(context.Background()) }()
	<-repository.readStarted

	newStorage, newActive := snapshotConfig(8, true, true, true)
	_, installed := manager.installConfigSnapshot(newStorage, newActive)
	require.True(t, installed)
	close(repository.release)
	require.NoError(t, <-reloadDone)

	active, ok := manager.Active()
	require.True(t, ok)
	require.Equal(t, int64(8), active.ConfigVersion)
	require.Equal(t, int64(8), manager.expected.Load())
	require.True(t, manager.expectedBlocking.Load())
}

func TestConfigManagerObservedInvalidationVersionRejectsOlderInstallation(t *testing.T) {
	manager := &ConfigManager{clock: fixedClock{now: time.Unix(1, 0)}}
	manager.observeExpectedVersion(8)

	storage, active := snapshotConfig(7, true, true, true)
	_, installed := manager.installConfigSnapshot(storage, active)
	require.False(t, installed)
	require.Nil(t, manager.snapshot.Load())
	require.Equal(t, int64(8), manager.expected.Load())
}
