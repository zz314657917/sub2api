package service

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

func newNativeAnthropicHangTestService(intervalSec int) *OpenAIGatewayService {
	return &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{
		StreamDataIntervalTimeout: intervalSec,
		MaxLineSize:               defaultMaxLineSize,
	}}}
}

func newHangingUpstreamResponse() (*http.Response, *io.PipeReader, *io.PipeWriter) {
	pr, pw := io.Pipe()
	return &http.Response{StatusCode: http.StatusOK, Body: pr, Header: make(http.Header)}, pr, pw
}

func miniAnthropicSSEStream() string {
	return strings.Join([]string{
		"event: message_start",
		`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":10,"output_tokens":1}}}`,
		"",
		"event: content_block_start",
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		"",
		"event: content_block_delta",
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		"",
		"event: content_block_stop",
		`data: {"type":"content_block_stop","index":0}`,
		"",
		"event: message_delta",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`,
		"",
		"event: message_stop",
		`data: {"type":"message_stop"}`,
		"",
		"",
	}, "\n")
}

func toolAnthropicSSEStream() string {
	return strings.Join([]string{
		"event: message_start", `data: {"type":"message_start","message":{"id":"msg_tool","type":"message","role":"assistant","content":[],"model":"glm-4.7","usage":{"input_tokens":10}}}`,
		"", "event: content_block_start", `data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"lookup","input":{}}}`,
		"", "event: content_block_delta", `data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"query\":\"status\"}"}}`,
		"", "event: content_block_stop", `data: {"type":"content_block_stop","index":0}`,
		"", "event: message_delta", `data: {"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":5}}`,
		"", "event: message_stop", `data: {"type":"message_stop"}`, "",
	}, "\n")
}

func TestAnthropicNativeLinePump_TimesOutWithoutData(t *testing.T) {
	pr, _ := io.Pipe()
	scanner := bufio.NewScanner(pr)
	defer pr.Close()
	pump := newAnthropicNativeLinePump(scanner, 50*time.Millisecond)
	defer pump.stop()
	start := time.Now()
	_, err := pump.next()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected interval timeout, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("timeout not respected: %v", elapsed)
	}
}

func TestAnthropicNativeLinePump_DataResetsTimer(t *testing.T) {
	pr, pw := io.Pipe()
	scanner := bufio.NewScanner(pr)
	pump := newAnthropicNativeLinePump(scanner, 100*time.Millisecond)
	defer pump.stop()
	defer pr.Close()
	go func() {
		_, _ = pw.Write([]byte("event: ping\n"))
		time.Sleep(300 * time.Millisecond)
		_ = pw.Close()
	}()
	line, err := pump.next()
	if err != nil || line != "event: ping" {
		t.Fatalf("expected first line, got %q err=%v", line, err)
	}
	start := time.Now()
	_, err = pump.next()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected interval timeout after data stops, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("timeout not respected: %v", elapsed)
	}
}

func TestCCStreamingFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	start := time.Now()
	res, err := svc.handleCCStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, start, true)
	_ = pw.Close()
	_ = pr.Close()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") || res == nil {
		t.Fatalf("expected timeout with result, result=%v err=%v", res, err)
	}
}

func TestCCBufferedFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	_, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	_ = pw.Close()
	_ = pr.Close()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected timeout, got %v", err)
	}
	if !strings.Contains(rec.Body.String(), "Upstream stream data interval timeout") {
		t.Fatalf("expected timeout response, got %q", rec.Body.String())
	}
}

func TestResponsesStreamingFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	res, err := svc.handleResponsesStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	_ = pw.Close()
	_ = pr.Close()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") || res == nil {
		t.Fatalf("expected timeout with result, result=%v err=%v", res, err)
	}
}

func TestResponsesStreamingFromNativeAnthropic_ClientDisconnectDrainsUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	// 首次写出即失败，模拟客户端断开；上游末尾 message_delta 才携带最终 output_tokens。
	c.Writer = &failingGinWriter{ResponseWriter: c.Writer, failAfter: 0}

	resp, pr, pw := newHangingUpstreamResponse()
	go func() {
		_, _ = pw.Write([]byte(miniAnthropicSSEStream()))
		_ = pw.Close()
	}()
	defer pr.Close()

	res, err := svc.handleResponsesStreamingFromNativeAnthropic(
		resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil {
		t.Fatalf("expected nil error after draining disconnected client, got %v", err)
	}
	if res == nil || !res.ClientDisconnect {
		t.Fatalf("expected disconnected result, got %+v", res)
	}
	if res.Usage.InputTokens != 10 || res.Usage.OutputTokens != 5 {
		t.Fatalf("expected complete usage after drain, got %+v", res.Usage)
	}
}

func TestResponsesBufferedFromNativeAnthropic_HangTimesOut(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(1)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	_, err := svc.handleResponsesBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	_ = pw.Close()
	_ = pr.Close()
	if err == nil || !strings.Contains(err.Error(), "stream data interval timeout") {
		t.Fatalf("expected timeout, got %v", err)
	}
	if !strings.Contains(rec.Body.String(), "Upstream stream data interval timeout") {
		t.Fatalf("expected timeout response, got %q", rec.Body.String())
	}
}

func TestCCStreamingFromNativeAnthropic_HappyPathStillConverts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	go func() {
		_, _ = pw.Write([]byte(miniAnthropicSSEStream()))
		_ = pw.Close()
	}()
	defer pr.Close()
	res, err := svc.handleCCStreamingFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now(), true)
	if err != nil || res == nil {
		t.Fatalf("unexpected result=%v err=%v", res, err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Hello") || !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("expected converted stream, got %q", body)
	}
}

func TestCCBufferedFromNativeAnthropic_HappyPathStillConverts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp, pr, pw := newHangingUpstreamResponse()
	go func() {
		_, _ = pw.Write([]byte(miniAnthropicSSEStream()))
		_ = pw.Close()
	}()
	defer pr.Close()
	res, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil || res == nil {
		t.Fatalf("unexpected result=%v err=%v", res, err)
	}
	if !strings.Contains(rec.Body.String(), "Hello") || res.Usage.InputTokens != 10 || res.Usage.OutputTokens != 5 {
		t.Fatalf("expected converted response and usage, body=%q usage=%+v", rec.Body.String(), res.Usage)
	}
}

func TestCCBufferedFromNativeAnthropic_ToolArgumentsAreValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(toolAnthropicSSEStream())), Header: http.Header{}}
	_, err := svc.handleCCBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					Function struct {
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Choices) != 1 || len(body.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("body=%s", rec.Body.String())
	}
	args := body.Choices[0].Message.ToolCalls[0].Function.Arguments
	if !json.Valid([]byte(args)) || args != `{"query":"status"}` {
		t.Fatalf("arguments=%q", args)
	}
}

func TestResponsesBufferedFromNativeAnthropic_ToolArgumentsAreValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newNativeAnthropicHangTestService(5)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(toolAnthropicSSEStream())), Header: http.Header{}}
	_, err := svc.handleResponsesBufferedFromNativeAnthropic(resp, c, "glm-4.7", "glm-4.7", "glm-4.7", nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Output []struct {
			Type      string `json:"type"`
			Arguments string `json:"arguments"`
		} `json:"output"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Output) != 1 || body.Output[0].Type != "function_call" {
		t.Fatalf("body=%s", rec.Body.String())
	}
	if args := body.Output[0].Arguments; !json.Valid([]byte(args)) || args != `{"query":"status"}` {
		t.Fatalf("arguments=%q", args)
	}
}
