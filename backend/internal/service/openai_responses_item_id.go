package service

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Invalid replayed IDs are removed rather than rewritten because a fabricated
// ID may point at a different upstream object.
func shouldStripOpenAIResponsesInputItemID(itemType, id string) bool {
	itemType = strings.TrimSpace(itemType)
	if id == "" {
		return false
	}
	if prefix, constrained := openAIResponsesInputItemIDPrefix(itemType); constrained {
		return !strings.HasPrefix(id, prefix)
	}
	return false
}

func openAIResponsesInputItemIDPrefix(itemType string) (string, bool) {
	switch strings.TrimSpace(itemType) {
	case "message":
		return "msg", true
	case "reasoning":
		return "rs", true
	case "custom_tool_call", "tool_search_call":
		return openAIResponsesToolCallIDPrefix(itemType), true
	case "custom_tool_call_output":
		// Local continuation output item IDs retain the generic fc namespace;
		// only their paired call_id is ctc_.
		return "fc", true
	default:
		if isCodexToolCallInputType(itemType) {
			return openAIResponsesToolCallIDPrefix(itemType), true
		}
		return "", false
	}
}

func openAIResponsesToolCallIDPrefix(itemType string) string {
	switch strings.TrimSpace(itemType) {
	case "custom_tool_call", "custom_tool_call_output":
		return "ctc"
	case "tool_search_call", "tool_search_output":
		return "tsc"
	default:
		return "fc"
	}
}

func sanitizeOpenAIResponsesInputItemIDs(body []byte) ([]byte, bool, error) {
	input := gjson.GetBytes(body, "input")
	if !input.IsArray() {
		return body, false, nil
	}

	items := make([][]byte, 0)
	changed := false
	var sanitizeErr error
	index := 0
	input.ForEach(func(_, item gjson.Result) bool {
		currentIndex := index
		index++
		itemBody := []byte(item.Raw)
		if item.IsObject() {
			itemType := item.Get("type")
			id := item.Get("id")
			if itemType.Type == gjson.String && id.Type == gjson.String &&
				shouldStripOpenAIResponsesInputItemID(itemType.String(), id.String()) {
				itemBody, sanitizeErr = sjson.DeleteBytes(itemBody, "id")
				if sanitizeErr != nil {
					sanitizeErr = fmt.Errorf("delete input.%d.id: %w", currentIndex, sanitizeErr)
					return false
				}
				changed = true
			}
		}
		items = append(items, itemBody)
		return true
	})
	if sanitizeErr != nil {
		return nil, false, sanitizeErr
	}
	if !changed {
		return body, false, nil
	}

	rebuiltInput := make([]byte, 0, len(input.Raw))
	rebuiltInput = append(rebuiltInput, '[')
	for i, item := range items {
		if i > 0 {
			rebuiltInput = append(rebuiltInput, ',')
		}
		rebuiltInput = append(rebuiltInput, item...)
	}
	rebuiltInput = append(rebuiltInput, ']')

	sanitized, err := sjson.SetRawBytes(body, "input", rebuiltInput)
	if err != nil {
		return nil, false, fmt.Errorf("replace sanitized input: %w", err)
	}
	return sanitized, true, nil
}
