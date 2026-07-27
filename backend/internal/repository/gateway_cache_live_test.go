//go:build integration

package repository

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func (s *GatewayCacheSuite) TestLiveCallRoundTripPreservesAttestationCiphertext() {
	store, ok := s.cache.(service.LiveCallStore)
	require.True(s.T(), ok, "gateway cache must expose the optional Live call store")

	record := &service.LiveCallRecord{
		CallID:                "call_round_trip_attestation",
		CallHash:              "hash_round_trip_attestation",
		AccountID:             11,
		APIKeyID:              22,
		UserID:                33,
		GroupID:               44,
		SubscriptionID:        55,
		LeaseID:               "lease-round-trip",
		Model:                 "gpt-live-test",
		CreatedAt:             time.Now().Add(-time.Second),
		ExpiresAt:             time.Now().Add(time.Hour),
		Controller:            service.LiveControllerPending,
		AttestationCiphertext: "encrypted-devicecheck-attestation",
	}

	require.NoError(s.T(), store.SaveLiveCall(s.ctx, record, time.Minute))
	loaded, err := store.GetLiveCall(s.ctx, record.CallHash)
	require.NoError(s.T(), err)
	require.Equal(s.T(), record.AttestationCiphertext, loaded.AttestationCiphertext)
}
