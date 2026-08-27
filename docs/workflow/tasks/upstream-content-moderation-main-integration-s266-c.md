# Upstream Content Moderation Main Integration S266-C

## Task ID

upstream-content-moderation-main-integration-s266-c

## Role

Planner and Final Evaluator: Codex. Generator: Codex Controller performing a
Git-only integration of two already reviewed product commits. QA: an independent
native `gpt-5.6-terra` QA Worker after the Controller completes the integration
and protection checks. No new product implementation is authorized.

## Goal

Integrate the independently accepted S266-A and S266-B content-moderation
product commits into the current local `main`, preserving all unrelated main
history and the protected `outputs/` tree. Re-run the combined acceptance suite
on the resulting main worktree. Do not push, deploy, rebuild containers, or
contact live external services.

## Success Criteria

- Starting point is local `main@2a3664747`, with `origin/main@e5b62a9b9`, a
  clean tracked/index state, and only untracked `outputs/`. The 20-file outputs
  manifest SHA-256 is
  `2996311A4EC1458EEC9C2AE4327D5D5EAA695C878783DE984AF841BBF0A79145`.
- Integrate S266-A product commit `c2cd7a0a1` before S266-B product commit
  `eeed2369f`, retaining separate commits and `-x` provenance. Their patch IDs
  must remain `922e3bc0fc4ccf5c1bd1fecc41e33e8894ac8d0a` and
  `1c0333b260eb320ffeb89982757756a9e8c26df1` respectively.
- The product integration changes only the exact union of
  `git diff-tree --no-commit-id --name-only -r c2cd7a0a1` and
  `git diff-tree --no-commit-id --name-only -r eeed2369f`. The frozen main delta
  has zero path overlap with either product commit; Pixel Cafe, wallet/billing,
  workflow state, lockfiles and `outputs/**` must remain unchanged.
- Combined fresh-main focused backend tests, server compilation, frontend
  Vitest/typecheck/build, gofmt, diff, provenance, conflict, index and protected
  outputs gates pass. Known independent tagged API snapshot and full repository
  billing-fixture drift remain outside this integration.
- The QA report records a first-line `PASS`, `FAIL`, or `BLOCKED` verdict and
  the Final Evaluator updates the candidate workflow/handoff evidence. Local
  `main` must not be pushed or deployed.

## Context

- Main repo: `F:/mcplugins/sub2api`
- Evidence worktree:
  `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266`
- S266-A independent QA: `bec523227`
- S266-B independent QA: `89f2d8869`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`,
  `knowledge/tasks/current-task.md`, and both S266-A/B contracts in the evidence
  worktree.

## Allowed Paths

- In local `main`, only the exact 67-path union reported by the two product
  commits above; the commits contain 21 and 56 paths with 10 shared owners.
- In the evidence worktree only:
  `docs/workflow/tasks/upstream-content-moderation-main-integration-s266-c.md`,
  `docs/workflow/qa-reports/upstream-content-moderation-main-integration-s266-c-qa.md`,
  `docs/workflow/status.md`, `docs/workflow/main-log.md`, and
  `knowledge/tasks/current-task.md`.

## Denied Paths

- Any local-main path outside the exact product-commit union, especially
  `outputs/**`, Pixel Cafe sources/tests/assets, wallet/billing changes,
  workflow/knowledge files, dependencies, lockfiles, deployment/container or
  production configuration.
- New product code, test rewrites, schema generation, any migration other than
  the already reviewed migration 237 file, shared data, live provider/SMTP/
  Redis/PostgreSQL calls, container update, staging, push, force operation or
  history rewrite.
- The protected untracked worker directory
  `upstream-content-moderation-cyber-policy-s266-b/` in the evidence worktree.

## Constraints

- Integrate A then B one batch at a time with non-interactive `git cherry-pick
  -x`; stop on any conflict instead of resolving across an owner boundary.
- Capture and compare main HEAD, exact status, staged/unmerged index, product
  scope, patch IDs and outputs manifest before and after integration.
- Do not use broad `git add`, reset, checkout, clean, stash or amend. Do not
  absorb concurrent work; if main changes after the frozen baseline, stop and
  re-contract against the new state.
- Migration 237 may be committed as source but must not be applied to any
  database in this task.

## Acceptance Commands

```powershell
Set-Location F:/mcplugins/sub2api/backend
go test ./internal/service -list 'Cyber|ContentModeration|KeywordMatcher|OpenAI.*Policy'
go test ./internal/handler -list 'Cyber|OpenAI'
go test ./internal/handler/admin -list 'ContentModeration|Cyber|Settings|RiskControl'
go test ./internal/repository -list 'Cyber|ContentModeration|OpsError'
go test ./internal/service -run 'Cyber|ContentModeration|KeywordMatcher|OpenAI.*Policy' -count=10
go test ./internal/handler -run 'Cyber|OpenAI' -count=10
go test ./internal/handler/admin -run 'ContentModeration|Cyber|Settings|RiskControl' -count=10
go test ./internal/repository -run 'Cyber|ContentModeration|OpsError' -count=10
go test ./migrations -run 'ContentModerationMatchedKeyword' -count=1
go test ./internal/service ./internal/handler ./internal/handler/admin -count=1
go test ./cmd/server -run '^$' -count=1

Set-Location F:/mcplugins/sub2api/frontend
node node_modules/vitest/vitest.mjs run src/features/prompt-audit/__tests__/integrationSurface.spec.ts src/views/admin/__tests__/RiskControlView.spec.ts
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build

Set-Location F:/mcplugins/sub2api
git diff --check 2a3664747..HEAD
git ls-files -u
git diff --cached --name-only
```

## Output

- QA writes only
  `docs/workflow/qa-reports/upstream-content-moderation-main-integration-s266-c-qa.md`
  in the evidence worktree.
- The report first line must be
  `### PASS: upstream-content-moderation-main-integration-s266-c`,
  `### FAIL: upstream-content-moderation-main-integration-s266-c`, or
  `### BLOCKED: upstream-content-moderation-main-integration-s266-c`.
- The Final Evaluator records integration commits, commands, baseline failures,
  protected-state comparison and final verdict in workflow/handoff files.

## Stop Rules

- Stop if local `main` is no longer exactly `2a3664747` before the first
  cherry-pick, tracked/index state is dirty, or outputs manifest changes.
- Stop on any cherry-pick conflict, product-scope drift, patch-ID mismatch,
  unexpected migration/dependency/lockfile change, protected-path change, or
  empty focused-test discovery.
- Stop if verification requires live external state or a product fix. A product
  defect must return to a separate fix/retest contract; it cannot be patched
  during integration.

## Budget

- developer_worker_model: not used; integration is Controller-owned
- qa_worker_model: `gpt-5.6-terra`
- no explicit token or USD budget requested
