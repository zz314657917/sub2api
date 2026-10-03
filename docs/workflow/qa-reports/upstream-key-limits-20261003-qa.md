### PASS: upstream-key-limits-20261003

2026-10-03 独立 QA；作用域为 approved 合同的选定回归、完整生产构建与改动审查。**scoped PASS 不代表 service 全包、运行服务或生产验收通过。**

基线 HEAD：`af190154bc62e8d2a66db8f19018ae6a7727a9d6`。工作树：`E:/codex-worktrees/sub2api/upstream-integration-20261003`。先读合同、contract review、developer result 与独立 baseline QA；按 `review-and-verification` 执行。隔离树没有 AGENTS.md，使用主控提供的仓库规则与合同。

## Findings

- 本任务限定范围未发现明确阻断问题。独立实际执行 **18 个顶层测试，18 PASS、0 FAIL**；子测试不重复计数。
- 全 service unit 命令仍编译失败、实际执行 **0 个测试**。首批错误文件、行号及原因与 `upstream-baseline-20261003-qa.md` 一致，保持 BLOCKED，不归入通过测试数。
- 8 个业务/配置/测试文件开始与结束 SHA256 全部一致。QA 没有改业务文件、替换生产源码、删测试或伪造生产类型。
- `git diff --name-only af190154b` 与 untracked 文件清单核对：产品修改正好合同允许的 8 个路径；另有进入 QA 前已存在的 `docs/workflow/status.md` 修改，本 QA 保留。本轮唯一写入为本报告；未 commit、push、deploy。

## Executed Checks

以下命令实际运行于 Windows PowerShell；backend 命令工作目录为上述树的 `backend`。

1. `go test ./internal/config -run TestAPIKeyCreateConfig -v -count=1`：退出码 0，2 个顶层测试 PASS，`ok .../internal/config 0.381s`。
   - `TestAPIKeyCreateConfigDefaultsAndZero`
   - `TestAPIKeyCreateConfigRejectsNegative`
   - 实际 Load 覆盖默认 200/60、两个独立零值开关及两个负值拒绝。配置测试输出的 TOTP/allowlist warning 是 Load 的既有行为，不是验收失败。

2. `go test ./internal/repository -run TestAPIKeyCreateCount -v -count=1`：退出码 0，4 个顶层测试 PASS，`ok .../internal/repository 2.530s`。
   - `TestAPIKeyCreateCountConcurrentRealService`
   - `TestAPIKeyCreateCountFixedTTLAndIsolation`
   - `TestAPIKeyCreateCountConcurrentAtomicAdmissions`
   - `TestAPIKeyCreateCountUnavailable`
   - 使用测试创建的 miniredis 和真实 `NewAPIKeyCache`、`NewAPIKeyService`。真实 Create 并发 100 次，自定义/生成各 50：成功 60、小时拒绝 40、repository mock 持久化调用 60。
   - 原子计数并发 100 次返回值互不重复；60 阈值恰好 60 次 admission。首次 TTL 为 1 小时；推进 20 分钟再递增仍为 40 分钟；窗口到期后重建 1 小时；用户隔离。
   - 旧失败计数 increment/delete 不删除新小时计数。nil Redis 和主动关闭的本测试 miniredis 返回错误。输出 `127.0.0.1:60526` 拒绝连接为该负面测试主动关闭的本地 miniredis；没有连接共享 Redis/数据库/provider。

3. 在 `backend/internal/service` 执行：

   ```powershell
   $pkg = go list -json . | ConvertFrom-Json
   Write-Output "production GoFiles: $($pkg.GoFiles.Count)"
   $files = @($pkg.GoFiles) + @('api_key_create_limits_test.go')
   go test -tags unit -v -count=1 @files
   ```

   当前实际生产 `GoFiles` 共 **423** 个，未替换生产实现。退出码 0，8 个顶层测试 PASS，`ok command-line-arguments 0.569s`：
   - `TestAPIKeyCreateLimitsCustomAndGeneratedShareQuota`
   - `TestAPIKeyCreateLimitsCountAndIndependentDisable`
   - `TestAPIKeyCreateLimitsPersistenceFailureConsumesAttempt`
   - `TestAPIKeyCreateLimitsCacheFailOpen`
   - `TestAPIKeyCreateLimitsNilConfigDisablesGuards`
   - `TestAPIKeyCreateLimitsValidationAndFailedCustomCounter`
   - `TestAPIKeyCreateLimitsEnsureInitialKeyConsumesOnlyCreation`
   - `TestAPIKeyCreateLimitsUpdateDeleteDoNotRefund`
   - 覆盖实际 Create/EnsureInitialKey/Update/Delete；数量满额、超额、查询错误；开关独立关闭；数据库失败消耗 slot；nil/缺失 optional capability/error cache 放行；nil cfg；旧重复 custom 失败计数、非法请求及旧失败阈值不消耗新小时计数；已有 Key bootstrap 不消耗；update/delete 不退款。

4. 同一 service 目录与同一 `$pkg.GoFiles`：

   ```powershell
   $files = @($pkg.GoFiles) + @(
     'api_key_service_cache_test.go',
     'api_key_service_delete_test.go',
     'concurrency_service_test.go'
   )
   go test -tags unit -v -count=1 -run 'TestApiKeyService_Delete_Default|TestApiKeyService_Update_Default|TestAPIKeyService_Cafe|TestAPIKeyService_ExpiredCafe' @files
   ```

   退出码 0，4 个顶层测试 PASS，`ok command-line-arguments 0.560s`：
   - `TestAPIKeyService_CafeBindingExpiryBoundsAuthCacheTTL`
   - `TestAPIKeyService_ExpiredCafeBindingSnapshotForcesAuthReload`
   - `TestApiKeyService_Delete_DefaultKeyRejected`
   - `TestApiKeyService_Update_DefaultKeyAllowed`

5. `go build ./...`：退出码 0，完整生产构建通过。
6. `go test -tags unit ./internal/service -run TestAPIKeyCreateLimits -count=1`：退出码 1，`[build failed]`，执行 0 测试；独立复现的首批错误：
   - `ops_health_score_test.go:442` / `usage_leaderboard_reward.go:1042`：`stringPtr` 重定义；
   - `billing_service_test.go:1395,1417`：`computeTokenBreakdown` 参数数量不匹配；
   - `billing_service_unified_test.go:32`：`calculateCostInternal` 参数数量不匹配；
   - `gateway_context_management_test.go:637`：3 个变量接收 2 个返回值；
   - `proxy_update_probe_invalidation_test.go:55,56,63`：`FallbackMode` / `ExpiryWarnDays` 字段不存在、`FallbackModeNone` 未定义。
   - 与独立基线报告相同；没有为通过此命令修改旧测试。
7. 根目录 `git diff --check`：退出码 0，仅既有 status.md LF/CRLF 提示。`gofmt -l` 对 7 个 Go 文件无输出、退出码 0：

   ```powershell
   gofmt -l backend/internal/config/config.go backend/internal/config/api_key_create_config_test.go backend/internal/service/api_key_service.go backend/internal/service/api_key_create_limits.go backend/internal/service/api_key_create_limits_test.go backend/internal/repository/api_key_cache.go backend/internal/repository/api_key_create_count_test.go
   ```

8. 开始及结束执行 `Get-FileHash -Algorithm SHA256` 对下表 8 个文件比较，全部相同：

   | 路径（工作树相对） | 开始及结束相同 SHA256 |
   | --- | --- |
   | backend/internal/config/config.go | 26483F35E1DEB3D7D030D77D3C823B2BE9EE75572741E4828A184038C174D3F1 |
   | backend/internal/config/api_key_create_config_test.go | 06C8E37E25E18015F1A735E8BA8781AD41D241DCCAB0AC41D97096E47B90CE14 |
   | backend/internal/service/api_key_service.go | 1B39C348585D7C0DE2FB9F0AC056446F41F02CD4F625297B3687FCF6F38A2176 |
   | backend/internal/service/api_key_create_limits.go | 31832B27536F20CEE245587309360DD26FC003E592EE4B83CDF015185DA0D58C |
   | backend/internal/service/api_key_create_limits_test.go | EE1735D21ED8E6C3C1A1DE6D997666B559117803C18B1293B7130D58D7AA1F51 |
   | backend/internal/repository/api_key_cache.go | 0248D11131E5A795EF58F453CB98B5233941D33EAF83D5086FA67D22C67D3131 |
   | backend/internal/repository/api_key_create_count_test.go | DFE022F46C29B50921BADFC42C2C8235673EDE5599AFD0B5B9D526838398445D |
   | deploy/config.example.yaml | E84CAC4A4A47E920AFBF81F30B8AD9197957B03E8B99073BAE9B0D79B6EBCC8C |

## Static Contract Review

- `api_key_repo.go:60-62` activeQuery 仅过滤 `DeletedAtIsNil`；`:911-913` CountByUserID 只增加 owner 条件。计数包含 inactive、expired 和 managed 的所有未删除 Key；没有新增数据库锁或事务，因此并发可能超越数量 cap。
- `api_key_service.go:845` 新 gate 位于请求/IP/group/route/pool/custom/生成校验后、`:874` 实际持久化前。数量 gate 在小时 gate 之前；数量查询失败返回错误。
- `api_key_create_limits.go` nil cfg 关闭新 guard；nil cache、未实现 optional interface、Redis 错误都小时 fail-open。原 `APIKeyCache` 不要求新增方法，旧 doubles 保持兼容。
- `api_key_cache.go` 新前缀为 `apikey:create:hour:`，Lua 原子 INCR，仅第一次设置 3600 秒 TTL。原失败计数 `apikey:ratelimit:` 方法、24 小时 TTL 与 update/delete 清理代码 diff 未改变；没有新小时计数退款入口。
- `EnsureInitialKey` 先 count，有 Key 返回；无 Key 调真实 Create，测试实际证明新小时额度消耗。`auth_service.go:879` 错误处理文件未修改。
- Cafe `createMembershipManagedKey` 和 `ensureManagedKey` 在 `cafe_room_activation_service.go:379,800` 直接调 repository Create，绕开普通 service gate，因此小时豁免；生命周期代码与 repository Create 没有修改。
- 无新增 dependency、migration、Ent、wire、frontend 变更；原 route/group/Cafe 校验 diff 保持。

## Unverified Risks

- 数量是 count-before-create 软防护；并发硬数量上限不在合同或本次证据范围。
- Redis 不可用/缺能力时小时限制失效；nil cfg 新限制关闭。固定窗口从第一次有效尝试开始，非日历整点或滑动窗口。被小时拒绝的尝试仍递增，持久化失败、更新或删除不退款。
- 显式 GoFiles 形成 `command-line-arguments` 包，绕开其他旧测试编译阻塞，不等价于全包通过。18 项定向 PASS 与全 service unit BLOCKED 必须同时保留。
- repository 持久化使用 mock，Redis 使用 miniredis；真实数据库状态矩阵、真实 Cafe 激活事务、auth 登录、运行服务、真实 Redis/真实 provider 和生产部署未验收。
- 直接 repository/Ent/SQL 创建者不受普通 APIKeyService.Create 小时 gate 约束，不声称覆盖全部创建入口。

## Recommendation

本合同范围独立 QA PASS，可供主控继续精确集成；保留全 service unit 旧编译 BLOCKED、数量 soft cap 和小时 Redis fail-open 限制。要求全包或生产验收时，应另立范围修复基线测试并补运行态证据。
