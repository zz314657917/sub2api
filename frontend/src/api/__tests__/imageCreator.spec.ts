import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
  },
}))

import { createImageTask, getImageTask, listImageTasks } from '@/api/imageCreator'

describe('imageCreator api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    post.mockResolvedValue({ data: { id: 123, status: 'pending' } })
    get.mockResolvedValue({ data: { tasks: [], images: [] } })
  })

  it('creates a server-side image task with api_key_id instead of a plaintext key', async () => {
    await createImageTask({
      apiKeyId: 10,
      model: 'gpt-image-2',
      prompt: 'draw cats',
      size: '1024x1024',
      quality: 'auto',
      count: 4,
      outputFormat: 'png',
      background: 'auto',
    })

    expect(post).toHaveBeenCalledTimes(1)
    const [url, payload] = post.mock.calls[0]
    expect(url).toBe('/user/image-creator/tasks')
    expect(payload).toMatchObject({
      api_key_id: 10,
      model: 'gpt-image-2',
      prompt: 'draw cats',
      count: 4,
      output_format: 'png',
    })
    expect(payload).not.toHaveProperty('api_key')
    expect(payload).not.toHaveProperty('apiKey')
  })

  it('loads recent tasks and polls a task by id', async () => {
    await listImageTasks()
    await getImageTask(123)

    expect(get).toHaveBeenNthCalledWith(1, '/user/image-creator/tasks')
    expect(get).toHaveBeenNthCalledWith(2, '/user/image-creator/tasks/123')
  })
})
