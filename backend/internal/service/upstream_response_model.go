package service

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	upstreamResponseModelObserverContextKey = "upstream_response_model_observer"
	upstreamResponseModelMaxLength          = 200
)

type upstreamResponseModelObserver struct {
	first, terminal string
	conflict        bool
}

func (o *upstreamResponseModelObserver) Observe(model string, terminal bool) {
	model = normalizeObservedUpstreamResponseModel(model)
	if model == "" {
		return
	}
	if current := o.Model(); current != "" && !strings.EqualFold(current, model) {
		o.conflict = true
	}
	if terminal {
		o.terminal = model
		return
	}
	if o.first == "" {
		o.first = model
	}
}
func normalizeObservedUpstreamResponseModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	runes := []rune(model)
	if len(runes) > upstreamResponseModelMaxLength {
		return string(runes[:upstreamResponseModelMaxLength])
	}
	return model
}
func (o *upstreamResponseModelObserver) ObserveOpenAI(payload []byte, eventType string) {
	o.Observe(firstValidTrimmedGJSONModel(payload, "response.model", "model"), isUpstreamResponseModelTerminalEvent(eventType))
}
func (o *upstreamResponseModelObserver) ObserveAnthropic(payload []byte) {
	o.Observe(firstValidTrimmedGJSONModel(payload, "message.model", "model"), false)
}
func (o *upstreamResponseModelObserver) ObserveGemini(payload []byte) {
	o.Observe(firstValidTrimmedGJSONModel(payload, "modelVersion", "response.modelVersion", "response.response.modelVersion"), true)
}
func (o *upstreamResponseModelObserver) Model() string {
	if o == nil {
		return ""
	}
	if o.terminal != "" {
		return o.terminal
	}
	return o.first
}
func (o *upstreamResponseModelObserver) Conflict() bool { return o != nil && o.conflict }
func beginUpstreamResponseModelObservation(c *gin.Context) *upstreamResponseModelObserver {
	o := &upstreamResponseModelObserver{}
	if c != nil {
		c.Set(upstreamResponseModelObserverContextKey, o)
	}
	return o
}
func upstreamResponseModelObserverFromContext(c *gin.Context) *upstreamResponseModelObserver {
	if c == nil {
		return nil
	}
	v, ok := c.Get(upstreamResponseModelObserverContextKey)
	if !ok {
		return nil
	}
	o, _ := v.(*upstreamResponseModelObserver)
	return o
}
func observedUpstreamResponseModel(c *gin.Context) string {
	return upstreamResponseModelObserverFromContext(c).Model()
}
func observedUpstreamResponseModelConflict(c *gin.Context) bool {
	return upstreamResponseModelObserverFromContext(c).Conflict()
}

func observeOpenAISSEBody(observer *upstreamResponseModelObserver, body string) {
	if observer == nil || strings.TrimSpace(body) == "" {
		return
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := []byte(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		observer.ObserveOpenAI(payload, strings.TrimSpace(gjson.GetBytes(payload, "type").String()))
	}
}
func firstValidTrimmedGJSONModel(payload []byte, paths ...string) string {
	if len(payload) == 0 {
		return ""
	}
	for _, path := range paths {
		value := gjson.GetBytes(payload, path)
		if !value.Exists() || value.Type != gjson.String {
			continue
		}
		if model := strings.TrimSpace(value.String()); model != "" {
			if !gjson.ValidBytes(payload) {
				return ""
			}
			return model
		}
	}
	return ""
}
func isUpstreamResponseModelTerminalEvent(eventType string) bool {
	switch strings.TrimSpace(eventType) {
	case "response.completed", "response.done", "response.failed", "response.incomplete", "response.cancelled", "response.canceled":
		return true
	}
	return false
}
func upstreamModelMismatch(sentModel, responseModel string) *bool {
	responseModel = strings.TrimSpace(responseModel)
	if responseModel == "" {
		return nil
	}
	mismatch := strings.TrimSpace(sentModel) == "" || !upstreamModelsMatchForAudit(sentModel, responseModel)
	return &mismatch
}
func upstreamModelsMatchForAudit(sentModel, responseModel string) bool {
	if strings.EqualFold(sentModel, responseModel) {
		return true
	}
	sent := canonicalGrokBuildRuntimeModel(sentModel)
	return sent != "" && sent == canonicalGrokBuildRuntimeModel(responseModel)
}
func canonicalGrokBuildRuntimeModel(model string) string {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "grok-4.5", "grok-4.5-latest", "grok-4.5-build":
		return "grok-4.5-build"
	case "grok-4.6", "grok-4.6-latest", "grok-4.6-build":
		return "grok-4.6-build"
	}
	return ""
}
func upstreamSentModel(requestedModel, upstreamModel string) string {
	if sent := strings.TrimSpace(upstreamModel); sent != "" {
		return sent
	}
	return strings.TrimSpace(requestedModel)
}
