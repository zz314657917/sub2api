package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type openAIQuotaResetExtraWriterStub struct {
	updates map[string]any
}

func (s *openAIQuotaResetExtraWriterStub) GetByID(context.Context, int64) (*Account, error) {
	return &Account{}, nil
}

func (s *openAIQuotaResetExtraWriterStub) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	s.updates = updates
	return nil
}

func TestOpenAIQuotaServiceCacheResetCreditsSnapshot(t *testing.T) {
	t.Run("persists complete positive snapshot", func(t *testing.T) {
		repo := &openAIQuotaResetExtraWriterStub{}
		svc := newOpenAIQuotaService(repo, nil, nil, nil)
		credits := &OpenAIRateLimitResetCredits{
			AvailableCount: 1,
			Credits:        []OpenAIRateLimitResetCreditDetail{{ExpiresAt: "2026-08-07T00:00:00Z"}},
		}

		require.NoError(t, svc.CacheResetCreditsSnapshot(context.Background(), 42, credits))
		require.Equal(t, credits, repo.updates[openAIQuotaResetCreditsKey])
	})

	t.Run("rejects incomplete positive snapshot without overwriting cache", func(t *testing.T) {
		repo := &openAIQuotaResetExtraWriterStub{}
		svc := newOpenAIQuotaService(repo, nil, nil, nil)

		err := svc.CacheResetCreditsSnapshot(context.Background(), 42, &OpenAIRateLimitResetCredits{AvailableCount: 1})

		require.Error(t, err)
		require.Nil(t, repo.updates)
	})
}
