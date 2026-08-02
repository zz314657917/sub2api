import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import UserErrorDetailModal from '../UserErrorDetailModal.vue'

const { getMyErrorDetail } = vi.hoisted(() => ({
  getMyErrorDetail: vi.fn(),
}))

vi.mock('@/api', () => ({
  usageAPI: {
    getMyErrorDetail,
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      te: () => true,
    }),
  }
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div v-if="show" data-test="base-dialog"><slot /></div>',
})

describe('UserErrorDetailModal', () => {
  beforeEach(() => {
    getMyErrorDetail.mockReset()
  })

  it('loads only after opening and renders the user-safe detail allowlist', async () => {
    getMyErrorDetail.mockResolvedValue({
      id: 41,
      created_at: '2026-08-01T00:00:00Z',
      model: 'safe-model',
      inbound_endpoint: '/v1/responses',
      status_code: 502,
      upstream_status_code: 503,
      category: 'upstream',
      platform: 'openai',
      message: 'safe message',
      error_body: '{"error":"safe body"}',
      key_name: 'owned-key',
      key_deleted: false,
      ip_address: '198.51.100.7',
      user_agent: 'secret-agent',
      email: 'hidden@example.com',
      account_name: 'hidden-account',
      upstream_endpoint: 'https://private-upstream.example/v1',
      retry_count: 9,
      key_prefix: 'sk-hidden',
      owner_source: 'hidden-owner',
    })

    const wrapper = mount(UserErrorDetailModal, {
      props: { show: false, errorId: 41 },
      global: { stubs: { BaseDialog: BaseDialogStub } },
    })

    expect(getMyErrorDetail).not.toHaveBeenCalled()

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(getMyErrorDetail).toHaveBeenCalledWith(41)
    expect(wrapper.get('[data-test="user-error-detail"]').text()).toContain('owned-key')
    expect(wrapper.text()).toContain('safe-model')
    expect(wrapper.text()).toContain('safe message')
    expect(wrapper.text()).toContain('safe body')
    expect(wrapper.text()).not.toContain('198.51.100.7')
    expect(wrapper.text()).not.toContain('secret-agent')
    expect(wrapper.text()).not.toContain('hidden@example.com')
    expect(wrapper.text()).not.toContain('hidden-account')
    expect(wrapper.text()).not.toContain('private-upstream.example')
    expect(wrapper.text()).not.toContain('sk-hidden')
    expect(wrapper.text()).not.toContain('hidden-owner')
  })

  it('clears loaded detail when closed', async () => {
    getMyErrorDetail.mockResolvedValue({
      id: 42,
      created_at: '2026-08-01T00:00:00Z',
      model: 'safe-model',
      inbound_endpoint: '/v1/messages',
      status_code: 400,
      category: 'invalid_request',
      platform: 'anthropic',
      message: 'invalid',
      error_body: '',
      key_name: 'owned-key',
      key_deleted: false,
    })

    const wrapper = mount(UserErrorDetailModal, {
      props: { show: false, errorId: 42 },
      global: { stubs: { BaseDialog: BaseDialogStub } },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()
    expect(wrapper.find('[data-test="user-error-detail"]').exists()).toBe(true)

    await wrapper.setProps({ show: false })
    expect(wrapper.find('[data-test="user-error-detail"]').exists()).toBe(false)
  })
})
