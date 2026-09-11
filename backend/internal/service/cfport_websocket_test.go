package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func TestCFPortWebSocketReservedToolAliasRoundTripAcrossTurns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	setCodexToolNameReverse(c, nil)

	first, reverse, changed, err := aliasOpenAIOAuthReservedToolNamesBody([]byte(`{"type":"response.create","tools":[{"type":"function","name":"python"}]}`))
	if err != nil || !changed {
		t.Fatalf("first-turn alias failed: changed=%v err=%v", changed, err)
	}
	if err := mergeCodexToolNameReverse(c, reverse); err != nil {
		t.Fatalf("register first-turn reverse alias: %v", err)
	}
	if !strings.Contains(string(first), codexPythonToolAlias) {
		t.Fatalf("upstream first turn lost alias: %s", first)
	}

	// A model may emit the first call only after a later response.create. The
	// retained reverse map must still restore the client-visible reserved name.
	_, laterReverse, laterChanged, err := aliasOpenAIOAuthReservedToolNamesBody([]byte(`{"type":"response.create","input":[{"type":"message","content":"next"}]}`))
	if err != nil || laterChanged {
		t.Fatalf("unexpected second-turn alias result: changed=%v err=%v", laterChanged, err)
	}
	if err := mergeCodexToolNameReverse(c, laterReverse); err != nil {
		t.Fatalf("merge no-tools second turn: %v", err)
	}
	restored := restoreCodexToolNamesFromContext(c, []byte(`{"type":"response.output_item.done","item":{"type":"function_call","name":"python__sub2api","call_id":"fc_1"}}`))
	if strings.Contains(string(restored), codexPythonToolAlias) || !strings.Contains(string(restored), `"name":"python"`) {
		t.Fatalf("downstream alias was not restored: %s", restored)
	}
}

func TestCFPortWebSocketFrameAdapterRestoresModelAndToolAcrossTurns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serverErr := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := coderws.Accept(w, r, nil)
		if err != nil {
			serverErr <- err
			return
		}
		defer func() { _ = conn.CloseNow() }()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		setCodexToolNameReverse(c, map[string]string{codexPythonToolAlias: codexReservedPythonToolName})
		frameConn := &openAIWSClientFrameConn{
			conn: conn,
			restoreResponseModel: func(payload []byte) []byte {
				return replaceOpenAIWSMessageModel(payload, "mapped-model", "client-model")
			},
			restoreToolNames: func(payload []byte) []byte {
				return restoreCodexToolNamesFromContext(c, payload)
			},
		}
		for _, payload := range [][]byte{
			[]byte(`{"type":"response.output_item.done","response":{"model":"mapped-model"},"item":{"type":"function_call","name":"python__sub2api","call_id":"fc_1","arguments":"python__sub2api"},"sequence_number":9007199254740993}`),
			[]byte(`{"type":"response.completed","response":{"model":"mapped-model","output":[{"type":"function_call","name":"python__sub2api","call_id":"fc_2"}]},"content":"python__sub2api","sequence_number":9007199254740993}`),
		} {
			writeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := frameConn.WriteFrame(writeCtx, coderws.MessageText, payload)
			cancel()
			if err != nil {
				serverErr <- err
				return
			}
		}
		serverErr <- nil
	}))
	defer server.Close()

	dialCtx, cancelDial := context.WithTimeout(context.Background(), time.Second)
	clientConn, _, err := coderws.Dial(dialCtx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	cancelDial()
	if err != nil {
		t.Fatalf("dial frame adapter server: %v", err)
	}
	defer func() { _ = clientConn.CloseNow() }()
	for turn := 1; turn <= 2; turn++ {
		readCtx, cancelRead := context.WithTimeout(context.Background(), time.Second)
		msgType, payload, readErr := clientConn.Read(readCtx)
		cancelRead()
		if readErr != nil {
			t.Fatalf("read restored turn %d: %v", turn, readErr)
		}
		if msgType != coderws.MessageText || strings.Contains(string(payload), "mapped-model") {
			t.Fatalf("turn %d did not restore frame fields: %s", turn, payload)
		}
		if !strings.Contains(string(payload), "client-model") || !strings.Contains(string(payload), `"name":"python"`) || !strings.Contains(string(payload), "9007199254740993") {
			t.Fatalf("turn %d lost restored/large-number fields: %s", turn, payload)
		}
		if turn == 1 && !strings.Contains(string(payload), `"arguments":"python__sub2api"`) {
			t.Fatalf("turn 1 rewrote tool argument content: %s", payload)
		}
		if turn == 2 && !strings.Contains(string(payload), `"content":"python__sub2api"`) {
			t.Fatalf("turn 2 rewrote ordinary content: %s", payload)
		}
	}
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("frame adapter server: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("frame adapter server did not finish")
	}
}

func TestCFPortWebSocketReservedToolAliasRejectsCaseCollision(t *testing.T) {
	_, _, changed, err := aliasOpenAIOAuthReservedToolNamesBody([]byte(`{"tools":[{"type":"function","name":"python"},{"type":"function","name":"Python"}]}`))
	if err == nil || changed {
		t.Fatalf("case-colliding aliases must be rejected: changed=%v err=%v", changed, err)
	}
}

func TestCFPortWebSocketReservedToolOwnersRejectCrossTurnCollisions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	newContext := func() *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		setCodexToolNameReverse(c, nil)
		c.Set(codexWSToolNameOwnersKey, nil)
		return c
	}
	register := func(t *testing.T, c *gin.Context, body string) {
		t.Helper()
		if err := registerCodexWSToolNameOwners(c, []byte(body)); err != nil {
			t.Fatalf("register %s: %v", body, err)
		}
	}

	c := newContext()
	register(t, c, `{"type":"response.create","tools":[{"type":"function","name":"python"}]}`)
	aliased, reverse, changed, err := aliasOpenAIOAuthReservedToolNamesBody([]byte(`{"type":"response.create","tools":[{"type":"function","name":"python"}]}`))
	if err != nil || !changed {
		t.Fatalf("alias initial python: changed=%v err=%v", changed, err)
	}
	if err := mergeCodexToolNameReverse(c, reverse); err != nil {
		t.Fatalf("merge initial python: %v", err)
	}
	if err := registerCodexWSToolNameOwners(c, []byte(`{"type":"response.create","tools":[{"type":"function","name":"python__sub2api"}]}`)); err == nil {
		t.Fatal("first python then native alias must be rejected")
	}
	if err := registerCodexWSToolNameOwners(c, []byte(`{"type":"response.create","tools":[{"type":"function","name":"Python"}]}`)); err == nil {
		t.Fatal("first python then case-variant Python must be rejected")
	}
	if err := registerCodexWSToolNameOwners(c, []byte(`{"type":"response.create","input":[{"type":"message","content":"no tools"}]}`)); err != nil {
		t.Fatalf("no-tools follow-up must retain existing binding: %v", err)
	}
	restored := restoreCodexToolNamesFromContext(c, []byte(`{"type":"response.output_item.done","item":{"type":"function_call","name":"python__sub2api","call_id":"fc_late"}}`))
	if !strings.Contains(string(restored), `"name":"python"`) {
		t.Fatalf("no-tools follow-up lost earlier reverse binding: %s", restored)
	}
	if !strings.Contains(string(aliased), codexPythonToolAlias) {
		t.Fatalf("initial upstream payload did not alias python: %s", aliased)
	}

	c = newContext()
	register(t, c, `{"type":"response.create","tools":[{"type":"function","name":"python__sub2api"}]}`)
	if err := registerCodexWSToolNameOwners(c, []byte(`{"type":"response.create","tools":[{"type":"function","name":"python"}]}`)); err == nil {
		t.Fatal("first native alias then python must be rejected")
	}
}

func TestCFPortWebSocketBufferedAliasRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	setCodexToolNameReverse(c, map[string]string{codexPythonToolAlias: codexReservedPythonToolName})
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_cfport\",\"output\":[{\"type\":\"function_call\",\"name\":\"python__sub2api\",\"call_id\":\"fc_1\"}]}}\n\n"))}
	finalResponse, _, _, err := (&OpenAIGatewayService{}).readOpenAICompatBufferedTerminal(resp, c, "cfport buffered alias", "req_cfport")
	if err != nil || finalResponse == nil {
		t.Fatalf("buffered terminal failed: response=%v err=%v", finalResponse, err)
	}
	encoded, marshalErr := json.Marshal(finalResponse)
	if marshalErr != nil {
		t.Fatalf("marshal buffered terminal response: %v", marshalErr)
	}
	if strings.Contains(string(encoded), codexPythonToolAlias) || !strings.Contains(string(encoded), `"name":"python"`) {
		t.Fatalf("buffered terminal leaked alias: %s", encoded)
	}
}

func TestCFPortWebSocketPassthroughTurnModelRestoration(t *testing.T) {
	meta := newOpenAIWSPassthroughUsageMeta("client-model", []byte(`{"type":"response.create","model":"client-model"}`))
	meta.initFromFirstFrame([]byte(`{"type":"response.create","model":"upstream-model"}`), "upstream-model")
	requestModel, upstreamModel := meta.turnModels("")
	if requestModel != "client-model" || upstreamModel != "upstream-model" {
		t.Fatalf("first turn models = %q, %q", requestModel, upstreamModel)
	}
	meta.updateFromResponseCreate([]byte(`{"type":"response.create","model":"mapped-second"}`), "mapped-second", "client-second")
	requestModel, upstreamModel = meta.turnModels("")
	if requestModel != "client-second" || upstreamModel != "mapped-second" {
		t.Fatalf("later turn models = %q, %q", requestModel, upstreamModel)
	}
	restored := replaceOpenAIWSMessageModel([]byte(`{"type":"response.completed","response":{"model":"mapped-second"}}`), upstreamModel, requestModel)
	if !strings.Contains(string(restored), "client-second") || strings.Contains(string(restored), "mapped-second") {
		t.Fatalf("passthrough model was not restored: %s", restored)
	}
}
