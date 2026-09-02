### PASS: upstream-v0185-codex-bootstrap-s289

# QA Report

## Task ID
upstream-v0185-codex-bootstrap-s289

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-v0185-codex-bootstrap-s289.md`
- `docs/workflow/contract-reviews/upstream-v0185-codex-bootstrap-s289-review.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`; protected dirty aggregate hash remains `0e467987fd7aec5fc451983bdb8f8216f97ba69c`
- commands run:
```text
backend: go test ./internal/handler -run 'TestNormalizeCodex(Delegation|Automation)Bootstrap' -count=10 -> PASS
backend: go test ./internal/handler -count=1 -> PASS (27.540s)
backend: go test ./cmd/server -run '^$' -count=1 -> PASS
backend: go build ./... -> PASS
root: gofmt -d backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_codex_bootstrap_test.go -> no output
root: git diff --check -- backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_codex_bootstrap_test.go -> PASS
root: git diff --name-only --diff-filter=U -> no output
```
- manual checks:
```text
integration timing -> normalization runs after JSON validity and before HTTP function_call_output context validation
strict candidates -> non-empty previous_response_id, call/reference anchors, non-empty or non-string call_id, mixed/unknown call output and recursively duplicated JSON members all retain the original body
delegation envelope -> exact unnamespaced, attribute-free XML root with one non-empty source_thread_id and input is required
automation payload -> codex_app.automation_update plus exact memory path, safe ID, valid last-run and non-empty prompt is required
payload preservation -> Decoder.UseNumber retains large integer lexemes; the input slice is replaced at its original index; a normalized body is idempotent
```

## Findings
- 未发现明确业务问题。实现与上游 `1be69e56`、`421a83282` 的 normalizer 逻辑一致；本地接入点满足在原有关联预校验前转换的合同要求。
- 新增测试覆盖了数值精度、重复 `type` JSON 成员、主要 context 拒绝、完整 XML 负例、顺序/幂等、automation CRLF 和非法 last-run。递归重复成员及全部拒绝分支另由代码审查确认；上游的更宽负例矩阵未逐条复制到本地，属于低风险测试覆盖缺口，不阻塞本合同验收。

## Bug Owner Recommendation
`none`

## Root Cause
`none`

## Retest Scope
- 无修复项。后续若扩展 normalizer 支持的 Codex 工具或请求形态，应补充对应的负例矩阵并重跑完整 handler 测试。

## Knowledge Promotion
- `none`

## Unverified Risks
- 真实 OpenAI provider、Responses WebSocket、数据库、容器、部署和浏览器 smoke 不属于本 Sprint，未执行。
