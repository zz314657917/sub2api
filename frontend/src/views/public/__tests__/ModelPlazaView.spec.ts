import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { ModelMarketCatalog } from '@/api/modelMarket'
import ModelPlazaView from '../ModelPlazaView.vue'

const getCatalogMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/modelMarket', () => ({
  modelMarketAPI: {
    getCatalog: getCatalogMock
  }
}))

const modelPlazaSource = readFileSync(resolve(process.cwd(), 'src/views/public/ModelPlazaView.vue'), 'utf8')
const mountedWrappers: Array<{ unmount: () => void }> = []

function createCatalog(): ModelMarketCatalog {
  const accountGroups = [
    {
      id: 1,
      name: '优惠预览组',
      platform: 'openai',
      rate_multiplier: 0.5,
      image_rate_independent: false,
      image_rate_multiplier: 1,
      effective_rate_multiplier: 0.5
    },
    {
      id: 2,
      name: '标准预览组',
      platform: 'openai',
      rate_multiplier: 1,
      image_rate_independent: false,
      image_rate_multiplier: 1,
      effective_rate_multiplier: 1
    }
  ]

  return {
    version: 1,
    description: '分组定价以管理员发布的可见规则为准。',
    groups: [
      {
        id: 'chat-openai',
        title: 'OpenAI 推理',
        category: 'chat',
        platform: 'openai',
        description: '通用推理模型',
        sort_order: 1,
        enabled: true,
        supported_groups: accountGroups,
        rows: [
          {
            id: 'gpt-4o',
            model: 'gpt-4o',
            input_price: '✪5 / 百万 tokens',
            output_price: '✪15 / 百万 tokens',
            our_price: '✪10 / 百万 tokens',
            note: '128K 上下文',
            sort_order: 1,
            enabled: true
          },
          {
            id: 'gpt-4-1-mini',
            model: 'gpt-4.1-mini',
            input_price: '✪1 / 百万 tokens',
            output_price: '✪4 / 百万 tokens',
            our_price: '✪3 / 百万 tokens',
            note: '1M 上下文',
            sort_order: 2,
            enabled: true
          },
          {
            id: 'gpt-disabled',
            model: 'gpt-disabled',
            our_price: '✪1',
            sort_order: 3,
            enabled: false
          }
        ]
      },
      {
        id: 'image-openai',
        title: '图像生成',
        category: 'image',
        platform: 'openai-image',
        description: '按尺寸展示图像价格',
        sort_order: 2,
        enabled: true,
        rows: Array.from({ length: 8 }, (_, index) => ({
          id: `image-${index + 1}`,
          spec: `${512 + index * 128} x ${512 + index * 128}`,
          our_price: `✪${index + 1}`,
          official_price: `$${index + 2}`,
          saving: `${10 + index}%`,
          note: index === 0 ? '高清规格' : `图像规格 ${index + 1}`,
          sort_order: index + 1,
          enabled: true
        }))
      },
      {
        id: 'video-generic',
        title: '视频生成',
        category: 'video',
        platform: 'video',
        description: '按任务展示视频价格',
        sort_order: 3,
        enabled: true,
        rows: [
          {
            id: 'video-hd',
            spec: '1080p / 10 秒',
            our_price: '✪20',
            note: '高清规格',
            sort_order: 1,
            enabled: true
          },
          {
            id: 'video-standard',
            spec: '720p / 5 秒',
            our_price: '✪8',
            note: '标准规格',
            sort_order: 2,
            enabled: true
          }
        ]
      }
    ]
  }
}

function mountModelPlaza() {
  const wrapper = mount(ModelPlazaView, {
    global: {
      stubs: {
        Icon: true,
        ModelIcon: true,
        PublicRevealBackdrop: true,
        PublicTopNav: true,
        RouterLink: {
          props: ['to'],
          template: '<a :href="to"><slot /></a>'
        }
      }
    }
  })
  mountedWrappers.push(wrapper)
  return wrapper
}

describe('ModelPlazaView model discovery', () => {
  beforeEach(() => {
    getCatalogMock.mockReset()
    getCatalogMock.mockResolvedValue(createCatalog())
  })

  afterEach(() => {
    mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  })

  it('uses a full-width search row and a flat filter row for tabs and result count', async () => {
    const wrapper = mountModelPlaza()
    await flushPromises()

    const toolbar = wrapper.get('.model-toolbar')
    const searchBox = toolbar.get('.model-search-box')
    const filterRow = toolbar.get('.model-filter-row')

    expect(searchBox.element.parentElement).toBe(toolbar.element)
    expect(filterRow.element.parentElement).toBe(toolbar.element)
    expect(filterRow.get('.model-category-tabs').element.parentElement).toBe(filterRow.element)
    expect(filterRow.get('[data-testid="model-result-count"]').element.parentElement).toBe(filterRow.element)
    expect(modelPlazaSource).not.toMatch(/\.model-toolbar\s*\{[^}]*(?:border:|background:|box-shadow:|backdrop-filter:)/s)
    expect(modelPlazaSource).not.toMatch(/\.model-category-tabs\s*\{[^}]*(?:border:|background:|box-shadow:|backdrop-filter:)/s)
  })

  it('filters matching rows, combines search with category, and restores rows after clearing search', async () => {
    const wrapper = mountModelPlaza()
    await flushPromises()

    const searchInput = wrapper.get('[data-testid="model-search-input"]')
    await searchInput.setValue('gpt-4.1-mini')

    expect(wrapper.findAll('.model-pricing-table tbody tr')).toHaveLength(1)
    expect(wrapper.get('.model-card-grid').text()).toContain('gpt-4.1-mini')
    expect(wrapper.get('.model-card-grid').text()).not.toContain('gpt-4o')
    expect(wrapper.get('[data-testid="model-result-count"]').text()).toContain('共 1 个型号/规格，来自 1 个分组')

    await wrapper.get('.model-search-clear').trigger('click')
    expect(wrapper.findAll('.model-pricing-table tbody tr')).toHaveLength(12)
    expect(wrapper.get('[data-testid="model-result-count"]').text()).toContain('共 12 个型号/规格，来自 3 个分组')

    await searchInput.setValue('高清规格')
    expect(wrapper.findAll('.model-pricing-table tbody tr')).toHaveLength(2)
    expect(wrapper.get('[data-testid="model-result-count"]').text()).toContain('共 2 个型号/规格，来自 2 个分组')

    const imageTab = wrapper.findAll('.model-category-tabs button').find((button) => button.text().startsWith('图像'))
    expect(imageTab).toBeDefined()
    await imageTab!.trigger('click')

    expect(wrapper.findAll('.model-pricing-table tbody tr')).toHaveLength(1)
    expect(wrapper.get('.model-card-grid').attributes('class')).toContain('model-card-grid')
    expect(wrapper.get('.model-market-card').attributes('data-category')).toBe('image')
    expect(wrapper.get('[data-testid="model-result-count"]').text()).toContain('共 1 个型号/规格，来自 1 个分组')
  })

  it('provides a finite mobile preview with explicit expand and collapse controls', async () => {
    const wrapper = mountModelPlaza()
    await flushPromises()

    const imageCard = wrapper.findAll('[data-testid="model-market-card"]')
      .find((card) => card.attributes('data-category') === 'image')
    expect(imageCard).toBeDefined()
    expect(imageCard!.findAll('tbody tr')).toHaveLength(8)
    expect(imageCard!.findAll('tbody tr.is-mobile-preview-hidden')).toHaveLength(2)

    const toggle = imageCard!.get('[data-testid="model-mobile-preview-toggle"]')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(toggle.text()).toContain('展开其余 2 个型号/规格')

    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('true')
    expect(toggle.text()).toContain('收起完整列表')
    expect(imageCard!.findAll('tbody tr.is-mobile-preview-hidden')).toHaveLength(0)

    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(imageCard!.findAll('tbody tr.is-mobile-preview-hidden')).toHaveLength(2)
    expect(modelPlazaSource).toMatch(/\.model-table-wrap\.is-scrollable \{\s*max-height: none;\s*overflow-y: visible;/)
  })

  it('labels group pricing as a preview, explains credits, and links to the tutorial', async () => {
    const wrapper = mountModelPlaza()
    await flushPromises()

    expect(wrapper.get('.model-tutorial-link').attributes('href')).toBe('/tutorial/getting-started')
    expect(wrapper.get('.model-price-context').text()).toContain('当前选择的账号分组仅用于价格预览')
    expect(wrapper.get('.model-price-context').text()).toContain('不代表匿名访问者或登录账号的实际分组')
    expect(wrapper.get('.model-price-context').text()).toContain('✪ 是本站额度单位，不代表人民币或美元')
    expect(wrapper.get('.model-plaza-description').text()).toContain('分组定价以管理员发布的可见规则为准。')
    expect((wrapper.get('.model-group-rate-select select').element as HTMLSelectElement).value).toBe('2')
    expect(wrapper.get('.model-market-card[data-category="chat"] .model-price-value').text()).toBe('✪10 / 百万 tokens')
  })

  it('keeps the existing catalog visible when a refresh fails', async () => {
    getCatalogMock
      .mockReset()
      .mockResolvedValueOnce(createCatalog())
      .mockRejectedValueOnce(new Error('网络超时'))

    const wrapper = mountModelPlaza()
    await flushPromises()
    expect(wrapper.text()).toContain('gpt-4o')

    await wrapper.get('.model-refresh-button').trigger('click')
    await flushPromises()

    expect(getCatalogMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('gpt-4o')
    expect(wrapper.find('.model-card-grid').exists()).toBe(true)
    expect(wrapper.find('.model-message-card').exists()).toBe(false)
    expect(wrapper.get('.model-refresh-warning').text()).toContain('目录刷新失败，正在显示上次成功加载的内容')
    expect(wrapper.get('.model-refresh-warning').text()).toContain('网络超时')
  })
})
