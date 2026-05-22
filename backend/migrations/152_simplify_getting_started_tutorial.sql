UPDATE tutorial_pages
SET description = '安装 Codex App，创建 codex 密钥，写入配置并启动验证。',
    content_md = $md$
# 快速开始

按下面 4 步完成第一次接入；完整配置见 Codex 教程。

[[callout type="tip" title="新手最快路线"]]
1. 安装或打开 Codex App。
2. 登录落叶网络控制台，创建 codex 类型密钥并复制 Token。
3. 按 Codex 教程写入 config.toml 和 auth.json。
4. 启动 Codex App，确认模型能正常响应。
[[/callout]]

## 安装 Codex App

打开 Codex App 安装页，按系统提示安装桌面端。

[[link-button href="https://chatgpt.com/codex" label="打开 Codex App"]]

## 创建 API 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 `codex`。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="选择分组和密钥类型后创建"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]

## Base URL

OpenAI 兼容工具通常使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.top/v1
[[/command]]

Claude Code 兼容工具通常使用：

[[command title="Anthropic 兼容 Base URL"]]
https://ai.3zapi.top
[[/command]]

## 下一步：Codex 配置

打开 Codex 教程，复制配置文件示例并写入你的 Token。

[[link-button href="/tutorial/codex" label="查看 Codex 配置"]]
$md$,
    updated_at = NOW()
WHERE slug = 'getting-started'
  AND (
    content_md LIKE '%第一次接入先按新手最快路线走%'
    OR content_md LIKE '%第一次接入先走 Codex App 路线%'
  );
