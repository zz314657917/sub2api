package openai

import "testing"

const (
	testClaudeCodeOriginator = "Claude Code"
	testClaudeCodeUserAgent  = "Claude Code/0.5.0 (Macos 15.5; arm64) iTerm2.app (Claude Code; 1.0.4)"
)

func TestIsAllowedClientMatch(t *testing.T) {
	entry := AllowedClientEntry{Originator: "Claude Code", UAContains: []string{"Claude Code/"}}

	tests := []struct {
		name       string
		ua         string
		originator string
		want       bool
	}{
		{name: "real signature", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, want: true},
		{name: "case insensitive", ua: "claude code/0.5.0 (macos)", originator: "claude code", want: true},
		{name: "trim originator", ua: testClaudeCodeUserAgent, originator: "  Claude Code  ", want: true},
		{name: "originator suffix rejected", ua: testClaudeCodeUserAgent, originator: "Claude Code Extra", want: false},
		{name: "empty originator rejected", ua: testClaudeCodeUserAgent, originator: "", want: false},
		{name: "codex originator rejected", ua: testClaudeCodeUserAgent, originator: "codex_cli_rs", want: false},
		{name: "missing ua marker rejected", ua: "curl/8.0", originator: testClaudeCodeOriginator, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAllowedClientMatch(tt.ua, tt.originator, entry); got != tt.want {
				t.Fatalf("IsAllowedClientMatch(%q, %q) = %v, want %v", tt.ua, tt.originator, got, tt.want)
			}
		})
	}
}

func TestIsAllowedClientMatch_UnsafePresetRejected(t *testing.T) {
	tests := []struct {
		name  string
		entry AllowedClientEntry
	}{
		{name: "empty originator", entry: AllowedClientEntry{Originator: "", UAContains: []string{"Claude Code/"}}},
		{name: "empty ua contains", entry: AllowedClientEntry{Originator: "Claude Code", UAContains: nil}},
		{name: "whitespace ua marker", entry: AllowedClientEntry{Originator: "Claude Code", UAContains: []string{"   "}}},
		{name: "mixed empty ua marker", entry: AllowedClientEntry{Originator: "Claude Code", UAContains: []string{"", "Claude Code/"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsAllowedClientMatch(testClaudeCodeUserAgent, testClaudeCodeOriginator, tt.entry) {
				t.Fatal("unsafe allowed client preset should not match")
			}
		})
	}
}

func TestMatchAllowedClients(t *testing.T) {
	tests := []struct {
		name       string
		ua         string
		originator string
		clientIDs  []string
		want       bool
	}{
		{name: "claude code preset", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, clientIDs: []string{AllowedClientClaudeCode}, want: true},
		{name: "spoofed originator rejected", ua: testClaudeCodeUserAgent, originator: "my_client", clientIDs: []string{AllowedClientClaudeCode}, want: false},
		{name: "empty list rejected", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, clientIDs: nil, want: false},
		{name: "unknown preset rejected", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, clientIDs: []string{"unknown_client"}, want: false},
		{name: "id normalized", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, clientIDs: []string{"  Claude_Code "}, want: true},
		{name: "any preset match", ua: testClaudeCodeUserAgent, originator: testClaudeCodeOriginator, clientIDs: []string{"unknown_client", AllowedClientClaudeCode}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchAllowedClients(tt.ua, tt.originator, tt.clientIDs); got != tt.want {
				t.Fatalf("MatchAllowedClients(%q, %q, %v) = %v, want %v", tt.ua, tt.originator, tt.clientIDs, got, tt.want)
			}
		})
	}
}
