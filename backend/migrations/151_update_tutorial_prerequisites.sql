UPDATE tutorial_pages
SET content_md = $md$
# 快速开始

第一次接入先按新手最快路线走，只需要准备账号、创建 Codex 密钥、写入配置并启动。Codex App 不需要先安装 Git 或 Node.js；这两个放到 CLI、Claude Code 或代码项目阶段再准备。CC Switch、Cockpit Tools、MinePilotQA 和其他工具都可以后续再看。

[[callout type="tip" title="新手最快路线"]]
1. 登录落叶网络控制台，打开 API 密钥页面。
2. 创建 Codex 类型密钥并复制 Token。
3. 打开 Codex 教程写入 config.toml 和 auth.json。
4. 需要 CLI、Claude Code 或代码项目协作时，再安装 Node.js/npm 和 Git。
[[/callout]]

## 准备工具

- Windows、Linux 或 macOS 电脑。
- 落叶网络 API Token 和 Base URL。
- Codex App：不把 Git、Node.js 作为启动前置条件。
- Codex CLI：需要 Node.js/npm，建议使用当前 LTS 版本；当前 npm 包声明支持 Node.js 16+。
- Claude Code：npm 安装路径需要 Node.js 18+；Windows 原生使用还会依赖 Git Bash 或 WSL。
- Git：不是 Codex App 的硬性要求，但做代码项目、版本回滚和团队协作时强烈建议安装。

[[command title="按需验证开发工具"]]
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
$md$,
    updated_at = NOW()
WHERE slug = 'getting-started'
  AND content_md LIKE '%1. 安装 Git 和 Node.js。%'
  AND content_md LIKE '%Node.js v18 或更高版本。%'
  AND updated_at = created_at;

UPDATE tutorial_pages
SET content_md = replace(
        replace(
            content_md,
            '新手优先使用 Codex App。先创建 `codex` 类型 API 密钥，然后写入 `config.toml` 和 `auth.json`。',
            '新手优先使用 Codex App。Codex App 的安装和登录不要求先安装 Git 或 Node.js；Node.js/npm 主要用于 Codex CLI，Git 主要用于代码项目版本管理。先创建 `codex` 类型 API 密钥，然后写入 `config.toml` 和 `auth.json`。'
        ),
        E'## CLI 用户可选安装\n\n[[command title="npm 镜像安装 Codex CLI"]]',
        E'## CLI 用户可选安装\n\n只有使用终端版 Codex CLI 时才需要 Node.js/npm。当前 npm 包声明支持 Node.js 16+，新机器建议直接安装当前 LTS 版本。\n\n[[command title="npm 镜像安装 Codex CLI"]]'
    ),
    updated_at = NOW()
WHERE slug = 'codex'
  AND content_md LIKE '%新手优先使用 Codex App。先创建 `codex` 类型 API 密钥%'
  AND content_md LIKE '%## CLI 用户可选安装%'
  AND updated_at = created_at;

UPDATE tutorial_pages
SET content_md = replace(
        content_md,
        E'## 安装\n\n[[command title="安装 Claude Code"]]',
        E'## 安装\n\nnpm 安装路径需要 Node.js 18+。Windows 用户如果不走 WSL，也要准备 Git for Windows / Git Bash；如果使用官方原生安装器，可以先按官方安装器流程执行，再用 `claude doctor` 检查环境。\n\n[[command title="安装 Claude Code"]]'
    ),
    updated_at = NOW()
WHERE slug = 'claude-code'
  AND content_md LIKE '%## 安装%'
  AND content_md NOT LIKE '%npm 安装路径需要 Node.js 18+%'
  AND updated_at = created_at;

UPDATE tutorial_pages
SET content_md = replace(
        replace(
            content_md,
            '换镜像源，再确认 Node.js 版本是 v18 或更高。',
            '先看是哪类工具安装失败。Codex App 通常不需要 Node.js；Codex CLI 需要 Node.js/npm，建议当前 LTS；Claude Code 的 npm 安装路径需要 Node.js 18+。npm 下载慢时再换镜像源。'
        ),
        'Git 和 Node.js 建议默认安装，不要手动改路径。',
        '如果是 Codex App，先确认是否已经安装桌面端并登录；如果是 Codex CLI 或 Claude Code，再确认 Node.js/npm 是否在 PATH 里。代码项目里找不到 git 时，再检查 Git for Windows / Git Bash 安装路径。'
    ),
    updated_at = NOW()
WHERE slug = 'faq'
  AND content_md LIKE '%换镜像源，再确认 Node.js 版本是 v18 或更高。%'
  AND content_md LIKE '%Git 和 Node.js 建议默认安装，不要手动改路径。%'
  AND updated_at = created_at;
