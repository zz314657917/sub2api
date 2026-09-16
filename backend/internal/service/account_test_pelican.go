package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
)

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
type pelicanUpstream struct{ HTTPUpstream }

func (u pelicanUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, profile *tlsfingerprint.Profile) (*http.Response, error) {
	resp, err := u.HTTPUpstream.DoWithTLS(req, proxy, id, concurrency, profile)
	if err != nil {
		return nil, errors.New("pelican upstream request failed")
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("pelican upstream response missing")
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(strings.NewReader("provider request failed"))
	} else {
		// Bound raw SSE overhead before the shared line parser buffers it.
		resp.Body = &pelicanLimitedBody{Reader: io.LimitReader(resp.Body, 8<<20), Closer: resp.Body}
	}
	return resp, nil
}

var pelicanHTMLStart = regexp.MustCompile(`(?is)^\s*(?:<!doctype\s+html[^>]*>\s*)?<html(?:\s[^>]*)?>`)
var pelicanHTMLEnd = regexp.MustCompile(`(?is)</html>\s*$`)

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
			return "", errors.New("pelican provider request failed")
		case "test_complete":
			if !event.Success {
				return "", errors.New("pelican provider request failed")
			}
			completed = true
		case "content":
			if text.Len()+len(event.Text) > 256*1024 {
				return "", errors.New("pelican HTML exceeds limit")
			}
			text.WriteString(event.Text)
		}
	}
	html := strings.TrimSpace(text.String())
	if strings.HasPrefix(html, "```html") {
		html = strings.TrimSpace(strings.TrimPrefix(html, "```html"))
		html = strings.TrimSpace(strings.TrimSuffix(html, "```"))
	}
	if !completed || !utf8.ValidString(html) || !pelicanHTMLStart.MatchString(html) || !pelicanHTMLEnd.MatchString(html) {
		return "", errors.New("pelican provider returned incomplete HTML")
	}
	return html, nil
}

// RunPelicanTest is the narrow provider adapter used by the scheduled runner.
// It deliberately reuses the account test path so TLS profile, proxy and OAuth
// authentication stay identical to the validated administrative probe path.
func (s *AccountTestService) RunPelicanTest(ctx context.Context, accountID int64, model string) (*PelicanResult, error) {
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
	c.Writer = &cappedPelicanWriter{ResponseWriter: c.Writer, remaining: 4 << 20, cancel: cancel}
	started := time.Now()
	// Use the established proxy/TLS/auth path without copying its mutex fields.
	probe := &AccountTestService{accountRepo: s.accountRepo, httpUpstream: pelicanUpstream{s.httpUpstream}, cfg: s.cfg, tlsFPProfileService: s.tlsFPProfileService}
	err = probe.testOpenAIAccountConnection(c, account, model, PelicanPrompt, AccountTestModeDefault)
	if c.Writer.(*cappedPelicanWriter).overflow {
		return nil, errors.New("pelican response exceeds limit")
	}
	if err != nil || runCtx.Err() != nil {
		return nil, errors.New("pelican provider request failed")
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
