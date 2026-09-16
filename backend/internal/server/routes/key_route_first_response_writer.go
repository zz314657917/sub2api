package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const keyRouteFirstResponsePreludeLimit = 64 << 10

// keyRouteFirstResponseWriter stages a protocol prelude privately. Its Size
// and Written state intentionally describe staged output, because existing
// handlers use those signals for same-account retry guards. committed is a
// separate downstream-visible boundary.
type keyRouteFirstResponseWriter struct {
	real       gin.ResponseWriter
	header     http.Header
	status     int
	size       int
	buf        bytes.Buffer
	stream     bool
	committed  bool
	onBusiness func()
	mu         sync.Mutex
	expired    bool
	gate       *keyRouteFirstResponseGate
}

type keyRouteFirstResponseGate struct {
	mu                 sync.Mutex
	expired, committed bool
}

func (g *keyRouteFirstResponseGate) expire() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.committed {
		return false
	}
	g.expired = true
	return true
}
func (g *keyRouteFirstResponseGate) claim() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.expired || g.committed {
		return false
	}
	g.committed = true
	return true
}

func (g *keyRouteFirstResponseGate) isExpired() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.expired
}

func newKeyRouteFirstResponseWriter(real gin.ResponseWriter, stream bool, gate *keyRouteFirstResponseGate) *keyRouteFirstResponseWriter {
	header := make(http.Header, len(real.Header()))
	for key, values := range real.Header() {
		header[key] = append([]string(nil), values...)
	}
	return &keyRouteFirstResponseWriter{real: real, header: header, status: http.StatusOK, size: -1, stream: stream, gate: gate}
}

func (w *keyRouteFirstResponseWriter) Header() http.Header { return w.header }
func (w *keyRouteFirstResponseWriter) Status() int         { return w.status }
func (w *keyRouteFirstResponseWriter) Size() int           { return w.size }
func (w *keyRouteFirstResponseWriter) Written() bool       { return w.size >= 0 }
func (w *keyRouteFirstResponseWriter) WriteHeader(code int) {
	if code > 0 && !w.Written() {
		w.status = code
	}
}
func (w *keyRouteFirstResponseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
	}
}
func (w *keyRouteFirstResponseWriter) WriteString(value string) (int, error) {
	return w.Write([]byte(value))
}
func (w *keyRouteFirstResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	if w.committed {
		n, err := w.real.Write(data)
		w.size += n
		return n, err
	}
	if w.buf.Len()+len(data) > keyRouteFirstResponsePreludeLimit {
		if err := w.Commit(); err != nil {
			return 0, err
		}
		n, err := w.real.Write(data)
		w.size += n
		return n, err
	}
	n, _ := w.buf.Write(data)
	w.size += n
	if w.stream && keyRouteFirstResponseMeaningful(w.buf.Bytes()) {
		if err := w.Commit(); err != nil {
			return n, err
		}
	}
	return n, nil
}
func (w *keyRouteFirstResponseWriter) Flush() {
	w.WriteHeaderNow()
	if w.committed {
		w.real.Flush()
	}
}
func (w *keyRouteFirstResponseWriter) CloseNotify() <-chan bool { return w.real.CloseNotify() }
func (w *keyRouteFirstResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("first-response dispatcher does not support hijack")
}
func (w *keyRouteFirstResponseWriter) Pusher() http.Pusher { return nil }
func (w *keyRouteFirstResponseWriter) Commit() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.committed {
		return nil
	}
	if w.expired || (w.gate != nil && !w.gate.claim()) {
		return errors.New("first response timeout won")
	}
	w.committed = true
	if w.onBusiness != nil {
		w.onBusiness()
	}
	for key := range w.real.Header() {
		w.real.Header().Del(key)
	}
	for key, values := range w.header {
		w.real.Header()[key] = append([]string(nil), values...)
	}
	w.real.WriteHeader(w.status)
	if w.buf.Len() > 0 {
		if _, err := w.real.Write(w.buf.Bytes()); err != nil {
			return err
		}
	}
	w.buf.Reset()
	return nil
}
func (w *keyRouteFirstResponseWriter) Committed() bool                  { return w.committed }
func (w *keyRouteFirstResponseWriter) SetFirstBusiness(callback func()) { w.onBusiness = callback }
func (w *keyRouteFirstResponseWriter) Expire() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.committed {
		return false
	}
	w.expired = true
	return true
}

func (w *keyRouteFirstResponseWriter) isExpired() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.expired
}

func keyRouteFirstResponseMeaningful(chunk []byte) bool {
	for _, line := range bytes.Split(chunk, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte(":")) {
			continue
		}
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		if bytes.Equal(payload, []byte("[DONE]")) {
			return true
		}
		var value any
		if json.Unmarshal(payload, &value) != nil {
			continue
		}
		if keyRouteFirstResponseJSONMeaningful(value) {
			return true
		}
	}
	return false
}

func keyRouteFirstResponseJSONMeaningful(value any) bool {
	object, ok := value.(map[string]any)
	if !ok {
		return false
	}
	if keyRouteFirstResponseHasUsage(value) {
		return true
	}
	if event, _ := object["type"].(string); event != "" {
		switch event {
		case "response.created", "response.in_progress", "message_start", "message_delta", "ping":
			return false
		case "response.completed", "response.done":
			return true
		}
		if strings.HasSuffix(event, ".delta") && (keyRouteFirstResponseNonEmpty(object["delta"]) || keyRouteFirstResponseNonEmpty(object["arguments"])) {
			return true
		}
	}
	return keyRouteFirstResponseHasContent(object)
}

func keyRouteFirstResponseHasUsage(value any) bool {
	switch current := value.(type) {
	case map[string]any:
		if usage, ok := current["usage"].(map[string]any); ok && keyRouteFirstResponseUsageMeaningful(usage) {
			return true
		}
		for _, child := range current {
			if keyRouteFirstResponseHasUsage(child) {
				return true
			}
		}
	case []any:
		for _, child := range current {
			if keyRouteFirstResponseHasUsage(child) {
				return true
			}
		}
	}
	return false
}

func keyRouteFirstResponseUsageMeaningful(usage map[string]any) bool {
	for key, value := range usage {
		normalized := strings.ToLower(key)
		if !(strings.Contains(normalized, "input") || strings.Contains(normalized, "output") || strings.Contains(normalized, "cache") || strings.Contains(normalized, "token")) {
			continue
		}
		switch number := value.(type) {
		case float64:
			if number > 0 {
				return true
			}
		case json.Number:
			if parsed, err := number.Float64(); err == nil && parsed > 0 {
				return true
			}
		}
	}
	return false
}

func keyRouteFirstResponseHasContent(value any) bool {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			normalized := strings.ToLower(key)
			switch normalized {
			case "content", "text", "reasoning", "thinking", "tool_calls", "tool_use", "partial_json", "arguments", "output_text":
				if keyRouteFirstResponseNonEmpty(child) {
					return true
				}
			}
			if keyRouteFirstResponseHasContent(child) {
				return true
			}
		}
	case []any:
		for _, child := range current {
			if keyRouteFirstResponseHasContent(child) {
				return true
			}
		}
	}
	return false
}
func keyRouteFirstResponseNonEmpty(value any) bool {
	switch current := value.(type) {
	case string:
		return strings.TrimSpace(current) != ""
	case []any:
		return len(current) > 0
	case map[string]any:
		return len(current) > 0
	default:
		return value != nil
	}
}

var _ gin.ResponseWriter = (*keyRouteFirstResponseWriter)(nil)
var _ io.StringWriter = (*keyRouteFirstResponseWriter)(nil)
