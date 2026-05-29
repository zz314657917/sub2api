package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageCreatorStorageGovernanceHandler struct {
	service *service.ImageCreatorStorageGovernanceService
}

func NewImageCreatorStorageGovernanceHandler(service *service.ImageCreatorStorageGovernanceService) *ImageCreatorStorageGovernanceHandler {
	return &ImageCreatorStorageGovernanceHandler{service: service}
}

func (h *ImageCreatorStorageGovernanceHandler) GetStats(c *gin.Context) {
	if h == nil || h.service == nil {
		response.InternalError(c, "image creator storage governance service is unavailable")
		return
	}
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *ImageCreatorStorageGovernanceHandler) Cleanup(c *gin.Context) {
	if h == nil || h.service == nil {
		response.InternalError(c, "image creator storage governance service is unavailable")
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.Cleanup(c.Request.Context(), strings.TrimSpace(req.Action))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
