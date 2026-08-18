export function applyInterceptWarmup(
  credentials: Record<string, unknown>,
  enabled: boolean,
  mode: 'create' | 'edit'
): void {
  if (enabled) {
    credentials.intercept_warmup_requests = true
  } else if (mode === 'edit') {
    delete credentials.intercept_warmup_requests
  }
}

export type CnAccountMode = 'payg' | 'coding'
export type CnApiProtocol = 'chat_completions' | 'anthropic' | 'responses'

export interface CnBaseUrlPreset {
  mode: CnAccountMode
  protocol: CnApiProtocol
  label: string
  url: string
}

export const CN_BASE_URL_PRESETS: Record<'kimi' | 'zhipu' | 'deepseek', CnBaseUrlPreset[]> = {
  kimi: [
    { mode: 'payg', protocol: 'chat_completions', label: 'Moonshot', url: 'https://api.moonshot.cn/v1' },
    { mode: 'payg', protocol: 'anthropic', label: 'Moonshot Anthropic', url: 'https://api.moonshot.cn/anthropic' },
    { mode: 'coding', protocol: 'chat_completions', label: 'Kimi For Coding', url: 'https://api.kimi.com/coding/v1' },
    { mode: 'coding', protocol: 'anthropic', label: 'Kimi Coding Anthropic', url: 'https://api.kimi.com/coding' }
  ],
  zhipu: [
    { mode: 'payg', protocol: 'chat_completions', label: 'GLM PaaS', url: 'https://open.bigmodel.cn/api/paas/v4' },
    { mode: 'payg', protocol: 'anthropic', label: 'GLM Anthropic', url: 'https://open.bigmodel.cn/api/anthropic' },
    { mode: 'coding', protocol: 'chat_completions', label: 'GLM Coding', url: 'https://open.bigmodel.cn/api/coding/paas/v4' },
    { mode: 'coding', protocol: 'anthropic', label: 'GLM Coding Anthropic', url: 'https://open.bigmodel.cn/api/anthropic' }
  ],
  deepseek: [
    { mode: 'payg', protocol: 'chat_completions', label: 'DeepSeek', url: 'https://api.deepseek.com' },
    { mode: 'payg', protocol: 'anthropic', label: 'DeepSeek Anthropic', url: 'https://api.deepseek.com/anthropic' },
    { mode: 'payg', protocol: 'responses', label: 'DeepSeek Responses', url: 'https://api.deepseek.com' }
  ]
}

export function defaultCNBaseUrl(
  platform: string,
  mode: CnAccountMode,
  protocol: CnApiProtocol = 'chat_completions'
): string {
  return CN_BASE_URL_PRESETS[platform as keyof typeof CN_BASE_URL_PRESETS]
    ?.find(candidate => candidate.mode === mode && candidate.protocol === protocol)?.url ?? ''
}

export function cnQuotaCellVisible(platform: string, accountMode: string): boolean {
  return (platform === 'kimi' || platform === 'zhipu') && accountMode === 'coding'
}

export function cnBalanceCellVisible(platform: string, accountMode: string): boolean {
  return (platform === 'kimi' || platform === 'deepseek') && accountMode !== 'coding'
}
