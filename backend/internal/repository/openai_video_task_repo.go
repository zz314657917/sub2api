package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIVideoTaskRepository struct {
	db *sql.DB
}

func NewOpenAIVideoTaskRepository(db *sql.DB) service.OpenAIVideoTaskRepository {
	return &openAIVideoTaskRepository{db: db}
}

func (r *openAIVideoTaskRepository) UpsertSubmitted(ctx context.Context, input *service.OpenAIVideoTaskUpsertInput) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	if input == nil || strings.TrimSpace(input.TaskID) == "" {
		return nil, errors.New("openai video task_id is required")
	}
	provider := normalizeVideoTaskProvider(input.Provider)
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "submitted"
	}
	query := `
		INSERT INTO openai_video_tasks (
			task_id,
			provider,
			user_id,
			api_key_id,
			group_id,
			account_id,
			model,
			billing_model,
			upstream_model,
			channel_id,
			original_model,
			channel_mapped_model,
			billing_model_source,
			model_mapping_chain,
			status,
			billing_status,
			estimated_cost,
			reserved_cost,
			refunded_cost,
			request_payload_hash,
			submit_response,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW(), NOW()
		)
		ON CONFLICT (provider, task_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			api_key_id = EXCLUDED.api_key_id,
			group_id = EXCLUDED.group_id,
			account_id = EXCLUDED.account_id,
			model = EXCLUDED.model,
			billing_model = EXCLUDED.billing_model,
			upstream_model = EXCLUDED.upstream_model,
			channel_id = EXCLUDED.channel_id,
			original_model = EXCLUDED.original_model,
			channel_mapped_model = EXCLUDED.channel_mapped_model,
			billing_model_source = EXCLUDED.billing_model_source,
			model_mapping_chain = EXCLUDED.model_mapping_chain,
			status = EXCLUDED.status,
			estimated_cost = CASE
				WHEN openai_video_tasks.estimated_cost > 0 THEN openai_video_tasks.estimated_cost
				ELSE EXCLUDED.estimated_cost
			END,
			reserved_cost = CASE
				WHEN openai_video_tasks.reserved_cost > 0 THEN openai_video_tasks.reserved_cost
				ELSE EXCLUDED.reserved_cost
			END,
			request_payload_hash = EXCLUDED.request_payload_hash,
			submit_response = EXCLUDED.submit_response,
			updated_at = NOW()
		RETURNING ` + openAIVideoTaskColumns
	return scanOpenAIVideoTaskRow(ctx, r.db, query,
		strings.TrimSpace(input.TaskID),
		provider,
		input.UserID,
		input.APIKeyID,
		openAIVideoTaskNullableInt64Arg(input.GroupID),
		input.AccountID,
		input.Model,
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModel),
		openAIVideoTaskNullableTrimmedStringArg(input.UpstreamModel),
		openAIVideoTaskNullablePositiveInt64Arg(input.ChannelID),
		openAIVideoTaskNullableTrimmedStringArg(input.OriginalModel),
		openAIVideoTaskNullableTrimmedStringArg(input.ChannelMappedModel),
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModelSource),
		openAIVideoTaskNullableTrimmedStringArg(input.ModelMappingChain),
		status,
		service.OpenAIVideoTaskBillingPending,
		normalizeNonNegativeCost(input.EstimatedCost),
		normalizeNonNegativeCost(input.ReservedCost),
		float64(0),
		openAIVideoTaskNullableTrimmedStringArg(input.RequestPayloadHash),
		openAIVideoTaskNullableJSONBytesArg(input.SubmitResponse),
	)
}

func (r *openAIVideoTaskRepository) ReserveBalance(ctx context.Context, input *service.OpenAIVideoTaskUpsertInput) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	if input == nil || strings.TrimSpace(input.TaskID) == "" {
		return nil, errors.New("openai video task_id is required")
	}
	amount := normalizeNonNegativeCost(input.ReservedCost)
	if amount <= 0 {
		return nil, errors.New("openai video task reserved cost must be positive")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err := deductWelfareVoucherThenBalance(ctx, tx, input.UserID, amount, welfareVoucherOperationOpenAIVideo, input.TaskID, true); err != nil {
		return nil, err
	}
	task, err := upsertSubmittedOpenAIVideoTask(ctx, tx, input, service.OpenAIVideoTaskBillingReserved, amount, amount)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return task, nil
}

func (r *openAIVideoTaskRepository) BindSubmitted(ctx context.Context, placeholderTaskID string, input *service.OpenAIVideoTaskUpsertInput) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	placeholderTaskID = strings.TrimSpace(placeholderTaskID)
	if input == nil || placeholderTaskID == "" || strings.TrimSpace(input.TaskID) == "" {
		return nil, errors.New("openai video task_id is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	provider := normalizeVideoTaskProvider(input.Provider)
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "submitted"
	}
	query := `
		UPDATE openai_video_tasks
		SET
			task_id = $3,
			user_id = $4,
			api_key_id = $5,
			group_id = $6,
			account_id = $7,
			model = $8,
			billing_model = $9,
			upstream_model = $10,
			channel_id = $11,
			original_model = $12,
			channel_mapped_model = $13,
			billing_model_source = $14,
			model_mapping_chain = $15,
			status = $16,
			estimated_cost = CASE WHEN estimated_cost > 0 THEN estimated_cost ELSE $17 END,
			reserved_cost = CASE WHEN reserved_cost > 0 THEN reserved_cost ELSE $18 END,
			request_payload_hash = $19,
			submit_response = $20,
			updated_at = NOW()
		WHERE provider = $1 AND task_id = $2
		RETURNING ` + openAIVideoTaskColumns
	task, err := scanOpenAIVideoTaskTx(ctx, tx, query,
		provider,
		placeholderTaskID,
		strings.TrimSpace(input.TaskID),
		input.UserID,
		input.APIKeyID,
		openAIVideoTaskNullableInt64Arg(input.GroupID),
		input.AccountID,
		input.Model,
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModel),
		openAIVideoTaskNullableTrimmedStringArg(input.UpstreamModel),
		openAIVideoTaskNullablePositiveInt64Arg(input.ChannelID),
		openAIVideoTaskNullableTrimmedStringArg(input.OriginalModel),
		openAIVideoTaskNullableTrimmedStringArg(input.ChannelMappedModel),
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModelSource),
		openAIVideoTaskNullableTrimmedStringArg(input.ModelMappingChain),
		status,
		normalizeNonNegativeCost(input.EstimatedCost),
		normalizeNonNegativeCost(input.ReservedCost),
		openAIVideoTaskNullableTrimmedStringArg(input.RequestPayloadHash),
		openAIVideoTaskNullableJSONBytesArg(input.SubmitResponse),
	)
	if err != nil {
		return nil, err
	}
	if err := rebindWelfareVoucherDeductions(ctx, tx, input.UserID, welfareVoucherOperationOpenAIVideo, placeholderTaskID, strings.TrimSpace(input.TaskID)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return task, nil
}

func (r *openAIVideoTaskRepository) GetByTaskID(ctx context.Context, provider string, taskID string) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, sql.ErrNoRows
	}
	query := `SELECT ` + openAIVideoTaskColumns + ` FROM openai_video_tasks WHERE provider = $1 AND task_id = $2`
	return scanOpenAIVideoTaskRow(ctx, r.db, query, normalizeVideoTaskProvider(provider), taskID)
}

func (r *openAIVideoTaskRepository) UpdateStatus(ctx context.Context, input *service.OpenAIVideoTaskStatusUpdate) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	if input == nil || strings.TrimSpace(input.TaskID) == "" {
		return nil, errors.New("openai video task_id is required")
	}
	status := openAIVideoTaskNullableTrimmedStringArg(input.Status)
	billingStatus := openAIVideoTaskNullableTrimmedStringArg(input.BillingStatus)
	query := `
		UPDATE openai_video_tasks
		SET
			status = COALESCE($3, status),
			billing_status = COALESCE($4, billing_status),
			usage_log_id = COALESCE($5, usage_log_id),
			last_status_response = COALESCE($6, last_status_response),
			completed_at = COALESCE($7, completed_at),
			billed_at = COALESCE($8, billed_at),
			updated_at = NOW()
		WHERE provider = $1 AND task_id = $2
		RETURNING ` + openAIVideoTaskColumns
	return scanOpenAIVideoTaskRow(ctx, r.db, query,
		normalizeVideoTaskProvider(input.Provider),
		strings.TrimSpace(input.TaskID),
		status,
		billingStatus,
		openAIVideoTaskNullableInt64Arg(input.UsageLogID),
		openAIVideoTaskNullableJSONBytesArg(input.LastStatusResponse),
		openAIVideoTaskNullableTimeArg(input.CompletedAt),
		openAIVideoTaskNullableTimeArg(input.BilledAt),
	)
}

func (r *openAIVideoTaskRepository) RefundReserved(ctx context.Context, provider string, taskID string) (*service.OpenAIVideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("openai video task repository db is nil")
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, sql.ErrNoRows
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	task, err := scanOpenAIVideoTaskTx(ctx, tx, `
		UPDATE openai_video_tasks
		SET
			billing_status = $3,
			refunded_cost = reserved_cost,
			updated_at = NOW()
		WHERE provider = $1
			AND task_id = $2
			AND billing_status = $4
			AND reserved_cost > 0
		RETURNING `+openAIVideoTaskColumns,
		normalizeVideoTaskProvider(provider),
		taskID,
		service.OpenAIVideoTaskBillingRefunded,
		service.OpenAIVideoTaskBillingReserved,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return r.GetByTaskID(ctx, provider, taskID)
	}
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("openai video task refund returned nil task")
	}
	_, err = refundWelfareVoucherDeductions(ctx, tx, task.UserID, task.ReservedCost, welfareVoucherOperationOpenAIVideo, task.TaskID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return task, nil
}

const openAIVideoTaskColumns = `
	id,
	task_id,
	provider,
	user_id,
	api_key_id,
	group_id,
	account_id,
	model,
	billing_model,
	upstream_model,
	channel_id,
	original_model,
	channel_mapped_model,
	billing_model_source,
	model_mapping_chain,
	status,
	billing_status,
	estimated_cost,
	reserved_cost,
	refunded_cost,
	usage_log_id,
	request_payload_hash,
	submit_response,
	last_status_response,
	created_at,
	updated_at,
	completed_at,
	billed_at`

func scanOpenAIVideoTaskRow(ctx context.Context, db *sql.DB, query string, args ...any) (*service.OpenAIVideoTask, error) {
	return scanOpenAIVideoTask(ctx, db, query, args...)
}

func scanOpenAIVideoTaskTx(ctx context.Context, tx *sql.Tx, query string, args ...any) (*service.OpenAIVideoTask, error) {
	return scanOpenAIVideoTask(ctx, tx, query, args...)
}

type openAIVideoTaskQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func scanOpenAIVideoTask(ctx context.Context, q openAIVideoTaskQuerier, query string, args ...any) (*service.OpenAIVideoTask, error) {
	var row openAIVideoTaskScan
	if err := q.QueryRowContext(ctx, query, args...).Scan(
		&row.ID,
		&row.TaskID,
		&row.Provider,
		&row.UserID,
		&row.APIKeyID,
		&row.GroupID,
		&row.AccountID,
		&row.Model,
		&row.BillingModel,
		&row.UpstreamModel,
		&row.ChannelID,
		&row.OriginalModel,
		&row.ChannelMappedModel,
		&row.BillingModelSource,
		&row.ModelMappingChain,
		&row.Status,
		&row.BillingStatus,
		&row.EstimatedCost,
		&row.ReservedCost,
		&row.RefundedCost,
		&row.UsageLogID,
		&row.RequestPayloadHash,
		&row.SubmitResponse,
		&row.LastStatusResponse,
		&row.CreatedAt,
		&row.UpdatedAt,
		&row.CompletedAt,
		&row.BilledAt,
	); err != nil {
		return nil, err
	}
	return row.toService(), nil
}

type openAIVideoTaskScan struct {
	ID                 int64
	TaskID             string
	Provider           string
	UserID             int64
	APIKeyID           int64
	GroupID            sql.NullInt64
	AccountID          int64
	Model              string
	BillingModel       sql.NullString
	UpstreamModel      sql.NullString
	ChannelID          sql.NullInt64
	OriginalModel      sql.NullString
	ChannelMappedModel sql.NullString
	BillingModelSource sql.NullString
	ModelMappingChain  sql.NullString
	Status             string
	BillingStatus      string
	EstimatedCost      float64
	ReservedCost       float64
	RefundedCost       float64
	UsageLogID         sql.NullInt64
	RequestPayloadHash sql.NullString
	SubmitResponse     []byte
	LastStatusResponse []byte
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        sql.NullTime
	BilledAt           sql.NullTime
}

func (r openAIVideoTaskScan) toService() *service.OpenAIVideoTask {
	out := &service.OpenAIVideoTask{
		ID:                 r.ID,
		TaskID:             r.TaskID,
		Provider:           r.Provider,
		UserID:             r.UserID,
		APIKeyID:           r.APIKeyID,
		AccountID:          r.AccountID,
		Model:              r.Model,
		BillingModel:       openAIVideoTaskNullStringValue(r.BillingModel),
		UpstreamModel:      openAIVideoTaskNullStringValue(r.UpstreamModel),
		ChannelID:          openAIVideoTaskNullInt64Value(r.ChannelID),
		OriginalModel:      openAIVideoTaskNullStringValue(r.OriginalModel),
		ChannelMappedModel: openAIVideoTaskNullStringValue(r.ChannelMappedModel),
		BillingModelSource: openAIVideoTaskNullStringValue(r.BillingModelSource),
		ModelMappingChain:  openAIVideoTaskNullStringValue(r.ModelMappingChain),
		Status:             r.Status,
		BillingStatus:      r.BillingStatus,
		EstimatedCost:      r.EstimatedCost,
		ReservedCost:       r.ReservedCost,
		RefundedCost:       r.RefundedCost,
		RequestPayloadHash: openAIVideoTaskNullStringValue(r.RequestPayloadHash),
		SubmitResponse:     r.SubmitResponse,
		LastStatusResponse: r.LastStatusResponse,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
	if r.GroupID.Valid {
		value := r.GroupID.Int64
		out.GroupID = &value
	}
	if r.UsageLogID.Valid {
		value := r.UsageLogID.Int64
		out.UsageLogID = &value
	}
	if r.CompletedAt.Valid {
		value := r.CompletedAt.Time
		out.CompletedAt = &value
	}
	if r.BilledAt.Valid {
		value := r.BilledAt.Time
		out.BilledAt = &value
	}
	return out
}

func normalizeVideoTaskProvider(provider string) string {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return service.OpenAIVideoTaskProviderOpenAI
	}
	return provider
}

func openAIVideoTaskNullableTrimmedStringArg(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func openAIVideoTaskNullableInt64Arg(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func openAIVideoTaskNullablePositiveInt64Arg(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func openAIVideoTaskNullableJSONBytesArg(value []byte) any {
	if len(value) == 0 || !json.Valid(value) {
		return nil
	}
	return string(value)
}

func openAIVideoTaskNullableTimeArg(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}

func openAIVideoTaskNullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func openAIVideoTaskNullInt64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func upsertSubmittedOpenAIVideoTask(ctx context.Context, tx *sql.Tx, input *service.OpenAIVideoTaskUpsertInput, billingStatus string, estimatedCost float64, reservedCost float64) (*service.OpenAIVideoTask, error) {
	provider := normalizeVideoTaskProvider(input.Provider)
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "submitted"
	}
	query := `
		INSERT INTO openai_video_tasks (
			task_id,
			provider,
			user_id,
			api_key_id,
			group_id,
			account_id,
			model,
			billing_model,
			upstream_model,
			channel_id,
			original_model,
			channel_mapped_model,
			billing_model_source,
			model_mapping_chain,
			status,
			billing_status,
			estimated_cost,
			reserved_cost,
			request_payload_hash,
			submit_response,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW(), NOW()
		)
		ON CONFLICT (provider, task_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			api_key_id = EXCLUDED.api_key_id,
			group_id = EXCLUDED.group_id,
			account_id = EXCLUDED.account_id,
			model = EXCLUDED.model,
			billing_model = EXCLUDED.billing_model,
			upstream_model = EXCLUDED.upstream_model,
			channel_id = EXCLUDED.channel_id,
			original_model = EXCLUDED.original_model,
			channel_mapped_model = EXCLUDED.channel_mapped_model,
			billing_model_source = EXCLUDED.billing_model_source,
			model_mapping_chain = EXCLUDED.model_mapping_chain,
			status = EXCLUDED.status,
			billing_status = EXCLUDED.billing_status,
			estimated_cost = EXCLUDED.estimated_cost,
			reserved_cost = EXCLUDED.reserved_cost,
			request_payload_hash = EXCLUDED.request_payload_hash,
			submit_response = EXCLUDED.submit_response,
			updated_at = NOW()
		RETURNING ` + openAIVideoTaskColumns
	return scanOpenAIVideoTaskTx(ctx, tx, query,
		strings.TrimSpace(input.TaskID),
		provider,
		input.UserID,
		input.APIKeyID,
		openAIVideoTaskNullableInt64Arg(input.GroupID),
		input.AccountID,
		input.Model,
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModel),
		openAIVideoTaskNullableTrimmedStringArg(input.UpstreamModel),
		openAIVideoTaskNullablePositiveInt64Arg(input.ChannelID),
		openAIVideoTaskNullableTrimmedStringArg(input.OriginalModel),
		openAIVideoTaskNullableTrimmedStringArg(input.ChannelMappedModel),
		openAIVideoTaskNullableTrimmedStringArg(input.BillingModelSource),
		openAIVideoTaskNullableTrimmedStringArg(input.ModelMappingChain),
		status,
		billingStatus,
		normalizeNonNegativeCost(estimatedCost),
		normalizeNonNegativeCost(reservedCost),
		openAIVideoTaskNullableTrimmedStringArg(input.RequestPayloadHash),
		openAIVideoTaskNullableJSONBytesArg(input.SubmitResponse),
	)
}

func normalizeNonNegativeCost(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
