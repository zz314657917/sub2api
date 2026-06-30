### DONE: upstream-main-v0141-model-not-found-s32

# Worker Result

## Task ID
upstream-main-v0141-model-not-found-s32

## Status
done

## Summary
- Ported the local-relevant part of upstream `fcd3bc1272b3e283c172f153db4d75911cd93357`.
- Added shared no-account error classification for gateway handlers.
- Added model availability diagnosis for both Anthropic/Gemini gateway service and OpenAI gateway service.
- Changed unsupported-model no-account failures to `404 model_not_found` only when the account pool is non-empty and no configured account supports the requested model.
- Preserved `503 api_error` for empty pools, lookup failures, compact-account unsupported errors, and transient capacity exhaustion.
- Kept local `ForUser` account-selection calls in OpenAI paths so user-private pool behavior remains unchanged.

## Changed Files
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/no_account_error.go`
- `backend/internal/handler/no_account_error_test.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/service/gateway_model_availability.go`
- `backend/internal/service/gateway_model_availability_test.go`
- `backend/internal/service/openai_gateway_model_availability.go`
- `docs/workflow/tasks/upstream-main-v0141-model-not-found-s32.md`

## Commands Run
```text
gofmt -w backend/internal/handler/no_account_error.go backend/internal/handler/no_account_error_test.go backend/internal/handler/gateway_handler.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/handler/gemini_v1beta_handler.go backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_embeddings.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_images.go backend/internal/service/gateway_model_availability.go backend/internal/service/gateway_model_availability_test.go backend/internal/service/openai_gateway_model_availability.go -> pass
go test -tags=unit ./internal/handler -run "TestClassifyNoAccountError" -count=1 -> pass
go test -tags=unit ./internal/handler -run "TestClassifyNoAccountError|Test.*Gateway|Test.*OpenAI|Test.*Messages|Test.*ChatCompletions|Test.*Images|Test.*Embeddings" -count=1 -> pass
go test -tags=unit ./internal/service -run "TestDiagnoseModelAvailabilityForPlatform" -count=1 -> blocked in main worktree by unrelated proxy/account dirty tree
clean worktree + same S32 patch: go test -tags=unit ./internal/service -run "TestDiagnoseModelAvailabilityForPlatform" -count=1 -> pass
clean worktree + same S32 patch: go test -tags=unit ./internal/handler -run "TestClassifyNoAccountError" -count=1 -> pass
git diff --check -- <S32 allowed paths> -> pass with LF-to-CRLF warnings for docs/workflow/status.md and docs/workflow/main-log.md
```

## Test Output
```text
main worktree handler:
ok  	github.com/Wei-Shaw/sub2api/internal/handler	6.504s

main worktree broader handler:
ok  	github.com/Wei-Shaw/sub2api/internal/handler	0.202s

main worktree service blocked before S32 tests:
internal\service\admin_service_delete_test.go:627:38: cannot use repo (variable of type *proxyRepoStub) as ProxyRepository value in struct literal: *proxyRepoStub does not implement ProxyRepository (missing method CountUserOwnedAccountsByProxyID)

clean worktree service:
ok  	github.com/Wei-Shaw/sub2api/internal/service	12.491s

clean worktree handler:
ok  	github.com/Wei-Shaw/sub2api/internal/handler	10.330s
```

## Risks
- Main worktree `./internal/service` package currently cannot compile because unrelated user-owned-proxy changes added a `ProxyRepository` method while older test stubs are dirty/incomplete. S32 did not modify those paths.
- Verification is unit-level and code-path level; no live gateway request was issued.
- Deterministic 404 depends on account `model_mapping` inspection and deliberately ignores transient state.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None for S32. Main worktree service package has unrelated compile blockers documented above.
