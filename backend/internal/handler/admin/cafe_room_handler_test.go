package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cafeRoomHandlerRepositoryStub struct {
	room       *service.CafeRoom
	listParams pagination.PaginationParams
	listStatus string
	listZone   string
	listSearch string
}

func newCafeRoomHandlerRepositoryStub() *cafeRoomHandlerRepositoryStub {
	accountID := int64(30)
	return &cafeRoomHandlerRepositoryStub{room: &service.CafeRoom{
		ID:        1,
		Code:      "ROOM-001",
		Name:      "Room 1",
		PlanID:    10,
		AccountID: &accountID,
		ZoneKey:   "featured",
		ThemeKey:  "warm_wood",
		Status:    service.CafeRoomStatusEnabled,
		Plan: &service.CafeRoomPlan{
			ID:              10,
			TargetGroupID:   20,
			FulfillmentMode: service.CafeRoomFulfillmentMode,
			GroupPlatform:   "openai",
			GroupAccessMode: service.CafeRoomGroupAccessMode,
			TotalShares:     4,
			SeatCount:       4,
			TimeoutMinutes:  60,
		},
	}}
}

func (r *cafeRoomHandlerRepositoryStub) List(_ context.Context, params pagination.PaginationParams, status, zone, search string) ([]service.CafeRoom, *pagination.PaginationResult, error) {
	r.listParams = params
	r.listStatus = status
	r.listZone = zone
	r.listSearch = search
	return []service.CafeRoom{*r.room}, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 1, Pages: 1}, nil
}

func (r *cafeRoomHandlerRepositoryStub) GetByID(_ context.Context, id int64) (*service.CafeRoom, error) {
	if id != r.room.ID {
		return nil, service.ErrCafeRoomNotFound
	}
	copy := *r.room
	return &copy, nil
}

func (r *cafeRoomHandlerRepositoryStub) GetPlan(context.Context, int64) (*service.CafeRoomPlan, error) {
	copy := *r.room.Plan
	return &copy, nil
}

func (r *cafeRoomHandlerRepositoryStub) GetAccount(context.Context, int64) (string, string, []int64, error) {
	return service.StatusActive, "openai", []int64{20}, nil
}

func (r *cafeRoomHandlerRepositoryStub) HasOperationalAccount(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (r *cafeRoomHandlerRepositoryStub) HasLiveRound(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *cafeRoomHandlerRepositoryStub) Create(_ context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	copy := *room
	copy.ID = 1
	copy.Plan = r.room.Plan
	r.room = &copy
	return &copy, nil
}

func (r *cafeRoomHandlerRepositoryStub) Update(_ context.Context, room *service.CafeRoom) (*service.CafeRoom, error) {
	copy := *room
	r.room = &copy
	return &copy, nil
}

func (r *cafeRoomHandlerRepositoryStub) Delete(context.Context, int64) error {
	return nil
}

func (r *cafeRoomHandlerRepositoryStub) CreateOpenRound(_ context.Context, roomID int64, now time.Time) (*service.CafeRound, error) {
	return &service.CafeRound{
		ID:                100,
		PlanID:            r.room.PlanID,
		CafeRoomID:        &roomID,
		AssignedAccountID: r.room.AccountID,
		Status:            service.CafeRoundStatusOpen,
		TotalShares:       4,
		TotalSeats:        4,
		DeadlineAt:        now.Add(time.Hour),
	}, nil
}

func newCafeRoomHandlerTestRouter(repo *cafeRoomHandlerRepositoryStub) *gin.Engine {
	handler := NewCafeRoomHandler(service.NewCafeRoomService(repo))
	router := gin.New()
	router.GET("/rooms", handler.List)
	router.POST("/rooms", handler.Create)
	router.POST("/rooms/bulk", handler.BulkCreate)
	router.GET("/rooms/:id", handler.Get)
	router.PATCH("/rooms/:id", handler.Update)
	router.DELETE("/rooms/:id", handler.Delete)
	router.POST("/rooms/:id/open-round", handler.OpenRound)
	return router
}

func TestCafeRoomHandlerListUsesPaginatedEnvelopeAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newCafeRoomHandlerRepositoryStub()
	router := newCafeRoomHandlerTestRouter(repo)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/rooms?page=2&page_size=5&status=enabled&zone=featured&search=Room", nil)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 2, repo.listParams.Page)
	require.Equal(t, 5, repo.listParams.PageSize)
	require.Equal(t, "enabled", repo.listStatus)
	require.Equal(t, "featured", repo.listZone)
	require.Equal(t, "Room", repo.listSearch)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	data := payload["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
	require.Len(t, data["items"].([]any), 1)
}

func TestCafeRoomHandlerCreateAndOpenRoundDoNotExposeAccountSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newCafeRoomHandlerRepositoryStub()
	router := newCafeRoomHandlerTestRouter(repo)

	create := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/rooms", strings.NewReader(`{"code":"ROOM-001","name":"Room 1","plan_id":10,"account_id":30,"status":"enabled"}`))
	createReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(create, createReq)
	require.Equal(t, http.StatusCreated, create.Code)
	require.Contains(t, create.Body.String(), `"code":"ROOM-001"`)
	require.NotContains(t, create.Body.String(), "credentials")
	require.NotContains(t, create.Body.String(), "oauth")
	require.NotContains(t, create.Body.String(), "proxy_url")

	openRound := httptest.NewRecorder()
	openRoundReq := httptest.NewRequest(http.MethodPost, "/rooms/1/open-round", nil)
	router.ServeHTTP(openRound, openRoundReq)
	require.Equal(t, http.StatusCreated, openRound.Code)
	require.Contains(t, openRound.Body.String(), `"status":"open"`)
}

func TestCafeRoomHandlerRejectsInvalidIDsAndStatuses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newCafeRoomHandlerTestRouter(newCafeRoomHandlerRepositoryStub())

	for _, request := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/rooms/0"},
		{method: http.MethodPatch, path: "/rooms/nope", body: `{}`},
		{method: http.MethodDelete, path: "/rooms/-1"},
		{method: http.MethodPost, path: "/rooms/0/open-round"},
	} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(request.method, request.path, strings.NewReader(request.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code, request.path)
		require.Contains(t, recorder.Body.String(), "INVALID_ID", request.path)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rooms", strings.NewReader(`{"code":"ROOM-002","name":"Room 2","plan_id":10,"account_id":30,"status":"online"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "CAFE_ROOM_INVALID_STATUS")
}
