package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// History replay needs a declaration, not permission to perform a new search.
// Lite puts hosted declarations inside input; compact and API-key requests do
// not participate in this OAuth-only compatibility rule.
func openAIToolsContainWebSearch(raw any) bool {
	tools, _ := raw.([]any)
	for _, value := range tools {
		if tool, ok := value.(map[string]any); ok &&
			strings.HasPrefix(strings.TrimSpace(firstNonEmptyString(tool["type"])), "web_search") {
			return true
		}
	}
	return false
}

func shouldPinOpenAIWebSearchHistoryToolChoice(choice any) bool {
	if choice == nil {
		return true
	}
	if value, ok := choice.(string); ok {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "", "auto", "none":
			return true
		}
	}
	return false
}

func ensureOpenAIOAuthWebSearchToolForHistory(body map[string]any, lite bool) bool {
	input, ok := body["input"].([]any)
	if !ok {
		return false
	}
	hasSearch, declared, additional := false, false, -1
	for i, value := range input {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		switch strings.TrimSpace(firstNonEmptyString(item["type"])) {
		case "web_search_call":
			hasSearch = true
		case "additional_tools":
			if openAIToolsContainWebSearch(item["tools"]) {
				return false
			}
			tools, _ := item["tools"].([]any)
			declared = declared || len(tools) > 0
			if additional < 0 {
				additional = i
			}
		}
	}
	if !hasSearch || openAIToolsContainWebSearch(body["tools"]) {
		return false
	}
	tools, _ := body["tools"].([]any)
	declared = declared || len(tools) > 0
	tool := map[string]any{"type": "web_search", "external_web_access": false}
	switch {
	case !lite:
		body["tools"] = append(tools, tool)
	case additional >= 0:
		item, _ := input[additional].(map[string]any)
		existing, _ := item["tools"].([]any)
		item["tools"] = append(existing, tool)
	default:
		at := len(input)
		if at > 0 {
			if last, ok := input[at-1].(map[string]any); ok &&
				strings.TrimSpace(firstNonEmptyString(last["type"])) == "compaction_trigger" {
				at--
			}
		}
		item := map[string]any{"type": "additional_tools", "role": "developer", "tools": []any{tool}}
		next := make([]any, 0, len(input)+1)
		next = append(next, input[:at]...)
		next = append(next, item)
		next = append(next, input[at:]...)
		body["input"] = next
	}
	if !declared && shouldPinOpenAIWebSearchHistoryToolChoice(body["tool_choice"]) {
		body["tool_choice"] = "none"
	}
	return true
}

func ensureOpenAIOAuthWebSearchToolForHistoryBody(body []byte, lite bool) ([]byte, bool, error) {
	if !bytes.Contains(body, []byte("web_search_call")) || !gjson.ValidBytes(body) {
		return body, false, nil
	}
	var request map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber() // Do not round unrelated large metadata integers.
	if err := decoder.Decode(&request); err != nil || request == nil {
		return body, false, nil
	}
	if !ensureOpenAIOAuthWebSearchToolForHistory(request, lite) {
		return body, false, nil
	}
	next, err := marshalOpenAIUpstreamJSON(request)
	if err != nil {
		return body, false, fmt.Errorf("declare OAuth web_search history tool: %w", err)
	}
	return next, true, nil
}

// Keep the established monolithic normalizer in place; extend only its callers
// owned by this worker, instead of editing the shared compatibility owner.
func normalizeOpenAIWSOAuthHistoryCompatibilityBody(body []byte, account *Account, lite, compact bool) ([]byte, bool, error) {
	next, changed, err := normalizeOpenAIResponsesWebSocketCompatibilityBody(body, account)
	if err != nil || compact || account == nil || !account.IsOpenAIOAuth() {
		return next, changed, err
	}
	next, historyChanged, err := ensureOpenAIOAuthWebSearchToolForHistoryBody(next, lite)
	return next, changed || historyChanged, err
}
