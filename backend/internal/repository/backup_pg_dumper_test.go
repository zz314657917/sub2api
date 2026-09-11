package repository

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func pgDumperTestCommand(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPgDumperHelperProcess$")
	cmd.Env = append(os.Environ(), "SUB2API_TEST_PG_DUMP_HELPER=1")
	return cmd
}

func TestPgDumperHelperProcess(t *testing.T) {
	if os.Getenv("SUB2API_TEST_PG_DUMP_HELPER") != "1" {
		return
	}
	_, err := io.WriteString(os.Stdout, "backup")
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func testPgDumper(t *testing.T, command func(context.Context, string, ...string) *exec.Cmd) (*PgDumper, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &PgDumper{cfg: &config.DatabaseConfig{Host: "db", Port: 5432, User: "u", DBName: "d"}, db: db, commandContext: command}, mock
}

func expectLock(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT pg_try_advisory_lock($1)")).WithArgs(migrationsAdvisoryLockID).
		WillReturnRows(sqlmock.NewRows([]string{"locked"}).AddRow(true))
}

func expectUnlock(mock sqlmock.Sqlmock) {
	mock.ExpectExec(regexp.QuoteMeta("SELECT pg_advisory_unlock($1)")).WithArgs(migrationsAdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestPgDumperHoldsMigrationLockUntilReaderClose(t *testing.T) {
	var mock sqlmock.Sqlmock
	dumper, createdMock := testPgDumper(t, func(ctx context.Context, name string, args ...string) *exec.Cmd {
		require.Equal(t, "pg_dump", name)
		require.NoError(t, mock.ExpectationsWereMet())
		return pgDumperTestCommand(ctx)
	})
	mock = createdMock
	expectLock(mock)
	reader, err := dumper.Dump(context.Background())
	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "backup", string(data))
	expectUnlock(mock)
	require.Error(t, mock.ExpectationsWereMet())
	require.NoError(t, reader.Close())
	require.NoError(t, mock.ExpectationsWereMet())
	require.NoError(t, reader.Close())
}

func TestPgDumperReleasesLockWhenProcessStartFails(t *testing.T) {
	dumper, mock := testPgDumper(t, func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "/missing/pg_dump")
	})
	expectLock(mock)
	expectUnlock(mock)
	reader, err := dumper.Dump(context.Background())
	require.Nil(t, reader)
	require.ErrorContains(t, err, "start pg_dump")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgDumperUnlockFailureIsReported(t *testing.T) {
	dumper, mock := testPgDumper(t, func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return pgDumperTestCommand(ctx)
	})
	expectLock(mock)
	reader, err := dumper.Dump(context.Background())
	require.NoError(t, err)
	_, err = io.ReadAll(reader)
	require.NoError(t, err)
	mock.ExpectExec(regexp.QuoteMeta("SELECT pg_advisory_unlock($1)")).WithArgs(migrationsAdvisoryLockID).WillReturnError(errors.New("unlock failed"))
	require.ErrorContains(t, reader.Close(), "release backup migration lock")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgDumperRejectsNilDB(t *testing.T) {
	dumper := &PgDumper{cfg: &config.DatabaseConfig{}}
	reader, err := dumper.Dump(context.Background())
	require.Nil(t, reader)
	require.ErrorContains(t, err, "nil sql db")
}
