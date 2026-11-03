package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupFast20261103Migration(t *testing.T) {
	raw, err := FS.ReadFile("254_group_force_openai_fast.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(raw))
	require.Contains(t, sql, "alter table groups")
	require.Contains(t, sql, "add column if not exists force_openai_fast boolean not null default false")
	require.NotContains(t, sql, "update groups")
	require.NotContains(t, sql, "drop ")
}
