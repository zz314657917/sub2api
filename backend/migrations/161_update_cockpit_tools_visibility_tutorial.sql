UPDATE tutorial_pages
SET content_md = replace(
        content_md,
        E'登录后可以一键添加到 Cockpit Tools，浏览器会唤起 Cockpit Tools。若不能唤起，检查协议注册和浏览器权限。\n\n[[callout type="tip" title="恢复 Codex 会话"]]\n导入账号后，如果 Codex 会话不可见，优先检查账号可见性和当前 Provider 是否选中。\n[[/callout]]',
        E'登录后可以一键添加到 Cockpit Tools，浏览器会唤起 Cockpit Tools。若不能唤起，检查协议注册和浏览器权限。\n\n导入账号后，在 Cockpit Tools 左侧点击 Codex 图标，确认当前账号已经切到落叶网络 Provider。\n\n## 恢复历史对话\n\n如果 Codex App 已经登录但历史对话没有显示，可以用 Cockpit Tools 修复可见性。\n\n1. 打开 Cockpit Tools。\n2. 点击左侧 Codex 图标。\n3. 点击顶部的会话管理。\n4. 点击修复可见性。\n5. 回到 Codex App，刷新或重新打开会话列表。\n\n[[callout type="tip" title="恢复 Codex 会话"]]\n导入账号后，如果 Codex 会话不可见，先确认当前 Provider 已选中落叶网络，再进入会话管理点击修复可见性。\n[[/callout]]'
    ),
    updated_at = NOW()
WHERE slug = 'cockpit-tools'
  AND content_md LIKE '%导入账号后，如果 Codex 会话不可见，优先检查账号可见性和当前 Provider 是否选中%'
  AND content_md NOT LIKE '%## 恢复历史对话%';
