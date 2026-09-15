package handler

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSetClaudeCodeClientContext_ParsedRequestProbeWithoutSystemPrompt(t *testing.T) {
	c, _ := newHelperTestContext(http.MethodPost, "/v1/messages")
	c.Request.Header.Set("User-Agent", "claude-cli/2.1.260 (external, cli)")
	SetClaudeCodeClientContext(c, nil, &service.ParsedRequest{Model: "claude-sonnet-4-5", MaxTokens: 1})
	require.True(t, service.IsClaudeCodeClient(c.Request.Context()))
}
