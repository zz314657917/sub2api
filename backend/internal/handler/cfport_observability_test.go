package handler

import (
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestCFPortRequestBodyReadErrorKinds(t *testing.T) {
	if got := requestBodyReadErrorKind(io.ErrUnexpectedEOF); got != "truncated_body" {
		t.Fatalf("unexpected EOF kind = %q", got)
	}
	if got := requestContentEncodingCategory("Br"); got != "other" {
		t.Fatalf("unknown encoding = %q", got)
	}
	maxErr := &http.MaxBytesError{Limit: 7}
	if got := requestBodyReadErrorKind(errors.Join(maxErr, io.ErrUnexpectedEOF)); got != "max_bytes" {
		t.Fatalf("max bytes precedence = %q", got)
	}
}

func TestCFPortCaptureWriterCapturesOnlyTerminalSSE(t *testing.T) {
	w := &opsCaptureWriter{limit: 1024}
	w.captureResponseChunk([]byte("data: ordinary text\n\n"), http.StatusOK)
	if w.buf.Len() != 0 {
		t.Fatalf("captured ordinary success output: %q", w.buf.String())
	}
	w.captureResponseChunk([]byte("event: response."), http.StatusOK)
	w.captureResponseChunk([]byte("failed\ndata: {\"type\":\"response.failed\",\"error\":{\"code\":\"rate_limit_exceeded\",\"message\":\"later\"}}\n\n"), http.StatusOK)
	parsed := parseOpsErrorResponse(w.buf.Bytes())
	if !parsed.StreamFailure || parsed.ErrorType != "rate_limit_error" || parsed.Message != "later" {
		t.Fatalf("unexpected parsed terminal SSE: %+v body=%q", parsed, w.buf.String())
	}
}

func TestCFPortSSEMarkerDoesNotMatchData(t *testing.T) {
	if got := findOpsTerminalSSEStart([]byte("data: prose event: error is not a frame\n\n")); got >= 0 {
		t.Fatalf("false marker at %d", got)
	}
	if _, ok := parseOpsSSEFailure([]byte("data: {\"note\":\"event: response.failed\"}\n\n")); ok {
		t.Fatal("ordinary data was classified as terminal failure")
	}
}

func TestCFPortStreamFailureStatus(t *testing.T) {
	if got := inferStreamFailureStatus(nil, parsedOpsError{ErrorType: "rate_limit_error", StreamFailure: true}); got != http.StatusTooManyRequests {
		t.Fatalf("rate limit status = %d", got)
	}
	if got := inferStreamFailureStatus(nil, parsedOpsError{Code: "authentication_failed", StreamFailure: true}); got != http.StatusUnauthorized {
		t.Fatalf("auth status = %d", got)
	}
}
