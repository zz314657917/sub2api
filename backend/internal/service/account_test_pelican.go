package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/util/logredact"
	"github.com/gin-gonic/gin"
)

const pelicanTestInstructions = "Generate only one complete HTML document containing a single 16:9 inline SVG artwork with viewBox=\"0 0 960 540\". The artwork must show a complete pelican riding a complete bicycle with two full wheels, hubs, spokes, and tires. Use only inline CSS @keyframes or declarative SVG animation, and make it autoplay. Do not output a title, explanation, card, navigation, controls, play/pause/replay buttons, Markdown fences, JavaScript, Canvas, iframe, audio/video, links, remote images, remote fonts, or any external resource. Return only the HTML source, beginning with <!DOCTYPE html> or <html> and ending with </html>. Do not inspect directories, create or edit files, invoke tools, or describe file operations."

type pelicanRequestFailure struct{ safe string }

func (e *pelicanRequestFailure) Error() string { return e.safe }
func (e *pelicanRequestFailure) Unwrap() error { return ErrPelicanUpstreamRequest }

var pelicanPrivateValue = regexp.MustCompile(`(?i)(?:https?://[^\s<>"']+|bearer\s+[^\s<>"']+|sk-[a-z0-9_-]+)`)

func newPelicanRequestFailure(message string) *pelicanRequestFailure {
	message = pelicanPrivateValue.ReplaceAllString(message, "[redacted]")
	return &pelicanRequestFailure{safe: pelicanLogPreview(message, 320)}
}

type pelicanObservedBody struct {
	io.ReadCloser
	failure *error
}

func (b *pelicanObservedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && err != io.EOF {
		var timeout net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
			*b.failure = ErrPelicanUpstreamTimeout
		} else {
			*b.failure = fmt.Errorf("%w: response body read failed", ErrPelicanUpstreamClosed)
		}
	}
	return n, err
}

type cappedPelicanWriter struct {
	gin.ResponseWriter
	remaining int
	overflow  bool
	cancel    context.CancelFunc
}

func (w *cappedPelicanWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		w.overflow = true
		w.cancel()
		limit := w.remaining
		if limit < 0 {
			limit = 0
		}
		p = p[:limit]
	}
	n, e := w.ResponseWriter.Write(p)
	w.remaining -= n
	if w.overflow {
		return n, errors.New("pelican response exceeds limit")
	}
	return n, e
}
func (w *cappedPelicanWriter) WriteString(v string) (int, error) { return w.Write([]byte(v)) }

type pelicanLimitedBody struct {
	io.Reader
	io.Closer
}
type pelicanUpstream struct {
	HTTPUpstream
	failure *error
}

func (u pelicanUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, profile *tlsfingerprint.Profile) (*http.Response, error) {
	resp, err := u.HTTPUpstream.DoWithTLS(req, proxy, id, concurrency, profile)
	if err != nil {
		var timeout net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
			*u.failure = ErrPelicanUpstreamTimeout
		} else {
			*u.failure = newPelicanRequestFailure(err.Error())
		}
		return nil, *u.failure
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("pelican upstream response missing")
	}
	if resp.StatusCode != http.StatusOK {
		*u.failure = newPelicanRequestFailure(fmt.Sprintf("HTTP %d", resp.StatusCode))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(strings.NewReader("provider request failed"))
	} else {
		// Bound raw SSE overhead before the shared line parser buffers it.
		observed := &pelicanObservedBody{ReadCloser: resp.Body, failure: u.failure}
		resp.Body = &pelicanLimitedBody{Reader: io.LimitReader(observed, 8<<20), Closer: observed}
	}
	return resp, nil
}

var pelicanHTMLStart = regexp.MustCompile(`(?is)^\s*(?:<!doctype\s+html[^>]*>\s*)?<html(?:\s[^>]*)?>`)
var pelicanHTMLEnd = regexp.MustCompile(`(?is)</html>\s*$`)
var pelicanDocumentStart = regexp.MustCompile(`(?i)<!doctype\s+html\b[^>]*>|<html(?:\s[^>]*)?>`)
var pelicanHTMLFence = regexp.MustCompile("(?im)^\\s*```html[ \\t]*\\r?\\n")

func extractPelicanDocument(output string) string {
	// The first HTML fence is authoritative, even if its document is incomplete.
	if fence := pelicanHTMLFence.FindStringIndex(output); fence != nil {
		output = output[fence[1]:]
		if end := strings.Index(output, "```"); end >= 0 {
			output = output[:end]
		}
	} else if strings.Contains(output, "```") {
		// Retain compatibility with an unlabelled code fence.
		parts := strings.Split(output, "```")
		if len(parts) > 1 {
			output = parts[1]
		}
	}
	if start := pelicanDocumentStart.FindStringIndex(output); start != nil {
		output = output[start[0]:]
	}
	if end := strings.LastIndex(strings.ToLower(output), "</html>"); end >= 0 {
		output = output[:end+len("</html>")]
	}
	return strings.TrimSpace(output)
}

type pelicanIncompleteHTMLError struct {
	reason string
}

func (e *pelicanIncompleteHTMLError) Error() string {
	return "pelican incomplete html: " + e.reason
}

func (e *pelicanIncompleteHTMLError) Unwrap() error {
	return ErrPelicanIncompleteHTML
}

func pelicanIncompleteHTMLReason(completed, validUTF8, hasStart, hasEnd bool) string {
	reasons := make([]string, 0, 4)
	if !completed {
		reasons = append(reasons, "未收到完成标记")
	}
	if !validUTF8 {
		reasons = append(reasons, "内容不是有效 UTF-8")
	}
	if !hasStart {
		reasons = append(reasons, "缺少 HTML 起始标签")
	}
	if !hasEnd {
		reasons = append(reasons, "缺少 HTML 结束标签")
	}
	return strings.Join(reasons, "；")
}

func pelicanLogPreview(value string, maxRunes int) string {
	value = logredact.RedactText(value, "api_key", "token", "authorization", "cookie")
	value = strings.NewReplacer("\r", "\\r", "\n", "\\n", "\t", "\\t").Replace(value)
	runes := []rune(value)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "..."
	}
	return value
}

func logPelicanIncompleteHTML(body, html string, completed, validUTF8, hasStart, hasEnd bool) {
	const previewRunes = 240
	head := html
	tail := html
	runes := []rune(html)
	if len(runes) > previewRunes {
		head = string(runes[:previewRunes])
		tail = string(runes[len(runes)-previewRunes:])
	}
	logger.LegacyPrintf("service.account_test_pelican",
		"incomplete_html completed=%t valid_utf8=%t has_html_start=%t has_html_end=%t body_bytes=%d html_bytes=%d html_runes=%d head=%q tail=%q",
		completed, validUTF8, hasStart, hasEnd, len(body), len(html), utf8.RuneCountInString(html),
		pelicanLogPreview(head, previewRunes), pelicanLogPreview(tail, previewRunes))
}

func parsePelicanEvents(body string) (string, error) {
	var text strings.Builder
	completed := false
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var event TestEvent
		if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &event) != nil {
			return "", errors.New("invalid pelican event")
		}
		switch event.Type {
		case "error":
			return "", ErrPelicanUpstreamRequest
		case "test_complete":
			if !event.Success {
				return "", ErrPelicanUpstreamRequest
			}
			completed = true
		case "content":
			if text.Len()+len(event.Text) > 4<<20 {
				return "", errors.New("pelican HTML exceeds limit")
			}
			text.WriteString(event.Text)
		}
	}
	html := extractPelicanDocument(text.String())
	if len(html) > 256*1024 {
		return "", &pelicanIncompleteHTMLError{reason: "HTML 超过大小限制"}
	}
	validUTF8 := utf8.ValidString(html)
	hasStart := pelicanHTMLStart.MatchString(html)
	hasEnd := pelicanHTMLEnd.MatchString(html)
	if !completed || !validUTF8 || !hasStart || !hasEnd {
		logPelicanIncompleteHTML(body, html, completed, validUTF8, hasStart, hasEnd)
		return "", &pelicanIncompleteHTMLError{reason: pelicanIncompleteHTMLReason(completed, validUTF8, hasStart, hasEnd)}
	}
	return html, nil
}

// RunPelicanTest is the narrow provider adapter used by the scheduled runner.
// It deliberately reuses the account test path so TLS profile, proxy and OAuth
// authentication stay identical to the validated administrative probe path.
func (s *AccountTestService) RunPelicanTest(ctx context.Context, accountID int64, model, prompt, reasoningEffort string) (*PelicanResult, error) {
	if s == nil || s.accountRepo == nil || s.httpUpstream == nil {
		return nil, errors.New("pelican provider unavailable")
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, errors.New("pelican account unavailable")
	}
	if !pelicanAccountAvailable(account) || strings.TrimSpace(model) == "" || isOpenAIImageModel(account.GetMappedModel(model)) {
		return nil, errors.New("pelican account or model unsupported")
	}
	rec := httptest.NewRecorder()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	req := httptest.NewRequest("POST", "/", nil).WithContext(runCtx)
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set("pelican_test", true)
	c.Set("pelican_reasoning_effort", reasoningEffort)
	c.Writer = &cappedPelicanWriter{ResponseWriter: c.Writer, remaining: 4 << 20, cancel: cancel}
	started := time.Now()
	// Use the established proxy/TLS/auth path without copying its mutex fields.
	var upstreamFailure error
	probe := &AccountTestService{accountRepo: s.accountRepo, httpUpstream: pelicanUpstream{HTTPUpstream: s.httpUpstream, failure: &upstreamFailure}, cfg: s.cfg, tlsFPProfileService: s.tlsFPProfileService}
	err = probe.testOpenAIAccountConnection(c, account, model, prompt, AccountTestModeDefault)
	if c.Writer.(*cappedPelicanWriter).overflow {
		return nil, errors.New("pelican response exceeds limit")
	}
	if err != nil || runCtx.Err() != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrPelicanUpstreamTimeout
		}
		if upstreamFailure != nil {
			return nil, upstreamFailure
		}
		return nil, err
	}
	html, err := parsePelicanEvents(rec.Body.String())
	if err != nil {
		return nil, err
	}
	for _, credential := range []string{account.GetOpenAIApiKey(), account.GetOpenAIAccessToken()} {
		if len(credential) >= 8 && strings.Contains(html, credential) {
			return nil, errors.New("pelican response contains private data")
		}
	}
	now := time.Now()
	return &PelicanResult{Status: "success", HTML: html, CharCount: utf8.RuneCountInString(html), LatencyMS: time.Since(started).Milliseconds(), StartedAt: started, FinishedAt: &now}, nil
}
