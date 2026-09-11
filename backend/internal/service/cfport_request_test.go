package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCFPortReservedToolAliasKeepsAllowedToolsAndDoesNotRewriteContent(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"python","parameters":{"type":"object"}}],"tool_choice":{"type":"allowed_tools","tools":[{"type":"function","name":"python"}]},"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"python"}]},{"type":"additional_tools","tools":[{"type":"function","name":"python"}]}]}`)
	normalized, reverse, changed, err := aliasOpenAIOAuthReservedToolNamesBody(body)
	if err != nil || !changed || reverse[codexPythonToolAlias] != "python" {
		t.Fatalf("alias result changed=%v reverse=%v err=%v", changed, reverse, err)
	}
	var request map[string]any
	if err := json.Unmarshal(normalized, &request); err != nil {
		t.Fatal(err)
	}
	if got := request["tool_choice"].(map[string]any)["type"]; got != "allowed_tools" {
		t.Fatalf("allowed_tools changed to %v", got)
	}
	selected := request["tool_choice"].(map[string]any)["tools"].([]any)[0].(map[string]any)["name"]
	if selected != codexPythonToolAlias {
		t.Fatalf("nested allowed_tools name was not aliased: %v", selected)
	}
	response := []byte(`{"type":"function_call","call_id":"fc_x","name":"python__sub2api","output":{"name":"python__sub2api"},"content":"python__sub2api"}`)
	restored := restoreCodexToolNamesInJSON(response, reverse)
	if strings.Contains(string(restored), `"content":"python"`) || strings.Contains(string(restored), `"output":{"name":"python"}`) {
		t.Fatalf("arbitrary content was restored: %s", restored)
	}
	if !strings.Contains(string(restored), `"name":"python"`) {
		t.Fatalf("tool name was not restored: %s", restored)
	}
}

func TestCFPortNativeToolIDsAndReferences(t *testing.T) {
	if got := normalizeCodexCallIDForItemType("custom_tool_call", "call_x"); got != "ctc_x" {
		t.Fatalf("custom id=%q", got)
	}
	if got := normalizeCodexCallIDForItemType("tool_search_call", "call_x"); got != "tsc_x" {
		t.Fatalf("search id=%q", got)
	}
	body := []byte(`{"input":[{"type":"custom_tool_call","id":"ctc_ok"},{"type":"tool_search_call","id":"fc_wrong"},{"type":"function_call","id":"ctc_wrong"}]}`)
	normalized, changed, err := sanitizeOpenAIResponsesInputItemIDs(body)
	if err != nil || !changed {
		t.Fatalf("sanitize changed=%v err=%v", changed, err)
	}
	if shouldStripOpenAIResponsesInputItemID("custom_tool_call_output", "fc_item") || !shouldStripOpenAIResponsesInputItemID("custom_tool_call_output", "ctc_item") {
		t.Fatal("custom output item ID must retain local fc-only validation")
	}
	if strings.Contains(string(normalized), "fc_wrong") || strings.Contains(string(normalized), "ctc_wrong") || !strings.Contains(string(normalized), "ctc_ok") {
		t.Fatalf("unexpected ID sanitation: %s", normalized)
	}
	filtered := filterCodexInputWithOptions([]any{
		map[string]any{"type": "custom_tool_call", "call_id": "call_native"},
		map[string]any{"type": "item_reference", "id": "call_native"},
	}, codexInputFilterOptions{PreserveReferences: true})
	if got := filtered[1].(map[string]any)["id"]; got != "ctc_native" {
		t.Fatalf("reference must follow native call namespace, got %v", got)
	}
}

func TestCFPortCompatibilityPreservesAstraProAndNumbers(t *testing.T) {
	astra := []byte(`{"model":"gpt-6-astra","reasoning":{"mode":"pro"},"n":9007199254740993}`)
	normalized, changed, err := normalizeOpenAIResponsesReasoningMode(astra)
	if err != nil || changed || string(normalized) != string(astra) {
		t.Fatalf("astra changed=%v body=%s err=%v", changed, normalized, err)
	}
	body := []byte(`{"model":"gpt-5.5","reasoning":{"mode":"pro"},"prompt":"hello","commands":["x"],"n":9007199254740993}`)
	normalized, changed, err = normalizeOpenAIOAuthResponsesCompatibilityBody(body)
	if err != nil || !changed || !strings.Contains(string(normalized), "9007199254740993") || strings.Contains(string(normalized), `"prompt"`) {
		t.Fatalf("OAuth compatibility body=%s changed=%v err=%v", normalized, changed, err)
	}
	normalized, changed, err = normalizeOpenAIOAuthResponsesCompatibilityBody([]byte(`{"prompt":"hello","input":null}`))
	if err != nil || !changed || !strings.Contains(string(normalized), `"input":"hello"`) {
		t.Fatalf("null input was not replaced: %s err=%v", normalized, err)
	}
	normalized, changed, err = normalizeOpenAIResponsesReasoningMode([]byte(`{"model":"gpt-5.5","reasoning":{"mode":"pro"}}`))
	if err != nil || !changed || !strings.Contains(string(normalized), `"effort":"max"`) || strings.Contains(string(normalized), `"mode"`) {
		t.Fatalf("reasoning body=%s changed=%v err=%v", normalized, changed, err)
	}
	normalized, changed, err = normalizeOpenAIResponsesReasoningMode([]byte(`{"model":"gpt-5.5","reasoning":{"mode":"legacy"}}`))
	if err != nil || !changed || strings.Contains(string(normalized), `"mode"`) {
		t.Fatalf("non-pro mode was not removed: %s err=%v", normalized, err)
	}
	req := map[string]any{"model": "gpt-5.5", "input": []any{}, "chat_template_kwargs": true, "truncation": "auto", "stop_sequences": []any{"x"}}
	applyCodexOAuthTransform(req, false, false)
	for _, key := range []string{"chat_template_kwargs", "truncation", "stop_sequences"} {
		if _, exists := req[key]; exists {
			t.Fatalf("transform retained unsupported %s", key)
		}
	}
	wsBody := []byte(`{"model":"gpt-5.5","chat_template_kwargs":true,"truncation":"auto","stop_sequences":["x"],"tools":[{"type":"function","parameters":{"type":null}}]}`)
	normalized, changed, err = normalizeOpenAIResponsesWebSocketCompatibilityBody(wsBody, &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth})
	if err != nil || !changed || strings.Contains(string(normalized), "chat_template_kwargs") || strings.Contains(string(normalized), "truncation") || strings.Contains(string(normalized), "stop_sequences") || !strings.Contains(string(normalized), `"type":"object"`) {
		t.Fatalf("WS compatibility incomplete: %s changed=%v err=%v", normalized, changed, err)
	}
	whitespaceSchema := []byte(`{"tools":[{"type":"function","parameters":{"type": null}}]}`)
	normalized, changed, err = sanitizeOpenAIResponsesToolParameterTypes(whitespaceSchema)
	if err != nil || !changed || !strings.Contains(string(normalized), `"type":"object"`) {
		t.Fatalf("whitespace null type was not normalized: %s changed=%v err=%v", normalized, changed, err)
	}
}

func TestCFPortSchemaInputAndRejectedFieldHelpers(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","parameters":{"type":"object","properties":{"x":{"pattern":"(?=x)foo"}}}}],"text":{"format":{"type":"json_schema","schema":{"properties":{"a":{"uniqueItems":true}}}}}}`)
	normalized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns(body)
	if err != nil || !changed || strings.Contains(string(normalized), `"pattern"`) {
		t.Fatalf("pattern body=%s changed=%v err=%v", normalized, changed, err)
	}
	normalized, changed, err = normalizeOpenAIResponseFormatSchemasBody(normalized)
	if err != nil || !changed || strings.Contains(string(normalized), "uniqueItems") || !strings.Contains(string(normalized), `"type":"object"`) {
		t.Fatalf("schema body=%s changed=%v err=%v", normalized, changed, err)
	}
	retry, reason, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(400, []byte(`{"input":[{"type":"reasoning","status":"completed","content":null}]}`), []byte(`{"error":{"code":"unknown_parameter","param":"input[0].status","message":"Unknown parameter input[0].status"}}`))
	if err != nil || !changed || reason == "" || strings.Contains(string(retry), "status") {
		t.Fatalf("retry=%s reason=%q changed=%v err=%v", retry, reason, changed, err)
	}
	request := map[string]any{"input": []any{map[string]any{"type": "function_call_output", "call_id": "gone"}, map[string]any{"type": "function_call", "call_id": "kept"}, map[string]any{"type": "function_call_output", "call_id": "kept"}}}
	if !sanitizeOpenAIResponsesOrphanToolOutputs(request, request["input"].([]any), false) || strings.Contains(fmtJSON(request), "gone") {
		t.Fatalf("orphan output not removed: %s", fmtJSON(request))
	}
	tooLong := strings.Repeat("界", openAIResponsesInputTextMaxChars+1)
	input := map[string]any{"input": []any{map[string]any{"type": "message", "content": []any{map[string]any{"text": tooLong}}}}}
	if !truncateOpenAIResponsesInputText(input) || len([]rune(input["input"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string))) != openAIResponsesInputTextMaxChars {
		t.Fatal("Unicode input truncation failed")
	}
}

func TestCFPortCompactionTriggerOrderPreservesLargeNumbers(t *testing.T) {
	body := []byte(`{"input":[{"type":"compaction_trigger"},{"type":"message","n":9007199254740993},{"type":"compaction_trigger"}]}`)
	normalized, changed, err := NormalizeCompactionTriggerInputOrder(body)
	if err != nil || !changed {
		t.Fatalf("compaction changed=%v err=%v", changed, err)
	}
	if !strings.Contains(string(normalized), "9007199254740993") {
		t.Fatalf("large integer lost: %s", normalized)
	}
	var payload map[string]any
	if err := json.Unmarshal(normalized, &payload); err != nil {
		t.Fatal(err)
	}
	input := payload["input"].([]any)
	if len(input) != 2 || input[1].(map[string]any)["type"] != "compaction_trigger" {
		t.Fatalf("trigger not collapsed/moved: %s", normalized)
	}
}

func TestCFPortRejectedRetryNullContentAndExactCacheOnly(t *testing.T) {
	tests := []struct{ name, body, response, want string }{
		{"reasoning", `{"input":[{"type":"reasoning","content":null}]}`, `{"error":{"code":"invalid_type","param":"input[0].content","message":"Invalid type for input[0].content: got null"}}`, `"content"`},
		{"message", `{"input":[{"type":"message","content":null}]}`, `{"error":{"code":"invalid_type","param":"input[0].content","message":"Invalid type for input[0].content: got null"}}`, `"content":""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(400, []byte(tt.body), []byte(tt.response))
			if err != nil || !changed {
				t.Fatalf("changed=%v err=%v", changed, err)
			}
			if tt.name == "reasoning" && strings.Contains(string(next), tt.want) {
				t.Fatalf("reasoning null content remains: %s", next)
			}
			if tt.name == "message" && !strings.Contains(string(next), tt.want) {
				t.Fatalf("message null content not normalized: %s", next)
			}
		})
	}
	top := []byte(`{"prompt_cache_breakpoint":true,"input":[{"prompt_cache_breakpoint":true},{"prompt_cache_breakpoint":true}]}`)
	next, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(400, top, []byte(`{"error":{"code":"invalid_parameter","param":"prompt_cache_breakpoint","message":"prompt_cache_breakpoint is not supported on this model"}}`))
	if err != nil || !changed || !strings.HasPrefix(string(next), `{"input"`) || strings.Count(string(next), "prompt_cache_breakpoint") != 2 {
		t.Fatalf("top cache deletion was not exact: %s err=%v", next, err)
	}
	next, _, changed, err = normalizeOpenAIResponsesRejectedFieldRetryBody(400, top, []byte(`{"error":{"code":"invalid_parameter","param":"input[1].prompt_cache_breakpoint","message":"input[1].prompt_cache_breakpoint is not supported on this model"}}`))
	if err != nil || !changed || strings.Count(string(next), "prompt_cache_breakpoint") != 2 || !strings.Contains(string(next), `"input":[{"prompt_cache_breakpoint":true},{}`) {
		t.Fatalf("indexed cache deletion was not exact: %s err=%v", next, err)
	}
}

func TestCFPortRejectedRetryRefusesAmbiguousOrUnrelatedErrors(t *testing.T) {
	body := []byte(`{"input":[{"type":"message","status":"done","content":null}],"prompt_cache_breakpoint":true}`)
	responses := [][]byte{
		[]byte(`{"error":{"code":"unknown_parameter","param":"input[0].status","message":"unsupported parameter input[0].namespace"}}`),
		[]byte(`{"error":{"code":"invalid_request_error","param":"input[0].status","message":"ordinary upstream error"}}`),
	}
	for _, response := range responses {
		next, reason, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(400, body, response)
		if err != nil || changed || next != nil || reason != "" {
			t.Fatalf("unrelated error mutated request: next=%s reason=%q changed=%v err=%v", next, reason, changed, err)
		}
	}
	next, _, changed, err := normalizeOpenAIResponsesRejectedFieldRetryBody(500, body, []byte(`{"error":{"code":"unknown_parameter","param":"input[0].status"}}`))
	if err != nil || changed || next != nil {
		t.Fatalf("non-400 must not retry: next=%s changed=%v err=%v", next, changed, err)
	}
}

func TestCFPortRejectedRetryStateSharesBudgetAndRejectsDuplicates(t *testing.T) {
	budget := &openAIResponsesRejectedFieldRetryBudget{}
	first := newOpenAIResponsesRejectedFieldRetryStateWithBudget([]byte(`{"n":0}`), budget)
	second := newOpenAIResponsesRejectedFieldRetryStateWithBudget([]byte(`{"n":0}`), budget)
	if first.Allow([]byte(`{"n":0}`)) || !first.Allow([]byte(`{"n":1}`)) || second.Allow([]byte(`{"n":0}`)) {
		t.Fatal("duplicate request body was allowed")
	}
	for i := 2; i <= maxOpenAIResponsesRejectedFieldRetries; i++ {
		if !second.Allow([]byte(fmtJSON(map[string]int{"n": i}))) {
			t.Fatalf("shared budget rejected attempt %d too early", i)
		}
	}
	if first.Allow([]byte(`{"n":99}`)) {
		t.Fatal("shared retry budget exceeded")
	}
}

func TestCFPortAliasCollisionAndReferenceAmbiguityStayAtomic(t *testing.T) {
	request := map[string]any{"tools": []any{map[string]any{"type": "function", "name": "python"}, map[string]any{"type": "function", "name": codexPythonToolAlias}}}
	before := fmtJSON(request)
	reverse, changed, err := aliasOpenAIOAuthReservedToolNames(request)
	if err == nil || changed || reverse != nil || fmtJSON(request) != before {
		t.Fatalf("alias collision partially mutated request: changed=%v reverse=%v err=%v body=%s", changed, reverse, err, fmtJSON(request))
	}
	input := []any{
		map[string]any{"type": "function_call", "call_id": "call_same"},
		map[string]any{"type": "custom_tool_call", "call_id": "call_same"},
		map[string]any{"type": "item_reference", "id": "call_same"},
		map[string]any{"type": "item_reference", "id": "call_unknown"},
	}
	filtered := filterCodexInputWithOptions(input, codexInputFilterOptions{PreserveReferences: true})
	if got := filtered[2].(map[string]any)["id"]; got != "call_same" {
		t.Fatalf("ambiguous reference changed to %v", got)
	}
	if got := filtered[3].(map[string]any)["id"]; got != "call_unknown" {
		t.Fatalf("unknown reference changed to %v", got)
	}
}

func fmtJSON(value any) string { body, _ := json.Marshal(value); return string(body) }
