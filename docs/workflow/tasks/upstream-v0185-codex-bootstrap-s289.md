---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-codex-bootstrap-s289
base_commit: e6845b4ea
upstream_refs:
  - 1be69e56
  - 421a83282
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-02
---

# Upstream v0.1.185 Codex bootstrap without call id S289

## Goal

手工适配上游 `1be69e56` 与 `421a83282`：仅把经过严格形态校验、但缺失
`call_id` 的 Codex delegation 或 scheduled automation bootstrap 从
`function_call_output` 改为普通 user message，避免本地 HTTP Responses 预校验直接
400。其它工具输出继续走原有关联校验。

## Allowed Paths

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_codex_bootstrap_test.go`
- `docs/workflow/worker-results/upstream-v0185-codex-bootstrap-s289-result.md`
- `docs/workflow/qa-reports/upstream-v0185-codex-bootstrap-s289-qa.md`

## Invariants

- 仅在没有非空 `previous_response_id`、call/reference anchor、非空 `call_id`、重复 JSON 键或混合/未知 call output 时转换。
- delegation 必须是无 namespace、无属性、仅含非空 `source_thread_id` 和 `input` 的完整 XML envelope。
- automation 仅接受 `codex_app.automation_update`，并校验 ID、memory 路径、last-run 时间戳及非空 prompt。
- JSON 数字精度、输入顺序与非目标字段保持；不改 service、路由、权限、计费、数据库、前端、容器或保护 dirty paths。

## Acceptance

From `backend`:

```text
go test ./internal/handler -run 'TestNormalizeCodex(Delegation|Automation)Bootstrap' -count=10
go test ./internal/handler -count=1
go test ./cmd/server -run '^$' -count=1
go build ./...
gofmt -d internal/handler/openai_gateway_handler.go internal/handler/openai_codex_bootstrap_test.go
```

From repository root:

```text
git diff --check -- backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_codex_bootstrap_test.go
git diff --name-only --diff-filter=U
```

真实 OpenAI provider、WebSocket、数据库、容器、部署和浏览器 smoke 不属于本 Sprint。
