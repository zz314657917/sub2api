### PASS: upstream-response-model-audit

# Contract Review

## Task ID
upstream-response-model-audit

## Contract Checked
- `docs/workflow/tasks/upstream-response-model-audit.md`
- `docs/workflow/tasks/upstream-response-model-audit-spec.md`

## Evidence
- Baseline is explicitly pinned to `ebd3db125c52c16496465c7697b4b110f94e4774`; it is an ancestor of the isolated worktree `HEAD` and the worktree was clean before this review artifact.
- Compared the authorized behavior references `db0bff82c`, `6e34fb09c`, and `c46d07ca0` with the local topology. The local checkout consolidates persistence, filtering, totals, statistics, trends, batch writes, and exports in `backend/internal/repository/usage_log_repo.go`; dashboard cache owners are `backend/internal/handler/admin/dashboard_query_cache.go` and `dashboard_snapshot_v2_handler.go`. All are explicitly allowlisted.
- Confirmed DTO isolation: ordinary-user handlers serialize `dto.UsageLogFromService`, while administrator handlers serialize `dto.UsageLogFromServiceAdmin`. The contract limits the two audit fields to the administrator DTO and includes direct mapper/handler tests for the non-leak boundary.
- Confirmed all local forwarding owners needed for HTTP, SSE, protocol conversions, Antigravity, WebSocket V2, WebSocket HTTP bridge, and passthrough relay are allowlisted. The contract uses the local monolithic `openai_ws_forwarder.go`, not upstream split-file names.
- Verified the finite Grok equivalence groups match `c46d07ca0`: `{grok-4.5, grok-4.5-latest, grok-4.5-build}` and `{grok-4.6, grok-4.6-latest, grok-4.6-build}`.
- Verified protocol semantics now define OpenAI terminal events, Anthropic first-declaration handling, Gemini latest-valid-`modelVersion` behavior (including the two wrapped paths), and strict WS response-ID/active-turn ownership. Foreign IDs, ambiguous/id-less events, new turns, and failed-attempt retries cannot leak an observation across attempts.
- Verified the contract provides dedicated disposable PostgreSQL creation/readiness/DSN/label-check/removal commands, and a task-owned browser recipe with mock-only API routing, fixtures, viewport checks, screenshots, explicit close, and precise owned-process cleanup. Direct local frontend commands are authorized when pnpm 11 attempts a new install and rejects locked build scripts.

## Findings
- The initial review required six corrections: restore the `contract-draft` phase; define Gemini/WS/retry ownership; enumerate Grok aliases; make PostgreSQL/browser acceptance reproducible; add production-path non-interference cases; and add the migration collision/recovery gate. The amended contract now covers each item, including named HTTP/SSE/WS non-interference tests and the pre-edit plus pre-integration 246/247 collision checks.
- No blocking ambiguity remains in the authorized local adaptation. The draft metadata is correct before review; the controller must now perform the P/G/E transition by recording this PASS, setting the contract front matter to `status: approved` and `review_verdict: PASS`, then moving `docs/workflow/status.md` to `contract-approved` before dispatch.
- The tagged `go test -tags unit ./internal/service -run '^TestUpstreamResponseModel'` command has documented pre-edit fixture compile drift. It is retained solely as baseline reproduction/isolation evidence and must never be reported as a passing task check. Untagged production-path tests, the isolated PostgreSQL test, and the build remain required PASS gates.

## Gate Checks
- success_criteria_testable: **yes**
- allowed_paths_explicit: **yes**
- denied_paths_explicit: **yes**
- acceptance_commands_executable: **yes**
- worker_model_confirmed: **yes** (`gpt-5.6-terra`)
- base_commit_confirmed: **yes** (`ebd3db125c52c16496465c7697b4b110f94e4774`)
- openspec_traceable: **yes** (`docs/workflow/tasks/upstream-response-model-audit-spec.md`)

## Verdict
`PASS` — implementation may start only after the controller applies the required contract/status metadata transition. Scope remains limited to the exact local allowlist; no user DTO exposure, billing/routing/account-state behavior change, main-worktree write, shared database, provider access, push, or deployment is authorized.

## Scope Amendment Review — `openai_ws_forwarder_success_test.go`

### PASS: upstream-response-model-audit scope-amendment-ws-success-test

- Reviewed the complete uncommitted diff for `backend/internal/service/openai_ws_forwarder_success_test.go`: exactly four added assertions, with no production-code change, fixture rewrite, import change, or test-flow change.
- The assertions extend two existing real `OpenAIGatewayService.Forward` WebSocket success tests. They verify the raw upstream response model is captured and that a non-conflicting response leaves the diagnostic conflict flag false; the second case also preserves its pre-existing assertion that the client-visible model is rewritten to the original request model.
- This is a narrower and stronger owner than duplicating the production WS success fixtures into a new path. It directly satisfies the contract's required WS production-path/non-interference evidence and does not expand provider, billing, routing, account-state, or payload behavior scope.

**Approved amendment:** add exactly `backend/internal/service/openai_ws_forwarder_success_test.go` to the contract allowlist, limited to assertions of `UpstreamResponseModel` and `UpstreamResponseModelConflict` in existing real WebSocket Forward success tests. Any fixture restructuring, new test scenario, production-file modification, or assertion outside response-model audit behavior requires another amendment review.
