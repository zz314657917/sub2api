<template>
  <BaseDialog
    :show="show"
    :title="t('admin.tickets.createForUser.title')"
    width="normal"
    @close="emit('close')"
  >
    <form v-if="user" id="user-ticket-message-form" class="space-y-5" @submit.prevent="handleSubmit">
      <div class="flex items-center gap-3 rounded-lg bg-gray-50 p-4 dark:bg-dark-700">
        <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
          <span class="text-sm font-semibold text-primary-700 dark:text-primary-300">
            {{ user.email.charAt(0).toUpperCase() }}
          </span>
        </div>
        <div class="min-w-0">
          <p class="truncate font-medium text-gray-900 dark:text-white">{{ user.email }}</p>
          <p class="text-sm text-gray-500 dark:text-dark-400">#{{ user.id }} · {{ user.username || '-' }}</p>
        </div>
      </div>

      <div>
        <label class="input-label">{{ t('admin.tickets.createForUser.subject') }}</label>
        <input
          v-model.trim="form.title"
          class="input"
          :placeholder="t('admin.tickets.createForUser.subjectPlaceholder')"
          required
        />
      </div>

      <div>
        <label class="input-label">{{ t('admin.tickets.createForUser.content') }}</label>
        <textarea
          v-model.trim="form.content"
          rows="6"
          class="input resize-y"
          :placeholder="t('admin.tickets.createForUser.contentPlaceholder')"
          required
        ></textarea>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="user-ticket-message-form" class="btn btn-primary" :disabled="submitting">
          <Icon name="mail" size="sm" />
          {{ submitting ? t('common.saving') : t('admin.tickets.createForUser.action') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { useAppStore } from '@/stores/app'
import type { AdminUser } from '@/types'

const props = defineProps<{
  show: boolean
  user: AdminUser | null
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()
const submitting = ref(false)
const form = reactive({
  title: '',
  content: '',
})

watch(
  () => props.show,
  (show) => {
    if (show) {
      form.title = ''
      form.content = ''
    }
  }
)

async function handleSubmit() {
  if (!props.user) return
  if (!form.title.trim()) {
    appStore.showError(t('tickets.titleRequired'))
    return
  }
  if (!form.content.trim()) {
    appStore.showError(t('admin.tickets.contentRequired'))
    return
  }

  submitting.value = true
  try {
    await adminTicketsAPI.createForUser(props.user.id, {
      title: form.title.trim(),
      content: form.content.trim(),
    })
    appStore.showSuccess(t('admin.tickets.createForUser.success'))
    emit('success')
    emit('close')
  } catch (error: any) {
    appStore.showError(error.message || t('admin.tickets.createForUser.failed'))
  } finally {
    submitting.value = false
  }
}
</script>
