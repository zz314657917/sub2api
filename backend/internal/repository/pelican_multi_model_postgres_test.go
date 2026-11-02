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
	p4 := plan(2, []string{"gpt-d", "gpt-e"})
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
	p4, e = repo.Claim(c, p4.ID, true, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	completedD := put(p4.ID, 2, 21, "gpt-d", "success", now)
	exec(`UPDATE pelican_test_results SET run_generation=$1 WHERE id=$2`, p4.RunGeneration, completedD)
	for page, wantGroup := range []int64{2, 1} {
		paged, err := repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 0, 0, page+1, 1)
		wantItems := 3
		if wantGroup == 1 {
			wantItems = 4
		}
		if err != nil {
			t.Fatal(err)
		}
		if paged.Total != 2 || len(paged.Items) != wantItems {
			t.Fatalf("group pagination page %d: %+v", page+1, paged)
		}
		for _, item := range paged.Items {
			if item.GroupID != wantGroup {
				t.Fatal("retained rounds split across group pages")
			}
		}
	}
	// Priorities apply before pagination, ignore missing IDs and preserve access.
	for page, wantGroup := range []int64{1, 2} {
		paged, err := repo.ListGalleryWithGroupOrder(c, service.PelicanAuthorization{Admin: true}, 0, 0, page+1, 1, []int64{999, 1})
		if err != nil || paged.Total != 2 || len(paged.Items) == 0 {
			t.Fatalf("ordered pagination: %+v %v", paged, err)
		}
		for _, row := range paged.Items {
			if row.GroupID != wantGroup {
				t.Fatalf("priority applied after pagination: page=%d group=%d", page+1, row.GroupID)
			}
		}
	}
	ordered, err := repo.ListGalleryWithGroupOrder(c, service.PelicanAuthorization{Admin: true}, 0, 0, 1, 24, []int64{1})
	if err != nil || ordered.Items[0].GroupID != 1 || ordered.Items[len(ordered.Items)-1].GroupID != 2 {
		t.Fatalf("final SELECT lost group priority: %+v %v", ordered, err)
	}
	ordered, err = repo.ListGalleryWithGroupOrder(c, service.PelicanAuthorization{GroupIDs: []int64{2}}, 0, 0, 1, 1, []int64{1, 999, 2})
	if err != nil || ordered.Total != 1 || len(ordered.Groups) != 1 {
		t.Fatalf("ordering changed authorization: %+v %v", ordered, err)
	}
	for _, row := range ordered.Items {
		if row.GroupID != 2 {
			t.Fatal("priority leaked unauthorized group")
		}
	}
	ordered, err = repo.ListGalleryWithGroupOrder(c, service.PelicanAuthorization{Admin: true}, 2, 20, 1, 1, []int64{1})
	if err != nil || ordered.Total != 1 || len(ordered.Items) != 1 || ordered.Items[0].GroupID != 2 || ordered.Items[0].AccountID != 20 {
		t.Fatalf("ordering changed group/account filters: %+v %v", ordered, err)
	}
	// Select group one explicitly: the newly claimed group two now sorts first.
	gallery, e := repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 1, 0, 1, 1)
	if e != nil || gallery.Total != 1 || len(gallery.Items) != 4 {
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
			if (x.Status == "success" && x.RoundNumber != 1) || (x.Status == "failed" && x.RoundNumber != 2) {
				t.Fatalf("legacy round numbers=%+v", x)
			}
		}
	}
	gallery, e = repo.ListGallery(c, service.PelicanAuthorization{GroupIDs: []int64{2}}, 0, 0, 1, 1)
	if e != nil || gallery.Total != 1 || len(gallery.Items) != 3 || gallery.Items[0].GroupID != 2 {
		t.Fatalf("authorization=%+v %v", gallery, e)
	}
	var completedExistingFound, runningPlaceholderFound bool
	for _, item := range gallery.Items {
		if item.PlanID == p4.ID && item.ModelID == "gpt-d" {
			completedExistingFound = item.Status == "success" && item.AccountID == 21 && item.ResultID == completedD && item.RoundNumber == p4.RunGeneration
		}
		if item.PlanID == p4.ID && item.ModelID == "gpt-e" {
			runningPlaceholderFound = item.Status == "running" && item.AccountID == 0 && item.ResultID == 0 && item.RoundNumber == p4.RunGeneration
		}
	}
	if !completedExistingFound || !runningPlaceholderFound {
		t.Fatalf("missing first-run running placeholder: %+v", gallery.Items)
	}
	gallery, e = repo.ListGallery(c, service.PelicanAuthorization{Admin: true}, 0, 10, 1, 1)
	if e != nil || len(gallery.Items) != 3 {
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
	if e != nil || len(gallery.Items) != 3 {
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

	t.Run("retained rounds use distinct artwork and bounded model rails", func(t *testing.T) {
		multi := plan(9, []string{"gpt-a", "gpt-b"})
		exec(`UPDATE pelican_test_plans SET max_results=5,run_generation=30 WHERE id=$1`, multi.ID)
		var aIDs []int64
		for j := 0; j < 7; j++ {
			id := put(multi.ID, 9, 90, "gpt-a", "success", now.Add(time.Duration(j)*time.Minute))
			exec(`UPDATE pelican_test_results SET run_generation=$1 WHERE id=$2`, 11+j, id)
			aIDs = append(aIDs, id)
		}
		for j := 0; j < 3; j++ {
			id := put(multi.ID, 9, 91, "gpt-b", "success", now.Add(time.Duration(j)*time.Minute))
			exec(`UPDATE pelican_test_results SET run_generation=$1 WHERE id=$2`, 20+j, id)
		}
		load := func(auth service.PelicanAuthorization, account int64) *service.PelicanListResponse {
			t.Helper()
			got, err := repo.ListGallery(c, auth, 9, account, 1, 1)
			if err != nil {
				t.Fatal(err)
			}
			return got
		}
		admin := service.PelicanAuthorization{Admin: true}
		got := load(admin, 0)
		if got.Total != 1 || len(got.Items) != 8 {
			t.Fatalf("five retained rounds plus three of second model: %+v", got)
		}
		for j, item := range got.Items[:5] {
			wantID := aIDs[6-j]
			if item.ModelID != "gpt-a" || item.ResultID != wantID ||
				item.ArtworkResultID == nil || *item.ArtworkResultID != wantID ||
				item.RoundNumber != int64(17-j) || item.HistoryCount != 7 ||
				item.ArtworkFinishedAt == nil || !item.ArtworkFinishedAt.Equal(*item.FinishedAt) {
				t.Fatalf("round %d must retain its own artwork/time: %+v", j, item)
			}
		}
		if len(load(admin, 90).Items) != 5 || len(load(service.PelicanAuthorization{GroupIDs: []int64{1}}, 0).Items) != 0 {
			t.Fatal("account filter or authorization regressed")
		}
		skipped := put(multi.ID, 9, 90, "gpt-a", "skipped", now.Add(8*time.Minute))
		failed := put(multi.ID, 9, 90, "gpt-a", "failed", now.Add(9*time.Minute))
		got = load(admin, 0)
		if got.Items[0].ResultID != failed || got.Items[0].ArtworkResultID == nil || *got.Items[0].ArtworkResultID != aIDs[6] ||
			got.Items[1].ResultID != skipped || got.Items[1].ArtworkResultID != nil {
			t.Fatal("only latest failure may fall back to previous success")
		}
		running, err := repo.Claim(c, multi.ID, true, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		done := put(multi.ID, 9, 91, "gpt-b", "success", now.Add(10*time.Minute))
		exec(`UPDATE pelican_test_results SET run_generation=$1 WHERE id=$2`, running.RunGeneration, done)
		got = load(admin, 0)
		runningCount := 0
		for _, item := range got.Items {
			if item.Status == "running" {
				runningCount++
				if item.ModelID != "gpt-a" || item.ResultID != 0 || item.RoundNumber != running.RunGeneration {
					t.Fatalf("incorrect running placeholder: %+v", item)
				}
			}
			if item.ResultID == done && (item.Status != "success" || item.RoundNumber != running.RunGeneration) {
				t.Fatalf("completed model marked running: %+v", item)
			}
		}
		if runningCount != 1 || len(got.Items) != 10 {
			t.Fatalf("one pending model plus saved rounds: %+v", got.Items)
		}
		// Even legacy plans with a higher retention value display at most 20.
		exec(`UPDATE pelican_test_plans SET running_until=NULL,max_results=50 WHERE id=$1`, multi.ID)
		for j := 0; j < 25; j++ {
			id := put(multi.ID, 9, 90, "gpt-a", "success", now.Add(time.Duration(11+j)*time.Minute))
			exec(`UPDATE pelican_test_results SET run_generation=$1 WHERE id=$2`, 32+j, id)
		}
		got = load(admin, 0)
		if len(got.Items) != 24 || got.Items[0].RoundNumber != 56 || got.Items[19].RoundNumber != 37 || got.Items[20].ModelID != "gpt-b" {
			t.Fatalf("20 round cap and configured model order: %+v", got.Items)
		}
		exec(`UPDATE pelican_test_plans SET max_results=5 WHERE id=$1`, multi.ID)
		if _, err := repo.Cleanup(c, multi.ID, "expired"); err != nil {
			t.Fatal(err)
		}
		got = load(admin, 0)
		if len(got.Items) != 9 || got.Items[0].RoundNumber != 56 || got.Items[4].RoundNumber != 52 {
			t.Fatalf("cleanup changed stable generation numbers: %+v", got.Items)
		}
		exec(`UPDATE groups SET status='disabled' WHERE id=9`)
		if len(load(service.PelicanAuthorization{GroupIDs: []int64{9}}, 0).Items) != 0 || len(load(admin, 0).Items) != 9 {
			t.Fatal("disabled group visibility regressed")
		}
		exec(`UPDATE groups SET deleted_at=NOW() WHERE id=9`)
		if len(load(admin, 0).Items) != 0 {
			t.Fatal("deleted group is visible")
		}
	})
}
