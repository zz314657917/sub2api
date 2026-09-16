package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type pelicanAPI interface {
	ListGallery(context.Context, service.PelicanAuthorization, int64, int64, int, int) (*service.PelicanListResponse, error)
	ListHistory(context.Context, service.PelicanAuthorization, int64, int64) ([]service.PelicanResult, error)
	GetResult(context.Context, service.PelicanAuthorization, int64) (*service.PelicanResult, error)
	ListPlans(context.Context) ([]service.PelicanPlan, error)
	CreatePlan(context.Context, service.PelicanPlanInput) (*service.PelicanPlan, error)
	UpdatePlan(context.Context, int64, service.PelicanPlanInput) (*service.PelicanPlan, error)
	DeletePlan(context.Context, int64) error
	RunNow(context.Context, int64) error
}

type pelicanGroupAccess interface {
	GetAvailableGroups(context.Context, int64) ([]service.Group, error)
}

func pelicanError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPelicanNotFound):
		response.NotFound(c, "Test result or plan not found")
	case errors.Is(err, service.ErrPelicanConflict):
		response.Error(c, http.StatusConflict, "Test plan is already running")
	case errors.Is(err, service.ErrPelicanInvalid):
		response.BadRequest(c, "Invalid or unsupported test plan")
	default:
		response.InternalError(c, "Unable to process pelican test request")
	}
}

// PelicanTestHandler exposes redacted results and administrator-only plans.
type PelicanTestHandler struct {
	svc    pelicanAPI
	groups pelicanGroupAccess
}

func NewPelicanTestHandler(svc *service.PelicanTestService, groups *service.APIKeyService) *PelicanTestHandler {
	return &PelicanTestHandler{svc: svc, groups: groups}
}

func (h *PelicanTestHandler) authorization(c *gin.Context) (service.PelicanAuthorization, bool) {
	c.Header("Cache-Control", "private, no-store")
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return service.PelicanAuthorization{}, false
	}
	role, _ := middleware.GetUserRoleFromContext(c)
	if role == service.RoleAdmin {
		return service.PelicanAuthorization{Admin: true}, true
	}
	groups, err := h.groups.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.InternalError(c, "Unable to resolve test group access")
		return service.PelicanAuthorization{}, false
	}
	auth := service.PelicanAuthorization{GroupIDs: make([]int64, 0, len(groups))}
	for _, group := range groups {
		auth.GroupIDs = append(auth.GroupIDs, group.ID)
	}
	return auth, true
}

func pelicanAdmin(c *gin.Context) bool {
	role, ok := middleware.GetUserRoleFromContext(c)
	if !ok || role != service.RoleAdmin {
		response.Forbidden(c, "Administrator access required")
		return false
	}
	return true
}

func pelicanPositiveID(c *gin.Context, value string, optional bool) (int64, bool) {
	if value == "" && optional {
		return 0, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid identifier")
		return 0, false
	}
	return id, true
}

func (h *PelicanTestHandler) List(c *gin.Context) {
	auth, ok := h.authorization(c)
	if !ok {
		return
	}
	groupID, ok := pelicanPositiveID(c, c.Query("group_id"), true)
	if !ok {
		return
	}
	accountID, ok := pelicanPositiveID(c, c.Query("account_id"), true)
	if !ok {
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 100000 {
		response.BadRequest(c, "Invalid page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "24"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		response.BadRequest(c, "Invalid page size")
		return
	}
	result, err := h.svc.ListGallery(c.Request.Context(), auth, groupID, accountID, page, pageSize)
	if err != nil {
		pelicanError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PelicanTestHandler) History(c *gin.Context) {
	auth, ok := h.authorization(c)
	if !ok {
		return
	}
	planID, ok := pelicanPositiveID(c, c.Query("plan_id"), false)
	if !ok {
		return
	}
	accountID, ok := pelicanPositiveID(c, c.Query("account_id"), false)
	if !ok {
		return
	}
	result, err := h.svc.ListHistory(c.Request.Context(), auth, planID, accountID)
	if err != nil {
		pelicanError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PelicanTestHandler) Result(c *gin.Context) {
	auth, ok := h.authorization(c)
	if !ok {
		return
	}
	id, ok := pelicanPositiveID(c, c.Param("id"), false)
	if !ok {
		return
	}
	result, err := h.svc.GetResult(c.Request.Context(), auth, id)
	if err != nil {
		pelicanError(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	response.Success(c, result)
}

func (h *PelicanTestHandler) ListPlans(c *gin.Context) {
	if !pelicanAdmin(c) {
		return
	}
	plans, err := h.svc.ListPlans(c.Request.Context())
	if err != nil {
		pelicanError(c, err)
		return
	}
	response.Success(c, plans)
}

func (h *PelicanTestHandler) SavePlan(c *gin.Context) {
	if !pelicanAdmin(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var input service.PelicanPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid test plan")
		return
	}
	if c.Param("id") == "" {
		plan, err := h.svc.CreatePlan(c.Request.Context(), input)
		if err != nil {
			pelicanError(c, err)
			return
		}
		response.Success(c, plan)
		return
	}
	id, ok := pelicanPositiveID(c, c.Param("id"), false)
	if !ok {
		return
	}
	plan, err := h.svc.UpdatePlan(c.Request.Context(), id, input)
	if err != nil {
		pelicanError(c, err)
		return
	}
	response.Success(c, plan)
}

func (h *PelicanTestHandler) DeletePlan(c *gin.Context) {
	if !pelicanAdmin(c) {
		return
	}
	id, ok := pelicanPositiveID(c, c.Param("id"), false)
	if !ok {
		return
	}
	if err := h.svc.DeletePlan(c.Request.Context(), id); err != nil {
		pelicanError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *PelicanTestHandler) RunPlan(c *gin.Context) {
	if !pelicanAdmin(c) {
		return
	}
	id, ok := pelicanPositiveID(c, c.Param("id"), false)
	if !ok {
		return
	}
	if err := h.svc.RunNow(c.Request.Context(), id); err != nil {
		pelicanError(c, err)
		return
	}
	response.Accepted(c, gin.H{"queued": true})
}
