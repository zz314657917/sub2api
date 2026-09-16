# Sub2API 上游选择性迁移总计划

## 当前核对快照（2026-09-16，优先于下方历史阶段）

核对主线 `1e0ce5878`；未刷新远端，不将下方旧 upstream SHA 称为最新。

| 项目 | 当前证据 | 剩余门禁 |
| --- | --- | --- |
| S302 WS cap | `8a35e9540` 已提交 | 不重复迁移 |
| S303 会话保留 | 后端 `b675e7f3c`、前端 `ed91196f8`、专项回归 `1e0ce5878`；独立 QA PASS | 真实 DB/认证浏览器仍未验证 |
| Accept-Encoding | `d7a08661f` 已提交 | 不重复迁移 |
| Proxy list malformed | `frontend/src/api/admin/proxies.ts` 已有数组校验 | 不重复实现同等校验 |
| Claude max_tokens probe | `bb3dde9e4` 已提交 | 已知旧 strict-validator fixture 失败须单独归因 |
| MiniMax allowlist | 本地没有匹配 adapter，不能仅加名单 | 新平台范围另行批准 |
| Image cached billing / aliases | `23ddcc8d6` / `568fd9ced` | 真实 provider 验收未完成 |
| Native Images / account parity | `dc851ea3f` / `6ed89efe2`；独立 scoped QA | 真实图片生成/编辑与结算验收缺专用资源确认 |
| Claude message output_config | `d48fca7d3`；独立 scoped QA | 真实 provider 未验证，不阻止已批准本地合入 |
| Gemini finishReason | 缺上游 signal owner；本地未知结束原因默认 end_turn | 特定上游错误归因缺陷不适用，不导入另一套框架 |
| Firefox / CF challenge | `9eb120dd4` 尚未迁入；本地 privacy factory 仍为 Chrome | 独立设计及真实账号/测试环境确认后执行 |
| OpenCode | 长期新平台专项 | 架构、跨协议范围待明确批准 |
| Composite | P1 registry 已存在，提交 `b25cc223a`；独立 Terra QA 已补，定向测试/build/Ent 重放通过 | Wire 生成缺两个既有 provider，local QA BLOCKED；真实 DB、P2/P3 仍未通过/未授权 |

Image account / Claude 两项完整 service 回归均保留真实 `FAIL`：四个旧失败
已在干净 `dc851ea3f` 复现并取得精确例外审查。完整 JSON `Action=fail`
枚举替代截断日志判断；该结论不是全量测试绿，也不是发布批准。证据分别位于
`docs/workflow/integrations/image2-account-parity-20260916/` 与
`docs/workflow/integrations/claude-mid-output-config-20260916/`。

Composite P1 独立审计已完成：`go generate ./ent` exit 0 且没有生成漂移；
`go generate ./cmd/server` exit 1，缺 `[]service.ProxyRepository` 与
`service.newUserTrialConsumer`。定向 service/handler/routes 测试与 backend build
通过，但不能用编译成功替代 Wire 可重现生成。证据：
`docs/workflow/integrations/composite-p1-audit-20260916/`。未修改业务或生成代码。

当前下一步：单独确认 Wire 基线修复范围，不能在 P1 QA 合同中顺手修补；等待真实
provider/DB/认证浏览器测试资源及允许操作范围，不调用用户现有账号、不操作数据库/容器。
S303 本地验收证据：`docs/workflow/integrations/s303-regression-closure-20260916/`。
Firefox 资源设计：`docs/workflow/plans/openai-privacy-cf-resource-plan-20260916.md`。
下方“当前下一步”为初建计划时的历史文本，不再调度已完成的 S302/S303 业务迁移。

OpenCode 源码审计补充：缓存上游初始功能 `242907854` 连同后续
`7c008bd8c`、`981279c99`、`efcc2252e`、`22dffa1bb` 涉及平台、三协议
dispatch/model mapping、计费、session、402 与 quota。初始提交包含 Ent
schema/迁移和前后端管理入口；本地未找到 PlatformOpenCode/PlatformOpencode。
因此它不是可直接补齐的 allowlist，不能将“平台不存在”视为小修授权。
初始功能实际触达 104 个路径；上游迁移编号 238 与本地已有的
`238_usage_billing_settlement_consistency.sql` 冲突，必须重新设计本地迁移编号，
不得覆盖既有账务迁移文件。

## 总目标

在本地长期分叉且主工作树存在用户改动的前提下，持续从
`upstream/main` 按行为拆分、逐项验证并小步迁移。禁止整体 merge、rebase
或按提交历史批量 cherry-pick。

当前基线：本地 `main=8a35e9540`，上游 `upstream/main=bdb42e22f`，共同祖先
为 `18790386a76f12ae5721e557dc652c346ca699d5`。两边历史和文件拓扑差异很大，
只能采用本地拓扑上的手工适配。

## 阶段一：已完成

### S302 OpenAI WS 容量上限

已迁移 `956f4672e` 的行为：`mode_router_v2` 下正并发容量应用账号类型
系数并受全局硬上限约束，非正并发仍不可调度。定向测试、`go build ./...`、
格式和 diff 检查通过，提交为 `8a35e9540`。

## 阶段二：认证可靠性

### S303 JWT 临时用户查询错误

来源：`781a02aea`。

目标：区分 `ErrUserNotFound` 与数据库临时/内部错误，避免故障期间误报
`401` 或清除有效会话。先审查本地 middleware、token refresh client 和错误
契约，再分别覆盖后端和前端 owner。不得修改 token 格式、schema、session
存储或全局错误处理。

验收：后端 found/not-found/transient/internal 四类测试；涉及前端时增加
refresh 行为测试；运行定向 Go/Vitest、构建、typecheck、格式、精确路径和
冲突检查。真实数据库和认证浏览器作为独立 runtime 证据记录。

## 阶段三：低风险兼容队列

每项单独合同、单独 diff 和单独 QA，按以下顺序评估：

1. `213de0797`：`Accept-Encoding` header casing。
2. `188e3a9f9`：拒绝 malformed proxy-list response。
3. `4e5d67df3`：Claude Code `max_tokens=1` probe 兼容。
4. `5968fd0ed`：MiniMax monitor provider allowlist；先确认本地 adapter 和
   monitor registry 完整，再决定是否迁移。

这些任务不得混入支付、计费、模型目录、数据库迁移、部署或容器改动。

## 阶段四：运行态和拓扑专项

以下项目必须在独立设计和资源确认后处理：

- `9eb120dd4`：Firefox 指纹与 Cloudflare challenge 识别，需要真实 provider
  smoke。
- `8ea4dc56f`：Gemini `finishReason` 归因，需核对本地多层 provider 拓扑。
- `d8326fccf`：Claude mid-conversation output config，涉及完整 gateway 链路。
- `c0d511937`、`7ccc8a6f5`：Image 2.5/OAuth，涉及模型目录、计费和 provider。
- OpenCode 系列：新平台和跨协议链路，不能按小补丁处理。

## 阶段五：Composite 路由长期 Epic

Composite Gateway Selection 不进入短期队列。它依赖 Ent schema/migration、
repository、resolver、admin API、上下文传播和 Gateway dispatch。只有在
明确批准数据库与管理 API 范围后，才按持久化、resolver、dispatch、runtime
四个子阶段重新立项。

## 统一执行规则

- 每个 Sprint 使用独立 worktree、合同、合同审查、worker report 和 QA report。
- 保护主工作树现有 dirty 文件及 `.tmp-tutorial-live.html`、
  `pelican-bicycle.html`、`outputs/`、`sub2api`。
- Developer 与独立 QA 使用 `gpt-5.6-terra`；不可用时记录 `BLOCKED`。
- 本地代码 PASS 不等于真实 provider、数据库、容器或部署 PASS。
- 每个 Sprint 完成后由 Final Evaluator 决定 `PASS/FAIL/BLOCKED`，再进入下一项。

## 当前下一步

为 S303 建立合同并完成独立审查；S302 保持已提交状态，不回滚、不重做。
