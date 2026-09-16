package admin

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type wireProviderAdminService struct{ service.AdminService }

type wireProviderQuotaResetAdminService struct{ service.AdminService }

func (wireProviderQuotaResetAdminService) AdminResetCafeRateLimitUsage(context.Context, *int64) (*service.AdminCafeQuotaResetResult, error) {
	return nil, nil
}

func TestWireProviderCafeRoomHandlerPreservesDependenciesAndSupportedReset(t *testing.T) {
	cafeRoomService := &service.CafeRoomService{}
	activation := &service.CafeRoomActivationService{}
	settings := &service.SettingService{}
	adminService := &wireProviderQuotaResetAdminService{}

	h := ProvideCafeRoomHandler(cafeRoomService, activation, settings, adminService)

	if h == nil {
		t.Fatal("provider returned nil handler")
	}
	if h.service != cafeRoomService || h.activation != activation || h.settings != settings {
		t.Fatal("provider did not preserve the cafe room constructor dependencies")
	}
	if h.quotaReset != adminService {
		t.Fatal("provider did not inject the supported quota reset capability")
	}
}

func TestWireProviderCafeRoomHandlerLeavesUnsupportedAndNilResetUnset(t *testing.T) {
	tests := []struct {
		name         string
		adminService service.AdminService
	}{
		{
			name:         "unsupported admin service",
			adminService: &wireProviderAdminService{},
		},
		{
			name: "nil admin service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ProvideCafeRoomHandler(nil, nil, nil, tt.adminService)
			if h == nil {
				t.Fatal("provider returned nil handler")
			}
			if h.quotaReset != nil {
				t.Fatal("provider injected an unsupported quota reset capability")
			}
		})
	}
}
