# Controller evidence: upstream-response-model-audit

- Base: ebd3db125c52c16496465c7697b4b110f94e4774. Dedicated branch/worktree established; main dirty Pelican files were neither copied nor modified.
- P/G/E skill applied. Precise local monolithic topology checked; upstream split source files will not be imported. Independent contract review dispatched before development.
- Initial Terra scope-agent request failed with HTTP 429 (request 0dc90261-cf00-4c6b-95a9-7643c80e3300); retry on the same model succeeded. No model substitution.
- Baseline `go build ./...`: PASS before edits.
- Baseline `go test ./internal/handler/admin ./internal/handler/dto ./internal/repository ./internal/pkg/usagestats -run 'Test.*(Usage|Dashboard|Migration)' -count=1`: admin, DTO and repository PASS. usagestats reported no tests to run; explicitly not counted as acceptance.
- Baseline `go test -tags unit ./internal/service -run '^TestUpstreamResponseModel' -count=1` fails compilation before selection: duplicate stringPtr; old billing computeTokenBreakdown/calculateCostInternal signatures; buildCountTokensRequest arity; removed proxy FallbackMode/ExpiryWarnDays. These reproduce known baseline fixture drift and are not repaired under this contract.
- pge-doctor strict: no errors; historical status size/current-task reference warnings. Main workflow belongs to another active task; delivery uses task-specific artifacts.
- Docker server 27.2.0 available; no task containers started yet. Frozen frontend dependency installation requested in isolated worktree only.
- Status: contract review pending; no business implementation started.
- Frontend dependencies installed with frozen lockfile; pnpm 11 initially rejected locked esbuild/vue-demi install scripts. Task-only temporary allowBuilds configuration permitted exactly those two, then was removed. No package/lockfile changes. pnpm exec auto-install repeats that policy check, so direct local .cmd entries are used and documented in contract.
- Baseline direct Vitest: UsageFilters (5), UsageTable (15), UsageView (7) all PASS, 27 actual tests. No browser/provider/DB runtime acceptance claimed yet.
- Baseline frontend `vue-tsc -b` and `vite build`: PASS; existing chunk-size/Browserslist warnings only.
- Contract independently PASS after explicit protocol/turn/retry semantics, non-interference assertions, migration collision checks and executable runtime recipes were added. Approved contract dispatched to Terra Developer; separate Terra QA prepares isolated mock-browser harness concurrently, without treating preparation as final acceptance.
- Task PostgreSQL: `sub2api-response-model-audit-pg`, ID `32ea540c3efad732e46d0fb3c865da13cad5bc4d807b95931a1036b9d4921321`, loopback `127.0.0.1:56287`, task label verified, ready. PGDATA is tmpfs, no mounts/volumes. Initial empty setup container `544e8992...` was removed with its anonymous image-declared volume before replacement by tmpfs instance. No existing database/container touched.
- Active workers: `/root/audit_developer` owns gateway integration and coordinates independent Terra frontend/persistence children; `/root/audit_qa` independently prepares then verifies runtime evidence. No Developer/QA commits permitted.
- QA preparation reached actual `/admin/usage` using task fake auth and fail-closed API mocks. Exact screenshot fixture sent model `gpt-5.6-chat-thinking`, observed `chat-thinking-ev3-gpt56-v2`; variant fixture `gpt-5.5` / `gpt-5.5-2026-08-01`. Early wording corrections sent to Developer; preparation session and Vite closed. Final acceptance pending frozen implementation.
- Mid-build controller review found missing XLSX audit columns despite filter propagation; returned to original frontend owner. Existing export is XLSX (not CSV); contract clarified format preservation. NULL must stay blank while false remains distinguishable.
- Separate read-only Terra owner-chain audit found incomplete integration in generic nonstream/protocol conversion, OpenAI nonstream/SSE fallback/chat/raw/messages/native, legacy WS, Antigravity and Gemini chat compat. Detailed findings sent to Developer; implementation remains build, not final QA. Observer-only tests renamed to non-interference do not satisfy required production Forward/RecordUsage evidence; Developer instructed to replace with effective tests and execute real PG coverage.

## Final review and remediation

- Three rounds of production owner-chain review closed the identified forwarding gaps. HTTP, SSE and WS paired tests now exercise production Forward plus RecordUsage with local transports/stub repositories; WS uses a capture dialer, not a live provider.
- Independent frontend QA passed 34 focused tests, scoped lint, typecheck/build and actual admin usage page desktop/390px mock-browser checks. Downloaded XLSX was parsed for requested/sent/response/mismatch values, including NULL blank. Browser/Vite task processes were cleaned.
- Independent backend QA passed the service prefix, WS package, admin/DTO/repository gates, build and dedicated PostgreSQL migration/roundtrip/filter/index-recovery checks. Tagged service tests retain the reproduced baseline compilation failures and are not counted as passing.
- Final QA found missing id-less terminal/foreign-ID audit isolation in the WS relay. Returned to Developer: preserve existing response ID, timing, usage and billing lifecycle; isolate only audit state. Acceptance remains pending until independent revalidation.
- An out-of-allowlist admin test edit was moved into the allowed upstream_response_model_audit_test.go; the original file's diff is zero. Independent admin revalidation passed.
- Pre-integration: main remains ebd3db125c52c16496465c7697b4b110f94e4774, main index empty, migrations246/247 free. The controller restored only its own isolated status.md change; shared main workflow snapshots will not enter these commits. All 67 current business/test files fit the reviewed allowlist; gofmt clean and no pnpm-workspace.yaml drift. Dirty tracked-file hashes captured outside the repository for preservation checks.

## Acceptance and integration handoff

- Independent Terra QA final verdict: PASS after relay remediation. Id-less audit observation uses its own owner; terminal overrides use distinct model names in regressions; foreign IDs and ended turns do not receive those values; unknown fallback association yields no audit field. Existing request IDs, timing and usage lifecycle remain unchanged. Final service audit tests, full WS relay package and backend build passed independently.
- QA report: docs/workflow/qa-reports/upstream-response-model-audit-qa.md. No real provider or deployed-system acceptance is claimed. Tagged service baseline failures remain excluded.
- Exact task PostgreSQL ID 32ea540c3efad732e46d0fb3c865da13cad5bc4d807b95931a1036b9d4921321, task label and tmpfs were rechecked before removal. Post-removal exact container filter returned empty. Browser/Vite already cleaned by QA; no business container or database touched.
- Controller review: allowlist, migration numbers, Admin-only DTO exposure, nullable filtering/cache keys, generated schema offsets, model-only observation and WS regression diff checked. No remaining blocking finding. Staged whitespace checks passed.
- Backend commit: 91e9fde20. Frontend commit: f9434b081. Task evidence is committed separately. The isolated branch is accepted for safe fast-forward integration; recheck main HEAD, index and dirty overlap immediately before merging. Shared workflow/current-task files are intentionally excluded because they belong to concurrent tasks.
- Recovery: revert this batch's frontend/backend commits in reverse order; do not rewrite history. No push or deployment is authorized by this delivery.

## Final integration status: BLOCKED by concurrent dirty path

- Final pre-merge gate stopped before `git merge` was invoked. Main HEAD remains ebd3db125c52c16496465c7697b4b110f94e4774 and its index remains empty.
- New concurrent main change: frontend/src/types/index.ts adds `service_store_enabled?: boolean` to PublicSettings. This path overlaps the accepted batch's AdminUsageLog type additions, although the hunks are different. The approved integration boundary requires retaining the accepted branch when a business path overlaps dirty work; no stash, overwrite or forced integration was performed.
- Accepted commits: 91e9fde20 (backend), f9434b081 (frontend), 9acd9bab2 (contract/review/QA evidence). This final status is a separate documentation commit. Branch codex/upstream-response-model-audit and its isolated worktree are retained.
- Concurrent tasks also changed their workflow/sidebar/router files during this review window. The controller made no main worktree writes and did not absorb those changes.
- Resume after the owner resolves/commits the overlapping main change: recheck HEAD and all dirty intersections; if main advanced, incorporate that committed main state in this isolated branch and rerun affected validation before a safe fast-forward. Source implementation and isolated acceptance are complete; integration, push, migration on the business database and deployment have not occurred.
