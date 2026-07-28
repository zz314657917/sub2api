package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type imageStorageSettingsRepo struct {
	values map[string]string
}

func newImageStorageSettingsRepo() *imageStorageSettingsRepo {
	return &imageStorageSettingsRepo{values: make(map[string]string)}
}

func (r *imageStorageSettingsRepo) Get(_ context.Context, key string) (*Setting, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return &Setting{Key: key, Value: value}, nil
}

func (r *imageStorageSettingsRepo) GetValue(_ context.Context, key string) (string, error) {
	return r.values[key], nil
}

func (r *imageStorageSettingsRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *imageStorageSettingsRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (r *imageStorageSettingsRepo) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

func (r *imageStorageSettingsRepo) GetAll(_ context.Context) (map[string]string, error) {
	values := make(map[string]string, len(r.values))
	for key, value := range r.values {
		values[key] = value
	}
	return values, nil
}

func (r *imageStorageSettingsRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type imageStorageSettingsEncryptor struct{}

func (imageStorageSettingsEncryptor) Encrypt(value string) (string, error) {
	return "ENC:" + value, nil
}

func (imageStorageSettingsEncryptor) Decrypt(value string) (string, error) {
	if len(value) < 4 || value[:4] != "ENC:" {
		return "", errors.New("not encrypted")
	}
	return value[4:], nil
}

type imageStorageSettingsStore struct{}

func (imageStorageSettingsStore) Save(_ context.Context, _ string, _ string, _ []byte) (string, error) {
	return "https://images.example.test/result.png", nil
}

func TestImageStorageSettingsHotUpdateAndSecretRedaction(t *testing.T) {
	repo := newImageStorageSettingsRepo()
	backup := NewBackupService(repo, &config.Config{
		Totp: config.TotpConfig{EncryptionKeyConfigured: true},
	}, imageStorageSettingsEncryptor{}, nil, nil)
	var factoryCalls int
	settings := NewImageStorageSettingService(
		repo,
		imageStorageSettingsEncryptor{},
		backup,
		func(_ context.Context, cfg *config.ImageStorageConfig) (ImageStorage, error) {
			factoryCalls++
			require.Equal(t, "generated-images", cfg.Bucket)
			require.Equal(t, "secret-value", cfg.SecretAccessKey)
			return imageStorageSettingsStore{}, nil
		},
		config.ImageStorageConfig{},
	)

	updated, err := settings.Update(context.Background(), ImageStorageSettings{
		Enabled:         true,
		Bucket:          "generated-images",
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-value",
	})
	require.NoError(t, err)
	require.Empty(t, updated.SecretAccessKey)
	require.Contains(t, repo.values[settingKeyImageStorageConfig], "ENC:secret-value")

	public, err := settings.Get(context.Background())
	require.NoError(t, err)
	require.Empty(t, public.SecretAccessKey)
	require.True(t, settings.SecretConfigured(context.Background()))

	_, enabled := settings.Resolver()()
	require.True(t, enabled)
	require.Equal(t, 1, factoryCalls)

	_, err = settings.Update(context.Background(), ImageStorageSettings{Enabled: false})
	require.NoError(t, err)
	_, enabled = settings.Resolver()()
	require.False(t, enabled)
}
