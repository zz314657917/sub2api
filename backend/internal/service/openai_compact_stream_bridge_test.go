package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func newCompactBridgeTestContext(t *testing.T, clientStream bool) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses/compact", nil)
	if clientStream {
		MarkOpenAICompactClientStream(c)
	}
	return c, rec
}

func parseCompactBridgeEvents(t *testing.T, body string) [][2]string {
	t.Helper()
	var events [][2]string
	for _, block := range strings.Split(strings.TrimSpace(body), "\n\n") {
		if strings.HasPrefix(strings.TrimSpace(block), ":") {
			continue
		}
		var eventType, data string
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "event: "):
				eventType = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				data = strings.TrimPrefix(line, "data: ")
			}
		}
		if eventType != "" {
			events = append(events, [2]string{eventType, data})
		}
	}
	return events
}

func TestRemoteCompactOutputPreservesRawDoneItem(t *testing.T) {
	body := strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"ignored"}`,
		`data: {"type":"response.output_item.done","item":{"id":"cmp_1","type":"compaction","encrypted_content":"opaque","summary":[{"text":"raw"}],"future":{"x":1}}}`,
		`data: {"type":"response.completed","response":{"id":"resp_1","output":[]}}`,
	}, "\n")

	output, ok := reconstructResponseOutputFromSSE(body)
	require.True(t, ok)
	items := gjson.ParseBytes(output).Array()
	require.Len(t, items, 1)
	require.Equal(t, "compaction", items[0].Get("type").String())
	require.Equal(t, "opaque", items[0].Get("encrypted_content").String())
	require.Equal(t, "raw", items[0].Get("summary.0.text").String())
	require.Equal(t, int64(1), items[0].Get("future.x").Int())
}

func TestRemoteCompactOutputCollectsAddedFallbackWithOtherDoneItem(t *testing.T) {
	body := strings.Join([]string{
		`data: {"type":"response.output_item.added","item":{"id":"cmp_2","type":"compaction_summary","encrypted_content":"added"}}`,
		`data: {"type":"response.output_item.done","item":{"id":"msg_1","type":"message","content":[{"type":"output_text","text":"hi"}]}}`,
	}, "\n")

	output, ok := reconstructResponseOutputFromSSE(body)
	require.True(t, ok)
	items := gjson.ParseBytes(output).Array()
	require.Len(t, items, 2)
	require.Equal(t, "message", items[0].Get("type").String())
	require.Equal(t, "compaction_summary", items[1].Get("type").String())
}

func TestRemoteCompactOutputSupplementsNonEmptyTerminalOutput(t *testing.T) {
	c, _ := newCompactBridgeTestContext(t, false)
	finalResponse := []byte(`{"id":"resp_2","output":[{"id":"msg_2","type":"message"}]}`)
	body := `data: {"type":"response.output_item.done","item":{"id":"cmp_3","type":"compaction","encrypted_content":"supplement"}}`

	patched := supplementCompactionItemFromSSE(c, finalResponse, body)
	items := gjson.GetBytes(patched, "output").Array()
	require.Len(t, items, 2)
	require.Equal(t, "compaction", items[1].Get("type").String())
	require.Equal(t, "supplement", items[1].Get("encrypted_content").String())
}

func TestRemoteCompactSSEBridgeWritesDoneAndCompleted(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	response := []byte(`{
		"id":"resp_bridge",
		"output":[{"id":"cmp_bridge","type":"compaction","encrypted_content":"payload"}],
		"usage":{"input_tokens":7,"output_tokens":2,"total_tokens":9}
	}`)

	require.True(t, writeOpenAICompactSSEBridge(c, http.StatusOK, response))
	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	events := parseCompactBridgeEvents(t, rec.Body.String())
	require.Len(t, events, 2)
	require.Equal(t, "response.output_item.done", events[0][0])
	require.Equal(t, "compaction", gjson.Get(events[0][1], "item.type").String())
	require.Equal(t, "response.completed", events[1][0])
	require.Equal(t, "resp_bridge", gjson.Get(events[1][1], "response.id").String())
}

func TestRemoteCompactSSEToJSONRawItemBridgesToClientStream(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, true)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.output_item.done","item":{"id":"cmp_e2e","type":"compaction","encrypted_content":"raw-e2e"}}`,
		``,
		`data: {"type":"response.completed","response":{"id":"resp_e2e","output":[],"usage":{"input_tokens":3,"output_tokens":1,"total_tokens":4}}}`,
		``,
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}

	svc := &OpenAIGatewayService{}
	result, err := svc.handleNonStreamingResponse(context.Background(), resp, c, &Account{ID: 1, Type: AccountTypeOAuth}, "gpt-5.5", "gpt-5.5")
	require.NoError(t, err)
	require.NotNil(t, result)
	events := parseCompactBridgeEvents(t, rec.Body.String())
	require.Len(t, events, 2)
	require.Equal(t, "raw-e2e", gjson.Get(events[0][1], "item.encrypted_content").String())
}

func TestRemoteCompactPathBasedSSEToJSONRemainsJSON(t *testing.T) {
	c, rec := newCompactBridgeTestContext(t, false)
	upstreamSSE := strings.Join([]string{
		`data: {"type":"response.output_item.done","item":{"id":"cmp_path","type":"compaction","encrypted_content":"raw-path"}}`,
		``,
		`data: {"type":"response.completed","response":{"id":"resp_path","output":[],"usage":{"input_tokens":2,"output_tokens":1,"total_tokens":3}}}`,
		``,
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}

	svc := &OpenAIGatewayService{}
	_, err := svc.handleNonStreamingResponse(context.Background(), resp, c, &Account{ID: 1, Type: AccountTypeOAuth}, "gpt-5.5", "gpt-5.5")
	require.NoError(t, err)
	require.Equal(t, "resp_path", gjson.Get(rec.Body.String(), "id").String())
	require.Equal(t, "compaction", gjson.Get(rec.Body.String(), "output.0.type").String())
	require.NotContains(t, rec.Body.String(), "event:")
}

func TestRemoteCompactAPIKeyPassthroughForcesJSONAccept(t *testing.T) {
	c, _ := newCompactBridgeTestContext(t, false)
	c.Request.Header.Set("Accept", "text/event-stream")
	svc := &OpenAIGatewayService{}
	req, err := svc.buildUpstreamRequestOpenAIPassthrough(context.Background(), c, &Account{Type: AccountTypeAPIKey}, []byte(`{"model":"gpt-5.5"}`), "token")
	require.NoError(t, err)
	require.Equal(t, "application/json", req.Header.Get("Accept"))
}
