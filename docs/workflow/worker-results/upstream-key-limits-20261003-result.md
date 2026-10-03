DONE: upstream-key-limits-20261003

实现及本 worker 定向验证完成，等待独立 QA；不代表完整 service unit 套件或部署验收通过。

## Findings

- 修改仅合同允许的 8 个代码/配置路径及本报告。保留隔离树既有 `docs/workflow/status.md` 修改；未写共享 F: 树、未 commit/push/deploy。
- `api_key_create` 默认 200/60；零值独立关闭、负值拒绝。nil cfg 不启用限制。数量错误保留上游 Forbidden `API_KEY_COUNT_EXCEEDED`；小时错误为 `API_KEY_CREATE_RATE_LIMITED`。
- 在实际 `APIKeyService.Create` 已有验证及生成之后、持久化之前检查数量和小时次数；自定义与生成 Key 共用 quota。
- 通过可选 `APIKeyCreateCounter` 保持原 `APIKeyCache` doubles 兼容。生产构造器真实返回实现该接口的 cache。Redis Lua 原子 INCR，首次设置 3600 秒 TTL，后续不延长；独立 key prefix `apikey:create:hour:`。
- 旧失败计数方法、TTL、update/delete 清理行为未改。新计数没有退款/删除入口；后续数据库创建失败仍消耗额度。

## Executed Checks

根目录 `E:/codex-worktrees/sub2api/upstream-integration-20261003`，backend 下执行：

1. `go test ./internal/config -run TestAPIKeyCreateConfig -v -count=1`：退出码 0，2 个顶层测试 PASS；默认、独立零值、两个负值配置实际 Load 验证。
2. `go test ./internal/repository -run TestAPIKeyCreateCount -v -count=1`：退出码 0，4 个顶层测试 PASS。使用本测试创建的 miniredis，无共享 Redis。
   - 真实 `NewAPIKeyService` 和 `NewAPIKeyCache` 执行 100 次并发 Create（custom/generated 各半），60 成功、40 小时拒绝、持久化 60 次。
   - 100 次并发 Lua 计数值唯一，按 60 阈值恰好 60 次 admission。
   - 初始 TTL 一小时；FastForward 20 分钟后再增仍 40 分钟；到期重建窗口一小时；用户隔离。
   - 旧失败计数 increment/delete 不清除小时计数；nil Redis 和关闭 miniredis 返回错误。
3. service 目录实际全部当前生产 `GoFiles`（423 个）加新增测试文件：

   ```powershell
   $pkg = go list -json . | ConvertFrom-Json
   $files = @($pkg.GoFiles) + @('api_key_create_limits_test.go')
   go test -tags unit -v -count=1 @files
   ```

   退出码 0，8 个顶层测试 PASS，包含实际 Create/EnsureInitialKey/Update/Delete：custom/generated quota，count 满/超额/error，各开关独立关闭，数据库失败不返还，nil/缺少可选能力/错误 cache fail-open，nil cfg，重复 custom 保留旧失败计数，非法请求不消耗，旧失败阈值不消耗，默认 bootstrap 创建消耗一次、已有 Key 跳过、再次创建遭小时拒绝，以及 update/delete 不退款。

4. 现有 default/Cafe 定向回归，仍使用 423 个实际生产文件：

   ```powershell
   $files = @($pkg.GoFiles) + @(
     'api_key_service_cache_test.go',
     'api_key_service_delete_test.go',
     'concurrency_service_test.go'
   )
   go test -tags unit -v -count=1 -run 'TestApiKeyService_Delete_Default|TestApiKeyService_Update_Default|TestAPIKeyService_Cafe|TestAPIKeyService_ExpiredCafe' @files
   ```

   退出码 0，4 个顶层测试 PASS：
   `TestAPIKeyService_CafeBindingExpiryBoundsAuthCacheTTL`、
   `TestAPIKeyService_ExpiredCafeBindingSnapshotForcesAuthReload`、
   `TestApiKeyService_Delete_DefaultKeyRejected`、
   `TestApiKeyService_Update_DefaultKeyAllowed`。

5. `go build ./...`：退出码 0；`git diff --check`：退出码 0，仅既有 status.md LF/CRLF 提示。
6. `go test -tags unit ./internal/service -run TestAPIKeyCreateLimits -count=1`：退出码 1，实际执行 0 测试；复现独立基线 QA 已记录的同一编译阻塞：
   - `ops_health_score_test.go:442` / `usage_leaderboard_reward.go:1042` 的 stringPtr 重定义；
   - `billing_service_test.go:1395,1417`、`billing_service_unified_test.go:32` 参数签名漂移；
   - `gateway_context_management_test.go:637` 返回数漂移；
   - `proxy_update_probe_invalidation_test.go:55-63` Proxy/UpdateProxyInput 字段漂移。

最终成功集合合计 **18 个顶层测试、0 失败**（不将重跑及子测试重复计数）。开发中新增服务测试首次 6 PASS/1 panic，次次 7 PASS/1 panic，原因均为测试桩未实现 Update/Delete 调用的 ListByUserID/Delete，已在允许测试文件补齐后 8/8 PASS。尝试更大的原始 initial/default/Cafe 文件集合两次因缺失其他测试 helper 未编译（0 执行）；没有复制或伪造生产类型，最终用上述已执行集合替代，完整原始 initial 文件仍未执行。

## Static Call Path Review

- `api_key_repo.go:60-62` activeQuery 仅排除软删除；`:911-913` CountByUserID 增加 owner 条件，没有 status/expiry/managed 排除，所以名字 max_active 实际计入全部非删除 Key。
- `api_key_service.go` EnsureInitialKey 先 count，有 Key 直接返回；没有 Key 调真实 Create。auth 现有错误处理未改。
- `cafe_room_activation_service.go:355` createMembershipManagedKey 与 `:769` ensureManagedKey 直接调用 `s.apiKeyRepo.Create`（repository 接口），未经过普通 APIKeyService.Create，因此小时限额豁免。两个生命周期函数和 repository Create 未修改。
- 数量 gate 位于所有现有 validation/route/group/Cafe guard 之后；小时 gate 位于数量 gate 后。没有 migrations/dependencies/wire/frontend 变更。

## Unverified Risks

- 数量为非事务 count-before-create 软防护：并发可超过 count cap；不是硬上限。
- Redis nil/错误或缺少 optional interface 时小时 guard fail-open；cache 健康时的原子计数并发 admission 有证据，Redis 不可用时没有小时保障。nil cfg 所有新 guard 关闭。
- 固定窗口从第一次有效创建尝试开始，不是日历整点，也不是滑动窗口。拒绝的小时尝试仍增计数；后续持久化失败消耗 slot。
- 显式 GoFiles 避开既有测试编译漂移，不等价于全包测试 PASS。数据库、真实 Cafe 激活事务、auth 真实登录、运行服务和部署未验收。计数包含所有非删除状态获生产查询静态证据，未跑真实数据库状态矩阵。
- 直接 repository/Ent/SQL 创建者不受普通 service 小时 gate 约束；不声称覆盖每个创建入口。

## Recommendation

可进入独立 QA；保留完整 service unit 基线 BLOCKED、soft count cap 与 Redis fail-open 限制。独立 QA 应复跑最终 18 个定向测试和完整构建，不将开发自述作为最终 PASS。
