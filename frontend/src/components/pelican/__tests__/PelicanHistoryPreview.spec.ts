import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PelicanHistoryPreview from '../PelicanHistoryPreview.vue'
import PelicanSandboxFrame from '../PelicanSandboxFrame.vue'

const api = vi.hoisted(() => ({ result: vi.fn() }))
vi.mock('@/api/pelicanTests', () => ({ getPelicanResult: api.result }))
describe('history animated previews', () => {
  beforeEach(() => vi.clearAllMocks())
  it('loads a sandboxed preview, replays and emits enlargement', async () => {
    api.result.mockResolvedValue({ status: 'success', html: '<html><body>animation</body></html>' })
    const wrapper = mount(PelicanHistoryPreview, { props: { resultId: 12 } })
    await flushPromises()
    expect(api.result).toHaveBeenCalledWith(12, expect.any(AbortSignal))
    expect(wrapper.get('iframe').attributes('sandbox')).toBe('')
    await wrapper.get('.preview').trigger('click')
    expect(wrapper.emitted('enlarge')).toHaveLength(1)
    await wrapper.findAll('button').find(button => button.text() === '重新播放')!.trigger('click')
    expect(wrapper.getComponent(PelicanSandboxFrame).props('replayKey')).toBe(1)
    wrapper.unmount()
  })
  it('allows a failed preview request to be retried and cancels on close', async () => {
    api.result.mockRejectedValueOnce(new Error('unavailable')).mockResolvedValueOnce({ status: 'success', html: '<html>ok</html>' })
    const wrapper = mount(PelicanHistoryPreview, { props: { resultId: 13 } })
    await flushPromises()
    expect(wrapper.text()).toContain('作品加载失败')
    await wrapper.get('button').trigger('click'); await flushPromises()
    expect(wrapper.find('iframe').exists()).toBe(true)
    const signal = api.result.mock.calls[1][1] as AbortSignal
    wrapper.unmount()
    expect(signal.aborted).toBe(true)
  })
})
