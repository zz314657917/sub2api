package service

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// openAICodexTurnStateHeader is intentionally handled outside the generic
// response-header allowlist. Codex sends it back on the next request.
const openAICodexTurnStateHeader = "x-codex-turn-state"

type openAICodexTurnStateOrigin struct {
	accountID int64
	expiresAt time.Time
}

// openAICodexTurnStateSeed uses the original inbound session, before OAuth
// session isolation. A missing authenticated API key intentionally disables
// tracking so anonymous and internal callers cannot share provenance.
func openAICodexTurnStateSeed(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	apiKeyID := getAPIKeyIDFromContext(c)
	if apiKeyID <= 0 {
		return ""
	}
	sessionID := strings.TrimSpace(c.Request.Header.Get("session-id"))
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.Request.Header.Get("session_id"))
	}
	if sessionID == "" {
		return ""
	}
	return strconv.FormatInt(apiKeyID, 10) + "\x00" + sessionID
}

func extractOpenAICodexTurnState(upstream http.Header) string {
	if upstream == nil {
		return ""
	}
	return strings.TrimSpace(upstream.Get(openAICodexTurnStateHeader))
}

// stageOpenAICodexTurnState writes (or clears) only the response header. It
// deliberately does not record provenance: an HTTP streaming attempt may still
// fail over before the downstream response is committed.
func stageOpenAICodexTurnState(dst http.Header, upstream http.Header) string {
	if dst == nil {
		return ""
	}
	key := http.CanonicalHeaderKey(openAICodexTurnStateHeader)
	state := extractOpenAICodexTurnState(upstream)
	if state == "" {
		dst.Del(key)
		return ""
	}
	dst.Set(key, state)
	return state
}

func (s *OpenAIGatewayService) noteOpenAICodexTurnStateCommitted(c *gin.Context, account *Account, state string) {
	if state == "" || c == nil || c.Writer == nil || !c.Writer.Written() {
		return
	}
	s.noteOpenAICodexTurnStateProvenance(c, account)
}

func (s *OpenAIGatewayService) noteOpenAICodexTurnStateProvenance(c *gin.Context, account *Account) {
	if s == nil || account == nil || account.ID <= 0 {
		return
	}
	seed := openAICodexTurnStateSeed(c)
	if seed == "" {
		return
	}
	s.openaiCodexTurnStateOrigins.Store(seed, openAICodexTurnStateOrigin{
		accountID: account.ID,
		expiresAt: time.Now().Add(s.openAIWSSessionStickyTTL()),
	})
	s.sweepOpenAICodexTurnStateOrigins()
	if s.openaiCodexTurnStateNoteHook != nil {
		s.openaiCodexTurnStateNoteHook(c)
	}
}

func (s *OpenAIGatewayService) guardOpenAICodexTurnStateEcho(c *gin.Context, account *Account, h http.Header) {
	if s == nil || account == nil || account.ID <= 0 || h == nil || strings.TrimSpace(h.Get(openAICodexTurnStateHeader)) == "" {
		return
	}
	seed := openAICodexTurnStateSeed(c)
	if seed == "" {
		return
	}
	raw, ok := s.openaiCodexTurnStateOrigins.Load(seed)
	if !ok {
		return
	}
	origin, ok := raw.(openAICodexTurnStateOrigin)
	if !ok || (!origin.expiresAt.IsZero() && time.Now().After(origin.expiresAt)) {
		s.openaiCodexTurnStateOrigins.Delete(seed)
		return
	}
	if origin.accountID != account.ID {
		h.Del(openAICodexTurnStateHeader)
	}
}

func (s *OpenAIGatewayService) sweepOpenAICodexTurnStateOrigins() {
	if s == nil || s.openaiCodexTurnStateWrites.Add(1)%256 != 0 {
		return
	}
	now := time.Now()
	s.openaiCodexTurnStateOrigins.Range(func(key, value any) bool {
		origin, ok := value.(openAICodexTurnStateOrigin)
		if !ok || (!origin.expiresAt.IsZero() && now.After(origin.expiresAt)) {
			s.openaiCodexTurnStateOrigins.Delete(key)
		}
		return true
	})
}
