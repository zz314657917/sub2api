# Gemini 价格、计费本地适配及上游缺失功能表

复核日期：2026-11-13，Asia/Shanghai。任务 ID 沿用开始日 `gemini-upstream-20261111`。
本报告固定比较 2026-11-11 已拉取的上游 `530c5e3a4f9c73ca2b0881191d0ef8e0d82c5484`
（包含 v0.2.21），不声称它是 2026-11-13 的最新远端提交。
原本地 HEAD 为 `6c6039ae6`；并行排行榜/Key 页面提交后，两次刷新组合输入，最终基于 `7252ea446` 合回。
本批 11 个产品/测试文件已合入本地工作区并完成复测；原有未提交改动按原字节核对保留。

## 本批适配范围

| 项目 | 具体行为 |
| --- | --- |
| Gemini 价格快照 | 新增确实缺失的 11 条标准模型记录，快照 250→261 条；旧记录及其他供应商原字节后缀保持不变 |
| 计费兜底 | 8 个新聊天/图片标识具备有限兜底；动态、自定义、渠道与分组价格继续优先；明确保留显式零价 |
| 图片价格 | 保留图片 token 价及每图元数据；新图片模型在无动态服务时使用对应单价，保留既有 1K/2K/4K 倍率与分组覆盖 |
| 同账号重试 | `gateway.max_same_account_retries_gemini` 默认 4，范围 0–10，不含首次请求；三入口共用，签名恢复两阶段独立于普通重试预算 |
| 非流式思考隔离 | Messages/Chat 不把布尔 `thought=true` 的文字混入正文，工具、图片和包含思考 token 的用量保持 |

Live/Embedding 三条仅补目录元数据，未新增 Realtime、Embedding 或音视频计费接口。
已有 Flash 3.6/3.7/3.8、Argon、URI 图片计数、输入 token 下限及断流结算不重复迁入。
模型发现候选列表与价格目录是两件事，本批没有同步前端/后端默认模型列表。

## 新增价格记录

下列数字来自固定上游代码快照，单位为 **USD/百万 token**；图片列为 **USD/张的目录基础价**。
它们用于说明本次代码对齐，不代表本轮独立核验了 Google 当前官方报价；不自动按日期调整价格。
破折号表示该字段未提供，不能解释为免费。Priority 的已有字段也随记录保留。

| 模型 | 输入 | 输出 | 缓存读 | 图片输出基础价 | 本次范围 |
| --- | ---: | ---: | ---: | ---: | --- |
| `gemini-3-flash` | 0.5 | 3 | 0.05 | — | 目录 + 有限兜底 |
| `gemini-3-pro-image` | 2 | 12 | — | 0.134 | 目录 + 有限兜底 |
| `gemini-3.1-flash-image` | 0.5 | 3 | — | 0.0672 | 目录 + 有限兜底 |
| `gemini-3.1-flash-lite-image` | 0.3 | 2.5 | — | 0.034 | 目录 + 有限兜底 |
| `gemini-3.1-flash-image-preview` | 0.5 | 3 | — | 0.0672 | 目录 + 有限兜底 |
| `gemini-3.1-flash-lite` | 0.25 | 1.5 | 0.025 | — | 目录 + 有限兜底 |
| `gemini-3.1-flash-lite-preview` | 0.25 | 1.5 | 0.025 | — | 目录 + 有限兜底 |
| `gemini-3.1-flash-live-preview` | 0.75 | 4.5 | — | — | 仅目录元数据 |
| `gemini-3.5-flash` | 1.5 | 9 | 0.15 | — | 目录 + 有限兜底 |
| `gemini-embedding-2` | 0.2 | 0 | — | — | 仅目录元数据 |
| `gemini-embedding-2-preview` | 0.2 | 0 | — | — | 仅目录元数据 |

## 上游有、本地仍缺失或仅部分具备的候选

以下项目只做源码筛选，尚未在本批实施。P1/P2 是建议顺序；涉及持久化、权限、收费模式的项目另立合同。

| 优先级 | 功能 | 本地差距 | 上游依据 |
| --- | --- | --- | --- |
| P1 | 刷新令牌原子轮换 | 当前 GET/DEL 分离，缺原子消费闭环 | `7122efee9` |
| P1 | 调度 hydrate 失败释放槽位 | 错误分支未调用已获得的 release | `4ac713c4d` |
| P1 | 价格同步 hash 锚点 | 正文 hash 不一致后仍可能保存远程锚点，阻碍后续重同步 | `92d35a694` |
| P1 | Sonnet 5.5 缓存读价 | 本地兜底/快照与固定上游仍有差价 | `508bb9560` |
| P1 | 第三方 Gemini / Vertex 配额区分 | 仍可能套用 AI Studio 模拟额度及午夜冷却语义 | `2588675ce` |
| P1 | Gemini 生图分组门禁 | 原生图像意图识别及映射后复核尚未对齐 | `47cb84b8b` |
| P1 | SSE/非流式读取上限 | Gemini 部分 ReadString、Antigravity 部分 ReadAll 尚无对应有界读取 | `d446c4a6b` |
| P1/P2 | Antigravity 协议兼容 | direct SSE、响应 JSON Schema 归一、工具媒体与并行结果顺序待补 | `c35418032` / `0b535b5c0` / `aae31c743` / `e2ce0b43d` |
| P2 | Gemini 显式缓存 | 缺 CRUD、绑定所有权、存储费、修改串行化/回滚恢复；需整组设计 | `d5a802f84` / `26b31a83e` / `43a3c197c` |
| P2 | Vertex API Key 与配额上限 | SA/模型查询已有，API Key 认证与 quota 闭环仍缺 | `81e6ec534` / `d6120b5d8` |
| P2 | 模型目录与映射并集 | 默认列表旧，多账号声明映射时的列表/详情并集缺失 | `31564b2e8` / `de44e3a0b` |
| P2 | Gemini 思考档位与流式摘要 | 已有签名/用量基础，缺完整预算/档位映射与摘要流 | `64211de8d` |
| P2 | Antigravity 真实配额窗口/主动查询 | 缺真实 5h/weekly 分组窗口、Pro Argon 行及绕缓存 Query | `bf4e16b90` / `455751402` / `4ac1675b5` |
| P1/P2 | OpenAI 缓存写入 TTL 分桶 | 总量已有，30m/12h 的数据库/计费/展示未贯通 | `fe093d76b` |
| P2 | Claude Fable/Mythos 5.2 | 缺专属目录、协议与兜底价整组适配 | `eb2da3448` / `3e8c2ee3f` |
| P2 | 账号刷新与设置保存体验 | 缺刷新保留表格/额度、设置页底部保存浮标 | `4a22a75b3` / `5965ac4e9` |

建议下一批先评估刷新令牌、槽位释放、价格同步、Sonnet 缓存读四个窄范围项目；
Gemini 则优先处理第三方配额、生图门禁、direct SSE 和 JSON Schema。

## Findings：最终结果

**本批本地源码适配 PASS，未提交、推送或部署。** 交付核验时间：2026-11-13T09:38:24+08:00。
独立开发、合同审核、QA 均使用矩阵指定的 `gpt-6.1-sol`，未替换模型。
未发现本批尚未解决的明确实现问题；保留一项已在原始基线复现的 unit 路由失败：
`TestAPIKeyResolveForModelRequestAllowsGeminiModelsOnOpenAIGroup`。
本轮没有修改该失败用例的断言。

Controller 的 Priority 显式零价账单探针曾发现价格标志丢失（账单 7.1），
Developer 已仅对本批有限 Gemini 模型修正，并加入实际计费回归；红/绿日志均保留。
旧错误策略夹具只补一行显式默认重试值 `4`，保持所有旧断言与其余字节不变，
该增补合同已独立审核通过。

## Executed Checks：实际验收

所有 Go 命令由 Controller 实际执行。worker 的本机文件工具被策略拒绝后，
按获批合同采用冻结源码包、补丁交付和真实日志审查；独立 QA 没有自行执行命令或计算哈希。
其 R4 实现审查与 R5 新组合输入补充审查均 PASS；下表“本地主目录”结果是 Controller 合回后实跑，
不冒充 worker 事后执行证据。旧 BLOCKED、R2 夹具失败和零价红日志保留。

| 检查 | 最终隔离组合 R5 | 合回后的本地主目录 |
| --- | --- | --- |
| 新增测试 default | 113 pass | 113 pass |
| 新增测试 unit | 113 pass | 113 pass |
| 相关回归 default | 883 pass | 883 pass |
| 相关回归 unit | 1099 pass + 同一既有路由失败 | 1099 pass + 同一既有路由失败 |
| 仅排除该精确既有失败后的 unit 余项 | 1099 pass，exit 0 | 1099 pass，exit 0 |
| `go build -mod=readonly ./...` | exit 0 | exit 0 |
| 10 个 Go 文件 `gofmt -l` | exit 0，空输出 | exit 0，空输出 |
| 任务文件 `git diff --check` | exit 0，空输出 | exit 0，空输出 |
| 源码及组合输入哈希 | 无漂移 | 11 个目标匹配，非目标无漂移 |

计数含父测试与子测试，未声称全库测试全绿。
原始基线 default 771 pass、unit 987 pass + 同一既有路由失败、build pass；
原始 unit 余项 987 pass。命令、退出码和完整事件日志保存在证据目录。

合回前再次比对原目标文件与主目录输入，候选输入共 4990 个；
合回后确认 HEAD `7252ea4469e12bff33a6d763d28aced6abbd515e`、逻辑暂存区以及全部冻结非目标文件保持，
没有额外源文件加入。价格目录仅新增 11 条，所有旧记录及原始文本后缀保持。

## Unverified Risks：边界

- 上述价格来自 2026-11-11 固定上游源码，本轮没有独立核验 Google 当前官方报价。
- 未调用真实 provider、生产数据库/Redis，也未重算历史账单或改变现有收费模式。
- 没有容器或服务更新；当前运行实例不会因源码合回自动生效。
- 没有运行浏览器或 `embed` 标签验收；并行 Key 页面和网页嵌入改动只按原字节保留，不纳入本批 UI 验收。
- Live/Embedding 只补目录元数据，不能据此认为对应端点已支持。其余缺失功能表是源码差异筛选，尚未实施。

## Recommendation：后续顺序

可继续使用本地源码构建；若要更新运行实例，需另行明确部署目标。
下一批建议先拆刷新令牌、槽位释放、价格同步、Sonnet 缓存读四个窄范围修复，
再处理 Gemini 第三方配额、生图门禁、协议兼容；显式缓存和 TTL 分桶另做设计与迁移评审。

## 文件与复核入口

| 产物 | 路径 |
| --- | --- |
| 本批合同 | [gemini-upstream-20261111.md](F:/mcplugins/sub2api/docs/workflow/tasks/gemini-upstream-20261111.md) |
| 独立合同审核 | [review.md](F:/mcplugins/sub2api/docs/workflow/contract-reviews/gemini-upstream-20261111-review.md) |
| 独立 QA 及主控交付记录 | [qa-report.md](F:/mcplugins/sub2api/docs/workflow/qa-reports/gemini-upstream-20261111.md) |
| 开发补丁交付记录 | [worker-result.md](F:/mcplugins/sub2api/docs/workflow/worker-results/gemini-upstream-20261111-result.md) |
| 详细源码差异表（含各项源码锚点、上游完整 SHA） | [audit-gap-table.md](E:/codex-runtime/sub2api/gemini-upstream-20261111/audit-gap-table.md) |
| 合回保护核对与最终测试证据 | [delivery-verification.json](E:/codex-runtime/sub2api/gemini-upstream-20261111/delivery-verification.json) |
| 全部运行证据 | `E:/codex-runtime/sub2api/gemini-upstream-20261111/` |

详细差异表保留审计当时状态，“实施中”和旧行号按历史快照理解；
本报告列出的已合入功能和最终文件哈希为交付事实。
sidecar 写出详细表后最终消息曾因网关 503 中断，未将该终止状态当作测试通过。

### 本批 11 个产品/测试文件

- `backend/internal/config/config.go`
- `backend/internal/config/gemini_retry_config_20261111_test.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_error_policy_single_check_test.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_pricing_20261111_test.go`
- `backend/internal/service/gemini_pricing_defaults.go`
- `backend/internal/service/gemini_retry_thought_20261111_test.go`
- `backend/internal/service/pricing_service.go`
- `backend/resources/model-pricing/model_prices_and_context_window.json`

### 最终证据指纹

| 产物 | SHA-256 |
| --- | --- |
| 合同 | `f621047f202b4a68e884413898379850aa765365f234e05d9573a933438e1158` |
| 任务补丁 | `31f29481adf68c90ec1cf2d678686ff358ef9a0c9eddc58f90421205243f14fc` |
| R5 组合冻结 | `42b93494e0f41a8a87566730b3376037806775dbc90500f75881aec4fec4f10b` |
| R5 QA 证据清单 | `2ded57ad5eeb84e6bb02619541e315fd862e8b3b0b02ea1bc482416649db2a58` |
| 合回保护记录 | `a4e033b64947d31d2055dfbac5823fe77344d670f56886bcc19a8bd588089063` |
| 最终主目录交付核对 | `6b6d00cd8cffb1382fb57669a3d446230abbe852fdc253224683c0f437eb9170` |
