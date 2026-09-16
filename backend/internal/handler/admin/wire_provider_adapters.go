package admin

import "github.com/Wei-Shaw/sub2api/internal/service"

// ProvideCafeRoomHandler preserves the optional quota-reset hook that Wire
// cannot infer from NewCafeRoomHandlerWithActivation's constructor signature.
func ProvideCafeRoomHandler(
	cafeRoomService *service.CafeRoomService,
	activation *service.CafeRoomActivationService,
	settings *service.SettingService,
	adminService service.AdminService,
) *CafeRoomHandler {
	h := NewCafeRoomHandlerWithActivation(cafeRoomService, activation, settings)
	if quotaReset, ok := adminService.(service.CafeQuotaResetService); ok {
		h.SetQuotaResetService(quotaReset)
	}
	return h
}
