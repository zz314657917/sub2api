package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	openAIOAuth429FallbackCooldown     = 5 * time.Second
	openAIOAuth429RetryWindow          = 2 * time.Minute
	openAIOAuth429RetryDelay           = 500 * time.Millisecond
	openAIOAuth429MaxRetryDelay        = 8 * time.Second
	openAIOAuth429MaxAccountAttempts   = 3
	openAIStopSchedulingBridgeCooldown = 2 * time.Minute
)

type openAIOAuth429Disposition uint8

const (
	openAIOAuth429Transient openAIOAuth429Disposition = iota
	openAIOAuth429Quota5h
	openAIOAuth429Quota7d
	openAIOAuth429QuotaReset
)

// isOpenAIOAuthAccount deliberately follows the local account predicate. The
// setup-token type is OAuth-like for some credential flows, but
// Account.IsOpenAIOAuth currently scopes this fast path to the standard
// OpenAI OAuth account type; widening it here would change scheduling
// semantics outside this upstream fix.
func isOpenAIOAuthAccount(account *Account) bool {
	return account != nil && account.IsOpenAIOAuth()
}

// classifyOpenAIOAuth429 distinguishes a short-lived OAuth 429 from an
// exhausted account quota. Codex window headers take precedence over generic
// reset signals so a 7d/5h quota is paused until its authoritative reset.
func classifyOpenAIOAuth429(headers http.Header, responseBody []byte) (openAIOAuth429Disposition, *time.Time) {
	if snapshot := ParseCodexRateLimitHeaders(headers); snapshot != nil {
		if normalized := snapshot.Normalize(); normalized != nil {
			if normalized.Used7dPercent != nil && *normalized.Used7dPercent >= 100 {
				if normalized.Reset7dSeconds != nil && *normalized.Reset7dSeconds > 0 {
					now := time.Now()
					resetAt := now.Add(time.Duration(*normalized.Reset7dSeconds) * time.Second)
					return openAIOAuth429Quota7d, &resetAt
				}
				return openAIOAuth429Quota7d, nil
			}
			if normalized.Used5hPercent != nil && *normalized.Used5hPercent >= 100 {
				if normalized.Reset5hSeconds != nil && *normalized.Reset5hSeconds > 0 {
					now := time.Now()
					resetAt := now.Add(time.Duration(*normalized.Reset5hSeconds) * time.Second)
					return openAIOAuth429Quota5h, &resetAt
				}
				return openAIOAuth429Quota5h, nil
			}
		}
	}
	if resetAt := calculateOpenAI429ResetTime(headers); resetAt != nil {
		return openAIOAuth429QuotaReset, resetAt
	}
	if resetUnix := parseOpenAIRateLimitResetTime(responseBody); resetUnix != nil {
		resetAt := time.Unix(*resetUnix, 0)
		return openAIOAuth429QuotaReset, &resetAt
	}
	return openAIOAuth429Transient, nil
}

func isOpenAIAccountRuntimeManaged(account *Account) bool {
	return account != nil && (account.Platform == PlatformOpenAI || account.Platform == PlatformGrok)
}

func (s *OpenAIGatewayService) openAIAccountRuntimeBlockLock(accountID int64) *sync.Mutex {
	if s == nil || accountID <= 0 {
		return nil
	}
	actual, _ := s.openaiAccountRuntimeBlockLocks.LoadOrStore(accountID, &sync.Mutex{})
	mu, ok := actual.(*sync.Mutex)
	if !ok {
		mu = &sync.Mutex{}
		s.openaiAccountRuntimeBlockLocks.Store(accountID, mu)
	}
	return mu
}

// BlockAccountScheduling installs or extends an in-process account pause. It
// never shortens an existing pause, which matters when a short transient signal
// races with a longer quota reset.
func (s *OpenAIGatewayService) BlockAccountScheduling(account *Account, until time.Time, _ string) {
	if s == nil || !isOpenAIAccountRuntimeManaged(account) || account.ID <= 0 {
		return
	}
	if !until.After(time.Now()) {
		until = time.Now().Add(openAIStopSchedulingBridgeCooldown)
	}
	mu := s.openAIAccountRuntimeBlockLock(account.ID)
	if mu == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if current, ok := s.openaiAccountRuntimeBlockUntil.Load(account.ID); ok {
		if currentUntil, ok := current.(time.Time); ok && currentUntil.After(until) {
			return
		}
	}
	s.openaiAccountRuntimeBlockUntil.Store(account.ID, until)
}

func (s *OpenAIGatewayService) ClearAccountSchedulingBlock(accountID int64) {
	if s == nil || accountID <= 0 {
		return
	}
	if mu := s.openAIAccountRuntimeBlockLock(accountID); mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	s.openaiAccountRuntimeBlockUntil.Delete(accountID)
	s.openaiOAuth429RetryStartedAt.Delete(accountID)
}

func (s *OpenAIGatewayService) isOpenAIAccountRuntimeBlocked(account *Account) bool {
	if s == nil || !isOpenAIAccountRuntimeManaged(account) || account.ID <= 0 {
		return false
	}
	mu := s.openAIAccountRuntimeBlockLock(account.ID)
	if mu == nil {
		return false
	}
	mu.Lock()
	defer mu.Unlock()
	value, ok := s.openaiAccountRuntimeBlockUntil.Load(account.ID)
	if !ok {
		return false
	}
	until, ok := value.(time.Time)
	if !ok || until.IsZero() || !time.Now().Before(until) {
		s.openaiAccountRuntimeBlockUntil.Delete(account.ID)
		return false
	}
	return true
}

func (s *OpenAIGatewayService) openAIOAuth429RetryWindowActive(account *Account) bool {
	if s == nil || !isOpenAIOAuthAccount(account) || account.IsShadow() {
		return false
	}
	now := time.Now()
	value, _ := s.openaiOAuth429RetryStartedAt.LoadOrStore(account.ID, now)
	startedAt, ok := value.(time.Time)
	if !ok {
		startedAt = now
		s.openaiOAuth429RetryStartedAt.Store(account.ID, startedAt)
	}
	if now.Before(startedAt.Add(openAIOAuth429RetryWindow)) {
		return true
	}
	s.openaiOAuth429RetryStartedAt.Delete(account.ID)
	return false
}

func (s *OpenAIGatewayService) openAIOAuth429RetryDeadline(account *Account) time.Time {
	if s == nil || !isOpenAIOAuthAccount(account) || account.IsShadow() {
		return time.Time{}
	}
	value, ok := s.openaiOAuth429RetryStartedAt.Load(account.ID)
	if !ok {
		return time.Time{}
	}
	startedAt, ok := value.(time.Time)
	if !ok {
		return time.Time{}
	}
	return startedAt.Add(openAIOAuth429RetryWindow)
}

func openAIOAuth429SameAccountRetryDelay(headers http.Header, deadline time.Time) time.Duration {
	delay := openAIOAuth429RetryDelay
	if resetAt := parseRetryAfterResetTime(headers, time.Now()); resetAt != nil && resetAt.After(time.Now()) {
		delay = resetAt.Sub(time.Now())
	}
	if delay > openAIOAuth429MaxRetryDelay {
		delay = openAIOAuth429MaxRetryDelay
	}
	if !deadline.IsZero() {
		remaining := time.Until(deadline)
		if delay > remaining {
			delay = remaining
		}
	}
	if delay < 0 {
		return 0
	}
	return delay
}

// parseRetryAfterResetTime accepts both the delta-seconds and HTTP-date forms
// allowed by RFC 9110. It is intentionally bounded by callers rather than
// clamping here, so an explicit upstream delay can still be compared with the
// request-local retry deadline.
func parseRetryAfterResetTime(headers http.Header, now time.Time) *time.Time {
	if headers == nil {
		return nil
	}
	raw := strings.TrimSpace(headers.Get("Retry-After"))
	if raw == "" {
		return nil
	}
	if seconds, err := strconv.ParseFloat(raw, 64); err == nil {
		resetAt := now.Add(time.Duration(seconds * float64(time.Second)))
		return &resetAt
	}
	if parsed, err := http.ParseTime(raw); err == nil {
		return &parsed
	}
	return nil
}

func (s *OpenAIGatewayService) markOpenAIOAuth429RateLimited(_ context.Context, account *Account, headers http.Header, responseBody []byte) {
	if s == nil || !isOpenAIOAuthAccount(account) || account.IsShadow() {
		return
	}
	s.recordOpenAIOAuth429(account)
	disposition, resetAt := classifyOpenAIOAuth429(headers, responseBody)
	if disposition == openAIOAuth429Transient && s.openAIOAuth429RetryWindowActive(account) {
		return
	}
	cooldownUntil := time.Now().Add(openAIOAuth429FallbackCooldown)
	if resetAt != nil && resetAt.After(time.Now()) {
		cooldownUntil = *resetAt
	} else if s.rateLimitService != nil {
		if cooldown, ok := s.rateLimitService.get429FallbackCooldown(context.Background(), account); ok && cooldown > 0 {
			cooldownUntil = time.Now().Add(cooldown)
		}
	}
	s.BlockAccountScheduling(account, cooldownUntil, "429")
	s.openaiOAuth429RetryStartedAt.Delete(account.ID)
}

func (s *OpenAIGatewayService) shouldRetryOpenAIOAuth429OnSameAccount(account *Account, statusCode int, shouldDisable bool) bool {
	return s.shouldRetryOpenAIOAuth429OnSameAccountWithResponse(account, statusCode, shouldDisable, nil, nil)
}

func (s *OpenAIGatewayService) shouldRetryOpenAIOAuth429OnSameAccountWithResponse(account *Account, statusCode int, shouldDisable bool, headers http.Header, responseBody []byte) bool {
	if shouldDisable || statusCode != http.StatusTooManyRequests || !isOpenAIOAuthAccount(account) || account.IsShadow() || s.isOpenAIAccountRuntimeBlocked(account) {
		return false
	}
	disposition, _ := classifyOpenAIOAuth429(headers, responseBody)
	return disposition == openAIOAuth429Transient && s.openAIOAuth429RetryWindowActive(account)
}

// ShouldRetryOpenAIOAuth429 is consumed by RateLimitService to avoid turning
// the first transient OAuth 429 into a durable scheduler cooldown.
func (s *OpenAIGatewayService) ShouldRetryOpenAIOAuth429(account *Account, headers http.Header, responseBody []byte) bool {
	if s == nil || !isOpenAIOAuthAccount(account) || account.IsShadow() || s.isOpenAIAccountRuntimeBlocked(account) {
		return false
	}
	disposition, _ := classifyOpenAIOAuth429(headers, responseBody)
	return disposition == openAIOAuth429Transient && s.openAIOAuth429RetryWindowActive(account)
}

func (s *OpenAIGatewayService) applyOpenAIOAuth429Retry(account *Account, statusCode int, shouldDisable bool, headers http.Header, responseBody []byte, failoverErr *UpstreamFailoverError) *UpstreamFailoverError {
	if failoverErr == nil || !s.shouldRetryOpenAIOAuth429OnSameAccountWithResponse(account, statusCode, shouldDisable, headers, responseBody) {
		return failoverErr
	}
	failoverErr.RetryableOnSameAccount = true
	if failoverErr.SameAccountRetryLimit <= 0 {
		failoverErr.SameAccountRetryLimit = openAIOAuth429MaxAccountAttempts
	}
	if failoverErr.SameAccountRetryBackoffBase <= 0 {
		failoverErr.SameAccountRetryBackoffBase = openAIOAuth429RetryDelay
	}
	return failoverErr
}
