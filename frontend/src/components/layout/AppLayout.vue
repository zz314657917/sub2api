<template>
  <div class="console-shell" :class="{ 'console-shell-dashboard': isUserDashboard, 'console-shell-viewport': viewport }">

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="console-main min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64', { 'console-main-viewport': viewport }]"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main
        class="console-content"
        :class="{
          'console-content-dense': route.meta.denseWorkspace,
          'console-content-dashboard': isUserDashboard,
          'console-content-viewport': viewport
        }"
      >
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

withDefaults(defineProps<{ viewport?: boolean }>(), { viewport: false })

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const isUserDashboard = computed(() => route.name === 'Dashboard')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>

<style>
.console-shell-viewport { height: 100dvh; min-height: 0; overflow: hidden; }
.console-main-viewport { display: flex; min-height: 0; height: 100%; flex-direction: column; }
.console-content-viewport { display: flex; min-height: 0; flex: 1 1 auto; overflow: hidden; }
.console-content-viewport > * { min-height: 0; width: 100%; }
</style>
