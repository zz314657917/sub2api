package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
)

// StreamBillingFirstResponseUncommitted distinguishes buffered preludes from
// bytes actually committed by the first-response attempt writer.
func StreamBillingFirstResponseUncommitted(ctx context.Context, c *gin.Context) bool {
	if ctx == nil || ctx.Err() == nil || !ctxkey.IsKeyRouteFirstResponseAttempt(ctx) {
		return false
	}
	if c == nil || c.Writer == nil {
		return true
	}
	if writer, ok := c.Writer.(interface{ Committed() bool }); ok {
		return !writer.Committed()
	}
	// Without an explicit commit signal, buffered prelude bytes cannot prove delivery.
	return true
}

func (u *ClaudeUsage) hasObservedTokens() bool {
	return u != nil && (u.InputTokens > 0 || u.OutputTokens > 0 ||
		u.CacheCreationInputTokens > 0 || u.CacheReadInputTokens > 0 ||
		u.CacheCreation5mTokens > 0 || u.CacheCreation1hTokens > 0 || u.ImageOutputTokens > 0)
}

// A retryable attempt must never also expose a billable partial result.
func partialStreamUsageResult(ctx context.Context, c *gin.Context, resp *http.Response, stream *streamingResult, model, upstreamModel string, start time.Time, err error) *ForwardResult {
	var failover *UpstreamFailoverError
	if StreamBillingFirstResponseUncommitted(ctx, c) || errors.As(err, &failover) || stream == nil || !stream.usage.hasObservedTokens() {
		return nil
	}
	return &ForwardResult{
		RequestID:                     resp.Header.Get("x-request-id"),
		Usage:                         *stream.usage,
		Model:                         model,
		UpstreamModel:                 upstreamModel,
		UpstreamResponseModel:         observedUpstreamResponseModel(c),
		UpstreamResponseModelConflict: observedUpstreamResponseModelConflict(c),
		Stream:                        true,
		Duration:                      time.Since(start),
		FirstTokenMs:                  stream.firstTokenMs,
		ClientDisconnect:              stream.clientDisconnect,
	}
}
