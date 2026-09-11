package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

func normalizeOpenAIResponseFormatSchemas(reqBody map[string]any) bool {
	if reqBody == nil {
		return false
	}
	changed := false
	var normalize func(map[string]any)
	normalize = func(format map[string]any) {
		if strings.TrimSpace(firstNonEmptyString(format["type"])) != "json_schema" {
			return
		}
		if schema, ok := format["schema"].(map[string]any); ok && normalizeOpenAIResponseJSONSchema(schema) {
			changed = true
		}
		if nested, ok := format["json_schema"].(map[string]any); ok {
			if schema, ok := nested["schema"].(map[string]any); ok && normalizeOpenAIResponseJSONSchema(schema) {
				changed = true
			}
		}
	}
	if text, ok := reqBody["text"].(map[string]any); ok {
		if format, ok := text["format"].(map[string]any); ok {
			normalize(format)
		}
	}
	if format, ok := reqBody["response_format"].(map[string]any); ok {
		normalize(format)
	}
	return changed
}
func normalizeOpenAIResponseJSONSchema(schema map[string]any) bool {
	changed := false
	for _, key := range []string{"uniqueItems", "minProperties"} {
		if _, ok := schema[key]; ok {
			delete(schema, key)
			changed = true
		}
	}
	if schema["type"] == nil {
		if schema["properties"] != nil {
			schema["type"] = "object"
			changed = true
		} else if schema["items"] != nil {
			schema["type"] = "array"
			changed = true
		}
	}
	visit := func(value any) {
		if child, ok := value.(map[string]any); ok && normalizeOpenAIResponseJSONSchema(child) {
			changed = true
		}
	}
	for _, key := range []string{"additionalProperties", "additionalItems", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems"} {
		visit(schema[key])
	}
	if items, ok := schema["items"].([]any); ok {
		for _, child := range items {
			visit(child)
		}
	} else {
		visit(schema["items"])
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf", "prefixItems"} {
		if children, ok := schema[key].([]any); ok {
			for _, child := range children {
				visit(child)
			}
		}
	}
	for _, key := range []string{"properties", "patternProperties", "$defs", "definitions", "dependentSchemas", "dependencies"} {
		if children, ok := schema[key].(map[string]any); ok {
			for _, child := range children {
				visit(child)
			}
		}
	}
	return changed
}
func normalizeOpenAIResponseFormatSchemasBody(body []byte) ([]byte, bool, error) {
	if len(body) == 0 || (strings.TrimSpace(gjson.GetBytes(body, "text.format.type").String()) != "json_schema" && strings.TrimSpace(gjson.GetBytes(body, "response_format.type").String()) != "json_schema") {
		return body, false, nil
	}
	var request map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		return body, false, fmt.Errorf("normalize responses schema body: %w", err)
	}
	if !normalizeOpenAIResponseFormatSchemas(request) {
		return body, false, nil
	}
	normalized, err := json.Marshal(request)
	if err != nil {
		return body, false, fmt.Errorf("serialize normalized responses schema body: %w", err)
	}
	return normalized, true, nil
}

// sanitizeOpenAIResponsesToolParameterTypes only repairs an explicit null
// schema type. A missing type is valid JSON Schema and must remain omitted.
func sanitizeOpenAIResponsesToolParameterTypes(body []byte) ([]byte, bool, error) {
	// JSON permits whitespace around the colon, so only gate on the two tokens;
	// the decoded map below still distinguishes explicit null from omission.
	if len(body) == 0 || !bytes.Contains(body, []byte(`"type"`)) || !bytes.Contains(body, []byte(`null`)) {
		return body, false, nil
	}
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return body, false, err
	}
	changed := false
	var visitTools func(any, int)
	visitTools = func(raw any, depth int) {
		if depth > 4 {
			return
		}
		tools, ok := raw.([]any)
		if !ok {
			return
		}
		for _, rawTool := range tools {
			tool, ok := rawTool.(map[string]any)
			if !ok {
				continue
			}
			for _, candidate := range []any{tool["parameters"], func() any {
				if function, ok := tool["function"].(map[string]any); ok {
					return function["parameters"]
				}
				return nil
			}()} {
				if parameters, ok := candidate.(map[string]any); ok {
					if value, exists := parameters["type"]; exists && value == nil {
						parameters["type"] = "object"
						changed = true
					}
				}
			}
			visitTools(tool["tools"], depth+1)
		}
	}
	visitTools(root["tools"], 0)
	if input, ok := root["input"].([]any); ok {
		for _, raw := range input {
			if item, ok := raw.(map[string]any); ok {
				visitTools(item["tools"], 0)
			}
		}
	}
	if !changed {
		return body, false, nil
	}
	normalized, err := json.Marshal(root)
	if err != nil {
		return body, false, err
	}
	return normalized, true, nil
}
