package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestServiceStoreDefaultsPersistsFalseAndPreservesOmission(t *testing.T) {
	ctx := context.Background()
	repo := &partialPayloadSettingRepoStub{values: map[string]string{}}
	svc := NewSettingService(repo, &config.Config{})
	updates := 0
	svc.SetOnUpdateCallback(func() { updates++ })

	settings, err := svc.GetAllSettings(ctx)
	require.NoError(t, err)
	require.True(t, settings.ServiceStoreEnabled, "missing legacy key must preserve the enabled Store")

	public, err := svc.GetPublicSettings(ctx)
	require.NoError(t, err)
	require.True(t, public.ServiceStoreEnabled)

	require.NoError(t, svc.UpdateSettings(ctx, &SystemSettings{ServiceStoreEnabled: true}))
	require.Equal(t, "true", repo.values[SettingKeyServiceStoreEnabled])
	require.Equal(t, 1, updates, "persisting Store state must invalidate cached public settings")

	settings, err = svc.GetAllSettings(ctx)
	require.NoError(t, err)
	require.True(t, settings.ServiceStoreEnabled)

	require.NoError(t, svc.UpdateSettings(ctx, &SystemSettings{ServiceStoreEnabled: false}))
	require.Equal(t, "false", repo.values[SettingKeyServiceStoreEnabled])
	require.Equal(t, 2, updates)

	settings, err = svc.GetAllSettings(ctx)
	require.NoError(t, err)
	require.False(t, settings.ServiceStoreEnabled)

	require.NoError(t, svc.UpdateSettingsOmitting(ctx, &SystemSettings{ServiceStoreEnabled: true}, OmittedSettingKeys{
		SettingKeyServiceStoreEnabled: {},
	}))
	require.Equal(t, "false", repo.values[SettingKeyServiceStoreEnabled])
	_, wrote := repo.updates[SettingKeyServiceStoreEnabled]
	require.False(t, wrote, "omitted field must not overwrite explicit false")
	require.Equal(t, 3, updates)
}

func TestServiceStorePublicInjectionParity(t *testing.T) {
	ctx := context.Background()
	svc := NewSettingService(&partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeyServiceStoreEnabled: "false",
	}}, &config.Config{})

	public, err := svc.GetPublicSettings(ctx)
	require.NoError(t, err)
	require.False(t, public.ServiceStoreEnabled)

	injected, err := svc.GetPublicSettingsForInjection(ctx)
	require.NoError(t, err)
	payload, ok := injected.(*PublicSettingsInjectionPayload)
	require.True(t, ok)
	require.False(t, payload.ServiceStoreEnabled)
}
