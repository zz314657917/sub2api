<template>
  <header class="public-top-shell">
    <nav class="public-top-nav public-page-shell mx-auto">
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
          :class="{ 'router-link-active': item.activePaths.includes(route.path) }"
        >
          <PixelIcon :name="item.icon" size="xs" />
          {{ item.label }}
        </router-link>
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

        <button
          type="button"
          class="public-icon-button"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <PixelIcon v-if="isDark" name="sun" size="sm" />
          <PixelIcon v-else name="moon" size="sm" />
        </button>

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
        :class="{ 'router-link-active': item.activePaths.includes(route.path) }"
      >
        <PixelIcon :name="item.icon" size="xs" />
        {{ item.label }}
      </router-link>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import PixelIcon from '@/components/icons/PixelIcon.vue'
import type { PixelIconName } from '@/components/icons/pixelIconTypes'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const { t } = useI18n()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const isDark = ref(
  typeof document !== 'undefined'
    ? document.documentElement.classList.contains('dark')
    : false
)

const navItems = computed<Array<{ to: string; label: string; icon: PixelIconName; activePaths: string[] }>>(() => [
  { to: '/home', label: t('home.navHome'), icon: 'panel', activePaths: ['/home', '/'] },
  { to: '/tutorial', label: t('home.navTutorial'), icon: 'book', activePaths: ['/tutorial'] },
  { to: '/models', label: t('home.navModels'), icon: 'cube', activePaths: ['/models'] }
])

function toggleTheme(): void {
  isDark.value = !isDark.value
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>

<style scoped>
.public-top-shell {
  position: sticky;
  top: 0;
  z-index: 40;
  width: 100%;
  border-bottom: 1px solid rgba(220, 215, 255, 0.1);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.055), rgba(255, 255, 255, 0.018)),
    rgba(4, 3, 16, 0.88);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    0 16px 42px rgba(3, 2, 12, 0.22);
  backdrop-filter: blur(18px);
}

.public-top-nav {
  display: flex;
  width: 100%;
  min-height: 3.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.45rem 1rem;
}

.public-page-shell {
  width: min(100%, 72rem);
}

.public-brand {
  display: inline-flex;
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
  border: 1px solid rgba(180, 189, 255, 0.34);
  background:
    linear-gradient(180deg, rgba(118, 96, 210, 0.48), rgba(35, 29, 84, 0.56)),
    rgba(255, 255, 255, 0.06);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 0 0 1px rgba(9, 6, 32, 0.4),
    0 8px 18px rgba(6, 4, 24, 0.28);
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
  border: 1px solid rgba(229, 224, 255, 0.13);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.095), rgba(255, 255, 255, 0.035)),
    rgba(7, 8, 22, 0.48);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.1),
    0 14px 32px rgba(3, 2, 12, 0.2);
  padding: 0.18rem;
  backdrop-filter: blur(18px);
}

.public-nav-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  min-height: 2rem;
  padding: 0.25rem 0.62rem;
  color: rgba(226, 221, 247, 0.66);
  font-size: 0.78rem;
  font-weight: 800;
  transition:
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.public-nav-pill:hover,
.public-nav-pill.router-link-active {
  background: rgba(255, 255, 255, 0.075);
  color: white;
  box-shadow: inset 0 0 0 1px rgba(229, 224, 255, 0.12);
}

.public-nav-pill .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.78);
  --pixel-glyph-accent: rgba(127, 255, 167, 0.72);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.public-nav-actions {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.public-icon-button {
  display: inline-flex;
  height: 2.35rem;
  width: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(229, 224, 255, 0.2);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.045)),
    rgba(97, 79, 171, 0.34);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    inset 0 -2px 0 rgba(0, 0, 0, 0.2),
    0 7px 16px rgba(5, 3, 18, 0.22);
  color: rgba(255, 255, 255, 0.82);
  transition: all 160ms ease;
}

.public-icon-button:hover {
  border-color: rgba(255, 255, 255, 0.38);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.06)),
    rgba(118, 97, 196, 0.46);
  color: white;
}

.public-icon-button .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.86);
  --pixel-glyph-accent: rgba(174, 183, 214, 0.86);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.public-locale-switcher :deep(button) {
  min-height: 2.35rem;
  border-radius: 0;
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
  border-radius: 0;
  border-color: rgba(206, 198, 255, 0.22);
  background: rgba(18, 12, 45, 0.94);
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
  border: 2px solid #153c1e;
  background: linear-gradient(#6fbf43, #328033);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 3px 0 #123118,
    0 8px 18px rgba(5, 3, 18, 0.24);
  padding: 0.35rem 0.9rem;
  color: white;
  font-size: 0.75rem;
  font-weight: 800;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.35);
  transition:
    transform 120ms ease,
    filter 120ms ease,
    box-shadow 120ms ease;
}

.public-nav-button:hover {
  filter: brightness(1.06);
  transform: translateY(-1px);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 4px 0 #123118,
    0 10px 20px rgba(5, 3, 18, 0.28);
}

.public-nav-button:active {
  transform: translateY(2px);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.2),
    inset 0 -2px 0 rgba(0, 0, 0, 0.22),
    0 1px 0 #123118,
    0 6px 14px rgba(5, 3, 18, 0.22);
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
  border: 1px solid rgba(229, 224, 255, 0.13);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.095), rgba(255, 255, 255, 0.035)),
    rgba(7, 8, 22, 0.48);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1);
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
