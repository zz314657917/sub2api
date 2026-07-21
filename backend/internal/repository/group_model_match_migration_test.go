package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type s91MigrationInvalidator struct {
	keys []string
}

func (i *s91MigrationInvalidator) InvalidateAuthCacheByKey(_ context.Context, key string) {
	i.keys = append(i.keys, key)
}

func (*s91MigrationInvalidator) InvalidateAuthCacheByUserID(context.Context, int64) {}

func (*s91MigrationInvalidator) InvalidateAuthCacheByGroupID(context.Context, int64) {}

func newS91MigrationClient(t *testing.T) *dbent.Client {
	t.Helper()
	name := fmt.Sprintf("file:s91_group_model_match_%s?mode=memory&cache=shared&_fk=1", t.Name())
	db, err := sql.Open("sqlite", name)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func createS91MigrationUser(t *testing.T, ctx context.Context, client *dbent.Client, id int64) {
	t.Helper()
	_, err := client.User.Create().
		SetEmail(fmt.Sprintf("s91-%d@example.com", id)).
		SetPasswordHash("test-password-hash").
		SetRole(service.RoleUser).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
}

func createS91MigrationGroup(t *testing.T, ctx context.Context, client *dbent.Client, name string, patterns []string) int64 {
	t.Helper()
	g, err := client.Group.Create().
		SetName(name).
		SetStatus(service.StatusActive).
		SetPlatform(service.PlatformOpenAI).
		SetModelMatchPatterns(patterns).
		Save(ctx)
	require.NoError(t, err)
	return g.ID
}

func createS91LegacyKey(t *testing.T, ctx context.Context, client *dbent.Client, id int64, key string, routes []domain.APIKeyMultiGroupRoute) {
	t.Helper()
	_, err := client.APIKey.Create().
		SetUserID(id).
		SetKey(key).
		SetName(key).
		SetStatus(service.StatusActive).
		SetMultiGroupRoutes(routes).
		Save(ctx)
	require.NoError(t, err)
}

func TestS91GroupModelMatchMigrationDryRunAndSwitch(t *testing.T) {
	ctx := context.Background()
	client := newS91MigrationClient(t)
	createS91MigrationUser(t, ctx, client, 1)
	configuredGroupID := createS91MigrationGroup(t, ctx, client, "configured", []string{"gpt-*"})
	createS91LegacyKey(t, ctx, client, 1, "s91-legacy-key", []domain.APIKeyMultiGroupRoute{{
		GroupID:       configuredGroupID,
		Priority:      1,
		Weight:        1,
		Enabled:       true,
		ModelPatterns: []string{"gpt-*"},
	}})

	invalidator := &s91MigrationInvalidator{}
	migration := NewGroupModelMatchMigration(client, invalidator)

	report, err := migration.Switch(ctx, true)
	require.NoError(t, err)
	require.Len(t, report.LegacyAPIKeys, 1)
	legacy, err := client.APIKey.Query().Where().Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, legacy.MultiGroupRoutes[0].ModelPatterns)
	require.Empty(t, invalidator.keys)

	_, err = migration.Switch(ctx, false)
	require.NoError(t, err)
	legacy, err = client.APIKey.Query().Where().Only(ctx)
	require.NoError(t, err)
	require.Nil(t, legacy.MultiGroupRoutes[0].ModelPatterns)
	require.Equal(t, []string{"s91-legacy-key"}, invalidator.keys)
}

func TestS91GroupModelMatchMigrationBlocksUnconfiguredAndReportsConflict(t *testing.T) {
	ctx := context.Background()
	client := newS91MigrationClient(t)
	createS91MigrationUser(t, ctx, client, 1)
	configuredGroupID := createS91MigrationGroup(t, ctx, client, "configured", []string{"gpt-*"})
	unconfiguredGroupID := createS91MigrationGroup(t, ctx, client, "unconfigured", nil)
	createS91LegacyKey(t, ctx, client, 1, "s91-key-a", []domain.APIKeyMultiGroupRoute{{
		GroupID:       configuredGroupID,
		ModelPatterns: []string{"gpt-*"},
	}})
	createS91LegacyKey(t, ctx, client, 1, "s91-key-b", []domain.APIKeyMultiGroupRoute{{
		GroupID:       configuredGroupID,
		ModelPatterns: []string{"claude-*"},
	}, {
		GroupID:       unconfiguredGroupID,
		ModelPatterns: []string{"claude-*"},
	}})

	invalidator := &s91MigrationInvalidator{}
	migration := NewGroupModelMatchMigration(client, invalidator)
	report, err := migration.Preflight(ctx)
	require.NoError(t, err)
	require.Len(t, report.UnconfiguredGroups, 1)
	require.Equal(t, unconfiguredGroupID, report.UnconfiguredGroups[0].ID)
	require.Len(t, report.Conflicts, 1)
	require.Equal(t, configuredGroupID, report.Conflicts[0].GroupID)

	_, err = migration.Switch(ctx, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unconfigured")
	require.Empty(t, invalidator.keys)
	keys, err := client.APIKey.Query().All(ctx)
	require.NoError(t, err)
	for _, key := range keys {
		require.NotNil(t, key.MultiGroupRoutes[0].ModelPatterns)
	}
}
