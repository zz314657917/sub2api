package admin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ticketService interface {
	ListAdminTickets(ctx context.Context, filter service.TicketListFilter) ([]service.SupportTicket, int64, error)
	CreateAdminTicketForUser(ctx context.Context, userID int64, adminID int64, input service.CreateTicketInput) (*service.SupportTicket, error)
	GetAdminTicket(ctx context.Context, ticketID int64) (*service.TicketDetail, error)
	AddAdminMessage(ctx context.Context, adminID int64, ticketID int64, input service.AddTicketMessageInput) (*service.TicketDetail, error)
	MarkAdminRead(ctx context.Context, ticketID int64) (*service.SupportTicket, error)
	CloseAdminTicket(ctx context.Context, ticketID int64) (*service.SupportTicket, error)
	ReopenAdminTicket(ctx context.Context, ticketID int64) (*service.SupportTicket, error)
}

type TicketHandler struct {
	svc ticketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func (h *TicketHandler) List(c *gin.Context) {
	filter := parseTicketListFilter(c)
	items, total, err := h.svc.ListAdminTickets(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.ToTickets(items), total, filter.Page, filter.PageSize)
}

func (h *TicketHandler) CreateForUser(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	userID, ok := parseUserIDParam(c)
	if !ok {
		return
	}
	var input service.CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ticket, err := h.svc.CreateAdminTicketForUser(c.Request.Context(), userID, subject.UserID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.ToTicket(*ticket))
}

func (h *TicketHandler) Get(c *gin.Context) {
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	detail, err := h.svc.GetAdminTicket(c.Request.Context(), ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.ToTicketDetail(detail))
}

func (h *TicketHandler) AddMessage(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	var input service.AddTicketMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	detail, err := h.svc.AddAdminMessage(c.Request.Context(), subject.UserID, ticketID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.ToTicketDetail(detail))
}

func (h *TicketHandler) MarkRead(c *gin.Context) {
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	ticket, err := h.svc.MarkAdminRead(c.Request.Context(), ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": dto.ToTicket(*ticket)})
}

func (h *TicketHandler) Close(c *gin.Context) {
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	ticket, err := h.svc.CloseAdminTicket(c.Request.Context(), ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": dto.ToTicket(*ticket)})
}

func (h *TicketHandler) Reopen(c *gin.Context) {
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	ticket, err := h.svc.ReopenAdminTicket(c.Request.Context(), ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": dto.ToTicket(*ticket)})
}

func parseTicketListFilter(c *gin.Context) service.TicketListFilter {
	page := parsePositiveIntQuery(c.Query("page"), 1)
	pageSize := parsePositiveIntQuery(c.Query("page_size"), 20)
	userID := int64(0)
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			userID = parsed
		}
	}
	return service.TicketListFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		TicketType: strings.TrimSpace(c.Query("ticket_type")),
		Search:     strings.TrimSpace(c.Query("search")),
		UserID:     userID,
		EventType:  strings.TrimSpace(c.Query("event_type")),
		EventKey:   strings.TrimSpace(c.Query("event_key")),
		DateFrom:   parseTicketTimeQuery(c.Query("date_from")),
		DateTo:     parseTicketTimeQuery(c.Query("date_to")),
		UnreadOnly: parseTicketBoolQuery(c.Query("unread_only")),
		SortBy:     strings.TrimSpace(c.Query("sort_by")),
		SortOrder:  strings.TrimSpace(c.Query("sort_order")),
		Page:       page,
		PageSize:   pageSize,
	}
}

func parsePositiveIntQuery(raw string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseTicketTimeQuery(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed
	}
	if parsed, err := time.Parse("2006-01-02", raw); err == nil {
		return parsed
	}
	return time.Time{}
}

func parseTicketBoolQuery(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseTicketIDParam(c *gin.Context) (int64, bool) {
	ticketID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || ticketID <= 0 {
		response.BadRequest(c, "invalid ticket id")
		return 0, false
	}
	return ticketID, true
}

func parseUserIDParam(c *gin.Context) (int64, bool) {
	userID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || userID <= 0 {
		response.BadRequest(c, "invalid user id")
		return 0, false
	}
	return userID, true
}
