package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentModerationMatchedKeywordMigrationIsIdempotent(t *testing.T) {
	path := filepath.Join("237_content_moderation_matched_keyword.sql")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	sql := strings.ToLower(string(raw))
	require.Contains(t, sql, "alter table content_moderation_logs")
	require.Contains(t, sql, "add column if not exists matched_keyword")
	require.Contains(t, sql, "varchar(255)")
	require.Contains(t, sql, "not null default ''")
	require.NotContains(t, sql, "drop column")
}
