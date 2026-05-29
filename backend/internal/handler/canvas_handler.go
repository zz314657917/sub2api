package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type canvasService interface {
	ListCanvases(ctx context.Context, userID int64, filters service.CanvasListFilters) ([]service.CanvasListItem, int, error)
	GetCanvas(ctx context.Context, userID int64, canvasID int64) (*service.CanvasDocument, error)
	SaveCanvas(ctx context.Context, userID int64, input service.CanvasSaveInput) (*service.CanvasDocument, error)
	DeleteCanvas(ctx context.Context, userID int64, canvasID int64) error
	CreateRun(ctx context.Context, userID int64, input service.CanvasRunCreateInput) (*service.CanvasRun, error)
	ListRuns(ctx context.Context, userID int64, filters service.CanvasRunListFilters) ([]service.CanvasRun, int, error)
	GetRun(ctx context.Context, userID int64, runID int64) (*service.CanvasRun, error)
	CancelRun(ctx context.Context, userID int64, runID int64) (*service.CanvasRun, error)
	ListModels(ctx context.Context, userID int64) (service.CanvasModelCatalog, error)
}

type CanvasHandler struct {
	svc canvasService
}

func NewCanvasHandler(svc *service.CanvasService) *CanvasHandler {
	return &CanvasHandler{svc: svc}
}

type canvasListResponse struct {
	Items  []service.CanvasListItem `json:"items"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

type canvasItemResponse struct {
	Item *service.CanvasDocument `json:"item"`
}

type canvasRunListResponse struct {
	Items  []service.CanvasRun `json:"items"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type canvasRunItemResponse struct {
	Item *service.CanvasRun `json:"item"`
}

func (h *CanvasHandler) ListCanvases(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	filters := service.CanvasListFilters{
		Limit:  parseBoundedQueryInt(c, "limit", 20, 1, 100),
		Offset: parseBoundedQueryInt(c, "offset", 0, 0, 100000),
	}
	items, total, err := h.svc.ListCanvases(c.Request.Context(), userID, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasListResponse{Items: items, Total: total, Limit: filters.Limit, Offset: filters.Offset})
}

func (h *CanvasHandler) GetCanvas(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	canvasID, ok := parsePositiveInt64Param(c, "id", "invalid canvas id")
	if !ok {
		return
	}
	item, err := h.svc.GetCanvas(c.Request.Context(), userID, canvasID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasItemResponse{Item: item})
}

func (h *CanvasHandler) SaveCanvas(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	var input service.CanvasSaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.SaveCanvas(c.Request.Context(), userID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasItemResponse{Item: item})
}

func (h *CanvasHandler) UpdateCanvas(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	canvasID, ok := parsePositiveInt64Param(c, "id", "invalid canvas id")
	if !ok {
		return
	}
	var input service.CanvasSaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	input.ID = canvasID
	item, err := h.svc.SaveCanvas(c.Request.Context(), userID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasItemResponse{Item: item})
}

func (h *CanvasHandler) DeleteCanvas(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	canvasID, ok := parsePositiveInt64Param(c, "id", "invalid canvas id")
	if !ok {
		return
	}
	if err := h.svc.DeleteCanvas(c.Request.Context(), userID, canvasID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *CanvasHandler) CreateRun(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	var input service.CanvasRunCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.CreateRun(c.Request.Context(), userID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasRunItemResponse{Item: item})
}

func (h *CanvasHandler) ListRuns(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	filters := service.CanvasRunListFilters{
		CanvasID: parseOptionalPositiveInt64Query(c, "canvas_id"),
		Limit:    parseBoundedQueryInt(c, "limit", 20, 1, 100),
		Offset:   parseBoundedQueryInt(c, "offset", 0, 0, 100000),
	}
	items, total, err := h.svc.ListRuns(c.Request.Context(), userID, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasRunListResponse{Items: items, Total: total, Limit: filters.Limit, Offset: filters.Offset})
}

func (h *CanvasHandler) GetRun(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	runID, ok := parsePositiveInt64Param(c, "id", "invalid canvas run id")
	if !ok {
		return
	}
	item, err := h.svc.GetRun(c.Request.Context(), userID, runID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasRunItemResponse{Item: item})
}

func (h *CanvasHandler) CancelRun(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	runID, ok := parsePositiveInt64Param(c, "id", "invalid canvas run id")
	if !ok {
		return
	}
	item, err := h.svc.CancelRun(c.Request.Context(), userID, runID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, canvasRunItemResponse{Item: item})
}

func (h *CanvasHandler) ListModels(c *gin.Context) {
	userID, ok := currentCanvasUserID(c)
	if !ok {
		return
	}
	catalog, err := h.svc.ListModels(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}

func currentCanvasUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}

func parseOptionalPositiveInt64Query(c *gin.Context, name string) int64 {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}
