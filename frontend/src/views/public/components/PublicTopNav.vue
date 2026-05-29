<template>
  <header class="public-top-shell">
    <nav class="public-top-nav">
      <router-link to="/home" class="public-brand">
        <span class="public-brand-logo">
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </span>
        <span class="min-w-0">
          <span class="block truncate text-sm font-black tracking-normal text-white sm:text-base">
            {{ siteName }}
          </span>
        </span>
      </router-link>

      <div class="public-nav-center">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="public-nav-pill"
          :class="{ 'router-link-active': isNavItemActive(item) }"
        >
          <PixelIcon :name="item.icon" size="xs" />
          {{ item.label }}
        </router-link>
        <button
          v-if="hasSupportButton"
          type="button"
          class="public-nav-pill public-nav-action-pill"
          @click="openSupportPopup"
        >
          <PixelIcon name="support" size="xs" />
          {{ t('home.navContact') }}
        </button>
      </div>

      <div class="public-nav-actions">
        <LocaleSwitcher class="public-locale-switcher" />

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="public-icon-button"
          :title="t('home.viewDocs')"
        >
          <PixelIcon name="book" size="sm" />
        </a>

        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="public-nav-button">
          {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
        </router-link>
      </div>
    </nav>

    <div class="public-mobile-nav public-page-shell mx-auto">
      <router-link
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="public-nav-pill"
        :class="{ 'router-link-active': isNavItemActive(item) }"
      >
        <PixelIcon :name="item.icon" size="xs" />
        {{ item.label }}
      </router-link>
      <button
        v-if="hasSupportButton"
        type="button"
        class="public-nav-pill public-nav-action-pill"
        @click="openSupportPopup"
      >
        <PixelIcon name="support" size="xs" />
        {{ t('home.navContact') }}
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import PixelIcon from '@/components/icons/PixelIcon.vue'
import type { PixelIconName } from '@/components/icons/pixelIconTypes'
import { openSupportPopup } from '@/utils/supportPopup'
import { hasSupportContent } from '@/utils/supportContent'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const { t } = useI18n()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const hasSupportButton = computed(() =>
  hasSupportContent(appStore.cachedPublicSettings, contactInfo.value)
)

const navItems = computed<Array<{ to: string; label: string; icon: PixelIconName; activePaths: string[] }>>(() => [
  { to: '/home', label: t('home.navHome'), icon: 'panel', activePaths: ['/home', '/'] },
  { to: '/tutorial', label: t('home.navTutorial'), icon: 'book', activePaths: ['/tutorial'] },
  { to: '/models', label: t('home.navModels'), icon: 'cube', activePaths: ['/models'] }
])

function isNavItemActive(item: { activePaths: string[] }): boolean {
  return item.activePaths.some((path) => route.path === path || (path !== '/' && route.path.startsWith(`${path}/`)))
}
</script>

<style scoped>
.public-top-shell {
  --public-border: rgba(226, 232, 240, 0.14);
  --public-border-strong: rgba(226, 232, 240, 0.22);
  --public-surface-soft: rgba(255, 255, 255, 0.07);
  --public-surface-hover: rgba(255, 255, 255, 0.11);
  --public-surface-raised:
    linear-gradient(180deg, rgba(255, 255, 255, 0.105), rgba(255, 255, 255, 0.055)),
    rgba(6, 13, 18, 0.56);
  --public-muted: rgba(222, 232, 255, 0.66);
  --public-ring: rgba(119, 255, 173, 0.32);
  --public-shadow-soft: 0 10px 24px rgba(0, 0, 0, 0.14);
  position: sticky;
  top: 0;
  z-index: 40;
  width: 100%;
  border-bottom: 1px solid var(--public-border);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.02)),
    rgba(4, 3, 16, 0.88);
  box-shadow: 0 12px 30px rgba(3, 7, 18, 0.18);
  backdrop-filter: blur(18px);
}

.public-top-nav {
  position: relative;
  display: flex;
  width: 100%;
  min-height: 3.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.45rem clamp(1rem, 3vw, 3rem);
}

.public-page-shell {
  width: min(100%, 72rem);
}

.public-brand {
  display: inline-flex;
  flex: 1 1 0;
  min-width: 0;
  align-items: center;
  gap: 0.78rem;
  color: white;
}

.public-brand-logo {
  display: inline-flex;
  height: 2.2rem;
  width: 2.2rem;
  flex: 0 0 2.2rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--public-border-strong);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.14), rgba(99, 102, 241, 0.1)),
    var(--public-surface-soft);
  box-shadow: var(--public-shadow-soft);
  padding: 0.22rem;
}

.public-nav-center {
  position: absolute;
  left: 50%;
  top: 50%;
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  transform: translate(-50%, -50%);
  border: 1px solid var(--public-border);
  border-radius: 8px;
  background: var(--public-surface-raised);
  box-shadow: var(--public-shadow-soft);
  padding: 0.18rem;
  backdrop-filter: blur(18px);
}

.public-nav-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  min-height: 2rem;
  border-radius: 6px;
  padding: 0.25rem 0.62rem;
  color: var(--public-muted);
  font-size: 0.78rem;
  font-weight: 800;
  transition:
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.public-nav-pill:hover,
.public-nav-pill.router-link-active {
  background: var(--public-surface-hover);
  color: white;
  box-shadow: inset 0 0 0 1px var(--public-border);
}

.public-nav-action-pill {
  border: 0;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}

.public-nav-action-pill:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.public-nav-pill .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.78);
  --pixel-glyph-accent: rgba(127, 255, 167, 0.72);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.public-nav-actions {
  display: flex;
  flex: 1 1 0;
  align-items: center;
  justify-content: flex-end;
  gap: 0.4rem;
}

.public-icon-button {
  display: inline-flex;
  height: 2.35rem;
  width: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--public-border-strong);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.105), rgba(255, 255, 255, 0.04)),
    var(--public-surface-soft);
  box-shadow: var(--public-shadow-soft);
  color: rgba(255, 255, 255, 0.82);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.public-icon-button:hover {
  border-color: rgba(119, 255, 173, 0.34);
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.16), rgba(20, 184, 166, 0.06)),
    var(--public-surface-hover);
  color: white;
  transform: translateY(-1px);
}

.public-icon-button:focus-visible,
.public-nav-button:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.public-icon-button .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.86);
  --pixel-glyph-accent: rgba(174, 183, 214, 0.86);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.public-locale-switcher :deep(button) {
  min-height: 2.35rem;
  border-radius: 8px;
  padding: 0.35rem 0.45rem;
  color: rgba(232, 228, 255, 0.62);
  font-size: 0.75rem;
  font-weight: 800;
}

.public-locale-switcher :deep(button:hover) {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

.public-locale-switcher :deep(.absolute) {
  border-radius: 8px;
  border-color: var(--public-border-strong);
  background: rgba(8, 15, 29, 0.96);
  color: white;
  backdrop-filter: blur(16px);
}

.public-locale-switcher :deep(.absolute button) {
  background: transparent;
  color: rgba(237, 234, 255, 0.82);
}

.public-locale-switcher :deep(.absolute button:hover) {
  background: rgba(124, 102, 214, 0.24);
  color: white;
}

.public-locale-switcher :deep(.absolute button[class*="bg-primary-50"]) {
  border-left: 2px solid rgba(112, 255, 169, 0.84);
  background:
    linear-gradient(90deg, rgba(56, 172, 104, 0.24), rgba(66, 51, 143, 0.18)),
    rgba(24, 18, 58, 0.96);
  color: #f1fff5;
}

.public-locale-switcher :deep(.absolute button[class*="bg-primary-50"] svg) {
  color: rgba(112, 255, 169, 0.92);
}

.public-nav-button {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(119, 255, 173, 0.34);
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.18), rgba(20, 184, 166, 0.08)),
    rgba(5, 15, 18, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    var(--public-shadow-soft);
  padding: 0.35rem 0.9rem;
  color: #eafff0;
  font-size: 0.75rem;
  font-weight: 800;
  text-shadow: none;
  transition:
    transform 120ms ease,
    border-color 120ms ease,
    background 120ms ease,
    filter 120ms ease,
    box-shadow 120ms ease;
}

.public-nav-button:hover {
  border-color: rgba(119, 255, 173, 0.58);
  background:
    linear-gradient(180deg, rgba(119, 255, 173, 0.26), rgba(20, 184, 166, 0.13)),
    rgba(6, 28, 24, 0.86);
  filter: brightness(1.06);
  transform: translateY(-1px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.14),
    0 12px 24px rgba(0, 0, 0, 0.24);
}

.public-nav-button:active {
  transform: translateY(1px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.1),
    0 6px 14px rgba(0, 0, 0, 0.2);
}

.public-mobile-nav {
  display: none;
  gap: 0.28rem;
  overflow-x: auto;
  padding: 0 0.75rem 0.48rem;
  scrollbar-width: none;
}

.public-mobile-nav::-webkit-scrollbar {
  display: none;
}

.public-mobile-nav .public-nav-pill {
  flex: 0 0 auto;
  border: 1px solid var(--public-border);
  background: var(--public-surface-raised);
  backdrop-filter: blur(18px);
}

@media (max-width: 1023px) {
  .public-nav-center {
    display: none;
  }

  .public-mobile-nav {
    display: flex;
  }
}

@media (max-width: 640px) {
  .public-top-nav {
    padding: 0.5rem 0.7rem;
  }

  .public-brand {
    gap: 0.55rem;
  }

  .public-brand-logo {
    height: 2rem;
    width: 2rem;
    flex-basis: 2rem;
  }

  .public-icon-button {
    height: 2.2rem;
    width: 2.2rem;
  }

  .public-nav-button {
    min-height: 2.2rem;
    padding: 0.3rem 0.65rem;
  }
}
</style>
