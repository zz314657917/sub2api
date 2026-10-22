import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, shallowMount } from '@vue/test-utils'
import ChannelStatusView from '../ChannelStatusView.vue'
import MonitorCardGrid from '@/components/user/monitor/MonitorCardGrid.vue'

const mocks = vi.hoisted(() => ({ list: vi.fn(), status: vi.fn(), capacity: vi.fn(), showError: vi.fn() }))
vi.mock('@/api/channelMonitor', () => ({ list: mocks.list, status: mocks.status }))
vi.mock('@/api/user', () => ({ getAccountCapacityPools: mocks.capacity }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: { channel_monitor_enabled: true }, showError: mocks.showError }),
}))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({ t: (key: string) => key }),
}))
enableAutoUnmount(afterEach)
afterEach(() => {
  vi.restoreAllMocks()
  vi.useRealTimers()
  localStorage.clear()
})
const item = { id: 1, name: 'Monitor', primary_model: 'model', primary_status: 'operational', availability_7d: 99, extra_models: [] }
const detail = (value: number) => ({
  id: 1, models: [{ model: 'model', availability_15d: value, availability_30d: value }],
})
beforeEach(async () => {
  vi.useFakeTimers()
  vi.resetAllMocks()
  vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)
  localStorage.clear()
  localStorage.setItem('channel-status-auto-refresh', JSON.stringify({ enabled: true, interval_seconds: 30 }))
  // jsdom dispatches storage events through a zero-delay timeout.
  await vi.advanceTimersByTimeAsync(0)
  mocks.list.mockImplementation(async () => ({ items: [{ ...item }] }))
  mocks.status.mockResolvedValue(detail(90))
  mocks.capacity.mockResolvedValue({ mine: null, shared: null, external: null })
})
function mountView() {
  return shallowMount(ChannelStatusView, {
    global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      MonitorHero: {
        props: ['loading'], emits: ['update:window', 'refresh'],
        template: `<div :data-loading="loading">
          <button class="fifteen" @click="$emit('update:window', '15d')" />
          <button class="thirty" @click="$emit('update:window', '30d')" />
          <button class="seven" @click="$emit('update:window', '7d')" />
          <button class="refresh" @click="$emit('refresh')" />
        </div>`,
      },
      MonitorCardGrid: false,
      MonitorCard: { props: ['availabilityValue'], template: '<output>{{ availabilityValue }}</output>' },
    } },
  })
}

function deferred() {
  let resolve!: (value: unknown) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

function deferCapacity(completion: string) {
  const pending = deferred()
  let signal!: AbortSignal
  mocks.capacity.mockImplementationOnce(({ signal: current }: { signal: AbortSignal }) => {
    signal = current
    if (completion === 'abort-rejection') {
      signal.addEventListener('abort', () => pending.reject(new DOMException('Aborted', 'AbortError')), { once: true })
    }
    return pending.promise
  })
  return { resolve: pending.resolve, get signal() { return signal } }
}

async function openDetails() {
  const wrapper = mountView()
  await flushPromises()
  await wrapper.get('.fifteen').trigger('click')
  await flushPromises()
  expect(wrapper.get('output').text()).toBe('90')
  return wrapper
}

describe('ChannelStatusView local detail refresh', () => {
  it.each(['fifteen', 'thirty'])('refreshes cached %s-day availability automatically and manually', async (tab) => {
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('.' + tab).trigger('click')
    await flushPromises()
    expect(wrapper.get('output').text()).toBe('90')
    mocks.status.mockResolvedValue(detail(98))
    await vi.advanceTimersByTimeAsync(61_000)
    await flushPromises()
    expect(wrapper.get('output').text()).toBe('98')
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(mocks.capacity).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-loading]').attributes('data-loading')).toBe('false')
    mocks.status.mockResolvedValue(detail(95))
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(wrapper.get('output').text()).toBe('95')
    expect(mocks.status).toHaveBeenCalledTimes(3)
    expect(mocks.capacity).toHaveBeenCalledTimes(3)
    await wrapper.get('.seven').trigger('click')
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(3)
  })

  it('keeps seven-day refresh on the list endpoint with no extra details', async () => {
    const wrapper = mountView()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(61_000)
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(3)
    expect(mocks.capacity).toHaveBeenCalledTimes(3)
    expect(mocks.status).not.toHaveBeenCalled()
    expect(wrapper.get('output').text()).toBe('99')
  })

  it('retains existing seven-day extra-model details without additional refresh requests', async () => {
    mocks.list.mockImplementation(async () => ({ items: [{ ...item, extra_models: [{ model: 'extra' }] }] }))
    const wrapper = mountView()
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(61_000)
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(3)
    expect(mocks.status).toHaveBeenCalledTimes(1)
  })

  it.each(['automatic', 'manual'])('retains the first pending seven-day extra-model detail during %s refresh without replacement', async (refresh) => {
    const pending = deferred()
    const extraDetail = { id: 1, models: [
      ...detail(93).models,
      { model: 'extra', availability_7d: 97, availability_15d: 96, availability_30d: 95 },
    ] }
    mocks.list.mockImplementation(async () => ({ items: [{ ...item, extra_models: [{ model: 'extra' }] }] }))
    // A duplicate request must not hide the loss of the original response.
    mocks.status.mockReturnValue(new Promise(() => {})).mockReturnValueOnce(pending.promise)
    const wrapper = mountView()
    await flushPromises()
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    expect(cache[1]).toBeUndefined()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    if (refresh === 'automatic') await vi.advanceTimersByTimeAsync(61_000)
    else await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(mocks.capacity).toHaveBeenCalledTimes(2)
    pending.resolve(extraDetail)
    await flushPromises()
    expect(cache[1]).toEqual(extraDetail)
    expect(cache[1].models.find((model: { model: string }) => model.model === 'extra')?.availability_7d).toBe(97)
    expect(mocks.status).toHaveBeenCalledTimes(1)
    await wrapper.get('.fifteen').trigger('click')
    await flushPromises()
    expect(wrapper.get('output').text()).toBe('93')
    expect(mocks.status).toHaveBeenCalledTimes(1)
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it.each(['list', 'capacity'])('pauses while the initial %s load is pending', async (pending) => {
    let finish!: (value: unknown) => void
    const response = new Promise(resolve => { finish = resolve })
    if (pending === 'list') mocks.list.mockReturnValueOnce(response)
    else mocks.capacity.mockReturnValueOnce(response)
    mountView()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(61_000)
    expect(mocks.list).toHaveBeenCalledTimes(1)
    expect(mocks.capacity).toHaveBeenCalledTimes(1)
    finish(pending === 'list' ? { items: [item] } : { mine: null, shared: null, external: null })
    await flushPromises()
    await vi.advanceTimersByTimeAsync(61_000)
    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(mocks.capacity).toHaveBeenCalledTimes(2)
  })

  it('pauses automatic reload while hidden and resumes after visibility returns', async () => {
    mountView()
    await flushPromises()
    vi.spyOn(document, 'hidden', 'get').mockReturnValue(true)
    await vi.advanceTimersByTimeAsync(60_000)
    expect(mocks.list).toHaveBeenCalledTimes(1)
    expect(mocks.capacity).toHaveBeenCalledTimes(1)
    vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)
    await vi.advanceTimersByTimeAsync(61_000)
    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(mocks.capacity).toHaveBeenCalledTimes(2)
  })

  it('aborts superseded list/capacity requests and prevents late refresh after unmount', async () => {
    const intervalStarted = vi.spyOn(globalThis, 'setInterval')
    const intervalStopped = vi.spyOn(globalThis, 'clearInterval')
    const listSignals: AbortSignal[] = []
    const capacitySignals: AbortSignal[] = []
    const finish: Array<(value: unknown) => void> = []
    mocks.list.mockImplementation(({ signal }: { signal: AbortSignal }) => {
      listSignals.push(signal)
      return new Promise(resolve => finish.push(resolve))
    })
    mocks.capacity.mockImplementation(({ signal }: { signal: AbortSignal }) => {
      capacitySignals.push(signal)
      return new Promise(resolve => finish.push(resolve))
    })
    const wrapper = mountView()
    await wrapper.get('.refresh').trigger('click')
    expect(listSignals[0].aborted).toBe(true)
    expect(capacitySignals[0].aborted).toBe(true)
    wrapper.unmount()
    expect(intervalStopped).toHaveBeenCalledWith(intervalStarted.mock.results[0].value)
    expect(listSignals[1].aborted).toBe(true)
    expect(capacitySignals[1].aborted).toBe(true)
    finish.forEach(resolve => resolve({ items: [item] }))
    await flushPromises()
    expect(mocks.status).not.toHaveBeenCalled()
    expect(mocks.showError).not.toHaveBeenCalled()
    vi.runAllTicks()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each(['late-response', 'abort-rejection'])('does not continue an automatic refresh after unmount with successful list and pending capacity: %s', async (completion) => {
    const wrapper = await openDetails()
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    const initialDetail = cache[1]
    const capacity = deferCapacity(completion)
    await vi.advanceTimersByTimeAsync(61_000)
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(mocks.capacity).toHaveBeenCalledTimes(2)
    expect(mocks.status).toHaveBeenCalledTimes(1)
    mocks.status.mockRejectedValueOnce(new Error('must not load after unmount'))
    wrapper.unmount()
    expect(capacity.signal.aborted).toBe(true)
    capacity.resolve({ mine: null, shared: null, external: null })
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    expect(cache[1]).toBe(initialDetail)
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each([
    ['tab', 'success'], ['tab', 'error'],
    ['refresh', 'success'], ['refresh', 'error'],
  ])('discards in-flight %s detail %s after unmount without cache/error effects', async (origin, completion) => {
    const wrapper = mountView()
    await flushPromises()
    if (origin === 'refresh') {
      await wrapper.get('.fifteen').trigger('click')
      await flushPromises()
    }
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    const initialDetail = cache[1]
    const pending = deferred()
    mocks.status.mockReturnValueOnce(pending.promise)
    await wrapper.get(origin === 'tab' ? '.fifteen' : '.refresh').trigger('click')
    await flushPromises()
    const expectedRequests = origin === 'tab' ? 1 : 2
    expect(mocks.status).toHaveBeenCalledTimes(expectedRequests)
    wrapper.unmount()
    if (completion === 'success') pending.resolve(detail(99))
    else pending.reject(new Error('stale detail failure'))
    await flushPromises()
    expect(cache[1]).toBe(initialDetail)
    expect(mocks.status).toHaveBeenCalledTimes(expectedRequests)
    expect(mocks.showError).not.toHaveBeenCalled()
    expect(vi.getTimerCount()).toBe(0)
  })

  it.each(['late-response', 'abort-rejection'])('does not continue a superseded refresh when its pending capacity completes: %s', async (completion) => {
    const wrapper = await openDetails()
    const capacity = deferCapacity(completion)
    await vi.advanceTimersByTimeAsync(61_000)
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(mocks.status).toHaveBeenCalledTimes(1)
    mocks.status.mockResolvedValue(detail(98))
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(capacity.signal.aborted).toBe(true)
    expect(mocks.list).toHaveBeenCalledTimes(3)
    expect(mocks.capacity).toHaveBeenCalledTimes(3)
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(wrapper.get('output').text()).toBe('98')
    capacity.resolve({ mine: null, shared: null, external: null })
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(wrapper.get('output').text()).toBe('98')
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it.each(['success', 'error'])('discards an older in-flight detail %s after a new refresh succeeds', async (completion) => {
    const wrapper = await openDetails()
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    const pending = deferred()
    mocks.status.mockReturnValueOnce(pending.promise).mockResolvedValue(detail(98))
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(2)
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(3)
    expect(wrapper.get('output').text()).toBe('98')
    const currentDetail = cache[1]
    if (completion === 'success') pending.resolve(detail(5))
    else pending.reject(new Error('superseded detail failure'))
    await flushPromises()
    expect(cache[1]).toBe(currentDetail)
    expect(wrapper.get('output').text()).toBe('98')
    expect(mocks.status).toHaveBeenCalledTimes(3)
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it.each(['success', 'error'])('discards an older ordinary detail %s after a forced detail for the same ID succeeds', async (completion) => {
    const pending = deferred()
    mocks.status.mockReturnValueOnce(pending.promise).mockResolvedValue(detail(98))
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('.fifteen').trigger('click')
    await flushPromises()
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    expect(cache[1]).toBeUndefined()
    expect(mocks.status).toHaveBeenCalledTimes(1)
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    const currentDetail = cache[1]
    expect(wrapper.get('output').text()).toBe('98')
    if (completion === 'success') pending.resolve(detail(5))
    else pending.reject(new Error('superseded ordinary detail failure'))
    await flushPromises()
    expect(cache[1]).toBe(currentDetail)
    expect(wrapper.get('output').text()).toBe('98')
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it.each(['success', 'error'])('discards a refresh-owned detail %s when superseded by a seven-day refresh without a replacement detail', async (completion) => {
    const wrapper = await openDetails()
    const cache = wrapper.findComponent(MonitorCardGrid).props('detailCache')
    const initialDetail = cache[1]
    const pending = deferred()
    mocks.status.mockReturnValueOnce(pending.promise)
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(2)
    await wrapper.get('.seven').trigger('click')
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.list).toHaveBeenCalledTimes(3)
    expect(mocks.capacity).toHaveBeenCalledTimes(3)
    if (completion === 'success') pending.resolve(detail(5))
    else pending.reject(new Error('superseded refresh detail failure'))
    await flushPromises()
    expect(cache[1]).toBe(initialDetail)
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(mocks.showError).not.toHaveBeenCalled()
  })

  it('still reports an active detail failure and preserves the previous cache', async () => {
    const wrapper = await openDetails()
    mocks.status.mockRejectedValueOnce(new Error('current detail failure'))
    await wrapper.get('.refresh').trigger('click')
    await flushPromises()
    expect(mocks.status).toHaveBeenCalledTimes(2)
    expect(wrapper.get('output').text()).toBe('90')
    expect(mocks.showError).toHaveBeenCalledTimes(1)
    expect(mocks.showError).toHaveBeenCalledWith('current detail failure')
    expect(wrapper.get('[data-loading]').attributes('data-loading')).toBe('false')
  })
})
