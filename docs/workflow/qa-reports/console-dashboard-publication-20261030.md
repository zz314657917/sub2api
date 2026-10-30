### PASS: console-dashboard-publication-20261030

User approved the current dashboard and requested organizing, committing and pushing the completed changes. Retain the existing credit symbol.

## Reviewed scope

33 product/design files comprise the original28 console design files plus the actual-spend follow-up, with overlaps deduplicated. Changes include the warm console shell, grouped navigation, three direct header shortcuts, dashboard layouts, today's Tokens, selected-range cache rate, model/Token analytics and per-record actual-credit composition. Four backend source/test files extend read-only user trend statistics; pricing, ledger writes and schema remain unchanged.

Six task-specific contract/review/QA documents accompany the sources; this report is the seventh document. Shared status/main-log/knowledge documents, Pelican work, dependency/security updates, local fixtures and generated build outputs are excluded.

## Fresh commit-snapshot verification

Prepared detached worktree E:/codex-worktrees/sub2api/console-publication-20261030 from8f202b732, copied only the33 intended source files and six evidence documents. Its backend module/lockfiles and all other sources are the committed baseline, so unrelated dirty backend changes do not serve as build prerequisites.

- Frontend targeted dashboard/layout/view tests:13 files,88 tests PASS.
- Frontend production build including vue-tsc PASS; output stays outside the repository. Reused the existing installed node_modules via a read-only junction; no dependency installation or lockfile change. This is not a fresh lockfile installation test.
- Repository focused tests with unit tag and fresh count: PASS.
- Full backend go build ./...: PASS using existing Go1.26.10 with committed go.mod/go.sum.
- Earlier independent candidate and integrated-main acceptance:13 SQL arithmetic cases and58 Chromium assertions PASS; two browser sessions closed with owned_processes=[]. See the two task QA reports.
- Original task target hashes are checked before staging; candidate/main staged trees must match. All prepublication tracked/untracked source bytes are protected externally and rechecked after commit.

Evidence: E:/codex-runtime/sub2api/dashboard-spend-share-20261030/publication/{manifest.json,frontend-tests.log,frontend-build.log,go-tests.log,go-build.log}. Publication record will include exact commit/tree/remote hashes and final preservation results.

## Boundaries

No production DB/provider runtime acceptance or deployment is claimed. Existing frontend chunk-size warning remains. Generated preview, processes and unrelated work stay local. Pushing main also publishes its three previously local ancestor commits f88b54e0c,424b96a6e and8f202b732; they are not part of this new commit's file delta.
