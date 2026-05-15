package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	apptimezone "github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

const (
	MembershipTierNormal = "normal"
	MembershipTierVIP    = "vip"
	MembershipTierSVIP   = "svip"

	MembershipGrantSourceAutoMonthlySpend = "auto_monthly_spend"

	MembershipGrantStatusActive  = "active"
	MembershipGrantStatusRevoked = "revoked"

	SettingKeyMembershipTiersConfig = "membership_tiers_config"

	defaultMembershipValidityDays = 30
	membershipEffectiveCacheTTL   = 30 * time.Second
)

type MembershipTierConfig struct {
	Level               string  `json:"level"`
	Label               string  `json:"label"`
	ThresholdAmount     float64 `json:"threshold_amount"`
	RateMultiplier      float64 `json:"rate_multiplier"`
	RPMLimit            int     `json:"rpm_limit"`
	TPMLimit            int     `json:"tpm_limit"`
	ImageActiveTasks    int     `json:"image_active_tasks"`
	SubscriptionGroupID int64   `json:"subscription_group_id"`
}

type MembershipSettings struct {
	Enabled      bool                   `json:"enabled"`
	ValidityDays int                    `json:"validity_days"`
	Tiers        []MembershipTierConfig `json:"tiers"`
}

type MembershipGrant struct {
	ID                  int64      `json:"id"`
	UserID              int64      `json:"user_id"`
	Tier                string     `json:"tier"`
	Source              string     `json:"source"`
	PeriodKey           string     `json:"period_key"`
	PeriodStart         time.Time  `json:"period_start"`
	PeriodEnd           time.Time  `json:"period_end"`
	QualifiedAmount     float64    `json:"qualified_amount"`
	StartsAt            time.Time  `json:"starts_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	Status              string     `json:"status"`
	SubscriptionID      *int64     `json:"subscription_id,omitempty"`
	SubscriptionGroupID *int64     `json:"subscription_group_id,omitempty"`
	SourceOrderID       *int64     `json:"source_order_id,omitempty"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty"`
	RevokeReason        *string    `json:"revoke_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type MembershipStatus struct {
	Enabled          bool                   `json:"enabled"`
	CurrentTier      string                 `json:"current_tier"`
	CurrentTierLabel string                 `json:"current_tier_label"`
	Benefits         MembershipTierConfig   `json:"benefits"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	CurrentMonthPaid float64                `json:"current_month_paid"`
	MonthPeriodStart time.Time              `json:"month_period_start"`
	MonthPeriodEnd   time.Time              `json:"month_period_end"`
	NextTier         *MembershipTierConfig  `json:"next_tier,omitempty"`
	AmountToNext     float64                `json:"amount_to_next"`
	Tiers            []MembershipTierConfig `json:"tiers"`
	Grant            *MembershipGrant       `json:"grant,omitempty"`
}

type MembershipRepository interface {
	GetMonthlyNetPaid(ctx context.Context, userID int64, start, end time.Time) (float64, error)
	UpsertAutoGrant(ctx context.Context, grant *MembershipGrant) (*MembershipGrant, bool, error)
	UpdateGrantSubscription(ctx context.Context, grantID, subscriptionID, groupID int64) error
	ListActiveAutoGrants(ctx context.Context, userID int64, now time.Time) ([]MembershipGrant, error)
	GetActiveHighestAutoGrant(ctx context.Context, userID int64, now time.Time) (*MembershipGrant, error)
	RevokeGrant(ctx context.Context, grantID int64, reason string) error
}

type MembershipBenefitResolver interface {
	GetEffectiveBenefits(ctx context.Context, userID int64) (MembershipTierConfig, error)
}

type membershipEffectiveCacheEntry struct {
	expiresAt time.Time
	benefits  MembershipTierConfig
}

type MembershipService struct {
	repo                 MembershipRepository
	settingRepo          SettingRepository
	subscriptionSvc      *SubscriptionService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	billingCacheService  *BillingCacheService

	cache sync.Map // userID -> membershipEffectiveCacheEntry
}

func NewMembershipService(repo MembershipRepository, settingRepo SettingRepository, subscriptionSvc *SubscriptionService, authCacheInvalidator APIKeyAuthCacheInvalidator, billingCacheService *BillingCacheService) *MembershipService {
	svc := &MembershipService{
		repo:                 repo,
		settingRepo:          settingRepo,
		subscriptionSvc:      subscriptionSvc,
		authCacheInvalidator: authCacheInvalidator,
		billingCacheService:  billingCacheService,
	}
	if billingCacheService != nil {
		billingCacheService.SetMembershipBenefitResolver(svc)
	}
	return svc
}

func DefaultMembershipSettings() MembershipSettings {
	return normalizeMembershipSettings(MembershipSettings{
		Enabled:      false,
		ValidityDays: defaultMembershipValidityDays,
		Tiers: []MembershipTierConfig{
			{
				Level:            MembershipTierNormal,
				Label:            "普通用户",
				ThresholdAmount:  0,
				RateMultiplier:   1,
				RPMLimit:         60,
				TPMLimit:         60000,
				ImageActiveTasks: 2,
			},
			{
				Level:            MembershipTierVIP,
				Label:            "VIP",
				ThresholdAmount:  50,
				RateMultiplier:   0.8,
				RPMLimit:         300,
				TPMLimit:         300000,
				ImageActiveTasks: 5,
			},
			{
				Level:            MembershipTierSVIP,
				Label:            "SVIP",
				ThresholdAmount:  100,
				RateMultiplier:   0.6,
				RPMLimit:         600,
				TPMLimit:         600000,
				ImageActiveTasks: 10,
			},
		},
	})
}

func (s *MembershipService) GetSettings(ctx context.Context) (MembershipSettings, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultMembershipSettings(), nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyMembershipTiersConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultMembershipSettings(), nil
		}
		return MembershipSettings{}, err
	}
	var settings MembershipSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return MembershipSettings{}, fmt.Errorf("parse membership settings: %w", err)
	}
	return normalizeMembershipSettings(settings), nil
}

func (s *MembershipService) UpdateSettings(ctx context.Context, settings MembershipSettings) (MembershipSettings, error) {
	if s == nil || s.settingRepo == nil {
		return MembershipSettings{}, fmt.Errorf("membership service unavailable")
	}
	normalized := normalizeMembershipSettings(settings)
	if err := validateMembershipSettings(normalized); err != nil {
		return MembershipSettings{}, err
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return MembershipSettings{}, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyMembershipTiersConfig, string(data)); err != nil {
		return MembershipSettings{}, err
	}
	s.invalidateAll()
	return normalized, nil
}

func (s *MembershipService) GetStatus(ctx context.Context, userID int64) (*MembershipStatus, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	now := apptimezone.Now()
	periodStart, periodEnd := membershipCurrentMonth(now)
	netPaid := 0.0
	if s.repo != nil {
		netPaid, err = s.repo.GetMonthlyNetPaid(ctx, userID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}
	}
	grant, err := s.activeHighestGrant(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	currentTier := s.effectiveLevelFromGrantAndSubscriptions(ctx, settings, userID, grant)
	benefits := settings.tierByLevel(currentTier)
	if benefits.Level == "" {
		benefits = settings.tierByLevel(MembershipTierNormal)
		currentTier = benefits.Level
	}
	var expiresAt *time.Time
	if grant != nil {
		v := grant.ExpiresAt
		expiresAt = &v
	}
	next := settings.nextTierByAmount(netPaid)
	amountToNext := 0.0
	if next != nil {
		amountToNext = roundMoney(math.Max(next.ThresholdAmount-netPaid, 0))
	}
	return &MembershipStatus{
		Enabled:          settings.Enabled,
		CurrentTier:      currentTier,
		CurrentTierLabel: benefits.Label,
		Benefits:         benefits,
		ExpiresAt:        expiresAt,
		CurrentMonthPaid: roundMoney(netPaid),
		MonthPeriodStart: periodStart,
		MonthPeriodEnd:   periodEnd,
		NextTier:         next,
		AmountToNext:     amountToNext,
		Tiers:            settings.Tiers,
		Grant:            grant,
	}, nil
}

func (s *MembershipService) RecalculateForUser(ctx context.Context, userID int64, sourceOrderID int64, reason string) error {
	return s.RecalculateForUserAt(ctx, userID, sourceOrderID, reason, apptimezone.Now())
}

func (s *MembershipService) RecalculateForUserAt(ctx context.Context, userID int64, sourceOrderID int64, reason string, periodAt time.Time) error {
	if s == nil || s.repo == nil || userID <= 0 {
		return nil
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	now := apptimezone.Now()
	if periodAt.IsZero() {
		periodAt = now
	}
	periodStart, periodEnd := membershipCurrentMonth(periodAt)
	periodKey := periodStart.Format("2006-01")

	netPaid, err := s.repo.GetMonthlyNetPaid(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return err
	}
	target := settings.qualifiedPaidTier(netPaid)
	if !settings.Enabled {
		target = settings.tierByLevel(MembershipTierNormal)
	}
	if membershipTierRank(target.Level) > membershipTierRank(MembershipTierNormal) && target.SubscriptionGroupID <= 0 {
		slog.Warn("membership tier is not bound to subscription group, skipping auto grant",
			"user_id", userID,
			"tier", target.Level,
			"period_key", periodKey,
		)
		target = settings.tierByLevel(MembershipTierNormal)
	}
	if err := s.revokeUnqualifiedCurrentPeriodGrants(ctx, userID, periodKey, netPaid, target.Level, reason); err != nil {
		return err
	}
	if settings.Enabled && membershipTierRank(target.Level) > membershipTierRank(MembershipTierNormal) {
		if err := s.ensureAutoGrant(ctx, userID, sourceOrderID, periodKey, periodStart, periodEnd, netPaid, target); err != nil {
			return err
		}
	}
	s.invalidateUser(userID)
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	return nil
}

func (s *MembershipService) GetEffectiveBenefits(ctx context.Context, userID int64) (MembershipTierConfig, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return MembershipTierConfig{}, err
	}
	if !settings.Enabled {
		benefits := settings.tierByLevel(MembershipTierNormal)
		benefits.RateMultiplier = 0
		benefits.RPMLimit = 0
		benefits.TPMLimit = 0
		return benefits, nil
	}
	now := time.Now()
	if cached, ok := s.cache.Load(userID); ok {
		entry, _ := cached.(membershipEffectiveCacheEntry)
		if entry.expiresAt.After(now) && entry.benefits.Level != "" {
			return entry.benefits, nil
		}
	}
	grant, err := s.activeHighestGrant(ctx, userID, apptimezone.Now())
	if err != nil {
		return MembershipTierConfig{}, err
	}
	level := s.effectiveLevelFromGrantAndSubscriptions(ctx, settings, userID, grant)
	benefits := settings.tierByLevel(level)
	if benefits.Level == "" {
		benefits = settings.tierByLevel(MembershipTierNormal)
	}
	s.cache.Store(userID, membershipEffectiveCacheEntry{
		expiresAt: now.Add(membershipEffectiveCacheTTL),
		benefits:  benefits,
	})
	return benefits, nil
}

func (s *MembershipService) ApplyRateMultiplier(ctx context.Context, userID int64, current float64) float64 {
	if s == nil || userID <= 0 {
		return current
	}
	benefits, err := s.GetEffectiveBenefits(ctx, userID)
	if err != nil {
		slog.Warn("membership multiplier lookup failed", "user_id", userID, "error", err)
		return current
	}
	if benefits.RateMultiplier > 0 && (current <= 0 || benefits.RateMultiplier < current) {
		return benefits.RateMultiplier
	}
	return current
}

func (s *MembershipService) ImageActiveTaskLimit(ctx context.Context, userID int64) int {
	if s == nil || userID <= 0 {
		return DefaultMembershipSettings().tierByLevel(MembershipTierNormal).ImageActiveTasks
	}
	benefits, err := s.GetEffectiveBenefits(ctx, userID)
	if err != nil {
		slog.Warn("membership image task limit lookup failed", "user_id", userID, "error", err)
		return DefaultMembershipSettings().tierByLevel(MembershipTierNormal).ImageActiveTasks
	}
	if benefits.ImageActiveTasks <= 0 {
		return DefaultMembershipSettings().tierByLevel(MembershipTierNormal).ImageActiveTasks
	}
	return benefits.ImageActiveTasks
}

func (s *MembershipService) ensureAutoGrant(ctx context.Context, userID, sourceOrderID int64, periodKey string, periodStart, periodEnd time.Time, netPaid float64, tier MembershipTierConfig) error {
	now := apptimezone.Now()
	var orderID *int64
	if sourceOrderID > 0 {
		orderID = &sourceOrderID
	}
	grant := &MembershipGrant{
		UserID:          userID,
		Tier:            tier.Level,
		Source:          MembershipGrantSourceAutoMonthlySpend,
		PeriodKey:       periodKey,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		QualifiedAmount: roundMoney(netPaid),
		StartsAt:        now,
		ExpiresAt:       now.AddDate(0, 0, s.validityDays(ctx)),
		Status:          MembershipGrantStatusActive,
		SourceOrderID:   orderID,
	}
	if tier.SubscriptionGroupID > 0 {
		gid := tier.SubscriptionGroupID
		grant.SubscriptionGroupID = &gid
	}
	stored, _, err := s.repo.UpsertAutoGrant(ctx, grant)
	if err != nil {
		return err
	}
	if stored == nil || tier.SubscriptionGroupID <= 0 || stored.SubscriptionID != nil {
		return nil
	}
	if s.subscriptionSvc == nil {
		return nil
	}
	sub, err := s.subscriptionSvc.AssignSubscription(ctx, &AssignSubscriptionInput{
		UserID:       userID,
		GroupID:      tier.SubscriptionGroupID,
		ValidityDays: s.validityDays(ctx),
		AssignedBy:   0,
		Notes:        membershipAutoGrantSubscriptionNote(tier.Level, periodKey),
	})
	if err != nil {
		if infraerrors.Reason(err) == infraerrors.Reason(ErrSubscriptionAssignConflict) {
			slog.Warn("membership auto subscription assignment skipped because an existing subscription conflicts",
				"user_id", userID,
				"tier", tier.Level,
				"group_id", tier.SubscriptionGroupID,
				"period_key", periodKey,
				"error", err,
			)
			return nil
		}
		return err
	}
	if sub != nil && sub.ID > 0 {
		if err := s.repo.UpdateGrantSubscription(ctx, stored.ID, sub.ID, tier.SubscriptionGroupID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MembershipService) revokeUnqualifiedCurrentPeriodGrants(ctx context.Context, userID int64, periodKey string, netPaid float64, targetLevel, reason string) error {
	grants, err := s.repo.ListActiveAutoGrants(ctx, userID, apptimezone.Now())
	if err != nil {
		return err
	}
	targetRank := membershipTierRank(targetLevel)
	for _, grant := range grants {
		if grant.PeriodKey != periodKey {
			continue
		}
		if membershipTierRank(grant.Tier) <= targetRank {
			continue
		}
		revokeReason := strings.TrimSpace(reason)
		if revokeReason == "" {
			revokeReason = "membership recalculated"
		}
		revokeReason = fmt.Sprintf("%s: net paid %.2f no longer qualifies %s", revokeReason, roundMoney(netPaid), grant.Tier)
		if grant.SubscriptionID != nil && s.subscriptionSvc != nil {
			if err := s.subscriptionSvc.RevokeSubscription(ctx, *grant.SubscriptionID); err != nil {
				slog.Warn("membership subscription revoke failed", "grant_id", grant.ID, "subscription_id", *grant.SubscriptionID, "error", err)
			}
		}
		if err := s.repo.RevokeGrant(ctx, grant.ID, revokeReason); err != nil {
			return err
		}
	}
	return nil
}

func membershipAutoGrantSubscriptionNote(tier, periodKey string) string {
	return fmt.Sprintf("auto membership grant: %s %s", normalizeMembershipLevel(tier), periodKey)
}

func (s *MembershipService) activeHighestGrant(ctx context.Context, userID int64, now time.Time) (*MembershipGrant, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	grant, err := s.repo.GetActiveHighestAutoGrant(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	return grant, nil
}

func (s *MembershipService) effectiveLevelFromGrantAndSubscriptions(ctx context.Context, settings MembershipSettings, userID int64, grant *MembershipGrant) string {
	level := MembershipTierNormal
	if grant != nil {
		level = normalizeMembershipLevel(grant.Tier)
	}
	if s == nil || s.subscriptionSvc == nil || userID <= 0 {
		return level
	}
	subs, err := s.subscriptionSvc.ListActiveUserSubscriptions(ctx, userID)
	if err != nil {
		slog.Warn("membership subscription level lookup failed", "user_id", userID, "error", err)
		return level
	}
	groupToLevel := make(map[int64]string, len(settings.Tiers))
	for _, tier := range settings.Tiers {
		if tier.SubscriptionGroupID > 0 {
			groupToLevel[tier.SubscriptionGroupID] = tier.Level
		}
	}
	for _, sub := range subs {
		if tierLevel := groupToLevel[sub.GroupID]; membershipTierRank(tierLevel) > membershipTierRank(level) {
			level = tierLevel
		}
	}
	return level
}

func (s *MembershipService) validityDays(ctx context.Context) int {
	settings, err := s.GetSettings(ctx)
	if err != nil || settings.ValidityDays <= 0 {
		return defaultMembershipValidityDays
	}
	return settings.ValidityDays
}

func (s *MembershipService) invalidateUser(userID int64) {
	if s == nil {
		return
	}
	s.cache.Delete(userID)
}

func (s *MembershipService) invalidateAll() {
	if s == nil {
		return
	}
	s.cache.Range(func(key, _ any) bool {
		s.cache.Delete(key)
		return true
	})
}

func normalizeMembershipSettings(settings MembershipSettings) MembershipSettings {
	if settings.ValidityDays <= 0 {
		settings.ValidityDays = defaultMembershipValidityDays
	}
	if settings.ValidityDays > MaxValidityDays {
		settings.ValidityDays = MaxValidityDays
	}
	defaults := DefaultMembershipSettingsNoNormalize()
	byLevel := make(map[string]MembershipTierConfig, len(defaults.Tiers))
	for _, tier := range defaults.Tiers {
		byLevel[tier.Level] = tier
	}
	for _, tier := range settings.Tiers {
		level := normalizeMembershipLevel(tier.Level)
		if level == "" {
			continue
		}
		base := byLevel[level]
		if strings.TrimSpace(tier.Label) != "" {
			base.Label = strings.TrimSpace(tier.Label)
		}
		base.ThresholdAmount = sanitizeMoney(tier.ThresholdAmount)
		if tier.RateMultiplier > 0 {
			base.RateMultiplier = tier.RateMultiplier
		}
		if tier.RPMLimit >= 0 {
			base.RPMLimit = tier.RPMLimit
		}
		if tier.TPMLimit >= 0 {
			base.TPMLimit = tier.TPMLimit
		}
		if tier.ImageActiveTasks >= 0 {
			base.ImageActiveTasks = tier.ImageActiveTasks
		}
		if tier.SubscriptionGroupID >= 0 {
			base.SubscriptionGroupID = tier.SubscriptionGroupID
		}
		base.Level = level
		byLevel[level] = base
	}
	settings.Tiers = []MembershipTierConfig{
		byLevel[MembershipTierNormal],
		byLevel[MembershipTierVIP],
		byLevel[MembershipTierSVIP],
	}
	return settings
}

func DefaultMembershipSettingsNoNormalize() MembershipSettings {
	return MembershipSettings{
		Enabled:      true,
		ValidityDays: defaultMembershipValidityDays,
		Tiers: []MembershipTierConfig{
			{Level: MembershipTierNormal, Label: "普通用户", ThresholdAmount: 0, RateMultiplier: 1, RPMLimit: 60, TPMLimit: 60000, ImageActiveTasks: 2},
			{Level: MembershipTierVIP, Label: "VIP", ThresholdAmount: 50, RateMultiplier: 0.8, RPMLimit: 300, TPMLimit: 300000, ImageActiveTasks: 5},
			{Level: MembershipTierSVIP, Label: "SVIP", ThresholdAmount: 100, RateMultiplier: 0.6, RPMLimit: 600, TPMLimit: 600000, ImageActiveTasks: 10},
		},
	}
}

func validateMembershipSettings(settings MembershipSettings) error {
	if settings.ValidityDays <= 0 || settings.ValidityDays > MaxValidityDays {
		return fmt.Errorf("validity_days must be between 1 and %d", MaxValidityDays)
	}
	normal := settings.tierByLevel(MembershipTierNormal)
	vip := settings.tierByLevel(MembershipTierVIP)
	svip := settings.tierByLevel(MembershipTierSVIP)
	if normal.Level == "" || vip.Level == "" || svip.Level == "" {
		return fmt.Errorf("normal, vip and svip tiers are required")
	}
	if normal.ThresholdAmount != 0 {
		return fmt.Errorf("normal threshold must be 0")
	}
	if vip.ThresholdAmount <= 0 || svip.ThresholdAmount <= vip.ThresholdAmount {
		return fmt.Errorf("tier thresholds must satisfy normal=0 < vip < svip")
	}
	for _, tier := range settings.Tiers {
		if tier.RateMultiplier <= 0 {
			return fmt.Errorf("%s rate_multiplier must be greater than 0", tier.Level)
		}
		if tier.RPMLimit < 0 || tier.TPMLimit < 0 || tier.ImageActiveTasks < 0 {
			return fmt.Errorf("%s limits must be non-negative", tier.Level)
		}
		if tier.SubscriptionGroupID < 0 {
			return fmt.Errorf("%s subscription_group_id must be non-negative", tier.Level)
		}
	}
	if settings.Enabled {
		if vip.SubscriptionGroupID <= 0 {
			return fmt.Errorf("vip subscription_group_id is required")
		}
		if svip.SubscriptionGroupID <= 0 {
			return fmt.Errorf("svip subscription_group_id is required")
		}
	}
	return nil
}

func normalizeMembershipLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case MembershipTierNormal, "":
		return MembershipTierNormal
	case MembershipTierVIP:
		return MembershipTierVIP
	case MembershipTierSVIP:
		return MembershipTierSVIP
	default:
		return ""
	}
}

func membershipTierRank(level string) int {
	switch normalizeMembershipLevel(level) {
	case MembershipTierSVIP:
		return 3
	case MembershipTierVIP:
		return 2
	case MembershipTierNormal:
		return 1
	default:
		return 0
	}
}

func membershipCurrentMonth(now time.Time) (time.Time, time.Time) {
	start := apptimezone.StartOfMonth(now)
	return start, start.AddDate(0, 1, 0)
}

func (settings MembershipSettings) tierByLevel(level string) MembershipTierConfig {
	level = normalizeMembershipLevel(level)
	for _, tier := range settings.Tiers {
		if tier.Level == level {
			return tier
		}
	}
	return MembershipTierConfig{}
}

func (settings MembershipSettings) qualifiedPaidTier(amount float64) MembershipTierConfig {
	tiers := append([]MembershipTierConfig(nil), settings.Tiers...)
	sort.SliceStable(tiers, func(i, j int) bool {
		return membershipTierRank(tiers[i].Level) > membershipTierRank(tiers[j].Level)
	})
	for _, tier := range tiers {
		if amount+0.000001 >= tier.ThresholdAmount {
			return tier
		}
	}
	return settings.tierByLevel(MembershipTierNormal)
}

func (settings MembershipSettings) nextTierByAmount(amount float64) *MembershipTierConfig {
	tiers := append([]MembershipTierConfig(nil), settings.Tiers...)
	sort.SliceStable(tiers, func(i, j int) bool {
		return tiers[i].ThresholdAmount < tiers[j].ThresholdAmount
	})
	for _, tier := range tiers {
		if tier.ThresholdAmount > amount+0.000001 {
			cp := tier
			return &cp
		}
	}
	return nil
}

func sanitizeMoney(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	return roundMoney(v)
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
