package handler

import (
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type GroupBuyHandler struct {
	groupBuyService *service.GroupBuyService
}

func NewGroupBuyHandler(groupBuyService *service.GroupBuyService) *GroupBuyHandler {
	return &GroupBuyHandler{groupBuyService: groupBuyService}
}

type createGroupBuyOrderRequest struct {
	PlanID        int64  `json:"plan_id" binding:"required"`
	ShareCount    int    `json:"share_count"`
	PaymentType   string `json:"payment_type" binding:"required"`
	OpenID        string `json:"openid"`
	ReturnURL     string `json:"return_url"`
	PaymentSource string `json:"payment_source"`
	IsMobile      *bool  `json:"is_mobile,omitempty"`
}

type bindGroupBuyKeyRequest struct {
	APIKeyID int64 `json:"api_key_id" binding:"required"`
}

func (h *GroupBuyHandler) ListPlans(c *gin.Context) {
	plans, err := h.groupBuyService.ListPlans(c.Request.Context(), false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plans)
}

func (h *GroupBuyHandler) Activity(c *gin.Context) {
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := parsePositiveInt(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	events, err := h.groupBuyService.Activity(c.Request.Context(), limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, events)
}

func (h *GroupBuyHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req createGroupBuyOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}
	result, err := h.groupBuyService.CreateOrder(c.Request.Context(), service.GroupBuyCreateOrderInput{
		UserID:          subject.UserID,
		PlanID:          req.PlanID,
		ShareCount:      req.ShareCount,
		PaymentType:     req.PaymentType,
		OpenID:          req.OpenID,
		ClientIP:        c.ClientIP(),
		IsMobile:        mobile,
		IsWeChatBrowser: isWeChatBrowser(c),
		SrcHost:         c.Request.Host,
		SrcURL:          c.Request.Referer(),
		ReturnURL:       req.ReturnURL,
		PaymentSource:   req.PaymentSource,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *GroupBuyHandler) MySeats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	seats, err := h.groupBuyService.ListMySeats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, seats)
}

func (h *GroupBuyHandler) MyOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	orders, result, err := h.groupBuyService.ListMyOrders(c.Request.Context(), subject.UserID, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, orders, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *GroupBuyHandler) BindKey(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	seatID, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req bindGroupBuyKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	seat, err := h.groupBuyService.BindKey(c.Request.Context(), subject.UserID, seatID, req.APIKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, seat)
}

func parsePositiveInt(raw string) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return 0, infraerrors.BadRequest("INVALID_INPUT", "invalid positive integer")
	}
	return v, nil
}

func parseGroupBuyPositiveInt64Param(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(c.Param(name)), 10, 64)
	if err != nil || v <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", name+" is invalid")
	}
	return v, nil
}
