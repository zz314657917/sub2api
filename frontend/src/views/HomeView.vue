<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="shouldRenderCustomHomeContent" class="min-h-screen">
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
  <div v-else class="home-page-root home-mimo-bg relative flex flex-col overflow-x-hidden">
    <PublicTopNav />

    <main class="home-main-stage relative z-10 overflow-hidden px-4 sm:px-6">
      <PublicRevealBackdrop variant="hero" />
      <section class="home-mimo-hero mx-auto" aria-label="Developer API workspace">
        <div class="home-mimo-copy">
          <h1 class="home-mimo-title" :aria-label="heroHeadline">
            <span class="home-mimo-title-line">{{ heroTitleTop }}</span>
            <span class="home-mimo-title-line home-mimo-title-line-sweep" :data-title="heroTitleBottom">
              {{ heroTitleBottom }}
            </span>
          </h1>

          <p class="home-mimo-description">
            <Transition name="home-subtitle-slide" mode="out-in">
              <span :key="heroDescriptionText" class="home-mimo-description-text">
                {{ heroDescriptionText }}
              </span>
            </Transition>
          </p>

          <button
            type="button"
            class="home-command-pill"
            :aria-label="`${t('home.copyApiEntry')}: ${publicApiEntryUrl}`"
            @click="copyApiEntryUrl"
          >
            <span>{{ t('home.apiEntryLabel') }}</span>
            <code>{{ publicApiEntryUrl }}</code>
            <span class="home-command-copy-state">
              <Icon :name="apiEntryCopied ? 'check' : 'copy'" size="xs" aria-hidden="true" />
              <span>{{ apiEntryCopied ? t('home.apiEntryCopied') : t('home.copyApiEntry') }}</span>
            </span>
          </button>

          <div class="home-mimo-actions">
            <router-link v-if="isAuthenticated" :to="dashboardPath" class="home-action-button is-primary">
              <span class="home-button-inner">
                <span>{{ t('home.goToDashboard') }}</span>
                <Icon name="arrowRight" size="sm" class="home-button-icon" aria-hidden="true" />
              </span>
            </router-link>
            <button v-else type="button" class="home-action-button is-primary" @click="openAuthPanel">
              <span class="home-button-inner">
                <span>{{ t('home.claimButton') }}</span>
                <Icon name="arrowRight" size="sm" class="home-button-icon" aria-hidden="true" />
              </span>
            </button>
            <router-link to="/tutorial/getting-started" class="home-action-button is-secondary">
              <span class="home-button-inner">
                <span>{{ t('home.quickStartButton') }}</span>
              </span>
            </router-link>
          </div>

          <ul class="home-trust-signals" :aria-label="t('home.trustSignalsLabel')">
            <li v-for="item in trustSignals" :key="item">
              <Icon name="check" size="xs" aria-hidden="true" />
              <span>{{ item }}</span>
            </li>
          </ul>
        </div>

        <AuthAccessPanel
          v-if="!isAuthenticated"
          id="home-auth-panel"
          class="home-auth-panel"
          :class="{ 'is-mobile-open': mobileAuthExpanded }"
          embedded
          :initial-mode="authMode"
        />
        <aside v-else class="home-account-workbench" :aria-label="t('home.accountWorkbench.ariaLabel')">
          <div class="home-account-workbench-header">
            <div class="home-account-user-chip" :title="accountWorkbenchUserName">
              <span>{{ accountWorkbenchUserName }}</span>
            </div>
          </div>

          <div class="home-account-balance-card">
            <span>{{ t('home.accountWorkbench.currentBalance') }}</span>
            <strong>{{ accountWorkbenchBalance }}</strong>
          </div>

          <div class="home-account-stat-grid">
            <article
              v-for="card in accountWorkbenchCards"
              :key="card.key"
              class="home-account-stat-card"
            >
              <span>{{ card.label }}</span>
              <strong>{{ card.value }}</strong>
            </article>
          </div>

          <p v-if="accountWorkbenchLoading" class="home-account-workbench-state">
            {{ t('home.accountWorkbench.loading') }}
          </p>
          <p v-else-if="accountWorkbenchError" class="home-account-workbench-state is-error">
            {{ t('home.accountWorkbench.loadFailed') }}
          </p>

          <router-link to="/purchase" class="home-account-workbench-link">
            <span>{{ t('home.accountWorkbench.rechargeBalance') }}</span>
            <Icon name="arrowRight" size="sm" aria-hidden="true" />
          </router-link>
        </aside>
      </section>
    </main>

    <section class="home-gateway-section relative z-10 px-4 sm:px-6" :aria-label="t('home.gatewayExplain.ariaLabel')">
      <div class="home-gateway-shell mx-auto">
        <div class="home-gateway-copy mx-auto">
          <p class="home-gateway-kicker">
            {{ t('home.gatewayExplain.kicker') }}
          </p>
          <h2 class="home-gateway-title">
            {{ t('home.gatewayExplain.title') }}
          </h2>
          <p class="home-gateway-description">
            {{ t('home.gatewayExplain.description') }}
          </p>
        </div>

        <div class="home-gateway-visual" :aria-label="t('home.gatewayExplain.mapLabel')">
          <img
            class="home-gateway-art"
            src="/public/gateway-network-art.webp"
            alt=""
            loading="eager"
            decoding="async"
          />
          <div class="home-gateway-label home-gateway-label-models">
            <span>{{ t('home.gatewayExplain.poolLabel') }}</span>
            <strong>{{ t('home.gatewayExplain.poolName') }}</strong>
            <small>{{ gatewayModels.join(' / ') }}</small>
          </div>
          <div class="home-gateway-label home-gateway-label-entry">
            <span>{{ t('home.gatewayExplain.gatewayLabel') }}</span>
            <strong>{{ t('home.gatewayExplain.gatewayName') }}</strong>
          </div>
          <div class="home-gateway-label home-gateway-label-accounts">
            <span>{{ t('home.gatewayExplain.accountPoolLabel') }}</span>
            <strong>{{ t('home.gatewayExplain.accountPoolName') }}</strong>
          </div>
        </div>
      </div>
    </section>

    <section class="home-model-carousel-section relative z-10 px-4 sm:px-6" :aria-label="t('home.modelCarousel.ariaLabel')">
      <div class="home-model-carousel-shell mx-auto">
        <div class="home-model-carousel-copy mx-auto">
          <p class="home-model-carousel-kicker">
            {{ t('home.modelCarousel.kicker') }}
          </p>
          <h2 class="home-model-carousel-title">
            {{ t('home.modelCarousel.title') }}
          </h2>
          <p class="home-model-carousel-description">
            {{ t('home.modelCarousel.description') }}
          </p>
        </div>

        <div class="home-model-carousel-panel" aria-hidden="true">
          <div class="home-model-carousel-window">
            <div
              v-for="(row, rowIndex) in modelCarouselRows"
              :key="`model-row-${rowIndex}`"
              :class="['home-model-carousel-track', { 'is-reverse': rowIndex % 2 === 1 }]"
            >
              <div
                v-for="setIndex in 2"
                :key="`model-row-${rowIndex}-set-${setIndex}`"
                class="home-model-carousel-set"
              >
                <span
                  v-for="item in row"
                  :key="`${setIndex}-${item.model}`"
                  class="home-model-carousel-tile"
                  :title="item.label"
                >
                  <ModelIcon :model="item.model" size="32px" />
                </span>
              </div>
            </div>
          </div>
        </div>
        <router-link to="/models" class="home-model-carousel-link">
          <span>{{ t('home.modelCarousel.viewAll') }}</span>
          <Icon name="arrowRight" size="xs" aria-hidden="true" />
        </router-link>
      </div>
    </section>

    <section class="home-faq-section relative z-10 px-4 sm:px-6" :aria-label="t('home.faq.ariaLabel')">
      <div class="home-faq-shell mx-auto">
        <div class="home-faq-copy mx-auto">
          <p class="home-faq-kicker">
            {{ t('home.faq.kicker') }}
          </p>
          <h2 class="home-faq-title">
            {{ t('home.faq.title') }}
          </h2>
          <p class="home-faq-description">
            {{ t('home.faq.description') }}
          </p>
        </div>

        <div class="home-faq-tabs" aria-hidden="true">
          <span
            v-for="tab in faqTabs"
            :key="tab"
            class="home-faq-tab"
          >
            {{ tab }}
          </span>
        </div>

        <div class="home-faq-list">
          <article
            v-for="(item, index) in faqItems"
            :key="item.question"
            :class="['home-faq-item', { 'is-open': activeFaqIndex === index }]"
          >
            <button
              class="home-faq-question-row"
              type="button"
              :aria-expanded="activeFaqIndex === index"
              :aria-controls="`home-faq-answer-${index}`"
              @click="activeFaqIndex = activeFaqIndex === index ? -1 : index"
            >
              <h3>{{ item.question }}</h3>
              <span class="home-faq-plus" aria-hidden="true">{{ activeFaqIndex === index ? '-' : '+' }}</span>
            </button>
            <p
              v-if="activeFaqIndex === index"
              :id="`home-faq-answer-${index}`"
              class="home-faq-answer"
            >
              {{ item.answer }}
            </p>
          </article>
        </div>
      </div>
    </section>

    <section class="home-final-cta-section relative z-10 px-4 sm:px-6" :aria-label="t('home.finalCta.ariaLabel')">
      <div class="home-final-cta-card mx-auto">
        <p class="home-final-cta-kicker">
          {{ t('home.finalCta.kicker') }}
        </p>
        <h2 class="home-final-cta-title">
          {{ t('home.finalCta.title') }}
        </h2>
        <p class="home-final-cta-description">
          {{ t('home.finalCta.description') }}
        </p>
        <router-link v-if="isAuthenticated" :to="dashboardPath" class="home-final-cta-button">
          <span>{{ t('home.goToDashboard') }}</span>
        </router-link>
        <button v-else type="button" class="home-final-cta-button" @click="openAuthPanel">
          <span>{{ t('home.finalCta.button') }}</span>
        </button>
      </div>
    </section>

    <footer class="home-footer relative z-10">
      <div class="home-footer-inner mx-auto">
        <div class="home-footer-bar">
          <p class="home-footer-copy">
            &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
          </p>
          <nav class="home-footer-links" :aria-label="t('home.footer.linksLabel')">
            <router-link to="/tutorial">{{ t('home.navTutorial') }}</router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
              {{ t('home.docs') }}
            </a>
            <router-link
              v-for="doc in footerLegalDocuments"
              :key="doc.documentId"
              :to="{ name: 'LegalDocument', params: { documentId: doc.documentId } }"
            >
              {{ doc.title }}
            </router-link>
          </nav>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore, useAppStore } from '@/stores'
import { usageAPI, type UserDashboardStats } from '@/api/usage'
import { formatCreditAmount } from '@/utils/credits'
import { useClipboard } from '@/composables/useClipboard'
import { toLegalDocumentLink, type LegalDocumentLink } from '@/utils/legalDocuments'
import ModelIcon from '@/components/common/ModelIcon.vue'
import AuthAccessPanel from '@/components/auth/AuthAccessPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import PublicTopNav from './public/components/PublicTopNav.vue'
import PublicRevealBackdrop from './public/components/PublicRevealBackdrop.vue'

type ModelCarouselItem = {
  model: string
  label: string
}

type HomeFaqItem = {
  question: string
  answer: string
}

type AccountWorkbenchCard = {
  key: string
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  authMode?: 'login' | 'register'
}>(), {
  authMode: 'register'
})

const gatewayModels = ['Claude', 'GPT', 'Gemini', 'Qwen', 'DeepSeek']
const modelCarouselRows: ModelCarouselItem[][] = [
  [
    { model: 'claude-sonnet-4-5', label: 'Claude' },
    { model: 'gpt-5.1', label: 'GPT' },
    { model: 'gemini-3-pro-preview', label: 'Gemini' },
    { model: 'qwen3-max', label: 'Qwen' },
    { model: 'deepseek-v3.2', label: 'DeepSeek' },
    { model: 'mistral-large-latest', label: 'Mistral' },
    { model: 'llama-4-maverick', label: 'Llama' },
    { model: 'command-r-plus', label: 'Cohere' },
    { model: 'grok-4', label: 'Grok' },
    { model: 'kimi-k2', label: 'Kimi' },
    { model: 'doubao-seed-1-6', label: 'Doubao' },
    { model: 'minimax-m1', label: 'MiniMax' }
  ],
  [
    { model: 'ernie-4.5', label: 'ERNIE' },
    { model: 'spark-x1', label: 'Spark' },
    { model: 'hunyuan-turbos', label: 'Hunyuan' },
    { model: '@cf/meta/llama-3.3-70b', label: 'Workers AI' },
    { model: 'midjourney-v7', label: 'Midjourney' },
    { model: 'perplexity-sonar', label: 'Perplexity' },
    { model: 'jina-embeddings-v4', label: 'Jina' },
    { model: 'openrouter-auto', label: 'OpenRouter' },
    { model: 'suno-v4.5', label: 'Suno' },
    { model: 'ollama-local', label: 'Ollama' },
    { model: 'glm-4.6', label: 'GLM' },
    { model: 'yi-large', label: 'Yi' }
  ]
]
const { t } = useI18n()
const route = useRoute()
const activeFaqIndex = ref(0)
const activeHeroDescriptionIndex = ref(0)
const mobileAuthExpanded = ref(route.path === '/login' || route.path === '/register')
const accountWorkbenchStats = ref<UserDashboardStats | null>(null)
const accountWorkbenchLoading = ref(false)
const accountWorkbenchError = ref(false)
let heroDescriptionTimer: number | undefined
const faqTabs = computed(() => ['home.faq.tabs.service', 'home.faq.tabs.billing', 'home.faq.tabs.usage'].map(key => t(key)))
const faqItems = computed<HomeFaqItem[]>(() => [0, 1, 2, 3].map(index => ({
  question: t(`home.faq.items.${index}.question`),
  answer: t(`home.faq.items.${index}.answer`)
})))

const authStore = useAuthStore()
const appStore = useAppStore()
const { copied: apiEntryCopied, copyToClipboard } = useClipboard()
const authMode = computed(() => props.authMode)
const trustSignals = computed(() => [
  t('home.trustSignals.compatible'),
  t('home.trustSignals.routing'),
  t('home.trustSignals.traceable')
])

// Site settings - directly from appStore (already initialized from injected config)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const shouldRenderCustomHomeContent = computed(() => route.path === '/home' && !!homeContent.value)
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const publicApiEntryUrl = computed(() => resolvePublicApiEntryUrl(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || ''))
const heroTitleTop = computed(() =>
  appStore.cachedPublicSettings?.home_hero_title_top?.trim() || t('home.heroTitleTop')
)
const heroTitleBottom = computed(() =>
  appStore.cachedPublicSettings?.home_hero_title_bottom?.trim() || t('home.heroTitleBottom')
)
const heroHeadline = computed(() => `${heroTitleTop.value} ${heroTitleBottom.value}`.trim())
const currentYear = new Date().getFullYear()

const footerLegalDocuments = computed(() =>
  (appStore.cachedPublicSettings?.login_agreement_documents ?? [])
    .map(doc => toLegalDocumentLink(doc, {
      terms: t('home.footer.terms'),
      privacy: t('home.footer.privacy'),
      'usage-policy': t('home.footer.usagePolicy'),
      'supported-regions': t('home.footer.supportedRegions'),
      'service-specific-terms': t('home.footer.serviceSpecificTerms')
    }))
    .filter((doc): doc is LegalDocumentLink => doc !== null)
)

async function copyApiEntryUrl(): Promise<void> {
  await copyToClipboard(publicApiEntryUrl.value, t('home.apiEntryCopied'))
}

async function openAuthPanel(): Promise<void> {
  mobileAuthExpanded.value = true
  await nextTick()

  const panel = document.getElementById('home-auth-panel')
  panel?.scrollIntoView({
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
    block: 'center'
  })
  panel
    ?.querySelector<HTMLInputElement>('input:not([type="hidden"]):not(:disabled)')
    ?.focus({ preventScroll: true })
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

function resolveCurrentOrigin(): string {
  if (typeof window === 'undefined' || !window.location?.origin) {
    return ''
  }

  return window.location.origin
}

function resolvePublicApiEntryUrl(value: string): string {
  const configured = value.trim()
  const origin = resolveCurrentOrigin()

  if (!configured) {
    return trimTrailingSlash(origin)
  }

  if (configured.startsWith('/') && origin) {
    return trimTrailingSlash(new URL(configured, origin).toString())
  }

  return trimTrailingSlash(configured)
}

function splitHeroSubtitles(value: string | null | undefined): string[] {
  return (value || '')
    .split(/\r?\n/)
    .map((text) => text.trim())
    .filter(Boolean)
}

const defaultHeroDescriptionTexts = computed(() => [
  t('home.heroDescription'),
  t('home.heroDescriptionAltModels'),
  t('home.heroDescriptionAltSupport')
].filter(Boolean))

const heroDescriptionTexts = computed(() => {
  const configured = splitHeroSubtitles(appStore.cachedPublicSettings?.home_hero_subtitles)
  return configured.length > 0 ? configured : defaultHeroDescriptionTexts.value
})

const heroDescriptionText = computed(() => {
  const descriptions = heroDescriptionTexts.value
  if (descriptions.length === 0) return ''
  return descriptions[activeHeroDescriptionIndex.value % descriptions.length]
})

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const accountWorkbenchUserName = computed(() => (
  authStore.user?.username?.trim()
  || authStore.user?.email?.trim()
  || t('home.dashboard')
))
const accountWorkbenchBalance = computed(() => formatHomeCredit(authStore.user?.balance ?? 0))
const accountWorkbenchCards = computed<AccountWorkbenchCard[]>(() => {
  const stats = accountWorkbenchStats.value
  return [
    {
      key: 'total-tokens',
      label: t('home.accountWorkbench.totalTokens'),
      value: formatHomeNumber(stats?.total_tokens ?? 0),
    },
    {
      key: 'total-requests',
      label: t('home.accountWorkbench.totalRequests'),
      value: formatHomeNumber(stats?.total_requests ?? 0),
    },
    {
      key: 'total-cost',
      label: t('home.accountWorkbench.totalCost'),
      value: formatHomeCredit(stats?.total_actual_cost ?? 0),
    },
  ]
})

function formatHomeNumber(value: number | null | undefined): string {
  return new Intl.NumberFormat(undefined, {
    notation: Number(value) >= 10000 ? 'compact' : 'standard',
    maximumFractionDigits: Number(value) >= 10000 ? 1 : 0,
  }).format(Number(value) || 0)
}

function formatHomeCredit(value: number | null | undefined): string {
  return formatCreditAmount(value, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

async function loadAccountWorkbench(): Promise<void> {
  if (!isAuthenticated.value || shouldRenderCustomHomeContent.value) {
    accountWorkbenchStats.value = null
    accountWorkbenchLoading.value = false
    accountWorkbenchError.value = false
    return
  }

  accountWorkbenchLoading.value = true
  accountWorkbenchError.value = false
  try {
    const [stats] = await Promise.all([
      usageAPI.getDashboardStats(),
      authStore.refreshUser().catch(() => undefined),
    ])
    accountWorkbenchStats.value = stats
  } catch (error) {
    console.error('Failed to load home account workbench:', error)
    accountWorkbenchError.value = true
  } finally {
    accountWorkbenchLoading.value = false
  }
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    document.documentElement.classList.add('dark')
  }
}

function startHeroDescriptionCarousel() {
  if (heroDescriptionTimer) {
    window.clearInterval(heroDescriptionTimer)
    heroDescriptionTimer = undefined
  }

  if (heroDescriptionTexts.value.length <= 1) return

  heroDescriptionTimer = window.setInterval(() => {
    const total = heroDescriptionTexts.value.length
    if (total <= 1) return
    activeHeroDescriptionIndex.value = (activeHeroDescriptionIndex.value + 1) % total
  }, 3600)
}

watch(heroDescriptionTexts, () => {
  activeHeroDescriptionIndex.value = 0
  startHeroDescriptionCarousel()
})

watch([isAuthenticated, shouldRenderCustomHomeContent], () => {
  void loadAccountWorkbench()
}, { immediate: true })

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  startHeroDescriptionCarousel()
})

onBeforeUnmount(() => {
  if (heroDescriptionTimer) {
    window.clearInterval(heroDescriptionTimer)
    heroDescriptionTimer = undefined
  }
})
</script>

<style scoped>
@import './public/public-page.css';

.home-mimo-bg {
  background: #faf9f5;
}

.home-page-root {
  min-height: 100vh;
  min-height: 100svh;
  font-family: var(--public-font-sans);
  font-weight: 400;
  -webkit-font-smoothing: antialiased;
  text-rendering: geometricPrecision;
}

.home-main-stage {
  flex: 0 0 auto;
  display: grid;
  align-items: center;
  box-sizing: border-box;
  min-height: min(44rem, calc(82vh - 3.75rem));
  min-height: min(44rem, calc(82svh - 3.75rem));
}

.home-mimo-hero {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(21rem, 28rem);
  align-items: center;
  justify-items: stretch;
  gap: clamp(2rem, 5vw, 4.5rem);
  width: min(100%, 76rem);
  min-height: auto;
  align-content: center;
  padding: clamp(3.2rem, 7.2vh, 4.8rem) 0 clamp(3.2rem, 7vh, 4.8rem);
  text-align: left;
}

.home-mimo-copy {
  display: grid;
  justify-items: start;
  min-width: 0;
}

.home-mimo-title {
  position: relative;
  display: grid;
  gap: 0.08em;
  margin-top: 0;
  max-width: 12em;
  color: #141413;
  font-family: var(--public-font-display);
  font-size: 4rem;
  font-weight: 760;
  line-height: 1.08;
  letter-spacing: 0;
  text-wrap: balance;
  word-break: keep-all;
}

.home-mimo-title-line {
  position: relative;
  display: block;
}

.home-mimo-title-line-sweep::after {
  content: attr(data-title);
  position: absolute;
  inset: 0;
  pointer-events: none;
  color: transparent;
  background:
    linear-gradient(
      105deg,
      rgba(20, 20, 19, 0) 0%,
      rgba(20, 20, 19, 0) 31%,
      rgba(204, 120, 92, 0.28) 39%,
      rgba(255, 248, 235, 0.98) 47%,
      rgba(255, 255, 255, 0.92) 50%,
      rgba(204, 120, 92, 0.72) 57%,
      rgba(20, 20, 19, 0) 68%,
      rgba(20, 20, 19, 0) 100%
    );
  background-size: 220% 100%;
  background-position: 132% 50%;
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  filter: drop-shadow(0 0 18px rgba(204, 120, 92, 0.18));
  opacity: 0;
  animation: home-title-light-sweep 10s cubic-bezier(0.45, 0, 0.2, 1) infinite;
}

@keyframes home-title-light-sweep {
  0%,
  12% {
    background-position: 132% 50%;
    opacity: 0;
  }

  20% {
    opacity: 1;
  }

  62% {
    background-position: -42% 50%;
    opacity: 1;
  }

  76%,
  100% {
    background-position: -42% 50%;
    opacity: 0;
  }
}

.home-mimo-description {
  display: grid;
  min-height: 3.9em;
  align-items: center;
  margin-top: 1.42rem;
  max-width: 55rem;
  color: #3d3d3a;
  font-size: 1.14rem;
  font-weight: 400;
  line-height: 1.7;
  text-wrap: balance;
}

.home-mimo-description-text {
  display: block;
}

.home-subtitle-slide-enter-active,
.home-subtitle-slide-leave-active {
  transition:
    opacity 320ms ease,
    transform 320ms ease,
    filter 320ms ease;
}

.home-subtitle-slide-enter-from {
  opacity: 0;
  filter: blur(3px);
  transform: translateY(0.45rem);
}

.home-subtitle-slide-leave-to {
  opacity: 0;
  filter: blur(3px);
  transform: translateY(-0.45rem);
}

.home-command-pill {
  display: inline-flex;
  width: min(100%, 36rem);
  min-height: 3.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 2.7rem;
  border: 1px solid #e6dfd8;
  border-radius: 8px;
  background: rgba(250, 249, 245, 0.82);
  padding: 0.45rem 1.25rem;
  color: #141413;
  cursor: pointer;
  font: inherit;
  text-align: left;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 10px 24px rgba(20, 20, 19, 0.035);
}

.home-command-pill:hover {
  border-color: #cc785c;
  background: #fffaf5;
}

.home-command-pill:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.home-command-pill > span:first-child {
  color: #8e8b82;
  font-size: 0.76rem;
  font-weight: 500;
}

.home-command-pill code {
  min-width: 0;
  overflow: hidden;
  color: #141413;
  font-family: var(--public-font-mono);
  font-size: 0.9rem;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-command-copy-state {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.34rem;
  color: #a9583e;
  font-size: 0.76rem;
  font-weight: 650;
  white-space: nowrap;
}

.home-mimo-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 1.1rem;
  margin-top: 2rem;
}

.home-trust-signals {
  display: flex;
  flex-wrap: wrap;
  gap: 0.62rem 1rem;
  margin: 1.15rem 0 0;
  padding: 0;
  color: #504f49;
  font-size: 0.78rem;
  list-style: none;
}

.home-trust-signals li {
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
}

.home-trust-signals svg {
  color: #a9583e;
}

.home-auth-panel {
  width: min(100%, 28rem);
}

.home-account-workbench {
  position: relative;
  display: grid;
  width: min(100%, 28rem);
  min-height: 24rem;
  align-content: start;
  gap: 0.95rem;
  overflow: hidden;
  border: 1px solid rgba(216, 206, 194, 0.82);
  border-radius: 24px;
  background:
    radial-gradient(circle at 86% 10%, rgba(204, 120, 92, 0.14), rgba(204, 120, 92, 0) 12rem),
    linear-gradient(180deg, rgba(250, 249, 245, 0.78), rgba(239, 233, 222, 0.66)),
    rgba(250, 249, 245, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 26px 70px rgba(75, 52, 40, 0.11);
  padding: 1.35rem;
  backdrop-filter: blur(22px) saturate(1.08);
}

.home-account-workbench-header {
  display: grid;
  text-align: left;
}

.home-account-user-chip {
  display: inline-flex;
  min-width: 0;
  max-width: 100%;
  min-height: 2.15rem;
  width: fit-content;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(216, 206, 194, 0.82);
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.68);
  color: #141413;
  font-size: 0.8rem;
  font-weight: 650;
  padding: 0.32rem 0.78rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 8px 18px rgba(20, 20, 19, 0.04);
}

.home-account-user-chip span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-account-balance-card {
  display: grid;
  gap: 0.55rem;
  border: 1px solid rgba(216, 206, 194, 0.88);
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(255, 250, 245, 0.86), rgba(239, 233, 222, 0.7)),
    rgba(250, 249, 245, 0.74);
  padding: 1.15rem;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 16px 32px rgba(20, 20, 19, 0.045);
}

.home-account-balance-card span,
.home-account-stat-card span {
  color: rgba(80, 75, 67, 0.7);
  font-size: 0.78rem;
  font-weight: 600;
  line-height: 1.4;
}

.home-account-balance-card strong {
  color: #141413;
  font-family: var(--public-font-display);
  font-size: 2.15rem;
  font-weight: 760;
  line-height: 1;
  letter-spacing: 0;
}

.home-account-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.68rem;
}

.home-account-stat-card {
  display: grid;
  min-height: 5.8rem;
  align-content: space-between;
  gap: 0.6rem;
  border: 1px solid rgba(216, 206, 194, 0.72);
  border-radius: 15px;
  background: rgba(255, 250, 245, 0.52);
  padding: 0.82rem 0.72rem;
}

.home-account-stat-card strong {
  min-width: 0;
  overflow: hidden;
  color: #2a2924;
  font-family: var(--public-font-mono);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-account-workbench-state {
  min-height: 1.2rem;
  color: rgba(80, 75, 67, 0.62);
  font-size: 0.78rem;
  line-height: 1.5;
}

.home-account-workbench-state.is-error {
  color: #a9583e;
}

.home-account-workbench-link {
  display: inline-flex;
  min-height: 2.9rem;
  align-items: center;
  justify-content: center;
  gap: 0.48rem;
  margin-top: auto;
  border: 1px solid #141413;
  border-radius: 999px;
  background: #141413;
  color: #fffaf5;
  font-size: 0.92rem;
  font-weight: 650;
  text-decoration: none;
  transition:
    transform 150ms ease,
    background-color 150ms ease,
    border-color 150ms ease;
}

.home-account-workbench-link:hover {
  border-color: #2a2926;
  background: #2a2926;
  transform: translateY(-1px);
}

.home-account-workbench-link:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.home-action-button {
  display: inline-flex;
  min-width: min(9.8rem, calc(100vw - 2rem));
  overflow: hidden;
  border: 1px solid #e6dfd8;
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.68);
  color: #141413;
  cursor: pointer;
  font: inherit;
  text-decoration: none;
  transition:
    transform 120ms ease,
    border-color 120ms ease,
    background 120ms ease;
}

.home-action-button.is-primary {
  border-color: #141413;
  background: #141413;
  color: #fffaf5;
}

.home-action-button:hover {
  border-color: #d8cec2;
  background: #efe9de;
  transform: translateY(-1px);
}

.home-action-button.is-primary:hover {
  border-color: #2a2926;
  background: #2a2926;
}

.home-action-button:active {
  transform: translateY(1px);
}

.home-action-button:focus-visible {
  outline: 2px solid var(--public-ring);
  outline-offset: 3px;
}

.home-button-inner {
  display: inline-flex;
  min-height: 2.72rem;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 0.52rem;
  padding: 0.58rem 1.22rem;
  font-weight: 500;
  letter-spacing: 0;
}

.home-button-icon {
  flex: 0 0 auto;
  color: currentColor;
  transition: transform 180ms ease;
}

.home-action-button:hover .home-button-icon {
  transform: translateX(2px);
}

.home-gateway-section {
  flex: 0 0 auto;
  display: grid;
  min-height: 88vh;
  min-height: 88svh;
  align-items: start;
  border-top: 1px solid #e6dfd8;
  background: #f3f0ea;
  padding-top: clamp(4rem, 7vh, 5.8rem);
  padding-bottom: clamp(5rem, 9vh, 7.2rem);
}

.home-gateway-shell {
  width: min(100%, 72rem);
  text-align: center;
}

.home-gateway-copy {
  max-width: 52rem;
}

.home-gateway-kicker {
  color: #a9583e;
  font-size: 0.82rem;
  font-weight: 650;
  line-height: 1.5;
}

.home-gateway-title {
  margin-top: 1rem;
  color: #141413;
  font-family: var(--public-font-display);
  font-size: 3.08rem;
  font-weight: 400;
  line-height: 1.1;
  letter-spacing: 0;
  text-wrap: balance;
}

.home-gateway-description {
  margin: 1.35rem auto 0;
  max-width: 45rem;
  color: #3d3d3a;
  font-size: 1.08rem;
  line-height: 1.85;
  text-wrap: balance;
}

.home-gateway-visual {
  position: relative;
  overflow: hidden;
  width: min(100%, 78rem);
  margin: clamp(3rem, 6vh, 4.8rem) auto 0;
  border: 1px solid rgba(216, 206, 194, 0.68);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.5), rgba(239, 233, 222, 0.42)),
    #f3f0ea;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 30px 76px rgba(20, 20, 19, 0.08);
}

.home-gateway-visual::after {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(180deg, rgba(243, 240, 234, 0.1), rgba(243, 240, 234, 0.26)),
    radial-gradient(circle at 50% 52%, transparent 0 42%, rgba(243, 240, 234, 0.34) 72%);
}

.home-gateway-art {
  display: block;
  width: 100%;
  aspect-ratio: 1536 / 840;
  object-fit: cover;
  object-position: center;
}

.home-gateway-label {
  position: absolute;
  z-index: 2;
  display: grid;
  min-width: 10.8rem;
  gap: 0.28rem;
  border: 1px solid rgba(216, 206, 194, 0.74);
  border-radius: 12px;
  background: rgba(250, 249, 245, 0.68);
  padding: 0.82rem 1rem;
  text-align: left;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.72),
    0 16px 36px rgba(20, 20, 19, 0.06);
  backdrop-filter: blur(16px) saturate(1.08);
}

.home-gateway-label::before {
  content: '';
  position: absolute;
  left: 0.92rem;
  top: 0.92rem;
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 999px;
  background: #cc785c;
  box-shadow: 0 0 0 4px rgba(204, 120, 92, 0.12);
}

.home-gateway-label span {
  padding-left: 1rem;
  color: #8e8b82;
  font-size: 0.78rem;
  font-weight: 500;
}

.home-gateway-label strong {
  display: block;
  color: #141413;
  font-size: 1rem;
  font-weight: 650;
}

.home-gateway-label small {
  color: rgba(61, 61, 58, 0.62);
  font-size: 0.72rem;
  font-weight: 500;
  line-height: 1.45;
}

.home-gateway-label-models {
  left: 5.6%;
  top: 11%;
}

.home-gateway-label-entry {
  left: 50%;
  bottom: 12%;
  transform: translateX(-50%);
}

.home-gateway-label-accounts {
  right: 6.2%;
  top: 18%;
}

.home-model-carousel-section {
  flex: 0 0 auto;
  display: grid;
  min-height: 50vh;
  min-height: 50svh;
  align-items: center;
  border-top: 1px solid #ded5ca;
  background: #d7d1c4;
  padding-top: clamp(3rem, 6vh, 4.6rem);
  padding-bottom: clamp(3.4rem, 7vh, 5.2rem);
}

.home-model-carousel-shell {
  width: 100%;
  max-width: 74rem;
  text-align: center;
}

.home-model-carousel-copy {
  max-width: 48rem;
}

.home-model-carousel-kicker {
  color: #a9583e;
  font-size: 0.82rem;
  font-weight: 650;
  line-height: 1.5;
}

.home-model-carousel-title {
  margin-top: 1rem;
  color: #2a2924;
  font-family: var(--public-font-display);
  font-size: 1.36rem;
  font-weight: 400;
  line-height: 1.6;
  letter-spacing: 0;
  text-wrap: balance;
}

.home-model-carousel-title strong {
  color: #cc785c;
  font-weight: 700;
}

.home-model-carousel-description {
  margin: 0.9rem auto 0;
  max-width: 35rem;
  color: rgba(61, 61, 58, 0.72);
  font-size: 0.94rem;
  line-height: 1.75;
  text-wrap: balance;
}

.home-model-carousel-panel {
  position: relative;
  overflow: hidden;
  margin: clamp(1.65rem, 3.2vh, 2.5rem) auto 0;
  width: 100%;
  max-width: 62rem;
  border: 1px solid rgba(216, 206, 194, 0.68);
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.7), rgba(239, 233, 222, 0.72)),
    rgba(250, 249, 245, 0.52);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.64),
    0 24px 55px rgba(20, 20, 19, 0.085);
}

.home-model-carousel-panel::before,
.home-model-carousel-panel::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  z-index: 2;
  width: min(9rem, 24%);
  pointer-events: none;
}

.home-model-carousel-panel::before {
  left: 0;
  background: linear-gradient(90deg, #e6e1d7 0%, rgba(230, 225, 215, 0.88) 28%, rgba(230, 225, 215, 0) 100%);
}

.home-model-carousel-panel::after {
  right: 0;
  background: linear-gradient(270deg, #e6e1d7 0%, rgba(230, 225, 215, 0.88) 28%, rgba(230, 225, 215, 0) 100%);
}

.home-model-carousel-window {
  display: grid;
  gap: 1.18rem;
  overflow: hidden;
  padding: clamp(1.35rem, 2.6vw, 2.05rem) 0;
}

.home-model-carousel-track {
  display: flex;
  width: max-content;
  animation: home-model-marquee 34s linear infinite;
}

.home-model-carousel-track.is-reverse {
  animation-direction: reverse;
  animation-duration: 38s;
}

.home-model-carousel-set {
  display: flex;
  flex: 0 0 auto;
  gap: 1.55rem;
  padding: 0 0.78rem;
}

.home-model-carousel-tile {
  display: inline-flex;
  width: 4rem;
  height: 4rem;
  align-items: center;
  justify-content: center;
  border: 1px solid #d8cec2;
  border-radius: 13px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.48), rgba(239, 233, 222, 0.58)),
    rgba(239, 233, 222, 0.5);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.48),
    0 7px 14px rgba(20, 20, 19, 0.035);
}

.home-model-carousel-link {
  display: inline-flex;
  min-height: 2.7rem;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  margin-top: 1.35rem;
  color: #2a2924;
  font-size: 0.86rem;
  font-weight: 650;
  text-decoration: underline;
  text-underline-offset: 0.24em;
}

.home-model-carousel-link:hover {
  color: #a9583e;
}

@keyframes home-model-marquee {
  from {
    transform: translateX(0);
  }

  to {
    transform: translateX(-50%);
  }
}

.home-faq-section {
  flex: 0 0 auto;
  border-top: 1px solid rgba(216, 206, 194, 0.72);
  background: #f5f0e8;
  padding-top: clamp(4.2rem, 8vh, 6.2rem);
  padding-bottom: clamp(3.8rem, 7.2vh, 5.8rem);
}

.home-faq-shell {
  width: 100%;
  max-width: 62rem;
}

.home-faq-copy {
  max-width: 42rem;
  text-align: center;
}

.home-faq-kicker,
.home-final-cta-kicker {
  color: #a9583e;
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.5;
  text-transform: uppercase;
}

.home-faq-title,
.home-final-cta-title {
  margin-top: 0.82rem;
  color: #141413;
  font-family: var(--public-font-display);
  font-size: 3.05rem;
  font-weight: 400;
  line-height: 1.12;
  letter-spacing: 0;
  text-wrap: balance;
}

.home-faq-description,
.home-final-cta-description {
  margin: 1rem auto 0;
  max-width: 34rem;
  color: rgba(61, 61, 58, 0.72);
  font-size: 0.98rem;
  line-height: 1.82;
  text-wrap: balance;
}

.home-faq-tabs {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.65rem;
  margin: 1.75rem auto 0;
}

.home-faq-tab {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(160, 153, 144, 0.42);
  border-radius: 999px;
  background: rgba(250, 249, 245, 0.62);
  color: #504b43;
  font-size: 0.84rem;
  font-weight: 560;
  padding: 0.42rem 0.95rem;
}

.home-faq-list {
  display: grid;
  gap: 0.72rem;
  margin: clamp(2rem, 4.2vh, 3rem) auto 0;
  max-width: 48rem;
}

.home-faq-item {
  border: 1px solid rgba(216, 206, 194, 0.92);
  border-radius: 18px;
  background: rgba(250, 249, 245, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.58),
    0 14px 28px rgba(20, 20, 19, 0.04);
  padding: 1.1rem 1.25rem;
}

.home-faq-item.is-open {
  background: rgba(250, 249, 245, 0.92);
}

.home-faq-question-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
}

.home-faq-question-row h3 {
  color: #25231f;
  font-size: 1rem;
  font-weight: 650;
  line-height: 1.45;
}

.home-faq-question-row:focus-visible {
  outline: 2px solid rgba(204, 120, 92, 0.52);
  outline-offset: 4px;
  border-radius: 10px;
}

.home-faq-plus {
  display: inline-flex;
  flex: 0 0 auto;
  width: 1.75rem;
  height: 1.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(160, 153, 144, 0.46);
  border-radius: 999px;
  color: #6c6a64;
  font-family: var(--public-font-mono);
  font-size: 0.95rem;
}

.home-faq-answer {
  margin-top: 0.8rem;
  max-width: 40rem;
  color: rgba(61, 61, 58, 0.74);
  font-size: 0.94rem;
  line-height: 1.78;
}

.home-final-cta-section {
  flex: 0 0 auto;
  background:
    radial-gradient(circle at 50% 10%, rgba(204, 120, 92, 0.14), rgba(204, 120, 92, 0) 34rem),
    #f5f0e8;
  padding-top: clamp(1rem, 2.6vh, 2rem);
  padding-bottom: clamp(4.2rem, 8vh, 6.4rem);
}

.home-final-cta-card {
  width: min(100%, 58rem);
  border: 1px solid rgba(216, 206, 194, 0.82);
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.9), rgba(239, 233, 222, 0.72)),
    #faf9f5;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 28px 70px rgba(75, 52, 40, 0.11);
  padding: clamp(2.4rem, 5.4vw, 4.5rem) clamp(1.35rem, 5vw, 4rem);
  text-align: center;
}

.home-final-cta-button {
  display: inline-flex;
  min-height: 3.05rem;
  align-items: center;
  justify-content: center;
  border: 0;
  margin-top: 1.75rem;
  border-radius: 999px;
  background: #cc785c;
  color: #fffaf5;
  cursor: pointer;
  font: inherit;
  font-size: 0.95rem;
  font-weight: 700;
  line-height: 1.2;
  padding: 0.88rem 1.55rem;
  text-decoration: none;
  transition: transform 180ms ease, background-color 180ms ease, box-shadow 180ms ease;
  box-shadow: 0 16px 32px rgba(204, 120, 92, 0.24);
}

.home-final-cta-button:hover {
  background: #a9583e;
  box-shadow: 0 18px 36px rgba(169, 88, 62, 0.26);
  transform: translateY(-1px);
}

.home-footer {
  flex: 0 0 auto;
  width: 100%;
  background: #d9d7d2;
  padding: clamp(4.2rem, 9vh, 6.2rem) max(1rem, env(safe-area-inset-left)) clamp(4.4rem, 9vh, 6.5rem) max(1rem, env(safe-area-inset-right));
}

.home-footer-inner {
  width: min(100%, 76rem);
  color: #24231f;
}

.home-footer-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.55rem;
  min-height: 3rem;
  padding: 0;
  text-align: center;
}

.home-footer-copy {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  color: #24231f;
  font-size: 0.78rem;
  line-height: 1.6;
  white-space: nowrap;
}

.home-footer-copy::after {
  content: '';
  display: inline-block;
  width: 1px;
  height: 0.9em;
  margin-left: 0.6rem;
  background: #6c6a64;
}

.home-footer-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0;
  color: #24231f;
  font-size: 0.78rem;
  font-weight: 400;
  line-height: 1.6;
}

.home-footer-links a {
  color: #24231f;
  text-decoration: underline;
  text-underline-offset: 0.18em;
  transition:
    color 150ms ease,
    text-decoration-color 150ms ease;
}

.home-footer-links a + a {
  margin-left: 0.6rem;
  border-left: 1px solid #6c6a64;
  padding-left: 0.6rem;
}

.home-footer-links a:hover {
  color: #a9583e;
}

@media (max-width: 640px) {
  .home-main-stage {
    align-items: center;
    min-height: auto;
  }

  .home-mimo-hero {
    display: grid;
    grid-template-columns: 1fr;
    width: min(100%, 34rem);
    min-height: auto;
    gap: 1.45rem;
    padding: 2.55rem 0 3rem;
    text-align: center;
  }

  .home-mimo-copy {
    justify-items: center;
  }

  .home-mimo-title {
    max-width: min(100%, 21rem);
    font-size: 2.55rem;
    line-height: 1.08;
    overflow-wrap: anywhere;
    word-break: normal;
  }

  .home-mimo-description {
    min-height: 4.9em;
    font-size: 1rem;
    line-height: 1.72;
  }

  .home-command-pill {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    justify-content: stretch;
    gap: 0.24rem 0.75rem;
    border-radius: 22px;
    padding: 0.72rem 0.9rem;
    text-align: left;
  }

  .home-command-pill > span:first-child {
    grid-column: 1 / -1;
  }

  .home-command-pill code {
    align-self: center;
  }

  .home-mimo-actions {
    width: 100%;
  }

  .home-action-button {
    width: 100%;
  }

  .home-action-button.is-secondary {
    min-width: 0;
    width: auto;
    border-color: transparent;
    background: transparent;
    text-decoration: underline;
    text-underline-offset: 0.24em;
  }

  .home-action-button.is-secondary .home-button-inner {
    min-height: 2.25rem;
    padding: 0.3rem 0.65rem;
  }

  .home-trust-signals {
    justify-content: center;
    margin-top: 0.75rem;
  }

  .home-auth-panel:not(.is-mobile-open) {
    display: none;
  }

  .home-account-workbench {
    width: 100%;
    min-height: auto;
    border-radius: 20px;
    padding: 1.08rem;
  }

  .home-account-workbench-header {
    justify-items: center;
    text-align: center;
  }

  .home-account-balance-card {
    text-align: center;
  }

  .home-account-balance-card strong {
    font-size: 1.86rem;
  }

  .home-account-stat-grid {
    grid-template-columns: 1fr;
  }

  .home-account-stat-card {
    min-height: auto;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
  }

  .home-account-stat-card strong {
    text-align: right;
  }

  .home-gateway-section {
    min-height: auto;
    padding-top: 3.8rem;
    padding-bottom: 4.8rem;
  }

  .home-gateway-title {
    margin-top: 0.82rem;
    font-size: 1.95rem;
    line-height: 1.14;
  }

  .home-gateway-description {
    font-size: 1rem;
    line-height: 1.75;
  }

  .home-gateway-visual {
    display: grid;
    gap: 0.72rem;
    overflow: visible;
    margin-top: 2.4rem;
    border-radius: 18px;
    background: transparent;
    box-shadow: none;
  }

  .home-gateway-visual::after {
    display: none;
  }

  .home-gateway-art {
    overflow: hidden;
    aspect-ratio: 1536 / 1024;
    border: 1px solid rgba(216, 206, 194, 0.68);
    border-radius: 18px;
    object-position: center;
    box-shadow: 0 18px 44px rgba(20, 20, 19, 0.06);
  }

  .home-gateway-label {
    position: static;
    min-width: 0;
    width: 100%;
    padding: 0.76rem 0.9rem;
    transform: none;
  }

  .home-gateway-label small {
    font-size: 0.7rem;
  }

  .home-model-carousel-section {
    min-height: auto;
    padding-top: 3.8rem;
    padding-bottom: 3.8rem;
  }

  .home-model-carousel-title {
    font-size: 1.1rem;
    line-height: 1.62;
  }

  .home-model-carousel-description {
    font-size: 0.92rem;
  }

  .home-model-carousel-panel {
    width: 100%;
    max-width: 100%;
    border-radius: 18px;
  }

  .home-model-carousel-shell,
  .home-model-carousel-panel {
    max-width: calc(100vw - 2rem);
  }

  .home-model-carousel-window {
    gap: 0.82rem;
    padding: 1.25rem 0;
  }

  .home-model-carousel-set {
    gap: 0.72rem;
    padding: 0 0.36rem;
  }

  .home-model-carousel-tile {
    width: 3.25rem;
    height: 3.25rem;
    border-radius: 11px;
  }

  .home-faq-section {
    padding-top: 3.5rem;
    padding-bottom: 3.2rem;
  }

  .home-faq-title,
  .home-final-cta-title {
    font-size: 2rem;
    line-height: 1.16;
  }

  .home-faq-description,
  .home-final-cta-description {
    font-size: 0.92rem;
    line-height: 1.72;
  }

  .home-faq-tabs {
    gap: 0.48rem;
    margin-top: 1.35rem;
  }

  .home-faq-tab {
    min-height: 2.16rem;
    font-size: 0.78rem;
    padding: 0.36rem 0.72rem;
  }

  .home-faq-list {
    gap: 0.62rem;
    margin-top: 1.55rem;
  }

  .home-faq-item {
    border-radius: 14px;
    padding: 0.96rem 1rem;
  }

  .home-faq-question-row h3 {
    font-size: 0.92rem;
  }

  .home-faq-answer {
    font-size: 0.88rem;
  }

  .home-final-cta-section {
    padding-top: 0.6rem;
    padding-bottom: 3.5rem;
  }

  .home-final-cta-card {
    border-radius: 20px;
    padding: 2.25rem 1.15rem;
  }

  .home-final-cta-button {
    width: 100%;
    max-width: 16rem;
  }

  .home-footer-bar {
    align-items: center;
    flex-direction: column;
    gap: 0.55rem;
    text-align: center;
  }

  .home-footer-copy {
    white-space: normal;
  }

  .home-footer-copy::after {
    display: none;
  }

  .home-footer-links {
    justify-content: center;
    font-size: 0.78rem;
  }
}

@media (max-width: 1023px) {
  .home-mimo-hero {
    grid-template-columns: 1fr;
    width: min(100%, 40rem);
    text-align: center;
  }

  .home-mimo-copy {
    justify-items: center;
  }

  .home-mimo-actions {
    justify-content: center;
  }
}

@media (min-width: 1536px) {
  .home-mimo-hero {
    width: min(100%, 80rem);
  }

  .home-mimo-title {
    font-size: 4.45rem;
  }
}

@media (max-width: 420px) {
  .home-mimo-title {
    font-size: 2.25rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-model-carousel-track {
    animation: none;
  }

  .home-subtitle-slide-enter-active,
  .home-subtitle-slide-leave-active {
    transition: none;
  }

  .home-mimo-title-line-sweep::after {
    animation: none;
    opacity: 0;
  }
}

</style>
