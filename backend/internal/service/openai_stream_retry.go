package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const OpenAIStreamRetryLimit = 3

// OpenAIStreamRetryError marks only replay-safe Responses streaming failures.
// Shared non-streaming and adapter failover errors deliberately remain untagged.
type OpenAIStreamRetryError struct {
	Failover *UpstreamFailoverError
}

func (e *OpenAIStreamRetryError) Error() string { return e.Failover.Error() }
func (e *OpenAIStreamRetryError) Unwrap() error { return e.Failover }

func OpenAIStreamRetryDelay(attempt int) time.Duration {
	return time.Duration(attempt) * 500 * time.Millisecond
}

func (s *OpenAIGatewayService) newResponsesStreamFailoverError(
	c *gin.Context, account *Account, passthrough bool, requestID string,
	payload []byte, message string, headers ...http.Header,
) error {
	if c != nil && c.Request != nil && c.Request.Context().Err() != nil {
		return fmt.Errorf("stream usage incomplete: %w", c.Request.Context().Err())
	}
	err := s.newOpenAIStreamFailoverError(c, account, passthrough, requestID, payload, message, headers...)
	// Preserve special capacity/overload, pool and OAuth429 policies first.
	if account == nil || !account.IsOpenAI() || account.IsPoolMode() || err.RetryableOnSameAccount ||
		openAIStreamFailedEventSemanticStatus(payload, message) != http.StatusBadGateway ||
		openAIStreamCredentialAuthFailure(payload) || isOpenAIUpstreamAccessStateError(message, payload) ||
		!openAIStreamFailedEventShouldFailover(payload, message) {
		return err
	}
	for _, path := range []string{"response.error.status_code", "error.status_code", "status_code"} {
		if status := gjson.GetBytes(payload, path).Int(); status >= 400 && status < 500 {
			return err
		}
	}
	code := openAIStreamFailedEventErrorCode(payload)
	if strings.HasPrefix(code, "invalid_") || strings.HasPrefix(code, "unsupported_") ||
		strings.HasPrefix(code, "billing_") {
		return err
	}
	switch code {
	case "insufficient_quota", "model_not_found", "permission_denied", "access_denied", "token_expired":
		return err
	}
	return &OpenAIStreamRetryError{Failover: err}
}

// OpenAIStreamRetryContext adds a temporary pin without changing the request
// context used for forwarding or replacing a caller's original account pin.
type openAIStreamRetryUserKey struct{}

func OpenAIStreamRetryContext(ctx context.Context, accountID, userID int64) context.Context {
	if _, pinned := pinnedAccountIDFromContext(ctx); pinned {
		return ctx
	}
	ctx = context.WithValue(ctx, openAIStreamRetryUserKey{}, userID)
	return context.WithValue(ctx, ctxkey.APIKeyPinnedAccountID, accountID)
}

const openAIStreamCommentBytesKey = "openai_stream_comment_bytes"

// OpenAIStreamWrittenSize excludes only comments written by this processor.
// Unknown writes, including an empty committed response, forbid replay.
func OpenAIStreamWrittenSize(c *gin.Context) int {
	size := OpenAICompactKeepaliveAdjustedWrittenSize(c)
	if c != nil && c.Writer != nil && c.Writer.Size() == 0 {
		return 0 // headers committed without a proven comment write
	}
	if size < 0 || c == nil {
		return size
	}
	n, _ := c.Get(openAIStreamCommentBytesKey)
	count, _ := n.(int)
	if count > 0 && size == count {
		return -1
	}
	return size - count
}

func OpenAIStreamCommentsCommitted(c *gin.Context) bool {
	if c == nil {
		return false
	}
	value, _ := c.Get(openAIStreamCommentBytesKey)
	count, _ := value.(int)
	return count > 0
}

func writeOpenAIStreamComment(c *gin.Context, comment string) error {
	replaySafe := OpenAIStreamWrittenSize(c) < 0
	n, err := c.Writer.WriteString(comment)
	// Never hide a write which followed unknown/semantic bytes.
	if err == nil && n == len(comment) && replaySafe {
		value, _ := c.Get(openAIStreamCommentBytesKey)
		count, _ := value.(int)
		c.Set(openAIStreamCommentBytesKey, count+n)
	}
	if err == nil {
		c.Writer.Flush()
	}
	return err
}

// Keep lifecycle/event lines out of bufio.Writer: its automatic 4KiB flush
// would otherwise commit the failed attempt before semantic output arrives.
type openAIStreamPreamble struct {
	strings.Builder
	limit int
}

func (p *openAIStreamPreamble) appendLine(line string) error {
	if len(line)+1 > p.limit-p.Len() {
		p.Reset()
		return errors.New("OpenAI stream preamble too large")
	}
	p.WriteString(line)
	p.WriteByte('\n')
	return nil
}
