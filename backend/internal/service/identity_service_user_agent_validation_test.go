package service

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

type userAgentValidationCache struct {
	fingerprint *Fingerprint
	setCalls    int
	lastSet     *Fingerprint
}

func (c *userAgentValidationCache) GetFingerprint(_ context.Context, _ int64) (*Fingerprint, error) {
	if c.fingerprint == nil {
		return nil, nil
	}
	clone := *c.fingerprint
	return &clone, nil
}

func (c *userAgentValidationCache) SetFingerprint(_ context.Context, _ int64, fp *Fingerprint) error {
	c.setCalls++
	clone := *fp
	c.lastSet = &clone
	c.fingerprint = &clone
	return nil
}

func (c *userAgentValidationCache) GetMaskedSessionID(_ context.Context, _ int64) (string, error) {
	return "", nil
}

func (c *userAgentValidationCache) SetMaskedSessionID(_ context.Context, _ int64, _ string) error {
	return nil
}

func userAgentValidationHeaders(ua string) http.Header {
	headers := http.Header{}
	if ua != "" {
		headers.Set("User-Agent", ua)
	}
	return headers
}

func TestIsAcceptableFingerprintUserAgent(t *testing.T) {
	cases := []struct {
		name string
		ua   string
		want bool
	}{
		{"official_claude", "claude-cli/2.1.92 (external, cli)", true},
		{"allowed_next_major", "claude-cli/4.0.0", true},
		{"valid_other_product", "some-sdk/999.2.3 (node)", true},
		{"local_suffix", "claude-cli/999.0.0-local", false},
		{"dev_suffix", "claude-cli/2.1.92-dev", false},
		{"build_suffix", "claude-cli/2.1.92+build", false},
		{"sentinel_major", "claude-cli/999.0.0", false},
		{"empty", "", false},
		{"missing_version", "claude-cli", false},
		{"two_segments", "claude-cli/2.1", false},
		{"leading_junk", "junk claude-cli/2.1.92", false},
		{"overlong", strings.Repeat("a", 300) + "/1.2.3", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isAcceptableFingerprintUserAgent(tc.ua))
		})
	}
}

func TestGetOrCreateFingerprintRejectsMalformedUserAgentOnCreate(t *testing.T) {
	cache := &userAgentValidationCache{}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/999.0.0-local"))

	require.NoError(t, err)
	require.Equal(t, defaultFingerprint.UserAgent, fp.UserAgent)
	require.Equal(t, defaultFingerprint.UserAgent, cache.lastSet.UserAgent)
}

func TestGetOrCreateFingerprintRejectsSentinelVersionOnUpgrade(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: "claude-cli/2.1.91", ClientID: "client-id", UpdatedAt: time.Now().Unix()}}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/999.0.0"))

	require.NoError(t, err)
	require.Equal(t, "claude-cli/2.1.91", fp.UserAgent)
	require.Zero(t, cache.setCalls)
}

func TestGetOrCreateFingerprintStillUpgradesOnValidNewerVersion(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: "claude-cli/2.1.91", ClientID: "client-id", UpdatedAt: time.Now().Unix()}}
	newUA := "claude-cli/2.1.93 (external, cli)"
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders(newUA))

	require.NoError(t, err)
	require.Equal(t, newUA, fp.UserAgent)
	require.Equal(t, 1, cache.setCalls)
}

func TestGetOrCreateFingerprintAcceptsValidUserAgentOnCreate(t *testing.T) {
	cache := &userAgentValidationCache{}
	ua := "claude-cli/" + claude.CLICurrentVersion + " (external, cli)"
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders(ua))

	require.NoError(t, err)
	require.Equal(t, ua, fp.UserAgent)
	require.NotEmpty(t, fp.ClientID)
}

func TestDefaultFingerprintUserAgentIsAcceptable(t *testing.T) {
	require.True(t, isAcceptableFingerprintUserAgent(defaultFingerprint.UserAgent))
}

func TestGetOrCreateFingerprintHealsPoisonedCacheUsingValidClientUA(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: "claude-cli/999.0.0-local", ClientID: "client-id", UpdatedAt: time.Now().Unix()}}
	realUA := "claude-cli/2.1.93 (external, cli)"
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders(realUA))

	require.NoError(t, err)
	require.Equal(t, realUA, fp.UserAgent)
	require.Equal(t, "client-id", fp.ClientID)
	require.Equal(t, "client-id", cache.lastSet.ClientID)
	require.Equal(t, 1, cache.setCalls)
}

func TestGetOrCreateFingerprintHealsPoisonedCacheWithoutValidClientUA(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: "claude-cli/999.0.0-local", ClientID: "client-id", UpdatedAt: time.Now().Unix()}}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/999.0.0-local"))

	require.NoError(t, err)
	require.Equal(t, defaultFingerprint.UserAgent, fp.UserAgent)
	require.Equal(t, "client-id", fp.ClientID)
	require.Equal(t, "client-id", cache.lastSet.ClientID)
	require.Equal(t, 1, cache.setCalls)
}

func TestGetOrCreateFingerprintDoesNotRewriteHealthyCache(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{UserAgent: "claude-cli/2.1.93", ClientID: "client-id", UpdatedAt: time.Now().Unix()}}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/2.1.92"))

	require.NoError(t, err)
	require.Equal(t, "claude-cli/2.1.93", fp.UserAgent)
	require.Zero(t, cache.setCalls)
}

func TestGetOrCreateFingerprintMissingUserAgentKeepsDefault(t *testing.T) {
	cache := &userAgentValidationCache{}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, http.Header{})

	require.NoError(t, err)
	require.Equal(t, defaultFingerprint.UserAgent, fp.UserAgent)
}

func TestDefaultFingerprintRetainsLocalDefaults(t *testing.T) {
	require.Equal(t, "claude-cli/2.1.92 (external, cli)", defaultFingerprint.UserAgent)
	require.Equal(t, "js", defaultFingerprint.StainlessLang)
	require.Equal(t, "0.70.0", defaultFingerprint.StainlessPackageVersion)
	require.Equal(t, "Linux", defaultFingerprint.StainlessOS)
	require.Equal(t, "arm64", defaultFingerprint.StainlessArch)
	require.Equal(t, "node", defaultFingerprint.StainlessRuntime)
	require.Equal(t, "v24.13.0", defaultFingerprint.StainlessRuntimeVersion)
}
