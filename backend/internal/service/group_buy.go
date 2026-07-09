package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyentitlement"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyevent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyrefund"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	GroupBuyProductTokenPinPinPin = "token_pinpinpin"

	GroupBuyPlanStatusActive   = "active"
	GroupBuyPlanStatusDisabled = "disabled"

	GroupBuyLaunchModeAuto   = "auto"
	GroupBuyLaunchModeManual = "manual"

	GroupBuyRoundStatusOpen       = "open"
	GroupBuyRoundStatusActivating = "activating"
	GroupBuyRoundStatusActive     = "active"
	GroupBuyRoundStatusFailed     = "failed"
	GroupBuyRoundStatusCancelled  = "cancelled"

	GroupBuySeatStatusLocked           = "locked"
	GroupBuySeatStatusReleased         = "released"
	GroupBuySeatStatusPaid             = "paid"
	GroupBuySeatStatusActive           = "active"
	GroupBuySeatStatusRefundPending    = "refund_pending"
	GroupBuySeatStatusRefundProcessing = "refund_processing"
	GroupBuySeatStatusRefunded         = "refunded"
	GroupBuySeatStatusCancelled        = "cancelled"

	GroupBuyEntitlementStatusActive   = "active"
	GroupBuyEntitlementStatusInactive = "inactive"

	GroupBuyRefundModeBalanceCredit  = "balance_credit"
	GroupBuyRefundModeProviderRefund = "provider_refund"

	GroupBuyRefundStatusProcessing      = "processing"
	GroupBuyRefundStatusSucceeded       = "succeeded"
	GroupBuyRefundStatusPendingProvider = "pending_provider"
	GroupBuyRefundStatusFailed          = "failed"

	defaultGroupBuyTotalShares     = 10
	defaultGroupBuyMaxUserShares   = 10
	groupBuyMinShareCount          = 1
	groupBuyMaxShareCount          = 10
	groupBuySubscriptionNotePrefix = "token_pinpinpin entitlement"

	groupBuyEventSharesLocked     = "shares_locked"
	groupBuyEventSharesPaid       = "shares_paid"
	groupBuyEventRoundCreated     = "round_created"
	groupBuyEventRoundActivated   = "round_activated"
	groupBuyEventRoundFailed      = "round_failed"
	groupBuyEventSharesReleased   = "shares_released"
	groupBuyEventEntitlementSync  = "entitlement_synced"
	groupBuyEventEntitlementBound = "entitlement_bound_key"
	groupBuyEventRefundProcessed  = "refund_processed"
)

var (
	ErrGroupBuyPlanNotFound       = infraerrors.NotFound("GROUP_BUY_PLAN_NOT_FOUND", "group buy plan not found")
	ErrGroupBuyPlanUnavailable    = infraerrors.Forbidden("GROUP_BUY_PLAN_UNAVAILABLE", "group buy plan is unavailable")
	ErrGroupBuyTargetGroupInvalid = infraerrors.BadRequest("GROUP_BUY_TARGET_GROUP_INVALID", "target groups must be active subscription groups")
	ErrGroupBuyTierMappingInvalid = infraerrors.BadRequest("GROUP_BUY_TIER_MAPPING_INVALID", "tier rules must cover the configured share range without gaps or overlaps")
	ErrGroupBuyShareUnavailable   = infraerrors.Conflict("GROUP_BUY_SHARE_UNAVAILABLE", "not enough shares are available in this round")
	ErrGroupBuyShareLimitExceeded = infraerrors.Conflict("GROUP_BUY_SHARE_LIMIT_EXCEEDED", "share count exceeds user limit")
	ErrGroupBuyRoundUnavailable   = infraerrors.Conflict("GROUP_BUY_ROUND_UNAVAILABLE", "no open round is available")
	ErrGroupBuySeatNotFound       = infraerrors.NotFound("GROUP_BUY_SEAT_NOT_FOUND", "group buy share batch not found")
	ErrGroupBuySeatNotActive      = infraerrors.BadRequest("GROUP_BUY_SEAT_NOT_ACTIVE", "share batch is not active")
	ErrGroupBuyRoundNotFound      = infraerrors.NotFound("GROUP_BUY_ROUND_NOT_FOUND", "group buy round not found")
	ErrGroupBuyInvalidStatus      = infraerrors.BadRequest("GROUP_BUY_INVALID_STATUS", "invalid group buy status")
	ErrGroupBuyDisabled           = infraerrors.Forbidden("GROUP_BUY_DISABLED", "Token拼拼拼 is disabled")
)

type GroupBuyService struct {
	entClient            *dbent.Client
	paymentSvc           *PaymentService
	settingSvc           *SettingService
	subscriptionSvc      *SubscriptionService
	apiKeySvc            *APIKeyService
	userRepo             UserRepository
	groupRepo            GroupRepository
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	now                  func() time.Time
}

func NewGroupBuyService(
	entClient *dbent.Client,
	paymentSvc *PaymentService,
	settingSvc *SettingService,
	subscriptionSvc *SubscriptionService,
	apiKeySvc *APIKeyService,
	userRepo UserRepository,
	groupRepo GroupRepository,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *GroupBuyService {
	svc := &GroupBuyService{
		entClient:            entClient,
		paymentSvc:           paymentSvc,
		settingSvc:           settingSvc,
		subscriptionSvc:      subscriptionSvc,
		apiKeySvc:            apiKeySvc,
		userRepo:             userRepo,
		groupRepo:            groupRepo,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
		now:                  time.Now,
	}
	if paymentSvc != nil {
		paymentSvc.SetGroupBuyFulfillment(svc)
	}
	return svc
}

func (s *GroupBuyService) requireEnabled(ctx context.Context) error {
	if s == nil || s.settingSvc == nil {
		return nil
	}
	if !s.settingSvc.IsGroupBuyEnabled(ctx) {
		return ErrGroupBuyDisabled
	}
	return nil
}

type GroupBuyPlanInput struct {
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	ProductKey         string              `json:"product_key"`
	TotalShares        int                 `json:"total_shares"`
	SeatCount          int                 `json:"seat_count"`
	PricePerShare      float64             `json:"price_per_share"`
	PricePerSeat       float64             `json:"price_per_seat"`
	PriceLabel         string              `json:"price_label"`
	QuotaPerShareLabel string              `json:"quota_per_share_label"`
	QuotaLabel         string              `json:"quota_label"`
	MaxSharesPerUser   int                 `json:"max_shares_per_user"`
	TargetGroupID      int64               `json:"target_group_id"`
	TierGroupIDs       map[string]int64    `json:"tier_group_ids"`
	TierGroups         []GroupBuyTierInput `json:"tier_groups"`
	TierRules          []GroupBuyTierInput `json:"tier_rules"`
	ValidityDays       int                 `json:"validity_days"`
	TimeoutMinutes     int                 `json:"timeout_minutes"`
	LaunchMode         string              `json:"launch_mode"`
	RefundMode         string              `json:"refund_mode"`
	AgreementText      string              `json:"agreement_text"`
	Status             string              `json:"status"`
	SortOrder          int                 `json:"sort_order"`
}

type GroupBuyTierInput struct {
	ShareCount    int    `json:"share_count"`
	MinShares     int    `json:"min_shares"`
	MaxShares     int    `json:"max_shares"`
	TargetGroupID int64  `json:"target_group_id"`
	Label         string `json:"label"`
}

type GroupBuyCreateOrderInput struct {
	UserID            int64
	PlanID            int64
	ShareCount        int
	PaymentType       string
	OpenID            string
	WechatResumeToken string
	ClientIP          string
	IsMobile          bool
	IsWeChatBrowser   bool
	SrcHost           string
	SrcURL            string
	ReturnURL         string
	PaymentSource     string
}

type GroupBuyPlanView struct {
	ID                 int64              `json:"id"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	ProductKey         string             `json:"product_key"`
	TotalShares        int                `json:"total_shares"`
	SeatCount          int                `json:"seat_count"`
	PricePerShare      float64            `json:"price_per_share"`
	PricePerSeat       float64            `json:"price_per_seat"`
	PriceLabel         string             `json:"price_label"`
	QuotaPerShareLabel string             `json:"quota_per_share_label"`
	QuotaLabel         string             `json:"quota_label"`
	MaxSharesPerUser   int                `json:"max_shares_per_user"`
	TargetGroupID      int64              `json:"target_group_id"`
	TargetGroup        *GroupBuyGroupView `json:"target_group,omitempty"`
	TierGroupIDs       map[string]int64   `json:"tier_group_ids"`
	TierGroups         []GroupBuyTierView `json:"tier_groups"`
	TierRules          []GroupBuyTierView `json:"tier_rules"`
	ValidityDays       int                `json:"validity_days"`
	TimeoutMinutes     int                `json:"timeout_minutes"`
	LaunchMode         string             `json:"launch_mode"`
	RefundMode         string             `json:"refund_mode"`
	AgreementText      string             `json:"agreement_text"`
	Status             string             `json:"status"`
	SortOrder          int                `json:"sort_order"`
	CurrentRound       *GroupBuyRoundView `json:"current_round,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type GroupBuyTierView struct {
	ShareCount    int                `json:"share_count"`
	MinShares     int                `json:"min_shares"`
	MaxShares     int                `json:"max_shares"`
	TargetGroupID int64              `json:"target_group_id"`
	Label         string             `json:"label"`
	TargetGroup   *GroupBuyGroupView `json:"target_group,omitempty"`
}

type GroupBuyGroupView struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Platform        string   `json:"platform"`
	DailyLimitUSD   *float64 `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUSD  *float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd,omitempty"`
}

type GroupBuyRoundView struct {
	ID              int64      `json:"id"`
	PlanID          int64      `json:"plan_id"`
	Status          string     `json:"status"`
	TotalShares     int        `json:"total_shares"`
	PaidShares      int        `json:"paid_shares"`
	ReservedShares  int        `json:"reserved_shares"`
	AvailableShares int        `json:"available_shares"`
	TotalSeats      int        `json:"total_seats"`
	PaidSeats       int        `json:"paid_seats"`
	ReservedSeats   int        `json:"reserved_seats"`
	AvailableSeats  int        `json:"available_seats"`
	DeadlineAt      time.Time  `json:"deadline_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	CloseReason     *string    `json:"close_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type GroupBuySeatView struct {
	ID                int64              `json:"id"`
	RoundID           int64              `json:"round_id"`
	PlanID            int64              `json:"plan_id"`
	UserID            int64              `json:"user_id"`
	OrderID           *int64             `json:"order_id,omitempty"`
	Status            string             `json:"status"`
	ShareCount        int                `json:"share_count"`
	SubscriptionID    *int64             `json:"subscription_id,omitempty"`
	BoundAPIKeyID     *int64             `json:"bound_api_key_id,omitempty"`
	LockedUntil       *time.Time         `json:"locked_until,omitempty"`
	PaidAt            *time.Time         `json:"paid_at,omitempty"`
	ActivatedAt       *time.Time         `json:"activated_at,omitempty"`
	ExpiresAt         *time.Time         `json:"expires_at,omitempty"`
	BoundAt           *time.Time         `json:"bound_at,omitempty"`
	RefundProcessedAt *time.Time         `json:"refund_processed_at,omitempty"`
	RefundNote        *string            `json:"refund_note,omitempty"`
	Plan              *GroupBuyPlanView  `json:"plan,omitempty"`
	Round             *GroupBuyRoundView `json:"round,omitempty"`
	Order             *PaymentOrderLite  `json:"order,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

type GroupBuyEntitlementView struct {
	ID               int64              `json:"id"`
	UserID           int64              `json:"user_id"`
	ProductKey       string             `json:"product_key"`
	Status           string             `json:"status"`
	ActiveShareCount int                `json:"active_share_count"`
	TargetGroupID    *int64             `json:"target_group_id,omitempty"`
	TargetGroup      *GroupBuyGroupView `json:"target_group,omitempty"`
	SubscriptionID   *int64             `json:"subscription_id,omitempty"`
	BoundAPIKeyID    *int64             `json:"bound_api_key_id,omitempty"`
	EntitlementLabel string             `json:"entitlement_label,omitempty"`
	LastActivatedAt  *time.Time         `json:"last_activated_at,omitempty"`
	ExpiresAt        *time.Time         `json:"expires_at,omitempty"`
	RefreshedAt      *time.Time         `json:"refreshed_at,omitempty"`
	DeactivatedAt    *time.Time         `json:"deactivated_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type GroupBuyMySeatsView struct {
	Entitlement *GroupBuyEntitlementView `json:"entitlement,omitempty"`
	Seats       []GroupBuySeatView       `json:"seats"`
}

type PaymentOrderLite struct {
	ID          int64      `json:"id"`
	Amount      float64    `json:"amount"`
	PayAmount   float64    `json:"pay_amount"`
	Currency    string     `json:"currency"`
	PaymentType string     `json:"payment_type"`
	OutTradeNo  string     `json:"out_trade_no"`
	Status      string     `json:"status"`
	OrderType   string     `json:"order_type"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type GroupBuyEventView struct {
	ID        int64          `json:"id"`
	PlanID    *int64         `json:"plan_id,omitempty"`
	RoundID   *int64         `json:"round_id,omitempty"`
	SeatID    *int64         `json:"seat_id,omitempty"`
	UserID    *int64         `json:"user_id,omitempty"`
	EventType string         `json:"event_type"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type GroupBuyCreateOrderResponse struct {
	*CreateOrderResponse
	Seat  *GroupBuySeatView  `json:"seat,omitempty"`
	Round *GroupBuyRoundView `json:"round,omitempty"`
}

func (s *GroupBuyService) ListPlans(ctx context.Context, includeDisabled bool) ([]GroupBuyPlanView, error) {
	if !includeDisabled {
		if err := s.requireEnabled(ctx); err != nil {
			return nil, err
		}
	}
	q := s.entClient.GroupBuyPlan.Query().
		Where(groupbuyplan.DeletedAtIsNil(), groupbuyplan.ProductKeyEQ(GroupBuyProductTokenPinPinPin)).
		Order(dbent.Asc(groupbuyplan.FieldSortOrder), dbent.Asc(groupbuyplan.FieldID)).
		WithTargetGroup().
		WithRounds(func(rq *dbent.GroupBuyRoundQuery) {
			rq.Where(groupbuyround.StatusEQ(GroupBuyRoundStatusOpen)).
				Order(dbent.Asc(groupbuyround.FieldDeadlineAt)).
				Limit(1)
		})
	if !includeDisabled {
		q = q.Where(groupbuyplan.StatusEQ(GroupBuyPlanStatusActive))
	}
	plans, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list group buy plans: %w", err)
	}
	out := make([]GroupBuyPlanView, 0, len(plans))
	for _, p := range plans {
		out = append(out, s.planView(ctx, p))
	}
	return out, nil
}

func (s *GroupBuyService) Activity(ctx context.Context, limit int) ([]GroupBuyEventView, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	events, err := s.entClient.GroupBuyEvent.Query().
		Order(dbent.Desc(groupbuyevent.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list group buy activity: %w", err)
	}
	out := make([]GroupBuyEventView, 0, len(events))
	for _, ev := range events {
		out = append(out, eventView(ev))
	}
	return out, nil
}

func (s *GroupBuyService) CreateOrder(ctx context.Context, input GroupBuyCreateOrderInput) (*GroupBuyCreateOrderResponse, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if input.UserID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	if input.PlanID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "plan_id is required")
	}
	if input.ShareCount <= 0 {
		input.ShareCount = 1
	}
	if input.ShareCount > groupBuyMaxShareCount {
		return nil, ErrGroupBuyShareLimitExceeded
	}
	req := CreateOrderRequest{
		UserID:          input.UserID,
		PaymentType:     input.PaymentType,
		OpenID:          input.OpenID,
		ClientIP:        input.ClientIP,
		IsMobile:        input.IsMobile,
		IsWeChatBrowser: input.IsWeChatBrowser,
		SrcHost:         input.SrcHost,
		SrcURL:          input.SrcURL,
		ReturnURL:       input.ReturnURL,
		PaymentSource:   input.PaymentSource,
		OrderType:       payment.OrderTypeGroupBuy,
		PlanID:          input.PlanID,
	}
	if normalized := NormalizeVisibleMethod(req.PaymentType); normalized != "" {
		req.PaymentType = normalized
	}

	cfg, err := s.paymentSvc.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}
	if err := s.paymentSvc.checkCancelRateLimit(ctx, input.UserID, cfg); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != StatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}
	plan, err := s.loadAvailablePlan(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}
	orderAmount := plan.PricePerShare * float64(input.ShareCount)
	req.Amount = orderAmount
	feeRate := cfg.RechargeFeeRate
	methodCurrency := payment.DefaultPaymentCurrency
	if s.paymentSvc.configService != nil {
		methodCurrency, err = s.paymentSvc.configService.ValidateMethodCurrencyConsistency(ctx, req.PaymentType)
		if err != nil {
			return nil, err
		}
	}
	payAmountStr, payAmount, err := calculateCreateOrderPayAmount(orderAmount, feeRate, methodCurrency)
	if err != nil {
		return nil, err
	}
	sel, err := s.paymentSvc.selectCreateOrderInstance(ctx, req, cfg, payAmount)
	if err != nil {
		return nil, err
	}
	if err := s.paymentSvc.validateSelectedCreateOrderInstance(ctx, req, sel); err != nil {
		return nil, err
	}
	selectedCurrency := payment.DefaultPaymentCurrency
	if sel != nil {
		selectedCurrency = paymentProviderConfigCurrency(sel.ProviderKey, sel.Config)
	}
	if selectedCurrency != methodCurrency {
		payAmountStr, payAmount, err = calculateCreateOrderPayAmount(orderAmount, feeRate, selectedCurrency)
		if err != nil {
			return nil, err
		}
	}
	if err := validateSelectedCreateOrderAmountCurrency(payAmountStr, sel); err != nil {
		return nil, err
	}
	oauthResp, err := s.paymentSvc.maybeBuildWeChatOAuthRequiredResponseForSelection(ctx, req, orderAmount, payAmount, feeRate, sel)
	if err != nil {
		return nil, err
	}
	if oauthResp != nil {
		return &GroupBuyCreateOrderResponse{CreateOrderResponse: oauthResp}, nil
	}

	order, seat, round, err := s.lockSharesAndCreateOrder(ctx, req, user, plan, input.ShareCount, cfg, feeRate, payAmount, sel)
	if err != nil {
		return nil, err
	}
	resp, err := s.paymentSvc.invokeProvider(ctx, order, req, cfg, orderAmount, payAmountStr, payAmount, nil, sel)
	if err != nil {
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).Save(ctx)
		_ = s.ReleaseGroupBuySeatForOrder(context.Background(), order.ID, "provider create failed")
		return nil, err
	}
	return &GroupBuyCreateOrderResponse{
		CreateOrderResponse: resp,
		Seat:                s.seatView(ctx, seat),
		Round:               roundView(round),
	}, nil
}

func (s *GroupBuyService) lockSharesAndCreateOrder(ctx context.Context, req CreateOrderRequest, user *User, plan *dbent.GroupBuyPlan, shareCount int, cfg *PaymentConfig, feeRate, payAmount float64, sel *payment.InstanceSelection) (*dbent.PaymentOrder, *dbent.GroupBuySeat, *dbent.GroupBuyRound, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	planQuery := tx.GroupBuyPlan.Query().
		Where(groupbuyplan.IDEQ(plan.ID), groupbuyplan.DeletedAtIsNil())
	lockedPlan, err := s.groupBuyPlanForUpdate(planQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, nil, ErrGroupBuyPlanNotFound
		}
		return nil, nil, nil, fmt.Errorf("lock group buy plan: %w", err)
	}
	if lockedPlan.Status != GroupBuyPlanStatusActive {
		return nil, nil, nil, ErrGroupBuyPlanUnavailable
	}
	if err := s.validatePlanTierRulesInTx(txCtx, tx, lockedPlan); err != nil {
		return nil, nil, nil, err
	}
	if err := s.paymentSvc.checkPendingLimit(txCtx, tx, req.UserID, cfg.MaxPendingOrders); err != nil {
		return nil, nil, nil, err
	}
	amount := lockedPlan.PricePerShare * float64(shareCount)
	if err := s.paymentSvc.checkDailyLimit(txCtx, tx, req.UserID, amount, cfg.DailyLimit); err != nil {
		return nil, nil, nil, err
	}
	round, err := s.findOpenRoundTx(txCtx, tx, lockedPlan, true)
	if err != nil {
		return nil, nil, nil, err
	}
	if round == nil {
		return nil, nil, nil, ErrGroupBuyRoundUnavailable
	}
	if round.TotalShares-round.PaidShares-round.ReservedShares < shareCount {
		return nil, nil, nil, ErrGroupBuyShareUnavailable
	}
	heldShares, err := s.countHeldSharesForUserTx(txCtx, tx, req.UserID)
	if err != nil {
		return nil, nil, nil, err
	}
	if heldShares+shareCount > lockedPlan.MaxSharesPerUser {
		return nil, nil, nil, ErrGroupBuyShareLimitExceeded.WithMetadata(map[string]string{
			"held_shares": strconv.Itoa(heldShares),
			"max_shares":  strconv.Itoa(lockedPlan.MaxSharesPerUser),
		})
	}
	outTradeNo, err := s.paymentSvc.allocateOutTradeNo(txCtx, tx)
	if err != nil {
		return nil, nil, nil, err
	}
	timeoutMin := cfg.OrderTimeoutMin
	if timeoutMin <= 0 {
		timeoutMin = defaultOrderTimeoutMin
	}
	expiresAt := s.now().Add(time.Duration(timeoutMin) * time.Minute)
	providerSnapshot := buildPaymentOrderProviderSnapshot(sel, req)
	selectedInstanceID := ""
	selectedProviderKey := ""
	if sel != nil {
		selectedInstanceID = strings.TrimSpace(sel.InstanceID)
		selectedProviderKey = strings.TrimSpace(sel.ProviderKey)
	}
	orderBuilder := tx.PaymentOrder.Create().
		SetUserID(req.UserID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetNillableUserNotes(psNilIfEmpty(user.Notes)).
		SetAmount(amount).
		SetPayAmount(payAmount).
		SetFeeRate(feeRate).
		SetRechargeCode("").
		SetOutTradeNo(outTradeNo).
		SetPaymentType(req.PaymentType).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetPlanID(lockedPlan.ID).
		SetSubscriptionDays(lockedPlan.ValidityDays).
		SetStatus(OrderStatusPending).
		SetExpiresAt(expiresAt).
		SetClientIP(req.ClientIP).
		SetSrcHost(req.SrcHost)
	if req.SrcURL != "" {
		orderBuilder.SetSrcURL(req.SrcURL)
	}
	if selectedInstanceID != "" {
		orderBuilder.SetProviderInstanceID(selectedInstanceID)
	}
	if selectedProviderKey != "" {
		orderBuilder.SetProviderKey(selectedProviderKey)
	}
	if providerSnapshot != nil {
		orderBuilder.SetProviderSnapshot(providerSnapshot)
	}
	order, err := orderBuilder.Save(txCtx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create group buy payment order: %w", err)
	}
	code := fmt.Sprintf("GB-%d-%d", order.ID, s.now().UnixNano()%100000)
	order, err = tx.PaymentOrder.UpdateOneID(order.ID).SetRechargeCode(code).Save(txCtx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("set group buy payment code: %w", err)
	}
	policySnapshot := buildGroupBuyPolicySnapshot(lockedPlan, s.now())
	seat, err := tx.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(lockedPlan.ID).
		SetUserID(req.UserID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusLocked).
		SetShareCount(shareCount).
		SetPolicySnapshot(policySnapshot).
		SetLockedUntil(expiresAt).
		Save(txCtx)
	if err != nil {
		return nil, nil, nil, translateGroupBuySeatCreateError(err)
	}
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		AddReservedShares(shareCount).
		AddReservedSeats(shareCount).
		SetUpdatedAt(s.now()).
		Save(txCtx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reserve group buy shares: %w", err)
	}
	s.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{
		PlanID:    &lockedPlan.ID,
		RoundID:   &round.ID,
		SeatID:    &seat.ID,
		UserID:    &req.UserID,
		EventType: groupBuyEventSharesLocked,
		Message:   "用户锁定 Token拼拼拼 份额",
		Metadata: map[string]any{
			"order_id":    order.ID,
			"amount":      order.Amount,
			"share_count": shareCount,
		},
	})
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, fmt.Errorf("commit group buy order transaction: %w", err)
	}
	return order, seat, round, nil
}

func (s *GroupBuyService) findOpenRoundTx(ctx context.Context, tx *dbent.Tx, plan *dbent.GroupBuyPlan, createIfAuto bool) (*dbent.GroupBuyRound, error) {
	now := s.now()
	roundQuery := tx.GroupBuyRound.Query().
		Where(
			groupbuyround.PlanIDEQ(plan.ID),
			groupbuyround.StatusEQ(GroupBuyRoundStatusOpen),
			groupbuyround.DeadlineAtGT(now),
		).
		Order(dbent.Asc(groupbuyround.FieldCreatedAt))
	rounds, err := s.groupBuyRoundForUpdate(roundQuery).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query open group buy rounds: %w", err)
	}
	for _, round := range rounds {
		if _, err := s.releaseExpiredLockedSeatsForRoundTx(ctx, tx, round.ID); err != nil {
			return nil, err
		}
		reloadQuery := tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(round.ID))
		reloaded, err := s.groupBuyRoundForUpdate(reloadQuery).Only(ctx)
		if err != nil {
			return nil, err
		}
		if reloaded.PaidShares+reloaded.ReservedShares < reloaded.TotalShares {
			return reloaded, nil
		}
	}
	if !createIfAuto || plan.LaunchMode == GroupBuyLaunchModeManual {
		return nil, nil
	}
	return s.createRoundForPlanTx(ctx, tx, plan)
}

func (s *GroupBuyService) createRoundForPlanTx(ctx context.Context, tx *dbent.Tx, plan *dbent.GroupBuyPlan) (*dbent.GroupBuyRound, error) {
	now := s.now()
	deadline := now.Add(time.Duration(plan.TimeoutMinutes) * time.Minute)
	round, err := tx.GroupBuyRound.Create().
		SetPlanID(plan.ID).
		SetStatus(GroupBuyRoundStatusOpen).
		SetTotalShares(plan.TotalShares).
		SetPaidShares(0).
		SetReservedShares(0).
		SetTotalSeats(plan.TotalShares).
		SetPaidSeats(0).
		SetReservedSeats(0).
		SetDeadlineAt(deadline).
		SetStartedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create group buy round: %w", err)
	}
	_, _ = tx.GroupBuyPlan.UpdateOneID(plan.ID).SetLastRoundCreatedAt(now).Save(ctx)
	s.createEventTx(ctx, tx.Client(), &groupBuyEventInput{
		PlanID:    &plan.ID,
		RoundID:   &round.ID,
		EventType: groupBuyEventRoundCreated,
		Message:   "Token拼拼拼 新车已开放",
		Metadata:  map[string]any{"total_shares": plan.TotalShares, "launch_mode": plan.LaunchMode},
	})
	return round, nil
}

func (s *GroupBuyService) countHeldSharesForUserTx(ctx context.Context, tx *dbent.Tx, userID int64) (int, error) {
	seats, err := tx.GroupBuySeat.Query().
		Where(
			groupbuyseat.UserIDEQ(userID),
			groupbuyseat.StatusIn(GroupBuySeatStatusLocked, GroupBuySeatStatusPaid, GroupBuySeatStatusActive, GroupBuySeatStatusRefundPending),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("count held group buy shares: %w", err)
	}
	total := 0
	now := s.now()
	for _, seat := range seats {
		if seat.Status == GroupBuySeatStatusLocked && seat.LockedUntil != nil && !seat.LockedUntil.After(now) {
			continue
		}
		if seat.Status == GroupBuySeatStatusActive && seat.ExpiresAt != nil && !seat.ExpiresAt.After(now) {
			continue
		}
		total += seat.ShareCount
	}
	return total, nil
}

func (s *GroupBuyService) releaseExpiredLockedSeatsForRoundTx(ctx context.Context, tx *dbent.Tx, roundID int64) (int, error) {
	now := s.now()
	seatQuery := tx.GroupBuySeat.Query().
		Where(
			groupbuyseat.RoundIDEQ(roundID),
			groupbuyseat.StatusEQ(GroupBuySeatStatusLocked),
			groupbuyseat.LockedUntilLTE(now),
		).
		WithOrder()
	seats, err := s.groupBuySeatForUpdate(seatQuery).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query expired group buy seats: %w", err)
	}
	releasedShares := 0
	for _, seat := range seats {
		order := seat.Edges.Order
		if order != nil && order.Status == OrderStatusPending {
			if _, err := tx.PaymentOrder.Update().
				Where(paymentorder.IDEQ(order.ID), paymentorder.StatusEQ(OrderStatusPending)).
				SetStatus(OrderStatusExpired).
				SetFailedAt(now).
				SetFailedReason("group buy share lock expired").
				Save(ctx); err != nil {
				return releasedShares, fmt.Errorf("expire group buy payment order: %w", err)
			}
		}
		if err := tx.GroupBuySeat.UpdateOneID(seat.ID).
			SetStatus(GroupBuySeatStatusReleased).
			SetUpdatedAt(now).
			Exec(ctx); err != nil {
			return releasedShares, fmt.Errorf("release expired group buy shares: %w", err)
		}
		releasedShares += seat.ShareCount
		s.createEventTx(ctx, tx.Client(), &groupBuyEventInput{
			PlanID:    &seat.PlanID,
			RoundID:   &seat.RoundID,
			SeatID:    &seat.ID,
			UserID:    &seat.UserID,
			EventType: groupBuyEventSharesReleased,
			Message:   "未支付份额已释放",
			Metadata:  map[string]any{"share_count": seat.ShareCount},
		})
	}
	if releasedShares > 0 {
		if _, err := tx.GroupBuyRound.UpdateOneID(roundID).AddReservedShares(-releasedShares).SetUpdatedAt(now).Save(ctx); err != nil {
			return releasedShares, fmt.Errorf("decrement released group buy shares: %w", err)
		}
		if _, err := tx.GroupBuyRound.UpdateOneID(roundID).AddReservedSeats(-releasedShares).SetUpdatedAt(now).Save(ctx); err != nil {
			return releasedShares, fmt.Errorf("decrement released legacy group buy seats: %w", err)
		}
	}
	return releasedShares, nil
}

func (s *GroupBuyService) HandleGroupBuyOrderPaid(ctx context.Context, orderID int64) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin group buy paid tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	seatQuery := tx.GroupBuySeat.Query().
		Where(groupbuyseat.OrderIDEQ(orderID))
	seat, err := s.groupBuySeatForUpdate(seatQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return ErrGroupBuySeatNotFound
		}
		return fmt.Errorf("load group buy share batch by order: %w", err)
	}
	if seat.Status == GroupBuySeatStatusActive || seat.Status == GroupBuySeatStatusPaid {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit idempotent group buy paid tx: %w", err)
		}
		return s.TryActivateRound(ctx, seat.RoundID)
	}
	if seat.Status != GroupBuySeatStatusLocked {
		return ErrGroupBuyInvalidStatus.WithMetadata(map[string]string{"seat_status": seat.Status})
	}
	roundQuery := tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(seat.RoundID))
	round, err := s.groupBuyRoundForUpdate(roundQuery).Only(txCtx)
	if err != nil {
		return fmt.Errorf("lock group buy round: %w", err)
	}
	now := s.now()
	if err := tx.GroupBuySeat.UpdateOneID(seat.ID).
		SetStatus(GroupBuySeatStatusPaid).
		SetPaidAt(now).
		SetUpdatedAt(now).
		Exec(txCtx); err != nil {
		return fmt.Errorf("mark group buy share batch paid: %w", err)
	}
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		AddPaidShares(seat.ShareCount).
		AddReservedShares(-seat.ShareCount).
		AddPaidSeats(seat.ShareCount).
		AddReservedSeats(-seat.ShareCount).
		SetUpdatedAt(now).
		Save(txCtx)
	if err != nil {
		return fmt.Errorf("update group buy round paid shares: %w", err)
	}
	s.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{
		PlanID:    &seat.PlanID,
		RoundID:   &seat.RoundID,
		SeatID:    &seat.ID,
		UserID:    &seat.UserID,
		EventType: groupBuyEventSharesPaid,
		Message:   "份额付款成功，等待满份成团",
		Metadata:  map[string]any{"order_id": orderID, "share_count": seat.ShareCount},
	})
	shouldActivate := round.PaidShares >= round.TotalShares
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit group buy paid tx: %w", err)
	}
	if shouldActivate {
		return s.TryActivateRound(ctx, round.ID)
	}
	return nil
}

func (s *GroupBuyService) TryActivateRound(ctx context.Context, roundID int64) error {
	if roundID <= 0 {
		return ErrGroupBuyRoundNotFound
	}
	round, plan, seats, claimed, err := s.claimRoundActivation(ctx, roundID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	now := s.now()
	expiresAt := now.AddDate(0, 0, plan.ValidityDays)
	userIDs := make(map[int64]struct{})
	activatedSeatIDs := make([]int64, 0, len(seats))
	for _, seat := range seats {
		userIDs[seat.UserID] = struct{}{}
		if seat.Status == GroupBuySeatStatusActive {
			activatedSeatIDs = append(activatedSeatIDs, seat.ID)
			continue
		}
		if _, err := s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
			SetStatus(GroupBuySeatStatusActive).
			SetActivatedAt(now).
			SetExpiresAt(expiresAt).
			SetUpdatedAt(now).
			Save(ctx); err != nil {
			return s.failActivation(ctx, round.ID, fmt.Errorf("mark group buy share batch active: %w", err))
		}
		activatedSeatIDs = append(activatedSeatIDs, seat.ID)
	}
	for userID := range userIDs {
		if _, err := s.RefreshUserEntitlement(ctx, userID); err != nil {
			return s.failActivation(ctx, round.ID, fmt.Errorf("refresh group buy entitlement: %w", err))
		}
	}
	updated, err := s.entClient.GroupBuyRound.Update().
		Where(groupbuyround.IDEQ(round.ID), groupbuyround.StatusEQ(GroupBuyRoundStatusActivating)).
		SetStatus(GroupBuyRoundStatusActive).
		SetClosedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark group buy round active: %w", err)
	}
	if updated > 0 {
		planID := plan.ID
		s.createEvent(ctx, &groupBuyEventInput{
			PlanID:    &planID,
			RoundID:   &round.ID,
			EventType: groupBuyEventRoundActivated,
			Message:   "Token拼拼拼 已满份成团，份额权益已开通",
			Metadata: map[string]any{
				"seat_ids":     activatedSeatIDs,
				"total_shares": round.TotalShares,
			},
		})
	}
	if plan.LaunchMode == GroupBuyLaunchModeAuto {
		if err := s.ensureAutoNextRound(ctx, plan.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupBuyService) claimRoundActivation(ctx context.Context, roundID int64) (*dbent.GroupBuyRound, *dbent.GroupBuyPlan, []*dbent.GroupBuySeat, bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, nil, nil, false, fmt.Errorf("begin group buy activation tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	roundQuery := tx.GroupBuyRound.Query().
		Where(groupbuyround.IDEQ(roundID))
	round, err := s.groupBuyRoundForUpdate(roundQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, nil, false, ErrGroupBuyRoundNotFound
		}
		return nil, nil, nil, false, fmt.Errorf("lock group buy round: %w", err)
	}
	if round.Status == GroupBuyRoundStatusActive || round.Status == GroupBuyRoundStatusFailed || round.Status == GroupBuyRoundStatusCancelled {
		if err := tx.Commit(); err != nil {
			return nil, nil, nil, false, err
		}
		return round, nil, nil, false, nil
	}
	if round.PaidShares < round.TotalShares {
		if err := tx.Commit(); err != nil {
			return nil, nil, nil, false, err
		}
		return round, nil, nil, false, nil
	}
	plan, err := tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(round.PlanID)).Only(txCtx)
	if err != nil {
		return nil, nil, nil, false, fmt.Errorf("load group buy plan: %w", err)
	}
	if err := s.validatePlanTierRulesInTx(txCtx, tx, plan); err != nil {
		return nil, nil, nil, false, err
	}
	seatQuery := tx.GroupBuySeat.Query().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusIn(GroupBuySeatStatusPaid, GroupBuySeatStatusActive))
	seats, err := s.groupBuySeatForUpdate(seatQuery).All(txCtx)
	if err != nil {
		return nil, nil, nil, false, fmt.Errorf("load paid group buy share batches: %w", err)
	}
	paidShares := 0
	for _, seat := range seats {
		paidShares += seat.ShareCount
	}
	if paidShares < round.TotalShares {
		if err := tx.Commit(); err != nil {
			return nil, nil, nil, false, err
		}
		return round, plan, seats, false, nil
	}
	if round.Status != GroupBuyRoundStatusActivating {
		round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
			SetStatus(GroupBuyRoundStatusActivating).
			SetUpdatedAt(s.now()).
			Save(txCtx)
		if err != nil {
			return nil, nil, nil, false, fmt.Errorf("claim group buy activation: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, false, fmt.Errorf("commit group buy activation claim: %w", err)
	}
	return round, plan, seats, true, nil
}

func (s *GroupBuyService) ensureAutoNextRound(ctx context.Context, planID int64) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin group buy auto round tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	planQuery := tx.GroupBuyPlan.Query().
		Where(groupbuyplan.IDEQ(planID), groupbuyplan.DeletedAtIsNil())
	plan, err := s.groupBuyPlanForUpdate(planQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			_ = tx.Commit()
			return nil
		}
		return err
	}
	if plan.Status != GroupBuyPlanStatusActive || plan.LaunchMode != GroupBuyLaunchModeAuto {
		_ = tx.Commit()
		return nil
	}
	existingQuery := tx.GroupBuyRound.Query().
		Where(groupbuyround.PlanIDEQ(plan.ID), groupbuyround.StatusEQ(GroupBuyRoundStatusOpen), groupbuyround.DeadlineAtGT(s.now())).
		Limit(1)
	existing, err := s.groupBuyRoundForUpdate(existingQuery).Exist(txCtx)
	if err != nil {
		return err
	}
	if !existing {
		if _, err := s.createRoundForPlanTx(txCtx, tx, plan); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *GroupBuyService) failActivation(ctx context.Context, roundID int64, cause error) error {
	if cause == nil {
		return nil
	}
	slog.Warn("group buy activation failed", "roundID", roundID, "error", cause)
	_, _ = s.entClient.GroupBuyEvent.Create().
		SetRoundID(roundID).
		SetEventType("activation_failed").
		SetMessage(cause.Error()).
		SetMetadata(map[string]any{"error": cause.Error()}).
		Save(ctx)
	return cause
}

func (s *GroupBuyService) ReleaseGroupBuySeatForOrder(ctx context.Context, orderID int64, reason string) error {
	if orderID <= 0 {
		return nil
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin group buy release tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	seatQuery := tx.GroupBuySeat.Query().
		Where(groupbuyseat.OrderIDEQ(orderID))
	seat, err := s.groupBuySeatForUpdate(seatQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			_ = tx.Commit()
			return nil
		}
		return fmt.Errorf("load group buy share batch for release: %w", err)
	}
	if seat.Status != GroupBuySeatStatusLocked {
		_ = tx.Commit()
		return nil
	}
	now := s.now()
	if err := tx.GroupBuySeat.UpdateOneID(seat.ID).
		SetStatus(GroupBuySeatStatusReleased).
		SetUpdatedAt(now).
		Exec(txCtx); err != nil {
		return fmt.Errorf("release group buy shares: %w", err)
	}
	if _, err := tx.GroupBuyRound.UpdateOneID(seat.RoundID).
		AddReservedShares(-seat.ShareCount).
		AddReservedSeats(-seat.ShareCount).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return fmt.Errorf("update group buy round release count: %w", err)
	}
	s.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{
		PlanID:    &seat.PlanID,
		RoundID:   &seat.RoundID,
		SeatID:    &seat.ID,
		UserID:    &seat.UserID,
		EventType: groupBuyEventSharesReleased,
		Message:   "未支付份额已释放",
		Metadata:  map[string]any{"order_id": orderID, "reason": reason, "share_count": seat.ShareCount},
	})
	return tx.Commit()
}

func (s *GroupBuyService) ExpireRounds(ctx context.Context) (int, error) {
	now := s.now()
	rounds, err := s.entClient.GroupBuyRound.Query().
		Where(groupbuyround.StatusEQ(GroupBuyRoundStatusOpen), groupbuyround.DeadlineAtLTE(now)).
		Order(dbent.Asc(groupbuyround.FieldDeadlineAt)).
		Limit(50).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query timed out group buy rounds: %w", err)
	}
	closed := 0
	for _, round := range rounds {
		ok, err := s.failTimedOutRound(ctx, round.ID)
		if err != nil {
			return closed, err
		}
		if ok {
			closed++
		}
	}
	return closed, nil
}

func (s *GroupBuyService) failTimedOutRound(ctx context.Context, roundID int64) (bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin group buy fail tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	roundQuery := tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))
	round, err := s.groupBuyRoundForUpdate(roundQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, ErrGroupBuyRoundNotFound
		}
		return false, err
	}
	if round.Status != GroupBuyRoundStatusOpen || round.PaidShares >= round.TotalShares || round.DeadlineAt.After(s.now()) {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		if round.PaidShares >= round.TotalShares {
			return false, s.TryActivateRound(ctx, round.ID)
		}
		return false, nil
	}
	now := s.now()
	reason := "round deadline reached before full shares"
	round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusFailed).
		SetClosedAt(now).
		SetCloseReason(reason).
		SetUpdatedAt(now).
		Save(txCtx)
	if err != nil {
		return false, fmt.Errorf("mark group buy round failed: %w", err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusPaid)).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("mark group buy shares refund pending: %w", err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked)).
		SetStatus(GroupBuySeatStatusReleased).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return false, fmt.Errorf("release locked shares in failed round: %w", err)
	}
	s.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{
		RoundID:   &round.ID,
		PlanID:    &round.PlanID,
		EventType: groupBuyEventRoundFailed,
		Message:   "Token拼拼拼 未在截止时间前满份，已进入退款处理",
		Metadata:  map[string]any{"reason": reason},
	})
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit group buy fail tx: %w", err)
	}
	return true, nil
}

func (s *GroupBuyService) RefreshExpiredEntitlements(ctx context.Context) (int, error) {
	now := s.now()
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(
			groupbuyseat.StatusEQ(GroupBuySeatStatusActive),
			groupbuyseat.ExpiresAtNotNil(),
			groupbuyseat.ExpiresAtLTE(now),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query expired group buy entitlement users: %w", err)
	}
	seen := make(map[int64]struct{}, len(seats))
	userIDs := make([]int64, 0, len(seats))
	for _, seat := range seats {
		if _, ok := seen[seat.UserID]; ok {
			continue
		}
		seen[seat.UserID] = struct{}{}
		userIDs = append(userIDs, seat.UserID)
	}
	for _, userID := range userIDs {
		if _, err := s.RefreshUserEntitlement(ctx, userID); err != nil {
			return 0, err
		}
	}
	return len(userIDs), nil
}

func (s *GroupBuyService) RefreshUserEntitlement(ctx context.Context, userID int64) (*GroupBuyEntitlementView, error) {
	if userID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	now := s.now()
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(
			groupbuyseat.UserIDEQ(userID),
			groupbuyseat.StatusEQ(GroupBuySeatStatusActive),
			groupbuyseat.Or(groupbuyseat.ExpiresAtIsNil(), groupbuyseat.ExpiresAtGT(now)),
		).
		WithPlan().
		Order(dbent.Desc(groupbuyseat.FieldActivatedAt), dbent.Desc(groupbuyseat.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load active group buy shares: %w", err)
	}
	totalShares := 0
	var maxExpires *time.Time
	var lastActivated *time.Time
	for _, seat := range seats {
		if seat.Edges.Plan == nil || seat.Edges.Plan.ProductKey != GroupBuyProductTokenPinPinPin {
			continue
		}
		totalShares += seat.ShareCount
		if seat.ExpiresAt != nil && (maxExpires == nil || seat.ExpiresAt.After(*maxExpires)) {
			t := *seat.ExpiresAt
			maxExpires = &t
		}
		if seat.ActivatedAt != nil && (lastActivated == nil || seat.ActivatedAt.After(*lastActivated)) {
			t := *seat.ActivatedAt
			lastActivated = &t
		}
	}
	if totalShares > groupBuyMaxShareCount {
		totalShares = groupBuyMaxShareCount
	}
	policyPlan, _ := s.latestProductPlan(ctx)

	ent, err := s.getOrCreateEntitlement(ctx, userID)
	if err != nil {
		return nil, err
	}
	if totalShares <= 0 || policyPlan == nil {
		ent, err = s.deactivateEntitlement(ctx, ent)
		if err != nil {
			return nil, err
		}
		return s.entitlementView(ctx, ent), nil
	}
	tierRule, ok := resolveTierRuleForShareCount(normalizedPlanTierRules(policyPlan), totalShares)
	if !ok {
		return nil, ErrGroupBuyTierMappingInvalid
	}
	targetGroupID := tierRule.TargetGroupID
	if targetGroupID <= 0 {
		return nil, ErrGroupBuyTierMappingInvalid
	}
	if err := s.validateTargetGroup(ctx, targetGroupID); err != nil {
		return nil, err
	}
	sub, err := s.ensureEntitlementSubscription(ctx, userID, targetGroupID, maxExpires, ent.ManagedSubscriptionID, ent.ID, totalShares)
	if err != nil {
		return nil, err
	}
	previousGroupID := psInt64Value(ent.TargetGroupID)
	ent, err = s.entClient.GroupBuyEntitlement.UpdateOneID(ent.ID).
		SetStatus(GroupBuyEntitlementStatusActive).
		SetActiveShareCount(totalShares).
		SetTargetGroupID(targetGroupID).
		SetSubscriptionID(sub.ID).
		SetManagedSubscriptionID(sub.ID).
		SetNillableLastActivatedAt(lastActivated).
		SetNillableExpiresAt(maxExpires).
		SetRefreshedAt(now).
		ClearDeactivatedAt().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update group buy entitlement: %w", err)
	}
	if ent.BoundAPIKeyID != nil && (previousGroupID == 0 || previousGroupID != targetGroupID) {
		_, _ = s.apiKeySvc.Update(ctx, *ent.BoundAPIKeyID, userID, UpdateAPIKeyRequest{GroupID: &targetGroupID})
	}
	s.createEvent(ctx, &groupBuyEventInput{
		UserID:    &userID,
		EventType: groupBuyEventEntitlementSync,
		Message:   "Token拼拼拼 份额权益已同步",
		Metadata: map[string]any{
			"active_share_count": totalShares,
			"target_group_id":    targetGroupID,
			"tier_label":         tierRule.Label,
			"subscription_id":    sub.ID,
		},
	})
	return s.entitlementView(ctx, ent), nil
}

func (s *GroupBuyService) getOrCreateEntitlement(ctx context.Context, userID int64) (*dbent.GroupBuyEntitlement, error) {
	ent, err := s.entClient.GroupBuyEntitlement.Query().
		Where(groupbuyentitlement.UserIDEQ(userID), groupbuyentitlement.ProductKeyEQ(GroupBuyProductTokenPinPinPin)).
		WithTargetGroup().
		Only(ctx)
	if err == nil {
		return ent, nil
	}
	if !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("load group buy entitlement: %w", err)
	}
	ent, err = s.entClient.GroupBuyEntitlement.Create().
		SetUserID(userID).
		SetProductKey(GroupBuyProductTokenPinPinPin).
		SetStatus(GroupBuyEntitlementStatusInactive).
		SetActiveShareCount(0).
		SetRefreshedAt(s.now()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create group buy entitlement: %w", err)
	}
	return ent, nil
}

func (s *GroupBuyService) deactivateEntitlement(ctx context.Context, ent *dbent.GroupBuyEntitlement) (*dbent.GroupBuyEntitlement, error) {
	now := s.now()
	managedSubscriptionID := ent.ManagedSubscriptionID
	if managedSubscriptionID == nil {
		managedSubscriptionID = ent.SubscriptionID
	}
	if managedSubscriptionID != nil {
		if sub, err := s.entClient.UserSubscription.Get(ctx, *managedSubscriptionID); err == nil && isGroupBuyManagedSubscription(sub) {
			_, _ = s.entClient.UserSubscription.UpdateOneID(sub.ID).
				SetStatus(SubscriptionStatusExpired).
				SetExpiresAt(now).
				Save(ctx)
			s.subscriptionSvc.InvalidateSubCache(sub.UserID, sub.GroupID)
		}
	}
	ent, err := s.entClient.GroupBuyEntitlement.UpdateOneID(ent.ID).
		SetStatus(GroupBuyEntitlementStatusInactive).
		SetActiveShareCount(0).
		ClearTargetGroupID().
		ClearSubscriptionID().
		ClearManagedSubscriptionID().
		ClearExpiresAt().
		SetRefreshedAt(now).
		SetDeactivatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("deactivate group buy entitlement: %w", err)
	}
	return ent, nil
}

func (s *GroupBuyService) ensureEntitlementSubscription(ctx context.Context, userID, groupID int64, expiresAt *time.Time, managedSubscriptionID *int64, entitlementID int64, shares int) (*dbent.UserSubscription, error) {
	now := s.now()
	if expiresAt == nil || !expiresAt.After(now) {
		t := now.AddDate(0, 0, 1)
		expiresAt = &t
	}
	note := fmt.Sprintf("%s shares=%d refreshed=%s", groupBuySubscriptionNotePrefix, shares, now.UTC().Format(time.RFC3339))
	if managedSubscriptionID != nil {
		if prev, err := s.entClient.UserSubscription.Get(ctx, *managedSubscriptionID); err == nil && isGroupBuyManagedSubscription(prev) {
			if prev.GroupID == groupID {
				notes := appendSubscriptionNotes(psStringValue(prev.Notes), note)
				sub, err := s.entClient.UserSubscription.UpdateOneID(prev.ID).
					SetStatus(SubscriptionStatusActive).
					SetExpiresAt(*expiresAt).
					SetSourceType("group_buy").
					SetSourceID(entitlementID).
					SetManagedByGroupBuy(true).
					SetNotes(notes).
					Save(ctx)
				if err != nil {
					return nil, fmt.Errorf("update group buy subscription: %w", err)
				}
				s.subscriptionSvc.InvalidateSubCache(userID, groupID)
				return sub, nil
			}
			_, _ = s.entClient.UserSubscription.UpdateOneID(prev.ID).SetStatus(SubscriptionStatusExpired).SetExpiresAt(now).Save(ctx)
			s.subscriptionSvc.InvalidateSubCache(prev.UserID, prev.GroupID)
		}
	}
	existing, err := s.entClient.UserSubscription.Query().
		Where(
			usersubscription.UserIDEQ(userID),
			usersubscription.GroupIDEQ(groupID),
			usersubscription.SourceTypeEQ("group_buy"),
			usersubscription.SourceIDEQ(entitlementID),
		).
		First(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("load existing group buy subscription: %w", err)
	}
	if existing != nil && isGroupBuyManagedSubscription(existing) {
		notes := appendSubscriptionNotes(psStringValue(existing.Notes), note)
		sub, err := s.entClient.UserSubscription.UpdateOneID(existing.ID).
			SetStatus(SubscriptionStatusActive).
			SetExpiresAt(*expiresAt).
			SetSourceType("group_buy").
			SetSourceID(entitlementID).
			SetManagedByGroupBuy(true).
			SetNotes(notes).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("restore group buy subscription: %w", err)
		}
		s.subscriptionSvc.InvalidateSubCache(userID, groupID)
		return sub, nil
	}
	windowStart := startOfDay(now)
	sub, err := s.entClient.UserSubscription.Create().
		SetUserID(userID).
		SetGroupID(groupID).
		SetStartsAt(now).
		SetExpiresAt(*expiresAt).
		SetStatus(SubscriptionStatusActive).
		SetSourceType("group_buy").
		SetSourceID(entitlementID).
		SetManagedByGroupBuy(true).
		SetDailyWindowStart(windowStart).
		SetWeeklyWindowStart(windowStart).
		SetMonthlyWindowStart(windowStart).
		SetAssignedAt(now).
		SetNotes(note).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create group buy subscription: %w", err)
	}
	s.subscriptionSvc.InvalidateSubCache(userID, groupID)
	return sub, nil
}

func (s *GroupBuyService) BindKey(ctx context.Context, userID, seatID, apiKeyID int64) (*GroupBuySeatView, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	seat, err := s.entClient.GroupBuySeat.Query().
		Where(groupbuyseat.IDEQ(seatID), groupbuyseat.UserIDEQ(userID)).
		WithPlan().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuySeatNotFound
		}
		return nil, fmt.Errorf("load group buy share batch: %w", err)
	}
	if seat.Status != GroupBuySeatStatusActive || (seat.ExpiresAt != nil && !seat.ExpiresAt.After(s.now())) {
		return nil, ErrGroupBuySeatNotActive
	}
	entView, err := s.RefreshUserEntitlement(ctx, userID)
	if err != nil {
		return nil, err
	}
	if entView == nil || entView.TargetGroupID == nil || entView.ActiveShareCount <= 0 {
		return nil, ErrGroupBuySeatNotActive
	}
	targetGroupID := *entView.TargetGroupID
	if _, err := s.apiKeySvc.Update(ctx, apiKeyID, userID, UpdateAPIKeyRequest{GroupID: &targetGroupID}); err != nil {
		return nil, err
	}
	now := s.now()
	ent, err := s.getOrCreateEntitlement(ctx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := s.entClient.GroupBuyEntitlement.UpdateOneID(ent.ID).
		SetBoundAPIKeyID(apiKeyID).
		SetRefreshedAt(now).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("bind group buy entitlement api key: %w", err)
	}
	updated, err := s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
		SetBoundAPIKeyID(apiKeyID).
		SetBoundAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("bind group buy api key: %w", err)
	}
	s.createEvent(ctx, &groupBuyEventInput{
		PlanID:    &seat.PlanID,
		RoundID:   &seat.RoundID,
		SeatID:    &seat.ID,
		UserID:    &userID,
		EventType: groupBuyEventEntitlementBound,
		Message:   "已绑定用户自己的平台 API Key",
		Metadata:  map[string]any{"api_key_id": apiKeyID, "target_group_id": targetGroupID},
	})
	return s.seatView(ctx, updated), nil
}

func (s *GroupBuyService) ListMySeats(ctx context.Context, userID int64) (*GroupBuyMySeatsView, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	entitlement, err := s.RefreshUserEntitlement(ctx, userID)
	if err != nil {
		return nil, err
	}
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(groupbuyseat.UserIDEQ(userID)).
		WithPlan(func(q *dbent.GroupBuyPlanQuery) { q.WithTargetGroup() }).
		WithRound().
		WithOrder().
		Order(dbent.Desc(groupbuyseat.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list group buy share batches: %w", err)
	}
	out := make([]GroupBuySeatView, 0, len(seats))
	for _, seat := range seats {
		out = append(out, *s.seatView(ctx, seat))
	}
	return &GroupBuyMySeatsView{Entitlement: entitlement, Seats: out}, nil
}

func (s *GroupBuyService) GetMyEntitlement(ctx context.Context, userID int64) (*GroupBuyEntitlementView, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	return s.RefreshUserEntitlement(ctx, userID)
}

func (s *GroupBuyService) ListMyOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]PaymentOrderLite, *pagination.PaginationResult, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, nil, err
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	q := s.entClient.PaymentOrder.Query().
		Where(paymentorder.UserIDEQ(userID), paymentorder.OrderTypeEQ(payment.OrderTypeGroupBuy))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count group buy orders: %w", err)
	}
	orders, err := q.Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list group buy orders: %w", err)
	}
	out := make([]PaymentOrderLite, 0, len(orders))
	for _, order := range orders {
		out = append(out, paymentOrderLite(order))
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (s *GroupBuyService) AdminCreatePlan(ctx context.Context, input GroupBuyPlanInput) (*GroupBuyPlanView, error) {
	if err := s.validatePlanInput(ctx, &input); err != nil {
		return nil, err
	}
	plan, err := s.entClient.GroupBuyPlan.Create().
		SetTitle(strings.TrimSpace(input.Title)).
		SetDescription(strings.TrimSpace(input.Description)).
		SetProductKey(input.ProductKey).
		SetTotalShares(input.TotalShares).
		SetSeatCount(input.TotalShares).
		SetPricePerShare(input.PricePerShare).
		SetPricePerSeat(input.PricePerShare).
		SetPriceLabel(strings.TrimSpace(input.PriceLabel)).
		SetQuotaPerShareLabel(strings.TrimSpace(input.QuotaPerShareLabel)).
		SetQuotaLabel(strings.TrimSpace(input.QuotaPerShareLabel)).
		SetMaxSharesPerUser(input.MaxSharesPerUser).
		SetTargetGroupID(input.TargetGroupID).
		SetTierGroupIds(input.TierGroupIDs).
		SetTierRules(tierRuleInputsToDomain(input.TierRules)).
		SetValidityDays(input.ValidityDays).
		SetTimeoutMinutes(input.TimeoutMinutes).
		SetLaunchMode(input.LaunchMode).
		SetRefundMode(normalizeGroupBuyRefundMode(input.RefundMode)).
		SetAgreementText(strings.TrimSpace(input.AgreementText)).
		SetStatus(normalizeGroupBuyPlanStatus(input.Status)).
		SetSortOrder(input.SortOrder).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create group buy plan: %w", err)
	}
	if plan.LaunchMode == GroupBuyLaunchModeAuto && plan.Status == GroupBuyPlanStatusActive {
		if err := s.ensureAutoNextRound(ctx, plan.ID); err != nil {
			return nil, err
		}
	}
	view := s.planView(ctx, plan)
	return &view, nil
}

func (s *GroupBuyService) AdminUpdatePlan(ctx context.Context, id int64, input GroupBuyPlanInput) (*GroupBuyPlanView, error) {
	if id <= 0 {
		return nil, ErrGroupBuyPlanNotFound
	}
	if err := s.validatePlanInput(ctx, &input); err != nil {
		return nil, err
	}
	plan, err := s.entClient.GroupBuyPlan.UpdateOneID(id).
		SetTitle(strings.TrimSpace(input.Title)).
		SetDescription(strings.TrimSpace(input.Description)).
		SetProductKey(input.ProductKey).
		SetTotalShares(input.TotalShares).
		SetSeatCount(input.TotalShares).
		SetPricePerShare(input.PricePerShare).
		SetPricePerSeat(input.PricePerShare).
		SetPriceLabel(strings.TrimSpace(input.PriceLabel)).
		SetQuotaPerShareLabel(strings.TrimSpace(input.QuotaPerShareLabel)).
		SetQuotaLabel(strings.TrimSpace(input.QuotaPerShareLabel)).
		SetMaxSharesPerUser(input.MaxSharesPerUser).
		SetTargetGroupID(input.TargetGroupID).
		SetTierGroupIds(input.TierGroupIDs).
		SetTierRules(tierRuleInputsToDomain(input.TierRules)).
		SetValidityDays(input.ValidityDays).
		SetTimeoutMinutes(input.TimeoutMinutes).
		SetLaunchMode(input.LaunchMode).
		SetRefundMode(normalizeGroupBuyRefundMode(input.RefundMode)).
		SetAgreementText(strings.TrimSpace(input.AgreementText)).
		SetStatus(normalizeGroupBuyPlanStatus(input.Status)).
		SetSortOrder(input.SortOrder).
		Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuyPlanNotFound
		}
		return nil, fmt.Errorf("update group buy plan: %w", err)
	}
	if plan.LaunchMode == GroupBuyLaunchModeAuto && plan.Status == GroupBuyPlanStatusActive {
		if err := s.ensureAutoNextRound(ctx, plan.ID); err != nil {
			return nil, err
		}
	}
	view := s.planView(ctx, plan)
	return &view, nil
}

func (s *GroupBuyService) AdminDeletePlan(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrGroupBuyPlanNotFound
	}
	now := s.now()
	n, err := s.entClient.GroupBuyPlan.Update().
		Where(groupbuyplan.IDEQ(id), groupbuyplan.DeletedAtIsNil()).
		SetDeletedAt(now).
		SetStatus(GroupBuyPlanStatusDisabled).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("delete group buy plan: %w", err)
	}
	if n == 0 {
		return ErrGroupBuyPlanNotFound
	}
	return nil
}

func (s *GroupBuyService) AdminCreateRound(ctx context.Context, planID int64) (*GroupBuyRoundView, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin manual group buy round tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	planQuery := tx.GroupBuyPlan.Query().
		Where(groupbuyplan.IDEQ(planID), groupbuyplan.DeletedAtIsNil())
	plan, err := s.groupBuyPlanForUpdate(planQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuyPlanNotFound
		}
		return nil, err
	}
	if plan.Status != GroupBuyPlanStatusActive {
		return nil, ErrGroupBuyPlanUnavailable
	}
	if err := s.validatePlanTierRulesInTx(txCtx, tx, plan); err != nil {
		return nil, err
	}
	existingQuery := tx.GroupBuyRound.Query().
		Where(groupbuyround.PlanIDEQ(plan.ID), groupbuyround.StatusEQ(GroupBuyRoundStatusOpen), groupbuyround.DeadlineAtGT(s.now())).
		Limit(1)
	existing, err := s.groupBuyRoundForUpdate(existingQuery).Only(txCtx)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return roundView(existing), nil
	}
	if !dbent.IsNotFound(err) {
		return nil, err
	}
	round, err := s.createRoundForPlanTx(txCtx, tx, plan)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return roundView(round), nil
}

func (s *GroupBuyService) AdminListRounds(ctx context.Context, status string, params pagination.PaginationParams) ([]GroupBuyRoundView, *pagination.PaginationResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	q := s.entClient.GroupBuyRound.Query()
	if strings.TrimSpace(status) != "" {
		q = q.Where(groupbuyround.StatusEQ(strings.TrimSpace(status)))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count group buy rounds: %w", err)
	}
	rounds, err := q.Order(dbent.Desc(groupbuyround.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list group buy rounds: %w", err)
	}
	out := make([]GroupBuyRoundView, 0, len(rounds))
	for _, round := range rounds {
		out = append(out, *roundView(round))
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (s *GroupBuyService) AdminCloseRound(ctx context.Context, roundID int64, reason string) error {
	if roundID <= 0 {
		return ErrGroupBuyRoundNotFound
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin group buy close tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	roundQuery := tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID))
	round, err := s.groupBuyRoundForUpdate(roundQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return ErrGroupBuyRoundNotFound
		}
		return err
	}
	if round.Status != GroupBuyRoundStatusOpen {
		return ErrGroupBuyInvalidStatus.WithMetadata(map[string]string{"round_status": round.Status})
	}
	now := s.now()
	if _, err := tx.GroupBuyRound.UpdateOneID(round.ID).
		SetStatus(GroupBuyRoundStatusCancelled).
		SetClosedAt(now).
		SetCloseReason(strings.TrimSpace(reason)).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return fmt.Errorf("close group buy round: %w", err)
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusPaid)).
		SetStatus(GroupBuySeatStatusRefundPending).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return err
	}
	if _, err := tx.GroupBuySeat.Update().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked)).
		SetStatus(GroupBuySeatStatusCancelled).
		SetUpdatedAt(now).
		Save(txCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *GroupBuyService) AdminRetryActivation(ctx context.Context, roundID int64) error {
	return s.TryActivateRound(ctx, roundID)
}

func (s *GroupBuyService) AdminProcessRefunds(ctx context.Context, roundID int64) (int, error) {
	round, err := s.entClient.GroupBuyRound.Query().Where(groupbuyround.IDEQ(roundID)).WithPlan().Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return 0, ErrGroupBuyRoundNotFound
		}
		return 0, err
	}
	if round.Status != GroupBuyRoundStatusFailed && round.Status != GroupBuyRoundStatusCancelled {
		return 0, ErrGroupBuyInvalidStatus.WithMetadata(map[string]string{"round_status": round.Status})
	}
	plan := round.Edges.Plan
	if plan == nil {
		return 0, ErrGroupBuyPlanNotFound
	}
	seats, err := s.entClient.GroupBuySeat.Query().
		Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusRefundPending)).
		WithOrder().
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list refund pending share batches: %w", err)
	}
	processed := 0
	for _, seat := range seats {
		if err := s.processSeatRefund(ctx, plan, seat); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func (s *GroupBuyService) processSeatRefund(ctx context.Context, plan *dbent.GroupBuyPlan, seat *dbent.GroupBuySeat) error {
	order := seat.Edges.Order
	if order == nil && seat.OrderID != nil {
		var err error
		order, err = s.entClient.PaymentOrder.Get(ctx, *seat.OrderID)
		if err != nil && !dbent.IsNotFound(err) {
			return err
		}
	}
	now := s.now()
	mode := normalizeGroupBuyRefundMode(plan.RefundMode)
	amount := plan.PricePerShare * float64(seat.ShareCount)
	if order != nil && order.Amount > 0 {
		amount = order.Amount
	}
	note := "Token拼拼拼 未满份退款"
	if mode == GroupBuyRefundModeProviderRefund {
		note = "Token拼拼拼 未满份，等待原路退款"
	}
	existingRefund, err := s.entClient.GroupBuyRefund.Query().
		Where(groupbuyrefund.SeatIDEQ(seat.ID)).
		Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return fmt.Errorf("load group buy refund record: %w", err)
	}
	if existingRefund != nil {
		switch existingRefund.Status {
		case GroupBuyRefundStatusSucceeded:
			_, _ = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
				SetStatus(GroupBuySeatStatusRefunded).
				SetNillableRefundProcessedAt(existingRefund.ProcessedAt).
				SetRefundNote(psStringValue(existingRefund.Note)).
				SetUpdatedAt(now).
				Save(ctx)
			return nil
		case GroupBuyRefundStatusPendingProvider, GroupBuyRefundStatusProcessing:
			_, _ = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
				SetStatus(GroupBuySeatStatusRefundProcessing).
				SetRefundNote(psStringValue(existingRefund.Note)).
				SetUpdatedAt(now).
				Save(ctx)
			return nil
		}
	}
	updated, err := s.entClient.GroupBuySeat.Update().
		Where(groupbuyseat.IDEQ(seat.ID), groupbuyseat.StatusEQ(GroupBuySeatStatusRefundPending)).
		SetStatus(GroupBuySeatStatusRefundProcessing).
		SetRefundNote(note).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("claim group buy refund: %w", err)
	}
	if updated == 0 {
		current, getErr := s.entClient.GroupBuySeat.Get(ctx, seat.ID)
		if getErr != nil {
			return getErr
		}
		if current.Status == GroupBuySeatStatusRefunded || current.Status == GroupBuySeatStatusRefundProcessing {
			return nil
		}
		return ErrGroupBuyInvalidStatus.WithMetadata(map[string]string{"seat_status": current.Status})
	}
	refund := existingRefund
	if refund == nil {
		create := s.entClient.GroupBuyRefund.Create().
			SetSeatID(seat.ID).
			SetUserID(seat.UserID).
			SetMode(mode).
			SetStatus(GroupBuyRefundStatusProcessing).
			SetAmount(amount).
			SetIdempotencyKey(groupBuyRefundIdempotencyKey(seat)).
			SetNote(note)
		if order != nil {
			create.SetOrderID(order.ID)
		}
		refund, err = create.Save(ctx)
		if err != nil {
			if dbent.IsConstraintError(err) {
				return nil
			}
			_, _ = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
				SetStatus(GroupBuySeatStatusRefundPending).
				SetUpdatedAt(now).
				Save(ctx)
			return fmt.Errorf("create group buy refund record: %w", err)
		}
	} else {
		refund, err = s.entClient.GroupBuyRefund.UpdateOneID(refund.ID).
			SetMode(mode).
			SetStatus(GroupBuyRefundStatusProcessing).
			SetAmount(amount).
			SetNote(note).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			_, _ = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
				SetStatus(GroupBuySeatStatusRefundPending).
				SetUpdatedAt(now).
				Save(ctx)
			return fmt.Errorf("retry group buy refund record: %w", err)
		}
	}
	if mode == GroupBuyRefundModeProviderRefund {
		if order != nil {
			_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).
				SetStatus(OrderStatusRefundPending).
				SetRefundAmount(order.PayAmount).
				SetRefundRequestReason(note).
				SetRefundRequestedAt(now).
				SetRefundRequestedBy("admin").
				Save(ctx)
		}
		if _, err := s.entClient.GroupBuyRefund.UpdateOneID(refund.ID).
			SetStatus(GroupBuyRefundStatusPendingProvider).
			SetProcessedAt(now).
			SetUpdatedAt(now).
			Save(ctx); err != nil {
			return fmt.Errorf("mark group buy provider refund pending: %w", err)
		}
		_, err := s.entClient.GroupBuySeat.UpdateOneID(seat.ID).SetRefundNote(note).SetUpdatedAt(now).Save(ctx)
		return err
	}
	if err := grantWelfareBalance(ctx, s.userRepo, seat.UserID, amount); err != nil {
		_, _ = s.entClient.GroupBuyRefund.UpdateOneID(refund.ID).
			SetStatus(GroupBuyRefundStatusFailed).
			SetNote(note + ": " + err.Error()).
			SetUpdatedAt(now).
			Save(ctx)
		_, _ = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
			SetStatus(GroupBuySeatStatusRefundPending).
			SetUpdatedAt(now).
			Save(ctx)
		return fmt.Errorf("credit group buy refund balance: %w", err)
	}
	s.invalidateBalanceCaches(ctx, seat.UserID)
	if order != nil {
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).
			SetStatus(OrderStatusRefunded).
			SetRefundAmount(amount).
			SetRefundReason(note).
			SetRefundAt(now).
			Save(ctx)
	}
	_, err = s.entClient.GroupBuyRefund.UpdateOneID(refund.ID).
		SetStatus(GroupBuyRefundStatusSucceeded).
		SetProcessedAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark group buy refund record succeeded: %w", err)
	}
	_, err = s.entClient.GroupBuySeat.UpdateOneID(seat.ID).
		SetStatus(GroupBuySeatStatusRefunded).
		SetRefundProcessedAt(now).
		SetRefundNote(note).
		SetUpdatedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark group buy refund processed: %w", err)
	}
	s.createEvent(ctx, &groupBuyEventInput{
		PlanID:    &seat.PlanID,
		RoundID:   &seat.RoundID,
		SeatID:    &seat.ID,
		UserID:    &seat.UserID,
		EventType: groupBuyEventRefundProcessed,
		Message:   "Token拼拼拼 未满份，已退回余额",
		Metadata:  map[string]any{"amount": amount, "share_count": seat.ShareCount},
	})
	return nil
}

func groupBuyRefundIdempotencyKey(seat *dbent.GroupBuySeat) string {
	if seat == nil {
		return "group_buy_refund_unknown"
	}
	if seat.OrderID != nil {
		return fmt.Sprintf("group_buy_refund_seat_%d_order_%d", seat.ID, *seat.OrderID)
	}
	return fmt.Sprintf("group_buy_refund_seat_%d", seat.ID)
}

func (s *GroupBuyService) loadAvailablePlan(ctx context.Context, planID int64) (*dbent.GroupBuyPlan, error) {
	plan, err := s.entClient.GroupBuyPlan.Query().
		Where(groupbuyplan.IDEQ(planID), groupbuyplan.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrGroupBuyPlanNotFound
		}
		return nil, fmt.Errorf("load group buy plan: %w", err)
	}
	if plan.Status != GroupBuyPlanStatusActive {
		return nil, ErrGroupBuyPlanUnavailable
	}
	if err := validateDomainTierRules(normalizedPlanTierRules(plan), plan.TotalShares); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *GroupBuyService) latestProductPlan(ctx context.Context) (*dbent.GroupBuyPlan, error) {
	return s.entClient.GroupBuyPlan.Query().
		Where(groupbuyplan.ProductKeyEQ(GroupBuyProductTokenPinPinPin), groupbuyplan.StatusEQ(GroupBuyPlanStatusActive), groupbuyplan.DeletedAtIsNil()).
		Order(dbent.Asc(groupbuyplan.FieldSortOrder), dbent.Desc(groupbuyplan.FieldID)).
		First(ctx)
}

func (s *GroupBuyService) validatePlanInput(ctx context.Context, input *GroupBuyPlanInput) error {
	if input == nil {
		return infraerrors.BadRequest("INVALID_INPUT", "plan input is required")
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "title is required")
	}
	input.ProductKey = strings.TrimSpace(input.ProductKey)
	if input.ProductKey == "" {
		input.ProductKey = GroupBuyProductTokenPinPinPin
	}
	if input.ProductKey != GroupBuyProductTokenPinPinPin {
		return infraerrors.BadRequest("INVALID_INPUT", "unsupported group buy product")
	}
	if input.TotalShares <= 0 {
		input.TotalShares = input.SeatCount
	}
	if input.TotalShares <= 0 {
		input.TotalShares = defaultGroupBuyTotalShares
	}
	if input.TotalShares > groupBuyMaxShareCount {
		input.TotalShares = groupBuyMaxShareCount
	}
	if input.PricePerShare <= 0 {
		input.PricePerShare = input.PricePerSeat
	}
	if input.PricePerShare <= 0 {
		return infraerrors.BadRequest("INVALID_INPUT", "price_per_share must be positive")
	}
	input.PriceLabel = strings.TrimSpace(input.PriceLabel)
	if len(input.PriceLabel) > 120 {
		return infraerrors.BadRequest("INVALID_INPUT", "price_label is too long")
	}
	if input.QuotaPerShareLabel == "" {
		input.QuotaPerShareLabel = input.QuotaLabel
	}
	if input.MaxSharesPerUser <= 0 {
		input.MaxSharesPerUser = defaultGroupBuyMaxUserShares
	}
	if input.MaxSharesPerUser > groupBuyMaxShareCount {
		input.MaxSharesPerUser = groupBuyMaxShareCount
	}
	if input.ValidityDays <= 0 {
		input.ValidityDays = 30
	}
	if input.TimeoutMinutes <= 0 {
		input.TimeoutMinutes = 1440
	}
	input.LaunchMode = normalizeGroupBuyLaunchMode(input.LaunchMode)
	input.RefundMode = normalizeGroupBuyRefundMode(input.RefundMode)
	input.Status = normalizeGroupBuyPlanStatus(input.Status)
	input.TierRules = normalizeTierRuleInputs(input.TotalShares, input.TargetGroupID, input.TierRules, input.TierGroups, input.TierGroupIDs)
	if err := validateTierRuleShape(input.TierRules, input.TotalShares); err != nil {
		return err
	}
	input.TierGroupIDs = tierRulesToExactMapping(input.TierRules, input.TotalShares)
	input.TargetGroupID = targetGroupIDForShareCount(input.TierRules, input.TotalShares)
	if input.TargetGroupID <= 0 {
		input.TargetGroupID = input.TierRules[len(input.TierRules)-1].TargetGroupID
	}
	return s.validateTierRules(ctx, input.TierRules)
}

func normalizeTierRuleInputs(totalShares int, fallbackGroupID int64, primary []GroupBuyTierInput, legacy []GroupBuyTierInput, raw map[string]int64) []GroupBuyTierInput {
	source := primary
	if len(source) == 0 {
		source = legacy
	}
	out := make([]GroupBuyTierInput, 0, len(source))
	for _, tier := range source {
		minShares := tier.MinShares
		maxShares := tier.MaxShares
		if minShares <= 0 && maxShares <= 0 && tier.ShareCount > 0 {
			minShares = tier.ShareCount
			maxShares = tier.ShareCount
		}
		if maxShares <= 0 {
			maxShares = minShares
		}
		label := strings.TrimSpace(tier.Label)
		if label == "" && minShares > 0 && maxShares > 0 {
			if minShares == maxShares {
				label = fmt.Sprintf("%d 份", minShares)
			} else {
				label = fmt.Sprintf("%d-%d 份", minShares, maxShares)
			}
		}
		if tier.TargetGroupID > 0 {
			out = append(out, GroupBuyTierInput{
				ShareCount:    tier.ShareCount,
				MinShares:     minShares,
				MaxShares:     maxShares,
				TargetGroupID: tier.TargetGroupID,
				Label:         label,
			})
		}
	}
	if len(out) > 0 {
		return normalizeTierRuleOrder(out)
	}
	if len(raw) > 0 {
		return exactTierMapToRules(raw, totalShares)
	}
	if fallbackGroupID > 0 {
		return []GroupBuyTierInput{{
			MinShares:     groupBuyMinShareCount,
			MaxShares:     totalShares,
			TargetGroupID: fallbackGroupID,
			Label:         "默认档位",
		}}
	}
	return nil
}

func exactTierMapToRules(raw map[string]int64, totalShares int) []GroupBuyTierInput {
	type exactTier struct {
		share   int
		groupID int64
	}
	items := make([]exactTier, 0, len(raw))
	for key, value := range raw {
		if value <= 0 {
			continue
		}
		share, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil || share < groupBuyMinShareCount || share > totalShares {
			continue
		}
		items = append(items, exactTier{share: share, groupID: value})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].share < items[j].share })
	out := make([]GroupBuyTierInput, 0, len(items))
	for i := 0; i < len(items); {
		start := items[i].share
		end := start
		groupID := items[i].groupID
		i++
		for i < len(items) && items[i].share == end+1 && items[i].groupID == groupID {
			end = items[i].share
			i++
		}
		label := fmt.Sprintf("%d 份", start)
		if start != end {
			label = fmt.Sprintf("%d-%d 份", start, end)
		}
		out = append(out, GroupBuyTierInput{MinShares: start, MaxShares: end, TargetGroupID: groupID, Label: label})
	}
	return out
}

func normalizeTierRuleOrder(rules []GroupBuyTierInput) []GroupBuyTierInput {
	out := make([]GroupBuyTierInput, 0, len(rules))
	for _, rule := range rules {
		out = append(out, rule)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].MinShares == out[j].MinShares {
			return out[i].MaxShares < out[j].MaxShares
		}
		return out[i].MinShares < out[j].MinShares
	})
	return out
}

func validateTierRuleShape(rules []GroupBuyTierInput, totalShares int) error {
	if totalShares <= 0 {
		totalShares = defaultGroupBuyTotalShares
	}
	if len(rules) == 0 {
		return ErrGroupBuyTierMappingInvalid
	}
	expected := groupBuyMinShareCount
	for _, rule := range normalizeTierRuleOrder(rules) {
		if rule.MinShares != expected || rule.MaxShares < rule.MinShares || rule.TargetGroupID <= 0 || rule.MaxShares > totalShares {
			return ErrGroupBuyTierMappingInvalid.WithMetadata(map[string]string{
				"expected_min_shares": strconv.Itoa(expected),
			})
		}
		expected = rule.MaxShares + 1
	}
	if expected != totalShares+1 {
		return ErrGroupBuyTierMappingInvalid.WithMetadata(map[string]string{"missing_share_count": strconv.Itoa(expected)})
	}
	return nil
}

func (s *GroupBuyService) validateTierRules(ctx context.Context, rules []GroupBuyTierInput) error {
	seen := map[int64]struct{}{}
	for _, rule := range rules {
		if _, ok := seen[rule.TargetGroupID]; ok {
			continue
		}
		seen[rule.TargetGroupID] = struct{}{}
		if err := s.validateTargetGroup(ctx, rule.TargetGroupID); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupBuyService) validatePlanTierRulesInTx(ctx context.Context, tx *dbent.Tx, plan *dbent.GroupBuyPlan) error {
	rules := normalizedPlanTierRules(plan)
	if err := validateDomainTierRules(rules, plan.TotalShares); err != nil {
		return err
	}
	seen := map[int64]struct{}{}
	for _, rule := range rules {
		if _, ok := seen[rule.TargetGroupID]; ok {
			continue
		}
		seen[rule.TargetGroupID] = struct{}{}
		if err := s.validateTargetGroupInTx(ctx, tx, rule.TargetGroupID); err != nil {
			return err
		}
	}
	return nil
}

func validateDomainTierRules(rules []domain.GroupBuyTierRule, totalShares int) error {
	if len(rules) == 0 {
		return ErrGroupBuyTierMappingInvalid
	}
	expected := groupBuyMinShareCount
	for _, rule := range rules {
		if rule.MinShares != expected || rule.MaxShares < rule.MinShares || rule.TargetGroupID <= 0 || rule.MaxShares > totalShares {
			return ErrGroupBuyTierMappingInvalid.WithMetadata(map[string]string{"expected_min_shares": strconv.Itoa(expected)})
		}
		expected = rule.MaxShares + 1
	}
	if expected != totalShares+1 {
		return ErrGroupBuyTierMappingInvalid.WithMetadata(map[string]string{"missing_share_count": strconv.Itoa(expected)})
	}
	return nil
}

func tierRuleInputsToDomain(rules []GroupBuyTierInput) []domain.GroupBuyTierRule {
	ordered := normalizeTierRuleOrder(rules)
	out := make([]domain.GroupBuyTierRule, 0, len(ordered))
	for _, rule := range ordered {
		label := strings.TrimSpace(rule.Label)
		if label == "" {
			if rule.MinShares == rule.MaxShares {
				label = fmt.Sprintf("%d 份", rule.MinShares)
			} else {
				label = fmt.Sprintf("%d-%d 份", rule.MinShares, rule.MaxShares)
			}
		}
		out = append(out, domain.GroupBuyTierRule{
			MinShares:     rule.MinShares,
			MaxShares:     rule.MaxShares,
			TargetGroupID: rule.TargetGroupID,
			Label:         label,
		})
	}
	return out
}

func domainTierRulesToInputs(rules []domain.GroupBuyTierRule) []GroupBuyTierInput {
	out := make([]GroupBuyTierInput, 0, len(rules))
	for _, rule := range rules {
		out = append(out, GroupBuyTierInput{
			MinShares:     rule.MinShares,
			MaxShares:     rule.MaxShares,
			TargetGroupID: rule.TargetGroupID,
			Label:         rule.Label,
		})
	}
	return out
}

func normalizedPlanTierRules(plan *dbent.GroupBuyPlan) []domain.GroupBuyTierRule {
	if plan == nil {
		return nil
	}
	rules := append([]domain.GroupBuyTierRule(nil), plan.TierRules...)
	if len(rules) == 0 {
		rules = tierRuleInputsToDomain(exactTierMapToRules(plan.TierGroupIds, plan.TotalShares))
	}
	if len(rules) == 0 && plan.TargetGroupID > 0 {
		rules = []domain.GroupBuyTierRule{{
			MinShares:     groupBuyMinShareCount,
			MaxShares:     plan.TotalShares,
			TargetGroupID: plan.TargetGroupID,
			Label:         "默认档位",
		}}
	}
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].MinShares == rules[j].MinShares {
			return rules[i].MaxShares < rules[j].MaxShares
		}
		return rules[i].MinShares < rules[j].MinShares
	})
	return rules
}

func tierRulesToExactMapping(rules []GroupBuyTierInput, totalShares int) map[string]int64 {
	out := map[string]int64{}
	for _, rule := range rules {
		maxShares := rule.MaxShares
		if maxShares > totalShares {
			maxShares = totalShares
		}
		for share := rule.MinShares; share <= maxShares; share++ {
			out[strconv.Itoa(share)] = rule.TargetGroupID
		}
	}
	return out
}

func resolveTierRuleForShareCount(rules []domain.GroupBuyTierRule, shareCount int) (domain.GroupBuyTierRule, bool) {
	for _, rule := range rules {
		if shareCount >= rule.MinShares && shareCount <= rule.MaxShares {
			return rule, true
		}
	}
	return domain.GroupBuyTierRule{}, false
}

func targetGroupIDForShareCount(rules []GroupBuyTierInput, shareCount int) int64 {
	rule, ok := resolveTierRuleForShareCount(tierRuleInputsToDomain(rules), shareCount)
	if !ok {
		return 0
	}
	return rule.TargetGroupID
}

func buildGroupBuyPolicySnapshot(plan *dbent.GroupBuyPlan, capturedAt time.Time) domain.GroupBuyPolicySnapshot {
	if plan == nil {
		return domain.GroupBuyPolicySnapshot{}
	}
	return domain.GroupBuyPolicySnapshot{
		ProductKey:          plan.ProductKey,
		PlanID:              plan.ID,
		TotalShares:         plan.TotalShares,
		QuotaPerShareLabel:  plan.QuotaPerShareLabel,
		TierRules:           normalizedPlanTierRules(plan),
		LegacyTierGroupIDs:  copyTierGroupIDs(plan.TierGroupIds),
		TargetGroupID:       plan.TargetGroupID,
		CapturedAtUnixMilli: capturedAt.UnixMilli(),
	}
}

func (s *GroupBuyService) validateTierGroupIDs(ctx context.Context, mapping map[string]int64) error {
	rules := exactTierMapToRules(mapping, groupBuyMaxShareCount)
	if err := validateTierRuleShape(rules, groupBuyMaxShareCount); err != nil {
		return err
	}
	return s.validateTierRules(ctx, rules)
}

func (s *GroupBuyService) validateTargetGroup(ctx context.Context, groupID int64) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return ErrGroupBuyTargetGroupInvalid.WithCause(err)
	}
	if group == nil || group.Status != StatusActive || !group.IsSubscriptionType() {
		return ErrGroupBuyTargetGroupInvalid
	}
	return nil
}

func (s *GroupBuyService) validateTargetGroupInTx(ctx context.Context, tx *dbent.Tx, groupID int64) error {
	group, err := tx.Group.Query().
		Where(dbgroup.IDEQ(groupID)).
		Only(ctx)
	if err != nil {
		return ErrGroupBuyTargetGroupInvalid.WithCause(err)
	}
	if group.Status != StatusActive || group.SubscriptionType != domain.SubscriptionTypeSubscription {
		return ErrGroupBuyTargetGroupInvalid
	}
	return nil
}

func (s *GroupBuyService) planView(ctx context.Context, p *dbent.GroupBuyPlan) GroupBuyPlanView {
	if p == nil {
		return GroupBuyPlanView{}
	}
	rules := normalizedPlanTierRules(p)
	view := GroupBuyPlanView{
		ID:                 p.ID,
		Title:              p.Title,
		Description:        psStringValue(p.Description),
		ProductKey:         p.ProductKey,
		TotalShares:        p.TotalShares,
		SeatCount:          p.TotalShares,
		PricePerShare:      p.PricePerShare,
		PricePerSeat:       p.PricePerShare,
		PriceLabel:         p.PriceLabel,
		QuotaPerShareLabel: p.QuotaPerShareLabel,
		QuotaLabel:         p.QuotaPerShareLabel,
		MaxSharesPerUser:   p.MaxSharesPerUser,
		TargetGroupID:      p.TargetGroupID,
		TierGroupIDs:       copyTierGroupIDs(p.TierGroupIds),
		ValidityDays:       p.ValidityDays,
		TimeoutMinutes:     p.TimeoutMinutes,
		LaunchMode:         p.LaunchMode,
		RefundMode:         p.RefundMode,
		AgreementText:      psStringValue(p.AgreementText),
		Status:             p.Status,
		SortOrder:          p.SortOrder,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
	if g := p.Edges.TargetGroup; g != nil {
		view.TargetGroup = groupViewFromEnt(g)
	}
	view.TierRules = s.tierRuleViews(ctx, rules)
	view.TierGroups = view.TierRules
	if len(p.Edges.Rounds) > 0 {
		view.CurrentRound = roundView(p.Edges.Rounds[0])
	}
	return view
}

func copyTierGroupIDs(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (s *GroupBuyService) tierRuleViews(ctx context.Context, rules []domain.GroupBuyTierRule) []GroupBuyTierView {
	out := make([]GroupBuyTierView, 0, len(rules))
	groupCache := map[int64]*GroupBuyGroupView{}
	for _, rule := range rules {
		groupID := rule.TargetGroupID
		tier := GroupBuyTierView{
			ShareCount:    rule.MaxShares,
			MinShares:     rule.MinShares,
			MaxShares:     rule.MaxShares,
			TargetGroupID: groupID,
			Label:         rule.Label,
		}
		if groupID > 0 {
			if cached, ok := groupCache[groupID]; ok {
				tier.TargetGroup = cached
			} else if g, err := s.entClient.Group.Get(ctx, groupID); err == nil {
				tier.TargetGroup = groupViewFromEnt(g)
				groupCache[groupID] = tier.TargetGroup
			}
		}
		out = append(out, tier)
	}
	return out
}

func groupViewFromEnt(g *dbent.Group) *GroupBuyGroupView {
	if g == nil {
		return nil
	}
	return &GroupBuyGroupView{
		ID:              g.ID,
		Name:            g.Name,
		Platform:        g.Platform,
		DailyLimitUSD:   g.DailyLimitUsd,
		WeeklyLimitUSD:  g.WeeklyLimitUsd,
		MonthlyLimitUSD: g.MonthlyLimitUsd,
	}
}

func (s *GroupBuyService) seatView(ctx context.Context, seat *dbent.GroupBuySeat) *GroupBuySeatView {
	if seat == nil {
		return nil
	}
	view := &GroupBuySeatView{
		ID:                seat.ID,
		RoundID:           seat.RoundID,
		PlanID:            seat.PlanID,
		UserID:            seat.UserID,
		OrderID:           seat.OrderID,
		Status:            seat.Status,
		ShareCount:        seat.ShareCount,
		SubscriptionID:    seat.SubscriptionID,
		BoundAPIKeyID:     seat.BoundAPIKeyID,
		LockedUntil:       seat.LockedUntil,
		PaidAt:            seat.PaidAt,
		ActivatedAt:       seat.ActivatedAt,
		ExpiresAt:         seat.ExpiresAt,
		BoundAt:           seat.BoundAt,
		RefundProcessedAt: seat.RefundProcessedAt,
		RefundNote:        seat.RefundNote,
		CreatedAt:         seat.CreatedAt,
		UpdatedAt:         seat.UpdatedAt,
	}
	if plan := seat.Edges.Plan; plan != nil {
		pv := s.planView(ctx, plan)
		view.Plan = &pv
	}
	if round := seat.Edges.Round; round != nil {
		view.Round = roundView(round)
	}
	if order := seat.Edges.Order; order != nil {
		ov := paymentOrderLite(order)
		view.Order = &ov
	}
	return view
}

func (s *GroupBuyService) entitlementView(ctx context.Context, ent *dbent.GroupBuyEntitlement) *GroupBuyEntitlementView {
	if ent == nil {
		return nil
	}
	refreshedAt := ent.RefreshedAt
	view := &GroupBuyEntitlementView{
		ID:               ent.ID,
		UserID:           ent.UserID,
		ProductKey:       ent.ProductKey,
		Status:           ent.Status,
		ActiveShareCount: ent.ActiveShareCount,
		TargetGroupID:    ent.TargetGroupID,
		SubscriptionID:   ent.SubscriptionID,
		BoundAPIKeyID:    ent.BoundAPIKeyID,
		LastActivatedAt:  ent.LastActivatedAt,
		ExpiresAt:        ent.ExpiresAt,
		RefreshedAt:      &refreshedAt,
		DeactivatedAt:    ent.DeactivatedAt,
		CreatedAt:        ent.CreatedAt,
		UpdatedAt:        ent.UpdatedAt,
	}
	if g := ent.Edges.TargetGroup; g != nil {
		view.TargetGroup = groupViewFromEnt(g)
	} else if ent.TargetGroupID != nil {
		if g, err := s.entClient.Group.Get(ctx, *ent.TargetGroupID); err == nil {
			view.TargetGroup = groupViewFromEnt(g)
		}
	}
	view.EntitlementLabel = s.entitlementTierLabel(ctx, ent)
	return view
}

func (s *GroupBuyService) entitlementTierLabel(ctx context.Context, ent *dbent.GroupBuyEntitlement) string {
	if ent == nil || ent.ActiveShareCount <= 0 {
		return ""
	}
	if plan, err := s.latestProductPlan(ctx); err == nil && plan != nil {
		if rule, ok := resolveTierRuleForShareCount(normalizedPlanTierRules(plan), ent.ActiveShareCount); ok {
			if label := strings.TrimSpace(rule.Label); label != "" {
				return label
			}
			if rule.MinShares == rule.MaxShares {
				return fmt.Sprintf("%d 份权益", rule.MinShares)
			}
			return fmt.Sprintf("%d-%d 份权益", rule.MinShares, rule.MaxShares)
		}
	}
	return fmt.Sprintf("%d 份权益", ent.ActiveShareCount)
}

func roundView(round *dbent.GroupBuyRound) *GroupBuyRoundView {
	if round == nil {
		return nil
	}
	available := round.TotalShares - round.PaidShares - round.ReservedShares
	if available < 0 {
		available = 0
	}
	return &GroupBuyRoundView{
		ID:              round.ID,
		PlanID:          round.PlanID,
		Status:          round.Status,
		TotalShares:     round.TotalShares,
		PaidShares:      round.PaidShares,
		ReservedShares:  round.ReservedShares,
		AvailableShares: available,
		TotalSeats:      round.TotalShares,
		PaidSeats:       round.PaidShares,
		ReservedSeats:   round.ReservedShares,
		AvailableSeats:  available,
		DeadlineAt:      round.DeadlineAt,
		StartedAt:       round.StartedAt,
		ClosedAt:        round.ClosedAt,
		CloseReason:     round.CloseReason,
		CreatedAt:       round.CreatedAt,
		UpdatedAt:       round.UpdatedAt,
	}
}

func paymentOrderLite(order *dbent.PaymentOrder) PaymentOrderLite {
	if order == nil {
		return PaymentOrderLite{}
	}
	return PaymentOrderLite{
		ID:          order.ID,
		Amount:      order.Amount,
		PayAmount:   order.PayAmount,
		Currency:    PaymentOrderCurrency(order),
		PaymentType: order.PaymentType,
		OutTradeNo:  order.OutTradeNo,
		Status:      order.Status,
		OrderType:   order.OrderType,
		CreatedAt:   order.CreatedAt,
		ExpiresAt:   order.ExpiresAt,
		PaidAt:      order.PaidAt,
		CompletedAt: order.CompletedAt,
	}
}

func eventView(ev *dbent.GroupBuyEvent) GroupBuyEventView {
	if ev == nil {
		return GroupBuyEventView{}
	}
	return GroupBuyEventView{
		ID:        ev.ID,
		PlanID:    ev.PlanID,
		RoundID:   ev.RoundID,
		SeatID:    ev.SeatID,
		UserID:    ev.UserID,
		EventType: ev.EventType,
		Message:   psStringValue(ev.Message),
		Metadata:  ev.Metadata,
		CreatedAt: ev.CreatedAt,
	}
}

type groupBuyEventInput struct {
	PlanID    *int64
	RoundID   *int64
	SeatID    *int64
	UserID    *int64
	EventType string
	Message   string
	Metadata  map[string]any
}

func (s *GroupBuyService) createEvent(ctx context.Context, input *groupBuyEventInput) {
	if s == nil || s.entClient == nil {
		return
	}
	s.createEventTx(ctx, s.entClient, input)
}

func (s *GroupBuyService) createEventTx(ctx context.Context, client *dbent.Client, input *groupBuyEventInput) {
	if client == nil || input == nil || strings.TrimSpace(input.EventType) == "" {
		return
	}
	b := client.GroupBuyEvent.Create().
		SetEventType(strings.TrimSpace(input.EventType)).
		SetMessage(strings.TrimSpace(input.Message))
	if input.PlanID != nil && *input.PlanID > 0 {
		b.SetPlanID(*input.PlanID)
	}
	if input.RoundID != nil && *input.RoundID > 0 {
		b.SetRoundID(*input.RoundID)
	}
	if input.SeatID != nil && *input.SeatID > 0 {
		b.SetSeatID(*input.SeatID)
	}
	if input.UserID != nil && *input.UserID > 0 {
		b.SetUserID(*input.UserID)
	}
	if input.Metadata != nil {
		b.SetMetadata(input.Metadata)
	}
	if _, err := b.Save(ctx); err != nil {
		slog.Warn("create group buy event failed", "event_type", input.EventType, "error", err)
	}
}

func normalizeGroupBuyRefundMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case GroupBuyRefundModeProviderRefund:
		return GroupBuyRefundModeProviderRefund
	default:
		return GroupBuyRefundModeBalanceCredit
	}
}

func normalizeGroupBuyPlanStatus(status string) string {
	switch strings.TrimSpace(status) {
	case GroupBuyPlanStatusDisabled:
		return GroupBuyPlanStatusDisabled
	default:
		return GroupBuyPlanStatusActive
	}
}

func normalizeGroupBuyLaunchMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case GroupBuyLaunchModeManual:
		return GroupBuyLaunchModeManual
	default:
		return GroupBuyLaunchModeAuto
	}
}

func translateGroupBuySeatCreateError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate") {
		return ErrGroupBuyInvalidStatus.WithCause(err)
	}
	return fmt.Errorf("create group buy share batch: %w", err)
}

func subscriptionNotesContainGroupBuyEntitlement(notes string) bool {
	for _, line := range strings.Split(notes, "\n") {
		normalized := strings.TrimSpace(line)
		if normalized == groupBuySubscriptionNotePrefix || strings.HasPrefix(normalized, groupBuySubscriptionNotePrefix+" ") {
			return true
		}
	}
	return false
}

func isGroupBuyManagedSubscription(sub *dbent.UserSubscription) bool {
	if sub == nil {
		return false
	}
	if sub.ManagedByGroupBuy || strings.EqualFold(strings.TrimSpace(sub.SourceType), "group_buy") {
		return true
	}
	return subscriptionNotesContainGroupBuyEntitlement(psStringValue(sub.Notes))
}

func (s *GroupBuyService) groupBuyPlanForUpdate(q *dbent.GroupBuyPlanQuery) *dbent.GroupBuyPlanQuery {
	if s.entClient != nil && s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *GroupBuyService) groupBuyRoundForUpdate(q *dbent.GroupBuyRoundQuery) *dbent.GroupBuyRoundQuery {
	if s.entClient != nil && s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *GroupBuyService) groupBuySeatForUpdate(q *dbent.GroupBuySeatQuery) *dbent.GroupBuySeatQuery {
	if s.entClient != nil && s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func psInt64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func (s *GroupBuyService) invalidateBalanceCaches(ctx context.Context, userID int64) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheService == nil {
		return
	}
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
	}()
}

func paginationResultFromTotal(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &pagination.PaginationResult{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Pages:    totalPages,
	}
}

func parseGroupBuyID(raw string, name string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", name+" is invalid")
	}
	return id, nil
}
