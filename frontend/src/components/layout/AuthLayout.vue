<template>
  <div class="auth-pixel-bg relative flex min-h-screen items-center justify-center overflow-hidden p-4 text-white">
    <div class="auth-blur-field pointer-events-none absolute inset-0"></div>
    <div class="auth-noise pointer-events-none absolute inset-0"></div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-7 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden border border-white/20 bg-white/12 shadow-[inset_0_1px_0_rgba(255,255,255,0.2),0_14px_26px_rgba(8,5,21,0.26)]"
          >
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="auth-brand-title mb-2 text-3xl font-black">
            {{ siteName }}
          </h1>
          <p class="text-sm font-medium text-violet-100/72">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="auth-panel p-8">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="auth-footer mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-violet-100/48">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => {
  const subtitle = appStore.cachedPublicSettings?.site_subtitle?.trim()
  if (!subtitle || subtitle === 'Subscription to API Conversion Platform') {
    return '智能编码国内解决方案'
  }
  return subtitle
})
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-pixel-bg {
  background:
    radial-gradient(circle at 50% 38%, rgba(169, 109, 255, 0.16) 0, transparent 34%),
    radial-gradient(circle at 18% 22%, rgba(105, 86, 255, 0.26) 0, transparent 30%),
    radial-gradient(circle at 82% 28%, rgba(168, 75, 255, 0.2) 0, transparent 28%),
    linear-gradient(180deg, #120b35 0%, #160932 48%, #080515 100%);
}

.auth-blur-field {
  background:
    radial-gradient(ellipse at 50% 32%, rgba(197, 118, 255, 0.28), transparent 36%),
    radial-gradient(ellipse at 34% 22%, rgba(74, 86, 255, 0.22), transparent 28%),
    radial-gradient(ellipse at 70% 24%, rgba(117, 41, 199, 0.28), transparent 30%),
    radial-gradient(ellipse at 54% 80%, rgba(77, 223, 255, 0.1), transparent 28%);
  filter: blur(44px);
  opacity: 0.94;
  transform: scale(1.05);
}

.auth-noise {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 54px 54px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.56), transparent 80%);
  opacity: 0.36;
}

.auth-brand-title {
  color: transparent;
  background: linear-gradient(98deg, #ffffff 0%, #c4a5ff 42%, #74d5ff 78%, #ffffff 100%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.auth-panel {
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.09);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 18px 44px rgba(8, 5, 21, 0.32);
  backdrop-filter: blur(20px);
}

.auth-footer :deep(a) {
  color: #ffd85d;
  font-weight: 800;
}

.auth-footer :deep(a:hover) {
  color: #ffec86;
}
</style>
