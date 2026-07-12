### DONE: upstream-gpt56-bare-alias-catalog-s72

# Worker Result

## Task ID

`upstream-gpt56-bare-alias-catalog-s72`

## Status

`done`

## Summary

- Added the bare `gpt-5.6` backend catalog alias and routed recognized bare aliases to canonical `gpt-5.6-sol` for OAuth/nil-account normalization and billing candidates.
- Replaced permissive GPT-5.6 substring matching with strict bare/variant/suffix parsing so `ultra`, `solstice`, `terrain`, and invalid explicit-variant suffixes remain passthrough values.
- Preserved explicit Sol/Terra/Luna identities and API-key compatible forwarding behavior.
- Exposed the bare alias in the frontend whitelist and preset mappings, and added `max` to bare/Sol/Terra/Luna OpenCode variants without changing existing limits.
- `deepseek-v4-pro` was unavailable with model 404 before dispatch; implementation used the current collaboration agent as the explicitly assigned fallback.

## Changed Files

- `backend/internal/pkg/openai/constants.go`
- `backend/internal/pkg/openai/constants_test.go`
- `backend/internal/service/openai_model_alias.go`
- `backend/internal/service/openai_model_alias_test.go`
- `backend/internal/service/openai_model_mapping_test.go`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/composables/__tests__/useModelWhitelist.spec.ts`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`
- `docs/workflow/worker-results/upstream-gpt56-bare-alias-catalog-s72-result.md`

## Commands Run

```text
gofmt -w <S72 allowed Go files> -> PASS
go test ./internal/pkg/openai ./internal/service -list <six-test-pattern> -> PASS, exactly 6 required tests
go test ./internal/pkg/openai ./internal/service -run <six-test-pattern> -count=1 -> PASS
go test ./internal/service -run "GPT56|UsageBillingModelCandidates|NormalizeOpenAIModelForUpstream" -count=1 -> PASS
vitest list useModelWhitelist.spec.ts -t <exact S72 test name> -> PASS, exactly 1 test
vitest list UseKeyModal.spec.ts -t <exact S72 test name> -> PASS, exactly 1 test
vitest run useModelWhitelist.spec.ts -t <exact S72 test name> -> PASS, 1/1
vitest run UseKeyModal.spec.ts -t <exact S72 test name> -> PASS, 1/1
pnpm --dir frontend run typecheck -> PASS
vitest run useModelWhitelist.spec.ts UseKeyModal.spec.ts -> PASS, 2 files / 24 tests
git diff --check -> PASS
conflict-marker scan over changed source/test files -> PASS, no matches
```

## Test Output

```text
Required backend discovery: 6/6
internal/pkg/openai: ok
internal/service required tests: ok
internal/service GPT56/model regressions: ok
Frontend exact test discovery: 1 + 1
Frontend targeted tests: 2/2
Frontend modified suites: 24/24
Frontend vue-tsc --noEmit: PASS
```

- The first backend discovery attempt exposed `undefined: require` in the newly added mapping test. The test was corrected to use the file's existing standard-library assertion style before all passing runs above.
- The temporary `frontend/node_modules` junction was removed after each frontend verification run.

## Risks

- No real OpenAI upstream request was performed; behavior is covered by normalization, account-scope, billing-candidate, catalog, and frontend configuration tests.
- The bare alias deliberately has no independent price and relies on the existing canonical Sol billing candidate.

## Knowledge Candidates

- None. This task only implements the approved project-specific alias contract.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- Not applicable.
