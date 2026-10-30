---
status: approved
review_verdict: PASS
task_id: pelican-radar-efforts-20261030
worker_model: gpt-6.1-sol
base_commit: 8f202b73282465a9b4e9d5bcfeb1c267362e72c4
spec_ref: docs/workflow/tasks/pelican-radar-20261030.md
---

# Task ID: pelican-radar-efforts-20261030

## Role

Controller: Planner/final integration. Independent gpt-6.1-sol Contract Reviewer, Developer and QA per existing Agent Matrix. Supplement to completed radar contract; build only after independent PASS.

## Goal

Show each selected model's selected reasoning efforts and their exact third-party scores together in one model card. Let administrators explicitly add all source-supported efforts for a model. Preserve existing selections and all original radar source/cache/auth contracts.

## Success Criteria

- Group DTO items by exact canonical model ID, preserving first model appearance and relative selected effort order. Header has ModelIcon/name once; each effort row has its own label and score/source max or exact unavailable/missing/unsupported status. Never substitute scores, select a best effort, call scores measured IQ, or add unselected combinations implicitly.
- Keep horizontal compact model cards, full desktop width, strip-only scroll on mobile, light/dark, original AppLayout/page/filters/gallery.
- User UI refinement during QA: display the source-defined maximum once as “满分 150” in the section heading (dynamic value); remove repeated `/ 150` suffixes from effort scores. Null maximum must not fabricate a number. Only Scores.vue and its tests need an R2 freeze; other four files stay unchanged.
- Add per-row admin action "添加全部档位" using catalog only, canonicalize explicit catalog aliases, append only missing supported pairs in source order, leave saved/unknown/unsupported pairs intact. Disable when catalog unavailable/unknown/no missing effort/disabled. Existing per-pair reorder/edit/remove remains.
- Raise settings validation cap from 20 to 64 pairs in backend/frontend to accommodate current source's 46 pairs. No DTO/storage shape change or automatic migration. Atomic batch add: if entire addition exceeds 64, disable with explanation; never silently truncate. Old settings and 20-pair configs remain compatible.
- Preview retains five selected models, expands each to every source-declared effort (29 pairs), derived exactly from frozen source.json and original weighted formula. Do not add other models; do not claim fixture as live backend.

## Allowed Paths

Developer in existing isolated candidate E:/codex-worktrees/sub2api/pelican-radar-20261030 only:
- frontend/src/components/pelican/PelicanReferenceScores.vue
- frontend/src/components/pelican/PelicanReferenceSettings.vue
- frontend/src/components/pelican/__tests__/PelicanReferenceScores.spec.ts
- frontend/src/components/pelican/__tests__/PelicanReferenceSettings.spec.ts
- backend/internal/service/pelican_reference_scores.go (selection bound only)
- backend/internal/service/pelican_reference_scores_test.go (bound compatibility tests only)

Controller only: this task's workflow/review/QA reports, minimal concurrent status/current-task/main-log updates, preview radar-fixture.json and README.md; evidence under E:/codex-runtime/sub2api/pelican-radar-efforts-20261030. Freeze main/candidate six targets before work; compare hashes before exact integration and preserve other dirty bytes. Reviewer/QA write external evidence and task-specific report only.

## Denied Paths

All unrelated source, dependencies, migrations, settings DTO/auth/cache/scoring algorithm, shared layout, production gallery/retention, git index/history/remotes, Docker/deployment, global memory. No commit, push, container update or real provider/DB operations.

## Constraints

Reuse previous isolated candidate/dependencies and evidence tooling. Use local Go1.26.10, GOTOOLCHAIN=local, GOFLAGS=-mod=readonly. npm.cmd; no embedded frontend dist. Do not modify task targets after source freeze while independent QA runs. Controller owns preview fixture generation; production default selections stay untouched.

## Acceptance Commands

- Focused Vue component tests plus PelicanTestsView regression tests; typecheck and targeted ESLint.
- Go TestPelicanReference settings/parser/cache focused tests, including exactly 64 accepted/65 rejected, existing 20 accepted, alias duplicate still rejected. No need to rerun unrelated handlers/build for single validation-bound change.
- New Vue tests: interleaved models grouped exactly, effort/score/status per row preserved and single model header; batch dedup/alias/source order/unsupported retention/catalog unavailable/disabled/64 boundary and no mutation.
- Independent QA on frozen six files: focused Vue/Go, typecheck, exact source-derived preview score oracle; real Edge browser original /pelican-tests at62081 desktop/mobile light/dark/sidebar and per-effort score mapping, no page overflow, original gallery/enlarge smoke. Record task-only session/profile/PIDs; close and verify owned browser/daemon zero.
- Controller final review of six-file delta, main drift protection, exact-copy integration, fresh focused main checks. Keep preview server for user.

## Output

Independent contract review docs/workflow/contract-reviews/pelican-radar-efforts-20261030-review.md; Developer/QA first-line PASS/FAIL/BLOCKED and commands/evidence/changed-file hashes. Final show existing original-page preview address and test limitations.

## Stop Rules

Contract not PASS; unavailable matrix model; unknown target drift; need broader DTO/auth/source semantics changes; cannot preserve concurrent work; unresolved QA failure. Retain original failed evidence and narrow fix/retest.
