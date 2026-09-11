package service

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestResponsesItemIDReplayRegression(t *testing.T) {
	for _, tc := range []struct {
		typ, id string
		strip   bool
	}{
		{"reasoning", "item_2481a3cf732826e29a4ac517", true},
		{"reasoning", "rs_valid", false},
		{" reasoning ", "item_invalid", true},
		{"message", "msg_valid", false},
		{"message", "item_invalid", true},
		{"function_call", "fc_valid", false},
		{"custom_tool_call", "ctc_valid", false},
		{"custom_tool_call", "fc_wrong", true},
		{"tool_search_call", "tsc_valid", false},
		{"tool_search_call", "item_invalid", true},
		{"custom_tool_call_output", "fc_valid", false},
		{"function_call_output", "item_output", false},
		{"future_item", "item_unknown", false},
	} {
		t.Run(tc.typ+"/"+tc.id, func(t *testing.T) {
			if got := shouldStripOpenAIResponsesInputItemID(tc.typ, tc.id); got != tc.strip {
				t.Fatalf("strip = %v, want %v", got, tc.strip)
			}
		})
	}
	input := []byte(`{"model":"gpt-test","input":[{"type":"reasoning","id":"item_2481a3cf732826e29a4ac517","encrypted_content":"opaque-content","summary":[]},{"type":"reasoning","id":"rs_valid","summary":[]},{"type":"custom_tool_call","id":"ctc_valid","call_id":"call_pair","name":"exec","input":"pwd"}]}`)
	output, changed, err := sanitizeOpenAIResponsesInputItemIDs(input)
	if err != nil || !changed {
		t.Fatalf("changed = %v, err = %v", changed, err)
	}
	if gjson.GetBytes(output, "input.0.id").Exists() {
		t.Fatal("invalid reasoning ID survived")
	}
	for path, want := range map[string]string{
		"input.0.encrypted_content": "opaque-content",
		"input.0.summary":           "[]",
		"input.1.id":                "rs_valid",
		"input.2.id":                "ctc_valid",
		"input.2.call_id":           "call_pair",
		"input.2.input":             "pwd",
	} {
		if got := gjson.GetBytes(output, path).String(); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	again, changed, err := sanitizeOpenAIResponsesInputItemIDs(output)
	if err != nil || changed || string(again) != string(output) {
		t.Fatalf("sanitization must be idempotent: changed = %v, err = %v", changed, err)
	}
}
