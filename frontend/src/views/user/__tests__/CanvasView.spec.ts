import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CanvasView from '../CanvasView.vue'

const listCanvases = vi.hoisted(() => vi.fn())
const getCanvas = vi.hoisted(() => vi.fn())
const createCanvas = vi.hoisted(() => vi.fn())
const updateCanvas = vi.hoisted(() => vi.fn())
const listCanvasRuns = vi.hoisted(() => vi.fn())
const createCanvasRun = vi.hoisted(() => vi.fn())
const listCanvasModels = vi.hoisted(() => vi.fn())
const keysList = vi.hoisted(() => vi.fn())
const getImageTask = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  }
})

vi.mock('@/api/canvas', () => ({
  listCanvases,
  getCanvas,
  createCanvas,
  updateCanvas,
  listCanvasRuns,
  createCanvasRun,
  listCanvasModels,
}))

vi.mock('@/api/keys', () => ({
  keysAPI: {
    list: keysList,
  },
}))

vi.mock('@/api/imageCreator', () => ({
  getImageTask,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

function makeCanvas(overrides: Record<string, unknown> = {}) {
  return {
    id: 'canvas_1',
    name: 'Campaign Canvas',
    description: 'Generate campaign images',
    model: 'gpt-image-2',
    node_count: 2,
    run_count: 1,
    created_at: '2026-05-20T00:00:00Z',
    updated_at: '2026-05-20T00:10:00Z',
    document: {
      nodes: [
        {
          id: 'node_prompt',
          type: 'prompt',
          title: 'Prompt',
          x: 80,
          y: 90,
          width: 170,
          height: 86,
          status: 'idle',
          config: { prompt: 'old prompt' },
        },
        {
          id: 'node_result',
          type: 'result',
          title: 'Result',
          x: 320,
          y: 90,
          width: 170,
          height: 86,
          status: 'done',
          config: {},
          result: {
            thumbnail_url: 'https://example.test/result.png',
            summary: 'rendered result',
          },
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
    ...overrides,
  }
}

function mountView() {
  return mount(CanvasView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: { props: ['name'], template: '<span />' },
      },
    },
  })
}

describe('CanvasView', () => {
  beforeEach(() => {
    listCanvases.mockReset().mockResolvedValue({
      items: [
        {
          id: 'canvas_1',
          name: 'Campaign Canvas',
          node_count: 2,
          run_count: 1,
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:10:00Z',
        },
      ],
      total: 1,
    })
    getCanvas.mockReset().mockResolvedValue(makeCanvas())
    createCanvas.mockReset().mockImplementation(async (payload) => ({
      ...makeCanvas({ id: 'canvas_created', created_at: '2026-05-20T00:00:00Z', updated_at: '2026-05-20T00:20:00Z' }),
      ...payload,
      node_count: payload.document.nodes.length,
    }))
    updateCanvas.mockReset().mockImplementation(async (id, payload) => ({
      ...makeCanvas({ id }),
      ...payload,
      node_count: payload.document.nodes.length,
      created_at: '2026-05-20T00:00:00Z',
      updated_at: '2026-05-20T00:30:00Z',
    }))
    listCanvasRuns.mockReset().mockResolvedValue({
      items: [
        {
          id: 'run_1',
          canvas_id: 'canvas_1',
          status: 'succeeded',
          api_key_id: 101,
          result_node_ids: ['node_result'],
          outputs: {
            node_result: {
              summary: 'latest output',
            },
          },
          created_at: '2026-05-20T00:11:00Z',
          updated_at: '2026-05-20T00:11:00Z',
        },
      ],
      total: 1,
    })
    createCanvasRun.mockReset().mockResolvedValue({
      id: 'run_2',
      canvas_id: 'canvas_1',
      status: 'queued',
      api_key_id: 101,
      created_at: '2026-05-20T00:12:00Z',
      updated_at: '2026-05-20T00:12:00Z',
    })
    listCanvasModels.mockReset().mockResolvedValue({
      items: [
        {
          id: 'gpt-image-2',
          name: 'gpt-image-2',
          provider: 'openai',
          capabilities: ['text_to_image', 'image_to_image'],
        },
      ],
    })
    getImageTask.mockReset()
    keysList.mockReset().mockResolvedValue({
      items: [
        {
          id: 101,
          user_id: 1,
          key: 'sk-redacted',
          name: 'Image Key',
          group_id: 7,
          multi_group_routes: [],
          account_pool_strategy: 'shared_only',
          status: 'active',
          ip_whitelist: [],
          ip_blacklist: [],
          last_used_at: null,
          quota: 0,
          quota_used: 0,
          expires_at: null,
          created_at: '2026-05-20T00:00:00Z',
          updated_at: '2026-05-20T00:00:00Z',
          group: {
            id: 7,
            name: 'OpenAI Images',
            description: null,
            platform: 'openai',
            rate_multiplier: 1,
            is_exclusive: false,
            status: 'active',
            subscription_type: 'standard',
            daily_limit_usd: null,
            weekly_limit_usd: null,
            monthly_limit_usd: null,
            allow_image_generation: true,
            image_rate_independent: false,
            image_rate_multiplier: 1,
            image_price_1k: null,
            image_price_2k: null,
            image_price_4k: null,
            claude_code_only: false,
            fallback_group_id: null,
            fallback_group_id_on_invalid_request: null,
            require_oauth_only: false,
            require_privacy_set: false,
            created_at: '2026-05-20T00:00:00Z',
            updated_at: '2026-05-20T00:00:00Z',
          },
          route_groups: [],
          rate_limit_5h: 0,
          rate_limit_1d: 0,
          rate_limit_7d: 0,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          window_5h_start: null,
          window_1d_start: null,
          window_7d_start: null,
          reset_5h_at: null,
          reset_1d_at: null,
          reset_7d_at: null,
        },
        {
          id: 102,
          name: 'Disabled Image Key',
          status: 'active',
          group_id: 8,
          group: {
            platform: 'openai',
            allow_image_generation: false,
          },
        },
      ],
    })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads canvases, opens the latest canvas, and renders all node type controls', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listCanvases).toHaveBeenCalledWith({ limit: 30, offset: 0 })
    expect(getCanvas).toHaveBeenCalledWith('canvas_1')
    expect(listCanvasRuns).toHaveBeenCalledWith({ canvas_id: 'canvas_1', limit: 8, offset: 0 })
    expect(listCanvasModels).toHaveBeenCalled()
    expect(keysList).toHaveBeenCalledWith(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    })
    expect(wrapper.find('[data-testid="canvas-view"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="canvas-node"]')).toHaveLength(2)
    expect(wrapper.find('[data-testid="canvas-node-preview-image"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="canvas-latest-run"]').text()).toContain('canvas.latestRun')

    for (const type of ['text', 'image', 'prompt', 'loop', 'group', 'text_to_image', 'image_to_image', 'result']) {
      expect(wrapper.find(`[data-testid="canvas-node-type-${type}"]`).exists()).toBe(true)
    }
  })

  it('creates a draft, adds a node, and saves it through the canvas API', async () => {
    listCanvases.mockResolvedValue({ items: [], total: 0 })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="canvas-new-button"]').trigger('click')
    await wrapper.find('[data-testid="canvas-name-input"]').setValue('Fresh Canvas')
    await wrapper.find('[data-testid="canvas-node-type-image_to_image"]').trigger('click')
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(createCanvas).toHaveBeenCalledWith(expect.objectContaining({
      name: 'Fresh Canvas',
      model: 'gpt-image-2',
      document: expect.objectContaining({
        nodes: expect.arrayContaining([
          expect.objectContaining({ type: 'image_to_image' }),
        ]),
      }),
    }))
    expect(showSuccess).toHaveBeenCalledWith('canvas.saveSuccess')
  })

  it('edits selected node config and saves it in the document payload', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="canvas-node-editor"]').exists()).toBe(true)
    await wrapper.find('[data-testid="canvas-node-title-input"]').setValue('Edited Prompt')
    await wrapper.find('[data-testid="canvas-node-config-prompt"]').setValue('draw a small robot')
    await wrapper.find('[data-testid="canvas-node-config-model"]').setValue('gpt-image-2')
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(updateCanvas).toHaveBeenCalledWith('canvas_1', expect.objectContaining({
      document: expect.objectContaining({
        nodes: expect.arrayContaining([
          expect.objectContaining({
            id: 'node_prompt',
            title: 'Edited Prompt',
            config: expect.objectContaining({
              prompt: 'draw a small robot',
              model: 'gpt-image-2',
            }),
          }),
        ]),
      }),
    }))
  })

  it('drags a node and saves the updated node coordinates', async () => {
    const wrapper = mountView()
    await flushPromises()

    const firstNode = wrapper.findAll('[data-testid="canvas-node"]')[0]
    await firstNode.trigger('mousedown', { button: 0, clientX: 100, clientY: 100 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 150, clientY: 130 }))
    window.dispatchEvent(new MouseEvent('mouseup'))
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(updateCanvas).toHaveBeenCalledWith('canvas_1', expect.objectContaining({
      document: expect.objectContaining({
        nodes: expect.arrayContaining([
          expect.objectContaining({
            id: 'node_prompt',
            x: 130,
            y: 120,
          }),
        ]),
      }),
    }))

    wrapper.unmount()
  })

  it('creates and deletes selected edges without duplicating the same edge', async () => {
    getCanvas.mockResolvedValue(makeCanvas({
      document: {
        nodes: [
          {
            id: 'node_prompt',
            type: 'prompt',
            title: 'Prompt',
            x: 80,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle',
            config: {},
          },
          {
            id: 'node_result',
            type: 'result',
            title: 'Result',
            x: 320,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle',
            config: {},
          },
        ],
        edges: [],
        viewport: { x: 0, y: 0, zoom: 1 },
      },
    }))
    const wrapper = mountView()
    await flushPromises()

    const nodes = wrapper.findAll('[data-testid="canvas-node"]')
    await nodes[0].trigger('click')
    await wrapper.find('[data-testid="canvas-create-edge-button"]').trigger('click')
    await nodes[1].trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-testid="canvas-edge"]')).toHaveLength(1)

    await nodes[0].trigger('click')
    await wrapper.find('[data-testid="canvas-create-edge-button"]').trigger('click')
    await nodes[1].trigger('click')
    await flushPromises()

    expect(wrapper.findAll('[data-testid="canvas-edge"]')).toHaveLength(1)

    await wrapper.find('[data-testid="canvas-edge"]').trigger('click')
    await wrapper.find('[data-testid="canvas-remove-edge-button"]').trigger('click')
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(updateCanvas).toHaveBeenCalledWith('canvas_1', expect.objectContaining({
      document: expect.objectContaining({
        edges: [],
      }),
    }))

    wrapper.unmount()
  })

  it('saves viewport after zooming and panning the canvas', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="canvas-zoom-in-button"]').trigger('click')
    await wrapper.find('[data-testid="canvas-stage"]').trigger('mousedown', { button: 0, clientX: 20, clientY: 30 })
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 70, clientY: 90 }))
    window.dispatchEvent(new MouseEvent('mouseup'))
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(updateCanvas).toHaveBeenCalledWith('canvas_1', expect.objectContaining({
      document: expect.objectContaining({
        viewport: {
          x: 50,
          y: 60,
          zoom: 1.1,
        },
      }),
    }))

    wrapper.unmount()
  })

  it('updates an existing canvas, saves before queueing, and refreshes runs', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('[data-testid="canvas-name-input"]').setValue('Updated Canvas')
    await wrapper.find('[data-testid="canvas-save-button"]').trigger('click')
    await flushPromises()

    expect(updateCanvas).toHaveBeenCalledWith('canvas_1', expect.objectContaining({
      name: 'Updated Canvas',
    }))

    await wrapper.find('[data-testid="canvas-run-button"]').trigger('click')
    await flushPromises()

    expect(createCanvasRun).toHaveBeenCalledWith({
      canvas_id: 'canvas_1',
      api_key_id: 101,
      model: 'gpt-image-2',
    })
    expect(updateCanvas).toHaveBeenCalledTimes(2)
    expect(listCanvasRuns).toHaveBeenLastCalledWith({ canvas_id: 'canvas_1', limit: 8, offset: 0 })
    expect(showSuccess).toHaveBeenCalledWith('canvas.runQueued')
  })

  it('polls canvas image tasks and renders the generated node image', async () => {
    getCanvas.mockResolvedValue(makeCanvas({
      document: {
        nodes: [
          {
            id: 'node_prompt',
            type: 'prompt',
            title: 'Prompt',
            x: 80,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle',
            config: { prompt: 'old prompt' },
          },
          {
            id: 'node_result',
            type: 'text_to_image',
            title: 'Text to image',
            x: 320,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle',
            config: {},
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
    }))
    listCanvasRuns.mockResolvedValue({
      items: [
        {
          id: 'run_1',
          canvas_id: 'canvas_1',
          status: 'running',
          api_key_id: 101,
          output: {
            mode: 'image_creator_tasks',
            image_tasks: [
              {
                node_id: 'node_result',
                task_id: 501,
                task_status: 'running',
              },
            ],
          },
          created_at: '2026-05-20T00:11:00Z',
          updated_at: '2026-05-20T00:11:00Z',
        },
      ],
      total: 1,
    })
    getImageTask.mockResolvedValue({
      id: 501,
      user_id: 1,
      api_key_id: 101,
      status: 'succeeded',
      model: 'gpt-image-2',
      prompt: 'paint a city',
      size: '1024x1024',
      quality: 'standard',
      output_format: 'png',
      background: 'auto',
      count: 1,
      expires_at: '2026-05-21T00:00:00Z',
      created_at: '2026-05-20T00:11:00Z',
      updated_at: '2026-05-20T00:12:00Z',
      images: [
        {
          id: 9001,
          task_id: 501,
          user_id: 1,
          url: '/api/v1/user/image-creator/images/9001/file',
          output_format: 'png',
          mime_type: 'image/png',
          byte_size: 1200,
          sha256: 'sha',
          expires_at: '2026-05-21T00:00:00Z',
          created_at: '2026-05-20T00:12:00Z',
        },
      ],
    })

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(getImageTask).toHaveBeenCalledWith(501)
    expect(wrapper.find('[data-testid="canvas-node-preview-image"]').attributes('src')).toBe('/api/v1/user/image-creator/images/9001/file')
    expect(wrapper.find('[data-testid="canvas-latest-run"]').text()).toContain('canvas.imageTaskSummary')

    wrapper.unmount()
  })

  it('renders canvas image task failures on the mapped node', async () => {
    getCanvas.mockResolvedValue(makeCanvas({
      document: {
        nodes: [
          {
            id: 'node_result',
            type: 'text_to_image',
            title: 'Text to image',
            x: 80,
            y: 90,
            width: 170,
            height: 86,
            status: 'idle',
            config: {},
          },
        ],
        edges: [],
      },
    }))
    listCanvasRuns.mockResolvedValue({
      items: [
        {
          id: 'run_1',
          canvas_id: 'canvas_1',
          status: 'running',
          api_key_id: 101,
          output: {
            image_tasks: [
              {
                node_id: 'node_result',
                task_id: 502,
                task_status: 'pending',
              },
            ],
          },
          created_at: '2026-05-20T00:11:00Z',
          updated_at: '2026-05-20T00:11:00Z',
        },
      ],
      total: 1,
    })
    getImageTask.mockResolvedValue({
      id: 502,
      user_id: 1,
      api_key_id: 101,
      status: 'failed',
      model: 'gpt-image-2',
      prompt: 'paint a city',
      size: '1024x1024',
      quality: 'standard',
      output_format: 'png',
      background: 'auto',
      count: 1,
      error_message: 'upstream refused the request',
      expires_at: '2026-05-21T00:00:00Z',
      created_at: '2026-05-20T00:11:00Z',
      updated_at: '2026-05-20T00:12:00Z',
      images: [],
    })

    const wrapper = mountView()
    await flushPromises()
    await flushPromises()

    expect(getImageTask).toHaveBeenCalledWith(502)
    expect(wrapper.find('[data-testid="canvas-node-error"]').text()).toContain('upstream refused the request')

    wrapper.unmount()
  })
})
