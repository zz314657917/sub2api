package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	CafeRoomStatusDraft            = "draft"
	CafeRoomStatusEnabled          = "enabled"
	CafeRoomStatusMaintenance      = "maintenance"
	CafeRoomStatusDisabled         = "disabled"
	CafeRoomFulfillmentMode        = "room_subscription"
	CafeRoomGroupAccessMode        = "room_managed"
	CafeRoundStatusOpen            = "open"
	CafeRoundStatusAwaitingAccount = "awaiting_account"
	cafeRoomAccountOptionMaxIDs    = 50
)

var (
	ErrCafeRoomNotFound               = errors.NotFound("CAFE_ROOM_NOT_FOUND", "cafe room not found")
	ErrCafeRoomInvalid                = errors.BadRequest("CAFE_ROOM_INVALID", "cafe room input is invalid")
	ErrCafePlanNotFound               = errors.NotFound("CAFE_PLAN_NOT_FOUND", "cafe room plan not found")
	ErrCafePlanInvalid                = errors.BadRequest("CAFE_PLAN_INVALID", "plan is not configured for room subscriptions")
	ErrCafeGroupInvalid               = errors.BadRequest("CAFE_GROUP_INVALID", "plan group is not configured for room management")
	ErrCafeAccountNotFound            = errors.NotFound("CAFE_ACCOUNT_NOT_FOUND", "cafe room account not found")
	ErrCafeAccountIncompatible        = errors.BadRequest("CAFE_ACCOUNT_INCOMPATIBLE", "account is not compatible with the room group")
	ErrCafeAccountAssigned            = errors.Conflict("CAFE_ACCOUNT_ALREADY_ASSIGNED", "account is already assigned to another cafe room")
	ErrCafeRoomLive                   = errors.Conflict("CAFE_ROOM_LIVE_ROUND", "room has a live round")
	ErrCafeRoundExists                = errors.Conflict("CAFE_ROOM_OPEN_ROUND_EXISTS", "room already has a live round")
	ErrCafeRoomDisabled               = errors.Conflict("CAFE_ROOM_DISABLED", "disabled rooms cannot open a round")
	ErrCafeRoomEnabled                = errors.Conflict("CAFE_ROOM_ENABLED", "enabled rooms cannot be deleted")
	ErrCafeRoundNotAwaitingAccount    = errors.Conflict("CAFE_ROUND_NOT_AWAITING_ACCOUNT", "cafe round is not awaiting an account")
	ErrCafeAccountTierMismatch        = errors.BadRequest("CAFE_ACCOUNT_TIER_MISMATCH", "account subscription tier does not match the cafe round")
	ErrCafeAccountAlreadyInUse        = errors.Conflict("CAFE_ACCOUNT_ALREADY_IN_USE", "account is already assigned to an active cafe round")
	ErrCafeFulfillmentDeadlineExpired = errors.Conflict("CAFE_FULFILLMENT_DEADLINE_EXPIRED", "cafe fulfillment deadline has expired")
)

type CafeRoom struct {
	ID           int64          `json:"id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	PlanID       int64          `json:"plan_id"`
	AccountID    *int64         `json:"account_id,omitempty"`
	ZoneKey      string         `json:"zone_key"`
	ThemeKey     string         `json:"theme_key"`
	SceneSlotKey string         `json:"scene_slot_key"`
	Status       string         `json:"status"`
	Featured     bool           `json:"featured"`
	SortOrder    int            `json:"sort_order"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Plan         *CafeRoomPlan  `json:"plan,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CafeRoomPlan struct {
	ID                        int64  `json:"id"`
	Title                     string `json:"title"`
	Status                    string `json:"status"`
	TargetGroupID             int64  `json:"target_group_id"`
	FulfillmentMode           string `json:"fulfillment_mode"`
	AutoCreateRoomKey         bool   `json:"auto_create_room_key"`
	TotalShares               int    `json:"total_shares"`
	SubscriptionTier          string `json:"subscription_tier"`
	MaxBuyers                 int    `json:"max_buyers"`
	MaxSharesPerUser          int    `json:"max_shares_per_user"`
	FulfillmentTimeoutMinutes int    `json:"fulfillment_timeout_minutes"`
	SeatCount                 int    `json:"seat_count"`
	TimeoutMinutes            int    `json:"timeout_minutes"`
	ValidityDays              int    `json:"validity_days"`
	GroupPlatform             string `json:"group_platform"`
	GroupAccessMode           string `json:"group_access_mode"`
	TargetGroupStatus         string `json:"target_group_status"`
}

type CafeRound struct {
	ID                int64     `json:"id"`
	PlanID            int64     `json:"plan_id"`
	CafeRoomID        *int64    `json:"cafe_room_id,omitempty"`
	AssignedAccountID *int64    `json:"assigned_account_id,omitempty"`
	RoomCodeSnapshot  *string   `json:"room_code_snapshot,omitempty"`
	RoomNameSnapshot  *string   `json:"room_name_snapshot,omitempty"`
	Status            string    `json:"status"`
	TotalShares       int       `json:"total_shares"`
	TotalSeats        int       `json:"total_seats"`
	DeadlineAt        time.Time `json:"deadline_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CafeRoomInput struct {
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	PlanID       int64          `json:"plan_id"`
	AccountID    *int64         `json:"account_id"`
	ZoneKey      string         `json:"zone_key"`
	ThemeKey     string         `json:"theme_key"`
	SceneSlotKey string         `json:"scene_slot_key"`
	Status       string         `json:"status"`
	Featured     bool           `json:"featured"`
	SortOrder    int            `json:"sort_order"`
	Metadata     map[string]any `json:"metadata"`
}

type CafeRoomUpdateInput struct {
	Code         *string         `json:"code"`
	Name         *string         `json:"name"`
	PlanID       *int64          `json:"plan_id"`
	AccountID    *int64          `json:"account_id"`
	ClearAccount bool            `json:"clear_account"`
	ZoneKey      *string         `json:"zone_key"`
	ThemeKey     *string         `json:"theme_key"`
	SceneSlotKey *string         `json:"scene_slot_key"`
	Status       *string         `json:"status"`
	Featured     *bool           `json:"featured"`
	SortOrder    *int            `json:"sort_order"`
	Metadata     *map[string]any `json:"metadata"`
}

type CafeRoomBulkInput struct {
	PlanID          int64   `json:"plan_id"`
	Quantity        int     `json:"quantity"`
	Count           int     `json:"-"`
	AccountIDs      []int64 `json:"-"`
	CodePrefix      string  `json:"code_prefix"`
	StartNumber     int     `json:"start_number"`
	ZoneKey         string  `json:"zone_key"`
	ThemeKey        string  `json:"theme_key"`
	CreateOpenRound bool    `json:"create_open_round"`
}

type CafeRoomBulkResult struct {
	Created []CafeRoomBulkCreated `json:"created"`
	Failed  []CafeRoomBulkFailure `json:"failed"`
}

type CafeRoomBulkCreated struct {
	Room  *CafeRoom  `json:"room"`
	Round *CafeRound `json:"round,omitempty"`
}

type CafeRoomBulkFailure struct {
	Index   int    `json:"index"`
	Code    string `json:"error_code"`
	Message string `json:"message"`
}

// CafeRoomAccountOption is deliberately narrower than the administrator Account DTO.
// It is safe to use in the Pixel Cafe room picker and never includes credentials.
type CafeRoomAccountOption struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Status      string `json:"status"`
	EmailMasked string `json:"email_masked,omitempty"`
}

type CafeRoomAccountOptionParams struct {
	Page          int
	PageSize      int
	Search        string
	PlanID        int64
	ExcludeRoomID int64
	IDs           []int64
}

type CafeRoomRepository interface {
	List(ctx context.Context, params pagination.PaginationParams, status, zone, search string) ([]CafeRoom, *pagination.PaginationResult, error)
	GetByID(ctx context.Context, id int64) (*CafeRoom, error)
	GetPlan(ctx context.Context, id int64) (*CafeRoomPlan, error)
	GetAccount(ctx context.Context, id int64) (status, platform string, groupIDs []int64, err error)
	ListAccountOptions(ctx context.Context, params CafeRoomAccountOptionParams) ([]CafeRoomAccountOption, *pagination.PaginationResult, error)
	HasOperationalAccount(ctx context.Context, accountID, excludeRoomID int64) (bool, error)
	HasLiveRound(ctx context.Context, roomID int64) (bool, error)
	Create(ctx context.Context, room *CafeRoom) (*CafeRoom, error)
	Update(ctx context.Context, room *CafeRoom) (*CafeRoom, error)
	Delete(ctx context.Context, id int64) error
	CreateOpenRound(ctx context.Context, roomID int64, now time.Time) (*CafeRound, error)
}

type CafeRoomService struct {
	repo CafeRoomRepository
	now  func() time.Time
}

func NewCafeRoomService(repo CafeRoomRepository) *CafeRoomService {
	return &CafeRoomService{repo: repo, now: time.Now}
}

func (s *CafeRoomService) List(ctx context.Context, params pagination.PaginationParams, status, zone, search string) ([]CafeRoom, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.InternalServer("CAFE_ROOM_SERVICE_UNAVAILABLE", "cafe room service is unavailable")
	}
	return s.repo.List(ctx, params, strings.TrimSpace(status), strings.TrimSpace(zone), strings.TrimSpace(search))
}

func (s *CafeRoomService) Get(ctx context.Context, id int64) (*CafeRoom, error) {
	if id <= 0 {
		return nil, ErrCafeRoomInvalid
	}
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrCafeRoomNotFound
	}
	return room, nil
}

func (s *CafeRoomService) ListAccountOptions(ctx context.Context, params CafeRoomAccountOptionParams) ([]CafeRoomAccountOption, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, errors.InternalServer("CAFE_ROOM_SERVICE_UNAVAILABLE", "cafe room service is unavailable")
	}
	if len(params.IDs) > 0 {
		if len(params.IDs) > cafeRoomAccountOptionMaxIDs {
			return nil, nil, ErrCafeRoomInvalid
		}
		return s.repo.ListAccountOptions(ctx, params)
	}
	if params.PlanID <= 0 {
		return nil, nil, ErrCafeRoomInvalid
	}
	plan, err := s.repo.GetPlan(ctx, params.PlanID)
	if err != nil {
		return nil, nil, err
	}
	if err := validateCafeOperationalPlan(plan); err != nil {
		return nil, nil, err
	}
	params.Search = strings.TrimSpace(params.Search)
	return s.repo.ListAccountOptions(ctx, params)
}

func (s *CafeRoomService) Create(ctx context.Context, input CafeRoomInput) (*CafeRoom, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	var err error
	input.Status, err = normalizeCafeRoomStatus(input.Status)
	if err != nil {
		return nil, err
	}
	if input.Code == "" || input.Name == "" || input.PlanID <= 0 {
		return nil, ErrCafeRoomInvalid
	}
	if err := s.validatePlan(ctx, input.PlanID); err != nil {
		return nil, err
	}
	room := &CafeRoom{Code: input.Code, Name: input.Name, PlanID: input.PlanID, AccountID: input.AccountID, ZoneKey: strings.TrimSpace(input.ZoneKey), ThemeKey: strings.TrimSpace(input.ThemeKey), SceneSlotKey: strings.TrimSpace(input.SceneSlotKey), Status: input.Status, Featured: input.Featured, SortOrder: input.SortOrder, Metadata: input.Metadata}
	if room.ZoneKey == "" {
		room.ZoneKey = "featured"
	}
	if room.ThemeKey == "" {
		room.ThemeKey = "warm_wood"
	}
	created, err := s.repo.Create(ctx, room)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *CafeRoomService) Update(ctx context.Context, id int64, input CafeRoomUpdateInput) (*CafeRoom, error) {
	room, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Code != nil {
		room.Code = strings.TrimSpace(*input.Code)
	}
	if input.Name != nil {
		room.Name = strings.TrimSpace(*input.Name)
	}
	if input.PlanID != nil {
		room.PlanID = *input.PlanID
	}
	if input.AccountID != nil {
		room.AccountID = input.AccountID
	}
	if input.ClearAccount {
		room.AccountID = nil
	}
	if input.ZoneKey != nil {
		room.ZoneKey = strings.TrimSpace(*input.ZoneKey)
	}
	if input.ThemeKey != nil {
		room.ThemeKey = strings.TrimSpace(*input.ThemeKey)
	}
	if input.SceneSlotKey != nil {
		room.SceneSlotKey = strings.TrimSpace(*input.SceneSlotKey)
	}
	if input.Status != nil {
		room.Status, err = normalizeCafeRoomStatus(*input.Status)
		if err != nil {
			return nil, err
		}
	}
	if input.Featured != nil {
		room.Featured = *input.Featured
	}
	if input.SortOrder != nil {
		room.SortOrder = *input.SortOrder
	}
	if input.Metadata != nil {
		room.Metadata = *input.Metadata
	}
	if room.Code == "" || room.Name == "" {
		return nil, ErrCafeRoomInvalid
	}
	if err := s.validatePlan(ctx, room.PlanID); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, room)
}

func (s *CafeRoomService) Delete(ctx context.Context, id int64) error {
	room, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	live, err := s.repo.HasLiveRound(ctx, id)
	if err != nil {
		return err
	}
	if live {
		return ErrCafeRoomLive
	}
	if room.Status == CafeRoomStatusEnabled {
		return ErrCafeRoomEnabled
	}
	return s.repo.Delete(ctx, id)
}

func (s *CafeRoomService) OpenRound(ctx context.Context, id int64) (*CafeRound, error) {
	if id <= 0 {
		return nil, ErrCafeRoomInvalid
	}
	room, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if room.Status != CafeRoomStatusEnabled {
		return nil, ErrCafeRoomDisabled
	}
	return s.repo.CreateOpenRound(ctx, id, s.now())
}

func (s *CafeRoomService) BulkCreate(ctx context.Context, input CafeRoomBulkInput) CafeRoomBulkResult {
	result := CafeRoomBulkResult{Created: []CafeRoomBulkCreated{}, Failed: []CafeRoomBulkFailure{}}
	quantity := input.Quantity
	if quantity <= 0 {
		quantity = input.Count
	}
	if input.PlanID <= 0 || quantity <= 0 || quantity > 100 {
		result.Failed = append(result.Failed, CafeRoomBulkFailure{Code: errors.Reason(ErrCafeRoomInvalid), Message: errors.Message(ErrCafeRoomInvalid)})
		return result
	}
	prefix := strings.TrimSpace(input.CodePrefix)
	if prefix == "" {
		prefix = "ROOM-"
	}
	if input.StartNumber < 1 {
		input.StartNumber = 1
	}
	for i := 0; i < quantity; i++ {
		code := prefix + formatCafeRoomNumber(input.StartNumber+i)
		room, err := s.Create(ctx, CafeRoomInput{Code: code, Name: code, PlanID: input.PlanID, ZoneKey: input.ZoneKey, ThemeKey: input.ThemeKey, Status: CafeRoomStatusEnabled})
		if err != nil {
			result.Failed = append(result.Failed, CafeRoomBulkFailure{Index: i, Code: errors.Reason(err), Message: errors.Message(err)})
			continue
		}
		var round *CafeRound
		if input.CreateOpenRound {
			round, err = s.OpenRound(ctx, room.ID)
			if err != nil {
				_ = s.repo.Delete(ctx, room.ID)
				result.Failed = append(result.Failed, CafeRoomBulkFailure{Index: i, Code: errors.Reason(err), Message: errors.Message(err)})
				continue
			}
		}
		result.Created = append(result.Created, CafeRoomBulkCreated{Room: room, Round: round})
	}
	return result
}

func (s *CafeRoomService) validatePlan(ctx context.Context, planID int64) error {
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return err
	}
	return validateCafeOperationalPlan(plan)
}

func (s *CafeRoomService) validatePlanAccount(ctx context.Context, planID, accountID, excludeRoomID int64, status string) error {
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return err
	}
	if plan == nil {
		return ErrCafePlanNotFound
	}
	if err := validateCafeOperationalPlan(plan); err != nil {
		return err
	}
	accountStatus, platform, groupIDs, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}
	if accountStatus == "" {
		return ErrCafeAccountNotFound
	}
	if accountStatus != StatusActive || platform != plan.GroupPlatform || !containsCafeGroupID(groupIDs, plan.TargetGroupID) {
		return ErrCafeAccountIncompatible
	}
	if status != CafeRoomStatusDisabled {
		assigned, err := s.repo.HasOperationalAccount(ctx, accountID, excludeRoomID)
		if err != nil {
			return err
		}
		if assigned {
			return ErrCafeAccountAssigned
		}
	}
	return nil
}

func validateCafeOperationalPlan(plan *CafeRoomPlan) error {
	if plan != nil {
		plan.TotalShares, plan.SubscriptionTier, plan.MaxBuyers, plan.MaxSharesPerUser, plan.FulfillmentTimeoutMinutes = normalizedCafePlanPolicy(
			plan.TotalShares, plan.SeatCount, plan.SubscriptionTier, plan.MaxBuyers, plan.MaxSharesPerUser, plan.FulfillmentTimeoutMinutes,
		)
	}
	if plan == nil || plan.Status != GroupBuyPlanStatusActive ||
		plan.FulfillmentMode != CafeRoomFulfillmentMode ||
		!plan.AutoCreateRoomKey || plan.ValidityDays <= 0 ||
		(plan.SubscriptionTier != "plus" && plan.SubscriptionTier != "pro") ||
		plan.TotalShares < 1 || plan.TotalShares > 10 || plan.MaxBuyers < 1 || plan.MaxBuyers > plan.TotalShares ||
		plan.MaxSharesPerUser < 1 || plan.MaxSharesPerUser > plan.TotalShares || plan.FulfillmentTimeoutMinutes <= 0 {
		return ErrCafePlanInvalid
	}
	if plan.TargetGroupID <= 0 || plan.GroupAccessMode != CafeRoomGroupAccessMode || plan.TargetGroupStatus != StatusActive {
		return ErrCafeGroupInvalid
	}
	return nil
}

func normalizedCafePlanPolicy(totalShares, seatCount int, tier string, maxBuyers, maxSharesPerUser, fulfillmentTimeoutMinutes int) (int, string, int, int, int) {
	if totalShares <= 0 {
		totalShares = seatCount
	}
	tier = strings.ToLower(strings.TrimSpace(tier))
	if tier == "" {
		tier = "plus"
	}
	if maxBuyers <= 0 && totalShares > 0 {
		maxBuyers = totalShares
		if maxBuyers > 4 {
			maxBuyers = 4
		}
	}
	if maxSharesPerUser <= 0 {
		maxSharesPerUser = totalShares
	}
	if fulfillmentTimeoutMinutes <= 0 {
		fulfillmentTimeoutMinutes = 1440
	}
	return totalShares, tier, maxBuyers, maxSharesPerUser, fulfillmentTimeoutMinutes
}

func normalizeCafeRoomStatus(status string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" {
		return CafeRoomStatusDraft, nil
	}
	switch normalized {
	case CafeRoomStatusDraft, CafeRoomStatusEnabled, CafeRoomStatusMaintenance, CafeRoomStatusDisabled:
		return normalized, nil
	default:
		return "", errors.BadRequest("CAFE_ROOM_INVALID_STATUS", fmt.Sprintf("unsupported room status %q", status))
	}
}

func containsCafeGroupID(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func formatCafeRoomNumber(value int) string {
	return fmt.Sprintf("%03d", value)
}
