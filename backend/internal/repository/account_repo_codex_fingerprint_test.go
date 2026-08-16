package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBulkUpdateCodexFingerprintOffRemovesKey(t *testing.T) {
	exec := &recordingSQLExecutor{}
	repo := newAccountRepositoryWithSQL(nil, exec, nil)

	_, err := repo.BulkUpdate(context.Background(), []int64{27}, service.AccountBulkUpdate{
		Extra: map[string]any{"codex_fingerprint_mode": nil},
	})

	require.NoError(t, err)
	require.NotEmpty(t, exec.queries)
	require.Contains(t, normalizeSQLWhitespace(exec.queries[0]), "- 'codex_fingerprint_mode'")
	payload, ok := exec.args[0][0].([]byte)
	require.True(t, ok)
	require.Equal(t, `{"codex_fingerprint_mode":null}`, string(payload))
}

func TestBulkUpdateCodexFingerprintModeKeepsOtherNullKeysUntouched(t *testing.T) {
	exec := &recordingSQLExecutor{}
	repo := newAccountRepositoryWithSQL(nil, exec, nil)

	_, err := repo.BulkUpdate(context.Background(), []int64{27}, service.AccountBulkUpdate{
		Extra: map[string]any{"unrelated_key": nil},
	})

	require.NoError(t, err)
	require.NotEmpty(t, exec.queries)
	require.NotContains(t, normalizeSQLWhitespace(exec.queries[0]), "- 'codex_fingerprint_mode'")
}
