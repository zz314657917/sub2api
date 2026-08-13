package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type backupLeaderLockSettingRepo struct {
	values   map[string]string
	getCalls int
}

func (r *backupLeaderLockSettingRepo) Get(context.Context, string) (*Setting, error) {
	return nil, nil
}

func (r *backupLeaderLockSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	r.getCalls++
	return r.values[key], nil
}

func (r *backupLeaderLockSettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *backupLeaderLockSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (r *backupLeaderLockSettingRepo) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}

func (r *backupLeaderLockSettingRepo) GetAll(context.Context) (map[string]string, error) {
	return r.values, nil
}

func (r *backupLeaderLockSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

func newBackupServiceLeaderLockTestService(db *sql.DB, repo *backupLeaderLockSettingRepo) *BackupService {
	svc := NewBackupService(repo, &config.Config{}, nil, nil, nil)
	svc.db = db
	return svc
}

func TestBackupServiceScheduledLeaderLockMissSkipsBeforeScheduleLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT pg_try_advisory_lock\\(\\$1\\)").
		WithArgs(hashAdvisoryLockID("backup:scheduled:leader")).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(false))
	repo := &backupLeaderLockSettingRepo{values: map[string]string{}}

	newBackupServiceLeaderLockTestService(db, repo).runScheduledBackup()

	require.Zero(t, repo.getCalls)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBackupServiceScheduledLeaderLockAcquisitionErrorFailsClosedAndLogs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT pg_try_advisory_lock\\(\\$1\\)").
		WithArgs(hashAdvisoryLockID("backup:scheduled:leader")).
		WillReturnError(errors.New("database unavailable"))
	repo := &backupLeaderLockSettingRepo{values: map[string]string{}}
	svc := newBackupServiceLeaderLockTestService(db, repo)
	var logged string
	svc.logf = func(_ string, format string, args ...any) { logged = format + ": " + args[0].(error).Error() }

	svc.runScheduledBackup()

	require.Contains(t, logged, "scheduled backup")
	require.Contains(t, logged, "database unavailable")
	require.Zero(t, repo.getCalls)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBackupServiceScheduledLeaderAcquiresAndReleasesLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	lockID := hashAdvisoryLockID("backup:scheduled:leader")
	mock.ExpectQuery("SELECT pg_try_advisory_lock\\(\\$1\\)").
		WithArgs(lockID).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(true))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\)").
		WithArgs(lockID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	repo := &backupLeaderLockSettingRepo{values: map[string]string{}}

	newBackupServiceLeaderLockTestService(db, repo).runScheduledBackup()

	require.GreaterOrEqual(t, repo.getCalls, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBackupServiceScheduledLeaderNilDBRunsWithoutLock(t *testing.T) {
	repo := &backupLeaderLockSettingRepo{values: map[string]string{}}

	newBackupServiceLeaderLockTestService(nil, repo).runScheduledBackup()

	require.GreaterOrEqual(t, repo.getCalls, 2)
}

func TestBackupServiceScheduledLeaderManualCreateBackupDoesNotAcquireLeaderLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := &backupLeaderLockSettingRepo{values: map[string]string{}}

	_, err = newBackupServiceLeaderLockTestService(db, repo).CreateBackup(context.Background(), "manual", 0)

	require.ErrorIs(t, err, ErrBackupS3NotConfigured)
	require.NoError(t, mock.ExpectationsWereMet())
}
