package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/require"
)

func TestNormalizePasskeyName(t *testing.T) {
	require.Equal(t, defaultPasskeyName, normalizePasskeyName("   "))
	require.Equal(t, "Laptop", normalizePasskeyName("  Laptop  "))

	longName := strings.Repeat("密", maxPasskeyNameLength+10)
	require.Len(t, []rune(normalizePasskeyName(longName)), maxPasskeyNameLength)
}

func TestVerifyPasskeyPassword(t *testing.T) {
	user := &User{}
	require.NoError(t, user.SetPassword("correct-password"))
	require.ErrorIs(t, verifyPasskeyPassword(user, ""), ErrPasswordRequired)
	require.ErrorIs(t, verifyPasskeyPassword(user, "wrong-password"), ErrPasswordIncorrect)
	require.NoError(t, verifyPasskeyPassword(user, "correct-password"))
}

func TestPasskeySummaryReportsCurrentBackupState(t *testing.T) {
	record := &PasskeyCredentialRecord{
		Credential: webauthn.Credential{
			Flags: webauthn.CredentialFlags{BackupEligible: true},
		},
	}
	require.False(t, passkeySummary(record).Backup)

	record.Credential.Flags.BackupState = true
	require.True(t, passkeySummary(record).Backup)
}

type passkeySettingRepoStub struct {
	value string
	err   error
}

func (s *passkeySettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *passkeySettingRepoStub) GetValue(context.Context, string) (string, error) {
	return s.value, s.err
}

func (s *passkeySettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *passkeySettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *passkeySettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *passkeySettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *passkeySettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestPasskeySettingEnabledIsOptIn(t *testing.T) {
	configured := NewSettingService(nil, &config.Config{
		WebAuthn: config.WebAuthnConfig{Enabled: true},
	})
	require.False(t, configured.passkeySettingEnabled(map[string]string{}))
	require.True(t, configured.passkeySettingEnabled(map[string]string{SettingKeyPasskeyEnabled: "true"}))
	require.False(t, configured.passkeySettingEnabled(map[string]string{SettingKeyPasskeyEnabled: "false"}))

	notConfigured := NewSettingService(nil, &config.Config{})
	require.False(t, notConfigured.passkeySettingEnabled(map[string]string{SettingKeyPasskeyEnabled: "true"}))
}

func TestPasskeyEnabledRequiresExplicitPersistedSetting(t *testing.T) {
	configured := &config.Config{WebAuthn: config.WebAuthnConfig{Enabled: true}}

	t.Run("missing setting defaults disabled", func(t *testing.T) {
		enabled, err := NewSettingService(&passkeySettingRepoStub{err: ErrSettingNotFound}, configured).
			PasskeyEnabled(context.Background())
		require.NoError(t, err)
		require.False(t, enabled)
	})

	t.Run("explicit true enables", func(t *testing.T) {
		enabled, err := NewSettingService(&passkeySettingRepoStub{value: "true"}, configured).
			PasskeyEnabled(context.Background())
		require.NoError(t, err)
		require.True(t, enabled)
	})

	t.Run("storage error fails closed", func(t *testing.T) {
		enabled, err := NewSettingService(&passkeySettingRepoStub{err: errors.New("database unavailable")}, configured).
			PasskeyEnabled(context.Background())
		require.Error(t, err)
		require.False(t, enabled)
	})
}
