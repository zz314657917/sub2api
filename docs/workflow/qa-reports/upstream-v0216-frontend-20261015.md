### PASS: upstream-v0216-frontend-20261015

2026-10-15，独立 QA 验收。依据主树 AGENTS.md、当前 Agent Matrix、已批准同任务 contract、独立 PASS contract review 及 Developer DONE 报告执行；与 Reviewer/Developer 分离，本轮自行执行全部验收命令，不以既有开发日志替代证据。当前 phase 为 qa，仅接受 forward recovery，不追认此前 provisional cherry-pick 的历史门禁。

## Findings

未发现本次 15 路径候选中需要阻断集成的明确问题。已审阅 base 到 working tree 的全部九个业务文件、六个新增测试，包含 Git diff 默认不包含的未跟踪 AccountGroupsCell 新测试。审阅内容保存在 `E:/codex-runtime/upstream-v0216-20261015/qa-reviewed-diff.log`。

- 三个支付倒计时均改为根据绝对截止时间计算剩余秒数；正常倒计时、暂停后的追赶及首次超时 tick 清理轮询均由真实组件测试覆盖。
- 倍率编辑使用 `Number.isFinite` 与 `< 0` 拒绝非有限值、负数；零倍率、正小数、科学计数及清空覆盖语义保留。新测还检查无效编辑不会覆盖合法未保存值。原 `GroupRateMultipliersModal.requests.v0215.spec.ts` 未修改，原有零倍率保存回归实际通过。
- Select 在 composition 期间直接返回；四个相关按键不拦截输入法，composition 结束后的 Enter 选择仍正常。
- Ops datetime-local 初始化使用本地时间格式化；原有 `toISOString()` API 提交路径保留，含跨日和 UTC 序列化断言。
- AccountGroupsCell 显式关闭截断并提供换行 classes；GroupBadge 其他调用默认 `truncate=true`。新测挂载真实组件，检查完整长名、默认截断、六组显示三组与 +3、body Teleport 全量弹层及 Escape 关闭。
- Gemini OAuth 复用既有 API 错误提取 helper，覆盖拦截器对象、旧 detail、空值 fallback 和失败后重试；没有引入新的 HTML 渲染方式。

日志存在非阻断警告：Browserslist 数据过旧、构建工具子进程 DEP0190 提示、静态/动态混合导入及大于 500 kB 的 chunk 提示。未为消除这些提示修改任何产品、依赖或锁文件。

## Executed Checks

候选根目录：`E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015`。测试、类型检查、lint、构建均在其 `frontend` 目录执行，Node 为 `D:/nobe/node.exe`。全部由本 QA 独立执行，无浏览器启动。

| 检查 | 实际结果 | 退出码 | 日志（统一位于 E:/codex-runtime/upstream-v0216-20261015/） |
| --- | --- | --- | --- |
| 合同必测 Vitest | 7 文件，41/41 PASS | 0 | qa-vitest-required.log |
| 相关回归 Vitest | 25 文件，139/139 PASS | 0 | qa-vitest-regression.log |
| 完整 vue-tsc --noEmit | PASS | 0 | qa-vue-tsc.log |
| 完整 Vite 生产构建 | PASS，按原配置实际生成候选 backend/internal/web/dist | 0 | qa-vite-build.log |
| 全部 15 路径 scoped ESLint | PASS，无 --fix | 0 | qa-eslint.log |
| git diff --check | working diff PASS | 0 | qa-diff-working.log |
| git diff --check BASE HEAD | committed diff PASS | 0 | qa-diff-committed.log |
| git diff --check BASE | base 到 working diff PASS | 0 | qa-diff-base-working.log |
| 未跟踪新增文件 whitespace 检查 | 无诊断，见下文原始退出码说明 | 1（no-index 文件差异） | qa-diff-untracked.log |
| 最终证据校验及 15 路径封存 | PASS | 0 | qa-finalize.log |

未跟踪文件执行的是 `git diff --no-index --check -- NUL frontend/src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts`。stdout/stderr 均为空；`--no-index` 对新增文件返回原始退出码 1，表示文件存在差异，并非 whitespace 失败。该结果按实际终端输出记录，未将其冒报为退出码 0。

必测命令：

```powershell
node node_modules/vitest/vitest.mjs run src/components/payment/__tests__/paymentCountdown.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts src/components/admin/group/__tests__/GroupRateMultipliersModal.requests.v0215.spec.ts src/components/common/__tests__/Select.ime.spec.ts src/views/admin/ops/components/__tests__/OpsDashboardHeader.timezone.spec.ts src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts src/composables/__tests__/useGeminiOAuth.errors.spec.ts --reporter=default --reporter=json --outputFile=E:/codex-runtime/upstream-v0216-20261015/qa-vitest-required.json
```

| 必测文件 | 实际测试数 |
| --- | ---: |
| paymentCountdown.spec.ts | 9 |
| GroupRateMultipliersModal.edit.spec.ts | 11 |
| GroupRateMultipliersModal.requests.v0215.spec.ts（未改） | 6 |
| Select.ime.spec.ts | 4 |
| OpsDashboardHeader.timezone.spec.ts | 3 |
| AccountGroupsCell.v0216.spec.ts | 3 |
| useGeminiOAuth.errors.spec.ts | 5 |

相关回归仅复用 `developer-vitest-regression-paths.txt` 的 25 路径参数，重新启动独立 Vitest 进程执行。包括支付组件、支付页面/API/store/i18n、disabledSelectors、Ops SLA 和 datetime-local 工具。完整展开命令保存在 QA 日志首部及 `qa-checks.json`，不是复用开发者运行结果。机器结果为 `qa-vitest-required.json`、`qa-vitest-regression.json`：所有文件测试数非零、所有 assertion 状态 passed；失败、skip/pending、todo 均为 0。两轮合计 180 次执行、171 个不同测试；paymentCountdown 的 9 项在两轮中重复。

其他命令：

```powershell
node node_modules/vue-tsc/bin/vue-tsc.js --noEmit
node node_modules/vite/bin/vite.js build
$env:NODE_PATH='E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/frontend/node_modules/.pnpm/node_modules'
node node_modules/eslint/bin/eslint.js <精确15路径，去掉frontend/前缀>
git diff --check
git diff --check de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d HEAD
git diff --check de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
```

ESLint 的全部展开参数见 `qa-eslint.log`、`qa-checks.json`，未执行 --fix。Vite 未覆盖 outDir，实际输出 `E:/codex-worktrees/sub2api/upstream-v0216-frontend-20261015/backend/internal/web/dist`；`git check-ignore` 确认其中 index.html 被忽略。该 index.html SHA-256 为 `216e0f8eb8bc03445c11a51e31a0bde39f36cf4252ef4cb61250d87a0dde7b1b`。构建资产留在候选忽略目录，不纳入封存产品清单。

### Scope、最终 Git 状态及字节保护

- 指定 base：`de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`。
- 最终候选 HEAD：`a19b132ec491ba99db319548c82f1e0edf68c393`，验收前后未变。
- base 到 working tree 的变更路径并集精确等于合同及主控 `integrate_frontend.py` 的 FILES 数组，共 15 条；只读取该脚本，未执行集成。
- 最终暂存变更列表为空。`git ls-files --stage -z` 内容 SHA-256：`2b371a00cd5abf01c1050611fc4422565d800bd842a2b32dc430059b91586544`，前后相同。
- 最终 working 修改仅 `frontend/src/components/admin/group/GroupRateMultipliersModal.vue` 和 `frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts`。
- 最终 untracked 列表仅 `frontend/src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts`。
- QA 前后全部 Git 可见候选文件字节哈希对照：`drift: []`；HEAD/index 均未变化。证据：`qa-before.json`、`qa-scope-before.json`、`qa-scope-after.json`。
- `backend/go.mod`、`backend/go.sum`、`frontend/package.json`、`frontend/pnpm-lock.yaml` 均逐字节等于指定 base 在当前 Git 属性下的检出结果，QA 前后也相同。`core.autocrlf=true` 下前三者为 CRLF，原始 Git blob 为 LF；报告明确区分 blob 与检出字节，没有声称两者 raw hash 相等。`git cat-file --filters BASE:path` 的基线检出哈希及原始 blob 哈希均保存于 scope JSON。
- 原 requests.v0215 测试既与基线 blob/检出字节一致，也在 QA 前后未变，SHA-256：`8755ece4f0748752b5bee0a6c40006d17eb29ec87578df37190cd9a76768867e`。
- 本 QA 未实现修复、修改产品/依赖/锁文件、暂存、commit、部署或派子代理；主树只新增本验收报告，其余输出位于指定外置 runtime。

### 15 路径 SHA-256 封存

manifest：`E:/codex-runtime/upstream-v0216-20261015/qa-sealed-files.json`，严格为 `{repo_relative_path: lowercase_sha256}` 字典，没有混入元数据；15 个文件的封存哈希与独立测试前捕获值一致。

manifest 文件本身 SHA-256：

```text
46d500c35b56db9450f7a530c6f9d9de83821287ad07ee95229ab88d7e5b53f1
```

封存集合：

```text
frontend/src/components/payment/PaymentQRDialog.vue
frontend/src/components/payment/PaymentStatusPanel.vue
frontend/src/views/user/PaymentQRCodeView.vue
frontend/src/components/payment/__tests__/paymentCountdown.spec.ts
frontend/src/components/admin/group/GroupRateMultipliersModal.vue
frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.edit.spec.ts
frontend/src/components/common/Select.vue
frontend/src/components/common/__tests__/Select.ime.spec.ts
frontend/src/views/admin/ops/components/OpsDashboardHeader.vue
frontend/src/views/admin/ops/components/__tests__/OpsDashboardHeader.timezone.spec.ts
frontend/src/components/account/AccountGroupsCell.vue
frontend/src/components/common/GroupBadge.vue
frontend/src/components/account/__tests__/AccountGroupsCell.v0216.spec.ts
frontend/src/composables/useGeminiOAuth.ts
frontend/src/composables/__tests__/useGeminiOAuth.errors.spec.ts
```

封存汇总与退出码：`E:/codex-runtime/upstream-v0216-20261015/qa-seal-summary.json`；可复查执行脚本：同目录 `qa-run.cjs`、`qa-finalize.cjs`。这些是外置 QA 证据文件，不属于产品 15 路径。

## Unverified Risks

- 未启动浏览器，未做实际宽度、换行、行高、响应式或真实输入法验收。jsdom 的文本、class、Teleport 和 KeyboardEvent 断言不构成真实浏览器或布局证据。
- 未做真实支付、支付 provider、Gemini OAuth provider 或生产 API/数据库集成验收；测试使用本地 mocks/fake timers，构建成功不能替代真实支付生命周期。
- 未执行主树集成、提交或部署。本 PASS 只适用于上述候选字节与本合同前端范围，不代表另行阻塞的 gateway 批次或生产发布通过。

## Recommendation

PASS：本合同要求的独立 QA 门禁通过，交主控按现有授权流程消费封存清单。主控执行外置 integrate_frontend.py 前应核对本 manifest hash、每个候选文件 hash 和目标基线保护；仅集成封存 15 路径，不复制 Vite 生成资产。候选发生任何字节变化后，本封存结论不覆盖新内容。

本 QA 已完成报告与封存并停手，不扩大验证、不重复测试、不实施集成或产品修改。
