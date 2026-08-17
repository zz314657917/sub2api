package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func cnProviderAccount(platform, mode, protocol, baseURL string) *Account {
	credentials := map[string]any{"api_key": "sk-test"}
	if mode != "" {
		credentials["account_mode"] = mode
	}
	if protocol != "" {
		credentials["api_protocol"] = protocol
	}
	if baseURL != "" {
		credentials["base_url"] = baseURL
	}
	return &Account{ID: 1, Platform: platform, Type: AccountTypeAPIKey, Credentials: credentials}
}

func TestCNProviderAccountDefaultsAndModes(t *testing.T) {
	cases := []struct {
		name     string
		account  *Account
		wantBase string
		wantMode string
	}{
		{"kimi payg default", cnProviderAccount(PlatformKimi, "", "", ""), DefaultKimiPayGBaseURL, ""},
		{"kimi coding", cnProviderAccount(PlatformKimi, AccountModeCoding, "", ""), DefaultKimiCodingBaseURL, AccountModeCoding},
		{"zhipu payg default", cnProviderAccount(PlatformZhipu, "", "", ""), DefaultZhipuPayGBaseURL, ""},
		{"zhipu coding", cnProviderAccount(PlatformZhipu, AccountModeCoding, "", ""), DefaultZhipuCodingBaseURL, AccountModeCoding},
		{"deepseek default", cnProviderAccount(PlatformDeepseek, AccountModeCoding, "", ""), DefaultDeepseekBaseURL, AccountModeCoding},
		{"custom base preserved", cnProviderAccount(PlatformKimi, AccountModeCoding, "", "https://relay.example/v1"), "https://relay.example/v1", AccountModeCoding},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, tc.account.IsCNProvider())
			require.Equal(t, tc.wantMode, tc.account.GetAccountMode())
			require.Equal(t, tc.wantBase, tc.account.GetOpenAIBaseURL())
		})
	}
}

func TestGetOpenAIProtocolAPIKey_CNProviders(t *testing.T) {
	kimi := cnProviderAccount(PlatformKimi, "", "", "")
	require.Equal(t, "sk-test", kimi.GetOpenAIProtocolAPIKey())
	require.False(t, kimi.IsOpenAIApiKey())

	notAPIKey := cnProviderAccount(PlatformDeepseek, "", "", "")
	notAPIKey.Type = AccountTypeOAuth
	require.Empty(t, notAPIKey.GetOpenAIProtocolAPIKey())

	openAI := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-openai"}}
	require.Equal(t, "sk-openai", openAI.GetOpenAIProtocolAPIKey())
}

func TestGetAPIProtocol(t *testing.T) {
	require.Equal(t, APIProtocolChatCompletions, cnProviderAccount(PlatformKimi, "", "", "").GetAPIProtocol())
	require.Equal(t, APIProtocolAnthropic, cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "").GetAPIProtocol())
	require.Equal(t, APIProtocolResponses, cnProviderAccount(PlatformDeepseek, "", APIProtocolResponses, "").GetAPIProtocol())
	require.Equal(t, APIProtocolChatCompletions, cnProviderAccount(PlatformKimi, "", APIProtocolResponses, "").GetAPIProtocol())
	require.Equal(t, APIProtocolChatCompletions, cnProviderAccount(PlatformZhipu, "", "invalid", "").GetAPIProtocol())
	require.Equal(t, APIProtocolChatCompletions, (&Account{Platform: PlatformOpenAI}).GetAPIProtocol())
}

func TestAnthropicProtocolBaseURL(t *testing.T) {
	require.Equal(t, DefaultKimiPayGAnthropicBaseURL, cnProviderAccount(PlatformKimi, "", APIProtocolAnthropic, "").GetAnthropicProtocolBaseURL())
	require.Equal(t, DefaultKimiCodingAnthropicBaseURL, cnProviderAccount(PlatformKimi, AccountModeCoding, APIProtocolAnthropic, "").GetAnthropicProtocolBaseURL())
	require.Equal(t, DefaultZhipuAnthropicBaseURL, cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "").GetAnthropicProtocolBaseURL())
	require.Equal(t, DefaultDeepseekAnthropicBaseURL, cnProviderAccount(PlatformDeepseek, "", APIProtocolAnthropic, "").GetAnthropicProtocolBaseURL())
	require.Equal(t, "https://relay.example/anthropic", cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "https://relay.example/anthropic").GetAnthropicProtocolBaseURL())
	require.Empty(t, cnProviderAccount(PlatformZhipu, "", "", "").GetAnthropicProtocolBaseURL())
}

func TestGetOpenAIFormatBaseURL_ProtocolAware(t *testing.T) {
	zhipu := cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "https://open.bigmodel.cn/api/anthropic")
	require.Equal(t, DefaultZhipuPayGBaseURL, zhipu.GetOpenAIFormatBaseURL())

	kimi := cnProviderAccount(PlatformKimi, AccountModeCoding, APIProtocolAnthropic, "https://api.kimi.com/coding")
	require.Equal(t, DefaultKimiCodingBaseURL, kimi.GetOpenAIFormatBaseURL())

	deepseek := cnProviderAccount(PlatformDeepseek, "", "", "https://relay.example")
	require.Equal(t, "https://relay.example", deepseek.GetOpenAIFormatBaseURL())
}

func TestBuildUpstreamModelsRequest_CNProviders(t *testing.T) {
	svc := &AccountTestService{cfg: &config.Config{}}
	cases := []struct {
		name string
		acct *Account
		url  string
	}{
		{"kimi", cnProviderAccount(PlatformKimi, "", "", ""), DefaultKimiPayGBaseURL + "/models"},
		{"kimi coding", cnProviderAccount(PlatformKimi, AccountModeCoding, "", ""), DefaultKimiCodingBaseURL + "/models"},
		{"zhipu", cnProviderAccount(PlatformZhipu, "", "", ""), DefaultZhipuPayGBaseURL + "/models"},
		{"zhipu coding", cnProviderAccount(PlatformZhipu, AccountModeCoding, "", ""), DefaultZhipuCodingBaseURL + "/models"},
		{"deepseek", cnProviderAccount(PlatformDeepseek, "", "", ""), DefaultDeepseekBaseURL + "/v1/models"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := svc.buildUpstreamModelsRequest(context.Background(), tc.acct)
			require.NoError(t, err)
			require.Equal(t, tc.url, req.URL.String())
			require.Equal(t, "Bearer sk-test", req.Header.Get("Authorization"))
		})
	}
}

func TestBuildUpstreamModelsRequest_AnthropicProtocol(t *testing.T) {
	svc := &AccountTestService{cfg: &config.Config{}}
	account := cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "https://open.bigmodel.cn/api/anthropic")
	req, err := svc.buildUpstreamModelsRequest(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, DefaultZhipuPayGBaseURL+"/models", req.URL.String())
	require.Equal(t, "Bearer sk-test", req.Header.Get("Authorization"))
}

func TestGetAnthropicAPIKeyAuthScheme_CNProvider(t *testing.T) {
	account := cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "")
	require.Equal(t, AnthropicAPIKeyAuthSchemeXAPIKey, account.GetAnthropicAPIKeyAuthScheme())

	account.Extra = map[string]any{anthropicAPIKeyAuthSchemeExtraKey: AnthropicAPIKeyAuthSchemeAuthorizationBearer}
	require.Equal(t, AnthropicAPIKeyAuthSchemeAuthorizationBearer, account.GetAnthropicAPIKeyAuthScheme())

	header := http.Header{}
	setAnthropicAPIKeyAuthHeader(header, account, "sk-test")
	require.Equal(t, []string{"Bearer sk-test"}, header["authorization"])
	require.NotContains(t, header, "x-api-key")
}
