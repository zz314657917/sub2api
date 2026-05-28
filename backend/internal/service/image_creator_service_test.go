package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func (r *fakeImageCreatorRepo) CreateTask(_ context.Context, task *ImageCreatorTask, maxActiveTasks int) error {
	if maxActiveTasks <= 0 {
		maxActiveTasks = 1
	}
	active := 0
	for _, existing := range r.tasks {
		if existing.UserID == task.UserID && (existing.Status == ImageCreatorTaskStatusPending || existing.Status == ImageCreatorTaskStatusRunning) {
			active++
			if active >= maxActiveTasks {
				return ErrImageCreatorActiveTaskExists
			}
		}
	}
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

func (r *fakeImageCreatorRepo) ListImagesForUser(_ context.Context, userID int64, limit int, offset int) ([]ImageCreatorManagedImage, int, error) {
	var userImages []ImageCreatorManagedImage
	for _, image := range r.images {
		if image.UserID != userID {
			continue
		}
		item := ImageCreatorManagedImage{ImageCreatorImage: image}
		if task := r.tasks[image.TaskID]; task != nil {
			item.TaskPrompt = task.Prompt
			item.TaskModel = task.Model
			item.TaskSize = task.Size
			item.TaskQuality = task.Quality
		}
		userImages = append(userImages, item)
	}
	total := len(userImages)
	if offset >= total {
		return []ImageCreatorManagedImage{}, total, nil
	}
	if limit <= 0 || offset+limit > total {
		limit = total - offset
	}
	return append([]ImageCreatorManagedImage(nil), userImages[offset:offset+limit]...), total, nil
}

func (r *fakeImageCreatorRepo) ListImagesForUserByIDs(_ context.Context, userID int64, ids []int64) ([]ImageCreatorImage, error) {
	idSet := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var out []ImageCreatorImage
	for _, image := range r.images {
		if image.UserID != userID {
			continue
		}
		if _, ok := idSet[image.ID]; ok {
			out = append(out, image)
		}
	}
	return out, nil
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
	results     []GeneratedImageAsset
	err         error
	calls       int
	inputs      []ImageCreatorGenerateRequest
	resultIndex int
}

type fakeImageCreatorObjectStore struct {
	objects map[string][]byte
	deleted []string
}

func newFakeImageCreatorObjectStore() *fakeImageCreatorObjectStore {
	return &fakeImageCreatorObjectStore{objects: make(map[string][]byte)}
}

func (s *fakeImageCreatorObjectStore) Upload(_ context.Context, key string, body io.Reader, _ string) (int64, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, err
	}
	s.objects[key] = append([]byte(nil), data...)
	return int64(len(data)), nil
}

func (s *fakeImageCreatorObjectStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	data, ok := s.objects[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *fakeImageCreatorObjectStore) Delete(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	delete(s.objects, key)
	return nil
}

func (s *fakeImageCreatorObjectStore) PresignURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}

func (s *fakeImageCreatorObjectStore) HeadBucket(_ context.Context) error {
	return nil
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
	count := maxInt(1, input.Count)
	assets := make([]GeneratedImageAsset, 0, count)
	for i := 0; i < count; i++ {
		assets = append(assets, g.results[g.resultIndex%len(g.results)])
		g.resultIndex++
	}
	return assets, nil
}

func TestDecodeImageCreatorBase64AcceptsDataURLWhitespaceAndURLSafeInput(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []byte
	}{
		{
			name: "data url with whitespace",
			raw:  "data:image/png;base64, aG Vs\n bG8=",
			want: []byte("hello"),
		},
		{
			name: "url safe without padding",
			raw:  "-_8",
			want: []byte{0xfb, 0xff},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeImageCreatorBase64(tc.raw)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestImageCreatorServiceCreateTaskNormalizesUnsupportedTransparentBackground(t *testing.T) {
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
		Model:        "gpt-image-1.5",
		Prompt:       "a neon city",
		Count:        1,
		OutputFormat: "png",
		Background:   "transparent",
	})

	require.NoError(t, err)
	require.Equal(t, "auto", task.Background)
	require.Equal(t, "auto", repo.tasks[task.ID].Background)
}

func TestImageCreatorServiceCreateTaskDefaultsRequestedOutputFormatToWebP(t *testing.T) {
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
		Prompt:       "a compressed image",
		Count:        1,
		OutputFormat: "png",
	})

	require.NoError(t, err)
	require.Equal(t, "webp", task.OutputFormat)
	require.Equal(t, "webp", repo.tasks[task.ID].OutputFormat)
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

func TestImageCreatorServiceCreateTaskAllowsEightImages(t *testing.T) {
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
		Prompt:       "eight images",
		Count:        8,
		OutputFormat: "png",
	})

	require.NoError(t, err)
	require.Equal(t, 8, task.Count)
	require.Equal(t, 8, repo.tasks[task.ID].Count)
}

func TestImageCreatorServiceCreateTaskRejectsWhenUserReachedActiveTaskLimit(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	repo.tasks[1] = &ImageCreatorTask{
		ID:        1,
		UserID:    42,
		APIKeyID:  10,
		Status:    ImageCreatorTaskStatusRunning,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	repo.tasks[2] = &ImageCreatorTask{
		ID:        2,
		UserID:    42,
		APIKeyID:  10,
		Status:    ImageCreatorTaskStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	_, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "wait for current task",
		Count:        1,
		OutputFormat: "png",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsTooManyRequests(err), "expected too many requests error, got %v", err)
	require.Equal(t, "IMAGE_CREATOR_TASK_LIMIT_EXCEEDED", infraerrors.Reason(err))
	require.Len(t, repo.tasks, 2)
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
		MaxSavedImages:       4,
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

func TestImageCreatorServiceProcessTaskKeepsAllImagesFromCurrentTask(t *testing.T) {
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

	require.Equal(t, 1, generator.calls)
	require.Len(t, generator.inputs, 1)
	require.Equal(t, 4, generator.inputs[0].Count)
	require.Len(t, repo.images, 4)
	require.Equal(t, []byte("one"), mustReadFile(t, repo.images[0].FilePath))
	require.Equal(t, []byte("two"), mustReadFile(t, repo.images[1].FilePath))
	require.Equal(t, []byte("three"), mustReadFile(t, repo.images[2].FilePath))
	require.Equal(t, []byte("four"), mustReadFile(t, repo.images[3].FilePath))
	require.FileExists(t, filepath.Join(dir, "42", "1-1.png"))
	require.Equal(t, ImageCreatorTaskStatusSucceeded, repo.tasks[task.ID].Status)
}

func TestImageCreatorServiceProcessTaskKeepsFourImageBatchWhenSavedLimitIsLower(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{
		{Data: []byte("image-1"), OutputFormat: "png"},
		{Data: []byte("image-2"), OutputFormat: "png"},
		{Data: []byte("image-3"), OutputFormat: "png"},
		{Data: []byte("image-4"), OutputFormat: "png"},
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

	require.Equal(t, 1, generator.calls)
	require.Len(t, generator.inputs, 1)
	require.Equal(t, 4, generator.inputs[0].Count)
	require.Len(t, repo.images, 4)
	expected := [][]byte{[]byte("image-1"), []byte("image-2"), []byte("image-3"), []byte("image-4")}
	for i, image := range repo.images {
		require.Equal(t, expected[i], mustReadFile(t, image.FilePath))
	}
	require.Equal(t, ImageCreatorTaskStatusSucceeded, repo.tasks[task.ID].Status)
}

func TestImageCreatorServiceProcessTaskBatchesEightImagesByFour(t *testing.T) {
	dir := t.TempDir()
	repo := newFakeImageCreatorRepo()
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{
		{Data: []byte("image-1"), OutputFormat: "png"},
		{Data: []byte("image-2"), OutputFormat: "png"},
		{Data: []byte("image-3"), OutputFormat: "png"},
		{Data: []byte("image-4"), OutputFormat: "png"},
		{Data: []byte("image-5"), OutputFormat: "png"},
		{Data: []byte("image-6"), OutputFormat: "png"},
		{Data: []byte("image-7"), OutputFormat: "png"},
		{Data: []byte("image-8"), OutputFormat: "png"},
	}}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Key:    "sk-test",
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, generator, ImageCreatorServiceOptions{
		StorageDir:           dir,
		MaxSavedImages:       4,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	task, err := svc.CreateTask(context.Background(), 42, ImageCreatorCreateTaskInput{
		APIKeyID:     10,
		Model:        "gpt-image-2",
		Prompt:       "eight versions",
		Count:        8,
		OutputFormat: "png",
	})
	require.NoError(t, err)

	require.NoError(t, svc.ProcessTask(context.Background(), task.ID))

	require.Equal(t, 2, generator.calls)
	require.Len(t, generator.inputs, 2)
	require.Equal(t, 4, generator.inputs[0].Count)
	require.Equal(t, 4, generator.inputs[1].Count)
	require.Len(t, repo.images, 8)
	for i, image := range repo.images {
		require.Equal(t, []byte(fmt.Sprintf("image-%d", i+1)), mustReadFile(t, image.FilePath))
	}
	require.Equal(t, ImageCreatorTaskStatusSucceeded, repo.tasks[task.ID].Status)
}

func TestImageCreatorServiceProcessTaskStoresGeneratedImagesInObjectStorage(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	store := newFakeImageCreatorObjectStore()
	generator := &fakeImageGenerator{results: []GeneratedImageAsset{{Data: []byte("object image"), OutputFormat: "png"}}}
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{key: &APIKey{
		ID:     10,
		UserID: 42,
		Key:    "sk-test",
		Status: StatusAPIKeyActive,
		Group:  &Group{Platform: PlatformOpenAI, Status: StatusActive, AllowImageGeneration: true},
	}}, generator, ImageCreatorServiceOptions{
		StorageDir:     t.TempDir(),
		StorageBackend: imageCreatorStorageBackendCOS,
		ObjectStorage:  &BackupS3Config{Bucket: "bucket-appid", AccessKeyID: "sid", SecretAccessKey: "secret", Prefix: "creator"},
		ObjectStoreFactory: func(context.Context, *BackupS3Config) (BackupObjectStore, error) {
			return store, nil
		},
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

	require.NoError(t, svc.ProcessTask(context.Background(), task.ID))

	require.Len(t, repo.images, 1)
	require.True(t, isImageCreatorObjectStoragePath(repo.images[0].FilePath))
	key := imageCreatorObjectStorageKey(repo.images[0].FilePath)
	require.Contains(t, key, "creator/42/")
	require.Equal(t, []byte("object image"), store.objects[key])
	file, err := svc.GetImageFile(context.Background(), 42, repo.images[0].ID)
	require.NoError(t, err)
	defer func() { _ = file.Body.Close() }()
	data, err := io.ReadAll(file.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("object image"), data)
	require.Equal(t, int64(len("object image")), file.SizeBytes)
}

func TestImageCreatorServiceAutoStorageBackendUsesObjectStorageWhenConfigured(t *testing.T) {
	opts := normalizeImageCreatorOptions(ImageCreatorServiceOptions{
		StorageBackend: "",
		ObjectStorage:  &BackupS3Config{Endpoint: "https://cos.example.com", Bucket: "bucket-appid", AccessKeyID: "sid", SecretAccessKey: "secret"},
	})

	require.Equal(t, imageCreatorStorageBackendCOS, opts.StorageBackend)
}

func TestImageCreatorServiceListImagesAttachesURLsAndTaskMetadata(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	repo.tasks[1] = &ImageCreatorTask{
		ID:      1,
		UserID:  42,
		Prompt:  "draw reusable image",
		Model:   "gpt-image-2",
		Size:    "1024x1024",
		Quality: "auto",
	}
	repo.images = append(repo.images, ImageCreatorImage{
		ID:           9,
		TaskID:       1,
		UserID:       42,
		OutputFormat: "png",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	images, total, err := svc.ListImages(context.Background(), 42, 20, 0)

	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, images, 1)
	require.Equal(t, "/api/v1/user/image-creator/images/9/file", images[0].URL)
	require.Equal(t, "draw reusable image", images[0].TaskPrompt)
	require.Equal(t, "gpt-image-2", images[0].TaskModel)
}

func TestImageCreatorServiceDeleteImagesOnlyDeletesCurrentUserImages(t *testing.T) {
	dir := t.TempDir()
	ownImage := filepath.Join(dir, "own.png")
	otherImage := filepath.Join(dir, "other.png")
	require.NoError(t, os.WriteFile(ownImage, []byte("own"), 0o600))
	require.NoError(t, os.WriteFile(otherImage, []byte("other"), 0o600))
	repo := newFakeImageCreatorRepo()
	repo.images = append(repo.images,
		ImageCreatorImage{ID: 9, TaskID: 1, UserID: 42, FilePath: ownImage, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)},
		ImageCreatorImage{ID: 10, TaskID: 2, UserID: 99, FilePath: otherImage, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)},
	)
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{})

	deleted, err := svc.DeleteImages(context.Background(), 42, []int64{9, 10, 9})

	require.NoError(t, err)
	require.Equal(t, 1, deleted)
	require.NoFileExists(t, ownImage)
	require.FileExists(t, otherImage)
	require.Len(t, repo.images, 1)
	require.Equal(t, int64(10), repo.images[0].ID)
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

func TestImageCreatorServiceCleanupExpiredDeletesObjectStorageImages(t *testing.T) {
	repo := newFakeImageCreatorRepo()
	store := newFakeImageCreatorObjectStore()
	store.objects["creator/42/old.png"] = []byte("old object")
	repo.images = append(repo.images, ImageCreatorImage{
		ID:        1,
		TaskID:    1,
		UserID:    42,
		FilePath:  imageCreatorObjectStoragePathPrefix + "creator/42/old.png",
		ExpiresAt: time.Now().Add(-time.Hour),
	})
	svc := NewImageCreatorServiceWithDeps(repo, fakeAPIKeyLookup{}, &fakeImageGenerator{}, ImageCreatorServiceOptions{
		StorageDir:     t.TempDir(),
		StorageBackend: imageCreatorStorageBackendS3,
		ObjectStorage:  &BackupS3Config{Bucket: "bucket", AccessKeyID: "access", SecretAccessKey: "secret", Prefix: "creator"},
		ObjectStoreFactory: func(context.Context, *BackupS3Config) (BackupObjectStore, error) {
			return store, nil
		},
		MaxSavedImages:       3,
		Retention:            7 * 24 * time.Hour,
		AutoStartWorker:      false,
		DisableAsyncOnCreate: true,
	})

	require.NoError(t, svc.CleanupExpired(context.Background()))

	require.Empty(t, repo.images)
	require.NotContains(t, store.objects, "creator/42/old.png")
	require.Equal(t, []string{"creator/42/old.png"}, store.deleted)
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}
