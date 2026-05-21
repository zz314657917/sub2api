package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	TutorialPageStatusDraft     = "draft"
	TutorialPageStatusPublished = "published"
)

var tutorialSlugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,78}[a-z0-9])?$`)

var (
	ErrTutorialPageNotFound        = infraerrors.NotFound("TUTORIAL_PAGE_NOT_FOUND", "tutorial page not found")
	ErrTutorialPageNilInput        = infraerrors.BadRequest("TUTORIAL_PAGE_INPUT_REQUIRED", "tutorial page input is required")
	ErrTutorialPageInvalidSlug     = infraerrors.BadRequest("TUTORIAL_PAGE_SLUG_INVALID", "tutorial page slug is invalid")
	ErrTutorialPageInvalidTitle    = infraerrors.BadRequest("TUTORIAL_PAGE_TITLE_INVALID", "tutorial page title is invalid")
	ErrTutorialPageInvalidStatus   = infraerrors.BadRequest("TUTORIAL_PAGE_STATUS_INVALID", "tutorial page status is invalid")
	ErrTutorialPageContentRequired = infraerrors.BadRequest("TUTORIAL_PAGE_CONTENT_REQUIRED", "tutorial page content is required")
	ErrTutorialPageSlugExists      = infraerrors.BadRequest("TUTORIAL_PAGE_SLUG_EXISTS", "tutorial page slug already exists")
)

type TutorialPage struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	Category    string
	SortOrder   int
	Status      string
	ContentMD   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

type TutorialPageListFilters struct {
	Status   string
	Category string
	Search   string
}

type CreateTutorialPageInput struct {
	Slug        string
	Title       string
	Description string
	Category    string
	SortOrder   int
	Status      string
	ContentMD   string
}

type UpdateTutorialPageInput struct {
	Slug        *string
	Title       *string
	Description *string
	Category    *string
	SortOrder   *int
	Status      *string
	ContentMD   *string
}

type TutorialPageRepository interface {
	Create(ctx context.Context, page *TutorialPage) error
	GetByID(ctx context.Context, id int64) (*TutorialPage, error)
	GetBySlug(ctx context.Context, slug string) (*TutorialPage, error)
	Update(ctx context.Context, page *TutorialPage) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params pagination.PaginationParams, filters TutorialPageListFilters) ([]TutorialPage, *pagination.PaginationResult, error)
	ListPublished(ctx context.Context) ([]TutorialPage, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*TutorialPage, error)
}

type TutorialPageService struct {
	repo TutorialPageRepository
}

func NewTutorialPageService(repo TutorialPageRepository) *TutorialPageService {
	return &TutorialPageService{repo: repo}
}

func (s *TutorialPageService) Create(ctx context.Context, input CreateTutorialPageInput) (*TutorialPage, error) {
	page, err := normalizeCreateTutorialPage(input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *TutorialPageService) GetByID(ctx context.Context, id int64) (*TutorialPage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TutorialPageService) GetBySlug(ctx context.Context, slug string) (*TutorialPage, error) {
	slug = strings.TrimSpace(slug)
	if !IsValidTutorialSlug(slug) {
		return nil, ErrTutorialPageInvalidSlug
	}
	return s.repo.GetBySlug(ctx, slug)
}

func (s *TutorialPageService) GetPublishedBySlug(ctx context.Context, slug string) (*TutorialPage, error) {
	slug = strings.TrimSpace(slug)
	if !IsValidTutorialSlug(slug) {
		return nil, ErrTutorialPageInvalidSlug
	}
	return s.repo.GetPublishedBySlug(ctx, slug)
}

func (s *TutorialPageService) List(ctx context.Context, params pagination.PaginationParams, filters TutorialPageListFilters) ([]TutorialPage, *pagination.PaginationResult, error) {
	filters.Status = strings.TrimSpace(filters.Status)
	if filters.Status != "" && !isValidTutorialStatus(filters.Status) {
		return nil, nil, ErrTutorialPageInvalidStatus
	}
	filters.Category = trimLimited(filters.Category, 80)
	filters.Search = trimLimited(filters.Search, 200)
	return s.repo.List(ctx, params, filters)
}

func (s *TutorialPageService) ListPublished(ctx context.Context) ([]TutorialPage, error) {
	return s.repo.ListPublished(ctx)
}

func (s *TutorialPageService) Update(ctx context.Context, id int64, input UpdateTutorialPageInput) (*TutorialPage, error) {
	page, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Slug != nil {
		page.Slug = strings.TrimSpace(*input.Slug)
	}
	if input.Title != nil {
		page.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		page.Description = trimLimited(*input.Description, 500)
	}
	if input.Category != nil {
		page.Category = trimLimited(*input.Category, 80)
	}
	if input.SortOrder != nil {
		page.SortOrder = *input.SortOrder
	}
	if input.Status != nil {
		page.Status = strings.TrimSpace(*input.Status)
	}
	if input.ContentMD != nil {
		page.ContentMD = strings.TrimSpace(*input.ContentMD)
	}
	if err := validateTutorialPage(page); err != nil {
		return nil, err
	}
	applyPublishedAt(page, time.Now().UTC())
	if err := s.repo.Update(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *TutorialPageService) SetStatus(ctx context.Context, id int64, status string) (*TutorialPage, error) {
	status = strings.TrimSpace(status)
	if !isValidTutorialStatus(status) {
		return nil, ErrTutorialPageInvalidStatus
	}
	return s.Update(ctx, id, UpdateTutorialPageInput{Status: &status})
}

func (s *TutorialPageService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func IsValidTutorialSlug(slug string) bool {
	return tutorialSlugPattern.MatchString(slug)
}

func normalizeCreateTutorialPage(input CreateTutorialPageInput) (*TutorialPage, error) {
	page := &TutorialPage{
		Slug:        strings.TrimSpace(input.Slug),
		Title:       strings.TrimSpace(input.Title),
		Description: trimLimited(input.Description, 500),
		Category:    trimLimited(input.Category, 80),
		SortOrder:   input.SortOrder,
		Status:      strings.TrimSpace(input.Status),
		ContentMD:   strings.TrimSpace(input.ContentMD),
	}
	if page.Status == "" {
		page.Status = TutorialPageStatusDraft
	}
	if err := validateTutorialPage(page); err != nil {
		return nil, err
	}
	applyPublishedAt(page, time.Now().UTC())
	return page, nil
}

func validateTutorialPage(page *TutorialPage) error {
	if page == nil {
		return ErrTutorialPageNilInput
	}
	if !IsValidTutorialSlug(page.Slug) {
		return ErrTutorialPageInvalidSlug
	}
	if strings.TrimSpace(page.Title) == "" || len([]rune(page.Title)) > 160 {
		return ErrTutorialPageInvalidTitle
	}
	if !isValidTutorialStatus(page.Status) {
		return ErrTutorialPageInvalidStatus
	}
	if strings.TrimSpace(page.ContentMD) == "" {
		return ErrTutorialPageContentRequired
	}
	return nil
}

func isValidTutorialStatus(status string) bool {
	return status == TutorialPageStatusDraft || status == TutorialPageStatusPublished
}

func applyPublishedAt(page *TutorialPage, now time.Time) {
	if page.Status == TutorialPageStatusPublished {
		if page.PublishedAt == nil {
			t := now
			page.PublishedAt = &t
		}
		return
	}
	page.PublishedAt = nil
}

func trimLimited(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}
