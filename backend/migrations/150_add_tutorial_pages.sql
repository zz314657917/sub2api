CREATE TABLE IF NOT EXISTS tutorial_pages (
    id           BIGSERIAL PRIMARY KEY,
    slug         VARCHAR(80)  NOT NULL,
    title        VARCHAR(160) NOT NULL,
    description  VARCHAR(500) NOT NULL DEFAULT '',
    category     VARCHAR(80)  NOT NULL DEFAULT '',
    sort_order   INTEGER      NOT NULL DEFAULT 0,
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',
    content_md   TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ  NULL,
    CONSTRAINT tutorial_pages_slug_check
        CHECK (slug ~ '^[a-z0-9][a-z0-9-]{0,78}[a-z0-9]$' OR slug ~ '^[a-z0-9]$'),
    CONSTRAINT tutorial_pages_status_check
        CHECK (status IN ('draft', 'published'))
);

CREATE UNIQUE INDEX IF NOT EXISTS tutorial_pages_slug_key
    ON tutorial_pages (slug);

CREATE INDEX IF NOT EXISTS idx_tutorial_pages_public_order
    ON tutorial_pages (sort_order, category, id)
    WHERE status = 'published';

CREATE INDEX IF NOT EXISTS idx_tutorial_pages_status
    ON tutorial_pages (status);

INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES
('getting-started', '快速开始', '准备账号、创建 API Key、确认 Base URL，然后选择要接入的工具。', '使用指南', 10, 'published', $md$
# 快速开始

第一次接入先按新手最快路线走，只需要准备账号、创建 Codex 密钥、写入配置并启动。CC Switch、Cockpit Tools、MinePilotQA 和其他工具都可以后续再看。

[[callout type="tip" title="新手最快路线"]]
1. 安装 Git 和 Node.js。
2. 登录落叶网络控制台，打开 API 密钥页面。
3. 创建 Codex 类型密钥并复制 Token。
4. 打开 Codex 教程写入 config.toml 和 auth.json。
[[/callout]]

## 准备基础工具

- Windows、Linux 或 macOS 电脑。
- Node.js v18 或更高版本。
- 落叶网络 API Token 和 Base URL。

[[command title="验证 Git / Node.js"]]
git --version
node -v
npm -v
[[/command]]

## 创建 API 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型按目标工具选择；Codex 用户请选择 `codex`。

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
$md$, NOW()),
('codex', 'Codex', 'Codex App / CLI 配置教程。', '工具配置', 20, 'published', $md$
# Codex 配置

新手优先使用 Codex App。先创建 `codex` 类型 API 密钥，然后写入 `config.toml` 和 `auth.json`。

## 安装 Codex App

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]

也可以从 Microsoft Store 或 https://chatgpt.com/codex 安装桌面端。

## 写入 config.toml

配置文件位置示例：`C:\Users\用户名\.codex\config.toml`。

[[command title="config.toml" lang="toml"]]
model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"
[[/command]]

## 写入 auth.json

配置文件位置示例：`C:\Users\用户名\.codex\auth.json`。

[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "输入你的key"
}
[[/command]]

## CLI 用户可选安装

[[command title="npm 镜像安装 Codex CLI"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
[[/command]]

## 启动

[[command title="启动 Codex"]]
codex
[[/command]]
$md$, NOW()),
('claude-code', 'Claude Code', 'Claude Code 命令行接入教程。', '工具配置', 30, 'published', $md$
# Claude Code 配置

Claude Code 使用 Anthropic 兼容环境变量。确认 API Token 后写入 `ANTHROPIC_AUTH_TOKEN` 和 `ANTHROPIC_BASE_URL`。

## 安装

[[command title="安装 Claude Code"]]
npm install -g @anthropic-ai/claude-code --registry=https://registry.npmmirror.com
claude --version
[[/command]]

## Windows PowerShell 环境变量

[[command title="PowerShell"]]
$env:ANTHROPIC_AUTH_TOKEN="你的令牌"
$env:ANTHROPIC_BASE_URL="https://ai.3zapi.top"

[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN","你的令牌","User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL","https://ai.3zapi.top","User")
[[/command]]

## 兜底配置文件

[[command title="settings.json" lang="json"]]
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "你的令牌",
    "ANTHROPIC_BASE_URL": "https://ai.3zapi.top",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
[[/command]]

## 启动

[[command title="启动 Claude Code"]]
claude
[[/command]]
$md$, NOW()),
('cc-switch', 'CC Switch', '桌面端统一管理 Codex 和 Claude Code 账号。', '工具配置', 40, 'published', $md$
# CC Switch 账号管理

CC Switch 适合需要多工具、多账号切换的用户。推荐从落叶网络控制台一键导入 Provider，也可以手动添加。

## 功能点

- Provider 管理。
- MCP 服务器管理。
- Prompts 管理。
- 多平台支持。
- Deep Link 导入：`ccswitch://`。

## macOS 安装

[[command title="Homebrew"]]
brew tap farion1231/ccswitch
brew install --cask cc-switch
[[/command]]

## Arch Linux 安装

[[command title="Arch Linux"]]
paru -S cc-switch-bin
[[/command]]

## Web 版本

[[command title="Linux Web"]]
wget https://github.com/farion1231/cc-switch/releases/latest/download/cc-switch-web-linux-x64.tar.gz
tar -xzf cc-switch-web-linux-x64.tar.gz
cd cc-switch-web/
./cc-switch-web
# open http://localhost:17666
[[/command]]
$md$, NOW()),
('cockpit-tools', 'Cockpit Tools', '网页面板管理 Codex 账号和会话。', '工具配置', 50, 'published', $md$
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

[[callout type="tip" title="恢复 Codex 会话"]]
导入账号后，如果 Codex 会话不可见，优先检查账号可见性和当前 Provider 是否选中。
[[/callout]]
$md$, NOW()),
('minepilotqa', 'MinePilotQA', '在 MinePilotQA 中添加 3zapi 服务商并测试连接。', '工具配置', 60, 'published', $md$
# MinePilotQA 接入

在 AI 配置里添加 3zapi 服务商，测试连接返回 HTTP 200 即成功。

## 配置服务商

Base URL：

[[command title="MinePilotQA Base URL"]]
https://ai.3zapi.top/
[[/command]]

检查项：

- Core 正常。
- HTTP 200。
- Provider models endpoint reachable。
- models endpoint reachable。

[[screenshot src="/tutorial/minepilotqa/provider-list.png" alt="服务商列表" caption="进入 Provider 管理"]]
[[screenshot src="/tutorial/minepilotqa/add-provider.png" alt="添加服务商" caption="添加 3zapi 服务商"]]
[[screenshot src="/tutorial/minepilotqa/connection-success.png" alt="连接成功" caption="连接测试成功"]]
$md$, NOW()),
('openclaw', 'OpenClaw', '自托管助手平台，使用 OpenAI 兼容 provider 接入。', '工具配置', 70, 'published', $md$
# OpenClaw 接入

OpenClaw 使用 OpenAI 兼容 provider 接入落叶网络。

参考：https://docs.easyrouter.io/zh/docs/apps/openclaw

## 初始化

[[command title="OpenClaw"]]
openclaw onboard
openclaw dashboard
[[/command]]

## openclaw.json

[[command title="openclaw.json" lang="json"]]
{
  "models": {
    "providers": {
      "luoye": {
        "api": "openai-completions",
        "name": "luoye",
        "baseURL": "https://ai.3zapi.top/v1",
        "envKey": "LUOYE_API_KEY",
        "models": [
          {
            "id": "gpt-5.3-codex",
            "contextWindow": 400000
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": {
        "primary": "luoye/gpt-5.3-codex"
      }
    }
  }
}
[[/command]]
$md$, NOW()),
('hermes-agent', 'Hermes-Agent', '终端 Agent，通过 Custom Endpoint 接入。', '工具配置', 80, 'published', $md$
# Hermes-Agent 接入

Hermes-Agent 通过 Custom Endpoint 接入落叶网络。

参考：https://docs.easyrouter.io/zh/docs/apps/hermes-agent

## 安装

[[command title="安装 Hermes-Agent"]]
curl -fsSL https://raw.githubusercontent.com/terryso/hermes-agent/main/install.sh | bash
[[/command]]

## 配置模型

[[command title="hermes model"]]
hermes model
# Provider: Custom Endpoint
# Base URL: https://ai.3zapi.top/v1
# API Key: 你的API Token
# Model: gpt-5.3-codex
[[/command]]

## 启动

[[command title="启动 Hermes"]]
hermes
[[/command]]
$md$, NOW()),
('faq', '常见问题', '安装、配置、鉴权和网络问题快速排查。', '使用指南', 90, 'published', $md$
# 常见问题

## 安装失败

换镜像源，再确认 Node.js 版本是 v18 或更高。

## 配置不生效

保存后关闭终端重开，仍不行就重启电脑。

## 命令找不到

Git 和 Node.js 建议默认安装，不要手动改路径。

## 鉴权失败

检查 Token 和 Base URL，注意不要多复制空格。

## 网络异常

npm 安装用镜像源，运行时确认代理地址可访问。

## 查看额度

到落叶网络控制台查看用量、记录和剩余额度。
$md$, NOW())
ON CONFLICT (slug) DO NOTHING;
