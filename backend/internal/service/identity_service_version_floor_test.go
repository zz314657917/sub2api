package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

func TestFloorClaudeCLIUserAgentVersion(t *testing.T) {
	floor := "claude-cli/" + claude.CLICurrentVersion
	tests := []struct {
		name    string
		ua      string
		want    string
		changed bool
	}{
		{"old version", "claude-cli/2.1.220 (external, cli)", floor + " (external, cli)", true},
		{"preserve suffix", "claude-cli/2.1.100 (external, claude-desktop-3p, agent-sdk/0.3.100)", floor + " (external, claude-desktop-3p, agent-sdk/0.3.100)", true},
		{"equal floor", floor + " (external, cli)", floor + " (external, cli)", false},
		{"newer version", "claude-cli/2.9.0 (external, cli)", "claude-cli/2.9.0 (external, cli)", false},
		{"other product", "opencode/1.2.3 (external, cli)", "opencode/1.2.3 (external, cli)", false},
		{"malformed", "claude-cli/abc (external, cli)", "claude-cli/abc (external, cli)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := floorClaudeCLIUserAgentVersion(tt.ua)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.changed, changed)
		})
	}
}

func TestGetOrCreateFingerprintFloorsStaleCachedUserAgent(t *testing.T) {
	cache := &userAgentValidationCache{fingerprint: &Fingerprint{
		UserAgent:               "claude-cli/2.1.220 (external, cli)",
		ClientID:                "client-id",
		StainlessPackageVersion: "0.91.1",
		UpdatedAt:               time.Now().Unix(),
	}}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/2.1.75 (external, cli)"))
	require.NoError(t, err)
	require.Equal(t, "claude-cli/"+claude.CLICurrentVersion+" (external, cli)", fp.UserAgent)
	require.Equal(t, fp.UserAgent, cache.lastSet.UserAgent)
	require.Equal(t, "0.91.1", fp.StainlessPackageVersion)
	require.Equal(t, 1, cache.setCalls)
}

func TestCreateFingerprintFloorsOldClientUserAgent(t *testing.T) {
	cache := &userAgentValidationCache{}
	fp, err := NewIdentityService(cache).GetOrCreateFingerprint(context.Background(), 1, userAgentValidationHeaders("claude-cli/2.1.75 (external, claude-desktop-3p)"))
	require.NoError(t, err)
	require.Equal(t, "claude-cli/"+claude.CLICurrentVersion+" (external, claude-desktop-3p)", fp.UserAgent)
}
