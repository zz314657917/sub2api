### DONE: upstream-openai-sticky-system-prefix-s245

## Changed files

- `backend/internal/service/openai_content_session_seed.go`
- `backend/internal/service/openai_content_session_seed_test.go`
- `docs/workflow/worker-results/upstream-openai-sticky-system-prefix-s245-result.md`

## Commands run

- Inspected approved workflow state, S245 spec addendum, contract, local direct-`gjson` owner, and upstream `e45490a36` first-parent behavior.
- `go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_ChatCompletions_(IgnoresLaterSystemMessages|UsesLeadingSystemDeveloperPrefix)$' -count=10`
- `go test ./internal/service -run '^TestDeriveOpenAIContentSessionSeed_' -count=1`
- `go test ./internal/service -count=1`
- `go test ./cmd/server -run '^$' -count=1`
- `gofmt -l internal/service/openai_content_session_seed.go internal/service/openai_content_session_seed_test.go`
- `git diff --check`; ancestry, scope, conflict/index, and protected-primary gates from the contract.

## Key output

- Focused leading-prefix regressions: PASS, `5.433s` at `-count=10`.
- Complete seed regression set: PASS, `0.067s`.
- Complete `internal/service`: PASS, `64.687s`.
- Server compile: PASS, `5.544s` (`[no tests to run]`, compile-only selector as required).
- Source and merge are ancestors of `upstream/main`; no conflict markers, unmerged entries, formatting output, or out-of-scope changed paths.
- Protected primary Pixel Cafe patch ID remained `370ac77de0e2f530ab652b99fb3eb35e809f4c84`; primary index stayed empty and `outputs/` retained two files.

## Risks

- No known implementation risk within the approved scope. Independent Terra QA and Controller review remain required before integration.

## Contract compliance

- Adapted only the leading contiguous Chat `system`/`developer` prefix guard in the existing direct-`gjson` scan.
- Did not import upstream's `86800a8cd` single-scan refactor and did not change Responses `input`, dependencies, gateway/scheduler behavior, primary worktree, provider traffic, browser automation, or remote state.
- Business commit contains exactly the two allowed product/test paths; this evidence commit contains only this report.

## knowledge_candidates

- none; the local direct-`gjson` adaptation is contract-specific and should remain in workflow evidence pending independent QA.
