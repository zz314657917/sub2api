# Image 2 selective integration continuation

Verified baseline 2026-09-16: main bb3dde9e4.

- I-01 cached image billing integrated as 23ddcc8d6.
- I-02 dated Image 2.5 aliases integrated as 568fd9ced.
- I-03 review fixes and I-04 native OAuth route remain incomplete. Contract: ../tasks/image2-native-port.md. Implement shared helpers, adapter, dispatch/fallback, usage and local mock regression in sequence. Real provider acceptance and account-test parity remain open requirements.
- The original image2-i04 worktree contains only an untracked experimental adapter with unresolved helpers; it is preserved, not accepted.
- MiniMax patch 5968fd0ed does not apply: local validation shares providerAdapters and no MiniMax adapter exists. Adding MiniMax support is a separate feature, not a missing allowlist fix.
- Preserve unrelated main worktree changes. No wholesale merge/rebase/cherry-pick, push, database, deployment or container update.
