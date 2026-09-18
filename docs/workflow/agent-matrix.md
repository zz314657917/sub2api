# Agent Matrix

| Role | Model | Responsibility |
| --- | --- | --- |
| Planner / Final Evaluator | Current controller | Scope, contract, integration and final evidence review |
| Contract Reviewer | gpt-5.6-terra | Independent contract review before implementation |
| Developer | gpt-5.6-terra | Implement only approved task allowlist |
| QA | gpt-5.6-terra | Independent acceptance, separate from Developer |

Developer and QA must be independently dispatched. If Terra is unavailable, report BLOCKED; no silent model substitution. Protect unrelated dirty work and use an isolated worktree. Contract review and final acceptance are separate gates. No push/deployment without user authorization.
