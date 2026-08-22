package migrations

import (
	"strings"
	"testing"
)

func TestGrokImageTutorialPages(t *testing.T) {
	content, err := FS.ReadFile("231_add_grok_image_tutorial_pages.sql")
	if err != nil {
		t.Fatalf("read Grok image tutorial migration: %v", err)
	}

	sql := string(content)
	pages := []struct {
		slug     string
		required []string
	}{
		{"grok-imagine-1-5", []string{"grok-imagine-1.5-apimart", "/v1/images/generations", "\x60quality\x60、\x60background\x60、\x60resolution\x60"}},
		{"grok-imagine-1-5-edit", []string{"grok-imagine-1.5-edit-apimart", "/v1/images/edits", "最多 1 个参考图"}},
	}
	for _, page := range pages {
		start := strings.Index(sql, "('"+page.slug+"',")
		if start < 0 {
			t.Fatalf("missing tutorial slug %q", page.slug)
		}
		end := strings.Index(sql[start:], "$md$, NOW())")
		if end < 0 {
			t.Fatalf("unterminated tutorial %q", page.slug)
		}
		block := sql[start : start+end]
		for _, required := range append(page.required, "400 invalid_request_error", "401 authentication_error", "402 payment_required", "429 rate_limit_error", "## cURL") {
			if !strings.Contains(block, required) {
				t.Fatalf("tutorial %q missing %q", page.slug, required)
			}
		}
	}
	for _, forbidden := range []string{"https://ai.3zapi.top", "https://ai.3zapi.com", "api.apimart.ai"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("migration contains forbidden host %q", forbidden)
		}
	}
}
