import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import GroupAccountPriorityModal from '../GroupAccountPriorityModal.vue'

const {
  getGroupAccounts,
  updateGroupAccountPriorities,
  showError,
  showSuccess
} = vi.hoisted(() => ({
  getGroupAccounts: vi.fn(),
  updateGroupAccountPriorities: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      getGroupAccounts,
      updateGroupAccountPriorities
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const VueDraggableStub = defineComponent({
  name: 'VueDraggable',
  props: {
    modelValue: { type: Array, required: true },
    disabled: { type: Boolean, default: false }
  },
  emits: ['update:modelValue', 'end'],
  template: '<div class="draggable-stub"><slot /></div>'
})

const accounts = [
  { id: 1, name: 'First account', priority: 1, account_groups: [{ group_id: 7, priority: 1 }] },
  { id: 2, name: 'Second account', priority: 2, account_groups: [{ group_id: 7, priority: 2 }] },
  { id: 3, name: 'Third account', priority: 3, account_groups: [{ group_id: 7, priority: 3 }] }
].map((account) => ({
  ...account,
  platform: 'openai',
  type: 'apikey',
  status: 'active'
}))

function mountModal() {
  return mount(GroupAccountPriorityModal, {
    props: {
      show: false,
      group: { id: 7, name: 'OpenAI group', platform: 'openai' }
    } as any,
    global: {
      stubs: {
        BaseDialog: {
          props: ['show'],
          template: '<div v-if="show"><slot /></div>'
        },
        Icon: true,
        PlatformIcon: true,
        VueDraggable: VueDraggableStub
      }
    }
  })
}

describe('GroupAccountPriorityModal drag sorting', () => {
  beforeEach(() => {
    getGroupAccounts.mockReset()
    updateGroupAccountPriorities.mockReset().mockResolvedValue({ message: 'ok' })
    showError.mockReset()
    showSuccess.mockReset()
    getGroupAccounts.mockResolvedValue(accounts)
  })

  it('uses a drag handle instead of arrow buttons and saves the dragged order', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.findAll('.account-priority-drag-handle')).toHaveLength(3)
    expect(wrapper.findAll('[title="admin.groups.accountPriority.moveUp"]').length).toBe(0)
    expect(wrapper.findAll('[title="admin.groups.accountPriority.moveDown"]').length).toBe(0)

    const draggable = wrapper.findComponent(VueDraggableStub)
    const currentRows = draggable.props('modelValue') as Array<{ account_id: number; group_priority: number }>
    await draggable.vm.$emit('update:modelValue', [...currentRows].reverse())
    await nextTick()
    await draggable.vm.$emit('end')
    await nextTick()

    const visibleNames = wrapper
      .findAll('.min-w-0 > .flex > .truncate')
      .map((node) => node.text())
    expect(visibleNames).toEqual(['Third account', 'Second account', 'First account'])

    const saveButton = wrapper.findAll('button').find((button) => button.text() === 'common.save')
    expect(saveButton).toBeDefined()
    await saveButton!.trigger('click')
    await flushPromises()

    expect(updateGroupAccountPriorities).toHaveBeenCalledWith(7, [
      { account_id: 3, priority: 1 },
      { account_id: 2, priority: 2 },
      { account_id: 1, priority: 3 }
    ])
  })
})
