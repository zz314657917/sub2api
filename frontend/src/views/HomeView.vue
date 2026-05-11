<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div v-else class="home-violet-bg relative min-h-screen overflow-hidden text-white">
    <div class="home-blur-field pointer-events-none absolute inset-0"></div>
    <div class="home-noise pointer-events-none absolute inset-0"></div>

    <header class="relative z-20 px-4 pt-4 sm:px-6 sm:pt-5">
      <nav
        class="mx-auto flex h-14 max-w-6xl items-center justify-between border border-white/14 bg-white/[0.07] px-3 shadow-[0_12px_32px_rgba(13,8,35,0.22)] backdrop-blur-xl sm:px-4"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden border border-white/20 bg-white/12 shadow-[inset_0_1px_0_rgba(255,255,255,0.2)]"
          >
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="truncate text-sm font-bold tracking-normal text-white sm:text-base">
            {{ siteName }}
          </span>
        </div>

        <div class="flex items-center gap-1.5 sm:gap-2">
          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="home-icon-button"
            :title="t('home.viewDocs')"
          >
            <PixelIcon name="book" size="sm" />
          </a>

          <button
            @click="toggleTheme"
            class="home-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <PixelIcon v-if="isDark" name="sun" size="sm" />
            <PixelIcon v-else name="moon" size="sm" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="home-nav-button hidden sm:inline-flex"
          >
            {{ t('home.dashboard') }}
          </router-link>
          <router-link v-else to="/login" class="home-nav-button hidden sm:inline-flex">
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 px-4 pb-8 pt-8 sm:px-6 sm:pb-10 sm:pt-9 lg:pt-10">
      <section class="mx-auto flex max-w-6xl flex-col gap-8 sm:gap-9">
        <div class="mx-auto flex w-full max-w-4xl flex-col items-center pt-5 text-center sm:pt-7 lg:pt-8">
          <div class="home-kicker mb-5">
            <span class="home-kicker-dot"></span>
            {{ t('home.heroEyebrow') }}
          </div>

          <h1
            class="home-title-sweep max-w-4xl text-[2.95rem] font-black leading-[0.98] tracking-normal text-white sm:text-[4.45rem] lg:text-[5.75rem]"
          >
            <span class="home-title-line block">{{ t('home.heroTitleTop') }}</span>
            <span class="home-title-line home-title-fill block">{{ t('home.heroTitleBottom') }}</span>
          </h1>

          <p
            class="mt-5 max-w-2xl text-base leading-7 text-violet-100/88 sm:mt-6 sm:text-lg sm:leading-8"
          >
            {{ t('home.heroDescription') }}
          </p>

          <div class="mt-7 flex flex-col items-center gap-3 sm:flex-row">
            <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="home-claim-button">
              <span class="home-button-inner">
                <PixelIcon :name="primaryActionIcon" size="sm" />
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.claimButton') }}
              </span>
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="home-secondary-link"
            >
              {{ t('home.docs') }}
              <PixelIcon name="arrow-right" size="xs" />
            </a>
          </div>

          <div class="mt-5 flex flex-wrap items-center justify-center gap-2.5 text-xs text-violet-50/78 sm:mt-6">
            <span v-for="tag in heroTags" :key="tag.label" class="home-chip">
              <PixelIcon :name="tag.icon" size="xs" />
              {{ tag.label }}
            </span>
          </div>
        </div>

        <div class="grid gap-3 md:grid-cols-3">
          <article v-for="item in featureCards" :key="item.title" class="home-feature-card">
            <div class="home-feature-icon-frame">
              <PixelIcon :name="item.icon" size="md" />
            </div>
            <div>
              <h2 class="text-sm font-bold text-white">{{ item.title }}</h2>
              <p class="mt-1.5 text-xs leading-5 text-violet-100/72">{{ item.description }}</p>
            </div>
          </article>
        </div>
      </section>
    </main>

    <a
      v-if="supportHref"
      :href="supportHref"
      target="_blank"
      rel="noopener noreferrer"
      class="home-support-button"
    >
      <PixelIcon name="support" size="sm" />
      {{ t('home.contactSupport') }}
    </a>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import PixelIcon from '@/components/icons/PixelIcon.vue'
import type { PixelIconName } from '@/components/icons/pixelIconTypes'

type FeatureIconName = Extract<PixelIconName, 'key' | 'shield' | 'usage'>
type HeroTagIconName = Extract<PixelIconName, 'link' | 'team' | 'usage' | 'spark'>
type PrimaryActionIconName = Extract<PixelIconName, 'dashboard' | 'gift'>

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '')

const heroTags = computed<Array<{ icon: HeroTagIconName; label: string }>>(() => [
  { icon: 'link', label: t('home.tags.directConnect') },
  { icon: 'team', label: t('home.tags.teamReady') },
  { icon: 'usage', label: t('home.tags.clearUsage') },
  { icon: 'spark', label: t('home.tags.routing') }
])

const featureCards = computed<Array<{ icon: FeatureIconName; title: string; description: string }>>(() => [
  {
    icon: 'key',
    title: t('home.features.unifiedGateway'),
    description: t('home.features.unifiedGatewayDesc')
  },
  {
    icon: 'shield',
    title: t('home.features.multiAccount'),
    description: t('home.features.multiAccountDesc')
  },
  {
    icon: 'usage',
    title: t('home.features.balanceQuota'),
    description: t('home.features.balanceQuotaDesc')
  }
])

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const supportHref = computed(() => normalizeSupportLink(contactInfo.value))

function normalizeSupportLink(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return ''

  try {
    const url = new URL(trimmed)
    return ['http:', 'https:', 'mailto:', 'tel:'].includes(url.protocol) ? url.href : ''
  } catch {
    return ''
  }
}

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const primaryActionIcon = computed<PrimaryActionIconName>(() =>
  isAuthenticated.value ? 'dashboard' : 'gift'
)

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.home-violet-bg {
  background:
    radial-gradient(circle at 50% 42%, rgba(169, 109, 255, 0.18) 0, transparent 34%),
    radial-gradient(circle at 18% 20%, rgba(105, 86, 255, 0.28) 0, transparent 30%),
    radial-gradient(circle at 82% 28%, rgba(168, 75, 255, 0.22) 0, transparent 28%),
    radial-gradient(circle at 48% 88%, rgba(62, 215, 255, 0.13) 0, transparent 32%),
    linear-gradient(180deg, #120b35 0%, #160932 48%, #080515 100%);
}

.home-blur-field {
  background:
    radial-gradient(ellipse at 50% 34%, rgba(197, 118, 255, 0.3), transparent 36%),
    radial-gradient(ellipse at 35% 22%, rgba(74, 86, 255, 0.24), transparent 28%),
    radial-gradient(ellipse at 70% 24%, rgba(117, 41, 199, 0.3), transparent 30%),
    radial-gradient(ellipse at 55% 78%, rgba(77, 223, 255, 0.11), transparent 28%);
  filter: blur(46px);
  opacity: 0.92;
  transform: scale(1.05);
}

.home-noise {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 54px 54px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.55), transparent 78%);
  opacity: 0.36;
}

.home-icon-button {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.82);
  transition: all 160ms ease;
}

.home-icon-button:hover {
  border-color: rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.14);
  color: white;
}

.home-icon-button .pixel-glyph,
.home-secondary-link .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.86);
  --pixel-glyph-accent: rgba(174, 183, 214, 0.86);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.home-nav-button {
  align-items: center;
  justify-content: center;
  min-height: 2.25rem;
  border: 2px solid #153c1e;
  background: linear-gradient(#6fbf43, #328033);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 4px 0 #123118;
  padding: 0.35rem 0.85rem;
  color: white;
  font-size: 0.75rem;
  font-weight: 800;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.35);
}

.home-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
  padding: 0.45rem 0.8rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: rgba(245, 239, 255, 0.82);
  backdrop-filter: blur(18px);
}

.home-kicker-dot {
  height: 0.45rem;
  width: 0.45rem;
  background: #6dff9b;
  box-shadow: 0 0 16px rgba(109, 255, 155, 0.8);
}

.home-title-sweep {
  position: relative;
  display: inline-block;
  padding: 0.03em 0.06em 0.08em;
}

.home-title-line {
  position: relative;
  z-index: 1;
}

.home-title-fill {
  color: transparent;
  background: linear-gradient(
    98deg,
    #8d67ff 0%,
    #b78dff 20%,
    #f4edff 39%,
    #99d9ff 54%,
    #55c6ff 72%,
    #9b76ff 100%
  );
  background-size: 240% 100%;
  background-position: 0% 50%;
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: title-fill-sweep 7.2s ease-in-out infinite;
}

@keyframes title-fill-sweep {
  0%,
  18% {
    background-position: 0% 50%;
  }
  55% {
    background-position: 100% 50%;
  }
  82%,
  100% {
    background-position: 0% 50%;
  }
}

.home-claim-button {
  display: inline-flex;
  min-width: min(18rem, calc(100vw - 2rem));
  border: 3px solid #5f4814;
  background: #9d6b19;
  box-shadow:
    0 5px 0 #4f360f,
    0 18px 34px rgba(13, 8, 35, 0.34);
  color: #2b1900;
  transition:
    transform 120ms ease,
    box-shadow 120ms ease,
    filter 120ms ease;
}

.home-claim-button:hover {
  filter: brightness(1.04);
  transform: translateY(-1px);
  box-shadow:
    0 6px 0 #4f360f,
    0 20px 38px rgba(13, 8, 35, 0.38);
}

.home-claim-button:active {
  transform: translateY(3px);
  box-shadow:
    0 2px 0 #4f360f,
    0 10px 22px rgba(13, 8, 35, 0.28);
}

.home-button-inner {
  display: inline-flex;
  min-height: 3.15rem;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  border: 2px solid rgba(255, 245, 168, 0.45);
  background:
    linear-gradient(rgba(255, 255, 255, 0.22), transparent 42%),
    linear-gradient(180deg, #ffd95d 0%, #e6a72c 100%);
  padding: 0.8rem 1.3rem;
  font-weight: 900;
  letter-spacing: 0;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.38);
}

.home-button-inner .pixel-glyph {
  --pixel-glyph-on: #2b1900;
  --pixel-glyph-accent: #835b19;
  --pixel-glyph-glow: transparent;
  filter: none;
}

.home-secondary-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: rgba(245, 239, 255, 0.8);
  font-size: 0.875rem;
  font-weight: 700;
  transition: color 150ms ease;
}

.home-secondary-link:hover {
  color: white;
}

.home-support-button {
  position: fixed;
  right: 1.25rem;
  bottom: 1.25rem;
  z-index: 30;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.65rem;
  border: 2px solid #153c1e;
  background: linear-gradient(#6fbf43, #328033);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 4px 0 #123118,
    0 14px 28px rgba(8, 5, 21, 0.32);
  padding: 0.6rem 0.9rem;
  color: white;
  font-size: 0.8rem;
  font-weight: 900;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.35);
  transition:
    transform 120ms ease,
    box-shadow 120ms ease,
    filter 120ms ease;
}

.home-support-button:hover {
  filter: brightness(1.05);
  transform: translateY(-1px);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 5px 0 #123118,
    0 16px 30px rgba(8, 5, 21, 0.36);
}

.home-support-button:active {
  transform: translateY(3px);
  box-shadow:
    inset 0 2px 0 rgba(255, 255, 255, 0.24),
    inset 0 -3px 0 rgba(0, 0, 0, 0.24),
    0 1px 0 #123118,
    0 8px 18px rgba(8, 5, 21, 0.26);
}

.home-support-button .pixel-glyph {
  --pixel-glyph-on: rgba(255, 255, 255, 0.94);
  --pixel-glyph-accent: rgba(205, 231, 214, 0.86);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.home-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.07);
  padding: 0.42rem 0.68rem 0.42rem 0.55rem;
  backdrop-filter: blur(16px);
}

.home-chip .pixel-glyph {
  --pixel-glyph-on: rgba(232, 229, 255, 0.86);
  --pixel-glyph-accent: rgba(174, 183, 214, 0.86);
  --pixel-glyph-glow: transparent;
  filter: none;
}

.home-feature-card {
  display: flex;
  min-height: 6.5rem;
  gap: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.075);
  padding: 1rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.1),
    0 12px 30px rgba(8, 5, 21, 0.24);
  backdrop-filter: blur(18px);
}

.home-feature-icon-frame {
  display: inline-flex;
  width: 2.55rem;
  height: 2.55rem;
  flex: 0 0 2.55rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.09);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12);
}

.home-feature-icon-frame .pixel-glyph {
  --pixel-glyph-on: rgba(238, 236, 255, 0.92);
  --pixel-glyph-accent: rgba(166, 176, 208, 0.92);
  --pixel-glyph-glow: transparent;
  filter: none;
}

@media (max-width: 640px) {
  .home-feature-card {
    min-height: auto;
  }

  .home-button-inner {
    min-height: 3rem;
  }

  .home-support-button {
    right: 1rem;
    bottom: 1rem;
    min-height: 2.5rem;
    padding: 0.55rem 0.75rem;
    font-size: 0.75rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-title-fill {
    animation: none;
    background-position: 54% 50%;
  }
}
</style>
