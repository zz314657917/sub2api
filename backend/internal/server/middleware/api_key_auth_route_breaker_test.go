package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type apiKeyRouteBreakerMiddlewareCacheStub struct {
	service.APIKeyCache
	failures  int
	successes int
	releases  int
	acquires  int
	lease     *service.APIKeyRouteBreakerLease
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) IsRouteGroupCooling(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) AcquireAPIKeyRouteBreaker(context.Context, service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
	s.acquires++
	if s.lease == nil {
		return nil, nil
	}
	lease := *s.lease
	return &lease, nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) SetRouteGroupCooldown(context.Context, int64, int64, time.Duration) error {
	return nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) DeleteRouteGroupCooldown(context.Context, int64, int64) error {
	return nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) RecordAPIKeyRouteBreakerSuccess(context.Context, service.APIKeyRouteBreakerLease) error {
	s.successes++
	return nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) RecordAPIKeyRouteBreakerFailure(context.Context, service.APIKeyRouteBreakerLease) error {
	s.failures++
	return nil
}

func (s *apiKeyRouteBreakerMiddlewareCacheStub) ReleaseAPIKeyRouteBreakerProbe(context.Context, service.APIKeyRouteBreakerLease) error {
	s.releases++
	return nil
}

func TestAPIKeyRouteBreakerStreamFailureAndBusiness4xx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(4)
	breakerKey, ok := service.NewAPIKeyRouteBreakerKey(groupID, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	cache := &apiKeyRouteBreakerMiddlewareCacheStub{}
	svc := service.NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	newKey := func(halfOpen bool) *service.APIKey {
		return &service.APIKey{
			ID:      7,
			GroupID: &groupID,
			Group:   &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, Hydrated: true},
			MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
				GroupID: groupID, Enabled: true, CooldownSeconds: 30,
			}},
			RouteBreakerLease: &service.APIKeyRouteBreakerLease{Key: breakerKey, Generation: 1, ProbeToken: 9, HalfOpen: halfOpen},
		}
	}

	streamRecorder := httptest.NewRecorder()
	streamContext, _ := gin.CreateTestContext(streamRecorder)
	streamContext.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	streamContext.Status(http.StatusOK)
	MarkAPIKeyRouteCooldown(streamContext, http.StatusTooManyRequests)
	applyAPIKeyRouteCooldownAfterRequest(streamContext, svc, newKey(false))
	require.Equal(t, 1, cache.failures)
	require.Equal(t, 0, cache.successes)
	require.Equal(t, 0, cache.releases)

	businessRecorder := httptest.NewRecorder()
	businessContext, _ := gin.CreateTestContext(businessRecorder)
	businessContext.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	businessContext.Status(http.StatusBadRequest)
	applyAPIKeyRouteCooldownAfterRequest(businessContext, svc, newKey(true))
	require.Equal(t, 1, cache.failures)
	require.Equal(t, 1, cache.releases)

	unknownRecorder := httptest.NewRecorder()
	unknownContext, _ := gin.CreateTestContext(unknownRecorder)
	unknownContext.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	unknownContext.Status(http.StatusInternalServerError)
	applyAPIKeyRouteCooldownAfterRequest(unknownContext, svc, newKey(false))
	require.Equal(t, 1, cache.failures)
}

func TestAPIKeyRouteBreakerUsesOriginalUpstreamStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(5)
	key, ok := service.NewAPIKeyRouteBreakerKey(groupID, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	cache := &apiKeyRouteBreakerMiddlewareCacheStub{}
	svc := service.NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	newLeaseKey := func() *service.APIKey {
		return &service.APIKey{
			GroupID:           &groupID,
			Group:             &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, Hydrated: true},
			MultiGroupRoutes:  []domain.APIKeyMultiGroupRoute{{GroupID: groupID, Enabled: true, CooldownSeconds: 30}},
			RouteBreakerLease: &service.APIKeyRouteBreakerLease{Key: key, Generation: 1, HalfOpen: true, ProbeToken: 7},
		}
	}
	for _, tc := range []struct {
		name           string
		upstreamStatus int
		finalStatus    int
		wantFailures   int
		wantReleases   int
	}{
		{name: "client_403_mapped_502", upstreamStatus: http.StatusForbidden, finalStatus: http.StatusBadGateway, wantFailures: 0, wantReleases: 1},
		{name: "upstream_503_mapped_502", upstreamStatus: http.StatusServiceUnavailable, finalStatus: http.StatusBadGateway, wantFailures: 1, wantReleases: 0},
		{name: "intermediate_503_final_success", upstreamStatus: http.StatusServiceUnavailable, finalStatus: http.StatusOK, wantFailures: 0, wantReleases: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			beforeFailures, beforeReleases := cache.failures, cache.releases
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			c.Set(service.OpsUpstreamStatusCodeKey, tc.upstreamStatus)
			c.Status(tc.finalStatus)
			applyAPIKeyRouteCooldownAfterRequest(c, svc, newLeaseKey())
			require.Equal(t, tc.wantFailures, cache.failures-beforeFailures)
			require.Equal(t, tc.wantReleases, cache.releases-beforeReleases)
		})
	}
}

func TestResolveAPIKeyForModelRequestReusesRequestRouteDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(14)
	breakerKey, ok := service.NewAPIKeyRouteBreakerKey(groupID, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	cache := &apiKeyRouteBreakerMiddlewareCacheStub{
		lease: &service.APIKeyRouteBreakerLease{
			Key:        breakerKey,
			Generation: 3,
			ProbeToken: 7,
			HalfOpen:   true,
		},
	}
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	key := &service.APIKey{
		ID:      9,
		Key:     "request-scoped-key",
		Status:  service.StatusActive,
		GroupID: &groupID,
		Group: &service.Group{
			ID:                 groupID,
			Platform:           service.PlatformOpenAI,
			Status:             service.StatusActive,
			Hydrated:           true,
			ModelMatchPatterns: []string{"gpt-*"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:  groupID,
			Enabled:  true,
			Priority: 1,
			Weight:   1,
		}},
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)

	first, firstOK := ResolveAPIKeyForModelRequest(c, apiKeyService, key, "gpt-5.6", false)
	require.True(t, firstOK)
	require.NotNil(t, first)
	require.NotNil(t, first.RouteBreakerLease)

	// Responses compact normalization may rewrite the path after the gateway
	// wrapper has already selected a route; it must not trigger a second acquire.
	c.Request.URL.Path = "/v1/responses/compact"
	second, secondOK := ResolveAPIKeyForModelRequest(c, apiKeyService, first, "gpt-5.6", true)
	require.True(t, secondOK)
	require.Same(t, first, second)
	require.Equal(t, 1, cache.acquires, "the same request must acquire one breaker lease")
}
