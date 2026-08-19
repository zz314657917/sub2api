import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, nextTick } from 'vue'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { TutorialPage, TutorialPageSummary } from '@/types'
import TutorialView from '../TutorialView.vue'
import { defaultQuickstartTutorialConfig } from '../tutorialQuickstart'

const { getBySlugMock, getQuickstartConfigMock, listMock, showErrorMock, showSuccessMock } = vi.hoisted(() => ({
  getBySlugMock: vi.fn(),
  getQuickstartConfigMock: vi.fn(),
  listMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
}))

vi.mock('@/api/tutorials', () => ({
  default: {
    getBySlug: getBySlugMock,
    getQuickstartConfig: getQuickstartConfigMock,
    list: listMock,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
  }),
}))

const timestamp = '2026-07-10T00:00:00Z'
const tutorialPages: TutorialPage[] = [
  {
    id: 1,
    slug: 'getting-started',
    title: '快速开始',
    description: '完成第一次 API 接入。',
    category: '入门',
    sort_order: 10,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: `
# 快速开始

这是第一段正文，直达详情时应立即可见。

## 安装

先完成安装。

\`\`\`bash
echo plain
\`\`\`

[[command title="短代码命令"]]
echo shortcode
[[/command]]

## 验证

[[screenshot src="/tutorial/example.png" alt="示例截图" caption="验证界面"]]
`.trim(),
  },
  {
    id: 2,
    slug: 'advanced',
    title: '进阶配置',
    description: '继续配置高级工具。',
    category: '工具配置',
    sort_order: 20,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: '# 进阶配置\n\n进阶正文。\n\n## 调整\n\n完成调整。',
  },
]

const summaries: TutorialPageSummary[] = tutorialPages.map(({ content_md: _content, ...page }) => page)
const imageTutorialPages: TutorialPage[] = [
  {
    id: 74,
    slug: 'gemini-3-pro-image-preview',
    title: 'Gemini 3 Pro Image Preview 图像生成',
    description: '使用 Gemini 3 Pro Image Preview 生成图像。',
    category: '图像模型',
    sort_order: 2242,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: '# Gemini 3 Pro Image Preview 图像生成\n\n图像教程正文。',
  },
  {
    id: 75,
    slug: 'gemini-3-pro-image-preview-official',
    title: 'Gemini 3 Pro Image Preview 官方图像生成',
    description: '使用 Gemini 3 Pro Image Preview 官方路径生成图像。',
    category: '图像模型',
    sort_order: 2243,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: '# Gemini 3 Pro Image Preview 官方图像生成\n\n官方图像教程正文。',
  },
  {
    id: 80,
    slug: 'doubao-seedance-4-5',
    title: '豆包 Seedance 4.5 图像生成',
    description: '使用豆包 Seedance 4.5 异步图像生成。',
    category: '图像模型',
    sort_order: 2248,
    status: 'published',
    created_at: timestamp,
    updated_at: timestamp,
    published_at: timestamp,
    content_md: '# 豆包 Seedance 4.5 图像生成\n\n异步图像教程正文。',
  },
]
const imageSummaries: TutorialPageSummary[] = imageTutorialPages.map(({ content_md: _content, ...page }) => page)
const mountedWrappers: VueWrapper[] = []
const scrollIntoViewMock = vi.fn()
const originalScrollIntoView = HTMLElement.prototype.scrollIntoView

class IntersectionObserverMock {
  observe = vi.fn()
  disconnect = vi.fn()
  unobserve = vi.fn()
}

function createTutorialRouter(): Router {
  const routeComponent = defineComponent({ template: '<div />' })
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/tutorial', name: 'Tutorial', component: routeComponent },
      { path: '/tutorial/:slug', name: 'TutorialPage', component: routeComponent },
      { path: '/keys', name: 'Keys', component: routeComponent },
      { path: '/models', name: 'Models', component: routeComponent },
    ],
  })
}

async function mountTutorial(path: string) {
  const router = createTutorialRouter()
  await router.push(path)
  await router.isReady()
  const wrapper = mount(TutorialView, {
    attachTo: document.body,
    global: {
      plugins: [router],
      stubs: {
        PublicRevealBackdrop: true,
        PublicTopNav: true,
      },
    },
  })
  mountedWrappers.push(wrapper)
  await flushPromises()
  await nextTick()
  return { router, wrapper }
}

function errorWithStatus(status: number, message: string) {
  return Object.assign(new Error(message), {
    response: { status, data: { message } },
  })
}

beforeEach(() => {
  vi.clearAllMocks()
  listMock.mockResolvedValue(summaries)
  getQuickstartConfigMock.mockResolvedValue(defaultQuickstartTutorialConfig)
  getBySlugMock.mockImplementation(async (slug: string) => {
    const page = tutorialPages.find((item) => item.slug === slug)
    if (!page) throw errorWithStatus(404, '教程不存在')
    return page
  })
  Object.defineProperty(window, 'IntersectionObserver', {
    configurable: true,
    value: IntersectionObserverMock,
  })
  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    configurable: true,
    value: function scrollIntoView(options?: ScrollIntoViewOptions) {
      scrollIntoViewMock(this.id, options)
    },
  })
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: vi.fn().mockResolvedValue(undefined) },
  })
})

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  document.body.innerHTML = ''
  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    configurable: true,
    value: originalScrollIntoView,
  })
})

describe('TutorialView reading flow', () => {
  it('uses the public quick-start configuration for the Base URL and generated commands', async () => {
    const config = JSON.parse(JSON.stringify(defaultQuickstartTutorialConfig))
    config.platforms[0].base_url = 'https://ai.3zapi.cc'
    getQuickstartConfigMock.mockResolvedValue(config)

    const { wrapper } = await mountTutorial('/tutorial')

    expect(wrapper.find('.tutorial-quickstart-facts').text()).toContain('https://ai.3zapi.cc')
    expect(wrapper.find('.tutorial-quickstart-steps').text()).toContain('base_url = "https://ai.3zapi.cc"')
    expect(wrapper.find('.tutorial-quickstart-code--large').text()).toContain('curl https://ai.3zapi.cc/responses')
  })

  it('keeps quick start separate from the searchable full tutorial directory', async () => {
    const { wrapper: quickstartWrapper } = await mountTutorial('/tutorial')

    expect(quickstartWrapper.find('.tutorial-quickstart').exists()).toBe(true)
    expect(quickstartWrapper.find('.tutorial-overview').exists()).toBe(false)
    expect(quickstartWrapper.find('.tutorial-index-controls').exists()).toBe(false)
    expect(quickstartWrapper.find('.guide-action-link').attributes('href')).toBe('/tutorial?view=library')
    expect(
      quickstartWrapper
        .findAll<HTMLButtonElement>('.tutorial-segmented-control[aria-label="选择模型平台"] button')
        .map((button) => button.text())
    ).toEqual(['ChatGPT / Codex', 'Claude'])

    const { wrapper } = await mountTutorial('/tutorial?view=library')

    expect(wrapper.find('.tutorial-quickstart').exists()).toBe(false)
    expect(wrapper.find('.tutorial-index-controls').exists()).toBe(true)
    expect(wrapper.findAll('.tutorial-category-group')).toHaveLength(2)
    expect(wrapper.find('.tutorial-article').exists()).toBe(false)

    await wrapper.find('input[type="search"]').setValue('进阶')

    const cards = wrapper.findAll('.tutorial-directory-card')
    expect(cards).toHaveLength(1)
    expect(cards[0].text()).toContain('进阶配置')
  })

  it('renders the quick-start guide and updates platform and terminal variants', async () => {
    const { wrapper } = await mountTutorial('/tutorial')

    expect(wrapper.find('.tutorial-quickstart').exists()).toBe(true)
    expect(wrapper.findAll('.tutorial-quickstart-step')).toHaveLength(5)
    expect(wrapper.find('.tutorial-quickstart-fact').text()).toContain('https://ai.3zapi.top')
    expect(wrapper.find('[aria-label="选择教程模式"]').exists()).toBe(false)
    expect(wrapper.find('.tutorial-quickstart-error-grid').text()).toContain('官方算力不足')
    expect(wrapper.find('.tutorial-quickstart-error-grid').text()).toContain('Selected model is at capacity. Please try a different model.')
    expect(wrapper.find('.tutorial-quickstart-step:nth-child(3)').text()).toContain(
      'C:\\Users\\你的用户名\\.codex\\config.toml'
    )
    expect(wrapper.find('.tutorial-quickstart-step:nth-child(3)').text()).toContain(
      'explorer "%USERPROFILE%\\.codex"'
    )

    const desktopDownloadLink = wrapper.get<HTMLAnchorElement>(
      'a.tutorial-quickstart-link[href="https://developers.openai.com/codex/app#getting-started"]'
    )
    expect(desktopDownloadLink.text()).toContain('下载 ChatGPT Desktop')
    expect(desktopDownloadLink.attributes('target')).toBe('_blank')
    expect(desktopDownloadLink.attributes('rel')).toBe('noopener noreferrer')

    const claudeButton = wrapper
      .findAll<HTMLButtonElement>('.tutorial-segmented-control button')
      .find((button) => button.text() === 'Claude')
    expect(claudeButton).toBeDefined()
    await claudeButton!.trigger('click')
    expect(wrapper.find('.tutorial-quickstart-facts').text()).toContain('Anthropic Messages')
    expect(wrapper.find('.tutorial-quickstart-step:nth-child(3)').text()).toContain('Claude Code 不使用 config.toml')
    expect(
      wrapper.find('a.tutorial-quickstart-link[href="https://developers.openai.com/codex/app#getting-started"]').exists()
    ).toBe(false)
    expect(wrapper.find('.tutorial-quickstart-step:nth-child(2) .tutorial-quickstart-code').text()).toContain(
      '@anthropic-ai/claude-code'
    )

    const unixButton = wrapper
      .findAll<HTMLButtonElement>('.tutorial-segmented-control button')
      .find((button) => button.text() === 'macOS / Linux')
    expect(unixButton).toBeDefined()
    await unixButton!.trigger('click')
    expect(wrapper.find('.tutorial-quickstart-step:nth-child(3)').text()).toContain('当前终端设置环境变量')
  })

  it('copies quick-start commands and shows feedback', async () => {
    const { wrapper } = await mountTutorial('/tutorial')
    const button = wrapper.find<HTMLButtonElement>('.tutorial-quickstart-code-head button')

    await button.trigger('click')
    await flushPromises()

    expect(button.text()).toBe('已复制')
    expect(showSuccessMock).toHaveBeenCalledWith('已复制命令')
  })

  it('opens detail without the index hero and keeps progress, mobile directory, toc, and hash history in sync', async () => {
    const { router, wrapper } = await mountTutorial('/tutorial/getting-started#安装-0')

    expect(wrapper.find('.tutorial-overview').exists()).toBe(false)
    expect(wrapper.find('.tutorial-article h1').text()).toBe('快速开始')
    expect(wrapper.find('.tutorial-article-head').text()).toContain('第 1 篇，共 2 篇')
    expect(wrapper.find('.tutorial-content').text()).toContain('这是第一段正文')
    expect(wrapper.find('.tutorial-sidebar').exists()).toBe(true)
    expect(wrapper.find('.tutorial-reader--detail > .tutorial-sidebar').exists()).toBe(true)
    expect(wrapper.find('.tutorial-reader--detail > .tutorial-detail-column').exists()).toBe(true)
    expect(wrapper.find('.tutorial-reader--detail > .tutorial-toc').exists()).toBe(true)
    expect(wrapper.find('.tutorial-content-shell > .tutorial-toc').exists()).toBe(false)
    expect(wrapper.find('.tutorial-mobile-directory-toggle').text()).toContain('快速开始')
    expect(wrapper.find('.tutorial-mobile-toc').element.nextElementSibling).toBe(
      wrapper.find('.tutorial-content').element
    )
    expect(wrapper.find('.tutorial-page-link--next').text()).toContain('进阶配置')
    expect(scrollIntoViewMock).toHaveBeenCalledWith('安装-0', { behavior: 'auto', block: 'start' })

    const directoryToggle = wrapper.find('.tutorial-mobile-directory-toggle')
    expect(directoryToggle.attributes('aria-expanded')).toBe('false')
    await directoryToggle.trigger('click')
    expect(directoryToggle.attributes('aria-expanded')).toBe('true')
    expect(wrapper.find('#tutorial-mobile-directory-list').isVisible()).toBe(true)

    const desktopTocButtons = wrapper.findAll('.tutorial-toc button')
    await desktopTocButtons[1].trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.hash).toBe('#验证-1')

    scrollIntoViewMock.mockClear()
    router.back()
    await flushPromises()
    await nextTick()
    expect(router.currentRoute.value.hash).toBe('#安装-0')
    expect(scrollIntoViewMock).toHaveBeenCalledWith('安装-0', { behavior: 'auto', block: 'start' })
  })

  it('uses concise image model names in navigation while preserving the full article title', async () => {
    listMock.mockResolvedValue([...summaries, ...imageSummaries])
    getBySlugMock.mockImplementation(async (slug: string) => {
      const page = [...tutorialPages, ...imageTutorialPages].find((item) => item.slug === slug)
      if (!page) throw errorWithStatus(404, '教程不存在')
      return page
    })

    const { wrapper } = await mountTutorial('/tutorial/gemini-3-pro-image-preview')
    const desktopLabels = wrapper.findAll('.tutorial-tab-link strong')

    expect(desktopLabels.map((label) => label.text())).toEqual(expect.arrayContaining([
      'Gemini 3 Pro',
      'Gemini 3 Pro 官方',
      'Seedance 4.5',
    ]))

    const activeDesktopLabel = desktopLabels.find(
      (label) => label.attributes('title') === 'Gemini 3 Pro Image Preview 图像生成'
    )
    expect(activeDesktopLabel?.text()).toBe('Gemini 3 Pro')

    const directoryToggle = wrapper.find('.tutorial-mobile-directory-toggle')
    expect(directoryToggle.find('strong').text()).toBe('Gemini 3 Pro')
    expect(directoryToggle.find('strong').attributes('title')).toBe('Gemini 3 Pro Image Preview 图像生成')

    await directoryToggle.trigger('click')
    const mobileLabel = wrapper
      .findAll('#tutorial-mobile-directory-list strong')
      .find((label) => label.attributes('title') === 'Gemini 3 Pro Image Preview 图像生成')
    expect(mobileLabel?.text()).toBe('Gemini 3 Pro')
    expect(wrapper.find('.tutorial-article h1').text()).toBe('Gemini 3 Pro Image Preview 图像生成')
  })

  it('separates detail loading, retryable error, and 404 states', async () => {
    let resolvePage: ((page: TutorialPage) => void) | undefined
    getBySlugMock.mockImplementationOnce(
      () =>
        new Promise<TutorialPage>((resolve) => {
          resolvePage = resolve
        })
    )

    const loadingMount = await mountTutorial('/tutorial/getting-started')
    expect(loadingMount.wrapper.find('.tutorial-detail-state--loading').exists()).toBe(true)
    resolvePage?.(tutorialPages[0])
    await flushPromises()
    await nextTick()
    expect(loadingMount.wrapper.find('.tutorial-article').exists()).toBe(true)
    loadingMount.wrapper.unmount()

    getBySlugMock
      .mockRejectedValueOnce(errorWithStatus(503, '服务暂不可用'))
      .mockResolvedValueOnce(tutorialPages[0])
    const errorMount = await mountTutorial('/tutorial/getting-started')
    expect(errorMount.wrapper.find('.tutorial-detail-state--error').text()).toContain('服务暂不可用')
    expect(errorMount.wrapper.find('.tutorial-detail-state--not-found').exists()).toBe(false)
    await errorMount.wrapper.find('.tutorial-state-actions button').trigger('click')
    await flushPromises()
    expect(errorMount.wrapper.find('.tutorial-article').exists()).toBe(true)
    errorMount.wrapper.unmount()

    getBySlugMock.mockRejectedValueOnce(errorWithStatus(404, '教程不存在'))
    const missingMount = await mountTutorial('/tutorial/missing')
    expect(missingMount.wrapper.find('.tutorial-detail-state--not-found').text()).toContain('教程不存在')
    expect(missingMount.wrapper.find('.tutorial-detail-state--error').exists()).toBe(false)
  })

  it('shows feedback on the exact ordinary and shortcode copy buttons', async () => {
    const { wrapper } = await mountTutorial('/tutorial/getting-started')
    const buttons = wrapper.findAll<HTMLButtonElement>('[data-copy-code]')
    const plainButton = buttons.find((button) => decodeURIComponent(button.attributes('data-copy-code')) === 'echo plain')
    const shortcodeButton = buttons.find(
      (button) => decodeURIComponent(button.attributes('data-copy-code')) === 'echo shortcode'
    )

    expect(plainButton).toBeDefined()
    expect(shortcodeButton).toBeDefined()
    await plainButton!.trigger('click')
    await flushPromises()
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('echo plain')
    expect(plainButton!.text()).toBe('已复制')
    expect(shortcodeButton!.text()).toBe('复制')

    vi.mocked(navigator.clipboard.writeText).mockRejectedValueOnce(new Error('clipboard denied'))
    await shortcodeButton!.trigger('click')
    await flushPromises()
    expect(shortcodeButton!.text()).toBe('复制失败')
    expect(plainButton!.text()).toBe('已复制')
    expect(showErrorMock).toHaveBeenCalledWith('复制失败，请手动选择命令')
  })

  it('opens screenshots from the keyboard, focuses the dialog, and restores focus after Escape', async () => {
    const { wrapper } = await mountTutorial('/tutorial/getting-started')
    const screenshot = wrapper.find<HTMLElement>('.tutorial-screenshot-card')

    expect(screenshot.attributes('tabindex')).toBe('0')
    expect(screenshot.attributes('role')).toBe('button')
    screenshot.element.focus()
    await screenshot.trigger('keydown', { key: 'Enter' })
    await nextTick()

    const closeButton = wrapper.find<HTMLButtonElement>('.tutorial-image-lightbox__close')
    expect(wrapper.find('.tutorial-image-lightbox').exists()).toBe(true)
    expect(document.activeElement).toBe(closeButton.element)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()
    expect(wrapper.find('.tutorial-image-lightbox').exists()).toBe(false)
    expect(document.activeElement).toBe(screenshot.element)
  })
})
