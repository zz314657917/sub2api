import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import SubscriptionPlanCard from '../SubscriptionPlanCard.vue'
import type { SubscriptionPlan } from '@/types/payment'
import { useAppStore } from '@/stores/app'
import { CREDIT_SYMBOL } from '@/utils/credits'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const basePlan: SubscriptionPlan = {
  id: 1,
  group_id: 10,
  group_platform: 'openai',
  group_name: 'OpenAI',
  rate_multiplier: 1,
  daily_limit_usd: null,
  weekly_limit_usd: null,
  monthly_limit_usd: null,
  name: 'Pro',
  description: 'Test plan',
  price: 330,
  original_price: 0,
  validity_days: 30,
  validity_unit: 'day',
  features: [],
  for_sale: true,
  sort_order: 1,
}

const mountPlanCard = (plan: SubscriptionPlan) => {
  const pinia = createPinia()
  setActivePinia(pinia)
  const appStore = useAppStore()
  appStore.cachedPublicSettings = {
    server_utc_offset: '+08:00',
  } as typeof appStore.cachedPublicSettings

  return mount(SubscriptionPlanCard, {
    props: { plan },
    global: { plugins: [pinia] },
  })
}

describe('SubscriptionPlanCard', () => {
  it('hides zero daily limit and keeps positive cycle limits visible', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      daily_limit_usd: 0,
      weekly_limit_usd: 350,
      monthly_limit_usd: 1250,
    })

    expect(wrapper.text()).not.toContain('payment.planCard.dailyLimit')
    expect(wrapper.text()).not.toContain(`${CREDIT_SYMBOL} 0`)
    expect(wrapper.text()).toContain('payment.planCard.weeklyLimit')
    expect(wrapper.text()).toContain(`${CREDIT_SYMBOL} 350.00`)
    expect(wrapper.text()).toContain('payment.planCard.monthlyLimit')
    expect(wrapper.text()).toContain(`${CREDIT_SYMBOL} 1,250.00`)
  })

  it('treats zero cycle limits as unlimited', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      daily_limit_usd: 0,
      weekly_limit_usd: 0,
      monthly_limit_usd: 0,
    })

    expect(wrapper.text()).toContain('payment.planCard.quota')
    expect(wrapper.text()).toContain('payment.planCard.unlimited')
    expect(wrapper.text()).not.toContain(`${CREDIT_SYMBOL} 0`)
  })

  it('does not expose Antigravity model scope names for OpenAI plans', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      group_platform: 'openai',
      supported_model_scopes: ['claude', 'gemini_text', 'gemini_image'],
    })

    expect(wrapper.text()).not.toContain('Claude')
    expect(wrapper.text()).not.toContain('Gemini')
    expect(wrapper.text()).not.toContain('Imagen')
  })

  it('renders plural admin-form validity units instead of days', () => {
    expect(mountPlanCard({ ...basePlan, validity_days: 1, validity_unit: 'months' }).text()).toContain('/ payment.perMonth')
    expect(mountPlanCard({ ...basePlan, validity_days: 3, validity_unit: 'months' }).text()).toContain('/ 3payment.months')
    expect(mountPlanCard({ ...basePlan, validity_days: 2, validity_unit: 'weeks' }).text()).toContain('/ 2payment.weeks')
    expect(mountPlanCard({ ...basePlan, validity_days: 30, validity_unit: 'day' }).text()).toContain('/ 30payment.days')
  })

  it('normalizes generated model scope feature copy', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      group_platform: 'antigravity',
      features: ['covers 3 model scopes'],
    })

    expect(wrapper.text()).toContain('payment.pricing.feature.gptModels')
    expect(wrapper.text()).not.toContain('covers 3 model scopes')
  })

  it('labels peak rate windows with the server UTC offset', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      peak_rate_enabled: true,
      peak_start: '14:00',
      peak_end: '18:00',
      peak_rate_multiplier: 2,
    })

    expect(wrapper.text()).toContain('payment.planCard.peakRate')
    expect(wrapper.text()).toContain('14:00-18:00 ×2 (UTC+08:00)')
  })

  it.each([
    ['long Chinese', '企业全球加速专业订阅套餐（含高级模型与优先支持）'],
    ['long English', 'Enterprise Global Acceleration Subscription with Priority Support'],
    ['unbroken token', 'EnterpriseGlobalAccelerationSubscriptionWithPrioritySupport1234567890'],
  ])('keeps the full %s plan title accessible in a bounded two-line area', (_label, name) => {
    const wrapper = mountPlanCard({ ...basePlan, name })
    const title = wrapper.get('h3')

    expect(title.text()).toBe(name)
    expect(title.attributes('title')).toBe(name)
    expect(title.classes()).toEqual(expect.arrayContaining([
      'min-w-0',
      'h-12',
      'break-words',
      'line-clamp-2',
      '[overflow-wrap:anywhere]',
    ]))
    expect(title.classes()).not.toContain('truncate')
  })

  it('keeps the title, badge, price, and purchase action in separate bounded regions', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      name: 'Enterprise Global Acceleration Subscription with Priority Support',
      price: 123.45,
      description: 'Includes advanced models and priority support.',
    })
    const title = wrapper.get('h3')
    const badge = wrapper.findAll('span').find((node) => node.text() === 'OpenAI')

    expect(title.element.parentElement?.classList).toContain('min-w-0')
    expect(title.element.parentElement?.classList).toContain('flex-1')
    expect(badge?.classes()).toContain('shrink-0')
    expect([...((badge?.element.parentElement?.classList) ?? [])]).toEqual(expect.arrayContaining([
      'flex',
      'items-center',
      'justify-end',
    ]))
    expect(badge?.element.parentElement?.textContent).toContain('/ 30payment.days')
    expect(wrapper.get('p').text()).toBe('Includes advanced models and priority support.')
    expect(wrapper.get('button').text()).toBe('payment.subscribeNow')
  })

  it('keeps short plan titles compact and aligned', () => {
    const wrapper = mountPlanCard({ ...basePlan, name: 'Pro', description: '' })
    const title = wrapper.get('h3')
    const badge = wrapper.findAll('span').find((node) => node.text() === 'OpenAI')

    expect(title.text()).toBe('Pro')
    expect(title.attributes('title')).toBe('Pro')
    expect(title.classes()).toEqual(expect.arrayContaining(['text-base', 'font-bold', 'h-12']))
    expect([...((badge?.element.parentElement?.classList) ?? [])]).toEqual(expect.arrayContaining([
      'flex',
      'items-center',
      'justify-end',
    ]))
    expect(badge?.element.parentElement?.textContent).toContain('/ 30payment.days')
  })
})
