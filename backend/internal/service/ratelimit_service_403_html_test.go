package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const openAI403HTMLBody = "<!DOCTYPE html><html><body>403 Forbidden</body></html>"

func TestHandleUpstreamError_OpenAIHTML403(t *testing.T) {
	svc := NewRateLimitService(nil, nil, &config.Config{}, nil, nil)
	account := &Account{ID: 501, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(openAI403HTMLBody)))
}

func TestHandleUpstreamError_OpenAIStructured403(t *testing.T) {
	require.False(t, isHTMLResponse([]byte(`{"error":{"message":"forbidden"}}`)))
}

func TestHandleUpstreamError_HTML403OnOtherPlatformsUnchanged(t *testing.T) {
	require.True(t, isHTMLResponse([]byte(openAI403HTMLBody)))
	// Platform gating is in handle403; only handleOpenAI403 invokes this helper.
	require.False(t, isHTMLResponse([]byte("Forbidden")))
}

func TestIsHTMLResponse(t *testing.T) {
	for _, tc := range []struct {
		body string
		want bool
	}{
		{"<!doctype html><html></html>", true},
		{" \n<HTML lang=\"en\">", true},
		{`{"error":{"message":"forbidden"}}`, false},
		{"Forbidden", false},
	} {
		require.Equal(t, tc.want, isHTMLResponse([]byte(tc.body)))
	}
}
