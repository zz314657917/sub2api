package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserProxyHandler struct {
	userProxyService *service.UserProxyService
}

func NewUserProxyHandler(userProxyService *service.UserProxyService) *UserProxyHandler {
	return &UserProxyHandler{userProxyService: userProxyService}
}

type userProxyCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Protocol string `json:"protocol" binding:"required,oneof=http https socks5 socks5h"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type userProxyUpdateRequest struct {
	Name     *string `json:"name"`
	Protocol *string `json:"protocol" binding:"omitempty,oneof=http https socks5 socks5h"`
	Host     *string `json:"host"`
	Port     *int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Status   *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

func (h *UserProxyHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	proxies, err := h.userProxyService.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.Proxy, 0, len(proxies))
	for i := range proxies {
		out = append(out, *dto.ProxyFromService(&proxies[i]))
	}
	response.Success(c, out)
}

func (h *UserProxyHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req userProxyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	executeUserIdempotentJSON(c, "user.proxies.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		proxy, err := h.userProxyService.Create(ctx, subject.UserID, service.CreateProxyRequest{
			Name:     strings.TrimSpace(req.Name),
			Protocol: strings.TrimSpace(req.Protocol),
			Host:     strings.TrimSpace(req.Host),
			Port:     req.Port,
			Username: strings.TrimSpace(req.Username),
			Password: strings.TrimSpace(req.Password),
		})
		if err != nil {
			return nil, err
		}
		return dto.ProxyFromService(proxy), nil
	})
}

func (h *UserProxyHandler) Update(c *gin.Context) {
	subject, proxyID, ok := h.requireSubjectAndProxyID(c)
	if !ok {
		return
	}
	var req userProxyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	proxy, err := h.userProxyService.Update(c.Request.Context(), subject.UserID, proxyID, service.UpdateProxyRequest{
		Name:     trimmedStringPtr(req.Name),
		Protocol: trimmedStringPtr(req.Protocol),
		Host:     trimmedStringPtr(req.Host),
		Port:     req.Port,
		Username: trimmedStringPtr(req.Username),
		Password: trimmedStringPtr(req.Password),
		Status:   trimmedStringPtr(req.Status),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.ProxyFromService(proxy))
}

func (h *UserProxyHandler) Delete(c *gin.Context) {
	subject, proxyID, ok := h.requireSubjectAndProxyID(c)
	if !ok {
		return
	}
	if err := h.userProxyService.Delete(c.Request.Context(), subject.UserID, proxyID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Proxy deleted successfully"})
}

func (h *UserProxyHandler) requireSubjectAndProxyID(c *gin.Context) (middleware2.AuthSubject, int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return middleware2.AuthSubject{}, 0, false
	}
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || proxyID <= 0 {
		response.BadRequest(c, "Invalid proxy ID")
		return middleware2.AuthSubject{}, 0, false
	}
	return subject, proxyID, true
}
