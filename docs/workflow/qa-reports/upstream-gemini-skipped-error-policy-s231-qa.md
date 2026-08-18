### PASS: upstream-gemini-skipped-error-policy-s231

## Scope

- QA base: `96d54717b`（包含 S231 业务 `5cf6f3fcd` 与 Controller report）。
- 独立 worktree：
  `E:/codex-worktrees/sub2api/upstream-gemini-skipped-error-policy-s231-qa`。
- 业务 diff `7b6515c62..5cf6f3fcd` 精确为四个 contract allowlist 文件；QA 未修改
  产品文件。

## Runtime Evidence

- `go test ./internal/service -run "TestGeminiForward(Native|Messages|AsChatCompletions).*|TestWriteGeminiMappedError_400KeepsUpstreamMessage|TestSkippedErrorPolicyFailoverError_CustomCodeMiss500HasNoSameAccountRetry" -count=10`
  PASS，`6.028s`。
- `go test ./internal/service -count=1` PASS，`77.427s`。
- `go test ./cmd/server -run "^$" -count=1` PASS，`9.062s`。
- focused 覆盖 native、Messages、Chat Completions 的 pool-mode 400、skipped 5xx
  failover、custom-code-miss 400 隐藏、400 message 保真，以及 custom 5xx 不携带
  同账号重试标记。

## Static / Git Evidence

- `gofmt -d` 四个 allowlist 文件：无输出。
- `git diff --check`：PASS。
- `git diff --name-only --diff-filter=U`、`git ls-files -u`：无输出。
- QA worktree `git status --short --branch`：仅 detached HEAD，工作树洁净。
- `git merge-base --is-ancestor ab0fcd1a0 upstream/main`：PASS。
- `git log ab0fcd1a0..upstream/main -- <four owners>`：无后续相关提交。
- 本地业务 patch-id：`e8c34a39abb58e03e4e00f52f646f408d5256af0`；上游
  patch-id：`67b07fc3b752ffe1f1ec43a08de3c28dad0cb60b`。差异符合 contract 记录的
  本地 `errorPolicy`/retry 拓扑适配，不是整包 cherry-pick。

## Protected Main

- 主工作区仍为 `main@3270b165e`，`origin/main@a865d8b6e`，
  `upstream/main@49504adc9`；未 push。
- 八个用户 dirty 文件 patch-id 保持：
  `d665008e...`、`d2b9be6d...`、`efdead98...`、`6e52c0c2...`、
  `c6076fd2...`、`da720d61...`、`228685a0...`、`08385fdd...`。
- 六个未跟踪教程 migration/test SHA256 保持：
  `D7EDF11F...`、`A426D11E...`、`854BBC7B...`、`C9676B55...`、
  `84C47AB0...`、`6D07FA3C...`；`outputs/` 仍未跟踪。

## Findings

- 未发现阻断项或 contract 越界。
- 初始 `-tags=unit` 命令的既有符号冲突已由 contract amendment 处理；默认标签
  focused、完整 service 和 server compile 均真实通过。
- 未调用真实 Gemini/provider、数据库、容器、部署或远端写入。
