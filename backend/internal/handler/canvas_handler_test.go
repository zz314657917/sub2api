package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeCanvasHandlerService struct {
	lastListUserID   int64
	lastSaveUserID   int64
	lastSaveInput    service.CanvasSaveInput
	lastRunUserID    int64
	lastRunInput     service.CanvasRunCreateInput
	lastCancelUserID int64
}

func (s *fakeCanvasHandlerService) ListCanvases(_ context.Context, userID int64, _ service.CanvasListFilters) ([]service.CanvasListItem, int, error) {
	s.lastListUserID = userID
	return []service.CanvasListItem{{
		ID:        12,
		UserID:    userID,
		Title:     "Campaign",
		Viewport:  map[string]any{},
		Metadata:  map[string]any{},
		NodeCount: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}, 1, nil
}

func (s *fakeCanvasHandlerService) GetCanvas(_ context.Context, userID int64, canvasID int64) (*service.CanvasDocument, error) {
	return &service.CanvasDocument{
		ID:        canvasID,
		UserID:    userID,
		Title:     "Campaign",
		Nodes:     []service.CanvasNode{{ID: "prompt", Type: service.CanvasNodeTypePrompt}},
		Edges:     []service.CanvasEdge{},
		Viewport:  map[string]any{},
		Metadata:  map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *fakeCanvasHandlerService) SaveCanvas(_ context.Context, userID int64, input service.CanvasSaveInput) (*service.CanvasDocument, error) {
	s.lastSaveUserID = userID
	s.lastSaveInput = input
	id := input.ID
	if id <= 0 {
		id = 20
	}
	return &service.CanvasDocument{
		ID:          id,
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Nodes:       input.Nodes,
		Edges:       input.Edges,
		Viewport:    input.Viewport,
		Metadata:    input.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (s *fakeCanvasHandlerService) DeleteCanvas(_ context.Context, _ int64, _ int64) error {
	return nil
}

func (s *fakeCanvasHandlerService) CreateRun(_ context.Context, userID int64, input service.CanvasRunCreateInput) (*service.CanvasRun, error) {
	s.lastRunUserID = userID
	s.lastRunInput = input
	return &service.CanvasRun{
		ID:          30,
		UserID:      userID,
		CanvasID:    input.CanvasID,
		Status:      service.CanvasRunStatusPending,
		TriggerType: "manual",
		Input:       map[string]any{},
		Output:      map[string]any{},
		Metadata:    map[string]any{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (s *fakeCanvasHandlerService) ListRuns(_ context.Context, userID int64, _ service.CanvasRunListFilters) ([]service.CanvasRun, int, error) {
	return []service.CanvasRun{{
		ID:        30,
		UserID:    userID,
		CanvasID:  12,
		Status:    service.CanvasRunStatusPending,
		Input:     map[string]any{},
		Output:    map[string]any{},
		Metadata:  map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}, 1, nil
}

func (s *fakeCanvasHandlerService) GetRun(_ context.Context, userID int64, runID int64) (*service.CanvasRun, error) {
	return &service.CanvasRun{
		ID:        runID,
		UserID:    userID,
		CanvasID:  12,
		Status:    service.CanvasRunStatusPending,
		Input:     map[string]any{},
		Output:    map[string]any{},
		Metadata:  map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *fakeCanvasHandlerService) CancelRun(_ context.Context, userID int64, runID int64) (*service.CanvasRun, error) {
	s.lastCancelUserID = userID
	return &service.CanvasRun{
		ID:        runID,
		UserID:    userID,
		CanvasID:  12,
		Status:    service.CanvasRunStatusCanceled,
		Input:     map[string]any{},
		Output:    map[string]any{},
		Metadata:  map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *fakeCanvasHandlerService) ListModels(_ context.Context, userID int64) (service.CanvasModelCatalog, error) {
	if userID <= 0 {
		return service.CanvasModelCatalog{}, nil
	}
	return service.DefaultCanvasModelCatalog(), nil
}

func canvasTestRouter(h *CanvasHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	})
	router.GET("/canvases", h.ListCanvases)
	router.POST("/canvases", h.SaveCanvas)
	router.PUT("/canvases/:id", h.UpdateCanvas)
	router.GET("/canvas-runs", h.ListRuns)
	router.POST("/canvas-runs", h.CreateRun)
	router.GET("/canvas-runs/:id", h.GetRun)
	router.POST("/canvas-runs/:id/cancel", h.CancelRun)
	router.GET("/canvas/models", h.ListModels)
	return router
}

func decodeCanvasHandlerResponse(t *testing.T, recorder *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var envelope response.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	return envelope
}

func TestCanvasHandlerListAndSaveUseAuthenticatedUser(t *testing.T) {
	fake := &fakeCanvasHandlerService{}
	router := canvasTestRouter(&CanvasHandler{svc: fake})

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/canvases?limit=10&offset=0", nil))
	require.Equal(t, http.StatusOK, listRecorder.Code)
	require.Equal(t, int64(42), fake.lastListUserID)
	listEnvelope := decodeCanvasHandlerResponse(t, listRecorder)
	listData := listEnvelope.Data.(map[string]any)
	require.Len(t, listData["items"].([]any), 1)

	saveRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/canvases", strings.NewReader(`{
		"title": "Campaign",
		"nodes": [{"id": "prompt", "type": "prompt", "position": {}, "data": {}}],
		"edges": []
	}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(saveRecorder, req)

	require.Equal(t, http.StatusOK, saveRecorder.Code)
	require.Equal(t, int64(42), fake.lastSaveUserID)
	require.Equal(t, "Campaign", fake.lastSaveInput.Title)
	saveEnvelope := decodeCanvasHandlerResponse(t, saveRecorder)
	saveData := saveEnvelope.Data.(map[string]any)
	require.Equal(t, "Campaign", saveData["item"].(map[string]any)["title"])
}

func TestCanvasHandlerUpdateOverridesPathID(t *testing.T) {
	fake := &fakeCanvasHandlerService{}
	router := canvasTestRouter(&CanvasHandler{svc: fake})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/canvases/12", strings.NewReader(`{"id": 99, "title": "Updated"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(12), fake.lastSaveInput.ID)
}

func TestCanvasHandlerRunRoutesUseAuthenticatedUser(t *testing.T) {
	fake := &fakeCanvasHandlerService{}
	router := canvasTestRouter(&CanvasHandler{svc: fake})

	createRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/canvas-runs", strings.NewReader(`{"canvas_id": 12, "model": "gpt-image-2"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, req)

	require.Equal(t, http.StatusOK, createRecorder.Code)
	require.Equal(t, int64(42), fake.lastRunUserID)
	require.Equal(t, int64(12), fake.lastRunInput.CanvasID)

	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/canvas-runs/30", nil))
	require.Equal(t, http.StatusOK, getRecorder.Code)

	cancelRecorder := httptest.NewRecorder()
	router.ServeHTTP(cancelRecorder, httptest.NewRequest(http.MethodPost, "/canvas-runs/30/cancel", nil))
	require.Equal(t, http.StatusOK, cancelRecorder.Code)
	require.Equal(t, int64(42), fake.lastCancelUserID)
}

func TestCanvasHandlerListModelsReturnsCatalog(t *testing.T) {
	router := canvasTestRouter(&CanvasHandler{svc: &fakeCanvasHandlerService{}})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/canvas/models", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	envelope := decodeCanvasHandlerResponse(t, recorder)
	data := envelope.Data.(map[string]any)
	require.Equal(t, "canvas_model_catalog", data["object"])
	require.NotEmpty(t, data["items"])
}
