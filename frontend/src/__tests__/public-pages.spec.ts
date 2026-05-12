import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const router = readFileSync(resolve(process.cwd(), 'src/router/index.ts'), 'utf8')
const tutorialView = readFileSync(resolve(process.cwd(), 'src/views/public/TutorialView.vue'), 'utf8')
const modelPlazaView = readFileSync(resolve(process.cwd(), 'src/views/public/ModelPlazaView.vue'), 'utf8')
const publicBackdrop = readFileSync(resolve(process.cwd(), 'src/views/public/components/PublicMatrixBackdrop.vue'), 'utf8')
const publicCss = readFileSync(resolve(process.cwd(), 'src/views/public/public-page.css'), 'utf8')
const sourceSiteBrandPattern = new RegExp(['Ti', 'Mi|TIM', 'ICC|tim', 'icc'].join(''), 'i')

describe('public tutorial and model plaza pages', () => {
  it('registers tutorial and model plaza as public routes', () => {
    expect(router).toContain("path: '/tutorial'")
    expect(router).toContain("name: 'Tutorial'")
    expect(router).toContain("path: '/models'")
    expect(router).toContain("name: 'ModelPlaza'")
    expect(router).toContain("'/tutorial'")
    expect(router).toContain("'/models'")
  })

  it('uses original Luoye Network tutorial content instead of copied source-site branding', () => {
    expect(tutorialView).toContain('落叶网络接入教程')
    expect(tutorialView).toContain('创建账户并领取试用额度')
    expect(tutorialView).toContain('接入 Codex / OpenAI SDK')
    expect(tutorialView).toContain("siteName || '落叶网络'")
    expect(tutorialView).not.toMatch(sourceSiteBrandPattern)
  })

  it('adds a public model plaza with searchable pricing cards', () => {
    expect(modelPlazaView).toContain('模型广场')
    expect(modelPlazaView).toContain('浏览可用的 AI 模型及其定价')
    expect(modelPlazaView).toContain('model-filter-panel')
    expect(modelPlazaView).toContain('model-card-grid')
    expect(modelPlazaView).toContain('显示倍率价格')
    expect(modelPlazaView).toContain('Prompt Caching')
    expect(modelPlazaView).toContain('userChannelsAPI.getAvailable')
    expect(modelPlazaView).toContain('userGroupsAPI.getUserGroupRates')
    expect(modelPlazaView).toContain('fallbackChannels')
    expect(modelPlazaView).toContain('formatModelPrice')
    expect(modelPlazaView).not.toMatch(sourceSiteBrandPattern)
  })

  it('shares the public matrix visual treatment on the tutorial page', () => {
    expect(tutorialView).toContain('PublicMatrixBackdrop')
    expect(publicBackdrop).toContain('public-matrix-rain')
    expect(publicBackdrop).toContain('Array.from({ length: 42 }')
    expect(publicCss).toContain('.public-page-header')
    expect(publicCss).toContain('.public-nav-button')
  })
})
