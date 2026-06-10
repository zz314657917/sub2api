package server

import (
	"reflect"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestAppendFrameOriginAddsHTTPOrigin(t *testing.T) {
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

func TestStudioBridgeFrameAncestorAllowlistFromSettings(t *testing.T) {
	allowlist := studioBridgeFrameAncestorAllowlistFromSettings(&service.StudioBridgeAppSettings{
		Enabled:              true,
		LaunchReturnURL:      "http://127.0.0.1:8081/auth/sub2api/launch",
		AllowedReturnDomains: []string{"luoye.example.com", "example.net"},
	})

	if !studioBridgeFrameAncestorAllowed("http://127.0.0.1:8081", allowlist) {
		t.Fatalf("expected local launch origin to be allowed")
	}
	if !studioBridgeFrameAncestorAllowed("https://luoye.example.com", allowlist) {
		t.Fatalf("expected configured return domain to be allowed")
	}
	if !studioBridgeFrameAncestorAllowed("https://studio.example.net", allowlist) {
		t.Fatalf("expected configured return subdomain to be allowed")
	}
	if studioBridgeFrameAncestorAllowed("http://luoye.example.com", allowlist) {
		t.Fatalf("expected non-HTTPS production domain to be rejected")
	}
	if studioBridgeFrameAncestorAllowed("https://evil.example.org", allowlist) {
		t.Fatalf("expected unrelated origin to be rejected")
	}
}

func TestStudioBridgeFrameAncestorAllowlistDisabled(t *testing.T) {
	allowlist := studioBridgeFrameAncestorAllowlistFromSettings(&service.StudioBridgeAppSettings{
		Enabled:         false,
		LaunchReturnURL: "http://127.0.0.1:8081/auth/sub2api/launch",
	})

	if studioBridgeFrameAncestorAllowed("http://127.0.0.1:8081", allowlist) {
		t.Fatalf("expected disabled bridge to reject frame ancestor")
	}
}
