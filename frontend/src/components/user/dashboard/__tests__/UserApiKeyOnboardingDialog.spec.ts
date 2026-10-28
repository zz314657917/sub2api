import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import UserApiKeyOnboardingDialog from '../UserApiKeyOnboardingDialog.vue'

const mocks = vi.hoisted(() => ({
  list: vi.fn(),
  copy: vi.fn(),
  support: vi.fn(),
}))
vi.mock('@/api/keys', () => ({ keysAPI: { list: mocks.list } }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: mocks.copy }) }))
vi.mock('@/utils/supportPopup', () => ({ openSupportPopup: mocks.support }))
vi.mock('vue-i18n', async () => {
  const { default: dashboard } = await import('@/i18n/locales/zh/dashboard')
  return {
    useI18n: () => ({
      t: (key: string, params: Record<string, unknown> = {}) => {
        const message = key.split('.').slice(1).reduce<unknown>((value, part) =>
          value && typeof value === 'object' ? (value as Record<string, unknown>)[part] : undefined, dashboard)
        return typeof message === 'string'
          ? message.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name] ?? ''))
          : key
      },
    }),
  }
})

const rawKey = 'sk-test-onboarding-not-a-real-credential-0123456789'
const makeKey = (overrides = {}) => ({
  id: 1, is_default: true, key: rawKey, status: 'active', expires_at: null, quota: 0, quota_used: 0,
  ...overrides,
})
const wrappers: ReturnType<typeof mount>[] = []
function render(props = {}) {
  const wrapper = mount(UserApiKeyOnboardingDialog, {
    props: { show: true, hasApiKey: true, ...props },
    global: {
      stubs: { Teleport: true, Transition: false },
    },
  })
  wrappers.push(wrapper)
  return wrapper
}
beforeEach(() => {
  vi.clearAllMocks()
  mocks.list.mockResolvedValue({ items: [makeKey()], total: 1 })
  mocks.copy.mockResolvedValue(true)
})
afterEach(() => { wrappers.splice(0).forEach(wrapper => wrapper.unmount()) })

describe('UserApiKeyOnboardingDialog quick setup', () => {
  it('links the executable without credentials and only displays a masked key', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.text()).toContain('两步完成 ChatGPT 配置')
    expect(wrapper.text()).toContain('下载配置助手')
    expect(wrapper.text()).toContain('复制你的密钥')
    expect(wrapper.text()).not.toContain(rawKey)
    expect(wrapper.get('code').text()).toBe('sk-tes...6789')
    expect(wrapper.get('[data-testid="setup-download"]').attributes('href')).toBe('/downloads/chatgpt-config-windows.exe')
    expect(mocks.list).toHaveBeenCalledWith(1, 1, { sort_by: 'id', sort_order: 'asc' }, { signal: expect.any(AbortSignal) })
    expect(mocks.copy).not.toHaveBeenCalled()
  })

  it('copies the full key on click, shows the next action and keeps the dialog open', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="setup-copy"]').trigger('click')
    await flushPromises()
    expect(mocks.copy).toHaveBeenCalledWith(rawKey, '已复制密钥')
    expect(wrapper.text()).toContain('已复制密钥')
    expect(wrapper.text()).toContain('打开助手 → 粘贴密钥 → 测试并保存。')
    expect(wrapper.emitted('skip')).toBeUndefined()
    expect(wrapper.text()).not.toContain(rawKey)
  })

  it('does not claim success when copying fails', async () => {
    mocks.copy.mockResolvedValue(false)
    const wrapper = render()
    await flushPromises()
    await wrapper.get('[data-testid="setup-copy"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-testid="setup-copy"]').text()).toBe('复制默认密钥')
  })

  it.each([
    [{ status: 'inactive' }, '默认密钥已停用'],
    [{ status: 'expired' }, '默认密钥已过期'],
    [{ expires_at: '2000-01-01T00:00:00Z' }, '默认密钥已过期'],
    [{ status: 'quota_exhausted' }, '默认密钥额度已用尽'],
    [{ quota: 1, quota_used: 1 }, '默认密钥额度已用尽'],
  ])('does not offer copying for unavailable keys: %j', async (overrides, expected) => {
    mocks.list.mockResolvedValue({ items: [makeKey(overrides)], total: 1 })
    const wrapper = render()
    await flushPromises()
    expect(wrapper.text()).toContain(expected)
    expect(wrapper.find('[data-testid="setup-copy"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('前往密钥管理')
    expect(mocks.copy).not.toHaveBeenCalled()
  })

  it('never substitutes another key when the default flag is absent', async () => {
    mocks.list.mockResolvedValue({ items: [makeKey({ is_default: false })], total: 1 })
    const wrapper = render()
    await flushPromises()
    expect(wrapper.find('[data-testid="setup-copy"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('还没有可复制的默认密钥')
    expect(mocks.copy).not.toHaveBeenCalled()
  })

  it('offers key creation when no key exists', async () => {
    mocks.list.mockResolvedValue({ items: [], total: 0 })
    const wrapper = render({ hasApiKey: false })
    await flushPromises()
    await wrapper.findAll('button').find(button => button.text() === '创建 API 密钥')!.trigger('click')
    expect(wrapper.emitted('create')).toHaveLength(1)
    expect(wrapper.emitted('skip')).toBeUndefined()
  })

  it('shows loading and retries a failed request without exposing server errors', async () => {
    let reject!: (reason: Error) => void
    mocks.list.mockReturnValueOnce(new Promise((_, rejectRequest) => { reject = rejectRequest }))
    const wrapper = render()
    expect(wrapper.text()).toContain('正在读取密钥')
    reject(new Error(rawKey))
    await flushPromises()
    expect(wrapper.text()).not.toContain(rawKey)
    expect(wrapper.text()).toContain('暂时无法读取密钥')
    await wrapper.findAll('button').find(button => button.text() === '重新读取密钥')!.trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="setup-copy"]').exists()).toBe(true)
  })

  it('ignores stale responses after closing and reopens with a fresh request', async () => {
    let resolve!: (value: unknown) => void
    mocks.list.mockReturnValueOnce(new Promise(resolveRequest => { resolve = resolveRequest }))
    const wrapper = render()
    const signal = mocks.list.mock.calls[0][3].signal as AbortSignal
    await wrapper.setProps({ show: false })
    expect(signal.aborted).toBe(true)
    await wrapper.setProps({ show: true })
    await flushPromises()
    resolve({ items: [makeKey({ key: 'sk-stale-secret', status: 'inactive' })] })
    await flushPromises()
    expect(wrapper.find('[data-testid="setup-copy"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('sk-stale-secret')
    expect(mocks.list).toHaveBeenCalledTimes(2)
  })

  it('does not fetch a key for a closed dialog', async () => {
    render({ show: false })
    await flushPromises()
    expect(mocks.list).not.toHaveBeenCalled()
  })

  it.each([
    ['reward', '完成首次真实调用后，可到福利中心领取 1 免费积分。'],
    ['trial', '首次调用时自动启用 API 试用积分'],
    ['wallet', '账户积分已到账'],
  ] as const)('preserves benefit guidance for %s', async (benefitKind, expected) => {
    const wrapper = render({ hasBenefit: true, benefitKind, benefitLabel: '1 免费积分' })
    await flushPromises()
    expect(wrapper.text()).toContain(expected)
    expect(wrapper.text()).not.toContain('GPT-5.5')
  })

  it('retains tutorial, support and skip actions', async () => {
    const wrapper = render()
    await flushPromises()
    const button = (text: string) => wrapper.findAll('button').find(item => item.text().startsWith(text))!
    await button('其他系统 / 手动配置').trigger('click')
    await button('联系客服').trigger('click')
    await button('暂时跳过').trigger('click')
    expect(wrapper.emitted('tutorial')).toHaveLength(1)
    expect(mocks.support).toHaveBeenCalledOnce()
    expect(wrapper.emitted('skip')).toHaveLength(1)
  })
})
