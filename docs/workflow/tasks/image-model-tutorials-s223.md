---
task_id: image-model-tutorials-s223
phase: contract-approved
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Image Model Tutorials S223

## Task ID

image-model-tutorials-s223

## Role

You are the independent `gpt-5.6-terra` Generator worker. Execute only this
approved documentation-data contract and do not expand into gateway behavior,
frontend layout, deployment, or database execution.

## Goal

Add nine published Chinese tutorial pages to the existing `tutorial_pages`
catalog for the image models shown in the user's screenshots. The pages must
use the site's own API URL and actual local gateway behavior while borrowing
only the reference site's information hierarchy, never its brand or copy.

## Success Criteria

- Migration `224_image_model_tutorial_pages.sql` inserts exactly one published
  page for each model below, all in category `图像模型`, with stable distinct
  sort orders and collision-safe `ON CONFLICT (slug) DO NOTHING` behavior:
  - `gpt-image-2`
  - `gpt-image-2-official`
  - `gemini-3-pro-image-preview`
  - `gemini-3-pro-image-preview-official`
  - `gemini-3.1-flash-image-preview`
  - `gemini-3.1-flash-image-preview-official`
  - `midjourney`
  - `doubao-seedance-4-0`
  - `doubao-seedance-4-5`
- Each page contains the exact model ID, endpoint, Bearer authentication,
  parameters, cURL, Python, and JavaScript examples, response handling, and
  concise common-error guidance. Examples use only `https://ai.3zapi.top` and
  placeholder text such as `YOUR_API_KEY`.
- GPT, Gemini, and Midjourney pages describe the local synchronous experience:
  the gateway waits for the upstream task and returns an OpenAI-compatible
  final body shaped like `{"created": ..., "data": [{"url": ...}]}`. They must
  not instruct users to poll an upstream task returned by the submit call.
- Seedream pages describe the current pass-through asynchronous experience:
  `POST /v1/images/generations` returns a task identifier and users query
  `GET /v1/tasks/{task_id}` until completion. They must show how to read the
  completed image URL from `data.result.images[0].url[0]` and warn that result
  URLs should be downloaded promptly.
- Shared local parameter facts are accurate:
  - GPT and Gemini use `POST /v1/images/generations`.
  - Midjourney uses `POST /v1/midjourney/generations` and may document only the
    locally parsed fields `size`, `version`, `speed`, `quality`, `stylize`,
    `chaos`, `weird`, `stop`, `niji`, `raw`, `tile`, and `image_urls`.
  - Gemini Pro documents `1k`, `2k`, and `4k`, with at most 14 reference images.
  - Gemini Flash documents the local `1k`, `2k`, and `4k` behavior, at most 14
    reference images, plus `google_search` and `google_image_search`. Do not
    advertise `0.5k`, because the current local payload normalizer does not
    preserve that tier.
  - Seedream enforces `reference image count + n <= 15`; the 4.5 page recommends
    only `2k` and `4k`.
- Standard and official variants are separate pages. The official GPT page may
  document `quality`, `background`, `moderation`, `output_format`,
  `output_compression`, and `mask_url`; the standard GPT page must not imply
  those official-only options are guaranteed.
- A focused Go test reads the embedded migration and proves all nine slugs and
  model IDs are present, pages are published, the local API host is used,
  collision handling is `DO NOTHING`, and forbidden reference branding or host
  strings are absent.

## Context

- Repo: `F:/mcplugins/sub2api`.
- Approval base: `a865d8b6eb06048f7cf7e3b983b65cf393197806`.
- Existing table and tutorial seed pattern:
  `backend/migrations/150_add_tutorial_pages.sql`.
- Runtime facts come from `backend/internal/service/openai_images.go` and
  `backend/internal/handler/openai_videos.go`; these files are read-only for
  this task.
- The reference documentation is layout and field research only. Do not copy
  its brand, domain, dashboard instructions, key-management wording, support
  links, or long verbatim prose.
- The migration is source code only. Do not execute it against shared,
  production, or user databases.

## Allowed Paths

- `backend/migrations/224_image_model_tutorial_pages.sql`
- `backend/migrations/image_model_tutorial_pages_test.go`
- `docs/workflow/worker-results/image-model-tutorials-s223-result.md`

## Denied Paths

- Existing migrations, `backend/internal/**`, `frontend/**`, configuration,
  dependencies, lockfiles, generated files, routes, services, and database
  runtime state.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, task contract edits,
  QA reports, `knowledge/**`, global memories, user account-modal changes, and
  `outputs/**`.
- Containers, deployment, push, shared/production databases, or external API
  calls with a real credential.

## Constraints

- Work only in the isolated S223 worktree created from the approval base.
- Preserve the existing tutorial CMS and renderer. Do not add frontend fallback
  copies; these pages belong in Tutorial Management through the migration.
- Use PostgreSQL dollar-quoted Markdown bodies and ASCII code blocks where
  practical. Chinese tutorial prose is expected and must remain valid UTF-8.
- Keep examples syntactically usable. Python should use `requests`; JavaScript
  should use native `fetch`. Do not add SDK dependencies.
- Use `ON CONFLICT (slug) DO NOTHING` so an administrator's existing page with
  the same slug is never overwritten.
- Do not claim live provider verification. This task validates documentation
  against repository behavior and static migration content only.

## Acceptance Commands

```powershell
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Set-Location E:/codex-worktrees/sub2api/image-model-tutorials-s223/backend
go test ./migrations -run '^TestImageModelTutorialPages$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S223 migration tutorial test failed' }
go test ./migrations -count=1
if ($LASTEXITCODE -ne 0) { throw 'S223 complete migrations package failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S223 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/image-model-tutorials-s223
rg -n -i 'apimart|api\.apimart\.ai|cdn\.apimart\.ai' backend/migrations/224_image_model_tutorial_pages.sql
if ($LASTEXITCODE -eq 0) { throw 'S223 migration contains forbidden reference branding or host' }
rg -n 'https://ai\.3zapi\.top' backend/migrations/224_image_model_tutorial_pages.sql
if ($LASTEXITCODE -ne 0) { throw 'S223 migration is missing the local API host' }
git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S223 diff check failed' }
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S223 conflict or unmerged index found'
}
```

## Output

- Write
  `docs/workflow/worker-results/image-model-tutorials-s223-result.md` with the
  required first-line verdict.
- Commit only the allowed migration, test, and worker report. Include changed
  files, commands, key output, known documentation risks, contract compliance,
  and `knowledge_candidates`.
- Report `BLOCKED` instead of changing gateway code if current behavior makes a
  documented path impossible.

## Stop Rules

- Stop if the nine screenshot model IDs cannot be represented by the existing
  tutorial schema or require frontend/backend CRUD changes.
- Stop if implementation requires modifying an existing migration, gateway
  behavior, dependency, lockfile, shared/production database, deployment, or
  user-owned files.
- Stop after two failed implementation rounds; do not integrate, push, deploy,
  or remove unrelated worktrees and branches.

## Budget

- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Contract Review

`PASS / contract-approved` (2026-08-17 12:13 +08:00): migration slot 224 is
free; `tutorial_pages` already provides published CMS pages and public listing,
so no CRUD or Vue source change is required. Repository inspection confirms the
local endpoint/model split, 14-image Gemini limit, 15 total Seedream limit,
Midjourney parsed fields, and internal polling for GPT/Gemini/Midjourney. It
also confirms Seedream is currently forwarded without that polling branch, so
its pages must use `/v1/tasks/{task_id}`. The contract excludes unsupported
Gemini Flash `0.5k`, reference branding, provider calls, live DB execution, and
all user-owned dirty paths.
