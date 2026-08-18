### PASS: upstream-codex-fingerprint-convergence-s233

## Findings

未发现明确问题。独立 QA 从 `775d4f7ed`（基线 `de08968f1`）执行；业务提交只包含 S233 contract 允许的 8 个产品/测试文件，没有引入上游专属 PAT service、resolver 或其他 denied path。

## Executed Checks

- `go test ./internal/service -run 'TestCodex|TestFetchCodexModelsManifest|TestOpenAICodex' -count=10`：PASS，28.47s。
- `go test ./internal/repository -run 'TestOpenAIOAuthServiceSuite' -count=10`：PASS，14.55s。
- `go test ./internal/service -count=1`：PASS，70.20s。
- `go test ./internal/repository -count=1`：PASS，5.39s。
- `go test ./cmd/server -run '^$' -count=1`：PASS，13.66s（无测试，仅完成编译）。
- `gofmt -l`（contract 指定 8 个文件）：PASS，无输出；`git diff --check de08968f1..775d4f7ed`：PASS；合计静态检查 0.261s。
- 精确 allowlist：PASS。差异文件恰为以下 8 个：
  `backend/internal/service/openai_codex_identity.go`
  `backend/internal/service/openai_codex_identity_test.go`
  `backend/internal/repository/openai_oauth_service.go`
  `backend/internal/repository/openai_oauth_service_test.go`
  `backend/internal/service/openai_codex_models_service.go`
  `backend/internal/service/openai_codex_models_service_test.go`
  `backend/internal/service/account_usage_service.go`
  `backend/internal/service/openai_codex_version_consistency_test.go`
- 冲突和索引：PASS。`git ls-files -u` 为空，QA worktree staged index 为空，未发现冲突标记。
- 上游 provenance：PASS。`a34123959` 是 `upstream/main` 的祖先，提交标题为 `fix(fingerprint): align credential-face identity with the real client and de-drift models version`，提交时间 `2026-08-18T15:20:08+08:00`。
- patch-id：上游 `a34123959` 为 `fa37558f618c93e2c1747dced1333067b6c46334`，本地 `775d4f7ed` 为 `9ebde5e6633eb301ade3368b2f9596eed68c49b2`。两者不同，符合本地缺失 PAT/resolver 拓扑后的行为级手工适配，不应强制整提交合并。
- 版本/旧指纹检查：`openAICodexProbeVersion` 在业务树中无引用；`codex-cli/0.91.0` 在业务树中无引用；`client_version` query 保留客户端值、低版本 `Version` header 回退的测试通过。
- 主线 dirty/untracked 保护：QA 前后主线 `F:/mcplugins/sub2api` 的默认 `git status --short` 列表一致（10 个已修改文件、6 组迁移/测试未跟踪项及 `outputs/`），主线 staged index 为空。QA 未修改主线用户内容。

## Unverified Risks

- 按 contract 禁止真实 provider、数据库、部署和容器操作；本结论覆盖本地单元/包测试和编译，不替代真实 OpenAI OAuth 或 Codex 上游运行态验证。
- `a34123959` 的 `openai_codex_pat_service.go` 和 resolver 拓扑未导入；若后续本地架构补齐这些 owner，应另开任务评估，而不是扩大 S233 scope。

## Contract Compliance

- OAuth exchange/refresh 使用 canonical Codex `User-Agent + originator` 配对，移除旧 `codex-cli/0.91.0` 半身份。
- manifest 空/过时 header version 使用本地 `codexCLIVersion`，`client_version` query 保持透传。
- quota probe 与 manifest 默认版本共用 `codexCLIVersion`，重复 `openAICodexProbeVersion` 已移除。
- 未执行任何 push、部署、容器、数据库或真实 provider 操作。

## Recommendation

可继续合入主线。保留运行态 provider 验证作为后续独立验收风险。
