package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type WelfareHandler struct {
	welfareService *service.WelfareService
}

func NewWelfareHandler(welfareService *service.WelfareService) *WelfareHandler {
	return &WelfareHandler{welfareService: welfareService}
}

func (h *WelfareHandler) GetOverview(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	result, err := h.welfareService.GetOverview(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *WelfareHandler) GetDailyCheckin(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	result, err := h.welfareService.GetDailyCheckin(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *WelfareHandler) ClaimDailyCheckin(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	result, err := h.welfareService.ClaimDailyCheckin(c.Request.Context(), subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *WelfareHandler) ClaimDailyCheckinMilestone(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	day, err := strconv.Atoi(c.Param("day"))
	if err != nil {
		response.BadRequest(c, "invalid milestone day")
		return
	}
	result, err := h.welfareService.ClaimDailyCheckinMilestone(c.Request.Context(), subject.UserID, day)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}
