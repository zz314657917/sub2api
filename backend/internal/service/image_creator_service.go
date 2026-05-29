package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	ImageCreatorTaskStatusPending   = "pending"
	ImageCreatorTaskStatusRunning   = "running"
	ImageCreatorTaskStatusSucceeded = "succeeded"
	ImageCreatorTaskStatusFailed    = "failed"

	defaultImageCreatorStorageDir                = "data/image-creator"
	defaultImageCreatorMaxSavedImages            = 16
	defaultImageCreatorRetention                 = 7 * 24 * time.Hour
	defaultImageCreatorWorkerInterval            = 30 * time.Second
	defaultImageCreatorTaskTimeout               = 30 * time.Minute
	defaultImageCreatorRequestTimeout            = 30 * time.Minute
	defaultImageCreatorCleanupBatchSize          = 100
	defaultImageCreatorListTaskLimit             = 20
	defaultImageCreatorActiveTaskLimit           = 2
	defaultImageCreatorObjectStoragePrefix       = "image-creator"
	imageCreatorStorageBackendLocal              = "local"
	imageCreatorStorageBackendAuto               = "auto"
	imageCreatorStorageBackendS3                 = "s3"
	imageCreatorStorageBackendCOS                = "cos"
	imageCreatorObjectStoragePathPrefix          = "s3://"
	maxImageCreatorTaskCount                     = 8
	maxImageCreatorUpstreamBatchCount            = 4
	imageCreatorMaxStoredImageBytes        int64 = 25 << 20
)

var errImageCreatorTaskNotRunnable = errors.New("image creator task is not runnable")

var ErrImageCreatorActiveTaskExists = errors.New("image creator active task already exists")

var imageCreatorViewCountPattern = regexp.MustCompile(`(?i)([0-9]{1,5})[[:space:]]*(?:个|张|种|幅)?[[:space:]]*(?:不同的?|different[[:space:]]+)?(?:角度|视角|机位|镜头|angles?|views?|perspectives?)`)

var imageCreatorChineseViewCounts = []struct {
	word  string
	count int
}{
	{word: "十", count: 10},
	{word: "九", count: 9},
	{word: "八", count: 8},
	{word: "七", count: 7},
	{word: "六", count: 6},
	{word: "五", count: 5},
	{word: "四", count: 4},
	{word: "三", count: 3},
	{word: "两", count: 2},
	{word: "二", count: 2},
	{word: "一", count: 1},
}

var imageCreatorAngleLabels = []string{
	"正面视角",
	"侧面视角",
	"背面视角",
	"俯视远景",
	"低机位仰视",
	"近景特写",
	"广角环境视角",
	"斜后方三分之四视角",
}

type ImageCreatorTask struct {
	ID                     int64               `json:"id"`
	UserID                 int64               `json:"user_id"`
	APIKeyID               int64               `json:"api_key_id"`
	Status                 string              `json:"status"`
	Model                  string              `json:"model"`
	Prompt                 string              `json:"prompt"`
	Size                   string              `json:"size"`
	Quality                string              `json:"quality"`
	OutputFormat           string              `json:"output_format"`
	Background             string              `json:"background"`
	Count                  int                 `json:"count"`
	ReferenceImagePath     string              `json:"-"`
	ReferenceImageMimeType string              `json:"reference_image_mime_type,omitempty"`
	ReferenceImageFilename string              `json:"reference_image_filename,omitempty"`
	ErrorMessage           string              `json:"error_message,omitempty"`
	StartedAt              *time.Time          `json:"started_at,omitempty"`
	CompletedAt            *time.Time          `json:"completed_at,omitempty"`
	ExpiresAt              time.Time           `json:"expires_at"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
	Images                 []ImageCreatorImage `json:"images,omitempty"`
	Metadata               map[string]any      `json:"metadata,omitempty"`
}

type ImageCreatorImage struct {
	ID            int64     `json:"id"`
	TaskID        int64     `json:"task_id"`
	UserID        int64     `json:"user_id"`
	FilePath      string    `json:"-"`
	URL           string    `json:"url,omitempty"`
	OutputFormat  string    `json:"output_format"`
	MimeType      string    `json:"mime_type"`
	ByteSize      int64     `json:"byte_size"`
	Width         int       `json:"width,omitempty"`
	Height        int       `json:"height,omitempty"`
	Resolution    string    `json:"resolution,omitempty"`
	AspectRatio   string    `json:"aspect_ratio,omitempty"`
	Orientation   string    `json:"orientation,omitempty"`
	Megapixels    float64   `json:"megapixels,omitempty"`
	SHA256        string    `json:"sha256"`
	RevisedPrompt string    `json:"revised_prompt,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type ImageCreatorManagedImage struct {
	ImageCreatorImage
	TaskPrompt  string `json:"task_prompt,omitempty"`
	TaskModel   string `json:"task_model,omitempty"`
	TaskSize    string `json:"task_size,omitempty"`
	TaskQuality string `json:"task_quality,omitempty"`
}

type ImageCreatorImageListFilters struct {
	Limit       int
	Offset      int
	Search      string
	StartDate   string
	EndDate     string
	Format      string
	Orientation string
	Resolution  string
	AspectRatio string
	MinWidth    int
	MinHeight   int
}

type ImageCreatorCreateTaskInput struct {
	APIKeyID               int64
	Model                  string
	Prompt                 string
	Size                   string
	Quality                string
	Count                  int
	OutputFormat           string
	Background             string
	ReferenceImage         []byte
	ReferenceImageMimeType string
	ReferenceImageFilename string
}

type ImageCreatorGenerateRequest struct {
	Model                  string
	Prompt                 string
	Size                   string
	Quality                string
	Count                  int
	OutputFormat           string
	Background             string
	ReferenceImagePath     string
	ReferenceImageMimeType string
	ReferenceImageFilename string
}

type GeneratedImageAsset struct {
	Data          []byte
	OutputFormat  string
	MimeType      string
	RevisedPrompt string
}

type ImageCreatorFile struct {
	Path                   string
	Body                   io.ReadCloser
	SizeBytes              int64
	ContentType            string
	FileName               string
	DownloadBytesPerSecond int64
}

type ImageCreatorRepository interface {
	CreateTask(ctx context.Context, task *ImageCreatorTask, maxActiveTasks int) error
	UpdateTaskReferenceImage(ctx context.Context, taskID int64, path string, mimeType string, filename string) error
	GetTaskByID(ctx context.Context, taskID int64) (*ImageCreatorTask, error)
	GetTaskForUser(ctx context.Context, userID int64, taskID int64) (*ImageCreatorTask, error)
	ListTasksForUser(ctx context.Context, userID int64, limit int) ([]ImageCreatorTask, error)
	GetImageForUser(ctx context.Context, userID int64, imageID int64) (*ImageCreatorImage, error)
	ClaimNextPendingTask(ctx context.Context, staleRunningAfter time.Duration) (*ImageCreatorTask, error)
	MarkTaskRunning(ctx context.Context, taskID int64) error
	MarkTaskSucceeded(ctx context.Context, taskID int64, warning string) error
	MarkTaskFailed(ctx context.Context, taskID int64, message string) error
	AddImage(ctx context.Context, image *ImageCreatorImage) error
	ListImagesForUser(ctx context.Context, userID int64, filters ImageCreatorImageListFilters) ([]ImageCreatorManagedImage, int, error)
	ListImagesForUserByIDs(ctx context.Context, userID int64, ids []int64) ([]ImageCreatorImage, error)
	ListPrunableImages(ctx context.Context, userID int64, keep int) ([]ImageCreatorImage, error)
	DeleteImagesByID(ctx context.Context, ids []int64) error
	ListExpiredImages(ctx context.Context, before time.Time, limit int) ([]ImageCreatorImage, error)
	DeleteExpiredTasks(ctx context.Context, before time.Time, limit int) ([]ImageCreatorTask, error)
	FailStaleRunningTasks(ctx context.Context, staleRunningAfter time.Duration, message string) error
}

type ImageCreatorAPIKeyLookup interface {
	GetByID(ctx context.Context, id int64) (*APIKey, error)
}

type ImageCreatorGenerator interface {
	GenerateImage(ctx context.Context, input ImageCreatorGenerateRequest, apiKey string) ([]GeneratedImageAsset, error)
}

type ImageCreatorServiceOptions struct {
	StorageDir             string
	StorageBackend         string
	ObjectStorage          *BackupS3Config
	ObjectStoreFactory     BackupObjectStoreFactory
	MaxSavedImages         int
	Retention              time.Duration
	WorkerInterval         time.Duration
	TaskTimeout            time.Duration
	RequestTimeout         time.Duration
	CleanupBatchSize       int
	DownloadBytesPerSecond int64
	AutoStartWorker        bool
	DisableAsyncOnCreate   bool
}

type ImageCreatorService struct {
	repo                   ImageCreatorRepository
	apiKeys                ImageCreatorAPIKeyLookup
	membershipService      *MembershipService
	generator              ImageCreatorGenerator
	storageDir             string
	storageBackend         string
	objectStoreCfg         *BackupS3Config
	objectStoreFactory     BackupObjectStoreFactory
	objectStore            BackupObjectStore
	objectStoreMu          sync.Mutex
	maxSavedImages         int
	retention              time.Duration
	workerInterval         time.Duration
	taskTimeout            time.Duration
	requestTimeout         time.Duration
	cleanupBatchSize       int
	downloadBytesPerSecond int64
	processOnCreate        bool
	disableAsyncOnCreate   bool
	stopCh                 chan struct{}
	startOnce              sync.Once
	stopOnce               sync.Once
}

func NewImageCreatorService(repo ImageCreatorRepository, apiKeyService *APIKeyService, membershipService *MembershipService, cfg *config.Config, storeFactories ...BackupObjectStoreFactory) *ImageCreatorService {
	opts := imageCreatorOptionsFromConfig(cfg)
	if len(storeFactories) > 0 {
		opts.ObjectStoreFactory = storeFactories[0]
	}
	generator := NewImageCreatorHTTPGenerator(cfg, opts.RequestTimeout)
	svc := NewImageCreatorServiceWithDeps(repo, apiKeyService, generator, opts)
	svc.membershipService = membershipService
	return svc
}

func NewImageCreatorServiceLegacy(repo ImageCreatorRepository, apiKeyService *APIKeyService, cfg *config.Config) *ImageCreatorService {
	return NewImageCreatorService(repo, apiKeyService, nil, cfg)
}

func NewImageCreatorServiceWithDeps(repo ImageCreatorRepository, apiKeys ImageCreatorAPIKeyLookup, generator ImageCreatorGenerator, opts ImageCreatorServiceOptions) *ImageCreatorService {
	opts = normalizeImageCreatorOptions(opts)
	svc := &ImageCreatorService{
		repo:                   repo,
		apiKeys:                apiKeys,
		generator:              generator,
		storageDir:             opts.StorageDir,
		storageBackend:         opts.StorageBackend,
		objectStoreCfg:         opts.ObjectStorage,
		objectStoreFactory:     opts.ObjectStoreFactory,
		maxSavedImages:         opts.MaxSavedImages,
		retention:              opts.Retention,
		workerInterval:         opts.WorkerInterval,
		taskTimeout:            opts.TaskTimeout,
		requestTimeout:         opts.RequestTimeout,
		cleanupBatchSize:       opts.CleanupBatchSize,
		downloadBytesPerSecond: opts.DownloadBytesPerSecond,
		processOnCreate:        !opts.AutoStartWorker,
		disableAsyncOnCreate:   opts.DisableAsyncOnCreate,
		stopCh:                 make(chan struct{}),
	}
	if opts.AutoStartWorker {
		svc.Start()
	}
	return svc
}

func imageCreatorOptionsFromConfig(cfg *config.Config) ImageCreatorServiceOptions {
	opts := ImageCreatorServiceOptions{
		AutoStartWorker: true,
	}
	if cfg == nil {
		return normalizeImageCreatorOptions(opts)
	}
	ic := cfg.ImageCreator
	opts.StorageDir = ic.StorageDir
	opts.StorageBackend = ic.StorageBackend
	opts.ObjectStorage = &BackupS3Config{
		Endpoint:        strings.TrimSpace(ic.ObjectStorage.Endpoint),
		Region:          strings.TrimSpace(ic.ObjectStorage.Region),
		Bucket:          strings.TrimSpace(ic.ObjectStorage.Bucket),
		AccessKeyID:     strings.TrimSpace(ic.ObjectStorage.AccessKeyID),
		SecretAccessKey: strings.TrimSpace(ic.ObjectStorage.SecretAccessKey),
		Prefix:          strings.Trim(strings.TrimSpace(ic.ObjectStorage.Prefix), "/"),
		ForcePathStyle:  ic.ObjectStorage.ForcePathStyle,
	}
	opts.MaxSavedImages = ic.MaxSavedImagesPerUser
	opts.Retention = time.Duration(ic.RetentionDays) * 24 * time.Hour
	opts.WorkerInterval = time.Duration(ic.WorkerIntervalSeconds) * time.Second
	opts.TaskTimeout = time.Duration(ic.TaskTimeoutSeconds) * time.Second
	opts.RequestTimeout = time.Duration(ic.RequestTimeoutSeconds) * time.Second
	opts.CleanupBatchSize = ic.CleanupBatchSize
	opts.DownloadBytesPerSecond = ic.DownloadBytesPerSecond
	return normalizeImageCreatorOptions(opts)
}

func normalizeImageCreatorOptions(opts ImageCreatorServiceOptions) ImageCreatorServiceOptions {
	if strings.TrimSpace(opts.StorageDir) == "" {
		opts.StorageDir = defaultImageCreatorStorageDir
	}
	if opts.ObjectStorage != nil {
		opts.ObjectStorage.Prefix = strings.Trim(strings.TrimSpace(opts.ObjectStorage.Prefix), "/")
		if opts.ObjectStorage.Prefix == "" {
			opts.ObjectStorage.Prefix = defaultImageCreatorObjectStoragePrefix
		}
	}
	opts.StorageBackend = normalizeImageCreatorStorageBackend(opts.StorageBackend, opts.ObjectStorage)
	if opts.MaxSavedImages <= 0 {
		opts.MaxSavedImages = defaultImageCreatorMaxSavedImages
	}
	if opts.Retention <= 0 {
		opts.Retention = defaultImageCreatorRetention
	}
	if opts.WorkerInterval <= 0 {
		opts.WorkerInterval = defaultImageCreatorWorkerInterval
	}
	if opts.TaskTimeout <= 0 {
		opts.TaskTimeout = defaultImageCreatorTaskTimeout
	}
	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = defaultImageCreatorRequestTimeout
	}
	if opts.CleanupBatchSize <= 0 {
		opts.CleanupBatchSize = defaultImageCreatorCleanupBatchSize
	}
	if opts.DownloadBytesPerSecond < 0 {
		opts.DownloadBytesPerSecond = 0
	}
	return opts
}

func normalizeImageCreatorStorageBackend(raw string, objectStorage *BackupS3Config) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return normalizeImageCreatorAutoStorageBackend(objectStorage)
	case imageCreatorStorageBackendAuto:
		return normalizeImageCreatorAutoStorageBackend(objectStorage)
	case imageCreatorStorageBackendLocal:
		return imageCreatorStorageBackendLocal
	case imageCreatorStorageBackendS3:
		return imageCreatorStorageBackendS3
	case imageCreatorStorageBackendCOS:
		return imageCreatorStorageBackendCOS
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeImageCreatorAutoStorageBackend(objectStorage *BackupS3Config) string {
	if objectStorage != nil &&
		strings.TrimSpace(objectStorage.Endpoint) != "" &&
		strings.TrimSpace(objectStorage.Bucket) != "" &&
		strings.TrimSpace(objectStorage.AccessKeyID) != "" &&
		strings.TrimSpace(objectStorage.SecretAccessKey) != "" {
		return imageCreatorStorageBackendCOS
	}
	return imageCreatorStorageBackendLocal
}

func (s *ImageCreatorService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		go s.workerLoop()
	})
}

func (s *ImageCreatorService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

func (s *ImageCreatorService) CreateTask(ctx context.Context, userID int64, input ImageCreatorCreateTaskInput) (*ImageCreatorTask, error) {
	if s == nil || s.repo == nil || s.apiKeys == nil {
		return nil, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "user_id is required")
	}
	input = normalizeCreateTaskInput(input)
	if input.APIKeyID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_API_KEY", "api_key_id is required")
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return nil, infraerrors.BadRequest("INVALID_PROMPT", "prompt is required")
	}
	if input.Count > maxImageCreatorTaskCount {
		return nil, infraerrors.BadRequest("INVALID_COUNT", fmt.Sprintf("count must be between 1 and %d", maxImageCreatorTaskCount))
	}

	apiKey, err := s.apiKeys.GetByID(ctx, input.APIKeyID)
	if err != nil || apiKey == nil {
		return nil, infraerrors.NotFound("API_KEY_NOT_FOUND", "api key not found")
	}
	if err := validateImageCreatorAPIKeyForUser(apiKey, userID); err != nil {
		return nil, err
	}
	maxActiveTasks := defaultImageCreatorActiveTaskLimit
	if s.membershipService != nil {
		maxActiveTasks = s.membershipService.ImageActiveTaskLimit(ctx, userID)
	}

	task := &ImageCreatorTask{
		UserID:       userID,
		APIKeyID:     input.APIKeyID,
		Status:       ImageCreatorTaskStatusPending,
		Model:        input.Model,
		Prompt:       input.Prompt,
		Size:         input.Size,
		Quality:      input.Quality,
		OutputFormat: input.OutputFormat,
		Background:   input.Background,
		Count:        input.Count,
		ExpiresAt:    time.Now().Add(s.retention),
	}
	if err := s.repo.CreateTask(ctx, task, maxActiveTasks); err != nil {
		if errors.Is(err, ErrImageCreatorActiveTaskExists) {
			return nil, infraerrors.TooManyRequests("IMAGE_CREATOR_TASK_LIMIT_EXCEEDED", fmt.Sprintf("active image generation task limit exceeded (%d)", maxActiveTasks))
		}
		return nil, err
	}
	if len(input.ReferenceImage) > 0 {
		path, err := s.saveReferenceImage(userID, task.ID, input.ReferenceImage, input.ReferenceImageMimeType, input.ReferenceImageFilename)
		if err != nil {
			_ = s.repo.MarkTaskFailed(context.Background(), task.ID, err.Error())
			return nil, err
		}
		task.ReferenceImagePath = path
		task.ReferenceImageMimeType = input.ReferenceImageMimeType
		task.ReferenceImageFilename = input.ReferenceImageFilename
		if err := s.repo.UpdateTaskReferenceImage(ctx, task.ID, path, input.ReferenceImageMimeType, input.ReferenceImageFilename); err != nil {
			_ = os.Remove(path)
			return nil, err
		}
	}

	if !s.disableAsyncOnCreate && s.processOnCreate {
		taskID := task.ID
		go func() {
			bg, cancel := context.WithTimeout(context.Background(), s.requestTimeout*time.Duration(maxInt(1, task.Count)))
			defer cancel()
			if err := s.ProcessTask(bg, taskID); err != nil && !errors.Is(err, errImageCreatorTaskNotRunnable) {
				logger.L().Warn("image creator async task failed", zap.Int64("task_id", taskID), zap.Error(err))
			}
		}()
	}
	return task, nil
}

func normalizeCreateTaskInput(input ImageCreatorCreateTaskInput) ImageCreatorCreateTaskInput {
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" {
		input.Model = "gpt-image-2"
	}
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.Size = strings.TrimSpace(input.Size)
	input.Quality = strings.TrimSpace(input.Quality)
	input.OutputFormat = normalizeImageCreatorRequestedOutputFormat(input.OutputFormat)
	input.Background = normalizeImageCreatorBackground(input.Model, input.Background)
	input.ReferenceImageMimeType = normalizeImageMimeType(input.ReferenceImageMimeType, input.ReferenceImage)
	input.ReferenceImageFilename = strings.TrimSpace(input.ReferenceImageFilename)
	if input.Count <= 0 {
		input.Count = 1
	}
	if input.Count <= 1 {
		if inferred := inferImageCreatorViewCount(input.Prompt); inferred > input.Count {
			input.Count = inferred
		}
	}
	return input
}

func validateImageCreatorAPIKeyForUser(apiKey *APIKey, userID int64) error {
	if apiKey == nil {
		return infraerrors.NotFound("API_KEY_NOT_FOUND", "api key not found")
	}
	if apiKey.UserID != userID {
		return infraerrors.Forbidden("API_KEY_FORBIDDEN", "api key does not belong to current user")
	}
	if apiKey.Status != StatusAPIKeyActive && apiKey.Status != StatusActive {
		return infraerrors.BadRequest("API_KEY_NOT_ACTIVE", "api key is not active")
	}
	if apiKey.Group == nil || apiKey.Group.Platform != PlatformOpenAI || !GroupAllowsImageGeneration(apiKey.Group) {
		return infraerrors.Forbidden("IMAGE_GENERATION_FORBIDDEN", ImageGenerationPermissionMessage())
	}
	return nil
}

func (s *ImageCreatorService) GetTask(ctx context.Context, userID int64, taskID int64) (*ImageCreatorTask, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if taskID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_TASK", "task id is required")
	}
	task, err := s.repo.GetTaskForUser(ctx, userID, taskID)
	if err != nil {
		return nil, infraerrors.NotFound("IMAGE_CREATOR_TASK_NOT_FOUND", "image task not found")
	}
	attachImageURLs(task)
	return task, nil
}

func (s *ImageCreatorService) ListTasks(ctx context.Context, userID int64, limit int) ([]ImageCreatorTask, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if limit <= 0 {
		limit = defaultImageCreatorListTaskLimit
	}
	tasks, err := s.repo.ListTasksForUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		attachImageURLs(&tasks[i])
	}
	return tasks, nil
}

func (s *ImageCreatorService) ListImages(ctx context.Context, userID int64, filters ImageCreatorImageListFilters) ([]ImageCreatorManagedImage, int, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if userID <= 0 {
		return nil, 0, infraerrors.BadRequest("INVALID_USER", "user_id is required")
	}
	filters = normalizeImageCreatorImageListFilters(filters)
	images, total, err := s.repo.ListImagesForUser(ctx, userID, filters)
	if err != nil {
		return nil, 0, err
	}
	for i := range images {
		attachImageCreatorImageDisplayFields(&images[i].ImageCreatorImage)
	}
	return images, total, nil
}

func normalizeImageCreatorImageListFilters(filters ImageCreatorImageListFilters) ImageCreatorImageListFilters {
	if filters.Limit <= 0 {
		filters.Limit = defaultImageCreatorListTaskLimit
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	filters.Search = strings.TrimSpace(filters.Search)
	filters.StartDate = normalizeImageCreatorDateFilter(filters.StartDate)
	filters.EndDate = normalizeImageCreatorDateFilter(filters.EndDate)
	filters.Format = normalizeImageCreatorFormatFilter(filters.Format)
	filters.Orientation = normalizeImageCreatorOrientationFilter(filters.Orientation)
	filters.Resolution = normalizeImageCreatorResolutionFilter(filters.Resolution)
	filters.AspectRatio = normalizeImageCreatorAspectRatioFilter(filters.AspectRatio)
	if filters.MinWidth < 0 {
		filters.MinWidth = 0
	}
	if filters.MinHeight < 0 {
		filters.MinHeight = 0
	}
	return filters
}

func normalizeImageCreatorDateFilter(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return ""
	}
	return value
}

func normalizeImageCreatorFormatFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "png", "webp":
		return strings.ToLower(strings.TrimSpace(value))
	case "jpg", "jpeg":
		return "jpeg"
	case "other":
		return "other"
	default:
		return ""
	}
}

func normalizeImageCreatorOrientationFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "landscape", "portrait", "square", "unknown":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeImageCreatorResolutionFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1080p", "2k", "4k", "unknown":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeImageCreatorAspectRatioFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1:1", "4:3", "3:4", "16:9", "9:16", "other", "unknown":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func (s *ImageCreatorService) DeleteImages(ctx context.Context, userID int64, ids []int64) (int, error) {
	if s == nil || s.repo == nil {
		return 0, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if userID <= 0 {
		return 0, infraerrors.BadRequest("INVALID_USER", "user_id is required")
	}
	ids = normalizeImageCreatorImageIDs(ids)
	if len(ids) == 0 {
		return 0, infraerrors.BadRequest("INVALID_IMAGE_IDS", "image ids are required")
	}
	images, err := s.repo.ListImagesForUserByIDs(ctx, userID, ids)
	if err != nil {
		return 0, err
	}
	if len(images) == 0 {
		return 0, infraerrors.NotFound("IMAGE_CREATOR_IMAGE_NOT_FOUND", "image not found")
	}
	deleteIDs := make([]int64, 0, len(images))
	for _, image := range images {
		deleteIDs = append(deleteIDs, image.ID)
	}
	if err := s.repo.DeleteImagesByID(ctx, deleteIDs); err != nil {
		return 0, err
	}
	for _, image := range images {
		s.removeStoredImageQuietly(ctx, image.FilePath)
	}
	return len(images), nil
}

func (s *ImageCreatorService) GetImageFile(ctx context.Context, userID int64, imageID int64) (*ImageCreatorFile, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	image, err := s.repo.GetImageForUser(ctx, userID, imageID)
	if err != nil || image == nil {
		return nil, infraerrors.NotFound("IMAGE_CREATOR_IMAGE_NOT_FOUND", "image not found")
	}
	if strings.TrimSpace(image.FilePath) == "" {
		return nil, infraerrors.NotFound("IMAGE_CREATOR_IMAGE_NOT_FOUND", "image file not found")
	}
	if isImageCreatorObjectStoragePath(image.FilePath) {
		body, err := s.openStoredImage(ctx, image.FilePath)
		if err != nil {
			return nil, infraerrors.NotFound("IMAGE_CREATOR_IMAGE_NOT_FOUND", "image file not found")
		}
		return &ImageCreatorFile{
			Body:                   body,
			SizeBytes:              image.ByteSize,
			ContentType:            firstNonEmptyString(image.MimeType, mimeTypeForOutputFormat(image.OutputFormat)),
			FileName:               fmt.Sprintf("image-%d.%s", image.ID, normalizeImageCreatorStoredOutputFormat(image.OutputFormat)),
			DownloadBytesPerSecond: s.downloadBytesPerSecond,
		}, nil
	}
	if _, err := os.Stat(image.FilePath); err != nil {
		return nil, infraerrors.NotFound("IMAGE_CREATOR_IMAGE_NOT_FOUND", "image file not found")
	}
	return &ImageCreatorFile{
		Path:                   image.FilePath,
		ContentType:            firstNonEmptyString(image.MimeType, mimeTypeForOutputFormat(image.OutputFormat)),
		FileName:               fmt.Sprintf("image-%d.%s", image.ID, normalizeImageCreatorStoredOutputFormat(image.OutputFormat)),
		DownloadBytesPerSecond: s.downloadBytesPerSecond,
	}, nil
}

func (s *ImageCreatorService) GetReferenceImageForUser(ctx context.Context, userID int64, imageID int64) (*ImageCreatorFile, error) {
	return s.GetImageFile(ctx, userID, imageID)
}

func (s *ImageCreatorService) ProcessTask(ctx context.Context, taskID int64) error {
	if s == nil || s.repo == nil {
		return infraerrors.InternalServer("IMAGE_CREATOR_UNAVAILABLE", "image creator service is unavailable")
	}
	if err := s.repo.MarkTaskRunning(ctx, taskID); err != nil {
		return errImageCreatorTaskNotRunnable
	}
	task, err := s.findTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	return s.processClaimedTask(ctx, task)
}

func (s *ImageCreatorService) findTaskByID(ctx context.Context, taskID int64) (*ImageCreatorTask, error) {
	return s.repo.GetTaskByID(ctx, taskID)
}

func (s *ImageCreatorService) processClaimedTask(ctx context.Context, task *ImageCreatorTask) error {
	if task == nil {
		return nil
	}
	apiKey, err := s.apiKeys.GetByID(ctx, task.APIKeyID)
	if err != nil || apiKey == nil {
		msg := "api key not found"
		_ = s.repo.MarkTaskFailed(context.Background(), task.ID, msg)
		return errors.New(msg)
	}
	if err := validateImageCreatorAPIKeyForUser(apiKey, task.UserID); err != nil {
		_ = s.repo.MarkTaskFailed(context.Background(), task.ID, imageCreatorTaskErrorMessage(err))
		return err
	}
	successCount := 0
	var lastErr error
	generateCount := maxInt(1, task.Count)
	plannedPrompts := planImageCreatorPrompts(task.Prompt, generateCount)
	for i := 0; i < generateCount && successCount < generateCount; {
		prompt := task.Prompt
		if i < len(plannedPrompts) {
			prompt = plannedPrompts[i]
		}
		batchCount := imageCreatorBatchCountForPlannedPrompts(plannedPrompts, i, generateCount)
		assets, err := s.generator.GenerateImage(ctx, ImageCreatorGenerateRequest{
			Model:                  task.Model,
			Prompt:                 prompt,
			Size:                   task.Size,
			Quality:                task.Quality,
			Count:                  batchCount,
			OutputFormat:           task.OutputFormat,
			Background:             task.Background,
			ReferenceImagePath:     task.ReferenceImagePath,
			ReferenceImageMimeType: task.ReferenceImageMimeType,
			ReferenceImageFilename: task.ReferenceImageFilename,
		}, apiKey.Key)
		if err != nil {
			lastErr = err
			i += batchCount
			continue
		}
		for _, asset := range assets {
			if successCount >= generateCount {
				break
			}
			if len(asset.Data) == 0 {
				continue
			}
			outputFormat := normalizeImageCreatorStoredOutputFormat(firstNonEmptyString(asset.OutputFormat, task.OutputFormat))
			mimeType := firstNonEmptyString(asset.MimeType, mimeTypeForOutputFormat(outputFormat))
			width, height := imageCreatorImageDimensions(asset.Data)
			filePath, hash, byteSize, err := s.saveGeneratedImage(ctx, task.UserID, task.ID, successCount+1, outputFormat, asset.Data)
			if err != nil {
				lastErr = err
				continue
			}
			image := &ImageCreatorImage{
				TaskID:        task.ID,
				UserID:        task.UserID,
				FilePath:      filePath,
				OutputFormat:  outputFormat,
				MimeType:      mimeType,
				ByteSize:      byteSize,
				Width:         width,
				Height:        height,
				SHA256:        hash,
				RevisedPrompt: strings.TrimSpace(asset.RevisedPrompt),
				ExpiresAt:     time.Now().Add(s.retention),
			}
			if err := s.repo.AddImage(ctx, image); err != nil {
				_ = s.removeStoredImage(context.Background(), filePath)
				lastErr = err
				continue
			}
			successCount++
		}
		i += batchCount
	}
	if pruneErr := s.pruneUserImages(ctx, task.UserID, successCount); pruneErr != nil && lastErr == nil {
		lastErr = pruneErr
	}
	if successCount == 0 {
		msg := "image generation failed"
		if lastErr != nil {
			msg = lastErr.Error()
		}
		_ = s.repo.MarkTaskFailed(context.Background(), task.ID, msg)
		return errors.New(msg)
	}
	warning := ""
	if successCount < task.Count {
		warning = fmt.Sprintf("generated %d of %d images", successCount, task.Count)
		if lastErr != nil {
			warning = warning + ": " + lastErr.Error()
		}
	}
	if err := s.repo.MarkTaskSucceeded(ctx, task.ID, warning); err != nil {
		return err
	}
	return nil
}

func (s *ImageCreatorService) CleanupExpired(ctx context.Context) error {
	if s == nil || s.repo == nil {
		return nil
	}
	now := time.Now()
	images, err := s.repo.ListExpiredImages(ctx, now, s.cleanupBatchSize)
	if err != nil {
		return err
	}
	if len(images) > 0 {
		ids := make([]int64, 0, len(images))
		for _, image := range images {
			ids = append(ids, image.ID)
		}
		if err := s.repo.DeleteImagesByID(ctx, ids); err != nil {
			return err
		}
		for _, image := range images {
			s.removeStoredImageQuietly(ctx, image.FilePath)
		}
	}
	tasks, err := s.repo.DeleteExpiredTasks(ctx, now, s.cleanupBatchSize)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		removeFileQuietly(task.ReferenceImagePath)
		for _, image := range task.Images {
			s.removeStoredImageQuietly(ctx, image.FilePath)
		}
	}
	return nil
}

func (s *ImageCreatorService) workerLoop() {
	if s.repo == nil {
		return
	}
	if err := s.repo.FailStaleRunningTasks(context.Background(), s.taskTimeout, "server restarted while image generation was running"); err != nil {
		logger.L().Warn("image creator fail stale tasks failed", zap.Error(err))
	}
	ticker := time.NewTicker(s.workerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		s.runWorkerOnce()
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
		}
	}
}

func (s *ImageCreatorService) runWorkerOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), s.requestTimeout)
	defer cancel()
	if err := s.repo.FailStaleRunningTasks(ctx, s.taskTimeout, "server restarted while image generation was running"); err != nil {
		logger.L().Warn("image creator fail stale tasks failed", zap.Error(err))
	}
	_ = s.CleanupExpired(ctx)
	task, err := s.repo.ClaimNextPendingTask(ctx, s.taskTimeout)
	if err != nil {
		logger.L().Warn("image creator claim task failed", zap.Error(err))
		return
	}
	if task == nil {
		return
	}
	processCtx, processCancel := s.taskProcessContext(ctx, task)
	defer processCancel()
	if err := s.processClaimedTask(processCtx, task); err != nil {
		logger.L().Warn("image creator process task failed", zap.Int64("task_id", task.ID), zap.Error(err))
	}
}

func (s *ImageCreatorService) taskProcessContext(parent context.Context, task *ImageCreatorTask) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	timeout := s.requestTimeout
	if timeout <= 0 {
		timeout = defaultImageCreatorRequestTimeout
	}
	count := 1
	if task != nil {
		count = maxInt(1, task.Count)
	}
	return context.WithTimeout(context.WithoutCancel(parent), timeout*time.Duration(count))
}

func (s *ImageCreatorService) pruneUserImages(ctx context.Context, userID int64, keepAtLeast int) error {
	if s.maxSavedImages <= 0 {
		return nil
	}
	keep := maxInt(s.maxSavedImages, keepAtLeast)
	images, err := s.repo.ListPrunableImages(ctx, userID, keep)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(images))
	for _, image := range images {
		ids = append(ids, image.ID)
	}
	if err := s.repo.DeleteImagesByID(ctx, ids); err != nil {
		return err
	}
	for _, image := range images {
		s.removeStoredImageQuietly(ctx, image.FilePath)
	}
	return nil
}

func imageCreatorTaskErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var appErr *infraerrors.ApplicationError
	if errors.As(err, &appErr) && strings.TrimSpace(appErr.Message) != "" {
		return appErr.Message
	}
	return err.Error()
}

func inferImageCreatorViewCount(prompt string) int {
	text := strings.ToLower(strings.TrimSpace(prompt))
	if text == "" {
		return 0
	}
	if match := imageCreatorViewCountPattern.FindStringSubmatch(text); len(match) == 2 {
		n, err := strconv.Atoi(match[1])
		if err == nil {
			return n
		}
	}
	viewTerms := []string{"角度", "视角", "机位", "镜头"}
	countSuffixes := []string{"", "个", "个不同", "个不同的", "种", "种不同", "种不同的", "张", "幅"}
	for _, candidate := range imageCreatorChineseViewCounts {
		for _, suffix := range countSuffixes {
			for _, term := range viewTerms {
				if strings.Contains(text, candidate.word+suffix+term) {
					return candidate.count
				}
			}
		}
	}
	return 0
}

func planImageCreatorPrompts(prompt string, count int) []string {
	count = maxInt(1, count)
	base := strings.TrimSpace(prompt)
	prompts := make([]string, count)
	if count == 1 || !imageCreatorPromptRequestsMultipleViews(base) {
		for i := range prompts {
			prompts[i] = base
		}
		return prompts
	}
	for i := range prompts {
		label := fmt.Sprintf("第 %d 个角度", i+1)
		if i < len(imageCreatorAngleLabels) {
			label = imageCreatorAngleLabels[i]
		}
		prompts[i] = fmt.Sprintf("%s\n\n第 %d 张（%s）：只生成一张完整图片，画面只使用这一个镜头角度；不要四宫格、不要拼图、不要分屏、不要在同一张图里展示多个角度。", base, i+1, label)
	}
	return prompts
}

func imageCreatorBatchCountForPlannedPrompts(prompts []string, start int, total int) int {
	if start < 0 || start >= total {
		return 1
	}
	limit := minInt(maxImageCreatorUpstreamBatchCount, total-start)
	if limit <= 1 {
		return 1
	}
	prompt := ""
	if start < len(prompts) {
		prompt = prompts[start]
	}
	count := 1
	for count < limit {
		next := ""
		if start+count < len(prompts) {
			next = prompts[start+count]
		}
		if next != prompt {
			break
		}
		count++
	}
	return count
}

func imageCreatorPromptRequestsMultipleViews(prompt string) bool {
	text := strings.ToLower(strings.TrimSpace(prompt))
	if inferImageCreatorViewCount(text) > 1 {
		return true
	}
	triggers := []string{
		"多个角度",
		"多个视角",
		"多角度",
		"多视角",
		"不同角度",
		"不同视角",
		"不同机位",
		"multiple angles",
		"multiple views",
		"different angles",
		"different views",
		"different perspectives",
	}
	for _, trigger := range triggers {
		if strings.Contains(text, trigger) {
			return true
		}
	}
	return false
}

func (s *ImageCreatorService) saveReferenceImage(userID int64, taskID int64, data []byte, mimeType string, filename string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	if int64(len(data)) > imageCreatorMaxStoredImageBytes {
		return "", infraerrors.BadRequest("REFERENCE_IMAGE_TOO_LARGE", "reference image is too large")
	}
	ext := imageExtension(normalizeImageMimeType(mimeType, data), filename, "png")
	dir := filepath.Join(s.storageDir, strconv.FormatInt(userID, 10), "refs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d-reference.%s", taskID, ext))
	return path, os.WriteFile(path, data, 0o600)
}

func (s *ImageCreatorService) saveGeneratedImage(ctx context.Context, userID int64, taskID int64, index int, outputFormat string, data []byte) (string, string, int64, error) {
	if len(data) == 0 {
		return "", "", 0, errors.New("image data is empty")
	}
	if int64(len(data)) > imageCreatorMaxStoredImageBytes {
		return "", "", 0, errors.New("generated image is too large")
	}
	outputFormat = normalizeImageCreatorStoredOutputFormat(outputFormat)
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	if s != nil && s.usesObjectStorage() {
		store, err := s.getImageObjectStore(ctx)
		if err != nil {
			return "", "", 0, err
		}
		key := s.generatedImageObjectKey(userID, taskID, index, outputFormat)
		size, err := store.Upload(ctx, key, bytes.NewReader(data), mimeTypeForOutputFormat(outputFormat))
		if err != nil {
			return "", "", 0, err
		}
		if size <= 0 {
			size = int64(len(data))
		}
		return imageCreatorObjectStoragePathPrefix + key, hash, size, nil
	}
	dir := filepath.Join(s.storageDir, strconv.FormatInt(userID, 10))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", 0, err
	}
	path := filepath.Join(dir, fmt.Sprintf("%d-%d.%s", taskID, index, outputFormat))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", "", 0, err
	}
	return path, hash, int64(len(data)), nil
}

func (s *ImageCreatorService) usesObjectStorage() bool {
	if s == nil {
		return false
	}
	return s.storageBackend == imageCreatorStorageBackendS3 || s.storageBackend == imageCreatorStorageBackendCOS
}

func (s *ImageCreatorService) getImageObjectStore(ctx context.Context) (BackupObjectStore, error) {
	if s == nil {
		return nil, errors.New("image creator service is unavailable")
	}
	if !s.usesObjectStorage() {
		return nil, errors.New("image object storage is not enabled")
	}
	if s.objectStoreFactory == nil {
		return nil, errors.New("image object storage factory is unavailable")
	}
	if s.objectStoreCfg == nil || !s.objectStoreCfg.IsConfigured() {
		return nil, errors.New("image object storage is not configured")
	}
	s.objectStoreMu.Lock()
	defer s.objectStoreMu.Unlock()
	if s.objectStore != nil {
		return s.objectStore, nil
	}
	store, err := s.objectStoreFactory(ctx, s.objectStoreCfg)
	if err != nil {
		return nil, err
	}
	s.objectStore = store
	return store, nil
}

func (s *ImageCreatorService) generatedImageObjectKey(userID int64, taskID int64, index int, outputFormat string) string {
	prefix := defaultImageCreatorObjectStoragePrefix
	if s != nil && s.objectStoreCfg != nil && strings.TrimSpace(s.objectStoreCfg.Prefix) != "" {
		prefix = strings.Trim(strings.TrimSpace(s.objectStoreCfg.Prefix), "/")
	}
	return fmt.Sprintf("%s/%d/%s/%d-%d.%s", prefix, userID, time.Now().UTC().Format("2006/01/02"), taskID, index, normalizeImageCreatorStoredOutputFormat(outputFormat))
}

func (s *ImageCreatorService) openStoredImage(ctx context.Context, path string) (io.ReadCloser, error) {
	if !isImageCreatorObjectStoragePath(path) {
		return os.Open(path)
	}
	store, err := s.getImageObjectStore(ctx)
	if err != nil {
		return nil, err
	}
	return store.Download(ctx, imageCreatorObjectStorageKey(path))
}

func (s *ImageCreatorService) removeStoredImage(ctx context.Context, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if !isImageCreatorObjectStoragePath(path) {
		return os.Remove(path)
	}
	store, err := s.getImageObjectStore(ctx)
	if err != nil {
		return err
	}
	return store.Delete(ctx, imageCreatorObjectStorageKey(path))
}

func (s *ImageCreatorService) removeStoredImageQuietly(ctx context.Context, path string) {
	if err := s.removeStoredImage(ctx, path); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.L().Warn("image creator remove stored image failed", zap.String("path", path), zap.Error(err))
	}
}

func isImageCreatorObjectStoragePath(path string) bool {
	return strings.HasPrefix(strings.TrimSpace(path), imageCreatorObjectStoragePathPrefix)
}

func imageCreatorObjectStorageKey(path string) string {
	return strings.TrimPrefix(strings.TrimSpace(path), imageCreatorObjectStoragePathPrefix)
}

type ImageCreatorHTTPGenerator struct {
	client  *http.Client
	baseURL string
}

func NewImageCreatorHTTPGenerator(cfg *config.Config, timeout time.Duration) *ImageCreatorHTTPGenerator {
	if timeout <= 0 {
		timeout = defaultImageCreatorRequestTimeout
	}
	return &ImageCreatorHTTPGenerator{
		client:  &http.Client{Timeout: timeout},
		baseURL: resolveImageCreatorGatewayBaseURL(cfg),
	}
}

func resolveImageCreatorGatewayBaseURL(cfg *config.Config) string {
	if cfg != nil {
		if raw := strings.TrimSpace(cfg.ImageCreator.LocalGatewayBaseURL); raw != "" {
			return strings.TrimRight(raw, "/")
		}
		host := strings.TrimSpace(cfg.Server.Host)
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		port := cfg.Server.Port
		if port <= 0 {
			port = 8080
		}
		return fmt.Sprintf("http://%s:%d", host, port)
	}
	return "http://127.0.0.1:8080"
}

func (g *ImageCreatorHTTPGenerator) GenerateImage(ctx context.Context, input ImageCreatorGenerateRequest, apiKey string) ([]GeneratedImageAsset, error) {
	if g == nil || g.client == nil {
		return nil, errors.New("image generator is unavailable")
	}
	input.Background = normalizeImageCreatorBackground(input.Model, input.Background)
	endpointMode := "generations"
	if strings.TrimSpace(input.ReferenceImagePath) != "" {
		endpointMode = "edits"
	}
	count := minInt(maxInt(1, input.Count), maxImageCreatorUpstreamBatchCount)
	endpoint := buildImageCreatorGatewayEndpoint(g.baseURL, endpointMode)
	var body io.Reader
	contentType := "application/json"
	if endpointMode == "edits" {
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		writeMultipartField(writer, "model", input.Model)
		writeMultipartField(writer, "prompt", input.Prompt)
		writeMultipartField(writer, "n", strconv.Itoa(count))
		writeMultipartField(writer, "response_format", "b64_json")
		writeMultipartField(writer, "size", input.Size)
		writeMultipartField(writer, "quality", input.Quality)
		writeMultipartField(writer, "output_format", normalizeImageCreatorRequestedOutputFormat(input.OutputFormat))
		writeMultipartField(writer, "background", input.Background)
		if err := addImageCreatorReferencePart(writer, input); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
		body = &buffer
		contentType = writer.FormDataContentType()
	} else {
		payload := map[string]any{
			"model":           input.Model,
			"prompt":          input.Prompt,
			"n":               count,
			"response_format": "b64_json",
		}
		if strings.TrimSpace(input.Size) != "" {
			payload["size"] = input.Size
		}
		if strings.TrimSpace(input.Quality) != "" {
			payload["quality"] = input.Quality
		}
		if strings.TrimSpace(input.OutputFormat) != "" {
			payload["output_format"] = normalizeImageCreatorRequestedOutputFormat(input.OutputFormat)
		}
		if strings.TrimSpace(input.Background) != "" {
			payload["background"] = input.Background
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 128<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("image gateway returned HTTP %d: %s", resp.StatusCode, extractImageCreatorGatewayError(respBody))
	}
	return g.assetsFromGatewayResponse(ctx, respBody, normalizeImageCreatorRequestedOutputFormat(input.OutputFormat))
}

func (g *ImageCreatorHTTPGenerator) assetsFromGatewayResponse(ctx context.Context, body []byte, fallbackFormat string) ([]GeneratedImageAsset, error) {
	var payload struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
			OutputFormat  string `json:"output_format"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	assets := make([]GeneratedImageAsset, 0, len(payload.Data))
	for _, item := range payload.Data {
		outputFormat := normalizeImageCreatorStoredOutputFormat(firstNonEmptyString(item.OutputFormat, fallbackFormat))
		var data []byte
		var err error
		switch {
		case strings.TrimSpace(item.B64JSON) != "":
			data, err = decodeImageCreatorBase64(item.B64JSON)
		case strings.TrimSpace(item.URL) != "":
			data, err = g.downloadImage(ctx, item.URL)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		assets = append(assets, GeneratedImageAsset{
			Data:          data,
			OutputFormat:  outputFormat,
			MimeType:      mimeTypeForOutputFormat(outputFormat),
			RevisedPrompt: strings.TrimSpace(item.RevisedPrompt),
		})
	}
	if len(assets) == 0 {
		return nil, errors.New("image response did not contain any image data")
	}
	return assets, nil
}

func (g *ImageCreatorHTTPGenerator) downloadImage(ctx context.Context, rawURL string) ([]byte, error) {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(rawURL)), "data:image/") {
		return decodeImageCreatorBase64(rawURL)
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported image url scheme: %s", parsed.Scheme)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download image returned HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, imageCreatorMaxStoredImageBytes))
}

func buildImageCreatorGatewayEndpoint(baseURL string, mode string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/images/" + mode
	}
	return base + "/v1/images/" + mode
}

func writeMultipartField(writer *multipart.Writer, name string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	_ = writer.WriteField(name, value)
}

func addImageCreatorReferencePart(writer *multipart.Writer, input ImageCreatorGenerateRequest) error {
	path := strings.TrimSpace(input.ReferenceImagePath)
	if path == "" {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	header := make(textproto.MIMEHeader)
	filename := strings.TrimSpace(input.ReferenceImageFilename)
	if filename == "" {
		filename = filepath.Base(path)
	}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, escapeQuotes(filename)))
	header.Set("Content-Type", firstNonEmptyString(input.ReferenceImageMimeType, "image/png"))
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	return err
}

func escapeQuotes(v string) string {
	return strings.ReplaceAll(v, `"`, `\"`)
}

func decodeImageCreatorBase64(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, ","); idx >= 0 && len(raw) >= len("data:") && strings.EqualFold(raw[:len("data:")], "data:") {
		raw = raw[idx+1:]
	}
	raw = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, raw)
	if raw == "" {
		return nil, errors.New("image data is empty")
	}
	unpadded := strings.TrimRight(raw, "=")
	padded := unpadded + strings.Repeat("=", (4-len(unpadded)%4)%4)

	if data, err := base64.StdEncoding.DecodeString(padded); err == nil {
		return data, nil
	}
	if data, err := base64.URLEncoding.DecodeString(padded); err == nil {
		return data, nil
	}
	if data, err := base64.RawStdEncoding.DecodeString(unpadded); err == nil {
		return data, nil
	}
	return base64.RawURLEncoding.DecodeString(unpadded)
}

func extractImageCreatorGatewayError(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if errObj, ok := payload["error"].(map[string]any); ok {
			if msg, ok := errObj["message"].(string); ok && strings.TrimSpace(msg) != "" {
				return strings.TrimSpace(msg)
			}
		}
		if msg, ok := payload["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return strings.TrimSpace(msg)
		}
	}
	return strings.TrimSpace(string(body))
}

func attachImageURLs(task *ImageCreatorTask) {
	if task == nil {
		return
	}
	for i := range task.Images {
		attachImageCreatorImageDisplayFields(&task.Images[i])
	}
}

func attachImageCreatorImageDisplayFields(image *ImageCreatorImage) {
	if image == nil {
		return
	}
	image.URL = imageCreatorImageURL(image.ID)
	if image.Width > 0 && image.Height > 0 {
		image.Resolution = fmt.Sprintf("%dx%d", image.Width, image.Height)
		image.AspectRatio = simplifiedImageCreatorAspectRatio(image.Width, image.Height)
		image.Orientation = imageCreatorOrientation(image.Width, image.Height)
		image.Megapixels = float64(image.Width) * float64(image.Height) / 1_000_000
	}
}

func imageCreatorImageURL(imageID int64) string {
	if imageID <= 0 {
		return ""
	}
	return fmt.Sprintf("/api/v1/user/image-creator/images/%d/file", imageID)
}

func normalizeImageCreatorImageIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func imageCreatorImageDimensions(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil && config.Width > 0 && config.Height > 0 {
		return config.Width, config.Height
	}
	return imageCreatorWebPDimensions(data)
}

func imageCreatorWebPDimensions(data []byte) (int, int) {
	if len(data) < 30 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0
	}
	chunk := string(data[12:16])
	switch chunk {
	case "VP8 ":
		if len(data) < 30 {
			return 0, 0
		}
		width := int(binary.LittleEndian.Uint16(data[26:28]) & 0x3fff)
		height := int(binary.LittleEndian.Uint16(data[28:30]) & 0x3fff)
		if width > 0 && height > 0 {
			return width, height
		}
	case "VP8L":
		if len(data) < 25 || data[20] != 0x2f {
			return 0, 0
		}
		b0, b1, b2, b3 := uint32(data[21]), uint32(data[22]), uint32(data[23]), uint32(data[24])
		width := int(1 + (((b1 & 0x3f) << 8) | b0))
		height := int(1 + (((b3 & 0x0f) << 10) | (b2 << 2) | ((b1 & 0xc0) >> 6)))
		if width > 0 && height > 0 {
			return width, height
		}
	case "VP8X":
		if len(data) < 30 {
			return 0, 0
		}
		width := 1 + int(uint32(data[24])|uint32(data[25])<<8|uint32(data[26])<<16)
		height := 1 + int(uint32(data[27])|uint32(data[28])<<8|uint32(data[29])<<16)
		if width > 0 && height > 0 {
			return width, height
		}
	}
	return 0, 0
}

func simplifiedImageCreatorAspectRatio(width int, height int) string {
	divisor := imageCreatorGCD(width, height)
	if divisor <= 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", width/divisor, height/divisor)
}

func imageCreatorGCD(a int, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func imageCreatorOrientation(width int, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if width == height {
		return "square"
	}
	if width > height {
		return "landscape"
	}
	return "portrait"
}

func normalizeImageCreatorRequestedOutputFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "jpg", "jpeg":
		return "jpeg"
	case "webp":
		return "webp"
	default:
		return "webp"
	}
}

func normalizeImageCreatorStoredOutputFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "jpg", "jpeg":
		return "jpeg"
	case "png":
		return "png"
	case "webp":
		return "webp"
	default:
		return "webp"
	}
}

func normalizeImageCreatorBackground(model string, background string) string {
	background = strings.TrimSpace(background)
	if strings.EqualFold(strings.TrimSpace(model), "gpt-image-1.5") && strings.EqualFold(background, "transparent") {
		return "auto"
	}
	return background
}

func normalizeImageMimeType(mimeType string, data []byte) string {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if strings.HasPrefix(mimeType, "image/") {
		return mimeType
	}
	if len(data) > 0 {
		detected := strings.ToLower(http.DetectContentType(data))
		if strings.HasPrefix(detected, "image/") {
			return detected
		}
	}
	return "image/png"
}

func mimeTypeForOutputFormat(format string) string {
	switch normalizeImageCreatorStoredOutputFormat(format) {
	case "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func imageExtension(mimeType string, filename string, fallback string) string {
	if ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), "."); ext == "png" || ext == "jpg" || ext == "jpeg" || ext == "webp" {
		if ext == "jpg" {
			return "jpeg"
		}
		return ext
	}
	exts, _ := mime.ExtensionsByType(mimeType)
	if len(exts) > 0 {
		ext := strings.TrimPrefix(strings.ToLower(exts[0]), ".")
		if ext == "jpg" {
			return "jpeg"
		}
		return ext
	}
	return fallback
}

func removeFileQuietly(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
