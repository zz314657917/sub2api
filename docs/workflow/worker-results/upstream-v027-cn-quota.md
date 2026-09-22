### PASS: upstream-v027-cn-quota

已将 `db8692d67` 的 CN Coding Plan 配额耗尽 403 适配到当前工作树。只有 Coding Plan
CN 账号命中明确的 `access_terminated_error` 或额度耗尽/重置文本时，才跳过既有通用 403
breaker：存在未来的 5h/weekly 快照时写入 `SetRateLimited` 至最早重置点；快照缺失、过期或
`SetRateLimited` 持久化失败时，回退到 10 分钟 `SetTempUnschedulable`。临时停调失败仅记录
可观测日志，不会对已识别的可恢复耗尽调用 `SetError`。

修复独立 QA 指出的运行时调度传播缺口：`SetRateLimited` 或回退
`SetTempUnschedulable` 持久化成功后才调用 `notifyAccountSchedulingBlocked`；两次写入均失败
不会留下虚假的进程内停调状态。

改动路径：

- `backend/internal/service/ratelimit_cn_providers.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/cn_quota_v027_test.go`

运行证据：

- `go test ./internal/service -run '^TestV027CNQuota' -count=1 -v`：PASS（含 runtime blocker
  成功写入后通知，以及双写失败不通知）。
- `go test ./internal/service -run 'CN.*403|403.*CN|Concurrency' -count=1 -v`：PASS。
- `go build ./...`：PASS（`BUILD_EXIT=0`）。
- `git diff HEAD --check`：PASS。
- `git diff --name-only --diff-filter=U`：空。

测试经 `HandleUpstreamError` 公共入口和 fake `AccountRepository` 验证：结构化与文本分类、未来/
过期/缺失快照、限流写入失败回退、临时停调写入失败、成功持久化后的 runtime blocker 通知、非
Coding Plan 账号及 Kimi 并发 403 的既有路径。未访问真实 provider、数据库、容器或付费服务。
