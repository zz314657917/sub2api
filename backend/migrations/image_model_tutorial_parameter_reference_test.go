package migrations

import (
	"strings"
	"testing"
)

func TestImageModelTutorialParameterReference(t *testing.T) {
	content, err := FS.ReadFile("230_expand_image_model_parameter_reference.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial parameter reference migration: %v", err)
	}

	sql := string(content)
	pages := []struct {
		slug     string
		required []string
	}{
		{"gpt-image-2", []string{"image_urls", "`1k`、`2k`、`4k`", "413 invalid_request_error"}},
		{"gpt-image-2-official", []string{"`1` 到 `4`", "`0` 到 `100`", "mask_url"}},
		{"gemini-3-pro-image-preview", []string{"最多 `14`", "n`", "429 rate_limit_error"}},
		{"gemini-3-pro-image-preview-official", []string{"最多 `14`", "任务对象", "403 permission_error"}},
		{"gemini-3-1-flash-image-preview", []string{"google_search", "google_image_search", "`0.5k`"}},
		{"gemini-3-1-flash-image-preview-official", []string{"google_search", "google_image_search", "`0.5k`"}},
		{"midjourney", []string{"`relax`、`fast`、`turbo`", "`0` 到 `1000`", "最多 `4`"}},
		{"doubao-seedance-4-0", []string{"合计不得超过 `15`", "data[0].task_id", "`failed`、`cancelled`、`canceled`"}},
		{"doubao-seedance-4-5", []string{"`2k` 或 `4k`", "合计最多 `15`", "error.message"}},
	}

	for _, page := range pages {
		block := tutorialParameterReferenceBlock(t, sql, page.slug)
		for _, section := range []string{"## 字段取值、返回和错误处理", "### 可用字段与取值", "### 成功返回", "### 错误状态与处理"} {
			if !strings.Contains(block, section) {
				t.Fatalf("tutorial %q is missing section %q", page.slug, section)
			}
		}
		for _, required := range page.required {
			if !strings.Contains(block, required) {
				t.Fatalf("tutorial %q is missing detailed reference %q", page.slug, required)
			}
		}
		for _, status := range []string{"400 invalid_request_error", "401 authentication_error", "402 payment_required", "429 rate_limit_error"} {
			if !strings.Contains(block, status) {
				t.Fatalf("tutorial %q is missing error guidance %q", page.slug, status)
			}
		}
	}

	for _, required := range []string{
		"regexp_replace(", "tutorial_references.reference_md", "updated_at = NOW()",
		"E'\\r?\\n\\r?\\n## cURL'", "## AI 调用清单", "## cURL",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("parameter reference migration is missing %q", required)
		}
	}
	for _, forbidden := range []string{"apimart", "ai.3zapi.top", "ai.3zapi.com"} {
		if strings.Contains(strings.ToLower(sql), forbidden) {
			t.Fatalf("parameter reference migration contains forbidden host or branding %q", forbidden)
		}
	}
}

func tutorialParameterReferenceBlock(t *testing.T, sql, slug string) string {
	t.Helper()
	startMarker := "('" + slug + "', $reference$"
	start := strings.Index(sql, startMarker)
	if start < 0 {
		t.Fatalf("parameter reference migration is missing slug %q", slug)
	}
	remainder := sql[start+len(startMarker):]
	end := strings.Index(remainder, "$reference$)")
	if end < 0 {
		t.Fatalf("parameter reference migration has unterminated content for %q", slug)
	}
	return remainder[:end]
}
