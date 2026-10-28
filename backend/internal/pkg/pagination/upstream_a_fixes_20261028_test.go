package pagination

import (
	"math"
	"testing"
)

func TestUpstreamAFixesPaginationOffsetSaturatesNativeInt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{name: "native max page default limit", page: math.MaxInt, pageSize: 20, want: (math.MaxInt/20 - 1) * 20},
		{name: "native max page max limit", page: math.MaxInt, pageSize: 1000, want: (math.MaxInt/1000 - 1) * 1000},
		{name: "zero page size uses default", page: 3, pageSize: 0, want: 40},
		{name: "negative page size uses default", page: 3, pageSize: -1, want: 40},
		{name: "oversized page size uses max", page: 3, pageSize: 1001, want: 2000},
		{name: "ordinary page", page: 3, pageSize: 7, want: 14},
		{name: "max page unit limit", page: math.MaxInt, pageSize: 1, want: math.MaxInt - 1},
		{name: "safe page boundary", page: math.MaxInt / 20, pageSize: 20, want: (math.MaxInt/20 - 1) * 20},
		{name: "unsafe end boundary", page: math.MaxInt/20 + 1, pageSize: 20, want: (math.MaxInt/20 - 1) * 20},
		{name: "native min page", page: math.MinInt, pageSize: 20, want: 0},
		{name: "zero page", page: 0, pageSize: 20, want: 0},
		{name: "max page zero size", page: math.MaxInt, pageSize: 0, want: (math.MaxInt/20 - 1) * 20},
		{name: "max page negative size", page: math.MaxInt, pageSize: math.MinInt, want: (math.MaxInt/20 - 1) * 20},
		{name: "max page max size", page: math.MaxInt, pageSize: math.MaxInt, want: (math.MaxInt/1000 - 1) * 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := PaginationParams{Page: tc.page, PageSize: tc.pageSize}
			got := p.Offset()
			if got != tc.want {
				t.Fatalf("Offset() = %d, want %d", got, tc.want)
			}
			if got < 0 || got > math.MaxInt-p.Limit() {
				t.Fatalf("unsafe offset/limit pair: offset=%d limit=%d", got, p.Limit())
			}
		})
	}
}

func TestUpstreamAFixesPaginationLimitKeepsDefaults(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ size, want int }{
		{math.MinInt, 20}, {-1, 20}, {0, 20}, {1, 1}, {20, 20},
		{1000, 1000}, {1001, 1000}, {math.MaxInt, 1000},
	} {
		requireSize := PaginationParams{PageSize: tc.size}.Limit()
		if requireSize != tc.want {
			t.Errorf("Limit(%d)=%d, want %d", tc.size, requireSize, tc.want)
		}
	}
}
