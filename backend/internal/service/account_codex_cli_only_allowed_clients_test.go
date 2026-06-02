package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccount_GetCodexCLIOnlyAllowedClients(t *testing.T) {
	tests := []struct {
		name    string
		account *Account
		want    []string
	}{
		{
			name: "oauth account reads any list",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Extra:    map[string]any{"codex_cli_only_allowed_clients": []any{"claude_code"}},
			},
			want: []string{"claude_code"},
		},
		{
			name: "oauth account reads string list",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Extra:    map[string]any{"codex_cli_only_allowed_clients": []string{"claude_code"}},
			},
			want: []string{"claude_code"},
		},
		{
			name: "blank and non-string values skipped",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Extra:    map[string]any{"codex_cli_only_allowed_clients": []any{"claude_code", 123, "", "  "}},
			},
			want: []string{"claude_code"},
		},
		{
			name: "non oauth account ignored",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Extra:    map[string]any{"codex_cli_only_allowed_clients": []any{"claude_code"}},
			},
			want: nil,
		},
		{
			name:    "empty extra ignored",
			account: &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.account.GetCodexCLIOnlyAllowedClients())
		})
	}
}
