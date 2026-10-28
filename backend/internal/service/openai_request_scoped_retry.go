package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const openAIRequestRetryMaxDelay = 8 * time.Second

// Only explicit empty strings in known semantic events are safe to hold back.
// Tool announcements/completions and unrecognised structures remain output.
func openAIStreamKnownEmptyEvent(data, eventType string) bool {
	if !gjson.Valid(data) || !gjson.Parse(data).IsObject() {
		return false
	}
	if gjson.Get(data, "type").Type != gjson.String || gjson.Get(data, "type").Str != eventType ||
		gjson.Get(data, "encrypted_content").Exists() {
		return false
	}
	empty := func(path string) bool {
		v := gjson.Get(data, path)
		return v.Type == gjson.String && v.Str == ""
	}
	switch eventType {
	case "response.output_text.delta", "response.reasoning_text.delta",
		"response.reasoning_summary_text.delta", "response.refusal.delta",
		"response.function_call_arguments.delta", "response.custom_tool_call_input.delta":
		return empty("delta")
	case "response.output_text.done", "response.reasoning_text.done", "response.reasoning_summary_text.done":
		return empty("text")
	case "response.refusal.done":
		return empty("refusal")
	case "response.content_part.added", "response.content_part.done",
		"response.reasoning_summary_part.added", "response.reasoning_summary_part.done":
		part := gjson.Get(data, "part")
		if !part.IsObject() {
			return false
		}
		kind := part.Get("type")
		if kind.Type != gjson.String {
			return false
		}
		summary := strings.HasPrefix(eventType, "response.reasoning_summary_part.")
		if summary && kind.Str != "summary_text" || !summary && kind.Str != "output_text" && kind.Str != "refusal" {
			return false
		}
		switch kind.Str {
		case "output_text", "summary_text":
			if !empty("part.text") {
				return false
			}
		case "refusal":
			if !empty("part.refusal") {
				return false
			}
		default:
			return false
		}
		safe := true
		part.ForEach(func(key, value gjson.Result) bool {
			switch key.Str {
			case "type", "text", "refusal":
				// An optional alternate string must also be explicitly empty.
				if key.Str != "type" && (value.Type != gjson.String || value.Str != "") {
					safe = false
				}
			case "annotations":
				safe = safe && value.IsArray() && len(value.Array()) == 0
			default:
				safe = false
			}
			return safe
		})
		return safe
	}
	return false
}

func openAIRequestOverloadHTTPStatus(status int) bool {
	switch status {
	case 400, 500, 502, 503, 529:
		return true
	}
	return false
}

// The same predicate guards metadata and HTTP account-side-effect suppression.
// Raw stream envelopes use 200; authoritative semantic statuses still win.
func openAIRequestScopedOverload(c *gin.Context, account *Account, status int, body []byte, stream bool) bool {
	if c == nil || c.Request == nil || account == nil || account.Platform != PlatformOpenAI ||
		account.IsAnthropicProtocol() || GetActualOpenAIUpstreamEndpoint(c) == "/v1/messages" {
		return false
	}
	switch account.Type {
	case AccountTypeAPIKey, AccountTypeOAuth, AccountTypeSetupToken:
	default:
		return false
	}
	if GetOpenAIClientTransport(c) == OpenAIClientTransportWS ||
		strings.EqualFold(strings.TrimSpace(c.Request.Header.Get("Upgrade")), "websocket") {
		return false
	}
	if raw, ok := c.Get(openAIClientTransportContextKey); ok && raw != nil &&
		strings.TrimSpace(fmt.Sprint(raw)) != "" && GetOpenAIClientTransport(c) != OpenAIClientTransportHTTP {
		return false
	}
	switch GetActualOpenAIUpstreamEndpoint(c) {
	case "", "/v1/responses", "/v1/chat/completions", "/v1/images/generations", "/v1/images/edits":
	default:
		return false
	}
	if !(stream && status == http.StatusOK) && !openAIRequestOverloadHTTPStatus(status) {
		return false
	}
	message := strings.TrimSpace(extractUpstreamErrorMessage(body))
	if !gjson.ValidBytes(body) {
		// Do not treat malformed JSON as a plain authoritative error sentence.
		return isOpenAIServerOverloadedMessage(string(body))
	}
	if !gjson.ParseBytes(body).IsObject() {
		return false
	}
	if isOpenAIModelCapacityError(message, body) || isOpenAIContextWindowError(message, body) ||
		isOpenAIUpstreamAccessStateError(message, body) {
		return false
	}
	if hit, _, _ := detectOpenAICyberPolicy(body); hit {
		return false
	}
	positive, conflict := false, false
	for _, prefix := range []string{"", "error.", "response.", "response.error."} {
		for _, key := range []string{"status", "status_code", "http_status"} {
			v := gjson.GetBytes(body, prefix+key)
			if v.Type == gjson.Number || v.Type == gjson.String {
				n, err := strconv.Atoi(v.String())
				if err == nil && n >= 400 && !openAIRequestOverloadHTTPStatus(n) {
					conflict = true
				}
			}
		}
		for _, key := range []string{"code", "type", "message"} {
			v := gjson.GetBytes(body, prefix+key)
			if v.Type != gjson.String {
				continue
			}
			value := strings.ToLower(strings.TrimSpace(v.Str))
			if key == "code" && isOpenAIServerOverloadedCode(value) ||
				key == "message" && isOpenAIServerOverloadedMessage(value) {
				positive = true
			}
			if key == "type" && (value == "error" || value == "response.failed") {
				continue
			}
			for _, marker := range []string{
				"auth", "credential", "api_key", "token_expired", "token_revoked", "token_invalid",
				"access_token", "access_denied", "access denied", "deactivat", "disabled", "suspend",
				"quota", "rate_limit", "rate limit", "too many requests", "billing", "payment",
				"permission", "forbidden", "invalid", "unsupported", "model_not_found", "not found",
				"context_length", "context_window", "context window", "maximum context",
				"content", "policy", "safety", "cyber", "capacity",
			} {
				if strings.Contains(value, marker) {
					conflict = true
				}
			}
		}
	}
	return positive && !conflict
}

// Parse before converting to time.Duration, so huge finite values cannot wrap.
func openAIRequestRetryValue(value string, unit time.Duration, now time.Time, date bool) time.Duration {
	value = strings.TrimSpace(value)
	if n, err := strconv.ParseFloat(value, 64); err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n > 0 {
		if n >= float64(openAIRequestRetryMaxDelay)/float64(unit) {
			return openAIRequestRetryMaxDelay
		}
		return time.Duration(n * float64(unit))
	}
	if date {
		if at, err := http.ParseTime(value); err == nil {
			delay := at.Sub(now)
			if delay > openAIRequestRetryMaxDelay {
				return openAIRequestRetryMaxDelay
			}
			if delay > 0 {
				return delay
			}
		}
	}
	return 0
}

func openAIRequestRetryDelay(headers http.Header, body []byte, now time.Time) time.Duration {
	for _, key := range []string{"Retry-After", "Retry-After-Ms"} {
		values := headers.Values(key)
		if len(values) == 0 {
			for name, raw := range headers {
				if strings.EqualFold(name, key) {
					values = raw
					break
				}
			}
		}
		for _, value := range values {
			unit := time.Second
			if key == "Retry-After-Ms" {
				unit = time.Millisecond
			}
			if delay := openAIRequestRetryValue(value, unit, now, key == "Retry-After"); delay > 0 {
				return delay
			}
		}
	}
	// JSON header maps are case insensitive, like real HTTP headers.
	for _, prefix := range []string{"error.headers", "response.error.headers"} {
		h := make(http.Header)
		gjson.GetBytes(body, prefix).ForEach(func(key, value gjson.Result) bool {
			if value.IsArray() {
				for _, item := range value.Array() {
					if item.Type == gjson.String || item.Type == gjson.Number {
						h.Add(key.Str, item.String())
					}
				}
			} else if value.Type == gjson.String || value.Type == gjson.Number {
				h.Add(key.Str, value.String())
			}
			return true
		})
		for _, key := range []string{"Retry-After", "Retry-After-Ms"} {
			unit := time.Second
			if key == "Retry-After-Ms" {
				unit = time.Millisecond
			}
			for _, value := range h.Values(key) {
				if delay := openAIRequestRetryValue(value, unit, now, key == "Retry-After"); delay > 0 {
					return delay
				}
			}
		}
	}
	return 0
}

func augmentOpenAIHTTPFailover(c *gin.Context, account *Account, resp *http.Response, body []byte, err *UpstreamFailoverError) *UpstreamFailoverError {
	if resp == nil || err == nil {
		return err
	}
	err.ResponseHeaders = resp.Header.Clone()
	if openAIRequestScopedOverload(c, account, resp.StatusCode, body, false) {
		err.RequestScopedTransient = true
		err.SameAccountRetryDelay = openAIRequestRetryDelay(resp.Header, body, time.Now())
	}
	return err
}

// Explicit opt-in for HTTP owners. Generic callers (notably WS bridges) keep
// handleOpenAIAccountUpstreamError's existing account health behaviour.
func (s *OpenAIGatewayService) handleOpenAIHTTPAccountUpstreamError(ctx context.Context, c *gin.Context, account *Account, status int, headers http.Header, body []byte, model ...string) bool {
	if openAIRequestScopedOverload(c, account, status, body, false) {
		return false
	}
	return s.handleOpenAIAccountUpstreamError(ctx, account, status, headers, body, model...)
}
