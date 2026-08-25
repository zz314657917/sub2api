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

func TestCreateSparkShadowHandlerUsesProtectedServiceContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminService := newStubAdminService()
	handler := NewOpenAIOAuthHandler(nil, adminService, nil, nil)
	router := gin.New()
	router.POST("/api/v1/admin/accounts/:id/shadow", handler.CreateShadow)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/42/shadow", bytes.NewBufferString(`{"name":"Spark","priority":9,"concurrency":3,"group_ids":[7]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(42), adminService.createdShadow.parentID)
	require.Equal(t, "Spark", adminService.createdShadow.name)
	require.Equal(t, []int64{7}, adminService.createdShadow.groupIDs)
	var body struct {
		Code int `json:"code"`
		Data struct {
			ParentAccountID *int64 `json:"parent_account_id"`
			QuotaDimension  string `json:"quota_dimension"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.NotNil(t, body.Data.ParentAccountID)
	require.Equal(t, int64(42), *body.Data.ParentAccountID)
	require.Equal(t, "spark", body.Data.QuotaDimension)
}
