# 当前任务快照

最后更新：2026-07-21 12:12 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- S88 修复多分组模型感知路由的错误默认兜底：候选路由全部不匹配时，
  原实现仍会把文字模型交给生图默认分组。
- 本轮只修改 service 路由边界、middleware 错误返回、定向测试和工作流
  证据；未修改持久化、计费、账号调度、前端或部署配置。

## 当前目标

- 已完成：模型感知多分组路由在默认组不兼容时 fail closed。
- 已授权并收口：本地 scoped commit。
- 未授权：推送、部署和容器更新。

## 本次已完成

- 默认组的 platform、`routing_scope` 或启用的显式路由规则与请求不兼容
  时，不再回落到该组。
- Middleware 返回 HTTP 403 和稳定错误码 `NO_MATCHING_GROUP_ROUTE`。
- 保留兼容默认组、匹配配置路由、单分组 key 和请求体解析前路由行为。
- 新增 service 8 个、middleware 1 个 S88 回归测试。

## 已确认事实

- S88 测试发现 9 个用例；service 和 middleware 定向测试均以
  `count=10` 通过。
- 既有 `ResolveForRequest` / `ResolveForModelRequest` 回归及
  `handler` / `server/routes` 编译检查通过。
- 完整 middleware 包通过。
- 完整 service 包只在既有 `group_peak_rate` 的 7 个时区断言失败；相同
  失败已在干净基线 `96021f068` 复现，且 `TestPeakMultiplier*` 隔离通过。
- S88 diff、allowlist、unmerged/conflict 和 whitespace 检查通过。

## 待验证点

- 未使用真实 API key 对文字、生图、视频和 embedding 请求做上游 smoke。
- 完整 service 包需等待 `group_peak_rate` 全局时区测试隔离问题另行修复。
- S88 源码和证据已作为本地 scoped commit 收口；尚未推送，运行环境未部署。

## 当前结论

- `PASS / source-committed`：S88 修复、定向回归和本地 scoped commit 已
  闭环，远端尚未更新。
- `not deployed`：未构建、替换或重启任何容器。

## 下一步

1. 普通“继续” -> 保持当前本地提交，等待新的明确任务。
2. 用户要求“推送” -> 正常推送当前 `main` 并核对远端 SHA。
3. 用户明确要求部署/更新容器 -> 新建独立任务并先获取容器锁。

## 验证记录

- 2026-07-21 11:15 +08:00：S88 发现、service/middleware `count=10`、
  既有路由回归及下游编译 PASS。
- 2026-07-21 11:17 +08:00：完整 middleware PASS；完整 service 仅复现
  已知 peak-rate 时区污染；隔离 peak 测试 PASS。
- 2026-07-21 12:12 +08:00：S88 精确 11 路径暂存、cached diff 和 fresh
  focused tests PASS；本地 scoped commit 收口，未推送。
