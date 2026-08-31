### DONE: upstream-v0184-frontend-compat-s277

# Worker Result

## Task ID
`upstream-v0184-frontend-compat-s277`

## Status
`done`

## Summary
完成 S277 allowlist 内三项前端兼容行为的定向回归核对，并修正兑换码有效 custom expiry 测试的期望值：严格本地 datetime parser 按 contract 截断到秒，因此 ISO 结果使用毫秒 `000`。未修改任何 denied path，未执行 commit、push 或外部状态操作。

## Changed Files
- `frontend/src/utils/format.ts`
- `frontend/src/utils/__tests__/formatDateTimeLocalInput.spec.ts`
- `frontend/src/views/admin/RedeemView.vue`
- `frontend/src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`
- `docs/workflow/worker-results/upstream-v0184-frontend-compat-s277-result.md`

## Commands Run
```text
cd frontend; pnpm.cmd exec vitest run src/utils/__tests__/formatDateTimeLocalInput.spec.ts src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts -> pass
cd frontend; pnpm.cmd run typecheck -> pass
cd frontend; pnpm.cmd run build -> pass
git diff --check -- frontend/src/utils/format.ts frontend/src/utils/__tests__/formatDateTimeLocalInput.spec.ts frontend/src/views/admin/RedeemView.vue frontend/src/views/admin/__tests__/RedeemView.batchUpdate.spec.ts frontend/src/components/keys/UseKeyModal.vue frontend/src/components/keys/__tests__/UseKeyModal.spec.ts -> pass
git diff --name-only --diff-filter=U -> empty
rg -n "CLAUDE_CODE_ATTRIBUTION_HEADER" frontend/src/components/keys/UseKeyModal.vue frontend/src/components/keys/__tests__/UseKeyModal.spec.ts -> only negative absence assertions in test; no generated-path or fixture occurrence
git status --short -> existing denied-path dirty changes preserved; no new denied-path edit
```

## Test Output
```text
Vitest: 3 files passed, 31 tests passed.
typecheck: vue-tsc --noEmit passed.
build: vue-tsc -b && vite build passed (1904 modules transformed).
Non-blocking existing warnings: stale Browserslist data, Vite dynamic-import chunk notices, and large chunks.
```

## Upstream Provenance
- `81e461f65`: `ported` - strict component-level local datetime parsing adapted to `src/utils/format.ts`.
- `b7aca87fd`: `ported` - focused parser valid/invalid regression coverage adapted locally.
- `5778739cd`: `ported` - Redeem batch custom expiry now parses strictly and serializes parsed seconds.
- `c03776604`: `ported` - remaining Claude settings JSON attribution override removed; nonessential traffic protection retained across variants.

## Risks
- No provider, database, container, deployment, or browser runtime smoke was run; these are outside the contract.
- Build warnings are pre-existing toolchain warnings and did not fail the build.
- Existing unrelated dirty paths (`backend/**`, `frontend/pnpm-lock.yaml`, Pixel Cafe, workflow docs, `knowledge/**`, `outputs/**`) remain untouched and require separate ownership.

## Knowledge Candidates
- `parseDateTimeLocalInput` returns second-resolution timestamps; custom expiry ISO serialization must therefore use `.000Z` for sub-second input.
- Claude settings JSON and shell variants retain `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` without emitting the attribution-header override.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
