### BLOCKED: upstream-routed-codex-model-catalog-s265

Source range: `22e1b8144..2abce6503` (manual adaptation retry).

No product commit was created. Workflow evidence was recorded in commit
`8d462fa1b71f2b72ebd7261e5b62c8521f2e5e6c`. The task worktree is clean at
`1696c7281` apart from that evidence commit.

Attempted command:

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-routed-codex-model-catalog-s265/backend
go test ./internal/service ./internal/handler ./internal/server/routes -run '^$'
```

The build stopped before tests because the clean local baseline lacks symbols
required by the routed catalog implementation, including `CompositeModelRoute`,
`DetectModelPlatform`, `IsGPTImageGenerationModel`,
`grokSupportsReasoningEffort`, `grokSupportsXHighReasoningEffort`, and
`claude.EffortLevelsForModel`. Introducing those requires the denied
Composite route/resolver or additional out-of-scope upstream changes, so the
contract stop rule applies. No network or live provider call was made.

Unverified risks: the routed catalog behavior, cache isolation, capability
intersection, account model discovery, and Use Key UI remain unimplemented on
this baseline.

Knowledge candidates: none.
