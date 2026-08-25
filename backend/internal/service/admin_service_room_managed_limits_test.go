package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGroupLimits_RoomManagedAlwaysUsesUnlimitedGroupLimits(t *testing.T) {
	daily, weekly, monthly := 25.0, 100.0, 400.0

	gotDaily, gotWeekly, gotMonthly := normalizeGroupLimits(
		GroupAccessModeRoomManaged,
		&daily,
		&weekly,
		&monthly,
	)

	require.Nil(t, gotDaily)
	require.Nil(t, gotWeekly)
	require.Nil(t, gotMonthly)
}

func TestNormalizeGroupLimits_NormalGroupKeepsPositiveLimits(t *testing.T) {
	daily, weekly, monthly := 25.0, 100.0, 400.0

	gotDaily, gotWeekly, gotMonthly := normalizeGroupLimits(
		GroupAccessModeNormal,
		&daily,
		&weekly,
		&monthly,
	)

	require.Equal(t, daily, *gotDaily)
	require.Equal(t, weekly, *gotWeekly)
	require.Equal(t, monthly, *gotMonthly)
}

func TestGroupLimits_RoomManagedIgnoresLegacyStoredValues(t *testing.T) {
	daily, weekly, monthly := 25.0, 100.0, 400.0
	group := &Group{
		AccessMode:      GroupAccessModeRoomManaged,
		DailyLimitUSD:   &daily,
		WeeklyLimitUSD:  &weekly,
		MonthlyLimitUSD: &monthly,
	}

	require.False(t, group.HasDailyLimit())
	require.False(t, group.HasWeeklyLimit())
	require.False(t, group.HasMonthlyLimit())
}
