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
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const openWebUILaunchTokenKeyPrefix = "sub2api:open_webui:launch:"
const openWebUISessionTokenKeyPrefix = "sub2api:open_webui:session:"
const openWebUISessionTokenMinTTL = 24 * time.Hour

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
	Get(ctx context.Context, key string) ([]byte, bool, error)
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
	SessionToken   string                `json:"session_token,omitempty"`
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
	Last4         string `json:"last4"`
	GroupID       int64  `json:"group_id"`
	GroupName     string `json:"group_name"`
	GroupPlatform string `json:"group_platform"`
}

type OpenWebUIAPIKeyOption struct {
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Last4                   string `json:"last4"`
	GroupID                 int64  `json:"group_id"`
	GroupName               string `json:"group_name"`
	GroupPlatform           string `json:"group_platform"`
	SupportsImageGeneration bool   `json:"supports_image_generation"`
}

type openWebUILaunchTokenPayload struct {
	UserID         int64     `json:"user_id"`
	APIKeyID       int64     `json:"api_key_id"`
	GatewayBaseURL string    `json:"gateway_base_url"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type openWebUISessionTokenPayload struct {
	UserID         int64     `json:"user_id"`
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

	if apiKeyID > 0 {
		if _, err := s.loadUsableAPIKey(ctx, userID, apiKeyID); err != nil {
			return nil, err
		}
	}

	token, err := randomLaunchToken()
	if err != nil {
		return nil, fmt.Errorf("generate open webui launch token: %w", err)
	}

	ttl := time.Duration(s.cfg.OpenWebUI.TokenTTLSeconds) * time.Second
	expiresAt := time.Now().UTC().Add(ttl)
	payload := openWebUILaunchTokenPayload{
		UserID:         userID,
		APIKeyID:       apiKeyID,
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

	user, err := s.loadUser(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}
	sessionToken, sessionExpiresAt, err := s.issueSessionToken(ctx, payload)
	if err != nil {
		return nil, err
	}
	result := &OpenWebUIRedeemResult{
		User: OpenWebUIRedeemUser{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
		GatewayBaseURL: payload.GatewayBaseURL,
		ExpiresAt:      sessionExpiresAt,
		SessionToken:   sessionToken,
	}
	if payload.APIKeyID <= 0 {
		return result, nil
	}

	apiKey, err := s.loadUsableAPIKey(ctx, payload.UserID, payload.APIKeyID)
	if err != nil {
		return nil, err
	}
	apiKeyOption, ok := openWebUIAPIKeyOptionFromAPIKey(apiKey)
	if !ok {
		return nil, ErrOpenWebUIKeyNotUsable
	}
	result.APIKey = OpenWebUIRedeemAPIKey{
		ID:            apiKey.ID,
		Key:           apiKey.Key,
		Name:          apiKey.Name,
		Last4:         apiKeyOption.Last4,
		GroupID:       apiKeyOption.GroupID,
		GroupName:     apiKeyOption.GroupName,
		GroupPlatform: apiKeyOption.GroupPlatform,
	}
	return result, nil
}

func (s *OpenWebUILaunchService) ListAPIKeys(ctx context.Context, userID int64) ([]OpenWebUIAPIKeyOption, error) {
	if s == nil || s.cfg == nil || !s.cfg.OpenWebUI.Enabled {
		return nil, ErrOpenWebUIDisabled
	}
	if _, err := s.loadUser(ctx, userID); err != nil {
		return nil, err
	}
	keys, _, err := s.apiKeyService.List(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 200}, APIKeyListFilters{Status: StatusAPIKeyActive})
	if err != nil {
		return nil, err
	}
	items := make([]OpenWebUIAPIKeyOption, 0, len(keys))
	for i := range keys {
		option, ok := openWebUIAPIKeyOptionFromAPIKey(&keys[i])
		if !ok {
			continue
		}
		if err := s.apiKeyService.CheckAPIKeyQuotaAndExpiry(&keys[i]); err != nil {
			continue
		}
		items = append(items, option)
	}
	return items, nil
}

func (s *OpenWebUILaunchService) ListAPIKeysBySession(ctx context.Context, sessionToken string) ([]OpenWebUIAPIKeyOption, error) {
	session, err := s.resolveSessionToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	return s.ListAPIKeys(ctx, session.UserID)
}

func (s *OpenWebUILaunchService) BindAPIKey(ctx context.Context, userID, apiKeyID int64, requestGatewayBaseURL string) (*OpenWebUIRedeemResult, error) {
	if s == nil || s.cfg == nil || !s.cfg.OpenWebUI.Enabled {
		return nil, ErrOpenWebUIDisabled
	}
	apiKey, err := s.loadUsableAPIKey(ctx, userID, apiKeyID)
	if err != nil {
		return nil, err
	}
	option, ok := openWebUIAPIKeyOptionFromAPIKey(apiKey)
	if !ok {
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
			Last4:         option.Last4,
			GroupID:       option.GroupID,
			GroupName:     option.GroupName,
			GroupPlatform: option.GroupPlatform,
		},
		GatewayBaseURL: s.resolveGatewayBaseURL(requestGatewayBaseURL),
		ExpiresAt:      time.Now().UTC().Add(time.Duration(s.cfg.OpenWebUI.TokenTTLSeconds) * time.Second),
	}, nil
}

func (s *OpenWebUILaunchService) BindAPIKeyBySession(ctx context.Context, sessionToken string, apiKeyID int64) (*OpenWebUIRedeemResult, error) {
	session, err := s.resolveSessionToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	result, err := s.BindAPIKey(ctx, session.UserID, apiKeyID, session.GatewayBaseURL)
	if err != nil {
		return nil, err
	}
	result.SessionToken = strings.TrimSpace(sessionToken)
	result.ExpiresAt = session.ExpiresAt
	return result, nil
}

func (s *OpenWebUILaunchService) ValidRedeemSecret(providedSecret string) bool {
	return s.validRedeemSecret(providedSecret)
}

func (s *OpenWebUILaunchService) loadUser(ctx context.Context, userID int64) (*User, error) {
	if s == nil || s.apiKeyService == nil || s.apiKeyService.userRepo == nil {
		return nil, infraerrors.InternalServer("OPEN_WEBUI_USER_REPOSITORY_UNAVAILABLE", "Open WebUI user repository is unavailable")
	}
	user, err := s.apiKeyService.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive() {
		return nil, ErrUserNotActive
	}
	return user, nil
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
	if err := s.apiKeyService.CheckAPIKeyQuotaAndExpiry(apiKey); err != nil {
		return nil, err
	}
	if _, ok := openWebUIAPIKeyOptionFromAPIKey(apiKey); !ok {
		return nil, ErrOpenWebUIKeyNotUsable
	}
	return apiKey, nil
}

func openWebUIAPIKeyOptionFromAPIKey(apiKey *APIKey) (OpenWebUIAPIKeyOption, bool) {
	if apiKey == nil || !apiKey.IsActive() {
		return OpenWebUIAPIKeyOption{}, false
	}
	groups := openWebUIImageGroupsForAPIKey(apiKey)
	if len(groups) == 0 {
		return OpenWebUIAPIKeyOption{}, false
	}
	group := groups[0]
	return OpenWebUIAPIKeyOption{
		ID:                      apiKey.ID,
		Name:                    apiKey.Name,
		Last4:                   openWebUIAPIKeyLast4(apiKey.Key),
		GroupID:                 group.ID,
		GroupName:               group.Name,
		GroupPlatform:           group.Platform,
		SupportsImageGeneration: true,
	}, true
}

func openWebUIImageGroupsForAPIKey(apiKey *APIKey) []*Group {
	if apiKey == nil {
		return nil
	}
	groups := make([]*Group, 0, 1+len(apiKey.MultiGroupRouteGroups))
	if apiKey.GroupID != nil && apiKey.Group != nil && apiKey.Group.IsActive() {
		groups = append(groups, apiKey.Group)
	}
	groups = append(groups, apiKey.MultiGroupRouteGroups...)
	seen := map[int64]struct{}{}
	out := make([]*Group, 0, len(groups))
	for _, group := range groups {
		if group == nil || !group.IsActive() || group.Platform != PlatformOpenAI || !group.AllowImageGeneration {
			continue
		}
		if _, ok := seen[group.ID]; ok {
			continue
		}
		seen[group.ID] = struct{}{}
		out = append(out, group)
	}
	return out
}

func openWebUIAPIKeyLast4(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 4 {
		return key
	}
	return key[len(key)-4:]
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

func (s *OpenWebUILaunchService) issueSessionToken(ctx context.Context, payload openWebUILaunchTokenPayload) (string, time.Time, error) {
	if s == nil || s.tokenStore == nil {
		return "", time.Time{}, infraerrors.InternalServer("OPEN_WEBUI_REDIS_UNAVAILABLE", "Open WebUI launch cache is unavailable")
	}
	token, err := randomLaunchToken()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate open webui session token: %w", err)
	}
	expiresAt := time.Now().UTC().Add(openWebUISessionTokenTTL(s.cfg))
	raw, err := json.Marshal(openWebUISessionTokenPayload{
		UserID:         payload.UserID,
		GatewayBaseURL: payload.GatewayBaseURL,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("marshal open webui session token: %w", err)
	}
	if err := s.tokenStore.Set(ctx, openWebUISessionTokenKeyPrefix+token, raw, time.Until(expiresAt)); err != nil {
		return "", time.Time{}, fmt.Errorf("store open webui session token: %w", err)
	}
	return token, expiresAt, nil
}

func (s *OpenWebUILaunchService) resolveSessionToken(ctx context.Context, sessionToken string) (*openWebUISessionTokenPayload, error) {
	if s == nil || s.tokenStore == nil {
		return nil, infraerrors.InternalServer("OPEN_WEBUI_REDIS_UNAVAILABLE", "Open WebUI launch cache is unavailable")
	}
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken == "" {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}
	raw, ok, err := s.tokenStore.Get(ctx, openWebUISessionTokenKeyPrefix+sessionToken)
	if err != nil {
		return nil, fmt.Errorf("load open webui session token: %w", err)
	}
	if !ok {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}
	var payload openWebUISessionTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}
	if payload.UserID <= 0 || payload.ExpiresAt.IsZero() || time.Now().UTC().After(payload.ExpiresAt) {
		return nil, ErrOpenWebUILaunchTokenInvalid
	}
	return &payload, nil
}

func openWebUISessionTokenTTL(cfg *config.Config) time.Duration {
	if cfg == nil {
		return openWebUISessionTokenMinTTL
	}
	ttl := time.Duration(cfg.OpenWebUI.TokenTTLSeconds) * time.Second
	if ttl < openWebUISessionTokenMinTTL {
		return openWebUISessionTokenMinTTL
	}
	return ttl
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
