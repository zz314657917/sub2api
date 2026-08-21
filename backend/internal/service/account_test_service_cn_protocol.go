package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
)

// testCNProviderChatCompletionsConnection probes the fixed Chat Completions
// endpoint selected by the CN account's existing protocol-aware helpers.
func (s *AccountTestService) testCNProviderChatCompletionsConnection(c *gin.Context, account *Account, modelID string, prompt string) error {
	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = openai.DefaultTestModel
	}
	testModelID = account.GetMappedModel(testModelID)

	authToken := strings.TrimSpace(account.GetOpenAIProtocolAPIKey())
	if authToken == "" {
		return s.sendErrorAndEnd(c, "No API key available")
	}

	baseURL, err := s.validateUpstreamBaseURL(account.GetOpenAIBaseURL())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
	}
	return s.testOpenAIChatCompletionsConnection(c, account, testModelID, prompt, baseURL, authToken)
}

// testCNProviderAnthropicConnection probes only the native Anthropic endpoint
// for an account explicitly configured with api_protocol=anthropic. In
// particular, it must never inherit the generic Claude tester's Anthropic
// default and beta query string.
func (s *AccountTestService) testCNProviderAnthropicConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()
	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = claude.DefaultTestModel
	}
	testModelID = account.GetMappedModel(testModelID)

	authToken := strings.TrimSpace(account.GetOpenAIProtocolAPIKey())
	if authToken == "" {
		return s.sendErrorAndEnd(c, "No API key available")
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetAnthropicProtocolBaseURL())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid Anthropic base URL: %s", err.Error()))
	}
	if hint := cnAnthropicBaseURLMisconfigHint(baseURL); hint != "" {
		return s.sendErrorAndEnd(c, hint)
	}
	apiURL := strings.TrimRight(baseURL, "/") + "/v1/messages"

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	payload, err := createTestPayload(testModelID)
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create Anthropic test payload")
	}
	payloadBytes, _ := json.Marshal(payload)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create Anthropic test request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("anthropic-version", "2023-06-01")
	for key, value := range claude.DefaultHeaders {
		req.Header.Set(key, value)
	}
	setAnthropicAPIKeyAuthHeader(req.Header, account, authToken)
	resp, err := s.doCNProviderAccountTestRequest(req, account)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Anthropic endpoint request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("Anthropic endpoint returned %d: %s", resp.StatusCode, string(body))
		if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && s.accountRepo != nil {
			_ = s.accountRepo.SetError(ctx, account.ID, errMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}
	return s.processClaudeStream(c, resp.Body)
}

// testCNProviderResponsesConnection probes DeepSeek's explicit native
// Responses protocol without consulting OpenAI capability metadata.
func (s *AccountTestService) testCNProviderResponsesConnection(c *gin.Context, account *Account, modelID string) error {
	ctx := c.Request.Context()
	testModelID := strings.TrimSpace(modelID)
	if testModelID == "" {
		testModelID = openai.DefaultTestModel
	}
	testModelID = account.GetMappedModel(testModelID)
	authToken := strings.TrimSpace(account.GetOpenAIProtocolAPIKey())
	if authToken == "" {
		return s.sendErrorAndEnd(c, "No API key available")
	}
	baseURL, err := s.validateUpstreamBaseURL(account.GetOpenAIBaseURL())
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Invalid base URL: %s", err.Error()))
	}
	apiURL := buildOpenAIResponsesURLForPlatform(account.Platform, baseURL)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	payload := createOpenAITestPayload(testModelID, false)
	delete(payload, "instructions")
	payloadBytes, _ := json.Marshal(payload)
	payloadBytes = normalizeDeepSeekResponsesRequestBody(account, payloadBytes)
	s.sendEvent(c, TestEvent{Type: "test_start", Model: testModelID})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return s.sendErrorAndEnd(c, "Failed to create Responses test request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := s.doCNProviderAccountTestRequest(req, account)
	if err != nil {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Responses endpoint request failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("Responses endpoint returned %d: %s", resp.StatusCode, string(body))
		if resp.StatusCode == http.StatusUnauthorized && s.accountRepo != nil {
			_ = s.accountRepo.SetError(ctx, account.ID, errMsg)
		}
		return s.sendErrorAndEnd(c, errMsg)
	}
	return s.processOpenAIStream(c, resp.Body)
}

func (s *AccountTestService) doCNProviderAccountTestRequest(req *http.Request, account *Account) (*http.Response, error) {
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	return s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
}

func cnAnthropicBaseURLMisconfigHint(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	path := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	if path == "" {
		return ""
	}
	if strings.Contains(path, "/paas/") || strings.HasSuffix(path, "/chat/completions") || strings.HasSuffix(path, "/responses") || openAIBaseURLHasVersionSuffix(path) {
		return fmt.Sprintf("API protocol is anthropic but base_url (%s) looks like an OpenAI-compatible endpoint; requests would hit {base}/v1/messages and 404. Set base_url to the provider's Anthropic endpoint (e.g. https://open.bigmodel.cn/api/anthropic) or switch api_protocol to chat_completions.", baseURL)
	}
	return ""
}
