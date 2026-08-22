package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTutorialMigrationsProvidePublishedAt(t *testing.T) {
	testCases := []struct {
		filename string
		slugs    []string
	}{
		{"231_add_grok_image_tutorial_pages.sql", []string{"grok-imagine-1-5", "grok-imagine-1-5-edit"}},
		{"232_replace_grok_15_with_20_tutorial_pages.sql", []string{"grok-imagine-2-0-ext", "grok-imagine-image-2-0"}},
		{"233_add_seedream_5_tutorial_pages.sql", []string{"seedream-5-0-pro", "seedream-5-0-lite"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.filename, func(t *testing.T) {
			content, err := FS.ReadFile(testCase.filename)
			require.NoError(t, err)

			sql := string(content)
			require.Contains(t, sql, "content_md, published_at)")
			for _, slug := range testCase.slugs {
				start := strings.Index(sql, "('"+slug+"',")
				require.NotEqualf(t, -1, start, "missing tutorial slug %q", slug)

				contentStart := strings.Index(sql[start:], "$md$\n")
				require.NotEqualf(t, -1, contentStart, "missing content delimiter for tutorial %q", slug)

				contentEnd := strings.Index(sql[start+contentStart+len("$md$\n"):], "$md$")
				require.NotEqualf(t, -1, contentEnd, "unterminated tutorial %q", slug)
				publishedAt := sql[start+contentStart+len("$md$\n")+contentEnd:]
				require.Truef(t, strings.HasPrefix(publishedAt, "$md$, NOW())"), "tutorial %q must provide published_at", slug)
			}
		})
	}
}
