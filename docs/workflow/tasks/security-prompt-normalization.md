---
status: approved
review_verdict: PASS
task_id: security-prompt-normalization
worker_model: gpt-5.6-terra
base_commit: c30bb5350006490eaf0dac314e6dc66746818930
spec_ref: docs/workflow/tasks/security-prompt-normalization.md
---

# Security branch batch: prompt normalization

## Task ID
security-prompt-normalization

## Role
Controller Planner/final evaluator; independent contract reviewer; separate gpt-5.6-terra Developer and QA.

## Goal
Port the prompt text normalization portion of bef505942df0f0f68b7f204e930dcd1efcec3d76 from codex/security-hardening to current main, in an isolated worktree, then safely fast-forward locally. The session-risk blocking portion and Composite gateway ownership remain separate pending batches; do not claim entire source branch integrated.

## Success Criteria
Apply NFKC, strip U+200B/U+200C/U+200D/U+2060/U+FEFF, replace NUL/backspace/vertical-tab/form-feed/U+000E/U+000F with spaces, trim outer whitespace, omit empty normalized segments before constructing prioritized scan text. Preserve the relative order of retained segments and readable internal newlines. Metadata/full prompt derived from this normalized scan representation follows original source behavior. Never mutate Request.Body, API/provider payloads, role selection, guard policy/action configuration, persistence/redaction gates or keyword evaluation semantics beyond changed scan input. Existing off/async/blocking and latest-turn behavior remain. No automatic temporary bans in this batch.
Exercise production snapshot extraction across supported text protocols (Chat, Responses, Anthropic, Gemini, images and WS), including full-width/zero-width bypass input, controls, empty segments, Unicode/emoji, prioritization, unchanged request bytes and redaction. Existing package tests remain passing. Prove normalized input reaches scan path via effective snapshot/scanner test; no paid provider required.

## Allowed Paths
- backend/internal/securityaudit/prompt_snapshot.go
- backend/internal/securityaudit/prompt_snapshot_test.go
- docs/workflow/tasks/security-prompt-normalization.md
- docs/workflow/contract-reviews/security-prompt-normalization-review.md
- docs/workflow/worker-results/security-prompt-normalization-result.md
- docs/workflow/qa-reports/security-prompt-normalization-qa.md
- docs/workflow/status.md (controller only, restore before integration)
- docs/workflow/main-log.md (controller only, restore before integration)

## Denied Paths
All other paths, main workspace changes, store/payment, migrations, dependencies, prompt_service.go/session-risk, Composite, knowledge/memories, production settings/data.

## Constraints
Isolated E:/codex-worktrees/sub2api/security-prompt-normalization; no wholesale cherry-pick. No push/deploy/browser/provider traffic. Runtime artifacts E:/codex-runtime/pge/sub2api/security-prompt-normalization. Preserve unrelated dirty files, no stash/force merge. Terra unavailable => BLOCKED, no substitution. No unrelated fixes or reformatting.

## Acceptance Commands
From backend: `go test ./internal/securityaudit -count=1`; `go test -tags unit ./internal/securityaudit -count=1`; `go build ./...`. Also targeted verbose normalization tests and gofmt/diff checks. No-tests-to-run is not PASS. Baseline failures must reproduce on immutable base; changed production behavior requires executable tests or block merge. Controller runs `git diff --cached --check`, verifies exact allowlist, main HEAD/index/dirty intersection and hashes before/after ff-only merge.

## Output
Developer result and independent QA report (first line ### PASS/FAIL/BLOCKED: security-prompt-normalization) with actual commands, findings and boundaries. Feature/evidence commits separate. Source branch stays until all other behaviors are audited or integrated.

## Stop Rules
Stop for wider allowlist, missing effective tests, changed guard policy, provider-body mutation, unrelated modifications, model unavailable or main conflict. Return issues to owner, never hide failures or lower acceptance.

## Controller source-branch inventory (2026-09-18)
Independent read-only audit of the final 22 source commits confirmed all 13 runtime topics already exist on main: ultrafast backend/UI, Astra aliases/pricing/Responses Lite exclusion/UI/cache/continuation, Claude billing/CLI baseline/override, Gemini models endpoint copy, and GLM-5.3 effort. Nine other commits contain historical evidence only; do not duplicate runtime changes merely to make ancestry identical.

The first seven Composite commits are partly superseded by main's b25cc223a administrator registry, schema, generated Ent and migration 241. Do not import the old migration 172. Main still lacks CompositeRouteResolver, account ownership and gateway selection integration. That remaining architectural batch needs scoped resolver/context/selection tests and generated Wire review, and overlaps uncommitted store changes in backend/internal/service/wire.go and backend/cmd/server/wire_gen.go. Preserve source branch and dirty work; do not force integration.

The other portion of bef505942 is not part of this batch: its risk counter is keyed by API key (user fallback), not conversation, escalates three Warn/Block results into subsequent blocking for ten minutes, and is process-local. It requires separate policy/behavior review before integration; this snapshot normalization batch must not enable it implicitly.
