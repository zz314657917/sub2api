# 任务状态说明

最后更新：2026-05-13

`knowledge/tasks/` 存放动态协作状态，默认不作为长期知识正文。

- `current-task.md`：当前任务快照，用于掉线恢复、继续做、交接。
- `timeline.md`：阶段时间轴，用于记录近期重点、关键决策、验证记录和遗留问题。

维护规则：

- 当前状态变化时更新 `current-task.md`。
- 阶段完成、暂停或需要归档时，向 `timeline.md` 顶部追加一条倒序记录。
- 长期稳定的项目结论写到 `knowledge/` 根下专题页，不写进任务快照。
- 不写密钥、token、账号、私有地址或未验证猜测。

Git 状态：

- `.gitignore` 当前忽略 `knowledge/tasks/`。
- 因此 `current-task.md` 和 `timeline.md` 默认是本地协作状态，不一定随代码提交。
- `knowledge/` 根下专题页未被忽略，适合作为可提交的长期知识库。
