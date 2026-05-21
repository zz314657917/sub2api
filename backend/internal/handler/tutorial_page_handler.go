package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
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

func (h *TutorialPageHandler) ListPublished(c *gin.Context) {
	items, err := h.tutorialService.ListPublished(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": dto.TutorialPageSummariesFromService(items)})
}

func (h *TutorialPageHandler) GetPublishedBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	item, err := h.tutorialService.GetPublishedBySlug(c.Request.Context(), slug)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.TutorialPageFromService(item))
}
