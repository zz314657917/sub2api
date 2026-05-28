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
      edges: [
        {
          id: 'edge_1',
          source_node_id: 'node_prompt',
          target_node_id: 'node_result',
        },
      ],
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
          status: 'queued',
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
    expect(wrapper.find('[data-testid="canvas-view"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="canvas-node"]')).toHaveLength(2)

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

  it('updates an existing canvas and queues a run', async () => {
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
      model: 'gpt-image-2',
    })
    expect(showSuccess).toHaveBeenCalledWith('canvas.runQueued')
  })
})
