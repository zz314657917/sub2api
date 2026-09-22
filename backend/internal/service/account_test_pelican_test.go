package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type pelicanExtractSaveRepo struct {
	PelicanTestRepository
	result *PelicanResult
}

func (r *pelicanExtractSaveRepo) SaveResult(_ context.Context, result *PelicanResult, _ int, _ int64) error {
	r.result = result
	return nil
}
func TestPelicanMinimumLengthUsesExtractedDocument(t *testing.T) {
	content, _ := json.Marshal(TestEvent{Type: "content", Text: strings.Repeat("说明", 300) + "<html>短</html>"})
	html, err := parsePelicanEvents("data: " + string(content) + "\ndata: {\"type\":\"test_complete\",\"success\":true}\n")
	if err != nil {
		t.Fatal(err)
	}
	repo := &pelicanExtractSaveRepo{}
	svc := &PelicanTestService{repo: repo}
	svc.save(context.Background(), &PelicanPlan{MinChars: 100}, 1, &PelicanResult{HTML: html}, "success", "", time.Now(), pelicanPromptSnapshot{})
	if repo.result.Status != "failed" || repo.result.HTML != "" || repo.result.CharCount >= 100 {
		t.Fatalf("unexpected result: %+v", repo.result)
	}
}

type pelicanBrokenReader struct{ err error }

func (r pelicanBrokenReader) Read([]byte) (int, error) { return 0, r.err }
func (r pelicanBrokenReader) Close() error             { return nil }
func TestPelicanBodyFailureTypes(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code string
	}{
		{io.ErrUnexpectedEOF, "upstream_stream_closed"},
		{context.DeadlineExceeded, "upstream_timeout"},
	} {
		var failure error
		body := &pelicanObservedBody{ReadCloser: pelicanBrokenReader{tc.err}, failure: &failure}
		_, _ = body.Read(make([]byte, 1))
		code, _ := classifyPelicanFailure(failure, nil)
		if code != tc.code {
			t.Fatalf("got %s want %s", code, tc.code)
		}
	}
}

func TestPelicanFencedDocumentExtraction(t *testing.T) {
	for _, tc := range []struct {
		text string
		ok   bool
	}{
		{"<!doctype html><html><body>作品</body></html>", true},
		{"说明文字\n<!DOCTYPE html><html><body>作品</body></html>\n尾注", true},
		{"说明文字<html>作品</html>尾注", true},
		{"只有解释文字", false},
		{"说明 <div>片段</div>", false},
		{"```html\n<html>截断\n```\n```html\n<html>完整</html>\n```", false},
		{"说明\n```html\n<html><body>作品</body></html>\n```\n完成", true},
		{"```HTML\r\n<!doctype html><html>作品</html>\r\n```", true},
		{"说明\n```html\n<html><body>截断\n```", false},
		{"说明\n```html\n<div>片段</div>\n```", false},
	} {
		content, _ := json.Marshal(TestEvent{Type: "content", Text: tc.text})
		body := "data: " + string(content) + "\ndata: {\"type\":\"test_complete\",\"success\":true}\n"
		html, err := parsePelicanEvents(body)
		if (err == nil) != tc.ok {
			t.Fatalf("text=%q err=%v", tc.text, err)
		}
		if tc.ok && (strings.Contains(html, "```") || strings.Contains(html, "说明")) {
			t.Fatalf("unexpected wrapper: %q", html)
		}
	}
}

func TestPelicanRequestFailureIsNotStreamFailure(t *testing.T) {
	err := newPelicanRequestFailure("upstream rejected request HTTP 503 https://private.test/?token=secret Bearer secret")
	code, safe := classifyPelicanFailure(err, nil)
	if code != "upstream_request_failed" || !strings.Contains(safe, "503") || strings.Contains(safe, "secret") || strings.Contains(safe, "private.test") {
		t.Fatalf("%s %s", code, safe)
	}
}

func TestPelicanResponsesPayloadPropagatesNonemptyPrompt(t *testing.T) {
	p := createOpenAITestPayloadWithExactPrompt("gpt-test", false, "  独立鹈鹕作品\n")
	input := p["input"].([]map[string]any)[0]
	content := input["content"].([]map[string]any)[0]
	if content["text"] != "  独立鹈鹕作品\n" {
		t.Fatalf("prompt was not propagated: %#v", content)
	}
}

func TestPelicanChatPayloadPreservesWhitespace(t *testing.T) {
	p := createOpenAIChatCompletionsTestPayloadWithExactPrompt("gpt-test", "  第一行\n第二行  ")
	messages := p["messages"].([]map[string]any)
	if messages[0]["content"] != "  第一行\n第二行  " {
		t.Fatalf("prompt was trimmed: %#v", messages[0])
	}
}

func TestPelicanResponsesPayloadPreservesEmptyProbe(t *testing.T) {
	p := createOpenAITestPayloadWithPrompt("gpt-test", false, "")
	input := p["input"].([]map[string]any)[0]
	content := input["content"].([]map[string]any)[0]
	if content["text"] != "hi" {
		t.Fatalf("empty prompt changed historic probe: %#v", content)
	}
}

func TestParsePelicanEventsReportsSpecificIncompleteHTMLReason(t *testing.T) {
	body := `data: {"type":"content","text":"说明文字"}` + "\n"
	_, err := parsePelicanEvents(body)
	if !errors.Is(err, ErrPelicanIncompleteHTML) {
		t.Fatalf("expected incomplete HTML error, got %v", err)
	}
	code, safe := classifyPelicanFailure(err, nil)
	if code != "incomplete_html" {
		t.Fatalf("unexpected code: %s", code)
	}
	if safe != "返回的 HTML 内容不完整：未收到完成标记；缺少 HTML 起始标签；缺少 HTML 结束标签" {
		t.Fatalf("unexpected safe error: %s", safe)
	}
}
