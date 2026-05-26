import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UserApiKeyOnboardingDialog from '../UserApiKeyOnboardingDialog.vue'

const { openSupportPopupMock } = vi.hoisted(() => ({
  openSupportPopupMock: vi.fn(),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'dashboard.onboarding.title': '先创建一个 API 密钥',
        'dashboard.onboarding.description': 'API 密钥是接入工具的凭证。',
        'dashboard.onboarding.badge': '新用户接入引导',
        'dashboard.onboarding.trialBadge': '新人福利',
        'dashboard.onboarding.trialTitle': '首次调用后领免费额度',
        'dashboard.onboarding.trialDescription': '创建 API 密钥后发起首次调用，API 专用试用额度会自动启用并抵扣，无需先充值。',
        'dashboard.onboarding.balanceDescription': '账户余额已发放到账号，创建 API 密钥后即可开始调用，系统会按实际消耗自动扣减。',
        'dashboard.onboarding.rewardDescription': '创建 API 密钥并完成一次真实调用后，到福利中心领取免费额度。',
        'dashboard.onboarding.trialQuotaFallback': 'API 试用额度',
        'dashboard.onboarding.trialQuotaAmount': '0.1 API 试用额度',
        'dashboard.onboarding.balanceFallback': '账户余额',
        'dashboard.onboarding.balanceAmount': '0.1 账户余额',
        'dashboard.onboarding.rewardFallback': '免费额度',
        'dashboard.onboarding.rewardAmount': '1 免费额度',
        'dashboard.onboarding.walletNotice': 'API 专用试用额度仅用于调用，不会显示为钱包余额。',
        'dashboard.onboarding.walletBalanceNotice': '可在钱包余额中查看，调用时按实际消耗扣减。',
        'dashboard.onboarding.pillAutoActivate': '首调用自动抵扣',
        'dashboard.onboarding.pillBalanceDeduct': '调用自动扣减',
        'dashboard.onboarding.pillGpt55': 'GPT-5.5',
        'dashboard.onboarding.pillImage2': 'Image2',
        'dashboard.onboarding.pillClaudeCode': 'Claude Code',
        'dashboard.onboarding.pillWelfareClaim': '福利中心领取',
        'dashboard.onboarding.pillNoRecharge': '无需充值体验',
        'dashboard.onboarding.createKey': '创建 API 密钥',
        'dashboard.onboarding.joinGroup': '联系客服',
        'dashboard.onboarding.viewTutorial': '查看教程',
        'dashboard.onboarding.skip': '暂时跳过',
        'dashboard.onboarding.stepKeyTitle': '创建密钥',
        'dashboard.onboarding.stepKeyDescription': '生成 API Key，接入 Codex、Claude Code 或其他工具。',
        'dashboard.onboarding.stepToolTitle': '接入工具',
        'dashboard.onboarding.stepToolDescription': '按教程配置工具。',
        'dashboard.onboarding.stepTrialTitle': '发起首次调用',
        'dashboard.onboarding.stepTrialDescription': '发送一次真实请求，试用额度会自动启用并抵扣。',
        'dashboard.onboarding.stepBalanceDescription': '使用新密钥发送一次真实请求，费用会从账户余额中扣减。',
        'dashboard.onboarding.stepRewardCallDescription': '发送一次真实请求，完成领取条件。',
        'dashboard.onboarding.stepRewardClaimTitle': '领取额度',
        'dashboard.onboarding.stepRewardTitle': '领取 1 免费额度',
        'dashboard.onboarding.stepRewardDescription': '到福利中心领取免费额度。',
        'dashboard.onboarding.stepUsageTitle': '查看用量',
        'dashboard.onboarding.stepUsageDescription': '查看消费和调用记录。',
      }
      return messages[key] ?? key
    },
  }),
}))

vi.mock('@/utils/supportPopup', () => ({
  openSupportPopup: openSupportPopupMock,
}))

describe('UserApiKeyOnboardingDialog', () => {
  beforeEach(() => {
    openSupportPopupMock.mockClear()
  })

  it('renders onboarding content and emits primary actions', async () => {
    const wrapper = mount(UserApiKeyOnboardingDialog, {
      props: {
        show: true,
        hasBenefit: true,
        benefitKind: 'reward',
        benefitLabel: '1 免费额度',
        benefitRewardLabel: '1 免费额度',
      },
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('首次调用后领免费额度')
    expect(wrapper.text()).toContain('创建 API 密钥并完成一次真实调用后，到福利中心领取免费额度')
    expect(wrapper.text()).toContain('GPT-5.5')
    expect(wrapper.text()).toContain('Image2')
    expect(wrapper.text()).toContain('Claude Code')
    expect(wrapper.text()).not.toContain('1 免费额度')
    expect(wrapper.text()).not.toContain('首次调用成功后')
    expect(wrapper.text()).toContain('创建密钥')
    expect(wrapper.text()).toContain('发起首次调用')
    expect(wrapper.text()).toContain('领取额度')
    expect(wrapper.find('.pointer-events-none').exists()).toBe(true)
    expect(wrapper.find('.items-center.justify-center').exists()).toBe(true)
    expect(wrapper.find('img[src="/onboarding/new-user-trial-popup-header.png"]').exists()).toBe(true)
    expect(wrapper.find('.modal-overlay').exists()).toBe(false)

    await wrapper.get('button.btn-primary').trigger('click')
    expect(wrapper.emitted('create')).toHaveLength(1)

    const buttonByText = (text: string) => wrapper.findAll('button').find(button => button.text() === text)
    await buttonByText('暂时跳过')?.trigger('click')
    await buttonByText('查看教程')?.trigger('click')
    await buttonByText('联系客服')?.trigger('click')

    expect(wrapper.emitted('skip')).toHaveLength(1)
    expect(wrapper.emitted('tutorial')).toHaveLength(1)
    expect(openSupportPopupMock).toHaveBeenCalledTimes(1)
  })

  it('falls back to plain API key onboarding when trial is unavailable', () => {
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
    expect(wrapper.text()).toContain('接入工具')
    expect(wrapper.text()).toContain('查看用量')
    expect(wrapper.text()).not.toContain('这笔额度仅用于 API 调用')
  })
})
