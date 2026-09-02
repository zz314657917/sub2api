### PASS: upstream-content-moderation-core-s266-a

# QA Report

## Task ID
upstream-content-moderation-core-s266-a

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-content-moderation-core-s266-a.md` (amended exact allowlist)
- `docs/workflow/worker-results/upstream-content-moderation-core-s266-a-result.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`; no out-of-scope product/test path. Controller-owned `docs/workflow/main-log.md` and amended contract are expressly outside the Worker allowlist.
- denied paths touched: `no`; `frontend/pnpm-lock.yaml` and `outputs/**` have no task diff.
- provenance: frozen base `e5b62a9b9` remains an ancestor of `HEAD`; `23f3d426c`, `1b2d8873b`, `815bc6c9b`, `8b37ba882`, `948b63c9c`, `0d7b6ae64`, and final fail-open source `af6928a26` are all reachable from `upstream/main`.
- integrity: `git diff --check` passed, `gofmt -d` was empty, `git ls-files -u` and `git diff --cached --name-only` were empty.

## Commands
```text
go test ./internal/service -list 'ContentModeration|KeywordMatcher' -> PASS; relevant tests discovered
go test ./internal/repository -list 'ContentModeration' -> PASS; relevant tests discovered
go test ./internal/handler/admin -list 'ContentModeration|RiskControl' -> PASS; includes TestContentModerationHandlerUpdateConfigMapsThresholdsAndProxyID
go test ./migrations -list 'ContentModerationMatchedKeyword' -> PASS; migration test discovered
go test ./internal/service -run 'ContentModeration|KeywordMatcher' -count=10 -> PASS
go test ./internal/repository -run 'ContentModeration' -count=10 -> PASS
go test ./internal/handler/admin -run 'ContentModeration|RiskControl' -count=10 -> PASS
go test ./migrations -run 'ContentModerationMatchedKeyword' -count=10 -> PASS
go test ./cmd/server -run '^$' -count=1 -> PASS
go test ./internal/service ./internal/repository ./internal/handler/admin -count=1 -> service 65.837s PASS; handler PASS; repository blocked only by pre-existing account_repo_upstream_billing_probe_update_test.go:559 fixture panic: updatedAccountRows expected 32 columns, actual 34
node node_modules/vitest/vitest.mjs run src/views/admin/__tests__/RiskControlView.spec.ts src/features/prompt-audit/__tests__/integrationSurface.spec.ts -> PASS, 2 files / 7 tests
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit -> PASS
node node_modules/vite/bin/vite.js build -> PASS, 1904 modules; existing browserslist, dynamic-import, and chunk-size warnings only
```

## Manual Checks
```text
final fail-open: checkSync logs moderation/backend errors then returns the allow decision -> PASS
proxy: selected ProxyID resolves through ProxyRepository and the shared httpclient pool; lookup/client failure returns an error with no direct fallback -> PASS by static review and focused service tests
proxy confidentiality: audit logging uses proxy id/name/protocol/host/port, not Proxy.URL credentials -> PASS by static review
matched_keyword: pre-block keyword hit sets the log field, repository INSERT/SELECT maps it, and UI renders it -> PASS
migration 237: ADD COLUMN IF NOT EXISTS matched_keyword VARCHAR(255) NOT NULL DEFAULT '' only -> PASS; existing rows read as empty
threshold/proxy handler payload: TestContentModerationHandlerUpdateConfigMapsThresholdsAndProxyID covers request mapping -> PASS
simple mode: risk-control retains featureFlag filtering without hideInSimpleMode -> PASS
```

## Findings
- No clear issue found. The previously out-of-scope tests and handler coverage gap are resolved by the amended contract and new handler regression, independently revalidated at x10.

## Bug Owner Recommendation
`none`

## Root Cause
- `none`

## Retest Scope
- none

## Unverified Risks
- No real provider, PostgreSQL, Redis, container, deployment, shared-data, or browser-runtime session was run, in accordance with the contract.
- `go test ./internal/service ./internal/repository ./internal/handler/admin -count=1` remains blocked in the repository subpackage only by the unrelated 32/34 `updatedAccountRows` fixture drift; focused repository S266-A tests pass x10.

## Knowledge Promotion
`none`
