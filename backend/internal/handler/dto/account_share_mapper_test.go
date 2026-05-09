package dto

import (
	"testing"

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
