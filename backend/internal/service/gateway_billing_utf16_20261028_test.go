package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func upstreamAFingerprintBody(t *testing.T, text string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": text}}})
	require.NoError(t, err)
	return body
}

func upstreamAFingerprintForSample(sample, version string) string {
	sum := sha256.Sum256([]byte("59cf53e54c78" + sample + version))
	return hex.EncodeToString(sum[:])[:3]
}

func TestUpstreamAFixesClaudeFingerprintUTF16Samples(t *testing.T) {
	for _, tc := range []struct{ name, text, sample string }{
		{"ascii", "abcdefghijklmnopqrstuvwxyz", "ehu"},
		{"empty", "", "000"}, {"short_ascii", "hello", "o00"},
		{"cjk", strings.Repeat("中", 21), "中中中"},
		{"short_cjk", "你好世界甲", "甲00"},
		{"emoji_before_samples", "😀abcdefghijklmnopqrstuv", "cfs"},
		{"unpaired_high", "abcd😀abcdefghijklmnopq", "\ufffdbo"},
		{"unpaired_low", "abc😀abcdefghijklmnopqr", "\ufffdcp"},
		// Source emoji are separate. Selected high+low halves join into U+1F603.
		{"recombined_pair", "abcd😀😃abcdefghijklmnop", "😃m"},
		{"unmatched_halves", "abcd😀x😃abcdefghijklmnop", "\ufffd\ufffdl"},
		{"emoji_at_last_sample", "abcdefghijklmnopqrst😀", "eh\ufffd"},
		{"long_prefix_bounded", "abcdefghijklmnopqrstuvwxyz" + strings.Repeat("😀中", 100000), "ehu"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := upstreamAFingerprintBody(t, tc.text)
			require.Equal(t, upstreamAFingerprintForSample(tc.sample, "2.1.92"), computeClaudeCodeFingerprint(body, "2.1.92"))
			// The first user text block, after a non-user message, is the same input.
			blockBody, err := json.Marshal(map[string]any{"messages": []any{
				map[string]any{"role": "assistant", "content": "ignored"},
				map[string]any{"role": "user", "content": []any{
					map[string]any{"type": "image", "source": "ignored"},
					map[string]any{"type": "text", "text": tc.text},
					map[string]any{"type": "text", "text": "ignored"},
				}},
			}})
			require.NoError(t, err)
			require.Equal(t, upstreamAFingerprintForSample(tc.sample, "2.1.92"), computeClaudeCodeFingerprint(blockBody, "2.1.92"))
		})
	}
}

func TestUpstreamAFixesClaudeFingerprintWireConsistency(t *testing.T) {
	for _, endpoint := range []string{"messages", "count_tokens"} {
		for _, mimic := range []bool{false, true} {
			for _, text := range []string{strings.Repeat("中", 21), "abcd😀😃abcdefghijklmnop", "abcd😀x😃abcdefghijklmnop"} {
				t.Run(endpoint+"/"+text+map[bool]string{true: "/mimic", false: "/passthrough"}[mimic], func(t *testing.T) {
					gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+endpoint, nil)
					body := upstreamAFingerprintBody(t, text)
					body, err := sjson.SetBytes(body, "model", "claude-haiku-4-5")
					require.NoError(t, err)
					block, err := buildBillingAttributionBlockJSON(body, "2.1.81")
					require.NoError(t, err)
					body, err = sjson.SetRawBytes(body, "system", append(append([]byte("["), block...), ']'))
					require.NoError(t, err)
					cachedUA := "claude-cli/2.9.0 (external, cli)"
					svc := &GatewayService{cfg: &config.Config{}, identityService: NewIdentityService(
						&userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: cachedUA, ClientID: "core-utf16", UpdatedAt: time.Now().Unix()}},
					)}
					account := &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeOAuth}
					var req *http.Request
					if endpoint == "messages" {
						req, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "test-token", "oauth", "claude-haiku-4-5", false, mimic)
					} else {
						req, err = svc.buildCountTokensRequest(context.Background(), c, account, body, "test-token", "oauth", "claude-haiku-4-5", mimic)
					}
					require.NoError(t, err)
					defer req.Body.Close()
					wire, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					wantUA := cachedUA
					if mimic {
						wantUA = claude.DefaultHeaders["User-Agent"]
					}
					require.Equal(t, wantUA, getHeaderRaw(req.Header, "User-Agent"))
					sample := "中中中"
					if strings.HasPrefix(text, "abcd😀😃") {
						sample = "😃m"
					} else if strings.HasPrefix(text, "abcd😀x") {
						sample = "\ufffd\ufffdl"
					}
					version := ExtractCLIVersion(wantUA)
					want := "cc_version=" + version + "." + upstreamAFingerprintForSample(sample, version) + ";"
					require.Contains(t, string(wire), want)
					require.Equal(t, text, gjson.GetBytes(wire, "messages.0.content").String())
					require.Equal(t, string(wire), string(syncBillingHeaderVersion(wire, wantUA)))
				})
			}
		}
	}
}
