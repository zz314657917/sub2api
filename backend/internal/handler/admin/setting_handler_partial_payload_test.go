package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newPartialPayloadSettingsHandler(values map[string]string) (*SettingHandler, *settingHandlerRepoStub) {
	repo := &settingHandlerRepoStub{values: values}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	return NewSettingHandler(svc, nil, nil, nil, nil, nil), repo
}

func updateSettingsPayload(t *testing.T, handler *SettingHandler, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateSettings(c)
	return rec
}

func TestUpdateSettingsPartialPayloadKeepsUnsentKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyRegistrationEnabled: "true",
		service.SettingKeySiteName:            "Example Gateway",
		service.SettingKeySiteSubtitle:        "Example Gateway Platform",
		service.SettingKeySMTPHost:            "smtp.example.com",
		service.SettingKeySMTPFrom:            "noreply@example.com",
		service.SettingKeyTurnstileEnabled:    "true",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"registration_enabled": false})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyRegistrationEnabled])
	require.Equal(t, "Example Gateway", repo.values[service.SettingKeySiteName])
	require.Equal(t, "Example Gateway Platform", repo.values[service.SettingKeySiteSubtitle])
	require.Equal(t, "smtp.example.com", repo.values[service.SettingKeySMTPHost])
	require.Equal(t, "noreply@example.com", repo.values[service.SettingKeySMTPFrom])
	require.Equal(t, "true", repo.values[service.SettingKeyTurnstileEnabled])
}

func TestUpdateSettingsExplicitEmptyValueClearsSentField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeySiteName: "Example Gateway",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"site_name": ""})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "", repo.values[service.SettingKeySiteName])
}

func TestUpdateSettingsExplicitFalseValueUpdatesSentField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyTurnstileEnabled: "true",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"turnstile_enabled": false})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyTurnstileEnabled])
}

func TestUpdateSettingsSMTPFromAliasIsWritable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeySMTPFrom: "old@example.com",
		service.SettingKeySiteName: "Example Gateway",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"smtp_from_email": "new@example.com"})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "new@example.com", repo.values[service.SettingKeySMTPFrom])
	require.Equal(t, "Example Gateway", repo.values[service.SettingKeySiteName])
}

func TestOmittedSettingKeysTracksValueFieldsAndAliases(t *testing.T) {
	omitted := omittedSettingKeys(map[string]json.RawMessage{
		"smtp_from_email":                     json.RawMessage(`"new@example.com"`),
		"turnstile_enabled":                   json.RawMessage(`false`),
		"smtp_port":                           json.RawMessage(`0`),
		"registration_email_suffix_whitelist": json.RawMessage(`[]`),
	})

	_, smtpFromOmitted := omitted[service.SettingKeySMTPFrom]
	_, siteNameOmitted := omitted[service.SettingKeySiteName]
	_, riskEnabledOmitted := omitted[service.SettingKeyRegistrationRiskEnabled]
	_, turnstileOmitted := omitted[service.SettingKeyTurnstileEnabled]
	_, smtpPortOmitted := omitted[service.SettingKeySMTPPort]
	_, emailSuffixesOmitted := omitted[service.SettingKeyRegistrationEmailSuffixWhitelist]
	require.False(t, smtpFromOmitted)
	require.True(t, siteNameOmitted)
	require.False(t, riskEnabledOmitted, "pointer fields retain their existing merge behavior")
	require.False(t, turnstileOmitted)
	require.False(t, smtpPortOmitted)
	require.False(t, emailSuffixesOmitted)
}
