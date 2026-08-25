package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	pixelCafeWorkstationDefaultCount   = 10
	pixelCafeWorkstationMinCount       = 1
	pixelCafeWorkstationMaxCount       = 50
	pixelCafeWorkstationLayoutMaxBytes = 4 * 1024
	pixelCafeWorkstationMinX           = 48.0
	pixelCafeWorkstationMaxX           = 912.0
	pixelCafeWorkstationMinY           = 72.0
	pixelCafeWorkstationMaxY           = 520.0
)

// PixelCafeWorkstationPosition is intentionally a public, presentation-only
// projection. It contains no room, user, account, or credential identifiers.
type PixelCafeWorkstationPosition struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type PixelCafeWorkstationLayout []PixelCafeWorkstationPosition

func DefaultPixelCafeWorkstationLayout() PixelCafeWorkstationLayout {
	return PixelCafeWorkstationLayout{
		{ID: 1, X: 340, Y: 250},
		{ID: 2, X: 445, Y: 250},
		{ID: 3, X: 550, Y: 250},
		{ID: 4, X: 655, Y: 250},
		{ID: 5, X: 760, Y: 250},
		{ID: 6, X: 360, Y: 362},
		{ID: 7, X: 465, Y: 362},
		{ID: 8, X: 570, Y: 362},
		{ID: 9, X: 675, Y: 362},
		{ID: 10, X: 780, Y: 362},
	}
}

func NormalizePixelCafeWorkstationLayout(layout PixelCafeWorkstationLayout) (PixelCafeWorkstationLayout, error) {
	count := len(layout)
	if count < pixelCafeWorkstationMinCount || count > pixelCafeWorkstationMaxCount {
		return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_COUNT_INVALID", "pixel cafe layout must contain between 1 and 50 workstations")
	}

	normalized := make(PixelCafeWorkstationLayout, 0, count)
	seen := make(map[int]struct{}, count)
	for _, item := range layout {
		if item.ID < 1 || item.ID > count {
			return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_ID_INVALID", "pixel cafe workstation ids must cover 1 through the layout count")
		}
		if _, exists := seen[item.ID]; exists {
			return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_ID_DUPLICATE", "pixel cafe workstation ids must be unique")
		}
		if math.IsNaN(item.X) || math.IsInf(item.X, 0) || math.IsNaN(item.Y) || math.IsInf(item.Y, 0) ||
			item.X < pixelCafeWorkstationMinX || item.X > pixelCafeWorkstationMaxX ||
			item.Y < pixelCafeWorkstationMinY || item.Y > pixelCafeWorkstationMaxY {
			return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_COORDINATE_INVALID", "pixel cafe workstation coordinates are outside the visible lobby")
		}
		seen[item.ID] = struct{}{}
		normalized = append(normalized, PixelCafeWorkstationPosition{
			ID: item.ID,
			X:  math.Round(item.X*10) / 10,
			Y:  math.Round(item.Y*10) / 10,
		})
	}

	sort.Slice(normalized, func(i, j int) bool { return normalized[i].ID < normalized[j].ID })
	for index, item := range normalized {
		if item.ID != index+1 {
			return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_ID_INVALID", "pixel cafe workstation ids must be contiguous from 1")
		}
	}
	return normalized, nil
}

func ParsePixelCafeWorkstationLayout(raw string) (PixelCafeWorkstationLayout, error) {
	if len(raw) > pixelCafeWorkstationLayoutMaxBytes {
		return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_TOO_LARGE", "pixel cafe layout is too large")
	}
	var layout PixelCafeWorkstationLayout
	if err := json.Unmarshal([]byte(raw), &layout); err != nil {
		return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_INVALID_JSON", "pixel cafe layout JSON is invalid")
	}
	return NormalizePixelCafeWorkstationLayout(layout)
}

func (s *SettingService) GetPixelCafeWorkstationLayout(ctx context.Context) (PixelCafeWorkstationLayout, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultPixelCafeWorkstationLayout(), nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyPixelCafeWorkstationLayout)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultPixelCafeWorkstationLayout(), nil
		}
		return nil, fmt.Errorf("get pixel cafe workstation layout: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return DefaultPixelCafeWorkstationLayout(), nil
	}
	layout, err := ParsePixelCafeWorkstationLayout(raw)
	if err != nil {
		// A malformed historical/manual setting must not make the public Cafe
		// unusable. Administrators can overwrite the safe default via the editor.
		return DefaultPixelCafeWorkstationLayout(), nil
	}
	return layout, nil
}

func (s *SettingService) SetPixelCafeWorkstationLayout(ctx context.Context, layout PixelCafeWorkstationLayout) (PixelCafeWorkstationLayout, error) {
	normalized, err := NormalizePixelCafeWorkstationLayout(layout)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal pixel cafe workstation layout: %w", err)
	}
	if len(data) > pixelCafeWorkstationLayoutMaxBytes {
		return nil, infraerrors.BadRequest("PIXEL_CAFE_LAYOUT_TOO_LARGE", "pixel cafe layout is too large")
	}
	if s == nil || s.settingRepo == nil {
		return nil, fmt.Errorf("setting repository is unavailable")
	}
	if err := s.settingRepo.Set(ctx, SettingKeyPixelCafeWorkstationLayout, string(data)); err != nil {
		return nil, fmt.Errorf("set pixel cafe workstation layout: %w", err)
	}
	s.notifyUpdate()
	return normalized, nil
}
