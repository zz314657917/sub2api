UPDATE tutorial_pages
SET content_md = $md$
# Cockpit Tools 账号管理

Cockpit Tools 适合用网页面板管理 Codex 账号和会话。

## 下载

项目地址：https://github.com/jlcodes99/cockpit-tools

当前教程使用版本：`v0.23.2`。

## 导入落叶网络

Base URL 使用：

[[command title="Cockpit Base URL"]]
https://ai.3zapi.top
[[/command]]

登录后可以一键添加到 Cockpit Tools，浏览器会唤起 Cockpit Tools。若不能唤起，检查协议注册和浏览器权限。

导入账号后，在 Cockpit Tools 左侧点击 Codex 图标，确认当前账号已经切到落叶网络 Provider。

[[callout type="tip" title="恢复 Codex 会话"]]
导入账号后，如果 Codex 会话不可见，先确认当前 Provider 已选中落叶网络；具体操作见恢复历史会话教程。
[[/callout]]

[[link-button href="/tutorial/recover-codex-session" label="查看恢复历史会话"]]
$md$,
    updated_at = NOW()
WHERE slug = 'cockpit-tools'
  AND (
    content_md LIKE '%## 恢复历史对话%'
    OR content_md LIKE '%导入账号后，如果 Codex 会话不可见，优先检查账号可见性和当前 Provider 是否选中%'
  );

INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES ('recover-codex-session', '恢复历史会话', '使用 Cockpit Tools 修复 Codex 历史会话可见性。', '排查', 55, 'published', $md$
# 恢复历史会话

如果 Codex App 已经登录但历史对话没有显示，可以用 Cockpit Tools 修复可见性。

## 操作步骤

1. 打开 Cockpit Tools。
2. 点击左侧 Codex 图标。

[[screenshot src="/tutorial/cockpit-tools/codex-sidebar.png" alt="Cockpit Tools 左侧 Codex 图标" caption="点击左侧 Codex 图标进入 Codex 页面"]]

3. 点击顶部的会话管理。
4. 点击修复可见性。

[[screenshot src="/tutorial/cockpit-tools/session-visibility.png" alt="Cockpit Tools 会话管理修复可见性" caption="进入会话管理后点击修复可见性"]]

5. 回到 Codex App，刷新或重新打开会话列表。

[[callout type="tip" title="看不到会话时请等待一下"]]
历史会话加载可能有延迟，修复后请等待一会儿，再刷新或重新打开会话列表。
[[/callout]]
$md$, NOW())
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    content_md = EXCLUDED.content_md,
    updated_at = NOW(),
    published_at = COALESCE(tutorial_pages.published_at, NOW());
