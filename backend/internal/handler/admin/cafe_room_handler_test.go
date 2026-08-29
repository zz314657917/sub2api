package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cafeRoomHandlerRepositoryStub struct {
	room         *service.CafeRoom
	listParams   pagination.PaginationParams
	listStatus   string
	listZone     string
	listSearch   string
	optionParams service.CafeRoomAccountOptionParams
}

type cafeRoundFulfillmentHandlerStub struct {
	pendingParams service.CafePendingRoundParams
	optionRoundID int64
	optionParams  service.CafeRoundAccountOptionParams
	assignRoundID int64
	assignAccount int64
}

type cafeWorkstationLayoutHandlerStub struct {
	layout service.PixelCafeWorkstationLayout
	writes int
}

type cafeQuotaResetHandlerStub struct {
	roomIDs []*int64
}

func (s *cafeQuotaResetHandlerStub) AdminResetCafeRateLimitUsage(_ context.Context, roomID *int64) (*service.AdminCafeQuotaResetResult, error) {
	s.roomIDs = append(s.roomIDs, roomID)
	result := &service.AdminCafeQuotaResetResult{AffectedKeys: 3, Scope: "all"}
	if roomID != nil {
		result.Scope = "room"
		result.RoomID = roomID
	}
	return result, nil
}

func (s *cafeWorkstationLayoutHandlerStub) GetPixelCafeWorkstationLayout(context.Context) (service.PixelCafeWorkstationLayout, error) {
	return s.layout, nil
}

func (s *cafeWorkstationLayoutHandlerStub) SetPixelCafeWorkstationLayout(_ context.Context, layout service.PixelCafeWorkstationLayout) (service.PixelCafeWorkstationLayout, error) {
	s.layout = layout
	s.writes++
	return layout, nil
}

func (s *cafeRoundFulfillmentHandlerStub) ListPendingRounds(_ context.Context, params service.CafePendingRoundParams) ([]service.CafePendingRound, *pagination.PaginationResult, error) {
	s.pendingParams = params
	return []service.CafePendingRound{{ID: 88, Status: service.GroupBuyRoundStatusAwaitingAccount, RoomID: 1, RoomCode: "ROOM-001", RoomName: "Room 1", SubscriptionTier: "plus", PaidShares: 10, TotalShares: 10, JoinedBuyers: 3, MaxBuyers: 4}}, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 1, Pages: 1}, nil
}

func (s *cafeRoundFulfillmentHandlerStub) ListRoundAccountOptions(_ context.Context, roundID int64, params service.CafeRoundAccountOptionParams) ([]service.CafeRoundAccountOption, *pagination.PaginationResult, error) {
	s.optionRoundID, s.optionParams = roundID, params
	return []service.CafeRoundAccountOption{{ID: 30, Name: "Cafe Plus", Platform: "openai", Status: service.StatusActive, PlanType: "plus", EmailMasked: "o***r@example.com"}}, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 1, Pages: 1}, nil
}

func (s *cafeRoundFulfillmentHandlerStub) AssignAccountAndActivateRound(_ context.Context, roundID, accountID int64) (*service.CafePendingRound, error) {
	s.assignRoundID, s.assignAccount = roundID, accountID
	return &service.CafePendingRound{ID: roundID, Status: service.GroupBuyRoundStatusActive, RoomID: 1, SubscriptionTier: "plus", PaidShares: 10, TotalShares: 10}, nil
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
			ID:                10,
			Status:            service.GroupBuyPlanStatusActive,
			TargetGroupID:     20,
			FulfillmentMode:   service.CafeRoomFulfillmentMode,
			AutoCreateRoomKey: true,
			GroupPlatform:     "openai",
			GroupAccessMode:   service.CafeRoomGroupAccessMode,
			TargetGroupStatus: service.StatusActive,
			TotalShares:       4,
			SeatCount:         4,
			TimeoutMinutes:    60,
			ValidityDays:      30,
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

func (r *cafeRoomHandlerRepositoryStub) ResolveDefaultRoomManagedGroupID(context.Context) (int64, error) {
	return 20, nil
}

func (r *cafeRoomHandlerRepositoryStub) GetAccount(context.Context, int64) (string, string, []int64, error) {
	return service.StatusActive, "openai", []int64{20}, nil
}

func (r *cafeRoomHandlerRepositoryStub) ListAccountOptions(_ context.Context, params service.CafeRoomAccountOptionParams) ([]service.CafeRoomAccountOption, *pagination.PaginationResult, error) {
	r.optionParams = params
	return []service.CafeRoomAccountOption{{ID: 30, Name: "Cafe account", Platform: "openai", Status: service.StatusActive, EmailMasked: "c***e@example.com"}}, &pagination.PaginationResult{Page: 1, PageSize: 20, Total: 1, Pages: 1}, nil
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
	if copy.Plan != nil && copy.PlanID <= 0 {
		copy.PlanID = 10
		copy.Plan.ID = 10
		copy.Plan.GroupPlatform = "openai"
		copy.Plan.GroupAccessMode = service.CafeRoomGroupAccessMode
		copy.Plan.TargetGroupStatus = service.StatusActive
	} else {
		copy.Plan = r.room.Plan
	}
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

func (r *cafeRoomHandlerRepositoryStub) PauseOpenRound(_ context.Context, roomID int64, now time.Time) (*service.CafeRound, error) {
	return &service.CafeRound{
		ID:          100,
		PlanID:      r.room.PlanID,
		CafeRoomID:  &roomID,
		Status:      service.GroupBuyRoundStatusCancelled,
		TotalShares: 4,
		TotalSeats:  4,
		DeadlineAt:  now.Add(time.Hour),
	}, nil
}

func newCafeRoomHandlerTestRouter(repo *cafeRoomHandlerRepositoryStub) *gin.Engine {
	handler := NewCafeRoomHandler(service.NewCafeRoomService(repo))
	router := gin.New()
	router.GET("/rooms", handler.List)
	router.GET("/rooms/account-options", handler.AccountOptions)
	router.POST("/rooms", handler.Create)
	router.POST("/rooms/bulk", handler.BulkCreate)
	router.GET("/rooms/:id", handler.Get)
	router.PATCH("/rooms/:id", handler.Update)
	router.DELETE("/rooms/:id", handler.Delete)
	router.POST("/rooms/:id/open-round", handler.OpenRound)
	router.POST("/rooms/:id/pause-round", handler.PauseRound)
	return router
}

func newCafeRoundFulfillmentHandlerTestRouter(activation cafeRoundFulfillmentService) *gin.Engine {
	handler := &CafeRoomHandler{activation: activation}
	router := gin.New()
	router.GET("/rounds/pending", handler.ListPendingRounds)
	router.GET("/rounds/:id/account-options", handler.ListRoundAccountOptions)
	router.POST("/rounds/:id/assign-account", handler.AssignRoundAccount)
	return router
}

func TestCafeRoomHandlerAccountOptionsAreBoundedAndRedacted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newCafeRoomHandlerRepositoryStub()
	router := newCafeRoomHandlerTestRouter(repo)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/rooms/account-options?plan_id=10&page=2&page_size=99&search=owner&exclude_room_id=1", nil)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(10), repo.optionParams.PlanID)
	require.Equal(t, int64(1), repo.optionParams.ExcludeRoomID)
	require.Equal(t, 2, repo.optionParams.Page)
	require.Equal(t, 50, repo.optionParams.PageSize)
	require.Equal(t, "owner", repo.optionParams.Search)
	for _, prohibited := range []string{"credentials", "api_key", "access_token", "base_url", "proxy"} {
		require.NotContains(t, recorder.Body.String(), prohibited)
	}
	require.Contains(t, recorder.Body.String(), `"email_masked":"c***e@example.com"`)

	for _, path := range []string{
		"/rooms/account-options?plan_id=0",
		"/rooms/account-options?plan_id=10&exclude_room_id=nope",
		"/rooms/account-options?ids=30,invalid",
		"/rooms/account-options",
	} {
		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusBadRequest, recorder.Code, path)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/rooms/account-options?ids=30,31", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []int64{30, 31}, repo.optionParams.IDs)
}

func TestCafeRoomHandlerReadsAndUpdatesWorkstationLayout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &cafeWorkstationLayoutHandlerStub{layout: service.DefaultPixelCafeWorkstationLayout()}
	handler := &CafeRoomHandler{settings: stub}
	router := gin.New()
	router.GET("/layout", handler.GetWorkstationLayout)
	router.PUT("/layout", handler.UpdateWorkstationLayout)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/layout", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":10`)

	draft := make(service.PixelCafeWorkstationLayout, 0, 50)
	for id := 1; id <= 50; id++ {
		draft = append(draft, service.PixelCafeWorkstationPosition{ID: id, X: 340, Y: 250})
	}
	draft[0].X = 410
	body, err := json.Marshal(draft)
	require.NoError(t, err)
	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/layout", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, stub.writes)
	require.Len(t, stub.layout, 50)
	require.Equal(t, float64(410), stub.layout[0].X)

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/layout", strings.NewReader(`{"workstations":[]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, 1, stub.writes)

	recorder = httptest.NewRecorder()
	oversized := `[{"id":1,"x":340,"y":250,"padding":"` + strings.Repeat("x", 5*1024) + `"}]`
	request = httptest.NewRequest(http.MethodPut, "/layout", strings.NewReader(oversized))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, 1, stub.writes)
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
	createReq := httptest.NewRequest(http.MethodPost, "/rooms", strings.NewReader(`{"name":"Room 1","description":"owned plan","status":"enabled","plan":{"subscription_tier":"pro","total_shares":10,"max_buyers":4,"max_shares_per_user":10,"price_per_share":12,"timeout_minutes":60,"fulfillment_timeout_minutes":1440,"validity_days":30,"target_group_id":20,"refund_mode":"balance_credit"}}`))
	createReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(create, createReq)
	require.Equal(t, http.StatusCreated, create.Code)
	require.Regexp(t, regexp.MustCompile(`"code":"ROOM-[A-Z2-9]{8}"`), create.Body.String())
	require.Contains(t, create.Body.String(), `"subscription_tier":"pro"`)
	require.Contains(t, create.Body.String(), `"fulfillment_timeout_minutes":1440`)
	require.NotContains(t, create.Body.String(), "credentials")
	require.NotContains(t, create.Body.String(), "oauth")
	require.NotContains(t, create.Body.String(), "proxy_url")

	openRound := httptest.NewRecorder()
	openRoundReq := httptest.NewRequest(http.MethodPost, "/rooms/1/open-round", nil)
	router.ServeHTTP(openRound, openRoundReq)
	require.Equal(t, http.StatusCreated, openRound.Code)
	require.Contains(t, openRound.Body.String(), `"status":"open"`)

	pauseRound := httptest.NewRecorder()
	pauseRoundReq := httptest.NewRequest(http.MethodPost, "/rooms/1/pause-round", nil)
	router.ServeHTTP(pauseRound, pauseRoundReq)
	require.Equal(t, http.StatusOK, pauseRound.Code)
	require.Contains(t, pauseRound.Body.String(), `"status":"cancelled"`)
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
		{method: http.MethodPost, path: "/rooms/0/pause-round"},
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

func TestCafeRoomHandlerQuotaResetEndpointsUseScopedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	quota := &cafeQuotaResetHandlerStub{}
	handler := NewCafeRoomHandler(nil)
	handler.SetQuotaResetService(quota)
	router := gin.New()
	router.POST("/rooms/reset-quotas", handler.ResetAllQuotas)
	router.POST("/rooms/:id/reset-quotas", handler.ResetRoomQuotas)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/rooms/42/reset-quotas", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, quota.roomIDs, 1)
	require.NotNil(t, quota.roomIDs[0])
	require.Equal(t, int64(42), *quota.roomIDs[0])
	require.Contains(t, recorder.Body.String(), `"affected_keys":3`)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/rooms/reset-quotas", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, quota.roomIDs, 2)
	require.Nil(t, quota.roomIDs[1])
}

func TestCafeRoomHandlerPendingFulfillmentEndpointsArePaginatedAndRedacted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	activation := &cafeRoundFulfillmentHandlerStub{}
	router := newCafeRoundFulfillmentHandlerTestRouter(activation)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/rounds/pending?page=2&page_size=15&search=%20ROOM%20", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 2, activation.pendingParams.Page)
	require.Equal(t, 15, activation.pendingParams.PageSize)
	require.Equal(t, "ROOM", activation.pendingParams.Search)
	require.Contains(t, recorder.Body.String(), `"subscription_tier":"plus"`)
	require.Contains(t, recorder.Body.String(), `"paid_shares":10`)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/rounds/88/account-options?page=1&page_size=10&search=owner", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(88), activation.optionRoundID)
	require.Equal(t, "owner", activation.optionParams.Search)
	require.Contains(t, recorder.Body.String(), `"email_masked":"o***r@example.com"`)
	for _, prohibited := range []string{"credentials", "access_token", "refresh_token", "api_key"} {
		require.NotContains(t, recorder.Body.String(), prohibited)
	}

	recorder = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/rounds/88/assign-account", strings.NewReader(`{"account_id":30}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(88), activation.assignRoundID)
	require.Equal(t, int64(30), activation.assignAccount)
	require.Contains(t, recorder.Body.String(), `"status":"active"`)
}

func TestCafeRoomHandlerPendingFulfillmentRejectsInvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newCafeRoundFulfillmentHandlerTestRouter(&cafeRoundFulfillmentHandlerStub{})
	for _, request := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/rounds/nope/account-options"},
		{method: http.MethodPost, path: "/rounds/0/assign-account", body: `{"account_id":30}`},
		{method: http.MethodPost, path: "/rounds/88/assign-account", body: `{"account_id":0}`},
	} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(request.method, request.path, strings.NewReader(request.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code, request.path)
		require.Contains(t, recorder.Body.String(), "INVALID_ID", request.path)
	}
}
