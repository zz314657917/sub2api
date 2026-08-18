---
task_id: upstream-openai-passthrough-model-discovery-s230-b
phase: contract-approved
base: 266ac829809d7cb11c301b8c4df907ce0eb7b5f5
---

# Task Contract

## Goal

适配上游 `1ea4150bf` 的 passthrough model discovery 修复：OpenAI passthrough 账号
不应被陈旧 `model_mapping` 限制可用模型白名单；发现列表应回退默认模型集，普通
OpenAI 账号与其他平台的映射列表保持不变。

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_hotpath_optimization_test.go`
- `docs/workflow/worker-results/upstream-openai-passthrough-model-discovery-s230-b-result.md`
- `docs/workflow/qa-reports/upstream-openai-passthrough-model-discovery-s230-b-qa.md`

## Denied Paths

- gateway handlers、Codex probe/fingerprint、billing、CN provider、frontend、migrations、
  dependencies、provider calls、push、deployment、containers、数据库
- 用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同 workflow report 外不得修改

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/service -run "TestGetAvailableModels_OpenAIPassthroughUsesDefaultFallback|TestGetAvailableModels_GlobalListPreservesMappedModelsWithOpenAIPassthrough|TestGetAvailableModels_ErrorAndGlobalListBranches" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/service/gateway_service.go internal/service/gateway_hotpath_optimization_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor 1ea4150bf upstream/main
Pop-Location
```

## Stop Rules

- passthrough 账号仍被 stale mapping 限制、普通账号映射改变、global list 丢失非 OpenAI 映射，或任何 denied path 变化，立即判定 FAIL。
