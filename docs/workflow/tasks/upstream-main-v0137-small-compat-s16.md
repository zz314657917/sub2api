---
task_id: upstream-main-v0137-small-compat-s16
role: Generator
phase: contract-approved
qa_mode: runtime
owner: codex
upstream_base: 4a5665da5
created_at: 2026-06-17
---

# Task Contract: upstream-main-v0137-small-compat-s16

## Goal

Continue the upstream `v0.1.137` review after S15 by porting only small, independent compatibility fixes that do not require migrations and do not touch local product customizations.

## Success Criteria

- Preserve Studio Bridge / LuoyeAI, payment package, model market, Canvas, tickets, and public page customizations.
- Port or equivalently implement these isolated upstream fixes:
  - `a67b10f46` / `44f579100`: anchor Responses API sticky-session hash to `input`.
  - `b256f9114`: intercept `max_tokens=1` Haiku Claude Code probes for streaming requests too.
  - `b88f8e4c0` / `2ce878892`: make OpenAI `/responses` APIKey probe verify tool-call capability using the mapped upstream model.
  - `56c62c59c` / `9e9e154f5`: include the actual client IP in API Key ACL denial messages.
- Keep S15 safe patches intact.
- Do not change `backend/cmd/server/VERSION` merely to match upstream.

## Allowed Paths

- `backend/internal/service/**`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_intercept_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `docs/workflow/**`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/VERSION`
- Studio Bridge / Canvas / payment / public page product files.
- Production deployment secrets or environment files.

## Constraints

- Prefer local equivalent ports over broad cherry-picks because local gateway parsing differs from upstream raw-range parsing.
- Do not port migration-heavy, OpenAI quota UI, cyber policy, channel monitor jitter, Claude OAuth system prompt block, or image-generation failover chains in this Sprint.
- If a candidate touches local product UX or schema/migrations, document it as skipped for a future Sprint.

## Acceptance Commands

- `cd backend && go test -tags=unit ./internal/service -run "TestParseGatewayRequest_ResponsesInput|TestGenerateSessionHash_ResponsesInputProducesHash|TestDecideResponsesProbeSupportRequiresFunctionCallOn2xx|TestOpenAIResponsesProbePayloadForcesFunctionCall|TestSelectResponsesProbeModelUsesMappedUpstreamModel|TestProbeOpenAIAPIKeyResponsesSupportPersistsToolCapability" -count=1`
- `cd backend && go test -tags=unit ./internal/handler -run "TestDetectInterceptType_MaxTokensOneHaiku|TestSendMockInterceptResponse_MaxTokensOneHaiku" -count=1`
- `cd backend && go test -tags=unit ./internal/server/middleware -run "TestAPIKeyAuthIPRestrictionDoesNotTrustSpoofedForwardHeaders|TestAPIKeyAuthIPRestrictionIncludesClientIPForBlacklistDenial" -count=1`
- `cd backend && go test -tags=unit ./internal/service -run "Test.*Billing|Test.*Thinking|Test.*Reasoning|Test.*Gateway|Test.*OpenAI|Test.*TokenRefresh|Test.*FilterThinking|Test.*ThinkingFilters|Test.*NormalizeChineseLLMThinking|Test.*ApplyThinkingEnabledFallback|Test.*GenerateSessionHash|TestParseGatewayRequest" -count=1`
- `cd backend && go test ./internal/repository -run "Test.*Decompress|Test.*HTTPUpstream" -count=1`
- `cd backend && go test ./internal/pkg/apicompat -count=1`
- `git diff --check`

## Output

- Implementation diff with short notes mapping each upstream commit to `ported`, `equivalent`, or `skipped`.
- Worker/result note at `docs/workflow/worker-results/upstream-main-v0137-small-compat-s16-result.md`.
- QA note at `docs/workflow/qa-reports/upstream-main-v0137-small-compat-s16-qa.md`.

## Stop Rules

- Stop before changing Ent schema, migrations, VERSION, product pages, payment behavior, Studio Bridge flow, or production secrets.
- Stop before broad merge/rebase/reset of `upstream/main`.
