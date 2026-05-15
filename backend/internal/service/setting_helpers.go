package service

import (
	"math"
	"strconv"
	"strings"
	"time"
)

func parseNonNegativeFloatSetting(raw string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}

func parseNonNegativeIntSetting(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 0 {
		return fallback
	}
	return v
}

func normalizeNonNegativeFloat(v float64) float64 {
	if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

func parseOptionalRFC3339Setting(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	utc := parsed.UTC()
	return &utc
}
