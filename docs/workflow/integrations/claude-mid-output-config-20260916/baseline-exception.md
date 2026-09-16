# Baseline regression evidence — 2026-09-16

Controller created a clean detached worktree at
E:/codex-worktrees/sub2api/integration-baseline-dc851ea3f from exact commit
dc851ea3fa3fd0beb80b29ed5a68d4ac14dd9779, with no candidate implementation.

Executed from backend:
`go test ./internal/service -run '^TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation$' -count=1`

Exit 1; package duration 5.561s; claude_code_validator_test.go:52: Should be false.
This is a pre-existing failure, not a green full-suite result. Both candidate
full suites independently reported the same failure. It must not be mixed into
the account/image or output-config patch.

Proposed narrowly scoped acceptance exception: still execute and report the full
service suite. Only this exact baseline test failure may remain for local scoped
integration; every other test must pass, including all existing account identity
and new contract cases. If another failure occurs, classify and fix before PASS.
This exception requires independent contract review. It is not release approval.

## Additional baseline failures — pending independent amendment review

Both independent Claude QA and controller ran in the clean baseline:
`go test ./internal/service -run '^(TestApplyCodexOAuthTransform_PreservesAllowedTools|TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested|TestParseSSEUsage_SelectiveParsing)$' -count=1`

Controller exit 1, duration 0.072s, all three reproduced:
- TestApplyCodexOAuthTransform_PreservesAllowedTools: openai_codex_transform_test.go:18, Should be false.
- TestApplyCodexOAuthTransform_CompactsOverlongCallIDsWhenPreserveRequested: openai_codex_transform_test.go:275, expected fc_6336... but actual fc_4774....
- TestParseSSEUsage_SelectiveParsing: openai_gateway_service_test.go:3259, expected 9, actual 1.

Propose extending the same narrow exception to these three exact name/location/
assertion failures. Still run the unfiltered full suite, enumerate every fail
event (prefer go test -json Action=fail to avoid log truncation), preserve exit
status FAIL, and block all other failures. No production fix or release claim.
