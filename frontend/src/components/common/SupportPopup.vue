<template>
  <Teleport to="body">
    <Transition name="support-popup-fade">
      <div
        v-if="show && hasContent"
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
              <p class="support-popup-card-caption">
                <span>{{ item.title }}</span>
                <span v-if="item.caption">{{ item.caption }}</span>
              </p>
            </article>

            <article v-if="contactCardText" class="support-popup-card support-popup-contact-card">
              <div class="support-popup-contact-icon">
                <Icon name="chatBubble" size="xl" />
              </div>
              <h3>{{ t('common.contactSupport') }}</h3>
              <p>{{ contactCardText }}</p>
            </article>
          </div>

          <p v-if="contactTextBelowImages" class="support-popup-contact-text">{{ contactTextBelowImages }}</p>
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
const contactInfo = computed(() =>
  (appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '').trim()
)
const contactCardText = computed(() => {
  if (popupItems.value.length > 0) return ''
  return contactInfo.value
})
const contactTextBelowImages = computed(() => {
  if (popupItems.value.length === 0) return ''
  const allCardsHaveCaption = popupItems.value.every((item) => item.caption)
  if (allCardsHaveCaption) return ''
  return (appStore.cachedPublicSettings?.contact_info || appStore.contactInfo || '').trim()
})

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
const hasContent = computed(() => hasItems.value || Boolean(contactCardText.value))

function close() {
  emit('close')
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.show) {
    close()
  }
}

watch(
  () => props.show && hasContent.value,
  (visible) => {
    if (typeof document === 'undefined') return
    document.body.classList.toggle('support-popup-open', visible)
  },
  { immediate: true },
)

watch(
  hasContent,
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
  background: rgba(15, 23, 42, 0.58);
  padding: 1.25rem;
  backdrop-filter: blur(8px);
}

.support-popup-panel {
  position: relative;
  width: min(100%, 42rem);
  max-height: min(88vh, 46rem);
  overflow-y: auto;
  border: 1px solid rgba(226, 232, 240, 0.78);
  border-radius: 0.95rem;
  background: rgba(248, 250, 252, 0.98);
  padding: 1.5rem 1.75rem 1.65rem;
  box-shadow: 0 28px 76px rgba(15, 23, 42, 0.34);
}

.dark .support-popup-panel {
  background: rgba(15, 23, 42, 0.96);
  color: #f8fafc;
}

.support-popup-close {
  position: absolute;
  right: 1rem;
  top: 0.92rem;
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
  padding: 0.1rem 2.5rem 1rem;
  text-align: center;
}

.support-popup-header h2 {
  color: #0f172a;
  font-size: 1.32rem;
  font-weight: 900;
  line-height: 1.2;
}

.dark .support-popup-header h2 {
  color: #f8fafc;
}

.support-popup-header p {
  margin-top: 0.48rem;
  color: #64748b;
  font-size: 0.9rem;
}

.dark .support-popup-header p {
  color: #cbd5e1;
}

.support-popup-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12.5rem, 14.2rem));
  justify-content: center;
  gap: 1.35rem 1.45rem;
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
  border: 2px solid rgba(203, 213, 225, 0.9);
  border-radius: 0.78rem;
  background: #f8fafc;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.9), 0 12px 28px rgba(15, 23, 42, 0.06);
  transition: border-color 160ms ease, box-shadow 160ms ease, transform 160ms ease;
}

.support-popup-image-wrap:hover {
  border-color: rgba(59, 130, 246, 0.86);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2), 0 16px 34px rgba(37, 99, 235, 0.14);
  transform: translateY(-1px);
}

.support-popup-image {
  height: calc(100% - 1.25rem);
  width: calc(100% - 1.25rem);
  object-fit: contain;
}

.support-popup-badge {
  position: absolute;
  left: 50%;
  top: 50%;
  max-width: calc(100% - 2rem);
  transform: translate(-50%, -50%);
  border-radius: 999px;
  background: rgba(220, 38, 38, 0.94);
  padding: 0.34rem 0.72rem;
  color: white;
  font-size: 0.8rem;
  font-weight: 800;
  line-height: 1.1;
  white-space: nowrap;
}

.support-popup-card-caption {
  display: grid;
  gap: 0.1rem;
  margin-top: 0.56rem;
  color: #2563eb;
  font-size: 0.86rem;
  font-weight: 800;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.support-popup-card-caption span + span {
  color: #64748b;
  font-size: 0.76rem;
  font-weight: 600;
}

.dark .support-popup-card-caption {
  color: #93c5fd;
}

.dark .support-popup-card-caption span + span {
  color: #cbd5e1;
}

.support-popup-card p {
  color: #64748b;
  font-size: 0.78rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.support-popup-contact-card {
  display: grid;
  justify-items: center;
  align-content: center;
  min-height: 12rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 0.65rem;
  padding: 1rem;
}

.support-popup-contact-text {
  margin-top: 1rem;
  color: #475569;
  font-size: 0.88rem;
  font-weight: 600;
  line-height: 1.5;
  text-align: center;
  overflow-wrap: anywhere;
}

.dark .support-popup-contact-text {
  color: #dbeafe;
}

.support-popup-contact-icon {
  display: inline-flex;
  height: 3.5rem;
  width: 3.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(37, 99, 235, 0.1);
  color: #2563eb;
}

.support-popup-contact-card h3 {
  margin-top: 0.7rem;
  color: #2563eb;
  font-size: 0.9rem;
  font-weight: 800;
}

.dark .support-popup-contact-icon {
  background: rgba(147, 197, 253, 0.14);
  color: #93c5fd;
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
