package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newAuthRoutesTestRouter(redisClient *redis.Client) *gin.Engine {
	return newAuthRoutesTestRouterWithConfig(redisClient, &config.Config{})
}

func newAuthRoutesTestRouterWithConfig(redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	RegisterAuthRoutes(
		v1,
		&handler.Handlers{
			Auth:    &handler.AuthHandler{},
			Setting: &handler.SettingHandler{},
		},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
			c.Next()
		}),
		servermiddleware.OptionalJWTAuthMiddleware(func(c *gin.Context) {
			c.Next()
		}),
		nil,
		redisClient,
		nil,
		cfg,
	)

	return router
}

func TestAuthRoutesRateLimitFailCloseWhenRedisUnavailable(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	router := newAuthRoutesTestRouter(rdb)
	paths := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/login/2fa",
		"/api/v1/auth/send-verify-code",
		"/api/v1/auth/oauth/pending/send-verify-code",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "203.0.113.10:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusTooManyRequests, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "rate limit exceeded", "path=%s", path)
	}
}

type modelPlazaRouteSettingRepoStub struct {
	values map[string]string
}

func (r *modelPlazaRouteSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, service.ErrSettingNotFound
}

func (r *modelPlazaRouteSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (r *modelPlazaRouteSettingRepoStub) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func (r *modelPlazaRouteSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (r *modelPlazaRouteSettingRepoStub) SetMultiple(_ context.Context, updates map[string]string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	for key, value := range updates {
		r.values[key] = value
	}
	return nil
}

func (r *modelPlazaRouteSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return r.values, nil
}

func (r *modelPlazaRouteSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type modelPlazaRouteUserRepoStub struct {
	service.UserRepository
	users map[int64]*service.User
}

func (r *modelPlazaRouteUserRepoStub) GetByID(_ context.Context, id int64) (*service.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *modelPlazaRouteUserRepoStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *modelPlazaRouteUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}

func TestModelPlazaRoutesFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.JWT.Secret = "model-plaza-route-test-secret-32bytes"
	cfg.JWT.AccessTokenExpireMinutes = 60

	users := &modelPlazaRouteUserRepoStub{users: map[int64]*service.User{
		1: {ID: 1, Email: "user@example.com", Role: "user", Status: service.StatusActive, TokenVersion: 1},
		2: {ID: 2, Email: "admin@example.com", Role: "admin", Status: service.StatusActive, TokenVersion: 1},
	}}
	settingRepo := &modelPlazaRouteSettingRepoStub{}
	settingService := service.NewSettingService(settingRepo, cfg)
	apiKeyService := service.NewAPIKeyService(nil, users, nil, nil, nil, nil, cfg)
	settingHandler := handler.NewSettingHandler(settingService, "test-version")
	settingHandler.SetAPIKeyService(apiKeyService)
	authService := service.NewAuthService(nil, users, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil)
	userService := service.NewUserService(users, nil, nil, nil)

	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterAuthRoutes(
		v1,
		&handler.Handlers{Auth: &handler.AuthHandler{}, Setting: settingHandler},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) { c.Next() }),
		servermiddleware.NewOptionalJWTAuthMiddleware(authService, userService, settingService, nil),
		nil,
		nil,
		settingService,
		cfg,
	)

	userToken, err := authService.GenerateToken(users.users[1])
	require.NoError(t, err)
	adminToken, err := authService.GenerateToken(users.users[2])
	require.NoError(t, err)

	tests := []struct {
		name        string
		enabled     bool
		requireAuth bool
		backendMode bool
		authorize   string
		wantStatus  int
	}{
		{name: "disabled returns not found", wantStatus: http.StatusNotFound},
		{name: "sign in required returns unauthorized", enabled: true, requireAuth: true, wantStatus: http.StatusUnauthorized},
		{name: "invalid JWT does not downgrade to anonymous", enabled: true, authorize: "Bearer invalid", wantStatus: http.StatusUnauthorized},
		{name: "backend mode blocks anonymous", enabled: true, backendMode: true, wantStatus: http.StatusForbidden},
		{name: "backend mode blocks regular user", enabled: true, backendMode: true, authorize: "Bearer " + userToken, wantStatus: http.StatusForbidden},
		{name: "backend mode permits administrator", enabled: true, backendMode: true, authorize: "Bearer " + adminToken, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, settingService.UpdateSettings(context.Background(), &service.SystemSettings{
				ModelPlazaEnabled:     tc.enabled,
				ModelPlazaRequireAuth: tc.requireAuth,
				BackendModeEnabled:    tc.backendMode,
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/model-market/catalog", nil)
			if tc.authorize != "" {
				req.Header.Set("Authorization", tc.authorize)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
