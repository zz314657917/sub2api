package claude

import (
	"strings"
	"unicode"
)

// IsOpus55 recognizes the fixed model ID, including existing provider, thinking
// and dated spellings, without treating other Opus 5 models as 5.5.
func IsOpus55(model string) bool {
	id := strings.ToLower(strings.TrimSpace(model))
	if slash := strings.LastIndexByte(id, '/'); slash >= 0 {
		id = id[slash+1:]
	}
	id = strings.TrimPrefix(id, "anthropic.")
	id = strings.TrimSuffix(id, "-thinking")
	if id == "claude-opus-5-5" {
		return true
	}
	suffix, ok := strings.CutPrefix(id, "claude-opus-5-5-")
	if !ok || len(suffix) != 8 {
		return false
	}
	for _, digit := range suffix {
		if !unicode.IsDigit(digit) {
			return false
		}
	}
	return true
}
