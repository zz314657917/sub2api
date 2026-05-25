import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import SubscriptionPlanCard from '../SubscriptionPlanCard.vue'
import type { SubscriptionPlan } from '@/types/payment'

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

const mountPlanCard = (plan: SubscriptionPlan) =>
  mount(SubscriptionPlanCard, {
    props: { plan },
  })

describe('SubscriptionPlanCard', () => {
  it('hides zero daily limit and keeps positive cycle limits visible', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      daily_limit_usd: 0,
      weekly_limit_usd: 350,
      monthly_limit_usd: 1250,
    })

    expect(wrapper.text()).not.toContain('payment.planCard.dailyLimit')
    expect(wrapper.text()).not.toContain('$0')
    expect(wrapper.text()).toContain('payment.planCard.weeklyLimit')
    expect(wrapper.text()).toContain('$350')
    expect(wrapper.text()).toContain('payment.planCard.monthlyLimit')
    expect(wrapper.text()).toContain('$1250')
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
    expect(wrapper.text()).not.toContain('$0')
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

  it('normalizes generated model scope feature copy', () => {
    const wrapper = mountPlanCard({
      ...basePlan,
      group_platform: 'antigravity',
      features: ['covers 3 model scopes'],
    })

    expect(wrapper.text()).toContain('payment.pricing.feature.gptModels')
    expect(wrapper.text()).not.toContain('covers 3 model scopes')
  })
})
