<template>
  <header class="public-top-shell">
    <nav class="public-top-nav">
      <router-link to="/home" class="public-brand">
        <span class="min-w-0">
          <span class="block truncate text-sm font-black tracking-normal sm:text-base">
            {{ siteName }}
          </span>
        </span>
      </router-link>

      <div class="public-nav-center">
        <router-link
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="public-nav-link"
          :class="{ 'router-link-active': isNavItemActive(item) }"
        >
          {{ item.label }}
        </router-link>
        <button
          type="button"
          class="public-nav-link public-nav-action-pill"
          @click="openSupportPopup"
        >
          {{ t('home.navContact') }}
        </button>
      </div>

      <div class="public-nav-actions">
        <LocaleSwitcher class="public-locale-switcher" />

        <router-link v-if="!isAuthenticated" to="/login" class="public-auth-button public-login-button">
          <span>{{ t('home.login') }}</span>
        </router-link>

        <router-link v-if="showDashboardButton" :to="dashboardPath" class="public-auth-button">
          <span>{{ t('home.goToDashboard') }}</span>
          <Icon name="arrowRight" size="xs" aria-hidden="true" />
        </router-link>

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="public-icon-button"
          :title="t('home.viewDocs')"
        >
          <Icon name="book" size="sm" aria-hidden="true" />
        </a>

      </div>
    </nav>

    <div class="public-mobile-nav public-page-shell mx-auto">
      <router-link
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="public-nav-link"
        :class="{ 'router-link-active': isNavItemActive(item) }"
      >
        {{ item.label }}
      </router-link>
      <button
        type="button"
        class="public-nav-link public-nav-action-pill"
        @click="openSupportPopup"
      >
        {{ t('home.navContact') }}
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { openSupportPopup } from '@/utils/supportPopup'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const { t } = useI18n()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const isHomeRoute = computed(() => route.path === '/home' || route.path === '/')
const showDashboardButton = computed(() => isAuthenticated.value && !isHomeRoute.value)

const navItems = computed<Array<{ to: string; label: string; activePaths: string[] }>>(() => [
  { to: '/home', label: t('home.navHome'), activePaths: ['/home', '/'] },
  { to: '/tutorial', label: t('home.navTutorial'), activePaths: ['/tutorial'] },
  { to: '/models', label: t('home.navModels'), activePaths: ['/models'] }
])

function isNavItemActive(item: { activePaths: string[] }): boolean {
  return item.activePaths.some((path) => route.path === path || (path !== '/' && route.path.startsWith(`${path}/`)))
}
</script>

<style scoped>
.public-top-shell {
  --public-nav-height: 3.75rem;
  --public-border: #e6dfd8;
  --public-border-strong: #d8cec2;
  --public-surface-soft: rgba(250, 249, 245, 0.74);
  --public-surface-hover: #efe9de;
  --public-surface-raised: rgba(250, 249, 245, 0.9);
  --public-muted: #6c6a64;
  --public-ring: rgba(204, 120, 92, 0.24);
  --public-shadow-soft: 0 8px 20px rgba(20, 20, 19, 0.05);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 80;
  width: 100%;
  border-bottom: 1px solid rgba(216, 206, 194, 0.62);
  background: rgba(250, 249, 245, 0.68);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 1px 0 rgba(20, 20, 19, 0.025),
    0 18px 42px rgba(20, 20, 19, 0.055);
  backdrop-filter: blur(22px) saturate(1.18);
  -webkit-backdrop-filter: blur(22px) saturate(1.18);
  font-family: var(--public-font-sans, Inter, "Noto Sans SC", "Source Han Sans SC", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", "Microsoft YaHei UI", sans-serif);
  font-weight: 400;
  -webkit-font-smoothing: antialiased;
  text-rendering: geometricPrecision;
}

.public-top-nav {
  position: relative;
  display: flex;
  width: 100%;
  min-height: var(--public-nav-height);
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.5rem clamp(1rem, 3.5vw, 3.5rem);
}

.public-page-shell {
  width: min(100%, 72rem);
}

.public-brand {
  display: inline-flex;
  flex: 1 1 0;
  min-width: 0;
  align-items: center;
  color: #141413;
}

.public-nav-center {
  position: absolute;
  left: 50%;
  top: 50%;
  display: inline-flex;
  align-items: center;
  gap: 1.25rem;
  transform: translate(-50%, -50%);
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
  backdrop-filter: none;
}

.public-nav-link {
  position: relative;
  display: inline-flex;
  align-items: center;
  min-height: 2.1rem;
  border-radius: 0;
  padding: 0.25rem 0;
  color: var(--public-muted);
  font-size: 0.82rem;
  font-weight: 450;
  transition:
    color 150ms ease,
    opacity 150ms ease;
}

.public-nav-link::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0.15rem;
  height: 1px;
  background: currentColor;
  opacity: 0;
  transform: scaleX(0.25);
  transform-origin: center;
  transition:
    opacity 150ms ease,
    transform 150ms ease;
}

.public-nav-link:hover,
.public-nav-link.router-link-active {
  color: #141413;
}

.public-nav-link:hover::after,
.public-nav-link.router-link-active::after {
  opacity: 1;
  transform: scaleX(1);
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
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.72);
  box-shadow: none;
  color: #141413;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.public-auth-button {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: 0.36rem;
  border: 1px solid #141413;
  border-radius: 999px;
  background: #141413;
  color: #fffaf5;
  font-size: 0.8rem;
  font-weight: 700;
  padding: 0.35rem 0.85rem;
  transition:
    background-color 150ms ease,
    border-color 150ms ease,
    transform 150ms ease;
}

.public-auth-button:hover {
  border-color: #2a2926;
  background: #2a2926;
  transform: translateY(-1px);
}

.public-icon-button:hover {
  border-color: rgba(204, 120, 92, 0.42);
  background: var(--public-surface-hover);
  color: #a9583e;
  transform: translateY(-1px);
}

.public-icon-button:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.public-locale-switcher :deep(button) {
  min-height: 2.35rem;
  border-radius: 999px;
  padding: 0.35rem 0.45rem;
  color: #6c6a64;
  font-size: 0.75rem;
  font-weight: 500;
}

.public-locale-switcher :deep(button:hover) {
  background: var(--public-surface-hover);
  color: #141413;
}

.public-locale-switcher :deep(.absolute) {
  border-radius: 10px;
  border-color: var(--public-border-strong);
  background: #faf9f5;
  color: #141413;
  backdrop-filter: blur(16px);
  box-shadow: var(--public-shadow);
}

.public-locale-switcher :deep(.absolute button) {
  background: transparent;
  color: #6c6a64;
}

.public-locale-switcher :deep(.absolute button:hover) {
  background: #efe9de;
  color: #141413;
}

.public-locale-switcher :deep(.absolute button[class*="bg-primary-50"]) {
  border-left: 2px solid #cc785c;
  background: #efe9de;
  color: #a9583e;
}

.public-locale-switcher :deep(.absolute button[class*="bg-primary-50"] svg) {
  color: #cc785c;
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

.public-mobile-nav .public-nav-link {
  flex: 0 0 auto;
  border: 1px solid var(--public-border);
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.68);
  padding: 0.26rem 0.62rem;
  backdrop-filter: blur(18px);
}

.public-mobile-nav .public-nav-link::after {
  display: none;
}

@media (max-width: 1023px) {
  .public-nav-center {
    display: none;
  }

  .public-mobile-nav {
    display: flex;
    background: rgba(250, 249, 245, 0.52);
    backdrop-filter: blur(18px) saturate(1.12);
    -webkit-backdrop-filter: blur(18px) saturate(1.12);
  }
}

@media (max-width: 640px) {
  .public-top-nav {
    padding: 0.5rem 0.7rem;
  }

  .public-icon-button {
    height: 2.2rem;
    width: 2.2rem;
  }

  .public-auth-button {
    min-height: 2.5rem;
    padding: 0.3rem 0.68rem;
  }

  .public-mobile-nav .public-nav-link {
    min-height: 2.5rem;
  }

  .public-login-button {
    padding-inline: 0.62rem;
  }

}
</style>
