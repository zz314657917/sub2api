package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type tutorialPageRepository struct {
	db *sql.DB
}

func NewTutorialPageRepository(db *sql.DB) service.TutorialPageRepository {
	return &tutorialPageRepository{db: db}
}

func (r *tutorialPageRepository) Create(ctx context.Context, page *service.TutorialPage) error {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, page.Slug, page.Title, page.Description, page.Category, page.SortOrder, page.Status, page.ContentMD, page.PublishedAt)
	if err := row.Scan(&page.ID, &page.CreatedAt, &page.UpdatedAt); err != nil {
		return translatePersistenceError(err, service.ErrTutorialPageNotFound, service.ErrTutorialPageSlugExists)
	}
	return nil
}

func (r *tutorialPageRepository) GetByID(ctx context.Context, id int64) (*service.TutorialPage, error) {
	return r.scanOne(ctx, `
		SELECT id, slug, title, description, category, sort_order, status, content_md, created_at, updated_at, published_at
		FROM tutorial_pages
		WHERE id = $1
	`, id)
}

func (r *tutorialPageRepository) GetBySlug(ctx context.Context, slug string) (*service.TutorialPage, error) {
	return r.scanOne(ctx, `
		SELECT id, slug, title, description, category, sort_order, status, content_md, created_at, updated_at, published_at
		FROM tutorial_pages
		WHERE slug = $1
	`, slug)
}

func (r *tutorialPageRepository) GetPublishedBySlug(ctx context.Context, slug string) (*service.TutorialPage, error) {
	return r.scanOne(ctx, `
		SELECT id, slug, title, description, category, sort_order, status, content_md, created_at, updated_at, published_at
		FROM tutorial_pages
		WHERE slug = $1 AND status = $2
	`, slug, service.TutorialPageStatusPublished)
}

func (r *tutorialPageRepository) Update(ctx context.Context, page *service.TutorialPage) error {
	row := r.db.QueryRowContext(ctx, `
		UPDATE tutorial_pages
		SET slug = $2,
		    title = $3,
		    description = $4,
		    category = $5,
		    sort_order = $6,
		    status = $7,
		    content_md = $8,
		    published_at = $9,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`, page.ID, page.Slug, page.Title, page.Description, page.Category, page.SortOrder, page.Status, page.ContentMD, page.PublishedAt)
	if err := row.Scan(&page.UpdatedAt); err != nil {
		return translatePersistenceError(err, service.ErrTutorialPageNotFound, service.ErrTutorialPageSlugExists)
	}
	return nil
}

func (r *tutorialPageRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tutorial_pages WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return service.ErrTutorialPageNotFound
	}
	return err
}

func (r *tutorialPageRepository) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.TutorialPageListFilters,
) ([]service.TutorialPage, *pagination.PaginationResult, error) {
	where, args := tutorialPageWhere(filters, false)
	countQuery := "SELECT COUNT(*) FROM tutorial_pages" + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	sortBy, sortOrder := tutorialPageSort(params)
	args = append(args, params.Limit(), params.Offset())
	query := fmt.Sprintf(`
		SELECT id, slug, title, description, category, sort_order, status, content_md, created_at, updated_at, published_at
		FROM tutorial_pages%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d
	`, where, sortBy, sortOrder, len(args)-1, len(args))
	items, err := r.scanMany(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *tutorialPageRepository) ListPublished(ctx context.Context) ([]service.TutorialPage, error) {
	return r.scanMany(ctx, `
		SELECT id, slug, title, description, category, sort_order, status, content_md, created_at, updated_at, published_at
		FROM tutorial_pages
		WHERE status = $1
		ORDER BY sort_order ASC, category ASC, id ASC
	`, service.TutorialPageStatusPublished)
}

func tutorialPageWhere(filters service.TutorialPageListFilters, publishedOnly bool) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if publishedOnly {
		args = append(args, service.TutorialPageStatusPublished)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	} else if filters.Status != "" {
		args = append(args, filters.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if filters.Category != "" {
		args = append(args, filters.Category)
		clauses = append(clauses, fmt.Sprintf("category = $%d", len(args)))
	}
	if filters.Search != "" {
		args = append(args, "%"+strings.ToLower(filters.Search)+"%")
		clauses = append(clauses, fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(slug) LIKE $%d OR LOWER(description) LIKE $%d OR LOWER(content_md) LIKE $%d)", len(args), len(args), len(args), len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func tutorialPageSort(params pagination.PaginationParams) (string, string) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := strings.ToUpper(params.NormalizedSortOrder(pagination.SortOrderAsc))
	switch sortBy {
	case "slug":
		return "slug", sortOrder
	case "title":
		return "title", sortOrder
	case "category":
		return "category", sortOrder
	case "status":
		return "status", sortOrder
	case "created_at":
		return "created_at", sortOrder
	case "updated_at":
		return "updated_at", sortOrder
	case "published_at":
		return "published_at", sortOrder
	case "", "sort_order":
		return "sort_order", sortOrder
	default:
		return "sort_order", "ASC"
	}
}

func (r *tutorialPageRepository) scanOne(ctx context.Context, query string, args ...any) (*service.TutorialPage, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	page, err := scanTutorialPage(row)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrTutorialPageNotFound, service.ErrTutorialPageSlugExists)
	}
	return page, nil
}

func (r *tutorialPageRepository) scanMany(ctx context.Context, query string, args ...any) ([]service.TutorialPage, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]service.TutorialPage, 0)
	for rows.Next() {
		page, err := scanTutorialPage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *page)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type tutorialPageScanner interface {
	Scan(dest ...any) error
}

func scanTutorialPage(scanner tutorialPageScanner) (*service.TutorialPage, error) {
	var page service.TutorialPage
	var publishedAt sql.NullTime
	if err := scanner.Scan(
		&page.ID,
		&page.Slug,
		&page.Title,
		&page.Description,
		&page.Category,
		&page.SortOrder,
		&page.Status,
		&page.ContentMD,
		&page.CreatedAt,
		&page.UpdatedAt,
		&publishedAt,
	); err != nil {
		return nil, err
	}
	if publishedAt.Valid {
		t := publishedAt.Time
		page.PublishedAt = &t
	}
	return &page, nil
}
