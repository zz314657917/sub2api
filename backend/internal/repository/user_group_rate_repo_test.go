package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSyncGroupRateMultipliersTransaction(t *testing.T) {
	for _, name := range []string{"empty", "duplicate zero wins", "upsert failure"} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := NewUserGroupRateRepository(db)
			var input []service.GroupRateMultiplierInput
			keep := "{}"
			if name != "empty" {
				input = []service.GroupRateMultiplierInput{{UserID: 7, RateMultiplier: 2}, {UserID: 7, RateMultiplier: 0}}
				keep = "{7}"
			}
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE user_group_rate_multipliers SET rate_multiplier = NULL, updated_at = NOW\(\) WHERE group_id = \$1 AND user_id <> ALL\(\$2\)`).WithArgs(4, keep).WillReturnResult(sqlmock.NewResult(0, 2))
			mock.ExpectExec(`DELETE FROM user_group_rate_multipliers WHERE group_id = \$1 AND rate_multiplier IS NULL AND rpm_override IS NULL`).WithArgs(4).WillReturnResult(sqlmock.NewResult(0, 1))
			var wantErr error
			if name != "empty" {
				upsert := mock.ExpectExec(`INSERT INTO user_group_rate_multipliers .* DO UPDATE SET rate_multiplier = EXCLUDED.rate_multiplier, updated_at = EXCLUDED.updated_at`).WithArgs(4, sqlmock.AnyArg(), "{7}", "{0}")
				if name == "upsert failure" {
					wantErr = errors.New("upsert failed")
					upsert.WillReturnError(wantErr)
				} else {
					upsert.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			if wantErr != nil {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			require.ErrorIs(t, repo.SyncGroupRateMultipliers(context.Background(), 4, input), wantErr)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSyncGroupRPMOverridesTransaction(t *testing.T) {
	for _, name := range []string{"empty", "last nil clears and zero wins", "delete failure"} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := NewUserGroupRateRepository(db)
			var input []service.GroupRPMOverrideInput
			keep := "{}"
			if name != "empty" {
				ten, zero := 10, 0
				input = []service.GroupRPMOverrideInput{{UserID: 6, RPMOverride: &ten}, {UserID: 7}, {UserID: 7, RPMOverride: &zero}, {UserID: 6}}
				keep = "{7}"
			}
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE user_group_rate_multipliers SET rpm_override = NULL, updated_at = NOW\(\) WHERE group_id = \$1 AND user_id <> ALL\(\$2\)`).WithArgs(4, keep).WillReturnResult(sqlmock.NewResult(0, 2))
			deleteRows := mock.ExpectExec(`DELETE FROM user_group_rate_multipliers WHERE group_id = \$1 AND rate_multiplier IS NULL AND rpm_override IS NULL`).WithArgs(4)
			var wantErr error
			if name == "delete failure" {
				wantErr = errors.New("delete failed")
				deleteRows.WillReturnError(wantErr)
				mock.ExpectRollback()
			} else {
				deleteRows.WillReturnResult(sqlmock.NewResult(0, 1))
				if name != "empty" {
					mock.ExpectExec(`INSERT INTO user_group_rate_multipliers .* DO UPDATE SET rpm_override = EXCLUDED.rpm_override, updated_at = EXCLUDED.updated_at`).WithArgs(4, sqlmock.AnyArg(), "{7}", "{0}").WillReturnResult(sqlmock.NewResult(0, 1))
				}
				mock.ExpectCommit()
			}
			require.ErrorIs(t, repo.SyncGroupRPMOverrides(context.Background(), 4, input), wantErr)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
