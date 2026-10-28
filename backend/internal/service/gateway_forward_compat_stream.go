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
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const anthropicDisconnectDrainLimit = 180 * time.Second

var (
	errGatewayAnthropicIdle  = errors.New("stream data interval timeout")
	errGatewayAnthropicDrain = errors.New("stream disconnect drain deadline exceeded")
)

type gatewayAnthropicLineEvent struct {
	line string
	err  error
}

type anthropicBillingBody struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func (b *anthropicBillingBody) Close() error {
	b.once.Do(func() { b.err = b.ReadCloser.Close() })
	return b.err
}

// This reader belongs to GatewayService only. The OpenAI native pump ABI is unchanged.
// All timers and Gin state belong to the consuming goroutine.
type gatewayAnthropicLinePump struct {
	events       chan gatewayAnthropicLineEvent
	done         chan struct{}
	exited       chan struct{}
	body         io.ReadCloser
	idle         *time.Timer
	deadline     *time.Timer
	interval     time.Duration
	drainIdle    time.Duration
	drainLimit   time.Duration
	disconnected bool
}

func newGatewayAnthropicLinePump(body io.ReadCloser, maxLine int, interval, drainIdle, drainLimit time.Duration) *gatewayAnthropicLinePump {
	p := &gatewayAnthropicLinePump{
		events: make(chan gatewayAnthropicLineEvent, 16), done: make(chan struct{}),
		exited: make(chan struct{}), body: body, interval: interval,
		drainIdle: drainIdle, drainLimit: drainLimit,
	}
	if interval > 0 {
		p.idle = time.NewTimer(interval)
	}
	go func() {
		defer close(p.exited)
		defer close(p.events)
		scanner := bufio.NewScanner(body)
		buf := getSSEScannerBuf64K()
		defer putSSEScannerBuf64K(buf)
		scanner.Buffer(buf[:0], maxLine)
		send := func(ev gatewayAnthropicLineEvent) bool {
			select {
			case p.events <- ev:
				return true
			case <-p.done:
				return false
			}
		}
		for scanner.Scan() {
			if !send(gatewayAnthropicLineEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			send(gatewayAnthropicLineEvent{err: err})
		}
	}()
	return p
}

func (p *gatewayAnthropicLinePump) idleC() <-chan time.Time {
	if p.idle != nil {
		return p.idle.C
	}
	return nil
}

func (p *gatewayAnthropicLinePump) deadlineC() <-chan time.Time {
	if p.deadline != nil {
		return p.deadline.C
	}
	return nil
}

func (p *gatewayAnthropicLinePump) readLine() {
	if p.idle != nil {
		p.idle.Reset(p.interval)
	}
}

func (p *gatewayAnthropicLinePump) disconnect() {
	if p.disconnected {
		return
	}
	p.disconnected = true
	if p.interval <= 0 {
		p.interval = p.drainIdle
	}
	if p.idle == nil {
		p.idle = time.NewTimer(p.interval)
	} else {
		p.idle.Reset(p.interval)
	}
	p.deadline = time.NewTimer(p.drainLimit)
}

func (p *gatewayAnthropicLinePump) stop() {
	close(p.done)
	_ = p.body.Close()
	if p.idle != nil {
		p.idle.Stop()
	}
	if p.deadline != nil {
		p.deadline.Stop()
	}
	<-p.exited
}

func (s *GatewayService) newAnthropicBillingPump(resp *http.Response) *gatewayAnthropicLinePump {
	maxLine, gap := defaultMaxLineSize, time.Duration(0)
	if s.cfg != nil {
		if s.cfg.Gateway.MaxLineSize > 0 {
			maxLine = s.cfg.Gateway.MaxLineSize
		}
		if s.cfg.Gateway.StreamDataIntervalTimeout > 0 {
			gap = time.Duration(s.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
		}
	}
	if _, ok := resp.Body.(*anthropicBillingBody); !ok {
		resp.Body = &anthropicBillingBody{ReadCloser: resp.Body}
	}
	return newGatewayAnthropicLinePump(resp.Body, maxLine, gap, anthropicDisconnectDrainLimit, anthropicDisconnectDrainLimit)
}

func anthropicClientGone(c *gin.Context) <-chan struct{} {
	if c != nil && c.Request != nil {
		return c.Request.Context().Done()
	}
	return nil
}

func anthropicBillingClientCanceled(ctx context.Context, c *gin.Context) bool {
	return (ctx != nil && ctx.Err() != nil) ||
		(c != nil && c.Request != nil && c.Request.Context().Err() != nil)
}

type anthropicCompatBillingStream struct {
	pump         *gatewayAnthropicLinePump
	clientGone   <-chan struct{}
	disconnected bool
}

func (s *GatewayService) newAnthropicCompatBillingStream(resp *http.Response, c *gin.Context) *anthropicCompatBillingStream {
	return &anthropicCompatBillingStream{pump: s.newAnthropicBillingPump(resp), clientGone: anthropicClientGone(c)}
}

func (st *anthropicCompatBillingStream) checkClient() {
	if st.clientGone != nil {
		select {
		case <-st.clientGone:
			st.disconnect()
		default:
		}
	}
}

func (st *anthropicCompatBillingStream) disconnect() {
	st.disconnected = true
	st.clientGone = nil
	st.pump.disconnect()
}

func (st *anthropicCompatBillingStream) next() (string, error) {
	for {
		st.checkClient()
		// The total drain deadline also applies while the upstream continuously sends.
		select {
		case <-st.pump.deadlineC():
			return "", errGatewayAnthropicDrain
		default:
		}
		select {
		case ev, ok := <-st.pump.events:
			if !ok {
				return "", io.EOF
			}
			st.pump.readLine()
			return ev.line, ev.err
		case <-st.clientGone:
			st.disconnect()
		case <-st.pump.idleC():
			// Buffered input is not upstream idleness.
			if len(st.pump.events) > 0 {
				st.pump.readLine()
				continue
			}
			return "", errGatewayAnthropicIdle
		case <-st.pump.deadlineC():
			return "", errGatewayAnthropicDrain
		}
	}
}

// read consumes complete SSE frames, including data-only frames. It never finalizes
// protocol output on EOF/read failure. An explicit terminal is the only success.
func (st *anthropicCompatBillingStream) read(c *gin.Context, usage *ClaudeUsage, process func(*apicompat.AnthropicStreamEvent) bool) error {
	eventName := ""
	var data []string
	dispatch := func() (bool, error) {
		payload := strings.Join(data, "\n")
		if eventName == "error" || gjson.Get(payload, "type").String() == "error" {
			return true, &sseStreamErrorEventError{RawData: payload}
		}
		if strings.TrimSpace(payload) == "[DONE]" {
			return true, nil
		}
		var event apicompat.AnthropicStreamEvent
		if json.Unmarshal([]byte(payload), &event) != nil {
			return false, nil
		}
		if event.Type == "" {
			event.Type = eventName
		}
		observer := upstreamResponseModelObserverFromContext(c)
		if observer == nil {
			observer = beginUpstreamResponseModelObservation(c)
		}
		observer.ObserveAnthropic([]byte(payload))
		if event.Type == "message_start" && event.Message != nil {
			mergeAnthropicUsage(usage, event.Message.Usage)
		}
		if event.Type == "message_delta" && event.Usage != nil {
			mergeAnthropicUsage(usage, *event.Usage)
		}
		mergeAnthropicCompatCacheDetails(usage, payload)
		st.checkClient()
		if process(&event) {
			st.disconnect()
		}
		return anthropicStreamEventIsTerminal(event.Type, payload) ||
			(event.Type == "message_delta" && event.Delta != nil && event.Delta.StopReason != ""), nil
	}
	for {
		line, err := st.next()
		if err != nil {
			// Scanner may deliver a final complete data line without an SSE blank
			// separator. Trust that tail only on EOF, never on a transport error.
			if errors.Is(err, io.EOF) && len(data) > 0 {
				payload := strings.TrimSpace(strings.Join(data, "\n"))
				if json.Valid([]byte(payload)) || payload == "[DONE]" {
					terminal, tailErr := dispatch()
					if terminal || tailErr != nil {
						return tailErr
					}
				}
			}
			return fmt.Errorf("stream usage incomplete: %w", err)
		}
		switch {
		case strings.TrimSpace(line) == "":
			terminal, err := dispatch()
			if terminal || err != nil {
				return err
			}
			eventName, data = "", nil
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
}

func mergeAnthropicCompatCacheDetails(usage *ClaudeUsage, payload string) {
	root := gjson.Parse(payload)
	node := root.Get("usage")
	if root.Get("type").String() == "message_start" {
		node = root.Get("message.usage")
	}
	if n := node.Get("cache_creation.ephemeral_5m_input_tokens").Int(); n > 0 {
		usage.CacheCreation5mTokens = int(n)
	}
	if n := node.Get("cache_creation.ephemeral_1h_input_tokens").Int(); n > 0 {
		usage.CacheCreation1hTokens = int(n)
	}
	if usage.CacheCreationInputTokens == 0 {
		usage.CacheCreationInputTokens = usage.CacheCreation5mTokens + usage.CacheCreation1hTokens
	}
}

func anthropicCompatBillingFailure(c *gin.Context, result *ForwardResult, err error, write func() bool) (*ForwardResult, error) {
	var sseErr *sseStreamErrorEventError
	if errors.As(err, &sseErr) {
		if result.ClientDisconnect {
			// There is no client left to retry. Keep already observed billing
			// usage and the upstream error, without writing another client frame.
			if !result.Usage.hasObservedTokens() {
				return nil, err
			}
			return result, err
		}
		// The handler's effective-output barrier decides whether this attempt can retry.
		if c.Writer.Size() > 0 {
			write()
		}
		return nil, &UpstreamFailoverError{StatusCode: http.StatusForbidden, ResponseBody: []byte(sseErr.RawData)}
	}
	if !result.ClientDisconnect {
		result.ClientDisconnect = write()
	}
	if !result.Usage.hasObservedTokens() {
		return nil, err
	}
	return result, err
}

func writeAnthropicCompatBillingFailure(c *gin.Context, state *apicompat.AnthropicEventToResponsesState, err error) bool {
	MarkResponseCommitted(c)
	message := "upstream stream ended without a terminal event"
	if !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
		message = sanitizeStreamError(err)
	}
	failure := map[string]string{"type": "stream_incomplete", "message": message}
	MarkOpsStreamError(c, failure["type"], message, http.StatusBadGateway)
	var frame string
	if state == nil {
		payload, _ := json.Marshal(map[string]any{"error": failure})
		frame = "data: " + string(payload) + "\n\n"
	} else {
		responseID := state.ResponseID
		if responseID == "" {
			responseID = "resp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		}
		payload, _ := json.Marshal(apicompat.ResponsesStreamEvent{
			Type: "response.failed", SequenceNumber: state.SequenceNumber,
			Response: &apicompat.ResponsesResponse{
				ID: responseID, Object: "response", Model: state.Model, CreatedAt: state.Created,
				Status: "failed", Output: []apicompat.ResponsesOutput{},
				Error: &apicompat.ResponsesError{Code: failure["type"], Message: message},
			},
		})
		state.SequenceNumber++
		frame = "event: response.failed\ndata: " + string(payload) + "\n\n"
	}
	if _, err := io.WriteString(c.Writer, frame); err != nil {
		return true
	}
	c.Writer.Flush()
	return false
}
