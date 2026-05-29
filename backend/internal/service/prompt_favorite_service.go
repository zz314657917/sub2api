package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	PromptFavoriteModeGenerate = "generate"
	PromptFavoriteModeEdit     = "edit"
)

type PromptFavorite struct {
	ID                 int64                         `json:"id"`
	UserID             int64                         `json:"user_id"`
	PromptID           string                        `json:"prompt_id"`
	Source             string                        `json:"source"`
	Title              string                        `json:"title"`
	Preview            string                        `json:"preview"`
	ReferenceImageURLs []string                      `json:"reference_image_urls"`
	Prompt             string                        `json:"prompt"`
	Author             string                        `json:"author"`
	Link               string                        `json:"link"`
	Mode               string                        `json:"mode"`
	Category           string                        `json:"category"`
	SubCategory        string                        `json:"sub_category"`
	Created            string                        `json:"created"`
	SourceLabel        string                        `json:"source_label"`
	IsNSFW             bool                          `json:"is_nsfw"`
	Localizations      map[string]PromptLocalization `json:"localizations"`
	FavoritedAt        time.Time                     `json:"favorited_at"`
	UpdatedAt          time.Time                     `json:"updated_at"`
}

type PromptFavoriteInput struct {
	PromptID           string                        `json:"prompt_id"`
	Source             string                        `json:"source"`
	Title              string                        `json:"title"`
	Preview            string                        `json:"preview"`
	ReferenceImageURLs []string                      `json:"reference_image_urls"`
	Prompt             string                        `json:"prompt"`
	Author             string                        `json:"author"`
	Link               string                        `json:"link"`
	Mode               string                        `json:"mode"`
	Category           string                        `json:"category"`
	SubCategory        string                        `json:"sub_category"`
	Created            string                        `json:"created"`
	SourceLabel        string                        `json:"source_label"`
	IsNSFW             bool                          `json:"is_nsfw"`
	Localizations      map[string]PromptLocalization `json:"localizations"`
}

type PromptLocalization struct {
	Title       string `json:"title,omitempty"`
	Prompt      string `json:"prompt,omitempty"`
	Category    string `json:"category,omitempty"`
	SubCategory string `json:"sub_category,omitempty"`
	Created     string `json:"created,omitempty"`
}

type PromptFavoriteRepository interface {
	ListPromptFavorites(ctx context.Context, userID int64) ([]PromptFavorite, error)
	UpsertPromptFavorite(ctx context.Context, userID int64, input PromptFavoriteInput) (*PromptFavorite, error)
	DeletePromptFavorite(ctx context.Context, userID int64, favoriteID int64) error
}

type PromptFavoriteService struct {
	repo PromptFavoriteRepository
}

func NewPromptFavoriteService(repo PromptFavoriteRepository) *PromptFavoriteService {
	return &PromptFavoriteService{repo: repo}
}

func (s *PromptFavoriteService) List(ctx context.Context, userID int64) ([]PromptFavorite, error) {
	if err := validatePromptFavoriteUser(userID); err != nil {
		return nil, err
	}
	return s.repo.ListPromptFavorites(ctx, userID)
}

func (s *PromptFavoriteService) Save(ctx context.Context, userID int64, input PromptFavoriteInput) (*PromptFavorite, error) {
	if err := validatePromptFavoriteUser(userID); err != nil {
		return nil, err
	}
	normalized := normalizePromptFavoriteInput(input)
	if err := validatePromptFavoriteInput(normalized); err != nil {
		return nil, err
	}
	return s.repo.UpsertPromptFavorite(ctx, userID, normalized)
}

func (s *PromptFavoriteService) Delete(ctx context.Context, userID int64, favoriteID int64) error {
	if err := validatePromptFavoriteUser(userID); err != nil {
		return err
	}
	if favoriteID <= 0 {
		return infraerrors.BadRequest("INVALID_PROMPT_FAVORITE_ID", "invalid prompt favorite id")
	}
	return s.repo.DeletePromptFavorite(ctx, userID, favoriteID)
}

func normalizePromptFavoriteInput(input PromptFavoriteInput) PromptFavoriteInput {
	input.PromptID = strings.TrimSpace(input.PromptID)
	input.Source = strings.TrimSpace(input.Source)
	input.Title = strings.TrimSpace(input.Title)
	input.Preview = strings.TrimSpace(input.Preview)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.Author = strings.TrimSpace(input.Author)
	input.Link = strings.TrimSpace(input.Link)
	input.Mode = normalizePromptFavoriteMode(input.Mode)
	input.Category = strings.TrimSpace(input.Category)
	input.SubCategory = strings.TrimSpace(input.SubCategory)
	input.Created = strings.TrimSpace(input.Created)
	input.SourceLabel = strings.TrimSpace(input.SourceLabel)
	input.ReferenceImageURLs = normalizePromptFavoriteReferenceURLs(input.ReferenceImageURLs)
	input.Localizations = normalizePromptFavoriteLocalizations(input.Localizations)
	return input
}

func normalizePromptFavoriteMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case PromptFavoriteModeEdit:
		return PromptFavoriteModeEdit
	default:
		return PromptFavoriteModeGenerate
	}
}

func normalizePromptFavoriteReferenceURLs(values []string) []string {
	const maxReferenceURLs = 16
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
		if len(out) >= maxReferenceURLs {
			break
		}
	}
	return out
}

func normalizePromptFavoriteLocalizations(values map[string]PromptLocalization) map[string]PromptLocalization {
	if len(values) == 0 {
		return map[string]PromptLocalization{}
	}
	out := make(map[string]PromptLocalization, len(values))
	for language, value := range values {
		language = strings.TrimSpace(language)
		if language == "" {
			continue
		}
		value.Title = strings.TrimSpace(value.Title)
		value.Prompt = strings.TrimSpace(value.Prompt)
		value.Category = strings.TrimSpace(value.Category)
		value.SubCategory = strings.TrimSpace(value.SubCategory)
		value.Created = strings.TrimSpace(value.Created)
		out[language] = value
	}
	if len(out) == 0 {
		return map[string]PromptLocalization{}
	}
	return out
}

func validatePromptFavoriteUser(userID int64) error {
	if userID <= 0 {
		return infraerrors.Unauthorized("UNAUTHORIZED", "authentication required")
	}
	return nil
}

func validatePromptFavoriteInput(input PromptFavoriteInput) error {
	if input.PromptID == "" {
		return infraerrors.BadRequest("PROMPT_ID_REQUIRED", "prompt id is required")
	}
	if input.Source == "" {
		return infraerrors.BadRequest("PROMPT_SOURCE_REQUIRED", "prompt source is required")
	}
	if input.Title == "" {
		return infraerrors.BadRequest("PROMPT_TITLE_REQUIRED", "prompt title is required")
	}
	if input.Prompt == "" {
		return infraerrors.BadRequest("PROMPT_REQUIRED", "prompt is required")
	}
	return nil
}
