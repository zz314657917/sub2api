# 当前任务快照

最后更新：2026-08-31 17:23 +08:00

## 背景

- S275 修复 APIMart 图片余额换算倍率被误展示成分组/用户计价倍率的问题。
- Earlier task snapshots were archived by pge-compact at 20260831T062627697Z.

## 当前目标

- S275 的实现、独立 QA 和本地提交均已完成；保持未推送状态。

## 本次已完成

- 产品与回归测试已提交为 `27daa1f2a`：只读投影
  `pricing_rate_multiplier` / `balance_conversion_multiplier`，用户端、管理端和导出
  分别展示计价倍率、余额换算与综合倍率。
- workflow contract/review/result/QA 与历史压缩归档已提交为 `641d0d341`。
- 未修改实际扣费、余额、`total_cost`、`actual_cost`、持久化
  `rate_multiplier`、schema、migration 或 repository SQL。

## 已确认事实

- 独立 `gpt-5.6-terra` QA 已通过 service/DTO focused x10、完整受影响 Go 包、
  server compile、43 个前端 Vitest、typecheck、gofmt、diff/冲突/allowlist 和受保护
  路径 SHA-256 门禁。
- 本轮提交前复跑 focused Go、43 个前端 Vitest、`vue-tsc --noEmit` 和两次 staged
  `git diff --check`，均通过。
- API-key route breaker、`admin_service.go`、Pixel Cafe 管理页和 `outputs/**` 的既有
  修改未进入上述提交。

## 待验证点

- 若未来由独立单条详情 API 展示非 official、仅靠 APIMart 账号触发的历史图片记录，
  需补 `GetByID` Account hydrate；验证方式：新增 repository/service 单条读取回归并核对
  DTO 拆分字段。当前详情抽屉使用已 hydrate 的列表行，不受影响。

## 当前结论

- `PASS / usage-billing-multiplier-breakdown-s275`：实现、证据和 handoff 已按独立提交
  整理；当前未 push、未更新容器、未操作数据库或共享数据。

## 下一步

- 仅在用户明确要求推送后执行普通 push -> 验证：比较 `HEAD`、`origin/main` 与远端
  `refs/heads/main`，确认无强推且 ahead/behind 收敛。
- 如需处理 `GetByID` 关联边界，另开 contract -> 验证：覆盖非 official APIMart
  账号图片的独立单条查询，不改变持久化账单。

## 验证记录

- `go test ./internal/service -run 'TestUsageRateMultiplierBreakdown' -count=1`：PASS。
- `go test ./internal/handler/dto -run 'TestUsageLogFromService.*MultiplierBreakdown' -count=1`：PASS。
- `vitest` 两个目标文件：2 files / 43 tests PASS。
- `pnpm run typecheck`：PASS。
- 产品提交 `27daa1f2a`；workflow 提交 `641d0d341`；均未推送。
