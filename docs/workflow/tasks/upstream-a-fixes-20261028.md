---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-a-fixes-20261028
worker_model: gpt-6.1-sol
base_commit: b3091c1c9a28884cf4b00fe97e87fc2f295b0512
spec_ref: docs/workflow/spec.md
last_verified: 2026-10-28
---

# Task Contract

## Task ID
upstream-a-fixes-20261028

## Role
Controller plans and integrates. Independent gpt-6.1-sol Contract Reviewer approves before implementation. Three disjoint gpt-6.1-sol Developers implement core/pagination/media slices; a separate gpt-6.1-sol QA independently validates the combined candidate.

## Goal
The user authorized starting the first recommended upstream batch A1–A7. Adapt eight pinned upstream commits to the local monolithic architecture without whole-history merge, unrelated feature changes or publishing.

## Context
- Main: F:/mcplugins/sub2api. Candidate: E:/codex-worktrees/sub2api/upstream-a-fixes-20261028.
- Evidence: E:/codex-runtime/sub2api/upstream-a-fixes-20261028.
- Upstream fixed pin f178e5ae32659e8901644a9f2378a0439ffa36ae. Provenance: de908ed6d, 3e18e78ab, 7bce2cfb4, f96528048, 9aa16c8c6, cde6ddfd8, 917dbdc70, 14a5fe305.
- Main-before.json records 12212 original files plus HEAD/index. Nine original dirty backend paths have byte backups under protected-backend/; none are target owners. Candidate initially uses committed HEAD and its dependencies. No copying unrelated dirty prerequisites to conceal failures. Main integration later verifies exact real dirty combination.
- No dependency upgrades. Go executable E:/tools/go1.26.10/go/bin/go.exe, GOROOT=E:/tools/go1.26.10/go, GOTOOLCHAIN=local. The external run-checks.mjs contains the fixed original acceptance commands; preserve its hash through review and QA.
- Global request body default is 256 MiB in current local configuration. A5 adopts upstream's 50 MiB per-part upload bound; do not change global limits, download bounds or existing custom configured limits.
- Runtime means in-process production functions and localhost HTTP/httptest fake upstream smoke, with no real paid providers, databases, containers or service deployments.

## Success Criteria
### A1 core: CN frequency 429
- Clear frequency-only wording plus a usable non-near-exhausted quota snapshot returns control to the existing short429 fallback. Use local cnAccountIsCodingPlan and local 5h/weekly quota keys, not absent upstream monthly/provider architecture.
- Any quota/balance wording, missing usable snapshot, malformed/nonfinite/out-of-range observed percentages, or any present known window >=90% must not establish short-cooldown eligibility. Preserve insufficient-balance recovery and genuine quota/reset behavior. One valid known window suffices when others are absent; all observed known windows must pass.
- Test the public upstream-error handling path's actual scheduling/repository writes and configured short delay, not just marker helpers. Keep unrelated providers and payg behavior unchanged.
### A2 core: DeepSeek cache accounting
- Add prompt_cache_hit_tokens as the last OpenAI usage cache-read fallback. Nested cached_tokens presence, including explicit0, keeps precedence. Reuse existing bounded non-negative integer parsing; invalid/negative/overflow values must not turn into chargeable huge cache tokens.
- Cover JSON Chat and streamed usage in actual existing parsing paths, non-DeepSeek standard fields unchanged. Exercise the real usage recorder using in-memory repositories for one request and duplicate request ID; preserve existing idempotency and prices, without altering billing implementation.
### A3 core: Claude fingerprint
- Sample JavaScript UTF-16 code units at zero-based indices4/7/20, pad missing units with ASCII0, join sampled units before UTF-8 conversion, replace unpaired surrogates as JavaScript does, SHA256(salt+sample+version), first3 hex.
- Keep ASCII outputs unchanged, bound work to the needed input prefix. Tests include CJK, emoji around sampled indices, recombined and unmatched surrogate halves, short/long input and header wire consistency. No change to site billing prices.
### A4 pagination
- Adapt both upstream fixes as one behavior: safe native-int decimal parsing and bounded page/offset so both offset and offset+effectiveLimit stay in int range.
- Offset uses Limit() consistently including invalid/oversized PageSize. Keep existing default20/max1000. Broad ParsePagination invalid numeric values fall back as before; valid huge page is clamped. Strict prompt-audit and Ops token-stat filters reject impossible offset pages with400.
- Local API-key listing already delegates to repository Offset()/Limit(); preserve that chain and do not introduce upstream's absent paginateAPIKeys. Verify actual repository paginateSlice and service mergeBalanceHistoryCodes return empty without panic on extreme pages. The existing dirty admin_service.go is read-only; fixing its shared Offset helper suffices. Cover nativeMaxInt, out-of-range strings, missing/invalid values, PageSize0/negative/above1000 and ordinary paging.
- No unrelated pagination API/response shape or database-query redesign.
### A5 media: upload rejection
- Read at most50 MiB+1 per multipart part and reject any part larger than50 MiB with a clear client400 through the existing parse/handler error flow. Exactly50 MiB works; >20 MiB <=50 MiB is not silently truncated. Preserve file bytes/content-type/model rewriting/masks/transparent-background behavior.
- Do not add absent upstream grok_media.go. Existing file-preserving multipart rewrite must not newly truncate; total request body guard remains active. No input image content logging.
### A6 media: Gemini output counting
- Count nonempty string data for inlineData/inline_data and nonempty string fileUri/file_uri for fileData/file_data with image MIME only; one part counts at most once.
- Preserve max-of-cumulative-SSE frames, per-forward reset and original/mapped model fallback. Empty/invalid/nonimage metadata does not count. Exercise observed URI-image count through the existing accounting path, not just an isolated predicate.
- No group image_count_billing switch, schema, full multi-image charging feature or pricing-policy change.
### A7 media: malformed Antigravity nonstream
- Transformer returns a typed/sentinel error on MALFORMED_FUNCTION_CALL in direct and wrapped responses instead of successful end_turn.
- Claude SSE-to-nonstream aggregation catches the finish reason before choosing last-with-parts, including a trailing failure frame without parts. Map to502 with explicit malformed_function_call through the local established error writer. Do not silently succeed, charge as success or retry a request after visible output.
- Preserve valid tool-call/normal responses and existing true-stream behavior; no broader error-policy changes. Test production collector with fake streamed response and actual response writer.
### Combined integration
- All new tests named TestUpstreamAFixes*, run every new declaration default/unit three times with no skip or omissions. Existing assertions remain intact; no stubs/overlays to hide product failures.
- Independent QA repeats original commands, compares any baseline failures by exact name and assertion, and independently probes required boundaries. Baseline failures require independent reproduction, never a new waiver by Developer.
- Before main integration freeze exact source hashes and prerequisites. Controller confirms current target bytes still equal pinned HEAD, original HEAD/index unchanged and no unexpected file drift; copy only approved source/test paths after independent PASS. Then repeat original commands and new tests against main's exact existing dependencies/dirty source. Non-target dirty bytes stay untouched.

## Allowed Paths
Core Developer:
- backend/internal/service/ratelimit_cn_providers.go
- backend/internal/service/ratelimit_cn_frequency_429_20261028_test.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_usage_deepseek_cache_20261028_test.go
- backend/internal/service/gateway_billing_block.go
- backend/internal/service/gateway_billing_utf16_20261028_test.go

Pagination Developer:
- backend/internal/pkg/pagination/pagination.go
- backend/internal/pkg/pagination/upstream_a_fixes_20261028_test.go
- backend/internal/pkg/response/response.go
- backend/internal/pkg/response/upstream_a_fixes_20261028_test.go
- backend/internal/handler/admin/ops_dashboard_handler.go
- backend/internal/handler/admin/upstream_a_pagination_20261028_test.go
- backend/internal/securityaudit/prompt_handler.go
- backend/internal/securityaudit/upstream_a_pagination_20261028_test.go
- backend/internal/service/upstream_a_pagination_20261028_test.go
- backend/internal/repository/upstream_a_pagination_20261028_test.go

Media Developer:
- backend/internal/service/openai_images.go
- backend/internal/service/openai_images_upload_20261028_test.go
- backend/internal/service/gemini_image_output_accounting.go
- backend/internal/service/gemini_image_output_20261028_test.go
- backend/internal/pkg/antigravity/response_transformer.go
- backend/internal/pkg/antigravity/upstream_a_malformed_20261028_test.go
- backend/internal/service/antigravity_gateway_service.go
- backend/internal/service/antigravity_nonstream_malformed_20261028_test.go

Each Developer writes their own external developer-<slice>/result.md and freeze.json. QA writes only external qa/ evidence and report. Controller alone owns contract, review, QA report and dedicated entries in spec/status/main-log/current-task.

## Denied Paths
All source files not explicitly listed, existing test assertions, frontend, package/lock files, dependencies, migrations/Ent, model catalogs/pricing, auth architecture, deployment/container/production/shared data and global knowledge/memory. No git add/commit/push/merge/rebase/cherry-pick or history rewrites. No real external provider, DB or container use.

## Constraints
- Developers share one isolated candidate with disjoint write sets; do not revert, rewrite or stage another slice's files. No whole-file upstream copy replacing local architecture.
- No approval inferred from mere elapsed time or pending review. Contract PASS precedes Developer dispatch; QA PASS precedes main source integration.
- Preserve source and raw evidence on failure. QA external probes may be copied temporarily only into a separately isolated QA checkout, removed afterwards, never alter candidate tests to make them pass.

## Acceptance Commands
Use GOROOT/GOTOOLCHAIN as above, GOMAXPROCS=4. Canonical external script:

```powershell
node E:/codex-runtime/sub2api/upstream-a-fixes-20261028/run-checks.mjs <root> <stage>
node E:/codex-runtime/sub2api/upstream-a-fixes-20261028/run-checks.mjs <root> <stage>-new new
```

First script runs concurrently with separate raw logs:
1. go test -json -p=2 -timeout=240s -count=1 ./internal/service -run 'CN|RateLimit|DeepSeek|CacheRead|Cache.*Usage|Usage.*Cache|Fingerprint|BillingHeader|GeminiImage|Images|Antigravity|Pagination|Paginate|GPT61|SubscriptionLongContext|UpstreamAFixes'
2. Same command with -tags=unit.
3. go test -json -p=2 -timeout=240s -count=1 ./internal/pkg/pagination ./internal/pkg/response ./internal/pkg/antigravity ./internal/handler/admin ./internal/securityaudit ./internal/repository -run 'Pagination|Paginate|Offset|ParseInt|OpsOpenAITokenStats|Antigravity|TransformGeminiToClaude|NonStreaming|UpstreamAFixes'
4. go build -p=2 ./...

Second script runs all TestUpstreamAFixes declarations across all seven affected packages default/unit -count=3. Verify declarations against JSON run events. In addition run complete ./internal/pkg/pagination ./internal/pkg/response ./internal/pkg/antigravity packages (no name filter); gofmt -l on every changed Go file; scoped git diff --check; git diff --name-only --diff-filter=U.

Supplement the unchanged fixed script with exact existing Gemini tests in default/unit:
go test -json -p=2 -timeout=240s -count=1 ./internal/service -run '^(TestCountGeminiInlineImageOutputs|TestHandleNativeNonStreamingResponse_FeedsImageCounter|TestGeminiMessagesCompatServiceForwardNative_BillsActualImageCount)$'
Record their original baseline and candidate results. Package-level skip for a package with no test declarations is not a skipped Test; no actual new or matched Test may be skipped.

Contract r2 resolves CR-001 from external contract-review-r1.md. Retain original FAIL and bindings; no baseline exclusions or weakened assertions.

## Output
Developer first line ### DONE/FAILED/BLOCKED: upstream-a-fixes-20261028-<slice>. Include changed paths, upstream provenance, red/green evidence, commands/exits, precise baseline differences, limitations and exact freeze hashes.
QA first line ### PASS/FAIL/BLOCKED: upstream-a-fixes-20261028. Include freeze/source/evidence hashes, original commands, independent probes, declaration execution audit, no skipped/new failures, and exact baseline comparison if needed.

## Stop Rules
Stop for denied paths/dependency requirements, failing unrelated original assertions not reproduced on baseline, authorization/pricing/architecture expansion, source input drift or required model unavailable. Report to controller, no silent model substitution or reduced tests.
