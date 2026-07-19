package service

import (
	"fmt"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestS79ParseGatewayRequestClaudeCodeLongContext(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		model    string
		want     string
	}{
		{name: "lowercase trailing selector", protocol: domain.PlatformAnthropic, model: "claude-opus-4-8[1m]", want: "claude-opus-4-8"},
		{name: "uppercase trailing selector", protocol: domain.PlatformAnthropic, model: "claude-opus-4-8[1M]", want: "claude-opus-4-8"},
		{name: "repeated trailing selectors", protocol: domain.PlatformAnthropic, model: "claude-opus-4-8[1M][1m]", want: "claude-opus-4-8"},
		{name: "selector in middle", protocol: domain.PlatformAnthropic, model: "claude-opus-4-8[1m]-preview", want: "claude-opus-4-8[1m]-preview"},
		{name: "selector only", protocol: domain.PlatformAnthropic, model: "[1m]", want: "[1m]"},
		{name: "non anthropic protocol", protocol: "responses", model: "claude-opus-4-8[1m]", want: "claude-opus-4-8[1m]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"model":%q,"system":"test","messages":[{"role":"user","content":"hi"}]}`, tt.model))
			parsed, err := ParseGatewayRequest(body, tt.protocol)
			require.NoError(t, err)
			require.Equal(t, tt.want, parsed.Model)
			require.Equal(t, tt.want, gjson.GetBytes(parsed.Body, "model").String())
			require.Equal(t, "test", parsed.System)
			require.Len(t, parsed.Messages, 1)
		})
	}
}
