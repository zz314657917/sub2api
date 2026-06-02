package repository

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildTicketWhere_SystemAuditFilters(t *testing.T) {
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)

	where, args := buildTicketWhere(service.TicketListFilter{
		UserID:     12,
		TicketType: service.TicketTypeSystem,
		EventType:  service.SystemTicketEventGroupChanged,
		EventKey:   "group_changed:12",
		DateFrom:   from,
		DateTo:     to,
	})

	require.Contains(t, where, "user_id = $1")
	require.Contains(t, where, "ticket_type = $2")
	require.Contains(t, where, "stm_event_type.event_type = $3")
	require.Contains(t, where, "stm_event_key.event_key ILIKE $4")
	require.Contains(t, where, "last_message_at >= $5")
	require.Contains(t, where, "last_message_at < $6")
	require.Equal(t, []any{
		int64(12),
		service.TicketTypeSystem,
		service.SystemTicketEventGroupChanged,
		"%group_changed:12%",
		from,
		to.AddDate(0, 0, 1),
	}, args)
}
