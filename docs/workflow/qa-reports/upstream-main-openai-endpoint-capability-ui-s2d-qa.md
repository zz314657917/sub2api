### PASS: upstream-main-openai-endpoint-capability-ui-s2d

# upstream-main-openai-endpoint-capability-ui-s2d QA Report

## Task ID
upstream-main-openai-endpoint-capability-ui-s2d

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-openai-endpoint-capability-ui-s2d.md`

## Evidence
- diff reviewed: yes
- denied paths touched: no
- commands run:
```text
git status --short --branch -> only contract-allowed frontend/workflow paths modified before QA
git diff --check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts -> pass, 1 file / 16 tests
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
```
- manual checks:
```text
Upstream source commit -> 37044b83e
Implementation preserves local modular i18n files and does not touch monolithic en.ts / zh.ts
OpenAI API Key edit path writes credentials.openai_capabilities only for non-default subsets
OpenAI API Key edit path deletes openai_responses_mode when text capability is disabled
Existing extra.openai_responses_supported is copied forward by editing extra from current account extra
OpenAI API Key create path passes through createAccountAndFinish, which now applies endpoint capability serialization
No backend, schema, migrations, gateway, billing, scheduling, or public API paths changed
```

## Findings
- 未发现当前 Sprint 2d 补丁引入的明确阻断问题。
- 新增目标 Vitest 覆盖了 embeddings-only 保存、恢复默认能力时省略 override、Responses mode 禁用态，以及 `openai_responses_supported` 保留。
- 前端 `typecheck` 与 `lint:check` 均通过。
- Vitest 输出包含既有 Browserslist/caniuse-lite 数据过期提示，不影响本次测试结果。

## Bug Owner Recommendation
none

## Root Cause
- none

## Retest Scope
- None.

## Unverified Risks
- 未执行浏览器级创建/编辑弹窗截图验收。
- 未启动完整前端/后端运行时做页面 smoke。
- 未使用真实 OpenAI API Key 做路由 smoke；该验证会依赖实际账号和上游额度。

## Knowledge Promotion
- none
