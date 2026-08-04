package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type pixelCafePresentationHandlerRepoStub struct {
	values map[string]string
}

func (s *pixelCafePresentationHandlerRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *pixelCafePresentationHandlerRepoStub) GetValue(context.Context, string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *pixelCafePresentationHandlerRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *pixelCafePresentationHandlerRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (s *pixelCafePresentationHandlerRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *pixelCafePresentationHandlerRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *pixelCafePresentationHandlerRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestPixelCafePresentationSettingsPublicResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSettingHandler(service.NewSettingService(&pixelCafePresentationHandlerRepoStub{values: map[string]string{
		service.SettingKeyPixelCafeEnabled:       "true",
		service.SettingKeyPixelCafeTitle:         "模型包间",
		service.SettingKeyPixelCafeDescription:   "按模型选择独立房间。",
		service.SettingKeyPixelCafeHeaderVisible: "false",
	}}, &config.Config{}), "test-version")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)
	h.GetPublicSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Code int `json:"code"`
		Data struct {
			PixelCafeEnabled       bool   `json:"pixel_cafe_enabled"`
			PixelCafeTitle         string `json:"pixel_cafe_title"`
			PixelCafeDescription   string `json:"pixel_cafe_description"`
			PixelCafeHeaderVisible bool   `json:"pixel_cafe_header_visible"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, 0, response.Code)
	require.True(t, response.Data.PixelCafeEnabled)
	require.Equal(t, "模型包间", response.Data.PixelCafeTitle)
	require.Equal(t, "按模型选择独立房间。", response.Data.PixelCafeDescription)
	require.False(t, response.Data.PixelCafeHeaderVisible)
}
