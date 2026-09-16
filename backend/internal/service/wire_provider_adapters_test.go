package service

import "testing"

type wireProviderHashCache struct{ ContentModerationHashCache }
type wireProviderGroupRepository struct{ GroupRepository }
type wireProviderUserRepository struct{ UserRepository }
type wireProviderAuthCacheInvalidator struct{ APIKeyAuthCacheInvalidator }
type wireProviderProxyRepository struct{ ProxyRepository }
type wireProviderSettingRepository struct{ SettingRepository }
type wireProviderContentModerationRepository struct{ ContentModerationRepository }

func TestWireProviderContentModerationAdapterForwardsDependencies(t *testing.T) {
	hashCache := &wireProviderHashCache{}
	groupRepo := &wireProviderGroupRepository{}
	userRepo := &wireProviderUserRepository{}
	authCacheInvalidator := &wireProviderAuthCacheInvalidator{}
	emailService := &EmailService{}
	proxyRepo := &wireProviderProxyRepository{}

	svc := ProvideContentModerationService(
		nil,
		nil,
		hashCache,
		groupRepo,
		userRepo,
		authCacheInvalidator,
		emailService,
		proxyRepo,
	)

	if svc == nil {
		t.Fatal("adapter returned nil service")
	}
	if svc.settingRepo != nil || svc.repo != nil {
		t.Fatal("adapter must preserve nil moderation repositories to avoid starting workers")
	}
	if svc.hashCache != hashCache || svc.groupRepo != groupRepo || svc.userRepo != userRepo || svc.authCacheInvalidator != authCacheInvalidator || svc.emailService != emailService || svc.proxyRepo != proxyRepo {
		t.Fatal("adapter did not forward all content moderation dependencies")
	}
}

func TestWireProviderContentModerationAdapterPreservesWorkerGuardInputs(t *testing.T) {
	hashCache := &wireProviderHashCache{}
	groupRepo := &wireProviderGroupRepository{}
	userRepo := &wireProviderUserRepository{}
	authCacheInvalidator := &wireProviderAuthCacheInvalidator{}
	emailService := &EmailService{}
	settingRepo := &wireProviderSettingRepository{}
	moderationRepo := &wireProviderContentModerationRepository{}

	tests := []struct {
		name        string
		settingRepo SettingRepository
		repo        ContentModerationRepository
	}{
		{
			name:        "non-nil setting repository",
			settingRepo: settingRepo,
		},
		{
			name: "non-nil moderation repository",
			repo: moderationRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := ProvideContentModerationService(
				tt.settingRepo,
				tt.repo,
				hashCache,
				groupRepo,
				userRepo,
				authCacheInvalidator,
				emailService,
				nil,
			)

			if svc.settingRepo != tt.settingRepo || svc.repo != tt.repo {
				t.Fatal("adapter did not preserve the content moderation worker guard inputs")
			}
			if svc.hashCache != hashCache || svc.groupRepo != groupRepo || svc.userRepo != userRepo || svc.authCacheInvalidator != authCacheInvalidator || svc.emailService != emailService {
				t.Fatal("adapter did not forward the remaining content moderation dependencies")
			}
			if svc.proxyRepo != nil {
				t.Fatal("adapter did not preserve a nil proxy repository")
			}
		})
	}
}

func TestWireProviderWelfareServiceImplementsNewUserTrialConsumer(t *testing.T) {
	var consumer newUserTrialConsumer = (*WelfareService)(nil)
	if consumer == nil {
		t.Fatal("WelfareService must satisfy newUserTrialConsumer")
	}
}
