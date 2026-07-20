import { describe, expect, it } from 'vitest'

import en from '../locales/en/admin/accounts'
import zh from '../locales/zh/admin/accounts'

describe('OpenAI WS mode locale descriptions', () => {
  it('documents the global v2 router requirement using local account modes', () => {
    const descriptions = [zh.openai.wsModeDesc, en.openai.wsModeDesc]

    for (const description of descriptions) {
      expect(description).toContain('gateway.openai_ws.mode_router_v2_enabled=true')
      expect(description).toContain('ctx_pool')
      expect(description).toContain('passthrough')
      expect(description).not.toContain('http_bridge')
    }
  })
})
