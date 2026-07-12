package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestToAdminTicket_IncludesMinimalUserSummary(t *testing.T) {
	ticket := service.SupportTicket{
		ID:     7,
		UserID: 42,
		User: &service.TicketUserSummary{
			ID:       42,
			Username: "alice",
			Email:    "alice@example.com",
		},
	}

	got := ToAdminTicket(ticket)
	require.NotNil(t, got.User)
	require.Equal(t, int64(42), got.User.ID)
	require.Equal(t, "alice", got.User.Username)
	require.Equal(t, "alice@example.com", got.User.Email)

	body, err := json.Marshal(got)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":7,"user_id":42,"title":"","status":"","ticket_type":"","last_message_preview":"","last_message_at":"0001-01-01T00:00:00Z","user_unread_count":0,"admin_unread_count":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","user":{"id":42,"username":"alice","email":"alice@example.com"}}`, string(body))
}

func TestToTicket_DoesNotExposeUserSummary(t *testing.T) {
	ticket := service.SupportTicket{
		ID:     7,
		UserID: 42,
		User: &service.TicketUserSummary{
			ID:       42,
			Username: "alice",
			Email:    "alice@example.com",
		},
	}

	userTicketBody, err := json.Marshal(ToTicket(ticket))
	require.NoError(t, err)
	require.NotContains(t, string(userTicketBody), `"user"`)

	detailBody, err := json.Marshal(ToTicketDetail(&service.TicketDetail{Ticket: &ticket}))
	require.NoError(t, err)
	require.NotContains(t, string(detailBody), `"user"`)
}
