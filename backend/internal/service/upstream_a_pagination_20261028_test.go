package service

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesBalancePaginationExtremePageReturnsEmpty(t *testing.T) {
	for _, size := range []int{math.MinInt, -1, 0, 1, 20, 1000, 1001, math.MaxInt} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			require.NotPanics(t, func() {
				got := mergeBalanceHistoryCodes(
					[]RedeemCode{{ID: 1}},
					[]RedeemCode{{ID: -1}},
					pagination.PaginationParams{Page: math.MaxInt, PageSize: size},
				)
				require.Empty(t, got)
			})
		})
	}
}

func TestUpstreamAFixesBalancePaginationOrdinaryPage(t *testing.T) {
	now := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	redeem := []RedeemCode{{ID: 1, CreatedAt: now}, {ID: 2, CreatedAt: now.Add(-2 * time.Hour)}}
	affiliate := []RedeemCode{{ID: -1, CreatedAt: now.Add(-time.Hour)}, {ID: -2, CreatedAt: now.Add(-3 * time.Hour)}}
	got := mergeBalanceHistoryCodes(
		redeem, affiliate,
		pagination.PaginationParams{Page: 2, PageSize: 2},
	)
	require.Equal(t, []RedeemCode{redeem[1], affiliate[1]}, got)
	for _, size := range []int{-1, 0, 1001} {
		got = mergeBalanceHistoryCodes(redeem, affiliate, pagination.PaginationParams{Page: 1, PageSize: size})
		require.Len(t, got, 4)
		require.Empty(t, mergeBalanceHistoryCodes(redeem, affiliate, pagination.PaginationParams{Page: 2, PageSize: size}))
	}
	require.Empty(t, mergeBalanceHistoryCodes(nil, nil, pagination.PaginationParams{Page: math.MaxInt, PageSize: 20}))
}
