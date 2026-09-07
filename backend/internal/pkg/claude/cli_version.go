package claude

import (
	"log/slog"
	"os"
	"strings"

	"golang.org/x/mod/semver"
)

// CLIVersionEnv is the optional operator override for the advertised Claude CLI version.
const CLIVersionEnv = "SUB2API_CLAUDE_CLI_VERSION"

// resolvedCLIVersion is initialized once so headers and billing cannot diverge if the
// process environment changes after startup.
var resolvedCLIVersion = resolveCLIVersion(os.Getenv(CLIVersionEnv))

// CLIVersion returns the effective Claude Code CLI version used for impersonation.
func CLIVersion() string {
	return resolvedCLIVersion
}

// IsSupportedCLIVersion accepts only plain three-part versions at or above the built-in pin.
func IsSupportedCLIVersion(version string) bool {
	version = strings.TrimSpace(version)
	if version == "" {
		return false
	}
	canonical := "v" + version
	if !semver.IsValid(canonical) || semver.Canonical(canonical) != canonical {
		return false
	}
	if semver.Prerelease(canonical) != "" || semver.Build(canonical) != "" {
		return false
	}
	return semver.Compare(canonical, "v"+CLICurrentVersion) >= 0
}

func resolveCLIVersion(raw string) string {
	version := strings.TrimSpace(raw)
	if version == "" {
		return CLICurrentVersion
	}
	if !IsSupportedCLIVersion(version) {
		slog.Warn("ignoring invalid Claude CLI version override; falling back to the built-in pin",
			"env", CLIVersionEnv,
			"value", version,
			"builtin", CLICurrentVersion,
			"requirement", "strict three-part semver, not older than the built-in pin")
		return CLICurrentVersion
	}
	return version
}
