package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	StudioBridgeAppLuoyeAI = "luoye-ai"
)

const (
	defaultStudioBridgeLaunchReturnURL = "http://127.0.0.1:8081/auth/sub2api/launch"
	defaultStudioBridgeRechargeURL     = "http://127.0.0.1:62080/purchase"
)

const (
	studioBridgeTokenKeyPrefix = "sub2api:studio_bridge:launch:"
	studioBridgeTokenTTL       = 5 * time.Minute
)

var (
	ErrStudioBridgeDisabled       = infraerrors.Forbidden("STUDIO_BRIDGE_DISABLED", "studio bridge is disabled")
	ErrStudioBridgeInvalidApp     = infraerrors.NotFound("STUDIO_BRIDGE_APP_NOT_FOUND", "studio bridge app not found")
	ErrStudioBridgeInvalidSecret  = infraerrors.Unauthorized("STUDIO_BRIDGE_INVALID_SECRET", "invalid studio bridge secret")
	ErrStudioBridgeInvalidToken   = infraerrors.Unauthorized("STUDIO_BRIDGE_TOKEN_INVALID", "launch token is invalid or expired")
	ErrStudioBridgeInvalidReturn  = infraerrors.BadRequest("STUDIO_BRIDGE_RETURN_URL_INVALID", "return url is not allowed")
	ErrStudioBridgeChargeKeyEmpty = infraerrors.BadRequest("STUDIO_BRIDGE_CHARGE_KEY_REQUIRED", "charge_key is required")
	ErrStudioBridgeAmountInvalid  = infraerrors.BadRequest("STUDIO_BRIDGE_AMOUNT_INVALID", "amount must be positive")
	ErrStudioBridgeConflict       = infraerrors.Conflict("STUDIO_BRIDGE_CHARGE_CONFLICT", "charge_key fingerprint conflict")
	ErrStudioBridgeInsufficient   = infraerrors.BadRequest("STUDIO_BRIDGE_INSUFFICIENT_BALANCE", "insufficient balance")
	ErrStudioBridgeGroupRequired  = infraerrors.BadRequest("STUDIO_BRIDGE_GROUP_REQUIRED", "default chat and image studio bridge groups are required when studio bridge is enabled")
)

type StudioBridgeService struct {
	settings *SettingService
	repo     StudioBridgeRepository
	store    StudioBridgeStore
}

type StudioBridgeStore interface {
	Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error
	GetDel(ctx context.Context, key string) ([]byte, bool, error)
}

type StudioBridgeRepository interface {
	GetUserSummary(ctx context.Context, userID int64, rechargeURL string, usageLimit int) (*StudioBridgeUserSummary, error)
	ReserveStudioBridgeCharge(ctx context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error)
	CommitStudioBridgeCharge(ctx context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error)
	RefundStudioBridgeCharge(ctx context.Context, cmd StudioBridgeChargeCommand) (*StudioBridgeChargeResult, error)
}

type StudioBridgeAppSettings struct {
	Enabled              bool     `json:"enabled"`
	SiteName             string   `json:"site_name"`
	AllowedReturnDomains []string `json:"allowed_return_domains"`
	LaunchReturnURL      string   `json:"launch_return_url"`
	RechargeReturnURL    string   `json:"recharge_return_url"`
	DefaultChatGroup     string   `json:"default_chat_group"`
	DefaultImageGroup    string   `json:"default_image_group"`
	DefaultVideoGroup    string   `json:"default_video_group"`
	InternalSecret       string   `json:"internal_secret,omitempty"`
	SecretConfigured     bool     `json:"secret_configured,omitempty"`
}

type StudioBridgeLaunch struct {
	LaunchURL string    `json:"launch_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type StudioBridgeRedeemResult struct {
	UserID            int64     `json:"user_id"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	ExpiresAt         time.Time `json:"expires_at"`
	DefaultChatGroup  string    `json:"default_chat_group"`
	DefaultImageGroup string    `json:"default_image_group"`
	DefaultVideoGroup string    `json:"default_video_group"`
}

type StudioBridgeUserSummary struct {
	UserID      int64                       `json:"user_id"`
	Email       string                      `json:"email"`
	Username    string                      `json:"username"`
	Balance     float64                     `json:"balance"`
	RechargeURL string                      `json:"recharge_url"`
	Usage       []StudioBridgeUsageSummary  `json:"usage"`
	Orders      []StudioBridgeRechargeOrder `json:"recent_recharges"`
}

type StudioBridgeUsageSummary struct {
	RequestID  string    `json:"request_id"`
	Model      string    `json:"model"`
	ActualCost float64   `json:"actual_cost"`
	CreatedAt  time.Time `json:"created_at"`
}

type StudioBridgeRechargeOrder struct {
	ID        int64      `json:"id"`
	Amount    float64    `json:"amount"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
}

type StudioBridgeChargeCommand struct {
	AppID              string  `json:"app_id"`
	UserID             int64   `json:"user_id"`
	ChargeKey          string  `json:"charge_key"`
	RefundForChargeKey string  `json:"refund_for_charge_key,omitempty"`
	Amount             float64 `json:"amount"`
	Reason             string  `json:"reason"`
	TaskID             string  `json:"task_id,omitempty"`
	Mode               string  `json:"mode,omitempty"`
	Model              string  `json:"model,omitempty"`
	ActorUserID        string  `json:"actor_user_id,omitempty"`
	TeamID             string  `json:"team_id,omitempty"`
}

type StudioBridgeChargeResult struct {
	ChargeKey    string  `json:"charge_key"`
	Status       string  `json:"status"`
	Applied      bool    `json:"applied"`
	Amount       float64 `json:"amount"`
	BalanceAfter float64 `json:"balance_after"`
}

type studioBridgeLaunchPayload struct {
	AppID     string    `json:"app_id"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewStudioBridgeService(settings *SettingService, repo StudioBridgeRepository, store StudioBridgeStore) *StudioBridgeService {
	return &StudioBridgeService{settings: settings, repo: repo, store: store}
}

func (s *StudioBridgeService) GetAppSettings(ctx context.Context, appID string) (*StudioBridgeAppSettings, error) {
	if normalizeStudioBridgeAppID(appID) != StudioBridgeAppLuoyeAI {
		return nil, ErrStudioBridgeInvalidApp
	}
	if s == nil || s.settings == nil {
		return defaultStudioBridgeAppSettings(), nil
	}
	cfg, err := s.settings.GetStudioBridgeLuoyeAISettings(ctx)
	if err != nil {
		return nil, err
	}
	return sanitizeStudioBridgeSettings(cfg), nil
}

func (s *StudioBridgeService) CreateLaunch(ctx context.Context, userID int64, appID, returnURL string) (*StudioBridgeLaunch, error) {
	cfg, err := s.loadEnabledApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	if s.store == nil {
		return nil, infraerrors.InternalServer("STUDIO_BRIDGE_STORE_UNAVAILABLE", "studio bridge token store is unavailable")
	}
	target, err := resolveStudioBridgeLaunchTarget(returnURL, cfg)
	if err != nil {
		return nil, err
	}
	token, err := randomStudioBridgeToken()
	if err != nil {
		return nil, fmt.Errorf("generate studio bridge token: %w", err)
	}
	expiresAt := time.Now().UTC().Add(studioBridgeTokenTTL)
	raw, err := json.Marshal(studioBridgeLaunchPayload{AppID: StudioBridgeAppLuoyeAI, UserID: userID, ExpiresAt: expiresAt})
	if err != nil {
		return nil, fmt.Errorf("marshal studio bridge launch token: %w", err)
	}
	if err := s.store.Set(ctx, studioBridgeTokenKeyPrefix+token, raw, studioBridgeTokenTTL); err != nil {
		return nil, fmt.Errorf("store studio bridge launch token: %w", err)
	}
	q := target.Query()
	q.Set("launch_token", token)
	target.RawQuery = q.Encode()
	return &StudioBridgeLaunch{LaunchURL: target.String(), ExpiresAt: expiresAt}, nil
}

func (s *StudioBridgeService) RedeemLaunch(ctx context.Context, appID, token, secret string) (*StudioBridgeRedeemResult, error) {
	cfg, err := s.loadEnabledApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	if !validStudioBridgeSecret(cfg.InternalSecret, secret) {
		return nil, ErrStudioBridgeInvalidSecret
	}
	if s.store == nil {
		return nil, infraerrors.InternalServer("STUDIO_BRIDGE_STORE_UNAVAILABLE", "studio bridge token store is unavailable")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrStudioBridgeInvalidToken
	}
	raw, ok, err := s.store.GetDel(ctx, studioBridgeTokenKeyPrefix+token)
	if err != nil {
		return nil, fmt.Errorf("redeem studio bridge launch token: %w", err)
	}
	if !ok {
		return nil, ErrStudioBridgeInvalidToken
	}
	var payload studioBridgeLaunchPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, ErrStudioBridgeInvalidToken
	}
	if payload.UserID <= 0 || payload.AppID != StudioBridgeAppLuoyeAI || payload.ExpiresAt.IsZero() || time.Now().UTC().After(payload.ExpiresAt) {
		return nil, ErrStudioBridgeInvalidToken
	}
	summary, err := s.repo.GetUserSummary(ctx, payload.UserID, cfg.RechargeReturnURL, 0)
	if err != nil {
		return nil, err
	}
	return &StudioBridgeRedeemResult{
		UserID:            summary.UserID,
		Email:             summary.Email,
		Username:          summary.Username,
		ExpiresAt:         payload.ExpiresAt,
		DefaultChatGroup:  cfg.DefaultChatGroup,
		DefaultImageGroup: cfg.DefaultImageGroup,
		DefaultVideoGroup: cfg.DefaultVideoGroup,
	}, nil
}

func (s *StudioBridgeService) GetUserSummary(ctx context.Context, appID string, userID int64, secret string) (*StudioBridgeUserSummary, error) {
	cfg, err := s.loadEnabledApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	if !validStudioBridgeSecret(cfg.InternalSecret, secret) {
		return nil, ErrStudioBridgeInvalidSecret
	}
	return s.repo.GetUserSummary(ctx, userID, cfg.RechargeReturnURL, 20)
}

func (s *StudioBridgeService) Reserve(ctx context.Context, cmd StudioBridgeChargeCommand, secret string) (*StudioBridgeChargeResult, error) {
	cfg, err := s.loadEnabledApp(ctx, cmd.AppID)
	if err != nil {
		return nil, err
	}
	if !validStudioBridgeSecret(cfg.InternalSecret, secret) {
		return nil, ErrStudioBridgeInvalidSecret
	}
	if err := normalizeStudioBridgeChargeCommand(&cmd); err != nil {
		return nil, err
	}
	return s.repo.ReserveStudioBridgeCharge(ctx, cmd)
}

func (s *StudioBridgeService) Commit(ctx context.Context, cmd StudioBridgeChargeCommand, secret string) (*StudioBridgeChargeResult, error) {
	cfg, err := s.loadEnabledApp(ctx, cmd.AppID)
	if err != nil {
		return nil, err
	}
	if !validStudioBridgeSecret(cfg.InternalSecret, secret) {
		return nil, ErrStudioBridgeInvalidSecret
	}
	if err := normalizeStudioBridgeChargeCommand(&cmd); err != nil {
		return nil, err
	}
	return s.repo.CommitStudioBridgeCharge(ctx, cmd)
}

func (s *StudioBridgeService) Refund(ctx context.Context, cmd StudioBridgeChargeCommand, secret string) (*StudioBridgeChargeResult, error) {
	cfg, err := s.loadEnabledApp(ctx, cmd.AppID)
	if err != nil {
		return nil, err
	}
	if !validStudioBridgeSecret(cfg.InternalSecret, secret) {
		return nil, ErrStudioBridgeInvalidSecret
	}
	if err := normalizeStudioBridgeChargeCommand(&cmd); err != nil {
		return nil, err
	}
	return s.repo.RefundStudioBridgeCharge(ctx, cmd)
}

func (s *StudioBridgeService) loadEnabledApp(ctx context.Context, appID string) (*StudioBridgeAppSettings, error) {
	if normalizeStudioBridgeAppID(appID) != StudioBridgeAppLuoyeAI {
		return nil, ErrStudioBridgeInvalidApp
	}
	if s == nil || s.settings == nil || s.repo == nil {
		return nil, infraerrors.InternalServer("STUDIO_BRIDGE_UNAVAILABLE", "studio bridge service is unavailable")
	}
	full, err := s.settings.GetStudioBridgeLuoyeAISettings(ctx)
	if err != nil {
		return nil, err
	}
	if !full.Enabled {
		return nil, ErrStudioBridgeDisabled
	}
	if err := validateStudioBridgeAppSettings(*full); err != nil {
		return nil, err
	}
	return full, nil
}

func defaultStudioBridgeAppSettings() *StudioBridgeAppSettings {
	return &StudioBridgeAppSettings{
		SiteName:             "落叶创艺",
		AllowedReturnDomains: []string{},
		LaunchReturnURL:      defaultStudioBridgeLaunchReturnURL,
		RechargeReturnURL:    defaultStudioBridgeRechargeURL,
	}
}

func sanitizeStudioBridgeSettings(cfg *StudioBridgeAppSettings) *StudioBridgeAppSettings {
	if cfg == nil {
		return defaultStudioBridgeAppSettings()
	}
	out := *cfg
	out.InternalSecret = ""
	out.SecretConfigured = strings.TrimSpace(cfg.InternalSecret) != ""
	return &out
}

func normalizeStudioBridgeAppID(appID string) string {
	return strings.ToLower(strings.TrimSpace(appID))
}

func normalizeStudioBridgeChargeCommand(cmd *StudioBridgeChargeCommand) error {
	cmd.AppID = normalizeStudioBridgeAppID(cmd.AppID)
	if cmd.AppID == "" {
		cmd.AppID = StudioBridgeAppLuoyeAI
	}
	cmd.ChargeKey = strings.TrimSpace(cmd.ChargeKey)
	cmd.RefundForChargeKey = strings.TrimSpace(cmd.RefundForChargeKey)
	cmd.Reason = strings.TrimSpace(cmd.Reason)
	cmd.TaskID = strings.TrimSpace(cmd.TaskID)
	cmd.Mode = strings.TrimSpace(cmd.Mode)
	cmd.Model = strings.TrimSpace(cmd.Model)
	cmd.ActorUserID = strings.TrimSpace(cmd.ActorUserID)
	cmd.TeamID = strings.TrimSpace(cmd.TeamID)
	if cmd.ChargeKey == "" {
		return ErrStudioBridgeChargeKeyEmpty
	}
	if cmd.Amount <= 0 {
		return ErrStudioBridgeAmountInvalid
	}
	return nil
}

func studioBridgeChargeFingerprint(cmd StudioBridgeChargeCommand) string {
	return cmd.Fingerprint()
}

func (cmd StudioBridgeChargeCommand) Fingerprint() string {
	return fmt.Sprintf("%s|%d|%s|%s|%.8f", cmd.AppID, cmd.UserID, cmd.ChargeKey, cmd.RefundForChargeKey, cmd.Amount)
}

func resolveStudioBridgeLaunchTarget(returnURL string, cfg *StudioBridgeAppSettings) (*url.URL, error) {
	returnURL = strings.TrimSpace(returnURL)
	if returnURL != "" {
		return validateStudioBridgeReturnURL(returnURL, cfg.AllowedReturnDomains)
	}
	if cfg == nil {
		return nil, ErrStudioBridgeInvalidReturn
	}
	return validateStudioBridgeConfiguredLaunchURL(cfg.LaunchReturnURL)
}

func validateStudioBridgeConfiguredLaunchURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !u.IsAbs() || strings.TrimSpace(u.Hostname()) == "" {
		return nil, ErrStudioBridgeInvalidReturn
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return nil, ErrStudioBridgeInvalidReturn
	}
	return u, nil
}

func validateStudioBridgeReturnURL(raw string, domains []string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !u.IsAbs() || strings.TrimSpace(u.Hostname()) == "" {
		return nil, ErrStudioBridgeInvalidReturn
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return nil, ErrStudioBridgeInvalidReturn
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return u, nil
		}
	}
	return nil, ErrStudioBridgeInvalidReturn
}

func validStudioBridgeSecret(expected, provided string) bool {
	expected = strings.TrimSpace(expected)
	provided = strings.TrimSpace(provided)
	if expected == "" || provided == "" || len(expected) != len(provided) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func randomStudioBridgeToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
