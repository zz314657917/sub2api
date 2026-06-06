package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	ModelMarketCategoryChat  = "chat"
	ModelMarketCategoryImage = "image"
	ModelMarketCategoryVideo = "video"

	modelMarketCatalogVersion = 11
)

type ModelMarketCatalog struct {
	Version   int                `json:"version"`
	UpdatedAt string             `json:"updated_at,omitempty"`
	Groups    []ModelMarketGroup `json:"groups"`
}

type ModelMarketGroup struct {
	ID                string                    `json:"id"`
	Title             string                    `json:"title"`
	Category          string                    `json:"category"`
	Platform          string                    `json:"platform,omitempty"`
	Description       string                    `json:"description,omitempty"`
	HideOfficialPrice bool                      `json:"hide_official_price,omitempty"`
	HideSaving        bool                      `json:"hide_saving,omitempty"`
	PriceMultiplier   float64                   `json:"price_multiplier,omitempty"`
	SupportedGroupIDs []int64                   `json:"supported_group_ids,omitempty"`
	SupportedGroups   []ModelMarketAccountGroup `json:"supported_groups,omitempty"`
	SortOrder         int                       `json:"sort_order"`
	Enabled           bool                      `json:"enabled"`
	Rows              []ModelMarketPriceRow     `json:"rows"`
}

type ModelMarketAccountGroup struct {
	ID                      int64   `json:"id"`
	Name                    string  `json:"name"`
	Platform                string  `json:"platform"`
	RateMultiplier          float64 `json:"rate_multiplier"`
	ImageRateIndependent    bool    `json:"image_rate_independent"`
	ImageRateMultiplier     float64 `json:"image_rate_multiplier"`
	EffectiveRateMultiplier float64 `json:"effective_rate_multiplier"`
}

type ModelMarketPriceRow struct {
	ID            string `json:"id"`
	Model         string `json:"model,omitempty"`
	Spec          string `json:"spec,omitempty"`
	InputPrice    string `json:"input_price,omitempty"`
	OutputPrice   string `json:"output_price,omitempty"`
	OurPrice      string `json:"our_price"`
	OfficialPrice string `json:"official_price,omitempty"`
	Saving        string `json:"saving,omitempty"`
	Note          string `json:"note,omitempty"`
	SortOrder     int    `json:"sort_order"`
	Enabled       bool   `json:"enabled"`
}

func (s *SettingService) GetModelMarketCatalog(ctx context.Context) (*ModelMarketCatalog, error) {
	var catalog *ModelMarketCatalog
	if s == nil || s.settingRepo == nil {
		catalog = DefaultModelMarketCatalog()
		return s.hydrateModelMarketSupportedGroups(ctx, catalog)
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelMarketCatalog)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			catalog = DefaultModelMarketCatalog()
			return s.hydrateModelMarketSupportedGroups(ctx, catalog)
		}
		return nil, fmt.Errorf("get model market catalog: %w", err)
	}
	catalog, err = ParseModelMarketCatalog(raw)
	if err != nil {
		return nil, err
	}
	return s.hydrateModelMarketSupportedGroups(ctx, catalog)
}

func (s *SettingService) SetModelMarketCatalog(ctx context.Context, catalog *ModelMarketCatalog) (*ModelMarketCatalog, error) {
	normalized, err := NormalizeModelMarketCatalog(catalog)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal model market catalog: %w", err)
	}
	if s == nil || s.settingRepo == nil {
		return nil, fmt.Errorf("setting repository is unavailable")
	}
	if err := s.settingRepo.Set(ctx, SettingKeyModelMarketCatalog, string(data)); err != nil {
		return nil, fmt.Errorf("set model market catalog: %w", err)
	}
	s.notifyUpdate()
	return normalized, nil
}

func (s *SettingService) ResetModelMarketCatalog(ctx context.Context) (*ModelMarketCatalog, error) {
	return s.SetModelMarketCatalog(ctx, DefaultModelMarketCatalog())
}

func (s *SettingService) hydrateModelMarketSupportedGroups(ctx context.Context, catalog *ModelMarketCatalog) (*ModelMarketCatalog, error) {
	if catalog == nil {
		return catalog, nil
	}
	if s == nil || s.modelMarketGroupReader == nil {
		return catalog, nil
	}

	for i := range catalog.Groups {
		group := &catalog.Groups[i]
		group.SupportedGroups = nil
		if len(group.SupportedGroupIDs) == 0 {
			continue
		}

		supported := make([]ModelMarketAccountGroup, 0, len(group.SupportedGroupIDs))
		for _, groupID := range group.SupportedGroupIDs {
			accountGroup, err := s.modelMarketGroupReader.GetByIDLite(ctx, groupID)
			if err != nil {
				if errors.Is(err, ErrGroupNotFound) {
					continue
				}
				return nil, fmt.Errorf("get model market supported group %d: %w", groupID, err)
			}
			if accountGroup == nil || !accountGroup.IsActive() {
				continue
			}
			supported = append(supported, modelMarketAccountGroupFromGroup(accountGroup, group.Category))
		}
		group.SupportedGroups = supported
	}
	return catalog, nil
}

func modelMarketAccountGroupFromGroup(group *Group, category string) ModelMarketAccountGroup {
	effectiveRate := group.RateMultiplier
	if category == ModelMarketCategoryImage && group.ImageRateIndependent {
		effectiveRate = group.ImageRateMultiplier
	}
	if effectiveRate < 0 {
		effectiveRate = 0
	}
	return ModelMarketAccountGroup{
		ID:                      group.ID,
		Name:                    strings.TrimSpace(group.Name),
		Platform:                strings.TrimSpace(group.Platform),
		RateMultiplier:          group.RateMultiplier,
		ImageRateIndependent:    group.ImageRateIndependent,
		ImageRateMultiplier:     group.ImageRateMultiplier,
		EffectiveRateMultiplier: effectiveRate,
	}
}

func ParseModelMarketCatalog(raw string) (*ModelMarketCatalog, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultModelMarketCatalog(), nil
	}
	var catalog ModelMarketCatalog
	if err := json.Unmarshal([]byte(raw), &catalog); err != nil {
		return nil, infraerrors.BadRequest("MODEL_MARKET_CATALOG_INVALID_JSON", "model market catalog JSON is invalid")
	}
	return NormalizeModelMarketCatalog(&catalog)
}

func NormalizeModelMarketCatalog(catalog *ModelMarketCatalog) (*ModelMarketCatalog, error) {
	if catalog == nil {
		return nil, infraerrors.BadRequest("MODEL_MARKET_CATALOG_EMPTY", "model market catalog is required")
	}
	out := *catalog
	if out.Version <= 0 {
		out.Version = modelMarketCatalogVersion
	}
	if len(out.Groups) > 80 {
		return nil, infraerrors.BadRequest("MODEL_MARKET_CATALOG_TOO_LARGE", "model market catalog can contain at most 80 groups")
	}
	groups := make([]ModelMarketGroup, 0, len(out.Groups))
	groupIDs := make(map[string]struct{}, len(out.Groups))
	for i, group := range out.Groups {
		normalized, err := normalizeModelMarketGroup(group, i)
		if err != nil {
			return nil, err
		}
		if _, exists := groupIDs[normalized.ID]; exists {
			return nil, infraerrors.BadRequest("MODEL_MARKET_GROUP_DUPLICATE", "model market group id must be unique")
		}
		groupIDs[normalized.ID] = struct{}{}
		groups = append(groups, normalized)
	}
	sortModelMarketGroups(groups)
	out.Groups = groups
	migrateModelMarketCatalog(&out)
	sortModelMarketGroups(out.Groups)
	return &out, nil
}

func normalizeModelMarketGroup(group ModelMarketGroup, index int) (ModelMarketGroup, error) {
	group.ID = normalizeModelMarketID(group.ID)
	group.Title = strings.TrimSpace(group.Title)
	group.Category = normalizeModelMarketCategory(group.Category)
	group.Platform = strings.TrimSpace(strings.ToLower(group.Platform))
	group.Description = strings.TrimSpace(group.Description)
	group.SupportedGroupIDs = normalizeModelMarketSupportedGroupIDs(group.SupportedGroupIDs)
	group.SupportedGroups = nil
	if group.PriceMultiplier < 0 {
		return group, infraerrors.BadRequest("MODEL_MARKET_GROUP_PRICE_MULTIPLIER_INVALID", "model market group price multiplier must be >= 0")
	}
	if group.PriceMultiplier == 0 {
		group.PriceMultiplier = 1
	}
	if group.ID == "" {
		group.ID = normalizeModelMarketID(group.Title)
	}
	if group.ID == "" {
		return group, infraerrors.BadRequest("MODEL_MARKET_GROUP_ID_REQUIRED", "model market group id is required")
	}
	if group.Title == "" {
		return group, infraerrors.BadRequest("MODEL_MARKET_GROUP_TITLE_REQUIRED", "model market group title is required")
	}
	if group.Category == "" {
		return group, infraerrors.BadRequest("MODEL_MARKET_GROUP_CATEGORY_INVALID", "model market group category must be chat, image, or video")
	}
	if group.SortOrder == 0 {
		group.SortOrder = (index + 1) * 100
	}
	if len(group.Rows) > 400 {
		return group, infraerrors.BadRequest("MODEL_MARKET_ROWS_TOO_LARGE", "model market group can contain at most 400 rows")
	}
	rows := make([]ModelMarketPriceRow, 0, len(group.Rows))
	rowIDs := make(map[string]struct{}, len(group.Rows))
	for i, row := range group.Rows {
		normalized, err := normalizeModelMarketRow(row, group.Category, i)
		if err != nil {
			return group, err
		}
		if _, exists := rowIDs[normalized.ID]; exists {
			return group, infraerrors.BadRequest("MODEL_MARKET_ROW_DUPLICATE", "model market row id must be unique within a group")
		}
		rowIDs[normalized.ID] = struct{}{}
		rows = append(rows, normalized)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].SortOrder == rows[j].SortOrder {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].SortOrder < rows[j].SortOrder
	})
	group.Rows = rows
	return group, nil
}

func normalizeModelMarketRow(row ModelMarketPriceRow, category string, index int) (ModelMarketPriceRow, error) {
	row.ID = normalizeModelMarketID(row.ID)
	row.Model = strings.TrimSpace(row.Model)
	row.Spec = strings.TrimSpace(row.Spec)
	row.InputPrice = strings.TrimSpace(row.InputPrice)
	row.OutputPrice = strings.TrimSpace(row.OutputPrice)
	row.OurPrice = normalizeModelMarketOurPriceCurrency(strings.TrimSpace(row.OurPrice))
	row.OfficialPrice = strings.TrimSpace(row.OfficialPrice)
	row.Saving = strings.TrimSpace(row.Saving)
	row.Note = strings.TrimSpace(row.Note)
	if row.ID == "" {
		switch category {
		case ModelMarketCategoryChat:
			row.ID = normalizeModelMarketID(row.Model)
		default:
			row.ID = normalizeModelMarketID(row.Spec)
		}
	}
	if row.ID == "" {
		return row, infraerrors.BadRequest("MODEL_MARKET_ROW_ID_REQUIRED", "model market row id is required")
	}
	if category == ModelMarketCategoryChat && row.Model == "" {
		return row, infraerrors.BadRequest("MODEL_MARKET_ROW_MODEL_REQUIRED", "chat model market row model is required")
	}
	if category != ModelMarketCategoryChat && row.Spec == "" {
		return row, infraerrors.BadRequest("MODEL_MARKET_ROW_SPEC_REQUIRED", "image/video model market row spec is required")
	}
	if row.OurPrice == "" {
		return row, infraerrors.BadRequest("MODEL_MARKET_ROW_PRICE_REQUIRED", "model market row our_price is required")
	}
	if row.SortOrder == 0 {
		row.SortOrder = (index + 1) * 100
	}
	return row, nil
}

func normalizeModelMarketCategory(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case ModelMarketCategoryChat, "text", "llm":
		return ModelMarketCategoryChat
	case ModelMarketCategoryImage:
		return ModelMarketCategoryImage
	case ModelMarketCategoryVideo:
		return ModelMarketCategoryVideo
	default:
		return ""
	}
}

func normalizeModelMarketSupportedGroupIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeModelMarketOurPriceCurrency(price string) string {
	return strings.ReplaceAll(price, "$", "¥")
}

func normalizeModelMarketID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func migrateModelMarketCatalog(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	if catalog.Version < 2 {
		for i := range catalog.Groups {
			group := &catalog.Groups[i]
			if isAPIMartGPTImage2OfficialModel(group.ID) || isAPIMartGPTImage2OfficialModel(group.Title) {
				group.Rows = gptImage2OfficialModelMarketRows()
			}
		}
	}
	if catalog.Version < 3 {
		ensureGPTImage2ModelMarketGroup(catalog)
	}
	if catalog.Version < 5 {
		applyModelMarketHiddenPriceColumnMigration(catalog)
	}
	if catalog.Version < 7 {
		ensureDoubaoSeedanceModelMarketGroups(catalog)
	}
	if catalog.Version < 8 {
		applyOfficialVideoModelMarketPricingMigration(catalog)
	}
	if catalog.Version < 9 {
		refreshModelMarketGroup(catalog, doubaoSeedanceVideoModelMarketGroup())
	}
	if catalog.Version < 11 {
		refreshModelMarketGroup(catalog, gptImage2OfficialModelMarketGroup())
	}
	applyModelMarketPriceMultiplierDefaults(catalog)
	applyModelMarketOurPriceCurrency(catalog)
	if catalog.Version < modelMarketCatalogVersion {
		catalog.Version = modelMarketCatalogVersion
	}
}

func sortModelMarketGroups(groups []ModelMarketGroup) {
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].SortOrder == groups[j].SortOrder {
			return groups[i].Title < groups[j].Title
		}
		return groups[i].SortOrder < groups[j].SortOrder
	})
}

func ensureGPTImage2ModelMarketGroup(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	for i := range catalog.Groups {
		group := &catalog.Groups[i]
		if normalizeModelMarketID(group.ID) == "gpt-image-2" || strings.EqualFold(strings.TrimSpace(group.Title), "gpt-image-2") {
			if len(group.Rows) == 0 {
				group.Rows = gptImage2ModelMarketRows()
			}
			if group.SortOrder == 0 {
				group.SortOrder = 390
			}
			return
		}
	}
	catalog.Groups = append(catalog.Groups, gptImage2ModelMarketGroup())
}

func ensureDoubaoSeedanceModelMarketGroups(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	ensureModelMarketGroup(catalog, doubaoSeedanceImageModelMarketGroup())
	ensureModelMarketGroup(catalog, doubaoSeedanceVideoModelMarketGroup())
}

func applyOfficialVideoModelMarketPricingMigration(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	refreshModelMarketGroup(catalog, klingV3OmniModelMarketGroup())
	refreshModelMarketGroup(catalog, klingV26ModelMarketGroup())
	refreshModelMarketGroup(catalog, wan27ModelMarketGroup())
	refreshModelMarketGroup(catalog, veo31FastModelMarketGroup())
	refreshModelMarketGroup(catalog, doubaoSeedanceVideoModelMarketGroup())
}

func ensureModelMarketGroup(catalog *ModelMarketCatalog, candidate ModelMarketGroup) {
	for i := range catalog.Groups {
		group := &catalog.Groups[i]
		if normalizeModelMarketID(group.ID) == candidate.ID || normalizeModelMarketID(group.Title) == candidate.ID {
			if len(group.Rows) == 0 {
				group.Rows = candidate.Rows
			}
			if group.Category == "" {
				group.Category = candidate.Category
			}
			if group.Platform == "" {
				group.Platform = candidate.Platform
			}
			if group.Description == "" {
				group.Description = candidate.Description
			}
			if group.SortOrder == 0 {
				group.SortOrder = candidate.SortOrder
			}
			if candidate.HideOfficialPrice {
				group.HideOfficialPrice = true
			}
			if candidate.HideSaving {
				group.HideSaving = true
			}
			if group.PriceMultiplier == 0 {
				group.PriceMultiplier = candidate.PriceMultiplier
			}
			return
		}
	}
	catalog.Groups = append(catalog.Groups, candidate)
}

func refreshModelMarketGroup(catalog *ModelMarketCatalog, candidate ModelMarketGroup) {
	for i := range catalog.Groups {
		group := &catalog.Groups[i]
		if normalizeModelMarketID(group.ID) != candidate.ID && normalizeModelMarketID(group.Title) != candidate.ID {
			continue
		}
		group.Category = candidate.Category
		group.Platform = candidate.Platform
		group.Description = candidate.Description
		group.HideOfficialPrice = candidate.HideOfficialPrice
		group.HideSaving = candidate.HideSaving
		if group.PriceMultiplier == 0 {
			group.PriceMultiplier = candidate.PriceMultiplier
		}
		if group.SortOrder == 0 {
			group.SortOrder = candidate.SortOrder
		}
		group.Rows = candidate.Rows
		return
	}
	catalog.Groups = append(catalog.Groups, candidate)
}

func applyModelMarketPriceMultiplierDefaults(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	for i := range catalog.Groups {
		if catalog.Groups[i].PriceMultiplier == 0 {
			catalog.Groups[i].PriceMultiplier = 1
		}
	}
}

func applyModelMarketOurPriceCurrency(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	for i := range catalog.Groups {
		for j := range catalog.Groups[i].Rows {
			catalog.Groups[i].Rows[j].OurPrice = normalizeModelMarketOurPriceCurrency(catalog.Groups[i].Rows[j].OurPrice)
		}
	}
}

func applyModelMarketHiddenPriceColumnMigration(catalog *ModelMarketCatalog) {
	if catalog == nil {
		return
	}
	for i := range catalog.Groups {
		group := &catalog.Groups[i]
		if isAPIMartGPTImage2OfficialModel(group.ID) || isAPIMartGPTImage2OfficialModel(group.Title) {
			group.HideOfficialPrice = true
			group.HideSaving = true
			for j := range group.Rows {
				row := &group.Rows[j]
				if row.OfficialPrice != "" {
					row.OurPrice = row.OfficialPrice
				}
				row.OfficialPrice = ""
				row.Saving = ""
			}
			continue
		}
		if group.Category == ModelMarketCategoryVideo {
			group.HideOfficialPrice = true
			group.HideSaving = true
			for j := range group.Rows {
				group.Rows[j].OfficialPrice = ""
				group.Rows[j].Saving = ""
			}
			continue
		}
		if group.Category != ModelMarketCategoryChat {
			if modelMarketRowsAllOfficialPriceEmpty(group.Rows) {
				group.HideOfficialPrice = true
			}
			if modelMarketRowsAllSavingEmpty(group.Rows) {
				group.HideSaving = true
			}
		}
	}
}

func modelMarketRowsAllOfficialPriceEmpty(rows []ModelMarketPriceRow) bool {
	for _, row := range rows {
		if strings.TrimSpace(row.OfficialPrice) != "" {
			return false
		}
	}
	return true
}

func modelMarketRowsAllSavingEmpty(rows []ModelMarketPriceRow) bool {
	for _, row := range rows {
		if strings.TrimSpace(row.Saving) != "" {
			return false
		}
	}
	return true
}

func gptImage2ModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "gpt-image-2",
		Title:             "gpt-image-2",
		Category:          ModelMarketCategoryImage,
		Platform:          "openai",
		Description:       "按输出分辨率计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         390,
		Enabled:           true,
		Rows:              gptImage2ModelMarketRows(),
	}
}

func gptImage2ModelMarketRows() []ModelMarketPriceRow {
	return []ModelMarketPriceRow{
		{ID: "1k", Spec: "1K", OurPrice: "¥0.04", SortOrder: 100, Enabled: true},
		{ID: "2k", Spec: "2K", OurPrice: "¥0.06", SortOrder: 200, Enabled: true},
		{ID: "4k", Spec: "4K", OurPrice: "¥0.1", SortOrder: 300, Enabled: true},
	}
}

func gptImage2OfficialModelMarketRows() []ModelMarketPriceRow {
	rows := make([]ModelMarketPriceRow, 0, len(apimartGPTImage2OfficialPriceRows)+1)
	rows = append(rows, ModelMarketPriceRow{
		ID:        "default",
		Spec:      "默认",
		OurPrice:  modelMarketOurPrice(apimartGPTImage2OfficialDefaultPrice),
		SortOrder: 100,
		Enabled:   true,
	})
	for i, row := range apimartGPTImage2OfficialPriceRows {
		rows = append(rows, ModelMarketPriceRow{
			ID:        apimartImagePriceKey(row.Size, row.Quality),
			Spec:      row.Size + " · " + modelMarketQualityLabel(row.Quality),
			OurPrice:  modelMarketOurPrice(row.Official),
			SortOrder: (i + 2) * 100,
			Enabled:   true,
		})
	}
	return rows
}

func gptImage2OfficialModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "gpt-image-2-official",
		Title:             "gpt-image-2-official",
		Category:          ModelMarketCategoryImage,
		Platform:          "openai",
		Description:       "按规格和质量档计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         400,
		Enabled:           true,
		Rows:              gptImage2OfficialModelMarketRows(),
	}
}

func doubaoSeedanceImageModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "doubao-seedance-image",
		Title:             "Doubao Seedance 图像",
		Category:          ModelMarketCategoryImage,
		Platform:          "bytedance",
		Description:       "豆包 Seedance 4.x，按默认规格单次计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         450,
		Enabled:           true,
		Rows:              doubaoSeedanceImageModelMarketRows(),
	}
}

func doubaoSeedanceImageModelMarketRows() []ModelMarketPriceRow {
	return []ModelMarketPriceRow{
		{ID: "doubao-seedance-4-0-default", Spec: "doubao-seedance-4-0 · 默认", OurPrice: modelMarketOurPrice(0.028), SortOrder: 100, Enabled: true},
		{ID: "doubao-seedance-4-5-default", Spec: "doubao-seedance-4-5 · 默认", OurPrice: modelMarketOurPrice(0.035), SortOrder: 200, Enabled: true},
	}
}

func doubaoSeedanceVideoModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "doubao-seedance-video",
		Title:             "Doubao Seedance 视频",
		Category:          ModelMarketCategoryVideo,
		Platform:          "bytedance",
		Description:       "豆包 Seedance 1.x/2.0，按模型变体和分辨率计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         900,
		Enabled:           true,
		Rows:              doubaoSeedanceVideoModelMarketRows(),
	}
}

func doubaoSeedanceVideoModelMarketRows() []ModelMarketPriceRow {
	return []ModelMarketPriceRow{
		{ID: "doubao-seedance-2-0-fast-480p-input", Spec: "doubao-seedance-2.0-fast · 480P-input", OurPrice: modelMarketOurPrice(0.0435), SortOrder: 100, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-480p", Spec: "doubao-seedance-2.0-fast · 480P", OurPrice: modelMarketOurPrice(0.073), SortOrder: 200, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-720p-input", Spec: "doubao-seedance-2.0-fast · 720P-input", OurPrice: modelMarketOurPrice(0.094), SortOrder: 300, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-720p", Spec: "doubao-seedance-2.0-fast · 720P", OurPrice: modelMarketOurPrice(0.157), SortOrder: 400, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-face-480p-input", Spec: "doubao-seedance-2.0-fast-face · 480P-input", OurPrice: modelMarketOurPrice(0.06), SortOrder: 500, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-face-480p", Spec: "doubao-seedance-2.0-fast-face · 480P", OurPrice: modelMarketOurPrice(0.1), SortOrder: 600, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-face-720p-input", Spec: "doubao-seedance-2.0-fast-face · 720P-input", OurPrice: modelMarketOurPrice(0.129), SortOrder: 700, Enabled: true},
		{ID: "doubao-seedance-2-0-fast-face-720p", Spec: "doubao-seedance-2.0-fast-face · 720P", OurPrice: modelMarketOurPrice(0.215), SortOrder: 800, Enabled: true},
		{ID: "doubao-seedance-2-0-480p-input", Spec: "doubao-seedance-2.0 · 480P-input", OurPrice: modelMarketOurPrice(0.055), SortOrder: 900, Enabled: true},
		{ID: "doubao-seedance-2-0-480p", Spec: "doubao-seedance-2.0 · 480P", OurPrice: modelMarketOurPrice(0.0907), SortOrder: 1000, Enabled: true},
		{ID: "doubao-seedance-2-0-720p-input", Spec: "doubao-seedance-2.0 · 720P-input", OurPrice: modelMarketOurPrice(0.118), SortOrder: 1100, Enabled: true},
		{ID: "doubao-seedance-2-0-720p", Spec: "doubao-seedance-2.0 · 720P", OurPrice: modelMarketOurPrice(0.1952), SortOrder: 1200, Enabled: true},
		{ID: "doubao-seedance-2-0-1080p-input", Spec: "doubao-seedance-2.0 · 1080P-input", OurPrice: modelMarketOurPrice(0.267), SortOrder: 1300, Enabled: true},
		{ID: "doubao-seedance-2-0-1080p", Spec: "doubao-seedance-2.0 · 1080P", OurPrice: modelMarketOurPrice(0.44), SortOrder: 1400, Enabled: true},
		{ID: "doubao-seedance-2-0-face-480p-input", Spec: "doubao-seedance-2.0-face · 480P-input", OurPrice: modelMarketOurPrice(0.075), SortOrder: 1500, Enabled: true},
		{ID: "doubao-seedance-2-0-face-480p", Spec: "doubao-seedance-2.0-face · 480P", OurPrice: modelMarketOurPrice(0.124), SortOrder: 1600, Enabled: true},
		{ID: "doubao-seedance-2-0-face-720p-input", Spec: "doubao-seedance-2.0-face · 720P-input", OurPrice: modelMarketOurPrice(0.161), SortOrder: 1700, Enabled: true},
		{ID: "doubao-seedance-2-0-face-720p", Spec: "doubao-seedance-2.0-face · 720P", OurPrice: modelMarketOurPrice(0.267), SortOrder: 1800, Enabled: true},
		{ID: "doubao-seedance-2-0-face-1080p-input", Spec: "doubao-seedance-2.0-face · 1080P-input", OurPrice: modelMarketOurPrice(0.375), SortOrder: 1900, Enabled: true},
		{ID: "doubao-seedance-2-0-face-1080p", Spec: "doubao-seedance-2.0-face · 1080P", OurPrice: modelMarketOurPrice(0.625), SortOrder: 2000, Enabled: true},
		{ID: "doubao-seedance-1-5-pro-480p", Spec: "doubao-seedance-1.5-pro · 480P", OurPrice: modelMarketOurPrice(0.0255), SortOrder: 2100, Enabled: true},
		{ID: "doubao-seedance-1-5-pro-720p", Spec: "doubao-seedance-1.5-pro · 720P", OurPrice: modelMarketOurPrice(0.055), SortOrder: 2200, Enabled: true},
		{ID: "doubao-seedance-1-5-pro-1080p", Spec: "doubao-seedance-1.5-pro · 1080P", OurPrice: modelMarketOurPrice(0.135), SortOrder: 2300, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-fast-480p", Spec: "doubao-seedance-1.0-pro-fast · 480P", OurPrice: modelMarketOurPrice(0.011), SortOrder: 2400, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-fast-720p", Spec: "doubao-seedance-1.0-pro-fast · 720P", OurPrice: modelMarketOurPrice(0.025), SortOrder: 2500, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-fast-1080p", Spec: "doubao-seedance-1.0-pro-fast · 1080P", OurPrice: modelMarketOurPrice(0.052), SortOrder: 2600, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-quality-480p", Spec: "doubao-seedance-1.0-pro-quality · 480P", OurPrice: modelMarketOurPrice(0.0255), SortOrder: 2700, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-quality-720p", Spec: "doubao-seedance-1.0-pro-quality · 720P", OurPrice: modelMarketOurPrice(0.055), SortOrder: 2800, Enabled: true},
		{ID: "doubao-seedance-1-0-pro-quality-1080p", Spec: "doubao-seedance-1.0-pro-quality · 1080P", OurPrice: modelMarketOurPrice(0.13), SortOrder: 2900, Enabled: true},
	}
}

func klingV3OmniModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "kling-v3-omni",
		Title:             "kling-v3-omni",
		Category:          ModelMarketCategoryVideo,
		Platform:          "video",
		Description:       "可灵 3.0 Omni，按能力规格计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         500,
		Enabled:           true,
		Rows: []ModelMarketPriceRow{
			{ID: "default", Spec: "默认", OurPrice: modelMarketOurPrice(0.084) + "/秒", SortOrder: 100, Enabled: true},
			{ID: "pro", Spec: "pro", OurPrice: modelMarketOurPrice(0.112) + "/秒", SortOrder: 200, Enabled: true},
			{ID: "sound", Spec: "sound", OurPrice: modelMarketOurPrice(0.112) + "/秒", SortOrder: 300, Enabled: true},
			{ID: "video", Spec: "video", OurPrice: modelMarketOurPrice(0.126) + "/秒", SortOrder: 400, Enabled: true},
			{ID: "pro-sound", Spec: "pro-sound", OurPrice: modelMarketOurPrice(0.14) + "/秒", SortOrder: 500, Enabled: true},
			{ID: "pro-video", Spec: "pro-video", OurPrice: modelMarketOurPrice(0.168) + "/秒", SortOrder: 600, Enabled: true},
			{ID: "4k", Spec: "4k", OurPrice: modelMarketOurPrice(0.5357) + "/秒", SortOrder: 700, Enabled: true},
			{ID: "4k-sound", Spec: "4k-sound", OurPrice: modelMarketOurPrice(0.5357) + "/秒", SortOrder: 800, Enabled: true},
		},
	}
}

func klingV26ModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "kling-v2-6",
		Title:             "kling-v2-6",
		Category:          ModelMarketCategoryVideo,
		Platform:          "video",
		Description:       "可灵 2.6，按默认和专业档计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         600,
		Enabled:           true,
		Rows: []ModelMarketPriceRow{
			{ID: "default", Spec: "默认", OurPrice: modelMarketOurPrice(0.046) + "/秒", SortOrder: 100, Enabled: true},
			{ID: "pro", Spec: "pro", OurPrice: modelMarketOurPrice(0.078125) + "/秒", SortOrder: 200, Enabled: true},
			{ID: "pro-sound", Spec: "pro-sound", OurPrice: modelMarketOurPrice(0.15625) + "/秒", SortOrder: 300, Enabled: true},
			{ID: "pro-sound-voice", Spec: "pro-sound-voice", OurPrice: modelMarketOurPrice(0.1875) + "/秒", SortOrder: 400, Enabled: true},
		},
	}
}

func wan27ModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "wan2-7",
		Title:             "wan2.7",
		Category:          ModelMarketCategoryVideo,
		Platform:          "video",
		Description:       "Wan 视频模型，按分辨率档计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         700,
		Enabled:           true,
		Rows: []ModelMarketPriceRow{
			{ID: "default", Spec: "默认", OurPrice: modelMarketOurPrice(0.083) + "/秒", SortOrder: 100, Enabled: true},
			{ID: "1080p", Spec: "1080P", OurPrice: modelMarketOurPrice(0.137) + "/秒", SortOrder: 200, Enabled: true},
		},
	}
}

func veo31FastModelMarketGroup() ModelMarketGroup {
	return ModelMarketGroup{
		ID:                "veo3-1-fast",
		Title:             "veo3.1-fast",
		Category:          ModelMarketCategoryVideo,
		Platform:          "video",
		Description:       "Google Veo fast 档，按视频规格计费",
		HideOfficialPrice: true,
		HideSaving:        true,
		PriceMultiplier:   1,
		SortOrder:         800,
		Enabled:           true,
		Rows: []ModelMarketPriceRow{
			{ID: "default", Spec: "默认", OurPrice: modelMarketOurPrice(0.225) + "/秒", SortOrder: 100, Enabled: true},
			{ID: "extend", Spec: "extend", OurPrice: modelMarketOurPrice(0.1) + "/秒", SortOrder: 200, Enabled: true},
			{ID: "4k", Spec: "4K", OurPrice: modelMarketOurPrice(0.3) + "/秒", SortOrder: 300, Enabled: true},
			{ID: "extend-4k", Spec: "EXTEND-4K", OurPrice: modelMarketOurPrice(0.3) + "/秒", SortOrder: 400, Enabled: true},
		},
	}
}

func modelMarketQualityLabel(quality string) string {
	switch normalizeAPIMartImageQuality(quality) {
	case "low":
		return "低"
	case "medium":
		return "中"
	case "high":
		return "高"
	default:
		return "自动"
	}
}

func modelMarketOurPrice(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 8, 64)
	formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	return "¥" + formatted
}

func DefaultModelMarketCatalog() *ModelMarketCatalog {
	return &ModelMarketCatalog{
		Version: modelMarketCatalogVersion,
		Groups: []ModelMarketGroup{
			{
				ID:              "chatgpt",
				Title:           "ChatGPT",
				Category:        ModelMarketCategoryChat,
				Platform:        "openai",
				Description:     "OpenAI 推理与编码模型",
				PriceMultiplier: 1,
				SortOrder:       100,
				Enabled:         true,
				Rows: []ModelMarketPriceRow{
					{ID: "gpt-5-5", Model: "gpt-5.5", InputPrice: "$2.5/M tokens", OutputPrice: "$15/M tokens", OurPrice: "¥5/M 输入 · ¥30/M 输出", SortOrder: 100, Enabled: true},
					{ID: "gpt-5-4", Model: "gpt-5.4", InputPrice: "$2.5/M tokens", OutputPrice: "$15/M tokens", OurPrice: "¥2.5/M 输入 · ¥15/M 输出", SortOrder: 200, Enabled: true},
					{ID: "gpt-5-4-mini", Model: "gpt-5.4-mini", InputPrice: "$0.75/M tokens", OutputPrice: "$4.5/M tokens", OurPrice: "¥0.75/M 输入 · ¥4.5/M 输出", SortOrder: 300, Enabled: true},
				},
			},
			{
				ID:              "gemini",
				Title:           "Gemini",
				Category:        ModelMarketCategoryChat,
				Platform:        "gemini",
				Description:     "Gemini 多模态与长上下文模型",
				PriceMultiplier: 1,
				SortOrder:       200,
				Enabled:         true,
				Rows: []ModelMarketPriceRow{
					{ID: "gemini-3-1-pro", Model: "gemini-3.1-pro", InputPrice: "$2/M tokens", OutputPrice: "$12/M tokens", OurPrice: "¥2/M 输入 · ¥12/M 输出", SortOrder: 100, Enabled: true},
					{ID: "gemini-3-1-flash", Model: "gemini-3.1-flash", InputPrice: "$0.5/M tokens", OutputPrice: "$3/M tokens", OurPrice: "¥0.5/M 输入 · ¥3/M 输出", SortOrder: 200, Enabled: true},
					{ID: "gemini-3-1-flash-lite", Model: "gemini-3.1-flash-lite", InputPrice: "$0.1/M tokens", OutputPrice: "$0.4/M tokens", OurPrice: "¥0.1/M 输入 · ¥0.4/M 输出", SortOrder: 300, Enabled: true},
				},
			},
			{
				ID:              "claude",
				Title:           "Claude",
				Category:        ModelMarketCategoryChat,
				Platform:        "anthropic",
				Description:     "Claude 长上下文与编码模型",
				PriceMultiplier: 1,
				SortOrder:       300,
				Enabled:         true,
				Rows: []ModelMarketPriceRow{
					{ID: "claude-opus-4-8", Model: "claude-opus-4.8", InputPrice: "$5/M tokens", OutputPrice: "$25/M tokens", OurPrice: "¥5/M 输入 · ¥25/M 输出", SortOrder: 100, Enabled: true},
					{ID: "claude-sonnet-4-6", Model: "claude-sonnet-4.6", InputPrice: "$3/M tokens", OutputPrice: "$15/M tokens", OurPrice: "¥3/M 输入 · ¥15/M 输出", SortOrder: 200, Enabled: true},
					{ID: "claude-haiku-4-5", Model: "claude-haiku-4.5", InputPrice: "$1/M tokens", OutputPrice: "$5/M tokens", OurPrice: "¥1/M 输入 · ¥5/M 输出", SortOrder: 300, Enabled: true},
				},
			},
			{
				ID:                "gpt-image-2",
				Title:             "gpt-image-2",
				Category:          ModelMarketCategoryImage,
				Platform:          "openai",
				Description:       "按输出分辨率计费",
				HideOfficialPrice: true,
				HideSaving:        true,
				PriceMultiplier:   1,
				SortOrder:         390,
				Enabled:           true,
				Rows:              gptImage2ModelMarketRows(),
			},
			gptImage2OfficialModelMarketGroup(),
			doubaoSeedanceImageModelMarketGroup(),
			klingV3OmniModelMarketGroup(),
			klingV26ModelMarketGroup(),
			wan27ModelMarketGroup(),
			veo31FastModelMarketGroup(),
			doubaoSeedanceVideoModelMarketGroup(),
		},
	}
}
