### PASS: git-branch-consolidation-s177

## Findings

- `main` was updated only from clean isolated integration worktrees. The final local integration
  commit is `8ced00f75`; the primary worktree's untracked `outputs/` files were preserved.
- `codex/main-s135-s136-publish-20260801-234733` was merged as a history-only reconciliation. The
  resulting merge tree stayed byte-identical to pre-merge `main`, so the old branch's superseded
  source state was not brought back; its focused OpenAI retry behavior passed on the isolated line.
- The stacked S156-S161 email line was resolved and verified as one explicit merge. It includes SMTP
  compatibility, registration alias dedup, OAuth email completion, dark-theme balance contrast,
  notification templates and operations report templates; it was not represented as a single S157
  alias-only change.
- The original S135 implementation is source-equivalent to the published mainline implementation,
  and the detached OpenAI Messages patch is patch-id equivalent to `9544a268a`; neither was merged a
  second time. The uncommitted i18n collision test was rejected because its full-directory glob does
  not match the locale assembly topology and would report false collisions.
- Retained candidates are not safe for blind merge: S132 stacks Passkey/Kimi/Model Plaza/Codex
  manifest work and has 13 conflict files; S169 has 32 conflict files plus dirty Prompt Audit work.
  The detached 429 WIP was fully superseded by current main and was archived before cleanup. Remote
  refs and upstream refs were not changed.

## Executed Checks

- Isolated merge trial in `E:/codex-worktrees/sub2api-branch-consolidation-20260804`.
- `go test ./internal/service -run "Test(IsOpenAIServerOverloadedError|OpenAISameAccountRetryPolicy|OpenAIGatewayService_Forward_(ModelCapacityError|ServerOverloaded)|OpenAIGatewayService_OpenAIPassthrough_(Capacity400|ServerOverloaded|429And529)|OpenAIStreaming(ResponseFailedBeforeOutputCapacity|ResponseFailedBeforeOutputServerOverloaded|PassthroughResponseFailedBeforeOutputCapacity|PassthroughResponseFailedBeforeOutputServerOverloaded))" -count=1` -> PASS.
- `go test ./internal/handler -run "Test(SameAccountRetryLimit|SameAccountRetryDelay|HandleFailoverError_SameAccountRetry)" -count=1` -> PASS.
- `go test ./... -run "^$"` -> PASS; all backend packages compiled with tests skipped.
- S156-S161 focused repository/service/handler/middleware tests -> PASS, including alias guards,
  OAuth completion, notification templates, operations reports and backend-mode OAuth boundaries.
- S156-S161 frontend Vitest -> PASS, 4 files / 43 tests; `pnpm run typecheck` and production build
  -> PASS, 1864 modules transformed.
- `go generate ./cmd/server` -> PASS; the conflict-resolved Wire output stayed stable and preserved
  both Prompt Audit coordinator wiring and notification-email wiring.
- Mainline 429 regressions -> PASS for safe `Retry-After` propagation, Messages rate-limit failover,
  Responses/passthrough pool retry and compact-keepalive output accounting. The Chat case remains
  source-identical but its unit-tag package is blocked by existing unrelated test compile failures.
- `git diff --check`, `git ls-files -u`, and worktree status checks -> PASS/clean for the temporary
  integration worktree and the removed candidate worktree.
- Mainline update: `git merge --ff-only codex/s177-integration` -> PASS.
- Mainline update: `git merge --ff-only codex/s157-integration-20260804` -> PASS.

## Cleanup Receipt

- Removed worktree and local branch `codex/main-s135-s136-publish-20260801-234733`.
- Removed temporary worktree and local branch `codex/s177-integration`.
- Created and retained recovery branch `backup/pre-s177-main-20260804` at the pre-update `main`.
- Created and retained recovery branch `backup/pre-s157-merge-20260804` before the email-stack merge.
- Removed local branches/worktrees for S156, stacked S157, client-IP S140, i18n S143, path-guard
  S136 and the source-equivalent original S135; removed the duplicate detached Messages worktree.
- Preserved four dirty worktree states as named stashes before cleanup: `d52a6b61f`, `ccac601e7`,
  `efd01586b` and `919c5052b`.
- Archived and removed the detached 429 WIP after confirming it would only revert newer mainline
  protections; recovery stash: `abeba42fc`.
- No remote branch deletion, push, Docker, database, deployment or production action occurred.

## Unverified Risks

- S132 and S169 still need independent behavior-level integration contracts and conflict resolution.
- Migration `190` was not executed against PostgreSQL. Real SMTP, OAuth provider callbacks,
  administrator browser flows, deployment and production behavior remain unverified for S156-S161.
- S176 remains `BLOCKED / browser-tool`; this branch task does not provide the required browser
  screenshot or change that QA verdict.
- The full runtime Go suite was not rerun here; the known Redis `127.0.0.1:1` baseline failure and
  external-provider/deployment behavior remain outside this Git-maintenance gate.

## Recommendation

`PASS / local-consolidation`：保留当前主线。后续只对 S132 或 S169 建立独立行为合同并逐提交
适配；不要整支合并，也不要删除远端 refs。
