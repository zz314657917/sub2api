### PASS: key-fast-policy-20261105

本判定仅覆盖冻结前端增量及下述证据，不代表后端、整项任务、生产环境或部署通过。独立 QA 与 Developer 分离；运行时 `x-codex-turn-metadata.model` 实际为 `gpt-6.1-sol`，符合 Agent Matrix。QA thread 为 `01a1a970-e9ec-7bb0-adc4-0fc5b6803cc7`，身份取证见 `qa-role-provenance.json`。

- 候选：`E:/codex-worktrees/sub2api/key-fast-policy-20261105`。
- 证据根目录：`E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend`。以下未写绝对路径的证据文件均位于此目录。
- 最终批准合同 R3：`7003aeb62b2186516058dabc306c61d16edda2d6f83d26336b958afc76497e61`。
- 独立审查：`docs/workflow/contract-reviews/key-fast-policy-20261105-review.md`，round 3，PASS；本轮读取到的报告 SHA256 为 `5d8446eb1964214f4f994c07893f38dc431ec2dd04336b177b99bb764434b35c`。

## Findings

未发现本次前端增量导致的新失败或阻断问题。完整定向验收仍有 **2 项既有失败，不能称为全绿**；本轮独立复现、保留断言并核对了归因：

1. `frontend/src/views/user/__tests__/GroupBuyView.spec.ts:263`：
   `GroupBuyView switches to the Pixel Cafe shell when the opt-in flag is enabled`。
   原有 mount 缺少嵌套 `ConfirmDialog.vue:39` 的 i18n 安装，抛出
   ``SyntaxError: Need to install with `app.use` function``，未到达页面文案断言。
2. `frontend/src/views/user/__tests__/KeysView.createQuery.spec.ts:445`：
   `KeysView create query does not overwrite a touched default group when rates load after the dialog opens`。
   原有异步默认分组场景在第 465 行期望 `group_id=2`，实际为 `null`。
   不是 create 参数数量或新增 Fast 字段断言失败。

候选为 404 项、402 过/2 失败；独立原源码完整 overlay 为 369 项、367 过/同 2 失败；独立原源码双文件 overlay 为 26 项、24 过/同 2 失败。三次结果的失败文件、完整用例名、完整消息及堆栈逐项相同，见 `failure-comparison.json`。没有删除、修改、跳过或弱化任何旧测试。

需保留的执行及输入说明：

- 第一次直接调用 ESLint JS 入口 exit 2，原因是绕过 pnpm `.bin/eslint.CMD` 所设置的 `NODE_PATH`，导致 `vue-eslint-parser` 无法解析；按现有启动器的原样 `NODE_PATH` 重跑，exit 0。未安装依赖、替换 parser 或调整 lint 配置。原始失败日志与成功日志分别保存在 `eslint.log`、`eslint-native-env.log`。
- 候选 12 个非目标前端文件与主目录 `before.json` 存在换行字节差异；它们全部精确匹配 `candidate-before.json`，CRLF 归一化后与主目录原文件一致，属于候选创建时已有差异。它们在 QA 期间未改变，不属于本任务交付，见 `preexisting-candidate-line-endings.json`。
- 初次主目录 4928 项保护核验有 `spec.md`、`status.md` 两份已登记的 Controller 文档差异；二者精确匹配当时 `controller-docs.json`。末次核验又观察到 Controller 将主目录 `status.md` 更新为 `phase: qa`，实际 SHA256 为 `ce948f70c284596e5bbbf345c8a742cad9fb4d8b282fcdea98ca60414d18ed13`，同样精确匹配更新后的 `controller-docs.json`。原始差异保留，未宣称所有主目录文件原始字节均未变化；没有主目录业务文件变化，QA 未写入这些文档。

## Executed Checks

### 合同、冻结及范围

R2 批准合同归档 `E:/codex-runtime/sub2api/key-fast-policy-20261105/contract-r2-approved.md` 的完整 SHA256 为：

`f54335d5d9cc8199c6b4de1b52aa57d8193de94e19b31c6442f1894d48b1cbce`

本轮独立核对 R3 原始字节及 round-3 PASS 绑定，并在内存中仅移除新增的后端旧测试 Allowed Path 和三行版本断言约束，重建 SHA256 精确等于 R2。R3 唯一新增授权是后端 `openai_policy_group_fast_20261103_test.go` 对应 snapshot version 断言 20→21；前端要求、范围、源码和 dist 均未变。因此 **manifest 所载 R2 前端源数据合法沿用到 R3 最终批准合同**，不需要重写 Developer freeze manifest。本 QA 没有修改该 manifest。

原 Developer manifest：

`E:/codex-runtime/sub2api/key-fast-policy-20261105/developer-frontend/frontend-freeze.json`

| 冻结对象 | 数量 | SHA256 |
| --- | ---: | --- |
| 原 manifest 原始字节 | 1 | `bbf2334ba576922396474536a9a54eb4696811675ebbc078ce3fd27acf94fc40` |
| 七源码条目集合 | 7 | `78b9876f5ffa97a776b23c852de7f5806eea59a8ec7d612e0f6b340fdeace4ce` |
| dist 条目集合 | 259 | `2f2e93c6288572429e7760f9463ba79bcade7f5e6c451990c0b0fee1294a6b93` |

集合哈希按 manifest 定义，计算对应有序条目数组 `JSON.stringify` 的 UTF-8 SHA256；本轮同时逐一验证了全部 266 个原始文件的大小和 SHA256，不只是校验集合文本。

| 冻结源码，候选根目录相对路径 | 原始文件 SHA256 |
| --- | --- |
| `frontend/src/types/index.ts` | `03869d9e97907c8e2c2d21bfb89b88018b820cc828c6a538b5d8f24faa84c786` |
| `frontend/src/api/keys.ts` | `19070bc8d74d2d0347593dcb2e3a8d1d6a60d00205bbd565609706329024c34b` |
| `frontend/src/views/user/KeysView.vue` | `80e9fc5a891fb659ff3c6a257cfa931134bcc0a545de3a4d46f4780fad8e9767` |
| `frontend/src/i18n/locales/zh/keys.ts` | `d436cf0a95a86edc34f7850e3139cdffa28b6116fe0abf9f207b6a862f688377` |
| `frontend/src/i18n/locales/en/keys.ts` | `01e2e1dba3a3d58a7aaf9dc2db5e268ff1cdf81c30391a5d5ea7e56171b1e54f` |
| `frontend/src/api/__tests__/keyFastPolicy.spec.ts` | `6b4a038b9a0cb4c66e5f9f2177fa37f061e79b821a91840549eb419b6ae78059` |
| `frontend/src/views/user/__tests__/keyFastPolicy.spec.ts` | `0d3ea59106683d59ff53419ee72d0f3def0e2bcc54db10befcc5a7124ee0a83a` |

核验结果：

- 开始、结束两轮七源码和 259 dist 原始字节全部匹配；dist 无额外文件。
- 对 1048 个原有前端输入核验，只有五个批准产品文件相对候选初始清单改变；两份新增测试属于批准路径。QA 期间原有前端输入零漂移，全部原有测试未改。
- 逐行审阅五产品文件 diff 和两份新测试：精确三值 union、create 追加第十二参数并保留旧十一参数/default omission、update 未变省略及显式回到 passthrough、旧响应 fallback、create/edit/save/reset、中文精确标签和对应英文说明均有实际实现。
- 新 mounted 测试保留真实 Select，点击选项并提交表单；API 测试保留 Axios JSON transform，只拦截 transport。不是纯 source-text 断言。涵盖保存后回显、失败后重试和托管密钥拒绝后不走备用写接口。
- 主目录只读核验 4928 个原跟踪文件；4926 个匹配 pre-task 原始字节，另两份 Controller 文档按上文对账。HEAD 保持 `0f6b739881ccf9e4ad423f464aeedcc9abd7f48a`，逻辑 index 保持不变。全量文件记录见 `main-protection.json`，后续并发文档变化见 `freeze-after.json`。
- `frontend/node_modules` 仍为指向 `F:/mcplugins/sub2api/frontend/node_modules` 的既有 junction。没有安装依赖。Vitest 全部加 `--no-cache`，未借共享缓存出具结果。

主控后续发布的完整候选 `E:/codex-runtime/sub2api/key-fast-policy-20261105/candidate-freeze.json` 也已绑定：

| 完整候选冻结对象 | 数量 | SHA256 |
| --- | ---: | --- |
| manifest 原始字节 | 1 | `b04c489046bdf948697558ee7a84d1841e624fa65fd6657183abd3cfbbdef9a7` |
| 完整产品 source map | 39 | `5226bfffd299cfdd4f17875d2acb0dae3e51fba2cfee713c0c31f33c3c6fa923` |
| 完整编译 input map | 3826 | `8eb2d607d846b8059df6fe917e145e7db57c0ebc52c49b0fad2c85cc4ce8688d` |

独立重算两个完整 map 的 `JSON.stringify` 哈希，核对其 R3 合同哈希，并实际逐文件验证其中 **1050 个前端编译输入**；它们与本轮测试时的 1048 个原有前端输入加两份新测试完全一致。七个前端 source 同时匹配 Developer 七文件冻结、完整 source map 和完整 input map。无前端漂移，证据为 `candidate-freeze-binding.json`；本轮没有将完整 map 的元数据校验冒充为对 32 个后端 source 或其他后端输入的独立文件/功能验收。

### 独立自跑结果

下表每行是一次独立执行，重叠测试不累计为不重复覆盖数。各次测试 pending/todo 均为 0。

| 检查 | 实际结果 | 证据 |
| --- | --- | --- |
| 新 mounted + Axios/API 测试 | 35/35，mounted 20、API 15；exit 0 | `new-tests.log`、`new-tests.json` |
| 既有 Group Fast 前端回归 | 26/26，mounted 14、group/type 12；exit 0 | `group-fast.log`、`group-fast.json` |
| 合同完整前端定向集 | 34 文件、404 项，402 过/2 旧失败；exit 1 | `acceptance.log`、`acceptance.json` |
| 原源码完整 baseline overlay | 32 文件、369 项，367 过/相同 2 失败；exit 1 | `baseline-full-overlay.log`、`baseline-full-overlay.json` |
| 原源码双文件 baseline overlay | 2 文件、26 项，24 过/相同 2 失败；exit 1 | `baseline-pair-overlay.log`、`baseline-pair-overlay.json` |
| `vue-tsc --noEmit` | exit 0 | `vue-tsc.log` |
| 七文件 scoped ESLint，原启动器环境 | exit 0 | `eslint-native-env.log` |
| 独立 Vite production build | exit 0、259 文件；全部与 Developer dist 逐字节一致 | `vite-build.log`、`independent-build-comparison.json` |
| scoped `git diff --check` | exit 0 | `diff-check.log`、`frontend-tracked.diff` |

独立 baseline 的五个产品源文件均由 `git show 0f6b739881ccf9e4ad423f464aeedcc9abd7f48a:<path>` 直接提取到证据目录，逐个验证原始 SHA256 等于 `before.json` 后，才通过 QA 自写 Vite pre-load plugin 在内存加载。没有把旧源码写回候选，没有修改原测试。完整 baseline 配置只排除本次新增 mounted 测试以还原原有 369 项，不排除任何旧测试；双文件重跑保留原有 26 项。加载命中包括 KeysView、keys API 及两种语言 keys locale；类型模块不需运行时加载。见 `baseline-inputs.json`、`baseline-overlay.config.mjs`、`baseline-overlay-loads.json`。

实际命令的工作目录为候选 `frontend`。使用已经安装的 Node 可执行文件：

```powershell
$qaNode = 'C:/Users/Administrator/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node.exe'
& $qaNode ./node_modules/vitest/vitest.mjs run src/views/user/__tests__/keyFastPolicy.spec.ts src/api/__tests__/keyFastPolicy.spec.ts --no-cache --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/new-tests.json
& $qaNode ./node_modules/vitest/vitest.mjs run src/views/admin/__tests__/GroupsView.openaiFast.spec.ts src/views/admin/__tests__/groupsOpenAIFast.spec.ts --no-cache --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/group-fast.json
& $qaNode ./node_modules/vitest/vitest.mjs run src/views/user/__tests__ src/api/__tests__/keyFastPolicy.spec.ts src/views/admin/__tests__/GroupsView.openaiFast.spec.ts src/views/admin/__tests__/groupsOpenAIFast.spec.ts --no-cache --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/acceptance.json
& $qaNode ./node_modules/vitest/vitest.mjs run --config E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/baseline-overlay.config.mjs --no-cache --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/baseline-full-overlay.json
& $qaNode ./node_modules/vitest/vitest.mjs run --config E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/baseline-overlay.config.mjs src/views/user/__tests__/KeysView.createQuery.spec.ts src/views/user/__tests__/GroupBuyView.spec.ts --no-cache --reporter=default --reporter=json --outputFile.json=E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/baseline-pair-overlay.json
& $qaNode ./node_modules/vue-tsc/bin/vue-tsc.js --noEmit
& $qaNode ./node_modules/vite/bin/vite.js build --outDir E:/codex-runtime/sub2api/key-fast-policy-20261105/qa-frontend/dist
```

ESLint 首次直调及成功重跑使用同一参数；成功重跑的子进程 `NODE_PATH` 逐字取自现有 `.bin/eslint.CMD`：

```powershell
$env:NODE_PATH = 'F:\mcplugins\sub2api\frontend\node_modules\.pnpm\eslint@8.57.1_supports-color@7.2.0\node_modules\eslint\node_modules;F:\mcplugins\sub2api\frontend\node_modules\.pnpm\eslint@8.57.1_supports-color@7.2.0\node_modules;F:\mcplugins\sub2api\frontend\node_modules\.pnpm\node_modules'
& $qaNode ./node_modules/eslint/bin/eslint.js src/types/index.ts src/api/keys.ts src/views/user/KeysView.vue src/i18n/locales/zh/keys.ts src/i18n/locales/en/keys.ts src/api/__tests__/keyFastPolicy.spec.ts src/views/user/__tests__/keyFastPolicy.spec.ts
```

实际执行通过 `spawn(nodeExe, args, {cwd, env, windowsHide:true})`，不是全局修改环境；原始绝对 argv、PID、开始时间和退出码见各 `*-execution.json` 及 `executed-commands.json`。候选根目录另执行：

```powershell
git --no-optional-locks diff --check -- frontend/src/types/index.ts frontend/src/api/keys.ts frontend/src/views/user/KeysView.vue frontend/src/i18n/locales/zh/keys.ts frontend/src/i18n/locales/en/keys.ts frontend/src/api/__tests__/keyFastPolicy.spec.ts frontend/src/views/user/__tests__/keyFastPolicy.spec.ts
```

独立构建只写 `qa-frontend/dist`，没有覆盖主控正在验收的外部 `key-fast-policy-20261105/dist` 或候选/主目录的产品 dist。Vite 保留现有大 chunk 警告，构建 exit 0。

### Controller 浏览器证据：核验和归属

**以下浏览器操作由 Controller 执行；不是本 QA 自跑浏览器。** 本 QA 没有启动浏览器或重开 fixture，独立读取原始日志/脚本、核对 profile/PID/fixture 路径、抽看截图并封存副本。

- 原始 `E:/codex-runtime/sub2api/key-fast-policy-20261105/browser-smoke-r3.log` 顶部 JSON 为 PASS；独立解析确认恰有 25 条 check，全部通过，`errors=[]`、`consoleErrors=[]`。
- 检查涵盖三态 create 默认/保存/edit 回显，force→filter→passthrough，刷新保留，priority/分组与全局优先级/WS 重连说明，以及 1440×1000、390×844 的 light/dark 四组合可见性和稳定后无页面横向 overflow。
- 已抽看 `desktop-light-r3.png`、`mobile-dark-r3.png`；四张 R3 截图均已封存。页面使用冻结 production dist；fixture `preview-server.json` 的 root 为 `E:/codex-runtime/sub2api/key-fast-policy-20261105/dist`，PID 44448、端口 62116。fixture 是 loopback 内存合成 CRUD，不是真实后端。
- R1 原始 FAIL 保留：resize 后侧栏动画瞬态触发 mobile-light overflow。R2 原始 timeout 保留：继承手机 viewport 后等待桌面 row 可见。R3 脚本显式设置初始桌面 viewport，等待文档宽度稳定及已有动画结束，原 overflow 和 CRUD 检查仍在，未删断言。
- session `key-fast-policy-20261105`；独立 profile `E:/codex-runtime/sub2api/key-fast-policy-20261105/browser-profile`；已读取进程记录确认 daemon 33296、browser 59416 的命令行明确归属此 session/profile。
- `browser-close.log` 记录 session 已关闭；最终 `browser-cleanup-final.json` 的 task browser、daemon、fixture 数组均为空且 `passed=true`。不采用首次 cleanup 的自匹配记录作为最终结论。
- 本 QA 另自跑只读 `check-owned-processes.ps1`：精确 profile 对应 Node/Chrome/Edge 进程和已记录 PID 33296/44448 均为零，见 `independent-browser-cleanup-check.json`；未终止任何进程。

全部导入证据保存在 `controller-browser/`，原文件路径、原始字节 SHA256 和归属在 `controller-browser-evidence.json`，保留 Controller 与独立 QA 的执行边界。

另独立核对主控结构化 `E:/codex-runtime/sub2api/key-fast-policy-20261105/browser-result.json`，原始 SHA256 为 `35627f7197f49742f6a1a13f1cd25bea002c21deb9c7997116e08381b8286a59`。其中 25 条 check 与 R3 原始日志逐项一致，cleanup 与最终清理 JSON 完全一致；script/log SHA、Developer dist freeze manifest SHA 及四张截图原始 SHA 全部相符，见 `browser-result-binding.json`。四截图哈希如下：

| Controller R3 截图 | SHA256 |
| --- | --- |
| desktop-light | `81218f6a87b6b1080e7a4b75cc0ed838e4a40a5837c95e9e596aa58080936c88` |
| desktop-dark | `a1616ad44da43da5d9d055667e3c18b23b484c2985a1f1610f074ead490b4ff7` |
| mobile-light | `019e04b9d9be521616ec3e9a596aa6fda29191540334b74ef1632c85a210964d` |
| mobile-dark | `16a98eea54383882d25b2784a4913080b90e95ef9437ad04b88bff6ceb6d2bf0` |

## Unverified Risks

- 后端仍由另一独立 QA 验收：持久化、所有权、Cafe 托管限制的真实服务行为、鉴权缓存 21、HTTP/WS 最终 tier/metadata/计费等，不由本前端报告授予通过。主控提及的 RawChat terminal 基线失败未在本轮执行或判定。
- 浏览器仅证明合成 CRUD fixture 下的实际构建页；没有连真实 provider、账户、数据库或迁移。25 条 check 和截图不等于完整键盘/屏幕阅读器无障碍专项认证。
- 两项旧前端失败仍存在，已独立证明不由本增量引入。本 PASS 不修复、不忽略它们，也不宣称原有大范围测试全部通过。
- 本报告只审当前冻结源及输入；Controller 后续若改变前端源、依赖或 dist，必须重新冻结并对受影响项重验。

## Recommendation

**前端范围 PASS，可提交给 Controller 合并后端独立验收结果并进行最终裁决。** 只允许使用上述精确七源码增量及其冻结证据；R3 后端最小补充不使 R2 前端冻结失效。主控应保留两项原有前端失败记录，并继续履行最终输入/hash、非任务文件保护和后端验收门禁。

QA 写入仅为本报告及指定外部 `qa-frontend` 目录中的脚本、基线提取副本、日志、构建和证据副本。没有改任何业务源码、原有测试、Developer manifest、原 dist、知识库、status/spec、依赖、主目录文件；没有 commit/push、provider 调用、数据库或容器操作。限定前端写集未发现越界；另行观察到的 Controller 流程文档变化已按原始 hash 记录和对账。
