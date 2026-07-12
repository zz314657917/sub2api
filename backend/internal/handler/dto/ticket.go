package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type Ticket struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	Title              string     `json:"title"`
	Status             string     `json:"status"`
	TicketType         string     `json:"ticket_type"`
	SystemKey          string     `json:"system_key,omitempty"`
	LastMessagePreview string     `json:"last_message_preview"`
	LastMessageAt      time.Time  `json:"last_message_at"`
	UserUnreadCount    int        `json:"user_unread_count"`
	AdminUnreadCount   int        `json:"admin_unread_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
}

type AdminTicket struct {
	Ticket
	User *TicketUserSummary `json:"user,omitempty"`
}

type TicketUserSummary struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type TicketMessage struct {
	ID           int64     `json:"id"`
	TicketID     int64     `json:"ticket_id"`
	SenderType   string    `json:"sender_type"`
	SenderUserID *int64    `json:"sender_user_id,omitempty"`
	Content      string    `json:"content"`
	EventType    string    `json:"event_type,omitempty"`
	EventKey     string    `json:"event_key,omitempty"`
	Metadata     any       `json:"metadata,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TicketDetail struct {
	Ticket   Ticket          `json:"ticket"`
	Messages []TicketMessage `json:"messages"`
}

type AdminTicketDetail struct {
	Ticket   AdminTicket     `json:"ticket"`
	Messages []TicketMessage `json:"messages"`
}

type TicketUnreadSummary struct {
	SupportUnread int `json:"support_unread"`
	SystemUnread  int `json:"system_unread"`
	TotalUnread   int `json:"total_unread"`
}

func ToTicket(ticket service.SupportTicket) Ticket {
	return Ticket{
		ID:                 ticket.ID,
		UserID:             ticket.UserID,
		Title:              ticket.Title,
		Status:             ticket.Status,
		TicketType:         ticket.TicketType,
		SystemKey:          ticket.SystemKey,
		LastMessagePreview: ticket.LastMessagePreview,
		LastMessageAt:      ticket.LastMessageAt,
		UserUnreadCount:    ticket.UserUnreadCount,
		AdminUnreadCount:   ticket.AdminUnreadCount,
		CreatedAt:          ticket.CreatedAt,
		UpdatedAt:          ticket.UpdatedAt,
		ClosedAt:           ticket.ClosedAt,
	}
}

func ToTickets(tickets []service.SupportTicket) []Ticket {
	if tickets == nil {
		return []Ticket{}
	}
	out := make([]Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, ToTicket(ticket))
	}
	return out
}

func ToAdminTicket(ticket service.SupportTicket) AdminTicket {
	adminTicket := AdminTicket{Ticket: ToTicket(ticket)}
	if ticket.User != nil {
		adminTicket.User = &TicketUserSummary{
			ID:       ticket.User.ID,
			Username: ticket.User.Username,
			Email:    ticket.User.Email,
		}
	}
	return adminTicket
}

func ToAdminTickets(tickets []service.SupportTicket) []AdminTicket {
	if tickets == nil {
		return []AdminTicket{}
	}
	out := make([]AdminTicket, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, ToAdminTicket(ticket))
	}
	return out
}

func ToTicketMessage(message service.SupportTicketMessage) TicketMessage {
	return TicketMessage{
		ID:           message.ID,
		TicketID:     message.TicketID,
		SenderType:   message.SenderType,
		SenderUserID: message.SenderUserID,
		Content:      message.Content,
		EventType:    message.EventType,
		EventKey:     message.EventKey,
		Metadata:     message.Metadata,
		CreatedAt:    message.CreatedAt,
	}
}

func ToTicketMessages(messages []service.SupportTicketMessage) []TicketMessage {
	if messages == nil {
		return []TicketMessage{}
	}
	out := make([]TicketMessage, 0, len(messages))
	for _, message := range messages {
		out = append(out, ToTicketMessage(message))
	}
	return out
}

func ToTicketDetail(detail *service.TicketDetail) TicketDetail {
	if detail == nil || detail.Ticket == nil {
		return TicketDetail{Messages: []TicketMessage{}}
	}
	return TicketDetail{
		Ticket:   ToTicket(*detail.Ticket),
		Messages: ToTicketMessages(detail.Messages),
	}
}

func ToAdminTicketDetail(detail *service.TicketDetail) AdminTicketDetail {
	if detail == nil || detail.Ticket == nil {
		return AdminTicketDetail{Messages: []TicketMessage{}}
	}
	return AdminTicketDetail{
		Ticket:   ToAdminTicket(*detail.Ticket),
		Messages: ToTicketMessages(detail.Messages),
	}
}

func ToTicketUnreadSummary(summary service.TicketUnreadSummary) TicketUnreadSummary {
	return TicketUnreadSummary{
		SupportUnread: summary.SupportUnread,
		SystemUnread:  summary.SystemUnread,
		TotalUnread:   summary.TotalUnread,
	}
}
