---
status: approved
review_verdict: PASS
task_id: image2-native-port
worker_model: gpt-5.6-terra
base_commit: bb3dde9e4
spec_ref: docs/workflow/plans/image2-selective-integration-20260915.md
---

## Task ID
image2-native-port

## Role
Terra Generator; separate contract Evaluator and Terra QA.

## Goal
Complete I-03/I-04 native OAuth Images behavior port from c0d511937, 1067e89fa and 23eef9ecc in the local consolidated gateway. Preserve local account mapping, API key splitting, APIMart and Responses fallback. This contract covers source and local mock acceptance; real provider acceptance remains a separate open Epic requirement.

## Success Criteria
1. Explicit supported models (gpt-image-1.5, gpt-image-2, both Image 2.5 models and their 2026-09-08 aliases) select native generation/edit; legacy and unknown snapshots retain Responses. Preserve configured Responses driver.
2. JSON/multipart requests preserve validated options, uploads/masks, n, stream, model mapping and response format. Native 404/405 falls back once to Responses, with no recursion loop. Other errors keep existing retry/cooldown boundaries.
3. Native JSON and SSE convert base64 to requested data URLs, report requested public model, derive actual PNG/JPEG/WebP image dimensions where locally supported, reject empty output, and preserve successful partial results and usage without unsafe retry after output.
4. Separate cached image input tokens and image output tokens flow to existing billing; retain cached image split in existing usage JSON payload without mutating caller maps or double charging. No schema change.
5. Implement in reviewable steps: shared image helpers and missing review behavior; direct adapter; dispatch/fallback; usage accounting; transport mock regression. Missing upstream helpers must be adapted to existing local owners, not replaced with stubs.
6. Tests exercise actual production functions with fake transport/provider: generation/edit, streaming/nonstreaming, 404/405 fallback, retryable pre-output failure, failure after output, empty output, driver-vs-image plan gates, actual size, model mapping and usage/cost. No external HTTP or secrets.
7. For JSON and SSE separately, assert mapped upstream model in captured request, requested public model in downstream payload, url output has expected MIME data URL and no b64_json, and b64_json output remains base64. Assert edit SSE uses image_edit prefix. Explicitly cover successful partial output followed by failure and assert no second upstream request after output.
8. RecordUsage regression must use captured usage repository and balance/subscription charging fakes: assert exact expected cached-image cost from fixed prices, persisted image_cache_read_tokens value and size entries, unchanged caller ImageSizeBreakdown map, and exactly one charge for one request. Repeat same request ID through the existing dedupe path and assert no second charge. Cover absent cached tokens (no metadata key), populated and nil size maps. Do not claim merely testing a JSON helper proves billing.

## Allowed Paths
- backend/internal/service/openai_images.go
- backend/internal/service/openai_images_responses.go
- backend/internal/service/openai_images_direct.go
- backend/internal/service/openai_images_native_helpers.go
- backend/internal/service/openai_images_test.go (existing Responses protocol fixtures only; preserve assertions and APIKey/APIMart contexts)
- backend/internal/service/openai_images_*_test.go
- backend/internal/service/openai_gateway_service.go (image usage record hunks only)
- backend/internal/service/openai_gateway_record_usage_test.go
- backend/internal/service/image_output_accounting.go
- backend/internal/service/billing_service.go (only missing native image token cost integration)
- docs/workflow/tasks/image2-native-port.md
- docs/workflow/worker-results/image2-native-port-result.md
- docs/workflow/qa-reports/image2-native-port-qa.md

## Denied Paths
All other business files; account_test_service.go; handler; wire; frontend; config defaults; Ent; migrations; dependencies; shared databases; secrets; production; containers; deployment; push. Never modify main worktree or other worktrees. No merge/rebase/cherry-pick, commits by worker, or wholesale upstream patch application.

## Constraints
Use apply_patch, preserve existing architecture. Existing E:/codex-worktrees/sub2api/image2-i04/openai_images_direct.go experiment (under backend/internal/service) is reference only and does not compile. Do not copy its unresolved dependencies. I-01/I-02 already committed. Runtime mock PASS must not be called real provider PASS. Account-test parity requires later scoped integration because its main-tree owner is dirty.

## Acceptance Commands
From backend: go test ./internal/service -run 'Test.*(OpenAIImages|CodexDirectImages|ImageOutput|ImageCache|RecordUsage)' -count=1; go build ./...
Also required: go test ./internal/service -run 'TestOpenAIGatewayServiceForwardImages' -count=1. The primary regex does not select this existing forwarding-test naming family.
Add native tests with TestCodexDirectImages or TestOpenAIImages prefixes so selection includes them. Use go test -tags unit only if selected test files require it; establish baseline compile errors separately without changing unrelated tests. Run gofmt -l on changed Go files, git diff --check, git diff --name-only, git ls-files -u. Record exact executed tests and mock assertions.

Required executable path/format gate (PowerShell from worktree root; intentionally includes tracked and untracked business files):

```powershell
$nativePaths = @((git diff --name-only bb3dde9e4); (git ls-files --others --exclude-standard)) | Sort-Object -Unique
$nativeAllowed = '^backend/internal/service/(openai_images\.go|openai_images_responses\.go|openai_images_direct\.go|openai_images_native_helpers\.go|openai_images_test\.go|openai_images_.*_test\.go|openai_gateway_service\.go|openai_gateway_record_usage_test\.go|image_output_accounting\.go|billing_service\.go)$'
$nativeDenied = @($nativePaths | Where-Object { $_ -notmatch '^docs/workflow/' -and $_ -notmatch $nativeAllowed })
if ($nativeDenied.Count -gt 0) { throw ('Denied paths: ' + ($nativeDenied -join ', ')) }
$nativeGo = @($nativePaths | Where-Object { $_ -match '\.go$' })
if ($nativeGo.Count -gt 0) { $nativeUnformatted = @(gofmt -l $nativeGo); if ($LASTEXITCODE -ne 0 -or $nativeUnformatted.Count -gt 0) { throw ('gofmt failed: ' + ($nativeUnformatted -join ', ')) } }
git diff --check bb3dde9e4
if ($LASTEXITCODE -ne 0) { throw 'diff check failed' }
if (@(git ls-files -u).Count -gt 0) { throw 'unmerged index' }
```

Additionally inspect `git diff bb3dde9e4` for hunk constraints and manually list ignored task evidence using `git status --short --ignored docs/workflow`; the docs exclusion above is only for controller/review evidence and does not broaden the worker's allowed documentation paths.

## Output
Worker report starts ### DONE/FAILED/BLOCKED: image2-native-port with changed paths, step coverage, commands, mock evidence and remaining gaps. Independent QA starts ### PASS/FAIL/BLOCKED: image2-native-port. No unsupported success claims.

## Stop Rules
Stop and report if schema/config/denied owner changes are needed or tests cannot be distinguished from baseline. Two implementation failures return to Planner. Do not shrink requirements to pass tests.
