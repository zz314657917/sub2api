//go:build unit

package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/intercept"
	"github.com/stretchr/testify/require"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestEnsureEmailAuthIdentityCreateErrorReturnsFalse(t *testing.T) {
	db, err := sql.Open("sqlite", "file:auth_service_identity_shadow?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	var authIdentityQueries int
	client.AuthIdentity.Intercept(intercept.TraverseFunc(func(context.Context, intercept.Query) error {
		authIdentityQueries++
		return nil
	}))
	client.AuthIdentity.Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(_ context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op().Is(ent.OpCreate) {
				return nil, errors.New("forced auth identity create failure")
			}
			return next.Mutate(context.Background(), m)
		})
	})

	svc := &AuthService{entClient: client}
	identity, created := svc.ensureEmailAuthIdentity(context.Background(), &User{
		ID:    123,
		Email: "shadow@example.com",
	}, "test")

	require.Nil(t, identity)
	require.False(t, created)
	require.Equal(t, 1, authIdentityQueries, "create failure should not be swallowed and followed by a reload query")
}
