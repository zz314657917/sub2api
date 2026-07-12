### PASS: upstream-v0151-followups-s71-s73

## Findings

- No blocking implementation issue remains across S71-S73.
- S71 independent review initially found the new user-scope locale keys under `betaPolicy` while the UI read `openaiFastPolicy`. The Generator moved all five English and Chinese keys, added a real-locale regression assertion, and both the reviewer and integration QA confirmed the finding closed before integration.
- All business changes remain inside the approved contracts. Payment concurrency `fc66a30ff`, Cyber migration, deployment, production configuration, and the unrelated main-worktree knowledge change were not touched.

## Executed Checks

- S71 independent integration QA: service discovery `4/4`, managed/API-key/OAuth HTTP upstream capture, parsed/passthrough WS first and follow-up capture, middleware `1/1`, DTO `1/1`, service regressions, frontend exact test, typecheck, and 14-path audit passed.
- S72 independent integration QA: backend discovery `6/6`, strict alias and billing-candidate regressions, two frontend exact tests, typecheck, and 10-path audit passed; a light S71 backend smoke also passed.
- S73 independent integration QA: discovery `2/2`, request-type matrix, leaderboard regressions, repository compile-only, production-hunk audit, and allowed-path audit passed; light S71/S72 backend smoke also passed.
- Final combined backend regression passed S71 `4`, S72 `6`, and S73 `2` exact tests plus HTTP, DTO, service, repository, leaderboard, and compile-only checks.
- Final combined frontend regression passed `SettingsView.spec.ts`, `useModelWhitelist.spec.ts`, and `UseKeyModal.spec.ts`: `3 files / 46 tests`; `vue-tsc --noEmit` passed.
- `git diff --check d471e58dc..5e4b08af9`, conflict-marker scan, worker/QA verdict audit, temporary junction cleanup, and the exact 35-path union audit passed.
- All six Generator/QA branches returned only patch-equivalent `-` entries from `git cherry`; their clean worktrees and local branches were removed after the final evidence was integrated.

## Unverified Risks

- No live OpenAI/Codex HTTP or WebSocket request was sent; transport behavior is proven by in-process HTTP servers and real local WebSocket relay capture.
- No real PostgreSQL instance was used for S73; SQL behavior is covered by SQLMock and exact SQL fragment assertions.
- Race, unrelated full backend suites, deployment, local container replacement, main merge, and remote push were not run.
- Frontend tests emit existing `router-link` stub and stale `caniuse-lite` warnings; they do not fail the suites.

## Recommendation

- `PASS / done`. The integration branch is ready for an explicit main-merge decision. Do not describe it as released until main merge and push are separately authorized and verified.
