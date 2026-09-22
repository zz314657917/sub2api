### FAIL: upstream-v027-remaining

# Contract Review

## Task ID

`upstream-v027-remaining`

## Contract Checked

`docs/workflow/tasks/upstream-v027-remaining.md` at base commit
`d6e6c34718e6eee6388391b346f40e1492e81d57`.

## Verdict

不允许分派 Developer。上游 25 个候选提交均能从本地 `upstream/main`
对象库解析，基线为干净工作树，且 Controller 已确认
`go test ./internal/pkg/apicompat -count=1` 与 `go build ./...` 通过；阻断原因
是当前合同不能把大批量跨域改动约束为可复核的实现和验收单元。

## Findings

1. Success Criteria 同时覆盖 DeepSeek 请求重写、Antigravity/Gemini 协议、OpenAI
   manifest 与取消后的亲和性持久化、CN 限流，以及 15 个前端改动。它们没有共享
   回滚或验收边界，不能作为一个 Developer ownership。
2. 后端 Allowed Paths 与已验证的本地/上游 topology 不一致。`881ab1b0c`
   实际涉及请求规范化和 apicompat 新 helper；本地规范化 owner 是
   `backend/internal/service/openai_gateway_service.go:7354`，并且必须保留
   `previous_response_id` 删除和 `store=false` 行为。`d7ee1ab6b` 的本地
   `bindHTTPResponseAccount` 也在该单体文件第 6668 行，但合同未把它作为
   独立的成功标准或回归测试 owner。
3. 合同列出的
   `backend/internal/service/antigravity_gateway_gemini.go` 与
   `backend/internal/service/antigravity_gateway_streaming.go` 在基线不存在。
   本地 Gemini 转发 owner 是
   `backend/internal/service/antigravity_gateway_service.go:2114`，而 SSE 兼容
   还可能涉及 `gemini_messages_compat_service.go` 和 handler。当前“经 controller
   approval 后可替换”的表述是开放式 allowlist，不能由 Developer 自行判断。
4. CN 限流必须按本地 `cnAccountIsCodingPlan` 判定，而不是上游
   `Account.IsCodingPlan()`；该函数位于
   `backend/internal/service/cn_provider_quota_service.go`。该现有 owner 没有
   被明确允许或禁止修改，也没有定义 403 分类、持久化失败、普通 403、并发 403 与
   reset 语义的逐项断言。
5. 前端范围“listed UI commits touched 的 production paths 和相关测试”不是路径
   allowlist；验收中的 `pnpm exec vitest run <changed tests>` 也不是可执行命令。
   同时合同要求 browser/runtime evidence，却未定义候选页面、启动命令、专用 profile
   路径、观察项和清理检查。按仓库浏览器进程边界，这些缺失会使 QA 无法完成验收。
6. `focused service/handler test selectors exercising named scenarios` 同样是占位描述；
   没有映射到具体 `-run` 正则或新增测试名，不能区分基线失败与本任务失败。当前
   `status.md` 仍指向已批准的另一项 S302；不得据此把本任务从 draft 推进。

## Required Amendments

1. 将合同拆为至少四份后端合同，每份独立 review、Developer、QA 和报告：
   - A：`881ab1b0c` + `acc05620c`，明确
     `responses_tool_output_media*.go`、`openai_gateway_service.go` 及精确测试文件；
     以 `httptest` localhost 上游断言 plain、多模态、多 tool-call 连续性、数值精度、
     `store=false` 和 `previous_response_id` 删除。
   - B：`0f4d8acaa`、`f79b8bf96`、`da74bf13a`，先由 Planner 写出
     `antigravity_gateway_service.go`、Gemini handler/compat 的实际函数级映射及
     精确 helper/test 路径；以 fake HTTP/SSE 测试断言裸模型 thinking 变体、SDK
     SSE 注释兼容和 account-mapped model listing。
   - C：`18bfa4bf2`、`2f16e0984`、`31f3003ff`、`d7ee1ab6b`，明确
     strict-provider role、manifest parse/validation 和
     `openai_gateway_service.go:6668` 的 bounded `context.WithoutCancel` 持久化；
     不得引入上游不存在于本地的 response-owner 子系统，必须保留现有绑定/guard，
     并测试 client cancel、超时预算和现有 `BindResponseAccount` 的单一持久化写入。
   - D：`db8692d67`，明确 `ratelimit_cn_providers.go`、`ratelimit_service.go`，并仅在
     必要时把 `cn_provider_quota_service.go` 以函数级理由加入 allowlist。测试需断言
     `cnAccountIsCodingPlan`、配额耗尽 403、普通 403、并发 403、快照 reset、无快照和
     `SetRateLimited`/`SetTempUnschedulable` 失败分支。
2. 前端 15 个候选提交另立一个或按 feature 拆分的合同。每份必须逐项写出生产文件和
   test 文件，声明“已等价/不适用/延后”的判定证据，并以真实 Vitest 文件路径替换
   `<changed tests>`。要求 browser evidence 时，同时列出页面、操作、断言、任务专用
   profile、启动 PID 和 `playwright-cli close` 后的精确清理检查。
3. 每份合同必须列出可直接复制执行的命令：精确 `go test ... -run '...' -count=1`、
   `go build ./...`、必要的 `gofmt -w <allowlist>`、`git diff --check`、允许路径
   diff 检查和 unmerged-index 检查；前端合同另列精确 Vitest、`pnpm exec vue-tsc -b`
   与 `pnpm build`。基线比较命令及其预期结果也必须固定。
4. 修订后才可把相应合同 frontmatter 改为 `status: approved`、
   `review_verdict: PASS`；本审核不授权 merge、cherry-pick、commit、push、部署、
   容器、数据库或主工作树操作。

## Gate Checks

- success_criteria_testable: `no`
- allowed_paths_explicit: `no`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `no`
- worker_model_confirmed: `yes` (`gpt-5.6-terra` for independent Developer and QA)
- base_commit_confirmed: `yes` (`d6e6c34718e6eee6388391b346f40e1492e81d57`, current HEAD)
- upstream_refs_confirmed: `yes` (`upstream/main` is `1c0a69c0...`; all named objects resolved)
- openspec_traceable: `not-applicable`

## Approval

当前合同保持 `draft` / `pending`。完成上述修订后，逐份重新提交独立 contract review。
