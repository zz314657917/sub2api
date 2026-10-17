import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import type { Group } from '@/types'
import AccountGroupsCell from '../AccountGroupsCell.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: { count: number }) => params ? `${key}:${params.count}` : key
  })
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ cachedPublicSettings: null }) }))
enableAutoUnmount(afterEach)

const longName = '完整分组名称不应被省略-' + 'LongUnbrokenGroupName'.repeat(8)
const groups = Array.from({ length: 6 }, (_, index) => ({
  id: index + 1,
  name: `${longName}-${index + 1}`,
  platform: 'openai',
  subscription_type: 'standard',
  rate_multiplier: 1
} as Group))

describe('account group full names', () => {
  it('renders full names with wrapping classes on the actual badges', () => {
    const wrapper = mount(AccountGroupsCell, { props: { groups: groups.slice(0, 2) } })
    expect(wrapper.classes()).toContain('max-w-56')
    expect(wrapper.get('div.flex').classes()).toContain('flex-wrap')
    expect(wrapper.classes()).not.toContain('overflow-hidden')
    const badges = wrapper.findAllComponents(GroupBadge)
    expect(badges).toHaveLength(2)
    badges.forEach((badge, index) => {
      expect(badge.props('truncate')).toBe(false)
      expect(badge.classes()).toEqual(expect.arrayContaining(['min-w-0', 'max-w-full']))
      const name = badge.get('span.whitespace-normal')
      expect(name.text()).toBe(groups[index]!.name)
      expect(name.classes()).toEqual(expect.arrayContaining(['min-w-0', '[overflow-wrap:anywhere]']))
      expect(name.classes()).not.toContain('truncate')
    })
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('keeps GroupBadge truncation enabled by default for other callers', () => {
    const wrapper = mount(GroupBadge, { props: { name: longName } })
    expect(wrapper.props('truncate')).toBe(true)
    expect(wrapper.get('span.truncate').text()).toBe(longName)
    expect(wrapper.find('.whitespace-normal').exists()).toBe(false)
  })

  it('expands +N to all groups in the body popover and closes on Escape', async () => {
    const wrapper = mount(AccountGroupsCell, { props: { groups }, attachTo: document.body })
    const badges = wrapper.findAllComponents(GroupBadge)
    expect(badges).toHaveLength(3)
    expect(badges.map(badge => badge.props('name'))).toEqual(groups.slice(0, 3).map(group => group.name))
    expect(wrapper.get('button').text()).toBe('+3')
    expect(document.body.querySelector('.fixed.z-50')).toBeNull()

    await wrapper.get('button').trigger('click')
    const popover = document.body.querySelector('.fixed.z-50')!
    expect(popover).not.toBeNull()
    expect(popover.textContent).toContain('admin.accounts.groupCountTotal:6')
    const names = [...popover.querySelectorAll('span.whitespace-normal')]
    expect(names.map(name => name.textContent)).toEqual(groups.map(group => group.name))
    names.forEach(name => {
      expect(name.classList.contains('[overflow-wrap:anywhere]')).toBe(true)
      expect(name.classList.contains('truncate')).toBe(false)
      expect(name.parentElement!.classList.contains('max-w-full')).toBe(true)
    })
    expect(popover.querySelector('.flex-wrap.overflow-y-auto')).not.toBeNull()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(document.body.querySelector('.fixed.z-50')).toBeNull()
    expect(wrapper.findAllComponents(GroupBadge)).toHaveLength(3)
    expect(wrapper.get('button').text()).toBe('+3')
  })
})
