package securityaudit

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesPromptPaginationRejectsImpossiblePage(t *testing.T) {
	calls := 0
	router := promptAdminRouter(&fakePromptAdminService{
		list: func(context.Context, EventFilter, int, int) (*EventPage, error) {
			calls++
			return &EventPage{}, nil
		},
	})
	for _, query := range []string{
		"page=" + strconv.Itoa(math.MaxInt),
		"page=" + strconv.Itoa(math.MaxInt) + "&page_size=100",
		"page=" + strconv.Itoa(math.MaxInt/20+1) + "&page_size=20",
		"page=" + strconv.Itoa(math.MaxInt) + "0",
		"page=bad", "page=0", "page=-1",
		"page_size=0", "page_size=-1", "page_size=1001",
	} {
		t.Run(query, func(t *testing.T) {
			before := calls
			recorder := promptAdminRequest(t, router, http.MethodGet, "/admin/prompt-audit/events?"+query, nil)
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.Contains(t, recorder.Body.String(), "prompt_audit_invalid_pagination")
			require.Equal(t, before, calls)
		})
	}
}

func TestUpstreamAFixesPromptPaginationKeepsOrdinaryPage(t *testing.T) {
	var gotPage, gotPageSize int
	router := promptAdminRouter(&fakePromptAdminService{
		list: func(_ context.Context, _ EventFilter, page, pageSize int) (*EventPage, error) {
			gotPage, gotPageSize = page, pageSize
			return &EventPage{Items: []*Event{}, Page: page, PageSize: pageSize}, nil
		},
	})
	for _, tc := range []struct {
		query      string
		page, size int
	}{
		{"", 1, 20},
		{"page=3&page_size=20", 3, 20},
		{"page=" + strconv.Itoa(math.MaxInt/20) + "&page_size=20", math.MaxInt / 20, 20},
		{"page=" + strconv.Itoa(math.MaxInt) + "&page_size=1", math.MaxInt, 1},
	} {
		t.Run(tc.query, func(t *testing.T) {
			recorder := promptAdminRequest(t, router, http.MethodGet, "/admin/prompt-audit/events?"+tc.query, nil)
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, tc.page, gotPage)
			require.Equal(t, tc.size, gotPageSize)
		})
	}
}
