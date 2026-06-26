---
task_id: upstream-main-v0138-followup-safe-patches-s21
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-26 00:00 +08:00
---

# Task Contract: upstream v0.1.138 follow-up safe patches S21

## Goal

Port the locally useful, low-risk follow-up patches found while reviewing upstream `v0.1.138` and post-release `upstream/main`, without wholesale merging upstream or changing local product strategy.

## Success Criteria

- `gpt-5.3-codex-spark` requests strip client-provided `image_generation` tools before image-generation permission checks and before forwarding to HTTP or WS upstream paths.
- OpenAI OAuth weekly limit reset in the admin account table requires an explicit confirmation dialog before consuming a reset credit.
- Admin usage stats display cache token totals with creation/read breakdown using fields already returned by the backend.
- Authenticated email binding and email-bind-code sending enforce the configured registration email suffix whitelist.
- Targeted backend tests, targeted frontend Vitest, denied-path audit, and `git diff --check` pass.

## Allowed Paths

- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_ingress_test.go`
- `backend/internal/service/auth_email_binding.go`
- `backend/internal/service/auth_service_email_bind_test.go`
- `frontend/src/api/admin/usage.ts`
- `frontend/src/components/admin/usage/UsageStatsCards.vue`
- `frontend/src/components/admin/usage/__tests__/UsageStatsCards.spec.ts`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `docs/workflow/tasks/upstream-main-v0138-followup-safe-patches-s21.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/upstream-main-v0138-followup-safe-patches-s21-result.md`
- `docs/workflow/qa-reports/upstream-main-v0138-followup-safe-patches-s21-qa.md`

## Denied Paths

- Ent schema/generated files and migrations.
- Payment fulfillment, affiliate rebate, subscription package logic.
- Scheduler strategy/config including prefer-soonest-reset.
- Claude mimicry `cch` signing removal and gateway forwarding settings.
- CI, deploy, README, VERSION, sponsor assets, public pages, Studio/Canvas, model market, payment pages, and unrelated knowledge files.

## Constraints

- Do not cherry-pick upstream commits blindly; adapt to local image permission gates and local UI structure.
- Keep the Spark tool-strip behavior independent of the Codex image-generation bridge flag, because the offending tool can be client-provided.
- Preserve non-Spark image-generation behavior.
- Do not stage or commit unless explicitly requested.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestApplyCodexOAuthTransform_.*ImageGenerationTool.*Spark|TestOpenAIGatewayService_Forward_StripsImageGenerationToolForSparkAPIKey|TestStripCodexSparkImageGenerationToolFromRawPayload|TestAuthServiceBindEmailIdentity_.*Suffix|TestAuthServiceSendEmailIdentityBindCode_.*Suffix" -count=1

cd F:/mcplugins/sub2api
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/OpenAIQuotaResetCell.spec.ts src/components/admin/usage/__tests__/UsageStatsCards.spec.ts"
git diff --check
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|backend/internal/service/payment_|backend/internal/service/subscription_|backend/internal/service/gateway_billing_|deploy/|README|README_|assets/partners/|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment|knowledge/05-current-focus.md)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in the allowed paths only.
- Worker-style result, QA report, and final report with `Findings / Executed Checks / Unverified Risks / Recommendation`.

## Stop Rules

- Stop if Spark tool stripping requires altering global image-generation routing semantics.
- Stop if email suffix enforcement reveals a product-policy ambiguity for already-bound users.
- Stop if frontend tests require broad component rewrites beyond the two touched UI surfaces.
- Stop if any patch requires denied paths.
