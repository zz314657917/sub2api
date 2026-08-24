package service

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// stripEmptyChatToolCallIdentityFromSSELine removes only present empty tool
// identity fields from raw Chat Completions stream deltas. Some compatible
// upstreams emit id/name in the first delta and empty strings in later
// argument-only deltas; clients that merge on field presence then overwrite
// the original identity with an empty value.
func stripEmptyChatToolCallIdentityFromSSELine(line string) string {
	payload, ok := extractOpenAISSEDataLine(line)
	if !ok {
		return line
	}
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" || trimmed == "[DONE]" {
		return line
	}
	rewritten, changed := stripEmptyChatToolCallIdentity([]byte(payload))
	if !changed {
		return line
	}
	prefixLen := len(line) - len(payload)
	if prefixLen < 0 {
		return line
	}
	return line[:prefixLen] + string(rewritten)
}

// stripEmptyChatToolCallIdentity deletes empty-string id and function.name
// fields under choices[*].delta.tool_calls[*]. Arguments, index, type,
// non-empty values, non-tool payloads, and invalid JSON are left untouched.
func stripEmptyChatToolCallIdentity(payload []byte) ([]byte, bool) {
	if len(payload) == 0 || !bytes.Contains(payload, []byte("tool_calls")) || !gjson.ValidBytes(payload) {
		return payload, false
	}
	choices := gjson.GetBytes(payload, "choices")
	if !choices.Exists() || !choices.IsArray() {
		return payload, false
	}
	updated := payload
	changed := false
	for choiceIndex, choice := range choices.Array() {
		delta := choice.Get("delta")
		if !delta.Exists() || !delta.IsObject() {
			continue
		}
		toolCalls := delta.Get("tool_calls")
		if !toolCalls.Exists() || !toolCalls.IsArray() {
			continue
		}
		for toolIndex, toolCall := range toolCalls.Array() {
			if id := toolCall.Get("id"); id.Exists() && id.Type == gjson.String && id.Str == "" {
				next, err := sjson.DeleteBytes(updated, "choices."+strconv.Itoa(choiceIndex)+".delta.tool_calls."+strconv.Itoa(toolIndex)+".id")
				if err != nil {
					return payload, false
				}
				updated = next
				changed = true
			}
			if name := toolCall.Get("function.name"); name.Exists() && name.Type == gjson.String && name.Str == "" {
				next, err := sjson.DeleteBytes(updated, "choices."+strconv.Itoa(choiceIndex)+".delta.tool_calls."+strconv.Itoa(toolIndex)+".function.name")
				if err != nil {
					return payload, false
				}
				updated = next
				changed = true
			}
		}
	}
	return updated, changed
}
