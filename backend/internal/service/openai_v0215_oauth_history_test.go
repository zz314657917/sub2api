package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestV0215OAuthHistoryMapAndRaw(t *testing.T) {
	for _, tc := range []struct {
		name, body, choice string
		lite, changed      bool
		path               string
	}{
		{"ordinary-empty", `{"input":[{"type":"web_search_call"}],"tools":[]}`, "none", false, true, "tools.0"},
		{"ordinary-function", `{"input":[{"type":"web_search_call"}],"tools":[{"type":"function","name":"local"}],"tool_choice":"auto"}`, "auto", false, true, "tools.1"},
		{"preserve-required", `{"input":[{"type":"web_search_call"}],"tool_choice":"required"}`, "required", false, true, "tools.0"},
		{"preserve-object", `{"input":[{"type":"web_search_call"}],"tool_choice":{"type":"allowed_tools","mode":"auto","tools":[]}}`, "", false, true, "tools.0"},
		{"existing-preview", `{"input":[{"type":"web_search_call"}],"tools":[{"type":"web_search_preview"}],"tool_choice":"auto"}`, "auto", false, false, ""},
		{"existing-additional", `{"input":[{"type":"additional_tools","tools":[{"type":"web_search"}]},{"type":"web_search_call"}]}`, "", true, false, ""},
		{"lite-new", `{"input":[{"type":"web_search_call"},{"type":"compaction_trigger"}],"tools":[]}`, "none", true, true, "input.1.tools.0"},
		{"lite-extend", `{"input":[{"type":"additional_tools","tools":[{"type":"function","name":"local"}]},{"type":"web_search_call"}],"tool_choice":"auto"}`, "auto", true, true, "input.0.tools.1"},
		{"lite-empty-additional", `{"input":[{"type":"additional_tools"},{"type":"web_search_call"}]}`, "none", true, true, "input.0.tools.0"},
		{"empty", `{}`, "", false, false, ""},
		{"string-input", `{"input":"web_search_call"}`, "", false, false, ""},
		{"not-history", `{"input":[{"type":"message","content":"web_search_call"}]}`, "", false, false, ""},
		{"mixed-items", `{"input":[null,4,{"type":"web_search_call"}],"metadata":{"integer":9007199254740993}}`, "none", false, true, "tools.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next, changed, err := ensureOpenAIOAuthWebSearchToolForHistoryBody([]byte(tc.body), tc.lite)
			require.NoError(t, err)
			require.Equal(t, tc.changed, changed)
			if !changed {
				require.Equal(t, tc.body, string(next))
			} else {
				require.Equal(t, "web_search", gjson.GetBytes(next, tc.path+".type").String())
				require.False(t, gjson.GetBytes(next, tc.path+".external_web_access").Bool())
				require.True(t, gjson.GetBytes(next, tc.path+".external_web_access").Exists())
				if tc.choice != "" {
					require.Equal(t, tc.choice, gjson.GetBytes(next, "tool_choice").String())
				}
				if tc.name == "preserve-object" {
					require.Equal(t, "allowed_tools", gjson.GetBytes(next, "tool_choice.type").String())
				}
				if tc.name == "lite-new" {
					require.Equal(t, "compaction_trigger", gjson.GetBytes(next, "input.2.type").String())
					require.Empty(t, gjson.GetBytes(next, "tools").Array())
				}
				if tc.name == "mixed-items" {
					require.Equal(t, "9007199254740993", gjson.GetBytes(next, "metadata.integer").Raw)
				}
			}
			again, changedAgain, err := ensureOpenAIOAuthWebSearchToolForHistoryBody(next, tc.lite)
			require.NoError(t, err)
			require.False(t, changedAgain)
			require.Equal(t, next, again)
			var request map[string]any
			require.NoError(t, json.Unmarshal([]byte(tc.body), &request))
			require.Equal(t, tc.changed, ensureOpenAIOAuthWebSearchToolForHistory(request, tc.lite))
		})
	}
}

func TestV0215OAuthHistoryMalformedAndScope(t *testing.T) {
	for _, body := range []string{"", "null", "[]", `{"input":[{"type":"web_search_call"}]`, `{"input":[{"type":6}]}`} {
		next, changed, err := ensureOpenAIOAuthWebSearchToolForHistoryBody([]byte(body), false)
		require.NoError(t, err)
		require.False(t, changed)
		require.Equal(t, body, string(next))
	}
	body := []byte(`{"model":"gpt-5.4","input":[{"type":"web_search_call"}],"tools":[]}`)
	for _, typ := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		for _, compact := range []bool{false, true} {
			next, _, err := normalizeOpenAIWSOAuthHistoryCompatibilityBody(body, &Account{Platform: PlatformOpenAI, Type: typ}, false, compact)
			require.NoError(t, err)
			require.Equal(t, typ == AccountTypeOAuth && !compact, gjson.GetBytes(next, "tools.0.type").String() == "web_search")
			if typ == AccountTypeAPIKey || compact {
				baseline, _, err := normalizeOpenAIResponsesWebSocketCompatibilityBody(body, &Account{Platform: PlatformOpenAI, Type: typ})
				require.NoError(t, err)
				require.Equal(t, baseline, next, "history rewrite must not change API-key/compact wire behavior")
			}
		}
	}
}

func TestV0215OAuthHistoryTransformLiteAndCompact(t *testing.T) {
	for _, lite := range []bool{false, true} {
		for _, compact := range []bool{false, true} {
			var request map[string]any
			require.NoError(t, json.Unmarshal([]byte(`{"model":"gpt-5.4","input":[{"type":"web_search_call","id":"ws_local"}],"tools":[]}`), &request))
			result := applyCodexOAuthTransformWithOptions(request, codexOAuthTransformOptions{ResponsesLite: lite, IsCompact: compact})
			require.NoError(t, result.Error)
			out, err := json.Marshal(request)
			require.NoError(t, err)
			path := "tools.0.type"
			if lite {
				path = "input.1.tools.0.type"
			}
			require.Equal(t, !compact, gjson.GetBytes(out, path).String() == "web_search")
		}
	}
}
