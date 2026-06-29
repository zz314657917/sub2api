package service

import "testing"

func TestOpenAIImageOutputCounterIgnoresTextOnlyDataArrays(t *testing.T) {
	counter := newOpenAIImageOutputCounter()
	counter.AddJSONResponse([]byte(`{
		"data":[
			{"type":"message","content":[{"type":"output_text","text":"hello"}]},
			{"text":"plain text only"}
		]
	}`))
	if got := counter.Count(); got != 0 {
		t.Fatalf("text-only data array should not count as image output, got %d", got)
	}
}

func TestOpenAIImageOutputCounterCountsOnlyImageDataItems(t *testing.T) {
	counter := newOpenAIImageOutputCounter()
	counter.AddJSONResponse([]byte(`{
		"data":[
			{"type":"message","text":"not an image"},
			{"b64_json":"image-a","size":"1024x1024"},
			{"url":"https://example.test/image.png","size":"2048x2048"}
		]
	}`))
	if got := counter.Count(); got != 2 {
		t.Fatalf("expected two image data items, got %d", got)
	}
	sizes := counter.Sizes()
	if len(sizes) != 2 || sizes[0] != "1024x1024" || sizes[1] != "2048x2048" {
		t.Fatalf("unexpected sizes: %#v", sizes)
	}
}

func TestOpenAIImageOutputCounterIgnoresEmptyImageGenerationCompleted(t *testing.T) {
	counter := newOpenAIImageOutputCounter()
	counter.AddSSEData([]byte(`{"type":"image_generation.completed","id":"ig_empty"}`))
	if got := counter.Count(); got != 0 {
		t.Fatalf("empty image_generation.completed should not count, got %d", got)
	}
}
