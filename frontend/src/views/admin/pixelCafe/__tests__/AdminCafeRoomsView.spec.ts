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
  listPlans,
  listAccounts,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  listRooms: vi.fn(),
  createRoom: vi.fn(),
  updateRoom: vi.fn(),
  removeRoom: vi.fn(),
  bulkCreate: vi.fn(),
  openRound: vi.fn(),
  listPlans: vi.fn(),
  listAccounts: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    cafeRooms: {
      list: listRooms,
      create: createRoom,
      update: updateRoom,
      remove: removeRoom,
      bulkCreate,
      openRound,
    },
    groupBuy: { listPlans },
    accounts: { list: listAccounts },
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
    account_id: 41,
    zone_key: 'openai',
    theme_key: 'warm_wood',
    scene_slot_key: 'openai-7',
    status,
    featured: true,
    sort_order: 7,
    plan: {
      id: 21,
      title: 'OpenAI Room 5 seats',
      target_group_id: 5,
      fulfillment_mode: 'room_subscription',
      total_shares: 5,
      seat_count: 5,
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
      { id: 21, title: 'OpenAI Room 5 seats', target_group_id: 5, fulfillment_mode: 'room_subscription' },
    ] })
    listAccounts.mockReset().mockResolvedValue({ items: [
      { id: 41, name: 'OpenAI account', platform: 'openai', status: 'active', concurrency: 3 },
      { id: 42, name: 'Second account', platform: 'openai', status: 'active', concurrency: 2 },
    ], total: 2, page: 1, page_size: 200, pages: 1 })
    createRoom.mockReset().mockResolvedValue({ data: room() })
    updateRoom.mockReset().mockResolvedValue({ data: room() })
    removeRoom.mockReset().mockResolvedValue({ data: { message: 'ok' } })
    openRound.mockReset().mockResolvedValue({ data: { id: 81, status: 'open' } })
    bulkCreate.mockReset().mockResolvedValue({ data: {
      created: [{ account_id: 41, room: room() }],
      failed: [{ account_id: 42, error_code: 'CAFE_ACCOUNT_ALREADY_ASSIGNED', message: 'assigned' }],
    } })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads rooms and exposes only room_subscription plans in the create form', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listRooms).toHaveBeenCalledWith(expect.objectContaining({ page: 1, page_size: 20, sort_by: 'sort_order' }))
    expect(listAccounts).toHaveBeenCalledWith(1, 200, { status: 'active', lite: 'true' })
    expect(wrapper.text()).toContain('OpenAI 七号房')

    const createButton = wrapper.findAll('button').find((button) => button.text().includes('新建房间'))
    await createButton?.trigger('click')
    expect(wrapper.text()).toContain('OpenAI Room 5 seats')
    expect(wrapper.text()).not.toContain('Legacy plan')
  })

  it('renders room management and embedded plan management together', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('OpenAI 七号房')
    expect(wrapper.find('[data-testid="embedded-group-buy"]').attributes('data-embedded')).toBe('true')
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
    await selects[1].setValue('41')
    await form.trigger('submit')
    await flushPromises()

    expect(createRoom).toHaveBeenCalledWith(expect.objectContaining({
      code: 'ROOM-008',
      name: '八号房',
      plan_id: 21,
      account_id: 41,
    }))
    const payload = createRoom.mock.calls[0][0]
    expect(payload).not.toHaveProperty('price')
    expect(payload).not.toHaveProperty('group_id')

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

  it('sends selected account IDs and renders per-item bulk failures', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('批量创建'))?.trigger('click')
    const form = wrapper.find('#cafe-room-bulk-form')
    const selects = form.findAll('select')
    await selects[0].setValue('21')
    await selects[1].setValue(['41', '42'])
    await form.trigger('submit')
    await flushPromises()

    expect(bulkCreate).toHaveBeenCalledWith(expect.objectContaining({ plan_id: 21, account_ids: [41, 42] }))
    expect(wrapper.text()).toContain('CAFE_ACCOUNT_ALREADY_ASSIGNED')
    expect(showSuccess).toHaveBeenCalled()
  })

  it('loads all active accounts across paginated dependency responses', async () => {
    listAccounts.mockReset()
      .mockResolvedValueOnce({
        items: [{ id: 41, name: 'First account', platform: 'openai', status: 'active', concurrency: 3 }],
        total: 201,
        page: 1,
        page_size: 200,
        pages: 2,
      })
      .mockResolvedValueOnce({
        items: [{ id: 241, name: 'Later account', platform: 'openai', status: 'active', concurrency: 2 }],
        total: 201,
        page: 2,
        page_size: 200,
        pages: 2,
      })

    const wrapper = mountView()
    await flushPromises()
    await wrapper.findAll('button').find((button) => button.text().includes('新建房间'))?.trigger('click')

    expect(listAccounts).toHaveBeenNthCalledWith(2, 2, 200, { status: 'active', lite: 'true' })
    expect(wrapper.find('#cafe-room-form').text()).toContain('Later account')
  })
})
