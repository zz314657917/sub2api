package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"strings"
	"time"
)

type pelicanTestRepository struct{ db *sql.DB }

func NewPelicanTestRepository(db *sql.DB) service.PelicanTestRepository {
	return &pelicanTestRepository{db: db}
}
func scanPelicanPlan(s interface{ Scan(...any) error }) (service.PelicanPlan, error) {
	var p service.PelicanPlan
	var usageDay sql.NullTime
	e := s.Scan(&p.ID, &p.GroupID, &p.GroupName, &p.ModelID, &p.IntervalMinutes, &p.Enabled, &p.MaxResults, &p.MinChars, &p.LastRunAt, &p.NextRunAt, &p.RunningUntil, &p.RunGeneration, &p.DailyCallLimit, &p.FailurePauseThreshold, &p.RetentionDays, &usageDay, &p.DailyCallsUsed, &p.ConsecutiveFailedRuns, &p.PauseReason, &p.LastRunCalls, &p.ReasoningEffort, &p.TimeoutSeconds, &p.CreatedAt, &p.UpdatedAt)
	if usageDay.Valid {
		p.UsageDay = usageDay.Time.Format("2006-01-02")
	}
	return p, e
}

const planCols = "id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,last_run_at,next_run_at,running_until,run_generation,daily_call_limit,failure_pause_threshold,retention_days,usage_day,daily_calls_used,consecutive_failed_runs,pause_reason,last_run_calls,reasoning_effort,timeout_seconds,created_at,updated_at"
const planListCols = "id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,last_run_at,next_run_at,running_until,run_generation,daily_call_limit,failure_pause_threshold,retention_days,(NOW() AT TIME ZONE 'Asia/Shanghai')::date AS usage_day,CASE WHEN usage_day IS NOT DISTINCT FROM (NOW() AT TIME ZONE 'Asia/Shanghai')::date THEN daily_calls_used ELSE 0 END AS daily_calls_used,consecutive_failed_runs,pause_reason,last_run_calls,reasoning_effort,timeout_seconds,created_at,updated_at"

func (r *pelicanTestRepository) CreatePlan(c context.Context, i service.PelicanPlanInput, n string) (*service.PelicanPlan, error) {
	q := "INSERT INTO pelican_test_plans (group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,daily_call_limit,failure_pause_threshold,retention_days,reasoning_effort,timeout_seconds,next_run_at) VALUES ($1,$2,$3,$4,FALSE,$5,$6,$7,$8,$9,$10,COALESCE(NULLIF($11,0),180),NOW()) RETURNING " + planListCols
	p, e := scanPelicanPlan(r.db.QueryRowContext(c, q, i.GroupID, n, i.ModelID, i.IntervalMinutes, i.MaxResults, i.MinChars, i.DailyCallLimit, i.FailurePauseThreshold, i.RetentionDays, i.ReasoningEffort, i.TimeoutSeconds))
	if e != nil {
		return nil, fmt.Errorf("create pelican plan: %w", e)
	}
	return &p, nil
}
func (r *pelicanTestRepository) UpdatePlan(c context.Context, id int64, i service.PelicanPlanInput, n string) (*service.PelicanPlan, error) {
	q := "UPDATE pelican_test_plans SET group_id=$2,group_name=$3,model_id=$4,interval_minutes=$5,enabled=CASE WHEN pause_reason='' THEN $6 ELSE FALSE END,max_results=$7,min_chars=$8,daily_call_limit=$9,failure_pause_threshold=$10,retention_days=$11,reasoning_effort=$12,timeout_seconds=COALESCE(NULLIF($13,0),180),updated_at=NOW() WHERE id=$1 AND (running_until IS NULL OR running_until<NOW()) RETURNING " + planListCols
	p, e := scanPelicanPlan(r.db.QueryRowContext(c, q, id, i.GroupID, n, i.ModelID, i.IntervalMinutes, i.Enabled, i.MaxResults, i.MinChars, i.DailyCallLimit, i.FailurePauseThreshold, i.RetentionDays, i.ReasoningEffort, i.TimeoutSeconds))
	if errors.Is(e, sql.ErrNoRows) {
		return nil, service.ErrPelicanConflict
	}
	return &p, e
}
func (r *pelicanTestRepository) DeletePlan(c context.Context, id int64) error {
	x, e := r.db.ExecContext(c, "DELETE FROM pelican_test_plans WHERE id=$1 AND (running_until IS NULL OR running_until<NOW())", id)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return service.ErrPelicanConflict
	}
	return nil
}
func (r *pelicanTestRepository) ListPlans(c context.Context) ([]service.PelicanPlan, error) {
	rows, e := r.db.QueryContext(c, "SELECT "+planListCols+" FROM pelican_test_plans ORDER BY id DESC")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]service.PelicanPlan, 0)
	for rows.Next() {
		p, e := scanPelicanPlan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *pelicanTestRepository) Claim(c context.Context, id int64, manual bool, now time.Time) (*service.PelicanPlan, error) {
	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	p, err := scanPelicanPlan(tx.QueryRowContext(c, "SELECT "+planCols+" FROM pelican_test_plans WHERE id=$1 FOR UPDATE", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrPelicanConflict
	}
	if err != nil {
		return nil, err
	}
	if p.RunningUntil != nil && !p.RunningUntil.Before(now) {
		return nil, service.ErrPelicanConflict
	}
	staleFinalized := false
	if p.RunningUntil != nil { // crash-safe finalization of an abandoned matching generation
		if _, err = tx.ExecContext(c, `UPDATE pelican_test_plans SET running_until=NULL,last_run_calls=round_attempted,
		 consecutive_failed_runs=CASE WHEN round_attempted=0 THEN consecutive_failed_runs WHEN round_successes>0 THEN 0 ELSE consecutive_failed_runs+1 END,
		 enabled=CASE WHEN round_attempted>0 AND round_successes=0 AND failure_pause_threshold>0 AND consecutive_failed_runs+1>=failure_pause_threshold THEN FALSE ELSE enabled END,
		 pause_reason=CASE WHEN round_attempted>0 AND round_successes=0 AND failure_pause_threshold>0 AND consecutive_failed_runs+1>=failure_pause_threshold THEN 'consecutive_failures' ELSE pause_reason END WHERE id=$1`, id); err != nil {
			return nil, err
		}
		p, err = scanPelicanPlan(tx.QueryRowContext(c, "SELECT "+planCols+" FROM pelican_test_plans WHERE id=$1", id))
		if err != nil {
			return nil, err
		}
		staleFinalized = true
	}
	if p.PauseReason != "" {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return nil, service.ErrPelicanPaused
	}
	if !manual && (!p.Enabled || p.NextRunAt == nil || p.NextRunAt.After(now)) {
		if staleFinalized {
			if err = tx.Commit(); err != nil {
				return nil, err
			}
		}
		return nil, service.ErrPelicanConflict
	}
	var exhausted bool
	if err = tx.QueryRowContext(c, "SELECT COALESCE(daily_call_limit>0 AND usage_day=(NOW() AT TIME ZONE 'Asia/Shanghai')::date AND daily_calls_used>=daily_call_limit,FALSE) FROM pelican_test_plans WHERE id=$1", id).Scan(&exhausted); err != nil {
		return nil, err
	}
	if exhausted {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return nil, service.ErrPelicanQuota
	}
	p, err = scanPelicanPlan(tx.QueryRowContext(c, "UPDATE pelican_test_plans SET running_until=$2::timestamptz + (timeout_seconds + 120) * interval '1 second',run_generation=run_generation+1,round_attempted=0,round_successes=0,last_run_at=$2,next_run_at=$2::timestamptz+(interval_minutes || ' minutes')::interval,updated_at=$2 WHERE id=$1 RETURNING "+planListCols, id, now))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *pelicanTestRepository) Finalize(c context.Context, id, generation int64) error {
	_, e := r.db.ExecContext(c, `UPDATE pelican_test_plans SET running_until=NULL,last_run_calls=round_attempted,
	 consecutive_failed_runs=CASE WHEN round_attempted=0 THEN consecutive_failed_runs WHEN round_successes>0 THEN 0 ELSE consecutive_failed_runs+1 END,
	 enabled=CASE WHEN round_attempted>0 AND round_successes=0 AND failure_pause_threshold>0 AND consecutive_failed_runs+1>=failure_pause_threshold THEN FALSE ELSE enabled END,
	 pause_reason=CASE WHEN round_attempted>0 AND round_successes=0 AND failure_pause_threshold>0 AND consecutive_failed_runs+1>=failure_pause_threshold THEN 'consecutive_failures' ELSE pause_reason END,
	 updated_at=NOW() WHERE id=$1 AND run_generation=$2 AND running_until IS NOT NULL`, id, generation)
	return e
}
func (r *pelicanTestRepository) Release(c context.Context, id, generation int64) error {
	return r.Finalize(c, id, generation)
}
func (r *pelicanTestRepository) ReserveAttempt(ctx context.Context, id, generation int64) error {
	q := `UPDATE pelican_test_plans SET usage_day=(NOW() AT TIME ZONE 'Asia/Shanghai')::date,
	 daily_calls_used=CASE WHEN usage_day IS DISTINCT FROM (NOW() AT TIME ZONE 'Asia/Shanghai')::date THEN 1 ELSE daily_calls_used+1 END,
	 round_attempted=round_attempted+1,updated_at=NOW()
	 WHERE id=$1 AND run_generation=$2 AND running_until>NOW() AND pause_reason=''
	 AND (daily_call_limit=0 OR CASE WHEN usage_day IS DISTINCT FROM (NOW() AT TIME ZONE 'Asia/Shanghai')::date THEN 0 ELSE daily_calls_used END < daily_call_limit)`
	x, err := r.db.ExecContext(ctx, q, id, generation)
	if err != nil {
		return err
	}
	n, err := x.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return service.ErrPelicanQuota
	}
	return nil
}
func (r *pelicanTestRepository) Resume(ctx context.Context, id int64) (*service.PelicanPlan, error) {
	p, err := scanPelicanPlan(r.db.QueryRowContext(ctx, "UPDATE pelican_test_plans SET enabled=TRUE,pause_reason='',consecutive_failed_runs=0,updated_at=NOW() WHERE id=$1 AND pause_reason='consecutive_failures' AND (running_until IS NULL OR running_until<NOW()) RETURNING "+planListCols, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrPelicanConflict
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *pelicanTestRepository) CanRunAccount(ctx context.Context, planID, generation, accountID int64) (bool, error) {
	var allowed bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM pelican_test_plans p
		JOIN groups g ON g.id=p.group_id AND g.deleted_at IS NULL AND g.status='active' AND g.platform='openai'
		JOIN account_groups ag ON ag.group_id=g.id AND ag.account_id=$3
		JOIN accounts a ON a.id=ag.account_id AND a.deleted_at IS NULL AND a.status='active'
		WHERE p.id=$1 AND p.run_generation=$2 AND p.running_until>NOW()
	)`, planID, generation, accountID).Scan(&allowed)
	return allowed, err
}
func authClause(a service.PelicanAuthorization, col string) (string, []any) {
	if a.Admin {
		return "", nil
	}
	if len(a.GroupIDs) == 0 {
		return " AND FALSE", nil
	}
	return " AND " + col + " = ANY($1)", []any{pq.Array(a.GroupIDs)}
}
func (r *pelicanTestRepository) ListGallery(c context.Context, a service.PelicanAuthorization, g, ac int64, page, size int) (*service.PelicanListResponse, error) {
	out := &service.PelicanListResponse{Items: make([]service.PelicanEntry, 0), Groups: make([]service.PelicanGroup, 0), Page: page, PageSize: size}
	if !a.Admin && len(a.GroupIDs) == 0 {
		return out, nil
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 24
	}
	out.Page = page
	out.PageSize = size
	where := " WHERE 1=1"
	args := []any{}
	n := 1
	if !a.Admin {
		where += " AND live_g.status='active' AND r.group_id=ANY($1)"
		args = append(args, pq.Array(a.GroupIDs))
		n++
	}
	if g > 0 {
		where += fmt.Sprintf(" AND r.group_id=$%d", n)
		args = append(args, g)
		n++
	}
	if ac > 0 {
		where += fmt.Sprintf(" AND r.account_id=$%d", n)
		args = append(args, ac)
		n++
	}
	latest := "(SELECT DISTINCT ON (plan_id,group_id) * FROM pelican_test_results ORDER BY plan_id,group_id,finished_at DESC,id DESC) r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups live_g ON live_g.id=r.group_id AND live_g.deleted_at IS NULL "
	q := "SELECT r.plan_id,r.group_id,live_g.name,r.account_id,r.model_id,r.status,r.latency_ms,r.char_count,r.min_chars,r.finished_at,(SELECT count(*) FROM pelican_test_results h WHERE h.plan_id=r.plan_id AND h.group_id=r.group_id),r.id,(SELECT s.id FROM pelican_test_results s WHERE s.plan_id=r.plan_id AND s.group_id=r.group_id AND s.status='success' ORDER BY s.finished_at DESC,s.id DESC LIMIT 1),r.error_message,r.error_code,r.error_message_safe,r.reasoning_effort FROM " + latest + where + " ORDER BY r.finished_at DESC,r.id DESC LIMIT $" + fmt.Sprint(n) + " OFFSET $" + fmt.Sprint(n+1)
	args = append(args, size, (page-1)*size)
	rows, e := r.db.QueryContext(c, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var x service.PelicanEntry
		e = rows.Scan(&x.PlanID, &x.GroupID, &x.GroupName, &x.AccountID, &x.ModelID, &x.Status, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.FinishedAt, &x.HistoryCount, &x.ResultID, &x.ArtworkResultID, &x.ErrorMessage, &x.ErrorCode, &x.ErrorMessageSafe, &x.ReasoningEffort)
		if e != nil {
			return nil, e
		}
		out.Items = append(out.Items, x)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if e = rows.Close(); e != nil {
		return nil, e
	}
	countQ := "SELECT count(*) FROM " + latest + where
	if e := r.db.QueryRowContext(c, countQ, args[:len(args)-2]...).Scan(&out.Total); e != nil {
		return nil, e
	}
	groupsQ := "SELECT DISTINCT g.id,g.name FROM groups g JOIN pelican_test_results r ON r.group_id=g.id JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id WHERE g.deleted_at IS NULL"
	var groupArgs []any
	if !a.Admin {
		groupsQ += " AND g.status='active' AND g.id=ANY($1)"
		groupArgs = append(groupArgs, pq.Array(a.GroupIDs))
	}
	groupRows, e := r.db.QueryContext(c, groupsQ+" ORDER BY g.name,g.id", groupArgs...)
	if e != nil {
		return nil, e
	}
	defer groupRows.Close()
	for groupRows.Next() {
		var group service.PelicanGroup
		if e = groupRows.Scan(&group.ID, &group.Name); e != nil {
			return nil, e
		}
		out.Groups = append(out.Groups, group)
	}
	return out, groupRows.Err()
}
func (r *pelicanTestRepository) ListHistory(c context.Context, a service.PelicanAuthorization, p, ac int64) ([]service.PelicanResult, error) {
	if !a.Admin && len(a.GroupIDs) == 0 {
		return []service.PelicanResult{}, nil
	}
	q := "SELECT r.id,r.plan_id,r.group_id,r.account_id,r.model_id,r.prompt_version,r.status,r.error_message,r.error_code,r.error_message_safe,r.latency_ms,r.char_count,r.min_chars,r.started_at,r.finished_at,r.reasoning_effort FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups g ON g.id=r.group_id AND g.deleted_at IS NULL WHERE 1=1"
	args := []any{}
	i := 1
	if !a.Admin {
		q += " AND g.status='active' AND r.group_id=ANY($1)"
		args = append(args, pq.Array(a.GroupIDs))
		i++
	}
	if p > 0 {
		q += fmt.Sprintf(" AND r.plan_id=$%d", i)
		args = append(args, p)
		i++
	}
	if ac > 0 {
		q += fmt.Sprintf(" AND r.account_id=$%d", i)
		args = append(args, ac)
		i++
	}
	q += " ORDER BY r.started_at DESC,r.id DESC LIMIT 50"
	rows, e := r.db.QueryContext(c, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]service.PelicanResult, 0)
	for rows.Next() {
		var x service.PelicanResult
		if e = rows.Scan(&x.ID, &x.PlanID, &x.GroupID, &x.AccountID, &x.ModelID, &x.PromptVersion, &x.Status, &x.ErrorMessage, &x.ErrorCode, &x.ErrorMessageSafe, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.StartedAt, &x.FinishedAt, &x.ReasoningEffort); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *pelicanTestRepository) GetResult(c context.Context, a service.PelicanAuthorization, id int64) (*service.PelicanResult, error) {
	if !a.Admin && len(a.GroupIDs) == 0 {
		return nil, service.ErrPelicanNotFound
	}
	q := "SELECT r.id,r.plan_id,r.group_id,r.account_id,r.model_id,r.prompt_version,r.status,r.error_message,r.error_code,r.error_message_safe,r.latency_ms,r.char_count,r.min_chars,r.started_at,r.finished_at,COALESCE(r.html,''),r.reasoning_effort FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups g ON g.id=r.group_id AND g.deleted_at IS NULL WHERE r.id=$1"
	args := []any{id}
	if !a.Admin {
		q += " AND g.status='active' AND r.group_id=ANY($2)"
		args = append(args, pq.Array(a.GroupIDs))
	}
	var x service.PelicanResult
	e := r.db.QueryRowContext(c, q, args...).Scan(&x.ID, &x.PlanID, &x.GroupID, &x.AccountID, &x.ModelID, &x.PromptVersion, &x.Status, &x.ErrorMessage, &x.ErrorCode, &x.ErrorMessageSafe, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.StartedAt, &x.FinishedAt, &x.HTML, &x.ReasoningEffort)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, service.ErrPelicanNotFound
	}
	if e != nil {
		return nil, e
	}
	return &x, nil
}
func (r *pelicanTestRepository) SaveResult(c context.Context, x *service.PelicanResult, max int, generation int64) error {
	if x == nil {
		return errors.New("nil pelican result")
	}
	if len(x.HTML) > 256*1024 {
		return errors.New("pelican HTML exceeds limit")
	}
	if x.FinishedAt == nil {
		n := time.Now()
		x.FinishedAt = &n
	}
	tx, e := r.db.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer func() { _ = tx.Rollback() }()
	var planID int64
	// Hold the plan row through pruning so a newer claim cannot overtake
	// an older generation's transaction after its initial fence check.
	e = tx.QueryRowContext(c, "SELECT id FROM pelican_test_plans WHERE id=$1 AND group_id=$2 AND run_generation=$3 AND running_until>NOW() FOR UPDATE", x.PlanID, x.GroupID, generation).Scan(&planID)
	if errors.Is(e, sql.ErrNoRows) {
		return service.ErrPelicanConflict
	}
	if e != nil {
		return e
	}
	inserted, e := tx.ExecContext(c, "INSERT INTO pelican_test_results (plan_id,group_id,account_id,model_id,prompt_version,status,error_message,error_code,error_message_safe,latency_ms,char_count,min_chars,run_generation,started_at,finished_at,html,reasoning_effort) SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17 WHERE EXISTS (SELECT 1 FROM pelican_test_plans WHERE id=$1 AND group_id=$2 AND run_generation=$13 AND running_until>NOW())", x.PlanID, x.GroupID, x.AccountID, x.ModelID, x.PromptVersion, x.Status, strings.TrimSpace(x.ErrorMessage), strings.TrimSpace(x.ErrorCode), strings.TrimSpace(x.ErrorMessageSafe), x.LatencyMS, x.CharCount, x.MinChars, generation, x.StartedAt, x.FinishedAt, x.HTML, x.ReasoningEffort)
	if e != nil {
		return e
	}
	n, e := inserted.RowsAffected()
	if e != nil {
		return e
	}
	if n == 0 {
		return service.ErrPelicanConflict
	}
	if x.Status == "success" {
		if _, e = tx.ExecContext(c, "UPDATE pelican_test_plans SET round_successes=round_successes+1 WHERE id=$1 AND run_generation=$2 AND running_until>NOW()", x.PlanID, generation); e != nil {
			return e
		}
	}
	_, e = tx.ExecContext(c, `DELETE FROM pelican_test_results r WHERE r.plan_id=$1 AND r.group_id=$2
	 AND (SELECT count(*) FROM pelican_test_results newer WHERE newer.plan_id=r.plan_id AND newer.group_id=r.group_id AND (newer.started_at,newer.id)>(r.started_at,r.id)) >= $3
	 AND NOT (r.status='success' AND NOT EXISTS (SELECT 1 FROM pelican_test_results newer_success WHERE newer_success.plan_id=r.plan_id AND newer_success.group_id=r.group_id AND newer_success.status='success' AND (newer_success.started_at,newer_success.id)>(r.started_at,r.id)))`, x.PlanID, x.GroupID, max)
	if e != nil {
		return e
	}
	return tx.Commit()
}

func (r *pelicanTestRepository) Cleanup(ctx context.Context, planID int64, scope string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var active bool
	if err = tx.QueryRowContext(ctx, "SELECT COALESCE(running_until>NOW(),FALSE) FROM pelican_test_plans WHERE id=$1 FOR UPDATE", planID).Scan(&active); errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrPelicanNotFound
	}
	if err != nil {
		return 0, err
	}
	if active {
		return 0, service.ErrPelicanConflict
	}
	var result sql.Result
	if scope == "failed" {
		result, err = tx.ExecContext(ctx, "DELETE FROM pelican_test_results WHERE plan_id=$1 AND status IN ('failed','skipped')", planID)
	} else {
		result, err = tx.ExecContext(ctx, cleanupSQL("AND plan_id=$1"), planID)
	}
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

func cleanupSQL(extra string) string {
	// Latest successful result per plan/group is protected even when it is
	// outside the rolling retention window or older than max_results.
	return `DELETE FROM pelican_test_results WHERE id IN (
	 WITH ranked AS (
	  SELECT r.id,r.plan_id,r.account_id,r.group_id,r.status,r.started_at,p.max_results,p.retention_days,
	   row_number() OVER (PARTITION BY r.plan_id,r.group_id ORDER BY r.started_at DESC,r.id DESC) rn,
	   row_number() OVER (PARTITION BY r.plan_id,r.group_id ORDER BY CASE WHEN r.status='success' THEN 0 ELSE 1 END,r.started_at DESC,r.id DESC) success_rank
	  FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id ` + extra + `
	 ) SELECT id FROM ranked WHERE (rn>max_results OR (retention_days>0 AND started_at < NOW()-(retention_days || ' days')::interval))
	  AND NOT (status='success' AND success_rank=1)
	)`
}
func (r *pelicanTestRepository) CleanupDue(ctx context.Context) error {
	// Each Cleanup call locks its plan row. A plan claimed after this inventory
	// simply reports conflict and is deferred to the next sweep.
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM pelican_test_plans ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = r.Cleanup(ctx, id, "expired"); err != nil && !errors.Is(err, service.ErrPelicanConflict) && !errors.Is(err, service.ErrPelicanNotFound) {
			return err
		}
	}
	return nil
}
