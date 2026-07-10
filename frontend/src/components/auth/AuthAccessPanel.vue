<template>
  <section class="auth-access-panel" :class="{ 'auth-access-panel--embedded': embedded }">
    <header class="auth-access-heading">
      <h2 class="auth-access-title">{{ panelTitle }}</h2>
      <p class="auth-access-subtitle">{{ panelSubtitle }}</p>
    </header>

    <div class="auth-access-tabs" role="tablist" :aria-label="t('auth.accessPanelLabel')">
      <button
        type="button"
        class="auth-access-tab"
        :class="{ 'is-active': activeMode === 'register' }"
        role="tab"
        :aria-selected="activeMode === 'register'"
        @click="activeMode = 'register'"
      >
        {{ t('auth.signUp') }}
      </button>
      <button
        type="button"
        class="auth-access-tab"
        :class="{ 'is-active': activeMode === 'login' }"
        role="tab"
        :aria-selected="activeMode === 'login'"
        @click="activeMode = 'login'"
      >
        {{ t('auth.signIn') }}
      </button>
    </div>

    <div class="auth-access-body">
      <LoginView
        v-if="activeMode === 'login'"
        embedded
        :show-title="false"
        @switch-to-register="activeMode = 'register'"
      />
      <RegisterView
        v-else
        embedded
        :show-title="false"
        @switch-to-login="activeMode = 'login'"
      />
    </div>

    <p class="auth-access-terms">
      {{ t('auth.accessAgreementPrefix') }}
      <template v-if="agreementDocumentLinks.length > 0">
        <template
          v-for="(doc, index) in agreementDocumentLinks"
          :key="doc.documentId"
        >
          <RouterLink
            :to="{ name: 'LegalDocument', params: { documentId: doc.documentId } }"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ doc.title }}
          </RouterLink>
          <span v-if="index < agreementDocumentLinks.length - 1">
            {{ t('auth.accessAgreementSeparator') }}
          </span>
        </template>
      </template>
      <span v-else>{{ t('auth.accessAgreementFallback') }}</span>
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LoginView from '@/views/auth/LoginView.vue'
import RegisterView from '@/views/auth/RegisterView.vue'
import { useAppStore } from '@/stores'
import { toLegalDocumentLink, type LegalDocumentLink } from '@/utils/legalDocuments'

const props = withDefaults(defineProps<{
  initialMode?: 'login' | 'register'
  embedded?: boolean
}>(), {
  initialMode: 'register',
  embedded: false
})

const { t } = useI18n()
const appStore = useAppStore()
const activeMode = ref<'login' | 'register'>(props.initialMode)

const panelTitle = computed(() =>
  activeMode.value === 'login' ? t('auth.accessLoginTitle') : t('auth.accessRegisterTitle')
)
const panelSubtitle = computed(() =>
  activeMode.value === 'login' ? t('auth.accessLoginSubtitle') : t('auth.accessRegisterSubtitle')
)
const agreementDocumentLinks = computed(() =>
  buildAgreementDocumentLinks()
)

watch(
  () => props.initialMode,
  (mode) => {
    activeMode.value = mode
  }
)

function buildAgreementDocumentLinks(): LegalDocumentLink[] {
  const titleFallbacks = {
    terms: t('home.footer.terms'),
    privacy: t('home.footer.privacy'),
    'usage-policy': t('home.footer.usagePolicy'),
    'supported-regions': t('home.footer.supportedRegions'),
    'service-specific-terms': t('home.footer.serviceSpecificTerms')
  }
  const links = (appStore.cachedPublicSettings?.login_agreement_documents ?? [])
    .map(doc => toLegalDocumentLink(doc, titleFallbacks))
    .filter((doc): doc is LegalDocumentLink => doc !== null)

  if (links.length <= 2) {
    return links
  }

  const preferred = [
    links.find(isTermsDocument),
    links.find(isPolicyDocument)
  ].filter((doc): doc is LegalDocumentLink => doc !== undefined)

  const unique = preferred.filter(
    (doc, index, list) => list.findIndex((item) => item.documentId === doc.documentId) === index
  )

  return unique.length > 0 ? unique : links.slice(0, 2)
}

function isTermsDocument(doc: LegalDocumentLink): boolean {
  const normalized = `${doc.documentId} ${doc.title}`.toLowerCase()
  return normalized.includes('terms')
}

function isPolicyDocument(doc: LegalDocumentLink): boolean {
  const normalized = `${doc.documentId} ${doc.title}`.toLowerCase()
  return normalized.includes('privacy') || normalized.includes('policy') || normalized.includes('usage')
}

</script>

<style scoped>
.auth-access-panel {
  width: min(100%, 28rem);
  min-height: 31rem;
  border: 1px solid rgba(216, 206, 194, 0.72);
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.9), rgba(239, 233, 222, 0.72)),
    rgba(250, 249, 245, 0.84);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.68),
    0 28px 72px rgba(75, 52, 40, 0.12);
  padding: clamp(1rem, 2vw, 1.35rem);
  backdrop-filter: blur(22px) saturate(1.08);
}

.auth-access-heading {
  margin-bottom: 1rem;
}

.auth-access-title {
  margin: 0;
  color: #141413;
  font-size: clamp(1.5rem, 2.2vw, 1.95rem);
  font-weight: 800;
  line-height: 1.12;
}

.auth-access-subtitle {
  margin: 0.42rem 0 0;
  color: #6c6a64;
  font-size: 0.9rem;
  font-weight: 500;
  line-height: 1.55;
}

.auth-access-panel--embedded {
  justify-self: center;
}

.auth-access-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem;
  margin-bottom: 1.15rem;
  border: 1px solid rgba(216, 206, 194, 0.68);
  border-radius: 999px;
  background: rgba(245, 240, 232, 0.78);
  padding: 0.28rem;
}

.auth-access-body {
  min-height: 15.75rem;
}

.auth-access-tab {
  min-height: 2.35rem;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: #6c6a64;
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 650;
  transition:
    background-color 160ms ease,
    color 160ms ease,
    box-shadow 160ms ease;
}

.auth-access-tab.is-active {
  background: #141413;
  color: #fffaf5;
  box-shadow: 0 10px 22px rgba(20, 20, 19, 0.12);
}

.auth-access-tab:focus-visible {
  outline: 2px solid rgba(204, 120, 92, 0.45);
  outline-offset: 2px;
}

.auth-access-terms {
  margin: 1rem auto 0;
  max-width: 22rem;
  color: #8e8b82;
  font-size: 0.76rem;
  font-weight: 500;
  line-height: 1.7;
  text-align: center;
}

.auth-access-terms a {
  color: #504f49;
  font-weight: 700;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 3px;
  transition: color 140ms ease;
}

.auth-access-terms a:hover {
  color: #141413;
}

@media (max-width: 640px) {
  .auth-access-panel {
    min-height: 0;
    border-radius: 18px;
    padding: 0.95rem;
  }

  .auth-access-body {
    min-height: 0;
  }
}
</style>
