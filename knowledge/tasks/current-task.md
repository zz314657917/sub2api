# 当前任务快照

最后更新：2026-08-12 13:30 +08:00

## 背景

- S211 将既有 `peak_rate_*` 分时因子扩展到标准与订阅分组；最终 Token
  倍率为原有效倍率乘以分时因子，窗口为服务器时区同日 `[start, end)`。
- S212 为单个账号在已有 `accounts.extra` 中增加可选的每日可用时段；窗口外
  仅排除新请求候选，不变更持久化账号状态、分组/API Key 绑定或在途请求。

## 当前目标

- S211/S212 已完成多智能体复核、独立 QA、分批本地提交、本地容器更新和
  Google Chrome 桌面/390px 验收；保持不推送远端。

## 本次已完成

- S212 已覆盖 Gateway、OpenAI、Gemini-compatible、私有池、快照、sticky、
  pinned 和 WebSocket 的请求开始时间调度，并在创建/编辑账号弹窗提供共享配置控件。
- 根据首轮 QA，合同仅增列必需的
  `antigravity_quota_scope.go`；模型级候选现使用
  `IsSchedulableWithContext`，新增回归证明窗口外拒绝、开始边界允许。
- 首轮独立 Terra QA 后，提交前四路审查发现 generic Gateway 非图片
  `per_request`、handler 首次时间复用、用户自助账号写入校验和前端输入边界缺口。
- 上述缺口已修复并补回归：按次媒体使用原有效倍率，日志倍率与实际扣费一致；
  handler 复用中间件 `RequestStartedAt`；用户账号写入严格校验；新写入仅接受
  `HH:MM`，历史单数字小时运行时继续兼容；OAuth/direct/edit 提交均有父组件防护。
- 原四个审查智能体第二轮复核无剩余代码阻断；S212 独立 Terra QA 按扩展后的
  contract 重跑完整门禁并裁定 `PASS`。
- 用户已明确授权多智能体复核、分批本地提交和更新本地容器。
- 已创建三批本地提交：`0d9004e71` 后端、`240ff4ce8` 前端、`0b836dc8e`
  workflow；`main` 相对 `origin/main` 本地领先，未推送。
- 本地 `sub2api` 已更新到 `sub2api:codex-20260812-1243-s211-s212`，镜像 ID
  `sha256:ee12bc3fa2eb...`，`http://127.0.0.1:62080/health` 返回 200；PostgreSQL、
  Redis、网络和 `deploy_sub2api_data` 未重建。
- 使用 Google Chrome stable headless、任务专用 profile 完成 1440x1000 和
  390x844 验收；分组/账号弹窗的页面与弹窗横向溢出均为 0，时区、公式、时间
  控件、校验提示和底部操作均可见。未提交任何表单。
- Chrome session 已关闭，任务 profile 对应的 Chrome 与 Playwright daemon 进程
  均归零；Docker guard 已使用原 token 正常释放。

## 已确认事实

- 分时窗口是服务器时区、同日、左闭右开 `[start, end)`；新写入严格校验，
  缺失或无效历史配置 fail-open。
- 所有自动重试、故障切换和异步选择沿用第一次捕获的 `RequestStartedAt`；不需要
  重建 API Key，也不改变账户状态、绑定、账单或媒体按次计费。
- S211/S212 的 review-fix amendment、result 和 QA 证据已收口；用户的
  `outputs/` 未被触碰。
- 回滚材料保留为容器 `sub2api-rollback-before-20260812-1243-s211-s212` 和镜像
  `sub2api:rollback-before-20260812-1243-s211-s212`，旧镜像 ID
  `sha256:5307a1216e7e...`。
- 容器最近日志仅见远程价格哈希拉取的 TLS handshake timeout；容器持续 healthy，
  `/health` 正常，该外部网络更新告警不属于 S211/S212 启动或功能错误。

## 待验证点

- 真实 provider、生产流量和远端发布未执行；如后续明确授权，再单独做对应 smoke
  或 push，不从本次本地验收推导生产结论。
- 远程价格哈希网络更新可在网络恢复后自然重试；当前不影响容器健康和本次功能。

## 当前结论

- `PASS / local-container`：S211/S212 已完成代码、contract、独立 QA、三批本地
  提交、本地容器更新和 Google Chrome 响应式验收；无剩余本地交付阻断。

## 下一步

1. 本任务无需继续修改；新需求另开 Sprint/contract。
2. 只有用户明确要求时才推送远端或执行真实 provider/生产流量验证。

## 验证记录

- S211 计费/请求时间第二轮复核：聚焦 `-count=10`、完整 service/handler/middleware、
  server compile、gofmt、diff 和索引门禁通过。
- S212 review-fix 独立 QA：聚焦 `-count=10`、完整 service、middleware、
  handler/admin、server compile、3 files / 42 frontend tests、lint、typecheck、build、
  gofmt、allowlist、diff 和索引门禁通过。
- Git：三批提交成功；最终 `git diff --check`、`git ls-files -u` 通过，工作树仅剩
  用户的 `knowledge/00-start-here.md`、`knowledge/05-current-focus.md` 和 `outputs/`。
- Docker：容器 running/healthy，镜像 `sha256:ee12bc3fa2eb...`，`/health` 200，
  回滚容器/镜像存在，Docker guard 状态为 missing（已释放）。
- Chrome：Google Chrome stable 独立 profile 验收桌面和 390x844；四张具名截图位于
  `E:/codex-artifacts/sub2api/s211-s212-container-20260812`，任务进程清理通过。
