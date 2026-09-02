### PASS: upstream-content-moderation-cyber-policy-s266-b

# Independent QA Report

## Findings

- No confirmed product defect was found in the S266-B product commit.
- The source and focused regressions confirm `cyber_policy` is terminal for Responses, Chat Completions, Messages, compatibility fallbacks, and Responses WebSocket: it is marked once, returned in the endpoint-compatible response shape, does not become an `UpstreamFailoverError`, and the forwarding loops therefore cannot select another account or append a fallback body.
- Session blocking is opt-in and fail-open. `CyberSessionBlockKey` requires a positive API-key ID plus an explicit header/body session signal, hashes their isolated combination, and returns empty for missing identity. The HTTP rejection runs before account selection; the WebSocket `BeforeTurn` checks `cyberBlockedThisConnection` before upstream I/O, while `AfterTurn` clears the per-turn mark and recorded flag.
- Cyber audit uses the Risk Control runtime snapshot and its group/model inclusion checks only; it ignores normal moderation enabled/mode/sample controls. Audit creates the moderation log before email, redacts the stored upstream text, and the exclusion switch passes the correct historical-cyber filter to automatic-ban counting.
- `RecordCyberPolicyUsageLog` carries upstream-observed tokens with `CyberBlocked=true`. The focused test evidence covers zero tokens producing a no-charge row and observed tokens being billed exactly once. The handler avoids a second partial-usage path when standalone cyber usage is selected.
- One initial invocation of the required service x10 command failed once in `TestRecordCyberPolicyEvent_RuntimeSnapshotRefreshFailureKeepsStaleScope` after its one-second asynchronous `Eventually` deadline. Re-running the complete service x10 command passed; the isolated test passed at `-count=20` and `-count=100`. This is recorded as a residual test-stability risk, not a reproducible product failure.

## Commands / Evidence

Executed from `E:/codex-worktrees/sub2api/upstream-content-moderation-parity-s266` on product commit `eeed2369f` (HEAD evidence commit `8ebb4fb5b`).

- Discovery (all non-empty, exit 0):
  - `go test ./internal/service -list 'Cyber|ContentModeration|OpenAI.*Policy'`
  - `go test ./internal/handler -list 'Cyber|OpenAI'`
  - `go test ./internal/handler/admin -list 'ContentModeration|Cyber|Settings'`
  - `go test ./internal/repository -list 'Cyber|ContentModeration|OpsError'`
- Required focused x10:
  - service: first run had the single asynchronous test failure noted above; fresh complete rerun exit 0 in 11.692s; isolated stale-snapshot test x20 and x100 both exit 0.
  - `go test ./internal/handler -run 'Cyber|OpenAI' -count=10` -> PASS, 10.490s.
  - `go test ./internal/handler/admin -run 'ContentModeration|Cyber|Settings' -count=10` -> PASS, 0.163s.
  - `go test ./internal/repository -run 'Cyber|ContentModeration|OpsError' -count=10` -> PASS; independent repeat 0.150s.
- Additional critical proof, both `-count=10` -> PASS:
  - service no-failover, single-body, WS-next-turn, mark-clear, session round-trip, zero/exact-token billing, redaction, audit scope/ban-count and log-before-email tests.
  - handler marked-event, duplicate-usage avoidance, fail-open session rejection, block-key plumbing, ops status/redaction and middleware-skip tests.
- `go test ./cmd/server -run '^$' -count=1` -> PASS.
- Frontend:
  - `node node_modules/vitest/vitest.mjs run src/features/prompt-audit/__tests__/integrationSurface.spec.ts src/views/admin/__tests__/RiskControlView.spec.ts` -> PASS, 2 files / 7 tests.
  - `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` -> PASS.
  - `node node_modules/vite/bin/vite.js build` -> PASS, 1904 modules, 22.73s (only existing Browserslist/chunk-size and dynamic-import warnings).
- Scope/integrity:
  - `gofmt -d` over all 43 changed Go files -> empty, exit 0.
  - `git diff --check eeed2369f^ eeed2369f` and working-tree `git diff --check` -> PASS.
  - `git ls-files -u`, `git diff --cached --name-only` -> empty.
  - Allowlist comparison: 56 product files changed, 62 allowed entries, 0 outside. No denied production/config/dependency/migration/container/shared-data paths are in the product commit.

## Scope / Provenance

- Current QA dispatch HEAD is `8ebb4fb5b` (`docs(workflow): dispatch cyber policy qa`), descending from Controller evidence `4fd7a2aa1`, which descends from product commit `eeed2369f`.
- The product parent is `79d8b3cc2`; `eeed2369f` contains exactly 56 contract-allowed product/test files. The protected untracked `upstream-content-moderation-cyber-policy-s266-b/` directory remained unentered, unstaged and unchanged.
- No provider, SMTP, Redis, PostgreSQL, container, deployment, staging or push action was performed. The QA wrote only this report; the contract-required Vite build regenerated ignored local web-dist artifacts but left no tracked working-tree modification.

## Known Baseline Failures

- Default `go test ./internal/server -run '^TestAPIContracts$' -count=1` passes with no tests because `TestAPIContracts` has the `unit` build tag.
- Independent `go test -tags=unit ./internal/server -run '^TestAPIContracts$' -count=1` fails four stale contract snapshots. The reported deltas are pre-S266 Group/Usage/Settings fields (for example `access_mode`, `allow_live`, long-context and broader settings fields); the S266-B `cyber_session_block_enabled` and `cyber_session_block_ttl_seconds` fields are already present in the expected settings body and are not failure deltas.
- Independent full `go test ./internal/repository -count=1` fails before S266-B focused tests at `account_repo_upstream_billing_probe_update_test.go:559`: SQL mock expects 32 values while the row has 34 columns. This is outside the S266-B product diff; the contract-focused repository x10 suite passes.

## Unverified Risks

- No real provider, SMTP, Redis, PostgreSQL, shared database, container, deployment, or browser-authenticated runtime smoke was allowed by the contract.
- The one observed stale-runtime-snapshot test timeout did not reproduce in the subsequent focused rerun or 100 isolated repetitions, but it remains worth monitoring in heavily loaded CI.
- The four tagged API snapshots and the independent repository billing fixture require separate baseline maintenance; neither was modified for S266-B.

## Final Recommendation

PASS for the approved S266-B scope. The product implementation and its allowed tests satisfy the contract under local mock/httptest evidence; retain the documented baseline and test-stability follow-ups outside this slice.
