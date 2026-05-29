package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type canvasRepository struct {
	db  *sql.DB
	sql sqlExecutor
}

func NewCanvasRepository(sqlDB *sql.DB) service.CanvasRepository {
	return &canvasRepository{db: sqlDB, sql: sqlDB}
}

func (r *canvasRepository) ListCanvases(ctx context.Context, userID int64, filters service.CanvasListFilters) ([]service.CanvasListItem, int, error) {
	query := `
		SELECT id, user_id, title, description, viewport, metadata, created_at, updated_at,
			(SELECT COUNT(*) FROM canvas_nodes WHERE canvas_id = canvas_documents.id) AS node_count,
			(SELECT COUNT(*) FROM canvas_edges WHERE canvas_id = canvas_documents.id) AS edge_count,
			COUNT(*) OVER() AS total
		FROM canvas_documents
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.sql.QueryContext(ctx, query, userID, filters.Limit, filters.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.CanvasListItem, 0)
	total := 0
	for rows.Next() {
		item, rowTotal, err := scanCanvasListItem(rows)
		if err != nil {
			return nil, 0, err
		}
		total = rowTotal
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *canvasRepository) GetCanvas(ctx context.Context, userID int64, canvasID int64) (*service.CanvasDocument, error) {
	doc, err := r.getCanvasHeader(ctx, r.sql, userID, canvasID)
	if err != nil {
		return nil, err
	}
	nodes, err := r.listCanvasNodes(ctx, canvasID)
	if err != nil {
		return nil, err
	}
	edges, err := r.listCanvasEdges(ctx, canvasID)
	if err != nil {
		return nil, err
	}
	doc.Nodes = nodes
	doc.Edges = edges
	return doc, nil
}

func (r *canvasRepository) SaveCanvas(ctx context.Context, userID int64, input service.CanvasSaveInput) (*service.CanvasDocument, error) {
	if r.db == nil {
		return nil, errors.New("canvas repository requires sql db")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var doc *service.CanvasDocument
	if input.ID > 0 {
		doc, err = r.updateCanvasHeader(ctx, tx, userID, input)
	} else {
		doc, err = r.insertCanvasHeader(ctx, tx, userID, input)
	}
	if err != nil {
		return nil, err
	}
	if input.ID > 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM canvas_edges WHERE canvas_id = $1 AND user_id = $2`, doc.ID, userID); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM canvas_nodes WHERE canvas_id = $1 AND user_id = $2`, doc.ID, userID); err != nil {
			return nil, err
		}
	}
	if err := insertCanvasNodes(ctx, tx, userID, doc.ID, input.Nodes); err != nil {
		return nil, err
	}
	if err := insertCanvasEdges(ctx, tx, userID, doc.ID, input.Edges); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	doc.Nodes = append([]service.CanvasNode(nil), input.Nodes...)
	doc.Edges = append([]service.CanvasEdge(nil), input.Edges...)
	return doc, nil
}

func (r *canvasRepository) DeleteCanvas(ctx context.Context, userID int64, canvasID int64) error {
	result, err := r.sql.ExecContext(ctx, `
		UPDATE canvas_documents
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, canvasID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *canvasRepository) CreateCanvasRun(ctx context.Context, userID int64, input service.CanvasRunCreateInput) (*service.CanvasRun, error) {
	inputJSON, err := marshalCanvasJSON(input.Input)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := marshalCanvasJSON(input.Metadata)
	if err != nil {
		return nil, err
	}
	query := `
		INSERT INTO canvas_runs (
			user_id, canvas_id, status, trigger_type, api_key_id, model, input, metadata
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb)
		RETURNING id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at
	`
	return scanCanvasRunRow(ctx, r.sql, query,
		userID,
		input.CanvasID,
		service.CanvasRunStatusPending,
		input.TriggerType,
		nullableCanvasInt64(input.APIKeyID),
		input.Model,
		string(inputJSON),
		string(metadataJSON),
	)
}

func (r *canvasRepository) MarkCanvasRunRunning(ctx context.Context, userID int64, runID int64) (*service.CanvasRun, error) {
	query := `
		UPDATE canvas_runs
		SET status = $3,
			started_at = COALESCE(started_at, NOW()),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status = $4
		RETURNING id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at
	`
	return scanCanvasRunRow(ctx, r.sql, query, runID, userID, service.CanvasRunStatusRunning, service.CanvasRunStatusPending)
}

func (r *canvasRepository) CompleteCanvasRun(ctx context.Context, userID int64, runID int64, input service.CanvasRunCompleteInput) (*service.CanvasRun, error) {
	outputJSON, err := marshalCanvasJSON(input.Output)
	if err != nil {
		return nil, err
	}
	query := `
		UPDATE canvas_runs
		SET status = $3,
			output = $4::jsonb,
			error_message = $5,
			completed_at = COALESCE(completed_at, NOW()),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status IN ($6, $7)
		RETURNING id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at
	`
	return scanCanvasRunRow(ctx, r.sql, query,
		runID,
		userID,
		input.Status,
		string(outputJSON),
		strings.TrimSpace(input.ErrorMessage),
		service.CanvasRunStatusPending,
		service.CanvasRunStatusRunning,
	)
}

func (r *canvasRepository) ListCanvasRuns(ctx context.Context, userID int64, filters service.CanvasRunListFilters) ([]service.CanvasRun, int, error) {
	clauses := []string{"user_id = $1"}
	args := []any{userID}
	if filters.CanvasID > 0 {
		args = append(args, filters.CanvasID)
		clauses = append(clauses, fmt.Sprintf("canvas_id = $%d", len(args)))
	}
	args = append(args, filters.Limit, filters.Offset)
	limitArg := len(args) - 1
	offsetArg := len(args)
	query := fmt.Sprintf(`
		SELECT id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at,
			COUNT(*) OVER() AS total
		FROM canvas_runs
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(clauses, " AND "), limitArg, offsetArg)
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	runs := make([]service.CanvasRun, 0)
	total := 0
	for rows.Next() {
		run, rowTotal, err := scanCanvasRunWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = rowTotal
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return runs, total, nil
}

func (r *canvasRepository) GetCanvasRun(ctx context.Context, userID int64, runID int64) (*service.CanvasRun, error) {
	query := canvasRunSelectSQL() + `
		WHERE id = $1 AND user_id = $2
	`
	return scanCanvasRunRow(ctx, r.sql, query, runID, userID)
}

func (r *canvasRepository) CancelCanvasRun(ctx context.Context, userID int64, runID int64) (*service.CanvasRun, error) {
	query := `
		UPDATE canvas_runs
		SET status = $3,
			canceled_at = COALESCE(canceled_at, NOW()),
			completed_at = COALESCE(completed_at, NOW()),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND status IN ($4, $5)
		RETURNING id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at
	`
	return scanCanvasRunRow(ctx, r.sql, query, runID, userID, service.CanvasRunStatusCanceled, service.CanvasRunStatusPending, service.CanvasRunStatusRunning)
}

func (r *canvasRepository) insertCanvasHeader(ctx context.Context, exec sqlExecutor, userID int64, input service.CanvasSaveInput) (*service.CanvasDocument, error) {
	viewportJSON, err := marshalCanvasJSON(input.Viewport)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := marshalCanvasJSON(input.Metadata)
	if err != nil {
		return nil, err
	}
	query := `
		INSERT INTO canvas_documents (user_id, title, description, viewport, metadata)
		VALUES ($1,$2,$3,$4::jsonb,$5::jsonb)
		RETURNING id, user_id, title, description, viewport, metadata, created_at, updated_at
	`
	return scanCanvasDocumentHeaderRow(ctx, exec, query, userID, input.Title, input.Description, string(viewportJSON), string(metadataJSON))
}

func (r *canvasRepository) updateCanvasHeader(ctx context.Context, exec sqlExecutor, userID int64, input service.CanvasSaveInput) (*service.CanvasDocument, error) {
	viewportJSON, err := marshalCanvasJSON(input.Viewport)
	if err != nil {
		return nil, err
	}
	metadataJSON, err := marshalCanvasJSON(input.Metadata)
	if err != nil {
		return nil, err
	}
	query := `
		UPDATE canvas_documents
		SET title = $3,
			description = $4,
			viewport = $5::jsonb,
			metadata = $6::jsonb,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, title, description, viewport, metadata, created_at, updated_at
	`
	return scanCanvasDocumentHeaderRow(ctx, exec, query, input.ID, userID, input.Title, input.Description, string(viewportJSON), string(metadataJSON))
}

func (r *canvasRepository) getCanvasHeader(ctx context.Context, exec sqlExecutor, userID int64, canvasID int64) (*service.CanvasDocument, error) {
	query := canvasDocumentSelectSQL() + `
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	return scanCanvasDocumentHeaderRow(ctx, exec, query, canvasID, userID)
}

func (r *canvasRepository) listCanvasNodes(ctx context.Context, canvasID int64) ([]service.CanvasNode, error) {
	query := `
		SELECT node_key, type, position, data, metadata
		FROM canvas_nodes
		WHERE canvas_id = $1
		ORDER BY id ASC
	`
	rows, err := r.sql.QueryContext(ctx, query, canvasID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	nodes := make([]service.CanvasNode, 0)
	for rows.Next() {
		node, err := scanCanvasNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (r *canvasRepository) listCanvasEdges(ctx context.Context, canvasID int64) ([]service.CanvasEdge, error) {
	query := `
		SELECT edge_key, source_node_key, target_node_key, source_handle, target_handle, data, metadata
		FROM canvas_edges
		WHERE canvas_id = $1
		ORDER BY id ASC
	`
	rows, err := r.sql.QueryContext(ctx, query, canvasID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	edges := make([]service.CanvasEdge, 0)
	for rows.Next() {
		edge, err := scanCanvasEdge(rows)
		if err != nil {
			return nil, err
		}
		edges = append(edges, edge)
	}
	return edges, rows.Err()
}

func insertCanvasNodes(ctx context.Context, exec sqlExecutor, userID int64, canvasID int64, nodes []service.CanvasNode) error {
	query := `
		INSERT INTO canvas_nodes (canvas_id, user_id, node_key, type, position, data, metadata)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7::jsonb)
	`
	for _, node := range nodes {
		positionJSON, err := marshalCanvasJSON(node.Position)
		if err != nil {
			return err
		}
		dataJSON, err := marshalCanvasJSON(node.Data)
		if err != nil {
			return err
		}
		metadataJSON, err := marshalCanvasJSON(node.Metadata)
		if err != nil {
			return err
		}
		if _, err := exec.ExecContext(ctx, query, canvasID, userID, node.ID, node.Type, string(positionJSON), string(dataJSON), string(metadataJSON)); err != nil {
			return err
		}
	}
	return nil
}

func insertCanvasEdges(ctx context.Context, exec sqlExecutor, userID int64, canvasID int64, edges []service.CanvasEdge) error {
	query := `
		INSERT INTO canvas_edges (
			canvas_id, user_id, edge_key, source_node_key, target_node_key,
			source_handle, target_handle, data, metadata
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb)
	`
	for _, edge := range edges {
		dataJSON, err := marshalCanvasJSON(edge.Data)
		if err != nil {
			return err
		}
		metadataJSON, err := marshalCanvasJSON(edge.Metadata)
		if err != nil {
			return err
		}
		if _, err := exec.ExecContext(
			ctx,
			query,
			canvasID,
			userID,
			edge.ID,
			edge.Source,
			edge.Target,
			nullableCanvasString(edge.SourceHandle),
			nullableCanvasString(edge.TargetHandle),
			string(dataJSON),
			string(metadataJSON),
		); err != nil {
			return err
		}
	}
	return nil
}

func canvasDocumentSelectSQL() string {
	return `
		SELECT id, user_id, title, description, viewport, metadata, created_at, updated_at
		FROM canvas_documents
	`
}

func canvasRunSelectSQL() string {
	return `
		SELECT id, user_id, canvas_id, status, trigger_type, api_key_id, model,
			input, output, error_message, metadata, started_at, completed_at,
			canceled_at, created_at, updated_at
		FROM canvas_runs
	`
}

func scanCanvasDocumentHeaderRow(ctx context.Context, exec sqlExecutor, query string, args ...any) (*service.CanvasDocument, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	doc, err := scanCanvasDocumentHeader(rows)
	if err != nil {
		return nil, err
	}
	return &doc, rows.Err()
}

func scanCanvasDocumentHeader(row imageCreatorTaskScanner) (service.CanvasDocument, error) {
	var doc service.CanvasDocument
	var viewport, metadata canvasJSONMap
	err := row.Scan(
		&doc.ID,
		&doc.UserID,
		&doc.Title,
		&doc.Description,
		&viewport,
		&metadata,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return doc, err
	}
	doc.Viewport = map[string]any(viewport)
	doc.Metadata = map[string]any(metadata)
	return doc, nil
}

func scanCanvasListItem(row imageCreatorTaskScanner) (service.CanvasListItem, int, error) {
	var item service.CanvasListItem
	var viewport, metadata canvasJSONMap
	var nodeCount, edgeCount, total int64
	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Description,
		&viewport,
		&metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
		&nodeCount,
		&edgeCount,
		&total,
	)
	if err != nil {
		return item, 0, err
	}
	item.Viewport = map[string]any(viewport)
	item.Metadata = map[string]any(metadata)
	item.NodeCount = int(nodeCount)
	item.EdgeCount = int(edgeCount)
	return item, int(total), nil
}

func scanCanvasNode(row imageCreatorTaskScanner) (service.CanvasNode, error) {
	var node service.CanvasNode
	var position, data, metadata canvasJSONMap
	err := row.Scan(&node.ID, &node.Type, &position, &data, &metadata)
	if err != nil {
		return node, err
	}
	node.Position = map[string]any(position)
	node.Data = map[string]any(data)
	node.Metadata = map[string]any(metadata)
	return node, nil
}

func scanCanvasEdge(row imageCreatorTaskScanner) (service.CanvasEdge, error) {
	var edge service.CanvasEdge
	var sourceHandle, targetHandle sql.NullString
	var data, metadata canvasJSONMap
	err := row.Scan(
		&edge.ID,
		&edge.Source,
		&edge.Target,
		&sourceHandle,
		&targetHandle,
		&data,
		&metadata,
	)
	if err != nil {
		return edge, err
	}
	edge.SourceHandle = nullStringValue(sourceHandle)
	edge.TargetHandle = nullStringValue(targetHandle)
	edge.Data = map[string]any(data)
	edge.Metadata = map[string]any(metadata)
	return edge, nil
}

func scanCanvasRunRow(ctx context.Context, exec sqlExecutor, query string, args ...any) (*service.CanvasRun, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	run, err := scanCanvasRun(rows)
	if err != nil {
		return nil, err
	}
	return &run, rows.Err()
}

func scanCanvasRunWithTotal(row imageCreatorTaskScanner) (service.CanvasRun, int, error) {
	var run service.CanvasRun
	var total int64
	if err := scanCanvasRunDest(row, &run, &total); err != nil {
		return run, 0, err
	}
	return run, int(total), nil
}

func scanCanvasRun(row imageCreatorTaskScanner) (service.CanvasRun, error) {
	var run service.CanvasRun
	if err := scanCanvasRunDest(row, &run); err != nil {
		return run, err
	}
	return run, nil
}

func scanCanvasRunDest(row imageCreatorTaskScanner, run *service.CanvasRun, extra ...any) error {
	var canvasID, apiKeyID sql.NullInt64
	var input, output, metadata canvasJSONMap
	var startedAt, completedAt, canceledAt sql.NullTime
	dest := []any{
		&run.ID,
		&run.UserID,
		&canvasID,
		&run.Status,
		&run.TriggerType,
		&apiKeyID,
		&run.Model,
		&input,
		&output,
		&run.ErrorMessage,
		&metadata,
		&startedAt,
		&completedAt,
		&canceledAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	}
	dest = append(dest, extra...)
	if err := row.Scan(dest...); err != nil {
		return err
	}
	if canvasID.Valid {
		run.CanvasID = canvasID.Int64
	}
	if apiKeyID.Valid {
		run.APIKeyID = apiKeyID.Int64
	}
	run.Input = map[string]any(input)
	run.Output = map[string]any(output)
	run.Metadata = map[string]any(metadata)
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	if canceledAt.Valid {
		run.CanceledAt = &canceledAt.Time
	}
	return nil
}

type canvasJSONMap map[string]any

func (m *canvasJSONMap) Scan(value any) error {
	if value == nil {
		*m = canvasJSONMap{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*m = canvasJSONMap{}
		return nil
	}
	if len(bytes) == 0 {
		*m = canvasJSONMap{}
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	*m = canvasJSONMap(decoded)
	return nil
}

func marshalCanvasJSON(value map[string]any) ([]byte, error) {
	if len(value) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(value)
}

func nullableCanvasString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableCanvasInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}
