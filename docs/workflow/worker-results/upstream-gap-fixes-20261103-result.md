### DONE: upstream-gap-fixes-20261103

## Findings

六项窄修复已完成，可交给 Controller 冻结并执行独立 QA。此报告是 Developer 交付，不是独立 QA 或集成批准。

- Contract SHA256：`b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`，指定 independent native review 为 PASS。
- 候选目录：`E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103`。
- HEAD：`7ea675c020ff2c5d5e4124ac5e249c9dc65034c5`，未提交。
- 所有本报告引用的日志、receipt、overlay 和冻结证据均在 `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103`。原始审计保留在相邻 `upstream-gap-audit-20261103`。
- 源码已于 `2026-11-03T04:03:04.469Z` 冻结，文件为 `developer-source-freeze.json`；冻结后仅继续只读验证和编写本报告。

## Changed Files

以下 8 个源码/测试路径和本报告为本 worker 的全部仓库改动：

1. `backend/internal/repository/usage_billing_repo.go`
2. `backend/internal/repository/usage_billing_deleted_account_test.go`（新增）
3. `backend/internal/service/openai_images.go`
4. `backend/internal/service/subscription_service.go`
5. `backend/internal/service/openai_gateway_service.go`
6. `backend/internal/service/upstream_gap_fixes_test.go`（新增）
7. `frontend/src/views/user/KeysView.vue`
8. `frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts`（新增）
9. `docs/workflow/worker-results/upstream-gap-fixes-20261103-result.md`（本报告）

`backend/internal/service/account_stats_pricing.go` 未修改，使用其现有可选参数。候选原有 `docs/workflow/agent-matrix.md` 修改保持原样，不属于本 worker 的改动。

## Implementation

1. 上游账号计数仅忽略 `ErrAccountNotFound`。余额/订阅结算、share ledger 顺序、幂等重试保留；真实数据库错误仍回滚。APIKey 与 Bedrock、存在/删除/数据库错误均有 SQLmock 场景。
2. 非流式图片不再从请求 `n` 初始化计费张数。实际返回一/二张分别计一/二张，无法计数时保留单张兜底；没有引入零收费策略。流式和拆分逻辑未修改。
3. 管理员 `ExtendSubscription` 在既有 `withSubscriptionUpdateTx` 内 `GetByIDForUpdate`，锁内计算及更新 expiry/status。成功提交后才失效缓存并普通回读。测试通过真实 ent transaction、SQLmock `SELECT ... FOR UPDATE` 和 SQL UPDATE，验证上下文、提交/回滚、缓存时序、失败不失效、两路并发 +5 累积 +10，以及无 entClient 的兼容入口。测试替身的行锁仅用于模拟数据库锁生命周期，生产代码没有新增全局锁。
4. 配额暂停只在没有有效 reset 时按旧快照自愈。绝对 reset 优先，相等即已到期；相对 reset 固定锚定 `codex_usage_updated_at`，不会每次检查重新延后。5h/7d、未来/过去/相等/非法/缺失 reset 和 update time 都有断言。
5. OpenAI `RecordUsage` 向既有账号统计管线传入账号 gate 与模型 OAuth 豁免的合并禁用标记，只影响模型文件价的长上下文阶梯；自定义规则、ApplyPricingToAccountStats、其它平台调用保持原优先级/签名。resolver 有无、账号/分组开关组合及 OAuth 特定模型均有控制项。账号关闭时统计 `.78`，开启时 `1.545`；客户费用按原有售价链路计算。无 resolver 的旧入口原本仅用账号 gate 决定客户阶梯，本批保持该行为。
6. 重置额度在 await 前捕获目标，使用实际返回的 `quota_used/status` 更新目标；当前选中 ID 仍相同时才更新编辑表单状态。实际 mounted KeysView 覆盖非零返回额度、切换/清空选中项、不激活状态、失败不突变、额度/速率用量分离，以及后续保存第二次 API 调用中的 `active`。

## Executed Checks

所有 Go 命令使用 `E:/tools/go1.26.10/go/bin/go.exe`，子进程环境为 `GOROOT=E:/tools/go1.26.10/go`、`GOTOOLCHAIN=local`，并设置 `GOPROXY=off`、`GOSUMDB=off`，均为 `-mod=readonly`。前端使用 `D:/nobe/node.exe` 及已安装 CLI，没有安装依赖。表内每个日志前缀对应 `.stdout.log`、`.stderr.log`、`.receipt.json`；receipt 保存完整可执行文件、参数、cwd、起止时间和实际退出码。

| 证据前缀 | 实际检查 | 退出码与结果 |
| --- | --- | --- |
| `baseline-related-unit-r3`（Controller 原始证据） | 未修改候选的原始相关 unit 基线 | exit 1；158 叶子通过、3 个既有失败，保留原证据 |
| `developer-new-tests-baseline` | `go test -mod=readonly -tags unit ./internal/service ./internal/repository -run '^TestGapFix' -count=1 -json`，生产源码修复前 | exit 1；28 失败、40 控制项通过 |
| `developer-final-tests-original-overlay` | 同上，加 `-overlay .../developer-baseline-overlay.json`，最终新测试对原 HEAD 生产源码 | exit 1；27 失败、41 控制项通过 |
| `developer-new-tests-candidate-r1` | 新增 Go 测试首轮候选 | exit 1；59 通过、9 个订阅测试夹具失败，详见下文；未删除负向断言 |
| `developer-new-tests-candidate-r2` | 相同新增 Go 定向命令，修正夹具后 | exit 0；68 叶子全部通过，无跳过 |
| `developer-acceptance-unit` | 合同第一条 Go 命令，加 `-json` | exit 1；41 叶子通过、2 个既有 quota API 失败，随后 panic，未声称全部测试执行 |
| `developer-acceptance-default` | 合同第二条 Go 命令，加 `-json` | exit 0；76 叶子通过，无跳过 |
| `developer-related-unit` | 59 个明确测试声明的精确正则，`go test -mod=readonly -tags unit ./internal/service -run <regex> -count=1 -json` | exit 0；59 叶子通过，无跳过 |
| `developer-backend-build` | `go build -mod=readonly ./...` | exit 0 |
| `developer-frontend-new-baseline` | 初版新测试对原组件 | exit 1；0 tests，i18n mock 缺少 createI18n 导出，属于收集/夹具错误，不能作为行为 RED |
| `developer-frontend-new-baseline-r2` | 修正部分 mock 后，对原组件运行新 mounted spec | exit 1；7 行为失败、1 控制项通过 |
| `developer-frontend-original-overlay` | `vitest run --config .../developer-baseline-vitest.config.mjs <三份 KeysView specs>`，最终测试和相同依赖，用 pre-load 插件加载原 HEAD KeysView | exit 1；30 通过、8 失败，其中 7 为新增行为 RED，1 为既有 create-query 失败 |
| `developer-frontend-tests-candidate-r1` | 三份合同规定的 KeysView specs | exit 1；36 通过、2 失败，其中新增后续保存夹具缺少必填分组；保留第二次 API 保存断言后补齐有效表单 |
| `developer-frontend-tests-candidate-r2` | `node node_modules/vitest/vitest.mjs run --cache=false src/views/user/__tests__/KeysView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/views/user/__tests__/KeysView.quotaReset.spec.ts` | exit 1；37 通过、1 个与原组件相同的既有 create-query 失败；新增 mounted spec 8/8、KeysView.spec 9/9 |
| `developer-frontend-typecheck` | `node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` | exit 0 |
| `developer-frontend-eslint` | `node node_modules/eslint/bin/eslint.js src/views/user/KeysView.vue src/views/user/__tests__/KeysView.quotaReset.spec.ts` | exit 2；严格 pnpm 布局无法从项目配置解析已安装的 vue-eslint-parser |
| `developer-frontend-eslint-installed-parser` | 同一命令，仅加进程级 NODE_PATH 指向现有 parser 的 node_modules | exit 0 |
| `developer-frontend-build` | `node node_modules/vite/bin/vite.js build --outDir E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/frontend-dist` | exit 0；产物仅输出到 runtime |
| `developer-diff-check` | `git diff --check -- <全部九个绑定源码/测试路径>` | exit 0 |
| `developer-gofmt-check` | 固定工具链 `gofmt -l <绑定的 Go 文件>` | exit 0，stdout 空；修改的 Go 文件已执行 gofmt -w |

定向既有回归的全部实际测试名和精确正则保存在 `developer-related-test-inventory.json`，覆盖 ForwardImages（含 GPTImage2 多图拆分、部分失败、APIKey streaming JSON/SSE、断开后排空）、BuildOpenAIImagesResponsesRequest、ResolveAccountStatsCost、TryModelFilePricing、AssignOrExtendSubscription 及 OAuth passthrough stream/compact 协议控制。Controller 的更大独立库存不作为本 worker 自测结果。

### Baseline findings and harness corrections

- 原始三项 Go 基线失败是 `TestCalculateCostWithLongContext_PreservesAndSplitsImageInputCost`、`TestOpenAIQuotaQueryUsageSendsCodexHeadersAndUsesEagerProxy`、`TestOpenAIQuotaServiceNotConfigured`。最后一项在 `openai_quota_service.go:187` nil dereference panic。合同第一条正则不命中第一项，但会遇到后两项及 panic；默认标签不包含这些 unit-only 失败。未修复或修改其 owner 文件及旧断言。
- Vue 旧失败为 `KeysView create query > does not overwrite a touched default group when rates load after the dialog opens`，预期 2、实际 null。候选和原 HEAD 组件的相同依赖/三文件运行均复现。
- Go r1 九个订阅失败是测试夹具问题：ent 实际 UPDATE 先写 updated_at，初版 SQL 预期次序不符；Ristretto MaxCost=10 不足以容纳实际条目。最终夹具增加容量并断言确实入缓存，按真实 SQL 次序验证，同时增加实际 SELECT FOR UPDATE 预期；原有回滚、提交前缓存存在、提交后失效断言均保留。
- 最终 Go 原始源码 overlay 的测试计数与最早 RED 相差一项，因为新测试最初误把无 resolver 的客户售价入口也当成受 group gate 约束；读取既有生产逻辑后，将该新控制项改为保持原有客户售价行为。旧测试从未修改。原始日志原样保留。
- 初次前端收集失败、首轮候选的有效表单缺失、首轮 lint 解析错误均单独记录，没有算作已修复生产缺陷或 PASS。

### Installed ESLint environment

精确环境记录：`developer-eslint-environment.json`。ESLint `8.57.1`、vue-eslint-parser `9.4.3`，均可在 `frontend/pnpm-lock.yaml` 查到对应版本。成功命令使用的子进程变量为：

```text
NODE_PATH=F:\mcplugins\sub2api\frontend\node_modules\.pnpm\vue-eslint-parser@9.4.3_esl_70a7b1a6f4ce77f7d80c7906f9f9f2e6\node_modules
```

没有改变环境持久配置、dependency declarations、lockfiles 或 ESLint 配置。

## Source SHA256

`developer-source-freeze.json` 绑定所有 9 个允许源码/测试路径。`developer-scope-review.json` 确认冻结后一致、无白名单外原始文件变化、主仓库已绑定的既有源码与 main-before 一致。

| 路径 | SHA256 |
| --- | --- |
| backend/internal/repository/usage_billing_repo.go | `4a01690416403c25713e02d638bce8897c02ff6160071c1d895357c8c215b111` |
| backend/internal/repository/usage_billing_deleted_account_test.go | `595ea8de66352ce633b77ec7c42ab0f24a954125c3af3e4e633453722725de8e` |
| backend/internal/service/openai_images.go | `384b054c35991ad90ae7acaaf34f118681d8339e3158626bb93f881251768d93` |
| backend/internal/service/subscription_service.go | `bf305852650a40ce783a4fc722e5a6d3ff4e934a07220fea3f0db1a575af996d` |
| backend/internal/service/openai_gateway_service.go | `1d325b359d3931d8e8b30d41eb76354fc1896f8b6276b911bd5bbad3f7fe20d9` |
| backend/internal/service/account_stats_pricing.go（未修改） | `4eabb069d8a736ce825892c9e69b2d499f63d0e40ca59d3ff97230ec4d8df4d4` |
| backend/internal/service/upstream_gap_fixes_test.go | `5352fdb1350c96fe5aa0bc68df39c47781909ddf8099f1b44f21da0d802b1293` |
| frontend/src/views/user/KeysView.vue | `27489944f4a59349261866f9e690d5f7fabaa45344fff856464cad9dc58bd0a6` |
| frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts | `2f2d3029c640ff603fdd8294f50e2c2f27d2bf2c74228c49bd28af2fd6219c5b` |

## Contract Compliance

- 在指定 baseline receipt 存在且收到 Controller GO 后才新增/修改源码，保留原始失败。
- 仓库写入仅位于指定隔离目录的允许路径；主仓库产品源码未修改。未触碰原有矩阵变更、知识库、依赖声明/锁、数据库 schema/migrations、配置、Pelican 源码。
- 没有启动其他 agent，没有安装依赖、调用 provider/在线数据库、提交、推送、合并、重置、清理、部署或操作容器。
- 未实现第七项 image-price unset 语义，未修改价格 resolver/billing-service 或旧断言。
- 尽量复用既有接口：账号统计 helper 保持原文件字节；管理员复用已有事务；配额 reset 提取解析 helper 供既有判断及暂停逻辑共用。

## Unverified Risks

- 广域 unit 与三份 Vue 合同命令有已确认基线失败，不能声称全套通过；panic 后未执行的覆盖由单独精确命令补证。
- 未进行真实 PostgreSQL 行锁/并发负载、Redis、provider、HTTP 服务启动、真实浏览器或线上验收。SQLmock 的事务/锁替身和 jsdom mounted 组件证据有此边界。
- 前端构建有既有 Browserslist 过期、静态/动态 import 混用和 chunk size 提示；本轮没有下载更新或改分包。
- 独立 QA 和主仓库集成由 Controller 另行执行，冻结后主仓库/候选若有变化必须重新绑定对应字节。

## Recommendation

交付冻结候选执行独立 QA 与 Controller 精确回归。不得将 DONE 理解为发布、部署或生产数据库/服务商验收通过。

## knowledge_candidates

本 worker 不写知识库。候选供 Controller 验证后决定是否沉淀：未来 quota reset 应优先于 stale snapshot 自愈；管理员绝对到期时间调整必须在数据库行锁事务中读改写；API Key 异步重置应捕获目标并消费服务端实际返回值。

## R2: Approved regression expectation alignment

本节仅追加 R2 记录，上方 R1 报告及历史证据保留。本轮严格按独立 PASS 的补充合同对齐两处既有测试，没有进一步修改任何生产或 R1 新增测试。

- 补充合同 SHA256：`a51b65f646f67a0a38dfda32bbdd63b2a1729c07b17ab5f20618cb79c8a9bb5c`。执行前已核对原始字节、指定独立 review 首行 PASS 及其哈希绑定。
- 父合同仍为 `b731e0c443ccc0b41f5e85f3a1b4fe8cc6ef0e09db8e3fa4b992540c84c72ce0`。
- R1 九个路径持续匹配 `candidate-freeze.json` 和 `developer-source-freeze.json`；未修改这九个文件及两个原始 freeze。
- R2 仓库改动仅为下面两处既有测试及本报告追加。Controller 负责正式 R2 freeze、独立 QA 副本和集成，本 worker 的 `developer-r2-source-hashes.json` 仅提供十一条路径的当前字节身份。

### R2 changed paths and exact scope

1. `backend/internal/service/openai_account_scheduler_test.go`：保留 `TestOpenAIGatewayService_SelectAccountForModelWithExclusions_StaleUsageSnapshotSkipsPause_Issue2994` 顶层名称、issue 背景、实际选号断言、35901/35902 ID 与 0/5 优先级。拆为 fresh per-case fixture：`without_known_future_reset_self_heals` 选择 35901，`with_known_future_reset_stays_paused` 选择 35902；仅修改直接矛盾的 reset 说明。相邻 fresh-exhausted、reset-reached 等测试不变。
2. `backend/internal/service/openai_subscription_long_context_test.go`：SettlementMatrix 的账号统计阶梯条件仅增加 `&& accountGate`，并更新对应两行说明；客户收费、凭据、resolver、group、tier、缓存、结算和所有其他断言、容差、循环及 case 名保持原样。
3. `docs/workflow/worker-results/upstream-gap-fixes-20261103-result.md`：仅追加本节。

两个测试最终 SHA256：

| 路径 | SHA256 |
| --- | --- |
| backend/internal/service/openai_account_scheduler_test.go | `b6fa14d686afe4702a5fcb5365b1903d54774633b255a4d7a8cf1004b9a9ca5c` |
| backend/internal/service/openai_subscription_long_context_test.go | `2302fb4c3c39d166c59f554822e633e163de3602774de8db0964b92461e1e51c` |

### R2 executed commands and actual exits

证据根目录仍为 `E:/codex-runtime/sub2api/upstream-gap-fixes-20261103`。所有新日志以 `developer-r2-` 开头，原始 FAIL、R1 报告 preimage、原始 freeze 均保留。每个测试日志前缀具有 `.stdout.log`、`.stderr.log` 和完整参数/实际退出码的 `.receipt.json`。

工具链仍为 `E:/tools/go1.26.10/go/bin/go.exe`；Go 子进程设置 `GOROOT=E:/tools/go1.26.10/go`、`GOTOOLCHAIN=local`、`GOPROXY=off`、`GOSUMDB=off`，未持久化环境或安装依赖。

定向正则 `FOCUSED` 为：

```text
^(TestOpenAIGatewayService_SelectAccountForModelWithExclusions_StaleUsageSnapshotSkipsPause_Issue2994|TestOpenAISubscriptionGPT61SettlementMatrix)$
```

| 日志前缀 | backend 下实际测试命令 | 实际结果 |
| --- | --- | --- |
| `developer-r2-red-unit` | `go test -mod=readonly -overlay E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/developer-r2-original-production-overlay.json -tags unit ./internal/service -run FOCUSED -count=1 -json` | exit 1；273 通过、17 个预期失败、0 skip |
| `developer-r2-red-default` | 上条去掉 `-tags unit` | exit 1；273 通过、同样 17 个预期失败、0 skip |
| `developer-r2-green-unit` | `go test -mod=readonly -tags unit ./internal/service -run FOCUSED -count=1 -json` | exit 0；290/290 叶子通过、0 skip |
| `developer-r2-green-default` | 上条去掉 `-tags unit` | exit 0；290/290 叶子通过、0 skip |
| `developer-r2-regression-unit` | `go test -mod=readonly -tags unit ./internal/service -run <existing-regression-regex.txt 的精确原内容> -count=1 -json` | exit 0；671/671 叶子通过、248 个顶层终态、0 skip |
| `developer-r2-regression-default` | 上条去掉 `-tags unit` | exit 0；620/620 叶子通过、207 个顶层终态、0 skip |

完整参数以对应 receipt 为准。`developer-r2-red-results.json`、`developer-r2-green-results.json`、`developer-r2-regression-results.json` 保存实际日志解析结果。

- RED overlay 只替换五个原生产路径，不替换测试；映射目标逐一核对原 HEAD 内容，未改 main 或 QA baseline。两标签的失败集合均精确为：future-reset 子例错误选中 35901 的 1 项，加 APIKey/account=false/n=272001、resolver/group/tier/cache 组合的 16 项账号统计失败。无 reset 自愈、account-on、OAuth 等控制项保留并通过。
- 定向矩阵共有 288 个叶子，加 scheduler 两个叶子共 290。第一次汇总脚本误算组合数、检查时错误期待 561 个通过，从而在两份真实 exit-1 receipt 生成后自身退出 1；没有改写或重复执行原 RED 测试来隐藏这一点。随后从原日志确认正确结果为 273 通过/17 失败，并在 `developer-r2-red-results.json` 记录该统计脚本错误。
- `developer-r2-inventory.json` 核对原选择的 20 个文件和全部 248 个原 Test 顶层声明、原顺序及精确回归正则哈希。新增 scheduler 子例使叶子计数准确增加一项，没有删除、改名、skip/xfail、减少循环或扩大容差。
- `developer-r2-gofmt-write`：固定工具链仅格式化两处允许测试，exit 0。
- `developer-r2-gofmt-check`：`gofmt -l` 仅检查两处测试，exit 0，stdout 空。
- `developer-r2-diff-check`：`git diff --check -- backend/internal/service/openai_account_scheduler_test.go backend/internal/service/openai_subscription_long_context_test.go`，exit 0。已逐行检查精确两测试 diff。

### R2 preservation and remaining limits

`developer-r2-pre-edit.json` 与 `developer-r2-report-preimage.md` 保留本轮修改前的测试哈希、R1 九路径及原报告。`developer-r2-source-hashes.json` 绑定 R1 九路径加本轮两处测试的十一条路径，整体仍为十个实际变化的源码/测试路径。冻结、原日志及原报告前缀的保全复核另见 `developer-r2-preservation.json`。

本轮无额外生产/测试修改，无其他 agent、安装、网络/provider/在线 DB、主仓库写入、Git staging/commit/push/history 操作或部署。原四项独立归因限制继续存在：长上下文图片输入成本旧失败、两个 quota API unit 失败（含 panic）、Vue create-query 旧失败；本轮固定库存全绿不代表父合同广域 Go/Vue 全套绿色，也没有重新裁定这些限制。

Developer R2 交付完成，可由 Controller 正式冻结并交独立 QA。尚无真实数据库/provider/浏览器或生产运行态验收结论。
