package service

import (
	"context"
	"errors"
	"io"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
)

// Gemini stays bound to the client context: it has no bounded detached drain.
// Only provider usage or completely written native image payloads are billable.
func geminiHasObservedTokens(usage *ClaudeUsage) bool {
	return usage != nil && (usage.InputTokens > 0 || usage.OutputTokens > 0 ||
		usage.CacheReadInputTokens > 0 || usage.CacheCreationInputTokens > 0 ||
		usage.CacheCreation5mTokens > 0 || usage.CacheCreation1hTokens > 0 ||
		usage.ImageOutputTokens > 0)
}

func geminiUncommittedCanceledAttempt(c *gin.Context) bool {
	if c == nil || c.Request == nil || !ctxkey.IsKeyRouteFirstResponseAttempt(c.Request.Context()) ||
		c.Request.Context().Err() == nil {
		return false
	}
	committed, ok := c.Writer.(interface{ Committed() bool })
	return !ok || !committed.Committed()
}

func geminiStreamReadError(c *gin.Context, err error) error {
	if errors.Is(err, io.EOF) && c != nil && c.Request != nil {
		if ctxErr := c.Request.Context().Err(); ctxErr != nil {
			return ctxErr
		}
	}
	return err
}

func geminiStreamAbortedWithBillableOutput(c *gin.Context, err error, usage *ClaudeUsage) bool {
	if err == nil || errors.Is(err, io.EOF) || geminiUncommittedCanceledAttempt(c) {
		return false
	}
	clientGone := errors.Is(err, context.Canceled) ||
		(c != nil && c.Request != nil && c.Request.Context().Err() != nil)
	return clientGone && (geminiHasObservedTokens(usage) || observedGeminiImageOutputs(c) > 0)
}

// writeSSE's legacy callers ignore write errors. Keep this attempt's first
// error and suppress every later write, including short writes without an error.
type geminiStreamWriter struct {
	writer io.Writer
	err    error
}

func (w *geminiStreamWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}
