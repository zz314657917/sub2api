### DONE: image-model-tutorials-s223

## Task ID
image-model-tutorials-s223

## Status
done

## Summary
- Added migration 224 with nine independently published Chinese image-model tutorials in the existing `tutorial_pages` catalog.
- R1 corrected tier handling to use `resolution` with pixel/ratio `size`, fixed Seedream envelope task-ID extraction and terminal-state polling, removed the pseudo reference asset, and replaced the official GPT transparent background example.
- Added focused embedded-migration checks for all required slugs, model IDs, local endpoint and authentication content, synchronous/asynchronous response guidance, limits, collision safety, forbidden branding/hosts, `resolution`, envelope task IDs, and tier-as-size regressions.

## Changed Files
- `backend/migrations/224_image_model_tutorial_pages.sql`
- `backend/migrations/image_model_tutorial_pages_test.go`
- `docs/workflow/worker-results/image-model-tutorials-s223-result.md`

## Commands Run
```text
R1: go test ./migrations -run '^TestImageModelTutorialPages$' -count=1 -> PASS
R1: go test ./migrations -count=1 -> PASS
R1: go test ./cmd/server -run '^$' -count=0 -> PASS
R1: rg -n -i 'apimart|api\.apimart\.ai|cdn\.apimart\.ai' backend/migrations/224_image_model_tutorial_pages.sql -> no matches
R1: rg -n 'https://ai\.3zapi\.top' backend/migrations/224_image_model_tutorial_pages.sql -> matches found
R1: git diff --check -> PASS
```

## Test Output
```text
R1: ok github.com/Wei-Shaw/sub2api/migrations 0.766s
R1: ok github.com/Wei-Shaw/sub2api/migrations 0.037s
R1: ok github.com/Wei-Shaw/sub2api/cmd/server 0.065s [no tests to run]
```

## Risks
- Documentation is validated only against the repository's current static gateway implementation. No shared database, live provider, or external API was called.
- Seedream result URLs are documented as time-sensitive according to the approved contract; retention duration is not verified here.

## Knowledge Candidates
- None. The gateway behavior described here remains implementation-coupled and should be rechecked if the image forwarding path changes.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.
