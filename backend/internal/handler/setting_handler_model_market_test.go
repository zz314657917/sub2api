package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type modelMarketSettingRepoStub struct {
	values map[string]string
}

func (s *modelMarketSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, service.ErrSettingNotFound
}

func (s *modelMarketSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (s *modelMarketSettingRepoStub) Set(context.Context, string, string) error { return nil }

func (s *modelMarketSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (s *modelMarketSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	return nil
}

func (s *modelMarketSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *modelMarketSettingRepoStub) Delete(context.Context, string) error { return nil }

func TestModelMarketCatalogRequiresEnabledFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSettingHandler(service.NewSettingService(&modelMarketSettingRepoStub{}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/model-market/catalog", nil)
	h.GetModelMarketCatalog(c)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestModelMarketCatalogRequiresAuthenticationWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSettingHandler(service.NewSettingService(&modelMarketSettingRepoStub{values: map[string]string{
		service.SettingKeyModelPlazaEnabled:     "true",
		service.SettingKeyModelPlazaRequireAuth: "true",
	}}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/model-market/catalog", nil)
	h.GetModelMarketCatalog(c)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPublicSettingsExposeModelPlazaPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSettingHandler(service.NewSettingService(&modelMarketSettingRepoStub{values: map[string]string{
		service.SettingKeyModelPlazaEnabled:     "true",
		service.SettingKeyModelPlazaRequireAuth: "true",
	}}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)
	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data struct {
			Enabled     bool `json:"model_plaza_enabled"`
			RequireAuth bool `json:"model_plaza_require_auth"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Data.Enabled)
	require.True(t, response.Data.RequireAuth)
}

func TestModelPlazaDescriptionRejectsMoreThan4000Characters(t *testing.T) {
	svc := service.NewSettingService(&modelMarketSettingRepoStub{}, &config.Config{})
	err := svc.UpdateSettings(context.Background(), &service.SystemSettings{
		ModelPlazaDescription: strings.Repeat("a", 4001),
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "model plaza description must not exceed 4000 characters")
}

func TestFilterModelMarketCatalogHidesUnauthorizedExclusiveGroups(t *testing.T) {
	catalog := &service.ModelMarketCatalog{Groups: []service.ModelMarketGroup{
		{ID: "general", Title: "General", Rows: []service.ModelMarketPriceRow{{ID: "general-row"}}},
		{
			ID:                "scoped",
			Title:             "Scoped",
			SupportedGroupIDs: []int64{10, 20, 30},
			SupportedGroups: []service.ModelMarketAccountGroup{
				{ID: 10, Name: "Public"},
				{ID: 20, Name: "Allowed", Exclusive: true},
				{ID: 30, Name: "Hidden", Exclusive: true},
			},
			Rows: []service.ModelMarketPriceRow{{ID: "scoped-row"}},
		},
		{
			ID:                "private-only",
			Title:             "Private only",
			SupportedGroupIDs: []int64{30},
			SupportedGroups:   []service.ModelMarketAccountGroup{{ID: 30, Name: "Hidden", Exclusive: true}},
			Rows:              []service.ModelMarketPriceRow{{ID: "private-row"}},
		},
	}}

	anonymous := filterModelMarketCatalog(catalog, nil)
	require.Len(t, anonymous.Groups, 2)
	require.Equal(t, []int64{10}, anonymous.Groups[1].SupportedGroupIDs)
	require.Equal(t, []service.ModelMarketAccountGroup{{ID: 10, Name: "Public"}}, anonymous.Groups[1].SupportedGroups)

	authorized := filterModelMarketCatalog(catalog, map[int64]struct{}{20: struct{}{}})
	require.Len(t, authorized.Groups, 2)
	require.Equal(t, []int64{10, 20}, authorized.Groups[1].SupportedGroupIDs)
	require.Equal(t, []int64{10, 20}, []int64{
		authorized.Groups[1].SupportedGroups[0].ID,
		authorized.Groups[1].SupportedGroups[1].ID,
	})
}
