<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import SupportPopup from '@/components/common/SupportPopup.vue'
import { resolveDocumentTitle } from '@/router/title'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import { useAppStore, useAuthStore, useSubscriptionStore, useAnnouncementStore } from '@/stores'
import { getSetupStatus } from '@/api/setup'
import { SUPPORT_POPUP_EVENT } from '@/utils/supportPopup'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const announcementStore = useAnnouncementStore()
const supportPopupVisible = ref(false)

/**
 * Update favicon dynamically
 * @param logoUrl - URL of the logo to use as favicon
 */
function getFaviconType(logoUrl: string) {
  const lowerUrl = logoUrl.trim().toLowerCase()
  const dataUrlMatch = /^data:([^;,]+)/.exec(lowerUrl)
  if (dataUrlMatch?.[1]?.startsWith('image/')) return dataUrlMatch[1]

  const urlWithoutQuery = lowerUrl.split(/[?#]/)[0]
  if (urlWithoutQuery.endsWith('.svg')) return 'image/svg+xml'
  if (urlWithoutQuery.endsWith('.webp')) return 'image/webp'
  if (urlWithoutQuery.endsWith('.png')) return 'image/png'
  if (urlWithoutQuery.endsWith('.jpg') || urlWithoutQuery.endsWith('.jpeg')) return 'image/jpeg'
  if (urlWithoutQuery.endsWith('.gif')) return 'image/gif'
  return 'image/x-icon'
}

function updateFavicon(logoUrl: string) {
  // Find existing favicon link or create new one
  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.type = getFaviconType(logoUrl)
  link.href = logoUrl
}

// Watch for site settings changes and update favicon/title
watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) {
      updateFavicon(newLogo)
    }
  },
  { immediate: true }
)

// Watch for authentication state and manage subscription data + announcements
function onVisibilityChange() {
  if (document.visibilityState === 'visible' && authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
}

function openSupportPopup() {
  supportPopupVisible.value = true
}

function closeSupportPopup() {
  supportPopupVisible.value = false
}

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated, oldValue) => {
    if (isAuthenticated) {
      // User logged in: preload subscriptions and start polling
      subscriptionStore.fetchActiveSubscriptions().catch((error) => {
        console.error('Failed to preload subscriptions:', error)
      })
      subscriptionStore.startPolling()

      // Announcements: new login vs page refresh restore
      if (oldValue === false) {
        // New login: delay 3s then force fetch
        setTimeout(() => announcementStore.fetchAnnouncements(true), 3000)
      } else {
        // Page refresh restore (oldValue was undefined)
        announcementStore.fetchAnnouncements()
      }

      // Register visibility change listener
      document.addEventListener('visibilitychange', onVisibilityChange)
    } else {
      // User logged out: clear data and stop polling
      subscriptionStore.clear()
      announcementStore.reset()
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  },
  { immediate: true }
)

// Route change trigger (throttled by store)
router.afterEach(() => {
  if (authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  window.removeEventListener(SUPPORT_POPUP_EVENT, openSupportPopup)
})

onMounted(async () => {
  window.addEventListener(SUPPORT_POPUP_EVENT, openSupportPopup)

  // Check if setup is needed
  try {
    const status = await getSetupStatus()
    if (status.needs_setup && route.path !== '/setup') {
      router.replace('/setup')
      return
    }
  } catch {
    // If setup endpoint fails, assume normal mode and continue
  }

  // Load public settings into appStore (will be cached for other components)
  await appStore.fetchPublicSettings()

  // Re-resolve document title now that siteName is available
  document.title = resolveDocumentTitle(route.meta.title, appStore.siteName, route.meta.titleKey as string, appStore.cachedPublicSettings)
})
</script>

<template>
  <NavigationProgress />
  <RouterView />
  <Toast />
  <AnnouncementPopup />
  <SupportPopup :show="supportPopupVisible" @close="closeSupportPopup" />
</template>
