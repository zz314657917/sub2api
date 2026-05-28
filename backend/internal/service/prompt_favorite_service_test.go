package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

type fakePromptFavoriteRepo struct {
	items        []PromptFavorite
	lastUserID   int64
	lastInput    PromptFavoriteInput
	deletedID    int64
	deleteUserID int64
}

func (r *fakePromptFavoriteRepo) ListPromptFavorites(_ context.Context, userID int64) ([]PromptFavorite, error) {
	r.lastUserID = userID
	out := make([]PromptFavorite, 0, len(r.items))
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakePromptFavoriteRepo) UpsertPromptFavorite(_ context.Context, userID int64, input PromptFavoriteInput) (*PromptFavorite, error) {
	r.lastUserID = userID
	r.lastInput = input
	item := PromptFavorite{
		ID:                 12,
		UserID:             userID,
		PromptID:           input.PromptID,
		Source:             input.Source,
		Title:              input.Title,
		Prompt:             input.Prompt,
		Mode:               input.Mode,
		ReferenceImageURLs: input.ReferenceImageURLs,
		Localizations:      input.Localizations,
	}
	return &item, nil
}

func (r *fakePromptFavoriteRepo) DeletePromptFavorite(_ context.Context, userID int64, favoriteID int64) error {
	r.deleteUserID = userID
	r.deletedID = favoriteID
	return nil
}

func TestPromptFavoriteServiceSaveNormalizesAndUsesCurrentUser(t *testing.T) {
	repo := &fakePromptFavoriteRepo{}
	svc := NewPromptFavoriteService(repo)

	item, err := svc.Save(context.Background(), 42, PromptFavoriteInput{
		PromptID:           " prompt-1 ",
		Source:             " banana-prompt-quicker ",
		Title:              " Poster ",
		Prompt:             " draw poster ",
		Mode:               "edit",
		ReferenceImageURLs: []string{" https://example.com/a.png ", "", "https://example.com/a.png", "https://example.com/b.png"},
		Localizations: map[string]PromptLocalization{
			" zh-CN ": {Title: " 海报 ", Prompt: " 画海报 "},
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(42), repo.lastUserID)
	require.Equal(t, "prompt-1", repo.lastInput.PromptID)
	require.Equal(t, "banana-prompt-quicker", repo.lastInput.Source)
	require.Equal(t, "Poster", repo.lastInput.Title)
	require.Equal(t, "draw poster", repo.lastInput.Prompt)
	require.Equal(t, PromptFavoriteModeEdit, repo.lastInput.Mode)
	require.Equal(t, []string{"https://example.com/a.png", "https://example.com/b.png"}, repo.lastInput.ReferenceImageURLs)
	require.Contains(t, repo.lastInput.Localizations, "zh-CN")
	require.Equal(t, "海报", repo.lastInput.Localizations["zh-CN"].Title)
	require.Equal(t, int64(42), item.UserID)
}

func TestPromptFavoriteServiceSaveRejectsMissingPrompt(t *testing.T) {
	svc := NewPromptFavoriteService(&fakePromptFavoriteRepo{})

	_, err := svc.Save(context.Background(), 42, PromptFavoriteInput{
		PromptID: "prompt-1",
		Source:   "source",
		Title:    "title",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
}

func TestPromptFavoriteServiceDeleteUsesCurrentUser(t *testing.T) {
	repo := &fakePromptFavoriteRepo{}
	svc := NewPromptFavoriteService(repo)

	require.NoError(t, svc.Delete(context.Background(), 42, 12))

	require.Equal(t, int64(42), repo.deleteUserID)
	require.Equal(t, int64(12), repo.deletedID)
}
