package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type imageCreatorRepository struct {
	sql sqlExecutor
	db  *sql.DB
}

func NewImageCreatorRepository(sqlDB *sql.DB) service.ImageCreatorRepository {
	return &imageCreatorRepository{sql: sqlDB, db: sqlDB}
}

func (r *imageCreatorRepository) CreateTask(ctx context.Context, task *service.ImageCreatorTask, maxActiveTasks int) error {
	if task == nil {
		return nil
	}
	if maxActiveTasks <= 0 {
		maxActiveTasks = 1
	}
	if r.db != nil {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback()
			}
		}()
		if err := r.createTaskWithLimit(ctx, tx, task, maxActiveTasks); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		committed = true
		return nil
	}
	return r.createTaskWithLimit(ctx, r.sql, task, maxActiveTasks)
}

func (r *imageCreatorRepository) createTaskWithLimit(ctx context.Context, exec sqlExecutor, task *service.ImageCreatorTask, maxActiveTasks int) error {
	if _, err := exec.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, task.UserID); err != nil {
		return err
	}
	query := `
		WITH active AS (
			SELECT COUNT(*) AS count
			FROM image_creator_tasks
			WHERE user_id = $1 AND status IN ($12, $13)
		)
		INSERT INTO image_creator_tasks (
			user_id, api_key_id, status, model, prompt, size, quality,
			output_format, background, image_count, expires_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		FROM active
		WHERE active.count < $14
		RETURNING id, created_at, updated_at
	`
	err := scanSingleRow(ctx, exec, query, []any{
		task.UserID,
		task.APIKeyID,
		task.Status,
		task.Model,
		task.Prompt,
		task.Size,
		task.Quality,
		task.OutputFormat,
		task.Background,
		task.Count,
		task.ExpiresAt,
		service.ImageCreatorTaskStatusPending,
		service.ImageCreatorTaskStatusRunning,
		maxActiveTasks,
	}, &task.ID, &task.CreatedAt, &task.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) || isUniqueConstraintViolation(err) {
		return service.ErrImageCreatorActiveTaskExists
	}
	return err
}

func (r *imageCreatorRepository) UpdateTaskReferenceImage(ctx context.Context, taskID int64, path string, mimeType string, filename string) error {
	query := `
		UPDATE image_creator_tasks
		SET reference_image_path = $2,
			reference_image_mime_type = $3,
			reference_image_filename = $4,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.sql.ExecContext(ctx, query, taskID, nullableString(path), nullableString(mimeType), nullableString(filename))
	return err
}

func (r *imageCreatorRepository) GetTaskByID(ctx context.Context, taskID int64) (*service.ImageCreatorTask, error) {
	task, err := r.getTask(ctx, "id = $1", []any{taskID})
	if err != nil {
		return nil, err
	}
	if err := r.attachImages(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageCreatorRepository) GetTaskForUser(ctx context.Context, userID int64, taskID int64) (*service.ImageCreatorTask, error) {
	task, err := r.getTask(ctx, "id = $1 AND user_id = $2", []any{taskID, userID})
	if err != nil {
		return nil, err
	}
	if err := r.attachImages(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageCreatorRepository) ListTasksForUser(ctx context.Context, userID int64, limit int) ([]service.ImageCreatorTask, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, user_id, api_key_id, status, model, prompt, size, quality,
			output_format, background, image_count,
			reference_image_path, reference_image_mime_type, reference_image_filename,
			error_message, started_at, completed_at, expires_at, created_at, updated_at
		FROM image_creator_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`
	rows, err := r.sql.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	tasks := make([]service.ImageCreatorTask, 0)
	for rows.Next() {
		task, err := scanImageCreatorTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range tasks {
		if err := r.attachImages(ctx, &tasks[i]); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

func (r *imageCreatorRepository) GetImageForUser(ctx context.Context, userID int64, imageID int64) (*service.ImageCreatorImage, error) {
	query := imageCreatorImageSelectSQL() + `
		WHERE id = $1 AND user_id = $2
	`
	rows, err := r.sql.QueryContext(ctx, query, imageID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	image, err := scanImageCreatorImage(rows)
	if err != nil {
		return nil, err
	}
	return &image, rows.Err()
}

func (r *imageCreatorRepository) ClaimNextPendingTask(ctx context.Context, staleRunningAfter time.Duration) (*service.ImageCreatorTask, error) {
	if staleRunningAfter <= 0 {
		staleRunningAfter = 30 * time.Minute
	}
	query := `
		WITH next AS (
			SELECT id
			FROM image_creator_tasks
			WHERE status = $1
			ORDER BY created_at ASC, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE image_creator_tasks AS tasks
		SET status = $2,
			started_at = COALESCE(started_at, NOW()),
			error_message = NULL,
			updated_at = NOW()
		FROM next
		WHERE tasks.id = next.id
		RETURNING tasks.id, tasks.user_id, tasks.api_key_id, tasks.status, tasks.model, tasks.prompt,
			tasks.size, tasks.quality, tasks.output_format, tasks.background, tasks.image_count,
			tasks.reference_image_path, tasks.reference_image_mime_type, tasks.reference_image_filename,
			tasks.error_message, tasks.started_at, tasks.completed_at, tasks.expires_at,
			tasks.created_at, tasks.updated_at
	`
	rowTask, err := r.scanTaskRow(ctx, query, service.ImageCreatorTaskStatusPending, service.ImageCreatorTaskStatusRunning)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return rowTask, err
}

func (r *imageCreatorRepository) MarkTaskRunning(ctx context.Context, taskID int64) error {
	query := `
		UPDATE image_creator_tasks
		SET status = $2,
			started_at = COALESCE(started_at, NOW()),
			error_message = NULL,
			updated_at = NOW()
		WHERE id = $1 AND status = $3
		RETURNING id
	`
	var id int64
	err := scanSingleRow(ctx, r.sql, query, []any{taskID, service.ImageCreatorTaskStatusRunning, service.ImageCreatorTaskStatusPending}, &id)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return err
}

func (r *imageCreatorRepository) MarkTaskSucceeded(ctx context.Context, taskID int64, warning string) error {
	query := `
		UPDATE image_creator_tasks
		SET status = $2,
			error_message = $3,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.sql.ExecContext(ctx, query, taskID, service.ImageCreatorTaskStatusSucceeded, nullableString(warning))
	return err
}

func (r *imageCreatorRepository) MarkTaskFailed(ctx context.Context, taskID int64, message string) error {
	query := `
		UPDATE image_creator_tasks
		SET status = $2,
			error_message = $3,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.sql.ExecContext(ctx, query, taskID, service.ImageCreatorTaskStatusFailed, strings.TrimSpace(message))
	return err
}

func (r *imageCreatorRepository) AddImage(ctx context.Context, image *service.ImageCreatorImage) error {
	if image == nil {
		return nil
	}
	query := `
		INSERT INTO image_creator_images (
			task_id, user_id, file_path, output_format, mime_type,
			byte_size, width, height, sha256, revised_prompt, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at
	`
	return scanSingleRow(ctx, r.sql, query, []any{
		image.TaskID,
		image.UserID,
		image.FilePath,
		image.OutputFormat,
		image.MimeType,
		image.ByteSize,
		image.Width,
		image.Height,
		image.SHA256,
		nullableString(image.RevisedPrompt),
		image.ExpiresAt,
	}, &image.ID, &image.CreatedAt)
}

func (r *imageCreatorRepository) ListImagesForUser(ctx context.Context, userID int64, filters service.ImageCreatorImageListFilters) ([]service.ImageCreatorManagedImage, int, error) {
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	where, args := buildImageCreatorImageListWhere(userID, filters)
	args = append(args, filters.Limit, filters.Offset)
	limitArg := len(args) - 1
	offsetArg := len(args)
	query := fmt.Sprintf(`
		SELECT images.id, images.task_id, images.user_id, images.file_path, images.output_format, images.mime_type,
			images.byte_size, images.width, images.height, images.sha256, images.revised_prompt, images.expires_at, images.created_at,
			tasks.prompt, tasks.model, tasks.size, tasks.quality,
			COUNT(*) OVER() AS total
		FROM image_creator_images AS images
		INNER JOIN image_creator_tasks AS tasks ON tasks.id = images.task_id
		WHERE `+where+`
		ORDER BY images.created_at DESC, images.id DESC
		LIMIT $%d OFFSET $%d
	`, limitArg, offsetArg)
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	images := make([]service.ImageCreatorManagedImage, 0)
	total := 0
	for rows.Next() {
		image, rowTotal, err := scanImageCreatorManagedImage(rows)
		if err != nil {
			return nil, 0, err
		}
		total = rowTotal
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return images, total, nil
}

func buildImageCreatorImageListWhere(userID int64, filters service.ImageCreatorImageListFilters) (string, []any) {
	clauses := []string{"images.user_id = $1"}
	args := []any{userID}
	addArg := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}
	if filters.Search != "" {
		value := "%" + strings.ToLower(filters.Search) + "%"
		arg := addArg(value)
		clauses = append(clauses, fmt.Sprintf(`(
			LOWER(tasks.prompt) LIKE %s OR
			LOWER(COALESCE(images.revised_prompt, '')) LIKE %s OR
			LOWER(tasks.model) LIKE %s OR
			LOWER(images.output_format) LIKE %s OR
			LOWER(images.mime_type) LIKE %s OR
			LOWER(images.sha256) LIKE %s
		)`, arg, arg, arg, arg, arg, arg))
	}
	if filters.StartDate != "" {
		if start, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			clauses = append(clauses, "images.created_at >= "+addArg(start))
		}
	}
	if filters.EndDate != "" {
		if end, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			clauses = append(clauses, "images.created_at < "+addArg(end.Add(24*time.Hour)))
		}
	}
	if filters.Format != "" {
		if filters.Format == "other" {
			clauses = append(clauses, "images.output_format NOT IN ('png', 'jpeg', 'jpg', 'webp')")
		} else if filters.Format == "jpeg" {
			clauses = append(clauses, "images.output_format IN ('jpeg', 'jpg')")
		} else {
			clauses = append(clauses, "images.output_format = "+addArg(filters.Format))
		}
	}
	if filters.MinWidth > 0 {
		clauses = append(clauses, "images.width >= "+addArg(filters.MinWidth))
	}
	if filters.MinHeight > 0 {
		clauses = append(clauses, "images.height >= "+addArg(filters.MinHeight))
	}
	if filters.Orientation != "" {
		clauses = append(clauses, imageCreatorOrientationWhere(filters.Orientation))
	}
	if filters.Resolution != "" {
		clauses = append(clauses, imageCreatorResolutionWhere(filters.Resolution))
	}
	if filters.AspectRatio != "" {
		clauses = append(clauses, imageCreatorAspectRatioWhere(filters.AspectRatio))
	}
	return strings.Join(clauses, " AND "), args
}

func imageCreatorOrientationWhere(value string) string {
	switch value {
	case "landscape":
		return "images.width > images.height"
	case "portrait":
		return "images.height > images.width"
	case "square":
		return "images.width > 0 AND images.width = images.height"
	case "unknown":
		return "(images.width <= 0 OR images.height <= 0)"
	default:
		return "TRUE"
	}
}

func imageCreatorResolutionWhere(value string) string {
	longSide := "GREATEST(images.width, images.height)"
	shortSide := "LEAST(images.width, images.height)"
	switch value {
	case "4k":
		return fmt.Sprintf("images.width > 0 AND images.height > 0 AND (%s >= 3200 OR %s >= 2400)", longSide, shortSide)
	case "2k":
		return fmt.Sprintf("images.width > 0 AND images.height > 0 AND (%s >= 1600 OR %s >= 1400) AND NOT (%s >= 3200 OR %s >= 2400)", longSide, shortSide, longSide, shortSide)
	case "1080p":
		return fmt.Sprintf("images.width > 0 AND images.height > 0 AND NOT (%s >= 1600 OR %s >= 1400)", longSide, shortSide)
	case "unknown":
		return "(images.width <= 0 OR images.height <= 0)"
	default:
		return "TRUE"
	}
}

func imageCreatorAspectRatioWhere(value string) string {
	switch value {
	case "1:1":
		return "images.width > 0 AND images.width = images.height"
	case "4:3":
		return "images.width > 0 AND images.height > 0 AND images.width * 3 = images.height * 4"
	case "3:4":
		return "images.width > 0 AND images.height > 0 AND images.width * 4 = images.height * 3"
	case "16:9":
		return "images.width > 0 AND images.height > 0 AND images.width * 9 = images.height * 16"
	case "9:16":
		return "images.width > 0 AND images.height > 0 AND images.width * 16 = images.height * 9"
	case "other":
		return `images.width > 0 AND images.height > 0
			AND images.width <> images.height
			AND images.width * 3 <> images.height * 4
			AND images.width * 4 <> images.height * 3
			AND images.width * 9 <> images.height * 16
			AND images.width * 16 <> images.height * 9`
	case "unknown":
		return "(images.width <= 0 OR images.height <= 0)"
	default:
		return "TRUE"
	}
}

func (r *imageCreatorRepository) ListImagesForUserByIDs(ctx context.Context, userID int64, ids []int64) ([]service.ImageCreatorImage, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := imageCreatorImageSelectSQL() + `
		WHERE user_id = $1 AND id = ANY($2)
		ORDER BY created_at DESC, id DESC
	`
	return r.queryImages(ctx, query, userID, pq.Array(ids))
}

func (r *imageCreatorRepository) ListPrunableImages(ctx context.Context, userID int64, keep int) ([]service.ImageCreatorImage, error) {
	if keep <= 0 {
		keep = 3
	}
	query := imageCreatorImageSelectSQL() + `
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		OFFSET $2
	`
	return r.queryImages(ctx, query, userID, keep)
}

func (r *imageCreatorRepository) DeleteImagesByID(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.sql.ExecContext(ctx, `DELETE FROM image_creator_images WHERE id = ANY($1)`, pq.Array(ids))
	return err
}

func (r *imageCreatorRepository) ListExpiredImages(ctx context.Context, before time.Time, limit int) ([]service.ImageCreatorImage, error) {
	if limit <= 0 {
		limit = 100
	}
	query := imageCreatorImageSelectSQL() + `
		WHERE expires_at < $1
		ORDER BY expires_at ASC, id ASC
		LIMIT $2
	`
	return r.queryImages(ctx, query, before, limit)
}

func (r *imageCreatorRepository) DeleteExpiredTasks(ctx context.Context, before time.Time, limit int) ([]service.ImageCreatorTask, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, user_id, api_key_id, status, model, prompt, size, quality,
			output_format, background, image_count,
			reference_image_path, reference_image_mime_type, reference_image_filename,
			error_message, started_at, completed_at, expires_at, created_at, updated_at
		FROM image_creator_tasks
		WHERE expires_at < $1
		ORDER BY expires_at ASC, id ASC
		LIMIT $2
	`
	rows, err := r.sql.QueryContext(ctx, query, before, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var tasks []service.ImageCreatorTask
	for rows.Next() {
		task, err := scanImageCreatorTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return tasks, nil
	}
	ids := make([]int64, 0, len(tasks))
	for i := range tasks {
		ids = append(ids, tasks[i].ID)
		if err := r.attachImages(ctx, &tasks[i]); err != nil {
			return nil, err
		}
	}
	if _, err := r.sql.ExecContext(ctx, `DELETE FROM image_creator_tasks WHERE id = ANY($1)`, pq.Array(ids)); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *imageCreatorRepository) FailStaleRunningTasks(ctx context.Context, staleRunningAfter time.Duration, message string) error {
	if staleRunningAfter <= 0 {
		staleRunningAfter = 30 * time.Minute
	}
	_, err := r.sql.ExecContext(ctx, `
		UPDATE image_creator_tasks
		SET status = $1,
			error_message = $2,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE status = $3
			AND started_at IS NOT NULL
			AND started_at < NOW() - ($4 * interval '1 second')
	`, service.ImageCreatorTaskStatusFailed, strings.TrimSpace(message), service.ImageCreatorTaskStatusRunning, int64(staleRunningAfter.Seconds()))
	return err
}

func (r *imageCreatorRepository) getTask(ctx context.Context, where string, args []any) (*service.ImageCreatorTask, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, api_key_id, status, model, prompt, size, quality,
			output_format, background, image_count,
			reference_image_path, reference_image_mime_type, reference_image_filename,
			error_message, started_at, completed_at, expires_at, created_at, updated_at
		FROM image_creator_tasks
		WHERE %s
	`, where)
	return r.scanTaskRow(ctx, query, args...)
}

func (r *imageCreatorRepository) scanTaskRow(ctx context.Context, query string, args ...any) (*service.ImageCreatorTask, error) {
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	task, err := scanImageCreatorTask(rows)
	if err != nil {
		return nil, err
	}
	return &task, rows.Err()
}

func (r *imageCreatorRepository) attachImages(ctx context.Context, task *service.ImageCreatorTask) error {
	if task == nil {
		return nil
	}
	query := imageCreatorImageSelectSQL() + `
		WHERE task_id = $1
		ORDER BY created_at ASC, id ASC
	`
	images, err := r.queryImages(ctx, query, task.ID)
	if err != nil {
		return err
	}
	task.Images = images
	return nil
}

func (r *imageCreatorRepository) queryImages(ctx context.Context, query string, args ...any) ([]service.ImageCreatorImage, error) {
	return queryImageCreatorImages(ctx, r.sql, query, args...)
}

func queryImageCreatorImages(ctx context.Context, exec sqlExecutor, query string, args ...any) ([]service.ImageCreatorImage, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	images := make([]service.ImageCreatorImage, 0)
	for rows.Next() {
		image, err := scanImageCreatorImage(rows)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, rows.Err()
}

type imageCreatorTaskScanner interface {
	Scan(dest ...any) error
}

func scanImageCreatorTask(row imageCreatorTaskScanner) (service.ImageCreatorTask, error) {
	var task service.ImageCreatorTask
	var refPath, refMime, refFilename, errMsg sql.NullString
	var startedAt, completedAt sql.NullTime
	err := row.Scan(
		&task.ID,
		&task.UserID,
		&task.APIKeyID,
		&task.Status,
		&task.Model,
		&task.Prompt,
		&task.Size,
		&task.Quality,
		&task.OutputFormat,
		&task.Background,
		&task.Count,
		&refPath,
		&refMime,
		&refFilename,
		&errMsg,
		&startedAt,
		&completedAt,
		&task.ExpiresAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return task, err
	}
	task.ReferenceImagePath = nullStringValue(refPath)
	task.ReferenceImageMimeType = nullStringValue(refMime)
	task.ReferenceImageFilename = nullStringValue(refFilename)
	task.ErrorMessage = nullStringValue(errMsg)
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	return task, nil
}

func imageCreatorImageSelectSQL() string {
	return `
		SELECT id, task_id, user_id, file_path, output_format, mime_type,
			byte_size, width, height, sha256, revised_prompt, expires_at, created_at
		FROM image_creator_images
	`
}

func scanImageCreatorImage(row imageCreatorTaskScanner) (service.ImageCreatorImage, error) {
	var image service.ImageCreatorImage
	var revisedPrompt sql.NullString
	err := row.Scan(
		&image.ID,
		&image.TaskID,
		&image.UserID,
		&image.FilePath,
		&image.OutputFormat,
		&image.MimeType,
		&image.ByteSize,
		&image.Width,
		&image.Height,
		&image.SHA256,
		&revisedPrompt,
		&image.ExpiresAt,
		&image.CreatedAt,
	)
	if err != nil {
		return image, err
	}
	image.RevisedPrompt = nullStringValue(revisedPrompt)
	return image, nil
}

func scanImageCreatorManagedImage(row imageCreatorTaskScanner) (service.ImageCreatorManagedImage, int, error) {
	var image service.ImageCreatorManagedImage
	var revisedPrompt sql.NullString
	var total int
	err := row.Scan(
		&image.ID,
		&image.TaskID,
		&image.UserID,
		&image.FilePath,
		&image.OutputFormat,
		&image.MimeType,
		&image.ByteSize,
		&image.Width,
		&image.Height,
		&image.SHA256,
		&revisedPrompt,
		&image.ExpiresAt,
		&image.CreatedAt,
		&image.TaskPrompt,
		&image.TaskModel,
		&image.TaskSize,
		&image.TaskQuality,
		&total,
	)
	if err != nil {
		return image, 0, err
	}
	image.RevisedPrompt = nullStringValue(revisedPrompt)
	return image, total, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
