//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type modelMarketSettingRepoStub struct {
	values map[string]string
}

func (s *modelMarketSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	if value, ok := s.values[key]; ok {
		return &Setting{Key: key, Value: value}, nil
	}
	return nil, ErrSettingNotFound
}

func (s *modelMarketSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *modelMarketSettingRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *modelMarketSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *modelMarketSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		_ = s.Set(ctx, key, value)
	}
	return nil
}

func (s *modelMarketSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *modelMarketSettingRepoStub) Delete(ctx context.Context, key string) error {
	delete(s.values, key)
	return nil
}

type modelMarketGroupRepoStub struct {
	groups map[int64]*Group
}

func (s *modelMarketGroupRepoStub) GetByID(ctx context.Context, id int64) (*Group, error) {
	return s.GetByIDLite(ctx, id)
}

func (s *modelMarketGroupRepoStub) GetByIDLite(ctx context.Context, id int64) (*Group, error) {
	if group, ok := s.groups[id]; ok {
		return group, nil
	}
	return nil, ErrGroupNotFound
}

func (s *modelMarketGroupRepoStub) Create(ctx context.Context, group *Group) error {
	panic("unexpected Create call")
}

func (s *modelMarketGroupRepoStub) Update(ctx context.Context, group *Group) error {
	panic("unexpected Update call")
}

func (s *modelMarketGroupRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}

func (s *modelMarketGroupRepoStub) DeleteCascade(ctx context.Context, id int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}

func (s *modelMarketGroupRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *modelMarketGroupRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, status, search string, isExclusive *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *modelMarketGroupRepoStub) ListActive(ctx context.Context) ([]Group, error) {
	panic("unexpected ListActive call")
}

func (s *modelMarketGroupRepoStub) ListActiveByPlatform(ctx context.Context, platform string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}

func (s *modelMarketGroupRepoStub) ExistsByName(ctx context.Context, name string) (bool, error) {
	panic("unexpected ExistsByName call")
}

func (s *modelMarketGroupRepoStub) GetAccountCount(ctx context.Context, groupID int64) (total int64, active int64, err error) {
	panic("unexpected GetAccountCount call")
}

func (s *modelMarketGroupRepoStub) DeleteAccountGroupsByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}

func (s *modelMarketGroupRepoStub) GetAccountIDsByGroupIDs(ctx context.Context, groupIDs []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}

func (s *modelMarketGroupRepoStub) BindAccountsToGroup(ctx context.Context, groupID int64, accountIDs []int64) error {
	panic("unexpected BindAccountsToGroup call")
}

func (s *modelMarketGroupRepoStub) UpdateSortOrders(ctx context.Context, updates []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

func TestSettingService_GetModelMarketCatalog_FallsBackToDefault(t *testing.T) {
	svc := NewSettingService(&modelMarketSettingRepoStub{}, &config.Config{})

	catalog, err := svc.GetModelMarketCatalog(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, catalog.Groups)
	require.Equal(t, "ChatGPT", catalog.Groups[0].Title)
	require.Equal(t, 1.0, catalog.Groups[0].PriceMultiplier)
	require.Equal(t, "gpt-image-2", catalog.Groups[3].Title)
	require.Equal(t, 1.0, catalog.Groups[3].PriceMultiplier)
	require.Len(t, catalog.Groups[3].Rows, 3)
	require.True(t, catalog.Groups[3].HideOfficialPrice)
	require.True(t, catalog.Groups[3].HideSaving)
	require.Equal(t, "¥0.04", requireModelMarketRow(t, catalog.Groups[3].Rows, "1k").OurPrice)
	require.Equal(t, "¥0.06", requireModelMarketRow(t, catalog.Groups[3].Rows, "2k").OurPrice)
	require.Equal(t, "¥0.1", requireModelMarketRow(t, catalog.Groups[3].Rows, "4k").OurPrice)
	require.Contains(t, catalog.Groups[4].Title, "gpt-image-2-official")
	require.Len(t, catalog.Groups[4].Rows, len(apimartGPTImage2OfficialPriceRows)+1)
	row := requireModelMarketRow(t, catalog.Groups[4].Rows, "1024x1024:auto")
	require.Equal(t, "1024x1024 · 自动", row.Spec)
	require.True(t, catalog.Groups[4].HideOfficialPrice)
	require.True(t, catalog.Groups[4].HideSaving)
	require.Empty(t, row.OfficialPrice)
	require.Empty(t, row.Saving)
	require.Equal(t, "¥0.0061", row.OurPrice)
	require.True(t, catalog.Groups[5].HideOfficialPrice)
	require.True(t, catalog.Groups[5].HideSaving)
	require.Empty(t, catalog.Groups[5].Rows[0].OfficialPrice)
	require.Empty(t, catalog.Groups[5].Rows[0].Saving)
	klingOmni := requireModelMarketGroup(t, catalog.Groups, "kling-v3-omni")
	require.Len(t, klingOmni.Rows, 8)
	require.Equal(t, "¥0.084/秒", requireModelMarketRow(t, klingOmni.Rows, "default").OurPrice)
	require.Equal(t, "¥0.5357/秒", requireModelMarketRow(t, klingOmni.Rows, "4k-sound").OurPrice)
	kling26 := requireModelMarketGroup(t, catalog.Groups, "kling-v2-6")
	require.Len(t, kling26.Rows, 4)
	require.Equal(t, "¥0.1875/秒", requireModelMarketRow(t, kling26.Rows, "pro-sound-voice").OurPrice)
	wan27 := requireModelMarketGroup(t, catalog.Groups, "wan2-7")
	require.Equal(t, "¥0.083/秒", requireModelMarketRow(t, wan27.Rows, "default").OurPrice)
	require.Equal(t, "¥0.137/秒", requireModelMarketRow(t, wan27.Rows, "1080p").OurPrice)
	veoFast := requireModelMarketGroup(t, catalog.Groups, "veo3-1-fast")
	require.Len(t, veoFast.Rows, 4)
	require.Equal(t, "¥0.225/秒", requireModelMarketRow(t, veoFast.Rows, "default").OurPrice)
	require.Equal(t, "¥0.3/秒", requireModelMarketRow(t, veoFast.Rows, "extend-4k").OurPrice)
	doubaoImage := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-image")
	require.Equal(t, ModelMarketCategoryImage, doubaoImage.Category)
	require.True(t, doubaoImage.HideOfficialPrice)
	require.True(t, doubaoImage.HideSaving)
	require.Len(t, doubaoImage.Rows, 2)
	require.Equal(t, "¥0.028", requireModelMarketRow(t, doubaoImage.Rows, "doubao-seedance-4-0-default").OurPrice)
	require.Equal(t, "¥0.035", requireModelMarketRow(t, doubaoImage.Rows, "doubao-seedance-4-5-default").OurPrice)
	doubaoVideo := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-video")
	require.Equal(t, ModelMarketCategoryVideo, doubaoVideo.Category)
	require.Equal(t, 1.0, doubaoVideo.PriceMultiplier)
	require.True(t, doubaoVideo.HideOfficialPrice)
	require.True(t, doubaoVideo.HideSaving)
	require.Len(t, doubaoVideo.Rows, 29)
	require.Equal(t, "doubao-seedance-2-0-fast-480p-input", doubaoVideo.Rows[0].ID)
	require.Equal(t, "doubao-seedance-1-0-pro-quality-1080p", doubaoVideo.Rows[len(doubaoVideo.Rows)-1].ID)
	require.Equal(t, "¥0.0435", requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-2-0-fast-480p-input").OurPrice)
	require.Equal(t, "¥0.625", requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-2-0-face-1080p").OurPrice)
	require.Empty(t, requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-2-0-face-1080p").OfficialPrice)
	require.Empty(t, requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-2-0-face-1080p").Saving)
}

func TestSettingService_SetModelMarketCatalog_PersistsNormalizedCatalog(t *testing.T) {
	repo := &modelMarketSettingRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	saved, err := svc.SetModelMarketCatalog(context.Background(), &ModelMarketCatalog{
		Groups: []ModelMarketGroup{
			{
				ID:              " ChatGPT ",
				Title:           "ChatGPT",
				Category:        "llm",
				PriceMultiplier: 1.25,
				Rows: []ModelMarketPriceRow{
					{Model: "gpt-test", OurPrice: "¥1/M"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, ModelMarketCategoryChat, saved.Groups[0].Category)
	require.Equal(t, 1.25, saved.Groups[0].PriceMultiplier)
	require.Equal(t, "gpt-test", saved.Groups[0].Rows[0].ID)

	var raw ModelMarketCatalog
	require.NoError(t, json.Unmarshal([]byte(repo.values[SettingKeyModelMarketCatalog]), &raw))
	require.Equal(t, "chatgpt", raw.Groups[0].ID)
	require.Equal(t, 1.25, raw.Groups[0].PriceMultiplier)
}

func TestNormalizeModelMarketCatalog_DefaultsPriceMultiplier(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Groups: []ModelMarketGroup{
			{
				ID:       "chatgpt",
				Title:    "ChatGPT",
				Category: ModelMarketCategoryChat,
				Rows: []ModelMarketPriceRow{
					{ID: "gpt-test", Model: "gpt-test", OurPrice: "¥1/M", Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, 1.0, catalog.Groups[0].PriceMultiplier)
}

func TestNormalizeModelMarketCatalog_RejectsNegativePriceMultiplier(t *testing.T) {
	_, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Groups: []ModelMarketGroup{
			{
				ID:              "chatgpt",
				Title:           "ChatGPT",
				Category:        ModelMarketCategoryChat,
				PriceMultiplier: -1,
				Rows: []ModelMarketPriceRow{
					{ID: "gpt-test", Model: "gpt-test", OurPrice: "¥1/M", Enabled: true},
				},
			},
		},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "price multiplier")
}

func TestSettingService_GetModelMarketCatalog_HydratesSupportedGroups(t *testing.T) {
	repo := &modelMarketSettingRepoStub{}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetModelMarketGroupReader(&modelMarketGroupRepoStub{
		groups: map[int64]*Group{
			10: {
				ID:             10,
				Name:           "基础图像",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.8,
				Status:         StatusActive,
			},
			20: {
				ID:                   20,
				Name:                 "图片低倍率",
				Platform:             PlatformOpenAI,
				RateMultiplier:       1.2,
				ImageRateIndependent: true,
				ImageRateMultiplier:  0.35,
				Status:               StatusActive,
			},
			30: {
				ID:             30,
				Name:           "停用分组",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.1,
				Status:         "inactive",
			},
		},
	})

	_, err := svc.SetModelMarketCatalog(context.Background(), &ModelMarketCatalog{
		Groups: []ModelMarketGroup{
			{
				ID:                "gpt-image-2",
				Title:             "gpt-image-2",
				Category:          ModelMarketCategoryImage,
				SupportedGroupIDs: []int64{20, 10, 20, 0, -1, 30, 404},
				Rows: []ModelMarketPriceRow{
					{ID: "1k", Spec: "1K", OurPrice: "$0.04", SortOrder: 100, Enabled: true},
				},
				Enabled: true,
			},
		},
	})
	require.NoError(t, err)

	catalog, err := svc.GetModelMarketCatalog(context.Background())
	require.NoError(t, err)
	group := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2")
	require.Equal(t, []int64{20, 10, 30, 404}, group.SupportedGroupIDs)
	require.Len(t, group.SupportedGroups, 2)
	require.Equal(t, int64(20), group.SupportedGroups[0].ID)
	require.Equal(t, "图片低倍率", group.SupportedGroups[0].Name)
	require.Equal(t, 0.35, group.SupportedGroups[0].EffectiveRateMultiplier)
	require.Equal(t, int64(10), group.SupportedGroups[1].ID)
	require.Equal(t, 0.8, group.SupportedGroups[1].EffectiveRateMultiplier)

	var raw ModelMarketCatalog
	require.NoError(t, json.Unmarshal([]byte(repo.values[SettingKeyModelMarketCatalog]), &raw))
	rawGroup := requireModelMarketGroup(t, raw.Groups, "gpt-image-2")
	require.Empty(t, rawGroup.SupportedGroups)
	require.Equal(t, []int64{20, 10, 30, 404}, rawGroup.SupportedGroupIDs)
}

func TestNormalizeModelMarketCatalog_RejectsInvalidRows(t *testing.T) {
	_, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Groups: []ModelMarketGroup{
			{
				ID:       "image",
				Title:    "Image",
				Category: ModelMarketCategoryImage,
				Rows: []ModelMarketPriceRow{
					{ID: "missing-price", Spec: "1024x1024"},
				},
			},
		},
	})
	require.Error(t, err)
}

func TestNormalizeModelMarketCatalog_MigratesTruncatedGPTImage2OfficialRows(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 1,
		Groups: []ModelMarketGroup{
			{
				ID:        "gpt-image-2-official",
				Title:     "gpt-image-2-official",
				Category:  ModelMarketCategoryImage,
				Enabled:   true,
				SortOrder: 100,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "$0.16872", OfficialPrice: "$0.2109", Saving: "20% ↓", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	require.Len(t, catalog.Groups[0].Rows, len(apimartGPTImage2OfficialPriceRows)+1)
	row := requireModelMarketRow(t, catalog.Groups[0].Rows, "2576x3216:medium")
	require.Equal(t, "2576x3216 · 中", row.Spec)
	require.Equal(t, "¥0.1408", row.OurPrice)
	require.Empty(t, row.OfficialPrice)
	require.Empty(t, row.Saving)
	require.True(t, catalog.Groups[0].HideOfficialPrice)
	require.True(t, catalog.Groups[0].HideSaving)
	gptImage2 := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2")
	require.Equal(t, "gpt-image-2", gptImage2.Title)
	require.True(t, gptImage2.HideOfficialPrice)
	require.True(t, gptImage2.HideSaving)
	require.Equal(t, "¥0.04", requireModelMarketRow(t, gptImage2.Rows, "1k").OurPrice)
	require.Equal(t, "¥0.06", requireModelMarketRow(t, gptImage2.Rows, "2k").OurPrice)
	require.Equal(t, "¥0.1", requireModelMarketRow(t, gptImage2.Rows, "4k").OurPrice)
}

func TestNormalizeModelMarketCatalog_MigratesVersion2WithGPTImage2Rows(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 2,
		Groups: []ModelMarketGroup{
			{
				ID:        "chatgpt",
				Title:     "ChatGPT",
				Category:  ModelMarketCategoryChat,
				Enabled:   true,
				SortOrder: 100,
				Rows: []ModelMarketPriceRow{
					{ID: "gpt-test", Model: "gpt-test", OurPrice: "¥1/M", SortOrder: 100, Enabled: true},
				},
			},
			{
				ID:        "gpt-image-2-official",
				Title:     "gpt-image-2-official",
				Category:  ModelMarketCategoryImage,
				Enabled:   true,
				SortOrder: 400,
				Rows:      gptImage2OfficialModelMarketRows(),
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	gptImage2 := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2")
	require.True(t, gptImage2.HideOfficialPrice)
	require.True(t, gptImage2.HideSaving)
	require.Equal(t, "¥0.04", requireModelMarketRow(t, gptImage2.Rows, "1k").OurPrice)
	require.Equal(t, "¥0.06", requireModelMarketRow(t, gptImage2.Rows, "2k").OurPrice)
	require.Equal(t, "¥0.1", requireModelMarketRow(t, gptImage2.Rows, "4k").OurPrice)
	official := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2-official")
	require.Len(t, official.Rows, len(apimartGPTImage2OfficialPriceRows)+1)
	require.True(t, official.HideOfficialPrice)
	require.True(t, official.HideSaving)
	defaultRow := requireModelMarketRow(t, official.Rows, "default")
	require.Equal(t, "¥0.2109", defaultRow.OurPrice)
	require.Empty(t, defaultRow.OfficialPrice)
	require.Empty(t, defaultRow.Saving)
}

func TestNormalizeModelMarketCatalog_MigratesVersion3HiddenColumns(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 3,
		Groups: []ModelMarketGroup{
			{
				ID:        "gpt-image-2-official",
				Title:     "gpt-image-2-official",
				Category:  ModelMarketCategoryImage,
				Enabled:   true,
				SortOrder: 400,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "$0.16872", OfficialPrice: "$0.2109", Saving: "20% ↓", SortOrder: 100, Enabled: true},
				},
			},
			{
				ID:        "kling-v3-omni",
				Title:     "kling-v3-omni",
				Category:  ModelMarketCategoryVideo,
				Enabled:   true,
				SortOrder: 500,
				Rows: []ModelMarketPriceRow{
					{ID: "720p", Spec: "720P 无音频", OurPrice: "$0.0672/秒", OfficialPrice: "6 Credit/秒", Saving: "可配置", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	official := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2-official")
	require.True(t, official.HideOfficialPrice)
	require.True(t, official.HideSaving)
	officialRow := requireModelMarketRow(t, official.Rows, "default")
	require.Equal(t, "¥0.2109", officialRow.OurPrice)
	require.Empty(t, officialRow.OfficialPrice)
	require.Empty(t, officialRow.Saving)
	video := requireModelMarketGroup(t, catalog.Groups, "kling-v3-omni")
	require.True(t, video.HideOfficialPrice)
	require.True(t, video.HideSaving)
	require.Len(t, video.Rows, 8)
	videoRow := requireModelMarketRow(t, video.Rows, "default")
	require.Equal(t, "¥0.084/秒", videoRow.OurPrice)
	require.Empty(t, videoRow.OfficialPrice)
	require.Empty(t, videoRow.Saving)
}

func TestNormalizeModelMarketCatalog_MigratesVersion4GPTImage2HiddenColumns(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 4,
		Groups: []ModelMarketGroup{
			{
				ID:        "gpt-image-2",
				Title:     "gpt-image-2",
				Category:  ModelMarketCategoryImage,
				Enabled:   true,
				SortOrder: 390,
				Rows: []ModelMarketPriceRow{
					{ID: "1k", Spec: "1K", OurPrice: "$0.04", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	gptImage2 := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2")
	require.True(t, gptImage2.HideOfficialPrice)
	require.True(t, gptImage2.HideSaving)
}

func TestNormalizeModelMarketCatalog_MigratesVersion6DoubaoSeedanceGroups(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 6,
		Groups: []ModelMarketGroup{
			{
				ID:        "chatgpt",
				Title:     "ChatGPT",
				Category:  ModelMarketCategoryChat,
				Enabled:   true,
				SortOrder: 100,
				Rows: []ModelMarketPriceRow{
					{ID: "gpt-test", Model: "gpt-test", OurPrice: "¥1/M", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	doubaoImage := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-image")
	require.Equal(t, ModelMarketCategoryImage, doubaoImage.Category)
	require.True(t, doubaoImage.HideOfficialPrice)
	require.True(t, doubaoImage.HideSaving)
	require.Equal(t, "¥0.028", requireModelMarketRow(t, doubaoImage.Rows, "doubao-seedance-4-0-default").OurPrice)
	require.Equal(t, "¥0.035", requireModelMarketRow(t, doubaoImage.Rows, "doubao-seedance-4-5-default").OurPrice)
	doubaoVideo := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-video")
	require.Equal(t, ModelMarketCategoryVideo, doubaoVideo.Category)
	require.True(t, doubaoVideo.HideOfficialPrice)
	require.True(t, doubaoVideo.HideSaving)
	require.Len(t, doubaoVideo.Rows, 29)
	require.Equal(t, "doubao-seedance-2-0-fast-480p-input", doubaoVideo.Rows[0].ID)
	require.Equal(t, "doubao-seedance-1-0-pro-quality-1080p", doubaoVideo.Rows[len(doubaoVideo.Rows)-1].ID)
	require.Equal(t, "¥0.011", requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-1-0-pro-fast-480p").OurPrice)
	require.Equal(t, "¥0.625", requireModelMarketRow(t, doubaoVideo.Rows, "doubao-seedance-2-0-face-1080p").OurPrice)
}

func TestNormalizeModelMarketCatalog_MigratesVersion7OfficialVideoPrices(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 7,
		Groups: []ModelMarketGroup{
			{
				ID:        "kling-v3-omni",
				Title:     "kling-v3-omni",
				Category:  ModelMarketCategoryVideo,
				Enabled:   true,
				SortOrder: 500,
				Rows: []ModelMarketPriceRow{
					{ID: "720p", Spec: "720P 无音频", OurPrice: "$0.0672/秒", SortOrder: 100, Enabled: true},
				},
			},
			{
				ID:        "kling-v2-6",
				Title:     "kling-v2-6",
				Category:  ModelMarketCategoryVideo,
				Enabled:   true,
				SortOrder: 600,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "$0.0368/秒", SortOrder: 100, Enabled: true},
				},
			},
			{
				ID:        "wan2-7",
				Title:     "wan2.7",
				Category:  ModelMarketCategoryVideo,
				Enabled:   true,
				SortOrder: 700,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "$0.0664/秒", SortOrder: 100, Enabled: true},
				},
			},
			{
				ID:        "veo3-1-fast",
				Title:     "veo3.1-fast",
				Category:  ModelMarketCategoryVideo,
				Enabled:   true,
				SortOrder: 800,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "$0.18/秒", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	klingOmni := requireModelMarketGroup(t, catalog.Groups, "kling-v3-omni")
	require.Len(t, klingOmni.Rows, 8)
	require.Equal(t, "¥0.084/秒", requireModelMarketRow(t, klingOmni.Rows, "default").OurPrice)
	require.Equal(t, "¥0.5357/秒", requireModelMarketRow(t, klingOmni.Rows, "4k").OurPrice)
	require.True(t, klingOmni.HideOfficialPrice)
	require.True(t, klingOmni.HideSaving)
	kling26 := requireModelMarketGroup(t, catalog.Groups, "kling-v2-6")
	require.Len(t, kling26.Rows, 4)
	require.Equal(t, "¥0.046/秒", requireModelMarketRow(t, kling26.Rows, "default").OurPrice)
	require.Equal(t, "¥0.1875/秒", requireModelMarketRow(t, kling26.Rows, "pro-sound-voice").OurPrice)
	wan27 := requireModelMarketGroup(t, catalog.Groups, "wan2-7")
	require.Equal(t, "¥0.083/秒", requireModelMarketRow(t, wan27.Rows, "default").OurPrice)
	require.Equal(t, "¥0.137/秒", requireModelMarketRow(t, wan27.Rows, "1080p").OurPrice)
	veoFast := requireModelMarketGroup(t, catalog.Groups, "veo3-1-fast")
	require.Len(t, veoFast.Rows, 4)
	require.Equal(t, "¥0.225/秒", requireModelMarketRow(t, veoFast.Rows, "default").OurPrice)
	require.Equal(t, "¥0.3/秒", requireModelMarketRow(t, veoFast.Rows, "extend-4k").OurPrice)
	doubaoVideo := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-video")
	require.Len(t, doubaoVideo.Rows, 29)
}

func TestNormalizeModelMarketCatalog_MigratesVersion8DoubaoSeedanceVideoOrder(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 8,
		Groups: []ModelMarketGroup{
			{
				ID:                "doubao-seedance-video",
				Title:             "Doubao Seedance 视频",
				Category:          ModelMarketCategoryVideo,
				SupportedGroupIDs: []int64{20, 10},
				Enabled:           true,
				SortOrder:         900,
				Rows: []ModelMarketPriceRow{
					{ID: "doubao-seedance-1-0-pro-fast-480p", Spec: "doubao-seedance-1-0-pro-fast - 480P", OurPrice: "$0.011", SortOrder: 100, Enabled: true},
					{ID: "doubao-seedance-2-0-fast-480p-input", Spec: "doubao-seedance-2.0-fast · 480P-input", OurPrice: "$0.0435", SortOrder: 2100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	doubaoVideo := requireModelMarketGroup(t, catalog.Groups, "doubao-seedance-video")
	require.Equal(t, []int64{20, 10}, doubaoVideo.SupportedGroupIDs)
	require.Len(t, doubaoVideo.Rows, 29)
	require.Equal(t, "doubao-seedance-2-0-fast-480p-input", doubaoVideo.Rows[0].ID)
	require.Equal(t, "doubao-seedance-2.0-fast · 480P-input", doubaoVideo.Rows[0].Spec)
	require.Equal(t, "¥0.0435", doubaoVideo.Rows[0].OurPrice)
	require.Equal(t, "doubao-seedance-1-5-pro-480p", doubaoVideo.Rows[20].ID)
	require.Equal(t, "doubao-seedance-1-0-pro-quality-1080p", doubaoVideo.Rows[len(doubaoVideo.Rows)-1].ID)
}

func TestNormalizeModelMarketCatalog_MigratesVersion10GPTImage2OfficialDisplayPrices(t *testing.T) {
	catalog, err := NormalizeModelMarketCatalog(&ModelMarketCatalog{
		Version: 10,
		Groups: []ModelMarketGroup{
			{
				ID:                "gpt-image-2-official",
				Title:             "gpt-image-2-official",
				Category:          ModelMarketCategoryImage,
				SupportedGroupIDs: []int64{20, 10},
				PriceMultiplier:   1.3,
				Enabled:           true,
				SortOrder:         400,
				Rows: []ModelMarketPriceRow{
					{ID: "default", Spec: "默认", OurPrice: "¥1.77156", SortOrder: 100, Enabled: true},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, modelMarketCatalogVersion, catalog.Version)
	official := requireModelMarketGroup(t, catalog.Groups, "gpt-image-2-official")
	require.Equal(t, []int64{20, 10}, official.SupportedGroupIDs)
	require.Equal(t, 1.3, official.PriceMultiplier)
	require.Len(t, official.Rows, len(apimartGPTImage2OfficialPriceRows)+1)
	require.Equal(t, "¥0.2109", requireModelMarketRow(t, official.Rows, "default").OurPrice)
	require.Equal(t, "¥0.1408", requireModelMarketRow(t, official.Rows, "2576x3216:medium").OurPrice)
	require.True(t, official.HideOfficialPrice)
	require.True(t, official.HideSaving)
}

func requireModelMarketGroup(t *testing.T, groups []ModelMarketGroup, id string) ModelMarketGroup {
	t.Helper()
	for _, group := range groups {
		if group.ID == id {
			return group
		}
	}
	require.Failf(t, "model market group not found", "id=%s", id)
	return ModelMarketGroup{}
}

func requireModelMarketRow(t *testing.T, rows []ModelMarketPriceRow, id string) ModelMarketPriceRow {
	t.Helper()
	for _, row := range rows {
		if row.ID == id {
			return row
		}
	}
	require.Failf(t, "model market row not found", "id=%s", id)
	return ModelMarketPriceRow{}
}
