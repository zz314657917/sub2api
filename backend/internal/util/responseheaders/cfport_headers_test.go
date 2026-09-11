package responseheaders

import (
	"net/http"
	"testing"
)

func TestCFPortReasoningIncludedHeaderIsForwarded(t *testing.T) {
	src := http.Header{"X-Reasoning-Included": []string{"true"}, "X-Private": []string{"no"}}
	got := FilterHeaders(src, nil)
	if got.Get("X-Reasoning-Included") != "true" {
		t.Fatalf("reasoning marker dropped: %#v", got)
	}
	if got.Get("X-Private") != "" {
		t.Fatalf("private header unexpectedly forwarded: %#v", got)
	}
}
