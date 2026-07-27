package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

const (
	LiveControllerPending  = "pending"
	LiveControllerObserver = "observer"
	LiveControllerProxy    = "proxy"
	LiveControllerClosed   = "closed"
)

var (
	ErrLiveUnavailable       = errors.New("live is unavailable")
	ErrLiveConcurrencyFull   = errors.New("live concurrency is full")
	ErrLiveCallNotFound      = errors.New("live call not found")
	ErrLiveIdentityMismatch  = errors.New("live call identity mismatch")
	ErrLiveControllerChanged = errors.New("live controller changed")
)

type LiveAttestationUnavailableError struct {
	Reason string
}

func (e *LiveAttestationUnavailableError) Error() string {
	if e == nil || e.Reason == "" {
		return "Live attestation is unavailable"
	}
	return "Live attestation is unavailable: " + e.Reason
}

type LiveCallRequest struct {
	SDP     string          `json:"sdp"`
	Session json.RawMessage `json:"session"`
}

type LiveCallIdentity struct {
	APIKeyID        int64
	UserID          int64
	GroupID         *int64
	SubscriptionID  *int64
	UserAgent       string
	IPAddress       string
	InboundEndpoint string
}

type LiveCallRecord struct {
	CallID                string
	CallHash              string
	AccountID             int64
	APIKeyID              int64
	UserID                int64
	GroupID               int64
	SubscriptionID        int64
	LeaseID               string
	Model                 string
	CreatedAt             time.Time
	ExpiresAt             time.Time
	Controller            string
	ControllerOwner       string
	UserAgent             string
	IPAddress             string
	InboundEndpoint       string
	AttestationCiphertext string
}

type LiveCallCreated struct {
	SDP      []byte
	CallID   string
	Location string
	Account  *Account
}

type LiveCallStore interface {
	SaveLiveCall(context.Context, *LiveCallRecord, time.Duration) error
	GetLiveCall(context.Context, string) (*LiveCallRecord, error)
	ClaimLiveController(context.Context, string, string, string) (bool, error)
	ReleaseLiveController(context.Context, string, string) (bool, error)
	GetLiveController(context.Context, string) (string, error)
	MarkLiveCallClosed(context.Context, string, time.Duration) (bool, error)
}

type LiveConcurrencyCache interface {
	AcquireLiveLease(context.Context, int64, int, int64, int, int64, string, bool) (bool, error)
	RefreshLiveLease(context.Context, int64, int64, int64, string) (bool, error)
	ReleaseLiveLease(context.Context, int64, int64, int64, string) error
}
