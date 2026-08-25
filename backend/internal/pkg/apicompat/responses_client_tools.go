package apicompat

import (
	"bytes"
	"encoding/json"
	"strings"
)

// ResponsesClientToolMapping records the reversible lowering applied before a
// native Responses request is sent to an upstream that only understands
// function tools.
type ResponsesClientToolMapping struct {
	CustomTools    map[string]bool
	ToolSearch     bool
	NamespaceTools map[string]ResponsesNamespaceName
}

// The monolithic gateway already handles namespace and tool-search tools in
// their native form. Keep the shared mapping shape for compatibility, but this
// adapter deliberately lowers only custom tools for function-only upstreams.
type ResponsesNamespaceName = NamespacedToolName

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

// AdaptResponsesClientTools lowers Codex client-only tools in req to
// ordinary function tools. It mutates req and returns the mapping required to
// restore the upstream response.
func AdaptResponsesClientTools(req map[string]any) (ResponsesClientToolMapping, bool, error) {
	if req == nil {
		return ResponsesClientToolMapping{}, false, nil
	}
	tools, ok := req["tools"].([]any)
	if !ok || len(tools) == 0 {
		return ResponsesClientToolMapping{}, false, nil
	}

	adapter := ResponsesClientToolMapping{CustomTools: make(map[string]bool)}
	lowered := make([]any, 0, len(tools))
	changed := false
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok {
			lowered = append(lowered, raw)
			continue
		}
		typ := strings.TrimSpace(stringValue(tool["type"]))
		name := strings.TrimSpace(stringValue(tool["name"]))
		switch typ {
		case "custom":
			if name == "" {
				lowered = append(lowered, raw)
				continue
			}
			copy := copyClientTool(tool)
			copy["type"] = "function"
			copy["parameters"] = json.RawMessage(customToolInputSchema)
			delete(copy, "format")
			adapter.CustomTools[name] = true
			lowered = append(lowered, copy)
			changed = true
		default:
			lowered = append(lowered, raw)
		}
	}
	if changed {
		req["tools"] = lowered
	}
	if rewriteClientToolHistory(req["input"], &adapter) {
		changed = true
	}
	if rewriteClientToolChoice(req, &adapter) {
		changed = true
	}
	if len(adapter.CustomTools) == 0 {
		adapter.CustomTools = nil
	}
	return adapter, changed, nil
}

// AdaptResponsesClientToolsWithInheritedMapping lowers client-tool history on
// a follow-up request that omits the session-level tools declaration. An
// explicitly present tools field, including an empty or malformed value,
// always replaces the inherited mapping and is handled by the ordinary
// declaration-driven adapter.
func AdaptResponsesClientToolsWithInheritedMapping(
	req map[string]any,
	inherited ResponsesClientToolMapping,
) (ResponsesClientToolMapping, bool, error) {
	if req == nil {
		return ResponsesClientToolMapping{}, false, nil
	}
	if _, toolsPresent := req["tools"]; toolsPresent {
		return AdaptResponsesClientTools(req)
	}
	if len(inherited.CustomTools) == 0 {
		return ResponsesClientToolMapping{}, false, nil
	}

	changed := rewriteClientToolHistory(req["input"], &inherited)
	if rewriteClientToolChoice(req, &inherited) {
		changed = true
	}
	return inherited, changed, nil
}

func copyClientTool(tool map[string]any) map[string]any {
	copy := make(map[string]any, len(tool))
	for key, value := range tool {
		copy[key] = value
	}
	return copy
}

func rewriteClientToolHistory(value any, adapter *ResponsesClientToolMapping) bool {
	changed := false
	var visit func(any)
	visit = func(value any) {
		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				visit(item)
			}
		case map[string]any:
			typ := strings.TrimSpace(stringValue(typed["type"]))
			switch typ {
			case "custom_tool_call":
				if adapter.CustomTools[strings.TrimSpace(stringValue(typed["name"]))] {
					typed["type"] = "function_call"
					typed["arguments"] = customToolCallArguments(stringValue(typed["input"]))
					delete(typed, "input")
					normalizeLoweredFunctionItemID(typed)
					changed = true
				}
			case "custom_tool_call_output":
				typed["type"] = "function_call_output"
				normalizeLoweredFunctionItemID(typed)
				normalizeClientToolOutput(typed)
				changed = true
			}
			for _, child := range typed {
				visit(child)
			}
		}
	}
	visit(value)
	return changed
}

// dropInvalidLoweredFunctionItemID removes Codex client-only item IDs such as
// ctc_*, ctco_*, tsc_*, and tso_* after their item type is lowered to the
// function protocol. Function upstreams validate these IDs with the fc prefix;
// call_id, which is preserved separately, is the tool call/output pairing key.
func normalizeLoweredFunctionItemID(item map[string]any) {
	id := strings.TrimSpace(stringValue(item["id"]))
	if id == "" || strings.HasPrefix(id, "fc_") {
		return
	}
	if recovered := retypedResponsesToolCallItemID(id, "function_call"); recovered != id {
		item["id"] = recovered
		return
	}
	delete(item, "id")
}

var responsesToolCallItemIDPrefixes = []string{"fc_", "ctc_", "tsc_"}

func responsesToolCallItemIDPrefix(itemType string) string {
	switch itemType {
	case "custom_tool_call":
		return "ctc_"
	case "tool_search_call":
		return "tsc_"
	case "function_call":
		return "fc_"
	default:
		return ""
	}
}

func retypedResponsesToolCallItemID(id, itemType string) string {
	want := responsesToolCallItemIDPrefix(itemType)
	if want == "" || id == "" || strings.HasPrefix(id, want) {
		return id
	}
	for _, known := range responsesToolCallItemIDPrefixes {
		if known != want && strings.HasPrefix(id, known) {
			return want + strings.TrimPrefix(id, known)
		}
	}
	return id
}

func retypeResponsesToolCallItemID(item map[string]any, itemType string) {
	id := strings.TrimSpace(stringValue(item["id"]))
	if retyped := retypedResponsesToolCallItemID(id, itemType); retyped != id {
		item["id"] = retyped
	}
}

func normalizeClientToolOutput(item map[string]any) {
	output, exists := item["output"]
	if !exists {
		return
	}
	if _, ok := output.(string); ok {
		return
	}
	if output == nil {
		item["output"] = ""
		return
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		item["output"] = ""
		return
	}
	item["output"] = string(encoded)
}

func rewriteClientToolChoice(req map[string]any, adapter *ResponsesClientToolMapping) bool {
	choice, ok := req["tool_choice"].(map[string]any)
	if !ok {
		return false
	}
	typ := strings.TrimSpace(stringValue(choice["type"]))
	name := strings.TrimSpace(stringValue(choice["name"]))
	if typ == "custom" && adapter.CustomTools[name] {
		choice["type"] = "function"
		return true
	}
	return false
}

func customToolCallArguments(input string) string {
	encoded, _ := json.Marshal(map[string]string{"input": input})
	return string(encoded)
}

func rawObjectString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

// RestoreResponsesClientToolPayload restores client tool calls in a non-stream
// native Responses JSON payload.
func RestoreResponsesClientToolPayload(payload []byte, mapping ResponsesClientToolMapping) ([]byte, bool, error) {
	if len(payload) == 0 {
		return payload, false, nil
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return payload, false, err
	}
	changed := restoreClientToolValue(value, &mapping)
	if !changed {
		return payload, false, nil
	}
	var rebuilt bytes.Buffer
	encoder := json.NewEncoder(&rebuilt)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return payload, false, err
	}
	rebuiltPayload := bytes.TrimSuffix(rebuilt.Bytes(), []byte("\n"))
	return rebuiltPayload, true, nil
}

func restoreClientToolValue(value any, adapter *ResponsesClientToolMapping) bool {
	changed := false
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			changed = restoreClientToolValue(item, adapter) || changed
		}
	case map[string]any:
		if strings.TrimSpace(stringValue(typed["type"])) == "function_call" {
			name := strings.TrimSpace(stringValue(typed["name"]))
			if adapter.CustomTools[name] {
				typed["type"] = "custom_tool_call"
				retypeResponsesToolCallItemID(typed, "custom_tool_call")
				typed["input"] = extractCustomToolCallInput(rawObjectString(typed["arguments"]))
				delete(typed, "arguments")
				delete(typed, "namespace")
				changed = true
			}
		}
		for _, child := range typed {
			changed = restoreClientToolValue(child, adapter) || changed
		}
	}
	return changed
}

// ResponsesClientToolStreamRestorer restores client tool stream lifecycles.
// It is intentionally stateful because custom tools need their function
// arguments buffered until the upstream signals the call is complete.
type ResponsesClientToolStreamRestorer struct {
	adapter  ResponsesClientToolMapping
	nextSeq  int
	seenSeq  bool
	calls    map[string]*responsesClientToolStreamCall
	byOutput map[int]*responsesClientToolStreamCall
}

type responsesClientToolStreamCall struct {
	kind         string
	name         string
	callID       string
	itemID       string
	clientItemID string
	outputIdx    int
	arguments    strings.Builder
}

func NewResponsesClientToolStreamRestorer(mapping ResponsesClientToolMapping) *ResponsesClientToolStreamRestorer {
	return &ResponsesClientToolStreamRestorer{adapter: mapping, calls: make(map[string]*responsesClientToolStreamCall), byOutput: make(map[int]*responsesClientToolStreamCall)}
}

// Restore transforms one upstream SSE event into zero or more client events.
// Returned sequence numbers are continuous even when function argument events
// are suppressed or a custom completion expands into two events.
func (r *ResponsesClientToolStreamRestorer) Restore(event ResponsesStreamEvent) []ResponsesStreamEvent {
	if r == nil {
		return []ResponsesStreamEvent{event}
	}
	if !r.seenSeq {
		r.nextSeq = event.SequenceNumber
		r.seenSeq = true
	}
	var out []ResponsesStreamEvent
	emit := func(event ResponsesStreamEvent) {
		event.SequenceNumber = r.nextSeq
		r.nextSeq++
		out = append(out, event)
	}

	switch event.Type {
	case "response.output_item.added":
		if call := r.recordItem(event); call != nil {
			if call.kind == "custom" {
				event.Item.Type = "custom_tool_call"
				event.Item.Input = ""
				event.Item.Arguments = ""
				event.Item.Namespace = ""
				if call.clientItemID != "" {
					event.Item.ID = call.clientItemID
				}
			} else {
				event.Item.Type = "tool_search_call"
				event.Item.Name = ""
				event.Item.Arguments = "{}"
				event.Item.Namespace = ""
			}
		}
		emit(event)
	case "response.function_call_arguments.delta":
		if call := r.callFor(event); call != nil {
			_, _ = call.arguments.WriteString(event.Delta)
			return nil
		}
		emit(event)
	case "response.function_call_arguments.done":
		if call := r.callFor(event); call != nil {
			if event.Arguments != "" {
				call.arguments.Reset()
				_, _ = call.arguments.WriteString(event.Arguments)
			}
			if call.kind == "custom" {
				input := extractCustomToolCallInput(call.arguments.String())
				if input != "" {
					emit(ResponsesStreamEvent{Type: "response.custom_tool_call_input.delta", OutputIndex: call.outputIdx, ItemID: call.clientItemID, Delta: input})
				}
				emit(ResponsesStreamEvent{Type: "response.custom_tool_call_input.done", OutputIndex: call.outputIdx, ItemID: call.clientItemID, CallID: call.callID, Name: call.name, Input: input})
			}
			return out
		}
		emit(event)
	case "response.output_item.done":
		if call := r.recordItem(event); call != nil {
			if call.kind == "custom" {
				event.Item.Type = "custom_tool_call"
				event.Item.Input = extractCustomToolCallInput(call.arguments.String())
				event.Item.Arguments = ""
				event.Item.Namespace = ""
			} else {
				event.Item.Type = "tool_search_call"
				event.Item.Name = ""
				event.Item.Arguments = call.arguments.String()
				if strings.TrimSpace(event.Item.Arguments) == "" {
					event.Item.Arguments = "{}"
				}
				event.Item.Namespace = ""
				if call.clientItemID != "" {
					event.Item.ID = call.clientItemID
				}
			}
			delete(r.calls, call.itemID)
			delete(r.calls, call.callID)
			delete(r.byOutput, call.outputIdx)
		}
		emit(event)
	default:
		// response.completed carries the non-stream representation.
		if event.Response != nil {
			restoreResponsesOutputClientTools(event.Response.Output, &r.adapter)
		}
		emit(event)
	}
	return out
}

// RestoreEvent restores one Responses SSE JSON data payload. Custom tool
// completions can expand to multiple payloads and proxy argument deltas can be
// intentionally dropped, hence the slice return value.
func (r *ResponsesClientToolStreamRestorer) RestoreEvent(payload []byte) ([][]byte, bool, error) {
	if len(payload) == 0 {
		return nil, false, nil
	}
	var wire struct {
		Type     string `json:"type"`
		Sequence int    `json:"sequence_number"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		return nil, false, err
	}
	if isResponsesClientToolTerminalEvent(wire.Type) {
		restored, changed, err := RestoreResponsesClientToolPayload(payload, r.adapter)
		if err != nil {
			return nil, false, err
		}
		return r.resequenceRaw(restored, wire.Sequence, changed)
	}
	if !clientToolLifecycleEvent(wire.Type) {
		return r.resequenceRaw(payload, wire.Sequence, false)
	}
	if !r.clientToolEventPayload(payload) {
		return r.resequenceRaw(payload, wire.Sequence, false)
	}
	var event ResponsesStreamEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, false, err
	}
	events := r.Restore(event)
	if len(events) == 1 {
		unchanged, err := json.Marshal(events[0])
		if err == nil && bytes.Equal(bytes.TrimSpace(unchanged), bytes.TrimSpace(payload)) {
			return [][]byte{payload}, false, nil
		}
	}
	result := make([][]byte, 0, len(events))
	for _, restored := range events {
		encoded, err := json.Marshal(restored)
		if err != nil {
			return nil, false, err
		}
		result = append(result, encoded)
	}
	return result, true, nil
}

func isResponsesClientToolTerminalEvent(typ string) bool {
	switch strings.TrimSpace(typ) {
	case "response.completed", "response.done", "response.incomplete", "response.failed", "response.cancelled", "response.canceled":
		return true
	default:
		return false
	}
}

func (r *ResponsesClientToolStreamRestorer) clientToolEventPayload(payload []byte) bool {
	var raw struct {
		ItemID      string `json:"item_id"`
		CallID      string `json:"call_id"`
		Name        string `json:"name"`
		OutputIndex int    `json:"output_index"`
		Item        *struct {
			Type   string `json:"type"`
			ID     string `json:"id"`
			CallID string `json:"call_id"`
			Name   string `json:"name"`
		} `json:"item"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false
	}
	if raw.Item != nil {
		if raw.Item.Type != "function_call" {
			return false
		}
		return r.adapter.CustomTools[raw.Item.Name] || r.calls[raw.Item.ID] != nil || r.calls[raw.Item.CallID] != nil
	}
	if r.calls[raw.ItemID] != nil || r.calls[raw.CallID] != nil || r.byOutput[raw.OutputIndex] != nil {
		return true
	}
	return false
}

func clientToolLifecycleEvent(typ string) bool {
	switch typ {
	case "response.output_item.added", "response.output_item.done", "response.function_call_arguments.delta", "response.function_call_arguments.done":
		return true
	default:
		return false
	}
}

// resequenceRaw deliberately keeps opaque upstream event fields untouched.
func (r *ResponsesClientToolStreamRestorer) resequenceRaw(payload []byte, sequence int, changed bool) ([][]byte, bool, error) {
	if !r.seenSeq {
		r.nextSeq, r.seenSeq = sequence, true
	}
	if r.nextSeq == sequence && !changed {
		r.nextSeq++
		return [][]byte{payload}, false, nil
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, false, err
	}
	raw["sequence_number"] = r.nextSeq
	r.nextSeq++
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, false, err
	}
	return [][]byte{encoded}, true, nil
}

func (r *ResponsesClientToolStreamRestorer) recordItem(event ResponsesStreamEvent) *responsesClientToolStreamCall {
	if event.Item == nil || event.Item.Type != "function_call" {
		return nil
	}
	name := event.Item.Name
	kind := ""
	if r.adapter.CustomTools[name] {
		kind = "custom"
	}
	if kind == "" {
		return nil
	}
	key := event.Item.ID
	if key == "" {
		key = event.Item.CallID
	}
	call := r.calls[key]
	if call == nil {
		call = &responsesClientToolStreamCall{
			kind:         kind,
			name:         name,
			callID:       event.Item.CallID,
			itemID:       event.Item.ID,
			clientItemID: retypedResponsesToolCallItemID(event.Item.ID, "custom_tool_call"),
			outputIdx:    event.OutputIndex,
		}
		r.calls[key] = call
		if call.callID != "" {
			r.calls[call.callID] = call
		}
		r.byOutput[call.outputIdx] = call
	}
	if event.Item.Arguments != "" {
		call.arguments.Reset()
		_, _ = call.arguments.WriteString(event.Item.Arguments)
	}
	return call
}

func (r *ResponsesClientToolStreamRestorer) callFor(event ResponsesStreamEvent) *responsesClientToolStreamCall {
	if call := r.calls[event.ItemID]; call != nil {
		return call
	}
	if call := r.byOutput[event.OutputIndex]; call != nil {
		return call
	}
	for _, call := range r.calls {
		if (event.CallID != "" && call.callID == event.CallID) || (event.ItemID == "" && event.Name != "" && call.name == event.Name) {
			return call
		}
	}
	return nil
}

func restoreResponsesOutputClientTools(outputs []ResponsesOutput, adapter *ResponsesClientToolMapping) {
	for index := range outputs {
		output := &outputs[index]
		if output.Type != "function_call" {
			continue
		}
		if adapter.CustomTools[output.Name] {
			output.Type = "custom_tool_call"
			output.ID = retypedResponsesToolCallItemID(output.ID, output.Type)
			output.Input = extractCustomToolCallInput(output.Arguments)
			output.Arguments = ""
			output.Namespace = ""
		}
	}
}
