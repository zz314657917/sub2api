package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	openAIWSClientReadLimitBytesDefault     int64 = 64 * 1024 * 1024
	openAIWSHTTPBridgeThresholdBytesDefault int64 = 15 * 1024 * 1024
	openAIWSHTTPBridgeErrorBodyLimitBytes         = 64 * 1024
)

const openAIWSHTTPBridgeToolStateContextKey = "openai_ws_http_bridge_tool_state"

type openAIWSHTTPBridgeToolState struct {
	ClientMapping apicompat.ResponsesClientToolMapping
	LoweredTools  json.RawMessage
}

func openAIWSHTTPBridgeToolStateFromContext(c *gin.Context) (openAIWSHTTPBridgeToolState, bool) {
	if c == nil {
		return openAIWSHTTPBridgeToolState{}, false
	}
	value, ok := c.Get(openAIWSHTTPBridgeToolStateContextKey)
	state, typed := value.(openAIWSHTTPBridgeToolState)
	return state, ok && typed
}

func setOpenAIWSHTTPBridgeToolState(c *gin.Context, state openAIWSHTTPBridgeToolState) {
	if c == nil {
		return
	}
	state.LoweredTools = append(json.RawMessage(nil), state.LoweredTools...)
	c.Set(openAIWSHTTPBridgeToolStateContextKey, state)
}

// ResolveOpenAIWSClientFirstMessageTimeout returns the effective client ingress deadline.
func ResolveOpenAIWSClientFirstMessageTimeout(cfg *config.Config) time.Duration {
	seconds := config.DefaultOpenAIWSClientFirstMessageTimeoutSeconds
	if cfg != nil && cfg.Gateway.OpenAIWS.ClientFirstMessageTimeoutSeconds > 0 {
		seconds = cfg.Gateway.OpenAIWS.ClientFirstMessageTimeoutSeconds
	}
	return time.Duration(seconds) * time.Second
}

func ResolveOpenAIWSClientReadLimitBytes(cfg *config.Config) int64 {
	if cfg == nil || cfg.Gateway.OpenAIWS.ClientReadLimitBytes <= 0 {
		return openAIWSClientReadLimitBytesDefault
	}
	return cfg.Gateway.OpenAIWS.ClientReadLimitBytes
}

func (s *OpenAIGatewayService) openAIWSHTTPBridgeEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.Gateway.OpenAIWS.HTTPBridgeEnabled
}

func (s *OpenAIGatewayService) openAIWSHTTPBridgeThresholdBytes() int64 {
	if s == nil || s.cfg == nil || s.cfg.Gateway.OpenAIWS.HTTPBridgeThresholdBytes <= 0 {
		return openAIWSHTTPBridgeThresholdBytesDefault
	}
	return s.cfg.Gateway.OpenAIWS.HTTPBridgeThresholdBytes
}

func (s *OpenAIGatewayService) shouldBridgeOpenAIWSHTTP(account *Account, payloadBytes int, previousResponseID string) bool {
	if account != nil && account.Platform == PlatformGrok {
		return true
	}
	if !s.openAIWSHTTPBridgeEnabled() {
		return false
	}
	if strings.TrimSpace(previousResponseID) != "" {
		return false
	}
	threshold := s.openAIWSHTTPBridgeThresholdBytes()
	return threshold > 0 && int64(payloadBytes) >= threshold
}

func prepareOpenAIWSHTTPBridgeBody(payload []byte) ([]byte, error) {
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, err
	}
	if body == nil {
		return nil, errors.New("response.create payload must be a JSON object")
	}
	delete(body, "type")
	delete(body, "generate")
	delete(body, "previous_response_id")
	body["stream"] = true
	return json.Marshal(body)
}

type openAIWSToolCallReplayCollector struct {
	items    []json.RawMessage
	seen     map[string]struct{}
	allItems []json.RawMessage
	allSeen  map[string]struct{}
}

func (c *openAIWSToolCallReplayCollector) AddEvent(eventType string, message []byte) {
	switch strings.TrimSpace(eventType) {
	case "response.output_item.done":
		item := gjson.GetBytes(message, "item")
		c.addAllItem(item)
		c.addItem(item)
	case "response.completed", "response.done":
		output := gjson.GetBytes(message, "response.output")
		if !output.IsArray() {
			return
		}
		for _, item := range output.Array() {
			c.addAllItem(item)
			c.addItem(item)
		}
	}
}

func (c *openAIWSToolCallReplayCollector) Items() []json.RawMessage {
	return cloneOpenAIWSRawMessages(c.items)
}

func (c *openAIWSToolCallReplayCollector) AllItems() []json.RawMessage {
	return cloneOpenAIWSRawMessages(c.allItems)
}

func (c *openAIWSToolCallReplayCollector) addAllItem(item gjson.Result) {
	if !item.Exists() || item.Type != gjson.JSON {
		return
	}
	raw := strings.TrimSpace(item.Raw)
	if raw == "" || !strings.HasPrefix(raw, "{") || strings.TrimSpace(item.Get("type").String()) == "" {
		return
	}
	key := strings.TrimSpace(item.Get("id").String())
	if key == "" {
		key = strings.TrimSpace(item.Get("call_id").String())
	}
	if key == "" {
		key = raw
	}
	if c.allSeen == nil {
		c.allSeen = make(map[string]struct{})
	}
	if _, ok := c.allSeen[key]; ok {
		return
	}
	c.allSeen[key] = struct{}{}
	c.allItems = append(c.allItems, json.RawMessage(raw))
}

func (c *openAIWSToolCallReplayCollector) addItem(item gjson.Result) {
	if !item.Exists() || item.Type != gjson.JSON {
		return
	}
	raw := strings.TrimSpace(item.Raw)
	if raw == "" || !strings.HasPrefix(raw, "{") {
		return
	}
	if !isCodexToolCallContextItemType(item.Get("type").String()) {
		return
	}
	key := strings.TrimSpace(item.Get("id").String())
	if key == "" {
		key = strings.TrimSpace(item.Get("call_id").String())
	}
	if key == "" {
		key = raw
	}
	if c.seen == nil {
		c.seen = make(map[string]struct{})
	}
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}
	c.items = append(c.items, json.RawMessage(raw))
}

func buildOpenAIWSHTTPBridgeErrorEvent(statusCode int, message string) []byte {
	message = strings.TrimSpace(message)
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if message == "" {
		message = "upstream request failed"
	}
	event := map[string]any{
		"type":   "error",
		"status": statusCode,
		"error": map[string]any{
			"type":    "upstream_error",
			"message": message,
		},
	}
	body, err := json.Marshal(event)
	if err != nil {
		return []byte(`{"type":"error","error":{"type":"upstream_error","message":"upstream request failed"}}`)
	}
	return body
}

func (s *OpenAIGatewayService) proxyOpenAIWSHTTPBridgeTurn(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	token string,
	payload []byte,
	payloadBytes int,
	originalModel string,
	imageBillingModel string,
	imageSizeTier string,
	imageQuality string,
	imageInputSize string,
	turn int,
	writeClientMessage func([]byte) error,
) (*OpenAIForwardResult, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	if s.httpUpstream == nil {
		return nil, errors.New("openai http upstream is nil")
	}
	if account == nil {
		return nil, errors.New("account is nil")
	}
	if writeClientMessage == nil {
		return nil, errors.New("client websocket writer is nil")
	}

	body, err := prepareOpenAIWSHTTPBridgeBody(payload)
	if err != nil {
		return nil, fmt.Errorf("prepare http bridge body: %w", err)
	}

	upstreamCtx, releaseUpstreamCtx := detachUpstreamContext(ctx)
	var upstreamReq *http.Request
	var clientToolMapping apicompat.ResponsesClientToolMapping
	if account.Platform == PlatformGrok {
		upstreamModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
		if originalModel != "" {
			if mappedModel := normalizeOpenAIModelForUpstream(account, account.GetMappedModel(originalModel)); mappedModel != "" {
				upstreamModel = mappedModel
			}
		}
		if upstreamModel == "" {
			upstreamModel = "grok-4.3"
		}
		body, err = patchGrokResponsesBody(body, upstreamModel)
		if err != nil {
			releaseUpstreamCtx()
			return nil, err
		}
		upstreamReq, err = buildGrokResponsesRequest(upstreamCtx, c, account, body, token)
	} else {
		inheritedState, _ := openAIWSHTTPBridgeToolStateFromContext(c)
		toolsPresent := gjson.GetBytes(body, "tools").Exists()
		body, clientToolMapping, err = adaptResponsesClientToolsForFunctionUpstreamWithMapping(body, "OpenAI WS HTTP bridge", inheritedState.ClientMapping)
		if err != nil {
			releaseUpstreamCtx()
			return nil, fmt.Errorf("adapt OpenAI WS HTTP bridge client tools: %w", err)
		}
		loweredTools := inheritedState.LoweredTools
		if toolsPresent {
			loweredTools = json.RawMessage(gjson.GetBytes(body, "tools").Raw)
		} else if len(loweredTools) > 0 {
			body, err = sjson.SetRawBytes(body, "tools", loweredTools)
			if err != nil {
				releaseUpstreamCtx()
				return nil, fmt.Errorf("inherit OpenAI WS HTTP bridge tools: %w", err)
			}
		}
		setOpenAIWSHTTPBridgeToolState(c, openAIWSHTTPBridgeToolState{ClientMapping: clientToolMapping, LoweredTools: loweredTools})
		upstreamReq, err = s.buildUpstreamRequestOpenAIPassthrough(upstreamCtx, c, account, body, token)
	}
	if err == nil && account.Platform != PlatformGrok && isOpenAIResponsesLiteWebSocketPayload(payload) {
		upstreamReq.Header.Set(responsesLiteHeader, "true")
	}
	releaseUpstreamCtx()
	if err != nil {
		return nil, err
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}
	if c != nil {
		c.Set("openai_passthrough", true)
		c.Set("openai_ws_http_bridge", true)
	}

	turnStart := time.Now()
	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		_ = writeClientMessage(buildOpenAIWSHTTPBridgeErrorEvent(http.StatusBadGateway, "Upstream request failed"))
		return nil, fmt.Errorf("upstream http bridge request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, openAIWSHTTPBridgeErrorBodyLimitBytes))
		upstreamMsg := sanitizeUpstreamErrorMessage(strings.TrimSpace(extractUpstreamErrorMessage(respBody)))
		if upstreamMsg == "" {
			upstreamMsg = http.StatusText(resp.StatusCode)
		}
		// A later bridge turn has no safe client-visible fallback after a 429.
		// Return a failover error before writing an error event so the handler can
		// switch accounts and replay the current turn with rebuilt context.
		if turn > 1 && account.Platform == PlatformOpenAI && resp.StatusCode == http.StatusTooManyRequests {
			if failoverErr := s.failoverOpenAIUpstreamHTTPError(ctx, c, account, resp, respBody, upstreamMsg, originalModel); failoverErr != nil {
				return nil, failoverErr
			}
		}
		if account.Platform != PlatformGrok {
			s.handleOpenAIAccountUpstreamError(ctx, account, resp.StatusCode, resp.Header, respBody, originalModel)
		}
		_ = writeClientMessage(buildOpenAIWSHTTPBridgeErrorEvent(resp.StatusCode, upstreamMsg))
		return nil, fmt.Errorf("upstream http bridge error: status=%d message=%s", resp.StatusCode, upstreamMsg)
	}

	responseID := ""
	usage := OpenAIUsage{}
	imageCounter := newOpenAIImageOutputCounter()
	var firstTokenMs *int
	reqStream := openAIWSPayloadBoolFromRaw(body, "stream", true)
	eventCount := 0
	tokenEventCount := 0
	terminalEventCount := 0
	replayCollector := &openAIWSToolCallReplayCollector{}
	firstEventType := ""
	lastEventType := ""
	upstreamTerminalEvent := ""
	sawDone := false
	wroteDownstream := false
	clientDisconnected := false
	mappedModel := ""
	needModelReplace := false
	var mappedModelBytes []byte
	if originalModel != "" {
		mappedModel = normalizeOpenAIModelForUpstream(account, account.GetMappedModel(originalModel))
		needModelReplace = mappedModel != "" && mappedModel != originalModel
		if needModelReplace {
			mappedModelBytes = []byte(mappedModel)
		}
	}

	resultWithUsage := func() *OpenAIForwardResult {
		imageCount := imageCounter.Count()
		result := &OpenAIForwardResult{
			RequestID:             responseID,
			Usage:                 usage,
			Model:                 originalModel,
			UpstreamModel:         mappedModel,
			ServiceTier:           extractOpenAIServiceTierFromBody(body),
			ReasoningEffort:       ApplyThinkingEnabledFallback(extractOpenAIReasoningEffortFromBody(body, mappedModel, originalModel), body, mappedModel),
			Stream:                reqStream,
			OpenAIWSMode:          true,
			UpstreamTerminalEvent: upstreamTerminalEvent,
			ResponseHeaders:       cloneHeader(resp.Header),
			Duration:              time.Since(turnStart),
			FirstTokenMs:          firstTokenMs,
		}
		if replayInput := replayCollector.Items(); len(replayInput) > 0 {
			result.wsReplayInput = replayInput
			result.wsReplayInputExists = true
		}
		result.wsAccountFailoverReplayInput = replayCollector.AllItems()
		if imageCount > 0 {
			result.ImageCount = imageCount
			result.ImageSize = imageSizeTier
			result.ImageQuality = imageQuality
			result.ImageInputSize = imageInputSize
			result.ImageOutputSizes = imageCounter.Sizes()
			result.BillingModel = imageBillingModel
		}
		return result
	}

	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	if hasOpenAIResponsesClientToolMapping(clientToolMapping) {
		resp.Body = newResponsesClientToolStreamBody(resp.Body, clientToolMapping, maxLineSize)
	}
	if account.Platform == PlatformGrok {
		resp.Body = newGrokResponsesBillingPingFilterBody(resp.Body, account, maxLineSize)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanBuf := getSSEScannerBuf64K()
	scanner.Buffer(scanBuf[:0], maxLineSize)
	defer putSSEScannerBuf64K(scanBuf)

	for scanner.Scan() {
		line := scanner.Text()
		data, ok := extractOpenAISSEDataLine(line)
		if !ok {
			continue
		}
		trimmedData := strings.TrimSpace(data)
		if trimmedData == "" {
			continue
		}
		if trimmedData == "[DONE]" {
			sawDone = true
			continue
		}

		upstreamMessage := []byte(trimmedData)
		eventType, eventResponseID, _ := parseOpenAIWSEventEnvelope(upstreamMessage)
		if responseID == "" && eventResponseID != "" {
			responseID = eventResponseID
		}
		if eventType != "" {
			eventCount++
			if firstEventType == "" {
				firstEventType = eventType
			}
			lastEventType = eventType
		}
		if isOpenAIWSTokenEvent(eventType) {
			tokenEventCount++
			if firstTokenMs == nil {
				ms := int(time.Since(turnStart).Milliseconds())
				firstTokenMs = &ms
			}
		}
		if openAIWSEventShouldParseUsage(eventType) {
			parseOpenAIWSResponseUsageFromCompletedEvent(upstreamMessage, &usage)
		}
		imageCounter.AddSSEData(upstreamMessage)

		if needModelReplace && len(mappedModelBytes) > 0 && openAIWSEventMayContainModel(eventType) && strings.Contains(trimmedData, mappedModel) {
			upstreamMessage = replaceOpenAIWSMessageModel(upstreamMessage, mappedModel, originalModel)
		}
		if s.toolCorrector != nil && openAIWSEventMayContainToolCalls(eventType) && openAIWSMessageLikelyContainsToolCalls(upstreamMessage) {
			if corrected, changed := s.toolCorrector.CorrectToolCallsInSSEBytes(upstreamMessage); changed {
				upstreamMessage = corrected
			}
		}
		if normalized, changed := normalizeOpenAIResponsesCustomToolNamespaces(upstreamMessage); changed {
			upstreamMessage = normalized
		}
		replayCollector.AddEvent(eventType, upstreamMessage)

		// A later-turn semantic rate limit must be retried on another account
		// before its error event reaches the client. Once any upstream output has
		// been written, preserving the live stream takes precedence and the
		// existing non-failover behavior remains unchanged.
		if turn > 1 && !wroteDownstream && (eventType == "error" || eventType == "response.failed") {
			errMessage := strings.TrimSpace(extractOpenAISSEErrorMessage(upstreamMessage))
			if errMessage == "" {
				errMessage = "upstream error event"
			}
			statusCode := openAIStreamFailureStatus(upstreamMessage, errMessage)
			if eventType == "error" {
				errCodeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(upstreamMessage)
				statusCode = openAIWSErrorHTTPStatusFromRaw(errCodeRaw, errTypeRaw)
				if statusCode == http.StatusTooManyRequests {
					s.persistOpenAIWSRateLimitSignal(ctx, account, resp.Header, upstreamMessage, errCodeRaw, errTypeRaw, errMsgRaw, mappedModel)
				}
			} else if statusCode == http.StatusTooManyRequests {
				s.handleOpenAIWSTerminalTransientFailure(ctx, account, mappedModel, resp.Header, upstreamMessage)
			}
			if statusCode == http.StatusTooManyRequests {
				return nil, s.newOpenAIStreamFailoverError(
					c,
					account,
					true,
					resp.Header.Get("x-request-id"),
					upstreamMessage,
					errMessage,
					resp.Header,
				)
			}
		}

		if !clientDisconnected {
			if err := writeClientMessage(upstreamMessage); err != nil {
				if isOpenAIWSClientDisconnectError(err) {
					clientDisconnected = true
					closeStatus, closeReason := summarizeOpenAIWSReadCloseError(err)
					logOpenAIWSModeInfo(
						"ingress_ws_http_bridge_client_disconnected_drain account_id=%d turn=%d close_status=%s close_reason=%s",
						account.ID,
						turn,
						closeStatus,
						truncateOpenAIWSLogValue(closeReason, openAIWSHeaderValueMaxLen),
					)
				} else {
					return nil, wrapOpenAIWSIngressTurnError(
						"write_client",
						fmt.Errorf("write client websocket event: %w", err),
						wroteDownstream,
					)
				}
			} else {
				wroteDownstream = true
			}
		}

		if eventType == "error" {
			s.handleOpenAIWSErrorEventTransientFailure(ctx, account, mappedModel, resp.Header, upstreamMessage)
			errCodeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(upstreamMessage)
			s.persistOpenAIWSRateLimitSignal(ctx, account, resp.Header, upstreamMessage, errCodeRaw, errTypeRaw, errMsgRaw, mappedModel)
			errMessage := strings.TrimSpace(errMsgRaw)
			if errMessage == "" {
				errMessage = "upstream error event"
			}
			return resultWithUsage(), errors.New(errMessage)
		}
		if isOpenAIWSTerminalEvent(eventType) {
			terminalEventCount++
			upstreamTerminalEvent = s.handleOpenAIWSTerminalTransientFailure(ctx, account, mappedModel, resp.Header, upstreamMessage)
			firstTokenMsValue := -1
			if firstTokenMs != nil {
				firstTokenMsValue = *firstTokenMs
			}
			logOpenAIWSModeInfo(
				"ingress_ws_http_bridge_turn_completed account_id=%d turn=%d response_id=%s payload_bytes=%d duration_ms=%d events=%d token_events=%d terminal_events=%d first_event=%s last_event=%s first_token_ms=%d client_disconnected=%v",
				account.ID,
				turn,
				truncateOpenAIWSLogValue(responseID, openAIWSIDValueMaxLen),
				payloadBytes,
				time.Since(turnStart).Milliseconds(),
				eventCount,
				tokenEventCount,
				terminalEventCount,
				truncateOpenAIWSLogValue(firstEventType, openAIWSLogValueMaxLen),
				truncateOpenAIWSLogValue(lastEventType, openAIWSLogValueMaxLen),
				firstTokenMsValue,
				clientDisconnected,
			)
			return resultWithUsage(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return resultWithUsage(), fmt.Errorf("read upstream http bridge stream: %w", err)
	}
	if sawDone && eventCount > 0 {
		return resultWithUsage(), nil
	}
	return resultWithUsage(), errors.New("upstream http bridge stream ended before terminal event")
}
