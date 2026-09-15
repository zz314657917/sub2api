package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClaudeCodeValidator_MaxTokensOneProbeAnyModel(t *testing.T) {
	validator := NewClaudeCodeValidator()
	for _, model := range []string{"claude-sonnet-4-5", "claude-opus-4-1", "claude-haiku-4-5"} {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/v1/messages", nil)
		req.Header.Set("User-Agent", "claude-cli/2.1.260 (external, cli)")
		require.True(t, validator.Validate(req, map[string]any{"model": model, "max_tokens": 1}))
	}
}
