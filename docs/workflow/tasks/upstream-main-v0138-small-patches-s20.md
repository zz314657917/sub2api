---
task_id: upstream-main-v0138-small-patches-s20
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-23 00:00 +08:00
---

# Task Contract: upstream v0.1.138 small patches S20

## Goal

Port the low-risk, locally relevant upstream `v0.1.138` fixes without wholesale merging `upstream/main` or touching local product customizations.

## Success Criteria

- Gemini tool schema cleanup removes upstream-unsupported `$defs` / `definitions` and normalizes nullable type arrays.
- OpenAI images Responses path recognizes `response.incomplete`, records no-output diagnostics, and preserves retry/failover semantics.
- Vertex Anthropic service-account path filters unsupported `anthropic-beta` tokens before forwarding.
- Claude Code validator accepts any billing-block `cc_entrypoint=` value while still requiring Claude CLI UA and entrypoint presence.
- GLM raw Chat Completions requests normalize OpenAI-style reasoning effort to GLM native `high` / `max`.
- OpenAI API-key chat-only accounts record `/v1/chat/completions` as the upstream endpoint in all OpenAI usage/ops recording paths.
- Admin promo-code edit can clear an existing expiry by sending `expires_at: 0`.
- Targeted backend tests and `git diff --check` pass.

## Allowed Paths

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_images_incomplete_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_anthropic_vertex_beta_filter_test.go`
- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `backend/internal/service/gateway_request.go`
- `backend/internal/service/gateway_request_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_chat_completions_raw_test.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/admin/promo_handler.go`
- `backend/internal/service/promo_service.go`
- `docs/workflow/tasks/upstream-main-v0138-small-patches-s20.md`
- `docs/workflow/status.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- Ent schema/generated files and migrations.
- Payment fulfillment, affiliate rebate, subscription package logic.
- Scheduler strategy/config except code already touched by Vertex builder inside `gateway_service.go`.
- Claude mimicry `cch` signing removal.
- Frontend UI, i18n, public pages, Studio/Canvas, model market, CI, deploy, README, VERSION, sponsor assets.

## Constraints

- Do not cherry-pick commits blindly; adapt patches to local divergences.
- Keep changes surgical and traceable to upstream fixes.
- Do not alter default scheduling, payment, auth policy, or product configuration behavior.
- Do not stage or commit unless explicitly requested.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestCleanToolSchema|TestExtractImagesUpstreamError|TestSummarizeNoOutputBody|TestImagesOAuthNonStreaming_CompletedNoImageTriggersSameAccountRetry|TestVertexBetaFilter|TestFilterVertexBetaTokens|TestClaudeCodeValidator|TestNormalizeGLMOpenAIReasoningEffort|TestForwardAsRawChatCompletions_NormalizesGLMReasoningEffortForUpstream" -count=1
go test ./internal/handler -run "Test.*OpenAI|Test.*ChatCompletions|Test.*Responses|Test.*Messages" -count=1
cd F:/mcplugins/sub2api
git diff --check
```

## Output

- Code diff in the allowed paths only.
- Final report with `Findings / Executed Checks / Unverified Risks / Recommendation`.

## Stop Rules

- Stop and report if an upstream patch requires denied paths.
- Stop and report if local code has an incompatible implementation requiring product or policy decisions.
- Stop and report if tests expose failures outside this task that cannot be attributed to the S20 diff.
