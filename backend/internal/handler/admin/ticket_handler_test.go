package admin

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseTicketListFilter_AuditParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/admin/tickets?ticket_type=system&event_type=group_changed&event_key=group_changed:12&date_from=2026-06-01&date_to=2026-06-02T13:00:00Z&sort_by=unread_first", nil)

	filter := parseTicketListFilter(c)

	require.Equal(t, service.TicketTypeSystem, filter.TicketType)
	require.Equal(t, service.SystemTicketEventGroupChanged, filter.EventType)
	require.Equal(t, "group_changed:12", filter.EventKey)
	require.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), filter.DateFrom)
	require.Equal(t, time.Date(2026, 6, 2, 13, 0, 0, 0, time.UTC), filter.DateTo)
	require.Equal(t, "unread_first", filter.SortBy)
}
