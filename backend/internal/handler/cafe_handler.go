package handler

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CafeHandler struct {
	service      *service.CafePublicService
	orderService *service.CafeRoomOrderService
}

func NewCafeHandler(cafeService *service.CafePublicService, orderService *service.CafeRoomOrderService) *CafeHandler {
	return &CafeHandler{service: cafeService, orderService: orderService}
}

type createCafeRoomOrderRequest struct {
	ShareCount        int    `json:"share_count" binding:"required"`
	PaymentType       string `json:"payment_type" binding:"required"`
	OpenID            string `json:"openid"`
	ReturnURL         string `json:"return_url"`
	PaymentSource     string `json:"payment_source"`
	IsMobile          *bool  `json:"is_mobile,omitempty"`
	AgreementAccepted bool   `json:"agreement_accepted"`
}

type reserveCafeRoomRequest struct {
	ShareCount        int  `json:"share_count" binding:"required"`
	AgreementAccepted bool `json:"agreement_accepted"`
}

type cafeRoomReservationIdempotencyPayload struct {
	UserID int64                 `json:"user_id"`
	RoomID int64                 `json:"room_id"`
	Input  reserveCafeRoomRequest `json:"input"`
}

type cafeRoomOrderIdempotencyPayload struct {
	UserID int64                      `json:"user_id"`
	RoomID int64                      `json:"room_id"`
	Input  createCafeRoomOrderRequest `json:"input"`
}

func (h *CafeHandler) Overview(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	roomLimit, err := cafePositiveQuery(c.Query("room_limit"), 8)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	overview, err := h.service.Overview(c.Request.Context(), subject.UserID, roomLimit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, overview)
}

func (h *CafeHandler) LobbyActivity(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	activity, err := h.service.LobbyActivity(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, activity)
}

func (h *CafeHandler) ListRooms(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	featured, err := cafeOptionalBool(c.Query("featured"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	rooms, result, err := h.service.List(c.Request.Context(), subject.UserID, service.CafePublicListParams{
		Page:     page,
		PageSize: pageSize,
		Zone:     c.Query("zone"),
		Featured: featured,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, rooms, &response.PaginationResult{
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
		Pages:    result.Pages,
	})
}

func (h *CafeHandler) GetRoom(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	roomID, err := cafeRoomID(c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	detail, err := h.service.Get(c.Request.Context(), subject.UserID, roomID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *CafeHandler) MyRooms(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	statuses, err := service.ParseCafeMyRoomStatuses(c.Query("status"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.MyRooms(c.Request.Context(), subject.UserID, service.CafeMyRoomsListParams{
		Page: page, PageSize: pageSize, Statuses: statuses,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, &response.PaginationResult{
		Total: result.Total, Page: result.Page, PageSize: result.PageSize, Pages: result.Pages,
	})
}

func (h *CafeHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.orderService == nil {
		response.ErrorFrom(c, infraerrors.InternalServer("CAFE_ORDER_SERVICE_UNAVAILABLE", "cafe room order service is unavailable"))
		return
	}
	roomID, err := cafeRoomID(c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req createCafeRoomOrderRequest
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.ShareCount <= 0 || strings.TrimSpace(req.PaymentType) == "" {
		response.BadRequest(c, "Invalid request: share_count and payment_type are required")
		return
	}
	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}
	executeUserIdempotentJSON(c, "cafe_room_order", cafeRoomOrderIdempotencyPayload{
		UserID: subject.UserID,
		RoomID: roomID,
		Input:  req,
	}, 24*time.Hour, func(ctx context.Context) (any, error) {
		return h.orderService.CreateOrder(ctx, service.CafeRoomOrderInput{
			UserID:            subject.UserID,
			RoomID:            roomID,
			ShareCount:        req.ShareCount,
			PaymentType:       req.PaymentType,
			OpenID:            req.OpenID,
			ClientIP:          c.ClientIP(),
			IsMobile:          mobile,
			IsWeChatBrowser:   isWeChatBrowser(c),
			SrcHost:           c.Request.Host,
			SrcURL:            c.Request.Referer(),
			ReturnURL:         req.ReturnURL,
			PaymentSource:     req.PaymentSource,
			AgreementAccepted: req.AgreementAccepted,
		})
	})
}

func (h *CafeHandler) ReserveShares(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok { response.Unauthorized(c, "User not authenticated"); return }
	if h.orderService == nil { response.ErrorFrom(c, infraerrors.InternalServer("CAFE_ORDER_SERVICE_UNAVAILABLE", "cafe room order service is unavailable")); return }
	roomID, err := cafeRoomID(c.Param("id")); if err != nil { response.ErrorFrom(c, err); return }
	var req reserveCafeRoomRequest
	decoder := json.NewDecoder(c.Request.Body); decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil { response.BadRequest(c, "Invalid request: "+err.Error()); return }
	if req.ShareCount <= 0 { response.BadRequest(c, "Invalid request: share_count is required"); return }
	executeUserIdempotentJSON(c, "cafe_room_reservation", cafeRoomReservationIdempotencyPayload{UserID: subject.UserID, RoomID: roomID, Input: req}, 24*time.Hour, func(ctx context.Context) (any, error) {
		return h.orderService.ReserveShares(ctx, service.CafeRoomReservationInput{UserID: subject.UserID, RoomID: roomID, ShareCount: req.ShareCount, AgreementAccepted: req.AgreementAccepted})
	})
}

func cafePositiveQuery(raw string, fallback int) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, infraerrors.BadRequest("CAFE_INVALID_QUERY", "room_limit must be a positive integer")
	}
	return parsed, nil
}

func cafeOptionalBool(raw string) (*bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, infraerrors.BadRequest("CAFE_INVALID_QUERY", "featured must be true or false")
	}
	return &parsed, nil
}

func cafeRoomID(raw string) (int64, error) {
	roomID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || roomID <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", "id is invalid")
	}
	return roomID, nil
}
