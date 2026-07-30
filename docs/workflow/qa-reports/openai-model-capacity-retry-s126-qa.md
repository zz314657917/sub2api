### PASS: openai-model-capacity-retry-s126

# QA Report

## Task ID

`openai-model-capacity-retry-s126`

## Verdict

`PASS / source-only / local-integrated`

## Contract Checked

- `docs/workflow/tasks/openai-model-capacity-retry-s126.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- business commit: `a986b86ef211c78603d732c16966d5bb3a30a5bf`
- local integration: primary `main` fast-forwarded from `7694d48db` to
  `a986b86ef`; the four target paths had no dirty or staged overlap.
- commands run:

```text
go test ./internal/service -run <S126 focused regex> -count=1 -> PASS
go test ./internal/service -run <broader OpenAI stream/passthrough regex> -count=1 -> PASS
go test ./internal/handler -run 'TestHandleFailoverError_(SameAccountRetry|BasicSwitch|IntegrationScenario)' -count=1 -> PASS
go test ./... -run '^$' -> PASS in isolated worktree
go test ./internal/service -count=1 -> FAIL in unrelated peak-rate global-timezone tests
go test ./internal/service -run '^TestPeakMultiplier' -count=1 -> PASS in isolation
gofmt -d <four S126 Go files> -> PASS, no output
git diff --check -> PASS
rg conflict markers -> PASS, no markers
Allowed Paths audit -> PASS
primary main: focused S126 service tests -> PASS
primary main: go test ./... -run '^$' -> PASS
```

- manual checks:

```text
normal non-pool HTTP 400 capacity -> failover returned, same-account retry true, client writer untouched
passthrough HTTP 400 capacity -> failover returned, same-account retry true, client writer untouched
standard and passthrough pre-output response.failed capacity -> same-account retry true
generic transient 400 and generic/server-overloaded stream failures -> failover retained, same-account retry false for non-pool account
passthrough non-capacity 400 -> original response still written to client
passthrough 429/529 -> existing failover retained, capacity-only retry flag remains false
context-window and policy/post-output behavior -> existing focused regressions retained
```

## Findings

- 未发现 S126 范围内的明确问题。
- The complete `internal/service` suite has a pre-existing global-timezone test
  isolation problem: other tests initialize `Asia/Shanghai` without restoring
  the package-global location, so peak-rate tests expecting UTC can fail when
  the whole package runs together. The failing peak tests pass alone, their
  source has no S126 diff, and all S126 focused tests plus full-repository
  compilation pass.
- Real OpenAI capacity pressure, wall-clock observation of three 500 ms retries,
  deployment, container refresh, push, and production client behavior were not
  executed.

## Bug Owner Recommendation

`integration-owner` for a separate cleanup of the unrelated timezone-global
test isolation issue; no S126 implementation fix is required.

## Root Cause

`none` for S126. The non-gating full service-suite failure is a separate
`test-bug` caused by mutable package-global timezone state.

## Retest Scope

- No S126 fix/retest loop is required. Before release, repeat the focused
  normal/passthrough/stream capacity tests and one real upstream smoke when a
  capacity response can be observed safely.

## Knowledge Promotion

- `none`
