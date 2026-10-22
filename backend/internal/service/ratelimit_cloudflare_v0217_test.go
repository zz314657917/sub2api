package service

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudflareEdge403NoAccountPenaltyV0217(t *testing.T) {
	bodies := []string{
		"<html><body>Cloudflare Error code: 1010</body></html>",
		"error code: 1010",
		`{"cloudflare_error":true,"code":1010}`,
		`{"errors":[{"nested":{"cloudflare_error":true,"error_code":" 1010 "}}]}`,
		`{"CLOUDFLARE_ERROR":true,"CODE":"1010"}`,
		`{"error":{"error_name":" BROWSER_SIGNATURE_BANNED "}}`,
	}
	for _, platform := range []string{PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformKimi, PlatformZhipu, PlatformDeepseek, PlatformAnthropic} {
		for i, body := range bodies {
			t.Run(fmt.Sprintf("%s/body%d", platform, i), func(t *testing.T) {
				svc, repo, counter := newHTML403Service(openAI403DisableThreshold)
				account := &Account{ID: 701, Platform: platform, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
				for attempt := 0; attempt < 4; attempt++ {
					require.False(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(body)))
				}
				require.Zero(t, counter.increments)
				require.Zero(t, repo.setErrorCalls)
				require.Zero(t, repo.tempCalls)
				require.Equal(t, StatusActive, account.Status)
				require.True(t, account.Schedulable)
			})
		}
	}
}

func TestCloudflareUnrelated403RetainsAccountPolicyV0217(t *testing.T) {
	for _, body := range []string{
		`{"error":{"code":1010,"message":"account banned"}}`,
		`{"cloudflare_error":false,"error_code":"1010"}`,
		`{"cloudflare_error":true,"code":1011}`,
		`{"cloudflare_error":true,"code":1010.5}`,
		`{"message":"error code: 1010"}`,
		`{"cloudflare_error":true,"code":1010`,
		`{"error_name":"browser_signature_banned_other"}`,
		`{"error":{"type":"permission_error","message":"invalid API key"}}`,
	} {
		for _, platform := range []string{PlatformOpenAI, PlatformKimi, PlatformGemini, PlatformAntigravity, PlatformAnthropic} {
			for _, count := range []int64{1, openAI403DisableThreshold} {
				t.Run(fmt.Sprintf("%s/count%d/%s", platform, count, body), func(t *testing.T) {
					svc, repo, counter := newHTML403Service(count)
					account := &Account{ID: 702, Platform: platform, Type: AccountTypeAPIKey}
					require.True(t, svc.HandleUpstreamError(context.Background(), account, http.StatusForbidden, http.Header{}, []byte(body)))
					if platform == PlatformOpenAI || IsCNProvider(platform) {
						require.Equal(t, 1, counter.increments)
						if count == 1 {
							require.Equal(t, 1, repo.tempCalls)
							require.Zero(t, repo.setErrorCalls)
						} else {
							require.Zero(t, repo.tempCalls)
							require.Equal(t, 1, repo.setErrorCalls)
						}
					} else {
						require.Zero(t, counter.increments)
						require.Zero(t, repo.tempCalls)
						require.Equal(t, 1, repo.setErrorCalls)
					}
				})
			}
		}
	}
}

func TestCloudflareSignalDoesNotBypass401V0217(t *testing.T) {
	svc, repo, counter := newHTML403Service(1)
	require.True(t, svc.HandleUpstreamError(context.Background(),
		&Account{ID: 703, Platform: PlatformGemini, Type: AccountTypeAPIKey}, http.StatusUnauthorized,
		http.Header{}, []byte(`{"cloudflare_error":true,"code":1010}`)))
	require.Equal(t, 1, repo.setErrorCalls)
	require.Zero(t, repo.tempCalls)
	require.Zero(t, counter.increments)
}

func TestCloudflareAntigravityValidationViolationStillDisableV0217(t *testing.T) {
	for _, body := range []string{
		`{"error":{"message":"VALIDATION_REQUIRED"}}`,
		`{"error":{"message":"violation of terms of service"}}`,
	} {
		t.Run(body, func(t *testing.T) {
			svc, repo, counter := newHTML403Service(1)
			require.True(t, svc.HandleUpstreamError(context.Background(),
				&Account{ID: 704, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
				http.StatusForbidden, http.Header{}, []byte(body)))
			require.Equal(t, 1, repo.setErrorCalls)
			require.Zero(t, repo.tempCalls)
			require.Zero(t, counter.increments)
		})
	}
}
