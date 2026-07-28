package service

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExtractClientSessionIDPrecedenceAndSanitization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Conversation-ID", " conversation ")
	req.Header.Set("session_id", "primary")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	require.Equal(t, "primary", ExtractClientSessionID(c))
	req.Header.Del("session_id")
	require.Equal(t, "conversation", ExtractClientSessionID(c))
	req.Header.Set("X-Conversation-ID", "bad\nvalue")
	require.Empty(t, ExtractClientSessionID(c))
}

func TestExtractClientSessionIDRejectsInvalidLengthAndControls(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	req.Header.Set("session_id", strings.Repeat("a", maxPersistedSessionIDLength+1))
	require.Empty(t, ExtractClientSessionID(c))
	req.Header.Set("session_id", "ok\tbad")
	require.Empty(t, ExtractClientSessionID(c))
}

func TestExtractClientSessionIDSupportsOpenCodeAndGrokHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-OpenCode-Session", "opencode-session")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	require.Equal(t, "opencode-session", ExtractClientSessionID(c))
	req.Header.Del("X-OpenCode-Session")
	req.Header.Set("X-Grok-Conv-Id", "grok-session")
	require.Equal(t, "grok-session", ExtractClientSessionID(c))
}
