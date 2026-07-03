### DONE: upstream-main-v0142-frontend-api-base-s39a

## Summary
- Ported the safe direct-request portion of upstream `2a58a57a7 fix(frontend): use configured API base for direct requests`.
- Added `frontend/src/api/url.ts` with `getAPIBaseURL`, `buildApiUrl`, and `buildGatewayUrl`.
- Updated direct `fetch` / WebSocket call sites to respect the configured API base:
  - admin ops QPS WebSocket
  - API client refresh call
  - setup client and restart polling
  - user/admin account test streaming calls
  - key usage direct gateway request
  - custom page markdown/image requests
  - Stripe popup payment-order polling
- Adapted upstream i18n edits to local split files under `frontend/src/i18n/locales/{en,zh}/admin/settings.ts`.
- Deferred `SettingsView.vue` callback suggestion edits because that file is already dirty.
- Deferred `8c2d9b9a1` because removing `gpt-5.3-codex` from default model lists is a product policy decision.

## Changed Files
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
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/PaymentView.spec.ts"`
  - Result: PASS, 2 files / 22 tests.
- `git diff --check -- <S39a allowed paths>`
  - Result: PASS, only LF/CRLF warnings for workflow docs.
- staged denied-path audit
  - Result: `NO_DENIED_PATHS` because no files were staged.

## Contract Compliance
- Did not edit `frontend/src/views/admin/SettingsView.vue`.
- Did not edit `frontend/src/components/keys/UseKeyModal.vue`, `frontend/src/composables/useModelWhitelist.ts`, or `backend/internal/pkg/openai/constants.go`.
- Did not touch backend, deploy, knowledge, migrations, Ent, or generated files.

## Risks
- OAuth callback URL suggestions in `SettingsView.vue` are still not ported; they should be handled later after the existing Settings dirty work is closed or isolated.
- `gpt-5.3-codex` default model removal remains deferred pending product confirmation.
