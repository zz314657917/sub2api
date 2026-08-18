### PASS: upstream-cn-provider-403-s229-c

# QA Report

## Task ID

upstream-cn-provider-403-s229-c

## Verdict

PASS

## Contract Checked

- `docs/workflow/tasks/upstream-cn-provider-403-s229-c.md`

## Evidence

- diff reviewed: yes
- allowed paths checked: yes; the business diff contains only the ratelimit owner and focused CN 403 tests, plus the Controller result report
- denied paths touched: no
- independent QA worktree: `E:/codex-worktrees/sub2api/upstream-cn-provider-403-s229-c-qa`
- commands run:

```text
backend/go test ./internal/service -run "TestHandleUpstreamError_CNProviderHTML403SkipsAccountPenalty|TestHandleUpstreamError_CNProviderStructured403TempUnschedulable|TestHandleUpstreamError_CNProviderStructured403ThresholdDisables" -count=10 -> PASS (0.364s)
backend/go test ./cmd/server -run "^$" -count=1 -> PASS (0.057s)
backend/go test ./internal/service -count=1 -> PASS (65.661s)
backend/gofmt -d internal/service/ratelimit_service.go internal/service/ratelimit_service_cn_403_test.go -> PASS (no output)
backend/git diff --check -> PASS
backend/git diff --name-only --diff-filter=U -> empty
backend/git ls-files -u -> empty
backend/git merge-base --is-ancestor 10c8b7020 upstream/main -> PASS
```

## Findings

未发现明确问题。

## Behavioral Coverage

- CN HTML 403 remains a request/path failure and does not increment account penalty state.
- Structured CN 403 reuses the existing OpenAI first-hit temporary cooldown behavior.
- Repeated structured CN 403 reaches the existing threshold permanent-disable behavior.
- OpenAI, Anthropic, Antigravity, billing, stream, and disconnect behavior remain outside this slice.

## Risks

- Validation uses local repository/counter stubs only; real provider, Redis, database, container, deployment, and push operations are excluded by contract.

## Bug Owner Recommendation

original-worker

## Root Cause

none

## Retest Scope

不适用；PASS。

## Knowledge Promotion

none
