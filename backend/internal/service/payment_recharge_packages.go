package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const (
	monthlyRechargeBonusAudit = "MONTHLY_RECHARGE_BONUS_GRANTED"
)

type rechargePackageOrderDecision struct {
	Package         RechargePackage
	BonusAvailable  bool
	Period          string
	BonusAmount     float64
	EffectiveCredit float64
}

func (s *PaymentService) resolveRechargePackageForOrder(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig) (*RechargePackage, bool, error) {
	if req.OrderType != payment.OrderTypeBalance {
		return nil, false, nil
	}
	packageID := strings.TrimSpace(req.RechargePackageID)
	if packageID == "" {
		return nil, false, infraerrors.BadRequest("RECHARGE_PACKAGE_REQUIRED", "balance recharge requires a recharge package")
	}
	pkg, ok := findEnabledRechargePackage(cfg.RechargePackages, packageID)
	if !ok {
		return nil, false, infraerrors.BadRequest("RECHARGE_PACKAGE_NOT_AVAILABLE", "recharge package is not available").
			WithMetadata(map[string]string{"package_id": packageID})
	}
	if (cfg.MinAmount > 0 && pkg.PayAmount < cfg.MinAmount) || (cfg.MaxAmount > 0 && pkg.PayAmount > cfg.MaxAmount) {
		return nil, false, infraerrors.BadRequest("INVALID_AMOUNT", "amount out of range").
			WithMetadata(map[string]string{"min": fmt.Sprintf("%.2f", cfg.MinAmount), "max": fmt.Sprintf("%.2f", cfg.MaxAmount)})
	}
	claimed, _, err := s.MonthlyRechargeBonusClaimStatus(ctx, req.UserID)
	if err != nil {
		return nil, false, err
	}
	return pkg, !claimed, nil
}

func findEnabledRechargePackage(packages []RechargePackage, id string) (*RechargePackage, bool) {
	id = strings.TrimSpace(id)
	for _, pkg := range packages {
		if pkg.Enabled && pkg.ID == id {
			cp := pkg
			return &cp, true
		}
	}
	return nil, false
}

func (s *PaymentService) MonthlyRechargeBonusClaimStatus(ctx context.Context, userID int64) (bool, *time.Time, error) {
	if s == nil || s.entClient == nil || userID <= 0 {
		return false, nil, nil
	}
	code := monthlyRechargeBonusRedeemCode(userID, monthlyRechargeBonusPeriod(s.nowTime()))
	reward, err := s.entClient.RedeemCode.Query().Where(redeemcode.CodeEQ(code)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, reward.UsedAt, nil
}

func (s *PaymentService) RechargePackageCheckoutViews(ctx context.Context, userID int64, packages []RechargePackage) ([]RechargePackageCheckoutView, bool, string, error) {
	claimed, claimedAt, err := s.MonthlyRechargeBonusClaimStatus(ctx, userID)
	if err != nil {
		return nil, false, "", err
	}
	claimedAtStr := ""
	if claimedAt != nil {
		claimedAtStr = claimedAt.UTC().Format(time.RFC3339)
	}
	views := make([]RechargePackageCheckoutView, 0, len(packages))
	for _, pkg := range packages {
		if !pkg.Enabled {
			continue
		}
		bonus := normalizePaymentPackageAmount(pkg.CreditedAmount - pkg.PayAmount)
		effectiveCredit := pkg.PayAmount
		effectiveBonus := 0.0
		if !claimed {
			effectiveCredit = pkg.CreditedAmount
			effectiveBonus = bonus
		}
		views = append(views, RechargePackageCheckoutView{
			ID:                      pkg.ID,
			Label:                   pkg.Label,
			PayAmount:               pkg.PayAmount,
			CreditedAmount:          pkg.CreditedAmount,
			BonusAmount:             bonus,
			EffectiveCreditedAmount: normalizePaymentPackageAmount(effectiveCredit),
			EffectiveBonusAmount:    normalizePaymentPackageAmount(effectiveBonus),
			SortOrder:               pkg.SortOrder,
		})
	}
	return views, claimed, claimedAtStr, nil
}

func (s *PaymentService) adjustBalanceOrderForMonthlyRechargePackage(ctx context.Context, o *dbent.PaymentOrder) (*dbent.PaymentOrder, error) {
	if s == nil || s.entClient == nil || o == nil || o.OrderType != payment.OrderTypeBalance {
		return o, nil
	}
	decision := monthlyRechargeDecisionFromOrder(o)
	if strings.TrimSpace(decision.Package.ID) == "" {
		return o, nil
	}
	claimed, err := s.tryClaimMonthlyRechargeBonus(ctx, o, decision)
	if err != nil {
		return nil, err
	}
	if claimed {
		return o, nil
	}
	if decision.BonusAmount <= 0 || o.Amount <= o.PayAmount {
		return o, nil
	}
	updated, err := s.entClient.PaymentOrder.UpdateOneID(o.ID).
		SetAmount(o.PayAmount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("remove duplicate monthly recharge bonus: %w", err)
	}
	s.writeAuditLog(ctx, o.ID, "MONTHLY_RECHARGE_BONUS_SKIPPED", "system", map[string]any{
		"package_id":      decision.Package.ID,
		"pay_amount":      o.PayAmount,
		"credited_amount": o.PayAmount,
		"bonus_amount":    0,
		"period":          decision.Period,
		"reason":          "monthly bonus already claimed",
	})
	return updated, nil
}

func (s *PaymentService) tryClaimMonthlyRechargeBonus(ctx context.Context, o *dbent.PaymentOrder, decision rechargePackageOrderDecision) (bool, error) {
	if s == nil || s.entClient == nil || o == nil || o.UserID <= 0 {
		return false, nil
	}
	code := monthlyRechargeBonusRedeemCode(o.UserID, decision.Period)
	now := time.Now().UTC()
	detail := map[string]any{
		"package_id":      decision.Package.ID,
		"pay_amount":      decision.Package.PayAmount,
		"credited_amount": decision.Package.CreditedAmount,
		"bonus_amount":    decision.BonusAmount,
		"period":          decision.Period,
		"redeemCode":      code,
	}
	detailJSON, _ := json.Marshal(detail)
	_, err := s.entClient.RedeemCode.Create().
		SetCode(code).
		SetType(RedeemTypeFirstRechargeBonus).
		SetValue(decision.BonusAmount).
		SetStatus(StatusUsed).
		SetUsedBy(o.UserID).
		SetUsedAt(now).
		SetNotes(fmt.Sprintf("monthly recharge bonus order %d", o.ID)).
		Save(ctx)
	if err != nil {
		if isRedeemCodeConflict(err) {
			return false, nil
		}
		return false, err
	}
	if decision.BonusAmount > 0 {
		if _, err := s.entClient.PaymentAuditLog.Create().
			SetOrderID(strconv.FormatInt(o.ID, 10)).
			SetAction(welfareFirstRechargeBonusAudit).
			SetDetail(string(detailJSON)).
			SetOperator("system").
			Save(ctx); err != nil {
			return false, err
		}
	}
	if _, err := s.entClient.PaymentAuditLog.Create().
		SetOrderID(strconv.FormatInt(o.ID, 10)).
		SetAction(monthlyRechargeBonusAudit).
		SetDetail(string(detailJSON)).
		SetOperator("system").
		Save(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func monthlyRechargeDecisionFromOrder(o *dbent.PaymentOrder) rechargePackageOrderDecision {
	if o == nil {
		return rechargePackageOrderDecision{}
	}
	snapshot := o.ProviderSnapshot
	packageID := stringSnapshotValue(snapshot, "recharge_package_id")
	if packageID == "" {
		return rechargePackageOrderDecision{}
	}
	payAmount := floatSnapshotValue(snapshot, "recharge_package_pay_amount")
	creditedAmount := floatSnapshotValue(snapshot, "recharge_package_credited_amount")
	if payAmount <= 0 {
		payAmount = o.PayAmount
	}
	if creditedAmount <= 0 {
		creditedAmount = o.Amount
	}
	period := stringSnapshotValue(snapshot, "monthly_recharge_bonus_period")
	if period == "" {
		period = monthlyRechargeBonusPeriod(timezone.Now())
	}
	return rechargePackageOrderDecision{
		Package: RechargePackage{
			ID:             packageID,
			Label:          stringSnapshotValue(snapshot, "recharge_package_label"),
			Enabled:        true,
			PayAmount:      normalizePaymentPackageAmount(payAmount),
			CreditedAmount: normalizePaymentPackageAmount(creditedAmount),
			SortOrder:      int(floatSnapshotValue(snapshot, "recharge_package_sort_order")),
		},
		Period:          period,
		BonusAmount:     normalizePaymentPackageAmount(creditedAmount - payAmount),
		EffectiveCredit: normalizePaymentPackageAmount(o.Amount),
	}
}

func monthlyRechargeBonusPeriod(now time.Time) string {
	return timezone.StartOfMonth(now).Format("200601")
}

func (s *PaymentService) nowTime() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return time.Now()
}

func monthlyRechargeBonusRedeemCode(userID int64, period string) string {
	return "MRB" + strings.ReplaceAll(period, "-", "") + "U" + strings.ToUpper(strconv.FormatInt(userID, 36))
}

func stringSnapshotValue(snapshot map[string]any, key string) string {
	if snapshot == nil {
		return ""
	}
	switch value := snapshot[key].(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func floatSnapshotValue(snapshot map[string]any, key string) float64 {
	if snapshot == nil {
		return 0
	}
	switch value := snapshot[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		v, _ := value.Float64()
		return v
	default:
		return 0
	}
}
