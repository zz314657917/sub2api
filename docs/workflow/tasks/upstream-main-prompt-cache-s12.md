# 上游合成 Sprint：`upstream-main-prompt-cache-s12`

## Summary

- 目标：在独立分支 `codex/upstream-main-prompt-cache-s12` 上合成 `v0.1.135` 后新增的低风险 OpenAI prompt cache 修复，不直接 merge `upstream/main`。
- 基线：本地 `main=c1cb19951`，上游 `upstream/main=b7cfe2462`。
- 范围：只移植 Chat Completions 兼容路径的 `prompt_cache_key` 传播与 `session_id` API Key 隔离修复。
- 不纳入：5h `ResetsAt` 前后端联动、ops 告警指标、前端 Select 高度、代理有效期/失败回退、README/VERSION/assets/skills 改动。

## Key Changes

- 候选提交按顺序移植或手工等价移植：
  - `d251487da`：`fix(openai): propagate prompt cache key for chat completions`。
- 允许路径：
  - `backend/internal/service/openai_gateway_chat_completions.go`
  - `backend/internal/service/openai_gateway_chat_completions_test.go`
  - `docs/workflow/tasks/upstream-main-prompt-cache-s12.md`
  - `docs/workflow/worker-results/upstream-main-prompt-cache-s12-result.md`
  - `docs/workflow/qa-reports/upstream-main-prompt-cache-s12-qa.md`
  - `docs/workflow/main-log.md`
- 禁止路径：
  - `backend/ent/**`
  - `backend/migrations/**`
  - `frontend/**`
  - `skills/**`
  - `.github/**`
  - `deploy/**`
  - `assets/**`
  - `README*`
  - `knowledge/**`
  - `docs/workflow/status.md`
  - `docs/workflow/spec.md`

## Public APIs / Interfaces

- 不新增数据库字段、migration、Ent schema、公开 DTO 字段或配置项。
- 行为变化：
  - Chat Completions 兼容转 Responses 时，API Key 账号会把非空 `prompt_cache_key` 写入上游 Responses body。
  - 上游 `session_id` 继续由 `prompt_cache_key` 派生，但加入当前 API Key ID 隔离，避免不同 API Key 共用同一 prompt cache session。
  - 如果请求 body 已有非空 `prompt_cache_key`，不覆盖调用方显式值。

## Test Plan

- 基础检查：
  - `git status --short --branch`
  - `git diff --check`
  - 路径审计：确认无 `frontend/`、`backend/ent/`、`backend/migrations/`、`skills/`、`assets/`、`README*`、`knowledge/`、`docs/workflow/status.md`、`docs/workflow/spec.md` 改动。
- 定向后端测试：
  - `go test ./internal/service -run "ForwardAsChatCompletions|PromptCache|Session|OpenAI|ChatCompletions" -count=1`
- 回归测试：
  - `go test ./internal/service -count=1`

## Assumptions

- 优先 `git cherry-pick -x d251487da`；若冲突触及 forbidden path，停止并标记 `DEFERRED`。
- 保留本地模型市场、APIMart 计费、工单、Canvas、Chat/Image Studio、OpenWebUI 和 workflow 文档。
- 本 Sprint 不处理 `16bc87693`、`f20e6bf76` 或 `f5cecea5b`；它们作为后续独立 Sprint 评估。
