package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func normalizeOpenAIOAuthResponsesCompatibilityFields(reqBody map[string]any) bool {
	if reqBody == nil {
		return false
	}
	changed := false
	if prompt, exists := reqBody["prompt"]; exists {
		if input, inputExists := reqBody["input"]; (!inputExists || input == nil) && prompt != nil {
			reqBody["input"] = prompt
			changed = true
		}
		delete(reqBody, "prompt")
		changed = true
	}
	if _, exists := reqBody["commands"]; exists {
		delete(reqBody, "commands")
		changed = true
	}
	input, _ := reqBody["input"].([]any)
	for _, value := range input {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if _, exists := item["internal_chat_message_metadata_passthrough"]; exists {
			delete(item, "internal_chat_message_metadata_passthrough")
			changed = true
		}
	}
	return changed
}
func normalizeOpenAIOAuthResponsesCompatibilityBody(body []byte) ([]byte, bool, error) {
	if len(body) == 0 || (!gjson.GetBytes(body, "prompt").Exists() && !gjson.GetBytes(body, "commands").Exists() && !gjson.GetBytes(body, "input").Exists()) {
		return body, false, nil
	}
	var request map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		return body, false, fmt.Errorf("decode OAuth Responses compatibility body: %w", err)
	}
	if !normalizeOpenAIOAuthResponsesCompatibilityFields(request) {
		return body, false, nil
	}
	normalized, err := json.Marshal(request)
	if err != nil {
		return body, false, fmt.Errorf("encode OAuth Responses compatibility body: %w", err)
	}
	return normalized, true, nil
}

// normalizeOpenAIResponsesReasoningMode retains Astra's native mode=pro; older OAuth models require effort=max instead.
func normalizeOpenAIResponsesReasoningMode(body []byte) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}
	mode := gjson.GetBytes(body, "reasoning.mode")
	if !mode.Exists() || mode.Type != gjson.String {
		return body, false, nil
	}
	if strings.EqualFold(strings.TrimSpace(gjson.GetBytes(body, "model").String()), "gpt-6-astra") {
		return body, false, nil
	}
	updated := body
	effort := gjson.GetBytes(body, "reasoning.effort")
	if strings.EqualFold(strings.TrimSpace(mode.String()), "pro") && (!effort.Exists() || effort.Type == gjson.Null || strings.TrimSpace(effort.String()) == "") {
		var err error
		updated, err = sjson.SetBytes(updated, "reasoning.effort", "max")
		if err != nil {
			return body, false, fmt.Errorf("set reasoning effort for mode=pro: %w", err)
		}
	}
	next, err := sjson.DeleteBytes(updated, "reasoning.mode")
	if err != nil {
		return body, false, fmt.Errorf("delete unsupported reasoning.mode: %w", err)
	}
	updated = next
	if reasoning := gjson.GetBytes(updated, "reasoning"); reasoning.Exists() && reasoning.IsObject() && len(reasoning.Map()) == 0 {
		updated, err = sjson.DeleteBytes(updated, "reasoning")
		if err != nil {
			return body, false, fmt.Errorf("delete empty reasoning object: %w", err)
		}
	}
	return updated, true, nil
}

func normalizeOpenAIResponsesWebSocketCompatibilityBody(body []byte, account *Account) ([]byte, bool, error) {
	normalized, changed := body, false
	if account != nil && account.IsOpenAI() && account.Type == AccountTypeAPIKey {
		if next, updated, err := normalizeOpenAIAPIKeyStoreFalseReasoningReplay(normalized, false); err != nil {
			return body, false, err
		} else if updated {
			normalized, changed = next, true
		}
	}
	if next, updated, err := sanitizeOpenAIResponsesInputItemIDs(normalized); err != nil {
		return body, false, fmt.Errorf("sanitize websocket Responses input item IDs: %w", err)
	} else if updated {
		normalized, changed = next, true
	}
	if account != nil && account.IsOpenAI() && account.IsOAuth() {
		if next, updated, err := normalizeOpenAIResponsesReasoningMode(normalized); err != nil {
			return body, false, err
		} else if updated {
			normalized, changed = next, true
		}
	}
	if account != nil && account.IsOpenAIOAuth() {
		if next, updated, err := normalizeOpenAIOAuthResponsesCompatibilityBody(normalized); err != nil {
			return body, false, err
		} else if updated {
			normalized, changed = next, true
		}
		for _, field := range openAIChatGPTInternalUnsupportedFields {
			if !gjson.GetBytes(normalized, field).Exists() {
				continue
			}
			next, err := sjson.DeleteBytes(normalized, field)
			if err != nil {
				return body, false, fmt.Errorf("normalize websocket body delete %s: %w", field, err)
			}
			normalized, changed = next, true
		}
	}
	if next, updated, err := normalizeOpenAIResponseFormatSchemasBody(normalized); err != nil {
		return body, false, err
	} else if updated {
		normalized, changed = next, true
	}
	if openAIRequestBodyImageGenerationToolNeedsNormalization(normalized) {
		var request map[string]any
		decoder := json.NewDecoder(bytes.NewReader(normalized))
		decoder.UseNumber()
		if err := decoder.Decode(&request); err != nil {
			return body, false, fmt.Errorf("normalize websocket image tool body: %w", err)
		}
		if normalizeOpenAIResponsesImageGenerationTools(request) {
			next, err := json.Marshal(request)
			if err != nil {
				return body, false, fmt.Errorf("serialize normalized websocket image tool body: %w", err)
			}
			normalized, changed = next, true
		}
	}
	if next, updated, err := sanitizeOpenAIResponsesToolSchemaPatterns(normalized); err != nil {
		return body, false, fmt.Errorf("normalize websocket tool schema patterns: %w", err)
	} else if updated {
		normalized, changed = next, true
	}
	if next, updated, err := sanitizeOpenAIResponsesToolParameterTypes(normalized); err != nil {
		return body, false, fmt.Errorf("normalize websocket tool parameter types: %w", err)
	} else if updated {
		normalized, changed = next, true
	}
	return normalized, changed, nil
}

func sanitizeOpenAIResponsesToolSchemaPatterns(body []byte) ([]byte, bool, error) {
	if len(body) == 0 || !bytes.Contains(body, []byte(`"pattern"`)) {
		return body, false, nil
	}
	var root any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return nil, false, err
	}
	changed := false
	var schema func(any)
	schema = func(value any) {
		node, ok := value.(map[string]any)
		if !ok {
			return
		}
		if pattern, ok := node["pattern"].(string); ok && hasRegexLookaround(pattern) {
			delete(node, "pattern")
			changed = true
		}
		for _, key := range []string{"additionalProperties", "additionalItems", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems"} {
			schema(node[key])
		}
		if values, ok := node["items"].([]any); ok {
			for _, value := range values {
				schema(value)
			}
		} else {
			schema(node["items"])
		}
		for _, key := range []string{"anyOf", "oneOf", "allOf", "prefixItems"} {
			if values, ok := node[key].([]any); ok {
				for _, value := range values {
					schema(value)
				}
			}
		}
		for _, key := range []string{"properties", "patternProperties", "$defs", "definitions", "dependentSchemas", "dependencies"} {
			if values, ok := node[key].(map[string]any); ok {
				for _, value := range values {
					schema(value)
				}
			}
		}
	}
	var tools func(any)
	tools = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			schema(node["parameters"])
			tools(node["function"])
			tools(node["tools"])
		case []any:
			for _, value := range node {
				tools(value)
			}
		}
	}
	if document, ok := root.(map[string]any); ok {
		tools(document["tools"])
		tools(document["input"])
	}
	if !changed {
		return body, false, nil
	}
	normalized, err := json.Marshal(root)
	if err != nil {
		return nil, false, err
	}
	return normalized, true, nil
}
func hasRegexLookaround(pattern string) bool {
	return strings.Contains(pattern, "(?=") || strings.Contains(pattern, "(?!") || strings.Contains(pattern, "(?<=") || strings.Contains(pattern, "(?<!")
}
