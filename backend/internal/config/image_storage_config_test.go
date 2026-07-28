package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadImageStorageConfigFromEnv(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("IMAGE_STORAGE_ENABLED", "true")
	t.Setenv("IMAGE_STORAGE_ENDPOINT", "https://s3.example.com")
	t.Setenv("IMAGE_STORAGE_REGION", "us-east-1")
	t.Setenv("IMAGE_STORAGE_BUCKET", "generated-images")
	t.Setenv("IMAGE_STORAGE_ACCESS_KEY_ID", "access-key")
	t.Setenv("IMAGE_STORAGE_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("IMAGE_STORAGE_PREFIX", "async-images/")
	t.Setenv("IMAGE_STORAGE_FORCE_PATH_STYLE", "true")
	t.Setenv("IMAGE_STORAGE_PUBLIC_BASE_URL", "https://cdn.example.com")
	t.Setenv("IMAGE_STORAGE_PRESIGN_EXPIRY_HOURS", "12")
	t.Setenv("IMAGE_STORAGE_MAX_DOWNLOAD_BYTES", "1048576")

	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.ImageStorage.Enabled)
	require.Equal(t, "https://s3.example.com", cfg.ImageStorage.Endpoint)
	require.Equal(t, "us-east-1", cfg.ImageStorage.Region)
	require.Equal(t, "generated-images", cfg.ImageStorage.Bucket)
	require.Equal(t, "access-key", cfg.ImageStorage.AccessKeyID)
	require.Equal(t, "secret-key", cfg.ImageStorage.SecretAccessKey)
	require.Equal(t, "async-images/", cfg.ImageStorage.Prefix)
	require.True(t, cfg.ImageStorage.ForcePathStyle)
	require.Equal(t, "https://cdn.example.com", cfg.ImageStorage.PublicBaseURL)
	require.Equal(t, 12, cfg.ImageStorage.PresignExpiry)
	require.Equal(t, int64(1048576), cfg.ImageStorage.MaxDownloadByte)
	require.True(t, cfg.ImageStorage.Active())
}
