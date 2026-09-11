package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// Adapted from upstream acce29af2. Stateless reasoning must be replayed by
// encrypted content, not by looking up an rs_* item that was never persisted.
func normalizeOpenAIAPIKeyStoreFalseReasoningReplay(body []byte, knownStoreFalse bool) ([]byte, bool, error) {
	if !knownStoreFalse && gjson.GetBytes(body, "store").Type != gjson.False {
		return body, false, nil
	}
	if !gjson.GetBytes(body, "input").IsArray() {
		return body, false, nil
	}
	if !json.Valid(body) {
		return body, false, fmt.Errorf("invalid stateless Responses JSON")
	}
	var request map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		return body, false, fmt.Errorf("decode stateless Responses replay: %w", err)
	}
	items, ok := request["input"].([]any)
	if !ok {
		return body, false, nil
	}
	filtered := make([]any, 0, len(items))
	changed := false
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			filtered = append(filtered, raw)
			continue
		}
		typ := strings.TrimSpace(firstNonEmptyString(item["type"]))
		id := strings.TrimSpace(firstNonEmptyString(item["id"]))
		switch typ {
		case "reasoning":
			encrypted, _ := item["encrypted_content"].(string)
			if strings.TrimSpace(encrypted) == "" {
				changed = true
				continue
			}
			if strings.HasPrefix(id, "rs_") {
				delete(item, "id")
				changed = true
			}
			if item["summary"] == nil {
				item["summary"] = []any{}
				changed = true
			}
		case "item_reference":
			if strings.HasPrefix(id, "rs_") {
				changed = true
				continue
			}
		}
		switch typ {
		case "reasoning", "message", "image_generation_call":
			if _, exists := item["call_id"]; exists {
				delete(item, "call_id")
				changed = true
			}
		}
		filtered = append(filtered, item)
	}
	if !changed {
		return body, false, nil
	}
	request["input"] = filtered
	normalized, err := json.Marshal(request)
	if err != nil {
		return body, false, fmt.Errorf("encode stateless Responses replay: %w", err)
	}
	return normalized, true, nil
}
