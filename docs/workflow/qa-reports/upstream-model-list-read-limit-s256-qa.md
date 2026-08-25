### PASS: upstream-model-list-read-limit-s256

# QA Report

## Task ID

upstream-model-list-read-limit-s256

## Verdict

PASS

## Contract Checked

- `docs/workflow/tasks/upstream-model-list-read-limit-s256.md`

## Evidence

- QA baseline: `d81022917`; business commit: `2f04ba987`; upstream behavior source: `847c0c452`.
- Fresh isolated worktree and branch: `E:/codex-worktrees/sub2api/upstream-model-list-read-limit-s256-qa` at `pge/upstream-model-list-read-limit-s256-qa`.
- Diff reviewed: yes. The business commit contains exactly the 13 allowed product/test/config/example/Wire paths; its evidence child adds only the allowed Developer result report.
- Denied paths touched: no. `git diff --cached --name-only`, `git diff --name-only`, and `git ls-files -u` were empty before this report was added.
- Provenance: `847c0c452` is reachable from `upstream/main`; local main `536c4bde3` is an ancestor of `2f04ba987`; business commit is an ancestor of `d81022917`.

### Commands run

```text
cd backend; go test ./internal/config -list 'Test.*ModelsList.*' -> PASS
cd backend; go test ./internal/pkg/antigravity -list 'Test.*FetchAvailableModels.*(Limit|limit)|Test.*ModelsList.*' -> PASS
cd backend; go test ./internal/service -list 'Test(ResolveModelsListReadLimit|FetchCodexModelsManifest.*Configured.*Limit|FetchUpstreamSupportedModels.*Configured.*Limit|Antigravity.*ModelsList.*Limit)' -> PASS (six matching tests discovered)

cd backend; go test ./internal/config -run 'Test.*ModelsList.*' -count=10 -> PASS (0.927s)
cd backend; go test ./internal/pkg/antigravity -run 'Test.*FetchAvailableModels.*(Limit|limit)|Test.*ModelsList.*' -count=10 -> PASS (1.446s)
cd backend; go test ./internal/service -run 'Test(ResolveModelsListReadLimit|FetchCodexModelsManifest.*Configured.*Limit|FetchUpstreamSupportedModels.*Configured.*Limit|Antigravity.*ModelsList.*Limit)' -count=10 -> PASS (5.447s)

cd backend; go test ./internal/config -count=1 -> PASS (1.508s)
cd backend; go test ./internal/pkg/antigravity -count=1 -> PASS (0.938s)
cd backend; go test ./internal/service -count=1 -timeout=3m -> PASS. The task-owned pass sentinel was started only after this command exited 0, observed, then precisely cleaned up.
cd backend; go test ./cmd/server -run '^$' -count=1 -> PASS (18.514s)

gofmt -d <all 12 allowed Go production/test files> -> PASS (no output)
git diff --check 2f04ba987^ 2f04ba987; git diff --check -> PASS
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' <all allowed Go files> -> PASS (none)
git diff --cached --name-only; git diff --name-only; git ls-files -u -> PASS (all empty before QA report)
git -C F:/mcplugins/sub2api status --short -> read-only protection check; 106 pre-existing user-owned entries, including S252 `backend/cmd/server/wire_gen.go`, were not modified or staged by QA
```

## Contract Checks

- `gateway.models_list_read_max_bytes` is registered at exactly `8 * 1024 * 1024`, appears in the example YAML as `8388608`, and `Config.Validate` rejects every `<= 0` value. The service resolver safely falls back to the same 8 MiB value for nil and zero-value configs.
- Generic `/models` reads, Codex manifests, Antigravity account-model reads, and Antigravity quota reads each resolve and pass the configured limit. The Antigravity client rejects non-positive input and uses `LimitReader(limit + 1)`.
- Codex uses `LimitReader(limit + 1)` and returns `OPENAI_CODEX_MODELS_UPSTREAM_FAILED` with `retryable: true` before parsing an oversized manifest. The pre-existing cache cap remains `codexModelsManifestCacheBodyLimit = 1 << 20` and rejects larger cache entries.
- Wire diff is exactly `NewAntigravityQuotaFetcher(proxyRepository, configConfig)`; no `wire.go` or unrelated generated topology changed.

## Findings

未发现明确问题。

## Scope

- Only the QA report was written by this Worker. No business source/test/config files, primary-worktree files, provider, database, container, deployment, or Git remote state were changed.
- No real provider, database, container, deploy, or push was attempted. Runtime evidence uses existing local package tests and `httptest`/stubs only.

## Risks

- This is code/package-level verification only. Real upstream provider behavior remains intentionally unverified under the contract boundary.
- Mainline integration still needs its separate protected-Wire application review against the user-owned S252 changes in the primary worktree.

## Bug Owner Recommendation

none

## Root Cause

none

## Retest Scope

None required.

## Knowledge Promotion

none
