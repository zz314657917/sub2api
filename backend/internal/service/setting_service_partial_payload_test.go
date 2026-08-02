package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type partialPayloadSettingRepoStub struct {
	values      map[string]string
	updates     map[string]string
	getMultiple func(context.Context, []string) (map[string]string, error)
}

func (s *partialPayloadSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *partialPayloadSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	return s.values[key], nil
}

func (s *partialPayloadSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *partialPayloadSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	if s.getMultiple != nil {
		return s.getMultiple(ctx, keys)
	}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *partialPayloadSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.updates[key] = value
		s.values[key] = value
	}
	return nil
}

func (s *partialPayloadSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *partialPayloadSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingServiceUpdateSettingsOmittingPreservesUnsentValue(t *testing.T) {
	repo := &partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeySiteName: "Example Gateway",
	}}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettingsOmitting(context.Background(), &SystemSettings{}, OmittedSettingKeys{
		SettingKeySiteName: {},
	})

	require.NoError(t, err)
	require.Equal(t, "Example Gateway", repo.values[SettingKeySiteName])
	_, wroteSiteName := repo.updates[SettingKeySiteName]
	require.False(t, wroteSiteName)
}

func TestSettingServiceUpdateSettingsOmittingExplicitEmptyValueClears(t *testing.T) {
	repo := &partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeySiteName: "Example Gateway",
	}}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.UpdateSettingsOmitting(context.Background(), &SystemSettings{SiteName: ""}, nil)

	require.NoError(t, err)
	require.Equal(t, "", repo.values[SettingKeySiteName])
	require.Equal(t, "", repo.updates[SettingKeySiteName])
}

func TestSettingServiceUpdateSettingsPersistsForwardedClientIPSettings(t *testing.T) {
	repo := &partialPayloadSettingRepoStub{values: map[string]string{}}
	cfg := &config.Config{}
	svc := NewSettingService(repo, cfg)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		APIKeyACLTrustForwardedIP: true,
		ForwardedClientIPHeaders:  []string{" x-cdn-ip ", "X-CDN-IP", "true-client-ip"},
	})

	require.NoError(t, err)
	require.Equal(t, "true", repo.updates[SettingKeyAPIKeyACLTrustForwardedIP])
	require.JSONEq(t, `["X-Cdn-Ip","True-Client-Ip"]`, repo.updates[SettingKeyForwardedClientIPHeaders])
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.True(t, runtimeSettings.TrustForwardedIP)
	require.Equal(t, []string{"X-Cdn-Ip", "True-Client-Ip"}, runtimeSettings.Headers)
}

func TestSettingServiceUpdateSettingsRejectsInvalidForwardedClientIPHeaders(t *testing.T) {
	repo := &partialPayloadSettingRepoStub{values: map[string]string{}}
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(true, []string{"X-Existing-IP"})
	svc := NewSettingService(repo, cfg)

	err := svc.UpdateSettings(context.Background(), &SystemSettings{
		APIKeyACLTrustForwardedIP: true,
		ForwardedClientIPHeaders:  []string{"X Invalid"},
	})

	require.Error(t, err)
	require.Nil(t, repo.updates)
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.True(t, runtimeSettings.TrustForwardedIP)
	require.Equal(t, []string{"X-Existing-Ip"}, runtimeSettings.Headers)
}

func TestSettingServiceUpdateSettingsOmittingForwardedClientIPSettingsPreservesStoredValues(t *testing.T) {
	repo := &partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeyAPIKeyACLTrustForwardedIP: "true",
		SettingKeyForwardedClientIPHeaders:  `["X-Existing-IP"]`,
	}}
	cfg := &config.Config{}
	svc := NewSettingService(repo, cfg)

	err := svc.UpdateSettingsOmitting(context.Background(), &SystemSettings{}, OmittedSettingKeys{
		SettingKeyAPIKeyACLTrustForwardedIP: {},
		SettingKeyForwardedClientIPHeaders:  {},
	})

	require.NoError(t, err)
	require.Equal(t, "true", repo.values[SettingKeyAPIKeyACLTrustForwardedIP])
	require.Equal(t, `["X-Existing-IP"]`, repo.values[SettingKeyForwardedClientIPHeaders])
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.True(t, runtimeSettings.TrustForwardedIP)
	require.Equal(t, []string{"X-Existing-Ip"}, runtimeSettings.Headers)
}

func TestLoadForwardedClientIPSettingsUsesValidPersistedValues(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(false, []string{"X-Config-IP"})
	repo := &partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeyAPIKeyACLTrustForwardedIP: "true",
		SettingKeyForwardedClientIPHeaders:  `[" x-cdn-ip ","X-CDN-IP","true-client-ip"]`,
	}}
	svc := NewSettingService(repo, cfg)

	require.NoError(t, svc.LoadForwardedClientIPSettings(context.Background()))
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.True(t, runtimeSettings.TrustForwardedIP)
	require.Equal(t, []string{"X-Cdn-Ip", "True-Client-Ip"}, runtimeSettings.Headers)
}

func TestLoadForwardedClientIPSettingsFallsBackToConfigWhenPersistedValuesAreMissing(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(true, []string{"X-Config-IP"})
	svc := NewSettingService(&partialPayloadSettingRepoStub{values: map[string]string{}}, cfg)

	require.NoError(t, svc.LoadForwardedClientIPSettings(context.Background()))
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.True(t, runtimeSettings.TrustForwardedIP)
	require.Equal(t, []string{"X-Config-Ip"}, runtimeSettings.Headers)
}

func TestLoadForwardedClientIPSettingsFailsClosedOnMalformedPersistedHeaders(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetForwardedClientIPSettings(true, []string{"X-Config-IP"})
	svc := NewSettingService(&partialPayloadSettingRepoStub{values: map[string]string{
		SettingKeyAPIKeyACLTrustForwardedIP: "true",
		SettingKeyForwardedClientIPHeaders:  `["X Invalid"]`,
	}}, cfg)

	err := svc.LoadForwardedClientIPSettings(context.Background())
	require.Error(t, err)
	runtimeSettings := cfg.ForwardedClientIPSettings()
	require.False(t, runtimeSettings.TrustForwardedIP)
	require.Empty(t, runtimeSettings.Headers)
}
