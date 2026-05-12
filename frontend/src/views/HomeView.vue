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
  <div v-else class="home-page-root home-violet-bg relative flex min-h-screen flex-col overflow-hidden text-white">
    <div class="home-matrix-rain pointer-events-none absolute inset-0" aria-hidden="true">
      <span
        v-for="column in matrixColumns"
        :key="column.left"
        class="home-matrix-column"
        :style="{
          left: column.left,
          animationDelay: column.delay,
          animationDuration: column.duration
        }"
      >
        {{ column.text }}
      </span>
    </div>
    <div class="home-blur-field pointer-events-none absolute inset-0"></div>
    <div class="home-noise pointer-events-none absolute inset-0"></div>

    <PublicTopNav />

    <main class="home-main-stage relative z-10 flex px-4 py-4 sm:px-6 sm:py-5 lg:py-5">
      <section class="home-hero-shell mx-auto flex flex-col justify-center gap-5 sm:gap-6">
        <div class="home-hero-content mx-auto flex w-full flex-col items-center text-center">
          <div class="home-kicker mb-4">
            <span class="home-kicker-dot"></span>
            {{ t('home.heroEyebrow') }}
          </div>

          <h1
            class="home-title-sweep text-[2.95rem] font-black leading-[0.98] tracking-normal text-white sm:text-[4.45rem] lg:text-[5.75rem]"
          >
            <span class="home-title-line home-title-fill block">{{ t('home.heroTitleTop') }}</span>
            <span class="home-title-line block">{{ t('home.heroTitleBottom') }}</span>
          </h1>

          <p
            class="home-hero-description mt-4 max-w-2xl text-base leading-7 text-violet-100/88 sm:mt-5 sm:text-lg sm:leading-8"
          >
            <span class="home-typewriter" :aria-label="currentHeroDescription">
              <span class="home-typewriter-text">{{ typedHeroDescription }}</span>
              <span class="home-typewriter-cursor" aria-hidden="true"></span>
            </span>
          </p>

          <div class="mt-5 flex flex-col items-center gap-3 sm:mt-6 sm:flex-row">
            <router-link :to="isAuthenticated ? dashboardPath : '/register'" class="home-claim-button">
              <span class="home-button-inner">
                <PixelIcon :name="primaryActionIcon" size="sm" />
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.claimButton') }}
              </span>
            </router-link>
          </div>

        </div>

      </section>
    </main>

    <footer class="home-footer relative z-10 mx-auto px-4 pb-6 sm:px-6">
      <div class="home-footer-inner mx-auto">
        <div class="home-footer-models" aria-label="Supported models">
          <span
            v-for="item in footerModelBadges"
            :key="item.label"
            class="home-footer-badge"
          >
            <span class="home-footer-badge-dot" :style="{ backgroundColor: item.color }"></span>
            {{ item.label }}
          </span>
        </div>

        <div class="home-footer-bar">
          <p class="home-footer-copy">
            &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
          </p>
          <nav class="home-footer-links" :aria-label="t('home.footer.linksLabel')">
            <router-link to="/tutorial">{{ t('home.navTutorial') }}</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
              {{ t('home.docs') }}
            </a>
            <router-link to="/legal/terms">{{ t('home.footer.terms') }}</router-link>
            <router-link to="/legal/privacy">{{ t('home.footer.privacy') }}</router-link>
            <router-link to="/legal/usage-policy">{{ t('home.footer.usagePolicy') }}</router-link>
          </nav>
        </div>
      </div>
    </footer>

    <a
      v-if="supportHref && !hasSupportPopupItems"
      :href="supportHref"
      target="_blank"
      rel="noopener noreferrer"
      class="home-support-button"
    >
      <PixelIcon name="support" size="sm" />
      {{ t('home.contactSupport') }}
    </a>

    <button
      v-else-if="hasSupportPopupItems"
      type="button"
      class="home-support-button"
      @click="openSupportPopup"
    >
      <PixelIcon name="support" size="sm" />
      {{ t('home.contactSupport') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import PixelIcon from '@/components/icons/PixelIcon.vue'
import type { PixelIconName } from '@/components/icons/pixelIconTypes'
import PublicTopNav from './public/components/PublicTopNav.vue'
import { openSupportPopup } from '@/utils/supportPopup'

type PrimaryActionIconName = Extract<PixelIconName, 'dashboard' | 'gift'>

type MatrixColumn = {
  left: string
  delay: string
  duration: string
  text: string
}

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const currentYear = new Date().getFullYear()

const matrixSeeds = [
  '01APIKEYCHATGPTCODEX',
  'MODELROUTER0101TOKEN',
  'PROXYDIRECTGPT01010',
  'TEAMKEYCODE110010',
  'OPENAIACCESS10101',
  'STREAMJSON01011',
  'LATENCYLOW00110',
  'QUOTAUSAGE10100'
]

const matrixColumnCount = 51

const matrixColumns: MatrixColumn[] = Array.from({ length: matrixColumnCount }, (_, index) => ({
  left: `${(index * 2.08) % 101}%`,
  delay: `${-(index % 13) * 0.64}s`,
  duration: `${8 + (index % 9) * 0.92}s`,
  text: matrixSeeds[index % matrixSeeds.length]
}))

const footerModelBadges = [
  { label: 'Claude', color: '#f97316' },
  { label: 'GPT', color: '#22c55e' },
  { label: 'Gemini', color: '#38bdf8' },
  { label: 'DeepSeek', color: '#a78bfa' },
  { label: 'Qwen', color: '#818cf8' }
]

const heroDescriptionTexts = computed(() =>
  [
    t('home.heroDescription'),
    t('home.heroDescriptionAltModels'),
    t('home.heroDescriptionAltSupport')
  ]
    .map((text) => text.trim())
    .filter(Boolean)
)

const typedHeroDescription = ref('')
const currentHeroDescription = computed(() => heroDescriptionTexts.value[heroDescriptionIndex] || '')
const prefersReducedHeroMotion = ref(false)
let heroDescriptionIndex = 0
let heroDescriptionCharIndex = 0
let heroDescriptionDeleting = false
let heroDescriptionStarted = false
let heroDescriptionTimer: number | null = null
let reducedMotionQuery: MediaQueryList | null = null

function clearHeroDescriptionTimer() {
  if (heroDescriptionTimer !== null) {
    window.clearTimeout(heroDescriptionTimer)
    heroDescriptionTimer = null
  }
}

function scheduleHeroDescriptionTick(delay: number) {
  clearHeroDescriptionTimer()
  heroDescriptionTimer = window.setTimeout(runHeroDescriptionTick, delay)
}

function resetHeroDescriptionTyping() {
  clearHeroDescriptionTimer()
  const texts = heroDescriptionTexts.value
  heroDescriptionIndex = 0
  heroDescriptionCharIndex = 0
  heroDescriptionDeleting = false

  if (texts.length === 0) {
    typedHeroDescription.value = ''
    return
  }

  if (prefersReducedHeroMotion.value || texts.length === 1) {
    typedHeroDescription.value = texts[0]
    return
  }

  typedHeroDescription.value = texts[0]
  heroDescriptionCharIndex = texts[0].length
  heroDescriptionDeleting = true
  scheduleHeroDescriptionTick(1800)
}

function runHeroDescriptionTick() {
  const texts = heroDescriptionTexts.value
  if (texts.length === 0) {
    typedHeroDescription.value = ''
    return
  }

  if (prefersReducedHeroMotion.value || texts.length === 1) {
    typedHeroDescription.value = texts[heroDescriptionIndex] || texts[0]
    return
  }

  const fullText = texts[heroDescriptionIndex] || texts[0]

  if (!heroDescriptionDeleting) {
    heroDescriptionCharIndex = Math.min(fullText.length, heroDescriptionCharIndex + 1)
    typedHeroDescription.value = fullText.slice(0, heroDescriptionCharIndex)

    if (heroDescriptionCharIndex < fullText.length) {
      scheduleHeroDescriptionTick(62)
      return
    }

    heroDescriptionDeleting = true
    scheduleHeroDescriptionTick(1800)
    return
  }

  heroDescriptionCharIndex = Math.max(0, heroDescriptionCharIndex - 1)
  typedHeroDescription.value = fullText.slice(0, heroDescriptionCharIndex)

  if (heroDescriptionCharIndex > 0) {
    scheduleHeroDescriptionTick(28)
    return
  }

  heroDescriptionDeleting = false
  heroDescriptionIndex = (heroDescriptionIndex + 1) % texts.length
  scheduleHeroDescriptionTick(320)
}

function handleReducedMotionChange(event: MediaQueryListEvent) {
  prefersReducedHeroMotion.value = event.matches
  resetHeroDescriptionTyping()
}

function initHeroDescriptionTyping() {
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
    reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    prefersReducedHeroMotion.value = reducedMotionQuery.matches
    reducedMotionQuery.addEventListener('change', handleReducedMotionChange)
  }

  heroDescriptionStarted = true
  resetHeroDescriptionTyping()
}

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const supportHref = computed(() => normalizeSupportLink(contactInfo.value))
const hasSupportPopupItems = computed(() => {
  const items = appStore.cachedPublicSettings?.support_popup_items
  return Array.isArray(items) && items.some((item) => item.title?.trim() && item.image_url?.trim())
})

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

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const primaryActionIcon = computed<PrimaryActionIconName>(() =>
  isAuthenticated.value ? 'dashboard' : 'gift'
)

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  initHeroDescriptionTyping()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

onBeforeUnmount(() => {
  clearHeroDescriptionTimer()
  reducedMotionQuery?.removeEventListener('change', handleReducedMotionChange)
})

watch(heroDescriptionTexts, () => {
  if (heroDescriptionStarted) {
    resetHeroDescriptionTyping()
  }
})
</script>

<style scoped>
.home-violet-bg {
  background:
    radial-gradient(circle at 50% 38%, rgba(74, 255, 147, 0.13) 0, transparent 34%),
    radial-gradient(circle at 18% 20%, rgba(105, 86, 255, 0.16) 0, transparent 30%),
    radial-gradient(circle at 82% 28%, rgba(88, 255, 158, 0.12) 0, transparent 30%),
    radial-gradient(circle at 48% 88%, rgba(62, 215, 255, 0.09) 0, transparent 32%),
    linear-gradient(180deg, #050914 0%, #08110f 48%, #03060a 100%);
}

.home-page-root {
  min-height: 100vh;
  min-height: 100svh;
}

.home-matrix-rain {
  overflow: hidden;
  opacity: 0.62;
  mix-blend-mode: screen;
  mask-image:
    linear-gradient(to bottom, transparent 0%, rgba(0, 0, 0, 0.92) 10%, rgba(0, 0, 0, 0.74) 46%, rgba(0, 0, 0, 0.22) 78%, transparent 100%),
    radial-gradient(circle at 50% 32%, transparent 0%, rgba(0, 0, 0, 0.78) 38%, rgba(0, 0, 0, 0.98) 100%);
  -webkit-mask-image:
    linear-gradient(to bottom, transparent 0%, rgba(0, 0, 0, 0.92) 10%, rgba(0, 0, 0, 0.74) 46%, rgba(0, 0, 0, 0.22) 78%, transparent 100%),
    radial-gradient(circle at 50% 32%, transparent 0%, rgba(0, 0, 0, 0.78) 38%, rgba(0, 0, 0, 0.98) 100%);
}

.home-matrix-rain::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 50% 40%, rgba(64, 255, 145, 0.1), transparent 38%),
    linear-gradient(90deg, transparent, rgba(87, 255, 151, 0.06), transparent);
  filter: blur(18px);
}

.home-matrix-column {
  position: absolute;
  top: -80vh;
  display: block;
  width: 1ch;
  background: linear-gradient(
    to bottom,
    rgba(202, 255, 212, 0.96) 0%,
    rgba(60, 255, 91, 0.82) 18%,
    rgba(42, 225, 68, 0.48) 56%,
    rgba(16, 145, 41, 0.14) 100%
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1.05;
  text-shadow:
    0 0 7px rgba(89, 255, 146, 0.62),
    0 0 18px rgba(24, 206, 94, 0.24);
  white-space: normal;
  word-break: break-all;
  animation: matrix-rain-fall linear infinite;
}

.home-matrix-column:nth-child(3n) {
  background-image: linear-gradient(
    to bottom,
    rgba(239, 255, 242, 0.98) 0%,
    rgba(82, 255, 105, 0.86) 22%,
    rgba(33, 210, 62, 0.42) 62%,
    rgba(12, 112, 31, 0.1) 100%
  );
  font-size: 0.72rem;
}

.home-matrix-column:nth-child(4n) {
  opacity: 0.56;
}

@keyframes matrix-rain-fall {
  0% {
    transform: translate3d(0, -10vh, 0);
  }
  100% {
    transform: translate3d(0, 190vh, 0);
  }
}

.home-blur-field {
  background:
    radial-gradient(ellipse at 50% 34%, rgba(52, 255, 128, 0.18), transparent 34%),
    radial-gradient(ellipse at 35% 22%, rgba(70, 80, 210, 0.14), transparent 28%),
    radial-gradient(ellipse at 70% 24%, rgba(45, 178, 105, 0.18), transparent 30%),
    radial-gradient(ellipse at 55% 78%, rgba(77, 223, 255, 0.07), transparent 28%);
  filter: blur(52px);
  opacity: 0.86;
  transform: scale(1.05);
}

.home-noise {
  background-image:
    linear-gradient(rgba(102, 255, 161, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(102, 255, 161, 0.028) 1px, transparent 1px);
  background-size: 54px 54px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.55), transparent 78%);
  opacity: 0.36;
}

.home-main-stage {
  flex: 1 1 auto;
  min-height: 0;
}

.home-hero-shell {
  width: min(100%, 72rem);
}

.home-hero-content {
  max-width: 56rem;
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
  max-width: 56rem;
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

.home-hero-description {
  min-height: 2em;
}

.home-typewriter {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: min(100%, 16rem);
}

.home-typewriter-text {
  overflow-wrap: anywhere;
}

.home-typewriter-cursor {
  display: inline-block;
  width: 0.12em;
  height: 1.15em;
  margin-left: 0.18em;
  background: rgba(216, 255, 226, 0.95);
  box-shadow: 0 0 16px rgba(109, 255, 155, 0.58);
  transform: translateY(0.12em);
  animation: hero-cursor-blink 940ms steps(2, start) infinite;
}

@keyframes hero-cursor-blink {
  0%,
  46% {
    opacity: 1;
  }

  47%,
  100% {
    opacity: 0;
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

.home-footer {
  flex: 0 0 auto;
  width: min(100%, 72rem);
}

.home-footer-inner {
  width: 100%;
  border-top: 1px solid rgba(221, 230, 255, 0.14);
  color: rgba(222, 232, 255, 0.62);
}

.home-footer-models {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.7rem 0 0.62rem;
}

.home-footer-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  min-height: 1.75rem;
  border: 1px solid rgba(222, 232, 255, 0.15);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
  padding: 0.28rem 0.72rem;
  color: rgba(235, 241, 255, 0.76);
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1;
  backdrop-filter: blur(14px);
}

.home-footer-badge-dot {
  width: 0.38rem;
  height: 0.38rem;
  border-radius: 999px;
  box-shadow: 0 0 12px currentColor;
}

.home-footer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-top: 1px solid rgba(221, 230, 255, 0.1);
  padding: 0.72rem 0 0;
}

.home-footer-copy {
  font-size: 0.86rem;
  line-height: 1.6;
}

.home-footer-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.9rem;
  font-size: 0.82rem;
  font-weight: 700;
}

.home-footer-links a {
  color: rgba(222, 232, 255, 0.62);
  transition: color 150ms ease;
}

.home-footer-links a:hover {
  color: rgba(255, 255, 255, 0.92);
}

@media (max-width: 640px) {
  .home-page-root {
    overflow-y: auto;
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

  .home-footer-bar {
    align-items: flex-start;
    flex-direction: column;
  }

  .home-footer-links {
    justify-content: flex-start;
  }
}

@media (min-width: 1536px) {
  .home-main-stage {
    align-items: center;
    padding-top: clamp(1rem, 2.4vh, 2rem);
    padding-bottom: clamp(1rem, 2.8vh, 2.4rem);
  }

  .home-hero-shell {
    width: min(100%, 82rem);
  }

  .home-hero-shell {
    gap: clamp(1.6rem, 3vh, 2.4rem);
  }

  .home-hero-content {
    max-width: 64rem;
    padding-top: 0;
  }

  .home-title-sweep {
    max-width: 64rem;
    font-size: clamp(5.45rem, 4.25vw, 6.35rem);
  }

  .home-hero-description {
    max-width: 46rem;
  }

}

@media (min-width: 1920px) {
  .home-hero-shell {
    width: min(100%, 92rem);
  }

  .home-hero-shell {
    gap: clamp(1.9rem, 3.2vh, 2.8rem);
  }

  .home-hero-content {
    max-width: 72rem;
  }

  .home-title-sweep {
    max-width: 72rem;
    font-size: clamp(5.8rem, 4vw, 6.85rem);
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-matrix-column {
    animation: none;
    transform: translate3d(0, 24vh, 0);
  }

  .home-title-fill {
    animation: none;
    background-position: 54% 50%;
  }

  .home-typewriter-cursor {
    animation: none;
  }
}
</style>
