package migrations

import (
	"strings"
	"testing"
)

func TestImageModelTutorialCurlFormatting(t *testing.T) {
	content, err := FS.ReadFile("229_format_image_tutorial_curl_examples.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial cURL migration: %v", err)
	}

	sql := string(content)
	pages := []struct {
		slug     string
		model    string
		endpoint string
	}{
		{"gpt-image-2", "gpt-image-2", "/v1/images/generations"},
		{"gpt-image-2-official", "gpt-image-2-official", "/v1/images/generations"},
		{"gemini-3-pro-image-preview", "gemini-3-pro-image-preview", "/v1/images/generations"},
		{"gemini-3-pro-image-preview-official", "gemini-3-pro-image-preview-official", "/v1/images/generations"},
		{"gemini-3-1-flash-image-preview", "gemini-3.1-flash-image-preview", "/v1/images/generations"},
		{"gemini-3-1-flash-image-preview-official", "gemini-3.1-flash-image-preview-official", "/v1/images/generations"},
		{"midjourney", "midjourney", "/v1/midjourney/generations"},
		{"doubao-seedance-4-0", "doubao-seedance-4-0", "/v1/images/generations"},
		{"doubao-seedance-4-5", "doubao-seedance-4-5", "/v1/images/generations"},
	}

	for _, page := range pages {
		block := tutorialCurlBlock(t, sql, page.slug)
		for _, required := range []string{
			"```bash\ncurl https://ai.3zapi.cc" + page.endpoint,
			"\"model\": \"" + page.model + "\"",
			"Authorization: Bearer YOUR_API_KEY",
			"Content-Type: application/json",
			"\n  -d '{\n",
			"\n  }'\n```",
		} {
			if !strings.Contains(block, required) {
				t.Fatalf("tutorial %q cURL block is missing %q", page.slug, required)
			}
		}
		if strings.Count(block, " \\\n") < 3 {
			t.Fatalf("tutorial %q must use multi-line cURL continuations", page.slug)
		}
	}

	for _, required := range []string{
		"regexp_replace(", "curl_examples.curl_md", "updated_at = NOW()",
		"$literal$\\n\\n## cURL$literal$", "E'\\n\\n## cURL'",
		"E'## cURL\\\\r?\\\\n\\\\r?\\\\n```bash\\\\r?\\\\n.*?\\\\r?\\\\n```'",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("cURL format migration is missing %q", required)
		}
	}
	if strings.Contains(sql, "E'\\\\n\\\\n## cURL'") {
		t.Fatal("cURL format migration must not write a literal newline escape before the heading")
	}
	for _, forbidden := range []string{"apimart", "ai.3zapi.top", "ai.3zapi.com"} {
		if strings.Contains(strings.ToLower(sql), forbidden) {
			t.Fatalf("cURL format migration contains forbidden host or branding %q", forbidden)
		}
	}
}

func tutorialCurlBlock(t *testing.T, sql, slug string) string {
	t.Helper()
	startMarker := "('" + slug + "', $curl$"
	start := strings.Index(sql, startMarker)
	if start < 0 {
		t.Fatalf("cURL format migration is missing slug %q", slug)
	}
	remainder := sql[start+len(startMarker):]
	end := strings.Index(remainder, "$curl$)")
	if end < 0 {
		t.Fatalf("cURL format migration has unterminated content for %q", slug)
	}
	return remainder[:end]
}
