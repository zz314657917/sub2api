package service

import (
	"context"
	"time"

	appTimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const groupUsageDateFormat = "2006-01-02"

type GroupUsageRollupRepository interface {
	SyncGroupUsageRollups(ctx context.Context, todayStart time.Time) error
}

func GroupUsageTimezoneName() string              { return appTimezone.Location().String() }
func GroupUsageTodayStart(at time.Time) time.Time { return appTimezone.StartOfDay(at).UTC() }
func GroupUsageYesterdayStart(at time.Time) time.Time {
	return appTimezone.StartOfDay(at).AddDate(0, 0, -1).UTC()
}
func GroupUsageDate(at time.Time) string {
	return at.In(appTimezone.Location()).Format(groupUsageDateFormat)
}
func ParseGroupUsageDate(value string) (time.Time, error) {
	return appTimezone.ParseInLocation(groupUsageDateFormat, value)
}
