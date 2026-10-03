package service

import (
	"context"
	"fmt"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrAPIKeyCountExceeded     = infraerrors.Forbidden("API_KEY_COUNT_EXCEEDED", "maximum number of API keys reached")
	ErrAPIKeyCreateRateLimited = infraerrors.TooManyRequests("API_KEY_CREATE_RATE_LIMITED", "too many API key creations, please try again later")
)

// APIKeyCreateCounter is optional to preserve existing APIKeyCache implementations.
// IncrementAPIKeyCreateCount atomically increments attempts in a fixed one-hour window.
type APIKeyCreateCounter interface {
	IncrementAPIKeyCreateCount(ctx context.Context, userID int64) (int64, error)
}

func (s *APIKeyService) checkAPIKeyCreateLimits(ctx context.Context, userID int64) error {
	if s.cfg == nil {
		return nil
	}
	maxKeys := s.cfg.APIKeyCreate.MaxActivePerUser
	maxHourly := s.cfg.APIKeyCreate.MaxPerUserPerHour
	if maxKeys > 0 {
		// CountByUserID includes all nondeleted keys. This is a soft guard;
		// concurrent creators can pass the same count before persistence.
		count, err := s.apiKeyRepo.CountByUserID(ctx, userID)
		if err != nil {
			return fmt.Errorf("count api keys: %w", err)
		}
		if count >= int64(maxKeys) {
			return ErrAPIKeyCountExceeded
		}
	}
	if maxHourly > 0 {
		// Unavailable caches (including implementations without this capability)
		// fail open, consistently with the existing creation error counter.
		if counter, ok := s.cache.(APIKeyCreateCounter); ok {
			count, err := counter.IncrementAPIKeyCreateCount(ctx, userID)
			if err == nil && count > int64(maxHourly) {
				return ErrAPIKeyCreateRateLimited
			}
		}
	}
	return nil
}
