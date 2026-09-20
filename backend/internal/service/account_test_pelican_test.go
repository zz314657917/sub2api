package service

import (
	"errors"
	"testing"
)

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
