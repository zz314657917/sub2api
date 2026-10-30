package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

var spendTrendColumns = []string{"date", "requests", "input_tokens", "output_tokens", "cache_creation_tokens", "cache_read_tokens", "total_tokens", "cost", "actual_cost", "actual_input_cost", "actual_output_cost", "actual_cache_creation_cost", "actual_cache_read_cost", "actual_other_cost"}

// Assert against the actual SQL submitted by both repository entrypoints rather
// than a duplicate Go allocation algorithm. SQL arithmetic execution is covered
// by the independent disposable-database acceptance harness.
func TestSpendBreakdownUserTrendQueries(t *testing.T) {
	for _, filtered := range []bool{false, true} {
		t.Run(map[bool]string{false: "user", true: "filtered"}[filtered], func(t *testing.T) {
			matcher := sqlmock.QueryMatcherFunc(func(_, actual string) error {
				for _, text := range []string{
					"COUNT(*) as requests", "SUM(total_cost)", "billing_status = 'applied'",
					"actual_cost * input_cost / NULLIF(total_cost, 0)",
					"actual_cost * output_cost / NULLIF(total_cost, 0)",
					"actual_cost * cache_creation_cost / NULLIF(total_cost, 0)",
					"actual_cost * cache_read_cost / NULLIF(total_cost, 0)",
					"(input_cost + output_cost + cache_creation_cost + cache_read_cost) <= total_cost",
					"total_cost > 0", "ELSE actual_cost END ELSE 0 END",
					"AS actual_other_cost", "GROUP BY date", "ORDER BY date ASC",
				} {
					require.Contains(t, actual, text)
				}
				for _, field := range []string{"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost"} {
					require.Contains(t, actual, field+" IS NOT NULL")
					require.Contains(t, actual, field+" >= 0")
					require.Contains(t, actual, field+"::text NOT IN ('NaN', 'Infinity', '-Infinity')")
				}
				require.NotContains(t, actual, "SUM(actual_cost) *")
				require.NotContains(t, actual, "image_input_cost")
				require.NotContains(t, actual, "image_output_cost")
				require.NotContains(t, actual, "rate_multiplier")
				if filtered {
					require.Contains(t, actual, "AND user_id = $3 AND api_key_id = $4")
				} else {
					require.Contains(t, actual, "WHERE user_id = $1 AND created_at >= $2 AND created_at < $3")
				}
				return nil
			})
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			repo := &usageLogRepository{sql: db}
			start, end := time.Now().UTC(), time.Now().UTC().Add(time.Hour)
			expect := mock.ExpectQuery("user trend")
			if filtered {
				expect.WithArgs(start, end, int64(7), int64(9))
			} else {
				expect.WithArgs(int64(7), start, end)
			}
			// Two mixed-price requests plus unclassified spend. The returned
			// category ratios deliberately differ from the reported token ratios.
			expect.WillReturnRows(sqlmock.NewRows(spendTrendColumns).
				AddRow("2026-10-30", 4, 9000, 100, 20, 300, 9420, 15, 10, 1, 6, 0.5, 1.5, 1).
				AddRow("2026-10-31", 2, 1, 2, 0, 0, 3, 8, 0, 0, 0, 0, 0, 0))
			var points []TrendDataPoint
			if filtered {
				points, err = repo.GetUsageTrendWithUsageFilters(context.Background(), start, end, "day", UsageLogFilters{UserID: 7, APIKeyID: 9})
			} else {
				points, err = repo.GetUserUsageTrendByUserID(context.Background(), 7, start, end, "day")
			}
			require.NoError(t, err)
			require.Len(t, points, 2)
			require.Equal(t, int64(4), points[0].Requests)
			require.Equal(t, float64(10), points[0].ActualCost)
			require.Equal(t, &usagestats.ActualCostBreakdown{Input: 1, Output: 6, CacheCreation: .5, CacheRead: 1.5, Other: 1}, points[0].ActualCostBreakdown)
			require.Equal(t, &usagestats.ActualCostBreakdown{}, points[1].ActualCostBreakdown)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpendBreakdownAdminTrendKeepsNineColumns(t *testing.T) {
	for _, granularity := range []string{"day", "hour", "month"} {
		t.Run(granularity, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := &usageLogRepository{sql: db}
			// Day/hour take the preaggregated path; month is the raw admin path.
			query := "FROM usage_logs"
			if granularity == "day" {
				query = "FROM usage_dashboard_daily"
			} else if granularity == "hour" {
				query = "FROM usage_dashboard_hourly"
			}
			mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows(spendTrendColumns[:9]).
				AddRow("2026-10-30", 3, 1, 2, 3, 4, 10, 4, 2))
			points, err := repo.GetUsageTrendWithUsageFilters(context.Background(), time.Now(), time.Now(), granularity, UsageLogFilters{})
			require.NoError(t, err)
			require.Len(t, points, 1)
			require.Nil(t, points[0].ActualCostBreakdown)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpendBreakdownTrendSerialization(t *testing.T) {
	for _, breakdown := range []*usagestats.ActualCostBreakdown{nil, {}, {Input: 1, Output: 2, CacheCreation: 3, CacheRead: 4, Other: 5}} {
		data, err := json.Marshal(TrendDataPoint{ActualCost: 15, ActualCostBreakdown: breakdown})
		require.NoError(t, err)
		if breakdown == nil {
			require.NotContains(t, string(data), "actual_cost_breakdown")
		} else {
			require.Contains(t, string(data), `"actual_cost_breakdown":{`)
			var roundtrip TrendDataPoint
			require.NoError(t, json.Unmarshal(data, &roundtrip))
			require.Equal(t, breakdown, roundtrip.ActualCostBreakdown)
			for _, name := range []string{"input", "output", "cache_creation", "cache_read", "other"} {
				require.True(t, strings.Contains(string(data), `"`+name+`":`))
			}
		}
	}
}

func TestSpendBreakdownTrendScanErrors(t *testing.T) {
	for _, rowError := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		repo := &usageLogRepository{sql: db}
		rows := sqlmock.NewRows(spendTrendColumns).AddRow("today", 1, 1, 1, 1, 1, 4, 1, 1, 0, 0, 0, 0, "bad")
		if rowError {
			rows.RowError(0, errors.New("read failed"))
		}
		mock.ExpectQuery("FROM usage_logs").WillReturnRows(rows)
		points, err := repo.GetUserUsageTrendByUserID(context.Background(), 7, time.Now(), time.Now(), "day")
		require.Error(t, err)
		require.Nil(t, points)
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		require.NoError(t, db.Close())
	}
}
