package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type imageTaskMemoryStore struct {
	task    *ImageTaskRecord
	ttl     time.Duration
	saveErr error
	getErr  error
}

type imageTaskResultStorage struct{}

func (imageTaskResultStorage) Save(_ context.Context, key, _ string, _ []byte) (string, error) {
	return "https://example.test/" + key, nil
}

type failingImageTaskResultStorage struct{}

func (failingImageTaskResultStorage) Save(_ context.Context, _ string, _ string, _ []byte) (string, error) {
	return "", errors.New("object storage unavailable")
}

func newImageTaskTestService(store ImageTaskStore) *ImageTaskService {
	uploader := NewImageResultUploader(imageTaskResultStorage{}, "images/", 1024, nil)
	return NewImageTaskServiceWithUploader(store, uploader, time.Hour, 10*time.Minute)
}

func (s *imageTaskMemoryStore) Save(_ context.Context, task *ImageTaskRecord, ttl time.Duration) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	copy := *task
	s.task = &copy
	s.ttl = ttl
	return nil
}

func (s *imageTaskMemoryStore) Get(_ context.Context, _ string) (*ImageTaskRecord, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.task == nil {
		return nil, ErrImageTaskNotFound
	}
	copy := *s.task
	return &copy, nil
}

func TestImageTaskServiceLifecycleAndOwnership(t *testing.T) {
	store := &imageTaskMemoryStore{}
	svc := newImageTaskTestService(store)
	owner := ImageTaskOwner{UserID: 7, APIKeyID: 9}

	created, err := svc.Create(context.Background(), owner)
	require.NoError(t, err)
	require.Equal(t, ImageTaskStatusProcessing, created.Status)
	require.Equal(t, created.ID, created.TaskID)
	require.Equal(t, "image.generation.task", created.Object)
	require.Equal(t, time.Hour, store.ttl)
	require.Equal(t, owner.UserID, store.task.UserID)
	require.Equal(t, owner.APIKeyID, store.task.APIKeyID)

	_, err = svc.Get(context.Background(), ImageTaskOwner{UserID: 7, APIKeyID: 10}, created.ID)
	require.ErrorIs(t, err, ErrImageTaskNotFound)

	result := json.RawMessage(`{"created":123,"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9JYlUAAAAASUVORK5CYII="}]}`)
	require.NoError(t, svc.Complete(context.Background(), created.ID, http.StatusOK, result))

	completed, err := svc.Get(context.Background(), owner, created.ID)
	require.NoError(t, err)
	require.Equal(t, ImageTaskStatusCompleted, completed.Status)
	require.Equal(t, http.StatusOK, completed.HTTPStatus)
	require.Contains(t, completed.ImageURL, "https://example.test/images/")
	require.NotContains(t, string(completed.Result), "b64_json")
	require.NotNil(t, completed.CompletedAt)
}

func TestImageTaskServiceInvalidResultBecomesFailed(t *testing.T) {
	store := &imageTaskMemoryStore{}
	svc := newImageTaskTestService(store)
	created, err := svc.Create(context.Background(), ImageTaskOwner{UserID: 1, APIKeyID: 2})
	require.NoError(t, err)

	require.NoError(t, svc.Complete(context.Background(), created.ID, http.StatusOK, json.RawMessage(`not-json`)))
	got, err := svc.Get(context.Background(), ImageTaskOwner{UserID: 1, APIKeyID: 2}, created.ID)
	require.NoError(t, err)
	require.Equal(t, ImageTaskStatusFailed, got.Status)
	require.Equal(t, http.StatusBadGateway, got.HTTPStatus)
	require.Contains(t, string(got.Error), "non-JSON")
}

func TestImageTaskServiceMapsStoreFailures(t *testing.T) {
	store := &imageTaskMemoryStore{saveErr: errors.New("redis down")}
	svc := newImageTaskTestService(store)

	_, err := svc.Create(context.Background(), ImageTaskOwner{UserID: 1, APIKeyID: 2})
	require.ErrorIs(t, err, ErrImageTaskUnavailable)
}

func TestImageTaskServiceOffloadFailureBecomesFailed(t *testing.T) {
	store := &imageTaskMemoryStore{}
	uploader := NewImageResultUploader(failingImageTaskResultStorage{}, "images/", 1024, nil)
	svc := NewImageTaskServiceWithUploader(store, uploader, time.Hour, time.Minute)
	owner := ImageTaskOwner{UserID: 1, APIKeyID: 2}
	created, err := svc.Create(context.Background(), owner)
	require.NoError(t, err)

	result := json.RawMessage(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9JYlUAAAAASUVORK5CYII="}]}`)
	require.NoError(t, svc.Complete(context.Background(), created.ID, http.StatusOK, result))

	got, err := svc.Get(context.Background(), owner, created.ID)
	require.NoError(t, err)
	require.Equal(t, ImageTaskStatusFailed, got.Status)
	require.Equal(t, http.StatusBadGateway, got.HTTPStatus)
	require.Empty(t, got.Result)
	require.Contains(t, string(got.Error), "failed to store generated image")
}

func TestImageTaskServiceDisableBlocksNewTasksButKeepsPolling(t *testing.T) {
	store := &imageTaskMemoryStore{}
	uploader := NewImageResultUploader(imageTaskResultStorage{}, "images/", 1024, nil)
	enabled := true
	svc := NewImageTaskServiceWithResolver(store, func() (*ImageResultUploader, bool) {
		return uploader, enabled
	}, time.Hour, time.Minute)
	owner := ImageTaskOwner{UserID: 1, APIKeyID: 2}

	created, err := svc.Create(context.Background(), owner)
	require.NoError(t, err)
	enabled = false

	_, err = svc.Create(context.Background(), owner)
	require.ErrorIs(t, err, ErrImageTaskUnavailable)
	require.True(t, svc.Pollable())
	got, err := svc.Get(context.Background(), owner, created.ID)
	require.NoError(t, err)
	require.Equal(t, ImageTaskStatusProcessing, got.Status)
}

func TestImageResultUploaderRejectsUnsafeDownloadURLs(t *testing.T) {
	uploader := NewImageResultUploader(imageTaskResultStorage{}, "images/", 1024, nil)
	for _, rawURL := range []string{
		"http://example.com/image.png",
		"https://127.0.0.1/image.png",
		"https://[::1]/image.png",
		"https://localhost/image.png",
	} {
		_, _, err := uploader.download(context.Background(), rawURL)
		require.Error(t, err, rawURL)
	}
}
