package dto

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestAccountFromServiceShallowMapsOwnerShareFields(t *testing.T) {
	ownerID := int64(42)
	out := AccountFromServiceShallow(&service.Account{
		ID:          1,
		Name:        "owner account",
		OwnerUserID: &ownerID,
		ShareMode:   service.AccountShareModePublic,
		ShareStatus: service.AccountShareStatusPendingReview,
	})

	if out.OwnerUserID == nil || *out.OwnerUserID != ownerID {
		t.Fatalf("owner_user_id = %#v", out.OwnerUserID)
	}
	if out.ShareMode != service.AccountShareModePublic {
		t.Fatalf("share_mode = %q", out.ShareMode)
	}
	if out.ShareStatus != service.AccountShareStatusPendingReview {
		t.Fatalf("share_status = %q", out.ShareStatus)
	}
}

func TestAccountFromServiceShallowMapsOpenAIOAuthShareDisplayFields(t *testing.T) {
	out := AccountFromServiceShallow(&service.Account{
		ID:       2,
		Name:     "openai oauth",
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Extra: map[string]any{
			"share_display_tier":          "plus",
			"share_display_percent_only":  true,
			"share_display_account_count": 3,
			"share_display_5h_limit":      500.0,
			"share_display_5h_used":       95.17,
			"share_display_7d_limit":      2160.0,
			"share_display_7d_used":       95.17,
		},
	})

	require.NotNil(t, out.ShareDisplayTier)
	require.Equal(t, "plus", *out.ShareDisplayTier)
	require.NotNil(t, out.ShareDisplayPercentOnly)
	require.True(t, *out.ShareDisplayPercentOnly)
	require.NotNil(t, out.ShareDisplayAccountCount)
	require.Equal(t, 3, *out.ShareDisplayAccountCount)
	require.NotNil(t, out.ShareDisplay5hLimit)
	require.Equal(t, 500.0, *out.ShareDisplay5hLimit)
	require.NotNil(t, out.ShareDisplay5hUsed)
	require.Equal(t, 95.17, *out.ShareDisplay5hUsed)
	require.NotNil(t, out.ShareDisplay7dLimit)
	require.Equal(t, 2160.0, *out.ShareDisplay7dLimit)
	require.NotNil(t, out.ShareDisplay7dUsed)
	require.Equal(t, 95.17, *out.ShareDisplay7dUsed)
}
