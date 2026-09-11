package service

import (
	"bytes"
	"strings"
)

const openAIResponsesInputTextMaxChars = 10000000

func sanitizeOpenAIResponsesOrphanToolOutputs(reqBody map[string]any, input []any, hasPreviousResponseID bool) bool {
	if len(input) == 0 || hasPreviousResponseID {
		return false
	}
	calls, references := map[string]struct{}{}, map[string]struct{}{}
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		typ := strings.TrimSpace(firstNonEmptyString(item["type"]))
		if typ == "item_reference" {
			if id := strings.TrimSpace(firstNonEmptyString(item["id"])); id != "" {
				references[id] = struct{}{}
			}
			continue
		}
		if isCodexToolCallContextItemType(typ) {
			if id := strings.TrimSpace(firstNonEmptyString(item["call_id"], item["id"])); id != "" {
				calls[id] = struct{}{}
			}
		}
	}
	normalized, changed := make([]any, 0, len(input)), false
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok || !isCodexToolCallOutputItemType(strings.TrimSpace(firstNonEmptyString(item["type"]))) {
			normalized = append(normalized, raw)
			continue
		}
		callID := strings.TrimSpace(firstNonEmptyString(item["call_id"]))
		if _, found := calls[callID]; found {
			normalized = append(normalized, raw)
			continue
		}
		if _, found := references[callID]; found {
			normalized = append(normalized, raw)
			continue
		}
		changed = true
	}
	if changed {
		reqBody["input"] = normalized
	}
	return changed
}

func truncateOpenAIResponsesInputText(reqBody map[string]any) bool {
	input, ok := reqBody["input"].([]any)
	if !ok {
		return false
	}
	changed := false
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if isCodexToolCallOutputItemType(strings.TrimSpace(firstNonEmptyString(item["type"]))) {
			if output, ok := item["output"].(string); ok {
				if value, truncated := truncateOpenAIResponsesInputString(output); truncated {
					item["output"] = value
					changed = true
				}
			}
		}
		if truncateOpenAIResponsesMessageText(item) {
			changed = true
		}
	}
	return changed
}
func openAIResponsesInputMayNeedTruncation(body []byte) bool {
	if len(body) <= openAIResponsesInputTextMaxChars {
		return false
	}
	return bytes.Contains(body, []byte(`"text"`)) && bytes.Contains(body, []byte(`"content"`)) || bytes.Contains(body, []byte("_tool_call_output"))
}
func truncateOpenAIResponsesMessageText(item map[string]any) bool {
	if strings.TrimSpace(firstNonEmptyString(item["type"])) != "message" && strings.TrimSpace(firstNonEmptyString(item["role"])) == "" {
		return false
	}
	parts, ok := item["content"].([]any)
	if !ok {
		return false
	}
	changed := false
	for _, raw := range parts {
		part, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := part["text"].(string); ok {
			if value, truncated := truncateOpenAIResponsesInputString(text); truncated {
				part["text"] = value
				changed = true
			}
		}
	}
	return changed
}
func truncateOpenAIResponsesInputString(value string) (string, bool) {
	if len(value) <= openAIResponsesInputTextMaxChars {
		return value, false
	}
	chars := 0
	for index := range value {
		if chars == openAIResponsesInputTextMaxChars {
			return value[:index], true
		}
		chars++
	}
	return value, false
}
