---
task_id: upstream-v0169-classifier-regression-integration
status: contract-approved
role: Generator
qa_mode: runtime
source_commits:
  - 2aef9905e
  - 55da5fc04
base_commit: 580ecea3c
---

# Task Contract

## Goal

Port the two remaining, independently applicable S169 regressions: accept the
official Claude Code security-monitor classifier's optional `<category>` element
and assert that mapped Anthropic count-token requests still remove `max_tokens`.

## Success Criteria

- Security-monitor recognition remains single-entry, type, prefix, minimum-length,
  and all other marker constrained while allowing `<category>` between `<block>yes` and `<reason>`.
- The mapping regression proves `ForwardCountTokens` strips `max_tokens` after model mapping.
- No routing, account, configuration, deployment, migration, or client behavior changes.

## Allowed Paths

- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go`
- `docs/workflow/tasks/upstream-v0169-classifier-regression-integration.md`
- `docs/workflow/qa-reports/upstream-v0169-classifier-regression-integration-qa.md`

## Denied Paths

- All migrations, configuration, frontend, deployment, Docker, CI, release,
  Prompt Audit, Passkey, pricing, account-selection, proxy-circuit, primary-worktree,
  remote, and push paths.

## Constraints

- Apply only source commits `2aef9905e` and `55da5fc04`.
- Do not weaken the security-monitor classifier beyond accepting the optional category element.
- Stop on a source conflict, changed marker coverage, or focused test failure.

## Acceptance Commands

```powershell
Set-Location backend
go test ./internal/service -run 'Test.*(ClaudeCodeValidator|CountTokens|ModelMappingPreservesOtherFields)' -count=1
go test ./... -run '^$'
go build ./...
Set-Location ..
git diff --check
git ls-files -u
```

## Contract Review

`PASS / contract-approved`: both source patches apply cleanly to `main@580ecea3c`.
The classifier keeps its independent long-prompt boundary and only relaxes one
literal output-adjacency marker. The count-token update changes a regression
assertion only; the corresponding production sanitization is already on main.
