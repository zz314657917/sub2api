### BLOCKED: upstream-v0200-composite-gateway-selection-s295-p1

# Findings

- 未发现 P1 静态实现、定向测试、Wire 编译或 backend build 的明确功能问题。
- `go generate ./cmd/server` 被既有 Wire 图缺失 `[]service.ProxyRepository` 和 `service.newUserTrialConsumer` 阻断。P1 仅手工同步了允许的 `cmd/server/wire_gen.go` 注入，`go test ./cmd/server -run '^$'` 已通过；不修改无关 provider。
- 当前会话没有获得独立 `gpt-5.6-terra` QA Worker 的执行证据。仓库 Agent Matrix 禁止以同一执行者的复核替代该门禁，因此本报告必须为 `BLOCKED`。

# Executed Checks

- 逐项核对批准合同的允许/拒绝路径：新增 schema、迁移、Ent 生成集、仓储、窄 service、GroupHandler、admin routes、Wire 和三份定向测试均在允许范围内。
- 审查 `preview`：仅调用路由仓储并按 `priority/id` 所给次序返回候选；没有读取账号、`model_mapping` 或 Gateway 上下文。
- 审查 `DeleteCascade`：在持有的 Ent 事务中软删除 `composite_model_routes`，然后软删除分组。
- `gofmt -w`（P1 Go 文件）通过。
- `go test ./internal/service -run 'CompositeRoute' -count=1` 通过。
- `go test ./internal/handler/admin -run 'CompositeRoute' -count=1` 通过。
- `go test ./internal/server/routes -run 'CompositeRoute' -count=1` 通过。
- `go test ./cmd/server -run '^$' -count=1` 通过。
- `go build ./...` 通过。
- `git diff --check` 通过；`git ls-files -u` 无输出。开始本轮时已存在的教程、Pixel Cafe、支付、lockfile、`outputs/`、`pelican-bicycle.html` 和 `sub2api` 脏项仍不在 P1 产品范围内。

# Unverified Risks

- 未执行迁移，故未验证 PostgreSQL 约束、索引、软删除和事务回滚的运行态行为。
- 未启动认证后台服务，故 admin auth/audit middleware 与 HTTP 响应只做了静态注册和 handler 单元验证。
- `go generate ./ent` 未在本轮完整重放；此前的 Windows 文件锁会污染合同外生成文件。runtime 初始化按上游同实体输出补齐并已随 build 编译。
- P2/P3 的 resolver、账号所有权、请求 context、Gateway 选择和 provider dispatch 明确不在本批，尚无运行态语义。
- 缺少独立 Terra QA Worker 复核，不能给出整体 PASS。

# Recommendation

保留 P1 代码和本地检查证据，先在明确授权的独立 `gpt-5.6-terra` QA Worker 中复跑定向命令、范围审计和生成器基线检查；在专属 PostgreSQL 环境执行迁移验证前，不发布、不启用 Composite 路由，也不推进 S295-P2/P3。

# 2026-09-09 复核补充

- 重新执行 `go test ./internal/service -run 'CompositeRoute' -count=1`、`go test ./internal/handler/admin -run 'CompositeRoute' -count=1`、`go test ./internal/server/routes -run 'CompositeRoute' -count=1`、`go test ./cmd/server -run '^$' -count=1`：全部通过。
- 重新执行 `go test ./internal/repository -run '^$' -count=1`、`go build ./...`、P1 文件 `gofmt -d`、`git diff --check` 和 `git ls-files -u`：全部通过或无输出。
- 白名单审计确认 P1 新增文件均在合同范围内；现有 8 个跟踪脏文件和未跟踪输出仍为既有改动，未被本任务吸收。
- `pge-doctor --strict`：20 项通过，1 项非阻断 warning（`docs/workflow/status.md` 压缩提示）。
- Hermes `tester` profile 当前为 `gpt-5.3-codex`，与仓库 Agent Matrix 要求的 `gpt-5.6-terra` 不一致，因此本轮不以该 profile 冒充独立 Terra QA。

# 2026-09-09 R4 资源复核

- 发现并核验任务专属资源：`sub2api-s293-test-postgres`（healthy，专属网络 `sub2api-s293-test-net`，容器地址 `172.18.0.3`）、`sub2api-s293-test-redis`（healthy，地址 `172.18.0.2`）和配套测试服务（healthy）。容器归属、专属卷和网络均已确认；没有操作共享 `sub2api-postgres`、`sub2api-redis` 或 `sub2api`。
- 使用专属容器内网地址实际运行 `go test ./internal/securityaudit -run 'Integration|RoundTrip|Runtime' -count=1`。宿主 Windows 到两个 Docker 内网地址均不可达：PostgreSQL 报 `connectex`，Redis 报 `i/o timeout`；4 个运行态用例因此失败在连接初始化阶段，未执行 migration 239、TRUNCATE 或业务断言。
- 本轮没有执行迁移，没有改变任何容器、卷、数据库或 Redis 数据。R4 仍为 `BLOCKED`，需要为该专属资源提供宿主可达的端口映射或在同一 Docker 网络内运行测试，并重新验证资源归属。
