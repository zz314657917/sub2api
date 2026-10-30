### PASS: pelican-radar-20261030

独立 Contract Review，2026-10-30。仅批准本 contract 进入实现；不代表实现、构建、运行态、浏览器或最终交付已经通过。未修改业务代码、其他任务文档、git index/history 或部署环境。

## Findings

未发现需要阻断实现的 contract 问题。当前版本已补齐稳定 read DTO、admin-only catalog route，以及关闭雷达时管理员显式读取目录的唯一 fetch 例外。末轮补审确认来源锚点修正为 `#stationmaster-round`；捕获首页确有该 ID，没有 `id="intelligence"`。

1. **评分来源与公式成立。** 本地捕获 JSON 为 `codex-four-task-results/1`、`codex-four-weighted-150-v2`，8 个模型、46 个 model/effort 组合、满分 150；实际权重 combined/pompeii/pelican/candy 分别为 60/45/22.5/22.5。独立计算全部 46 行，对照捕获首页 total panel 中同一精确 model/effort 的公开分数，46/46 一致，按首页两位小数精度比较无差异。combined 使用 final `score` 是必要条件：`gpt-6.1-sol/low` 的 `score=69`，但 `raw_score_total=72`。当前公式所得 106.165760869565... 对应首页 106.17；不能替换为 raw score 或 reward。
2. **日期、批次和状态语义可实现。** 源区间为 `2026-10-29T16:08:11.944316+08:00` 至 `2026-10-29T20:00:24.249331+08:00`，源快照为 `2026-10-29T20:00:25.692213+08:00`，`round_id=codex-four-261-10c-20261029`。源未提供每组合重复轮次数；contract 要求“来源批次”而非猜测次数正确。`gpt-6-luna/ultra` 和 `gpt-5.6-luna/ultra` 不在源 model_efforts，确应与未知模型/缺失行区分。
3. **缓存边界闭合。** 进程共享单实例、首次 singleflight、30 分钟 TTL、过期立即读旧值并独立刷新、失败保留快照及一分钟退避，足够约束实现。固定 HTTPS、禁止跳转、10 秒上限、2MiB 上限和禁止原始错误文本进入 DTO，避免把上游输入暴露成动态 URL/HTML/内部错误。配置顺序应从同一 immutable source snapshot 投影，不能把选择结果作为全局共享快照。
4. **旧设置兼容有明确可行路径。** 现有 `PelicanTestSettings` 被 `pelican_custom_settings_test.go:69` 直接 `!=` 比较；新增 slice 直接放入该结构会导致原测试编译失败。contract 的 optional pointer/object 允许 `Radar *...`，旧 JSON 与 defaults 均保留 nil，从而维持可比较性和原默认断言。旧 PUT omitted radar 保留原值，显式空 selections 清空，均可在现有 GetSettings/UpdateSettings 内局部完成，不需要 schema、wire 或其他业务模块变更。
5. **权限和路由归属已核实。** `backend/internal/server/routes/user.go` 当前注册 authenticated pelican GET，具备 JWT/BackendModeUserGuard/audit 链；`admin.go` 当前使用 AdminAuth 注册设置 GET/PUT。新增 reference read 应维持现有 `authorization` + service `allowUserAccess` 的 plaza-disabled/admin 语义；catalog 必须同时处于既有 admin 路由组并执行 handler admin guard。普通用户 read DTO 不含 prompt、settings、账号、分组或原始上游结构。
6. **范围及并行 dirty 可保留。** 当前抽查的 7 个已有任务目标，在 main 与候选上的 SHA256 全部一致：service、handler、user/admin route、frontend API/View/View spec。候选带入的其他 dirty prerequisites 不属于本任务修改权。contract 要求冻结目标并在集成前逐字节复核；发生漂移则停止精确覆盖，足以避免覆盖并行 gallery 或其他后端任务。

实现/QA 注意项（不增加业务范围，不作为额外阻断）：

- “disabled no-fetch” 的测试应明确针对普通 read；同时增加管理员显式 catalog 可 fetch、普通用户不能访问 catalog 的互补用例，避免把已批准的例外误判为失败。
- 缺失/null 分项不计总分，但 `candy.correct=false` 和 score 0 是有效值；用显式存在性/类型判断，不能用 truthiness。非法 task 对象类型、重复/冲突 model_names alias、声明之外的 row effort 必须属于 malformed snapshot 分支。
- 首次冷启动获取也应使用服务自身的有界 fetch context；单个等待者取消只终止自己的等待，不应让共享请求或其他等待者失败。建议分别覆盖冷启动和过期刷新，而非只验证后者。
- `PelicanReferenceScores` 请求在用户身份/角色切换或卸载后应失效，沿用现有 View 的 request-version/abort 风格；catalog 不可用时保留已配置与自定义 pairs，保存失败保留 draft。
- API DTO 的 `items`、catalog `models` 和 selections 建议输出 `[]`，不要意外输出 null；disabled DTO 不得触发 fetch，score 非 available 时必须为 null。
- 候选沿用的 dirty `go.mod` 要求 Go 1.26.10；当前 PATH 中的 Go 1.26.3 不满足。执行验收需显式使用现有 `E:/tools/go1.26.10/go/bin/go.exe` 并设置相应 GOROOT/PATH，而非降级 go.mod 或下载/修改依赖。

## Executed Checks

- 读取仓库 AGENTS、P/G/E 与 review-and-verification 技能、workflow status/matrix/spec、当前 contract。
- 读取捕获 JSON、首页及 source-ui，独立用 PowerShell 解析每行四分项并按精确 model/effort 匹配首页 total panel：`Rows=46, Compared=46, Mismatches=[]`。未运行或信任上游页面 JavaScript 来生成 oracle。
- 静态阅读现有 service settings/default/update/access/constructor、handler authorization/admin guard/settings、user/admin route 注册、frontend settings API/View 与相关旧 settings 测试。
- 检查候选 node_modules 中 Vite/Vitest 入口存在、`npm.cmd` 可解析；frontend package scripts 存在 test:run/typecheck，Vitest 能匹配新增 `.spec.ts`。进一步核实 dirty go.mod 要求 Go 1.26.10，且 `E:/tools/go1.26.10/go/bin/go.exe` 已存在；Go 验收必须采用该工具链。
- 读取候选 `git status --short`，确认存在与本任务无关的 dirty prerequisites；没有执行 stage/reset/clean。对上述 7 个既有目标做 main/candidate hash 比较，均一致。
- 本轮没有业务实现 diff，因此不声明 code diff、构建或测试通过。仅新增本评审文件。

评审绑定 SHA256：

| Artifact | SHA256 |
| --- | --- |
| `docs/workflow/tasks/pelican-radar-20261030.md` draft | `A527778E681F4189FFE274F75F67F9074E07EFC65E74A010651B9D7D1E2E95C0` |
| `E:/codex-runtime/sub2api/pelican-radar-20261030/source.json` | `200A64A9625FDC75CBE1FC36B1369CB56D2BA5DA55486D55902A23496D0CADE7` |
| `E:/codex-runtime/sub2api/pelican-radar-20261030/source-home.html` | `C502AD76DABC3F9CE37D9204F5BA71379C9BE636BDCF6E0F7CDB1E06EFAA8B92` |
| `E:/codex-runtime/sub2api/pelican-radar-20261030/source-ui.mjs` | `968A24558241B4778BEC68049E5F47EFFBB4923042A5C77E8BA980473B34B7DA` |
| `E:/codex-runtime/sub2api/pelican-radar-20261030/main-baseline.json` | `F9CB99B889FC2568084BDCF879B21FC0B83398C418B6AA127378A589B22AF44D` |

允许 controller 仅将该已审 draft 的 `status`/`review_verdict` 元数据改为 approved/PASS；业务 contract 内容变化需要再次独立评审。

## Unverified Risks

- 尚未执行实现后的 Go/Vue 测试、构建、race/并发验证、真实 handler smoke 或浏览器验收；这些均仍是后续独立 QA 门禁。
- 当前仅比较捕获的外部来源，没有重新联网抓取；远端后续数据变化必须经适配器严格验证，不得套用本次 46 行校验结果。
- main/candidate 的 7 个目标一致性是评审时点抽查，不替代实现前完整冻结、候选差异审查及集成前完整保护检查。
- 浏览器关闭及精确任务 PID/profile 清理、原画廊行为与布局保留，需要后续实证。

## Recommendation

可按当前 contract 批准并派发隔离 Developer；继续保留独立 QA 与最终 controller 集成复核。不得据本次 PASS 声称实现通过、真实供应商测试通过或授权 commit/push/container/deployment。
