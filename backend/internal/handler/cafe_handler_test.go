package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCafeHandlerMyRoomsAndPublicQueriesRejectInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewCafeHandler(nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		c.Next()
	})
	router.GET("/overview", handler.Overview)
	router.GET("/lobby-activity", handler.LobbyActivity)
	router.GET("/rooms", handler.ListRooms)
	router.GET("/rooms/:id", handler.GetRoom)
	router.GET("/my-rooms", handler.MyRooms)

	for _, path := range []string{"/overview?room_limit=0", "/rooms?featured=not-a-bool", "/rooms/0", "/my-rooms?status=active,unknown"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusBadRequest, recorder.Code, path)
	}
}

func TestCafeHandlerLobbyAndMyRoomsRequireAuthenticatedSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/overview", NewCafeHandler(nil, nil).Overview)
	router.GET("/lobby-activity", NewCafeHandler(nil, nil).LobbyActivity)
	router.GET("/my-rooms", NewCafeHandler(nil, nil).MyRooms)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/overview", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/my-rooms", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/lobby-activity", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestCafeHandlerRejectsClientOwnedOrderFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		c.Next()
	})
	router.POST("/rooms/:id/orders", NewCafeHandler(nil, &service.CafeRoomOrderService{}).CreateOrder)
	recorder := httptest.NewRecorder()
	recorderRequest := httptest.NewRequest(http.MethodPost, "/rooms/1/orders", bytes.NewBufferString(`{"seat_no":1,"payment_type":"alipay","agreement_accepted":true,"plan_id":99}`))
	recorderRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, recorderRequest)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
