package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerApplyOAuthCredentials_MergesExtraAndClearsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminSvc := newStubAdminService()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	router := gin.New()
	router.POST("/api/v1/admin/accounts/:id/apply-oauth-credentials", handler.ApplyOAuthCredentials)

	body := []byte(`{
		"type": "oauth",
		"credentials": {"access_token": "new-token", "refresh_token": "new-refresh"},
		"extra": {"account_uuid": "acc-1", "org_uuid": "org-1"}
	}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/3/apply-oauth-credentials", bytes.NewReader(body))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.updatedAccounts, 1)
	require.Equal(t, int64(3), adminSvc.updatedAccounts[0].id)
	require.Equal(t, "oauth", adminSvc.updatedAccounts[0].input.Type)
	require.Equal(t, "new-token", adminSvc.updatedAccounts[0].input.Credentials["access_token"])

	require.Len(t, adminSvc.extraUpdates, 1)
	require.Equal(t, int64(3), adminSvc.extraUpdates[0].id)
	require.Equal(t, "acc-1", adminSvc.extraUpdates[0].updates["account_uuid"])
	require.Equal(t, "org-1", adminSvc.extraUpdates[0].updates["org_uuid"])

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Equal(t, int64(3), resp.Data.ID)
}
