### DONE: group-model-match-legacy-compat-s273

# Worker Result

## Task ID
group-model-match-legacy-compat-s273

## Status
`done`

## Summary
- 修复 S272 引入的旧数据回归：非 pinned、无多组路由的单组 API Key 在
  `model_match_patterns` 为空时继续使用默认分组；有配置时仍严格按管理员
  模型规则匹配。
- 兼容分支不再使用通用 endpoint 的平台/路由 scope 列表误伤单组旧 Key，图片
  权限仍由 handler 返回稳定的 `permission_error`。多组和 pinned 路由的空规则
  继续 fail-closed。
- 补充 service 与 middleware 回归，覆盖空规则、严格匹配、图片权限、Grok
  单组规则，以及不完整默认快照不能绕过多组路由。

## Changed Files
- `backend/internal/service/api_key_routing.go`
- `backend/internal/service/api_key_routing_s273_test.go`
- `backend/internal/server/middleware/api_key_auth.go`
- `backend/internal/server/middleware/api_key_auth_s272_test.go`

## Commands Run
```text
go test ./internal/service -run 'TestS273|TestS91' -count=10 -> PASS
go test ./internal/server/middleware -run 'TestS272|TestS273' -count=10 -> PASS
go test ./internal/service -count=1 -> PASS (65.352s)
go test ./internal/server/middleware -count=1 -> PASS
go test ./internal/handler -count=1 -> PASS (32.900s)
go test ./internal/server/routes -run 'TestGatewayRoutes' -count=1 -> PASS
go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -w <changed Go files> -> PASS
git diff --check -> PASS
git ls-files -u -> PASS (no unmerged entries)
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.885s (focused x10)
ok github.com/Wei-Shaw/sub2api/internal/server/middleware 5.698s (focused x10)
ok github.com/Wei-Shaw/sub2api/internal/service 65.352s
ok github.com/Wei-Shaw/sub2api/internal/handler 32.900s
```

## Risks
- 未执行真实 provider、部署、容器更新、共享/生产数据库写入或 API smoke；这些
  均被 contract 明确排除且未获授权。
- 仓库既有 repository fixture 漂移（32/34）属于本轮前基线问题，未在本任务中
  修改或重新归因；本轮 handler 全包已通过。

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason
- None.
