---
task_id: upstream-gemini-skipped-error-policy-s231
phase: contract-approved
base: 6d14a6dd11cc7a572bbf56b6187ac25054a3bd1d
source: ab0fcd1a0e6fa8f9974ba21c6c3478bebdd07be4
upstream_ref: 49504adc9
qa_mode: runtime
---

# Task Contract

## Role

- Planner / Final Evaluator: Codex Controller
- Generator: isolated S231 implementation worktree
- Evaluator: separate clean S231 QA worktree and final Controller review

## Goal

按本地 Gemini 网关拓扑手工移植上游 `ab0fcd1a0` 的
`ErrorPolicySkipped` 错误语义修复，不直接 cherry-pick：

- `Skipped` 只跳过账号状态惩罚，不跳过可重试状态的换号；池模式继续保留既有
  `pool_mode_retry_status_codes` 同账号重试标记，自定义错误码未命中不获得该标记。
- 池模式遇到不可 failover 的上游 4xx 时，Gemini native 保留真实状态码和响应体，
  Messages 与 Chat Completions 按真实状态码映射，不再硬改成 500/502。
- 自定义错误码未命中且状态不可 failover 时，三协议统一返回 500 和固定文案
  `Upstream gateway error`，上游细节只进入现有 ops 日志。
- Gemini 400 映射响应优先回传已脱敏的上游 message，帮助客户端定位确定性参数错误。

## Success Criteria

- Native、Messages、Chat Completions 三条协议均覆盖 pool-mode 400、pool-mode 5xx、
  custom-code-miss 400 和 custom-code-miss 5xx 的写出/换号契约。
- 可 failover 的 `Skipped` 错误不写客户端响应；不可 failover 的 pool-mode 4xx 不改变
  上游状态语义；custom-code-miss 不泄露上游错误正文。
- `ErrorPolicyMatched`、`ErrorPolicyTempUnscheduled`、`ErrorPolicyNone`、Google project
  配置错误、count_tokens fallback、现有重试与账号状态处理保持不变。
- focused tests x10、完整 service 回归、server 编译、格式、范围、provenance、冲突/index
  和用户改动保护门禁全部通过。

## Allowed Paths

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_error_policy_test.go`
- `backend/internal/service/gemini_error_policy_skipped_write_test.go`
- `docs/workflow/worker-results/upstream-gemini-skipped-error-policy-s231-result.md`
- `docs/workflow/qa-reports/upstream-gemini-skipped-error-policy-s231-qa.md`

## Denied Paths

- 其他 gateway/handler、调度、计费、账号 schema、frontend、migrations、dependencies、
  Docker/containers、数据库、真实 provider、部署、push、远端写入
- 用户已有 dirty/untracked 文件、`knowledge/**`、`outputs/**`，除本合同列出的 workflow
  report 外不得修改

## Constraints

- 基线固定为 `main@6d14a6dd1`；上游来源必须可达最新抓取的
  `upstream/main@49504adc9`。
- 上游原 patch 在本地 `gemini_messages_compat_service.go` 和测试拓扑无法直接 apply，
  必须保留本地预计算 `errorPolicy`、count_tokens fallback、retry loop 和现有 ops 语义。
- 不调用真实 Gemini/provider，不写共享或生产状态，不安装或升级依赖。
- Controller review 与 QA 必须从不同 clean worktree 执行；QA 不修改产品文件。

## Acceptance Commands

```powershell
Push-Location backend
go test -tags=unit ./internal/service -run "TestGeminiForward(Native|AsChatCompletions|Messages)_.*Skipped|TestWriteGemini(MappedError|ChatCompletionsMappedError)_400KeepsUpstreamMessage|TestSkippedErrorPolicyFailoverError|TestGeminiErrorPolicyIntegration" -count=10
go test ./internal/service -count=1
go test ./cmd/server -run "^$" -count=1
gofmt -d internal/service/gemini_messages_compat_service.go internal/service/gemini_chat_completions_compat_service.go internal/service/gemini_error_policy_test.go internal/service/gemini_error_policy_skipped_write_test.go
git diff --check
git diff --name-only --diff-filter=U
git ls-files -u
git merge-base --is-ancestor ab0fcd1a0 upstream/main
Pop-Location
```

## Output

- Generator report:
  `docs/workflow/worker-results/upstream-gemini-skipped-error-policy-s231-result.md`
- QA report:
  `docs/workflow/qa-reports/upstream-gemini-skipped-error-policy-s231-qa.md`
- 两份报告必须记录首行 verdict、changed files、commands、风险、contract compliance 和
  upstream provenance。

## Stop Rules

- 任一协议仍将 pool-mode 不可 failover 4xx 硬改为 5xx、custom-code-miss 泄露上游正文、
  或可 failover `Skipped` 未换号，立即判定 FAIL。
- `Matched`/`TempUnscheduled`/`None`、账号惩罚、重试、count_tokens、Google project 400
  语义发生非合同变化，立即停止并回到 contract review。
- 出现 denied path、依赖、真实 provider、数据库、容器、部署、远端、冲突或
  unmerged-index 变化，立即停止。
