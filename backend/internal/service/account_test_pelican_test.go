package service

import "testing"

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
