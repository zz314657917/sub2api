package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type quickstartTutorialSettingRepoStub struct {
	values map[string]string
}

func (s *quickstartTutorialSettingRepoStub) Get(_ context.Context, key string) (*Setting, error) {
	value, err := s.GetValue(context.Background(), key)
	if err != nil {
		return nil, err
	}
	return &Setting{Key: key, Value: value}, nil
}

func (s *quickstartTutorialSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (s *quickstartTutorialSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = make(map[string]string)
	}
	s.values[key] = value
	return nil
}

func (s *quickstartTutorialSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (s *quickstartTutorialSettingRepoStub) SetMultiple(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if err := s.Set(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *quickstartTutorialSettingRepoStub) GetAll(_ context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *quickstartTutorialSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

func TestQuickstartTutorialConfig_DefaultAndPersistence(t *testing.T) {
	repo := &quickstartTutorialSettingRepoStub{}
	service := NewSettingService(repo, &config.Config{})

	defaults, err := service.GetQuickstartTutorialConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "https://ai.3zapi.top", defaults.Platforms[0].BaseURL)

	updated := *defaults
	updated.Platforms = append([]QuickstartTutorialPlatform(nil), defaults.Platforms...)
	updated.Platforms[0].BaseURL = "https://ai.3zapi.com"
	saved, err := service.SetQuickstartTutorialConfig(context.Background(), &updated)
	require.NoError(t, err)
	require.Equal(t, "https://ai.3zapi.com", saved.Platforms[0].BaseURL)
	require.Contains(t, repo.values[SettingKeyQuickstartTutorialConfig], "ai.3zapi.com")

	reloaded, err := service.GetQuickstartTutorialConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "https://ai.3zapi.com", reloaded.Platforms[0].BaseURL)
}

func TestQuickstartTutorialConfig_UsesConfiguredAPIBaseURLBeforeFirstSave(t *testing.T) {
	repo := &quickstartTutorialSettingRepoStub{values: map[string]string{
		SettingKeyAPIBaseURL: "https://ai.3zapi.com/v1",
	}}
	service := NewSettingService(repo, &config.Config{})

	cfg, err := service.GetQuickstartTutorialConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "https://ai.3zapi.com", cfg.Platforms[0].BaseURL)
	require.Equal(t, "https://ai.3zapi.com", cfg.Platforms[1].BaseURL)
}

func TestQuickstartTutorialConfig_RejectsUnsafeTextAndInvalidURL(t *testing.T) {
	configWithHTML := DefaultQuickstartTutorialConfig()
	configWithHTML.Header.Title = "<script>alert(1)</script>"
	_, err := NormalizeQuickstartTutorialConfig(configWithHTML)
	require.Error(t, err)
	require.Contains(t, err.Error(), "plain text")

	configWithInvalidURL := DefaultQuickstartTutorialConfig()
	configWithInvalidURL.Platforms = append([]QuickstartTutorialPlatform(nil), configWithInvalidURL.Platforms...)
	configWithInvalidURL.Platforms[0].BaseURL = "javascript:alert(1)"
	_, err = NormalizeQuickstartTutorialConfig(configWithInvalidURL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "http(s) URL")
}

func TestQuickstartTutorialConfig_RejectsMalformedStoredJSON(t *testing.T) {
	_, err := ParseQuickstartTutorialConfig(`{"platforms":`)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JSON is invalid")
}
