//go:build unit

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestS99SkippableSubscriptionRouteErrors(t *testing.T) {
	require.True(t, isSkippableSubscriptionRouteError(service.ErrSubscriptionNotFound))
	require.True(t, isSkippableSubscriptionRouteError(service.ErrSubscriptionExpired))
	require.True(t, isSkippableSubscriptionRouteError(service.ErrSubscriptionSuspended))
	require.False(t, isSkippableSubscriptionRouteError(service.ErrDailyLimitExceeded))
	require.False(t, isSkippableSubscriptionRouteError(errors.New("temporary database failure")))
}

func TestS99UnavailableSubscriptionAnnotationDoesNotMutateCachedAPIKey(t *testing.T) {
	groupID := int64(11)
	key := s99AuthAPIKey(groupID, 0)
	cfg := &config.Config{RunMode: config.RunModeStandard}
	subscriptionService := service.NewSubscriptionService(nil, &stubUserSubscriptionRepo{
		getActive: func(context.Context, int64, int64) (*service.UserSubscription, error) {
			return nil, service.ErrSubscriptionNotFound
		},
	}, nil, nil, cfg)
	t.Cleanup(subscriptionService.Stop)

	annotated := withUnavailableSubscriptionRouteGroups(context.Background(), key, subscriptionService)

	require.NotSame(t, key, annotated)
	require.Empty(t, key.UnavailableRouteGroupIDs)
	require.True(t, annotated.IsRouteGroupUnavailable(groupID))
}

func TestS99APIKeyAuthSkipsUnavailableSubscriptionRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	primaryID := int64(11)
	fallbackID := int64(22)
	key := s99AuthAPIKey(primaryID, fallbackID)
	cfg := &config.Config{RunMode: config.RunModeStandard}
	apiKeyService := service.NewAPIKeyService(&stubApiKeyRepo{
		getByKey: func(context.Context, string) (*service.APIKey, error) {
			clone := *key
			return &clone, nil
		},
	}, nil, nil, nil, nil, nil, cfg)
	subscriptionService := service.NewSubscriptionService(nil, &stubUserSubscriptionRepo{
		getActive: func(_ context.Context, _ int64, groupID int64) (*service.UserSubscription, error) {
			if groupID == primaryID {
				return nil, service.ErrSubscriptionNotFound
			}
			return nil, errors.New("unexpected subscription lookup")
		},
	}, nil, nil, cfg)
	t.Cleanup(subscriptionService.Stop)

	router := gin.New()
	router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(apiKeyService, subscriptionService, cfg)))
	router.GET("/t", func(c *gin.Context) {
		resolved, ok := GetAPIKeyFromContext(c)
		if !ok || resolved.GroupID == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"group_id": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"group_id": *resolved.GroupID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set("x-api-key", key.Key)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"group_id":22`)
}

func TestS99APIKeyAuthDoesNotFailOverOnUsageLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	primaryID := int64(11)
	fallbackID := int64(22)
	key := s99AuthAPIKey(primaryID, fallbackID)
	limit := 1.0
	key.Group.DailyLimitUSD = &limit
	now := time.Now()
	subscription := &service.UserSubscription{
		ID: 33, UserID: key.UserID, GroupID: primaryID,
		Status: service.SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour),
		DailyWindowStart: &now, DailyUsageUSD: limit + 1,
	}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	apiKeyService := service.NewAPIKeyService(&stubApiKeyRepo{
		getByKey: func(context.Context, string) (*service.APIKey, error) {
			clone := *key
			return &clone, nil
		},
	}, nil, nil, nil, nil, nil, cfg)
	subscriptionService := service.NewSubscriptionService(nil, &stubUserSubscriptionRepo{
		getActive: func(context.Context, int64, int64) (*service.UserSubscription, error) {
			clone := *subscription
			return &clone, nil
		},
	}, nil, nil, cfg)
	t.Cleanup(subscriptionService.Stop)

	router := newAuthTestRouter(apiKeyService, subscriptionService, cfg)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set("x-api-key", key.Key)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "USAGE_LIMIT_EXCEEDED")
}

func s99AuthAPIKey(primaryID, fallbackID int64) *service.APIKey {
	primary := &service.Group{
		ID: primaryID, Name: "expired subscription", Status: service.StatusActive,
		Hydrated: true, Platform: service.PlatformOpenAI,
		SubscriptionType: service.SubscriptionTypeSubscription,
		RoutingScope:     service.GroupRoutingScopeInference,
	}
	routes := []domain.APIKeyMultiGroupRoute{{
		GroupID: primaryID, Priority: 1, Weight: 1, Enabled: true,
	}}
	routeGroups := []*service.Group{}
	if fallbackID > 0 {
		routes = append(routes, domain.APIKeyMultiGroupRoute{
			GroupID: fallbackID, Priority: 2, Weight: 1, Enabled: true,
		})
		routeGroups = append(routeGroups, &service.Group{
			ID: fallbackID, Name: "fallback", Status: service.StatusActive,
			Hydrated: true, Platform: service.PlatformOpenAI,
			SubscriptionType: service.SubscriptionTypeStandard,
			RoutingScope:     service.GroupRoutingScopeInference,
		})
	}
	return &service.APIKey{
		ID: 100, UserID: 7, Key: "s99-key", Status: service.StatusActive,
		GroupID: &primaryID, Group: primary,
		MultiGroupRoutes: routes, MultiGroupRouteGroups: routeGroups,
		User: &service.User{
			ID: 7, Role: service.RoleUser, Status: service.StatusActive,
			Balance: 10, Concurrency: 1,
		},
	}
}
