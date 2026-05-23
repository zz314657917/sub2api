import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserApiKeyOnboardingDialog from '../UserApiKeyOnboardingDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'dashboard.onboarding.title': '先创建一个 API 密钥',
        'dashboard.onboarding.description': 'API 密钥是接入工具的凭证。',
        'dashboard.onboarding.createKey': '创建 API 密钥',
        'dashboard.onboarding.viewTutorial': '查看教程',
        'dashboard.onboarding.skip': '暂时跳过',
        'dashboard.onboarding.stepKeyTitle': '创建密钥',
        'dashboard.onboarding.stepKeyDescription': '生成你的第一个 API 密钥。',
        'dashboard.onboarding.stepToolTitle': '接入工具',
        'dashboard.onboarding.stepToolDescription': '按教程配置工具。',
        'dashboard.onboarding.stepUsageTitle': '查看用量',
        'dashboard.onboarding.stepUsageDescription': '查看消费和调用记录。',
      }
      return messages[key] ?? key
    },
  }),
}))

describe('UserApiKeyOnboardingDialog', () => {
  it('renders onboarding content and emits primary actions', async () => {
    const wrapper = mount(UserApiKeyOnboardingDialog, {
      props: {
        show: true,
      },
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('先创建一个 API 密钥')
    expect(wrapper.text()).toContain('创建密钥')
    expect(wrapper.text()).toContain('接入工具')
    expect(wrapper.text()).toContain('查看用量')
    expect(wrapper.find('.pointer-events-none').exists()).toBe(true)
    expect(wrapper.find('.items-center.justify-center').exists()).toBe(true)
    expect(wrapper.find('.from-blue-600.to-sky-500').exists()).toBe(true)
    expect(wrapper.find('.bg-blue-50').exists()).toBe(true)
    expect(wrapper.find('.modal-overlay').exists()).toBe(false)

    await wrapper.get('button.btn-primary').trigger('click')
    expect(wrapper.emitted('create')).toHaveLength(1)

    const buttons = wrapper.findAll('button.btn-secondary')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    expect(wrapper.emitted('skip')).toHaveLength(1)
    expect(wrapper.emitted('tutorial')).toHaveLength(1)
  })
})
