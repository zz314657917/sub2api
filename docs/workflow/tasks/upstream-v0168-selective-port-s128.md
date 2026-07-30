---
task_id: upstream-v0168-selective-port-s128
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Selectively adapt four isolated `v0.1.168` compatibility fixes to the current
local architecture: preserve an OAuth-mimic migrated system cache breakpoint,
emit Anthropic-compatible synthesized message IDs/schema, retain GPT-5.6
`max` reasoning effort in the Messages bridge, and add the Claude Sonnet 5
status alias. Do not merge or cherry-pick the upstream tag.

## Success Criteria

- When non-Claude-Code OAuth mimicry moves an array-form client `system`
  prompt into the leading synthetic user message, preserve the final original
  text block's complete `cache_control` object. Requests without that object
  remain uncached.
- Intercept, Gemini compatibility, and Antigravity synthesized Anthropic
  responses use `msg_01` plus 22 Base62 characters. The changed mock response
  fields match the intended Anthropic schema without changing routing,
  scheduling, or accounting behavior.
- An Anthropic request with `output_config.effort=max` reaches GPT-5.6
  Responses upstream as `max`; non-GPT-5.6 behavior continues to use the
  existing `max` to `xhigh` compatibility mapping.
- A `claude-sonnet-5` account rate-limit scope renders as `CSon5`, while the
  existing `claude-opus-5` alias remains intact.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-v0168-selective-port-s128`
- Source commits: `1631b19f8`, `248236ce6`, `46dba1939`, `32618e71e`.
- Equivalent local ports for `83b368553`, `1c26dc7ad`, and `d8ae153ae` are
  already present and are explicitly out of scope.

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_prompt_test.go`
- `backend/internal/service/gateway_anthropic_apikey_passthrough_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_intercept_test.go`
- `backend/internal/handler/gateway_handler_warmup_intercept_unit_test.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_chat_completions_compat_service_test.go`
- `backend/internal/pkg/antigravity/response_transformer.go`
- `backend/internal/pkg/antigravity/stream_transformer.go`
- `backend/internal/pkg/antigravity/response_transformer_test.go`
- `backend/internal/service/openai_compat_model.go`
- `backend/internal/service/openai_compat_model_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_usage_test.go`
- `frontend/src/components/account/AccountStatusIndicator.vue`
- `frontend/src/components/account/__tests__/AccountStatusIndicator.spec.ts`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `backend/internal/setup/**`
- `backend/internal/securityaudit/**`
- `backend/internal/handler/openai_codex_models_handler.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/config/**`
- `frontend/src/views/**`
- `frontend/src/components/modelPlaza/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `backend/cmd/server/VERSION`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- Any path not listed under Allowed Paths

## Constraints

- Adapt behavior to local file boundaries; do not force the upstream split-file
  structure into this repository.
- Preserve `cache_control` only on the synthetic system-instruction message;
  do not add new cache breakpoints or weaken the existing cache-control limit.
- Do not globally change the `max` to `xhigh` mapping. The exception must be
  bounded to the final GPT-5.6 upstream model.
- Keep generated identifiers cryptographically random where the current code
  already uses `crypto/rand`; fallback output must still preserve the `msg_01`
  prefix.
- Do not modify model routing, model substitution, billing, persistence,
  account selection, configuration, version metadata, deployment, containers,
  or the primary worktree.

## Acceptance Commands

Run from `backend`:

```powershell
go test ./internal/service -run "Test(RewriteSystemForNonClaudeCode|GatewayService_AnthropicOAuthMimic|OpenAICompat)" -count=1
go test ./internal/handler -run "Test(SendMockInterceptResponse|DetectInterceptType|GatewayHandlerMessages_InterceptWarmup)" -count=1
go test ./internal/pkg/antigravity -count=1
go test ./... -run "^$"
```

Run from the worktree root:

```powershell
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/AccountStatusIndicator.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
gofmt -d <changed Go files>
git diff --check
git diff --name-only HEAD
git ls-files -u
```

## Output

- A narrow implementation, focused regressions, and
  `docs/workflow/qa-reports/upstream-v0168-selective-port-s128-qa.md`.
- Codex owns implementation and QA; no worker is invoked because the user did
  not authorize a sub-agent.

## Stop Rules

- Stop if the implementation needs a migration, configuration change, model
  routing/billing change, Passkey/WebAuthn work, deployment change, or any
  denied path.
- Stop if preserving a cache breakpoint changes existing cache-control count
  enforcement or if GPT-5.6 requires changing the global effort mapping.
- Stop if an out-of-scope path appears in the diff or focused tests expose a
  pre-existing unrelated failure that cannot be isolated.

## Contract Review

`PASS`: the local OAuth mimic, synthetic response, Messages bridge, and status
alias seams cover all four behavior changes without requiring a migration,
configuration, routing, billing, deployment, or container change. The existing
model-aware `normalizeOpenAIReasoningEffortForModel` helper supports the
GPT-5.6-only exception while preserving the global compatibility mapping.

## Implementation Result

- The OAuth mimic now carries the final original array-system
  `cache_control` object onto its synthetic instruction message and leaves
  requests without a caller breakpoint unchanged.
- Intercept, Gemini Messages/Chat, and Antigravity fallback responses now use
  the `msg_01` Anthropic identifier shape; intercept payloads also carry the
  compatible stop/cache usage fields.
- `ForwardAsAnthropic` resolves the final upstream model before restoring
  `max` only for GPT-5.6, retaining `xhigh` for other OpenAI models.
- Account status rendering recognizes `claude-sonnet-5` as `CSon5` without
  changing the existing `COpus5` alias.

## QA Result

`PASS / source-only`: focused Go, Antigravity, frontend, compile, formatting,
diff, unmerged-index, and allowed-path checks passed. Real upstream protocol
calls, commit, merge, push, deployment, and container refresh remain out of
scope.
