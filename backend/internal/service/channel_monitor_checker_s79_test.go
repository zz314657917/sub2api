package service

import "testing"

func TestS79ExtractAnthropicMonitorText(t *testing.T) {
	extractAnthropic := func(body []byte) string {
		return extractMonitorResponseText(providerAdapters[MonitorProviderAnthropic], body)
	}

	t.Run("thinking before text", func(t *testing.T) {
		body := []byte(`{"content":[{"type":"thinking","thinking":"work"},{"type":"text","text":"  answer is 2  "}]}`)
		if got := extractAnthropic(body); got != "answer is 2" {
			t.Fatalf("Anthropic monitor extraction = %q, want %q", got, "answer is 2")
		}
	})

	t.Run("multiple text and tool blocks", func(t *testing.T) {
		body := []byte(`{"content":[{"type":"text","text":" first "},{"type":"tool_use","name":"lookup"},{"type":"text","text":"second"},{"type":"text","text":"  "}]}`)
		if got := extractAnthropic(body); got != "first\nsecond" {
			t.Fatalf("Anthropic monitor extraction = %q, want %q", got, "first\\nsecond")
		}
	})

	t.Run("empty and malformed", func(t *testing.T) {
		for _, body := range [][]byte{
			[]byte(`{"content":[]}`),
			[]byte(`{"content":[{"type":"thinking","thinking":"only"},{"type":"text","text":"  "}]}`),
			[]byte(`{"content":"not-an-array"}`),
			[]byte(`{"content":[{"type":"text"},null,"broken"]}`),
			[]byte(`{"content":`),
		} {
			if got := extractAnthropic(body); got != "" {
				t.Errorf("Anthropic monitor extraction for %q = %q, want empty", body, got)
			}
		}
	})

	t.Run("challenge matches aggregated text", func(t *testing.T) {
		body := []byte(`{"content":[{"type":"thinking","thinking":"work"},{"type":"text","text":"答案是"},{"type":"tool_use","name":"ignored"},{"type":"text","text":" 2 "}]}`)
		respText := extractAnthropic(body)
		if !validateChallenge(respText, "2") {
			t.Fatalf("validateChallenge(%q, %q) = false, want true", respText, "2")
		}
	})

	t.Run("other providers keep text path extraction", func(t *testing.T) {
		body := []byte(`{"choices":[{"message":{"content":"unchanged"}}]}`)
		if got := extractMonitorResponseText(providerAdapters[MonitorProviderOpenAI], body); got != "unchanged" {
			t.Fatalf("extractMonitorResponseText() = %q, want %q", got, "unchanged")
		}
	})
}
