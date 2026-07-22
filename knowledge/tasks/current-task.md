# 当前任务快照

最后更新：2026-07-22 14:28 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- S98/S99 组合功能提交 `34b1844ab` 和 workflow 收口提交 `be5d4114d` 已推送。
- 当前按已批准的上游差异计划推进独立小步迁移，不整体 merge `upstream/main`。

## 当前目标

- S100：移植上游 `d0fa8c63f`，将订阅剩余天数从向下截断改为正数向上取整。
- 保持 24 小时 duration 语义，不扩展成用户时区日历日，不触碰前端、计费、持久化或部署。

## 本次已完成

- `DaysRemaining()` 复用可确定测试的 `daysRemainingAt(now)`。
- 过期和恰好到期返回 `0`；不足一天返回 `1`；超过整天的余数向上取整。
- progress DTO 的剩余天数断言收敛为精确值。
- 上游新增测试移除 `unit` build tag，使其进入本地默认 service 测试集，避开既有 unit-tag 聚合编译漂移。

## 验证记录

- 测试发现同时列出 `TestCalculateProgress_BasicFields` 和 `TestUserSubscriptionDaysRemainingAt`。
- focused 和 broader progress service tests PASS。
- `gofmt -d`、`git diff --check`、冲突标记、精确路径和未合并索引检查 PASS。

## 当前结论

- `PASS / published`：S100 功能提交 `e567ecbb5` 已推送到 `origin/main`。
- 未部署、未更新容器；24 小时 duration 语义未改为日历日。

## 下一步

1. 进入下一独立 Sprint，手工迁移上游零散硬编码文案本地化。
2. Responses SSE 热路径优化随后单独处理，不引入上游整套 gateway 文件拆分。
