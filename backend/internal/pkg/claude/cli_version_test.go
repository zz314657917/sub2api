package claude

import "testing"

func TestIsSupportedCLIVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		want    bool
	}{
		{"builtin", CLICurrentVersion, true},
		{"higher patch", "2.1.259", true},
		{"higher minor", "2.2.0", true},
		{"higher major", "3.0.0", true},
		{"below builtin", "2.1.219", false},
		{"empty", "", false},
		{"two segments", "2.2", false},
		{"v prefix", "v2.2.0", false},
		{"prerelease", "2.2.0-local", false},
		{"build metadata", "2.2.0+build1", false},
		{"sentinel suffix", "999.0.0-local", false},
		{"numeric sentinel", "999.0.0", true},
		{"non numeric", "abc", false},
		{"extra segment", "2.2.0.1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSupportedCLIVersion(tc.version); got != tc.want {
				t.Fatalf("IsSupportedCLIVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}

func TestResolveCLIVersion(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"unset", "", CLICurrentVersion},
		{"whitespace", "   ", CLICurrentVersion},
		{"valid override", "2.1.259", "2.1.259"},
		{"trimmed override", "  2.1.259  ", "2.1.259"},
		{"invalid fallback", "not-a-version", CLICurrentVersion},
		{"lower fallback", "2.0.0", CLICurrentVersion},
		{"prerelease fallback", "2.2.0-local", CLICurrentVersion},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCLIVersion(tc.raw); got != tc.want {
				t.Fatalf("resolveCLIVersion(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestDefaultHeadersUserAgentMatchesCLIVersion(t *testing.T) {
	want := "claude-cli/" + CLIVersion() + " (external, cli)"
	if got := DefaultHeaders["User-Agent"]; got != want {
		t.Fatalf("DefaultHeaders[User-Agent] = %q, want %q", got, want)
	}
}

func TestCLIVersionDefaultsToBuiltinPin(t *testing.T) {
	if got := CLIVersion(); got != CLICurrentVersion {
		t.Fatalf("CLIVersion() = %q, want built-in pin %q (test process must not set %s)", got, CLICurrentVersion, CLIVersionEnv)
	}
}
