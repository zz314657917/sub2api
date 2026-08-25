### DONE: upstream-gemini-tool-schema-s254

# Worker Result

## Task ID

upstream-gemini-tool-schema-s254

## Status

done

## Summary

- 已在本隔离 worktree 行为级移植上游 `19da0f240`：`cleanToolSchema` 递归删除 Gemini 不支持的 `deprecated` 和 `exclusiveMinimum`。
- 标量 enum（字符串、布尔值、JSON 数字、Go 数值和 `nil`）统一转换为 JSON 字符串；任一对象、数组或其他不支持值会使整个 enum 被删除，不会部分透传。
- 保留本地既有 schema 清理、类型大写规范化、工具转换和 Web Search 拓扑；业务与测试变更已提交为 `b57277dc6`。

## Changed Files

- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service_test.go`

## Commands Run

```text
gofmt -w backend/internal/service/gemini_messages_compat_service.go backend/internal/service/gemini_messages_compat_service_test.go -> PASS
Push-Location backend; go test ./internal/service -run "TestCleanToolSchema_(DropsAmbiguousExclusiveMinimumWithoutConversion|RemovesNestedDeprecatedAndNormalizesMixedScalarEnum|DropsEnumWithNonScalarValue)" -count=10 -> PASS
Push-Location backend; go test ./internal/service -run "TestCleanToolSchema|TestConvertClaudeToolsToGeminiTools" -count=1 -> PASS
Push-Location backend; go test ./internal/service -count=1 -> PASS (ok github.com/Wei-Shaw/sub2api/internal/service 64.679s)
Push-Location backend; go test ./cmd/server -run '^$' -count=1 -> PASS
git diff --check -> PASS
rg -n "^(<<<<<<< .+|=======$|>>>>>>> .+)$" backend/internal/service/gemini_messages_compat_service.go backend/internal/service/gemini_messages_compat_service_test.go -> PASS (no matches)
git diff --cached --name-only -> PASS (empty before business staging; business commit contained only the two owner files)
git ls-files -u -> PASS (empty)
```

## Upstream Mapping

- Source `upstream/main@e2d9b823f`, commit `19da0f240`: add `deprecated` removal, normalize scalar enum values through `json.Marshal`, and delete invalid whole enums.
- Local adaptation: retained local `cleanToolSchema` topology at `main@249cbc223`; did not cherry-pick because the contract records `git apply --check` as topology-incompatible.

## Risks

- 无真实 Gemini provider 调用：合同明确禁止，验证限于本地 schema/转换单元与服务测试。
- `json.Marshal` 无法表示的数值（例如 NaN/Inf）会导致 enum 整体删除，符合“不向 Gemini 透传不支持值”的安全策略。

## Knowledge Candidates

- Gemini function declaration enum 只应包含字符串；将可 JSON 编码标量转为 JSON 字符串，并对复合/不可编码成员整体丢弃 enum，可避免上游 schema 拒绝。

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason

- N/A
