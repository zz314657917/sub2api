### DONE: upstream-main-compat-s78-ui

# Worker Result

## Task ID

`upstream-main-compat-s78`

## Status

`done`

## Summary

- Replaced the three runtime Stripe SDK imports with the side-effect-free
  `@stripe/stripe-js/pure` entrypoint and isolated Stripe in a dedicated Vite
  vendor chunk.
- Added a focused lazy-loading/chunking test and updated the existing Stripe
  view mock.
- Added the missing English and Chinese OpenAI Mobile RT/AT labels.
- The deferred `174ea22ee` Grok Codex template was not touched because the
  local checkout has no corresponding template.

## Changed Files

- `frontend/src/components/payment/StripePaymentInline.vue`
- `frontend/src/views/user/StripePaymentView.vue`
- `frontend/src/views/user/StripePopupView.vue`
- `frontend/src/views/user/__tests__/StripePaymentView.spec.ts`
- `frontend/src/views/user/__tests__/stripeLazyLoading.spec.ts`
- `frontend/vite.config.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`

## Commands Run

```text
npm.cmd run test:run -- src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/stripeLazyLoading.spec.ts -> PASS (2 files, 5 tests)
npm.cmd run typecheck -> PASS
npm.cmd run build -> PASS
git diff --check -> PASS
```

## Risks

- No authenticated browser payment flow was run; checkout behavior is unchanged
  by inspection and the focused Stripe view test.
- `174ea22ee` remains a separate deferred Grok UI/template slice.

## Knowledge Candidates

- none

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`
