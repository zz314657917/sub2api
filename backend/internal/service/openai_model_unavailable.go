package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const openAIModelUnavailableRequestKey = "openai_model_unavailable_request"

type openAIModelUnavailableRequest struct {
	accountID int64
	model     string
}

func rememberOpenAIModelUnavailableRequest(c *gin.Context, account *Account, body []byte) {
	if c != nil && account != nil {
		c.Set(openAIModelUnavailableRequestKey, openAIModelUnavailableRequest{
			accountID: account.ID, model: gjson.GetBytes(body, "model").String(),
		})
	}
}

// Keep this opt-in at HTTP text owners. Images have a separate driver/model
// rejection guard; WS and other platforms retain their existing policies.
func openAIModelUnavailableHTTPScope(c *gin.Context, account *Account) bool {
	if c == nil || c.Request == nil || account == nil || account.Platform != PlatformOpenAI ||
		account.IsAnthropicProtocol() {
		return false
	}
	switch account.Type {
	case AccountTypeAPIKey, AccountTypeOAuth, AccountTypeSetupToken:
	default:
		return false
	}
	if GetOpenAIClientTransport(c) == OpenAIClientTransportWS ||
		strings.EqualFold(c.Request.Header.Get("Upgrade"), "websocket") {
		return false
	}
	if raw, ok := c.Get(openAIClientTransportContextKey); ok && raw != nil &&
		strings.TrimSpace(fmt.Sprint(raw)) != "" && GetOpenAIClientTransport(c) != OpenAIClientTransportHTTP {
		return false
	}
	if strings.Contains(c.Request.URL.Path, "/images/") {
		return false
	}
	switch GetActualOpenAIUpstreamEndpoint(c) {
	case "", "/v1/responses", "/v1/chat/completions":
		return true
	default:
		return false
	}
}

// Only error fields are authoritative. Never search serialized request/debug
// metadata, nor recursively decode an echoed JSON message over an outer code.
func openAIModelUnavailableError(account *Account, status int, body []byte) bool {
	if status != http.StatusBadRequest && status != http.StatusNotFound {
		return false
	}
	var messages []string
	explicit, conflict := false, false
	if gjson.ValidBytes(body) {
		root := gjson.ParseBytes(body)
		if !root.IsObject() {
			return false
		}
		for _, prefix := range []string{"", "error.", "detail."} {
			code := root.Get(prefix + "code")
			if code.Exists() && code.Type != gjson.Null && strings.TrimSpace(code.String()) != "" {
				switch strings.ToLower(strings.TrimSpace(code.String())) {
				case "model_not_found", "model_not_available", "model_unavailable", "unsupported_model":
					explicit = true
				default:
					conflict = true
				}
			}
			for _, key := range []string{"message", "detail"} {
				v := root.Get(prefix + key)
				if v.Type == gjson.String {
					messages = append(messages, v.Str)
				}
			}
		}
		if root.Get("error").Type == gjson.String {
			messages = append(messages, root.Get("error").Str)
		}
	} else {
		text := strings.TrimSpace(string(body))
		// Malformed JSON/HTML must not become an authoritative plain sentence.
		if text == "" || strings.ContainsAny(text, "{}[]<>") {
			return false
		}
		messages = []string{text}
	}
	if conflict {
		return false
	}
	positive := false
	for _, message := range messages {
		m := strings.ToLower(strings.Join(strings.Fields(message), " "))
		if strings.HasPrefix(m, "{") || strings.HasPrefix(m, "[") {
			continue
		}
		// Parameter/capability errors describe the request, not model availability.
		if strings.Contains(m, "response_format") || strings.Contains(m, "context length") ||
			strings.Contains(m, "context window") || strings.Contains(m, "unsupported parameter") {
			return false
		}
		if strings.Contains(m, "model is not supported when using codex") {
			// Do this before generic "not supported" matching, including API keys.
			if account == nil || !account.IsOpenAIOAuth() || status != http.StatusBadRequest {
				return false
			}
			positive = true
			continue
		}
		if strings.Contains(m, "model not found") || m == "unknown model" ||
			strings.Contains(m, "unknown model:") || strings.HasPrefix(m, "unknown model ") ||
			m == "unknown provider for model" || strings.HasPrefix(m, "unknown provider for model:") ||
			strings.HasPrefix(m, "unknown provider for model ") {
			positive = true
		}
		// Require the sentence's subject to be a model; "route not found for
		// model X" and "model does not support parameter Y" do not qualify.
		if strings.HasPrefix(m, "model ") || strings.HasPrefix(m, "the model ") ||
			strings.HasPrefix(m, "the '") && strings.Contains(m, "' model ") {
			for _, phrase := range []string{" does not exist", " was not found", " is not found",
				" is not available", " is unavailable", " is not supported by "} {
				if strings.Contains(m, phrase) {
					positive = true
				}
			}
			if strings.HasSuffix(m, "model is not supported") {
				positive = true
			}
			// "For streaming/tools/schema" describes a capability mismatch.
			// Only account/provider/plan scopes establish model availability;
			// mentioning an account later in a capability error is insufficient.
			if _, scope, ok := strings.Cut(m, " is not supported for "); ok {
				for _, qualifier := range []string{"this ", "your ", "the ", "current "} {
					scope = strings.TrimPrefix(scope, qualifier)
				}
				fields := strings.Fields(scope)
				if len(fields) > 0 {
					switch strings.TrimRight(fields[0], ".,:;") {
					case "account", "provider", "plan", "subscription":
						positive = true
					}
				}
			}
		}
	}
	return explicit || positive
}

func openAIHTTPModelUnavailable(c *gin.Context, account *Account, status int, body []byte) bool {
	return openAIModelUnavailableHTTPScope(c, account) && openAIModelUnavailableError(account, status, body)
}

// upstreamModel is the value already sent, never an input to GetMappedModel.
// The scheduler currently looks up the mapped request slug before OAuth
// normalization/compact rewriting. Mirror that key as well when it differs,
// without mapping the effective upstream slug a second time.
func (s *OpenAIGatewayService) persistOpenAIModelUnavailable(ctx context.Context, c *gin.Context, account *Account, status int, upstreamModel string) {
	if s == nil || s.rateLimitService == nil || s.rateLimitService.accountRepo == nil ||
		!account.ShouldHandleErrorCode(status) || account.IsPoolMode() && !account.IsCustomErrorCodesEnabled() {
		return
	}
	keys := []string{normalizeModelRateLimitLookupKey(upstreamModel)}
	// Rejecting an image model on a text endpoint is not evidence that the
	// account cannot generate images. The image HTTP owner owns that decision.
	if isOpenAIImageModel(upstreamModel) {
		return
	}
	if raw, ok := c.Get(openAIModelUnavailableRequestKey); ok {
		if request, ok := raw.(openAIModelUnavailableRequest); ok && request.accountID == account.ID {
			for _, key := range account.modelRateLimitKeysForRequest(ctx, request.model) {
				if key != keys[0] && !isOpenAIImageModel(key) {
					keys = append(keys, key)
				}
			}
		}
	}
	resetAt := time.Now().Add(upstreamModelNotFoundCooldown)
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if err := s.rateLimitService.accountRepo.SetModelRateLimit(ctx, account.ID, key, resetAt); err != nil {
			slog.Warn("openai_model_unavailable_cooldown_failed", "account_id", account.ID, "model", key, "error", err)
		}
	}
}

func (s *OpenAIGatewayService) failoverOpenAIModelUnavailable(ctx context.Context, c *gin.Context, account *Account, resp *http.Response, body []byte, upstreamModel string) error {
	if resp == nil || !openAIHTTPModelUnavailable(c, account, resp.StatusCode, body) {
		return nil
	}
	// Existing explicit error presentation rules still own the response.
	if getBoundErrorPassthroughService(c) != nil {
		if rule := getBoundErrorPassthroughService(c).MatchRule(PlatformOpenAI, resp.StatusCode, body); rule != nil {
			return nil
		}
	}
	s.persistOpenAIModelUnavailable(ctx, c, account, resp.StatusCode, upstreamModel)
	committed := c.Writer.Written()
	if writer, ok := c.Writer.(interface{ Committed() bool }); ok {
		committed = writer.Committed()
	}
	if IsResponseCommitted(c) || committed {
		return errors.New("upstream model unavailable after client output")
	}
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
		Platform: account.Platform, AccountID: account.ID, AccountName: account.Name,
		UpstreamStatusCode: resp.StatusCode, UpstreamRequestID: resp.Header.Get("x-request-id"),
		Kind: "failover", Message: sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(body)),
	})
	// Pool retry settings and transient-capacity retry metadata must never turn
	// this deterministic rejection into another call to the same account.
	return &UpstreamFailoverError{
		StatusCode: resp.StatusCode, ResponseBody: body, ResponseHeaders: resp.Header.Clone(),
	}
}
