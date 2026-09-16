package routes

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type keyRouteFirstResponseRequest struct {
	model  string
	stream bool
	body   []byte
}

// Tests replace this with a short duration; the public configuration remains
// whole seconds and production always uses time.Second.
var keyRouteFirstResponseSecond = time.Second

// withKeyRouteFirstResponse invokes an existing final dispatch closure at most
// once per eligible group. Middleware is never replayed and each closure owns
// its attempt synchronously, so a cancelled A cannot write while B runs.
func withKeyRouteFirstResponse(c *gin.Context, apiKeyService *service.APIKeyService, dispatch gin.HandlerFunc) {
	pristine, ok := middleware.APIKeyFirstResponsePristine(c)
	// Disabled keys must follow their exact legacy path: do not consume the
	// body, resolve a second group, or acquire/release a breaker probe.
	if !ok || pristine.LargestFirstResponseTimeoutSeconds() <= 0 {
		dispatch(c)
		return
	}
	request, ok := keyRouteFirstResponseEligible(c)
	if !ok {
		dispatch(c)
		return
	}
	// Studio bridge has no API-key auth snapshot; keeping its legacy path avoids
	// claiming unsupported default-group semantics.
	if pristine.PinnedAccountID > 0 {
		dispatch(c)
		return
	}

	probe := c.Copy()
	probe.Request = cloneKeyRouteFirstResponseRequest(c.Request, request.body, c.Request.Context())
	totalGate := &keyRouteFirstResponseGate{}
	probe.Writer = newKeyRouteFirstResponseWriter(c.Writer, request.stream, totalGate)
	resolved, prepared := middleware.PrepareAPIKeyFirstResponseAttempt(probe, apiKeyService, pristine, request.model, nil, probe.Request.Context())
	if !prepared || resolved == nil || resolved.GroupID == nil {
		// Local validation/billing response was produced on the private copy.
		writer := probe.Writer.(*keyRouteFirstResponseWriter)
		if writer.Written() || writer.Status() >= http.StatusBadRequest {
			if err := writer.Commit(); err != nil && !writer.Committed() {
				writeKeyRouteFirstResponseUnavailable(c)
			}
		} else {
			writeKeyRouteFirstResponseUnavailable(c)
		}
		return
	}
	timeoutSeconds := resolved.FirstResponseTimeoutSeconds(*resolved.GroupID)
	if timeoutSeconds <= 0 {
		// Reuse the exact resolver decision (and its possible half-open lease) for
		// legacy handling. Re-selecting would perturb weighted routing or acquire
		// a second probe despite the feature being disabled.
		copyKeyRouteFirstResponseFinalContext(c, probe)
		dispatch(c)
		return
	}

	// Do not let the auth middleware's post-Next epilogue combine A's failure
	// with B's success. Finalization below is per immutable selected key.
	c.Set(string(middleware.ContextKeyAPIKeyFirstResponseDispatcher), true)
	totalSeconds := pristine.LargestFirstResponseTimeoutSeconds()
	if totalSeconds < timeoutSeconds {
		totalSeconds = timeoutSeconds
	}
	total := time.Duration(totalSeconds) * keyRouteFirstResponseSecond
	if total < 120*keyRouteFirstResponseSecond {
		total = 120 * keyRouteFirstResponseSecond
	}
	if total > 600*keyRouteFirstResponseSecond {
		total = 600 * keyRouteFirstResponseSecond
	}
	totalCtx, cancelTotal := context.WithCancel(probe.Request.Context())
	totalTimer := time.AfterFunc(total, func() {
		if totalGate.expire() {
			cancelTotal()
		}
	})
	defer func() { totalTimer.Stop(); cancelTotal() }()
	attempted := make(map[int64]struct{})
	firstAttempt := probe
	firstResolved := resolved
	var lastAttempt *gin.Context

	for {
		attempt := firstAttempt
		var writer *keyRouteFirstResponseWriter
		if attempt != nil {
			resolved = firstResolved
			writer = attempt.Writer.(*keyRouteFirstResponseWriter)
			attempt.Request = attempt.Request.WithContext(ctxkey.WithKeyRouteFirstResponseAttempt(totalCtx))
			firstAttempt = nil
		} else {
			attempt = c.Copy()
			attempt.Request = cloneKeyRouteFirstResponseRequest(c.Request, request.body, totalCtx)
			writer = newKeyRouteFirstResponseWriter(c.Writer, request.stream, totalGate)
			attempt.Writer = writer
			resolved, prepared = middleware.PrepareAPIKeyFirstResponseAttempt(attempt, apiKeyService, pristine, request.model, attempted, attempt.Request.Context())
			if !prepared || resolved == nil || resolved.GroupID == nil {
				if writer.Written() || writer.Status() >= http.StatusBadRequest {
					copyKeyRouteFirstResponseFinalContext(c, attempt)
					if err := writer.Commit(); err != nil && !writer.Committed() {
						writeKeyRouteFirstResponseUnavailable(c)
					}
				} else {
					copyKeyRouteFirstResponseFinalContext(c, lastAttempt)
					writeKeyRouteFirstResponseUnavailable(c)
				}
				return
			}
		}
		groupID := *resolved.GroupID
		attemptStartedAt := time.Now()
		attempted[groupID] = struct{}{}

		if totalGate.isExpired() || totalCtx.Err() != nil {
			// This selected group has not been attempted: release its probe,
			// rather than cooling a healthy group for an exhausted total budget.
			middleware.FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, middleware.APIKeyFirstResponseOutcome{})
			if c.Request.Context().Err() != nil {
				return
			}
			copyKeyRouteFirstResponseFinalContext(c, lastAttempt)
			writeKeyRouteFirstResponseUnavailable(c)
			return
		}
		perGroup := time.Duration(timeoutSeconds) * keyRouteFirstResponseSecond
		if configured := resolved.FirstResponseTimeoutSeconds(groupID); configured > 0 {
			perGroup = time.Duration(configured) * keyRouteFirstResponseSecond
		}
		attemptCtx, cancelAttempt := context.WithCancel(attempt.Request.Context())
		attemptTimer := time.AfterFunc(perGroup, func() {
			if writer.Expire() {
				cancelAttempt()
			}
		})
		writer.SetFirstBusiness(func() { attemptTimer.Stop(); totalTimer.Stop() })
		attempt.Request = attempt.Request.WithContext(ctxkey.WithKeyRouteFirstResponseAttempt(attemptCtx))
		dispatch(attempt)
		lastAttempt = attempt
		attemptTimer.Stop()
		cancelAttempt() // handler returned; this cannot interrupt a healthy live stream.

		// A non-streaming result becomes ready only after the handler returns.
		// Commit participates in the same locked gates as the timers. If a timer
		// wins between this check and Commit, the failure path below observes the
		// gate itself, not a separately published timeout flag.
		if !writer.Committed() && !request.stream && writer.Status() < http.StatusBadRequest && writer.Size() > 0 &&
			c.Request.Context().Err() == nil && !keyRouteFirstResponseTransientStatus(attempt) &&
			!writer.isExpired() && !totalGate.isExpired() {
			_ = writer.Commit()
		}
		if writer.Committed() {
			outcome := middleware.APIKeyFirstResponseOutcomeForContext(attempt)
			middleware.FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, outcome)
			logKeyRouteFirstResponse(groupID, time.Since(attemptStartedAt), "committed", "terminal", outcome.Failure)
			copyKeyRouteFirstResponseFinalContext(c, attempt)
			return
		}
		if c.Request.Context().Err() != nil {
			middleware.FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, middleware.APIKeyFirstResponseOutcome{})
			return
		}
		if writer.isExpired() || totalGate.isExpired() || keyRouteFirstResponseTransientStatus(attempt) {
			middleware.FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, middleware.APIKeyFirstResponseOutcome{Failure: true})
			if !totalGate.isExpired() && totalCtx.Err() == nil {
				logKeyRouteFirstResponse(groupID, time.Since(attemptStartedAt), "retry", "timeout_or_transient", true)
				continue
			}
			logKeyRouteFirstResponse(groupID, time.Since(attemptStartedAt), "exhausted", "deadline", true)
			copyKeyRouteFirstResponseFinalContext(c, attempt)
			writeKeyRouteFirstResponseUnavailable(c)
			return
		}
		// Local 4xx and other terminal protocol outcomes are never replayed.
		middleware.FinalizeAPIKeyFirstResponseAttempt(apiKeyService, resolved, middleware.APIKeyFirstResponseOutcome{})
		logKeyRouteFirstResponse(groupID, time.Since(attemptStartedAt), "terminal", "local_or_business", false)
		if err := writer.Commit(); err != nil && !writer.Committed() {
			writeKeyRouteFirstResponseUnavailable(c)
		}
		copyKeyRouteFirstResponseFinalContext(c, attempt)
		return
	}
}

func keyRouteFirstResponseEligible(c *gin.Context) (keyRouteFirstResponseRequest, bool) {
	if c == nil || c.Request == nil || c.Request.Method != http.MethodPost || !keyRouteFirstResponsePath(c.Request.URL.Path) {
		return keyRouteFirstResponseRequest{}, false
	}
	if middleware.HasForcePlatform(c) || strings.TrimSpace(c.GetHeader("Upgrade")) != "" || strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") {
		return keyRouteFirstResponseRequest{}, false
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(c.GetHeader("Content-Type"), ";")[0]))
	if mediaType != "application/json" {
		return keyRouteFirstResponseRequest{}, false
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return keyRouteFirstResponseRequest{}, false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if !gjson.ValidBytes(body) || !gjson.GetBytes(body, "@this").IsObject() {
		return keyRouteFirstResponseRequest{}, false
	}
	root := gjson.ParseBytes(body)
	modelValue := root.Get("model")
	if modelValue.Type != gjson.String {
		return keyRouteFirstResponseRequest{}, false
	}
	model := strings.TrimSpace(modelValue.String())
	if model == "" {
		return keyRouteFirstResponseRequest{}, false
	}
	for _, field := range []string{"previous_response_id", "conversation"} {
		if value := root.Get(field); value.Exists() && value.Type != gjson.Null {
			return keyRouteFirstResponseRequest{}, false
		}
	}
	stream := root.Get("stream")
	if stream.Exists() && stream.Type != gjson.True && stream.Type != gjson.False {
		return keyRouteFirstResponseRequest{}, false
	}
	if service.IsImageGenerationIntent(c.Request.URL.Path, model, body) || service.IsVideoGenerationIntent(c.Request.URL.Path, model, body) || root.Get("tools.#(type==\"image_generation\")").Exists() {
		return keyRouteFirstResponseRequest{}, false
	}
	return keyRouteFirstResponseRequest{model: model, stream: stream.Bool(), body: body}, true
}

func keyRouteFirstResponsePath(path string) bool {
	switch path {
	case "/v1/messages", "/v1/chat/completions", "/v1/responses", "/chat/completions", "/responses", "/backend-api/codex/responses":
		return true
	}
	return false
}
func cloneKeyRouteFirstResponseRequest(request *http.Request, body []byte, ctx context.Context) *http.Request {
	clone := request.Clone(ctx)
	clone.Body = io.NopCloser(bytes.NewReader(body))
	clone.ContentLength = int64(len(body))
	return clone
}
func keyRouteFirstResponseTransientStatus(c *gin.Context) bool {
	status := c.Writer.Status()
	if upstream, ok := c.Get(service.OpsUpstreamStatusCodeKey); ok {
		code, _ := upstream.(int)
		if status >= 400 && status < 500 {
			return status == http.StatusTooManyRequests && code == http.StatusTooManyRequests
		}
		return code == http.StatusTooManyRequests || code == 529 || code >= 500
	}
	return status == 529 || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}
func logKeyRouteFirstResponse(groupID int64, elapsed time.Duration, result, reason string, failed bool) {
	slog.Info("api key first response route", "group_id", groupID, "elapsed_ms", elapsed.Milliseconds(), "reason", reason, "switch_result", result, "failed", failed)
}
func copyKeyRouteFirstResponseFinalContext(target, source *gin.Context) {
	if target == nil || source == nil {
		return
	}
	for key, value := range source.Keys {
		target.Set(key, value)
	}
	if apiKey, ok := middleware.GetAPIKeyFromContext(source); ok {
		middleware.SetAPIKeyContext(target, apiKey)
	}
}
func writeKeyRouteFirstResponseUnavailable(c *gin.Context) {
	c.JSON(http.StatusGatewayTimeout, gin.H{"error": gin.H{"type": "upstream_unavailable", "message": "No route produced a first response before the timeout"}})
}
