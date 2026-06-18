package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/tidwall/gjson"
)

// openaiResponsesProbeTimeout 是探测请求的超时时长。
// 探测在后台异步执行，不阻塞账号创建/更新；留出余量给推理型模型先思考再产出工具调用。
const openaiResponsesProbeTimeout = 15 * time.Second

const responsesProbeMaxBodyBytes = 256 * 1024

// openaiResponsesProbePayload 构造探测使用的 Responses 请求体。
// 携带一个必选工具，只有响应 output 中出现 function_call 才视为工具路径可用。
//
// 注意：探测的目标是区分"端点存在"与"端点不存在"——只要上游返回非 404 的
// 4xx/5xx（如 400 invalid_request_error / 401 unauthorized / 422 等），
// 都视为"端点存在 → 支持 Responses"。仅 404 / 405 视为"端点不存在"。
func openaiResponsesProbePayload(modelID string) []byte {
	if strings.TrimSpace(modelID) == "" {
		modelID = openai.DefaultTestModel
	}
	body, _ := json.Marshal(map[string]any{
		"model": modelID,
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": "Call the probe_ping function with ok=true to acknowledge readiness. You must use the tool."},
				},
			},
		},
		"tools": []map[string]any{
			{
				"type":        "function",
				"name":        "probe_ping",
				"description": "Capability probe. Call to acknowledge.",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"ok": map[string]any{"type": "boolean"},
					},
					"required": []string{"ok"},
				},
			},
		},
		"tool_choice":       "required",
		"max_output_tokens": 512,
		"stream":            false,
	})
	return body
}

func selectResponsesProbeModel(account *Account) string {
	if account == nil {
		return openai.DefaultTestModel
	}
	mapping := account.GetModelMapping()
	candidates := make([]string, 0, len(mapping))
	for _, upstream := range mapping {
		upstream = strings.TrimSpace(upstream)
		if upstream == "" || strings.Contains(upstream, "*") {
			continue
		}
		candidates = append(candidates, upstream)
	}
	if len(candidates) == 0 {
		return openai.DefaultTestModel
	}
	sort.Strings(candidates)
	return candidates[0]
}

// ProbeOpenAIAPIKeyResponsesSupport 探测 OpenAI APIKey 账号上游是否支持
// /v1/responses 端点，并将结果持久化到 accounts.extra.openai_responses_supported。
//
// 调用时机：账号创建/更新后，且仅当 platform=openai && type=apikey 时。
//
// 探测策略（参见包文档 internal/pkg/openai_compat）：
//   - 上游 404 / 405 → 不支持，写 false
//   - 上游 2xx → 响应 output 含 function_call 才写 true，否则写 false
//   - 其他 4xx（401/422/400 等）/ 5xx → 端点存在但无法判定工具能力，保守写 true
//   - 网络层失败（连接错误、超时）→ 不写标记，保持 unknown
//     （后续请求仍按"现状即证据"默认走 Responses）
//
// 该方法是幂等的：重复调用会以最新探测结果覆盖标记。
//
// 关于失败处理：探测本身的失败不应阻塞账号创建——账号能创建/更新成功就够了，
// 探测结果只影响后续路由优化。所有错误都仅记录日志，不向调用方传播。
func (s *AccountTestService) ProbeOpenAIAPIKeyResponsesSupport(ctx context.Context, accountID int64) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_load_account_failed: account_id=%d err=%v", accountID, err)
		return
	}
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeAPIKey {
		// 仅 OpenAI APIKey 账号需要探测；其他账号类型无能力差异。
		return
	}

	apiKey := account.GetOpenAIApiKey()
	if apiKey == "" {
		logger.LegacyPrintf("service.openai_probe", "probe_skip_no_apikey: account_id=%d", accountID)
		return
	}
	baseURL := account.GetOpenAIBaseURL()
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_invalid_baseurl: account_id=%d base_url=%q err=%v", accountID, baseURL, err)
		return
	}

	probeURL := buildOpenAIResponsesURL(normalizedBaseURL)
	probeModel := selectResponsesProbeModel(account)

	probeCtx, cancel := context.WithTimeout(ctx, openaiResponsesProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodPost, probeURL, bytes.NewReader(openaiResponsesProbePayload(probeModel)))
	if err != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_build_request_failed: account_id=%d err=%v", accountID, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, s.tlsFPProfileService.ResolveTLSProfile(account))
	if err != nil {
		// 网络层失败：不写标记，保持 unknown，下次重试或由网关 fallback 处理
		logger.LegacyPrintf("service.openai_probe", "probe_request_failed: account_id=%d url=%s err=%v", accountID, probeURL, err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, responsesProbeMaxBodyBytes))
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, responsesProbeMaxBodyBytes))
	if readErr != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_read_body_failed: account_id=%d url=%s err=%v", accountID, probeURL, readErr)
		return
	}

	supported := decideResponsesProbeSupport(resp.StatusCode, bodyBytes)

	if err := s.accountRepo.UpdateExtra(ctx, accountID, map[string]any{
		openai_compat.ExtraKeyResponsesSupported: supported,
	}); err != nil {
		logger.LegacyPrintf("service.openai_probe", "probe_persist_failed: account_id=%d supported=%v err=%v", accountID, supported, err)
		return
	}

	logger.LegacyPrintf("service.openai_probe",
		"probe_done: account_id=%d base_url=%s probe_model=%s status=%d supported=%v",
		accountID, normalizedBaseURL, probeModel, resp.StatusCode, supported,
	)
}

// isResponsesEndpointSupportedByStatus 根据探测响应的 HTTP 状态码判定上游
// 是否暴露 /v1/responses 端点。
//
// 关键观察：第三方 OpenAI 兼容上游（DeepSeek/Kimi 等）对未知端点统一返回 404
// 或 405；而 OpenAI 官方/有 Responses 实现的上游会因为请求体最简（缺字段）
// 返回 400/422 等业务错误，但端点本身存在。
//
// 因此：仅 404 和 405 视为"端点不存在"，其他 status 视为"端点存在"。
//
// 5xx 也视为"端点存在"——上游偶发故障不应误判为不支持。
func isResponsesEndpointSupportedByStatus(status int) bool {
	switch status {
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return false
	}
	return true
}

func decideResponsesProbeSupport(status int, body []byte) bool {
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		return false
	}
	if status < 200 || status >= 300 {
		return true
	}
	return responsesProbeBodyHasFunctionCall(body)
}

func responsesProbeBodyHasFunctionCall(body []byte) bool {
	output := gjson.GetBytes(body, "output")
	if !output.IsArray() {
		return false
	}
	for _, item := range output.Array() {
		if strings.TrimSpace(item.Get("type").String()) == "function_call" {
			return true
		}
	}
	return false
}
