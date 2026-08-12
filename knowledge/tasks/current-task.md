# 当前任务快照

最后更新：2026-08-12 12:43 +08:00

## 背景

- S211 将既有 `peak_rate_*` 分时因子扩展到标准与订阅分组；最终 Token
  倍率为原有效倍率乘以分时因子，窗口为服务器时区同日 `[start, end)`。
- S212 为单个账号在已有 `accounts.extra` 中增加可选的每日可用时段；窗口外
  仅排除新请求候选，不变更持久化账号状态、分组/API Key 绑定或在途请求。

## 当前目标

- S211/S212 已通过提交前多智能体复核和独立 QA；现在按后端、前端、workflow
  三批本地提交，并更新本地 `sub2api` 容器，不推送远端。

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

## 已确认事实

- 分时窗口是服务器时区、同日、左闭右开 `[start, end)`；新写入严格校验，
  缺失或无效历史配置 fail-open。
- 所有自动重试、故障切换和异步选择沿用第一次捕获的 `RequestStartedAt`；不需要
  重建 API Key，也不改变账户状态、绑定、账单或媒体按次计费。
- S211/S212 的 review-fix amendment、result 和 QA 证据已收口；用户的
  `outputs/` 未被触碰。

## 待验证点

- 三批提交范围必须精确。验证：每批 `git diff --cached --name-only`、
  `git diff --cached --check` 和 `git ls-files -u` 通过，且不包含 `outputs/`、
  `knowledge/00-start-here.md` 或 `knowledge/05-current-focus.md`。
- 更新本地容器后验证健康状态、`/health`、日志、回滚镜像和桌面/390px 页面；
  远端推送、真实 provider 与生产流量仍不执行。

## 当前结论

- `PASS / pre-commit-review`：S211/S212 代码、contract 和独立 QA 证据可进入
  分批本地提交与本地容器更新。

## 下一步

1. 按后端、前端、workflow 三批本地提交；验证：每批 staged scope 精确且无 `outputs/`。
2. 使用 Docker guard 接管过期锁并更新本地 `sub2api`；验证：容器、端点和日志健康，
   guard 正常释放。
3. 使用任务专用 Playwright profile 验收桌面与 390px；验证：页面可用、无明显布局
   问题，且任务拥有的浏览器/session/daemon 被精确关闭。

## 验证记录

- S211 计费/请求时间第二轮复核：聚焦 `-count=10`、完整 service/handler/middleware、
  server compile、gofmt、diff 和索引门禁通过。
- S212 review-fix 独立 QA：聚焦 `-count=10`、完整 service、middleware、
  handler/admin、server compile、3 files / 42 frontend tests、lint、typecheck、build、
  gofmt、allowlist、diff 和索引门禁通过。
