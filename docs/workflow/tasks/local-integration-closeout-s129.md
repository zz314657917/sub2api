---
task_id: local-integration-closeout-s129
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Record the completed local integration of the selective `v0.1.168` S128 port
without altering its business behavior or publication state. Workflow and
handoff records must identify the implementation commit `85439ff50` and merge
commit `fbf4ea10e`, while preserving that `origin/main` has not been updated.

## Success Criteria

- `docs/workflow/status.md` reports S129 as the current completed sprint and
  accurately distinguishes local merge completion from remote publication.
- `docs/workflow/spec.md`, `docs/workflow/main-log.md`, and
  `knowledge/tasks/current-task.md` contain the same local-only facts and
  actionable next gates.
- The handoff keeps its eight standard sections, does not disclose sensitive
  configuration, and states that `outputs/` remains untracked.
- No source, dependency, deployment, container, generated, or remote Git state
  changes occur.

## Allowed Paths

- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/local-integration-closeout-s129.md`
- `docs/workflow/qa-reports/local-integration-closeout-s129-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/**`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `outputs/**`
- `.gitignore`
- Any path not listed under Allowed Paths

## Constraints

- Do not amend, reset, rebase, merge, push, fetch, or otherwise alter existing
  Git history or remote state. After final QA, one new documentation-only
  local commit is allowed to persist this handoff; it must not be pushed.
- Do not edit business code, tests, dependency locks, configuration, migrations,
  or deployment/container assets.
- Preserve the documented source-level limit: no real upstream OAuth,
  Anthropic, Gemini, Antigravity, database, or authenticated browser smoke was
  run for S128.

## Acceptance Commands

Run from the repository root:

```powershell
git diff --check
git diff --name-only
git ls-files -u
git status --short
git rev-parse --short HEAD
git rev-parse --short origin/main
git rev-list --left-right --count origin/main...HEAD
```

## Output

- Accurate S129 workflow and handoff records plus a source-only QA report.

## Stop Rules

- Stop if documenting the facts requires a source-code, deployment, container,
  remote Git, or `outputs/` change.
- Stop if the local and remote commit relationship cannot be verified from Git.

## Contract Review

`PASS`: the task is bounded to durable workflow and handoff facts. The local
commit relationship is directly verifiable (`origin/main` at `49af8e1bb`,
implementation at `85439ff50`, and local merge at `fbf4ea10e`), and the
approved paths exclude all runtime, release, and source-code surfaces.

## Implementation Result

- Workflow status and specification now describe the local S128 commit and
  merge, distinguish them from publication, and retain external runtime smoke
  as unverified.
- The main log records the contract and documentation build gates.
- The current-task handoff now points to publication preflight as the next
  action and preserves the untracked `outputs/` boundary.

## QA Result

`PASS / documentation-only`: the working diff is limited to the approved
workflow and handoff paths; `git diff --check` and unmerged-index checks pass,
and the recorded local/remote relationship matches Git. No source or remote
operation was performed.

## Contract Amendment

`PASS`: allow one new local documentation-only commit after QA. Without it,
the approved handoff would remain as a dirty working tree and block the next
publication preflight. The amendment neither changes existing history nor
authorizes a push, fetch, deployment, container change, or source edit.
