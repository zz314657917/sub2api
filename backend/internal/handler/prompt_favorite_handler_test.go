package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakePromptFavoriteHandlerService struct {
	items        []service.PromptFavorite
	lastUserID   int64
	lastInput    service.PromptFavoriteInput
	deletedID    int64
	deleteUserID int64
}

func (s *fakePromptFavoriteHandlerService) List(_ context.Context, userID int64) ([]service.PromptFavorite, error) {
	s.lastUserID = userID
	return append([]service.PromptFavorite(nil), s.items...), nil
}

func (s *fakePromptFavoriteHandlerService) Save(_ context.Context, userID int64, input service.PromptFavoriteInput) (*service.PromptFavorite, error) {
	s.lastUserID = userID
	s.lastInput = input
	item := service.PromptFavorite{
		ID:       12,
		UserID:   userID,
		PromptID: input.PromptID,
		Source:   input.Source,
		Title:    input.Title,
		Prompt:   input.Prompt,
	}
	s.items = append([]service.PromptFavorite{item}, s.items...)
	return &item, nil
}

func (s *fakePromptFavoriteHandlerService) Delete(_ context.Context, userID int64, favoriteID int64) error {
	s.deleteUserID = userID
	s.deletedID = favoriteID
	if len(s.items) > 0 {
		s.items = s.items[:0]
	}
	return nil
}

func promptFavoriteTestRouter(h *PromptFavoriteHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	})
	router.GET("/prompt-favorites", h.List)
	router.POST("/prompt-favorites", h.Save)
	router.DELETE("/prompt-favorites/:id", h.Delete)
	return router
}

func decodePromptFavoriteResponse(t *testing.T, recorder *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var envelope response.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	return envelope
}

func TestPromptFavoriteHandlerListUsesAuthenticatedUser(t *testing.T) {
	fake := &fakePromptFavoriteHandlerService{items: []service.PromptFavorite{
		{ID: 12, UserID: 42, PromptID: "p1", Source: "banana", Title: "Poster", Prompt: "draw"},
	}}
	router := promptFavoriteTestRouter(&PromptFavoriteHandler{svc: fake})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/prompt-favorites", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(42), fake.lastUserID)
	envelope := decodePromptFavoriteResponse(t, recorder)
	data := envelope.Data.(map[string]any)
	items := data["items"].([]any)
	require.Len(t, items, 1)
	require.Equal(t, "Poster", items[0].(map[string]any)["title"])
}

func TestPromptFavoriteHandlerSaveReturnsUpdatedItems(t *testing.T) {
	fake := &fakePromptFavoriteHandlerService{}
	router := promptFavoriteTestRouter(&PromptFavoriteHandler{svc: fake})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/prompt-favorites", strings.NewReader(`{
		"prompt_id": "p1",
		"source": "banana-prompt-quicker",
		"title": "Poster",
		"prompt": "draw poster",
		"mode": "generate"
	}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(42), fake.lastUserID)
	require.Equal(t, "p1", fake.lastInput.PromptID)
	envelope := decodePromptFavoriteResponse(t, recorder)
	data := envelope.Data.(map[string]any)
	require.Equal(t, "Poster", data["item"].(map[string]any)["title"])
	require.Len(t, data["items"].([]any), 1)
}

func TestPromptFavoriteHandlerDeleteUsesAuthenticatedUser(t *testing.T) {
	fake := &fakePromptFavoriteHandlerService{items: []service.PromptFavorite{
		{ID: 12, UserID: 42, PromptID: "p1", Source: "banana", Title: "Poster", Prompt: "draw"},
	}}
	router := promptFavoriteTestRouter(&PromptFavoriteHandler{svc: fake})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/prompt-favorites/12", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(42), fake.deleteUserID)
	require.Equal(t, int64(12), fake.deletedID)
	envelope := decodePromptFavoriteResponse(t, recorder)
	data := envelope.Data.(map[string]any)
	require.Empty(t, data["items"].([]any))
}
