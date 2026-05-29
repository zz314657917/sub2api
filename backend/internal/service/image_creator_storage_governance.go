package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	ImageCreatorStorageGovernanceActionExpiredImages = "expired_images"
	ImageCreatorStorageGovernanceActionOrphanFiles   = "orphan_files"
	ImageCreatorStorageGovernanceActionPreviewCache  = "preview_cache"
	ImageCreatorStorageGovernanceActionThumbCache    = "thumb_cache"
)

type ImageCreatorStorageGovernanceDBStats struct {
	ImageCount        int64
	ImageBytes        int64
	ExpiredImageCount int64
	ExpiredImageBytes int64
}

type ImageCreatorStorageGovernanceRepository interface {
	GetImageCreatorStorageDBStats(ctx context.Context, before time.Time) (ImageCreatorStorageGovernanceDBStats, error)
	ListImageCreatorReferencedPaths(ctx context.Context) ([]string, error)
	ListExpiredImageCreatorImages(ctx context.Context, before time.Time, limit int) ([]ImageCreatorImage, error)
	DeleteImageCreatorImagesByID(ctx context.Context, ids []int64) error
}

type ImageCreatorStorageGovernanceStats struct {
	StorageBackend string                            `json:"storage_backend"`
	StorageDir     string                            `json:"storage_dir,omitempty"`
	ScannedAt      time.Time                         `json:"scanned_at"`
	Images         ImageCreatorStorageGovernanceItem `json:"images"`
	ExpiredImages  ImageCreatorStorageGovernanceItem `json:"expired_images"`
	OrphanFiles    ImageCreatorStorageGovernanceItem `json:"orphan_files"`
	PreviewCache   ImageCreatorStorageGovernanceItem `json:"preview_cache"`
	ThumbCache     ImageCreatorStorageGovernanceItem `json:"thumb_cache"`
}

type ImageCreatorStorageGovernanceItem struct {
	Count       int64  `json:"count"`
	ByteSize    int64  `json:"byte_size"`
	Unsupported bool   `json:"unsupported,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type ImageCreatorStorageGovernanceCleanupResult struct {
	Action       string `json:"action"`
	Deleted      int64  `json:"deleted"`
	DeletedBytes int64  `json:"deleted_bytes"`
	Unsupported  bool   `json:"unsupported,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type ImageCreatorStorageGovernanceService struct {
	repo             ImageCreatorStorageGovernanceRepository
	imageCreator     *ImageCreatorService
	storageDir       string
	storageBackend   string
	cleanupBatchSize int
}

func NewImageCreatorStorageGovernanceService(repo ImageCreatorStorageGovernanceRepository, imageCreator *ImageCreatorService, cfg *config.Config) *ImageCreatorStorageGovernanceService {
	opts := imageCreatorOptionsFromConfig(cfg)
	return &ImageCreatorStorageGovernanceService{
		repo:             repo,
		imageCreator:     imageCreator,
		storageDir:       opts.StorageDir,
		storageBackend:   opts.StorageBackend,
		cleanupBatchSize: opts.CleanupBatchSize,
	}
}

func (s *ImageCreatorStorageGovernanceService) GetStats(ctx context.Context) (*ImageCreatorStorageGovernanceStats, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("image creator storage governance service is unavailable")
	}
	now := time.Now()
	dbStats, err := s.repo.GetImageCreatorStorageDBStats(ctx, now)
	if err != nil {
		return nil, err
	}
	stats := &ImageCreatorStorageGovernanceStats{
		StorageBackend: s.storageBackend,
		StorageDir:     s.storageDir,
		ScannedAt:      now.UTC(),
		Images: ImageCreatorStorageGovernanceItem{
			Count:    dbStats.ImageCount,
			ByteSize: dbStats.ImageBytes,
		},
		ExpiredImages: ImageCreatorStorageGovernanceItem{
			Count:    dbStats.ExpiredImageCount,
			ByteSize: dbStats.ExpiredImageBytes,
		},
	}
	if !s.supportsLocalScan() {
		stats.OrphanFiles = unsupportedStorageGovernanceItem()
		stats.PreviewCache = unsupportedStorageGovernanceItem()
		stats.ThumbCache = unsupportedStorageGovernanceItem()
		return stats, nil
	}
	referenced, err := s.repo.ListImageCreatorReferencedPaths(ctx)
	if err != nil {
		return nil, err
	}
	scan, err := s.scanLocalStorage(ctx, referenced)
	if err != nil {
		return nil, err
	}
	stats.OrphanFiles = storageGovernanceItemFromFiles(scan.OrphanFiles)
	stats.PreviewCache = storageGovernanceItemFromFiles(scan.PreviewCache)
	stats.ThumbCache = storageGovernanceItemFromFiles(scan.ThumbCache)
	return stats, nil
}

func (s *ImageCreatorStorageGovernanceService) Cleanup(ctx context.Context, action string) (*ImageCreatorStorageGovernanceCleanupResult, error) {
	action = strings.TrimSpace(action)
	switch action {
	case ImageCreatorStorageGovernanceActionExpiredImages:
		return s.cleanupExpiredImages(ctx)
	case ImageCreatorStorageGovernanceActionOrphanFiles:
		return s.cleanupLocalFileGroup(ctx, action, func(scan imageCreatorLocalStorageScan) []imageCreatorLocalFile {
			return scan.OrphanFiles
		})
	case ImageCreatorStorageGovernanceActionPreviewCache:
		return s.cleanupLocalFileGroup(ctx, action, func(scan imageCreatorLocalStorageScan) []imageCreatorLocalFile {
			return scan.PreviewCache
		})
	case ImageCreatorStorageGovernanceActionThumbCache:
		return s.cleanupLocalFileGroup(ctx, action, func(scan imageCreatorLocalStorageScan) []imageCreatorLocalFile {
			return scan.ThumbCache
		})
	default:
		return nil, infraerrors.BadRequest("INVALID_IMAGE_CREATOR_STORAGE_GOVERNANCE_ACTION", "invalid storage governance action")
	}
}

func (s *ImageCreatorStorageGovernanceService) cleanupExpiredImages(ctx context.Context) (*ImageCreatorStorageGovernanceCleanupResult, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("image creator storage governance service is unavailable")
	}
	result := &ImageCreatorStorageGovernanceCleanupResult{Action: ImageCreatorStorageGovernanceActionExpiredImages}
	for {
		images, err := s.repo.ListExpiredImageCreatorImages(ctx, time.Now(), s.cleanupBatchSize)
		if err != nil {
			return nil, err
		}
		if len(images) == 0 {
			return result, nil
		}
		ids := make([]int64, 0, len(images))
		for _, image := range images {
			ids = append(ids, image.ID)
		}
		if err := s.repo.DeleteImageCreatorImagesByID(ctx, ids); err != nil {
			return nil, err
		}
		for _, image := range images {
			if err := s.removeGovernedImagePath(ctx, image.FilePath); err != nil {
				return nil, err
			}
			result.Deleted++
			result.DeletedBytes += image.ByteSize
		}
		if len(images) < s.effectiveCleanupBatchSize() {
			return result, nil
		}
	}
}

func (s *ImageCreatorStorageGovernanceService) cleanupLocalFileGroup(ctx context.Context, action string, pick func(imageCreatorLocalStorageScan) []imageCreatorLocalFile) (*ImageCreatorStorageGovernanceCleanupResult, error) {
	result := &ImageCreatorStorageGovernanceCleanupResult{Action: action}
	if s == nil || s.repo == nil {
		return nil, errors.New("image creator storage governance service is unavailable")
	}
	if !s.supportsLocalScan() {
		result.Unsupported = true
		result.Reason = "unsupported_storage_backend"
		return result, nil
	}
	referenced, err := s.repo.ListImageCreatorReferencedPaths(ctx)
	if err != nil {
		return nil, err
	}
	scan, err := s.scanLocalStorage(ctx, referenced)
	if err != nil {
		return nil, err
	}
	files := pick(scan)
	for _, file := range files {
		if err := removeLocalStorageFile(s.storageDir, file.Path); err != nil {
			return nil, err
		}
		result.Deleted++
		result.DeletedBytes += file.ByteSize
	}
	return result, nil
}

func (s *ImageCreatorStorageGovernanceService) removeGovernedImagePath(ctx context.Context, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if isImageCreatorObjectStoragePath(path) {
		if s.imageCreator == nil || !s.imageCreator.usesObjectStorage() {
			return nil
		}
		return s.imageCreator.removeStoredImage(ctx, path)
	}
	return removeLocalStorageFile(s.storageDir, path)
}

func (s *ImageCreatorStorageGovernanceService) supportsLocalScan() bool {
	if s == nil {
		return false
	}
	return s.storageBackend == imageCreatorStorageBackendLocal
}

func (s *ImageCreatorStorageGovernanceService) effectiveCleanupBatchSize() int {
	if s == nil || s.cleanupBatchSize <= 0 {
		return defaultImageCreatorCleanupBatchSize
	}
	return s.cleanupBatchSize
}

type imageCreatorLocalFile struct {
	Path     string
	ByteSize int64
}

type imageCreatorLocalStorageScan struct {
	OrphanFiles  []imageCreatorLocalFile
	PreviewCache []imageCreatorLocalFile
	ThumbCache   []imageCreatorLocalFile
}

func (s *ImageCreatorStorageGovernanceService) scanLocalStorage(ctx context.Context, referencedPaths []string) (imageCreatorLocalStorageScan, error) {
	var scan imageCreatorLocalStorageScan
	base, err := absCleanPath(s.storageDir)
	if err != nil {
		return scan, err
	}
	if _, err := os.Stat(base); errors.Is(err, os.ErrNotExist) {
		return scan, nil
	} else if err != nil {
		return scan, err
	}
	referenced := make(map[string]struct{}, len(referencedPaths))
	for _, path := range referencedPaths {
		normalized, ok := normalizeLocalStoragePath(base, path)
		if ok {
			referenced[normalized] = struct{}{}
		}
	}
	err = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		file := imageCreatorLocalFile{Path: path, ByteSize: info.Size()}
		if isInImageCreatorCacheDir(base, path, true) {
			scan.PreviewCache = append(scan.PreviewCache, file)
			return nil
		}
		if isInImageCreatorCacheDir(base, path, false) {
			scan.ThumbCache = append(scan.ThumbCache, file)
			return nil
		}
		normalized, ok := normalizeLocalStoragePath(base, path)
		if !ok {
			return nil
		}
		if _, exists := referenced[normalized]; !exists {
			scan.OrphanFiles = append(scan.OrphanFiles, file)
		}
		return nil
	})
	return scan, err
}

func storageGovernanceItemFromFiles(files []imageCreatorLocalFile) ImageCreatorStorageGovernanceItem {
	var item ImageCreatorStorageGovernanceItem
	item.Count = int64(len(files))
	for _, file := range files {
		item.ByteSize += file.ByteSize
	}
	return item
}

func unsupportedStorageGovernanceItem() ImageCreatorStorageGovernanceItem {
	return ImageCreatorStorageGovernanceItem{Unsupported: true, Reason: "unsupported_storage_backend"}
}

func removeLocalStorageFile(baseDir string, path string) error {
	base, err := absCleanPath(baseDir)
	if err != nil {
		return err
	}
	normalized, ok := normalizeLocalStoragePath(base, path)
	if !ok {
		return nil
	}
	if err := os.Remove(normalized); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func normalizeLocalStoragePath(base string, path string) (string, bool) {
	path = strings.TrimSpace(path)
	if path == "" || isImageCreatorObjectStoragePath(path) {
		return "", false
	}
	candidate, err := absCleanPath(path)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return "", false
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", false
	}
	return candidate, true
}

func absCleanPath(path string) (string, error) {
	abs, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func isInImageCreatorCacheDir(base string, path string, preview bool) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts[:maxInt(0, len(parts)-1)] {
		name := strings.ToLower(strings.TrimSpace(part))
		if preview && (name == "preview" || name == "previews" || name == "preview-cache" || name == "preview_cache") {
			return true
		}
		if !preview && (name == "thumb" || name == "thumbs" || name == "thumbnail" || name == "thumbnails" || name == "thumb-cache" || name == "thumb_cache" || name == "thumbnail-cache" || name == "thumbnail_cache") {
			return true
		}
	}
	return false
}
