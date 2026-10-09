package admin

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestStoreEnabledAdminUpdatePreservesOmissionAndPersistsFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repo := newPartialPayloadSettingsHandler(map[string]string{
		service.SettingKeyServiceStoreEnabled: "true",
	})

	rec := updateSettingsPayload(t, handler, map[string]any{})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "true", repo.values[service.SettingKeyServiceStoreEnabled])

	rec = updateSettingsPayload(t, handler, map[string]any{"service_store_enabled": false})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "false", repo.values[service.SettingKeyServiceStoreEnabled])

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, false, envelope.Data["service_store_enabled"])

	changed := diffSettings(
		&service.SystemSettings{ServiceStoreEnabled: true},
		&service.SystemSettings{ServiceStoreEnabled: false},
		&service.AuthSourceDefaultSettings{},
		&service.AuthSourceDefaultSettings{},
		UpdateSettingsRequest{},
	)
	require.Contains(t, changed, "service_store_enabled")
}
