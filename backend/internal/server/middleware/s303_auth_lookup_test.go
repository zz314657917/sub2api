//go:build unit

package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type s303JWTUserReader struct {
	user *service.User
	err  error
}

func (r s303JWTUserReader) GetByID(context.Context, int64) (*service.User, error) {
	return r.user, r.err
}

func newS303AuthService() *service.AuthService {
	cfg := &config.Config{}
	cfg.JWT.Secret = "s303-test-secret"
	cfg.JWT.AccessTokenExpireMinutes = 60
	return service.NewAuthService(nil, nil, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil)
}

func s303ActiveUser(role string) *service.User {
	return &service.User{ID: 303, Email: "s303@example.com", Role: role, Status: service.StatusActive, TokenVersion: 1, TokenVersionResolved: true}
}

func TestS303JWTAuthLookupOutcomes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := newS303AuthService()
	token, err := authService.GenerateToken(s303ActiveUser("user"))
	require.NoError(t, err)

	tests := []struct {
		name       string
		reader     s303JWTUserReader
		wantStatus int
		wantCode   string
	}{
		{"active_user_reaches_handler", s303JWTUserReader{user: s303ActiveUser("user")}, http.StatusOK, ""},
		{"direct_not_found_is_unauthorized", s303JWTUserReader{err: service.ErrUserNotFound}, http.StatusUnauthorized, "USER_NOT_FOUND"},
		{"wrapped_not_found_is_unauthorized", s303JWTUserReader{err: fmt.Errorf("lookup: %w", service.ErrUserNotFound)}, http.StatusUnauthorized, "USER_NOT_FOUND"},
		{"timeout_is_internal_error", s303JWTUserReader{err: context.DeadlineExceeded}, http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"internal_error_is_internal_error", s303JWTUserReader{err: errors.New("database unavailable")}, http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reached := false
			router := gin.New()
			router.Use(jwtAuth(authService, tt.reader, nil))
			router.GET("/protected", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			require.Equalf(t, tt.wantStatus, response.Code, "body: %s", response.Body.String())
			require.Equal(t, tt.wantStatus == http.StatusOK, reached)
			if tt.wantCode != "" {
				require.Contains(t, response.Body.String(), tt.wantCode)
			}
		})
	}
}

func TestS303AdminAuthLookupOutcomesHTTPAndWebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := newS303AuthService()
	adminToken, err := authService.GenerateToken(s303ActiveUser(service.RoleAdmin))
	require.NoError(t, err)
	memberToken, err := authService.GenerateToken(s303ActiveUser("user"))
	require.NoError(t, err)

	tests := []struct {
		name       string
		user       *service.User
		err        error
		token      string
		wantStatus int
		wantCode   string
	}{
		{"active_admin_reaches_handler", s303ActiveUser(service.RoleAdmin), nil, adminToken, http.StatusOK, ""},
		{"direct_not_found_is_unauthorized", nil, service.ErrUserNotFound, adminToken, http.StatusUnauthorized, "USER_NOT_FOUND"},
		{"wrapped_not_found_is_unauthorized", nil, errors.Join(errors.New("lookup failed"), service.ErrUserNotFound), adminToken, http.StatusUnauthorized, "USER_NOT_FOUND"},
		{"timeout_is_internal_error", nil, context.DeadlineExceeded, adminToken, http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"internal_error_is_internal_error", nil, errors.New("database unavailable"), adminToken, http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"nonadmin_is_forbidden", s303ActiveUser("user"), nil, memberToken, http.StatusForbidden, "FORBIDDEN"},
	}

	for _, transport := range []string{"http", "websocket"} {
		for _, tt := range tests {
			t.Run(transport+"/"+tt.name, func(t *testing.T) {
				userRepo := &stubUserRepo{getByID: func(context.Context, int64) (*service.User, error) { return tt.user, tt.err }}
				userService := service.NewUserService(userRepo, nil, nil, nil)
				reached := false
				router := gin.New()
				router.Use(adminAuth(authService, userService, nil))
				router.GET("/admin", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })
				request := httptest.NewRequest(http.MethodGet, "/admin", nil)
				if transport == "websocket" {
					request.Header.Set("Upgrade", "websocket")
					request.Header.Set("Connection", "Upgrade")
					request.Header.Set("Sec-WebSocket-Protocol", "sub2api-admin, jwt."+tt.token)
				} else {
					request.Header.Set("Authorization", "Bearer "+tt.token)
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)

				require.Equal(t, tt.wantStatus, response.Code)
				require.Equal(t, tt.wantStatus == http.StatusOK, reached)
				if tt.wantCode != "" {
					require.Contains(t, response.Body.String(), tt.wantCode)
				}
			})
		}
	}
}
