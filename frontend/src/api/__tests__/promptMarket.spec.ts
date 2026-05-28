import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, del } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  del: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    delete: del,
  },
}))

import {
  createPromptFavorite,
  deletePromptFavorite,
  fetchPromptFavorites,
  fetchPromptMarketPrompts,
} from '@/api/promptMarket'

describe('promptMarket api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    del.mockReset()
    vi.restoreAllMocks()
  })

  it('parses banana and awesome prompt market sources', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async (url: RequestInfo | URL) => {
      const textUrl = String(url)
      if (textUrl.includes('banana-prompt-quicker')) {
        return new Response(JSON.stringify([
          {
            title: 'Banana poster',
            preview: 'https://example.com/banana.png',
            reference_image_urls: ['https://example.com/ref.png'],
            prompt: 'draw a banana poster',
            author: 'maker',
            category: 'Poster',
          },
        ]), { status: 200 })
      }
      return new Response(`## 海报

### Case 1: [Neon city](https://example.com/case) (by [artist](https://example.com/artist))

<img src="./images/neon.png" />

**提示词：**
\`\`\`
draw neon city
\`\`\`
`, { status: 200 })
    })

    const prompts = await fetchPromptMarketPrompts()

    expect(fetchMock).toHaveBeenCalledTimes(3)
    expect(prompts).toHaveLength(2)
    expect(prompts[0]).toMatchObject({
      source: 'banana-prompt-quicker',
      title: 'Banana poster',
      prompt: 'draw a banana poster',
    })
    expect(prompts[1]).toMatchObject({
      source: 'awesome-gpt-image-2-prompts',
      title: 'Neon city',
      prompt: 'draw neon city',
    })
    expect(prompts[1].preview).toContain('raw.githubusercontent.com/EvoLinkAI/awesome-gpt-image-2-prompts')
  })

  it('uses sub2api user prompt favorite endpoints', async () => {
    get.mockResolvedValueOnce({ data: { items: [] } })
    post.mockResolvedValueOnce({ data: { item: { id: 1 }, items: [] } })
    del.mockResolvedValueOnce({ data: { items: [] } })

    await fetchPromptFavorites()
    await createPromptFavorite({
      id: 'p1',
      source: 'banana-prompt-quicker',
      title: 'Poster',
      preview: 'https://example.com/p.png',
      referenceImageUrls: [],
      prompt: 'draw poster',
      author: 'maker',
      mode: 'generate',
      category: 'Poster',
      sourceLabel: 'banana-prompt-quicker',
      isNsfw: false,
    })
    await deletePromptFavorite(1)

    expect(get).toHaveBeenCalledWith('/user/prompt-favorites', {
      signal: undefined,
    })
    expect(post).toHaveBeenCalledWith('/user/prompt-favorites', expect.objectContaining({
      prompt_id: 'p1',
      source: 'banana-prompt-quicker',
      prompt: 'draw poster',
    }))
    expect(del).toHaveBeenCalledWith('/user/prompt-favorites/1')
  })
})
