# Task Contract

## Task ID
upstream-main-openai-endpoint-capability-ui-s2d

## Role
Codex acts as Planner, Generator, and Final Evaluator for this small frontend-focused Sprint. Implement only the OpenAI API Key endpoint capability UI clarification selected here, and stop if the patch requires backend, schema, gateway, protocol bridge, or monolithic i18n changes.

## Goal
Port the safe frontend subset of upstream `37044b83e fix(openai): clarify endpoint capability UI` onto the current upstream-sync branch after Sprint 2c. This Sprint exposes and clarifies the already-existing `credentials.openai_capabilities` routing control for OpenAI API Key accounts, and makes Responses mode visibly inapplicable when the text endpoint capability is disabled.

## Success Criteria
- OpenAI API Key create/edit forms show endpoint capability checkboxes for text generation and embeddings.
- Text endpoint label reflects the selected Responses mode:
  - `auto` -> Responses / Chat Completions auto wording.
  - `force_responses` -> Responses wording.
  - `force_chat_completions` -> Chat Completions wording.
- Disabling the text endpoint resets Responses mode to `auto`, disables the Responses mode selector, and shows an explanatory hint.
- Saving an embeddings-only API Key account writes `credentials.openai_capabilities = ['embeddings']`.
- Saving with both default capabilities omits `credentials.openai_capabilities`.
- Existing `extra.openai_responses_supported` probe result is preserved when editing.
- Local modular i18n files are used; upstream monolithic `en.ts` / `zh.ts` are not touched.
- No backend, schema, migrations, gateway routing, WS/Responses bridge, or public API behavior changes are introduced.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-account-usage-window-hints-s2c`
- Work branch: `codex/upstream-main-openai-endpoint-capability-ui-s2d`
- Upstream source commit: `37044b83e`
- Related backend capability support already exists locally:
  - `backend/internal/service/account.go`
  - `backend/internal/service/openai_account_scheduler.go`
  - `backend/internal/handler/openai_embeddings.go`
- Main worktree `F:/mcplugins/sub2api` has unrelated dirty changes and must not be modified.
- Local i18n is modular. Do not adopt upstream monolithic `frontend/src/i18n/locales/en.ts` / `zh.ts` structure.

## Allowed Paths
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/tasks/upstream-main-openai-endpoint-capability-ui-s2d.md`
- `docs/workflow/worker-results/upstream-main-openai-endpoint-capability-ui-s2d-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-endpoint-capability-ui-s2d-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/**`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS/Responses bridge redesign, OpenAI gateway routing redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Port manually if cherry-pick would touch denied monolithic i18n paths.
- Preserve local modular i18n and existing account modal behavior.
- Do not introduce Ent schema/codegen/migrations.
- Do not add real upstream smoke tests requiring live credentials.
- Do not change backend routing semantics; this Sprint only edits frontend serialization for an already-supported credential key.
- If the selected patch requires touching denied paths or broader account/gateway architecture, stop and split a new Sprint.

## Candidate Commit
- Primary: `37044b83e fix(openai): clarify endpoint capability UI`

## Explicitly Deferred
- `ed1b57c59 fix(openai): gate routing by endpoint capability` backend changes, because equivalent backend capability gating appears already present locally and is not part of this frontend Sprint.
- OpenAI OAuth refresh enrichment.
- Admin usage performance/deleted-user history.
- `user_platform_quotas`, DingTalk OAuth, payment/subscription/redeem/channel monitor migrations.
- OpenAI gateway / WS / Responses bridge redesign and response.failed stream handling.
- Any backend/runtime Docker smoke outside this frontend-only Sprint.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run lint:check
```

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-endpoint-capability-ui-s2d-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-openai-endpoint-capability-ui-s2d-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval, implementation, and QA events.

## Stop Rules
- Stop if implementing the selected commit requires touching denied paths.
- Stop if resolving conflicts requires adopting upstream monolithic i18n or broader account/OAuth/gateway architecture.
- Stop if tests fail for reasons requiring backend/schema/API/config changes beyond this frontend capability UI patch.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`
