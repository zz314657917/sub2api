package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpguts"
)

const openAICodexRoutingHintHeader = "x-codex-routing-hint"

// setOpenAICodexRoutingHint creates the gateway-owned Codex routing hint from
// the final upstream model and service tier. The same function also removes
// caller/account supplied values on non-OAuth paths.
func setOpenAICodexRoutingHint(headers http.Header, account *Account, model string, serviceTier string) {
	if headers == nil {
		return
	}
	deleteOpenAIHeaderEqualFold(headers, openAICodexRoutingHintHeader)
	if account == nil || !account.IsOpenAIOAuth() {
		return
	}

	model = strings.TrimSpace(model)
	if model == "" || strings.ContainsAny(model, ";=") {
		return
	}

	canonicalTier := normalizedOpenAIServiceTierValue(serviceTier)
	switch canonicalTier {
	case OpenAIFastTierPriority, OpenAIFastTierFlex:
	default:
		canonicalTier = ""
	}

	hint := "model=" + model
	if canonicalTier != "" {
		hint += ";tier=" + canonicalTier
	}
	if !httpguts.ValidHeaderFieldValue(hint) {
		return
	}
	headers.Set(openAICodexRoutingHintHeader, hint)
}

func deleteOpenAIHeaderEqualFold(headers http.Header, name string) {
	if headers == nil {
		return
	}
	name = strings.TrimSpace(name)
	for key := range headers {
		if strings.EqualFold(strings.TrimSpace(key), name) {
			delete(headers, key)
		}
	}
}

func stripOpenAILegacyResponsesBeta(headers http.Header) {
	if headers == nil {
		return
	}
	preserved := make([]string, 0)
	for key, values := range headers {
		if !strings.EqualFold(strings.TrimSpace(key), "OpenAI-Beta") {
			continue
		}
		delete(headers, key)
		for _, value := range values {
			parts := strings.Split(value, ",")
			kept := parts[:0]
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" || strings.EqualFold(part, "responses=experimental") {
					continue
				}
				kept = append(kept, part)
			}
			if len(kept) > 0 {
				preserved = append(preserved, strings.Join(kept, ", "))
			}
		}
	}
	for _, value := range preserved {
		headers.Add("OpenAI-Beta", value)
	}
}

func setOpenAICodexRoutingHintFromBody(headers http.Header, account *Account, body []byte) {
	fields := gjson.GetManyBytes(body, "model", "service_tier")
	setOpenAICodexRoutingHint(headers, account, fields[0].String(), fields[1].String())
}

// logOpenAIRoutingDiagnostics records only gateway-derived state. It never
// includes raw headers, routing hint values, tokens, or credentials.
func logOpenAIRoutingDiagnostics(
	ctx context.Context,
	account *Account,
	transport string,
	model string,
	serviceTier string,
	hintGenerated bool,
	wsAffinityDecision string,
) {
	if ctx == nil {
		ctx = context.Background()
	}
	accountID := int64(0)
	if account != nil {
		accountID = account.ID
	}
	logger.FromContext(ctx).Debug("openai routing decision",
		zap.String("component", "service.openai_routing"),
		zap.String("transport", strings.TrimSpace(transport)),
		zap.Int64("account_id", accountID),
		zap.String("final_model", strings.TrimSpace(model)),
		zap.String("final_service_tier", normalizedOpenAIServiceTierValue(serviceTier)),
		zap.Bool("routing_hint_generated", hintGenerated),
		zap.String("ws_affinity_decision", strings.TrimSpace(wsAffinityDecision)),
	)
}

func logOpenAIRoutingDiagnosticsFromBody(
	ctx context.Context,
	account *Account,
	transport string,
	headers http.Header,
	body []byte,
	wsAffinityDecision string,
) {
	fields := gjson.GetManyBytes(body, "model", "service_tier")
	logOpenAIRoutingDiagnostics(
		ctx,
		account,
		transport,
		fields[0].String(),
		fields[1].String(),
		strings.TrimSpace(headers.Get(openAICodexRoutingHintHeader)) != "",
		wsAffinityDecision,
	)
}
