import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import AdminCafeRoomsView from '../AdminCafeRoomsView.vue'

const {
  listRooms,
  createRoom,
  updateRoom,
  removeRoom,
  bulkCreate,
  openRound,
  listAccountOptions,
  listPendingRounds,
  listRoundAccountOptions,
  assignRoundAccount,
  getWorkstationLayout,
  updateWorkstationLayout,
  listPlans,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  listRooms: vi.fn(),
  createRoom: vi.fn(),
  updateRoom: vi.fn(),
  removeRoom: vi.fn(),
  bulkCreate: vi.fn(),
  openRound: vi.fn(),
  listAccountOptions: vi.fn(),
  listPendingRounds: vi.fn(),
  listRoundAccountOptions: vi.fn(),
  assignRoundAccount: vi.fn(),
  getWorkstationLayout: vi.fn(),
  updateWorkstationLayout: vi.fn(),
  listPlans: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    cafeRooms: {
      list: listRooms,
      listAccountOptions,
      listPendingRounds,
      listRoundAccountOptions,
      assignRoundAccount,
      getWorkstationLayout,
      updateWorkstationLayout,
      create: createRoom,
      update: updateRoom,
      remove: removeRoom,
      bulkCreate,
      openRound,
    },
    groupBuy: { listPlans },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

const labels: Record<string, string> = {
  'admin.pixelCafe.refresh': '刷新',
  'admin.pixelCafe.bulkCreate': '批量创建',
  'admin.pixelCafe.layout.open': '大厅布局',
  'admin.pixelCafe.layout.title': '编辑大厅电脑位置',
  'admin.pixelCafe.layout.save': '保存布局',
  'admin.pixelCafe.layout.saving': '保存中',
  'admin.pixelCafe.layout.reset': '重置默认位置',
  'admin.pixelCafe.layout.hint': '拖动电脑',
  'admin.pixelCafe.layout.snap': '吸附网格',
  'admin.pixelCafe.layout.count': '电脑工位数量',
  'admin.pixelCafe.layout.countRange': '可设置 {min}–{max} 个',
  'admin.pixelCafe.layout.decreaseCount': '减少一个电脑工位',
  'admin.pixelCafe.layout.increaseCount': '增加一个电脑工位',
  'admin.pixelCafe.layout.desktopOnly': '请使用桌面端',
  'admin.pixelCafe.layout.keyboardHint': '方向键微调',
  'admin.pixelCafe.success.layoutSaved': '大厅电脑布局已保存',
  'admin.pixelCafe.createRoom': '新建房间',
  'admin.pixelCafe.noRooms': '暂无房间',
  'admin.pixelCafe.noRoomPlans': '暂无 Room 计划',
  'admin.pixelCafe.actions.edit': '编辑',
  'admin.pixelCafe.actions.openRound': '开团',
  'admin.pixelCafe.actions.openingRound': '开团中',
  'admin.pixelCafe.actions.delete': '删除',
  'admin.pixelCafe.form.createTitle': '新建房间',
  'admin.pixelCafe.form.editTitle': '编辑房间',
  'admin.pixelCafe.form.save': '保存房间',
  'admin.pixelCafe.bulk.title': '批量创建房间',
  'admin.pixelCafe.bulk.submit': '开始创建',
  'admin.pixelCafe.accountDeferred': '成团后配号',
  'admin.pixelCafe.pending.title': '待配号轮次',
  'admin.pixelCafe.pending.description': '份额售罄后在这里绑定账号',
  'admin.pixelCafe.pending.assign': '选择账号',
  'admin.pixelCafe.pending.assignTitle': '绑定并开通',
  'admin.pixelCafe.pending.accountSearch': '搜索账号',
  'admin.pixelCafe.success.created': '房间已创建',
  'admin.pixelCafe.success.updated': '房间已更新',
  'admin.pixelCafe.success.deleted': '房间已删除',
  'admin.pixelCafe.success.roundOpened': 'open Round 已创建',
  'common.cancel': '取消',
  'common.close': '关闭',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        let value = labels[key] ?? key
        for (const [name, replacement] of Object.entries(params ?? {})) {
          value = value.replace(`{${name}}`, String(replacement))
        }
        return value
      },
    }),
  }
})

const DataTableStub = defineComponent({
  props: { data: { type: Array, default: () => [] } },
  template: `
    <div>
      <div v-for="row in data" :key="row.id" class="room-row">
        <slot name="cell-room" :row="row" />
        <slot name="cell-plan" :row="row" />
        <slot name="cell-account" :row="row" />
        <slot name="cell-status" :row="row" />
        <slot name="cell-actions" :row="row" />
      </div>
      <slot v-if="data.length === 0" name="empty" />
    </div>
  `,
})

const BaseDialogStub = defineComponent({
  props: { show: Boolean, title: String },
  template: '<section v-if="show" class="dialog"><h2>{{ title }}</h2><slot /><slot name="footer" /></section>',
})

const ConfirmDialogStub = defineComponent({
  props: { show: Boolean },
  emits: ['confirm', 'cancel'],
  template: '<div v-if="show" class="confirm-dialog"><button type="button" @click="$emit(\'confirm\')">confirm-delete</button></div>',
})

function room(status: 'enabled' | 'maintenance' = 'enabled') {
  return {
    id: 7,
    code: 'ROOM-007',
    name: 'OpenAI 七号房',
    plan_id: 21,
    account_id: null,
    zone_key: 'openai',
    theme_key: 'warm_wood',
    scene_slot_key: 'openai-7',
    status,
    featured: true,
    sort_order: 7,
    plan: {
      id: 21,
      title: 'ChatGPT Plus',
      target_group_id: 5,
      fulfillment_mode: 'room_subscription',
      total_shares: 5,
      subscription_tier: 'plus',
      max_buyers: 4,
      max_shares_per_user: 4,
      timeout_minutes: 60,
      validity_days: 30,
      group_platform: 'openai',
      group_access_mode: 'room_managed',
    },
    created_at: '',
    updated_at: '',
  }
}

function mountView() {
  return mount(AdminCafeRoomsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: { template: '<div><slot name="filters"/><slot name="table"/><slot name="pagination"/><slot/></div>' },
        DataTable: DataTableStub,
        Pagination: true,
        Select: true,
        EmptyState: { props: ['title'], template: '<div>{{ title }}</div>' },
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        Icon: true,
        AdminGroupBuyView: {
          props: { embedded: Boolean },
          template: '<section data-testid="embedded-group-buy" :data-embedded="String(embedded)">拼团工作区</section>',
        },
      },
    },
  })
}

describe('AdminCafeRoomsView', () => {
  beforeEach(() => {
    listRooms.mockReset().mockResolvedValue({ data: { items: [room()], total: 1, page: 1, page_size: 20, pages: 1 } })
    listPlans.mockReset().mockResolvedValue({ data: [
      { id: 20, title: 'Legacy plan', target_group_id: 4, fulfillment_mode: 'aggregate_tier' },
      { id: 21, title: 'ChatGPT Plus', target_group_id: 5, fulfillment_mode: 'room_subscription', subscription_tier: 'plus' },
    ] })
    listAccountOptions.mockReset().mockResolvedValue({ data: { items: [
      { id: 41, name: 'OpenAI account', platform: 'openai', status: 'active', email_masked: 'o***i@example.com' },
      { id: 42, name: 'Second account', platform: 'openai', status: 'active', email_masked: 's***d@example.com' },
    ], total: 2, page: 1, page_size: 20, pages: 1 } })
    listPendingRounds.mockReset().mockResolvedValue({ data: { items: [{
      id: 81, status: 'awaiting_account', room_id: 7, room_code: 'ROOM-007', room_name: 'OpenAI 七号房', subscription_tier: 'plus', paid_shares: 10, total_shares: 10, joined_buyers: 4, max_buyers: 4,
    }], total: 1, page: 1, page_size: 20, pages: 1 } })
    listRoundAccountOptions.mockReset().mockResolvedValue({ data: { items: [
      { id: 41, name: 'Plus account', platform: 'openai', status: 'active', plan_type: 'plus', email_masked: 'o***i@example.com' },
    ], total: 1, page: 1, page_size: 30, pages: 1 } })
    assignRoundAccount.mockReset().mockResolvedValue({ data: { id: 81, status: 'active' } })
    const layout = Array.from({ length: 10 }, (_, index) => ({ id: index + 1, x: 300 + index * 20, y: 200 + index * 10 }))
    getWorkstationLayout.mockReset().mockResolvedValue({ data: layout })
    updateWorkstationLayout.mockReset().mockImplementation(async draft => ({ data: draft }))
    createRoom.mockReset().mockResolvedValue({ data: room() })
    updateRoom.mockReset().mockResolvedValue({ data: room() })
    removeRoom.mockReset().mockResolvedValue({ data: { message: 'ok' } })
    openRound.mockReset().mockResolvedValue({ data: { id: 81, status: 'open' } })
    bulkCreate.mockReset().mockResolvedValue({ data: {
      created: [{ room: room() }],
      failed: [{ index: 2, error_code: 'CAFE_ROOM_CREATE_FAILED', message: 'failed' }],
    } })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads rooms and exposes only room_subscription plans in the create form', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listRooms).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20, sort_by: 'sort_order' }))
    expect(listAccountOptions).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('OpenAI 七号房')
    expect(wrapper.text()).toContain('成团后配号')

    const createButton = wrapper.findAll('button').find((button) => button.text().includes('新建房间'))
    await createButton?.trigger('click')
    expect(wrapper.text()).toContain('ChatGPT Plus')
    expect(wrapper.text()).not.toContain('Legacy plan')
  })

  it('renders room management and embedded plan management together', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('OpenAI 七号房')
    expect(wrapper.find('[data-testid="embedded-group-buy"]').attributes('data-embedded')).toBe('true')
  })

  it('loads, resizes, resets, edits, and saves one shared lobby workstation layout', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.text().includes('大厅布局'))?.trigger('click')
    await flushPromises()
    expect(getWorkstationLayout).toHaveBeenCalledTimes(1)
    expect(wrapper.findAll('[data-testid="cafe-layout-workstation"]')).toHaveLength(10)

    await wrapper.get('[data-testid="cafe-layout-count-input"]').setValue(12)
    expect(wrapper.findAll('[data-testid="cafe-layout-workstation"]')).toHaveLength(12)
    await wrapper.findAll('button').find(button => button.text().includes('重置默认位置'))?.trigger('click')
    expect(wrapper.findAll('[data-testid="cafe-layout-workstation"]')).toHaveLength(12)

    await wrapper.findAll('[data-testid="cafe-layout-workstation"]')[0].trigger('keydown', { key: 'ArrowRight' })
    await wrapper.findAll('button').find(button => button.text().includes('保存布局'))?.trigger('click')
    await flushPromises()

    const savedLayout = updateWorkstationLayout.mock.calls[0][0]
    expect(savedLayout).toHaveLength(12)
    expect(savedLayout.map((slot: { id: number }) => slot.id)).toEqual(Array.from({ length: 12 }, (_, index) => index + 1))
    expect(savedLayout[0]).toEqual(expect.objectContaining({ id: 1, x: 344 }))
    expect(showSuccess).toHaveBeenCalledWith('大厅电脑布局已保存')
  })

  it('submits create input without client-owned price or group fields and opens a round', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('新建房间'))?.trigger('click')
    const form = wrapper.find('#cafe-room-form')
    const inputs = form.findAll('input')
    await inputs[0].setValue('ROOM-008')
    await inputs[1].setValue('八号房')
    const selects = form.findAll('select')
    await selects[0].setValue('21')
    await form.trigger('submit')
    await flushPromises()

    expect(createRoom).toHaveBeenCalledWith(expect.objectContaining({
      code: 'ROOM-008',
      name: '八号房',
      plan_id: 21,
    }))
    const payload = createRoom.mock.calls[0][0]
    expect(payload).not.toHaveProperty('price')
    expect(payload).not.toHaveProperty('group_id')
    expect(payload).not.toHaveProperty('account_id')

    await wrapper.findAll('button').find((button) => button.text().trim() === '开团')?.trigger('click')
    await flushPromises()
    expect(openRound).toHaveBeenCalledWith(7)
  })

  it('updates and deletes a non-enabled room through the S145 endpoints', async () => {
    listRooms.mockResolvedValue({ data: { items: [room('maintenance')], total: 1, page: 1, page_size: 20, pages: 1 } })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().trim() === '编辑')?.trigger('click')
    const form = wrapper.find('#cafe-room-form')
    await form.findAll('input')[1].setValue('维护房间')
    await form.trigger('submit')
    await flushPromises()
    expect(updateRoom).toHaveBeenCalledWith(7, expect.objectContaining({ name: '维护房间' }))

    await wrapper.findAll('button').find((button) => button.text().trim() === '删除')?.trigger('click')
    await wrapper.findAll('button').find((button) => button.text() === 'confirm-delete')?.trigger('click')
    await flushPromises()
    expect(removeRoom).toHaveBeenCalledWith(7)
  })

  it('bulk creates by quantity and renders per-item failures', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('批量创建'))?.trigger('click')
    const form = wrapper.find('#cafe-room-bulk-form')
    const selects = form.findAll('select')
    await selects[0].setValue('21')
    await form.find('input[type="number"]').setValue(3)
    await form.trigger('submit')
    await flushPromises()

    expect(bulkCreate).toHaveBeenCalledWith(expect.objectContaining({ plan_id: 21, quantity: 3 }))
    expect(bulkCreate.mock.calls[0][0]).not.toHaveProperty('account_ids')
    expect(wrapper.text()).toContain('CAFE_ROOM_CREATE_FAILED')
    expect(showSuccess).toHaveBeenCalled()
  })

  it('searches and assigns accounts only from the pending-round workspace', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(listPendingRounds).toHaveBeenCalledWith({ page: 1, page_size: 20, search: undefined })
    expect(wrapper.find('[data-testid="cafe-pending-fulfillment"]').text()).toContain('10/10 份')
    await wrapper.findAll('button').filter(button => button.text() === '选择账号').at(-1)?.trigger('click')
    await flushPromises()
    expect(listRoundAccountOptions).toHaveBeenCalledWith(81, { page: 1, page_size: 30, search: undefined })
    expect(wrapper.text()).toContain('o***i@example.com')
    await wrapper.find('input[type="radio"]').setValue(41)
    await wrapper.findAll('button').filter(button => button.text() === '选择账号').at(-1)?.trigger('click')
    await flushPromises()
    expect(assignRoundAccount).toHaveBeenCalledWith(81, 41)
    expect(showSuccess).toHaveBeenCalled()
  })

  it('filters pending rounds and account candidates server-side', async () => {
    const wrapper = mountView()
    await flushPromises()
    const pendingSearch = wrapper.find('[data-testid="cafe-pending-fulfillment"] input[type="search"]')
    await pendingSearch.setValue('七号')
    await pendingSearch.trigger('change')
    await flushPromises()
    expect(listPendingRounds).toHaveBeenLastCalledWith({ page: 1, page_size: 20, search: '七号' })
    await wrapper.findAll('button').find(button => button.text() === '选择账号')?.trigger('click')
    await flushPromises()
    const accountSearch = wrapper.find('.dialog input[type="search"]')
    await accountSearch.setValue('owner')
    await accountSearch.trigger('change')
    await flushPromises()
    expect(listRoundAccountOptions).toHaveBeenLastCalledWith(81, { page: 1, page_size: 30, search: 'owner' })
  })
})
