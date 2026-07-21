package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type backupS95SettingRepo struct {
	values map[string]string
}

func (r *backupS95SettingRepo) Get(context.Context, string) (*Setting, error) {
	return nil, nil
}

func (r *backupS95SettingRepo) GetValue(_ context.Context, key string) (string, error) {
	return r.values[key], nil
}

func (r *backupS95SettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *backupS95SettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *backupS95SettingRepo) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

func (r *backupS95SettingRepo) GetAll(context.Context) (map[string]string, error) {
	return r.values, nil
}

func (r *backupS95SettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type backupS95Encryptor struct{}

func (backupS95Encryptor) Encrypt(value string) (string, error) {
	return "ENC:" + value, nil
}

func (backupS95Encryptor) Decrypt(value string) (string, error) {
	return strings.TrimPrefix(value, "ENC:"), nil
}

func newBackupS95Service(fixedKey bool, repo *backupS95SettingRepo) *BackupService {
	cfg := &config.Config{Totp: config.TotpConfig{EncryptionKeyConfigured: fixedKey}}
	return NewBackupService(repo, cfg, backupS95Encryptor{}, nil, nil)
}

func TestS95BackupServiceRejectsNewSecretWithEphemeralKey(t *testing.T) {
	repo := &backupS95SettingRepo{values: map[string]string{}}
	svc := newBackupS95Service(false, repo)

	_, err := svc.UpdateS3Config(context.Background(), BackupS3Config{
		Bucket:          "bucket",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
	})
	require.ErrorIs(t, err, ErrSecretEncryptionKeyNotConfigured)
	require.Empty(t, repo.values[settingKeyBackupS3Config])
}

func TestS95BackupServiceAllowsReuseWithEphemeralKey(t *testing.T) {
	repo := &backupS95SettingRepo{values: map[string]string{}}
	svc := newBackupS95Service(false, repo)

	_, err := svc.UpdateS3Config(context.Background(), BackupS3Config{
		Bucket:      "bucket",
		AccessKeyID: "access",
	})
	require.NoError(t, err)
	require.NotEmpty(t, repo.values[settingKeyBackupS3Config])
}

func TestS95BackupServiceAllowsNewSecretWithFixedKey(t *testing.T) {
	repo := &backupS95SettingRepo{values: map[string]string{}}
	svc := newBackupS95Service(true, repo)

	_, err := svc.UpdateS3Config(context.Background(), BackupS3Config{
		Bucket:          "bucket",
		AccessKeyID:     "access",
		SecretAccessKey: "secret",
	})
	require.NoError(t, err)
	require.Contains(t, repo.values[settingKeyBackupS3Config], "ENC:secret")
}

func TestS95BackupServiceReportsEncryptionKeyConfiguration(t *testing.T) {
	repo := &backupS95SettingRepo{values: map[string]string{}}
	require.True(t, newBackupS95Service(true, repo).EncryptionKeyConfigured())
	require.False(t, newBackupS95Service(false, repo).EncryptionKeyConfigured())
}
