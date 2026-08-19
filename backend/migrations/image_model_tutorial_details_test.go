package migrations

import (
	"strings"
	"testing"
)

func TestImageModelTutorialDetails(t *testing.T) {
	content, err := FS.ReadFile("228_expand_image_model_tutorial_details.sql")
	if err != nil {
		t.Fatalf("read image-model tutorial details migration: %v", err)
	}

	sql := string(content)
	pages := []struct {
		slug        string
		parameters  []string
		mustContain []string
	}{
		{"gpt-image-2", []string{"model", "prompt", "size", "resolution", "n"}, []string{"同步最终响应", "任务目标 + 主体 + 环境"}},
		{"gpt-image-2-official", []string{"model", "prompt", "size", "resolution", "n", "quality", "background", "moderation", "output_format", "output_compression", "mask_url"}, []string{"`auto`", "`opaque`", "透明"}},
		{"gemini-3-pro-image-preview", []string{"model", "prompt", "size", "resolution", "image_urls", "n"}, []string{"最多 14 张", "固定为 `1`"}},
		{"gemini-3-pro-image-preview-official", []string{"model", "prompt", "size", "resolution", "image_urls", "n"}, []string{"最多 14 张", "固定为 `1`"}},
		{"gemini-3-1-flash-image-preview", []string{"model", "prompt", "size", "resolution", "image_urls", "n", "google_search", "google_image_search"}, []string{"不要使用 `0.5k`", "时效性"}},
		{"gemini-3-1-flash-image-preview-official", []string{"model", "prompt", "size", "resolution", "image_urls", "n", "google_search", "google_image_search"}, []string{"`0.5k` 不是本地兼容输入", "使用边界"}},
		{"midjourney", []string{"model", "prompt", "size", "version", "speed", "quality", "stylize", "chaos", "weird", "stop", "niji", "raw", "tile", "image_urls"}, []string{"/v1/midjourney/generations", "同步最终响应"}},
		{"doubao-seedance-4-0", []string{"model", "prompt", "size", "resolution", "n", "image_urls"}, []string{"data[0].task_id", "data.result.images[0].url[0]"}},
		{"doubao-seedance-4-5", []string{"model", "prompt", "size", "resolution", "n", "image_urls"}, []string{"建议仅使用 `2k` 或 `4k`", "应用重启后应继续使用已保存的任务 ID"}},
	}

	for _, page := range pages {
		block := tutorialDetailsBlock(t, sql, page.slug)
		for _, section := range []string{"## 适用场景", "## 参数", "## 提示词编写", "## AI 调用清单"} {
			if !strings.Contains(block, section) {
				t.Fatalf("tutorial %q is missing section %q", page.slug, section)
			}
		}
		if strings.Count(block, "\n- `") != len(page.parameters) {
			t.Fatalf("tutorial %q must describe every parameter on its own line", page.slug)
		}
		for _, parameter := range page.parameters {
			if strings.Count(block, "\n- `"+parameter+"`:") != 1 {
				t.Fatalf("tutorial %q must describe parameter %q on its own line", page.slug, parameter)
			}
		}
		for _, marker := range page.mustContain {
			if !strings.Contains(block, marker) {
				t.Fatalf("tutorial %q is missing detailed guidance %q", page.slug, marker)
			}
		}
		if !strings.Contains(block, "https://ai.3zapi.cc") {
			t.Fatalf("tutorial %q does not use the local .cc API host", page.slug)
		}
	}

	for _, required := range []string{
		"regexp_replace(", "tutorial_details.details_md", "updated_at = NOW()",
		"E'## 参数\\\\r?\\\\n\\\\r?\\\\n.*?\\\\r?\\\\n## cURL'", "## cURL",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("tutorial detail migration is missing %q", required)
		}
	}
	for _, forbidden := range []string{"apimart", "ai.3zapi.top", "ai.3zapi.com"} {
		if strings.Contains(strings.ToLower(sql), forbidden) {
			t.Fatalf("tutorial detail migration contains forbidden host or branding %q", forbidden)
		}
	}
}

func tutorialDetailsBlock(t *testing.T, sql, slug string) string {
	t.Helper()
	startMarker := "('" + slug + "', $details$"
	start := strings.Index(sql, startMarker)
	if start < 0 {
		t.Fatalf("tutorial detail migration is missing slug %q", slug)
	}
	remainder := sql[start+len(startMarker):]
	end := strings.Index(remainder, "$details$)")
	if end < 0 {
		t.Fatalf("tutorial detail migration has unterminated content for %q", slug)
	}
	return remainder[:end]
}
