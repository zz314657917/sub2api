import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import UserEditModal from '../UserEditModal.vue'

const { updateUser } = vi.hoisted(() => ({
  updateUser: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: { update: updateUser },
    userAttributes: { updateUserAttributeValues: vi.fn() }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const user: AdminUser = {
  id: 42,
  username: 'leaderboard-user',
  email: 'leaderboard@example.com',
  role: 'user',
  balance: 0,
  concurrency: 1,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-07-11T00:00:00Z',
  updated_at: '2026-07-11T00:00:00Z',
  notes: '',
  exclude_from_leaderboard: true
}

describe('UserEditModal', () => {
  beforeEach(() => {
    updateUser.mockReset()
    updateUser.mockResolvedValue(user)
  })

  it('loads and submits the leaderboard participation flag, including false', async () => {
    const wrapper = mount(UserEditModal, {
      props: { show: true, user },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          UserAttributeForm: true,
          Icon: true
        }
      }
    })

    const checkbox = wrapper.get<HTMLInputElement>('#exclude-from-leaderboard')
    expect(checkbox.element.checked).toBe(true)

    await checkbox.setValue(false)
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateUser).toHaveBeenCalledWith(42, expect.objectContaining({
      exclude_from_leaderboard: false
    }))
  })
})
