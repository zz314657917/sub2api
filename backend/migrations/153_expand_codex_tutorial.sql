UPDATE tutorial_pages
SET description = 'Codex App / CLI 详细配置教程。',
    content_md = $md$
# Codex 配置

Codex 推荐使用落叶网络的 OpenAI 兼容接口。流程是：安装 Codex App，创建 codex 类型密钥，写入 config.toml 和 auth.json，然后启动验证。

[[callout type="tip" title="推荐路线"]]
新手先走 Codex App。Codex App 不要求先安装 Git 或 Node.js；Node.js/npm 只用于 Codex CLI，Git 主要用于代码项目版本管理和团队协作。
[[/callout]]

## 适用场景

- Codex App：推荐新手使用，先安装桌面端再写配置。
- Linux / macOS：优先使用 Codex CLI，需要 Node.js/npm；也可以下载 GitHub Release 二进制包。
- VS Code 或其他 IDE 插件：优先复用同一套 Base URL 和 API Key。

## 安装 Codex App

[[command title="Windows winget 安装"]]
winget install --id OpenAI.Codex -e
[[/command]]

也可以从 Microsoft Store 或 https://chatgpt.com/codex 安装桌面端。

## Linux / macOS 安装 Codex CLI

Linux 和 macOS 通常使用 Codex CLI。先确认已经安装 Node.js/npm，然后全局安装 Codex。

[[command title="Linux / macOS npm 安装"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

不想走 npm 时，可以到官方 GitHub Release 下载对应架构的二进制包；Linux 常见架构是 x86_64 或 arm64。下载后把 codex 可执行文件放到 PATH 目录里，再执行 codex --version 验证。

[[link-button href="https://github.com/openai/codex/releases" label="打开 Codex Releases"]]

## 创建 codex 密钥

登录落叶网络控制台，进入 API 密钥页面，点击创建密钥。名称可以自由填写，分组选择可用套餐，密钥类型请选择 codex。创建后复制 Token，后面写入 auth.json。

[[screenshot src="/tutorial/api-key/sidebar-api-key.png" alt="API 密钥入口" caption="左侧进入 API 密钥页面"]]
[[screenshot src="/tutorial/api-key/create-key-button.png" alt="创建密钥按钮" caption="点击右上角创建密钥"]]
[[screenshot src="/tutorial/api-key/create-key-dialog.png" alt="创建密钥弹窗" caption="密钥类型选择 codex"]]

[[link-button href="/keys" label="打开 API 密钥页面"]]

## 第一步：找到 Codex 配置文件夹

Codex 会读取一个叫 .codex 的文件夹。这个文件夹就放在你当前电脑账号的用户目录里，不是项目目录，也不是浏览器下载目录。

- Windows 示例：C:/Users/你的用户名/.codex/
- macOS / Linux 示例：/home/你的用户名/.codex/ 或 /Users/你的用户名/.codex/

不知道用户名也没关系，直接复制下面命令执行即可，它会自动进入你的用户目录并创建 .codex 文件夹。

[[command title="Windows PowerShell 创建配置文件夹"]]
mkdir $env:USERPROFILE\.codex -Force
explorer $env:USERPROFILE\.codex
[[/command]]

[[command title="Linux / macOS 创建配置文件夹"]]
mkdir -p ~/.codex
cd ~/.codex
pwd
[[/command]]

## 第二步：创建 config.toml

config.toml 是 Codex 的主配置文件，用来告诉 Codex 使用哪个模型、哪个服务商、哪个 Base URL。

Windows 用户：在刚打开的 .codex 文件夹里新建文本文件，命名为 config.toml。注意不要变成 config.toml.txt。

Linux / macOS 用户：在终端执行下面命令打开编辑器；如果文件不存在会自动创建。

[[command title="Linux / macOS 打开 config.toml"]]
nano ~/.codex/config.toml
[[/command]]

把下面内容完整粘进去，然后保存。

[[command title="config.toml" lang="toml"]]
model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"
[[/command]]

- model_provider = "luoye" 表示 Codex 使用下面这个自定义服务商。
- base_url 必须写成 `https://ai.3zapi.top/v1`，结尾带 `/v1`。
- env_key = "OPENAI_API_KEY" 要和 auth.json 里的字段名保持一致；CLI 环境变量方案也使用这个名字。
- wire_api = "responses" 表示使用 Responses 协议。

## 第三步：创建 auth.json

auth.json 用来保存你的 API Key。这里的 key 就是你在落叶网络控制台创建的 codex 类型密钥。

Windows 用户：在同一个 .codex 文件夹里新建 auth.json，再用记事本打开。

Linux / macOS 用户：

[[command title="Linux / macOS 打开 auth.json"]]
nano ~/.codex/auth.json
[[/command]]

把下面内容粘进去，并把中文提示替换成你的真实 API Key。

[[command title="auth.json" lang="json"]]
{
  "OPENAI_API_KEY": "替换成你的 codex 类型 API Key"
}
[[/command]]

如果 auth.json 已经存在，只需要更新 OPENAI_API_KEY 这一项，不要删除其他仍在使用的配置。

## 启动验证

关闭并重新打开 Codex App，进入一个项目后发送简单问题，确认模型能正常响应。

[[callout type="warning" title="失败时先查这三项"]]
1. Base URL 是否为 `https://ai.3zapi.top/v1`。
2. API Key 是否复制完整，并且密钥类型是 codex。
3. 当前分组是否有可用套餐和余额。
[[/callout]]

## CLI 用户可选安装

只有使用终端版 Codex CLI 时才需要 Node.js/npm。Linux/macOS 用户通常走这一条；新机器建议直接安装当前 LTS 版本。

[[command title="npm 镜像安装 Codex CLI"]]
npm install -g @openai/codex --registry=https://registry.npmmirror.com
codex --version
[[/command]]

安装后在终端执行 codex，CLI 会读取同一目录下的 config.toml 和 auth.json。

## 常见问题

- 401 或 unauthorized：优先检查 Token 是否复制完整、密钥类型是否为 codex。
- model not found：当前分组可能不支持配置里的模型，换成控制台可用模型后再试。
- Base URL 报错：OpenAI 兼容地址必须写 `https://ai.3zapi.top/v1`，不要漏掉 `/v1`。
- 配置不生效：修改配置后重启 Codex App；如果是 CLI，关闭旧终端后重新打开。
$md$,
    updated_at = NOW()
WHERE slug = 'codex'
  AND (
    content_md LIKE '%新手优先使用 Codex App。Codex App 的安装和登录不要求先安装 Git 或 Node.js%'
    OR content_md LIKE '%新手优先使用 Codex App。先创建 `codex` 类型 API 密钥%'
    OR content_md LIKE '%Codex 推荐使用落叶网络的 OpenAI 兼容接口。流程是：安装 Codex App%'
  );
