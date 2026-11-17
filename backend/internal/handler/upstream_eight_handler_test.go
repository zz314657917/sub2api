package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func upstreamEightPage(t *testing.T, name, output, extra string) string {
	t.Helper()
	return `{"type":"function_call_output","namespace":"agent","name":` + testJSON(t, name) + `,"output":` + testJSON(t, output) + extra + `}`
}

func TestUpstreamEightPageSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name, output, extra string
	}{
		{"codex_apps_open_page", `{"page_id":"p1","text":"中文 😀\n<script>untrusted</script>"}`, ""},
		{"codex_apps_open_page", `{"page_id":null}`, `,"call_id":""`},
		{"codex_apps_selected_page_part_1_of_1", "Companion Page snapshot part 1 of 1:\r\n{\"page_id\":\"p1\"}", ""},
		{"codex_apps_selected_page_part_1_of_2", "Companion Page snapshot part 1 of 2:\n{\"page_id\":\"p", ""},
		{"codex_apps_selected_page_part_2_of_2", "Companion Page snapshot part 2 of 2:\n1\"}", ""},
	} {
		t.Run(tc.name+"/"+tc.extra, func(t *testing.T) {
			body := []byte(`{"model":"gpt-5","previous_response_id":"resp_old","metadata":{"n":9007199254740993},"input":[{"role":"user","content":"before"},` +
				upstreamEightPage(t, tc.name, tc.output, tc.extra) + `,{"role":"assistant","content":"after"}]}`)
			got, changed := normalizeCodexPageContext(body)
			require.True(t, changed)
			require.Equal(t, "user", gjson.GetBytes(got, "input.1.role").String())
			require.Equal(t, "<external_"+tc.name+">\n"+tc.output+"\n</external_"+tc.name+">", gjson.GetBytes(got, "input.1.content.0.text").String())
			require.Equal(t, "before", gjson.GetBytes(got, "input.0.content").String())
			require.Equal(t, "after", gjson.GetBytes(got, "input.2.content").String())
			require.Equal(t, "9007199254740993", gjson.GetBytes(got, "metadata.n").Raw)
			require.Equal(t, "resp_old", gjson.GetBytes(got, "previous_response_id").String())
			again, changed := normalizeCodexPageContext(got)
			require.False(t, changed)
			require.Equal(t, got, again)
		})
	}
}

func TestUpstreamEightPageRejectsAmbiguity(t *testing.T) {
	valid := upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p"}`, "")
	for i, item := range []string{
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p"}`, `,"call_id":null`),
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p"}`, `,"call_id":12`),
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p"}`, `,"call_id":"call_1"`),
		strings.Replace(valid, `"namespace":"agent"`, `"namespace":"other"`, 1),
		strings.Replace(valid, `"namespace":"agent"`, `"namespace":"other","namespace":"agent"`, 1),
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p","page_id":"q"}`, ""),
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":12}`, ""),
		upstreamEightPage(t, "codex_apps_open_page", `{"page_id":""}`, ""),
		upstreamEightPage(t, "codex_apps_open_page", `[]`, ""),
		upstreamEightPage(t, "codex_apps_open_page", `null`, ""),
		upstreamEightPage(t, "codex_apps_unknown", `{"page_id":"p"}`, ""),
		upstreamEightPage(t, "codex_apps_selected_page_part_2_of_1", "Companion Page snapshot part 2 of 1:\n{}", ""),
		upstreamEightPage(t, "codex_apps_selected_page_part_1_of_2", "Companion Page snapshot part 2 of 2:\n{}", ""),
		upstreamEightPage(t, "codex_apps_selected_page_part_1_of_2", "Companion Page snapshot part 1 of 2:\n ", ""),
		upstreamEightPage(t, "codex_apps_selected_page_part_1_of_1", "Companion Page snapshot part 1 of 1:\n{", ""),
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			body := []byte(`{"model":"gpt-5","input":[` + item + `]}`)
			got, changed := normalizeCodexPageContext(body)
			require.False(t, changed)
			require.Equal(t, body, got)
		})
	}
	body := []byte(`{"input":[],"input":[` + valid + `]}`)
	got, changed := normalizeCodexPageContext(body)
	require.False(t, changed)
	require.Equal(t, body, got)
}

func TestUpstreamEightPageMixedBootstrap(t *testing.T) {
	automation := "Automation: review\nAutomation ID: wiki\nAutomation memory: $CODEX_HOME/automations/wiki/memory.md\nLast run: never\n\nReview."
	body := []byte(`{"model":"gpt-5","input":[` + upstreamEightPage(t, "codex_apps_open_page", `{"page_id":"p"}`, "") +
		`,{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, automation) +
		`},{"type":"function_call_output","namespace":"codex_app","name":"create_thread","output":` + testJSON(t, testCodexDelegationEnvelope) + `}]}`)
	got, changed := normalizeCodexPageContext(body)
	require.True(t, changed)
	got, changed = normalizeCodexAutomationBootstrap(got)
	require.True(t, changed)
	got, changed = normalizeCodexDelegationBootstrap(got)
	require.True(t, changed)
	require.Equal(t, automation, gjson.GetBytes(got, "input.1.content.0.text").String())
	require.Equal(t, testCodexDelegationEnvelope, gjson.GetBytes(got, "input.2.content.0.text").String())
	for i := 0; i < 3; i++ {
		require.Equal(t, "user", gjson.GetBytes(got, fmt.Sprintf("input.%d.role", i)).String())
	}
}

func TestUpstreamEightResponsesUTF8Ingress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, path string
		body       []byte
		message    string
	}{
		{"invalid-byte", "/v1/responses", []byte("{\"input\":\"\xff\"}"), "valid UTF-8"},
		{"invalid-compact", "/v1/responses/compact", []byte("{\"input\":\"\xc0\xaf\"}"), "valid UTF-8"},
		{"truncated-sequence", "/v1/responses", []byte("{\"input\":\"\xe4\xb8\"}"), "valid UTF-8"},
		{"unicode-control", "/v1/responses", []byte(`{"input":"中文 😀"}`), "model is required"},
		{"invalid-json-control", "/v1/responses", []byte(`{"input":`), "Failed to parse request body"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 10, User: &service.User{ID: 1}})
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})
				c.Next()
			})
			router.POST(tc.path, h.Responses)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader(tc.body))
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Equal(t, "invalid_request_error", gjson.GetBytes(response.Body.Bytes(), "error.type").String())
			require.Contains(t, gjson.GetBytes(response.Body.Bytes(), "error.message").String(), tc.message)
		})
	}
}
