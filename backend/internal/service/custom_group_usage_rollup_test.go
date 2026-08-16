package service

import (
	"testing"
	"time"

	appTimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
)

func TestGroupUsageDateUsesConfiguredTimezoneBoundary(t *testing.T) {
	require.NoError(t, appTimezone.Init("America/New_York"))
	t.Cleanup(func() { require.NoError(t, appTimezone.Init("Asia/Shanghai")) })

	start := GroupUsageTodayStart(time.Date(2026, 3, 9, 4, 30, 0, 0, time.UTC))
	require.Equal(t, time.Date(2026, 3, 9, 4, 0, 0, 0, time.UTC), start)
	require.Equal(t, "2026-03-09", GroupUsageDate(start))
}

func TestGroupUsageParseDateUsesConfiguredTimezone(t *testing.T) {
	require.NoError(t, appTimezone.Init("America/New_York"))
	t.Cleanup(func() { require.NoError(t, appTimezone.Init("Asia/Shanghai")) })

	parsed, err := ParseGroupUsageDate("2026-03-09")
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 3, 9, 4, 0, 0, 0, time.UTC), parsed.UTC())
}

func TestGroupUsageYesterdayStartHandlesDST(t *testing.T) {
	require.NoError(t, appTimezone.Init("America/New_York"))
	t.Cleanup(func() { require.NoError(t, appTimezone.Init("Asia/Shanghai")) })

	todayStart := GroupUsageTodayStart(time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC))
	yesterdayStart := GroupUsageYesterdayStart(todayStart)
	require.Equal(t, time.Date(2026, 3, 8, 5, 0, 0, 0, time.UTC), yesterdayStart)
	// The DST day was 23 hours, so natural-day calculation must not subtract 24h.
	require.Equal(t, 23*time.Hour, todayStart.Sub(yesterdayStart))
}
