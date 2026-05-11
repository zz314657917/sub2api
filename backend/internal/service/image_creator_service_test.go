package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type fakeImageCreatorRepo struct {
	tasks             map[int64]*ImageCreatorTask
	images            []ImageCreatorImage
	nextTaskID        int64
	nextImageID       int64
	prunedImagePaths  []string
	expiredImagePaths []string
	expiredRefPaths   []string
	failStaleCalls    int
	failStaleMessage  string
}

func newFakeImageCreatorRepo() *fakeImageCreatorRepo {
	return &fakeImageCreatorRepo{
		tasks:       make(map[int64]*ImageCreatorTask),
		nextTaskID:  1,
		nextImageID: 1,
	}
}

func (r *fakeImageCreatorRepo) CreateTask(_ context.Context, task *ImageCreatorTask) error {
	task.ID = r.nextTaskID
	r.nextTaskID++
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	cp := *task
	r.tasks[task.ID] = &cp
	return nil
}

func (r *fakeImageCreatorRepo) UpdateTaskReferenceImage(_ context.Context, taskID int64, path string, mimeType string, filename string) error {
	task := r.tasks[taskID]
	if task == nil {
		return errors.New("task not found")
	}
	task.ReferenceImagePath = path
	task.ReferenceImageMimeType = mimeType
	task.ReferenceImageFilename = filename
	return nil
}

func (r *fakeImageCreatorRepo) GetTaskByID(_ context.Context, taskID int64) (*ImageCreatorTask, error) {
	task := r.tasks[taskID]
	if task == nil {
		return nil, errors.New("task not found")
	}
	cp := *task
	cp.Images = r.imagesForTask(taskID)
	return &cp, nil
}

func (r *fakeImageCreatorRepo) GetTaskForUser(_ context.Context, userID int64, taskID int64) (*ImageCreatorTask, error) {
	task := r.tasks[taskID]
	if task == nil || task.UserID != userID {
		return nil, errors.New("task not found")
	}
	cp := *task
	cp.Images = r.imagesForTask(taskID)
	return &cp, nil
}

func (r *fakeImageCreatorRepo) ListTasksForUser(_ context.Context, userID int64, limit int) ([]ImageCreatorTask, error) {
	out := make([]ImageCreatorTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.UserID != userID {
			continue
		}
		cp := *task
		cp.Images = r.imagesForTask(task.ID)
		out = append(out, cp)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *fakeImageCreatorRepo) GetImageForUser(_ context.Context, userID int64, imageID int64) (*ImageCreatorImage, error) {
	for i := range r.images {
		if r.images[i].ID == imageID && r.images[i].UserID == userID {
			cp := r.images[i]
			return &cp, nil
		}
	}
	return nil, errors.New("image not found")
}

func (r *fakeImageCreatorRepo) ClaimNextPendingTask(_ context.Context, _ time.Duration) (*ImageCreatorTask, error) {
	for _, task := range r.tasks {
		if task.Status == ImageCreatorTaskStatusPending {
			task.Status = ImageCreatorTaskStatusRunning
			now := time.Now()
			task.StartedAt = &now
			cp := *task
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeImageCreatorRepo) MarkTaskRunning(_ context.Context, taskID int64) error {
	task := r.tasks[taskID]
	if task == nil {
		return errors.New("task not found")
	}
	task.Status = ImageCreatorTaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	return nil
}

func (r *fakeImageCreatorRepo) MarkTaskSucceeded(_ context.Context, taskID int64, warning string) error {
	task := r.tasks[taskID]
	if task == nil {
		return errors.New("task not found")
	}
	task.Status = ImageCreatorTaskStatusSucceeded
	task.ErrorMessage = warning
	now := time.Now()
	task.CompletedAt = &now
	return nil
}

func (r *fakeImageCreatorRepo) MarkTaskFailed(_ context.Context, taskID int64, message string) error {
	task := r.tasks[taskID]
	if task == nil {
		return errors.New("task not found")
	}
	task.Status = ImageCreatorTaskStatusFailed
	task.ErrorMessage = message
	now := time.Now()
	task.CompletedAt = &now
	return nil
}

func (r *fakeImageCreatorRepo) AddImage(_ context.Context, image *ImageCreatorImage) error {
	image.ID = r.nextImageID
	r.nextImageID++
	now := time.Now()
	image.CreatedAt = now
	cp := *image
	r.images = append(r.images, cp)
	return nil
}

func (r *fakeImageCreatorRepo) ListPrunableImages(_ context.Context, userID int64, keep int) ([]ImageCreatorImage, error) {
	var userImages []ImageCreatorImage
	for _, image := range r.images {
		if image.UserID == userID {
			userImages = append(userImages, image)
		}
	}
	if len(userImages) <= keep {
		return nil, nil
	}
	return append([]ImageCreatorImage(nil), userImages[:len(userImages)-keep]...), nil
}

func (r *fakeImageCreatorRepo) DeleteImagesByID(_ context.Context, ids []int64) error {
	deleteSet := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		deleteSet[id] = struct{}{}
	}
	out := r.images[:0]
	for _, image := range r.images {
		if _, ok := deleteSet[image.ID]; ok {
			r.prunedImagePaths = append(r.prunedImagePaths, image.FilePath)
			continue
		}
		out = append(out, image)
	}
	r.images = out
	return nil
}

func (r *fakeImageCreatorRepo) ListExpiredImages(_ context.Context, _ time.Time, _ int) ([]ImageCreatorImage, error) {
	out := make([]ImageCreatorImage, 0)
	for _, image := range r.images {
		if image.ExpiresAt.Before(time.Now()) {
			out = append(out, image)
		}
	}
	return out, nil
}

func (r *fakeImageCreatorRepo) DeleteExpiredTasks(_ context.Context, _ time.Time, _ int) ([]ImageCreatorTask, error) {
	var expired []ImageCreatorTask
	for _, task := range r.tasks {
		if !task.ExpiresAt.Before(time.Now()) {
			continue
		}
		cp := *task
		cp.Images = r.imagesForTask(task.ID)
		expired = append(expired, cp)
		r.expiredRefPaths = append(r.expiredRefPaths, task.ReferenceImagePath)
		delete(r.tasks, task.ID)
	}
	return expired, nil
}

func (r *fakeImageCreatorRepo) FailStaleRunningTasks(_ context.Context, staleRunningAfter time.Duration, message string) error {
	r.failStaleCalls++
	r.failStaleMessage = message
	cutoff := time.Now().Add(-staleRunningAfter)
	for _, task := range r.tasks {
		if task.Status != ImageCreatorTaskStatusRunning || task.StartedAt == nil || !task.StartedAt.Before(cutoff) {
			continue
		}
		task.Status = ImageCreatorTaskStatusFailed
		task.ErrorMessage = message
		now := time.Now()
		task.CompletedAt = &now
	}
	return nil
}

func (r *fakeImageCreatorRepo) imagesForTask(taskID int64) []ImageCreatorImage {
	out := make([]ImageCreatorImage, 0)
	for _, image := range r.images {
		if image.TaskID == taskID {
			out = append(out, image)
		}
	}
	return out
}

type fakeAPIKeyLookup struct {
	key *APIKey
}

func (f fakeAPIKeyLookup) GetByID(_ context.Context, _ int64) (*APIKey, error) {
	if f.key == nil {
		return nil, errors.New("key not found")
	}
	return f.key, nil
}

type fakeImageGenerator struct {
	results []GeneratedImageAsset
	err     error
	calls   int
	inputs  []ImageCreatorGenerateRequest
}

func (g *fakeImageGenerator) GenerateImage(_ context.Context, input ImageCreatorGenerateRequest, _ string) ([]GeneratedImageAsset, error) {
	g.calls++
	g.inputs = append(g.inputs, input)
	if g.err != nil {
		return nil, g.err
	}
	if len(g.results) == 0 {
		return nil, errors.New("no fake result")
	}
	result := g.results[(g.calls-1)%len(g.results)]
	return []GeneratedImageAsset{result}, nil
}

func TestImageCreatorServiceCreateTaskRejectsForeignAPIKey(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 99,
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	_, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "a neon city",
		Count:        1,
		OutputFormat: "png",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err), "expected forbidden error, got %v", err)
	require.Empty(t, repo.tasks)
}

func TestImageCreatorServiceCreateTaskRejectsDisabledImageGenerationGroup(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: false},
	}}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	_, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "a neon city",
		Count:        1,
		OutputFormat: "png",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsForbidden(err), "expected forbidden error, got %v", err)
	require.Empty(t, repo.tasks)
}

func TestImageCreatorServiceCreateTaskRejectsTooManyImages(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	_, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "too many images",
		Count:        9,
		OutputFormat: "png",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err), "expected bad request error, got %v", err)
	require.Equal(t, "INVALID_COUNT", infraerrors.Reason(err))
	require.Empty(t, repo.tasks)
}

func TestImageCreatorServiceCreateTaskInfersMultiAngleCountFromPrompt(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, &fakeImageGenerator{}, ImageCreatorServiceOptions{
		DisableAsyncOnCreate: true,
	})

	task, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "生成特朗普跳河的四个角度的图片",
		Count:        1,
		OutputFormat: "png",
	})

	require.NoError(t, err)
	require.Equal(t, 4, task.Count)
	require.Equal(t, 4, repo.tasks[task.ID].Count)
}

func TestImageCreatorServiceProcessTaskSplitsMultiAnglePrompt(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{
		{Data: []byte("front"), OutputFormat: "png"},
		{Data: []byte("side"), OutputFormat: "png"},
		{Data: []byte("rear"), OutputFormat: "png"},
		{Data: []byte("overhead"), OutputFormat: "png"},
	}}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Key:    "sk-test",
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, generator, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       8,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	task, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "生成特朗普跳河的四个角度的图片",
		Count:        4,
		OutputFormat: "png",
	})
	require.NoError(t, err)

	require.NoError(t, svc.ProcessTask(context.Background(), task.ID))

	require.Equal(t, 4, generator.calls)
	require.Len(t, generator.inputs, 4)
	require.Contains(t, generator.inputs[0].Prompt, "正面视角")
	require.Contains(t, generator.inputs[1].Prompt, "侧面视角")
	require.Contains(t, generator.inputs[2].Prompt, "背面视角")
	require.Contains(t, generator.inputs[3].Prompt, "俯视远景")
	for _, input := range generator.inputs {
		require.Contains(t, input.Prompt, "只生成一张完整图片")
		require.Contains(t, input.Prompt, "不要四宫格")
	}
	require.NotEqual(t, generator.inputs[0].Prompt, generator.inputs[1].Prompt)
}

func TestImageCreatorServiceProcessTaskPersistsOnlyLatestThreeImages(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{
		{Data: []byte("one"), OutputFormat: "png"},
		{Data: []byte("two"), OutputFormat: "png"},
		{Data: []byte("three"), OutputFormat: "png"},
		{Data: []byte("four"), OutputFormat: "png"},
	}}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Key:    "sk-test",
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, generator, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	task, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "four versions",
		Count:        4,
		OutputFormat: "png",
	})
	require.NoError(t, err)

	require.NoError(t, svc.ProcessTask(context.Background(), task.ID))

	require.Equal(t, 4, generator.calls)
	require.Len(t, repo.images, 3)
	require.Equal(t, []byte("two"), mustReadFile(t, repo.images[0].FilePath))
	require.Equal(t, []byte("three"), mustReadFile(t, repo.images[1].FilePath))
	require.Equal(t, []byte("four"), mustReadFile(t, repo.images[2].FilePath))
	require.NoFileExists(t, filepath.Join(dir, "42", "1-1.png"))
	require.Equal(t, ImageCreatorTaskStatusSucceeded, repo.tasks[task.ID].Status)
}

func TestImageCreatorServiceProcessTaskRevalidatesAPIKeyPermission(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	key := &APIKey{
		ID:     10,
		UserID: 42,
		Key:    "sk-test",
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{{Data: []byte("one"), OutputFormat: "png"}}}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: key}, generator, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	task, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "one version",
		Count:        1,
		OutputFormat: "png",
	})
	require.NoError(t, err)

	key.Group.AllowImageGeneration = false
	err = svc.ProcessTask(context.Background(), task.ID)

	require.Error(t, err)
	require.Zero(t, generator.calls)
	require.Empty(t, repo.images)
	require.Equal(t, ImageCreatorTaskStatusFailed, repo.tasks[task.ID].Status)
	require.Contains(t, repo.tasks[task.ID].ErrorMessage, ImageGenerationPermissionMessage())
}

func TestImageCreatorServiceRunWorkerOnceFailsStaleRunningTasks(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	startedAt := time.Now().Add(-2 * time.Hour)
	repo.tasks[1] = &ImageCreatorTask{
		ID:        1,
		UserID:    42,
		APIKeyID:  10,
		Status:    ImageCreatorTaskStatusRunning,
		StartedAt: &startedAt,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{
		StorageDir:           t.TempDir(),
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		TaskTimeout:          time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	svc.runWorkerOnce()

	require.Equal(t, 1, repo.failStaleCalls)
	require.Equal(t, "server restarted while image generation was running", repo.failStaleMessage)
	require.Equal(t, ImageCreatorTaskStatusFailed, repo.tasks[1].Status)
	require.Equal(t, "server restarted while image generation was running", repo.tasks[1].ErrorMessage)
}

func TestImageCreatorServiceCleanupExpiredDeletesImageAndReferenceFiles(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	expiredImage := filepath.Join(dir, "expired.png")
	expiredRef := filepath.Join(dir, "ref.png")
	require.NoError(t, os.WriteFile(expiredImage, []byte("old image"), 0o600))
	require.NoError(t, os.WriteFile(expiredRef, []byte("old ref"), 0o600))
	repo.tasks[1] = &ImageCreatorTask{
		ID:                 1,
		UserID:             42,
		Status:             ImageCreatorTaskStatusSucceeded,
		ReferenceImagePath: expiredRef,
		ExpiresAt:          time.Now().Add(-time.Hour),
	}
	repo.images = append(repo.images, ImageCreatorImage{
		ID:        1,
		TaskID:    1,
		UserID:    42,
		FilePath:  expiredImage,
		ExpiresAt: time.Now().Add(-time.Hour),
	})
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	require.NoError(t, svc.CleanupExpired(context.Background()))

	require.NoFileExists(t, expiredImage)
	require.NoFileExists(t, expiredRef)
	require.Empty(t, repo.images)
	require.Empty(t, repo.tasks)
}

func TestImageCreatorServiceCleanupExpiredTaskDeletesCascadedImageFiles(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	imagePath := filepath.Join(dir, "task-image.png")
	require.NoError(t, os.WriteFile(imagePath, []byte("task image"), 0o600))
	repo.tasks[1] = &ImageCreatorTask{
		ID:        1,
		UserID:    42,
		Status:    ImageCreatorTaskStatusSucceeded,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	repo.images = append(repo.images, ImageCreatorImage{
		ID:        1,
		TaskID:    1,
		UserID:    42,
		FilePath:  imagePath,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	require.NoError(t, svc.CleanupExpired(context.Background()))

	require.NoFileExists(t, imagePath)
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}
