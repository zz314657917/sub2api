package handler

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const studioBridgeSecretHeader = "X-Sub2API-Studio-Secret"

type StudioBridgeHandler struct {
	service *service.StudioBridgeService
}

func NewStudioBridgeHandler(service *service.StudioBridgeService) *StudioBridgeHandler {
	return &StudioBridgeHandler{service: service}
}

type studioBridgeLaunchRequest struct {
	AppID     string `json:"app_id"`
	ReturnURL string `json:"return_url"`
}

type studioBridgeRedeemRequest struct {
	AppID string `json:"app_id"`
	Token string `json:"launch_token" binding:"required"`
}

type studioBridgeUserSummaryRequest struct {
	AppID  string `json:"app_id"`
	UserID int64  `json:"user_id" binding:"required"`
}

type studioBridgeSessionProbeResponse struct {
	UserID int64 `json:"user_id"`
}

func (h *StudioBridgeHandler) Launch(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req studioBridgeLaunchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.AppID == "" {
		req.AppID = service.StudioBridgeAppLuoyeAI
	}
	result, err := h.service.CreateLaunch(c.Request.Context(), subject.UserID, req.AppID, req.ReturnURL)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *StudioBridgeHandler) Redeem(c *gin.Context) {
	var req studioBridgeRedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.AppID == "" {
		req.AppID = service.StudioBridgeAppLuoyeAI
	}
	result, err := h.service.RedeemLaunch(c.Request.Context(), req.AppID, req.Token, studioBridgeSecretFromRequest(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *StudioBridgeHandler) UserSummary(c *gin.Context) {
	var req studioBridgeUserSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.AppID == "" {
		req.AppID = service.StudioBridgeAppLuoyeAI
	}
	result, err := h.service.GetUserSummary(c.Request.Context(), req.AppID, req.UserID, studioBridgeSecretFromRequest(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *StudioBridgeHandler) SessionProbe(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	appID := strings.TrimSpace(c.Query("app_id"))
	if appID == "" {
		appID = service.StudioBridgeAppLuoyeAI
	}
	if err := h.service.ValidateSessionProbeOrigin(c.Request.Context(), appID, c.Query("parent_origin")); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, studioBridgeSessionProbeResponse{UserID: subject.UserID})
}

func (h *StudioBridgeHandler) Reserve(c *gin.Context) {
	h.handleCharge(c, h.service.Reserve)
}

func (h *StudioBridgeHandler) Commit(c *gin.Context) {
	h.handleCharge(c, h.service.Commit)
}

func (h *StudioBridgeHandler) Refund(c *gin.Context) {
	h.handleCharge(c, h.service.Refund)
}

func (h *StudioBridgeHandler) handleCharge(c *gin.Context, fn func(context.Context, service.StudioBridgeChargeCommand, string) (*service.StudioBridgeChargeResult, error)) {
	var req service.StudioBridgeChargeCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.AppID == "" {
		req.AppID = service.StudioBridgeAppLuoyeAI
	}
	result, err := fn(c.Request.Context(), req, studioBridgeSecretFromRequest(c))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func studioBridgeSecretFromRequest(c *gin.Context) string {
	if secret := strings.TrimSpace(c.GetHeader(studioBridgeSecretHeader)); secret != "" {
		return secret
	}
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if token, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}
