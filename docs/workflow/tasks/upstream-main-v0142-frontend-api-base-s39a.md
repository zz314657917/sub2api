---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 01:29 +08:00
---

# Task Contract

## Task ID
upstream-main-v0142-frontend-api-base-s39a

## Role
Codex acts as Planner and Final Evaluator. Implementation may be done by Codex directly only after this contract is approved. QA may be run by Codex directly after implementation.

## Goal
Port the safe core of upstream `2a58a57a7 fix(frontend): use configured API base for direct requests` so direct frontend `fetch` / WebSocket calls use the configured API base instead of assuming same-origin backend paths, while excluding dirty `SettingsView.vue` callback suggestion changes.

## Success Criteria
- Add a frontend API URL helper that derives API and gateway URLs from `VITE_API_BASE_URL` / current origin semantics.
- Update direct request call sites in clean files to use the helper:
  - admin ops QPS WebSocket
  - API client refresh call
  - setup API and setup restart polling
  - account test streaming calls
  - key usage direct gateway request
  - custom page markdown/image fetches
  - Stripe popup payment-order polling
- Do not modify `frontend/src/views/admin/SettingsView.vue` in this Sprint; OAuth callback URL suggestions remain deferred because that file is already dirty.
- Do not port `8c2d9b9a1` in this Sprint; removing `gpt-5.3-codex` from default model lists is a product policy decision and not a safe frontend API-base fix.
- Do not touch backend, migrations, Ent, deployment files, knowledge files, or unrelated dirty frontend files.

## Context
- Repo: `F:/mcplugins/sub2api`
- Base planning Sprint: `upstream-main-v0142-merge-plan-s35`
- Previous completed Sprint: `upstream-main-v0142-account-repo-count-s38a`
- Upstream release: `v0.1.142` / `60da9ba17`
- Original S39 candidates: `2a58a57a7`, `8c2d9b9a1`.
- Current precheck:
  - `2a58a57a7` touches 13 frontend files.
  - All `2a58a57a7` paths except `frontend/src/views/admin/SettingsView.vue` are currently clean.
  - `frontend/src/views/admin/SettingsView.vue` is already dirty in the current worktree and must not be edited by this Sprint.
  - `8c2d9b9a1` touches backend OpenAI defaults and frontend model whitelist/use-key UI; it changes visible model defaults and needs product confirmation before porting.

## Allowed Paths
- `frontend/src/api/admin/ops.ts`
- `frontend/src/api/client.ts`
- `frontend/src/api/setup.ts`
- `frontend/src/api/url.ts`
- `frontend/src/components/account/AccountTestModal.vue`
- `frontend/src/components/admin/account/AccountTestModal.vue`
- `frontend/src/i18n/locales/en/admin/settings.ts`
- `frontend/src/i18n/locales/zh/admin/settings.ts`
- `frontend/src/views/KeyUsageView.vue`
- `frontend/src/views/setup/SetupWizardView.vue`
- `frontend/src/views/user/CustomPageView.vue`
- `frontend/src/views/user/StripePopupView.vue`
- `docs/workflow/tasks/upstream-main-v0142-frontend-api-base-s39a.md`
- `docs/workflow/worker-results/upstream-main-v0142-frontend-api-base-s39a-result.md`
- `docs/workflow/qa-reports/upstream-main-v0142-frontend-api-base-s39a-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/composables/useModelWhitelist.ts`
- `backend/internal/pkg/openai/constants.go`
- `backend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `knowledge/**`
- `assets/**`
- `README*`
- `.github/**`
- Any unlisted dirty file.

## Constraints
- Do not merge/rebase `v0.1.142` or `upstream/main`.
- Do not cherry-pick the whole S39 candidate set.
- Keep S39a frontend-only and limited to direct-request API-base behavior.
- Preserve existing same-origin behavior when `VITE_API_BASE_URL` is unset.
- If implementation requires editing `SettingsView.vue` or model default lists, stop and mark that part deferred.
- Do not stage existing dirty files outside allowed paths.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/PaymentView.spec.ts"
git diff --check -- frontend/src/api/admin/ops.ts frontend/src/api/client.ts frontend/src/api/setup.ts frontend/src/api/url.ts frontend/src/components/account/AccountTestModal.vue frontend/src/components/admin/account/AccountTestModal.vue frontend/src/i18n/locales/en/admin/settings.ts frontend/src/i18n/locales/zh/admin/settings.ts frontend/src/views/KeyUsageView.vue frontend/src/views/setup/SetupWizardView.vue frontend/src/views/user/CustomPageView.vue frontend/src/views/user/StripePopupView.vue docs/workflow/tasks/upstream-main-v0142-frontend-api-base-s39a.md docs/workflow/worker-results/upstream-main-v0142-frontend-api-base-s39a-result.md docs/workflow/qa-reports/upstream-main-v0142-frontend-api-base-s39a-qa.md docs/workflow/status.md docs/workflow/main-log.md
git diff --cached --name-only | rg "^(frontend/src/views/admin/SettingsView.vue|frontend/src/components/keys/UseKeyModal.vue|frontend/src/composables/useModelWhitelist.ts|backend/|deploy/|knowledge/|assets/|README|README_|\.github/)" || echo NO_DENIED_PATHS
```

## Output
- Frontend code diff in allowed direct-request paths only.
- Worker result: `docs/workflow/worker-results/upstream-main-v0142-frontend-api-base-s39a-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-v0142-frontend-api-base-s39a-qa.md`
- Updated `docs/workflow/status.md` and appended `docs/workflow/main-log.md`

## Stop Rules
- Stop if `SettingsView.vue` must be edited to complete the safe direct-request behavior.
- Stop if the model default removal commit becomes necessary; split a product-policy Sprint instead.
- Stop if frontend typecheck failures come from unrelated dirty files and cannot be isolated.
- Stop if implementation requires backend, deploy, migration, knowledge, or generated changes.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0.05`

## Review Result
- Reviewed at: 2026-07-03 01:29 +08:00.
- Verdict: approved.
- Reason: required P/G/E contract fields are present; allowed paths are limited to currently clean frontend direct-request files and workflow artifacts; denied paths explicitly protect dirty `SettingsView.vue`, model default policy files, backend, deploy, knowledge, and unrelated dirty files.

## Contract Adaptation
- Updated at: 2026-07-03 01:34 +08:00.
- Reason: local i18n files are split under `frontend/src/i18n/locales/{en,zh}/admin/settings.ts`, unlike the upstream aggregate `frontend/src/i18n/locales/{en,zh}.ts`; the implementation and acceptance paths use the local split files.
