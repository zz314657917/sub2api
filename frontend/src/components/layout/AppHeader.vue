<template>
  <header class="console-header sticky top-0 z-30">
    <div class="flex h-16 items-center justify-between px-4 md:px-6">
      <div class="flex items-center gap-4">
        <button
          @click="toggleMobileSidebar"
          class="console-icon-button inline-flex h-10 w-10 items-center justify-center lg:hidden"
          aria-label="Toggle Menu"
        >
          <Icon name="menu" size="md" />
        </button>

        <div class="hidden lg:block">
          <h1 class="console-page-title text-lg font-semibold">
            {{ pageTitle }}
          </h1>
          <p v-if="pageDescription" class="console-page-description text-xs">
            {{ pageDescription }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <AnnouncementBell v-if="user" />

        <button
          v-if="hasSupportPopupItems"
          type="button"
          class="console-chip flex items-center gap-1.5 px-2.5 py-1.5 text-sm font-semibold transition-colors"
          @click="openSupportPopup"
        >
          <Icon name="chatBubble" size="sm" />
          <span class="hidden sm:inline">{{ t('common.contactSupport') }}</span>
        </button>

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="console-chip flex items-center gap-1.5 px-2.5 py-1.5 text-sm font-semibold transition-colors"
        >
          <Icon name="book" size="sm" />
          <span class="hidden sm:inline">{{ t('nav.docs') }}</span>
        </a>

        <LocaleSwitcher />
        <SubscriptionProgressMini v-if="user" />

        <div v-if="user" class="console-balance hidden items-center gap-2 px-3 py-1.5 sm:flex">
          <Icon name="dollar" size="sm" />
          <span class="text-sm font-semibold">
            ${{ user.balance?.toFixed(2) || '0.00' }}
          </span>
        </div>

        <div v-if="user" class="relative" ref="dropdownRef">
          <button
            @click="toggleDropdown"
            class="console-user-button flex items-center gap-2 p-1.5 transition-colors"
            aria-label="User Menu"
          >
            <div class="console-avatar flex h-8 w-8 items-center justify-center overflow-hidden text-sm font-semibold text-white">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                class="h-full w-full object-cover"
              >
              <span v-else>{{ userInitials }}</span>
            </div>
            <div class="hidden text-left md:block">
              <div class="text-sm font-medium text-white">
                {{ displayName }}
              </div>
              <div class="text-xs capitalize text-violet-200/60">
                {{ user.role }}
              </div>
            </div>
            <Icon name="chevronDown" size="xs" class="hidden md:block" />
          </button>

          <transition name="dropdown">
            <div v-if="dropdownOpen" class="dropdown right-0 mt-2 w-56">
              <div class="border-b border-white/10 px-4 py-3">
                <div class="text-sm font-medium text-white">
                  {{ displayName }}
                </div>
                <div class="text-xs text-violet-200/60">{{ user.email }}</div>
              </div>

              <div class="border-b border-white/10 px-4 py-2 sm:hidden">
                <div class="text-xs text-violet-200/60">
                  {{ t('common.balance') }}
                </div>
                <div class="text-sm font-semibold text-emerald-200">
                  ${{ user.balance?.toFixed(2) || '0.00' }}
                </div>
              </div>

              <div class="py-1">
                <router-link to="/profile" @click="closeDropdown" class="dropdown-item">
                  <Icon name="user" size="sm" />
                  {{ t('nav.profile') }}
                </router-link>

                <router-link to="/keys" @click="closeDropdown" class="dropdown-item">
                  <Icon name="key" size="sm" />
                  {{ t('nav.apiKeys') }}
                </router-link>

                <a
                  v-if="authStore.isAdmin"
                  href="https://github.com/Wei-Shaw/sub2api"
                  target="_blank"
                  rel="noopener noreferrer"
                  @click="closeDropdown"
                  class="dropdown-item"
                >
                  <Icon name="externalLink" size="sm" />
                  {{ t('nav.github') }}
                </a>
              </div>

              <div v-if="contactInfo" class="border-t border-white/10 px-4 py-2.5">
                <div class="flex items-center gap-2 text-xs text-violet-200/60">
                  <Icon name="chatBubble" size="xs" />
                  <span>{{ t('common.contactSupport') }}:</span>
                  <span class="font-medium text-violet-50">{{ contactInfo }}</span>
                </div>
              </div>

              <div v-if="showOnboardingButton" class="border-t border-white/10 py-1">
                <button @click="handleReplayGuide" class="dropdown-item w-full">
                  <Icon name="lightbulb" size="sm" />
                  {{ $t('onboarding.restartTour') }}
                </button>
              </div>

              <div class="border-t border-white/10 py-1">
                <button
                  @click="handleLogout"
                  class="dropdown-item w-full text-rose-300 hover:bg-rose-900/20"
                >
                  <Icon name="logout" size="sm" />
                  {{ t('nav.logout') }}
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import Icon from '@/components/icons/Icon.vue'
import { openSupportPopup } from '@/utils/supportPopup'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const onboardingStore = useOnboardingStore()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const contactInfo = computed(() => appStore.contactInfo)
const docUrl = computed(() => appStore.docUrl)
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const hasSupportPopupItems = computed(() => {
  const items = appStore.cachedPublicSettings?.support_popup_items
  return Array.isArray(items) && items.some((item) => item.title?.trim() && item.image_url?.trim())
})

const showOnboardingButton = computed(() => {
  return !authStore.isSimpleMode && user.value?.role === 'admin'
})

const userInitials = computed(() => {
  if (!user.value) return ''
  if (user.value.username) {
    return user.value.username.substring(0, 2).toUpperCase()
  }
  if (user.value.email) {
    const localPart = user.value.email.split('@')[0]
    return localPart.substring(0, 2).toUpperCase()
  }
  return ''
})

const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

const pageTitle = computed(() => {
  if (route.name === 'CustomPage') {
    const id = route.params.id as string
    const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
    const menuItem =
      publicItems.find((item) => item.id === id) ??
      (authStore.isAdmin
        ? adminSettingsStore.customMenuItems.find((item) => item.id === id)
        : undefined)
    if (menuItem?.label) return menuItem.label
  }
  const titleKey = route.meta.titleKey as string
  if (titleKey) {
    return t(titleKey)
  }
  return (route.meta.title as string) || ''
})

const pageDescription = computed(() => {
  const descKey = route.meta.descriptionKey as string
  if (descKey) {
    return t(descKey)
  }
  return (route.meta.description as string) || ''
})

function toggleMobileSidebar() {
  appStore.toggleMobileSidebar()
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide() {
  closeDropdown()
  onboardingStore.replay()
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
