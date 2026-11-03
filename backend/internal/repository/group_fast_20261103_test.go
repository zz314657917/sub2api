package repository

import (
	"context"
	"errors"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupFast20261103RepositoryWritesAndProjection(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
		stopped := errors.New("test intercepted before any database call")
		calls := 0
		client.Group.Use(func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
				value, exists := mutation.Field("force_openai_fast")
				require.True(t, exists, "both false and true must be explicitly persisted")
				require.Equal(t, enabled, value)
				calls++
				return nil, stopped
			})
		})
		repo := newGroupRepositoryWithSQL(client, db)
		group := &service.Group{ID: 7, Name: "test", Platform: service.PlatformOpenAI,
			Status: service.StatusActive, ForceOpenAIFast: enabled}
		require.ErrorIs(t, createGroupRecord(context.Background(), client, group), stopped)
		mock.ExpectQuery(`SELECT .* FROM "groups" WHERE`).
			WithArgs(int64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "platform", "status", "force_openai_fast"}).
				AddRow(int64(7), "test", service.PlatformOpenAI, service.StatusActive, enabled))
		require.ErrorIs(t, repo.Update(context.Background(), group), stopped)
		require.Equal(t, 2, calls)
		require.NoError(t, mock.ExpectationsWereMet())
		projected := groupEntityToService(&dbent.Group{ID: 7, Platform: service.PlatformOpenAI,
			Status: service.StatusActive, ForceOpenaiFast: enabled})
		require.Equal(t, enabled, projected.ForceOpenAIFast)
		require.True(t, projected.Hydrated)
	}
}

func TestGroupFast20261103RepositoryAuthProjection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := newAPIKeyRepositoryWithSQL(client, db)
	mock.ExpectQuery(`SELECT .* FROM "api_keys" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "group_id", "status"}).AddRow(1, 0, 7, "active"))
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT .*"groups"\."force_openai_fast".* FROM "groups" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "platform", "status", "force_openai_fast"}).
			AddRow(7, service.PlatformOpenAI, service.StatusActive, true))
	mock.ExpectQuery(`SELECT .* FROM "api_key_account_bindings" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT multi_group_routes`).
		WillReturnRows(sqlmock.NewRows([]string{"multi_group_routes"}).AddRow(nil))
	mock.ExpectQuery(`SELECT account_pool_strategy`).
		WillReturnRows(sqlmock.NewRows([]string{"account_pool_strategy"}).AddRow("shared"))
	key, err := repo.GetByKeyForAuth(context.Background(), "test")
	require.NoError(t, err)
	require.NotNil(t, key.Group)
	require.True(t, key.Group.ForceOpenAIFast)
	require.True(t, key.Group.Hydrated)
	require.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT .*"groups"\."force_openai_fast".* FROM "groups" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "platform", "status", "force_openai_fast"}).
			AddRow(7, service.PlatformComposite, service.StatusActive, true))
	key.MultiGroupRoutes = []domain.APIKeyMultiGroupRoute{{GroupID: 7}}
	require.NoError(t, repo.hydrateAPIKeyMultiGroupRouteGroups(context.Background(), key))
	require.Len(t, key.MultiGroupRouteGroups, 1)
	require.True(t, key.MultiGroupRouteGroups[0].ForceOpenAIFast)
	require.True(t, key.MultiGroupRouteGroups[0].Hydrated)
	require.NoError(t, mock.ExpectationsWereMet())
}
