package service

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserErrorCategoryAndFilter(t *testing.T) {
	for _, tc := range []struct {
		phase, typ, category string
	}{
		{"auth", "authentication_error", "auth"},
		{"request", "rate_limit_error", "rate_limit"},
		{"request", "billing_error", "quota"},
		{"upstream", "api_error", "upstream"},
		{"internal", "api_error", "internal"},
	} {
		if got := MapUserErrorCategory(tc.phase, tc.typ); got != tc.category {
			t.Fatalf("category(%q,%q)=%q, want %q", tc.phase, tc.typ, got, tc.category)
		}
	}
	phases, types := CategoryToFilter("quota")
	if len(phases) != 0 || len(types) != 2 {
		t.Fatalf("quota filter = phases=%v types=%v", phases, types)
	}
}

func TestToUserErrorRequestDetailWhitelist(t *testing.T) {
	clientIP := "192.0.2.1"
	upstreamStatus := 503
	out := ToUserErrorRequestDetail(&OpsErrorLogDetail{
		OpsErrorLog: OpsErrorLog{
			ID: 9, CreatedAt: time.Unix(10, 0).UTC(), RequestedModel: "requested-model",
			InboundEndpoint: "/v1/responses", StatusCode: 502, Phase: "upstream", Type: "api_error",
			Platform: "openai", Message: "upstream failed", ClientIP: &clientIP, UserEmail: "private@example.com",
			UpstreamEndpoint: "https://provider.invalid", GroupName: "my-group", UserAgent: "client/1.0", Stream: true,
		},
		ErrorBody: `{"error":"failed"}`, UpstreamStatusCode: &upstreamStatus,
	})
	if out == nil || out.Model != "requested-model" {
		t.Fatalf("unexpected user detail: %+v", out)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"user_email", "upstream_endpoint", "account_id", "api_key_id", "client_ip", "user_agent", "group_name", "request_type", "stream"} {
		if _, present := payload[forbidden]; present {
			t.Fatalf("sensitive field %q leaked: %s", forbidden, raw)
		}
	}
}
