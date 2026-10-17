---
type: contract-review
status: approved
verdict: PASS
task_id: upstream-v0216-test-compat-20261015
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
reviewer: independent-contract-reviewer
reviewer_model: gpt-6.1-sol
last_verified: 2026-10-17
reviewed_draft_sha256: C8CE72CF903E4C274DD3BBA02EFE4DE5668E8B595822F48893854A58A69C4332
---
### PASS: upstream-v0216-test-compat-20261015 — 2026-10-17 六文件修订

审阅 draft SHA-256：`C8CE72CF903E4C274DD3BBA02EFE4DE5668E8B595822F48893854A58A69C4332`。本轮独立 Contract Reviewer 仅批准合同范围及验收设计，不代表实现或 QA 已通过。无需修订合同；合同正文及 frontmatter 均未修改，主控可仅批准合同 frontmatter。

## Findings

未发现阻断本次六文件 test-only 修订的问题。

- 原五文件现存 diff 为 12 行新增、17 行删除，均为 helper 重命名、unused import 清理、参数/返回值兼容和移除不存在的 proxy 字段。现有费用、请求内容及两条 proxy 更新路径断言保留。
- 新增 `auth_service_email_bind_test.go` 对应 canonical tagged compile 的六处 `createEmailBindTestUser` 未定义及 stub 接口缺口。helper 应使用当前 Ent Create、传入的 email/username/password hash 和相邻 fixture 的初始化方式，不改变调用输入、断言或并发协调。该文件现有 13 个匹配邮件绑定/发码测试均被新增验收命令覆盖。
- `user_service.go:95,127` 要求 `Update(ctx, user, fields UserUpdateFields)`；当前 stub 在邮件绑定测试文件第 996 行仍为两参数。必须按实际调用 `auth_email_binding.go:78` 的 `{Email: true, PasswordHash: true}` 更新，并保留未选字段及邮箱索引一致性，不能只增加签名后继续整行 clone 覆盖。`Balance`、`TotalRecharged`、`TokenVersion` 不属于可任意写回字段；原 token revocation 测试及断言保持不变。
- 新增 `AdjustBalance(context.Context, int64, float64) (service.BalanceChange, error)`、`SetBalance(context.Context, int64, float64) (service.BalanceChange, error)` 必须对意外调用立即失败，例如 panic；禁止返回零值和 nil 假装成功。本轮 stub 场景不授权新增余额行为。已有真实 Ent 首绑默认额度测试继续保留。
- 验收要求 canonical unit、13 个邮件绑定/发码用例、相关回归、默认 tag service 编译及生产 build，并禁止 overlay、隐藏文件、改 tag、skip 和弱化断言；出现额外范围需求或断言失败则停止，足以守住 test-only 边界。
- Gemini 原正则未直接匹配 `TestCheckErrorPolicy_GeminiAccounts`、`TestHandleGeminiUpstreamError_GoogleOneCapacityExhaustedUsesTierCooldown`；合同已有“discover exact policy names if regex omits any”的强制要求，执行方必须补跑这两个现存名称并记录非零通过数及零 skips。此为现有要求的明确化，无需修改合同。

## Executed Checks

- 只读核对当前合同、Agent Matrix、workflow status/spec、既有评审、Developer BLOCKED 报告及 canonical 编译诊断日志。
- 核对隔离树 HEAD 为指定 base，仅原五个测试文件 dirty；逐行审阅这五文件完整 diff，`git diff --check` 退出 0。
- 核对生产 `UserRepository` / `UserUpdateFields`、repository Update 实现、邮件绑定实际调用、helper 六处调用及 stub；主树与隔离树这四个相关文件 SHA-256 一致。核对 billing/countTokens 当前签名和验收测试名称。
- 未运行编译、测试、浏览器或外部服务；未修改业务/测试文件、合同、workflow 状态，未 stage 或 commit。

## Unverified Risks

第六文件尚未实现，canonical 编译恢复及 13 个邮件绑定用例（含并发与 token revocation）实际结果待独立 Developer/QA 验证。现有编译日志仅证明当前阻断，不能证明修复后无后续错误；继续按 Stop Rules 处理，不扩大 allowlist。QA 六文件 hash seal、生产及依赖未变证据仍为实施后的必需门禁。

## Recommendation

PASS：主控仅批准合同 frontmatter 后，可安排六文件兼容实现及独立 QA；本报告不批准集成、提交或部署。

## 历史评审（2026-10-15，原文保留）

---
type: contract-review
status: approved
verdict: PASS
task_id: upstream-v0216-test-compat-20261015
base_commit: de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d
reviewer: independent-contract-reviewer
reviewer_model: gpt-6.1-sol
last_verified: 2026-10-15
---
### PASS: upstream-v0216-test-compat-20261015

Independent configured reviewer gpt-6.1-sol, agent 01a13ef0-b7b0-73e2-8d74-64fd90db4095.

Reviewed contract SHA-256: `52281201EBE81DD80CBDDC485D227D15E579E593B5CF056A9F1569A642E12E48`. No required amendments.

The five-file test-only scope exactly covers duplicate `stringPtr`, unused import, obsolete billing signatures, two-value countTokens builder, and removed proxy fields. Current API signatures were checked. Test names, business inputs, expected costs, both proxy update paths and assertions must remain. Commands cover nonzero tagged cases, regressions, default-tag compile and production build. No production edits or overlays are authorized.
