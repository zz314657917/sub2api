# Sub2API 上游选择性迁移总计划

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
