package service

import "testing"

func TestPelicanResponsesPayloadPropagatesNonemptyPrompt(t *testing.T) {
	p := createOpenAITestPayloadWithPrompt("gpt-test", false, "独立鹈鹕作品")
	input := p["input"].([]map[string]any)[0]
	content := input["content"].([]map[string]any)[0]
	if content["text"] != "独立鹈鹕作品" {
		t.Fatalf("prompt was not propagated: %#v", content)
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
