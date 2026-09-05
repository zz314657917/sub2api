package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/caferoom"
	"github.com/Wei-Shaw/sub2api/ent/caferoundmembership"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyround"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyseat"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
)

const cafeRoomConcurrentRoomCap = 3

var (
	ErrCafeOrderUnavailable = infraerrors.Conflict("CAFE_ORDER_UNAVAILABLE", "cafe room is not available for orders")
	ErrCafeSeatUnavailable  = infraerrors.Conflict("CAFE_SEAT_UNAVAILABLE", "selected cafe seat is no longer available")
	ErrCafeSeatHeld         = infraerrors.Conflict("CAFE_USER_SEAT_EXISTS", "user already has a seat in this cafe round")
	ErrCafeShareUnavailable = infraerrors.Conflict("CAFE_SHARE_UNAVAILABLE", "requested cafe shares are no longer available")
	ErrCafeBuyerLimit       = infraerrors.Conflict("CAFE_BUYER_LIMIT_REACHED", "cafe room participant limit has been reached")
	ErrCafeUserShareLimit   = infraerrors.Conflict("CAFE_USER_SHARE_LIMIT_REACHED", "user share limit would be exceeded")
	ErrCafeRoomCapExceeded  = infraerrors.Conflict("CAFE_ROOM_LIMIT_EXCEEDED", "concurrent cafe room seat limit reached")
	ErrCafeAgreementNeeded  = infraerrors.BadRequest("CAFE_AGREEMENT_REQUIRED", "cafe purchase agreement must be accepted")
)

// CafeRoomOrderInput intentionally excludes server-owned Room, Plan, Round and account facts.
type CafeRoomOrderInput struct {
	UserID            int64
	RoomID            int64
	ShareCount        int
	PaymentType       string
	OpenID            string
	ClientIP          string
	IsMobile          bool
	IsWeChatBrowser   bool
	SrcHost           string
	SrcURL            string
	ReturnURL         string
	PaymentSource     string
	AgreementAccepted bool
}

type CafeRoomOrderResponse struct {
	CreateOrderResponse
	RoomID     int64 `json:"room_id"`
	RoundID    int64 `json:"round_id"`
	ShareCount int   `json:"share_count"`
}

type CafeRoomReservationInput struct {
	UserID            int64
	RoomID            int64
	ShareCount        int
	AgreementAccepted bool
}

type CafeRoomReservationResponse struct {
	RoomID       int64  `json:"room_id"`
	RoundID      int64  `json:"round_id"`
	ReservationID int64 `json:"reservation_id"`
	ShareCount   int    `json:"share_count"`
	Status       string `json:"status"`
	TotalShares  int    `json:"total_shares"`
	ReservedShares int  `json:"reserved_shares"`
}

// CafeRoomOrderService owns the Room-specific order and seat-lock transaction.
// It delegates payment selection and provider invocation to PaymentService.
type CafeRoomOrderService struct {
	entClient   *dbent.Client
	paymentSvc  *PaymentService
	groupBuySvc *GroupBuyService
	settings    CafePublicSettings
	now         func() time.Time
}

func NewCafeRoomOrderService(entClient *dbent.Client, paymentSvc *PaymentService, groupBuySvc *GroupBuyService, settings CafePublicSettings) *CafeRoomOrderService {
	return &CafeRoomOrderService{
		entClient:   entClient,
		paymentSvc:  paymentSvc,
		groupBuySvc: groupBuySvc,
		settings:    settings,
		now:         time.Now,
	}
}

func (s *CafeRoomOrderService) CreateOrder(ctx context.Context, input CafeRoomOrderInput) (*CafeRoomOrderResponse, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if input.UserID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	if input.RoomID <= 0 || input.ShareCount <= 0 || strings.TrimSpace(input.PaymentType) == "" {
		return nil, infraerrors.BadRequest("CAFE_INVALID_ORDER", "room, share count and payment type are required")
	}
	if !input.AgreementAccepted {
		return nil, ErrCafeAgreementNeeded
	}
	if s.paymentSvc == nil || s.paymentSvc.configService == nil || s.paymentSvc.userRepo == nil || s.groupBuySvc == nil {
		return nil, infraerrors.InternalServer("CAFE_ORDER_SERVICE_UNAVAILABLE", "cafe room order service is unavailable")
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
	}
	if normalized := NormalizeVisibleMethod(req.PaymentType); normalized != "" {
		req.PaymentType = normalized
	}

	room, plan, err := s.loadAvailableRoom(ctx, input.RoomID)
	if err != nil {
		return nil, err
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
	if _, err := s.paymentSvc.userRepo.GetByID(ctx, input.UserID); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	orderAmount := plan.PricePerShare * float64(input.ShareCount)
	req.Amount = orderAmount
	methodCurrency, err := s.paymentSvc.configService.ValidateMethodCurrencyConsistency(ctx, req.PaymentType)
	if err != nil {
		return nil, err
	}
	payAmountStr, payAmount, err := calculateCreateOrderPayAmount(orderAmount, cfg.RechargeFeeRate, methodCurrency)
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
		payAmountStr, payAmount, err = calculateCreateOrderPayAmount(orderAmount, cfg.RechargeFeeRate, selectedCurrency)
		if err != nil {
			return nil, err
		}
	}
	if err := validateSelectedCreateOrderAmountCurrency(payAmountStr, sel); err != nil {
		return nil, err
	}
	if oauthResp, err := s.paymentSvc.maybeBuildWeChatOAuthRequiredResponseForSelection(ctx, req, orderAmount, payAmount, cfg.RechargeFeeRate, sel); err != nil {
		return nil, err
	} else if oauthResp != nil {
		return &CafeRoomOrderResponse{CreateOrderResponse: *oauthResp, RoomID: room.ID}, nil
	}

	order, round, err := s.lockSeatAndCreateOrder(ctx, req, room.ID, input.ShareCount, cfg, cfg.RechargeFeeRate, payAmount, sel)
	if err != nil {
		return nil, err
	}
	// The plan is read once before payment selection and once under the room/
	// plan lock. If an admin changed price or fee inputs in between, refuse to
	// call the provider with a stale amount rather than creating a mismatched
	// order that can never pass webhook amount validation.
	if math.Abs(order.Amount-orderAmount) > amountToleranceCNY || math.Abs(order.PayAmount-payAmount) > amountToleranceCNY {
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).Save(ctx)
		if releaseErr := s.groupBuySvc.ReleaseGroupBuySeatForOrder(context.Background(), order.ID, "cafe plan changed during order creation"); releaseErr != nil {
			slog.Error("release cafe seat after plan change failed", "order_id", order.ID, "error", releaseErr)
		}
		return nil, infraerrors.Conflict("CAFE_ORDER_CHANGED", "cafe room pricing changed, please retry")
	}
	result, err := s.paymentSvc.invokeProvider(ctx, order, req, cfg, orderAmount, payAmountStr, payAmount, nil, sel)
	if err != nil {
		if errors.Is(err, ErrPaymentProviderResponsePersist) {
			// The provider may already have created a payable order. Keep the
			// local order and share lock pending for webhook/expiry reconciliation.
			return nil, err
		}
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).Save(ctx)
		if releaseErr := s.groupBuySvc.ReleaseGroupBuySeatForOrder(context.Background(), order.ID, "provider create failed"); releaseErr != nil {
			slog.Error("release cafe seat after provider create failure failed", "order_id", order.ID, "error", releaseErr)
		}
		return nil, err
	}
	return &CafeRoomOrderResponse{CreateOrderResponse: *result, RoomID: room.ID, RoundID: round.ID, ShareCount: input.ShareCount}, nil
}

// ReserveShares records a temporary, unpaid reservation. It deliberately does
// not create a payment order or call a provider; payment begins only after the
// round reaches awaiting_payment.
func (s *CafeRoomOrderService) ReserveShares(ctx context.Context, input CafeRoomReservationInput) (*CafeRoomReservationResponse, error) {
	if err := s.requireEnabled(ctx); err != nil { return nil, err }
	if input.UserID <= 0 || input.RoomID <= 0 || input.ShareCount <= 0 { return nil, infraerrors.BadRequest("CAFE_INVALID_RESERVATION", "room and share count are required") }
	if !input.AgreementAccepted { return nil, ErrCafeAgreementNeeded }
	tx, err := s.entClient.Tx(ctx); if err != nil { return nil, fmt.Errorf("begin cafe reservation transaction: %w", err) }
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	room, err := s.cafeRoomForUpdate(tx.CafeRoom.Query().Where(caferoom.IDEQ(input.RoomID), caferoom.DeletedAtIsNil())).Only(txCtx)
	if err != nil { if dbent.IsNotFound(err) { return nil, ErrCafeRoomNotFound }; return nil, err }
	if room.Status != CafeRoomStatusEnabled { return nil, ErrCafeOrderUnavailable }
		plan, err := s.groupBuySvc.groupBuyPlanForUpdate(tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(room.PlanID), groupbuyplan.DeletedAtIsNil()).WithTargetGroup()).Only(txCtx)
	if err != nil { return nil, ErrCafePlanNotFound }
	if !isCafeOperationalPlanEntity(plan) { return nil, ErrCafeOrderUnavailable }
	round, err := s.groupBuySvc.groupBuyRoundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.CafeRoomIDEQ(room.ID), groupbuyround.StatusIn(CafeRoundStatusOpen, CafeRoundStatusReserving), groupbuyround.DeadlineAtGT(s.now()))).Only(txCtx)
	if err != nil { if dbent.IsNotFound(err) { return nil, ErrCafeRoundNotOpen }; return nil, err }
	if round.PlanID != plan.ID || round.CafeFulfillmentVersion != "membership_share" { return nil, ErrCafeOrderUnavailable }
	if round.TotalShares-round.PaidShares-round.ReservedShares < input.ShareCount { return nil, ErrCafeShareUnavailable }
	membership, mErr := tx.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.UserIDEQ(input.UserID)).Only(txCtx)
	if mErr != nil && !dbent.IsNotFound(mErr) { return nil, mErr }
	if dbent.IsNotFound(mErr) {
		buyers, err := tx.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.Or(caferoundmembership.PaidSharesGT(0), caferoundmembership.ReservedSharesGT(0))).Count(txCtx)
		if err != nil { return nil, err }
		if round.MaxBuyers != nil && buyers >= *round.MaxBuyers { return nil, ErrCafeBuyerLimit }
		membership, err = tx.CafeRoundMembership.Create().SetRoundID(round.ID).SetUserID(input.UserID).SetStatus(GroupBuySeatStatusLocked).Save(txCtx)
		if err != nil { return nil, err }
	}
	maxPerUser := round.TotalShares; if round.MaxSharesPerUser != nil { maxPerUser = *round.MaxSharesPerUser }
	if membership.PaidShares+membership.ReservedShares+input.ShareCount > maxPerUser { return nil, ErrCafeUserShareLimit }
		seat, seatErr := s.groupBuySvc.groupBuySeatForUpdate(tx.GroupBuySeat.Query().Where(groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.UserIDEQ(input.UserID), groupbuyseat.StatusEQ(GroupBuySeatStatusLocked), groupbuyseat.OrderIDIsNil())).Only(txCtx)
	if dbent.IsNotFound(seatErr) {
		seat, seatErr = tx.GroupBuySeat.Create().SetRoundID(round.ID).SetPlanID(plan.ID).SetUserID(input.UserID).SetStatus(GroupBuySeatStatusLocked).SetMembershipID(membership.ID).SetShareCount(input.ShareCount).SetPolicySnapshot(buildGroupBuyPolicySnapshot(plan, s.now())).SetLockedUntil(s.now().Add(2*time.Hour)).Save(txCtx)
	} else if seatErr == nil {
		seat, seatErr = tx.GroupBuySeat.UpdateOneID(seat.ID).SetShareCount(seat.ShareCount + input.ShareCount).SetLockedUntil(s.now().Add(2*time.Hour)).SetUpdatedAt(s.now()).Save(txCtx)
	}
	if seatErr != nil { return nil, translateCafeSeatCreateError(seatErr) }
	if _, err = tx.CafeRoundMembership.UpdateOneID(membership.ID).AddReservedShares(input.ShareCount).SetUpdatedAt(s.now()).Save(txCtx); err != nil { return nil, err }
	status := round.Status; reserved := round.ReservedShares + input.ShareCount
	update := tx.GroupBuyRound.UpdateOneID(round.ID).AddReservedShares(input.ShareCount).AddReservedSeats(input.ShareCount).SetUpdatedAt(s.now())
	if status == CafeRoundStatusOpen { status = CafeRoundStatusReserving; update.SetStatus(CafeRoundStatusReserving) }
	if round.PaidShares+reserved >= round.TotalShares { status = CafeRoundStatusAwaitingPayment; update.SetStatus(CafeRoundStatusAwaitingPayment) }
	round, err = update.Save(txCtx); if err != nil { return nil, err }
	s.groupBuySvc.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{PlanID: &plan.ID, RoundID: &round.ID, SeatID: &seat.ID, UserID: &input.UserID, EventType: groupBuyEventSharesLocked, Message: "用户预约像素网吧份额", Metadata: map[string]any{"share_count": input.ShareCount, "reservation": true}})
	if err := tx.Commit(); err != nil { return nil, err }
	if s.paymentSvc != nil && s.paymentSvc.systemTicketSvc != nil {
		members, listErr := s.entClient.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID)).All(ctx)
		if listErr == nil {
			for _, member := range members {
				event := NewCafeReservationChangedSystemTicketNotification(member.UserID, round.ID, status, map[string]any{
					"room_id": room.ID, "reserved_shares": round.ReservedShares, "total_shares": round.TotalShares,
				})
				s.paymentSvc.systemTicketSvc.NotifyEventBestEffort(ctx, "service.cafe", member.UserID, event)
			}
		}
	}
	return &CafeRoomReservationResponse{RoomID: room.ID, RoundID: round.ID, ReservationID: seat.ID, ShareCount: input.ShareCount, Status: status, TotalShares: round.TotalShares, ReservedShares: round.ReservedShares}, nil
}

func (s *CafeRoomOrderService) lockSeatAndCreateOrder(ctx context.Context, req CreateOrderRequest, roomID int64, shareCount int, cfg *PaymentConfig, feeRate, payAmount float64, sel *payment.InstanceSelection) (*dbent.PaymentOrder, *dbent.GroupBuyRound, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin cafe room order transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	roomQuery := tx.CafeRoom.Query().Where(caferoom.IDEQ(roomID), caferoom.DeletedAtIsNil())
	lockedRoom, err := s.cafeRoomForUpdate(roomQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrCafeRoomNotFound
		}
		return nil, nil, fmt.Errorf("lock cafe room: %w", err)
	}
	if lockedRoom.Status != CafeRoomStatusEnabled {
		return nil, nil, ErrCafeOrderUnavailable
	}
	planQuery := tx.GroupBuyPlan.Query().Where(groupbuyplan.IDEQ(lockedRoom.PlanID), groupbuyplan.DeletedAtIsNil()).WithTargetGroup()
	lockedPlan, err := s.groupBuySvc.groupBuyPlanForUpdate(planQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrCafePlanNotFound
		}
		return nil, nil, fmt.Errorf("lock cafe room plan: %w", err)
	}
	if !isCafeOperationalPlanEntity(lockedPlan) {
		return nil, nil, ErrCafeOrderUnavailable
	}
	if err := s.paymentSvc.checkPendingLimit(txCtx, tx, req.UserID, cfg.MaxPendingOrders); err != nil {
		return nil, nil, err
	}
	if err := s.paymentSvc.checkDailyLimit(txCtx, tx, req.UserID, lockedPlan.PricePerShare*float64(shareCount), cfg.DailyLimit); err != nil {
		return nil, nil, err
	}
	activeUser, err := s.cafeUserForUpdate(tx.User.Query().Where(user.IDEQ(req.UserID), user.StatusEQ(StatusActive))).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
		}
		return nil, nil, fmt.Errorf("lock cafe order user: %w", err)
	}
	roundQuery := tx.GroupBuyRound.Query().Where(
		groupbuyround.CafeRoomIDEQ(lockedRoom.ID),
		groupbuyround.StatusIn(CafeRoundStatusOpen, CafeRoundStatusAwaitingPayment),
		groupbuyround.DeadlineAtGT(s.now()),
	)
	round, err := s.groupBuySvc.groupBuyRoundForUpdate(roundQuery).Only(txCtx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrCafeOrderUnavailable
		}
		return nil, nil, fmt.Errorf("lock cafe room round: %w", err)
	}
	if round.PlanID != lockedPlan.ID || round.CafeFulfillmentVersion != "membership_share" {
		return nil, nil, ErrCafeOrderUnavailable
	}
	if _, err := s.groupBuySvc.releaseExpiredLockedSeatsForRoundTx(txCtx, tx, round.ID); err != nil {
		return nil, nil, err
	}
	round, err = s.groupBuySvc.groupBuyRoundForUpdate(tx.GroupBuyRound.Query().Where(groupbuyround.IDEQ(round.ID))).Only(txCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("reload cafe room round: %w", err)
	}
	var reservationSeat *dbent.GroupBuySeat
	if round.Status == CafeRoundStatusAwaitingPayment {
		reservationSeat, err = s.groupBuySvc.groupBuySeatForUpdate(tx.GroupBuySeat.Query().Where(
			groupbuyseat.RoundIDEQ(round.ID), groupbuyseat.UserIDEQ(req.UserID),
			groupbuyseat.StatusEQ(GroupBuySeatStatusLocked), groupbuyseat.OrderIDIsNil(),
			groupbuyseat.ShareCountEQ(shareCount),
		)).Only(txCtx)
		if err != nil {
			if dbent.IsNotFound(err) { return nil, nil, ErrCafeOrderUnavailable }
			return nil, nil, fmt.Errorf("lock cafe reservation: %w", err)
		}
	} else if shareCount <= 0 || shareCount > round.TotalShares || round.PaidShares+round.ReservedShares+shareCount > round.TotalShares {
		return nil, nil, ErrCafeShareUnavailable
	}
	if reservationSeat == nil {
		if err := s.checkConcurrentRoomCapTx(txCtx, tx, req.UserID, lockedRoom.ID); err != nil {
		return nil, nil, err
		}
	}
	membershipQuery := tx.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.UserIDEQ(req.UserID))
	if s.entClient.Driver().Dialect() != dialect.SQLite {
		membershipQuery = membershipQuery.ForUpdate()
	}
	membership, membershipErr := membershipQuery.Only(txCtx)
	if reservationSeat != nil {
		membership, membershipErr = tx.CafeRoundMembership.Query().Where(caferoundmembership.IDEQ(*reservationSeat.MembershipID)).Only(txCtx)
	}
	if membershipErr != nil && !dbent.IsNotFound(membershipErr) {
		return nil, nil, fmt.Errorf("lock cafe membership: %w", membershipErr)
	}
	needsBuyerSlot := reservationSeat == nil && (dbent.IsNotFound(membershipErr) || membership.PaidShares == 0 && membership.ReservedShares == 0)
	if needsBuyerSlot {
		buyers, err := tx.CafeRoundMembership.Query().Where(caferoundmembership.RoundIDEQ(round.ID), caferoundmembership.Or(caferoundmembership.PaidSharesGT(0), caferoundmembership.ReservedSharesGT(0))).Count(txCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("count cafe buyers: %w", err)
		}
		if round.MaxBuyers == nil || buyers >= *round.MaxBuyers {
			return nil, nil, ErrCafeBuyerLimit
		}
	}
	if reservationSeat == nil && dbent.IsNotFound(membershipErr) {
		membership, err = tx.CafeRoundMembership.Create().SetRoundID(round.ID).SetUserID(req.UserID).SetStatus(GroupBuySeatStatusLocked).Save(txCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("create cafe membership: %w", err)
		}
	}
	maxPerUser := round.TotalShares
	if round.MaxSharesPerUser != nil {
		maxPerUser = *round.MaxSharesPerUser
	}
	if reservationSeat == nil && membership.PaidShares+membership.ReservedShares+shareCount > maxPerUser {
		return nil, nil, ErrCafeUserShareLimit
	}

	outTradeNo, err := s.paymentSvc.allocateOutTradeNo(txCtx, tx)
	if err != nil {
		return nil, nil, err
	}
	timeoutMin := cfg.OrderTimeoutMin
	if timeoutMin <= 0 {
		timeoutMin = defaultOrderTimeoutMin
	}
	expiresAt := s.now().Add(time.Duration(timeoutMin) * time.Minute)
	providerSnapshot := buildPaymentOrderProviderSnapshot(sel, req)
	orderBuilder := tx.PaymentOrder.Create().
		SetUserID(req.UserID).
		SetUserEmail(activeUser.Email).
		SetUserName(activeUser.Username).
		SetNillableUserNotes(psNilIfEmpty(activeUser.Notes)).
		SetAmount(lockedPlan.PricePerShare * float64(shareCount)).
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
	if sel != nil {
		if instanceID := strings.TrimSpace(sel.InstanceID); instanceID != "" {
			orderBuilder.SetProviderInstanceID(instanceID)
		}
		if providerKey := strings.TrimSpace(sel.ProviderKey); providerKey != "" {
			orderBuilder.SetProviderKey(providerKey)
		}
	}
	if providerSnapshot != nil {
		orderBuilder.SetProviderSnapshot(providerSnapshot)
	}
	order, err := orderBuilder.Save(txCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("create cafe room payment order: %w", err)
	}
	order, err = tx.PaymentOrder.UpdateOneID(order.ID).SetRechargeCode(fmt.Sprintf("CAFE-%d-%d", order.ID, s.now().UnixNano()%100000)).Save(txCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("set cafe room payment code: %w", err)
	}
	seat := reservationSeat
	if seat == nil {
		seat, err = tx.GroupBuySeat.Create().
		SetRoundID(round.ID).
		SetPlanID(lockedPlan.ID).
		SetUserID(req.UserID).
		SetOrderID(order.ID).
		SetStatus(GroupBuySeatStatusLocked).
		SetMembershipID(membership.ID).
		SetShareCount(shareCount).
		SetPolicySnapshot(buildGroupBuyPolicySnapshot(lockedPlan, s.now())).
		SetLockedUntil(expiresAt).
		Save(txCtx)
		if err != nil {
			return nil, nil, translateCafeSeatCreateError(err)
		}
	} else {
		seat, err = tx.GroupBuySeat.UpdateOneID(seat.ID).SetOrderID(order.ID).SetLockedUntil(expiresAt).SetUpdatedAt(s.now()).Save(txCtx)
		if err != nil { return nil, nil, fmt.Errorf("attach payment to cafe reservation: %w", err) }
	}
	if reservationSeat == nil {
		round, err = tx.GroupBuyRound.UpdateOneID(round.ID).
		AddReservedShares(shareCount).
		AddReservedSeats(shareCount).
		SetUpdatedAt(s.now()).
		Save(txCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("reserve cafe room seat: %w", err)
		}
	}
	s.groupBuySvc.createEventTx(txCtx, tx.Client(), &groupBuyEventInput{
		PlanID:    &lockedPlan.ID,
		RoundID:   &round.ID,
		SeatID:    &seat.ID,
		UserID:    &req.UserID,
		EventType: groupBuyEventSharesLocked,
		Message:   "用户锁定像素网吧席位",
		Metadata:  map[string]any{"order_id": order.ID, "share_count": shareCount, "amount": order.Amount},
	})
	if reservationSeat == nil {
		if _, err := tx.CafeRoundMembership.UpdateOneID(membership.ID).AddReservedShares(shareCount).SetUpdatedAt(s.now()).Save(txCtx); err != nil {
		return nil, nil, fmt.Errorf("reserve cafe membership shares: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit cafe room order transaction: %w", err)
	}
	return order, round, nil
}

func (s *CafeRoomOrderService) checkConcurrentRoomCapTx(ctx context.Context, tx *dbent.Tx, userID, roomID int64) error {
	seats, err := s.groupBuySvc.groupBuySeatForUpdate(tx.GroupBuySeat.Query().Where(
		groupbuyseat.UserIDEQ(userID),
		groupbuyseat.StatusIn(GroupBuySeatStatusLocked, GroupBuySeatStatusPaid, GroupBuySeatStatusActive),
	).WithRound()).All(ctx)
	if err != nil {
		return fmt.Errorf("list current cafe room seats: %w", err)
	}
	roomIDs := make(map[int64]struct{})
	now := s.now()
	for _, seat := range seats {
		if seat.Status == GroupBuySeatStatusLocked && seat.LockedUntil != nil && !seat.LockedUntil.After(now) {
			continue
		}
		round := seat.Edges.Round
		if round == nil || round.CafeRoomID == nil {
			continue
		}
		roomIDs[*round.CafeRoomID] = struct{}{}
	}
	if _, alreadyPresent := roomIDs[roomID]; !alreadyPresent && len(roomIDs) >= cafeRoomConcurrentRoomCap {
		return ErrCafeRoomCapExceeded
	}
	return nil
}

func (s *CafeRoomOrderService) loadAvailableRoom(ctx context.Context, roomID int64) (*dbent.CafeRoom, *dbent.GroupBuyPlan, error) {
	if s == nil || s.entClient == nil {
		return nil, nil, infraerrors.InternalServer("CAFE_ORDER_SERVICE_UNAVAILABLE", "cafe room order service is unavailable")
	}
	room, err := s.entClient.CafeRoom.Query().Where(caferoom.IDEQ(roomID), caferoom.DeletedAtIsNil()).WithPlan(func(planQuery *dbent.GroupBuyPlanQuery) {
		planQuery.WithTargetGroup()
	}).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrCafeRoomNotFound
		}
		return nil, nil, fmt.Errorf("load cafe room: %w", err)
	}
	if room.Status != CafeRoomStatusEnabled || room.Edges.Plan == nil || !isCafeOperationalPlanEntity(room.Edges.Plan) || room.Edges.Plan.Edges.TargetGroup == nil {
		return nil, nil, ErrCafeOrderUnavailable
	}
	return room, room.Edges.Plan, nil
}

func isCafeOperationalPlanEntity(plan *dbent.GroupBuyPlan) bool {
	return plan != nil && plan.Status == GroupBuyPlanStatusActive && plan.FulfillmentMode == CafeRoomFulfillmentMode && plan.AutoCreateRoomKey && plan.ValidityDays > 0 && plan.Edges.TargetGroup != nil && plan.Edges.TargetGroup.Status == StatusActive && plan.Edges.TargetGroup.AccessMode == CafeRoomGroupAccessMode
}

func cafeAccountBelongsToGroup(item *dbent.Account, groupID int64) bool {
	if item == nil {
		return false
	}
	for _, group := range item.Edges.Groups {
		if group.ID == groupID {
			return true
		}
	}
	return false
}

func (s *CafeRoomOrderService) requireEnabled(ctx context.Context) error {
	if s == nil || s.settings == nil {
		return ErrCafePublicUnavailable
	}
	settings, err := s.settings.GetPublicSettings(ctx)
	if err != nil {
		return fmt.Errorf("load public cafe settings: %w", err)
	}
	if settings == nil || !settings.PixelCafeEnabled {
		return ErrCafeDisabled
	}
	return nil
}

func (s *CafeRoomOrderService) cafeRoomForUpdate(q *dbent.CafeRoomQuery) *dbent.CafeRoomQuery {
	if s.entClient != nil && s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func (s *CafeRoomOrderService) cafeUserForUpdate(q *dbent.UserQuery) *dbent.UserQuery {
	if s.entClient != nil && s.entClient.Driver().Dialect() != dialect.SQLite {
		return q.ForUpdate()
	}
	return q
}

func cafeRoundSeatCount(round *dbent.GroupBuyRound) int {
	if round == nil {
		return 0
	}
	if round.TotalSeats > 0 {
		return round.TotalSeats
	}
	return round.TotalShares
}

func translateCafeSeatCreateError(err error) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return ErrCafeSeatUnavailable.WithCause(err)
	}
	return fmt.Errorf("create cafe room seat: %w", err)
}
