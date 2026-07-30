package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

const settingKeyImageStorageConfig = "image_storage_config"

var ErrImageStorageIncomplete = errors.New("image storage is enabled but bucket/access_key_id/secret_access_key are incomplete")

// ImageStorageFactory is implemented in repository so service keeps no
// dependency on a particular object-storage provider.
type ImageStorageFactory func(ctx context.Context, cfg *config.ImageStorageConfig) (ImageStorage, error)

// ImageStorageSettings is the administrative, hot-reloadable configuration
// for gateway async image task results. It is separate from image_creator.
type ImageStorageSettings struct {
	Enabled       bool `json:"enabled"`
	ReuseBackupS3 bool `json:"reuse_backup_s3"`

	Bucket           string `json:"bucket"`
	Prefix           string `json:"prefix"`
	PublicBaseURL    string `json:"public_base_url"`
	PresignExpiry    int    `json:"presign_expiry_hours"`
	MaxDownloadBytes int64  `json:"max_download_bytes"`

	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	ForcePathStyle  bool   `json:"force_path_style"`
}

type ImageStorageSettingService struct {
	settingRepo SettingRepository
	encryptor   SecretEncryptor
	backup      *BackupService
	factory     ImageStorageFactory
	fallback    config.ImageStorageConfig

	mu       sync.Mutex
	resolved bool
	uploader *ImageResultUploader
	enabled  bool
}

func NewImageStorageSettingService(
	settingRepo SettingRepository,
	encryptor SecretEncryptor,
	backup *BackupService,
	factory ImageStorageFactory,
	fallback config.ImageStorageConfig,
) *ImageStorageSettingService {
	return &ImageStorageSettingService{
		settingRepo: settingRepo,
		encryptor:   encryptor,
		backup:      backup,
		factory:     factory,
		fallback:    fallback,
	}
}

func (s *ImageStorageSettingService) Resolver() ImageStorageResolver {
	return func() (*ImageResultUploader, bool) {
		return s.resolve()
	}
}

func (s *ImageStorageSettingService) resolve() (*ImageResultUploader, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.resolved {
		return s.uploader, s.enabled
	}
	s.resolved = true
	s.uploader, s.enabled = nil, false

	cfg, err := s.effectiveConfig(context.Background())
	if err != nil {
		logger.L().Warn("image_storage.settings_load_failed; async image tasks remain disabled", zap.Error(err))
		return nil, false
	}
	if !cfg.Enabled {
		return nil, false
	}
	if !cfg.IsConfigured() {
		logger.L().Warn("image_storage is enabled but incomplete; async image tasks remain disabled", zap.Strings("missing_keys", cfg.MissingCredentialKeys()))
		return nil, false
	}
	if s.factory == nil {
		logger.L().Error("image_storage.client_factory_unavailable; async image tasks remain disabled")
		return nil, false
	}
	storage, err := s.factory(context.Background(), cfg)
	if err != nil {
		logger.L().Error("image_storage.client_build_failed; async image tasks remain disabled", zap.Error(err))
		return nil, false
	}
	s.uploader = NewImageResultUploader(storage, cfg.Prefix, cfg.MaxDownloadByte, nil)
	s.enabled = true
	return s.uploader, true
}

func (s *ImageStorageSettingService) Invalidate() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.resolved = false
	s.uploader = nil
	s.enabled = false
	s.mu.Unlock()
}

// Get never returns a configured secret to an admin client.
func (s *ImageStorageSettingService) Get(ctx context.Context) (*ImageStorageSettings, error) {
	settings, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = settingsFromImageStorageConfig(s.fallback)
	}
	settings.SecretAccessKey = ""
	return settings, nil
}

func (s *ImageStorageSettingService) SecretConfigured(ctx context.Context) bool {
	settings, err := s.load(ctx)
	if err != nil || settings == nil {
		return strings.TrimSpace(s.fallback.SecretAccessKey) != ""
	}
	if settings.ReuseBackupS3 {
		cfg, err := s.backupCredentials(ctx)
		return err == nil && cfg != nil && strings.TrimSpace(cfg.SecretAccessKey) != ""
	}
	return strings.TrimSpace(settings.SecretAccessKey) != ""
}

// Update encrypts a new own-storage secret and preserves an existing secret
// when the UI submits an empty field. Saving invalidates the resolver cache,
// making changes effective for the following request without a restart.
func (s *ImageStorageSettingService) Update(ctx context.Context, in ImageStorageSettings) (*ImageStorageSettings, error) {
	normalizeImageStorageSettings(&in)
	if in.ReuseBackupS3 {
		in.Endpoint, in.Region, in.AccessKeyID, in.SecretAccessKey = "", "", "", ""
		in.ForcePathStyle = false
	} else if in.SecretAccessKey == "" {
		if old, err := s.load(ctx); err == nil && old != nil {
			in.SecretAccessKey = old.SecretAccessKey
		}
	} else {
		if s.backup == nil || !s.backup.EncryptionKeyConfigured() {
			return nil, ErrSecretEncryptionKeyNotConfigured
		}
		encrypted, err := s.encryptor.Encrypt(in.SecretAccessKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt image storage secret: %w", err)
		}
		in.SecretAccessKey = encrypted
	}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal image storage settings: %w", err)
	}
	if s.settingRepo == nil {
		return nil, errors.New("image storage settings repository is unavailable")
	}
	if err := s.settingRepo.Set(ctx, settingKeyImageStorageConfig, string(data)); err != nil {
		return nil, fmt.Errorf("save image storage settings: %w", err)
	}
	s.Invalidate()
	in.SecretAccessKey = ""
	return &in, nil
}

// TestConnection creates a client with the prospective settings only. It does
// not write a setting or contact a bucket.
func (s *ImageStorageSettingService) TestConnection(ctx context.Context, in ImageStorageSettings) error {
	normalizeImageStorageSettings(&in)
	if !in.ReuseBackupS3 && in.SecretAccessKey == "" {
		if old, err := s.load(ctx); err == nil && old != nil {
			in.SecretAccessKey = old.SecretAccessKey
		}
	}
	cfg, err := s.toImageStorageConfig(ctx, &in)
	if err != nil {
		return err
	}
	if !cfg.IsConfigured() {
		return ErrImageStorageIncomplete
	}
	if s.factory == nil {
		return errors.New("image storage factory is unavailable")
	}
	_, err = s.factory(ctx, cfg)
	return err
}

func (s *ImageStorageSettingService) effectiveConfig(ctx context.Context) (*config.ImageStorageConfig, error) {
	settings, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		fallback := s.fallback
		return &fallback, nil
	}
	return s.toImageStorageConfig(ctx, settings)
}

func (s *ImageStorageSettingService) toImageStorageConfig(ctx context.Context, in *ImageStorageSettings) (*config.ImageStorageConfig, error) {
	if in == nil {
		return nil, errors.New("image storage settings are nil")
	}
	cfg := &config.ImageStorageConfig{
		Enabled:         in.Enabled,
		Bucket:          in.Bucket,
		Prefix:          in.Prefix,
		PublicBaseURL:   in.PublicBaseURL,
		PresignExpiry:   in.PresignExpiry,
		MaxDownloadByte: in.MaxDownloadBytes,
		Endpoint:        in.Endpoint,
		Region:          in.Region,
		AccessKeyID:     in.AccessKeyID,
		SecretAccessKey: in.SecretAccessKey,
		ForcePathStyle:  in.ForcePathStyle,
	}
	if in.ReuseBackupS3 {
		backupCfg, err := s.backupCredentials(ctx)
		if err != nil {
			return nil, err
		}
		if backupCfg == nil {
			return nil, errors.New("image storage reuses backup S3 but no backup S3 configuration exists")
		}
		cfg.Endpoint = backupCfg.Endpoint
		cfg.Region = backupCfg.Region
		cfg.AccessKeyID = backupCfg.AccessKeyID
		cfg.SecretAccessKey = backupCfg.SecretAccessKey
		cfg.ForcePathStyle = backupCfg.ForcePathStyle
		if cfg.Bucket == "" {
			cfg.Bucket = backupCfg.Bucket
		}
	} else if cfg.SecretAccessKey != "" {
		decrypted, err := s.encryptor.Decrypt(cfg.SecretAccessKey)
		if err != nil {
			logger.L().Warn("image_storage secret decrypt failed; treating stored value as plaintext for compatibility", zap.Error(err))
		} else {
			cfg.SecretAccessKey = decrypted
		}
	}
	return cfg, nil
}

func (s *ImageStorageSettingService) backupCredentials(ctx context.Context) (*BackupS3Config, error) {
	if s == nil || s.backup == nil {
		return nil, errors.New("backup service is unavailable")
	}
	return s.backup.loadS3Config(ctx)
}

func (s *ImageStorageSettingService) load(ctx context.Context) (*ImageStorageSettings, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, settingKeyImageStorageConfig)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var settings ImageStorageSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, fmt.Errorf("parse image storage settings: %w", err)
	}
	return &settings, nil
}

func settingsFromImageStorageConfig(cfg config.ImageStorageConfig) *ImageStorageSettings {
	return &ImageStorageSettings{
		Enabled:          cfg.Enabled,
		Bucket:           cfg.Bucket,
		Prefix:           cfg.Prefix,
		PublicBaseURL:    cfg.PublicBaseURL,
		PresignExpiry:    cfg.PresignExpiry,
		MaxDownloadBytes: cfg.MaxDownloadByte,
		Endpoint:         cfg.Endpoint,
		Region:           cfg.Region,
		AccessKeyID:      cfg.AccessKeyID,
		SecretAccessKey:  cfg.SecretAccessKey,
		ForcePathStyle:   cfg.ForcePathStyle,
	}
}

func normalizeImageStorageSettings(in *ImageStorageSettings) {
	in.Bucket = strings.TrimSpace(in.Bucket)
	in.Endpoint = strings.TrimSpace(in.Endpoint)
	in.Region = strings.TrimSpace(in.Region)
	in.AccessKeyID = strings.TrimSpace(in.AccessKeyID)
	in.SecretAccessKey = strings.TrimSpace(in.SecretAccessKey)
	in.PublicBaseURL = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(in.PublicBaseURL), "/"))
	in.Prefix = strings.TrimSpace(in.Prefix)
	if in.Prefix == "" {
		in.Prefix = "images/"
	} else if !strings.HasSuffix(in.Prefix, "/") {
		in.Prefix += "/"
	}
	if in.Region == "" {
		in.Region = "auto"
	}
	if in.PresignExpiry <= 0 {
		in.PresignExpiry = 24
	}
	if in.MaxDownloadBytes <= 0 {
		in.MaxDownloadBytes = defaultImageMaxDownloadBytes
	}
}
