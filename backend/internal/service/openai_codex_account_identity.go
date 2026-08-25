package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const codexAccountIdentitySourceContextKey = "openai_codex_account_identity_source"

func (s *OpenAIGatewayService) prepareCodexAccountIdentitySource(ctx context.Context, c *gin.Context, account *Account) (*Account, error) {
	source := account
	if account != nil && account.IsShadow() {
		resolved, err := resolveCredentialAccount(ctx, s.accountRepo, account)
		if err != nil {
			return nil, err
		}
		source = resolved
	}
	if c != nil {
		c.Set(codexAccountIdentitySourceContextKey, source)
	}
	return source, nil
}

func codexAccountIdentitySource(c *gin.Context, fallback *Account) *Account {
	if c != nil {
		if value, ok := c.Get(codexAccountIdentitySourceContextKey); ok {
			if source, ok := value.(*Account); ok && source != nil {
				return source
			}
		}
	}
	return fallback
}

func isOpenAIOAuthLikeAccount(account *Account) bool {
	return account != nil && account.IsOpenAI() && (account.Type == AccountTypeOAuth || account.Type == AccountTypeSetupToken)
}

// codexAccountIdentityNamespace is credential scoped and deliberately never uses a local row ID.
func codexAccountIdentityNamespace(account *Account) string {
	if !isOpenAIOAuthLikeAccount(account) {
		return ""
	}
	accountID := strings.TrimSpace(account.GetChatGPTAccountID())
	if accountID == "" {
		accountID = strings.TrimSpace(account.GetCredential("chatgpt_account_id"))
	}
	if accountID != "" {
		if userID := strings.TrimSpace(account.GetCredential("chatgpt_user_id")); userID != "" {
			return "chatgpt:" + accountID + ":user:" + userID
		}
		return "chatgpt:" + accountID
	}
	if account.Type == AccountTypeSetupToken {
		if token := strings.TrimSpace(account.GetOpenAIAccessToken()); token != "" {
			sum := sha256.Sum256([]byte("openai-setup-token:" + token))
			return fmt.Sprintf("setup-token:%x", sum[:16])
		}
	}
	if refreshToken := strings.TrimSpace(account.GetCredential("refresh_token")); refreshToken != "" {
		sum := sha256.Sum256([]byte("openai-oauth-credential:" + refreshToken))
		return fmt.Sprintf("oauth-credential:%x", sum[:16])
	}
	return ""
}

func isolateOpenAIUpstreamSessionID(apiKeyID int64, account *Account, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if namespace := codexAccountIdentityNamespace(account); namespace != "" {
		sum := sha256.Sum256([]byte(fmt.Sprintf("u%d:a%s:%s", apiKeyID, namespace, raw)))
		return fmt.Sprintf("%x", sum[:8])
	}
	return isolateOpenAISessionID(apiKeyID, raw)
}

func scopeCodexAccountIdentityValue(account *Account, apiKeyID int64, kind, raw string) string {
	raw = strings.TrimSpace(raw)
	namespace := codexAccountIdentityNamespace(account)
	if raw == "" || namespace == "" {
		return raw
	}
	return deriveStableUUIDv4(fmt.Sprintf("sub2api:codex-account-identity:v1:user:%d:account:%s:kind:%s:value:%s", apiKeyID, namespace, kind, raw))
}

var codexAccountIdentityFields = []struct{ name, kind string }{
	{"installation_id", "installation"}, {"x-codex-installation-id", "installation"}, {"session_id", "session"}, {"session-id", "session"}, {"thread_id", "thread"}, {"thread-id", "thread"}, {"turn_id", "turn"}, {"turn-id", "turn"}, {"window_id", "window"}, {"x-codex-window-id", "window"}, {"x-client-request-id", "request"},
}

func applyCodexAccountIdentityFields(values map[string]any, account *Account, apiKeyID int64) bool {
	if values == nil || codexAccountIdentityNamespace(account) == "" {
		return false
	}
	changed := false
	for _, field := range codexAccountIdentityFields {
		if raw, ok := values[field.name].(string); ok && strings.TrimSpace(raw) != "" {
			next := scopeCodexAccountIdentityValue(account, apiKeyID, field.kind, raw)
			if next != raw {
				values[field.name] = next
				changed = true
			}
		}
	}
	return changed
}

func applyCodexAccountIdentityEmbeddedMetadata(values map[string]any, account *Account, apiKeyID int64) bool {
	raw, ok := values[openAIWSTurnMetadataHeader].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return false
	}
	metadata := map[string]any{}
	if json.Unmarshal([]byte(raw), &metadata) != nil || !applyCodexAccountIdentityFields(metadata, account, apiKeyID) {
		return false
	}
	rebuilt, err := json.Marshal(metadata)
	if err != nil {
		return false
	}
	values[openAIWSTurnMetadataHeader] = string(rebuilt)
	return true
}

func applyCodexAccountIdentityClientMetadataMap(body map[string]any, account *Account, apiKeyID int64) bool {
	if body == nil || codexAccountIdentityNamespace(account) == "" {
		return false
	}
	changed, originalSession := false, ""
	if metadata, _ := body["client_metadata"].(map[string]any); metadata != nil {
		originalSession, _ = metadata["session_id"].(string)
		changed = applyCodexAccountIdentityFields(metadata, account, apiKeyID) || applyCodexAccountIdentityEmbeddedMetadata(metadata, account, apiKeyID)
	}
	if raw, ok := body["prompt_cache_key"].(string); ok && strings.TrimSpace(raw) != "" {
		kind := "prompt-cache"
		if originalSession != "" && raw == originalSession {
			kind = "session"
		}
		next := scopeCodexAccountIdentityValue(account, apiKeyID, kind, raw)
		if next != raw {
			body["prompt_cache_key"] = next
			changed = true
		}
	}
	return changed
}

// applyCodexAccountIdentityClientMetadataRaw decodes only client_metadata, never the complete passthrough body.
func applyCodexAccountIdentityClientMetadataRaw(body []byte, account *Account, apiKeyID int64) ([]byte, bool, error) {
	if len(body) == 0 || codexAccountIdentityNamespace(account) == "" || !gjson.ParseBytes(body).IsObject() {
		return body, false, nil
	}
	next, changed, originalSession := body, false, ""
	if cm := gjson.GetBytes(body, "client_metadata"); cm.IsObject() {
		metadata := map[string]any{}
		if err := json.Unmarshal([]byte(cm.Raw), &metadata); err != nil {
			return body, false, fmt.Errorf("decode client_metadata for account identity: %w", err)
		}
		originalSession, _ = metadata["session_id"].(string)
		if applyCodexAccountIdentityFields(metadata, account, apiKeyID) || applyCodexAccountIdentityEmbeddedMetadata(metadata, account, apiKeyID) {
			raw, err := json.Marshal(metadata)
			if err != nil {
				return body, false, fmt.Errorf("encode account-scoped client_metadata: %w", err)
			}
			var setErr error
			next, setErr = sjson.SetRawBytes(next, "client_metadata", raw)
			if setErr != nil {
				return body, false, fmt.Errorf("splice account-scoped client_metadata: %w", setErr)
			}
			changed = true
		}
	}
	if cache := gjson.GetBytes(body, "prompt_cache_key"); cache.Type == gjson.String && strings.TrimSpace(cache.String()) != "" {
		raw, kind := cache.String(), "prompt-cache"
		if originalSession != "" && raw == originalSession {
			kind = "session"
		}
		scoped := scopeCodexAccountIdentityValue(account, apiKeyID, kind, raw)
		if scoped != raw {
			rewritten, err := sjson.SetBytes(next, "prompt_cache_key", scoped)
			if err != nil {
				return body, false, fmt.Errorf("splice account-scoped prompt_cache_key: %w", err)
			}
			next, changed = rewritten, true
		}
	}
	return next, changed, nil
}

func applyCodexAccountIdentityHeaders(headers http.Header, account *Account, apiKeyID int64) {
	if headers == nil || codexAccountIdentityNamespace(account) == "" {
		return
	}
	for _, field := range codexAccountIdentityFields {
		if field.name != "session_id" {
			if raw := strings.TrimSpace(headers.Get(field.name)); raw != "" {
				headers.Set(field.name, scopeCodexAccountIdentityValue(account, apiKeyID, field.kind, raw))
			}
		}
	}
	if raw := strings.TrimSpace(headers.Get(openAIWSTurnMetadataHeader)); raw != "" {
		metadata := map[string]any{}
		if json.Unmarshal([]byte(raw), &metadata) == nil && applyCodexAccountIdentityFields(metadata, account, apiKeyID) {
			if rebuilt, err := json.Marshal(metadata); err == nil {
				headers.Set(openAIWSTurnMetadataHeader, string(rebuilt))
			}
		}
	}
}
