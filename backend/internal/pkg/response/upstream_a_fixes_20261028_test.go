package response

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesParsePaginationBoundsAndFallbacks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nativeMax := strconv.FormatInt(int64(math.MaxInt), 10)
	for _, tc := range []struct {
		name         string
		query        string
		wantPage     int
		wantPageSize int
	}{
		{name: "native max clamps", query: "page=" + nativeMax + "&page_size=20", wantPage: math.MaxInt / 20, wantPageSize: 20},
		{name: "out of range falls back", query: "page=" + nativeMax + "0", wantPage: 1, wantPageSize: 20},
		{name: "invalid page falls back", query: "page=12x", wantPage: 1, wantPageSize: 20},
		{name: "zero page size falls back", query: "page_size=0", wantPage: 1, wantPageSize: 20},
		{name: "negative page size falls back", query: "page_size=-1", wantPage: 1, wantPageSize: 20},
		{name: "oversized page size falls back", query: "page_size=1001", wantPage: 1, wantPageSize: 20},
		{name: "ordinary page", query: "page=3&page_size=50", wantPage: 3, wantPageSize: 50},
		{name: "missing values", query: "", wantPage: 1, wantPageSize: 20},
		{name: "empty values", query: "page=&page_size=", wantPage: 1, wantPageSize: 20},
		{name: "zero page", query: "page=0", wantPage: 1, wantPageSize: 20},
		{name: "negative page", query: "page=-1", wantPage: 1, wantPageSize: 20},
		{name: "max page unit limit", query: "page=" + nativeMax + "&page_size=1", wantPage: math.MaxInt, wantPageSize: 1},
		{name: "max page max limit", query: "page=" + nativeMax + "&page_size=1000", wantPage: math.MaxInt / 1000, wantPageSize: 1000},
		{name: "unsafe end boundary", query: "page=" + strconv.Itoa(math.MaxInt/20+1), wantPage: math.MaxInt / 20, wantPageSize: 20},
		{name: "overflow wraps positive", query: "page=18446744073709551618", wantPage: 1, wantPageSize: 20},
		{name: "page size overflow wraps positive", query: "page_size=18446744073709551618", wantPage: 1, wantPageSize: 20},
		{name: "native max size falls back", query: "page_size=" + nativeMax, wantPage: 1, wantPageSize: 20},
		{name: "bad page size blocks limit", query: "page_size=bad&limit=5", wantPage: 1, wantPageSize: 20},
		{name: "limit alias clamps page", query: "page=" + nativeMax + "&limit=1000", wantPage: math.MaxInt / 1000, wantPageSize: 1000},
		{name: "invalid limit", query: "limit=bad", wantPage: 1, wantPageSize: 20},
		{name: "zero limit", query: "limit=0", wantPage: 1, wantPageSize: 20},
		{name: "negative limit", query: "limit=-1", wantPage: 1, wantPageSize: 20},
		{name: "oversized limit", query: "limit=1001", wantPage: 1, wantPageSize: 20},
		{name: "out of range limit", query: "limit=" + nativeMax + "0", wantPage: 1, wantPageSize: 20},
		{name: "padded decimal", query: "page=0003&page_size=0050", wantPage: 3, wantPageSize: 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/?"+tc.query, nil)
			page, pageSize := ParsePagination(c)
			require.Equal(t, tc.wantPage, page)
			require.Equal(t, tc.wantPageSize, pageSize)
			require.LessOrEqual(t, page, math.MaxInt/pageSize)

			Paginated(c, []string{}, 0, page, pageSize)
			require.Equal(t, http.StatusOK, w.Code)
			var envelope struct {
				Code int           `json:"code"`
				Data PaginatedData `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
			require.Zero(t, envelope.Code)
			require.Equal(t, tc.wantPage, envelope.Data.Page)
			require.Equal(t, tc.wantPageSize, envelope.Data.PageSize)
			require.Equal(t, 1, envelope.Data.Pages)
		})
	}
}

func TestUpstreamAFixesParseIntUsesNativeRange(t *testing.T) {
	nativeMax := strconv.FormatInt(int64(math.MaxInt), 10)
	value, err := parseInt(nativeMax)
	require.NoError(t, err)
	require.Equal(t, math.MaxInt, value)

	for _, raw := range []string{
		strconv.FormatUint(uint64(math.MaxInt)+1, 10),
		nativeMax + "0",
		"18446744073709551618",
	} {
		_, err = parseInt(raw)
		require.Error(t, err, "input=%s", raw)
	}

	for _, raw := range []string{"12x", "", "-1", "+1", " 1", "1 ", "1.5", "1_0", "１２"} {
		value, err = parseInt(raw)
		require.NoError(t, err, "input=%q", raw)
		require.Zero(t, value, "input=%q", raw)
	}
}
