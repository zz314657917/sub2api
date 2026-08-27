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
		service.SettingKeyRegistrationEnabled:         "true",
		service.SettingKeySiteName:                    "Example Gateway",
		service.SettingKeySiteSubtitle:                "Example Gateway Platform",
		service.SettingKeySMTPHost:                    "smtp.example.com",
		service.SettingKeySMTPFrom:                    "noreply@example.com",
		service.SettingKeyTurnstileEnabled:            "true",
		service.SettingKeyCyberSessionBlockEnabled:    "true",
		service.SettingKeyCyberSessionBlockTTLSeconds: "7200",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"registration_enabled": false})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyRegistrationEnabled])
	require.Equal(t, "Example Gateway", repo.values[service.SettingKeySiteName])
	require.Equal(t, "Example Gateway Platform", repo.values[service.SettingKeySiteSubtitle])
	require.Equal(t, "smtp.example.com", repo.values[service.SettingKeySMTPHost])
	require.Equal(t, "noreply@example.com", repo.values[service.SettingKeySMTPFrom])
	require.Equal(t, "true", repo.values[service.SettingKeyTurnstileEnabled])
	require.Equal(t, "true", repo.values[service.SettingKeyCyberSessionBlockEnabled])
	require.Equal(t, "7200", repo.values[service.SettingKeyCyberSessionBlockTTLSeconds])
}

func TestUpdateSettingsCyberSessionBlockFieldsAreWritable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyCyberSessionBlockEnabled:    "true",
		service.SettingKeyCyberSessionBlockTTLSeconds: "7200",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{
		"cyber_session_block_enabled":     false,
		"cyber_session_block_ttl_seconds": 1800,
	})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "false", repo.values[service.SettingKeyCyberSessionBlockEnabled])
	require.Equal(t, "1800", repo.values[service.SettingKeyCyberSessionBlockTTLSeconds])
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, false, envelope.Data["cyber_session_block_enabled"])
	require.Equal(t, float64(1800), envelope.Data["cyber_session_block_ttl_seconds"])
}

func TestUpdateSettingsRejectsInvalidCyberSessionBlockTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyCyberSessionBlockEnabled:    "true",
		service.SettingKeyCyberSessionBlockTTLSeconds: "7200",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{
		"cyber_session_block_ttl_seconds": 0,
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "true", repo.values[service.SettingKeyCyberSessionBlockEnabled])
	require.Equal(t, "7200", repo.values[service.SettingKeyCyberSessionBlockTTLSeconds])
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

func TestUpdateSettingsForwardedClientIPFieldsArePersistedAndReturned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyAPIKeyACLTrustForwardedIP: "false",
		service.SettingKeyForwardedClientIPHeaders:  `["X-Existing-IP"]`,
		service.SettingKeySiteName:                  "Example Gateway",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{
		"api_key_acl_trust_forwarded_ip": true,
		"forwarded_client_ip_headers":    []string{" x-cdn-ip ", "X-CDN-IP", "true-client-ip"},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[service.SettingKeyAPIKeyACLTrustForwardedIP])
	require.Equal(t, `["X-Cdn-Ip","True-Client-Ip"]`, repo.values[service.SettingKeyForwardedClientIPHeaders])

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, true, envelope.Data["api_key_acl_trust_forwarded_ip"])
	require.Equal(t, []any{"X-Cdn-Ip", "True-Client-Ip"}, envelope.Data["forwarded_client_ip_headers"])
}

func TestUpdateSettingsForwardedClientIPFieldsPreserveOmissionAndAllowExplicitClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyAPIKeyACLTrustForwardedIP: "true",
		service.SettingKeyForwardedClientIPHeaders:  `["X-Existing-IP"]`,
		service.SettingKeySiteName:                  "Example Gateway",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{"site_name": "Updated Gateway"})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[service.SettingKeyAPIKeyACLTrustForwardedIP])
	require.Equal(t, `["X-Existing-Ip"]`, repo.values[service.SettingKeyForwardedClientIPHeaders])

	rec = updateSettingsPayload(t, handler, map[string]any{
		"forwarded_client_ip_headers": []string{},
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "[]", repo.values[service.SettingKeyForwardedClientIPHeaders])
}

func TestUpdateSettingsRejectsInvalidForwardedClientIPHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyAPIKeyACLTrustForwardedIP: "false",
		service.SettingKeyForwardedClientIPHeaders:  `["X-Existing-IP"]`,
		service.SettingKeySiteName:                  "Example Gateway",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{
		"forwarded_client_ip_headers": []string{"X Invalid"},
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, "false", repo.values[service.SettingKeyAPIKeyACLTrustForwardedIP])
	require.Equal(t, `["X-Existing-IP"]`, repo.values[service.SettingKeyForwardedClientIPHeaders])
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
