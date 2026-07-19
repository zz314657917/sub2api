# Task Contract: upstream-main-compat-s79

## Task ID

`upstream-main-compat-s79`

## Status

`approved`

## Role

Direct Codex implementation of a narrowly scoped low-risk compatibility batch
from upstream `v0.1.161`. The batch ports behavior onto the local topology; it
does not merge upstream history.

## Goal

Port four independent fixes from live upstream snapshot
`d4b9797ff72024960a035cf22fdd8f213e149169` onto local baseline
`e50c512746f67e4b916b596c28b4f8010fd87dc5`:

- `72b29a7d4`: preserve paid Antigravity plan type when an ineligible tier is reported.
- `1a855d3e0`: extract all Anthropic text blocks for channel-monitor challenges.
- `2264a3308`: normalize leaked trailing Claude Code `[1m]` model selectors.
- `c3f759d13`: remove the hard-coded day unit from subscription-plan validity copy.

## Success Criteria

- A paid Antigravity tier remains `Pro`/`Ultra` when `IneligibleTiers` is
  non-empty, while `SubscriptionStatus` remains `abnormal` and the first reason
  is preserved. Free, unknown, or absent tiers with an ineligible result remain
  `Abnormal`.
- Anthropic monitor responses concatenate trimmed `content[]` entries whose
  `type` is `text`, in order, separated by a newline. Thinking/tool blocks are
  ignored and other provider adapters keep their existing path extraction.
- Anthropic requests strip one or more case-insensitive trailing `[1m]`
  selectors from a non-empty base model, update both `ParsedRequest.Model` and
  `ParsedRequest.Body`, and make all Messages/Count Tokens handler paths use the
  normalized body. Non-Anthropic requests, middle-position suffixes, and the
  suffix-only model remain unchanged.
- Existing payment locale keys remain stable, but English and Chinese plan
  validity labels and validation errors no longer claim the value is always in
  days.
- Focused Go/Vitest checks, frontend typecheck/build, allowlist audit, conflict
  scan, and `git diff --check` pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s79`
- Branch: `codex/upstream-main-compat-s79`
- Baseline: `e50c512746f67e4b916b596c28b4f8010fd87dc5`
- Upstream snapshot: `d4b9797ff72024960a035cf22fdd8f213e149169`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, and this contract.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `backend/internal/service/antigravity_subscription_service.go`
- `backend/internal/service/antigravity_subscription_test.go`
- `backend/internal/service/channel_monitor_checker.go`
- `backend/internal/service/channel_monitor_checker_s79_test.go`
- `backend/internal/service/gateway_request.go`
- `backend/internal/service/gateway_request_s79_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/openai_gateway_count_tokens.go`
- `backend/internal/handler/claude_long_context_body_s79_test.go`
- `frontend/src/i18n/locales/en/payment.ts`
- `frontend/src/i18n/locales/zh/payment.ts`
- `frontend/src/i18n/__tests__/paymentValidityLocales.spec.ts`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s79.md`
- `docs/workflow/worker-results/upstream-main-compat-s79-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s79-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `knowledge/**`, global memories, and handoff/timeline files.
- `backend/ent/**`, `backend/migrations/**`, Wire/generated code, repository,
  scheduler, billing, payment, subscription assignment, Grok media, WS relay,
  Responses stream state machines, and security/auth paths.
- `deploy/**`, Docker/Compose, VERSION, README, dependency manifests, and lockfiles.
- S78 Stripe/payment runtime files and any frontend path not listed above.
- Upstream commits or behavior outside the four selected slices.

## Constraints

- Work only in the isolated S79 worktree; do not modify the dirty primary checkout.
- Adapt to the local `[]byte` `ParsedRequest.Body` topology instead of copying
  upstream `RequestBodyRef` code.
- Keep the existing unit-tagged `gateway_request_test.go` and
  `channel_monitor_checker_body_test.go` untouched. S79 parser and monitor
  coverage live in dedicated default-tag files so the exact `go test -list`
  discovery gate cannot pass without executing the new regressions.
- Keep existing locale key names (`validityDays`, `validityDaysRequired`) to
  avoid unrelated view churn.
- Do not refactor provider adapters, parser ownership, payment UI, or account models.
- New contract/result/QA files under `docs/workflow/**` are matched by the
  repository's `docs/*` ignore rule and must be explicitly staged with
  `git add -f`; ordinary tracked workflow files use normal scoped staging.
- The three affected handlers must call one shared, tested parse-and-body
  propagation helper. The dedicated handler test must prove that the helper
  returns the normalized body and must inspect the Go AST to prove that
  `GatewayHandler.Messages`, `GatewayHandler.CountTokens`, and
  `OpenAIGatewayHandler.CountTokens` all use that helper. This keeps the test
  inside the allowlist while preventing a parser-only false positive.
- Do not push, deploy, update containers, or merge S79 automatically.

## Acceptance Commands

```powershell
Push-Location backend
$env:GOCACHE = 'F:/mcplugins/sub2api/.tmp/go-cache-s79'
$env:GOTMPDIR = 'F:/mcplugins/sub2api/.tmp/go-build'
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
$serviceTests = go test ./internal/service -list '^TestS79'
if ($LASTEXITCODE -ne 0) { throw 'S79 service test discovery failed' }
@(
  'TestS79NormalizeAntigravitySubscription',
  'TestS79ExtractAnthropicMonitorText',
  'TestS79ParseGatewayRequestClaudeCodeLongContext'
) | ForEach-Object {
  if ($serviceTests -notcontains $_) { throw "Missing S79 service test: $_" }
}
go test ./internal/service -run '^TestS79' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S79 service tests failed' }
$handlerTests = go test ./internal/handler -list '^TestS79'
if ($LASTEXITCODE -ne 0) { throw 'S79 handler test discovery failed' }
if ($handlerTests -notcontains 'TestS79AnthropicNormalizedBodyPropagation') {
  throw 'Missing S79 handler body-propagation test'
}
go test ./internal/handler -run '^TestS79AnthropicNormalizedBodyPropagation$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S79 handler body-propagation test failed' }
Pop-Location

Push-Location frontend
npm.cmd run test:run -- src/i18n/__tests__/paymentValidityLocales.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S79 locale test failed' }
npm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S79 typecheck failed' }
npm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S79 production build failed' }
Pop-Location

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S79 diff check failed' }
```

The locale spec contains exactly two named cases (English and Chinese). Each
case asserts that both `validityDays` and `validityDaysRequired` still exist and
equal the contract's unit-neutral copy, so all four values are covered without
renaming keys.

Evaluator additionally confirms changed paths are within the allowlist, no
conflict markers/unmerged entries exist, every named test is discovered, and
the primary checkout's protected `knowledge/**` files still match all three
recorded SHA-256 baselines.

### Pre-commit Tracking Gate

After the worker result and QA report exist, stage only the exact allowlisted
paths. Because the three new workflow evidence files are ignored, force-add
only these exact paths:

```powershell
git add -f -- docs/workflow/tasks/upstream-main-compat-s79.md docs/workflow/worker-results/upstream-main-compat-s79-result.md docs/workflow/qa-reports/upstream-main-compat-s79-qa.md
git ls-files --error-unmatch docs/workflow/tasks/upstream-main-compat-s79.md docs/workflow/worker-results/upstream-main-compat-s79-result.md docs/workflow/qa-reports/upstream-main-compat-s79-qa.md
if ($LASTEXITCODE -ne 0) { throw 'S79 workflow evidence is not tracked' }
```

The evaluator then audits `git diff --cached --name-only` against the complete
allowlist before a commit is permitted.

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s79-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s79-qa.md`
- Workflow log/status entries for contract review, implementation, QA, and verdict.

## Stop Rules

- Stop if implementation requires a denied path, migration, dependency update,
  deployment change, or product-level subscription/security decision.
- Stop if `[1m]` normalization would need to affect non-Anthropic protocols or
  a model suffix other than the exact trailing selector.
- Stop if any named S79 test is not discovered, even when `go test -run`
  otherwise exits successfully.
- Stop if the normalized body cannot be proven to propagate through all three
  affected handler methods within the allowed paths.
- Stop if unit-neutral copy requires renaming a locale key or editing a view.
- Stop if the contract, worker result, or QA report is not tracked at the
  pre-commit gate.
- Stop if any protected primary-checkout knowledge hash differs from its
  recorded baseline.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a small bounded compatibility batch`
- qa_mode: `independent Codex review plus fresh focused commands`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`
