### PASS: upstream-main-v0138-followup-safe-patches-s21

## Changed Files

- `backend/internal/service/auth_email_binding.go`
- `backend/internal/service/auth_service_email_bind_test.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `frontend/src/api/admin/usage.ts`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts`
- `frontend/src/components/admin/usage/UsageStatsCards.vue`
- `frontend/src/components/admin/usage/__tests__/UsageStatsCards.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/usage.ts`

## Summary

- Spark /responses requests now strip client-provided `image_generation` tools before local image permission checks and before HTTP/WS forwarding.
- Email identity binding and bind-code sending enforce `registration_email_suffix_whitelist`.
- OpenAI weekly reset now opens an explicit confirmation dialog before consuming a reset credit.
- Admin usage stat cards show total cache tokens with creation/read breakdown.

## Commands Run

- `go test ./internal/service -run "TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey|TestStripCodexSparkImageGenerationToolFromRawPayload|TestAuthServiceBindEmailIdentity_.*Suffix|TestAuthServiceSendEmailIdentityBindCode_.*Suffix" -count=1`
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/admin/usage/__tests__/UsageStatsCards.spec.ts"`
- `git diff --check`
- `git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|backend/internal/service/payment_|backend/internal/service/subscription_|backend/internal/service/gateway_billing_|deploy/|README|README_|assets/partners/|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment|knowledge/05-current-focus.md)" || echo NO_DENIED_PATHS`

## Risks

- No live OpenAI OAuth upstream or browser UI smoke was run; validation is targeted unit/component coverage.
