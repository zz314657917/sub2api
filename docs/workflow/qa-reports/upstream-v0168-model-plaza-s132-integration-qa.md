### PASS: upstream-v0168-model-plaza-s132-integration

# QA Report

## Task ID
upstream-v0168-model-plaza-s132-integration

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-v0168-model-plaza-s132-integration.md`

## Evidence
- Diff reviewed: yes. The manual conflict resolution retains current-main notification email, Pixel Cafe, user-error-request, Passkey, and security-audit wiring while adding the guarded Model Plaza behavior.
- Allowed paths checked: yes. All 35 changed source files are in the approved allowlist; this report is also an allowed artifact.
- Denied paths touched: no.
- Commands run:

```text
go test ./internal/server/routes -run '^TestModelPlazaRoutesFailClosed$' -count=1 -v -> PASS (six registered-route cases)
go test ./internal/service ./internal/handler ./internal/server/... -run 'Test.*(ModelMarket|ModelPlaza|OptionalJWT).*' -count=1 -> PASS
go test ./... -run '^$' -> PASS
go build ./... -> PASS
corepack.cmd pnpm@10.28.1 --dir frontend exec vitest run src/views/public/__tests__/ModelPlazaView.spec.ts src/__tests__/public-pages.spec.ts src/__tests__/public-smoke.spec.ts src/router/__tests__/guards.spec.ts -> PASS (68 tests)
corepack.cmd pnpm@10.28.1 --dir frontend run typecheck -> PASS
corepack.cmd pnpm@10.28.1 --dir frontend run build -> PASS
git diff --check -> PASS
git ls-files -u -> PASS (empty)
allowlist audit -> PASS
conflict-marker scan -> PASS
```

- Manual checks:

```text
missing model_plaza_enabled -> 404
model_plaza_require_auth with anonymous request -> 401
invalid JWT on catalog route -> 401; no anonymous downgrade
Backend Mode anonymous and ordinary user -> 403
Backend Mode administrator -> 200
model_plaza_description longer than 4000 Unicode code points -> rejected before persistence
```

## Findings
- No confirmed defects.
- The production frontend build reports existing chunk-size and dynamic-import warnings. It completes successfully and this task does not change the warned bundling boundaries.

## Bug Owner Recommendation
integration-owner

## Root Cause
none

## Retest Scope
- No fix retest is required. A future deployment must separately validate administrator configuration and real browser access; Docker, database, deployment, and production traffic were denied by this contract.

## Knowledge Promotion
none
