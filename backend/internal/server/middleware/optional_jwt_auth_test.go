package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type optionalJWTUserRepoStub struct {
	service.UserRepository
	users map[int64]*service.User
}

func (r *optionalJWTUserRepoStub) GetByID(_ context.Context, id int64) (*service.User, error) {
	return r.users[id], nil
}

func (r *optionalJWTUserRepoStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *optionalJWTUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}

func TestOptionalJWTAuth_AnonymousAllowedAndInvalidJWTRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-jwt-secret-32bytes-long!!!"
	cfg.JWT.AccessTokenExpireMinutes = 60
	user := &service.User{ID: 1, Email: "user@example.com", Role: "user", Status: service.StatusActive, TokenVersion: 1}
	userRepo := &optionalJWTUserRepoStub{users: map[int64]*service.User{user.ID: user}}
	authService := service.NewAuthService(nil, userRepo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil)
	userService := service.NewUserService(userRepo, nil, nil, nil)

	router := gin.New()
	router.Use(gin.HandlerFunc(NewOptionalJWTAuthMiddleware(authService, userService, nil, nil)))
	router.GET("/catalog", func(c *gin.Context) {
		_, authenticated := GetAuthSubjectFromContext(c)
		c.JSON(http.StatusOK, gin.H{"authenticated": authenticated})
	})

	anonymousRecorder := httptest.NewRecorder()
	router.ServeHTTP(anonymousRecorder, httptest.NewRequest(http.MethodGet, "/catalog", nil))
	require.Equal(t, http.StatusOK, anonymousRecorder.Code)
	var anonymousBody map[string]bool
	require.NoError(t, json.Unmarshal(anonymousRecorder.Body.Bytes(), &anonymousBody))
	require.False(t, anonymousBody["authenticated"])

	invalidRecorder := httptest.NewRecorder()
	invalidRequest := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	invalidRequest.Header.Set("Authorization", "Bearer invalid")
	router.ServeHTTP(invalidRecorder, invalidRequest)
	require.Equal(t, http.StatusUnauthorized, invalidRecorder.Code)
	var invalidBody ErrorResponse
	require.NoError(t, json.Unmarshal(invalidRecorder.Body.Bytes(), &invalidBody))
	require.Equal(t, "INVALID_TOKEN", invalidBody.Code)
}
