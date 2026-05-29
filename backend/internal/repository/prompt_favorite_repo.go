package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promptFavoriteRepository struct {
	sql sqlExecutor
}

func NewPromptFavoriteRepository(sqlDB *sql.DB) service.PromptFavoriteRepository {
	return &promptFavoriteRepository{sql: sqlDB}
}

func (r *promptFavoriteRepository) ListPromptFavorites(ctx context.Context, userID int64) ([]service.PromptFavorite, error) {
	query := promptFavoriteSelectSQL() + `
		WHERE user_id = $1
		ORDER BY favorited_at DESC, id DESC
	`
	rows, err := r.sql.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.PromptFavorite, 0)
	for rows.Next() {
		item, err := scanPromptFavorite(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *promptFavoriteRepository) UpsertPromptFavorite(ctx context.Context, userID int64, input service.PromptFavoriteInput) (*service.PromptFavorite, error) {
	referenceURLs, err := json.Marshal(input.ReferenceImageURLs)
	if err != nil {
		return nil, err
	}
	localizations, err := json.Marshal(input.Localizations)
	if err != nil {
		return nil, err
	}
	query := `
		INSERT INTO prompt_favorites (
			user_id, prompt_id, source, title, preview, reference_image_urls, prompt,
			author, link, mode, category, sub_category, created, source_label,
			is_nsfw, localizations
		)
		VALUES (
			$1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb
		)
		ON CONFLICT (user_id, source, prompt_id) DO UPDATE SET
			title = EXCLUDED.title,
			preview = EXCLUDED.preview,
			reference_image_urls = EXCLUDED.reference_image_urls,
			prompt = EXCLUDED.prompt,
			author = EXCLUDED.author,
			link = EXCLUDED.link,
			mode = EXCLUDED.mode,
			category = EXCLUDED.category,
			sub_category = EXCLUDED.sub_category,
			created = EXCLUDED.created,
			source_label = EXCLUDED.source_label,
			is_nsfw = EXCLUDED.is_nsfw,
			localizations = EXCLUDED.localizations,
			updated_at = NOW()
		RETURNING id, user_id, prompt_id, source, title, preview, reference_image_urls,
			prompt, author, link, mode, category, sub_category, created, source_label,
			is_nsfw, localizations, favorited_at, updated_at
	`
	var item service.PromptFavorite
	if err := scanSingleRow(ctx, r.sql, query, []any{
		userID,
		input.PromptID,
		input.Source,
		input.Title,
		input.Preview,
		string(referenceURLs),
		input.Prompt,
		input.Author,
		input.Link,
		input.Mode,
		input.Category,
		input.SubCategory,
		input.Created,
		input.SourceLabel,
		input.IsNSFW,
		string(localizations),
	}, promptFavoriteScanDest(&item)...); err != nil {
		return nil, err
	}
	if err := decodePromptFavoriteJSON(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *promptFavoriteRepository) DeletePromptFavorite(ctx context.Context, userID int64, favoriteID int64) error {
	_, err := r.sql.ExecContext(ctx, `DELETE FROM prompt_favorites WHERE id = $1 AND user_id = $2`, favoriteID, userID)
	return err
}

func promptFavoriteSelectSQL() string {
	return `
		SELECT id, user_id, prompt_id, source, title, preview, reference_image_urls,
			prompt, author, link, mode, category, sub_category, created, source_label,
			is_nsfw, localizations, favorited_at, updated_at
		FROM prompt_favorites
	`
}

func scanPromptFavorite(row imageCreatorTaskScanner) (service.PromptFavorite, error) {
	var item service.PromptFavorite
	if err := row.Scan(promptFavoriteScanDest(&item)...); err != nil {
		return item, err
	}
	if err := decodePromptFavoriteJSON(&item); err != nil {
		return item, err
	}
	return item, nil
}

func promptFavoriteScanDest(item *service.PromptFavorite) []any {
	return []any{
		&item.ID,
		&item.UserID,
		&item.PromptID,
		&item.Source,
		&item.Title,
		&item.Preview,
		(*promptFavoriteRawReferenceURLs)(item),
		&item.Prompt,
		&item.Author,
		&item.Link,
		&item.Mode,
		&item.Category,
		&item.SubCategory,
		&item.Created,
		&item.SourceLabel,
		&item.IsNSFW,
		(*promptFavoriteRawLocalizations)(item),
		&item.FavoritedAt,
		&item.UpdatedAt,
	}
}

type promptFavoriteRawReferenceURLs service.PromptFavorite
type promptFavoriteRawLocalizations service.PromptFavorite

func (p *promptFavoriteRawReferenceURLs) Scan(value any) error {
	return scanPromptFavoriteJSON(value, &((*service.PromptFavorite)(p).ReferenceImageURLs))
}

func (p *promptFavoriteRawLocalizations) Scan(value any) error {
	return scanPromptFavoriteJSON(value, &((*service.PromptFavorite)(p).Localizations))
}

func scanPromptFavoriteJSON(value any, target any) error {
	if value == nil {
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, target)
}

func decodePromptFavoriteJSON(item *service.PromptFavorite) error {
	if item.ReferenceImageURLs == nil {
		item.ReferenceImageURLs = []string{}
	}
	if item.Localizations == nil {
		item.Localizations = map[string]service.PromptLocalization{}
	}
	return nil
}
