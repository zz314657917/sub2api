# 响应模型审计：已授权规格

- Task ID: upstream-response-model-audit
- Baseline: ebd3db125c52c16496465c7697b4b110f94e4774
- Worktree: E:/codex-worktrees/sub2api/upstream-response-model-audit
- Branch: codex/upstream-response-model-audit
- Upstream references: db0bff82c, 6e34fb09c, c46d07ca0; selectively adapt behavior, not history or upstream file topology.

## 范围

采集上游响应声明的模型，对比实际发给上游的模型，并在管理员用量列表展示、筛选和导出。该信息是声明审计，不证明实际运行模型身份。

覆盖本地 OpenAI / Claude / Gemini / Antigravity 转发、HTTP 非流式、SSE 和 WebSocket；重试和 WS 轮次隔离。响应无模型时保留未知状态，禁止用请求模型伪造观测。保留原始响应模型，识别已知 Grok 别名，日期版本作为展示层模型变体。观察器不能改变原始透传、用量或计费行为。

新增 nullable upstream_response_model 和 upstream_model_mismatch。新增迁移暂用 246/247，执行前重查全主线迁移占用；不得修改旧迁移。管理员列表、筛选、统计、分页、导出和查询缓存使用一致条件。保留本地现有 XLSX 导出格式并新增审计列，无需另加 CSV。普通用户 DTO 不暴露上述管理员审计字段。

## 排除

不实现按响应模型计费、service_tier 行为、路由/封禁变更、无关 provider 或模型目录功能；不修改鹈鹕任务，不访问现有业务数据库，不调用付费 provider，不推送、不部署。

## 验收要求

- 有效生产路径测试：不同协议模型字段、无字段/非法字段、首帧与终帧优先级、冲突、重试与 WS 轮次隔离、映射后模型比较、Grok 别名、原始转发和计费不变。
- 专属 PostgreSQL：新增迁移与重复执行、旧行 NULL、实际写入/读回、true/false/NULL 筛选、分页/统计一致、并发索引；不得使用现有业务数据库。
- API/DTO：管理员可见、普通用户不可见；缓存键区分筛选条件。
- 前端组件测试、类型检查、构建；独立 profile 的管理员用量页桌面/移动端模拟验收（包含截图所示异名、日期变体、缺失模型及映射链），结束精确清理会话/进程。
- 基线错误先复现；不得将 no-tests-to-run 或跳过测试记为通过。受影响生产路径缺少有效证据则 BLOCKED。
- 独立 Terra contract review 通过后开发；原 Terra Developer 与另一独立 Terra QA；主控使用 review-and-verification 作最终裁决。
- 合回前核对 main HEAD/index/脏改/迁移编号，只有无覆盖的快进合并可执行；不 stash，不重写历史。生产代码与验收证据分别提交。

## 当前门禁

仅规格与隔离工作区已建立。精确文件白名单和可执行命令待本地拓扑核对后写入独立 contract；此规格不等于已通过 contract，不授权跳过评审。
