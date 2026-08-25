package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type pixelCafePresentationSettingsRepoStub struct {
	values map[string]string
}

func (s *pixelCafePresentationSettingsRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *pixelCafePresentationSettingsRepoStub) GetValue(context.Context, string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *pixelCafePresentationSettingsRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *pixelCafePresentationSettingsRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (s *pixelCafePresentationSettingsRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *pixelCafePresentationSettingsRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *pixelCafePresentationSettingsRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestPixelCafePresentationSettingsPublicProjection(t *testing.T) {
	custom, err := NewSettingService(&pixelCafePresentationSettingsRepoStub{values: map[string]string{
		SettingKeyPixelCafeEnabled:           "true",
		SettingKeyPixelCafeTitle:             " 模型包间 ",
		SettingKeyPixelCafeDescription:       " 按模型选择独立房间。 ",
		SettingKeyPixelCafeHeaderVisible:     "false",
		SettingKeyPixelCafeWorkstationLayout: `[{"id":1,"x":300,"y":200},{"id":2,"x":400,"y":200},{"id":3,"x":500,"y":200},{"id":4,"x":600,"y":200},{"id":5,"x":700,"y":200},{"id":6,"x":320,"y":340},{"id":7,"x":420,"y":340},{"id":8,"x":520,"y":340},{"id":9,"x":620,"y":340},{"id":10,"x":720,"y":340}]`,
	}}, &config.Config{}).GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, custom.PixelCafeEnabled)
	require.Equal(t, "模型包间", custom.PixelCafeTitle)
	require.Equal(t, "按模型选择独立房间。", custom.PixelCafeDescription)
	require.False(t, custom.PixelCafeHeaderVisible)
	require.Len(t, custom.PixelCafeWorkstationLayout, 10)
	require.Equal(t, float64(300), custom.PixelCafeWorkstationLayout[0].X)

	defaults, err := NewSettingService(&pixelCafePresentationSettingsRepoStub{values: map[string]string{}}, &config.Config{}).
		GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "像素网吧", defaults.PixelCafeTitle)
	require.Equal(t, "把每个模型分组变成一间可订阅的数字包间。", defaults.PixelCafeDescription)
	require.True(t, defaults.PixelCafeHeaderVisible)
	require.Equal(t, DefaultPixelCafeWorkstationLayout(), defaults.PixelCafeWorkstationLayout)
}
