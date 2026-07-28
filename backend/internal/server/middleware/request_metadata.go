package middleware

import (
	"strings"
	"unicode/utf8"
)

const maxPersistentUserAgentBytes = 512

// normalizePersistentText bounds attacker-controlled metadata before it is
// persisted in an audit record or used as a session-binding input.
func normalizePersistentText(value string, maxBytes int) string {
	value = strings.TrimSpace(strings.ToValidUTF8(value, ""))
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}
