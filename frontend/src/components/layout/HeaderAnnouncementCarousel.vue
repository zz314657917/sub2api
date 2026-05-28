<template>
  <button
    v-if="currentAnnouncement"
    type="button"
    class="header-announcement-carousel"
    data-testid="header-announcement-carousel"
    :title="currentTitle"
    :aria-label="`${t('announcements.title')}: ${currentAnnouncement.title}`"
    @click="$emit('select', currentAnnouncement)"
    @mouseenter="pause"
    @mouseleave="resume"
    @focus="pause"
    @blur="resume"
  >
    <span class="header-announcement-label">{{ t('announcements.title') }}</span>
    <span class="header-announcement-divider" aria-hidden="true"></span>

    <span class="header-announcement-viewport" aria-live="polite">
      <Transition name="header-announcement-slide" mode="out-in">
        <span :key="currentAnnouncement.id" class="header-announcement-content">
          <span class="header-announcement-title">{{ currentAnnouncement.title }}</span>
        </span>
      </Transition>
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserAnnouncement } from '@/types'

const ROTATE_INTERVAL_MS = 5000
const MAX_ITEMS = 3

const props = defineProps<{
  announcements: UserAnnouncement[]
}>()

defineEmits<{
  (event: 'select', announcement: UserAnnouncement): void
}>()

const { t } = useI18n()
const currentIndex = ref(0)
const paused = ref(false)
let timer: number | null = null

const visibleAnnouncements = computed(() =>
  [...props.announcements]
    .sort((a, b) => getAnnouncementTime(b) - getAnnouncementTime(a))
    .slice(0, MAX_ITEMS)
)

const currentAnnouncement = computed(() => {
  const items = visibleAnnouncements.value
  if (items.length === 0) return null
  return items[Math.min(currentIndex.value, items.length - 1)]
})

const currentTitle = computed(() => {
  if (!currentAnnouncement.value) return t('announcements.title')
  return currentAnnouncement.value.title
})

function getAnnouncementTime(announcement: UserAnnouncement) {
  const timestamp = Date.parse(announcement.created_at)
  return Number.isFinite(timestamp) ? timestamp : 0
}

function clearTimer() {
  if (timer === null) return
  window.clearInterval(timer)
  timer = null
}

function startTimer() {
  clearTimer()
  if (paused.value || visibleAnnouncements.value.length < 2) return

  timer = window.setInterval(() => {
    const total = visibleAnnouncements.value.length
    if (total === 0) {
      currentIndex.value = 0
      return
    }
    currentIndex.value = (currentIndex.value + 1) % total
  }, ROTATE_INTERVAL_MS)
}

function pause() {
  paused.value = true
}

function resume() {
  paused.value = false
}

watch(
  () => visibleAnnouncements.value.length,
  (total) => {
    if (currentIndex.value >= total) {
      currentIndex.value = 0
    }
    startTimer()
  },
  { immediate: true }
)

watch(paused, startTimer)

onBeforeUnmount(clearTimer)
</script>

<style scoped>
.header-announcement-carousel {
  display: flex;
  width: 100%;
  max-width: 680px;
  height: 36px;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  overflow: hidden;
  border: 0;
  background: transparent;
  color: #475569;
  padding: 0;
  transition:
    color 0.18s ease;
}

.header-announcement-carousel:hover,
.header-announcement-carousel:focus-visible {
  color: #111827;
  outline: none;
}

.header-announcement-label {
  flex-shrink: 0;
  color: #4338ca;
  padding: 0;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1rem;
}

.header-announcement-divider {
  width: 1px;
  height: 1rem;
  flex-shrink: 0;
  background: rgba(148, 163, 184, 0.5);
}

.header-announcement-viewport {
  display: flex;
  min-width: 0;
  flex: 0 1 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.header-announcement-content {
  display: flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.header-announcement-title {
  max-width: min(36rem, 100%);
  flex-shrink: 0;
  overflow: hidden;
  color: #111827;
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.1rem;
  text-overflow: ellipsis;
}

.header-announcement-slide-enter-active,
.header-announcement-slide-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}

.header-announcement-slide-enter-from {
  opacity: 0;
  transform: translateY(7px);
}

.header-announcement-slide-leave-to {
  opacity: 0;
  transform: translateY(-7px);
}

.dark .header-announcement-carousel {
  color: #cbd5e1;
}

.dark .header-announcement-carousel:hover,
.dark .header-announcement-carousel:focus-visible {
  color: #f8fafc;
}

.dark .header-announcement-label {
  color: #c7d2fe;
}

.dark .header-announcement-divider {
  background: rgba(71, 85, 105, 0.86);
}

.dark .header-announcement-title {
  color: #f8fafc;
}

@media (max-width: 1535px) {
  .header-announcement-title {
    max-width: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .header-announcement-slide-enter-active,
  .header-announcement-slide-leave-active {
    transition: none;
  }
}
</style>
