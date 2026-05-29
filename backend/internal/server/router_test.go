package server

import (
	"reflect"
	"testing"
)

func TestAppendFrameOriginAddsOpenWebUIChatOrigin(t *testing.T) {
	origins := appendFrameOrigin([]string{"https://checkout.airwallex.com"}, "https://image.3zapi.top/image")
	want := []string{"https://checkout.airwallex.com", "https://image.3zapi.top"}
	if !reflect.DeepEqual(origins, want) {
		t.Fatalf("origins = %#v, want %#v", origins, want)
	}
}

func TestAppendFrameOriginDeduplicatesOrigin(t *testing.T) {
	origins := appendFrameOrigin([]string{"https://image.3zapi.top"}, "https://image.3zapi.top/image")
	want := []string{"https://image.3zapi.top"}
	if !reflect.DeepEqual(origins, want) {
		t.Fatalf("origins = %#v, want %#v", origins, want)
	}
}

func TestFrameOriginFromURLRejectsNonHTTPURLs(t *testing.T) {
	if origin := frameOriginFromURL("javascript:alert(1)"); origin != "" {
		t.Fatalf("origin = %q, want empty", origin)
	}
}
