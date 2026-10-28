package repository

import (
	"math"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesRepositoryPaginationExtremePageReturnsEmpty(t *testing.T) {
	items := []int{1, 2, 3}
	for _, size := range []int{math.MinInt, -1, 0, 1, 20, 1000, 1001, math.MaxInt} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			require.NotPanics(t, func() {
				require.Empty(t, paginateSlice(items, pagination.PaginationParams{Page: math.MaxInt, PageSize: size}))
			})
		})
	}
}

func TestUpstreamAFixesRepositoryPaginationOrdinaryPage(t *testing.T) {
	require.Equal(t, []int{3, 4}, paginateSlice(
		[]int{1, 2, 3, 4, 5},
		pagination.PaginationParams{Page: 2, PageSize: 2},
	))
	items := make([]int, 1001)
	for i := range items {
		items[i] = i + 1
	}
	for _, tc := range []struct{ size, first, count int }{{0, 21, 20}, {-1, 21, 20}, {1001, 1001, 1}} {
		got := paginateSlice(items, pagination.PaginationParams{Page: 2, PageSize: tc.size})
		require.Len(t, got, tc.count)
		require.Equal(t, tc.first, got[0])
	}
	require.Equal(t, items[:20], paginateSlice(items, pagination.PaginationParams{Page: math.MinInt, PageSize: 20}))
	require.Empty(t, paginateSlice([]int{}, pagination.PaginationParams{Page: math.MaxInt, PageSize: 20}))
}
