<template>
  <div class="auth-pixel-bg relative flex min-h-screen items-center justify-center overflow-hidden p-4 text-[#141413]">
    <div class="auth-blur-field pointer-events-none absolute inset-0"></div>
    <div class="auth-noise pointer-events-none absolute inset-0"></div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-7 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div class="mb-4 inline-flex h-12 w-12 items-center justify-center overflow-hidden rounded-full border border-[#d8cec2] bg-[#faf9f5]/80 shadow-[inset_0_1px_0_rgba(255,255,255,0.7),0_12px_26px_rgba(20,20,19,0.08)]">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="auth-brand-title mb-2 text-3xl font-bold">
            {{ siteName }}
          </h1>
          <p class="text-sm font-medium text-[#6c6a64]">
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
      <div class="mt-8 text-center text-xs text-[#8e8b82]">
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
    radial-gradient(circle at 18% 12%, rgba(204, 120, 92, 0.14) 0, transparent 28rem),
    radial-gradient(circle at 84% 16%, rgba(216, 206, 194, 0.48) 0, transparent 26rem),
    linear-gradient(180deg, #faf9f5 0%, #f5f0e8 54%, #efe9de 100%);
}

.auth-blur-field {
  background:
    radial-gradient(ellipse at 50% 28%, rgba(204, 120, 92, 0.12), transparent 34%),
    radial-gradient(ellipse at 36% 18%, rgba(255, 248, 235, 0.72), transparent 28%),
    radial-gradient(ellipse at 70% 24%, rgba(216, 206, 194, 0.36), transparent 30%);
  filter: blur(38px);
  opacity: 0.86;
  transform: scale(1.05);
}

.auth-noise {
  background-image:
    linear-gradient(rgba(216, 206, 194, 0.24) 1px, transparent 1px),
    linear-gradient(90deg, rgba(216, 206, 194, 0.18) 1px, transparent 1px);
  background-size: 46px 46px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.56), transparent 80%);
  opacity: 0.28;
}

.auth-brand-title {
  color: #141413;
}

.auth-panel {
  border: 1px solid rgba(216, 206, 194, 0.82);
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(250, 249, 245, 0.92), rgba(239, 233, 222, 0.76)),
    rgba(250, 249, 245, 0.86);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.68),
    0 24px 64px rgba(75, 52, 40, 0.12);
  backdrop-filter: blur(20px);
}

.auth-footer :deep(a) {
  color: #a9583e;
  font-weight: 650;
}

.auth-footer :deep(a:hover) {
  color: #cc785c;
}
</style>
