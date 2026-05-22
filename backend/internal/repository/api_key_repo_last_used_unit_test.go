package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newAPIKeyRepoSQLite(t *testing.T) (*apiKeyRepository, *dbent.Client) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:api_key_repo_last_used?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	_, err = db.Exec("ALTER TABLE api_keys ADD COLUMN multi_group_routes TEXT NOT NULL DEFAULT '[]'")
	require.NoError(t, err)

	return &apiKeyRepository{client: client, dialect: dialect.SQLite}, client
}

func mustCreateAPIKeyRepoUser(t *testing.T, ctx context.Context, client *dbent.Client, email string) *service.User {
	t.Helper()
	u, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("test-password-hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	return userEntityToService(u)
}

func mustCreateAPIKeyRepoGroup(t *testing.T, ctx context.Context, client *dbent.Client, name string) *service.Group {
	t.Helper()
	g, err := client.Group.Create().
		SetName(name).
		SetStatus(service.StatusActive).
		SetPlatform(service.PlatformOpenAI).
		Save(ctx)
	require.NoError(t, err)
	return groupEntityToService(g)
}

func TestAPIKeyRepository_CreateWithLastUsedAt(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "create-last-used@test.com")

	lastUsed := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	key := &service.APIKey{
		UserID:     user.ID,
		Key:        "sk-create-last-used",
		Name:       "CreateWithLastUsed",
		Status:     service.StatusActive,
		LastUsedAt: &lastUsed,
	}

	require.NoError(t, repo.Create(ctx, key))
	require.NotNil(t, key.LastUsedAt)
	require.WithinDuration(t, lastUsed, *key.LastUsedAt, time.Second)

	got, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.NotNil(t, got.LastUsedAt)
	require.WithinDuration(t, lastUsed, *got.LastUsedAt, time.Second)
}

func TestAPIKeyRepository_UpdateLastUsed(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "update-last-used@test.com")

	key := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-update-last-used",
		Name:   "UpdateLastUsed",
		Status: service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	before, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.Nil(t, before.LastUsedAt)

	target := time.Now().UTC().Add(2 * time.Minute).Truncate(time.Second)
	require.NoError(t, repo.UpdateLastUsed(ctx, key.ID, target))

	after, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.NotNil(t, after.LastUsedAt)
	require.WithinDuration(t, target, *after.LastUsedAt, time.Second)
	require.WithinDuration(t, target, after.UpdatedAt, time.Second)
}

func TestAPIKeyRepository_UpdateLastUsedDeletedKey(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "deleted-last-used@test.com")

	key := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-update-last-used-deleted",
		Name:   "UpdateLastUsedDeleted",
		Status: service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))
	require.NoError(t, repo.Delete(ctx, key.ID))

	err := repo.UpdateLastUsed(ctx, key.ID, time.Now().UTC())
	require.ErrorIs(t, err, service.ErrAPIKeyNotFound)
}

func TestAPIKeyRepository_UpdateLastUsedDBError(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "db-error-last-used@test.com")

	key := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-update-last-used-db-error",
		Name:   "UpdateLastUsedDBError",
		Status: service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	require.NoError(t, client.Close())
	err := repo.UpdateLastUsed(ctx, key.ID, time.Now().UTC())
	require.Error(t, err)
}

func TestAPIKeyRepository_CreateDuplicateKey(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "duplicate-key@test.com")

	first := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-duplicate",
		Name:   "first",
		Status: service.StatusActive,
	}
	second := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-duplicate",
		Name:   "second",
		Status: service.StatusActive,
	}

	require.NoError(t, repo.Create(ctx, first))
	err := repo.Create(ctx, second)
	require.ErrorIs(t, err, service.ErrAPIKeyExists)
}

func TestAPIKeyRepository_ClearGroupIDByGroupIDClearsMultiGroupRoutes_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "clear-route-group@test.com")
	primary := mustCreateAPIKeyRepoGroup(t, ctx, client, "clear-route-primary")
	fallback := mustCreateAPIKeyRepoGroup(t, ctx, client, "clear-route-fallback")

	key := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-clear-route-group",
		Name:    "ClearRouteGroup",
		GroupID: &primary.ID,
		Status:  service.StatusActive,
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: primary.ID, Priority: 100, Weight: 1, CooldownSeconds: 30, Enabled: true},
			{GroupID: fallback.ID, Priority: 110, Weight: 1, CooldownSeconds: 30, Enabled: true},
		},
	}
	require.NoError(t, repo.Create(ctx, key))

	beforeKeys, err := repo.ListKeysByGroupID(ctx, primary.ID)
	require.NoError(t, err)
	require.Contains(t, beforeKeys, key.Key)

	affected, err := repo.ClearGroupIDByGroupID(ctx, primary.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, affected, int64(1))

	got, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.Nil(t, got.GroupID)
	require.Len(t, got.MultiGroupRoutes, 1)
	require.Equal(t, fallback.ID, got.MultiGroupRoutes[0].GroupID)
	require.Len(t, got.MultiGroupRouteGroups, 1)
	require.Equal(t, fallback.ID, got.MultiGroupRouteGroups[0].ID)

	primaryKeys, err := repo.ListKeysByGroupID(ctx, primary.ID)
	require.NoError(t, err)
	require.NotContains(t, primaryKeys, key.Key)

	fallbackKeys, err := repo.ListKeysByGroupID(ctx, fallback.ID)
	require.NoError(t, err)
	require.Contains(t, fallbackKeys, key.Key)
}
