package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const openWebUIRedeemSecretHeader = "X-Sub2API-OpenWebUI-Secret"

type OpenWebUIHandler struct {
	launchService *service.OpenWebUILaunchService
}

func NewOpenWebUIHandler(launchService *service.OpenWebUILaunchService) *OpenWebUIHandler {
	return &OpenWebUIHandler{launchService: launchService}
}

type openWebUILaunchRequest struct {
	APIKeyID int64 `json:"api_key_id"`
}

type openWebUIRedeemRequest struct {
	Token string `json:"token" binding:"required"`
}

type openWebUIAPIKeysRequest struct {
	SessionToken string `json:"session_token" binding:"required"`
}

type openWebUIBindingRequest struct {
	SessionToken string `json:"session_token" binding:"required"`
	APIKeyID     int64  `json:"api_key_id" binding:"required"`
}

func (h *OpenWebUIHandler) Launch(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req openWebUILaunchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	launch, err := h.launchService.CreateLaunch(c.Request.Context(), subject.UserID, req.APIKeyID, requestGatewayBaseURL(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, launch)
}

func (h *OpenWebUIHandler) Redeem(c *gin.Context) {
	var req openWebUIRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.launchService.Redeem(c.Request.Context(), req.Token, redeemSecretFromRequest(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

func (h *OpenWebUIHandler) APIKeys(c *gin.Context) {
	if !h.requireInternalSecret(c) {
		return
	}
	var req openWebUIAPIKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	items, err := h.launchService.ListAPIKeysBySession(c.Request.Context(), req.SessionToken)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *OpenWebUIHandler) BindAPIKey(c *gin.Context) {
	if !h.requireInternalSecret(c) {
		return
	}
	var req openWebUIBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.launchService.BindAPIKeyBySession(c.Request.Context(), req.SessionToken, req.APIKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *OpenWebUIHandler) requireInternalSecret(c *gin.Context) bool {
	if h == nil || h.launchService == nil || !h.launchService.ValidRedeemSecret(redeemSecretFromRequest(c)) {
		response.ErrorFrom(c, service.ErrOpenWebUIInvalidSecret)
		return false
	}
	return true
}

func redeemSecretFromRequest(c *gin.Context) string {
	if secret := strings.TrimSpace(c.GetHeader(openWebUIRedeemSecretHeader)); secret != "" {
		return secret
	}
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if token, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}

func requestGatewayBaseURL(c *gin.Context) string {
	host := firstForwardedValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" && c.Request != nil {
		host = c.Request.Host
	}
	if host == "" {
		return ""
	}

	proto := firstForwardedValue(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		proto = "http"
		if c.Request != nil && c.Request.TLS != nil {
			proto = "https"
		}
	}
	return strings.ToLower(proto) + "://" + host + "/v1"
}

func firstForwardedValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if i := strings.IndexByte(value, ','); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}
