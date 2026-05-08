package service

import (
	"math"
	"strconv"
	"strings"
)

func parseNonNegativeFloatSetting(raw string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
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
