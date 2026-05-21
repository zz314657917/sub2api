package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type TutorialPageHandler struct {
	tutorialService *service.TutorialPageService
}

func NewTutorialPageHandler(tutorialService *service.TutorialPageService) *TutorialPageHandler {
	return &TutorialPageHandler{tutorialService: tutorialService}
}

type createTutorialPageRequest struct {
	Slug        string `json:"slug" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status" binding:"omitempty,oneof=draft published"`
	ContentMD   string `json:"content_md" binding:"required"`
}

type updateTutorialPageRequest struct {
	Slug        *string `json:"slug"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	SortOrder   *int    `json:"sort_order"`
	Status      *string `json:"status" binding:"omitempty,oneof=draft published"`
	ContentMD   *string `json:"content_md"`
}

type updateTutorialPageStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=draft published"`
}

func (h *TutorialPageHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "sort_order"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
	}
	items, pageResult, err := h.tutorialService.List(c.Request.Context(), params, service.TutorialPageListFilters{
		Status:   strings.TrimSpace(c.Query("status")),
		Category: strings.TrimSpace(c.Query("category")),
		Search:   strings.TrimSpace(c.Query("search")),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.TutorialPageSummary, 0, len(items))
	for i := range items {
		out = append(out, *dto.TutorialPageSummaryFromService(&items[i]))
	}
	response.Paginated(c, out, pageResult.Total, pageResult.Page, pageResult.PageSize)
}

func (h *TutorialPageHandler) Get(c *gin.Context) {
	id, ok := parseTutorialID(c)
	if !ok {
		return
	}
	item, err := h.tutorialService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.TutorialPageFromService(item))
}

func (h *TutorialPageHandler) Create(c *gin.Context) {
	var req createTutorialPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.tutorialService.Create(c.Request.Context(), service.CreateTutorialPageInput{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
		ContentMD:   req.ContentMD,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.TutorialPageFromService(item))
}

func (h *TutorialPageHandler) Update(c *gin.Context) {
	id, ok := parseTutorialID(c)
	if !ok {
		return
	}
	var req updateTutorialPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.tutorialService.Update(c.Request.Context(), id, service.UpdateTutorialPageInput{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
		ContentMD:   req.ContentMD,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.TutorialPageFromService(item))
}

func (h *TutorialPageHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseTutorialID(c)
	if !ok {
		return
	}
	var req updateTutorialPageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.tutorialService.SetStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.TutorialPageFromService(item))
}

func (h *TutorialPageHandler) Delete(c *gin.Context) {
	id, ok := parseTutorialID(c)
	if !ok {
		return
	}
	if err := h.tutorialService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Tutorial page deleted successfully"})
}

func parseTutorialID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TUTORIAL_PAGE_ID", "invalid tutorial page id"))
		return 0, false
	}
	return id, true
}
