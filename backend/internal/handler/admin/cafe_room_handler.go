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

type CafeRoomHandler struct {
	service *service.CafeRoomService
}

func NewCafeRoomHandler(cafeRoomService *service.CafeRoomService) *CafeRoomHandler {
	return &CafeRoomHandler{service: cafeRoomService}
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

func parseCafeRoomPositiveInt64Param(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param(name)), 10, 64)
	if err != nil || id <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", name+" is invalid")
	}
	return id, nil
}
