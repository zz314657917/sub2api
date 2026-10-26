package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

func pelicanPreferredRow(ids string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(strings.Split(planCols, ",")).AddRow(
		int64(7), int64(3), "group", "m", 15, true, 20, 100,
		nil, now, nil, int64(3), 0, 0, 0, nil, 0, 0, "", 0, "", 180, now, now, "{m}", ids,
	)
}

func TestPelicanPreferredPersistence(t *testing.T) {
	for index, ids := range [][]int64{{9, 2, 8}, {}, nil} {
		for _, operation := range []string{"create", "update", "list", "claim", "resume"} {
			t.Run(fmt.Sprintf("%s/%d", operation, index), func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				array := "{}"
				if len(ids) > 0 {
					array = "{9,2,8}"
				}
				input := service.PelicanPlanInput{GroupID: 3, ModelID: "m", IntervalMinutes: 15, MaxResults: 20, MinChars: 100, PreferredAccountIDs: ids}
				repo := NewPelicanTestRepository(db)
				var plan *service.PelicanPlan
				switch operation {
				case "create":
					mock.ExpectQuery("INSERT INTO pelican_test_plans .*preferred_account_ids").
						WithArgs(int64(3), "group", "m", 15, 20, 100, 0, 0, 0, "", 0, `{"m"}`, array).
						WillReturnRows(pelicanPreferredRow(array))
					plan, err = repo.CreatePlan(context.Background(), input, "group")
				case "update":
					mock.ExpectQuery("UPDATE pelican_test_plans SET .*preferred_account_ids=\\$15").
						WithArgs(int64(7), int64(3), "group", "m", 15, false, 20, 100, 0, 0, 0, "", 0, `{"m"}`, array).
						WillReturnRows(pelicanPreferredRow(array))
					plan, err = repo.UpdatePlan(context.Background(), 7, input, "group")
				case "list":
					mock.ExpectQuery("SELECT .*preferred_account_ids FROM pelican_test_plans").WillReturnRows(pelicanPreferredRow(array))
					var plans []service.PelicanPlan
					plans, err = repo.ListPlans(context.Background())
					if len(plans) == 1 {
						plan = &plans[0]
					}
				case "claim":
					now := time.Now()
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .*preferred_account_ids FROM pelican_test_plans .*FOR UPDATE").
						WithArgs(int64(7)).WillReturnRows(pelicanPreferredRow(array))
					mock.ExpectQuery("SELECT COALESCE").WithArgs(int64(7)).
						WillReturnRows(sqlmock.NewRows([]string{"exhausted"}).AddRow(false))
					mock.ExpectQuery("UPDATE pelican_test_plans SET running_until=.*RETURNING .*preferred_account_ids").
						WithArgs(int64(7), now).WillReturnRows(pelicanPreferredRow(array))
					mock.ExpectCommit()
					plan, err = repo.Claim(context.Background(), 7, true, now)
				case "resume":
					mock.ExpectQuery("UPDATE pelican_test_plans SET enabled=TRUE.*RETURNING .*preferred_account_ids").
						WithArgs(int64(7)).WillReturnRows(pelicanPreferredRow(array))
					plan, err = repo.Resume(context.Background(), 7)
				}
				want := ids
				if want == nil {
					want = []int64{}
				}
				if err != nil || plan == nil || !reflect.DeepEqual(plan.PreferredAccountIDs, want) {
					t.Fatalf("plan=%+v want=%v err=%v", plan, want, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestPelicanPreferredNilNeverSerializesNULL(t *testing.T) {
	// Assert the SQL driver value directly as well as through the queries above.
	value, err := pq.Array(pelicanInputPreferredAccounts(service.PelicanPlanInput{})).Value()
	if err != nil || value != "{}" {
		t.Fatalf("value=%v err=%v", value, err)
	}
}
