### DONE: upstream-gemini-skipped-error-policy-s231

## Summary

按批准 contract 手工移植上游 `ab0fcd1a0` 的 Gemini `ErrorPolicySkipped`
语义到本地单体网关。业务提交：`5cf6f3fcd5d7c9df5700bf5a06908b156903aab0`。

- native：池模式不可 failover 4xx 保留真实状态码和响应体；自定义错误码未命中
  的不可 failover 状态返回 500 固定文案；可 failover 状态仍返回
  `UpstreamFailoverError`。
- Messages / Chat Completions：池模式 400 保留真实映射和已脱敏上游 message；
  custom-code miss 隐藏上游细节；可 failover skipped 状态换号，池模式同账号重试
  标记仍由既有 `pool_mode_retry_status_codes` 控制。
- 保留旧 `poolModeSkippedFailoverError` 包装以兼容现有本地测试，新的生产路径统一
  使用 `skippedErrorPolicyFailoverError`。

## Changed Files

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_error_policy_test.go`
- `backend/internal/service/gemini_error_policy_skipped_write_test.go`

## Verification

- `go test ./internal/service -run "TestGeminiForward(Native|Messages|AsChatCompletions).*|TestWriteGeminiMappedError_400KeepsUpstreamMessage|TestSkippedErrorPolicyFailoverError_CustomCodeMiss500HasNoSameAccountRetry" -count=10`：PASS，9 个 S231 场景，约 `5.890s`。
- `go test ./internal/service -count=1`：PASS，`66.068s`。
- `go test ./cmd/server -run "^$" -count=1`：PASS，`5.508s`。
- `gofmt -d` 四个 allowlist 文件：PASS。
- `git diff --check`：PASS。
- `git diff --name-only --diff-filter=U` 与 `git ls-files -u`：无冲突、无 unmerged index。
- `git merge-base --is-ancestor ab0fcd1a0 upstream/main`：PASS；直接 `git apply --check`
  已确认原 patch 因本地 `errorPolicy`/retry 拓扑差异不能整包应用。

## Contract Compliance

- 产品差异严格为上述四个 allowlist 文件；无依赖、迁移、frontend、数据库、容器、
  provider、部署、push 或用户 dirty/untracked 路径变化。
- 真实 Gemini/provider、共享数据库、容器和远端均未访问或写入。
- 初始 `-tags=unit` focused 命令复现了仓库既有无关符号冲突；按已批准 amendment，
  S231 回归改为默认构建标签并已通过。

## Risks / Follow-up

- 独立 QA 尚未执行；主线集成前必须在从该业务提交创建的独立 QA worktree 重跑
  focused、完整 service、server compile、scope/provenance 和用户改动保护门禁。
