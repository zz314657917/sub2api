package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

// userTrendSpendColumns is appended only to user-scoped raw trend queries.
// Keep all arithmetic numeric until scanning, so storage-rounding inconsistencies
// fail the strict weight check rather than being silently normalized.
func userTrendSpendColumns() string {
	components := []string{"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost"}
	valid := []string{"total_cost > 0", "total_cost::text NOT IN ('NaN', 'Infinity', '-Infinity')"}
	for _, component := range components {
		valid = append(valid, component+" IS NOT NULL", component+" >= 0",
			component+"::text NOT IN ('NaN', 'Infinity', '-Infinity')")
	}
	valid = append(valid, "("+strings.Join(components, " + ")+") <= total_cost")
	condition := strings.Join(valid, " AND ")
	allocated := make([]string, len(components))
	columns := make([]string, 0, 5)
	for i, component := range components {
		allocated[i] = "actual_cost * " + component + " / NULLIF(total_cost, 0)"
		columns = append(columns, fmt.Sprintf(
			"COALESCE(SUM(CASE WHEN billing_status = 'applied' AND %s THEN %s ELSE 0 END), 0) AS actual_%s",
			condition, allocated[i], component))
	}
	// GREATEST guards only the numerical residue of valid divisions; invalid
	// metadata always leaves the entire applied actual_cost in Other.
	columns = append(columns, fmt.Sprintf(
		"COALESCE(SUM(CASE WHEN billing_status = 'applied' THEN CASE WHEN %s THEN GREATEST(actual_cost - (%s), 0) ELSE actual_cost END ELSE 0 END), 0) AS actual_other_cost",
		condition, strings.Join(allocated, " + ")))
	return ",\n" + strings.Join(columns, ",\n")
}

func scanUserTrendRows(rows *sql.Rows) ([]TrendDataPoint, error) {
	results := make([]TrendDataPoint, 0)
	for rows.Next() {
		var row TrendDataPoint
		var breakdown usagestats.ActualCostBreakdown
		if err := rows.Scan(&row.Date, &row.Requests, &row.InputTokens, &row.OutputTokens,
			&row.CacheCreationTokens, &row.CacheReadTokens, &row.TotalTokens,
			&row.Cost, &row.ActualCost, &breakdown.Input, &breakdown.Output,
			&breakdown.CacheCreation, &breakdown.CacheRead, &breakdown.Other); err != nil {
			return nil, err
		}
		row.ActualCostBreakdown = &breakdown
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
