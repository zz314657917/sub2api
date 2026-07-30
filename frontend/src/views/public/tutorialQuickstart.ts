export type QuickstartPlatformID = 'codex' | 'claude'

export interface QuickstartTutorialHeader {
  kicker: string
  title: string
  description: string
  library_action_label: string
  keys_action_label: string
  platform_control_label: string
  terminal_control_label: string
}

export interface QuickstartTutorialPlatform {
  id: QuickstartPlatformID
  label: string
  client_name: string
  base_url: string
  base_url_description: string
  auth_hint: string
  protocol: string
  model_hint: string
}

export interface QuickstartTutorialFacts {
  base_url_label: string
  auth_label: string
  auth_description: string
  protocol_label: string
  protocol_description: string
  model_label: string
  model_description: string
}

export interface QuickstartTutorialSection {
  kicker: string
  title: string
  description: string
}

export interface QuickstartTutorialDesktopTile {
  number: string
  title: string
  description: string
}

export interface QuickstartTutorialConfig {
  version: number
  header: QuickstartTutorialHeader
  platforms: QuickstartTutorialPlatform[]
  facts: QuickstartTutorialFacts
  desktop: QuickstartTutorialSection & { tiles: QuickstartTutorialDesktopTile[] }
  api: QuickstartTutorialSection
  api_hint: string
  troubleshooting: QuickstartTutorialSection
  errors: Array<{ code: string; title: string; description: string }>
}

export const defaultQuickstartTutorialConfig: QuickstartTutorialConfig = {
  version: 1,
  header: {
    kicker: 'QUICK START',
    title: '使用文档',
    description: '选择模型平台和终端环境，按步骤完成第一次接入。',
    library_action_label: '查看完整教程',
    keys_action_label: '查看密钥',
    platform_control_label: '模型平台',
    terminal_control_label: '系统 / 终端'
  },
  platforms: [
    {
      id: 'codex',
      label: 'ChatGPT / Codex',
      client_name: 'ChatGPT / Codex',
      base_url: 'https://ai.3zapi.top',
      base_url_description: 'ChatGPT / Codex 使用根地址，无需追加 /v1。',
      auth_hint: 'Bearer / OPENAI_API_KEY',
      protocol: 'OpenAI Responses',
      model_hint: 'Codex 模型列表'
    },
    {
      id: 'claude',
      label: 'Claude',
      client_name: 'Claude',
      base_url: 'https://ai.3zapi.top',
      base_url_description: 'Claude 兼容工具使用根地址。',
      auth_hint: 'Bearer / ANTHROPIC_AUTH_TOKEN',
      protocol: 'Anthropic Messages',
      model_hint: 'Claude 模型列表'
    }
  ],
  facts: {
    base_url_label: 'Base URL',
    auth_label: '鉴权',
    auth_description: '密钥来自控制台的 API 密钥页面。',
    protocol_label: '协议',
    protocol_description: '端点必须和客户端的协议模式一致。',
    model_label: '模型',
    model_description: '将示例模型替换为当前账号可用的模型 ID。'
  },
  desktop: {
    kicker: 'DESKTOP',
    title: '接入桌面端',
    description: '桌面端、CLI 和配置文件共用同一套 Base URL 与 API Key。',
    tiles: [
      { number: '01', title: '完成 CLI 配置', description: '先按上方步骤写入 Base URL 和鉴权。' },
      { number: '02', title: '复用 auth.json', description: '同一份本地密钥可以供 CLI、桌面端和兼容插件使用。' },
      { number: '03', title: '启动桌面端', description: '打开项目后发送简单问题，确认模型响应和额度均正常。' }
    ]
  },
  api: {
    kicker: 'API',
    title: '在你自己的程序里调用 API',
    description: '使用当前平台对应的兼容协议发起一次最小请求。'
  },
  api_hint: '多轮对话请按实际上游能力决定是否使用 `store` 与 `previous_response_id`。',
  troubleshooting: {
    kicker: 'TROUBLESHOOTING',
    title: '常见错误码',
    description: '先确认鉴权、模型、余额和请求协议，再查看服务端返回信息。'
  },
  errors: [
    { code: '400', title: '请求格式错误', description: '检查 JSON、模型参数和所使用的 API 协议是否匹配。' },
    { code: '401', title: '密钥无效', description: '确认 API Key 完整、未过期，并使用正确的鉴权字段。' },
    { code: '402', title: '余额或权益不足', description: '补充余额、兑换权益码或确认当前账号仍有有效套餐。' },
    { code: '403', title: '无模型或分组权限', description: '当前密钥可能未绑定可用分组，或模型不在允许列表。' },
    { code: '404', title: '端点或模型不存在', description: '确认 Claude 使用 Messages，Codex 使用 Responses，模型以列表为准。' },
    { code: '429', title: '请求过快', description: '降低并发，等待 Retry-After 后重试，避免重复提交。' },
    { code: 'CAPACITY', title: '官方算力不足', description: 'Selected model is at capacity. Please try a different model. 请切换其他模型后重试。' },
    { code: '5xx', title: '链路异常', description: '系统会自动回落；持续出现时记录请求 ID 并联系管理员。' }
  ]
}
