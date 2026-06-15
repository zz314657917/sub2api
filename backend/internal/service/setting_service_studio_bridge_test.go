package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type writableStudioBridgeSettingRepo struct {
	values map[string]string
}

func (r *writableStudioBridgeSettingRepo) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *writableStudioBridgeSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (r *writableStudioBridgeSettingRepo) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	r.values[key] = value
	return nil
}

func (r *writableStudioBridgeSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *writableStudioBridgeSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *writableStudioBridgeSettingRepo) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *writableStudioBridgeSettingRepo) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type studioBridgeGroupRepoStub struct {
	groups []Group
}

func (r *studioBridgeGroupRepoStub) Create(context.Context, *Group) error {
	panic("unexpected Create call")
}

func (r *studioBridgeGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	panic("unexpected GetByID call")
}

func (r *studioBridgeGroupRepoStub) GetByIDLite(context.Context, int64) (*Group, error) {
	panic("unexpected GetByIDLite call")
}

func (r *studioBridgeGroupRepoStub) Update(context.Context, *Group) error {
	panic("unexpected Update call")
}

func (r *studioBridgeGroupRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (r *studioBridgeGroupRepoStub) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}

func (r *studioBridgeGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (r *studioBridgeGroupRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (r *studioBridgeGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	return append([]Group(nil), r.groups...), nil
}

func (r *studioBridgeGroupRepoStub) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}

func (r *studioBridgeGroupRepoStub) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}

func (r *studioBridgeGroupRepoStub) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}

func (r *studioBridgeGroupRepoStub) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}

func (r *studioBridgeGroupRepoStub) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}

func (r *studioBridgeGroupRepoStub) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}

func (r *studioBridgeGroupRepoStub) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

func TestSettingServiceRepairLocalStudioBridgeDefaults(t *testing.T) {
	t.Setenv("STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET", "local-secret")
	raw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:              false,
		SiteName:             "落叶创艺",
		AllowedReturnDomains: []string{"example.com"},
		LaunchReturnURL:      "https://example.com/auth/sub2api/launch",
		RechargeReturnURL:    "https://example.com/purchase",
		DefaultChatGroup:     "4",
		DefaultImageGroup:    "4",
	})
	require.NoError(t, err)

	repo := &writableStudioBridgeSettingRepo{values: map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyStudioBridgeLuoyeAI: raw,
	}}
	svc := NewSettingService(repo, &config.Config{})
	svc.SetStudioBridgeDefaultGroupReader(&studioBridgeGroupRepoStub{groups: []Group{
		{ID: 8, Status: StatusActive, RoutingScope: GroupRoutingScopeInference, SortOrder: 20},
		{ID: 7, Status: StatusActive, RoutingScope: GroupRoutingScopeImage, AllowImageGeneration: true, SortOrder: 10},
	}})

	require.NoError(t, svc.InitializeDefaultSettings(context.Background()))
	updated := parseStudioBridgeAppSettings(repo.values[SettingKeyStudioBridgeLuoyeAI])
	require.True(t, updated.Enabled)
	require.Equal(t, "local-secret", updated.InternalSecret)
	require.Equal(t, defaultStudioBridgeLaunchReturnURL, updated.LaunchReturnURL)
	require.Equal(t, defaultStudioBridgeRechargeURL, updated.RechargeReturnURL)
	require.Equal(t, []string{"127.0.0.1", "localhost"}, updated.AllowedReturnDomains)
	require.Equal(t, "7", updated.DefaultImageGroup)
	require.Equal(t, "8", updated.DefaultChatGroup)
	require.Equal(t, []StudioBridgeDefaultAPIRoute{
		{GroupID: "8", Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, TextOnly: true},
		{GroupID: "7", Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ImageOnly: true},
	}, updated.DefaultAPIRoutes)
}

func TestSettingServiceRepairLocalStudioBridgeDefaultsDoesNotOverwriteProduction(t *testing.T) {
	t.Setenv("STUDIO_BRIDGE_LUOYE_AI_INTERNAL_SECRET", "local-secret")
	raw, err := marshalStudioBridgeAppSettings(StudioBridgeAppSettings{
		Enabled:              true,
		SiteName:             "落叶创艺",
		AllowedReturnDomains: []string{"luoye.example.org"},
		LaunchReturnURL:      "https://luoye.example.org/auth/sub2api/launch",
		RechargeReturnURL:    "https://luoye.example.org/purchase",
		DefaultChatGroup:     "11",
		DefaultImageGroup:    "12",
		InternalSecret:       "prod-secret",
	})
	require.NoError(t, err)
	repo := &writableStudioBridgeSettingRepo{values: map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyStudioBridgeLuoyeAI: raw,
	}}
	svc := NewSettingService(repo, &config.Config{})
	require.NoError(t, svc.InitializeDefaultSettings(context.Background()))
	require.Equal(t, raw, repo.values[SettingKeyStudioBridgeLuoyeAI])
}

func TestSettingServiceParsesLegacyStudioBridgeGroupsAsDefaultAPIRoutes(t *testing.T) {
	raw := `{"enabled":true,"default_chat_group":"11","default_image_group":"12","default_video_group":"13"}`

	parsed := parseStudioBridgeAppSettings(raw)

	require.Equal(t, []StudioBridgeDefaultAPIRoute{
		{GroupID: "11", Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, TextOnly: true},
		{GroupID: "12", Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ImageOnly: true},
		{GroupID: "13", Priority: 1, Weight: 1, CooldownSeconds: 30, Enabled: true, ModelPatterns: []string{"doubao-seedance-*", "*-video-*"}},
	}, parsed.DefaultAPIRoutes)
}
