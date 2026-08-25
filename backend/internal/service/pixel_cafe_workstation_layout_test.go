package service

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type pixelCafeWorkstationLayoutRepoStub struct {
	value  string
	exists bool
	writes int
}

func (s *pixelCafeWorkstationLayoutRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *pixelCafeWorkstationLayoutRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if key != SettingKeyPixelCafeWorkstationLayout || !s.exists {
		return "", ErrSettingNotFound
	}
	return s.value, nil
}

func (s *pixelCafeWorkstationLayoutRepoStub) Set(_ context.Context, key, value string) error {
	if key != SettingKeyPixelCafeWorkstationLayout {
		panic("unexpected setting key")
	}
	s.value = value
	s.exists = true
	s.writes++
	return nil
}

func (s *pixelCafeWorkstationLayoutRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *pixelCafeWorkstationLayoutRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *pixelCafeWorkstationLayoutRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *pixelCafeWorkstationLayoutRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestPixelCafeWorkstationLayoutDefaultsAndMalformedFallback(t *testing.T) {
	ctx := context.Background()
	repo := &pixelCafeWorkstationLayoutRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	layout, err := svc.GetPixelCafeWorkstationLayout(ctx)
	require.NoError(t, err)
	require.Equal(t, DefaultPixelCafeWorkstationLayout(), layout)

	repo.exists = true
	repo.value = `[{"id":1,"x":1,"y":1}]`
	layout, err = svc.GetPixelCafeWorkstationLayout(ctx)
	require.NoError(t, err)
	require.Equal(t, DefaultPixelCafeWorkstationLayout(), layout)
}

func TestPixelCafeWorkstationLayoutPersistsNormalizedSharedLayout(t *testing.T) {
	ctx := context.Background()
	repo := &pixelCafeWorkstationLayoutRepoStub{}
	svc := NewSettingService(repo, &config.Config{})
	draft := DefaultPixelCafeWorkstationLayout()
	draft[0].X = 401.26
	draft[0].Y = 278.74
	draft[0], draft[9] = draft[9], draft[0]

	saved, err := svc.SetPixelCafeWorkstationLayout(ctx, draft)
	require.NoError(t, err)
	require.Equal(t, 1, repo.writes)
	require.Equal(t, 1, saved[0].ID)
	require.Equal(t, 401.3, saved[0].X)
	require.Equal(t, 278.7, saved[0].Y)

	loaded, err := svc.GetPixelCafeWorkstationLayout(ctx)
	require.NoError(t, err)
	require.Equal(t, saved, loaded)
}

func TestPixelCafeWorkstationLayoutAcceptsVariableCounts(t *testing.T) {
	ctx := context.Background()
	for _, count := range []int{1, pixelCafeWorkstationDefaultCount, pixelCafeWorkstationMaxCount} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			repo := &pixelCafeWorkstationLayoutRepoStub{}
			svc := NewSettingService(repo, &config.Config{})
			draft := make(PixelCafeWorkstationLayout, 0, count)
			for id := 1; id <= count; id++ {
				draft = append(draft, PixelCafeWorkstationPosition{ID: id, X: 340, Y: 250})
			}

			saved, err := svc.SetPixelCafeWorkstationLayout(ctx, draft)
			require.NoError(t, err)
			require.Len(t, saved, count)
			require.Equal(t, count, saved[count-1].ID)
			require.Equal(t, 1, repo.writes)

			encoded, err := json.Marshal(saved)
			require.NoError(t, err)
			require.LessOrEqual(t, len(encoded), pixelCafeWorkstationLayoutMaxBytes)
		})
	}
}

func TestPixelCafeWorkstationLayoutRejectsInvalidDraftWithoutWrite(t *testing.T) {
	ctx := context.Background()
	for name, mutate := range map[string]func(PixelCafeWorkstationLayout) PixelCafeWorkstationLayout{
		"empty": func(PixelCafeWorkstationLayout) PixelCafeWorkstationLayout { return nil },
		"too many": func(PixelCafeWorkstationLayout) PixelCafeWorkstationLayout {
			layout := make(PixelCafeWorkstationLayout, 0, pixelCafeWorkstationMaxCount+1)
			for id := 1; id <= pixelCafeWorkstationMaxCount+1; id++ {
				layout = append(layout, PixelCafeWorkstationPosition{ID: id, X: 340, Y: 250})
			}
			return layout
		},
		"non contiguous": func(layout PixelCafeWorkstationLayout) PixelCafeWorkstationLayout {
			layout = layout[:9]
			layout[8].ID = 10
			return layout
		},
		"duplicate id": func(layout PixelCafeWorkstationLayout) PixelCafeWorkstationLayout {
			layout[1].ID = layout[0].ID
			return layout
		},
		"out of bounds": func(layout PixelCafeWorkstationLayout) PixelCafeWorkstationLayout {
			layout[0].X = 0
			return layout
		},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &pixelCafeWorkstationLayoutRepoStub{}
			svc := NewSettingService(repo, &config.Config{})
			_, err := svc.SetPixelCafeWorkstationLayout(ctx, mutate(DefaultPixelCafeWorkstationLayout()))
			require.Error(t, err)
			require.Zero(t, repo.writes)
		})
	}
}
