import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const homeView = readFileSync(resolve(process.cwd(), 'src/views/HomeView.vue'), 'utf8')
const zhLocale = readFileSync(resolve(process.cwd(), 'src/i18n/locales/zh.ts'), 'utf8')
const mainEntry = readFileSync(resolve(process.cwd(), 'src/main.ts'), 'utf8')
const pixelUi = readFileSync(resolve(process.cwd(), 'src/styles/pixel-ui.css'), 'utf8')

describe('home page visual direction', () => {
  it('uses the compact coding-solution hero copy', () => {
    expect(zhLocale).toContain("heroEyebrow: '国内直连 · 统一密钥 · 可试用'")
    expect(zhLocale).toContain("heroTitleTop: 'ChatGPT'")
    expect(zhLocale).toContain("heroTitleBottom: '国内直连方案'")
    expect(zhLocale).toContain("heroDescription: '即刻体验 ChatGPT 最新模型，添加客服领取试用额度。'")
    expect(zhLocale).toContain("unifiedGatewayDesc: '无需挂代理翻墙，密钥接入编码模型与上游能力。'")
    expect(zhLocale).toContain("claimButton: '注册领取试用'")
    expect(zhLocale).toContain("goToDashboard: '登录控制台'")
    expect(homeView).toContain("isAuthenticated ? dashboardPath : '/register'")
  })

  it('drops the old terminal and tiled background from the default homepage', () => {
    expect(homeView).not.toContain('terminal-container')
    expect(homeView).not.toContain('bg-mesh-gradient')
    expect(homeView).toContain('home-violet-bg')
    expect(homeView).toContain('home-matrix-rain')
    expect(homeView).toContain('matrixColumns')
    expect(homeView).toContain('const matrixColumnCount = 51')
    expect(homeView).toContain('@keyframes matrix-rain-fall')
    expect(homeView).toContain('home-blur-field')
  })

  it('adds a clipped gradient sweep inside the hero title text', () => {
    expect(homeView).toContain('home-title-sweep')
    expect(homeView).toContain('home-title-fill')
    expect(homeView).toContain('<span class="home-title-line home-title-fill block">{{ t(\'home.heroTitleTop\') }}</span>')
    expect(homeView).toContain('<span class="home-title-line block">{{ t(\'home.heroTitleBottom\') }}</span>')
    expect(homeView).toContain('background-clip: text')
    expect(homeView).toContain('@keyframes title-fill-sweep')
    expect(homeView).toContain('@media (prefers-reduced-motion: reduce)')
  })

  it('uses public contact info for a safe support shortcut', () => {
    expect(homeView).toContain('contactInfo')
    expect(homeView).toContain('supportHref')
    expect(homeView).toContain('home-support-button')
    expect(homeView).toContain("['http:', 'https:', 'mailto:', 'tel:']")
    expect(zhLocale).toContain("contactSupport: '添加客服'")
  })

  it('uses the full-width translucent top navigation treatment', () => {
    expect(homeView).toContain('home-top-shell')
    expect(homeView).toContain('home-top-nav')
    expect(homeView).toContain('home-page-shell')
    expect(homeView).toContain('home-brand-logo')
    expect(homeView).toContain("t('home.navFeatures')")
    expect(homeView).toContain('home-locale-switcher')
    expect(homeView).toContain('backdrop-filter: blur(18px)')
    expect(homeView).toContain('width: 100%')
    expect(homeView).toContain('rgba(4, 3, 16, 0.62)')
    expect(zhLocale).toContain("navFeatures: '功能特性'")
  })

  it('keeps the hero responsive on wide desktop screens', () => {
    expect(homeView).toContain('home-page-root')
    expect(homeView).toContain('home-main-stage')
    expect(homeView).toContain('home-hero-shell')
    expect(homeView).toContain('home-hero-content')
    expect(homeView).toContain('home-feature-grid')
    expect(homeView).toContain('@media (min-width: 1536px)')
    expect(homeView).toContain('@media (min-width: 1920px)')
    expect(homeView).toContain('min-height: calc(100vh - 3.75rem)')
    expect(homeView).toContain('width: min(100%, 92rem)')
    expect(homeView).toContain('gap: clamp(3.5rem, 5.4vh, 5rem)')
    expect(homeView).toContain('font-size: clamp(6.25rem, 4.35vw, 7.4rem)')
  })

  it('keeps the documentation link out of the primary hero action row', () => {
    expect(homeView).not.toContain('home-secondary-link')
    expect(homeView).not.toContain("t('home.docs')")
  })

  it('keeps pixel icons in a shared visual style file', () => {
    expect(mainEntry).toContain("import './styles/pixel-ui.css'")
    expect(homeView).toContain('home-feature-icon-frame')
    expect(homeView).toContain('<PixelIcon :name="item.icon" size="md" />')
    expect(homeView).toContain('<PixelIcon :name="primaryActionIcon" size="sm" />')
    expect(homeView).toContain('<PixelIcon name="support" size="sm" />')
    expect(homeView).not.toContain('heroTags')
    expect(homeView).not.toContain('home-chip')
    expect(homeView).not.toContain('.home-chip-icon')
    expect(homeView).not.toContain('.home-feature-pixel-icon')
    expect(homeView).not.toContain("from '@/components/icons/Icon.vue'")
    expect(pixelUi).toContain('.pixel-glyph')
    expect(pixelUi).toContain('.pixel-glyph__pixel--1')
    expect(pixelUi).toContain('.pixel-glyph__pixel--2')
    expect(pixelUi).not.toContain('.pixel-icon--direct')
    expect(pixelUi).not.toContain('.pixel-icon-frame')
  })
})
