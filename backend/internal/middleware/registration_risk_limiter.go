package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	registrationRiskKeyPrefix          = "registration_risk:"
	registrationCreatedContextKey      = "registration_risk.account_created"
	defaultRegistrationRiskWindow      = 24 * time.Hour
	defaultRegistrationRiskShortWindow = 10 * time.Minute
)

var (
	ErrRegistrationIPLimitExceeded = infraerrors.TooManyRequests(
		"REGISTRATION_IP_LIMIT_EXCEEDED",
		"too many accounts registered from this IP, please try again later",
	)
	ErrRegistrationRiskLimitExceeded = infraerrors.TooManyRequests(
		"REGISTRATION_RISK_LIMIT_EXCEEDED",
		"too many registration attempts, please try again later",
	)
)

type registrationRiskRequest struct {
	Email string `json:"email"`
}

type registrationRiskRedis interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type RegistrationRiskConfigProvider interface {
	GetRegistrationRiskLimitConfig(ctx context.Context) config.RegistrationRiskLimitConfig
}

type RegistrationRiskLimiter struct {
	redis    registrationRiskRedis
	cfg      config.RegistrationRiskLimitConfig
	provider RegistrationRiskConfigProvider
}

func NewRegistrationRiskLimiter(redisClient *redis.Client, cfg config.RegistrationRiskLimitConfig, providers ...RegistrationRiskConfigProvider) *RegistrationRiskLimiter {
	limiter := &RegistrationRiskLimiter{
		redis: redisClient,
		cfg:   cfg,
	}
	if len(providers) > 0 {
		limiter.provider = providers[0]
	}
	return limiter
}

func (l *RegistrationRiskLimiter) LimitRegistrationEntry() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := l.effectiveConfig(c.Request.Context())
		if !l.enabled(cfg) {
			c.Next()
			return
		}
		if l.redis == nil {
			response.ErrorFrom(c, ErrRegistrationRiskLimitExceeded)
			c.Abort()
			return
		}

		body, err := readAndRestoreRequestBody(c)
		if err != nil {
			response.BadRequest(c, "Invalid request: "+err.Error())
			c.Abort()
			return
		}

		clientIP := ip.GetClientIP(c)
		userAgent := c.GetHeader("User-Agent")
		emailDomain := registrationRiskEmailDomain(body)

		if err := l.checkSuccessfulRegistrationIPLimit(c.Request.Context(), cfg, clientIP); err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if err := l.countShortWindow(c.Request.Context(), cfg, "ip_ua_attempt", hashParts(clientIP, userAgent), cfg.IPUserAgentAttempts); err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if emailDomain != "" {
			if err := l.countShortWindow(c.Request.Context(), cfg, "email_domain_attempt", hashParts(emailDomain), cfg.EmailDomainAttempts); err != nil {
				response.ErrorFrom(c, err)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func MarkRegistrationCreated(c *gin.Context) {
	if c != nil {
		c.Set(registrationCreatedContextKey, true)
	}
}

func (l *RegistrationRiskLimiter) ReserveMarkedSuccessfulRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := l.effectiveConfig(c.Request.Context())
		if !l.enabled(cfg) || l.redis == nil {
			c.Next()
			return
		}
		limit := cfg.SuccessfulRegistrationsPerIP
		if limit <= 0 {
			c.Next()
			return
		}
		clientIP := ip.GetClientIP(c)
		if strings.TrimSpace(clientIP) == "" {
			c.Next()
			return
		}

		key := successfulRegistrationIPKey(clientIP)
		count, err := l.incrWithTTL(c.Request.Context(), key, l.successWindow(cfg)).Result()
		if err != nil {
			log.Printf("[RegistrationRisk] failed to reserve successful registration: %v", err)
			response.ErrorFrom(c, ErrRegistrationRiskLimitExceeded)
			c.Abort()
			return
		}
		if count > int64(limit) {
			l.releaseSuccessfulRegistration(c.Request.Context(), key)
			response.ErrorFrom(c, ErrRegistrationIPLimitExceeded)
			c.Abort()
			return
		}

		c.Next()

		created, _ := c.Get(registrationCreatedContextKey)
		if created != true {
			l.releaseSuccessfulRegistration(c.Request.Context(), key)
		}
	}
}

func (l *RegistrationRiskLimiter) effectiveConfig(ctx context.Context) config.RegistrationRiskLimitConfig {
	if l == nil {
		return config.RegistrationRiskLimitConfig{}
	}
	if l.provider != nil {
		return l.provider.GetRegistrationRiskLimitConfig(ctx)
	}
	return l.cfg
}

func (l *RegistrationRiskLimiter) enabled(cfg config.RegistrationRiskLimitConfig) bool {
	return l != nil && cfg.Enabled
}

func (l *RegistrationRiskLimiter) checkSuccessfulRegistrationIPLimit(ctx context.Context, cfg config.RegistrationRiskLimitConfig, clientIP string) error {
	limit := cfg.SuccessfulRegistrationsPerIP
	if limit <= 0 || strings.TrimSpace(clientIP) == "" {
		return nil
	}
	count, err := l.redis.Get(ctx, successfulRegistrationIPKey(clientIP)).Int()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		log.Printf("[RegistrationRisk] redis get error: %v", err)
		return ErrRegistrationRiskLimitExceeded
	}
	if count >= limit {
		return ErrRegistrationIPLimitExceeded
	}
	return nil
}

func (l *RegistrationRiskLimiter) countShortWindow(ctx context.Context, cfg config.RegistrationRiskLimitConfig, kind, identity string, limit int) error {
	if limit <= 0 || strings.TrimSpace(identity) == "" {
		return nil
	}
	count, err := l.incrWithTTL(ctx, registrationRiskKey(kind, identity), l.shortWindow(cfg)).Result()
	if err != nil {
		log.Printf("[RegistrationRisk] redis incr error: %v", err)
		return ErrRegistrationRiskLimitExceeded
	}
	if count > int64(limit) {
		return ErrRegistrationRiskLimitExceeded
	}
	return nil
}

func (l *RegistrationRiskLimiter) incrWithTTL(ctx context.Context, key string, ttl time.Duration) *redis.IntCmd {
	cmd := l.redis.Incr(ctx, key)
	if cmd.Err() != nil {
		return cmd
	}
	if cmd.Val() == 1 {
		if err := l.redis.Expire(ctx, key, ttl).Err(); err != nil {
			return redis.NewIntResult(cmd.Val(), err)
		}
	}
	return cmd
}

func (l *RegistrationRiskLimiter) releaseSuccessfulRegistration(ctx context.Context, key string) {
	if l == nil || l.redis == nil || strings.TrimSpace(key) == "" {
		return
	}
	if err := l.redis.Decr(ctx, key).Err(); err != nil {
		log.Printf("[RegistrationRisk] failed to release successful registration reservation: %v", err)
	}
}

func (l *RegistrationRiskLimiter) successWindow(cfg config.RegistrationRiskLimitConfig) time.Duration {
	if l == nil || cfg.WindowHours <= 0 {
		return defaultRegistrationRiskWindow
	}
	return time.Duration(cfg.WindowHours) * time.Hour
}

func (l *RegistrationRiskLimiter) shortWindow(cfg config.RegistrationRiskLimitConfig) time.Duration {
	if l == nil || cfg.ShortWindowSeconds <= 0 {
		return defaultRegistrationRiskShortWindow
	}
	return time.Duration(cfg.ShortWindowSeconds) * time.Second
}

func readAndRestoreRequestBody(c *gin.Context) ([]byte, error) {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func registrationRiskEmailDomain(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var req registrationRiskRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return ""
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return ""
	}
	return email[at+1:]
}

func successfulRegistrationIPKey(clientIP string) string {
	return registrationRiskKey("success_ip", hashParts(clientIP))
}

func registrationRiskKey(kind, identity string) string {
	return registrationRiskKeyPrefix + kind + ":" + identity
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(strings.TrimSpace(part)))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
