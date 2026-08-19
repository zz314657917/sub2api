package migrations

import (
	"regexp"
	"strings"
	"testing"
)

func TestImageModelTutorialPages(t *testing.T) {
	content, err := FS.ReadFile("224_image_model_tutorial_pages.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial migration: %v", err)
	}

	sql := string(content)
	domainContent, err := FS.ReadFile("225_update_image_model_tutorial_domains.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial domain migration: %v", err)
	}
	domainSQL := string(domainContent)
	ccDomainContent, err := FS.ReadFile("226_update_image_model_tutorial_domains_to_cc.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial .cc domain migration: %v", err)
	}
	ccDomainSQL := string(ccDomainContent)
	parametersContent, err := FS.ReadFile("227_format_image_model_tutorial_parameters.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial parameter migration: %v", err)
	}
	parametersSQL := string(parametersContent)
	pages := []struct {
		slug       string
		model      string
		parameters []string
	}{
		{"gpt-image-2", "gpt-image-2", []string{"model", "prompt", "size", "resolution", "n"}},
		{"gpt-image-2-official", "gpt-image-2-official", []string{"model", "prompt", "size", "resolution", "n", "quality", "background", "moderation", "output_format", "output_compression", "mask_url"}},
		{"gemini-3-pro-image-preview", "gemini-3-pro-image-preview", []string{"model", "prompt", "size", "resolution", "image_urls", "n"}},
		{"gemini-3-pro-image-preview-official", "gemini-3-pro-image-preview-official", []string{"model", "prompt", "size", "resolution", "image_urls", "n"}},
		{"gemini-3-1-flash-image-preview", "gemini-3.1-flash-image-preview", []string{"model", "prompt", "size", "resolution", "image_urls", "n", "google_search", "google_image_search"}},
		{"gemini-3-1-flash-image-preview-official", "gemini-3.1-flash-image-preview-official", []string{"model", "prompt", "size", "resolution", "image_urls", "n", "google_search", "google_image_search"}},
		{"midjourney", "midjourney", []string{"model", "prompt", "size", "version", "speed", "quality", "stylize", "chaos", "weird", "stop", "niji", "raw", "tile", "image_urls"}},
		{"doubao-seedance-4-0", "doubao-seedance-4-0", []string{"model", "prompt", "size", "resolution", "n", "image_urls"}},
		{"doubao-seedance-4-5", "doubao-seedance-4-5", []string{"model", "prompt", "size", "resolution", "n", "image_urls"}},
	}
	validSlug := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,78}[a-z0-9])?$`)
	for _, page := range pages {
		if !validSlug.MatchString(page.slug) {
			t.Fatalf("tutorial slug %q violates tutorial_pages_slug_check", page.slug)
		}
		if strings.Count(sql, "'"+page.slug+"'") != 1 {
			t.Fatalf("expected exactly one seeded slug %q", page.slug)
		}
		if !strings.Contains(sql, "`"+page.model+"`") {
			t.Fatalf("migration is missing model ID %q", page.model)
		}
		if strings.Count(domainSQL, "'"+page.slug+"'") != 1 {
			t.Fatalf("domain migration must target tutorial %q exactly once", page.slug)
		}
		if strings.Count(ccDomainSQL, "'"+page.slug+"'") != 1 {
			t.Fatalf(".cc domain migration must target tutorial %q exactly once", page.slug)
		}
		parameterBlockStart := "('" + page.slug + "', $params$"
		parameterBlockOffset := strings.Index(parametersSQL, parameterBlockStart)
		if parameterBlockOffset < 0 {
			t.Fatalf("parameter migration is missing tutorial %q", page.slug)
		}
		parameterBlock := parametersSQL[parameterBlockOffset+len(parameterBlockStart):]
		parameterBlockEnd := strings.Index(parameterBlock, "$params$)")
		if parameterBlockEnd < 0 {
			t.Fatalf("parameter migration has an unterminated block for tutorial %q", page.slug)
		}
		parameterBlock = parameterBlock[:parameterBlockEnd]
		if strings.Count(parameterBlock, "\n- `") != len(page.parameters) {
			t.Fatalf("tutorial %q must render exactly one list item per parameter", page.slug)
		}
		for _, parameter := range page.parameters {
			parameterLine := "\n- `" + parameter + "`:"
			if strings.Count(parameterBlock, parameterLine) != 1 {
				t.Fatalf("tutorial %q must render parameter %q on its own line", page.slug, parameter)
			}
		}
	}

	for _, required := range []string{
		"regexp_replace(", "page.content_md", "btrim(parameter_blocks.parameters_md, E'\\r\\n')", "updated_at = NOW()",
		"E'## 参数\\\\r?\\\\n\\\\r?\\\\n.*?\\\\r?\\\\n## cURL'",
	} {
		if !strings.Contains(parametersSQL, required) {
			t.Fatalf("parameter migration is missing required content %q", required)
		}
	}

	for _, required := range []string{
		"replace(", "content_md,", "'https://ai.3zapi.top'", "'https://ai.3zapi.com'",
		"updated_at = NOW()", "content_md LIKE '%https://ai.3zapi.top%'",
	} {
		if !strings.Contains(domainSQL, required) {
			t.Fatalf("domain migration is missing required content %q", required)
		}
	}

	for _, required := range []string{
		"replace(", "content_md,", "'https://ai.3zapi.com'", "'https://ai.3zapi.cc'",
		"updated_at = NOW()", "content_md LIKE '%https://ai.3zapi.com%'",
	} {
		if !strings.Contains(ccDomainSQL, required) {
			t.Fatalf(".cc domain migration is missing required content %q", required)
		}
	}

	for _, required := range []string{
		"'图像模型'", "'published'", "https://ai.3zapi.top", "Authorization: Bearer YOUR_API_KEY",
		"## cURL", "## Python", "## JavaScript", "import requests", "fetch(",
		"ON CONFLICT (slug) DO NOTHING",
		"POST https://ai.3zapi.top/v1/images/generations",
		"POST https://ai.3zapi.top/v1/midjourney/generations",
		"{\"created\": 0, \"data\": [{\"url\": \"...\"}]}",
		"GET /v1/tasks/{task_id}", "data.result.images[0].url[0]", "结果 URL 应尽快下载",
		"最多 14 张", "1k`、`2k`、`4k`", "google_search", "google_image_search",
		"兼容流程中 `n` 固定为 `1`", "参考图数量加 `n` 不得超过 15", "推荐 `resolution` 仅使用 `2k` 或 `4k`",
		"\"resolution\":\"1k\"", "\"resolution\":\"2k\"", "\"resolution\":\"4k\"",
		"created[\"data\"][0][\"task_id\"]", "created.data[0].task_id", "Seedream task failed",
		"`size`、`version`、`speed`、`quality`、`stylize`、`chaos`、`weird`、`stop`、`niji`、`raw`、`tile`、`image_urls`",
		"`quality`、`background`、`moderation`、`output_format`、`output_compression`、`mask_url`",
		"不要使用 `0.5k`",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration is missing required tutorial content %q", required)
		}
	}

	for _, marker := range []string{"## cURL", "## Python", "## JavaScript", "Authorization: Bearer YOUR_API_KEY"} {
		if strings.Count(sql, marker) < len(pages) {
			t.Fatalf("expected every page to include %q", marker)
		}
	}
	for _, forbidden := range []string{"apimart", "api.apimart.ai", "cdn.apimart.ai"} {
		if strings.Contains(strings.ToLower(sql), forbidden) {
			t.Fatalf("migration contains forbidden reference branding or host %q", forbidden)
		}
	}
	for _, forbiddenHost := range []string{"https://example.com", "http://"} {
		if strings.Contains(strings.ToLower(sql), forbiddenHost) {
			t.Fatalf("migration contains non-local example host %q", forbiddenHost)
		}
	}
	for _, forbiddenTierAsSize := range []string{
		"\"size\":\"1k\"", "\"size\":\"2k\"", "\"size\":\"4k\"",
		"\"size\": \"1k\"", "\"size\": \"2k\"", "\"size\": \"4k\"",
		"`size` 支持", "`size`: `1k`", "推荐 `size` 仅使用", "`size` 不是 `1k`",
	} {
		if strings.Contains(sql, forbiddenTierAsSize) {
			t.Fatalf("migration documents a resolution tier through size %q", forbiddenTierAsSize)
		}
	}
}
