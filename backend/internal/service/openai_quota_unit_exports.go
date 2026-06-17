//go:build unit

package service

import "context"

// NewOpenAIQuotaServiceForTest exposes the narrow test constructor to other
// unit-test packages without requiring fakes to implement full repositories.
func NewOpenAIQuotaServiceForTest(
	accountRepo interface {
		GetByID(ctx context.Context, id int64) (*Account, error)
	},
	proxyRepo interface {
		GetByID(ctx context.Context, id int64) (*Proxy, error)
	},
	tokenProvider *OpenAITokenProvider,
	privacyClientFactory PrivacyClientFactory,
) *OpenAIQuotaService {
	return newOpenAIQuotaService(accountRepo, proxyRepo, tokenProvider, privacyClientFactory)
}

// SetOpenAIQuotaURLsForTest redirects upstream quota endpoints for unit tests.
func SetOpenAIQuotaURLsForTest(serverURL string) func() {
	oldUsageURL := chatGPTUsageURL
	oldResetURL := chatGPTRateLimitResetURL
	chatGPTUsageURL = serverURL + "/backend-api/wham/usage"
	chatGPTRateLimitResetURL = serverURL + "/backend-api/wham/rate-limit-reset-credits/consume"
	return func() {
		chatGPTUsageURL = oldUsageURL
		chatGPTRateLimitResetURL = oldResetURL
	}
}
