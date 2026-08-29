package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CafeRoomHandler struct {
	service    *service.CafeRoomService
	activation cafeRoundFulfillmentService
	settings   cafeWorkstationLayoutService
	quotaReset service.CafeQuotaResetService
}

type cafeWorkstationLayoutService interface {
	GetPixelCafeWorkstationLayout(context.Context) (service.PixelCafeWorkstationLayout, error)
	SetPixelCafeWorkstationLayout(context.Context, service.PixelCafeWorkstationLayout) (service.PixelCafeWorkstationLayout, error)
}

type cafeRoundFulfillmentService interface {
	ListPendingRounds(context.Context, service.CafePendingRoundParams) ([]service.CafePendingRound, *pagination.PaginationResult, error)
	ListRoundAccountOptions(context.Context, int64, service.CafeRoundAccountOptionParams) ([]service.CafeRoundAccountOption, *pagination.PaginationResult, error)
	AssignAccountAndActivateRound(context.Context, int64, int64) (*service.CafePendingRound, error)
}

func NewCafeRoomHandler(cafeRoomService *service.CafeRoomService) *CafeRoomHandler {
	return &CafeRoomHandler{service: cafeRoomService}
}

func NewCafeRoomHandlerWithActivation(cafeRoomService *service.CafeRoomService, activation *service.CafeRoomActivationService, settings *service.SettingService) *CafeRoomHandler {
	return &CafeRoomHandler{service: cafeRoomService, activation: activation, settings: settings}
}

// SetQuotaResetService wires the optional admin-only quota reset capability.
// It is kept as a setter so existing lightweight handler constructors/tests do
// not need an AdminService dependency.
func (h *CafeRoomHandler) SetQuotaResetService(reset service.CafeQuotaResetService) {
	h.quotaReset = reset
}

func (h *CafeRoomHandler) GetWorkstationLayout(c *gin.Context) {
	layout, err := h.settings.GetPixelCafeWorkstationLayout(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, layout)
}

func (h *CafeRoomHandler) UpdateWorkstationLayout(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4*1024)
	var layout service.PixelCafeWorkstationLayout
	if err := c.ShouldBindJSON(&layout); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	updated, err := h.settings.SetPixelCafeWorkstationLayout(c.Request.Context(), layout)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *CafeRoomHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	rooms, result, err := h.service.List(c.Request.Context(), pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    strings.TrimSpace(c.Query("sort_by")),
		SortOrder: strings.TrimSpace(c.Query("sort_order")),
	}, c.Query("status"), c.Query("zone"), c.Query("search"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if result == nil {
		response.PaginatedWithResult(c, rooms, nil)
		return
	}
	response.PaginatedWithResult(c, rooms, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *CafeRoomHandler) AccountOptions(c *gin.Context) {
	params, err := parseCafeRoomAccountOptionParams(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items, result, err := h.service.ListAccountOptions(c.Request.Context(), params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{Total: result.Total, Page: result.Page, PageSize: result.PageSize, Pages: result.Pages})
}

func (h *CafeRoomHandler) Create(c *gin.Context) {
	var req service.CafeRoomInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	room, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, room)
}

func (h *CafeRoomHandler) Get(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	room, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, room)
}

func (h *CafeRoomHandler) Update(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req service.CafeRoomUpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	room, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, room)
}

func (h *CafeRoomHandler) Delete(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func (h *CafeRoomHandler) ResetRoomQuotas(c *gin.Context) {
	if h.quotaReset == nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CAFE_QUOTA_RESET_UNAVAILABLE", "cafe quota reset is unavailable"))
		return
	}
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result, err := h.quotaReset.AdminResetCafeRateLimitUsage(c.Request.Context(), &id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *CafeRoomHandler) ResetAllQuotas(c *gin.Context) {
	if h.quotaReset == nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CAFE_QUOTA_RESET_UNAVAILABLE", "cafe quota reset is unavailable"))
		return
	}
	result, err := h.quotaReset.AdminResetCafeRateLimitUsage(c.Request.Context(), nil)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *CafeRoomHandler) BulkCreate(c *gin.Context) {
	var req service.CafeRoomBulkInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	response.Created(c, h.service.BulkCreate(c.Request.Context(), req))
}

func (h *CafeRoomHandler) OpenRound(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	round, err := h.service.OpenRound(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, round)
}

func (h *CafeRoomHandler) PauseRound(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	round, err := h.service.PauseOpenRound(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, round)
}

func (h *CafeRoomHandler) ListPendingRounds(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.activation.ListPendingRounds(c.Request.Context(), service.CafePendingRoundParams{Page: page, PageSize: pageSize, Search: strings.TrimSpace(c.Query("search"))})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{Total: result.Total, Page: result.Page, PageSize: result.PageSize, Pages: result.Pages})
}

func (h *CafeRoomHandler) ListRoundAccountOptions(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.activation.ListRoundAccountOptions(c.Request.Context(), id, service.CafeRoundAccountOptionParams{Page: page, PageSize: pageSize, Search: strings.TrimSpace(c.Query("search"))})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{Total: result.Total, Page: result.Page, PageSize: result.PageSize, Pages: result.Pages})
}

func (h *CafeRoomHandler) AssignRoundAccount(c *gin.Context) {
	id, err := parseCafeRoomPositiveInt64Param(c, "id")
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req struct {
		AccountID int64 `json:"account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AccountID <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_ID", "account_id is invalid"))
		return
	}
	item, err := h.activation.AssignAccountAndActivateRound(c.Request.Context(), id, req.AccountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func parseCafeRoomPositiveInt64Param(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param(name)), 10, 64)
	if err != nil || id <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", name+" is invalid")
	}
	return id, nil
}

func parseCafeRoomAccountOptionParams(c *gin.Context) (service.CafeRoomAccountOptionParams, error) {
	page, pageSize := response.ParsePagination(c)
	if pageSize > 50 {
		pageSize = 50
	}
	params := service.CafeRoomAccountOptionParams{Page: page, PageSize: pageSize, Search: strings.TrimSpace(c.Query("search"))}
	if raw := strings.TrimSpace(c.Query("plan_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return params, infraerrors.BadRequest("INVALID_ID", "plan_id is invalid")
		}
		params.PlanID = value
	}
	if raw := strings.TrimSpace(c.Query("exclude_room_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return params, infraerrors.BadRequest("INVALID_ID", "exclude_room_id is invalid")
		}
		params.ExcludeRoomID = value
	}
	if raw := strings.TrimSpace(c.Query("ids")); raw != "" {
		seen := make(map[int64]struct{})
		for _, value := range strings.Split(raw, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil || id <= 0 {
				return params, infraerrors.BadRequest("INVALID_ID", "ids contains an invalid account id")
			}
			if _, exists := seen[id]; !exists {
				params.IDs = append(params.IDs, id)
				seen[id] = struct{}{}
			}
		}
		if len(params.IDs) > 50 {
			return params, infraerrors.BadRequest("INVALID_ID", "ids exceeds the maximum of 50 account ids")
		}
	}
	if len(params.IDs) == 0 && params.PlanID <= 0 {
		return params, infraerrors.BadRequest("INVALID_ID", "plan_id is required unless ids is provided")
	}
	return params, nil
}
