---
status: approved
verdict: PASS
task_id: upstream-v0215-ws-oauth-20261009
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
reviewer: current-session-default
last_verified: 2026-10-10
---
### PASS: upstream-v0215-ws-oauth-20261009

## Findings
- 未发现必须修订的合同问题。PASS 仅指合同可执行，不指实现完成或计费/运行态验收通过。
- 本地 handler 已有 BeforeRequest/BeforeTurn/AfterTurn 与异步 RecordUsage 入口，APIKeyService.GetByKey 经过现有 auth cache；无须引入缺失的 upstream profit-gate 架构或改数据库。
- HTTP reasoning 重试、OAuth transform 和 passthrough owners 均在本地 monolithic `openai_gateway_service.go` 等 allowlist 路径；新增 `openai_oauth_web_search_history.go` 是明确授权的新 helper，不是结构缺失 blocker。
- relay Usage、解析累加和 adapter 的 turn/session 映射都在 allowlist；本地 relay 与上游 merge/finalize 拓扑有差异，应做行为适配而非复制整文件。
- 与 protocol/frontend 实际业务 owner、约定测试前缀无重叠；保留其他 worker owners，不创建同时匹配其他合同测试模式的混合职责文件。

## Executed Checks
- 读取 spec、三合同、isolated status/matrix；检查六个上游 commit 的路径及 73381fc9c、b6a157c19、a4ea8059b 具体 diff。
- 核对本地 handler hooks、GetByKey 缓存实现、HTTP bounded retry、OAuth compact/Lite 入口、relay usage 和 adapter 传播位置。
- HEAD 等于合同 base_commit；6 个现存精确业务路径 SHA256 均等于外置 `baseline.json`。
- 读取主控 `E:/codex-runtime/upstream-v0215-20261009/baseline-checks.md`。包测试、生产 build PASS 为主控证据；reviewer 未重跑 baseline、调用 provider/DB 或做完整安全/计费审计。

## Gates / Unverified Risks
- 本 PASS 后由主控批准 draft/pending frontmatter；用户授权本轮 current-session-default native agents 已写入 spec，不改全局矩阵。
- 新 handler/service 测试名必须含 `V0215`，明确非零 test counts。service baseline fixture 漂移允许外置精确 compile overlay；必须记录原错误、完整替换内容、命令和测试选择，不能修改/屏蔽本轮测试或替换生产实现来伪造通过。
- 建议生产构建外置：`go build -o E:/codex-runtime/upstream-v0215-20261009/ws-oauth-server.exe ./cmd/server`。baseline 编译漂移不授权修全库或进入 denied paths。
- WS 必测首轮无 BeforeTurn 的 fallback、同 key/group/platform/subscription 才 refresh、lookup error/身份变化回连接快照、Before/After 交错及异步任务延迟。刷新只改变该 turn 计费 group，保留 local 路由、trial、subscription、ledger、QuotaPlatform、ChannelUsageFields 和取消/输出边界；不能成为重授权策略。
- image usage 必测 Responses/Chat details、managed image fallback、usage 优先、turn/session 累加和重复终态不双计；按本地 relay 实际流程证明，不照搬 upstream helper 名称。
- HTTP reasoning 必测准确 code/text 联合条件、最多一次、无 encrypted content 不重试、无 replay/输出边界不变、其他 thinking error 不重试。
- OAuth web_search 必测 ordinary tools/Lite additional_tools、existing declaration/tool_choice、malformed/empty input、API-key/compact 不变及 fake-local-upstream payload；禁止真实付费 provider 或共享 DB/Redis。
- 最终相对 copied baseline 检查 minimal delta、其他 owner 和受保护本地定价行为；provider/生产部署未验证且不在授权范围。

## Recommendation
合同可批准进入实现；独立 QA、fake-local-upstream 边界测试、production build 和 baseline preservation 门禁仍须完成。只新增本报告，不改合同、业务文件、依赖或知识库，不提交、推送或部署。
