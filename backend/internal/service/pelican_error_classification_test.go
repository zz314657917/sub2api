package service

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestClassifyPelicanFailure(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"timeout", context.DeadlineExceeded, "upstream_timeout"},
		{"stream", io.ErrUnexpectedEOF, "upstream_stream_closed"},
		{"request", ErrPelicanUpstreamRequest, "upstream_request_failed"},
		{"html", ErrPelicanIncompleteHTML, "incomplete_html"},
		{"other", errors.New("provider rejected request"), "generation_failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, safe := classifyPelicanFailure(tc.err, context.Background())
			if code != tc.want || safe == "" {
				t.Fatalf("code=%q safe=%q, want %q and non-empty safe message", code, safe, tc.want)
			}
		})
	}
}

func TestClassifyPelicanFailureDoesNotExposeProviderDetails(t *testing.T) {
	code, safe := classifyPelicanFailure(errors.New("authorization Bearer secret-token https://provider.test/?key=abc"), context.Background())
	if code != "generation_failed" || safe != "生成失败" {
		t.Fatalf("code=%q safe=%q", code, safe)
	}
}
