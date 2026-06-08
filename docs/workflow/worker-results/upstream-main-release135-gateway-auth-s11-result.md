### DONE: upstream-main-release135-gateway-auth-s11

## Summary

- Created branch `codex/upstream-main-release135-gateway-auth-s11` from local `main@87424d0c6`.
- Reviewed upstream release `v0.1.135=8c782bcc8` and upstream `main=acbcb50de`.
- Ported the approved gateway/auth/session hardening candidates without directly merging `upstream/main`.
- Kept changes inside approved backend and workflow paths.

## Candidate Results

- `154e0ed6c`: `CHERRY_PICKED` as `094e8b438`. Non-streaming JSON response paths reset `Content-Type` to `application/json`.
- `1a86c6ce1`: `CHERRY_PICKED` as `15f01494f`. API Key exclusive group authorization is enforced at request time.
- `87dd5f5d7`: `CHERRY_PICKED` as `0b595254f`. Mismatched `previous_response_id` is stripped after OpenAI group switching.
- `9a0e43980`: `CHERRY_PICKED` as `0fcd80d8b`. Cross-group mismatch protection moved into the effective WSv2 path.
- `a43629635`: `CHERRY_PICKED` as `064f35021`. OpenAI sticky session group membership is validated before reuse.
- `217f85999`: `CHERRY_PICKED` as `9d69c1c09`. OpenAI `/responses` transport errors are converted into failover errors, and persistent transport faults can temporarily unschedule accounts.

## Local Adaptations

- `541fe39c1`: Adapted transport error handling to local scheduling semantics. The local implementation uses `AccountRepository.SetTempUnschedulable` as the source of truth instead of upstream-only in-memory scheduler blocking.
- `d1812704c`: Added `ValidateFunctionCallOutputContextBytes` and test coverage for handler-side raw tool continuation validation required by the WS previous response stripping path.

## Deferred / Skipped

- `af19d4432`: `DEFERRED`. Proxy expiry/fallback feature requires Ent schema, migrations, frontend/API contract changes, and is outside this Sprint.
- Tag-after changes such as `d251487da`: `DEFERRED`. Prompt cache key propagation is a possible next backend-only Sprint candidate.
- README/assets/skills/version-only changes: `SKIPPED`. These paths are outside the approved backend hardening scope.

## Commits

- `094e8b438` fix: force Content-Type to application/json on non-streaming responses
- `15f01494f` fix: enforce exclusive group access for api keys
- `0b595254f` fix(openai): 切组后剥离失配的 previous_response_id，修复跨组会话鉴权失败
- `0fcd80d8b` fix(openai): 跨组会话失配保护移到生效的 WSv2 路径并补测
- `064f35021` fix: validate OpenAI sticky session groups
- `9d69c1c09` fix(openai): /responses 传输层错误转 failover + 持久故障临时摘除账号
- `541fe39c1` fix(openai): adapt release135 transport and sticky tests
- `d1812704c` fix(openai): add raw tool continuation validation

## Changed Files

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/repository/api_key_repo.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/api_key_auth_cache.go`
- `backend/internal/service/api_key_auth_cache_impl.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/channel_service_test.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_tool_continuation.go`
- `backend/internal/service/openai_tool_continuation_test.go`
- `backend/internal/service/openai_upstream_transport_error.go`
- `backend/internal/service/openai_upstream_transport_error_handle_test.go`
- `backend/internal/service/openai_upstream_transport_error_test.go`
- `backend/internal/service/openai_ws_state_store.go`
- `docs/workflow/tasks/upstream-main-release135-gateway-auth-s11.md`
- `docs/workflow/worker-results/upstream-main-release135-gateway-auth-s11-result.md`
- `docs/workflow/qa-reports/upstream-main-release135-gateway-auth-s11-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check main..HEAD`
- denied path audit against `main..HEAD`
- `go test ./internal/service -run "OpenAI|Gateway|Responses|ChatCompletions|Sticky|Previous|Transport|Failover|ContentType" -count=1`
- `go test ./internal/server/middleware -run "APIKey|Group|Exclusive|Allowed" -count=1`
- `go test ./internal/repository -run "APIKey|AllowedGroups" -count=1`
- `go test ./internal/service ./internal/server/middleware ./internal/repository -count=1`
- `go test ./internal/handler ./internal/server -run "OpenAI|Gateway|APIKey|Contract" -count=1`

## Notes

- No `frontend/`, `backend/ent/`, `backend/migrations/`, `skills/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, or `docs/workflow/spec.md` changes were made.
- Local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, and workflow docs were preserved.
