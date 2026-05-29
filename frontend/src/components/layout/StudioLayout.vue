<template>
  <div class="studio-layout-shell">
    <header class="studio-layout-topbar">
      <div class="studio-layout-left">
        <RouterLink
          :to="dashboardPath"
          class="studio-layout-brand"
          :title="dashboardTitle"
        >
          <span class="studio-layout-logo">
            <img :src="siteLogo || '/logo.png'" alt="" />
          </span>
          <span class="studio-layout-brand-copy">
            <span class="studio-layout-site">{{ siteName }}</span>
            <span class="studio-layout-return">{{ dashboardTitle }}</span>
          </span>
        </RouterLink>

        <nav class="studio-layout-nav" :aria-label="t('nav.myAccount')">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="studio-layout-nav-link"
            :title="item.label"
          >
            <Icon :name="item.icon" size="sm" />
            <span>{{ item.label }}</span>
          </RouterLink>
        </nav>
      </div>

      <div class="studio-layout-title">
        <span class="studio-layout-title-icon">
          <Icon name="sparkles" size="sm" />
        </span>
        <span class="studio-layout-title-copy">
          <span>{{ t('chatImageStudio.title') }}</span>
          <small>{{ t('chatImageStudio.subtitle') }}</small>
        </span>
      </div>

      <div class="studio-layout-actions">
        <RouterLink to="/keys" class="studio-layout-icon-action" :title="t('nav.apiKeys')">
          <Icon name="key" size="sm" />
        </RouterLink>

        <div v-if="user" class="studio-layout-balance">
          <Icon name="dollar" size="sm" />
          <span>${{ user.balance?.toFixed(2) || '0.00' }}</span>
        </div>

        <LocaleSwitcher class="studio-layout-locale" />

        <div v-if="user" ref="dropdownRef" class="studio-layout-user">
          <button
            type="button"
            class="studio-layout-user-button"
            aria-label="User Menu"
            @click="toggleDropdown"
          >
            <span class="studio-layout-avatar">
              <img v-if="avatarUrl" :src="avatarUrl" :alt="displayName" />
              <span v-else>{{ userInitials }}</span>
            </span>
            <span class="studio-layout-user-copy">
              <span>{{ displayName }}</span>
              <small>{{ user.role }}</small>
            </span>
            <Icon name="chevronDown" size="xs" class="studio-layout-user-chevron" />
          </button>

          <transition name="dropdown">
            <div v-if="dropdownOpen" class="studio-layout-dropdown">
              <RouterLink :to="dashboardPath" class="studio-layout-dropdown-item" @click="closeDropdown">
                <Icon name="home" size="sm" />
                {{ dashboardTitle }}
              </RouterLink>
              <RouterLink to="/profile" class="studio-layout-dropdown-item" @click="closeDropdown">
                <Icon name="user" size="sm" />
                {{ t('nav.profile') }}
              </RouterLink>
              <RouterLink to="/keys" class="studio-layout-dropdown-item" @click="closeDropdown">
                <Icon name="key" size="sm" />
                {{ t('nav.apiKeys') }}
              </RouterLink>
              <RouterLink to="/usage" class="studio-layout-dropdown-item" @click="closeDropdown">
                <Icon name="chart" size="sm" />
                {{ t('nav.usage') }}
              </RouterLink>
              <button
                type="button"
                class="studio-layout-dropdown-item studio-layout-dropdown-danger"
                @click="handleLogout"
              >
                <Icon name="logout" size="sm" />
                {{ t('nav.logout') }}
              </button>
            </div>
          </transition>
        </div>
      </div>
    </header>

    <main class="studio-layout-content">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

type IconName = InstanceType<typeof Icon>['$props']['name']

interface StudioNavItem {
  path: string
  label: string
  icon: IconName
}

const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const user = computed(() => authStore.user)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => appStore.siteLogo)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const dashboardTitle = computed(() => authStore.isAdmin ? t('admin.dashboard.title') : t('nav.dashboard'))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')

const navItems = computed<StudioNavItem[]>(() => [
  { path: dashboardPath.value, label: dashboardTitle.value, icon: 'home' },
  { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
  { path: '/usage', label: t('nav.usage'), icon: 'chart' },
])

const userInitials = computed(() => {
  if (!user.value) return ''
  if (user.value.username) return user.value.username.substring(0, 2).toUpperCase()
  if (user.value.email) return user.value.email.split('@')[0].substring(0, 2).toUpperCase()
  return ''
})

const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

function toggleDropdown(): void {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown(): void {
  dropdownOpen.value = false
}

async function handleLogout(): Promise<void> {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleClickOutside(event: MouseEvent): void {
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
.studio-layout-shell {
  min-height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  background: linear-gradient(135deg, rgb(2 6 23) 0%, rgb(15 23 42) 52%, rgb(6 18 39) 100%);
  color: rgb(226 232 240);
}

.studio-layout-topbar {
  position: sticky;
  top: 0;
  z-index: 40;
  display: grid;
  min-height: 4rem;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 1rem;
  border-bottom: 1px solid rgb(148 163 184 / 0.18);
  background: rgb(2 6 23 / 0.82);
  padding: 0.625rem 1rem;
  backdrop-filter: blur(18px);
}

.studio-layout-left,
.studio-layout-actions,
.studio-layout-brand,
.studio-layout-title,
.studio-layout-nav,
.studio-layout-nav-link,
.studio-layout-icon-action,
.studio-layout-balance,
.studio-layout-user-button {
  display: flex;
  align-items: center;
}

.studio-layout-left {
  min-width: 0;
  gap: 0.75rem;
}

.studio-layout-brand {
  min-width: 0;
  gap: 0.625rem;
  border-radius: 0.5rem;
  padding: 0.25rem;
  color: rgb(248 250 252);
  transition: background-color 0.15s ease;
}

.studio-layout-brand:hover {
  background: rgb(255 255 255 / 0.07);
}

.studio-layout-logo {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.24);
  border-radius: 0.5rem;
  background: rgb(15 23 42 / 0.78);
}

.studio-layout-logo img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.studio-layout-brand-copy,
.studio-layout-title-copy,
.studio-layout-user-copy {
  display: grid;
  min-width: 0;
  gap: 0.0625rem;
}

.studio-layout-site,
.studio-layout-title-copy > span,
.studio-layout-user-copy > span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 700;
  line-height: 1.1;
}

.studio-layout-site {
  max-width: 8rem;
  font-size: 0.9rem;
}

.studio-layout-return,
.studio-layout-title-copy small,
.studio-layout-user-copy small {
  overflow: hidden;
  color: rgb(148 163 184);
  font-size: 0.7rem;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-layout-nav {
  min-width: 0;
  gap: 0.375rem;
}

.studio-layout-nav-link,
.studio-layout-icon-action,
.studio-layout-balance,
.studio-layout-user-button {
  border: 1px solid rgb(148 163 184 / 0.22);
  border-radius: 0.5rem;
  background: rgb(15 23 42 / 0.6);
  color: rgb(203 213 225);
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
}

.studio-layout-nav-link {
  min-height: 2.25rem;
  gap: 0.45rem;
  padding: 0 0.75rem;
  font-size: 0.82rem;
  font-weight: 650;
}

.studio-layout-nav-link:hover,
.studio-layout-icon-action:hover,
.studio-layout-user-button:hover {
  border-color: rgb(52 211 153 / 0.45);
  background: rgb(16 185 129 / 0.12);
  color: rgb(236 253 245);
}

.studio-layout-title {
  min-width: 0;
  justify-content: center;
  gap: 0.625rem;
  color: rgb(248 250 252);
}

.studio-layout-title-icon {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgb(16 185 129 / 0.15);
  color: rgb(110 231 183);
}

.studio-layout-title-copy {
  text-align: center;
}

.studio-layout-title-copy > span {
  max-width: 16rem;
  font-size: 0.95rem;
}

.studio-layout-title-copy small {
  max-width: 20rem;
}

.studio-layout-actions {
  min-width: 0;
  justify-content: flex-end;
  gap: 0.5rem;
}

.studio-layout-icon-action {
  height: 2.25rem;
  width: 2.25rem;
  justify-content: center;
}

.studio-layout-balance {
  min-height: 2.25rem;
  gap: 0.375rem;
  padding: 0 0.75rem;
  color: rgb(134 239 172);
  font-size: 0.85rem;
  font-weight: 750;
}

.studio-layout-user {
  position: relative;
}

.studio-layout-user-button {
  min-height: 2.5rem;
  gap: 0.5rem;
  padding: 0.25rem 0.5rem 0.25rem 0.25rem;
}

.studio-layout-avatar {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.5rem;
  background: linear-gradient(135deg, rgb(59 130 246), rgb(16 185 129));
  color: white;
  font-size: 0.75rem;
  font-weight: 800;
}

.studio-layout-avatar img {
  height: 100%;
  width: 100%;
  object-fit: cover;
}

.studio-layout-user-copy {
  max-width: 8.5rem;
  text-align: left;
}

.studio-layout-user-chevron {
  flex-shrink: 0;
}

.studio-layout-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  z-index: 60;
  width: 13rem;
  overflow: hidden;
  border: 1px solid rgb(148 163 184 / 0.24);
  border-radius: 0.75rem;
  background: rgb(15 23 42 / 0.96);
  box-shadow: 0 20px 48px rgb(0 0 0 / 0.32);
}

.studio-layout-dropdown-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.625rem;
  padding: 0.7rem 0.85rem;
  color: rgb(226 232 240);
  font-size: 0.875rem;
  text-align: left;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.studio-layout-dropdown-item:hover {
  background: rgb(255 255 255 / 0.08);
  color: rgb(255 255 255);
}

.studio-layout-dropdown-danger {
  color: rgb(252 165 165);
}

.studio-layout-content {
  height: calc(100vh - 4rem);
  height: calc(100dvh - 4rem);
  min-height: 0;
  overflow: hidden;
  padding: 0.75rem;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.16s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.96) translateY(-0.25rem);
}

@media (max-width: 1180px) {
  .studio-layout-topbar {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .studio-layout-title {
    display: none;
  }
}

@media (max-width: 920px) {
  .studio-layout-brand-copy,
  .studio-layout-nav-link span,
  .studio-layout-locale {
    display: none;
  }

  .studio-layout-nav-link {
    width: 2.25rem;
    justify-content: center;
    padding: 0;
  }
}

@media (max-width: 640px) {
  .studio-layout-topbar {
    gap: 0.5rem;
    padding-inline: 0.625rem;
  }

  .studio-layout-nav {
    display: none;
  }

  .studio-layout-balance {
    padding-inline: 0.55rem;
  }

  .studio-layout-user-copy,
  .studio-layout-user-chevron {
    display: none;
  }

  .studio-layout-user-button {
    padding-right: 0.25rem;
  }

  .studio-layout-content {
    padding: 0.5rem;
  }
}
</style>
