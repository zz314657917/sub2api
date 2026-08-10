### DONE: upstream-v0173-selective-fixes-s207

# Worker Result

## Task ID
`upstream-v0173-selective-fixes-s207`

## Status
`done`

## Summary

- Behaviorally ported the five approved `v0.1.173` fixes onto the local divergent topology.
- Gemini native and Claude-compatible paths now bill valid returned inline images, including custom and mapped model names, while preserving the one-image model fallback.
- Gemini pool-mode 429 and temporary-unscheduled policy decisions no longer fall through into account-level rate limiting across Messages, Native, and Chat Completions paths.
- Missing Web Search settings return a normally cached disabled config; reopening `BaseDialog` resets only its body scroll.
- Grok OAuth returns `GROK_OAUTH_CLIENT_NOT_CONFIGURED` without consuming a valid Exchange session, and Antigravity fallback reports the model actually forwarded.
- The schema/migration/admin audit prerequisite `db0bff82c` remains excluded.

## Changed Files

- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/antigravity_gateway_service_test.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_image_output_accounting.go`
- `backend/internal/service/gemini_image_output_accounting_test.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/grok_oauth_service.go`
- `backend/internal/service/upstream_v0173_selective_fixes_test.go`
- `backend/internal/service/websearch_config.go`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/__tests__/BaseDialog.spec.ts`
- S207 workflow and handoff evidence files allowed by the contract.

## Commands Run

```text
go test ./internal/service -run '<S207 focused regex>' -count=1 -> PASS
go test ./internal/service -run 'TestAntigravityGatewayService_ForwardGemini|TestGeminiMessagesCompatService|TestWebSearch' -count=1 -> PASS
go test ./internal/service -run 'TestGemini(ChatCompletions|MessagesCompat|Native)_(TempUnscheduled429|PoolMode429)DoesNotSetAccountRateLimit' -count=1 -> PASS
go test ./internal/service -count=1 -> PASS (61.428s latest run)
go test ./cmd/server -run '^$' -count=0 -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/components/common/__tests__/BaseDialog.spec.ts -> PASS (1/1)
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS
git diff --check -> PASS
allowlist / conflict marker / unmerged index checks -> PASS
v0.1.173 provenance for b6eb6c1ef, cbc2a3dd4, 2526a0422, 769f1c2de, 6e34fb09c -> PASS
db0bff82c ancestor exclusion -> PASS
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 61.428s
ok github.com/Wei-Shaw/sub2api/cmd/server 5.506s [no tests to run]
BaseDialog.spec.ts: 1 test passed
vite: 1871 modules transformed, production build completed
```

## Risks

- No real Gemini, Grok, or Antigravity provider call was made.
- No shared database/Redis, container, deployment, remote push, or production traffic was exercised.
- Frontend build retained existing chunk-size, dynamic-import, Browserslist-age, and Node deprecation warnings; none blocked the build and none were introduced by the two BaseDialog files.

## Knowledge Candidates

- None. The behavior and integration boundary are recorded in repository workflow evidence.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.
