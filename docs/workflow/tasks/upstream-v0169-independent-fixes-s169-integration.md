---
task_id: upstream-v0169-independent-fixes-s169-integration
status: contract-approved
role: Generator
qa_mode: runtime
source_commits:
  - ed70b8861
  - 72d43e62f
  - 20fb2cfc5
  - 92092ee62
  - 38bbb1fe3
  - 5aa57672c (PlatformComposite constant hunk only)
base_commit: aad349f84
---

# Task Contract

## Goal

Adapt only five independently applicable S169 fixes onto current `main`:
remove `max_tokens` from Anthropic count-token requests, expand Composite
available-channel groups by platform, log successful OPS cleanup at info level,
recognize the official Claude Code security-monitor prompt, and skip
permanently unschedulable accounts during token refresh.

## Success Criteria

- Anthropic `/v1/messages/count_tokens` request sanitization removes
  `max_tokens` without changing normal message behavior.
- Available Channels expands Composite model groups for the configured platform
  while preserving ordinary group isolation and empty behavior.
- OPS cleanup success is observable at info level without changing cleanup
  scheduling or persistence.
- Claude Code security-monitor recognition remains constrained to the official
  long prompt shape and accepts the upstream `auto` mode variant.
- Permanently unschedulable token-refresh accounts are skipped; ordinary
  retryable accounts retain their existing path.

## Allowed Paths

- `backend/internal/service/{gateway_service.go,gateway_anthropic_apikey_passthrough_test.go,gateway_context_management_test.go,ops_cleanup_service.go,claude_code_validator.go,claude_code_validator_test.go,token_refresh_service.go,token_refresh_service_test.go,domain_constants.go}`
- `backend/internal/domain/constants.go`
- `backend/internal/handler/{available_channel_handler.go,available_channel_handler_test.go}`
- `docs/workflow/tasks/upstream-v0169-independent-fixes-s169-integration.md`
- `docs/workflow/qa-reports/upstream-v0169-independent-fixes-s169-integration-qa.md`

## Denied Paths

- `backend/migrations/**`, `backend/go.mod`, `backend/go.sum`, `frontend/**`.
- `.github/**`, `Dockerfile*`, `deploy/**`, `outputs/**`, `knowledge/**`,
  `docs/workflow/status.md`, and `docs/workflow/main-log.md`.
- Passkey, Prompt Audit, SMTP, pricing, release, account scheduler, proxy
  circuit, Docker/CI, database/container, deployment, remote, and primary
  worktree changes.

## Constraints

- Apply only source commits `ed70b8861`, `72d43e62f`, `20fb2cfc5`,
  `92092ee62`, and `38bbb1fe3`, plus only the `PlatformComposite` constant
  hunk from `5aa57672c` and its `domain.PlatformComposite = "composite"`
  dependency, with no additional behavior.
- Stop on an apply conflict, a denied-path requirement, a weakened security
  classifier boundary, or a failing focused regression.
- Do not access Docker, databases, external providers, deployment, remotes, or
  push.

## Acceptance Commands

```powershell
go test ./internal/service ./internal/handler -run 'Test.*(CountTokens|AvailableChannel|Cleanup|ClaudeCode|TokenRefresh)' -count=1
go test ./... -run '^$'
go build ./...
git diff --check
git ls-files -u
```

## Output

- One local commit containing the adapted source, the approved contract, and a
  QA report whose first line is `### PASS`, `### FAIL`, or `### BLOCKED`.

## Contract Review

`PASS / contract-approved`: the five functional source commits pass
`git apply --check` against `main@aad349f84`. The available-channel compile
gate identified one required source dependency: the single
`PlatformComposite = domain.PlatformComposite` re-export hunk from
`5aa57672c`, whose source branch depends on the exact domain constant
`PlatformComposite = "composite"`. Those two lines are independently reviewed
and added to the allowlist; the rest of `5aa57672c` remains excluded because
its tests conflict with the current Gateway behavior. The rejected S169 commits
otherwise overlap current-main implementation or require denied deployment, CI,
migrations, Prompt Audit, SMTP, pricing, release, or frontend scope.
