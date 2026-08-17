package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
)

const (
	cnExtraSuffix5hUsed       = "5h_used_percent"
	cnExtraSuffix5hReset      = "5h_reset_at"
	cnExtraSuffixWeeklyUsed   = "weekly_used_percent"
	cnExtraSuffixWeeklyReset  = "weekly_reset_at"
	cnExtraSuffixUsageUpdated = "usage_updated_at"
	cnQuotaMaxBodyBytes       = 256 * 1024
	cnProbeUpstreamTimeout    = 15 * time.Second
)

func cnExtraKey(provider, suffix string) string { return provider + "_" + suffix }

type CNQuotaTier struct {
	Window      string  `json:"window"`
	UsedPercent float64 `json:"used_percent"`
	ResetAt     string  `json:"reset_at,omitempty"`
}

type CNProviderQuotaProbeResult struct {
	Provider        string        `json:"provider"`
	Source          string        `json:"source"`
	Success         bool          `json:"success"`
	CredentialValid bool          `json:"credential_valid"`
	Tiers           []CNQuotaTier `json:"tiers,omitempty"`
	PlanLevel       string        `json:"plan_level,omitempty"`
	StatusCode      int           `json:"status_code,omitempty"`
	FetchedAt       int64         `json:"fetched_at"`
	Persisted       bool          `json:"persisted"`
	Error           string        `json:"error,omitempty"`
}

type CNProviderQuotaService struct {
	accountRepo  AccountRepository
	proxyRepo    ProxyRepository
	httpUpstream HTTPUpstream
	cfg          *config.Config
}

func NewCNProviderQuotaService(accountRepo AccountRepository, proxyRepo ProxyRepository, httpUpstream HTTPUpstream, cfg *config.Config) *CNProviderQuotaService {
	return &CNProviderQuotaService{accountRepo: accountRepo, proxyRepo: proxyRepo, httpUpstream: httpUpstream, cfg: cfg}
}

// QueryUsage refreshes only Kimi/Zhipu Coding Plan snapshots. Failed probes do
// not write Extra, preserving the previous snapshot.
func (s *CNProviderQuotaService) QueryUsage(ctx context.Context, accountID int64) (*CNProviderQuotaProbeResult, error) {
	if s == nil || s.accountRepo == nil || s.httpUpstream == nil {
		return nil, fmt.Errorf("cn provider quota service is not configured")
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, fmt.Errorf("cn provider account %d not found", accountID)
	}
	provider := cnCodingPlanProvider(account)
	if provider == "" {
		return nil, fmt.Errorf("account is not a kimi/zhipu coding plan account")
	}
	apiKey := cnProviderAPIKey(account)
	if apiKey == "" {
		return nil, fmt.Errorf("account api_key is empty")
	}
	targetURL, authorization := "", ""
	switch provider {
	case PlatformKimi:
		targetURL, authorization = kimiQuotaURL(account.GetOpenAIBaseURL()), "Bearer "+apiKey
	case PlatformZhipu:
		targetURL, authorization = zhipuQuotaURL(account.GetOpenAIBaseURL()), apiKey
	}
	// Validate the final endpoint before constructing a request that carries credentials.
	targetURL, err = cnValidateProbeURL(s.cfg, targetURL)
	if err != nil {
		return nil, fmt.Errorf("cn quota target rejected: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, cnProbeUpstreamTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(callCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build quota request: %w", err)
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpUpstream.Do(req, s.resolveProxyURL(ctx, account), account.ID, maxInt(account.Concurrency, 1))
	if err != nil {
		return nil, fmt.Errorf("quota upstream request: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("quota upstream response is empty")
	}
	defer resp.Body.Close()
	result := &CNProviderQuotaProbeResult{Provider: provider, Source: "coding_plan", StatusCode: resp.StatusCode, FetchedAt: time.Now().UTC().Unix()}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, cnQuotaMaxBodyBytes))
	if readErr != nil {
		result.Error = "response_read_failed"
		return result, nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		result.Error = fmt.Sprintf("authentication failed (HTTP %d)", resp.StatusCode)
		return result, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		result.Error = fmt.Sprintf("API error (HTTP %d)", resp.StatusCode)
		return result, nil
	}
	if !gjson.ValidBytes(body) {
		result.Error = "invalid_response"
		return result, nil
	}
	if provider == PlatformZhipu && gjson.GetBytes(body, "success").Exists() && !gjson.GetBytes(body, "success").Bool() {
		result.Error = "API error: " + strings.TrimSpace(gjson.GetBytes(body, "msg").String())
		return result, nil
	}
	if provider == PlatformKimi {
		result.Tiers = parseKimiUsageTiers(body)
	} else {
		data := gjson.GetBytes(body, "data")
		if !data.Exists() || !data.Get("limits").IsArray() {
			result.Error = "invalid_response"
			return result, nil
		}
		result.Tiers = parseZhipuTokenTiers(data)
		result.PlanLevel = strings.TrimSpace(data.Get("level").String())
	}
	if len(result.Tiers) == 0 {
		result.Error = "invalid_response"
		return result, nil
	}
	result.Success, result.CredentialValid = true, true
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, cnQuotaExtraUpdates(provider, result.Tiers, time.Now().UTC())); err == nil {
		result.Persisted = true
	}
	return result, nil
}

func (s *CNProviderQuotaService) resolveProxyURL(ctx context.Context, account *Account) string {
	if account == nil || account.ProxyID == nil {
		return ""
	}
	if account.Proxy != nil {
		return account.Proxy.URL()
	}
	if s.proxyRepo == nil {
		return ""
	}
	proxy, err := s.proxyRepo.GetByID(ctx, *account.ProxyID)
	if err != nil || proxy == nil {
		return ""
	}
	return proxy.URL()
}

func cnAccountIsCodingPlan(account *Account) bool {
	return account != nil && account.GetAccountMode() == AccountModeCoding
}

func cnProviderAPIKey(account *Account) string {
	if account == nil {
		return ""
	}
	return strings.TrimSpace(account.GetCredential("api_key"))
}

func cnCodingPlanProvider(account *Account) string {
	if !cnAccountIsCodingPlan(account) {
		return ""
	}
	if account.Platform == PlatformKimi || account.Platform == PlatformZhipu {
		return account.Platform
	}
	return ""
}

func kimiQuotaURL(base string) string {
	return strings.TrimSuffix(strings.TrimRight(base, "/"), "/v1") + "/v1/usages"
}
func zhipuQuotaURL(base string) string {
	return zhipuQuotaHost(base) + "/api/monitor/usage/quota/limit"
}
func zhipuQuotaHost(base string) string {
	if strings.Contains(strings.ToLower(base), "z.ai") {
		return "https://api.z.ai"
	}
	return "https://open.bigmodel.cn"
}

func cnParseF64(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func parseKimiUsageTiers(body []byte) []CNQuotaTier {
	var tiers []CNQuotaTier
	gjson.GetBytes(body, "limits").ForEach(func(_, limit gjson.Result) bool {
		detail := limit.Get("detail")
		if !detail.Exists() {
			return true
		}
		max, _ := cnParseF64(detail.Get("limit").Value())
		remaining, _ := cnParseF64(detail.Get("remaining").Value())
		if max > 0 {
			used := (max - remaining) / max * 100
			if used < 0 {
				used = 0
			}
			tiers = append(tiers, CNQuotaTier{Window: "5h", UsedPercent: used, ResetAt: cnNormalizeResetTime(detail.Get("resetTime").Value())})
		}
		return false
	})
	usage := gjson.GetBytes(body, "usage")
	if usage.Exists() {
		max, _ := cnParseF64(usage.Get("limit").Value())
		remaining, _ := cnParseF64(usage.Get("remaining").Value())
		if max > 0 {
			used := (max - remaining) / max * 100
			if used < 0 {
				used = 0
			}
			tiers = append(tiers, CNQuotaTier{Window: "weekly", UsedPercent: used, ResetAt: cnNormalizeResetTime(usage.Get("resetTime").Value())})
		}
	}
	return tiers
}

func parseZhipuTokenTiers(data gjson.Result) []CNQuotaTier {
	type entry struct {
		unit     int64
		used     float64
		reset    string
		resetAt  int64
		hasReset bool
	}
	var knownFive, knownWeekly *entry
	var fallback []entry
	data.Get("limits").ForEach(func(_, item gjson.Result) bool {
		// CREDIT_LIMIT is a different metric and must never fill a token window.
		if strings.ToUpper(strings.TrimSpace(item.Get("type").String())) != "TOKENS_LIMIT" {
			return true
		}
		used, _ := cnParseF64(item.Get("percentage").Value())
		reset := cnNormalizeResetTime(item.Get("nextResetTime").Value())
		entry := entry{unit: item.Get("unit").Int(), used: used, reset: reset, hasReset: reset != ""}
		if entry.hasReset {
			if parsed, err := time.Parse(time.RFC3339, reset); err == nil {
				entry.resetAt = parsed.Unix()
			}
		}
		switch entry.unit {
		case 3:
			if knownFive == nil {
				knownFive = &entry
			} else {
				fallback = append(fallback, entry)
			}
		case 6:
			if knownWeekly == nil {
				knownWeekly = &entry
			} else {
				fallback = append(fallback, entry)
			}
		default:
			fallback = append(fallback, entry)
		}
		return true
	})
	// Unknown units use the upstream fallback: a bucket without reset is 5h
	// before reset-ordered buckets, rather than sorting known unit values.
	sort.SliceStable(fallback, func(i, j int) bool {
		if fallback[i].hasReset != fallback[j].hasReset {
			return !fallback[i].hasReset
		}
		return fallback[i].resetAt < fallback[j].resetAt
	})
	for _, entry := range fallback {
		if knownFive == nil {
			value := entry
			knownFive = &value
		} else if knownWeekly == nil {
			value := entry
			knownWeekly = &value
		}
	}
	tiers := make([]CNQuotaTier, 0, 2)
	if knownFive != nil {
		tiers = append(tiers, CNQuotaTier{Window: "5h", UsedPercent: knownFive.used, ResetAt: knownFive.reset})
	}
	if knownWeekly != nil {
		tiers = append(tiers, CNQuotaTier{Window: "weekly", UsedPercent: knownWeekly.used, ResetAt: knownWeekly.reset})
	}
	return tiers
}
func hasCNTier(tiers []CNQuotaTier, window string) bool {
	for _, tier := range tiers {
		if tier.Window == window {
			return true
		}
	}
	return false
}
func cnQuotaExtraUpdates(provider string, tiers []CNQuotaTier, now time.Time) map[string]any {
	updates := map[string]any{
		cnExtraKey(provider, cnExtraSuffixUsageUpdated): now.UTC().Format(time.RFC3339),
	}
	for _, tier := range tiers {
		switch tier.Window {
		case "5h":
			updates[cnExtraKey(provider, cnExtraSuffix5hUsed)] = tier.UsedPercent
			if tier.ResetAt != "" {
				updates[cnExtraKey(provider, cnExtraSuffix5hReset)] = tier.ResetAt
			}
		case "weekly":
			updates[cnExtraKey(provider, cnExtraSuffixWeeklyUsed)] = tier.UsedPercent
			if tier.ResetAt != "" {
				updates[cnExtraKey(provider, cnExtraSuffixWeeklyReset)] = tier.ResetAt
			}
		}
	}
	return updates
}
func cnNormalizeResetTime(raw any) string {
	if value, ok := cnParseF64(raw); ok && value > 0 {
		if value < 1_000_000_000_000 {
			value *= 1000
		}
		return time.UnixMilli(int64(value)).UTC().Format(time.RFC3339)
	}
	text, ok := raw.(string)
	if !ok {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(text))
	if err != nil {
		return ""
	}
	return parsed.UTC().Format(time.RFC3339)
}
