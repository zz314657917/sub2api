package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type tutorialPageRepoStub struct {
	pages []TutorialPage
}

func (r *tutorialPageRepoStub) Create(ctx context.Context, page *TutorialPage) error {
	cp := *page
	if cp.ID == 0 {
		cp.ID = int64(len(r.pages) + 1)
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now().UTC()
	}
	cp.UpdatedAt = cp.CreatedAt
	*page = cp
	r.pages = append(r.pages, cp)
	return nil
}

func (r *tutorialPageRepoStub) GetByID(ctx context.Context, id int64) (*TutorialPage, error) {
	for i := range r.pages {
		if r.pages[i].ID == id {
			cp := r.pages[i]
			return &cp, nil
		}
	}
	return nil, ErrTutorialPageNotFound
}

func (r *tutorialPageRepoStub) GetBySlug(ctx context.Context, slug string) (*TutorialPage, error) {
	for i := range r.pages {
		if r.pages[i].Slug == slug {
			cp := r.pages[i]
			return &cp, nil
		}
	}
	return nil, ErrTutorialPageNotFound
}

func (r *tutorialPageRepoStub) GetPublishedBySlug(ctx context.Context, slug string) (*TutorialPage, error) {
	page, err := r.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if page.Status != TutorialPageStatusPublished {
		return nil, ErrTutorialPageNotFound
	}
	return page, nil
}

func (r *tutorialPageRepoStub) List(ctx context.Context, params pagination.PaginationParams, filters TutorialPageListFilters) ([]TutorialPage, *pagination.PaginationResult, error) {
	items := make([]TutorialPage, 0, len(r.pages))
	for _, page := range r.pages {
		if filters.Status != "" && page.Status != filters.Status {
			continue
		}
		items = append(items, page)
	}
	return items, &pagination.PaginationResult{Total: int64(len(items)), Page: params.Page, PageSize: params.PageSize}, nil
}

func (r *tutorialPageRepoStub) ListPublished(ctx context.Context) ([]TutorialPage, error) {
	items := make([]TutorialPage, 0, len(r.pages))
	for _, page := range r.pages {
		if page.Status == TutorialPageStatusPublished {
			items = append(items, page)
		}
	}
	return items, nil
}

func (r *tutorialPageRepoStub) Update(ctx context.Context, page *TutorialPage) error {
	for i := range r.pages {
		if r.pages[i].ID == page.ID {
			cp := *page
			cp.UpdatedAt = time.Now().UTC()
			*page = cp
			r.pages[i] = cp
			return nil
		}
	}
	return ErrTutorialPageNotFound
}

func (r *tutorialPageRepoStub) Delete(ctx context.Context, id int64) error {
	for i := range r.pages {
		if r.pages[i].ID == id {
			r.pages = append(r.pages[:i], r.pages[i+1:]...)
			return nil
		}
	}
	return ErrTutorialPageNotFound
}

func TestTutorialPageServiceCreateValidatesSlugAndContent(t *testing.T) {
	svc := NewTutorialPageService(&tutorialPageRepoStub{})

	_, err := svc.Create(context.Background(), CreateTutorialPageInput{
		Slug:      "Bad_Slug",
		Title:     "教程",
		ContentMD: "正文",
	})
	if !errors.Is(err, ErrTutorialPageInvalidSlug) {
		t.Fatalf("expected invalid slug, got %v", err)
	}

	_, err = svc.Create(context.Background(), CreateTutorialPageInput{
		Slug:      "valid-slug",
		Title:     "教程",
		ContentMD: "",
	})
	if !errors.Is(err, ErrTutorialPageContentRequired) {
		t.Fatalf("expected content required, got %v", err)
	}
}

func TestTutorialPageServicePublishedAccessAndStatusSwitch(t *testing.T) {
	repo := &tutorialPageRepoStub{}
	svc := NewTutorialPageService(repo)

	page, err := svc.Create(context.Background(), CreateTutorialPageInput{
		Slug:      "codex",
		Title:     "Codex",
		ContentMD: "# Codex",
		Status:    TutorialPageStatusDraft,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	if _, err = svc.GetPublishedBySlug(context.Background(), "codex"); !errors.Is(err, ErrTutorialPageNotFound) {
		t.Fatalf("draft should not be publicly visible, got %v", err)
	}

	published, err := svc.SetStatus(context.Background(), page.ID, TutorialPageStatusPublished)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.PublishedAt == nil {
		t.Fatalf("expected published_at to be set")
	}

	publicPage, err := svc.GetPublishedBySlug(context.Background(), "codex")
	if err != nil {
		t.Fatalf("published page should be visible: %v", err)
	}
	if publicPage.Slug != "codex" {
		t.Fatalf("unexpected public slug %q", publicPage.Slug)
	}

	unpublished, err := svc.SetStatus(context.Background(), page.ID, TutorialPageStatusDraft)
	if err != nil {
		t.Fatalf("unpublish: %v", err)
	}
	if unpublished.PublishedAt != nil {
		t.Fatalf("expected published_at to be cleared")
	}
}

func TestTutorialPageServicePreservesPublishedAtOnContentUpdate(t *testing.T) {
	now := time.Now().Add(-time.Hour).UTC()
	repo := &tutorialPageRepoStub{
		pages: []TutorialPage{{
			ID:          10,
			Slug:        "codex",
			Title:       "Codex",
			Status:      TutorialPageStatusPublished,
			ContentMD:   "# Old",
			PublishedAt: &now,
		}},
	}
	svc := NewTutorialPageService(repo)
	title := "Codex 新版"
	content := "# New"
	status := TutorialPageStatusPublished

	updated, err := svc.Update(context.Background(), 10, UpdateTutorialPageInput{
		Title:     &title,
		ContentMD: &content,
		Status:    &status,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.PublishedAt == nil || !updated.PublishedAt.Equal(now) {
		t.Fatalf("expected published_at to be preserved, got %v", updated.PublishedAt)
	}
}
