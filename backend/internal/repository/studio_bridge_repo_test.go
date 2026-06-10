package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestStudioBridgeRepositoryResolveChargeUsageRefsUsesDefaultKey(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	mock.ExpectQuery(`WITH existing_key AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), nil, int64(88)))

	refs, err := repo.resolveChargeUsageRefs(context.Background(), db, service.StudioBridgeChargeCommand{
		UserID: 42,
		Mode:   "chat",
	})
	require.NoError(t, err)
	require.Equal(t, int64(77), refs.apiKeyID)
	require.Equal(t, int64(88), refs.accountID)
	require.False(t, refs.groupID.Valid)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudioBridgeRepositoryResolveChargeUsageRefsPreservesDefaultKeyGroup(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	mock.ExpectQuery(`WITH existing_key AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "image").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), int64(9), int64(88)))

	refs, err := repo.resolveChargeUsageRefs(context.Background(), db, service.StudioBridgeChargeCommand{
		UserID: 42,
		Mode:   "edit",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err)
	require.Equal(t, int64(77), refs.apiKeyID)
	require.Equal(t, int64(88), refs.accountID)
	require.True(t, refs.groupID.Valid)
	require.Equal(t, int64(9), refs.groupID.Int64)
	require.NoError(t, mock.ExpectationsWereMet())
}
