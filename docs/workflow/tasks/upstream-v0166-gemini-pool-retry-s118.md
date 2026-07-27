---
task_id: upstream-v0166-gemini-pool-retry-s118
status: done
role: Developer Worker
qa_mode: runtime
---

# Task Contract

## Goal

Adapt upstream `fd7e2039d` so Gemini API-key pool forwarding preserves the
existing configured same-account retry decision for failover-worthy upstream
errors that are skipped by the error-policy layer.

## Success Criteria

- Pool-mode Gemini 429 enters failover and sets `RetryableOnSameAccount=true`
  when its configured retryable-status policy allows it.
- Pool-mode 500 enters failover but keeps same-account retry disabled unless
  500 is explicitly configured as retryable.
- Pool-mode 400 remains a mapped/pass-through error; non-pool behavior and
  error-policy matched/temporary-unscheduled behavior stay unchanged.
- Chat-completions, HTTP messages, and native messages use consistent pool
  failover semantics without changing retry limits, cooldowns, scheduling, or
  account-selection logic.

## Allowed Paths

- backend/internal/service/gemini_chat_completions_compat_service.go
- backend/internal/service/gemini_messages_compat_service.go
- backend/internal/service/gemini_pool_retry_test.go
- docs/workflow/status.md
- docs/workflow/spec.md
- docs/workflow/main-log.md
- docs/workflow/tasks/upstream-v0166-gemini-pool-retry-s118.md
- docs/workflow/worker-results/upstream-v0166-gemini-pool-retry-s118-result.md
- docs/workflow/qa-reports/upstream-v0166-gemini-pool-retry-s118-qa.md

## Denied Paths

- backend/ent/**
- backend/migrations/**
- backend/internal/handler/**
- backend/internal/repository/**
- backend/internal/server/**
- backend/internal/service/** except the listed Gemini files
- frontend/**
- deploy/**
- Dockerfile*
- knowledge/**
- outputs/**

## Constraints

- Reuse `shouldFailoverGeminiUpstreamError`, `IsPoolMode`, and
  `IsPoolModeRetryableStatus`; do not add a new retry policy or make retry
  unconditional for the same account.
- Preserve current operation-error logging and response bodies.
- Do not use the dirty primary worktree. All changes remain isolated until
  independently reviewed and merged.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v0166-gemini-pool-retry-s118/backend
go test ./internal/service -run "TestGeminiPoolModeSkippedFailover" -count=1
go test ./... -run "^$"
gofmt -d internal/service/gemini_chat_completions_compat_service.go internal/service/gemini_messages_compat_service.go internal/service/gemini_error_policy_test.go
cd E:/codex-worktrees/sub2api/upstream-v0166-gemini-pool-retry-s118
git diff --check
git diff --name-only HEAD
```

## Output

- Narrow Gemini service adaptation, focused regression evidence, developer
  result, QA report, and an allowlist-constrained diff.

## Stop Rules

- Stop if the change requires a scheduler, persistence, cooldown, retry-count,
  route, or frontend change.
- Stop if configuring the same-account retry flag cannot be separated from
  global error-policy behavior.
