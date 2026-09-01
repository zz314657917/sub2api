package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var ErrUsageBillingRequestIDRequired = errors.New("usage billing request_id is required")
var ErrUsageBillingRequestConflict = errors.New("usage billing request fingerprint conflict")
var ErrUsageBillingCommandInvalid = errors.New("usage billing command is invalid")

// UsageBillingCommand describes one billable request that must be applied at most once.
type UsageBillingCommand struct {
	UsageLogID int64
	// FinalizeUsageLog controls whether Apply may close the usage log/outbox.
	// Composite trial settlement keeps this false until its overage and trial
	// consumption steps have both completed.
	FinalizeUsageLog   bool
	RequestID          string
	APIKeyID           int64
	RequestFingerprint string
	RequestPayloadHash string

	UserID              int64
	AccountID           int64
	GroupID             *int64
	SubscriptionID      *int64
	AccountType         string
	Platform            string
	Model               string
	ServiceTier         string
	ReasoningEffort     string
	BillingType         int8
	InputTokens         int
	OutputTokens        int
	CacheCreationTokens int
	CacheReadTokens     int
	ImageCount          int
	MediaType           string

	BalanceCost        float64
	PrepaidBalanceCost float64
	// RequireBalanceCheck keeps known-amount reservations (images, videos and
	// Studio Bridge) strict. Ordinary post-response usage billing may record an
	// overdraft so a delivered upstream response cannot be billed as free.
	RequireBalanceCheck bool
	SubscriptionCost    float64
	APIKeyQuotaCost     float64
	APIKeyRateLimitCost float64
	AccountQuotaCost    float64

	AccountShareOwnerUserID      int64
	AccountShareOwnerRatePercent float64
	AccountShareFreezeHours      int
}

func (c *UsageBillingCommand) Normalize() {
	if c == nil {
		return
	}
	c.RequestID = strings.TrimSpace(c.RequestID)
	if strings.TrimSpace(c.RequestFingerprint) == "" {
		c.RequestFingerprint = buildUsageBillingFingerprint(c)
	}
	c.quantizeMonetaryFields()
}

// UsageBillingMonetaryScale aligns command amounts with PostgreSQL NUMERIC(20,8).
const UsageBillingMonetaryScale = 8

// quantizeMonetaryFields normalizes every amount that is persisted as billing
// cost. It must run after request-fingerprint generation so retries keep the
// raw-value fingerprint produced before this normalization was introduced.
func (c *UsageBillingCommand) quantizeMonetaryFields() {
	c.BalanceCost = QuantizeUsageBillingAmount(c.BalanceCost)
	c.PrepaidBalanceCost = QuantizeUsageBillingAmount(c.PrepaidBalanceCost)
	c.SubscriptionCost = QuantizeUsageBillingAmount(c.SubscriptionCost)
	c.APIKeyQuotaCost = QuantizeUsageBillingAmount(c.APIKeyQuotaCost)
	c.APIKeyRateLimitCost = QuantizeUsageBillingAmount(c.APIKeyRateLimitCost)
	c.AccountQuotaCost = QuantizeUsageBillingAmount(c.AccountQuotaCost)
}

// QuantizeUsageBillingAmount rounds an amount to UsageBillingMonetaryScale
// fractional digits using PostgreSQL NUMERIC's half-away-from-zero behavior.
// Decimal arithmetic avoids binary multiplication and division drift at
// rounding boundaries. Nonfinite inputs are deliberately preserved.
func QuantizeUsageBillingAmount(v float64) float64 {
	if v == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	quantized, _ := decimal.NewFromFloat(v).Round(UsageBillingMonetaryScale).Float64()
	return quantized
}

func buildUsageBillingFingerprint(c *UsageBillingCommand) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf(
		"%d|%d|%d|%s|%s|%s|%s|%d|%d|%d|%d|%d|%d|%s|%d|%0.10f|%0.10f|%0.10f|%0.10f|%0.10f|%0.10f|%d",
		c.UserID,
		c.AccountID,
		c.APIKeyID,
		strings.TrimSpace(c.AccountType),
		strings.TrimSpace(c.Model),
		strings.TrimSpace(c.ServiceTier),
		strings.TrimSpace(c.ReasoningEffort),
		c.BillingType,
		c.InputTokens,
		c.OutputTokens,
		c.CacheCreationTokens,
		c.CacheReadTokens,
		c.ImageCount,
		strings.TrimSpace(c.MediaType),
		valueOrZero(c.SubscriptionID),
		c.BalanceCost,
		c.PrepaidBalanceCost,
		c.SubscriptionCost,
		c.APIKeyQuotaCost,
		c.APIKeyRateLimitCost,
		c.AccountQuotaCost,
		c.AccountShareOwnerUserID,
	)
	if payloadHash := strings.TrimSpace(c.RequestPayloadHash); payloadHash != "" {
		raw += "|" + payloadHash
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func HashUsageRequestPayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func valueOrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// AccountQuotaState holds the post-increment quota state returned by the DB transaction.
// All values are post-update (i.e., already include the increment).
type AccountQuotaState struct {
	TotalUsed    float64
	TotalLimit   float64
	DailyUsed    float64
	DailyLimit   float64
	WeeklyUsed   float64
	WeeklyLimit  float64
	MonthlyUsed  float64
	MonthlyLimit float64
}

type UsageBillingApplyResult struct {
	Applied              bool
	APIKeyQuotaExhausted bool
	NewBalance           *float64           // post-deduction balance (nil = no balance deduction)
	BalanceOverdrafted   bool               // true when usage billing recorded a balance below zero
	VoucherCost          float64            // voucher amount consumed before wallet balance
	BalanceCost          float64            // wallet balance amount consumed after vouchers
	PrepaidBalanceCost   float64            // balance cost already deducted before this billing application
	QuotaState           *AccountQuotaState // post-increment quota state (nil = no quota increment)
}

type UsageBillingRepository interface {
	Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error)
}

// UsageBillingLedgerReconciler repairs the single usage-log ledger row from
// the complete settlement payload. It is intentionally optional so existing
// billing repository implementations and test doubles retain the base Apply
// contract. This matters for trial billing, where primary and overage commands
// have different request_id values but share one usage_log_id ledger row.
type UsageBillingLedgerReconciler interface {
	ReconcileUsageBillingEntry(ctx context.Context, payload *UsageBillingSettlementPayload) error
}

// UsageBillingSettlementTask is a durable, local-only replay of a billing command.
// It deliberately contains no upstream credentials or request/response body.
type UsageBillingSettlementTask struct {
	ID         int64
	UsageLogID int64
	Command    UsageBillingCommand
	Payload    UsageBillingSettlementPayload
	Attempts   int
}

// UsageBillingSettlementPayload is the durable replay snapshot for a request.
// Trial billing is deliberately represented as a composite: the primary
// command, optional wallet/subscription overage, and the idempotent trial-pool
// consumption must all finish before the usage log is marked applied.
type UsageBillingSettlementPayload struct {
	Version int                       `json:"version"`
	Primary UsageBillingCommand       `json:"primary"`
	Overage *UsageBillingCommand      `json:"overage,omitempty"`
	Trial   *UsageBillingTrialPayload `json:"trial,omitempty"`
}

type UsageBillingTrialPayload struct {
	TrialID        int64   `json:"trial_id"`
	UserID         int64   `json:"user_id"`
	TrialRequestID string  `json:"trial_request_id"`
	RequestID      string  `json:"request_id"`
	Amount         float64 `json:"amount"`
	Model          string  `json:"model"`
	APIKeyID       int64   `json:"api_key_id"`
}

// UsageBillingSettlementRepository is optional so existing billing test doubles
// remain source compatible with UsageBillingRepository.
type UsageBillingSettlementRepository interface {
	CreatePending(ctx context.Context, log *UsageLog, cmd *UsageBillingCommand) error
	CreatePendingPayload(ctx context.Context, log *UsageLog, payload *UsageBillingSettlementPayload) error
	MarkPendingError(ctx context.Context, usageLogID int64, billingErr error) error
	MarkApplied(ctx context.Context, usageLogID int64) error
	ClaimPending(ctx context.Context, limit int, lease time.Duration) ([]UsageBillingSettlementTask, error)
	MarkRetry(ctx context.Context, taskID int64, attempts int, billingErr error, nextAttempt time.Time, terminal bool) error
}

// UsageBillingSettlementOwnershipRepository is an optional extension for
// callers that must not race an existing processing outbox lease. owned means
// this caller may run the synchronous settlement; settled means an earlier
// caller already completed it.
type UsageBillingSettlementOwnershipRepository interface {
	CreatePendingPayloadWithOwnership(ctx context.Context, log *UsageLog, payload *UsageBillingSettlementPayload) (owned bool, settled bool, err error)
}
