package openai

import "strings"

const (
	// AllowedClientClaudeCode is the Claude Code Codex plugin preset.
	AllowedClientClaudeCode = "claude_code"
)

type AllowedClientEntry struct {
	Originator string
	UAContains []string
}

var allowedClientRegistry = map[string]AllowedClientEntry{
	AllowedClientClaudeCode: {
		Originator: "Claude Code",
		UAContains: []string{"Claude Code/"},
	},
}

func IsAllowedClientMatch(userAgent, originator string, entry AllowedClientEntry) bool {
	wantOriginator := normalizeCodexClientHeader(entry.Originator)
	if wantOriginator == "" {
		return false
	}
	if normalizeCodexClientHeader(originator) != wantOriginator {
		return false
	}
	if len(entry.UAContains) == 0 {
		return false
	}
	ua := normalizeCodexClientHeader(userAgent)
	for _, marker := range entry.UAContains {
		normalizedMarker := normalizeCodexClientHeader(marker)
		if normalizedMarker == "" {
			return false
		}
		if !strings.Contains(ua, normalizedMarker) {
			return false
		}
	}
	return true
}

func MatchAllowedClients(userAgent, originator string, clientIDs []string) bool {
	for _, id := range clientIDs {
		entry, ok := allowedClientRegistry[normalizeCodexClientHeader(id)]
		if !ok {
			continue
		}
		if IsAllowedClientMatch(userAgent, originator, entry) {
			return true
		}
	}
	return false
}
