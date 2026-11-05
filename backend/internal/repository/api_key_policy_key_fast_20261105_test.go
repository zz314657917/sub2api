package repository

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestKeyFast20261105RepositoryRoundTripAndUpdateMask(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "key-fast-20261105@test.local")
	for i, policy := range []string{"", "passthrough", "force", "filter"} {
		key := &service.APIKey{UserID: user.ID, Key: "sk-key-fast-" + string(rune('a'+i)),
			Name: "key fast", Status: service.StatusActive, OpenAIFastPolicy: policy, QuotaUsed: 17}
		require.NoError(t, repo.Create(ctx, key))
		want := service.NormalizeAPIKeyOpenAIFastPolicy(policy)
		byID, err := repo.GetByID(ctx, key.ID)
		require.NoError(t, err)
		require.Equal(t, want, byID.OpenAIFastPolicy)
		byKey, err := repo.GetByKey(ctx, key.Key)
		require.NoError(t, err)
		require.Equal(t, want, byKey.OpenAIFastPolicy)
		auth, err := repo.GetByKeyForAuth(ctx, key.Key)
		require.NoError(t, err)
		require.Equal(t, want, auth.OpenAIFastPolicy)
		stored, err := client.APIKey.Get(ctx, key.ID)
		require.NoError(t, err)
		require.Equal(t, want, stored.OpenaiFastPolicy)
		// A stale editor object must not reset usage or unrequested policy.
		_, err = client.APIKey.UpdateOneID(key.ID).SetUsage5h(9).Save(ctx)
		require.NoError(t, err)
		key.OpenAIFastPolicy, key.QuotaUsed, key.Usage5h = "force", 0, 0
		require.NoError(t, repo.Update(ctx, key, service.APIKeyUpdateFields{Name: true}))
		stored, err = client.APIKey.Get(ctx, key.ID)
		require.NoError(t, err)
		require.Equal(t, want, stored.OpenaiFastPolicy)
		require.NoError(t, repo.Update(ctx, key, service.APIKeyUpdateFields{OpenAIFastPolicy: true}))
		stored, err = client.APIKey.Get(ctx, key.ID)
		require.NoError(t, err)
		require.Equal(t, "force", stored.OpenaiFastPolicy)
		require.Equal(t, float64(17), stored.QuotaUsed)
		require.Equal(t, float64(9), stored.Usage5h)
	}
	list, result, err := repo.ListByUserID(ctx, user.ID, pagination.PaginationParams{Page: 1, PageSize: 10}, service.APIKeyListFilters{})
	require.NoError(t, err)
	require.EqualValues(t, 4, result.Total)
	require.Len(t, list, 4)
	for _, key := range list {
		require.Equal(t, "force", key.OpenAIFastPolicy)
	}
	// Ent clients outside the user service still get the persisted default and
	// cannot write arbitrary policy strings.
	entity, err := client.APIKey.Create().SetUserID(user.ID).SetName("default").SetKey("sk-ent-key-fast-default").Save(ctx)
	require.NoError(t, err)
	require.Equal(t, "passthrough", entity.OpenaiFastPolicy)
	_, err = client.APIKey.UpdateOneID(entity.ID).SetOpenaiFastPolicy("invalid").Save(ctx)
	require.Error(t, err)
	_, err = client.APIKey.Create().SetUserID(user.ID).SetName("invalid").SetKey("sk-ent-key-fast-invalid").SetOpenaiFastPolicy("invalid").Save(ctx)
	require.Error(t, err)
	require.Equal(t, "passthrough", apiKeyEntityToService(&dbent.APIKey{}).OpenAIFastPolicy)
}

func TestKeyFast20261105RepositoryAuthSelectAndSQLMask(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	repo := newAPIKeyRepositoryWithSQL(client, db)
	mock.ExpectQuery(`SELECT .*"api_keys"\."openai_fast_policy".* FROM "api_keys" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "openai_fast_policy"}).AddRow(7, 0, "active", "filter"))
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT .* FROM "api_key_account_bindings" WHERE`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT multi_group_routes`).WillReturnRows(sqlmock.NewRows([]string{"multi_group_routes"}).AddRow(nil))
	mock.ExpectQuery(`SELECT account_pool_strategy`).WillReturnRows(sqlmock.NewRows([]string{"account_pool_strategy"}).AddRow("shared_only"))
	auth, err := repo.GetByKeyForAuth(context.Background(), "sk-test")
	require.NoError(t, err)
	require.Equal(t, "filter", auth.OpenAIFastPolicy)
	require.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectExec(`^UPDATE "api_keys" SET "updated_at" = \$1, "openai_fast_policy" = \$2 WHERE`).
		WithArgs(sqlmock.AnyArg(), "force", int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), &service.APIKey{ID: 7, OpenAIFastPolicy: "force"},
		service.APIKeyUpdateFields{OpenAIFastPolicy: true}))
	require.NoError(t, mock.ExpectationsWereMet())
	require.Contains(t, apikey.Columns, apikey.FieldOpenaiFastPolicy)
}

func TestKeyFast20261105RedisSnapshotRoundTripAndDelete(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewAPIKeyCache(rdb)
	ctx := context.Background()
	for _, policy := range []string{"passthrough", "force", "filter"} {
		entry := &service.APIKeyAuthCacheEntry{Snapshot: &service.APIKeyAuthSnapshot{
			Version: 21, APIKeyID: 7, OpenAIFastPolicy: policy,
		}}
		require.NoError(t, cache.SetAuthCache(ctx, "key-fast-test", entry, time.Minute))
		out, err := cache.GetAuthCache(ctx, "key-fast-test")
		require.NoError(t, err)
		require.Equal(t, policy, out.Snapshot.OpenAIFastPolicy)
		require.Equal(t, 21, out.Snapshot.Version)
		require.NoError(t, cache.DeleteAuthCache(ctx, "key-fast-test"))
		out, err = cache.GetAuthCache(ctx, "key-fast-test")
		require.ErrorIs(t, err, redis.Nil)
		require.Nil(t, out)
	}
}
