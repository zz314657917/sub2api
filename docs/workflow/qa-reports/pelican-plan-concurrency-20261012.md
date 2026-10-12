### PASS: pelican-plan-concurrency-20261012

# 独立 QA：pelican-plan-concurrency-20261012

- 日期：2026-10-12。
- 隔离工作区：`E:/codex-worktrees/sub2api/pelican-plan-concurrency-20261012`。
- HEAD：`a62b9246397f1caf41f0011db353698164e848f7`，与 approved contract 的 base_commit 一致。
- 依据：主工作区 approved contract、隔离工作区 Developer report、实际代码 diff 和本轮独立执行证据。
- 使用 `review-and-verification`；本报告独立于 Developer，Controller 保留最终验收权。

## Findings

未发现明确阻断问题。生产 diff 仅为构造函数中 `planGate` 容量从 1 改为 4；两个测试文件变更对应批准范围，未见业务范围漂移。

1. **四路同时执行与容量拒绝**：新增 fake executor 用每计划 release channel 阻塞执行，四个 started 通知后断言 active/maxActive 均为 4，因此证明真实 service goroutine 执行重叠，而非只接受四个请求。手动、定时共享 `claimAndRun`，满容量下两者均不进入 Claim、不增加 ReserveAttempt、不推进计划 5 的 due 时间。计划 1 完成并释放 gate 后，计划 5 的定时重试成功。
2. **同计划保护及失败回收**：相同 ID 的第二次启动在空余 gate 下到达 mutex 保护的 fake Claim，返回 conflict；占用槽和额度仍各为 1，其他计划能填满余下三槽。settings read error、feature disabled、Claim error 各重复六次并断言 gate=0，然后成功启动四个计划，覆盖失败槽泄漏。
3. **Stop 与父级取消**：四个 executor 在收到取消后被第二个退出屏障阻塞，Stop 尚未返回、gate 仍为 4；取消 Stop context 及已过期 deadline 分别保留 wrapped Canceled/DeadlineExceeded。新手动/定时启动拒绝；允许 executor 退出后 Stop 成功，并断言 active round、execution、account lease、gate 全部归零。另一个测试让 Claim 本身收到取消后继续阻塞，证明 Stop 等待已登记的 Claim。父 context 取消测试验证四轮 Finalize 及 gate 清空。
4. **计划内模型顺序与既有规则**：两个计划并行，每个 first model 均由 channel 屏障控制，second model 在 first 释放后才进入。最终 overlap/leaseError=false，最大全局执行数=2，lease 全部释放。完整 Pelican 测试还覆盖既有额度停止、失败后继续下一模型、父取消、timeout 配置、feature switch 和 HTTP 请求取消不影响已接受轮次。
5. **实现安全性静态核对**：gate 获取和 `wg.Add` 保持在 Stop 使用的同一 mutex 下；Stop 设置 stopped 后才等待，关闭新登记入口。成功 run 的 Finalize、gate 释放先于外层 `wg.Done`；settings/Claim 失败同步归还 gate 和计数。定时循环仍调用同一路径。真实 repository 的事务 `SELECT ... FOR UPDATE`、running_until 判断、run_generation 增量及后续 generation fencing 均未修改。

新增测试没有以 sleep 制造并发；1 ms ticker 仅观察 fake Finalize 通知之后执行的 deferred gate 释放。所用 bounded channel 等待在失败时给出超时；Stop barrier 测试有 defer 解锁，避免断言失败把 cleanup 永久挂住。更新的 runner fixture 保留传入计划 ID，修正旧的单槽期望合理。

## Executed Checks

下列 Go 命令均在隔离工作区 `backend` 执行，独立运行，未使用外部 overlay。

| 命令 | 本轮证据 |
| --- | --- |
| `go test -p 1 ./internal/service -run Pelican -count=1` | PASS，`ok .../internal/service 1.304s` |
| `go test -p 1 ./internal/service -run 'PelicanPlanConcurrency\|PelicanRunnerRejectsManualRunAtCapacity' -count=20` | PASS，`ok .../internal/service 0.640s`；七个并发测试及 runner 测试各重复 20 次 |
| `go build -p 1 ./internal/service` | PASS，退出码 0 |
| `git diff --check` | PASS，无输出；在工作区根目录运行 |
| `gofmt -l backend/internal/service/pelican_test_service.go backend/internal/service/pelican_test_runner_test.go backend/internal/service/pelican_plan_concurrency_test.go` | PASS，无输出 |
| `go version` | `go version go1.26.3 windows/amd64` |
| `go env CGO_ENABLED` | `0` |
| `Get-Command gcc -ErrorAction SilentlyContinue \| Select-Object -ExpandProperty Source` | 无输出；当前 PATH 无 GCC |
| `go test -race -p 1 ./internal/service -run 'PelicanPlanConcurrency\|PelicanRunnerRejectsManualRunAtCapacity' -count=1` | 环境 BLOCKED：`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` |

精准 diff 核对：tracked 修改只有两个允许的 service/runner 文件，新增 concurrency 测试为 untracked 文件；生产 diff 一行。新增本 QA 报告位于批准路径，沿用仓库 `docs/*` 忽略规则。未编辑业务代码、未 commit/push、未部署、未调用 provider、未修改数据库、未操作浏览器。

## Unverified Risks

- Race detector 未运行，不能把重复测试通过视为 data race 检测通过。
- 新 fake Claim 证明 service 尊重 repository conflict 和 generation 参数，不等价于真实 PostgreSQL 并发验收；真实 repository 的原子性保留由本轮静态检查确认。
- 定时覆盖调用 scheduler 的共同入口 `claimAndRun(..., false)`，未等待真实 30 秒 ticker。
- fake account leases 按独立计划/账号记录，不证明真实多个计划竞争同一账号的吞吐行为；真实 selector/account enforcement 代码未修改，现有对应 Pelican selector 测试随全量测试通过。
- 四槽属于服务实例，本测试及实现不提供跨进程统一四槽上限。
- 本轮是 approved contract 指定的本地 fake service 执行验收，不代表生产数据库、provider、容器或浏览器验收。

## Recommendation

批准范围内的代码及受控执行验收 PASS，可交 Controller 做最终验收并决定后续流程。无需为本轮通过扩大到数据库或 provider 操作；如后续要求 race 或真实环境验收，应单独补足上述证据。
