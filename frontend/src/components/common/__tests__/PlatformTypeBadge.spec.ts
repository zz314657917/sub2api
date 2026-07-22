import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import PlatformTypeBadge from '../PlatformTypeBadge.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const mountBadge = (planType?: string, platform: 'openai' | 'anthropic' = 'openai') => mount(PlatformTypeBadge, {
  props: {
    platform,
    type: 'oauth',
    planType
  },
  global: {
    stubs: {
      PlatformIcon: true,
      Icon: true
    }
  }
})

describe('PlatformTypeBadge plan labels', () => {
  it('normalizes K12 and Pro aliases', () => {
    expect(mountBadge('k12').text()).toContain('K12')
    expect(mountBadge('ChatGPTPro').text()).toContain('Pro')
  })

  it('categorizes unknown and missing OpenAI plan values', () => {
    expect(mountBadge('education').text()).toContain('admin.accounts.planTypeOther')
    expect(mountBadge().text()).toContain('admin.accounts.planTypeUnrecognized')
  })

  it('does not add an unrecognized plan label to another platform', () => {
    expect(mountBadge(undefined, 'anthropic').text()).not.toContain('admin.accounts.planTypeUnrecognized')
  })
})
