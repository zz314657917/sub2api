package admin

import (
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type GroupBuyHandler struct {
	groupBuyService *service.GroupBuyService
}

func NewGroupBuyHandler(groupBuyService *service.GroupBuyService) *GroupBuyHandler {
	return &GroupBuyHandler{groupBuyService: groupBuyService}
}

type closeGroupBuyRoundRequest struct {
	Reason string `json:"reason"`
}

func (h *GroupBuyHandler) ListPlans(c *gin.Context) {
	plans, err := h.groupBuyService.ListPlans(c.Request.Context(), true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plans)
}

func (h *GroupBuyHandler) CreatePlan(c *gin.Context) {
	var req service.GroupBuyPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.groupBuyService.AdminCreatePlan(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, plan)
}

func (h *GroupBuyHandler) UpdatePlan(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req service.GroupBuyPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.groupBuyService.AdminUpdatePlan(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plan)
}

func (h *GroupBuyHandler) DeletePlan(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.groupBuyService.AdminDeletePlan(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *GroupBuyHandler) ListRounds(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	rounds, result, err := h.groupBuyService.AdminListRounds(c.Request.Context(), strings.TrimSpace(c.Query("status")), pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, rounds, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *GroupBuyHandler) CreateRound(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	round, err := h.groupBuyService.AdminCreateRound(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, round)
}

func (h *GroupBuyHandler) CloseRound(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req closeGroupBuyRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.groupBuyService.AdminCloseRound(c.Request.Context(), id, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *GroupBuyHandler) RetryActivation(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.groupBuyService.AdminRetryActivation(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *GroupBuyHandler) ProcessRefunds(c *gin.Context) {
	id, err := parseGroupBuyPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	processed, err := h.groupBuyService.AdminProcessRefunds(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"processed": processed})
}

func parseGroupBuyPositiveInt64Param(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(c.Param(name)), 10, 64)
	if err != nil || v <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", name+" is invalid")
	}
	return v, nil
}
