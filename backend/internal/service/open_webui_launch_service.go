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

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const openWebUILaunchTokenKeyPrefix = "sub2api:open_webui:launch:"

var (
	ErrOpenWebUIDisabled           = infraerrors.Forbidden("OPEN_WEBUI_DISABLED", "Open WebUI launch is disabled")
	ErrOpenWebUIInvalidSecret      = infraerrors.Unauthorized("OPEN_WEBUI_INVALID_SECRET", "invalid Open WebUI redeem secret")
	ErrOpenWebUILaunchTokenInvalid = infraerrors.Unauthorized("OPEN_WEBUI_LAUNCH_TOKEN_INVALID", "launch token is invalid or expired")
	ErrOpenWebUIKeyNotUsable       = infraerrors.BadRequest("OPEN_WEBUI_API_KEY_NOT_USABLE", "selected API key cannot be used by Open WebUI")
)

type OpenWebUILaunchService struct {
	apiKeyService *APIKeyService
	tokenStore    OpenWebUILaunchTokenStore
	cfg           *config.Config
}

type OpenWebUILaunchTokenStore interface {
	Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error
	GetDel(ctx context.Context, key string) ([]byte, bool, error)
}

type OpenWebUILaunch struct {
	LaunchURL string    `json:"launch_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type OpenWebUIRedeemResult struct {
	User           OpenWebUIRedeemUser   `json:"user"`
	APIKey         OpenWebUIRedeemAPIKey `json:"api_key"`
	GatewayBaseURL string                `json:"gateway_base_url"`
	ExpiresAt      time.Time             `json:"expires_at"`
}

type OpenWebUIRedeemUser struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type OpenWebUIRedeemAPIKey struct {
	ID            int64  `json:"id"`
	Key           string `json:"key"`
	Name          string `json:"name"`
	GroupID       int64  `json:"group_id"`
	GroupName     string `json:"group_name"`
	GroupPlatform string `json:"group_platform"`
}

type openWebUILaunchTokenPayload struct {
	UserID         int64     `json:"user_id"`
	APIKeyID       int64     `json:"api_key_id"`
	GatewayBaseURL string    `json:"gateway_base_url"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func NewOpenWebUILaunchService(apiKeyService *APIKeyService, tokenStore OpenWebUILaunchTokenStore, cfg *config.Config) *OpenWebUILaunchService {
	return &OpenWebUILaunchService{
		apiKeyService: apiKeyService,
		tokenStore:    tokenStore,
		cfg:           cfg,
	}
}

func (s *OpenWebUILaunchService) CreateLaunch(ctx context.Context, userID, apiKeyID int64, requestGatewayBaseURL string) (*OpenWebUILaunch, error) {
	if s == nil || s.cfg == nil || !s.cfg.OpenWebUI.Enabled {
		return nil, ErrOpenWebUIDisabled
	}
	if s.tokenStore == nil {
		return nil, infraerrors.InternalServer("OPEN_WEBUI_REDIS_UNAVAILABLE", "Open WebUI launch cache is unavailable")
	}
	if apiKeyID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_API_KEY_ID", "api_key_id is required")
	}

	apiKey, err := s.loadUsableAPIKey(ctx, userID, apiKeyID)
	if err != nil {
		return nil, err
	}

	token, err := randomLaunchToken()
	if err != nil {
		return nil, fmt.Errorf("generate open webui launch token: %w", err)
	}

	ttl := time.Duration(s.cfg.OpenWebUI.TokenTTLSeconds) * time.Second
	expiresAt := time.Now().UTC().Add(ttl)
	payload := openWebUILaunchTokenPayload{
		UserID:         userID,
		APIKeyID:       apiKey.ID,
		GatewayBaseURL: s.resolveGatewayBaseURL(requestGatewayBaseURL),
		ExpiresAt:      expiresAt,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal open webui launch token: %w", err)
	}
	if err := s.tokenStore.Set(ctx, openWebUILaunchTokenKeyPrefix+token, raw, ttl); err != nil {
		return nil, fmt.Errorf("store open webui launch token: %w", err)
	}

	launchURL, err := s.buildLaunchURL(token)
	if err != nil {
		return nil, err
	}

	return &OpenWebUILaunch{
		LaunchURL: launchURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *OpenWebUILaunchService) Redeem(ctx context.Context, token, providedSecret string) (*OpenWebUIRedeemResult, error) {
	if s == nil || s.cfg == nil || !s.cfg.OpenWebUI.Enabled {
		return nil, ErrOpenWebUIDisabled
	}
	if s.tokenStore == nil {
		return nil, infraerrors.InternalServer("OPEN_WEBUI_REDIS_UNAVAILABLE", "Open WebUI launch cache is unavailable")
	}
	if !s.validRedeemSecret(providedSecret) {
		return nil, ErrOpenWebUIInvalidSecret
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}

	raw, ok, err := s.tokenStore.GetDel(ctx, openWebUILaunchTokenKeyPrefix+token)
	if err != nil {
		return nil, fmt.Errorf("redeem open webui launch token: %w", err)
	}
	if !ok {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}

	var payload openWebUILaunchTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}
	if payload.ExpiresAt.IsZero() || time.Now().UTC().After(payload.ExpiresAt) {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}

	apiKey, err := s.loadUsableAPIKey(ctx, payload.UserID, payload.APIKeyID)
	if err != nil {
		return nil, err
	}
	if apiKey.User == nil || apiKey.Group == nil || apiKey.GroupID == nil {
		return nil, ErrOpenWebUIKeyNotUsable
	}

	return &OpenWebUIRedeemResult{
		User: OpenWebUIRedeemUser{
			ID:       apiKey.User.ID,
			Email:    apiKey.User.Email,
			Username: apiKey.User.Username,
		},
		APIKey: OpenWebUIRedeemAPIKey{
			ID:            apiKey.ID,
			Key:           apiKey.Key,
			Name:          apiKey.Name,
			GroupID:       *apiKey.GroupID,
			GroupName:     apiKey.Group.Name,
			GroupPlatform: apiKey.Group.Platform,
		},
		GatewayBaseURL: payload.GatewayBaseURL,
		ExpiresAt:      payload.ExpiresAt,
	}, nil
}

func (s *OpenWebUILaunchService) loadUsableAPIKey(ctx context.Context, userID, apiKeyID int64) (*APIKey, error) {
	apiKey, err := s.apiKeyService.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}
	if apiKey.UserID != userID {
		return nil, ErrInsufficientPerms
	}
	if apiKey.User == nil || !apiKey.User.IsActive() {
		return nil, ErrUserNotActive
	}
	if !apiKey.IsActive() {
		return nil, infraerrors.BadRequest("API_KEY_INACTIVE", "api key is not active")
	}
	if apiKey.GroupID == nil || apiKey.Group == nil || !apiKey.Group.IsActive() {
		return nil, ErrOpenWebUIKeyNotUsable
	}
	if err := s.apiKeyService.CheckAPIKeyQuotaAndExpiry(apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *OpenWebUILaunchService) validRedeemSecret(providedSecret string) bool {
	expected := strings.TrimSpace(s.cfg.OpenWebUI.RedeemSecret)
	if expected == "" {
		return false
	}
	provided := strings.TrimSpace(providedSecret)
	if provided == "" {
		return false
	}
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func (s *OpenWebUILaunchService) buildLaunchURL(token string) (string, error) {
	base, err := url.Parse(strings.TrimRight(s.cfg.OpenWebUI.ChatURL, "/"))
	if err != nil || !base.IsAbs() || strings.TrimSpace(base.Host) == "" {
		return "", infraerrors.InternalServer("OPEN_WEBUI_CHAT_URL_INVALID", "Open WebUI chat URL is invalid")
	}
	launchPath := strings.TrimSpace(s.cfg.OpenWebUI.LaunchPath)
	if launchPath == "" {
		launchPath = "/api/v1/auths/sub2api/launch"
	}
	rel := &url.URL{Path: launchPath}
	target := base.ResolveReference(rel)
	q := target.Query()
	q.Set("token", token)
	target.RawQuery = q.Encode()
	return target.String(), nil
}

func (s *OpenWebUILaunchService) resolveGatewayBaseURL(requestGatewayBaseURL string) string {
	if configured := strings.TrimRight(strings.TrimSpace(s.cfg.OpenWebUI.GatewayBaseURL), "/"); configured != "" {
		return configured
	}
	if frontendURL := strings.TrimRight(strings.TrimSpace(s.cfg.Server.FrontendURL), "/"); frontendURL != "" {
		return frontendURL + "/v1"
	}
	if requestGatewayBaseURL = strings.TrimRight(strings.TrimSpace(requestGatewayBaseURL), "/"); requestGatewayBaseURL != "" {
		return requestGatewayBaseURL
	}
	return "http://127.0.0.1:8080/v1"
}

func randomLaunchToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
