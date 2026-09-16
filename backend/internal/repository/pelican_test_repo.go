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
	e := s.Scan(&p.ID, &p.GroupID, &p.GroupName, &p.ModelID, &p.IntervalMinutes, &p.Enabled, &p.MaxResults, &p.MinChars, &p.LastRunAt, &p.NextRunAt, &p.RunningUntil, &p.RunGeneration, &p.CreatedAt, &p.UpdatedAt)
	return p, e
}

const planCols = "id,group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,last_run_at,next_run_at,running_until,run_generation,created_at,updated_at"

func (r *pelicanTestRepository) CreatePlan(c context.Context, i service.PelicanPlanInput, n string) (*service.PelicanPlan, error) {
	q := "INSERT INTO pelican_test_plans (group_id,group_name,model_id,interval_minutes,enabled,max_results,min_chars,next_run_at) VALUES ($1,$2,$3,$4,FALSE,$5,$6,NOW()) RETURNING " + planCols
	p, e := scanPelicanPlan(r.db.QueryRowContext(c, q, i.GroupID, n, i.ModelID, i.IntervalMinutes, i.MaxResults, i.MinChars))
	if e != nil {
		return nil, fmt.Errorf("create pelican plan: %w", e)
	}
	return &p, nil
}
func (r *pelicanTestRepository) UpdatePlan(c context.Context, id int64, i service.PelicanPlanInput, n string) (*service.PelicanPlan, error) {
	q := "UPDATE pelican_test_plans SET group_id=$2,group_name=$3,model_id=$4,interval_minutes=$5,enabled=$6,max_results=$7,min_chars=$8,updated_at=NOW() WHERE id=$1 AND (running_until IS NULL OR running_until<NOW()) RETURNING " + planCols
	p, e := scanPelicanPlan(r.db.QueryRowContext(c, q, id, i.GroupID, n, i.ModelID, i.IntervalMinutes, i.Enabled, i.MaxResults, i.MinChars))
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
	rows, e := r.db.QueryContext(c, "SELECT "+planCols+" FROM pelican_test_plans ORDER BY id DESC")
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
	q := "UPDATE pelican_test_plans SET running_until=$2::timestamptz + interval '30 minutes',run_generation=run_generation+1,last_run_at=$2,next_run_at=$2::timestamptz+(interval_minutes || ' minutes')::interval,updated_at=$2 WHERE id=$1 AND (running_until IS NULL OR running_until<$2) AND ($3 OR (enabled AND next_run_at <= $2)) RETURNING " + planCols
	p, e := scanPelicanPlan(r.db.QueryRowContext(c, q, id, now, manual))
	if errors.Is(e, sql.ErrNoRows) {
		return nil, service.ErrPelicanConflict
	}
	return &p, e
}
func (r *pelicanTestRepository) Release(c context.Context, id, generation int64) error {
	_, e := r.db.ExecContext(c, "UPDATE pelican_test_plans SET running_until=NULL WHERE id=$1 AND run_generation=$2", id, generation)
	return e
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
		where += " AND live_g.status='active' AND live_a.status='active' AND r.group_id=ANY($1)"
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
	q := "SELECT r.plan_id,r.group_id,live_g.name,r.account_id,r.model_id,r.status,r.latency_ms,r.char_count,r.min_chars,r.finished_at,(SELECT count(*) FROM pelican_test_results h WHERE h.plan_id=r.plan_id AND h.account_id=r.account_id AND h.group_id=r.group_id),r.id,(SELECT max(s.id) FROM pelican_test_results s WHERE s.plan_id=r.plan_id AND s.account_id=r.account_id AND s.group_id=r.group_id AND s.status='success'),r.error_message FROM (SELECT DISTINCT ON (plan_id,account_id,group_id) * FROM pelican_test_results ORDER BY plan_id,account_id,group_id,finished_at DESC,id DESC) r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups live_g ON live_g.id=r.group_id AND live_g.deleted_at IS NULL JOIN accounts live_a ON live_a.id=r.account_id AND live_a.deleted_at IS NULL JOIN account_groups ag ON ag.group_id=r.group_id AND ag.account_id=r.account_id " + where + " ORDER BY r.finished_at DESC LIMIT $" + fmt.Sprint(n) + " OFFSET $" + fmt.Sprint(n+1)
	args = append(args, size, (page-1)*size)
	rows, e := r.db.QueryContext(c, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var x service.PelicanEntry
		e = rows.Scan(&x.PlanID, &x.GroupID, &x.GroupName, &x.AccountID, &x.ModelID, &x.Status, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.FinishedAt, &x.HistoryCount, &x.ResultID, &x.ArtworkResultID, &x.ErrorMessage)
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
	countQ := "SELECT count(*) FROM (SELECT DISTINCT r.plan_id,r.account_id FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups live_g ON live_g.id=r.group_id AND live_g.deleted_at IS NULL JOIN accounts live_a ON live_a.id=r.account_id AND live_a.deleted_at IS NULL JOIN account_groups ag ON ag.group_id=r.group_id AND ag.account_id=r.account_id " + where + ") x"
	if e := r.db.QueryRowContext(c, countQ, args[:len(args)-2]...).Scan(&out.Total); e != nil {
		return nil, e
	}
	groupsQ := "SELECT DISTINCT g.id,g.name FROM groups g JOIN pelican_test_results r ON r.group_id=g.id JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN account_groups ag ON ag.group_id=g.id AND ag.account_id=r.account_id JOIN accounts a ON a.id=r.account_id AND a.deleted_at IS NULL WHERE g.deleted_at IS NULL"
	var groupArgs []any
	if !a.Admin {
		groupsQ += " AND g.status='active' AND a.status='active' AND g.id=ANY($1)"
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
	q := "SELECT r.id,r.plan_id,r.group_id,r.account_id,r.model_id,r.prompt_version,r.status,r.error_message,r.latency_ms,r.char_count,r.min_chars,r.started_at,r.finished_at FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups g ON g.id=r.group_id AND g.deleted_at IS NULL JOIN accounts a ON a.id=r.account_id AND a.deleted_at IS NULL JOIN account_groups ag ON ag.group_id=r.group_id AND ag.account_id=r.account_id WHERE 1=1"
	args := []any{}
	i := 1
	if !a.Admin {
		q += " AND g.status='active' AND a.status='active' AND r.group_id=ANY($1)"
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
		if e = rows.Scan(&x.ID, &x.PlanID, &x.GroupID, &x.AccountID, &x.ModelID, &x.PromptVersion, &x.Status, &x.ErrorMessage, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.StartedAt, &x.FinishedAt); e != nil {
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
	q := "SELECT r.id,r.plan_id,r.group_id,r.account_id,r.model_id,r.prompt_version,r.status,r.error_message,r.latency_ms,r.char_count,r.min_chars,r.started_at,r.finished_at,COALESCE(r.html,'') FROM pelican_test_results r JOIN pelican_test_plans p ON p.id=r.plan_id AND p.group_id=r.group_id JOIN groups g ON g.id=r.group_id AND g.deleted_at IS NULL JOIN accounts a ON a.id=r.account_id AND a.deleted_at IS NULL JOIN account_groups ag ON ag.group_id=r.group_id AND ag.account_id=r.account_id WHERE r.id=$1"
	args := []any{id}
	if !a.Admin {
		q += " AND g.status='active' AND a.status='active' AND r.group_id=ANY($2)"
		args = append(args, pq.Array(a.GroupIDs))
	}
	var x service.PelicanResult
	e := r.db.QueryRowContext(c, q, args...).Scan(&x.ID, &x.PlanID, &x.GroupID, &x.AccountID, &x.ModelID, &x.PromptVersion, &x.Status, &x.ErrorMessage, &x.LatencyMS, &x.CharCount, &x.MinChars, &x.StartedAt, &x.FinishedAt, &x.HTML)
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
	inserted, e := tx.ExecContext(c, "INSERT INTO pelican_test_results (plan_id,group_id,account_id,model_id,prompt_version,status,error_message,latency_ms,char_count,min_chars,run_generation,started_at,finished_at,html) SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14 WHERE EXISTS (SELECT 1 FROM pelican_test_plans WHERE id=$1 AND group_id=$2 AND run_generation=$11 AND running_until>NOW())", x.PlanID, x.GroupID, x.AccountID, x.ModelID, x.PromptVersion, x.Status, strings.TrimSpace(x.ErrorMessage), x.LatencyMS, x.CharCount, x.MinChars, generation, x.StartedAt, x.FinishedAt, x.HTML)
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
	_, e = tx.ExecContext(c, "DELETE FROM pelican_test_results WHERE id IN (WITH ranked AS (SELECT id, row_number() OVER (ORDER BY started_at DESC,id DESC) AS rn, max(id) FILTER (WHERE status='success') OVER () AS last_success FROM pelican_test_results WHERE plan_id=$1 AND account_id=$2 AND group_id=$3) SELECT id FROM ranked WHERE rn > $4 AND id IS DISTINCT FROM last_success)", x.PlanID, x.AccountID, x.GroupID, max)
	if e != nil {
		return e
	}
	return tx.Commit()
}
