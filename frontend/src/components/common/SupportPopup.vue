<template>
  <Teleport to="body">
    <Transition name="support-popup-fade">
      <div
        v-if="show && hasItems"
        class="support-popup-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="popupTitle"
        @click.self="close"
      >
        <div class="support-popup-panel">
          <button
            type="button"
            class="support-popup-close"
            :aria-label="t('common.close')"
            @click="close"
          >
            <Icon name="x" size="md" />
          </button>

          <div class="support-popup-header">
            <h2>{{ popupTitle }}</h2>
            <p v-if="popupDescription">{{ popupDescription }}</p>
          </div>

          <div class="support-popup-grid">
            <article
              v-for="item in popupItems"
              :key="item.id || item.image_url"
              class="support-popup-card"
            >
              <div class="support-popup-image-wrap">
                <img :src="item.image_url" :alt="item.title" class="support-popup-image" />
                <span v-if="item.badge" class="support-popup-badge">{{ item.badge }}</span>
              </div>
              <h3>{{ item.title }}</h3>
              <p v-if="item.caption">{{ item.caption }}</p>
            </article>
          </div>

          <p v-if="popupFooter" class="support-popup-footer">{{ popupFooter }}</p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Icon from '@/components/icons/Icon.vue'
import type { SupportPopupItem } from '@/types'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const popupTitle = computed(() =>
  (appStore.cachedPublicSettings?.support_popup_title || '').trim() || t('home.contactSupport')
)
const popupDescription = computed(() =>
  (appStore.cachedPublicSettings?.support_popup_description || '').trim()
)
const popupFooter = computed(() =>
  (appStore.cachedPublicSettings?.support_popup_footer || '').trim()
)

const popupItems = computed<SupportPopupItem[]>(() => {
  const raw = appStore.cachedPublicSettings?.support_popup_items
  if (!Array.isArray(raw)) return []
  return raw
    .map((item) => ({
      id: item.id || '',
      title: (item.title || '').trim(),
      image_url: (item.image_url || '').trim(),
      caption: (item.caption || '').trim(),
      badge: (item.badge || '').trim(),
    }))
    .filter((item) => item.title && item.image_url)
})

const hasItems = computed(() => popupItems.value.length > 0)

function close() {
  emit('close')
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.show) {
    close()
  }
}

watch(
  () => props.show && hasItems.value,
  (visible) => {
    if (typeof document === 'undefined') return
    document.body.classList.toggle('support-popup-open', visible)
  },
  { immediate: true },
)

watch(
  hasItems,
  (available) => {
    if (!available && props.show) {
      close()
    }
  },
)

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  if (typeof document === 'undefined') return
  document.removeEventListener('keydown', handleKeydown)
  document.body.classList.remove('support-popup-open')
})
</script>

<style scoped>
.support-popup-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(17, 24, 39, 0.58);
  padding: 1.25rem;
  backdrop-filter: blur(8px);
}

.support-popup-panel {
  position: relative;
  width: min(100%, 42rem);
  max-height: min(88vh, 46rem);
  overflow-y: auto;
  border-radius: 1rem;
  background: rgba(248, 250, 252, 0.96);
  padding: 1.75rem;
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.36);
}

.dark .support-popup-panel {
  background: rgba(15, 23, 42, 0.96);
  color: #f8fafc;
}

.support-popup-close {
  position: absolute;
  right: 1rem;
  top: 1rem;
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  color: #64748b;
  transition: background 160ms ease, color 160ms ease;
}

.support-popup-close:hover {
  background: rgba(148, 163, 184, 0.18);
  color: #0f172a;
}

.dark .support-popup-close:hover {
  color: #f8fafc;
}

.support-popup-header {
  padding: 0.25rem 2.5rem 1rem;
  text-align: center;
}

.support-popup-header h2 {
  color: #111827;
  font-size: 1.32rem;
  font-weight: 800;
  line-height: 1.2;
}

.dark .support-popup-header h2 {
  color: #f8fafc;
}

.support-popup-header p {
  margin-top: 0.45rem;
  color: #64748b;
  font-size: 0.9rem;
}

.dark .support-popup-header p {
  color: #cbd5e1;
}

.support-popup-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 1.1rem;
}

.support-popup-card {
  min-width: 0;
  text-align: center;
}

.support-popup-image-wrap {
  position: relative;
  display: flex;
  aspect-ratio: 1 / 1;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.42);
  border-radius: 0.65rem;
  background: white;
}

.support-popup-image {
  height: 100%;
  width: 100%;
  object-fit: contain;
}

.support-popup-badge {
  position: absolute;
  left: 50%;
  top: 50%;
  max-width: calc(100% - 2rem);
  transform: translate(-50%, -50%);
  border-radius: 999px;
  background: rgba(220, 38, 38, 0.92);
  padding: 0.3rem 0.7rem;
  color: white;
  font-size: 0.8rem;
  font-weight: 800;
  line-height: 1.1;
  white-space: nowrap;
}

.support-popup-card h3 {
  margin-top: 0.55rem;
  color: #2563eb;
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.dark .support-popup-card h3 {
  color: #93c5fd;
}

.support-popup-card p {
  margin-top: 0.25rem;
  color: #64748b;
  font-size: 0.78rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.dark .support-popup-card p {
  color: #cbd5e1;
}

.support-popup-footer {
  margin-top: 1.35rem;
  color: #64748b;
  font-size: 0.85rem;
  text-align: center;
}

.dark .support-popup-footer {
  color: #cbd5e1;
}

.support-popup-fade-enter-active,
.support-popup-fade-leave-active {
  transition: opacity 180ms ease;
}

.support-popup-fade-enter-from,
.support-popup-fade-leave-to {
  opacity: 0;
}

.support-popup-fade-enter-active .support-popup-panel,
.support-popup-fade-leave-active .support-popup-panel {
  transition: transform 180ms ease;
}

.support-popup-fade-enter-from .support-popup-panel,
.support-popup-fade-leave-to .support-popup-panel {
  transform: translateY(0.5rem) scale(0.98);
}

@media (max-width: 640px) {
  .support-popup-overlay {
    align-items: flex-end;
    padding: 0.75rem;
  }

  .support-popup-panel {
    max-height: 86vh;
    padding: 1.25rem;
  }

  .support-popup-header {
    padding-inline: 2rem;
  }

  .support-popup-grid {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
body.support-popup-open {
  overflow: hidden;
}
</style>
