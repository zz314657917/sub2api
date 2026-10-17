### DONE: upstream-v0216-frontend-20261015

2026-10-15，独立 Developer 开发交付；不是独立 QA PASS。依据已批准合同及同任务独立 PASS contract review，仅执行 forward recovery，不追认此前 cherry-pick 的历史门禁。已停手，交由不同 QA 验收最终候选。

## Changed paths

本轮实际修改仅以下三个隔离树文件，以及本报告：

- `E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/src/components/admin/group/GroupRateMultipliersModal.vue`
- `E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts`
- `E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts`（新增）
- `F:/mcplugins/sub2api/docs/workflow/worker-results/upstream-v0216-frontend-20261015.md`（新增；唯一主树写入）

倍率生产代码仅将 `!Number.isFinite(num) || num <= 0` 改为 `!Number.isFinite(num) || num < 0`。允许零值覆盖并保存，保留清空覆盖和正小数/科学计数语义；拒绝负数、NaN、Infinity、-Infinity、溢出及无效文本。测试还验证无效编辑不会覆盖已暂存的合法倍率。number input 会将非有限文本归一为空，因此无效输入测试通过 DOM value getter 提供原始 change 值，实际执行组件事件处理器，避免误测成清空行为。

AccountGroupsCell 新测挂载真实 AccountGroupsCell/GroupBadge，覆盖完整长名称、容器及名称换行 classes、其他调用默认 `truncate=true`、六组默认显示三组和 `+3`、body Teleport 弹层包含全部六组及 Escape 关闭。

其他既有候选代码仅审阅，未修改。原 `GroupRateMultipliersModal.requests.v0215.spec.ts` 保持逐字节不变。

## Executed checks

执行目录：`E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend`。

1. 合同指定必测：7 文件、41/41 PASS，零跳过、零失败：

   ```powershell
   node node_modules/vitest/vitest.mjs run src/components/payment/__tests__/paymentCountdown.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.v0215.spec.ts src/components/common/__tests__/Select.ime.spec.ts src/views/admin/ops/components/__tests__/OpsDashboardHeader.timezone.spec.ts src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts src/composables/__tests__/useGeminiOAuth.errors.spec.ts
   ```

   edit 11 项、requests.v0215 6 项、AccountGroupsCell 3 项、paymentCountdown 9 项、Select IME 4 项、Ops timezone 3 项、Gemini errors 5 项。日志：`E:/codex-runtime/upstream-v0216-20261015/developer-vitest-required.log`。

2. 相关既有回归：25 文件、139/139 PASS，零跳过、零失败。包括支付组件、支付页面/API/store/i18n、disabledSelectors、Ops SLA 和 datetime-local 工具。用精确路径数组执行 `node node_modules/vitest/vitest.mjs run @files`；完整参数数组保存于 `E:/codex-runtime/upstream-v0216-20261015/developer-vitest-regression-paths.txt`，日志为同目录 `developer-vitest-regression.log`。其中 paymentCountdown 9 项在两轮中重复执行；两轮合计 180 次执行，171 个不同测试。

3. `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit`：退出码 0。日志：`E:/codex-runtime/upstream-v0216-20261015/developer-vue-tsc.log`（成功无输出）。

4. Scoped ESLint：候选相对 base 的全部 15 个 Vue/TS 文件，退出码 0，无 `--fix`：

   ```powershell
   $env:NODE_PATH='E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/node_modules/.pnpm/node_modules'
   node node_modules/eslint/bin/eslint.js @files
   ```

   精确路径数组：`E:/codex-runtime/upstream-v0216-20261015/developer-eslint-paths.txt`；日志：同目录 `developer-eslint.log`（成功无输出）。

5. `git diff --check` 和 `git diff --check de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d HEAD`：均退出码 0。日志：runtime 下 `developer-diff-check.log`、`developer-candidate-diff-check.log`。

按主控本轮明确指示未重跑 Vite：主控报告此前完整 PASS；这不是本 Developer 对最终候选的构建证据，最终 Vite 由独立 QA 执行。

## Scope/hash evidence

- 隔离树基线清洁，HEAD 始终为 `a19b132ec491ba99db319548c82f1e0edf68c393`。
- 开始时对隔离树 Git 可见的 4675 个文件保存 SHA-256；结束对照仅上述两个既有文件变化，另新增一个指定测试，其余 4673 个基线文件无 hash 变化或删除。
- `GroupRateMultipliersModal.requests.v0215.spec.ts`、其他候选代码、依赖清单及锁文件均在未变化集合。
- 最终 index 为空；Git status 仅两个指定修改及一个指定新增。未新建 worktree、提交、暂存、push、部署或派子代理。
- 完整前后 hash/范围证据：`E:/codex-runtime/upstream-v0216-20261015/developer-before.json`、`developer-scope-audit.json`。精确 15 路径最终候选 manifest 留给独立 QA 生成。
- 主树 12111 文件保护由主控独立核对；本 Developer 未复跑或扩大该核对，只写上述授权报告。

## Remaining verification

未进行浏览器验收。jsdom 中完整文本和 class 断言不证明实际浏览器换行、宽度或行高效果；无真实支付/provider/runtime 验收声明。独立 QA 和最终候选 Vite 尚待执行；没有主树集成。Vitest 提示 Browserslist 数据过旧，测试仍全部通过，未更新依赖。
