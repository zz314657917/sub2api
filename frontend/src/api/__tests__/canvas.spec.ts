import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, put, del } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    put,
    delete: del,
  },
}))

import {
  cancelCanvasRun,
  createCanvas,
  createCanvasRun,
  getCanvas,
  getCanvasRun,
  listCanvasModels,
  listCanvasRuns,
  listCanvases,
  updateCanvas,
} from '@/api/canvas'

describe('canvas api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    del.mockReset()
  })

  it('maps backend canvas summaries and documents into the view model', async () => {
    get
      .mockResolvedValueOnce({
        data: {
          items: [
            {
              id: 12,
              title: 'Campaign',
              description: 'Launch images',
              node_count: 2,
              created_at: '2026-05-20T00:00:00Z',
              updated_at: '2026-05-20T00:10:00Z',
            },
          ],
          total: 1,
          limit: 20,
          offset: 0,
        },
      })
      .mockResolvedValueOnce({
        data: {
          item: {
            id: 12,
            title: 'Campaign',
            description: 'Launch images',
            metadata: { model: 'gpt-image-2' },
            viewport: { x: 0, y: 0, zoom: 1 },
            created_at: '2026-05-20T00:00:00Z',
            updated_at: '2026-05-20T00:10:00Z',
            nodes: [
              {
                id: 'node_prompt',
                type: 'prompt',
                position: { x: 80, y: 90, width: 170, height: 86 },
                data: {
                  title: 'Prompt',
                  status: 'queued',
                  config: { prompt: 'draw' },
                  result: { summary: 'done' },
                  error: { message: 'provider failed' },
                },
              },
            ],
            edges: [
              {
                id: 'edge_1',
                source: 'node_prompt',
                target: 'node_result',
              },
            ],
          },
        },
      })

    await expect(listCanvases({ limit: 20, offset: 0 })).resolves.toMatchObject({
      items: [{ id: '12', name: 'Campaign', node_count: 2 }],
      total: 1,
    })
    await expect(getCanvas('12')).resolves.toMatchObject({
      id: '12',
      name: 'Campaign',
      model: 'gpt-image-2',
      document: {
        nodes: [
          {
            id: 'node_prompt',
            type: 'prompt',
            title: 'Prompt',
            x: 80,
            y: 90,
            status: 'queued',
            config: { prompt: 'draw' },
            result: { summary: 'done' },
            error: { message: 'provider failed' },
          },
        ],
        edges: [
          {
            id: 'edge_1',
            source_node_id: 'node_prompt',
            target_node_id: 'node_result',
          },
        ],
      },
    })

    expect(get).toHaveBeenNthCalledWith(1, '/user/canvases', {
      params: { limit: 20, offset: 0 },
    })
    expect(get).toHaveBeenNthCalledWith(2, '/user/canvases/12')
  })

  it('maps view model writes into backend canvas payloads', async () => {
    post.mockResolvedValueOnce({
      data: {
        item: {
          id: 20,
          title: 'Fresh Canvas',
          description: 'Draft',
          metadata: { model: 'gpt-image-2' },
          viewport: { x: 0, y: 0, zoom: 1 },
          nodes: [],
          edges: [],
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:00:00Z',
        },
      },
    })
    put.mockResolvedValueOnce({
      data: {
        item: {
          id: 20,
          title: 'Fresh Canvas',
          metadata: { model: 'gpt-image-2' },
          nodes: [],
          edges: [],
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:10:00Z',
        },
      },
    })

    const payload = {
      name: 'Fresh Canvas',
      description: 'Draft',
      model: 'gpt-image-2',
      document: {
        nodes: [
          {
            id: 'node_prompt',
            type: 'prompt' as const,
            title: 'Prompt',
            x: 80,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle' as const,
            config: { prompt: 'draw' },
          },
        ],
        edges: [
          {
            id: 'edge_1',
            source_node_id: 'node_prompt',
            target_node_id: 'node_result',
          },
        ],
        viewport: { x: 0, y: 0, zoom: 1 },
      },
    }

    await createCanvas(payload)
    await updateCanvas('20', payload)

    const expectedPayload = expect.objectContaining({
      title: 'Fresh Canvas',
      description: 'Draft',
      metadata: expect.objectContaining({ model: 'gpt-image-2' }),
      nodes: [
        expect.objectContaining({
          id: 'node_prompt',
          type: 'prompt',
          position: expect.objectContaining({ x: 80, y: 90 }),
          data: expect.objectContaining({ title: 'Prompt', config: { prompt: 'draw' } }),
        }),
      ],
      edges: [
        expect.objectContaining({
          id: 'edge_1',
          source: 'node_prompt',
          target: 'node_result',
        }),
      ],
    })
    expect(post).toHaveBeenCalledWith('/user/canvases', expectedPayload)
    expect(put).toHaveBeenCalledWith('/user/canvases/20', expectedPayload)
  })

  it('maps canvas run and model endpoints', async () => {
    get
      .mockResolvedValueOnce({
        data: {
          items: [
            {
              id: 30,
              canvas_id: 20,
              api_key_id: 44,
              status: 'pending',
              model: 'gpt-image-2',
              input: { nodes: 1 },
              output: { node_result: { summary: 'created' } },
              outputs: { node_result: { summary: 'created' } },
              result_node_ids: ['node_result'],
              canceled_at: '2026-05-20T00:03:00Z',
              created_at: '2026-05-20T00:00:00Z',
              updated_at: '2026-05-20T00:00:00Z',
            },
          ],
          total: 1,
        },
      })
      .mockResolvedValueOnce({
        data: {
          item: {
            id: 30,
            canvas_id: 20,
            api_key_id: 44,
            status: 'pending',
            output: { node_result: { summary: 'created' } },
            created_at: '2026-05-20T00:00:00Z',
            updated_at: '2026-05-20T00:00:00Z',
          },
        },
      })
      .mockResolvedValueOnce({
        data: {
          items: [
            {
              id: 'gpt-image-2',
              display_name: 'gpt-image-2',
              capabilities: ['image'],
            },
          ],
        },
      })
    post.mockResolvedValueOnce({
      data: {
        item: {
          id: 31,
          canvas_id: 20,
          api_key_id: 44,
          status: 'pending',
          model: 'gpt-image-2',
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:00:00Z',
        },
      },
    })

    await expect(listCanvasRuns({ canvas_id: '20', limit: 8 })).resolves.toMatchObject({
      items: [{
        id: '30',
        canvas_id: '20',
        api_key_id: 44,
        status: 'queued',
        output: { node_result: { summary: 'created' } },
        outputs: { node_result: { summary: 'created' } },
        canceled_at: '2026-05-20T00:03:00Z',
      }],
    })
    await expect(getCanvasRun('30')).resolves.toMatchObject({
      id: '30',
      api_key_id: 44,
      status: 'queued',
      output: { node_result: { summary: 'created' } },
    })
    await expect(listCanvasModels()).resolves.toMatchObject({
      items: [
        {
          id: 'gpt-image-2',
          capabilities: ['text_to_image', 'image_to_image'],
          supports_image_input: true,
          supports_image_output: true,
        },
      ],
    })
    await expect(createCanvasRun({ canvas_id: '20', api_key_id: 44, model: 'gpt-image-2' })).resolves.toMatchObject({
      id: '31',
      canvas_id: '20',
      api_key_id: 44,
      status: 'queued',
    })

    expect(post).toHaveBeenCalledWith('/user/canvas-runs', {
      canvas_id: 20,
      api_key_id: 44,
      model: 'gpt-image-2',
    })
  })

  it('cancels canvas runs and maps canceled timestamps', async () => {
    post.mockResolvedValueOnce({
      data: {
        item: {
          id: 30,
          canvas_id: 20,
          api_key_id: 44,
          status: 'canceled',
          canceled_at: '2026-05-20T00:03:00Z',
          completed_at: '2026-05-20T00:03:00Z',
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:03:00Z',
        },
      },
    })

    await expect(cancelCanvasRun('30')).resolves.toMatchObject({
      id: '30',
      canvas_id: '20',
      status: 'canceled',
      canceled_at: '2026-05-20T00:03:00Z',
      completed_at: '2026-05-20T00:03:00Z',
    })

    expect(post).toHaveBeenCalledWith('/user/canvas-runs/30/cancel')
  })
})
