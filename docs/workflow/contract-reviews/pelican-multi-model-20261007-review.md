### PASS: pelican-multi-model-20261007

评审日期：2026-10-07。角色：独立 Contract Reviewer。当前工具会话 `x-codex-turn-metadata.model` 确认为 `gpt-6.1-sol`，满足仓库 Agent Matrix；未使用替代模型。

审阅合同：`docs/workflow/tasks/pelican-multi-model-20261007.md`。只读核对工作区：`E:/codex-worktrees/sub2api/pelican-multi-model-20261007`，HEAD 为 `10e542353d3a6ddb8c26996e6d35cf6122278329`，`git status --short` 无输出。没有修改业务代码、调用供应商或操作数据库。

## 结论与门禁

合同的目标、成功标准、允许路径、禁止路径、隔离运行要求和验收证据足够明确，技术方案在现有 Pelican 架构内可实现，无需扩展 provider adapter、依赖或安全边界。没有阻断性合同修订。

本 PASS 仅批准合同的可行性，不代表实现或运行验收通过。合同目前仍为 `status: draft` / `review_verdict: pending`；主控须批准 frontmatter 后才可启动 Developer。共享 `docs/workflow/status.md` 仍记录旧批次 `phase: done`，主控须在调度前更新当前任务与 phase，不能把历史状态当作当前开发门禁。

## 源码核对与实现约束

1. **模型存储和兼容可行。** `PelicanPlan` / `PelicanPlanInput` 目前使用单一 `ModelID`；repository 的列投影、扫描、Create/Update 和 Claim 都在允许文件中。`modelsForGroup` 已提供本地模型目录和现有分组/平台资格校验。新字段须维护列表顺序并逐模型校验，所有返回路径统一提供 `model_ids` 与首项 `model_id`。显式无效列表不可回退到旧字段；仅省略新字段才允许旧 `model_id` 输入。迁移只加字段、回填和必要索引/约束，不合并或删除计划/历史。数据库类型选择属于实现细节。

2. **串行调度和租约可行，但须同时修改两个预算。** service `run`（708–775 行）当前只运行一次，整体 deadline 为 `timeout + 60s`；repository `Claim`（129 行）租约为 `timeout + 120s`。不能仅把单模型调用包进循环。按模型数扩展整体 deadline 与租约，并给选择、结果保存、finalize 留明确有界余量；租约应长于整体 deadline 和必要收尾余量。每模型 helper 在返回时立即取消该模型 context、释放 `ReleaseFunc`，不能把释放 defer 放在整个轮次函数中。模型超时不应取消下一模型；父 context 取消则停止后续模型。每轮使用已 claim 的不可变模型快照和同一 prompt snapshot。

3. **配额和失败暂停可沿用现有机制。** `ReserveAttempt`（149–166 行）原子地检查 generation、有效租约和上海日边界，再增加 `daily_calls_used` / `round_attempted`；`Finalize` 与 abandoned-round recovery 已按 `round_successes` / `round_attempted` 结算。每个模型在资格通过后、调用 tester 前保留一次 attempt，失败不退额度；资格 skip 不调用 reservation。额度耗尽或 reservation 异常不可继续发起供应商调用。保留“有成功归零、尝试但全失败只加一次、全 skip 不加”的轮次语义，`estimated_calls_per_run` 按配置模型数返回，不沿用现有账户可调度时才返回 1 的估算。

4. **历史和保留必须覆盖全部查询。** 现有 results 已有 `model_id`，不需重写旧结果。repository `ListGallery`（229–230 行）的 latest、history count、成功 fallback，目前仅按 plan/group；`SaveResult`（375–377 行）的立即保留、`cleanupSQL`（419–430 行）的计数/天数保留和成功保护，也仅按 plan/group。以上全部增加 model 维度，不能只修改 gallery。`ListHistory`（275–311 行）新增可选 model 条件且放在 `LIMIT 50` 前，保留现有授权和 group join。移除模型不会即时删除结果，历史接口不应因模型已不在当前列表而拒绝读取；常规保留/手动清理继续按合同规定执行。

5. **分组分页可行，必须保留筛选次序。** 当前 gallery 对 plan/group 取 latest，再按 `account_id` 过滤（224–230 行），不是“先筛账户再取该账户最新”。多模型后应对 plan/group/model 取 latest，再用现有账户过滤语义构建可见候选集，然后按 group 聚合、排序、分页，最后返回选中 group 的候选子项；不得为筛选命中的 group 额外返回不匹配账户的子项。`total` 是候选 group 数；组排序使用候选子项最大 `finished_at` 再 group ID，子项使用固定 plan 顺序和配置模型顺序。同模型来自不同 plan 不合并。旧计划不因 `enabled=false` 被隐藏，现有展示本就包含禁用计划；“active plan/model entries”在此指现存计划当前选择的模型，并继续遵循既有分组状态授权规则。

6. **前端有兼容扩展入口。** 现有 `ModelIcon` 位于 `frontend/src/components/common/ModelIcon.vue`，接受 `model` 属性并已有未知模型 fallback，允许直接复用无需修改它。`PelicanPlansPanel.vue` 的 edit/watch/loadModels 已有 stale-response 保护，多选改造须保留版本和 group 检查。`PelicanTestsView.vue` 的 card key、refresh 时选中项检查目前只使用 plan/group（45、270 行），必须补 model 身份；组卡以 group ID 标识，子条目至少以 plan/group/model 标识，history 请求传当前 model。保留 result-ID artwork cache 和用户/授权切换清理，避免按 group 或 model 单独缓存 HTML。

7. **展示和安全范围明确。** 图标、加粗模型名、clock、时间 tooltip 和独立 duration 都可在允许 view/panel 或任务专用展示 helper 内完成。展示旧成功 fallback 时，应明确区分当前轮次状态/时间与实际展示作品的生成时间，避免把失败时间标成旧作品生成时间。维持现有 preview、安全 sandbox、replay/enlarge/focus 和 580px/180px 约束；不改 sanitizer 或 iframe 安全实现。

## 验收可执行性

现有 service、repository 和 handler Pelican 测试文件均为普通 Go test 文件，所抽查文件没有特殊 build tag。可从 `backend` 执行 `go build ./internal/service ./internal/repository ./internal/handler`，并执行 `go test ./internal/service ./internal/repository ./internal/handler -run Pelican -count=1`；实际既有 fixture 编译阻断应单独记录，不得为了测试通过修改无关 owner。

repository 的现有 PostgreSQL 用例在缺少 `PELICAN_TEST_POSTGRES_DSN` 时会 skip；成本控制用例另外使用 `PELICAN_COST_TEST_DSN`。QA 必须指向任务独占临时数据库，并确认新增验收用例实际运行，不能把 skip 当 PostgreSQL PASS。

frontend 已有 `test:run`、`typecheck`、`build` 脚本，Pelican 测试集中在 `src/components/pelican/__tests__`。可从 `frontend` 执行 `npm.cmd run test:run -- src/components/pelican/__tests__`、`npm.cmd run typecheck`、`npm.cmd run build`，用不带 `--fix` 的 scoped ESLint 验证。浏览器 fixture 须补多计划同组同模型、移除模型后历史可读、账户筛选不泄漏子项、组分页不拆分和移动端堆叠，并证明任务 profile/PID/daemon 清理完成。

本评审只核对源码和脚本入口，没有执行编译、测试、SQL 或浏览器验收。上述实现注意事项均源自合同成功标准和现有源码，不增加业务范围；最终是否满足由独立 QA 和主控 Final Evaluator 判断。
