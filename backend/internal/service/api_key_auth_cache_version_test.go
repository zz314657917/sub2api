package service

import (
	"testing"
	"time"
)

func TestAPIKeyService_RejectsV10AuthSnapshotWithoutModelsListConfig(t *testing.T) {
	groupID := int64(9)
	svc := &APIKeyService{}

	apiKey, ok, err := svc.applyAuthCacheEntry("k-legacy-models-list", &APIKeyAuthCacheEntry{
		Snapshot: &APIKeyAuthSnapshot{
			Version:  10,
			APIKeyID: 1,
			UserID:   2,
			GroupID:  &groupID,
			Status:   StatusActive,
			User: APIKeyAuthUserSnapshot{
				ID:          2,
				Status:      StatusActive,
				Role:        RoleUser,
				Balance:     10,
				Concurrency: 3,
			},
			Group: &APIKeyAuthGroupSnapshot{
				ID:               groupID,
				Name:             "openai",
				Platform:         PlatformOpenAI,
				Status:           StatusActive,
				SubscriptionType: SubscriptionTypeStandard,
				RateMultiplier:   1,
			},
		},
	})

	if err != nil {
		t.Fatalf("expected stale snapshot to be ignored without error, got %v", err)
	}
	if ok {
		t.Fatalf("expected v10 auth snapshot to be rejected after models_list_config was added")
	}
	if apiKey != nil {
		t.Fatalf("expected no API key from stale snapshot, got %#v", apiKey)
	}
}

func TestAPIKeyService_PinnedSnapshotRoundTripAndRouteStaysOnBoundGroup(t *testing.T) {
	groupID := int64(9)
	managedSourceID := int64(66)
	bindingExpiresAt := time.Now().Add(time.Hour)
	svc := &APIKeyService{}
	apiKey := &APIKey{
		ID:                      1,
		UserID:                  2,
		Key:                     "cafe-pinned-key",
		GroupID:                 &groupID,
		Status:                  StatusAPIKeyActive,
		PinnedAccountID:         77,
		ManagedBindingID:        88,
		ManagedBindingExpiresAt: &bindingExpiresAt,
		ManagedSourceType:       APIKeyManagedSourceCafeRoomSeat,
		ManagedSourceID:         &managedSourceID,
		User:                    &User{ID: 2, Status: StatusActive, Role: RoleUser},
		Group: &Group{
			ID:               groupID,
			Name:             "cafe-claude",
			Platform:         PlatformAnthropic,
			Status:           StatusActive,
			Hydrated:         true,
			SubscriptionType: SubscriptionTypeStandard,
			AccessMode:       GroupAccessModeRoomManaged,
		},
	}

	snapshot := svc.snapshotFromAPIKey(nil, apiKey)
	if snapshot.Version != apiKeyAuthSnapshotVersion || snapshot.PinnedAccountID != 77 || snapshot.ManagedBindingID != 88 || snapshot.ManagedBindingExpiresAt == nil || !snapshot.ManagedBindingExpiresAt.Equal(bindingExpiresAt) || snapshot.Group == nil || snapshot.Group.AccessMode != GroupAccessModeRoomManaged {
		t.Fatalf("pinned auth snapshot did not preserve binding facts: %#v", snapshot)
	}
	roundTrip := svc.snapshotToAPIKey(apiKey.Key, snapshot)
	if roundTrip.PinnedAccountID != 77 || roundTrip.ManagedBindingID != 88 || roundTrip.ManagedBindingExpiresAt == nil || !roundTrip.ManagedBindingExpiresAt.Equal(bindingExpiresAt) || !roundTrip.IsCafeRoomManaged() || roundTrip.Group == nil || roundTrip.Group.AccessMode != GroupAccessModeRoomManaged {
		t.Fatalf("pinned auth snapshot did not round-trip: %#v", roundTrip)
	}
	if got := svc.ResolveForRequest(nil, roundTrip, "/v1/messages", ""); got == nil || got.GroupID == nil || *got.GroupID != groupID || len(got.MultiGroupRoutes) != 0 {
		t.Fatalf("pinned request should retain exactly its bound group: %#v", got)
	}
	if got := svc.ResolveForModelRequest(nil, roundTrip, "/v1/responses", "", "gpt-5", false); got != nil {
		t.Fatalf("pinned key must reject a model route that would require another platform: %#v", got)
	}
}

func TestAuthCacheEntryExpired_RejectsCafePinWithoutBindingExpiry(t *testing.T) {
	now := time.Now()
	entry := &APIKeyAuthCacheEntry{Snapshot: &APIKeyAuthSnapshot{
		PinnedAccountID:   77,
		ManagedBindingID:  88,
		ManagedSourceType: APIKeyManagedSourceCafeRoomSeat,
	}}

	if !authCacheEntryExpired(entry, now) {
		t.Fatal("a Cafe pinned snapshot without Binding expiry must force an auth reload")
	}
}
