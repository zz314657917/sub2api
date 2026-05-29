package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type fakeImageCreatorStorageGovernanceRepo struct {
	stats          ImageCreatorStorageGovernanceDBStats
	referenced     []string
	expiredBatches [][]ImageCreatorImage
	deletedIDs     []int64
}

func (r *fakeImageCreatorStorageGovernanceRepo) GetImageCreatorStorageDBStats(_ context.Context, _ time.Time) (ImageCreatorStorageGovernanceDBStats, error) {
	return r.stats, nil
}

func (r *fakeImageCreatorStorageGovernanceRepo) ListImageCreatorReferencedPaths(_ context.Context) ([]string, error) {
	return append([]string(nil), r.referenced...), nil
}

func (r *fakeImageCreatorStorageGovernanceRepo) ListExpiredImageCreatorImages(_ context.Context, _ time.Time, _ int) ([]ImageCreatorImage, error) {
	if len(r.expiredBatches) == 0 {
		return nil, nil
	}
	next := r.expiredBatches[0]
	r.expiredBatches = r.expiredBatches[1:]
	return append([]ImageCreatorImage(nil), next...), nil
}

func (r *fakeImageCreatorStorageGovernanceRepo) DeleteImageCreatorImagesByID(_ context.Context, ids []int64) error {
	r.deletedIDs = append(r.deletedIDs, ids...)
	return nil
}

func TestImageCreatorStorageGovernanceStatsCountsLocalFiles(t *testing.T) {
	dir := t.TempDir()
	managed := writeTestGovernanceFile(t, dir, "42/1-0.webp", "managed")
	writeTestGovernanceFile(t, dir, "42/orphan.webp", "orphan")
	writeTestGovernanceFile(t, dir, "previews/a.webp", "preview")
	writeTestGovernanceFile(t, dir, "thumbs/a.webp", "thumb")

	repo := &fakeImageCreatorStorageGovernanceRepo{
		stats: ImageCreatorStorageGovernanceDBStats{
			ImageCount:        1,
			ImageBytes:        7,
			ExpiredImageCount: 1,
			ExpiredImageBytes: 7,
		},
		referenced: []string{managed},
	}
	svc := NewImageCreatorStorageGovernanceService(repo, nil, &config.Config{
		ImageCreator: config.ImageCreatorConfig{
			StorageDir:            dir,
			StorageBackend:        "local",
			CleanupBatchSize:      10,
			MaxSavedImagesPerUser: 16,
			RetentionDays:         7,
		},
	})

	stats, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), stats.Images.Count)
	require.Equal(t, int64(1), stats.ExpiredImages.Count)
	require.Equal(t, int64(1), stats.OrphanFiles.Count)
	require.Equal(t, int64(len("orphan")), stats.OrphanFiles.ByteSize)
	require.Equal(t, int64(1), stats.PreviewCache.Count)
	require.Equal(t, int64(len("preview")), stats.PreviewCache.ByteSize)
	require.Equal(t, int64(1), stats.ThumbCache.Count)
	require.Equal(t, int64(len("thumb")), stats.ThumbCache.ByteSize)
}

func TestImageCreatorStorageGovernanceRejectsPathTraversalCleanup(t *testing.T) {
	dir := t.TempDir()
	outside := writeTestGovernanceFile(t, t.TempDir(), "outside.webp", "outside")
	require.NoError(t, removeLocalStorageFile(dir, outside))
	_, err := os.Stat(outside)
	require.NoError(t, err)
}

func TestImageCreatorStorageGovernanceCleanupExpiredImagesRemovesOnlyStorageDirFiles(t *testing.T) {
	dir := t.TempDir()
	inside := writeTestGovernanceFile(t, dir, "42/expired.webp", "expired")
	outside := writeTestGovernanceFile(t, t.TempDir(), "outside.webp", "outside")
	repo := &fakeImageCreatorStorageGovernanceRepo{
		expiredBatches: [][]ImageCreatorImage{{
			{ID: 1, FilePath: inside, ByteSize: int64(len("expired"))},
			{ID: 2, FilePath: outside, ByteSize: int64(len("outside"))},
		}},
	}
	svc := NewImageCreatorStorageGovernanceService(repo, nil, &config.Config{
		ImageCreator: config.ImageCreatorConfig{
			StorageDir:            dir,
			StorageBackend:        "local",
			CleanupBatchSize:      10,
			MaxSavedImagesPerUser: 16,
			RetentionDays:         7,
		},
	})

	result, err := svc.Cleanup(context.Background(), ImageCreatorStorageGovernanceActionExpiredImages)
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Deleted)
	require.Equal(t, []int64{1, 2}, repo.deletedIDs)
	_, err = os.Stat(inside)
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(outside)
	require.NoError(t, err)
}

func TestImageCreatorStorageGovernanceObjectStorageLocalActionsUnsupported(t *testing.T) {
	dir := t.TempDir()
	repo := &fakeImageCreatorStorageGovernanceRepo{}
	svc := NewImageCreatorStorageGovernanceService(repo, nil, &config.Config{
		ImageCreator: config.ImageCreatorConfig{
			StorageDir:            dir,
			StorageBackend:        "s3",
			MaxSavedImagesPerUser: 16,
			RetentionDays:         7,
			ObjectStorage: config.ImageCreatorObjectStorageConfig{
				Endpoint:        "https://s3.example.com",
				Bucket:          "bucket",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				Prefix:          "image-creator",
			},
		},
	})

	stats, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	require.True(t, stats.OrphanFiles.Unsupported)
	require.True(t, stats.PreviewCache.Unsupported)
	require.True(t, stats.ThumbCache.Unsupported)

	result, err := svc.Cleanup(context.Background(), ImageCreatorStorageGovernanceActionOrphanFiles)
	require.NoError(t, err)
	require.True(t, result.Unsupported)
	require.Zero(t, result.Deleted)
}

func writeTestGovernanceFile(t *testing.T, root string, relative string, content string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}
