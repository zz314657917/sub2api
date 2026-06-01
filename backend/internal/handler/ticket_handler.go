package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ticketService interface {
	ListUserTickets(ctx context.Context, userID int64, filter service.TicketListFilter) ([]service.SupportTicket, int64, error)
	GetUserUnreadSummary(ctx context.Context, userID int64) (service.TicketUnreadSummary, error)
	CreateUserTicket(ctx context.Context, userID int64, input service.CreateTicketInput) (*service.SupportTicket, error)
	GetUserTicket(ctx context.Context, userID int64, ticketID int64) (*service.TicketDetail, error)
	AddUserMessage(ctx context.Context, userID int64, ticketID int64, input service.AddTicketMessageInput) (*service.TicketDetail, error)
	MarkUserRead(ctx context.Context, userID int64, ticketID int64) (*service.SupportTicket, error)
	CloseUserTicket(ctx context.Context, userID int64, ticketID int64) (*service.SupportTicket, error)
}

type TicketHandler struct {
	svc ticketService
}

func NewTicketHandler(svc *service.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func (h *TicketHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	filter := parseTicketListFilter(c)
	items, total, err := h.svc.ListUserTickets(c.Request.Context(), subject.UserID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.ToTickets(items), total, filter.Page, filter.PageSize)
}

func (h *TicketHandler) UnreadSummary(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	summary, err := h.svc.GetUserUnreadSummary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.ToTicketUnreadSummary(summary))
}

func (h *TicketHandler) Create(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var input service.CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ticket, err := h.svc.CreateUserTicket(c.Request.Context(), subject.UserID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.ToTicket(*ticket))
}

func (h *TicketHandler) Get(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	detail, err := h.svc.GetUserTicket(c.Request.Context(), subject.UserID, ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.ToTicketDetail(detail))
}

func (h *TicketHandler) AddMessage(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
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
	detail, err := h.svc.AddUserMessage(c.Request.Context(), subject.UserID, ticketID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.ToTicketDetail(detail))
}

func (h *TicketHandler) MarkRead(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	ticket, err := h.svc.MarkUserRead(c.Request.Context(), subject.UserID, ticketID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ticket": dto.ToTicket(*ticket)})
}

func (h *TicketHandler) Close(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ticketID, ok := parseTicketIDParam(c)
	if !ok {
		return
	}
	ticket, err := h.svc.CloseUserTicket(c.Request.Context(), subject.UserID, ticketID)
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
		UnreadOnly: parseTicketBoolQuery(c.Query("unread_only")),
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
