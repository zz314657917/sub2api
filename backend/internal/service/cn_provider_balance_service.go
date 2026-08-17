package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
)

const cnBalanceMaxBodyBytes = 256 * 1024

type CNProviderBalanceEntry struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}
type CNProviderBalanceResult struct {
	Provider   string                   `json:"provider"`
	Success    bool                     `json:"success"`
	Balance    float64                  `json:"balance"`
	Currency   string                   `json:"currency,omitempty"`
	Balances   []CNProviderBalanceEntry `json:"balances,omitempty"`
	Available  bool                     `json:"available"`
	StatusCode int                      `json:"status_code,omitempty"`
	FetchedAt  int64                    `json:"fetched_at"`
	Persisted  bool                     `json:"persisted"`
	Error      string                   `json:"error,omitempty"`
}
type CNProviderBalanceService struct {
	accountRepo  AccountRepository
	proxyRepo    ProxyRepository
	httpUpstream HTTPUpstream
	cfg          *config.Config
}

func NewCNProviderBalanceService(accountRepo AccountRepository, proxyRepo ProxyRepository, httpUpstream HTTPUpstream, cfg *config.Config) *CNProviderBalanceService {
	return &CNProviderBalanceService{accountRepo: accountRepo, proxyRepo: proxyRepo, httpUpstream: httpUpstream, cfg: cfg}
}
func (s *CNProviderBalanceService) QueryBalance(ctx context.Context, accountID int64) (*CNProviderBalanceResult, error) {
	if s == nil || s.accountRepo == nil || s.httpUpstream == nil {
		return nil, fmt.Errorf("cn provider balance service is not configured")
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return nil, fmt.Errorf("cn provider account %d not found", accountID)
	}
	if (account.Platform != PlatformKimi && account.Platform != PlatformDeepseek) || cnAccountIsCodingPlan(account) {
		return nil, fmt.Errorf("account has no CN payg balance endpoint")
	}
	apiKey := cnProviderAPIKey(account)
	if apiKey == "" {
		return nil, fmt.Errorf("account api_key is empty")
	}
	targetURL, err := cnValidateProbeURL(s.cfg, cnBalanceURL(account))
	if err != nil {
		return nil, fmt.Errorf("cn balance target rejected: %w", err)
	}
	callCtx, cancel := context.WithTimeout(ctx, cnProbeUpstreamTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(callCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build balance request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpUpstream.Do(req, s.resolveProxyURL(ctx, account), account.ID, maxInt(account.Concurrency, 1))
	if err != nil {
		return nil, fmt.Errorf("balance upstream request: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("balance upstream response is empty")
	}
	defer resp.Body.Close()
	result := &CNProviderBalanceResult{Provider: account.Platform, StatusCode: resp.StatusCode, FetchedAt: time.Now().UTC().Unix(), Available: true}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, cnBalanceMaxBodyBytes))
	if readErr != nil {
		result.Error = "response_read_failed"
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
	if account.Platform == PlatformKimi {
		value := gjson.GetBytes(body, "data.available_balance")
		balance, ok := cnParseF64(value.Value())
		if !value.Exists() || !ok {
			result.Error = "invalid_response"
			return result, nil
		}
		result.Balances = []CNProviderBalanceEntry{{Currency: "CNY", Balance: balance}}
	} else {
		if available := gjson.GetBytes(body, "is_available"); available.Exists() {
			result.Available = available.Bool()
		}
		balanceInfos := gjson.GetBytes(body, "balance_infos")
		if !balanceInfos.IsArray() {
			result.Error = "invalid_response"
			return result, nil
		}
		valid := true
		balanceInfos.ForEach(func(_, info gjson.Result) bool {
			balance, ok := cnParseF64(info.Get("total_balance").Value())
			if !ok {
				valid = false
				return false
			}
			currency := strings.ToUpper(strings.TrimSpace(info.Get("currency").String()))
			if currency == "" {
				currency = "CNY"
			}
			result.Balances = append(result.Balances, CNProviderBalanceEntry{Currency: currency, Balance: balance})
			return true
		})
		if !valid || len(result.Balances) == 0 {
			result.Error = "invalid_response"
			return result, nil
		}
	}
	result.Balance, result.Currency, result.Success = result.Balances[0].Balance, result.Balances[0].Currency, true
	balances := make([]any, 0, len(result.Balances))
	for _, entry := range result.Balances {
		balances = append(balances, map[string]any{"currency": entry.Currency, "balance": entry.Balance})
	}
	updates := map[string]any{cnExtraKey(result.Provider, "balance"): result.Balance, cnExtraKey(result.Provider, "balance_currency"): result.Currency, cnExtraKey(result.Provider, "balance_available"): result.Available, cnExtraKey(result.Provider, "balance_updated_at"): time.Now().UTC().Format(time.RFC3339), cnExtraKey(result.Provider, "balances"): balances}
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, updates); err == nil {
		result.Persisted = true
	}
	return result, nil
}
func (s *CNProviderBalanceService) resolveProxyURL(ctx context.Context, account *Account) string {
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
func cnBalanceURL(account *Account) string {
	if account == nil {
		return ""
	}
	switch account.Platform {
	case PlatformKimi:
		return "https://api.moonshot.cn/v1/users/me/balance"
	case PlatformDeepseek:
		return strings.TrimRight(account.GetOpenAIFormatBaseURL(), "/") + "/user/balance"
	}
	return ""
}
func allCNBalancesBelowThreshold(result *CNProviderBalanceResult, threshold float64) bool {
	if result == nil || len(result.Balances) == 0 {
		return result == nil || result.Balance < threshold
	}
	for _, entry := range result.Balances {
		if entry.Balance >= threshold {
			return false
		}
	}
	return true
}
