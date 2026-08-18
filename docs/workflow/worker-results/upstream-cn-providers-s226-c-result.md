### PASS: upstream-cn-providers-s226-c

# S226-C Controller Result

## Task ID

`upstream-cn-providers-s226-c`

## Status

`done`

## Summary

Controller 接管并完成国产供应商多协议网关与失败处理：

- Kimi/Zhipu/DeepSeek 保持精确平台调度；接入 Chat Completions、Responses、原生 Anthropic Messages/count_tokens 路径。
- DeepSeek Responses 使用 `/responses`，强制 `store=false` 并移除 `previous_response_id`；补齐 HTTP/WS/回退路径的 CN API key/base URL。
- 四条 Anthropic-native CC/Responses 读循环使用 `gateway.stream_data_interval_timeout` 泵，idle 时关闭上游 body，并保留已累计 usage。
- CN 余额不足走可恢复临时停调，Coding Plan 429 冷却到快照中的最早 reset；非 CN 行为保持原路径。

业务提交：`24873abf1`  业务 patch-id：`d6ee6e8e161ad9343b86f8092e55a4be9e2fbe88`

## Changed Files

本提交严格包含 20 个 S226-C allowlist 内的业务/测试文件；无其他路径变更。

## Commands Run

```text
gofmt -w <20 C business/test files> -> PASS
gofmt -d <20 C business/test files> -> no output
git diff --check -> PASS
git diff --name-only f6b380e21 + git ls-files --others -> exact C allowlist
git diff --name-only --diff-filter=U -> empty
git ls-files -u -> empty
git merge-base --is-ancestor 901a0439f upstream/main -> PASS
git merge-base --is-ancestor 4b667ccd4 upstream/main -> PASS
git merge-base --is-ancestor e72854538 upstream/main -> PASS
```

## Test Output

```text
17 focused tests (contract 16 + CN credential/WS regression), all discoverable, -count=10 -> PASS
go test ./internal/service ./internal/handler ./internal/server/routes -count=1 -> PASS
  internal/service 66.376s
  internal/handler 27.215s
  internal/server/routes 1.520s
go test ./cmd/server -run '^$' -count=1 -> PASS
```

## Risks

- 未执行真实国产供应商请求、Redis、数据库、容器、部署或 push；验收使用本地 mock/HTTPUpstream、完整 Go 回归和编译。
- S226-D 前端、独立 QA 和主线集成尚未开始；本报告只裁决 S226-C Controller gate。

## Knowledge Candidates

- 无需新增长期知识；行为和验收证据已由本报告及 C contract 固化。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Protected Main Worktree Evidence

在主工作区复核通过：account modal patch-id `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`、TutorialView patch-id `a07a7c33f09d9fa0e308a1bddf6bf0ee9d7cf671`、knowledge patch-id `2abee47db90ce1d54e1f9ba7d1a3cc2d633c2374`、backend tutorial tests patch-id `a81fbffbe14121ef62387f28cfee09a6d247ac94`；六个 C0 教程 migration/test 文件仍未跟踪且 SHA256 未变；`outputs/` 仍未跟踪。
