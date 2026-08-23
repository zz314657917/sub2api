# Upstream Chat File Input S246

## Task ID

`upstream-chat-file-input-s246`

## Role

Developer Worker and independent QA Worker both use `gpt-5.6-terra` in
separate executions and evidence paths. Codex is Planner and Final Evaluator.
The implementation must follow the approved contract without widening scope.

## Goal

Behaviorally adapt upstream source `4d4a0be1a` as merged by `6244090c1` so a
Chat Completions `type:"file"` content part is converted into a Responses
`type:"input_file"` part instead of being silently dropped.

## Success Criteria

- A Chat file part containing `file_data` forwards `filename` and `file_data`
  as an `input_file` while preserving surrounding text-part order.
- A Chat file part containing `file_id` forwards it as `input_file.file_id`.
- A file part with neither `file_data` nor `file_id` is skipped, matching the
  existing empty-media behavior and avoiding an unusable upstream part.
- If both supported payload fields are present, the converter preserves both;
  no validation or normalization outside the upstream behavior is introduced.
- Existing text/image conversion, empty-content fallback, Responses output
  conversion, custom tools, and streamed tool-name `omitempty` behavior remain
  unchanged.
- Focused regressions pass repeatedly; complete `apicompat`, service/server
  compilation, formatting, scope, provenance, and protected-primary gates pass.

## Frozen Base And Provenance

- Frozen product base: local
  `main@da825994fa8376d5769452fa48369b3f10bec012`.
- Upstream source: `4d4a0be1ad2a994bcf8ca175444a9eb2facb28b4`.
- Upstream merge: `6244090c1c0efb49a71b0f29529988249103db7d`.
- Upstream audit tip:
  `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- Source and merge are ancestors of the audit tip and touch the same three
  owners. Their raw patch IDs differ because the merge resolved against newer
  `types.go` history; the merge first-parent behavior is authoritative.
- The merge patch applies to the converter and test owner but conflicts in
  local `types.go`, which contains additional local compatibility work including
  S239 streamed tool-name omission. Manually add only the file-part DTO fields
  while preserving all local fields and JSON tags.

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`,
  `docs/workflow/agent-matrix.md`, `docs/workflow/spec.md`, and this contract.
- Owners:
  `backend/internal/pkg/apicompat/types.go`,
  `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`, and
  `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`.

## Allowed Paths

- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `docs/workflow/worker-results/upstream-chat-file-input-s246-result.md`
- `docs/workflow/qa-reports/upstream-chat-file-input-s246-qa.md`

## Denied Paths

- All other backend, frontend, generated, schema, migration, dependency,
  lockfile, configuration, deployment, container, and workflow product paths.
- `docs/workflow/status.md`, `docs/workflow/spec.md`,
  `docs/workflow/main-log.md`, this contract, and all `knowledge/**` paths are
  Controller-owned and denied to workers.
- All user-owned dirty and untracked paths in the primary worktree, including
  the eleven current Pixel Cafe paths and `outputs/`.
- Remote writes, push, force operations, history rewrites, real provider
  traffic, shared/production data, and browser automation.

## Constraints

- Keep the existing DTO and converter topology; do not redesign content-part
  parsing, introduce validation, or change unknown-part behavior.
- Preserve the local `ChatFunctionCall.Name` `omitempty` behavior and all S242/
  S243 custom-tool and replay compatibility fields in `types.go`.
- Do not add `file_url`, upload/download support, MIME inspection, size limits,
  base64 validation, gateway routes, billing behavior, or security-audit
  extraction under this contract.
- Do not install or update dependencies, call a real provider, or touch shared
  services.
- Do not stage, overwrite, revert, or format unrelated work.

## Acceptance Commands

From `backend/` in the isolated worktree:

```powershell
go test ./internal/pkg/apicompat -run '^TestChatCompletionsToResponses_(FilePartFileData|FilePartFileID|EmptyFilePartSkipped)$' -count=10
go test ./internal/pkg/apicompat -count=1
go test ./internal/service -run '^$' -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/pkg/apicompat/types.go internal/pkg/apicompat/chatcompletions_to_responses.go internal/pkg/apicompat/chatcompletions_responses_test.go
```

From the worktree root:

```powershell
git diff --check
git diff --cached --name-only
git ls-files -u
git merge-base --is-ancestor 4d4a0be1ad2a994bcf8ca175444a9eb2facb28b4 upstream/main
git merge-base --is-ancestor 6244090c1c0efb49a71b0f29529988249103db7d upstream/main
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend/internal/pkg/apicompat/types.go backend/internal/pkg/apicompat/chatcompletions_to_responses.go backend/internal/pkg/apicompat/chatcompletions_responses_test.go
```

The Controller must additionally verify the exact business/evidence commit
allowlists, source/merge first-parent scope, local S239 `omitempty` preservation,
empty index/conflict state, and primary-worktree protected snapshot.

The protected primary-worktree patch ID is scoped to these eleven user-owned
paths only:

- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/SceneFallback.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/features/pixelCafe/renderer/assetManifest.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/types/pixelCafe.ts`

Their combined stable patch ID must remain
`370ac77de0e2f530ab652b99fb3eb35e809f4c84`. The primary staged/unmerged index
must remain empty, and `outputs/` must retain its two pre-existing untracked
files.

## Output

- Developer produces one business commit containing only the three product/test
  paths and one separate evidence commit containing only
  `docs/workflow/worker-results/upstream-chat-file-input-s246-result.md`.
- The Developer report first line must be exactly
  `### DONE: upstream-chat-file-input-s246`,
  `### BLOCKED: upstream-chat-file-input-s246`, or
  `### FAILED: upstream-chat-file-input-s246`.
- Independent QA may modify only
  `docs/workflow/qa-reports/upstream-chat-file-input-s246-qa.md`; its first line
  must be exactly `### PASS: upstream-chat-file-input-s246`,
  `### FAIL: upstream-chat-file-input-s246`, or
  `### BLOCKED: upstream-chat-file-input-s246`.
- Reports list changed files, commands run, key output, risks, contract
  compliance, and `knowledge_candidates` without unrelated long logs.

## Stop Rules

- Stop if `gpt-5.6-terra` is unavailable; do not silently replace the model.
- Stop if implementation requires any path outside the allowlist, file upload/
  download semantics, validation policy, gateway/security-audit changes,
  dependencies, frontend/schema changes, browser automation, or external state.
- Stop if the focused selector discovers no tests, a baseline failure is owned
  outside this contract, local S239 `omitempty` behavior changes, or any
  protected-primary path changes unexpectedly.

## Budget

- worker_mode: native `gpt-5.6-terra`
- qa_worker_mode: native `gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- developer_max_budget_usd: `0.10`
- qa_max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Status

`contract-approved`

## Worker Output

Same requirements as `Output`; this compatibility heading is retained for the
worker dispatcher.
