# Task Contract

- Task ID: `upstream-v0173-selective-fixes-s207`
- Role: Generator and Evaluator (direct Codex implementation; no worker is delegated).
- Goal: Behaviorally adapt the bounded `v0.1.173` fixes `b6eb6c1ef`, `cbc2a3dd4`, `2526a0422`, and `769f1c2de`, plus the independently applicable Antigravity fallback-model correctness from `6e34fb09c`, onto the local divergent topology without merging the release or its large prerequisite feature chains.

## Success Criteria

- Gemini native and Claude-compatible forwarding bill by the maximum number of valid inline image parts observed in one upstream payload, reset observation per forwarding attempt, preserve the existing model-name fallback, and also recognize the mapped upstream model.
- Gemini API-key pool-mode 429 responses do not write account-level rate-limit state unless enabled custom error codes match; non-pool, OAuth, 401/403/529, temporary-unscheduled, and custom-code behavior remains intact across messages and chat-completions paths.
- A missing Web Search setting is treated as a normal disabled empty configuration and cached with the normal TTL; reopening `BaseDialog` resets only its body scroll position.
- Grok OAuth external flows return `GROK_OAUTH_CLIENT_NOT_CONFIGURED` instead of dereferencing a nil client; ExchangeCode validation does not consume a valid session before this configuration check.
- Antigravity native Gemini fallback reports the actual fallback model in `ForwardResult.UpstreamModel`. The response-observer hot-path portion of `6e34fb09c` is excluded because local `main` does not contain prerequisite `db0bff82c`; S207 must not introduce its 76-file schema, migration, usage-audit, or admin-frontend chain.
- Focused tagged/default Go tests, complete service regression, server compile, focused Vitest, frontend typecheck/build, formatting, scope, provenance and Git-integrity gates pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/upstream-v0173-s207`
- Frozen base: `ebc3438e4c89c9ad588fa7be707fbb196af03c28`
- Upstream tag: `v0.1.173` annotated tag object `9e2a27ad39201a14074982bae331c4610161586a`, release commit `29009f0b2ea14edf3b11ae2564fb617ff91a03b4`.
- Upstream sources: `b6eb6c1ef`, `cbc2a3dd4`, `2526a0422`, `769f1c2de`, and `6e34fb09c`; prerequisite audit commit `db0bff82c` is intentionally denied.
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`, the five upstream patches, and the related local service/test files below.

## Allowed Paths

- `backend/internal/service/gemini_image_output_accounting.go`
- `backend/internal/service/gemini_image_output_accounting_test.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_error_policy_test.go`
- `backend/internal/service/websearch_config.go`
- `backend/internal/service/websearch_config_test.go`
- `backend/internal/service/grok_oauth_service.go`
- `backend/internal/service/grok_oauth_service_test.go`
- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/antigravity_gateway_service_test.go`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/__tests__/BaseDialog.spec.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/tasks/upstream-v0173-selective-fixes-s207.md`
- `docs/workflow/worker-results/upstream-v0173-selective-fixes-s207-result.md`
- `docs/workflow/qa-reports/upstream-v0173-selective-fixes-s207-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`
- `knowledge/tasks/timeline.md`

## Denied Paths

- `backend/ent/**`, `backend/migrations/**`, `backend/internal/repository/**`, `backend/internal/handler/**`, `backend/go.mod`, `backend/go.sum`, and every frontend path except the two `BaseDialog` paths above.
- Upstream `db0bff82c`, migration `194`/`195`, migration `220`, Grok full integration, Channel Monitor V2, deployment, containers, provider credentials, shared database/Redis, production traffic, remote push, and unrelated release changes.

## Constraints

- Port behavior into the local monolithic Gemini/Antigravity topology; do not replace local files with upstream split-file versions.
- Preserve local billing-model candidates, image-size tiers, failover lifecycle, same-account retry, custom error policy precedence, account cooldown rules, and client-disconnect accounting.
- Image counting accepts only valid JSON `inlineData`/`inline_data` parts with non-empty data and an `image/*` MIME type. Use maximum-per-payload observation to avoid duplicate billing on cumulative streams.
- Keep the `6e34fb09c` adaptation executable and useful without adding dead observer code or silently pulling its schema-changing prerequisite.
- No push, deployment, container update, provider call, schema or migration operation.

## Acceptance Commands

```powershell
git rev-parse HEAD
go test -tags=unit ./internal/service -run 'Test(CountGeminiInlineImageOutputs|ObserveGeminiImageOutputs|BeginGeminiImageOutputObservation|ResolveGeminiImageCount|HandleNativeNonStreamingResponse_FeedsImageCounter|HandleGeminiUpstreamError_PoolMode429|GrokOAuthService.*MissingClient|AntigravityGatewayService_ForwardGemini_FallbackReportsActualUpstreamModel|LoadWebSearchConfigFromDB_MissingSetting)' -count=1
go test ./internal/service -run 'TestAntigravityGatewayService_ForwardGemini|TestGeminiMessagesCompatService|TestWebSearch' -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=0
gofmt -w <changed Go files>
corepack.cmd pnpm --dir frontend exec vitest run src/components/common/__tests__/BaseDialog.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend run build
git diff --check
git diff --name-only ebc3438e4c89c9ad588fa7be707fbb196af03c28...HEAD
git diff --name-only --diff-filter=U
rg -n '^(<<<<<<<|=======|>>>>>>>)' <changed files>
foreach ($commit in @('b6eb6c1ef','cbc2a3dd4','2526a0422','769f1c2de','6e34fb09c')) { git merge-base --is-ancestor $commit 29009f0b2ea14edf3b11ae2564fb617ff91a03b4; if ($LASTEXITCODE -ne 0) { throw "$commit is not contained in v0.1.173" } }
```

## Output

- Write `docs/workflow/worker-results/upstream-v0173-selective-fixes-s207-result.md` with a first-line `### DONE`, `### BLOCKED`, or `### FAILED` verdict.
- Write `docs/workflow/qa-reports/upstream-v0173-selective-fixes-s207-qa.md` with an evidence-backed `PASS`, `FAIL`, or `BLOCKED` verdict.
- Record P/G/E transitions in `docs/workflow/main-log.md`; update handoff files only after final evidence exists.
- Commit only allowed files to the S207 branch. After final PASS, fast-forward local `main` only; do not push or deploy.

## Stop Rules

- Stop if any slice requires schema/migration, repository persistence changes, production configuration, provider credentials, or a frontend path outside `BaseDialog`.
- Stop if image observation can leak across failover attempts, overcount cumulative streams, or bypass the existing billing resolver.
- Stop if pool-mode protection weakens custom error code precedence, OAuth cooldowns, or 401/403/529 handling.
- Stop if the Antigravity fallback correction cannot be proved without importing `db0bff82c`.
