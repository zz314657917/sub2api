### PASS: upstream-v0215-frontend-20261009

## Latest Retest — 2026-10-10 / scoped local QA only

### Findings / 当前裁决

- **初始 P1/P2 已由原 Developer 修复，原独立 QA 原样 retest 通过；当前未发现剩余本合同明确缺陷。** 本 PASS 仅为本轮前端最小 delta 和相关 mounted/deferred 测试验收，**不代表 full typecheck、full production build、真实 UI/视觉、backend、部署或全库 PASS**。
- 未用 Developer DONE/128+6 自述替代证据；本次重新读实际 Ops source、正式 6 例，再亲自执行下面全部命令。
- 原外置 `ops-race.spec.ts`、`stepup-locale.spec.ts` 和 QA `vitest.config.mjs` **保持未改**，重测前后 SHA256 比较一致。初始 3 FAIL 日志 `ops-race-resolved.log`、初始完整 FAIL 报告快照 `retest-ops/initial-fail-report.md` 均保留，不覆盖为 GREEN；对应摘要在 `retest-ops/verification.json`。下方“Initial QA”是历史，不能误读为当前未关闭缺陷。

### Ops source 审查与关闭证据

实际文件：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/OpsSettingsDialog.vue`。

- 第 26-38 行：实例 `settingsRequestId` / `disposed`；失效函数递增代际并清空 `settingsLoaded/loading`；`onBeforeUnmount` 先置 disposed 再失效。
- 第 70-96 行：load 入口拒绝 hidden/disposed；每次新 load 自增代际；`isCurrentRequest` 同时判断 version、show、未 disposed；`Promise.all` 完成后先校验，再写 runtime/email/advanced/thresholds 和 loaded。旧成功无法写回或解锁当前失败窗口。
- 第 97-103 行：catch 在 toast/log 前校验；finally 仅当前请求可结束 loading。旧失败不能跨 reopen 弹错或清当前 loading。
- 第 107-113 行：show watcher 为 `flush:'sync'`，关闭同步失效，新打开新代际；与 unmount guard 合闭环。
- 第 216 行起保存流程仍为原 guard/validation/Promise.all；未改 API 或配置归一化。关闭/卸载清 loaded 后无法保存；成功加载当前数据后仍可正常保存。
- 正式 `OpsSettingsDialog.loading.v0215.spec.ts` 原 3 例保留、新增 3 条并发回归，独立全文件 **6/6 PASS**，包含在完整 128 中。新回归断言涵盖所有 update API 禁止调用、thresholds/loaded 保持、旧 error/finally 隔离、当前成功后可保存以及 unmount 不 toast/不 save。
- 原 QA 外置 3 条（实际保存按钮交互、旧 error/finally、新旧窗口与卸载）**原样 3/3 PASS**。未删例、弱化断言、生产 stub 或 QA 代修业务。

### Fresh Executed Commands / Test Counts

工作目录：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend`；证据目录：`E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/retest-ops/`。

```powershell
$rt = 'E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010'
node node_modules/vitest/vitest.mjs run --config "$rt/vitest.config.mjs" "$rt/ops-race.spec.ts" "$rt/stepup-locale.spec.ts"
$tests = Get-Content "$rt/tests-final.json" -Raw | ConvertFrom-Json
node node_modules/vitest/vitest.mjs run --config "$rt/vitest.config.mjs" @tests
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build --config "$rt/vite.config.mjs"
```

- 原外置复现/step-up/实际 root locale：**2 files / 6 PASS，exit 0**，`external-original.log`。
- 完整相关集合：**28 files / 128 PASS，exit 0**，`related-full.log`。原 125 全保留 + formal Ops 新 3；旧 UsageTable **完整 19/19**，Ops formal **6/6**；全部无 `-t` 过滤、无 skip。合计 **30 files / 134 PASS / 0 FAIL**。
- 独立重跑 typecheck **exit 2**，仍唯一相同诊断：
  `src/views/user/AirwallexPaymentView.vue(103,36): error TS2307: Cannot find module '@airwallex/components-sdk' or its corresponding type declarations.`
  修前/修后完全相同，没有本轮新增 TS 诊断，**full typecheck 未过**；`typecheck.log`。
- 真实生产源码 Vite bundling **exit 1**，188 modules transformed 后仍相同 Airwallex unresolved import，**full build 未过**；`build.log`。仍只使用原外置配置将 checker 独立运行、output/cache 外置；没有 install/externalize/stub，不声称缺包之后阶段已验证。
- `git diff --check -- frontend/src` **exit 0**，`diff-check.log`。只写自身 QA 报告和 QA runtime；源码、tests/config/依赖、F 树、backend、其他报告未由本 QA 修改。

### Fresh scope audit / Recommendation

- 重新运行 baseline SHA256 审查：仍 **20 source / 17 new test files / 1 narrow existing test**，allowlist 外新增/修改 **0**、baseline 缺失 **0**；两 group copied listener/debounce cleanup 保持；UsageTable 与 immutable baseline 仅指定两处替换 **逐字节相符**。
- 与初始 QA 保存的 existing-file baseline delta 逐文件比较，**仅 OpsSettingsDialog.vue 的差异块变化**，其他已有 source/legacy-test 的差异块均保持；正式新 test 文件实际读取并全执行，未将其原本未保留 hash 的历史逐字不变当作已证实事实。
- 当前全量 delta 与历史 diff 证据：runtime `scope.json` / `baseline-delta.patch`、`retest-ops/scope.log`、`retest-ops/initial-baseline-delta.patch`、`verification.json`。
- **可交主控作 scoped 前端 integration gate PASS**；不必等其他 backend Developer，也不覆盖他们的独立裁决。Airwallex baseline dependency blocker 和真实浏览器/视觉/后端/部署未验证边界必须继续保留。

---

## Initial QA — historical FAIL, retained evidence / superseded by retest above

初始首行裁决：`### FAIL: upstream-v0215-frontend-20261009`。以下原始发现和命令结果保持，P1/P2 已由上方原样独立重测关闭；不是当前状态。

## Findings

1. **[P1] OpsSettingsDialog 重开加载失败仍能被旧成功重新解锁保存。**
   `E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/OpsSettingsDialog.vue:57-85,197`：
   A 次打开的 thresholds promise 保持 pending，关闭并重开触发 B；B 加载失败后保存正确禁用，但 A 随后成功仍无条件赋值并设置 `settingsLoaded=true`，使 B 失败窗口可以保存 A 的旧配置。独立 shallow-mounted/deferred 测试实际点击保存按钮后，`updateMetricThresholds` 收到一次 `{sla_percent_min:97, ttft_p99_ms_max:500, request_error_rate_percent_max:5, upstream_error_rate_percent_max:5}`。这违反“Ops 失败加载不能 save”，不是 Airwallex 环境 blocker，也不因开发者 107 个测试通过而关闭。
2. **[P2] 同一 owner 的旧错误/结束状态未隔离。**
   `OpsSettingsDialog.vue:80-85,89-93`：旧请求在新一轮加载期间失败仍 `showError`，且无条件清除共享 loading；卸载后失败仍弹错。两个独立 deferred 用例均实际失败（旧错误 toast 被调用一次）。它们与上项属于同一加载生命周期根因；不能把“新窗口失败禁止保存”的顺序测试当并发隔离证据。

外置复现：`E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/ops-race.spec.ts`；最终有效日志 `ops-race-resolved.log`：**3 tests 执行，3 FAIL，exit 1**。初始两次外置测试接线失败的日志也保留，分别是 Vitest absolute-import hoisting 与外置 i18n mock module identity 问题；修正只涉及 QA runtime 配置/测试，不涉及生产替换，最终失败是实际行为断言。

## Executed Checks

### 独立性与范围

- 独立读 approved QA 合同、QA PASS review、frontend child approved 合同及 revision PASS review、Developer 最新 `### DONE` 报告；未采用 Developer 自述作为通过证据。
- 唯一检验 worktree：`E:/codex-worktrees/sub2api/upstream-v0215-20261009`。无源代码、仓库 tests/config/依赖、F 树、backend、其他报告、knowledge 修改；无 commit/push/deploy/install/browser。
- 使用 immutable `E:/codex-runtime/upstream-v0215-20261009/baseline.json` 和 `baseline/` 重新计算整个前端文件 SHA256、审阅 baseline delta，而非仅相对 HEAD。增量为 **20 source、17 new tests、1 narrow existing test**；allowlist 外增量与 baseline 缺失文件均 **0**。证据：runtime `scope.json`、`baseline-delta.patch`。
- 两个 copied group modal 的 `clearTimeout(searchTimeout)` 和 `document.removeEventListener('click', handleClickOutside)` 均保持；mounted cleanup 两例通过。旧 UsageTable 仍保留全部 19 cases，修改仅 neutral mock label 与 marker expected。
- node_modules 是共享依赖 junction，仅读取。Vitest 外置 wrapper 导入用户指定的 `E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs`，cache=false，cacheDir 位于 QA runtime `noSharedCache`；为仓库外 QA tests 增加绝对 module aliases，不改变生产实现。

### 测试命令、数量与退出码

工作目录：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend`。

```powershell
$rt = 'E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010'
$tests = Get-Content "$rt/tests-final.json" -Raw | ConvertFrom-Json
node node_modules/vitest/vitest.mjs run --config "$rt/vitest.config.mjs" @tests
node node_modules/vitest/vitest.mjs run --config "$rt/vitest.config.mjs" "$rt/stepup-locale.spec.ts"
node node_modules/vitest/vitest.mjs run --config "$rt/vitest.config.mjs" "$rt/ops-race.spec.ts"
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build --config "$rt/vite.config.mjs"
```

- 首轮独立 new/related：**21 files / 88 PASS，exit 0**，包含新 17 files/79 和旧 4 files/9。
- narrow UsageTable 修复后 final rerun：**28 files / 125 PASS，exit 0**。包括上述 88、完整旧 UsageTable **19/19 PASS**（本合同集合合计 107），另加既有 BaseDialog、BaseDialog.ids、ProxySelector.testing、useStepUp、formatMultiplier、opsFormatters 六文件 **18 PASS**。完整明确文件集合在 `tests-final.json`，无 `-t` 过滤、无删例/跳例。日志 `vitest-final.log`。
- 独立 external step-up retry-failure/reset 与实际 root locale wiring：**1 file / 3 PASS，exit 0**，日志 `stepup-locale.log`。
- 独立 external Ops overlapping requests：**1 file / 3 FAIL，exit 1**，日志 `ops-race-resolved.log`。综合有效最终集合 **128 PASS / 3 FAIL**；不是 scoped PASS。
- 自身业务路径 `git diff --check`：exit 0；未修改 index。

### 16 项最小 delta / 实际覆盖

| 项 | 独立审查与实际执行覆盖 | 结论 |
| --- | --- | --- |
| RPM | mounted 编辑：空白/小数/负数不覆盖；0、整数、科学计数整数保存；Number.isInteger 最小改动 | PASS（6） |
| group multipliers | group 切换、关闭重开、旧成功/失败与 loading、卸载；0 保留且保存；copied cleanup 不动 | PASS（6+cleanup 2） |
| channel autofill | 首次空条目单模型填价；已有模型追加保持 null；多模型不共用首价；手工价不改 | PASS（4） |
| scaled prices | 两参数 API 保持；1e-10/1e-20 数量级、常规/0/null | PASS（10） |
| sub-byte | 0.5/0.01 Bytes、0/1024、真实 ops byte-rate formatter | PASS（5） |
| neutral long context | 仅 applied 行标记；中英 label/title；真实 en.ts/zh.ts → admin/index.ts → usage 接线；account 与 GroupID pricing 0.0x 保持 | PASS（new 3、legacy 19、外置 locale 2） |
| dialog stack | 最上层 Esc、child 不可 Esc 时 parent 不关闭、重开调整栈、scroll lock/unmount | PASS（new 3，另既有 2） |
| disabled selectors | 实际打开后 disabled 关闭；重新 enabled 可打开，不发更新 | PASS（2，另 ProxySelector 1） |
| backup polling | initial load deferred 后卸载，backup/restore 两条均不复活 timers | PASS（2） |
| cleanup responses | 分页旧成功/失败、关闭重开、关闭/卸载旧失败 | PASS（5） |
| scheduled results | plan 切换、折叠重开、旧成功/错误/loading、关闭/卸载 | PASS（6） |
| template lists | 重开旧成功/失败、旧 finally 不污染 loading、关闭/卸载 | PASS（5） |
| ops save guard | 常规 pending、失败、顺序重开 guard 用例通过，但并发旧成功可解锁失败新窗口保存；旧错误跨生命周期污染 | **FAIL（Developer 3 PASS 不足；QA 3 FAIL）** |
| alert whole minutes | sustained/cooldown 小数拒绝；1440/0 接受，fractional threshold 保留 | PASS（3） |
| announcement read | 实际 Pinia store + mounted Bell：失败未读/详情不清，不报成功；重试成功才关 | PASS（1，另相关 store/locale 2） |
| step-up concurrent | 同 promise identity、全部 waiter verify/cancel、重新 prompt；mounted 双 caller；独立补测其中 retry 失败、另一成功及 reset | PASS（new 13、旧 9、外置 1） |

这些是 jsdom mounted/shallow-mounted interaction 和 mock API deferred promise 验证，**不是真实 UI/视觉、浏览器或真实后端验收**。

### Typecheck / production bundle：baseline blocker 保留

独立 typecheck **exit 2**，唯一诊断：

```text
src/views/user/AirwallexPaymentView.vue(103,36): error TS2307: Cannot find module '@airwallex/components-sdk' or its corresponding type declarations.
```

与主控 immutable baseline-checks 和该 unchanged payment owner 一致，**本次实际诊断无新增类型错误，但 full typecheck 未通过**。

独立生产源码 bundling **exit 1**，166 modules transformed，确切阻塞：

```text
[vite]: Rollup failed to resolve import "@airwallex/components-sdk" from "E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/user/AirwallexPaymentView.vue?vue&type=script&setup=true&lang.ts".
```

外置 Vite config 继承原生产配置，仅去掉环境缺 shell 入口的 checker 插件（另行真实 node vue-tsc 检查）、指定 QA runtime output/cache。**没有 externalize、install、生产 stub 或更换生产源；full production build 未通过**。提前阻塞意味着不能证明缺包之后的全部 bundling 阶段无错误。日志 `typecheck.log` / `build.log`；共享/backend dist 未写。

## Unverified Risks

- 全量 typecheck 和生产 bundle 因 baseline Airwallex 缺依赖未过；没有宣称 full repository PASS。
- 未执行 browser、视觉、真实 API、数据库、provider、部署；Go/backend 不在本 QA 范围。
- baseline 自述与当前独立复跑区分：已确认当前唯一 TS2307 和真实 bundling 缺包错误，未复制完整 baseline 环境后另造 build PASS。

## Recommendation

**回主控 / 原 Frontend Developer 修复 OpsSettingsDialog 请求代际和卸载失效边界；本 QA 不修源。** 当前结论 scoped **FAIL**，不是等待其他 backend Developer。应保留全部 tests，加入旧成功在当前失败后到达、旧失败在新加载中到达、卸载后失败三条并发回归，修后重跑本报告完整相关集合和外置复现。Airwallex baseline blocker 仍单列，不与本轮产品缺陷混淆。

报告：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/docs/workflow/qa-reports/upstream-v0215-frontend-20261009-qa.md`。
证据：`E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/`。
