### PASS: image-model-tutorials-s223

# QA Report

## Task ID
image-model-tutorials-s223

## Verdict
PASS

## Scope
- Reviewed candidate `aa85faf32c59196d064e95ba3edd907bdeabe2ec` against approved base `a865d8b6eb06048f7cf7e3b983b65cf393197806` in `E:/codex-worktrees/sub2api/image-model-tutorials-s223`.
- Candidate changes exactly three Developer-owned files: migration 224, its focused test, and the Developer result. The worktree and index were clean with no unmerged entries before this QA report.
- Read the S223 contract, migration, migration test, Developer result, and read-only image/video runtime owners.

## Evidence
- Diff reviewed: yes.
- Allowed paths checked: yes.
- Denied paths touched: no.
- Nine distinct, published `图像模型` rows were verified with ordered sorts 2240 through 2248 and exact required slugs/model IDs. The migration ends with `ON CONFLICT (slug) DO NOTHING`.
- Every page has the local host, Bearer authentication, parameter section, cURL, Python `requests`, JavaScript `fetch`, and response/troubleshooting section. Branding and URL scans found no APIMart name/domain, pseudo reference asset, or non-local example host.
- `openai_images.go` confirms GPT/Gemini/Midjourney select the internal asynchronous upstream branch, submit and poll before `buildAPIMartOpenAIImagesResponse` returns the OpenAI-compatible final body. Seedream is deliberately outside that branch; its tutorials correctly document pass-through task polling through `/v1/tasks/{task_id}`.
- Gemini uses `resolution` for 1k/2k/4k examples, limits references to 14, fixes `n` at 1, and Flash only rejects rather than advertises 0.5k. Source checks confirm 14-reference and Seedream reference-plus-output limit 15, Flash search fields, the Midjourney payload whitelist, and resolution normalization.
- Seedream examples extract only `data[0].task_id`, poll `GET /v1/tasks/{task_id}` through completed/succeeded/success, read `data.result.images[0].url[0]`, warn about prompt download, keep reference plus `n` at most 15, and recommend only 2k/4k for 4.5. No root-level task-ID extraction, tier-as-size example, or official GPT transparent-background example was found.

## Commands Run
```text
go test ./migrations -run '^TestImageModelTutorialPages$' -count=10 -> PASS (ok github.com/Wei-Shaw/sub2api/migrations 0.044s)
go test ./migrations -count=1 -> PASS (ok github.com/Wei-Shaw/sub2api/migrations 0.043s)
go test ./cmd/server -run '^$' -count=0 -> PASS (ok github.com/Wei-Shaw/sub2api/cmd/server 0.063s; no tests to run)
rg -n -i 'apimart|api\.apimart\.ai|cdn\.apimart\.ai' backend/migrations/224_image_model_tutorial_pages.sql -> PASS (no matches)
rg -n 'https://ai\.3zapi\.top' backend/migrations/224_image_model_tutorial_pages.sql -> PASS (40 matching lines)
git diff --check a865d8b6eb06048f7cf7e3b983b65cf393197806..aa85faf32c59196d064e95ba3edd907bdeabe2ec -> PASS
git diff --name-only --diff-filter=U; git ls-files -u -> PASS (both empty)
git status --short -> PASS (empty before QA report)
```

## Findings
- No explicit defect found.

## Risks
- This is static repository validation only. No migration, shared or production database, provider endpoint, or credentialed API request was executed. Provider-side behavior and temporary result URL retention remain outside the verified scope.

## Integration Decision
- Integrable: yes. The candidate is within the approved Developer scope; this commit contains only the independent QA report.

## Bug Owner Recommendation
none

## Root Cause
none

## Retest Scope
- None required for this candidate. Re-run the focused migration test and static content checks if the migration or image forwarding semantics change.

## Knowledge Promotion
none
