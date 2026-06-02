package handler

import (
	"net/http"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *GatewayHandler) resolveAPIKeyForModelRequest(c *gin.Context, apiKey *service.APIKey, requestedModel string, imageIntent bool) (*service.APIKey, bool) {
	resolved, ok := middleware2.ResolveAPIKeyForModelRequest(c, h.apiKeyService, apiKey, requestedModel, imageIntent)
	if !ok {
		return nil, false
	}
	return resolved, true
}

func (h *OpenAIGatewayHandler) resolveAPIKeyForModelRequest(c *gin.Context, apiKey *service.APIKey, requestedModel string, imageIntent bool) (*service.APIKey, bool) {
	resolved, ok := middleware2.ResolveAPIKeyForModelRequest(c, h.apiKeyService, apiKey, requestedModel, imageIntent)
	if !ok {
		return nil, false
	}
	return resolved, true
}

func (h *OpenAIGatewayHandler) ensureImageGenerationAllowed(c *gin.Context, apiKey *service.APIKey) bool {
	if service.GroupAllowsImageGeneration(apiKey.Group) {
		return true
	}
	h.errorResponse(c, http.StatusForbidden, "permission_error", service.ImageGenerationPermissionMessage())
	return false
}
