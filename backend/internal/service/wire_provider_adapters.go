package service

// ProvideContentModerationService gives Wire a fixed-argument provider while
// preserving NewContentModerationService's public variadic constructor.
func ProvideContentModerationService(
	settingRepo SettingRepository,
	repo ContentModerationRepository,
	hashCache ContentModerationHashCache,
	groupRepo GroupRepository,
	userRepo UserRepository,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	emailService *EmailService,
	proxyRepo ProxyRepository,
) *ContentModerationService {
	return NewContentModerationService(
		settingRepo,
		repo,
		hashCache,
		groupRepo,
		userRepo,
		authCacheInvalidator,
		emailService,
		proxyRepo,
	)
}
