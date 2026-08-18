---
task_id: upstream-codex-usage-probe-model-s230-a
phase: contract-approved
base: a7834f0889aff7579b6bb9402393c4d82f2416f5
---

# Task Contract

## Goal

适配上游 `16e4f7ecc` 的 Codex OAuth 额度探针模型兼容性修复：为额度探针定义
专用 `CodexUsageProbeModel = "codex-auto-review"`，并仅在
`probeOpenAICodexSnapshot` 使用该模型；普通 OpenAI 账号测试模型保持不变。

## Allowed Paths

- `backend/internal/pkg/openai/constants.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/openai_codex_usage_probe_model_test.go`
- `docs/workflow/worker-results/upstream-codex-usage-probe-model-s230-a-result.md`
- `docs/workflow/qa-reports/upstream-codex-usage-probe-model-s230-a-qa.md`

## Denied Paths

- 其他探针、网关、计费、指纹、CN provider、frontend、migrations、dependencies、
  provider calls、push、deployment、containers、数据库
- 用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同 workflow report 外不得修改

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestCodexUsageProbeModel|TestOpenAICodexVersionConsistency" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/pkg/openai/constants.go internal/service/account_usage_service.go internal/service/openai_codex_usage_probe_model_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 16e4f7ecc upstream/main
Pop-Location
```

## Stop Rules

- 不得修改普通 `DefaultTestModel` 语义；不得扩展到其他 OpenAI probe 或测试路径。
- 任何 denied path、依赖、provider、数据库、容器、部署、远端、冲突或 unmerged-index 变化立即停止。
