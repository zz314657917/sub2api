package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type TutorialPage struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	SortOrder   int        `json:"sort_order"`
	Status      string     `json:"status"`
	ContentMD   string     `json:"content_md"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type TutorialPageSummary struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	SortOrder   int        `json:"sort_order"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

func TutorialPageFromService(page *service.TutorialPage) *TutorialPage {
	if page == nil {
		return nil
	}
	return &TutorialPage{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Description: page.Description,
		Category:    page.Category,
		SortOrder:   page.SortOrder,
		Status:      page.Status,
		ContentMD:   page.ContentMD,
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
		PublishedAt: page.PublishedAt,
	}
}

func TutorialPageSummaryFromService(page *service.TutorialPage) *TutorialPageSummary {
	if page == nil {
		return nil
	}
	return &TutorialPageSummary{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Description: page.Description,
		Category:    page.Category,
		SortOrder:   page.SortOrder,
		Status:      page.Status,
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
		PublishedAt: page.PublishedAt,
	}
}

func TutorialPageSummariesFromService(items []service.TutorialPage) []TutorialPageSummary {
	out := make([]TutorialPageSummary, 0, len(items))
	for i := range items {
		out = append(out, *TutorialPageSummaryFromService(&items[i]))
	}
	return out
}
