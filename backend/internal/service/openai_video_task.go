package service

import (
	"context"
	"errors"
	"time"
)

var ErrOpenAIVideoPricingUnavailable = errors.New("openai video pricing unavailable")

const (
	OpenAIVideoTaskProviderOpenAI = "openai"

	OpenAIVideoTaskBillingPending           = "pending"
	OpenAIVideoTaskBillingReserved          = "reserved"
	OpenAIVideoTaskBillingCharged           = "charged"
	OpenAIVideoTaskBillingFailedNoCost      = "failed_no_charge"
	OpenAIVideoTaskBillingReservationFailed = "reservation_failed"
	OpenAIVideoTaskBillingRefunded          = "refunded"
)

type OpenAIVideoTask struct {
	ID                 int64
	TaskID             string
	Provider           string
	UserID             int64
	APIKeyID           int64
	GroupID            *int64
	AccountID          int64
	Model              string
	BillingModel       string
	UpstreamModel      string
	ChannelID          int64
	OriginalModel      string
	ChannelMappedModel string
	BillingModelSource string
	ModelMappingChain  string
	Status             string
	BillingStatus      string
	EstimatedCost      float64
	ReservedCost       float64
	RefundedCost       float64
	UsageLogID         *int64
	RequestPayloadHash string
	SubmitResponse     []byte
	LastStatusResponse []byte
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        *time.Time
	BilledAt           *time.Time
}

type OpenAIVideoTaskUpsertInput struct {
	TaskID             string
	Provider           string
	UserID             int64
	APIKeyID           int64
	GroupID            *int64
	AccountID          int64
	Model              string
	BillingModel       string
	UpstreamModel      string
	ChannelID          int64
	OriginalModel      string
	ChannelMappedModel string
	BillingModelSource string
	ModelMappingChain  string
	Status             string
	EstimatedCost      float64
	ReservedCost       float64
	RequestPayloadHash string
	SubmitResponse     []byte
}

type OpenAIVideoTaskStatusUpdate struct {
	TaskID             string
	Provider           string
	Status             string
	BillingStatus      string
	UsageLogID         *int64
	LastStatusResponse []byte
	CompletedAt        *time.Time
	BilledAt           *time.Time
}

type OpenAIVideoTaskRepository interface {
	UpsertSubmitted(ctx context.Context, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error)
	GetByTaskID(ctx context.Context, provider string, taskID string) (*OpenAIVideoTask, error)
	UpdateStatus(ctx context.Context, input *OpenAIVideoTaskStatusUpdate) (*OpenAIVideoTask, error)
	ReserveBalance(ctx context.Context, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error)
	BindSubmitted(ctx context.Context, placeholderTaskID string, input *OpenAIVideoTaskUpsertInput) (*OpenAIVideoTask, error)
	RefundReserved(ctx context.Context, provider string, taskID string) (*OpenAIVideoTask, error)
}
