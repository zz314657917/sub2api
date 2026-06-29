package service

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExtractImagesUpstreamError_IncompleteIsRetryable(t *testing.T) {
	body := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_1\"}}\n\n" +
		"data: {\"type\":\"response.incomplete\",\"response\":{\"id\":\"resp_1\",\"status\":\"incomplete\",\"incomplete_details\":{\"reason\":\"max_output_tokens\"}}}\n\n"

	got := extractOpenAIImagesUpstreamError([]byte(body))
	if got == nil {
		t.Fatal("incomplete event should produce an upstream error, got nil")
	}
	if got.StatusCode != http.StatusBadGateway {
		t.Fatalf("incomplete(max_output_tokens) should be 502 retryable, got %d", got.StatusCode)
	}
	if got.Code != "response_incomplete" {
		t.Fatalf("unexpected code %q", got.Code)
	}
	if !strings.Contains(got.Message, "max_output_tokens") {
		t.Fatalf("message should carry reason, got %q", got.Message)
	}
}

func TestExtractImagesUpstreamError_IncompleteContentFilterNotRetryable(t *testing.T) {
	body := "data: {\"type\":\"response.incomplete\",\"response\":{\"id\":\"r\",\"status\":\"incomplete\",\"incomplete_details\":{\"reason\":\"content_filter\"}}}\n\n"

	got := extractOpenAIImagesUpstreamError([]byte(body))
	if got == nil {
		t.Fatal("content_filter incomplete should produce error")
	}
	if got.StatusCode != http.StatusBadRequest {
		t.Fatalf("content_filter should be 400, got %d", got.StatusCode)
	}
}

func TestExtractImagesUpstreamError_ErrorAndFailedUnchanged(t *testing.T) {
	errBody := "data: {\"type\":\"error\",\"error\":{\"type\":\"image_generation_user_error\",\"code\":\"moderation_blocked\",\"message\":\"rejected\"}}\n\n"
	if got := extractOpenAIImagesUpstreamError([]byte(errBody)); got == nil || got.StatusCode != http.StatusBadRequest {
		t.Fatalf("moderation_blocked should still be 400, got %+v", got)
	}
}

func TestSummarizeNoOutputBody_ExtractsDiagnostics(t *testing.T) {
	body := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"r\"}}\n\n" +
		"data: {\"type\":\"response.in_progress\",\"response\":{\"id\":\"r\",\"status\":\"in_progress\"}}\n\n"

	summary := summarizeOpenAIImagesNoOutputBody([]byte(body))
	if !strings.HasPrefix(summary, "no_image_output") {
		t.Fatalf("summary should start with marker, got %q", summary)
	}
	if !strings.Contains(summary, "last_event=response.in_progress") {
		t.Fatalf("summary should capture last event type, got %q", summary)
	}
	if !strings.Contains(summary, "status=in_progress") {
		t.Fatalf("summary should capture response status, got %q", summary)
	}
}

func TestSummarizeNoOutputBody_IncompleteReasonAndTruncation(t *testing.T) {
	long := strings.Repeat("x", 2000)
	body := "data: {\"type\":\"response.incomplete\",\"response\":{\"status\":\"incomplete\",\"incomplete_details\":{\"reason\":\"max_output_tokens\"},\"junk\":\"" + long + "\"}}\n\n"

	summary := summarizeOpenAIImagesNoOutputBody([]byte(body))
	if !strings.Contains(summary, "incomplete_reason=max_output_tokens") {
		t.Fatalf("should capture incomplete reason, got %q", summary[:120])
	}
	if !strings.Contains(summary, "truncated") {
		t.Fatalf("oversized body should be truncated, len=%d", len(summary))
	}
}

func TestImagesOAuthNonStreaming_CompletedNoImageTriggersSameAccountRetry(t *testing.T) {
	upstreamSSE := "event: response.created\n" +
		"data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_x\",\"status\":\"in_progress\",\"model\":\"gpt-5.4-mini-2026-03-17\",\"output\":[]}}\n\n" +
		"event: response.completed\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_x\",\"status\":\"completed\",\"model\":\"gpt-5.4-mini-2026-03-17\",\"output\":[],\"tool_usage\":{\"image_gen\":{\"output_tokens\":0}}}}\n\n"

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}

	svc := &OpenAIGatewayService{}
	_, _, _, err := svc.handleOpenAIImagesOAuthNonStreamingResponse(resp, c, "b64_json", "gpt-image-2")

	if err == nil {
		t.Fatal("completed-but-no-image should return an error")
	}
	var failoverErr *UpstreamFailoverError
	if !errors.As(err, &failoverErr) {
		t.Fatalf("expected *UpstreamFailoverError to trigger retry, got %T: %v", err, err)
	}
	if failoverErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", failoverErr.StatusCode)
	}
	if !failoverErr.RetryableOnSameAccount {
		t.Fatal("soft-failure should prefer same-account retry")
	}
}

func TestImagesOAuthNonStreaming_IncompleteTriggersFailover(t *testing.T) {
	body := "data: {\"type\":\"response.incomplete\",\"response\":{\"id\":\"resp_i\",\"status\":\"incomplete\",\"incomplete_details\":{\"reason\":\"max_output_tokens\"}}}\n\n"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"x-request-id": []string{"rid_incomplete"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	svc := &OpenAIGatewayService{}
	_, _, _, err := svc.handleOpenAIImagesOAuthNonStreamingResponse(resp, c, "b64_json", "gpt-image-2")

	var failoverErr *UpstreamFailoverError
	if !errors.As(err, &failoverErr) {
		t.Fatalf("incomplete should return *UpstreamFailoverError, got %T: %v", err, err)
	}
	if failoverErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", failoverErr.StatusCode)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("failover should not write an upstream error response before handler retry, got recorder code %d", rec.Code)
	}
}

func TestImagesOAuthNonStreaming_IncompleteContentFilterWritesClientError(t *testing.T) {
	body := "data: {\"type\":\"response.incomplete\",\"response\":{\"id\":\"resp_i\",\"status\":\"incomplete\",\"incomplete_details\":{\"reason\":\"content_filter\"}}}\n\n"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	svc := &OpenAIGatewayService{}
	_, _, _, err := svc.handleOpenAIImagesOAuthNonStreamingResponse(resp, c, "b64_json", "gpt-image-2")

	var upstreamErr *OpenAIImagesUpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("content_filter incomplete should return *OpenAIImagesUpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", upstreamErr.StatusCode)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected written 400 response, got %d", rec.Code)
	}
}

func TestExtractModelRefusal_EmptyWhenNoText(t *testing.T) {
	body := "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_no_text\",\"output\":[]}}\n\n"
	if got := extractOpenAIImagesModelRefusal([]byte(body)); got != "" {
		t.Fatalf("expected no refusal text, got %q", got)
	}
}

func TestImagesOAuthNonStreaming_ContentRefusalReturns400NoRetry(t *testing.T) {
	body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"I cannot help create that image.\"}\n\n" +
		"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_refusal\",\"status\":\"completed\",\"output\":[]}}\n\n"
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	svc := &OpenAIGatewayService{}
	_, _, _, err := svc.handleOpenAIImagesOAuthNonStreamingResponse(resp, c, "b64_json", "gpt-image-2")

	var upstreamErr *OpenAIImagesUpstreamError
	if !errors.As(err, &upstreamErr) {
		t.Fatalf("expected *OpenAIImagesUpstreamError, got %T: %v", err, err)
	}
	if upstreamErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", upstreamErr.StatusCode)
	}
	if upstreamErr.Code != "content_policy_violation" {
		t.Fatalf("unexpected code %q", upstreamErr.Code)
	}
	var failoverErr *UpstreamFailoverError
	if errors.As(err, &failoverErr) {
		t.Fatalf("content refusal must not return failover error: %#v", failoverErr)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected written 400 response, got %d", rec.Code)
	}
}
