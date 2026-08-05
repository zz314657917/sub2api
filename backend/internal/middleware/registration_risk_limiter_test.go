package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type fakeRegistrationRiskRedis struct {
	mu      sync.Mutex
	values  map[string]int
	expires map[string]time.Duration
	err     error
}

func newFakeRegistrationRiskRedis() *fakeRegistrationRiskRedis {
	return &fakeRegistrationRiskRedis{
		values:  make(map[string]int),
		expires: make(map[string]time.Duration),
	}
}

func (f *fakeRegistrationRiskRedis) Incr(_ context.Context, key string) *redis.IntCmd {
	if f.err != nil {
		return redis.NewIntResult(0, f.err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.values[key]++
	return redis.NewIntResult(int64(f.values[key]), nil)
}

func (f *fakeRegistrationRiskRedis) Decr(_ context.Context, key string) *redis.IntCmd {
	if f.err != nil {
		return redis.NewIntResult(0, f.err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.values[key]--
	return redis.NewIntResult(int64(f.values[key]), nil)
}

func (f *fakeRegistrationRiskRedis) Expire(_ context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	if f.err != nil {
		return redis.NewBoolResult(false, f.err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expires[key] = expiration
	return redis.NewBoolResult(true, nil)
}

func (f *fakeRegistrationRiskRedis) Get(_ context.Context, key string) *redis.StringCmd {
	if f.err != nil {
		return redis.NewStringResult("", f.err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	value, ok := f.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(strconv.Itoa(value), nil)
}

func (f *fakeRegistrationRiskRedis) set(key string, value int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.values[key] = value
}

type fakeRegistrationRiskConfigProvider struct {
	cfg config.RegistrationRiskLimitConfig
}

func (f *fakeRegistrationRiskConfigProvider) GetRegistrationRiskLimitConfig(context.Context) config.RegistrationRiskLimitConfig {
	return f.cfg
}

func newRegistrationRiskTestRouter(redis registrationRiskRedis, cfg config.RegistrationRiskLimitConfig, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limiter := &RegistrationRiskLimiter{redis: redis, cfg: cfg}
	router.POST("/register", limiter.LimitRegistrationEntry(), limiter.ReserveMarkedSuccessfulRegistration(), handler)
	return router
}

func performRegistrationRiskRequest(router *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.20:12345"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "risk-test-agent")
	req.Header.Set("CF-Connecting-IP", "203.0.113.20")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRegistrationRiskLimiterBlocksSuccessfulRegistrationsPerIP(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:                      true,
		SuccessfulRegistrationsPerIP: 1,
		WindowHours:                  24,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		MarkRegistrationCreated(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	first := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusOK, first.Code)

	second := performRegistrationRiskRequest(router, `{"email":"second@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Contains(t, second.Body.String(), "REGISTRATION_IP_LIMIT_EXCEEDED")
}

func TestRegistrationRiskLimiterReleasesUnmarkedSuccessfulResponse(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:                      true,
		SuccessfulRegistrationsPerIP: 1,
		WindowHours:                  24,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pending": true})
	})

	first := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusOK, first.Code)
	second := performRegistrationRiskRequest(router, `{"email":"second@example.com"}`)
	require.Equal(t, http.StatusOK, second.Code)

	count, err := redis.Get(context.Background(), successfulRegistrationIPKey("203.0.113.20")).Int()
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestRegistrationRiskLimiterKeepsMarkedSuccessfulResponse(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:                      true,
		SuccessfulRegistrationsPerIP: 2,
		WindowHours:                  24,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		MarkRegistrationCreated(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	first := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusOK, first.Code)

	count, err := redis.Get(context.Background(), successfulRegistrationIPKey("203.0.113.20")).Int()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRegistrationRiskLimiterKeepsMarkedFailedResponse(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:                      true,
		SuccessfulRegistrationsPerIP: 2,
		WindowHours:                  24,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		MarkRegistrationCreated(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
	})

	w := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	count, err := redis.Get(context.Background(), successfulRegistrationIPKey("203.0.113.20")).Int()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRegistrationRiskLimiterBlocksIPUserAgentAttempts(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:             true,
		IPUserAgentAttempts: 1,
		ShortWindowSeconds:  600,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "business validation failed"})
	})

	first := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusBadRequest, first.Code)

	second := performRegistrationRiskRequest(router, `{"email":"second@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Contains(t, second.Body.String(), "REGISTRATION_RISK_LIMIT_EXCEEDED")
}

func TestRegistrationRiskLimiterBlocksEmailDomainAttempts(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:             true,
		EmailDomainAttempts: 1,
		ShortWindowSeconds:  600,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "business validation failed"})
	})

	first := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusBadRequest, first.Code)

	second := performRegistrationRiskRequest(router, `{"email":"second@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Contains(t, second.Body.String(), "REGISTRATION_RISK_LIMIT_EXCEEDED")
}

func TestRegistrationRiskLimiterDisabledByConfig(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:             false,
		IPUserAgentAttempts: 1,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "business validation failed"})
	})

	require.Equal(t, http.StatusBadRequest, performRegistrationRiskRequest(router, `{"email":"first@example.com"}`).Code)
	require.Equal(t, http.StatusBadRequest, performRegistrationRiskRequest(router, `{"email":"second@example.com"}`).Code)
}

func TestRegistrationRiskLimiterUsesDynamicProviderConfig(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	provider := &fakeRegistrationRiskConfigProvider{
		cfg: config.RegistrationRiskLimitConfig{
			Enabled:             true,
			IPUserAgentAttempts: 1,
			ShortWindowSeconds:  60,
		},
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limiter := &RegistrationRiskLimiter{
		redis: redis,
		cfg: config.RegistrationRiskLimitConfig{
			Enabled: false,
		},
		provider: provider,
	}
	router.POST("/register", limiter.LimitRegistrationEntry(), func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "business validation failed"})
	})

	require.Equal(t, http.StatusBadRequest, performRegistrationRiskRequest(router, `{"email":"first@example.com"}`).Code)
	second := performRegistrationRiskRequest(router, `{"email":"second@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, second.Code)
	require.Contains(t, second.Body.String(), "REGISTRATION_RISK_LIMIT_EXCEEDED")
}

func TestRegistrationRiskLimiterFailCloseOnRedisError(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	redis.err = context.DeadlineExceeded
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:             true,
		IPUserAgentAttempts: 1,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		MarkRegistrationCreated(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := performRegistrationRiskRequest(router, `{"email":"first@example.com"}`)
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "REGISTRATION_RISK_LIMIT_EXCEEDED")
}

func TestRegistrationRiskLimiterStoresExpectedTTLs(t *testing.T) {
	redis := newFakeRegistrationRiskRedis()
	cfg := config.RegistrationRiskLimitConfig{
		Enabled:                      true,
		SuccessfulRegistrationsPerIP: 10,
		WindowHours:                  12,
		IPUserAgentAttempts:          10,
		ShortWindowSeconds:           120,
	}
	router := newRegistrationRiskTestRouter(redis, cfg, func(c *gin.Context) {
		MarkRegistrationCreated(c)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	require.Equal(t, http.StatusOK, performRegistrationRiskRequest(router, `{"email":"first@example.com"}`).Code)

	var hasShortTTL bool
	var hasSuccessTTL bool
	for key, ttl := range redis.expires {
		if strings.Contains(key, "ip_ua_attempt") && ttl == 120*time.Second {
			hasShortTTL = true
		}
		if strings.Contains(key, "success_ip") && ttl == 12*time.Hour {
			hasSuccessTTL = true
		}
	}
	require.True(t, hasShortTTL)
	require.True(t, hasSuccessTTL)
}
