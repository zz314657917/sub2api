package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeImageCreatorHandlerService struct {
	createUserID int64
	createInput  service.ImageCreatorCreateTaskInput
	task         *service.ImageCreatorTask
	tasks        []service.ImageCreatorTask
	file         *service.ImageCreatorFile
}

func (s *fakeImageCreatorHandlerService) CreateTask(_ context.Context, userID int64, input service.ImageCreatorCreateTaskInput) (*service.ImageCreatorTask, error) {
	s.createUserID = userID
	s.createInput = input
	if s.task != nil {
		return s.task, nil
	}
	return &service.ImageCreatorTask{ID: 123, UserID: userID, APIKeyID: input.APIKeyID, Status: service.ImageCreatorTaskStatusPending}, nil
}

func (s *fakeImageCreatorHandlerService) ListTasks(_ context.Context, _ int64, _ int) ([]service.ImageCreatorTask, error) {
	return s.tasks, nil
}

func (s *fakeImageCreatorHandlerService) GetTask(_ context.Context, userID int64, taskID int64) (*service.ImageCreatorTask, error) {
	if s.task != nil {
		return s.task, nil
	}
	return &service.ImageCreatorTask{ID: taskID, UserID: userID, Status: service.ImageCreatorTaskStatusRunning}, nil
}

func (s *fakeImageCreatorHandlerService) GetImageFile(_ context.Context, _ int64, _ int64) (*service.ImageCreatorFile, error) {
	return s.file, nil
}

func imageCreatorTestRouter(h *ImageCreatorHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	})
	router.POST("/tasks", h.CreateTask)
	router.GET("/tasks", h.ListTasks)
	router.GET("/tasks/:id", h.GetTask)
	router.GET("/images/:id/file", h.GetImageFile)
	return router
}

func decodeHandlerResponse(t *testing.T, recorder *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var envelope response.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	return envelope
}

func TestImageCreatorHandlerCreateTaskUsesAuthenticatedUserAndAPIKeyID(t *testing.T) {
	fake := &fakeImageCreatorHandlerService{}
	router := imageCreatorTestRouter(&ImageCreatorHandler{svc: fake})

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{
		"api_key_id": 10,
		"api_key": "sk-should-not-be-accepted",
		"model": "gpt-image-2",
		"prompt": "draw a durable task",
		"size": "1024x1024",
		"quality": "auto",
		"count": 4,
		"output_format": "png",
		"background": "auto"
	}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(42), fake.createUserID)
	require.Equal(t, int64(10), fake.createInput.APIKeyID)
	require.Equal(t, "draw a durable task", fake.createInput.Prompt)
	require.Equal(t, 4, fake.createInput.Count)
	require.Equal(t, "png", fake.createInput.OutputFormat)

	envelope := decodeHandlerResponse(t, recorder)
	require.Equal(t, 0, envelope.Code)
	data := envelope.Data.(map[string]any)
	require.Equal(t, float64(123), data["id"])
	require.Equal(t, service.ImageCreatorTaskStatusPending, data["status"])
}

func TestImageCreatorHandlerListTasksReturnsTasksAndFlattenedImages(t *testing.T) {
	expires := time.Now().Add(7 * 24 * time.Hour)
	fake := &fakeImageCreatorHandlerService{tasks: []service.ImageCreatorTask{
		{
			ID:     123,
			UserID: 42,
			Status: service.ImageCreatorTaskStatusSucceeded,
			Images: []service.ImageCreatorImage{
				{ID: 9, TaskID: 123, UserID: 42, URL: "/api/v1/user/image-creator/images/9/file", ExpiresAt: expires},
			},
		},
	}}
	router := imageCreatorTestRouter(&ImageCreatorHandler{svc: fake})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/tasks", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	envelope := decodeHandlerResponse(t, recorder)
	data := envelope.Data.(map[string]any)
	require.Len(t, data["tasks"].([]any), 1)
	images := data["images"].([]any)
	require.Len(t, images, 1)
	require.Equal(t, float64(9), images[0].(map[string]any)["id"])
}

func TestImageCreatorHandlerServesImageFileForAuthenticatedOwner(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "generated.png")
	require.NoError(t, os.WriteFile(path, []byte("pngdata"), 0o600))
	fake := &fakeImageCreatorHandlerService{file: &service.ImageCreatorFile{
		Path:        path,
		ContentType: "image/png",
		FileName:    "image-9.png",
	}}
	router := imageCreatorTestRouter(&ImageCreatorHandler{svc: fake})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/images/9/file", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "pngdata", recorder.Body.String())
	require.Contains(t, recorder.Header().Get("Content-Type"), "image/png")
}
