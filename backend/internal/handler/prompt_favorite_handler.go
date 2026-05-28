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

type promptFavoriteService interface {
	List(ctx context.Context, userID int64) ([]service.PromptFavorite, error)
	Save(ctx context.Context, userID int64, input service.PromptFavoriteInput) (*service.PromptFavorite, error)
	Delete(ctx context.Context, userID int64, favoriteID int64) error
}

type PromptFavoriteHandler struct {
	svc promptFavoriteService
}

func NewPromptFavoriteHandler(svc *service.PromptFavoriteService) *PromptFavoriteHandler {
	return &PromptFavoriteHandler{svc: svc}
}

type promptFavoriteListResponse struct {
	Items []service.PromptFavorite `json:"items"`
}

type promptFavoriteSaveResponse struct {
	Item  *service.PromptFavorite  `json:"item"`
	Items []service.PromptFavorite `json:"items"`
}

func (h *PromptFavoriteHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	items, err := h.svc.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, promptFavoriteListResponse{Items: normalizePromptFavoriteItems(items)})
}

func (h *PromptFavoriteHandler) Save(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var input service.PromptFavoriteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.Save(c.Request.Context(), subject.UserID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items, err := h.svc.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, promptFavoriteSaveResponse{Item: item, Items: normalizePromptFavoriteItems(items)})
}

func (h *PromptFavoriteHandler) Delete(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	favoriteID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || favoriteID <= 0 {
		response.BadRequest(c, "invalid prompt favorite id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), subject.UserID, favoriteID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	items, err := h.svc.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, promptFavoriteListResponse{Items: normalizePromptFavoriteItems(items)})
}

func normalizePromptFavoriteItems(items []service.PromptFavorite) []service.PromptFavorite {
	if items == nil {
		return []service.PromptFavorite{}
	}
	return items
}
