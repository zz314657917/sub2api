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

    await wrapper.findAll('button').find(button => button.text() === 'Mac / Linux / 手动配置')!.trigger('click')
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

  it('keeps three primary steps and collapses optional help', async () => {
    const { wrapper } = await mountTutorial('/tutorial')
    await wrapper.findAll('button').find(button => button.text() === 'Mac / Linux / 手动配置')!.trigger('click')
    expect(wrapper.findAll('.tutorial-quickstart-step')).toHaveLength(3)
    expect(wrapper.find('.tutorial-quickstart-facts').exists()).toBe(false)
    expect(wrapper.find('.tutorial-quickstart-tile-grid').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('首次必做')
    for (const selector of ['.tutorial-install', '.tutorial-directory', '.tutorial-api', '.tutorial-troubleshooting']) {
      expect(wrapper.get<HTMLDetailsElement>(selector).element.open).toBe(false)
    }
    const config = wrapper.get('.tutorial-quickstart-step:nth-child(2) > .tutorial-quickstart-code pre code').text()
    expect(config).toContain('model_provider = "OpenAI"')
    expect(config).toContain('model = "gpt-6.1-sol"')
    expect(config).toContain('review_model = "gpt-6.1-sol"')
    expect(config).toContain('base_url = "https://ai.3zapi.com"')
    expect(config).toContain('requires_openai_auth = true')
    expect(config).toContain('stream_max_retries = 1')
    expect(config).not.toContain('env_key')
    expect(config).not.toContain('disable_response_storage')
    expect(config).not.toContain('network_access')
    expect(config).not.toContain('windows_wsl_setup_acknowledged')
    expect(wrapper.get('.tutorial-quickstart-code--large').text())
      .toContain('"model":"gpt-6.1-sol"')
    expect(JSON.parse(wrapper.get('.tutorial-quickstart-step:nth-child(3) pre code').text()))
      .toEqual({ OPENAI_API_KEY: '替换成你的 API Key' })
    expect(wrapper.get('.tutorial-install').text()).toContain('下载 Codex App')
    expect(wrapper.get('.tutorial-install').text()).toContain('@openai/codex')
    expect(wrapper.get('.tutorial-troubleshooting').text()).toContain('官方算力不足')

    const claude = wrapper.findAll('.tutorial-segmented-control button').find(button => button.text() === 'Claude')!
    await claude.trigger('click')
    expect(wrapper.findAll('.tutorial-quickstart-step')).toHaveLength(3)
    expect(wrapper.find('.tutorial-directory').exists()).toBe(false)
    expect(wrapper.get('.tutorial-install').text()).toContain('@anthropic-ai/claude-code')
    expect(wrapper.get('.tutorial-quickstart-step:nth-child(2) pre code').text()).toContain('ANTHROPIC_AUTH_TOKEN')
    expect(wrapper.get('.tutorial-quickstart-step:nth-child(3) pre code').text()).toBe('claude')
    const unix = wrapper.findAll('.tutorial-segmented-control button').find(button => button.text() === 'macOS / Linux')!
    await unix.trigger('click')
    expect(wrapper.get('.tutorial-quickstart-step:nth-child(2) pre code').text()).toContain('export ANTHROPIC_AUTH_TOKEN')
  })

  it('copies quick-start commands and shows feedback', async () => {
    const { wrapper } = await mountTutorial('/tutorial')
    await wrapper.findAll('button').find(button => button.text() === 'Mac / Linux / 手动配置')!.trigger('click')
    const button = wrapper.find<HTMLButtonElement>('.tutorial-quickstart-step > .tutorial-quickstart-code .tutorial-quickstart-code-head button')

    await button.trigger('click')
    await flushPromises()

    expect(button.text()).toBe('已复制')
    expect(showSuccessMock).toHaveBeenCalledWith('已复制命令')
  })

  it('starts with four illustrated Windows helper steps and keeps technical setup out of the beginner flow', async () => {
    const { wrapper } = await mountTutorial('/tutorial')

    expect(wrapper.get('h1').text()).toBe('第一次使用，跟着这四步')
    expect(wrapper.findAll('.tutorial-helper-step h2').map(heading => heading.text())).toEqual([
      '先下载安装官方 ChatGPT',
      '登录网站，复制你的密钥',
      '下载配置助手，粘贴密钥并保存',
      '发送第一条消息',
    ])
    const steps = wrapper.findAll('.tutorial-helper-step')
    const officialDownload = steps[0].get('[data-testid="tutorial-official-download"]')
    expect(officialDownload.attributes('href')).toBe('https://chatgpt.com/download/')
    expect(officialDownload.attributes('target')).toBe('_blank')
    expect(officialDownload.attributes('rel')).toBe('noopener noreferrer')
    expect(steps[0].find('[data-testid="tutorial-helper-download"]').exists()).toBe(false)
    expect(steps[2].find('[data-testid="tutorial-helper-download"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="tutorial-helper-download"]').attributes('href')).toBe('/downloads/chatgpt-config-windows.exe')
    expect(wrapper.get('[data-testid="tutorial-helper-download"]').attributes('download')).toBe('ChatGPT配置助手.exe')
    expect(wrapper.get('.tutorial-helper-step a[href="/keys"]').exists()).toBe(true)
    expect(wrapper.find('.tutorial-quickstart-steps').exists()).toBe(false)
    expect(wrapper.find('[aria-label="选择系统和终端"]').exists()).toBe(false)
    expect(wrapper.find('.tutorial-api').exists()).toBe(false)
    expect(wrapper.get('.tutorial-helper-screenshot img').attributes('src')).toBe('/tutorial/config-helper/windows-setup.png')
    expect(wrapper.get('.tutorial-beginner').text()).toContain('实际配置的是 Codex 桌面功能')
    expect(wrapper.get('.tutorial-beginner').text()).toContain('收到正常文字回复，才算接入成功')
    expect(wrapper.get<HTMLDetailsElement>('.tutorial-helper-help').element.open).toBe(false)

    await wrapper.get('.tutorial-helper-message button').trigger('click')
    await flushPromises()
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('你好，请用一句话介绍你能帮我做什么。')
    expect(wrapper.get('.tutorial-helper-message button').text()).toBe('已复制')

    await wrapper.get('.tutorial-helper-screenshot').trigger('keydown', { key: 'Enter' })
    expect(wrapper.get('[role="dialog"] img').attributes('src')).toBe(new URL('/tutorial/config-helper/windows-setup.png', window.location.href).href)
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('switches between helper, manual Codex and Claude without losing configured API settings', async () => {
    const config = JSON.parse(JSON.stringify(defaultQuickstartTutorialConfig))
    config.platforms[0].base_url = 'https://example.com/custom'
    getQuickstartConfigMock.mockResolvedValue(config)
    const { wrapper } = await mountTutorial('/tutorial')
    const click = async (label: string) => {
      await wrapper.findAll('button').find(button => button.text() === label)!.trigger('click')
    }

    await click('Claude')
    expect(wrapper.find('.tutorial-beginner').exists()).toBe(false)
    expect(wrapper.findAll('.tutorial-quickstart-step')).toHaveLength(3)
    expect(wrapper.get('.tutorial-quickstart-step:nth-child(2) pre code').text()).toContain('ANTHROPIC_AUTH_TOKEN')
    await click('ChatGPT / Codex')
    expect(wrapper.findAll('.tutorial-helper-step')).toHaveLength(4)

    await click('Mac / Linux / 手动配置')
    expect(wrapper.get('.tutorial-quickstart-steps').text()).toContain('base_url = "https://example.com/custom"')
    await click('macOS / Linux')
    expect(wrapper.get('.tutorial-directory').text()).toContain('~/.codex')
    await click('Windows 助手（推荐）')
    expect(wrapper.find('[aria-label="选择系统和终端"]').exists()).toBe(false)
    expect(wrapper.findAll('.tutorial-helper-step')).toHaveLength(4)
    await click('Mac / Linux / 手动配置')
    expect(wrapper.get('[aria-label="选择系统和终端"] button[aria-pressed="true"]').text()).toBe('macOS / Linux')
  })

  it('offers the beginner route above a CMS-managed Codex article without replacing its content', async () => {
    const page = { ...tutorialPages[1], slug: 'codex', title: 'Codex', content_md: '# Codex\n\n管理员自定义内容。' }
    listMock.mockResolvedValue([{ ...summaries[1], slug: 'codex', title: 'Codex' }])
    getBySlugMock.mockResolvedValue(page)
    const { wrapper, router } = await mountTutorial('/tutorial')
    await wrapper.findAll('button').find(button => button.text() === 'Mac / Linux / 手动配置')!.trigger('click')
    await router.push('/tutorial/codex')
    await flushPromises()

    expect(wrapper.get('.tutorial-content').text()).toContain('管理员自定义内容')
    expect(wrapper.get('.tutorial-article-beginner-link').attributes('href')).toBe('/tutorial')
    await wrapper.get('.tutorial-article-beginner-link').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/tutorial')
    expect(wrapper.findAll('.tutorial-helper-step')).toHaveLength(4)
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
