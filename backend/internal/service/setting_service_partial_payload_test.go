package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type partialPayloadSettingRepoStub struct {
	values  map[string]string
	updates map[string]string
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
	panic("unexpected GetMultiple call")
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
