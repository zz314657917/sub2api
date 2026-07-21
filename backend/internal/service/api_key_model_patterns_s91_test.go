package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

// These tests stop before persistence and group validation. Embedding the
// repository interfaces keeps the fixture focused on the boundary under test.
type s91APIKeyUserRepo struct {
	UserRepository
	user *User
}

func (r *s91APIKeyUserRepo) GetByID(context.Context, int64) (*User, error) {
	return r.user, nil
}

type s91APIKeyRepo struct {
	APIKeyRepository
	key *APIKey
}

func (r *s91APIKeyRepo) GetByID(context.Context, int64) (*APIKey, error) {
	return r.key, nil
}

func TestS91CreateRejectsLegacyModelPatterns(t *testing.T) {
	svc := &APIKeyService{
		userRepo: &s91APIKeyUserRepo{user: &User{ID: 7, Status: StatusActive}},
	}

	_, err := svc.Create(context.Background(), 7, CreateAPIKeyRequest{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:       11,
			ModelPatterns: []string{"gpt-*"},
		}},
	})

	require.ErrorIs(t, err, ErrAPIKeyModelPatternsManagedByGroup)
}

func TestS91UpdateRejectsLegacyModelPatterns(t *testing.T) {
	svc := &APIKeyService{
		apiKeyRepo: &s91APIKeyRepo{key: &APIKey{ID: 9, UserID: 7, Key: "s91-key"}},
	}

	_, err := svc.Update(context.Background(), 9, 7, UpdateAPIKeyRequest{
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{
			GroupID:       11,
			ModelPatterns: []string{"gpt-*"},
		}},
	})

	require.ErrorIs(t, err, ErrAPIKeyModelPatternsManagedByGroup)
}
