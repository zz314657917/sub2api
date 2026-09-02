---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0200-group-pricing-layout-s290
worker_model: gpt-5.6-terra
base_commit: 6050139a3
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-02
---

# Upstream v0.2.0 Group Pricing Layout S290

## Task ID
upstream-v0200-group-pricing-layout-s290

## Role
你是 P/G/E 流程里的 Generator worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal
手工适配上游 `1a33dc8cc` 的分组模型定价布局修复，消除窄屏下分组模型定价六项默认价格的横向挤压，同时保持所有表单字段、事件与计价语义不变。

## Success Criteria
- 新建和编辑分组的模型定价弹窗使用既有 `wide` 宽度，并在窄屏允许说明与“添加定价”按钮自然换行。
- 两个模型定价标题行均允许换行、说明容器可收缩、添加控件不收缩；分组弹窗中实际渲染的六项默认 Token 价格采用本地字段数对应的响应式网格。
- `IntervalRow` 保持其现有字段、事件与响应式网格标记，但分组创建/编辑调用方固定传入 `hide-token-intervals=true`，不得为本 Sprint 改变该既有语义；Token 区间行的真实浏览器验收不属于分组弹窗验收。
- 不引入横向滚动，不删除字段或改变输入、emit、校验逻辑；源级测试必须断言两个 `wide` 弹窗、两个可换行标题行、添加控件不收缩，以及默认价格与共享区间两个响应式网格标记。
- 定向 Vitest、前端 typecheck 和生产构建通过；浏览器验收使用任务专属 CLI session，采集实际 profile/PID，并在桌面与移动视口对创建、编辑分组中已展开的六项默认 Token 价格检查无横向溢出。表单必须取消，不保存；仍须执行 session/profile 清理门禁。

## Context
- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/spec.md`, `docs/workflow/status.md`
- Source: upstream `1a33dc8cc`; 直接 patch 因本地组件拓扑分叉而失败，必须保持本地六字段布局做行为级适配。
- QA 发现旧合同把 Token 区间行列为分组弹窗浏览器验收目标，但本地两个调用均显式传入 `hide-token-intervals=true`。这是不可达的验收条件，不是 S290 的 UI 缺陷。

## Allowed Paths
- `frontend/src/components/admin/channel/IntervalRow.vue`
- `frontend/src/components/admin/channel/PricingEntryCard.vue`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts`
- `docs/workflow/worker-results/upstream-v0200-group-pricing-layout-s290-result.md`

## Denied Paths
- `backend/**`
- `frontend/pnpm-lock.yaml`
- `frontend/src/views/admin/pixelCafe/**`
- `backend/internal/pkg/apicompat/**`
- `backend/internal/service/admin_service.go`
- `knowledge/**`
- `outputs/**`
- 数据库 migration、Ent 生成物、路由、鉴权、计费、生产配置、容器、部署与未列入 Allowed Paths 的任何文件。

## Constraints
- 保持最小改动，不做无关重构或格式化。
- 使用现有 Tailwind 与组件样式，不新增依赖，不改 API 或 i18n 文案。
- 不回滚、覆盖或暂存既有脏改；受保护业务脏改 hash 必须保持 `0e467987fd7aec5fc451983bdb8f8216f97ba69c`。
- 不直接 cherry-pick 或整体 merge 上游。
- 保留 `GroupsView` 两处 `:hide-token-intervals="true"`；不得为了浏览器证据而启用、保存或改变 Token 区间计费行为。
- 浏览器只允许访问本机 Vite；不得复用用户浏览器 profile、不得使用生产 URL、真实管理员凭据或共享数据。

## Acceptance Commands
```powershell
Set-Location F:/mcplugins/sub2api/frontend
pnpm.cmd exec vitest run src/views/admin/__tests__/groupsModelsListLayout.spec.ts
pnpm.cmd run typecheck
pnpm.cmd run build

Set-Location F:/mcplugins/sub2api
git diff --check -- frontend/src/components/admin/channel/IntervalRow.vue frontend/src/components/admin/channel/PricingEntryCard.vue frontend/src/views/admin/GroupsView.vue frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts
git diff --name-only --diff-filter=U

$protected = @(
  'backend/internal/pkg/apicompat',
  'backend/internal/service/admin_service.go',
  'frontend/pnpm-lock.yaml',
  'frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue'
)
$actualProtectedHash = git diff -- $protected | git hash-object --stdin
if ($actualProtectedHash -ne '0e467987fd7aec5fc451983bdb8f8216f97ba69c') {
  throw "Protected dirty diff changed: $actualProtectedHash"
}

# Browser protocol. Start Vite in a task-owned process and record $vite.Id.
Set-Location F:/mcplugins/sub2api/frontend
$vite = Start-Process -FilePath pnpm.cmd -ArgumentList 'exec','vite','--host','127.0.0.1','--port','5174','--strictPort' -WorkingDirectory $PWD -PassThru -WindowStyle Hidden
$session = 'sub2api-s290-pricing-layout'
$profile = $null
$browserStarted = Get-Date
try {
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session open http://127.0.0.1:5174 --headed
  # The browser must have started with this task, and must expose a non-default user-data directory.
  $browser = Get-CimInstance Win32_Process | Where-Object {
    $_.Name -match 'chrome|chromium' -and
    $_.CommandLine -match '--user-data-dir=' -and
    [Management.ManagementDateTimeConverter]::ToDateTime($_.CreationDate) -ge $browserStarted.AddSeconds(-2)
  } | Select-Object -Last 1
  if ($null -eq $browser) { throw 'No task browser profile found' }
  $profile = [regex]::Match($browser.CommandLine, '--user-data-dir=(?:"(?<p>[^"]+)"|(?<p>\S+))').Groups['p'].Value
  if ([string]::IsNullOrWhiteSpace($profile) -or $profile -match 'Google\\Chrome\\User Data') {
    throw "Refusing non-task browser profile: $profile"
  }
  [pscustomobject]@{ session = $session; vite_pid = $vite.Id; browser_pid = $browser.ProcessId; profile = $profile } | Format-List

  # Snapshot before every element reference. Do not authenticate unless safely provisioned local credentials exist.
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session snapshot
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session resize 1440 900
  # Open the local group pricing modal only with safely provisioned local credentials; otherwise report this as unreachable.
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session eval 'document.documentElement.scrollWidth === document.documentElement.clientWidth'
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session screenshot
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session resize 390 844
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session snapshot
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session eval 'document.documentElement.scrollWidth === document.documentElement.clientWidth'
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session screenshot
} finally {
  npx.cmd --yes --package @playwright/cli playwright-cli -s=$session close 2>$null
  Get-Process -Id $vite.Id -ErrorAction SilentlyContinue | Stop-Process
  $ownedResidual = if ($profile) {
    Get-CimInstance Win32_Process | Where-Object {
      $_.CommandLine -match [regex]::Escape($profile) -or
      ($_.CommandLine -match 'playwright-cli|cliDaemon' -and $_.CommandLine -match [regex]::Escape($session))
    }
  }
  $ownedResidual | ForEach-Object { Stop-Process -Id $_.ProcessId }
  $remainingOwned = if ($profile) {
    Get-CimInstance Win32_Process | Where-Object {
      $_.CommandLine -match [regex]::Escape($profile) -or
      ($_.CommandLine -match 'playwright-cli|cliDaemon' -and $_.CommandLine -match [regex]::Escape($session))
    }
  }
  $remainingOwned | Select-Object ProcessId, Name, CommandLine
  if ($remainingOwned) { throw 'Task-owned browser or cliDaemon process remains' }
}
```

## Output
- 按 `C:/Users/Administrator/.codex/templates/worker-result.md` 写 worker report。
- Worker report 第一行必须是 `### DONE: upstream-v0200-group-pricing-layout-s290`、`### BLOCKED: upstream-v0200-group-pricing-layout-s290` 或 `### FAILED: upstream-v0200-group-pricing-layout-s290`。
- 列出 changed files、commands run、浏览器 session、实际 profile 路径、Vite/Browser/cliDaemon PID、截图路径或其不可执行原因、关闭/清理证据、risks 与 knowledge_candidates。

## Stop Rules
- Contract 不清、验收命令不可执行或需要改 Denied Paths 时停止并报告。
- 无法证明 browser profile 非用户默认路径、无法证明残留进程归属、需要管理员凭据、生产服务或真实共享数据时，不得尝试绕过认证；关闭 CLI session，停止已记录的 Vite PID，将浏览器验证标记为未验证。
- 连续失败或出现组件 owner 争议时，停止 worker loop，交由 Codex 裁决。

## Budget
- worker_mode: `claude-bare-gpt-5.6-terra`
- qa_worker_mode: `codex-agent-gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Worker Output
- 兼容旧脚本字段；内容同 `Output`。
