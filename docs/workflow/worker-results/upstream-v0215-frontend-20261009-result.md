### DONE: upstream-v0215-frontend-20261009

## Findings / Verdict
- 合同 approved/PASS、revision review PASS 已核对。只在 `E:/codex-worktrees/sub2api/upstream-v0215-20261009` 修改 allowlist source/test 和本报告；无 commit/push/deploy、无 F 主树编辑、无后端修改。
- 最新 Ops QA feedback 修复后：新测试 **17 文件 / 82 tests PASS**；相关既有测试 **11 文件 / 46 tests PASS**，合计 **28 文件 / 128 tests PASS**（原独立 QA 的 125 项保留，加 3 条正式回归），exit 0。原外置 Ops 复现 **3/3 PASS**，外置 step-up/locale **3/3 PASS**；全部最终有效集合 **30 文件 / 134 tests PASS**。
- narrow revision 独立 PASS 后，仅向既有 `E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts` 增加真实 neutral 英文 mock label `Long context`，并将 marker expected 从 `x2` 改为 `Long context`。其余 copied 内容、用例名及断言保持不变，完整 **19/19 PASS**，本轮再次随完整集合重跑；未跳过用例。
- 独立 QA 的三条 Ops 并发失败属于真实产品缺陷，之前顺序 guard 测试覆盖不足，不是 baseline 问题。已仅在 Ops owner 范围加入请求代际、关闭/卸载失效及 success/error/finally 保护，原业务加载、配置归一化、保存流程不变；正式新增三例在修复前全部 RED，修后全部 GREEN。
- DONE 表示合同范围实现与 scoped/related mounted 测试完成，等待独立 QA 重跑；不代表独立 QA 已 PASS 或 full build/typecheck PASS。
- typecheck 仅有既知 baseline `AirwallexPaymentView.vue(103,36)` TS2307，找不到 `@airwallex/components-sdk`；**新增类型错误 0**。生产 bundling 同样因缺失该依赖失败；未补包、未 externalize/stub 来宣称生产构建通过。
- 最初 Vite checker 调用共享依赖缺少的 `vue-tsc` shell 入口失败；外置临时配置移除 checker 后独立检查 bundling。真实类型检查仍使用指定 node 入口，没有绕过新类型错误。

## Changed Files
相对 copied dirty baseline：**20 source + 17 new tests + 1 narrow existing test + 1 own report**。全量前端比较：allowlist 外新增/修改 **0**；payments/store、manifest/lockfile、root Vite/Vitest config 和其他现有测试均保持 baseline。

上轮 narrow revision 仅修改既有测试 `E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts` 的上述两个位置及本报告，与 immutable baseline 构造预期两处替换后逐字比较一致。本轮 Ops 修复只修改以下两文件及 own report；对修复前全量 frontend/src SHA256 比较，其他文件变化 **0**：
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/OpsSettingsDialog.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/__tests__/OpsSettingsDialog.loading.v0215.spec.ts

未修改独立 QA 报告、复现文件、QA runtime 配置、其他 owner、payments/store、依赖或 root 配置。

### Source
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/group/GroupRPMOverridesModal.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/group/GroupRateMultipliersModal.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/channel/PricingEntryCard.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/UsageTable.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/UsageCleanupDialog.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/account/ScheduledTestsPanel.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/monitor/MonitorTemplateManagerDialog.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/BaseDialog.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/Select.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/ProxySelector.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/AnnouncementBell.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/BackupView.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/OpsSettingsDialog.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/utils/pricing.ts
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/utils/format.ts
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/composables/useStepUp.ts
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/stores/announcements.ts
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/i18n/locales/en/admin/usage.ts
- E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/i18n/locales/zh/admin/usage.ts

### New tests / Coverage
以下测试名均对应当前工作区 `frontend/src` 下真实文件；以 git show 逐块适配选定行为，没有整文件覆盖本地 source。

| Commit / 行为 | 新测试绝对路径 | PASS |
| --- | --- | ---: |
| 7b1aa9314 RPM 空白/小数/负数不覆盖；0、整数、科学计数整数 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.edit.v0215.spec.ts | 6 |
| 8811c125e group 切换/重开旧响应、错误、loading、卸载；零倍率保存 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.v0215.spec.ts | 6 |
| 440784914 仅空条目首次单模型 autofill，不改已有/多模型/手工价格 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/channel/__tests__/PricingEntryCard.modelDefaultPrice.v0215.spec.ts | 4 |
| 5ace0699b 科学计数法数量级；保留本地两参数 API | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/utils/__tests__/pricing.formatScaled.v0215.spec.ts | 10 |
| 3a1a37b7b 小于 1 byte 和真实 ops byte-rate formatter | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/utils/__tests__/formatBytes.v0215.spec.ts | 5 |
| d5f490e10 en/zh 已加载 usage 文案、仅 applied 行标记；零账户/GroupID 定价倍率 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/__tests__/UsageTable.v0215.spec.ts | 3 |
| e0019da46 Escape 仅最上层；不可 Escape child、重开、scroll lock | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/__tests__/BaseDialog.escape.v0215.spec.ts | 3 |
| 8167d85b3 Select/ProxySelector 禁用关闭，再启用可打开 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/__tests__/disabledSelectors.v0215.spec.ts | 2 |
| 00a922c42 deferred initial load 在卸载后不重启 backup/restore polling | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/__tests__/BackupView.v0215.spec.ts | 2 |
| 8a495b0a5 cleanup 页码旧成功/失败、重开、关闭/卸载后失败 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/usage/__tests__/UsageCleanupDialog.v0215.spec.ts | 5 |
| 8f636513f scheduled plan 切换/折叠重开、旧错误/loading、关闭/卸载 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/account/__tests__/ScheduledTestsPanel.results.v0215.spec.ts | 6 |
| 4551fb0d8 template 重开旧成功/失败、loading、关闭/卸载 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/monitor/__tests__/MonitorTemplateManagerDialog.requests.v0215.spec.ts | 5 |
| 81bd85ec5 Ops load/save guard；QA fix：旧成功不能解锁失败新窗口、旧错误/finally 不污染新加载、卸载失效 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/__tests__/OpsSettingsDialog.loading.v0215.spec.ts | 6 |
| b07b1fb5d whole-minute sustained/cooldown；保留 fractional threshold | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/views/admin/ops/components/__tests__/OpsAlertRulesCard.duration.v0215.spec.ts | 3 |
| cee76030b 实际 Pinia store 失败不清 read 状态/不提示成功/不关闭；重试成功 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/common/__tests__/AnnouncementBell.v0215.spec.ts | 1 |
| c64145312 pending promise identity、全部 waiter、再次 prompt；mounted 并发 action verify/cancel | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/composables/__tests__/useStepUp.v0215.spec.ts | 13 |
| copied baseline 两个 group modal listener/debounce 卸载清理保留 | E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend/src/components/admin/group/__tests__/groupListenerCleanup.v0215.spec.ts | 2 |

## Executed Commands / Evidence
工作目录：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend`。所有 Vitest 使用指定外置 config（2 workers、cache=false、cacheDir 在 E runtime）。共享 node_modules junction 仅读取，未安装/改依赖。

上轮 narrow revision 执行的 88 + 19 测试集合与命令（历史 PASS）；最新 Ops 修复扩展到完整 128 + external 6 的命令见下面最新小节：
```powershell
$tests = @(
  'src/components/admin/group/__tests__/GroupRPMOverridesModal.edit.v0215.spec.ts',
  'src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.v0215.spec.ts',
  'src/components/admin/channel/__tests__/PricingEntryCard.modelDefaultPrice.v0215.spec.ts',
  'src/utils/__tests__/pricing.formatScaled.v0215.spec.ts',
  'src/utils/__tests__/formatBytes.v0215.spec.ts',
  'src/components/admin/usage/__tests__/UsageTable.v0215.spec.ts',
  'src/components/common/__tests__/BaseDialog.escape.v0215.spec.ts',
  'src/components/common/__tests__/disabledSelectors.v0215.spec.ts',
  'src/views/admin/__tests__/BackupView.v0215.spec.ts',
  'src/components/admin/usage/__tests__/UsageCleanupDialog.v0215.spec.ts',
  'src/components/admin/account/__tests__/ScheduledTestsPanel.results.v0215.spec.ts',
  'src/components/admin/monitor/__tests__/MonitorTemplateManagerDialog.requests.v0215.spec.ts',
  'src/views/admin/ops/components/__tests__/OpsSettingsDialog.loading.v0215.spec.ts',
  'src/views/admin/ops/components/__tests__/OpsAlertRulesCard.duration.v0215.spec.ts',
  'src/components/common/__tests__/AnnouncementBell.v0215.spec.ts',
  'src/composables/__tests__/useStepUp.v0215.spec.ts',
  'src/components/admin/group/__tests__/groupListenerCleanup.v0215.spec.ts',
  'src/components/account/__tests__/UsageProgressBar.spec.ts',
  'src/stores/__tests__/announcements.markAll.spec.ts',
  'src/views/admin/ops/components/__tests__/OpsSettingsDialog.spec.ts',
  'src/i18n/__tests__/announcementLocales.spec.ts'
)
node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs @tests
# exit 0; 21 files / 88 tests PASS (其中 baseline UsageProgressBar 6/6)
node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs src/components/admin/usage/__tests__/UsageTable.spec.ts
# exit 0; narrow expected/mock label 修订后全部 19 tests PASS
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
# exit 2; 仅已知 SDK TS2307，无新增错误
node node_modules/vite/bin/vite.js build --config E:/codex-runtime/upstream-v0215-20261009/frontend-worker/vite.config.mjs
# exit 1; baseline SDK unresolved (282 modules transformed)
```
- `git diff --check`：exit 0；仅已有 status.md 的 LF/CRLF 提示，无 diff whitespace 问题。
- 对 copied baseline 全量前端比较和逐块 source delta review：20 source / 17 new tests / 1 narrow existing test；非 allowlist delta 0。业务 GroupID 定价代码未改、nullish 零倍率保留，mounted 验证为 `0.0x`。两个 dirty group modal 的 listener removal 和 `clearTimeout(searchTimeout)` 均保留且有 mounted 验证。
- `BaseDialog` 本地 baseline 没有上游 `openDialogs` 集合，最小补入实例栈及 show/unmount 维护，不替换本地 UI/focus/listener 逻辑；避免子弹窗关闭时移除父弹窗 scroll lock。
- locale 只向现有 en/zh `admin/usage.ts` 各加两个 key，确认原 admin index 加载链；没有增加或使用未加载的 `resources.ts`。
- 首轮 73 tests PASS；补充测试首轮出现一个本地格式期待错误（0x vs 0.0x），已修正并以 88/88 重测关闭。上轮 narrow revision 重跑旧测试 19/19 和新/相关 88/88，均 exit 0。本轮 Ops fix 再次真实执行完整测试、typecheck 和生产 bundling，最新结果见下面；不以旧日志宣称 full PASS。

外置证据目录：`E:/codex-runtime/upstream-v0215-20261009/frontend-worker/`
- `vitest-final.log`、`typecheck-final.log`、`usage-legacy-final.log`、`build.log`、`build-bundling.log`
- narrow revision 历史：`usage-narrow-final.log`（19/19 PASS）、`vitest-narrow-final.log`（88/88 PASS）、`usage-narrow-delta.patch`（仅 mock label 和 marker expected 两处变化）。旧 `usage-legacy-final.log` 留作修订前失败历史，不是当前结果。
- `scope-check.json`、`frontend-delta.patch`、`source-baseline.json`、`tests.json`（初始 16 文件清单，最终精确集合以上面 21 文件为准）
- 临时 `vite.config.mjs` 的 outDir=`E:/codex-runtime/upstream-v0215-20261009/frontend-worker/dist`、cacheDir=`E:/codex-runtime/upstream-v0215-20261009/frontend-worker/vite-cache`；未写 backend dist 或 root config。

## Latest Ops QA Fix — 2026-10-10
对应原独立 QA Findings，不修改 QA 的 FAIL 裁决；等待原 QA retest。

### 实现及正式回归
- 每次 load 增加 `settingsRequestId`，成功赋值前校验代际、`props.show` 和 `disposed`，因此 runtime/email/advanced/thresholds 以及 `settingsLoaded=true` 都只能由当前请求写入。
- 旧失败在 catch 入口退出，不写 toast/log；finally 仅当前请求可结束 loading。关闭使用同步 show watcher 失效代际并清空 `settingsLoaded/loading`；onBeforeUnmount 同样失效并置 disposed。
- 保留原 3 条 Ops 测试，新增 3 条有明确名称的 mounted/deferred 正式回归：旧成功最后到达不能写 thresholds/解锁保存或调用任何 update API；旧失败不能弹 toast/结束当前 loading，当前成功后可保存正确配置；卸载后旧失败不能 toast/解锁保存。
- 先运行正式 6 例的修复前 RED：原 3 PASS、新 3 FAIL，exit 1，`ops-regression-red.log`。实际产品代码修复后同文件 **6/6 PASS**，包含在最终完整集合，不是跳例/过滤或替换复现。

### 实际命令及最终结果
工作目录仍为 `E:/codex-worktrees/sub2api/upstream-v0215-20261009/frontend`，无 `-t`、无 skip、无依赖安装：
```powershell
$tests = Get-Content E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/tests-final.json -Raw | ConvertFrom-Json
node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/vitest.config.mjs @tests
# exit 0; 28 files / 128 PASS (原 125 + 正式新增 3)，UsageTable 仍完整 19/19
node node_modules/vitest/vitest.mjs run --config E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-repro.config.mjs E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/ops-race.spec.ts E:/codex-runtime/upstream-v0215-20261009/qa-frontend-20261010/stepup-locale.spec.ts
# exit 0; 原外置 tests 保持未改，2 files / 6 PASS (Ops 3 + step-up/locale 3)
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
# exit 2; 仅已知 Airwallex SDK TS2307，新类型错误 0，full typecheck 未过
node node_modules/vite/bin/vite.js build --config E:/codex-runtime/upstream-v0215-20261009/frontend-worker/vite.config.mjs
# exit 1; 191 modules transformed 后缺失 Airwallex SDK，full build 未过
```

最新 own runtime 证据：
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-related.log`
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-external-final.log`
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-typecheck.log`
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-build.log`
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-scope.json`
- `E:/codex-runtime/upstream-v0215-20261009/frontend-worker/ops-fix-delta.patch`

外置 repro 首次接线失败（文件加载限制，0 tests、exit 1）留在 `ops-fix-external.log`，不算产品验证；只在 own 外置 `ops-repro.config.mjs` 加入精确 runtime/workspace fs.allow 后以未修改的原两份 QA specs 重跑 6/6 PASS。该 wrapper 只读继承 QA 模块 aliases，缓存写入 own runtime，2 workers/cache=false；未修改 QA config 或共享依赖。`ops-fix-exits.json` 保留首次执行 external=1 的历史，最新有效 external exit 0 单独记录在 `ops-fix-external-final.exit`。

## Risks / Recommendation
- **相关测试门禁已关闭**：独立 review PASS 后在扩展 allowlist 内只改旧 marker expected 和必要 mock label；全部 19 个既有用例及原 88 个新/相关用例均重跑 PASS。请独立 QA 重跑确认，不把 Developer 报告等同 QA 裁决。
- **full/生产构建及完整 typecheck 仍未验证通过**：已知 Airwallex SDK 缺依赖，未补包、未动 payments。独立 QA 可确认与 baseline 一致并判定 scoped QA；本报告不宣称全量 build/typecheck PASS。
- 未执行浏览器、真实 API、视觉、生产环境或部署验收；mounted/deferred tests 不能替代上述证据。
- 已实现 16 项、三条 Ops QA feedback 修复及完整相关测试可交付原独立 QA retest；未修改或替代原 QA FAIL 裁决。full build 限制继续单列风险，不声明全量验收通过或发版。
- knowledge_candidates: none。

绝对报告路径：`E:/codex-worktrees/sub2api/upstream-v0215-20261009/docs/workflow/worker-results/upstream-v0215-frontend-20261009-result.md`
