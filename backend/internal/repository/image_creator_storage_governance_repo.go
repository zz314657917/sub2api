package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type imageCreatorStorageGovernanceRepository struct {
	sql sqlExecutor
}

func NewImageCreatorStorageGovernanceRepository(sqlDB *sql.DB) service.ImageCreatorStorageGovernanceRepository {
	return &imageCreatorStorageGovernanceRepository{sql: sqlDB}
}

func (r *imageCreatorStorageGovernanceRepository) GetImageCreatorStorageDBStats(ctx context.Context, before time.Time) (service.ImageCreatorStorageGovernanceDBStats, error) {
	var stats service.ImageCreatorStorageGovernanceDBStats
	rows, err := r.sql.QueryContext(ctx, `
		SELECT
			COUNT(*) AS image_count,
			COALESCE(SUM(byte_size), 0) AS image_bytes,
			COUNT(*) FILTER (WHERE expires_at < $1) AS expired_image_count,
			COALESCE(SUM(byte_size) FILTER (WHERE expires_at < $1), 0) AS expired_image_bytes
		FROM image_creator_images
	`, before)
	if err != nil {
		return stats, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return stats, sql.ErrNoRows
	}
	if err := rows.Scan(&stats.ImageCount, &stats.ImageBytes, &stats.ExpiredImageCount, &stats.ExpiredImageBytes); err != nil {
		return stats, err
	}
	return stats, rows.Err()
}

func (r *imageCreatorStorageGovernanceRepository) ListImageCreatorReferencedPaths(ctx context.Context) ([]string, error) {
	rows, err := r.sql.QueryContext(ctx, `
		SELECT file_path
		FROM image_creator_images
		WHERE file_path IS NOT NULL AND file_path <> ''
		UNION
		SELECT reference_image_path
		FROM image_creator_tasks
		WHERE reference_image_path IS NOT NULL AND reference_image_path <> ''
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	paths := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

func (r *imageCreatorStorageGovernanceRepository) ListExpiredImageCreatorImages(ctx context.Context, before time.Time, limit int) ([]service.ImageCreatorImage, error) {
	if limit <= 0 {
		limit = 100
	}
	query := imageCreatorImageSelectSQL() + `
		WHERE expires_at < $1
		ORDER BY expires_at ASC, id ASC
		LIMIT $2
	`
	return queryImageCreatorImages(ctx, r.sql, query, before, limit)
}

func (r *imageCreatorStorageGovernanceRepository) DeleteImageCreatorImagesByID(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.sql.ExecContext(ctx, `DELETE FROM image_creator_images WHERE id = ANY($1)`, pq.Array(ids))
	return err
}
