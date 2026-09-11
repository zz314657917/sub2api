package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	codexReservedPythonToolName = "python"
	codexPythonToolAlias        = "python__sub2api"
	codexToolNameReverseKey     = "openai_codex_tool_name_reverse"
)

type codexToolNameField struct {
	object    map[string]any
	key, name string
}

func aliasOpenAIOAuthReservedToolNames(reqBody map[string]any) (map[string]string, bool, error) {
	if reqBody == nil {
		return nil, false, nil
	}
	fields := collectOpenAIResponsesToolNameFields(reqBody)
	owners, reverse := map[string]string{}, map[string]string{}
	for _, field := range fields {
		aliased := aliasOpenAIOAuthReservedToolName(field.name)
		original := field.name
		if aliased != field.name {
			original = strings.TrimSpace(field.name)
		}
		if previous, exists := owners[aliased]; exists && previous != original {
			return nil, false, fmt.Errorf("tool names %q and %q both normalize to %q", previous, original, aliased)
		}
		owners[aliased] = original
		if aliased != field.name {
			reverse[aliased] = original
		}
	}
	if len(reverse) == 0 {
		return nil, false, nil
	}
	for _, field := range fields {
		if aliased := aliasOpenAIOAuthReservedToolName(field.name); aliased != field.name {
			field.object[field.key] = aliased
		}
	}
	return reverse, true, nil
}

func aliasOpenAIOAuthReservedToolName(name string) string {
	if strings.EqualFold(strings.TrimSpace(name), codexReservedPythonToolName) {
		return codexPythonToolAlias
	}
	return name
}

func collectOpenAIResponsesToolNameFields(reqBody map[string]any) []codexToolNameField {
	fields := make([]codexToolNameField, 0, 8)
	appendName := func(object map[string]any, key string) {
		if name, ok := object[key].(string); ok && strings.TrimSpace(name) != "" {
			fields = append(fields, codexToolNameField{object, key, name})
		}
	}
	var collectTools func(any)
	collectTools = func(raw any) {
		tools, ok := raw.([]any)
		if !ok {
			return
		}
		for _, rawTool := range tools {
			tool, ok := rawTool.(map[string]any)
			if !ok {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(firstNonEmptyString(tool["type"])), "namespace") {
				appendName(tool, "name")
			}
			if function, ok := tool["function"].(map[string]any); ok {
				appendName(function, "name")
			}
			collectTools(tool["tools"])
		}
	}
	collectTools(reqBody["tools"])
	collectTools(reqBody["functions"])
	if choice, ok := reqBody["tool_choice"].(map[string]any); ok {
		if !strings.EqualFold(strings.TrimSpace(firstNonEmptyString(choice["type"])), "namespace") {
			appendName(choice, "name")
		}
		if function, ok := choice["function"].(map[string]any); ok {
			appendName(function, "name")
		}
		// allowed_tools carries nested declarations/selections; they must follow
		// the same alias map without broadening the policy to auto.
		collectTools(choice["tools"])
	}
	if input, ok := reqBody["input"].([]any); ok {
		for _, rawItem := range input {
			item, ok := rawItem.(map[string]any)
			if !ok {
				continue
			}
			typ := strings.ToLower(strings.TrimSpace(firstNonEmptyString(item["type"])))
			if typ == "additional_tools" {
				collectTools(item["tools"])
			}
			if strings.HasSuffix(typ, "_call") || typ == "tool_call" {
				appendName(item, "name")
				if function, ok := item["function"].(map[string]any); ok {
					appendName(function, "name")
				}
			}
		}
	}
	return fields
}

func aliasOpenAIOAuthReservedToolNamesBody(body []byte) ([]byte, map[string]string, bool, error) {
	if len(body) == 0 || !containsASCIIFold(body, []byte(codexReservedPythonToolName)) {
		return body, nil, false, nil
	}
	var reqBody map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&reqBody); err != nil {
		return body, nil, false, fmt.Errorf("decode OAuth reserved tool names: %w", err)
	}
	reverse, changed, err := aliasOpenAIOAuthReservedToolNames(reqBody)
	if err != nil || !changed {
		return body, reverse, false, err
	}
	normalized, err := json.Marshal(reqBody)
	if err != nil {
		return body, nil, false, fmt.Errorf("encode OAuth reserved tool names: %w", err)
	}
	return normalized, reverse, true, nil
}

func containsASCIIFold(haystack, needle []byte) bool {
	return bytes.Contains(bytes.ToLower(haystack), bytes.ToLower(needle))
}

func setCodexToolNameReverse(c *gin.Context, reverse map[string]string) {
	if c == nil {
		return
	}
	if len(reverse) == 0 {
		c.Set(codexToolNameReverseKey, nil)
		return
	}
	copyMap := make(map[string]string, len(reverse))
	for aliased, original := range reverse {
		copyMap[aliased] = original
	}
	c.Set(codexToolNameReverseKey, copyMap)
}
func codexToolNameReverseFromContext(c *gin.Context) map[string]string {
	if c == nil {
		return nil
	}
	raw, ok := c.Get(codexToolNameReverseKey)
	if !ok {
		return nil
	}
	reverse, _ := raw.(map[string]string)
	return reverse
}
func restoreCodexToolNamesInJSON(data []byte, reverse map[string]string) []byte {
	if len(data) == 0 || len(reverse) == 0 || !json.Valid(data) {
		return data
	}
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil || !restoreCodexToolNameFields(decoded, reverse) {
		return data
	}
	restored, err := json.Marshal(decoded)
	if err != nil {
		return data
	}
	return restored
}
func restoreCodexToolNamesFromContext(c *gin.Context, data []byte) []byte {
	return restoreCodexToolNamesInJSON(data, codexToolNameReverseFromContext(c))
}
func restoreCodexToolNameFields(value any, reverse map[string]string) bool {
	changed := false
	switch typed := value.(type) {
	case map[string]any:
		typ := strings.ToLower(strings.TrimSpace(firstNonEmptyString(typed["type"])))
		_, hasCallID := typed["call_id"]
		if name, ok := typed["name"].(string); ok && (hasCallID || strings.HasSuffix(typ, "_call") || typ == "tool_call") {
			if original, exists := reverse[name]; exists {
				typed["name"] = original
				changed = true
			}
		}
		for _, child := range typed {
			if restoreCodexToolNameFields(child, reverse) {
				changed = true
			}
		}
	case []any:
		for _, child := range typed {
			if restoreCodexToolNameFields(child, reverse) {
				changed = true
			}
		}
	}
	return changed
}
