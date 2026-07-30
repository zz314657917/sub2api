import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultQuickstartTutorialConfig } from '@/views/public/tutorialQuickstart'

const { get, post, put } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: { get, post, put },
}))

import {
  getQuickstartConfig,
  resetQuickstartConfig,
  updateQuickstartConfig,
} from '@/api/admin/tutorials'

describe('admin tutorial API', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
  })

  it('uses the dedicated quick-start configuration endpoints', async () => {
    get.mockResolvedValueOnce({ data: defaultQuickstartTutorialConfig })
    put.mockResolvedValueOnce({ data: defaultQuickstartTutorialConfig })
    post.mockResolvedValueOnce({ data: defaultQuickstartTutorialConfig })

    await getQuickstartConfig()
    await updateQuickstartConfig(defaultQuickstartTutorialConfig)
    await resetQuickstartConfig()

    expect(get).toHaveBeenCalledWith('/admin/tutorials/quickstart-config')
    expect(put).toHaveBeenCalledWith('/admin/tutorials/quickstart-config', defaultQuickstartTutorialConfig)
    expect(post).toHaveBeenCalledWith('/admin/tutorials/quickstart-config/reset')
  })
})
