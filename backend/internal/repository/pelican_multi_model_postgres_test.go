package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPelicanMultiModelPostgres(t *testing.T) {
	dsn := os.Getenv("PELICAN_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PELICAN_TEST_POSTGRES_DSN not set; requires disposable PostgreSQL")
	}
	db, e := sql.Open("postgres", dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	c := context.Background()
	schema := fmt.Sprintf("qa_multi_%d", time.Now().UnixNano())
	exec := func(q string, args ...any) {
		t.Helper()
		if _, e := db.ExecContext(c, q, args...); e != nil {
			t.Fatal(e)
		}
	}
	exec("CREATE SCHEMA " + schema)
	defer db.ExecContext(c, "DROP SCHEMA "+schema+" CASCADE")
	exec("SET search_path TO " + schema)
	for _, f := range []string{"243_pelican_tests.sql", "244_pelican_cost_cleanup.sql", "245_pelican_reasoning_effort.sql", "247_pelican_result_reasoning_effort.sql", "248_pelican_plan_timeout.sql", "249_pelican_result_error_details.sql"} {
		b, e := os.ReadFile("../../migrations/" + f)
		if e != nil {
			t.Fatal(e)
		}
		exec(string(b))
	}
	exec(`INSERT INTO pelican_test_plans(id,group_id,model_id,interval_minutes)VALUES(99,9,'legacy',15)`)
	b, e := os.ReadFile("../../migrations/251_pelican_multi_model.sql")
	if e != nil {
		t.Fatal(e)
	}
	exec(string(b))
	preferredMigration, err := os.ReadFile("../../migrations/252_pelican_preferred_accounts.sql")
	if err != nil {
		t.Fatal(err)
	}
	exec(string(preferredMigration))
	var backfill string
	if e = db.QueryRow(`SELECT model_ids[1] FROM pelican_test_plans WHERE id=99`).Scan(&backfill); e != nil || backfill != "legacy" {
		t.Fatalf("backfill=%q %v", backfill, e)
	}
	for _, arr := range []string{`'{}'`, `ARRAY['a','a']`, `ARRAY[' ']`, `ARRAY[NULL]`, `ARRAY[' a']`} {
		if _, e = db.Exec(`UPDATE pelican_test_plans SET model_ids=` + arr + ` WHERE id=99`); e == nil {
			t.Fatalf("invalid list accepted %s", arr)
		}
	}
	exec(`CREATE TABLE groups(id BIGINT PRIMARY KEY,name TEXT,platform TEXT,status TEXT,deleted_at TIMESTAMPTZ);INSERT INTO groups VALUES(1,'one','openai','active',NULL),(2,'two','openai','active',NULL),(9,'legacy','openai','active',NULL)`)
	repo := &pelicanTestRepository{db: db}
	plan := func(g int64, ids []string) *service.PelicanPlan {
		t.Helper()
		p, e := repo.CreatePlan(c, service.PelicanPlanInput{GroupID: g, ModelID: ids[0], ModelIDs: ids, IntervalMinutes: 15, MinChars: 100, MaxResults: 2, DailyCallLimit: 2, TimeoutSeconds: 30}, "qa")
		if e != nil {
			t.Fatal(e)
		}
		return p
	}
	p := plan(1, []string{"gpt-a", "gpt-b"})
	p2 := plan(1, []string{"gpt-a"})
	p3 := plan(2, []string{"gpt-c"})
	put := func(pid, g, account int64, m, status string, at time.Time) int64 {
		t.Helper()
		var id int64
		e := db.QueryRow(`INSERT INTO pelican_test_results(plan_id,group_id,account_id,model_id,prompt_version,status,min_chars,started_at,finished_at)VALUES($1,$2,$3,$4,'qa',$5,100,$6,$6)RETURNING id`, pid, g, account, m, status, at).Scan(&id)
		if e != nil {
			t.Fatal(e)
		}
		return id
	}
	now := time.Now()
	oldA := put(p.ID, 1, 10, "gpt-a", "success", now.Add(-time.Hour))
	put(p.ID, 1, 11, "gpt-a", "failed", now)
	put(p.ID, 1, 10, "gpt-b", "success", now.Add(-time.Minute))
	put(p2.ID, 1, 10, "gpt-a", "success", now.Add(-2*time.Minute))
	put(p3.ID, 2, 20, "gpt-c", "success", now.Add(-3*time.Minute))
	gallery, e := repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 0, 0, 1, 1)
	if e != nil || gallery.Total != 2 || len(gallery.Items) != 3 {
		t.Fatalf("group paging=%+v %v", gallery, e)
	}
	for _, x := range gallery.Items {
		if x.GroupID != 1 {
			t.Fatal("group split")
		}
		if x.PlanID == p.ID && x.ModelID == "gpt-a" {
			if x.ArtworkResultID == nil || *x.ArtworkResultID != oldA || x.HistoryCount != 2 {
				t.Fatalf("fallback=%+v", x)
			}
		}
	}
	gallery, e = repo.ListGallery(c, service.PelicanAuthorization{GroupIDs: []int64{2}}, 0, 0, 1, 1)
	if e != nil || gallery.Total != 1 || len(gallery.Items) != 1 || gallery.Items[0].GroupID != 2 {
		t.Fatalf("authorization=%+v %v", gallery, e)
	}
	gallery, e = repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 0, 10, 1, 1)
	if e != nil || len(gallery.Items) != 2 {
		t.Fatalf("account filter=%+v %v", gallery, e)
	}
	for j := 0; j < 55; j++ {
		put(p.ID, 1, 10, "gpt-b", "failed", now.Add(time.Duration(j)*time.Second))
	}
	history, e := repo.ListHistoryForModel(c, service.PelicanAuthorization{Admin: true}, p.ID, 0, "gpt-a")
	if e != nil || len(history) != 2 {
		t.Fatalf("filter before limit %d %v", len(history), e)
	}
	exec(`UPDATE pelican_test_plans SET model_ids=ARRAY['gpt-b'] WHERE id=$1`, p.ID)
	history, e = repo.ListHistoryForModel(c, service.PelicanAuthorization{Admin: true}, p.ID, 0, "gpt-a")
	if e != nil || len(history) != 2 {
		t.Fatal("removed history lost")
	}
	gallery, e = repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 1, 0, 1, 1)
	if e != nil || len(gallery.Items) != 2 {
		t.Fatalf("removed gallery=%+v %v", gallery, e)
	}
	// Quota reservations are atomic, even when invoked concurrently.
	exec(`UPDATE pelican_test_plans SET model_ids=ARRAY['gpt-a','gpt-b'],max_results=2 WHERE id=$1`, p.ID)
	claimed, e := repo.Claim(c, p.ID, true, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	if time.Until(*claimed.RunningUntil) < 190*time.Second {
		t.Fatal("lease too short")
	}
	var wg sync.WaitGroup
	success := make(chan bool, 8)
	for j := 0; j < 8; j++ {
		wg.Add(1)
		go func() { defer wg.Done(); success <- repo.ReserveAttempt(c, p.ID, claimed.RunGeneration) == nil }()
	}
	wg.Wait()
	close(success)
	n := 0
	for ok := range success {
		if ok {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("quota=%d", n)
	}
	if e = repo.SaveResult(c, &service.PelicanResult{PlanID: p.ID, GroupID: 1, AccountID: 10, ModelID: "gpt-b", Status: "failed", MinChars: 100, StartedAt: now, FinishedAt: &now}, 2, claimed.RunGeneration-1); e != service.ErrPelicanConflict {
		t.Fatalf("stale write=%v", e)
	}
	if e = repo.Finalize(c, p.ID, claimed.RunGeneration); e != nil {
		t.Fatal(e)
	}
	var failures, used int
	e = db.QueryRow(`SELECT consecutive_failed_runs,daily_calls_used FROM pelican_test_plans WHERE id=$1`, p.ID).Scan(&failures, &used)
	if e != nil || failures != 1 || used != 2 {
		t.Fatalf("round count %d/%d %v", failures, used, e)
	}
	if _, e = repo.Cleanup(c, p.ID, "expired"); e != nil {
		t.Fatal(e)
	}
	// Successful rows stay protected separately for both models.
	exec(`UPDATE pelican_test_plans SET retention_days=1 WHERE id=$1`, p.ID)
	exec(`UPDATE pelican_test_results SET started_at=NOW()-interval '5 days',finished_at=NOW()-interval '5 days' WHERE plan_id=$1`, p.ID)
	if _, e = repo.Cleanup(c, p.ID, "expired"); e != nil {
		t.Fatal(e)
	}
	var protected int
	e = db.QueryRow(`SELECT count(*) FROM pelican_test_results WHERE plan_id=$1 AND status='success'`, p.ID).Scan(&protected)
	if e != nil || protected != 2 {
		t.Fatalf("protected=%d %v", protected, e)
	}
}
