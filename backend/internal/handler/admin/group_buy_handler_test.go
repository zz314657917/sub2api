package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGroupBuyHandlerRejectsInvalidRoundIDsBeforeServiceAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGroupBuyHandler(nil)
	router := gin.New()
	router.GET("/admin/group-buy/rounds/:id/seats", h.ListRoundSeats)
	router.POST("/admin/group-buy/rounds/:id/process-refunds", h.ProcessRefunds)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/admin/group-buy/rounds/0/"+map[string]string{
			http.MethodGet:  "seats",
			http.MethodPost: "process-refunds",
		}[method], nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		require.Equal(t, http.StatusBadRequest, res.Code)
		require.Contains(t, res.Body.String(), "INVALID_ID")
	}
}
