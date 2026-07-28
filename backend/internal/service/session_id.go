package service

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const maxPersistedSessionIDLength = 255

// ExtractClientSessionID returns an explicit client correlation identifier for
// usage records. It is deliberately separate from GenerateSessionHash: usage
// correlation must never be synthesized from request content or cache keys.
func ExtractClientSessionID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	for _, header := range []string{
		"session_id",
		"X-Session-Id",
		"X-Session-Affinity",
		"X-OpenCode-Session",
		"X-Conversation-ID",
		"X-Grok-Conv-Id",
		"conversation_id",
		"X-Claude-Code-Session-Id",
	} {
		if value := sanitizeClientSessionID(c.GetHeader(header)); value != "" {
			return value
		}
	}
	return ""
}

func sanitizeClientSessionID(raw string) string {
	if !utf8.ValidString(raw) {
		return ""
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	count := 0
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return ""
		}
		count++
		if count > maxPersistedSessionIDLength {
			return ""
		}
	}
	return value
}
