import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('AppHeader public shortcuts', () => {
  it('keeps tutorial, model plaza, and support actions in the console header', () => {
    expect(componentSource).toContain('to="/tutorial"')
    expect(componentSource).toContain("t('home.navTutorial')")
    expect(componentSource).toContain('to="/models"')
    expect(componentSource).toContain("t('home.navModels')")
    expect(componentSource).toContain('hasSupportButton')
    expect(componentSource).toContain('hasSupportContent')
    expect(componentSource).toContain('@click="openSupportPopup"')
  })

  it('mounts the header announcement carousel between page title and actions', () => {
    expect(componentSource).toContain('HeaderAnnouncementCarousel')
    expect(componentSource).toContain('<AnnouncementBell v-if="user" ref="announcementBellRef" />')
    expect(componentSource).toContain(':announcements="headerAnnouncements"')
    expect(componentSource).toContain('@select="openAnnouncementDetail"')
    expect(componentSource).toContain('announcementBellRef.value?.openDetail(announcement)')
  })

  it('uses the configured Pixel Cafe title for the GroupBuy header when enabled', () => {
    expect(componentSource).toContain("import { resolveGroupBuyProductName, resolvePixelCafeTitle } from '@/utils/groupBuyProduct'")
    expect(componentSource).toContain('appStore.cachedPublicSettings?.pixel_cafe_enabled === true')
    expect(componentSource).toContain('resolvePixelCafeTitle(appStore.cachedPublicSettings)')
  })
})
