<template>
  <div>
    <!-- Tags display -->
    <div class="relative">
      <div
        class="flex min-h-[2.5rem] flex-wrap gap-1.5 rounded-lg border border-gray-200 bg-white p-2 dark:border-dark-600 dark:bg-dark-800"
        @click="focusInput"
      >
        <span
          v-for="(model, idx) in models"
          :key="idx"
          class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-sm"
          :class="getPlatformTagClass(props.platform || '')"
        >
          {{ model }}
          <button
            type="button"
            @click.stop="removeModel(idx)"
            class="ml-0.5 rounded-full p-0.5 hover:bg-primary-200 dark:hover:bg-primary-800"
          >
            <Icon name="x" size="xs" />
          </button>
        </span>
        <input
          ref="inputRef"
          v-model="inputValue"
          type="text"
          class="min-w-[120px] flex-1 border-none bg-transparent text-sm outline-none placeholder:text-gray-400 dark:text-white"
          :placeholder="models.length === 0 ? placeholder : ''"
          @focus="showSuggestions = true"
          @input="showSuggestions = true"
          @keydown.enter.prevent="addModel"
          @keydown.tab.prevent="addModel"
          @keydown.backspace="handleBackspace"
          @keydown.delete="handleBackspace"
          @keydown.escape.prevent="showSuggestions = false"
          @paste="handlePaste"
          @blur="handleBlur"
        />
      </div>

      <div
        v-if="showSuggestionDropdown"
        class="absolute z-50 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
      >
        <div v-if="loadingSuggestions" class="px-3 py-2 text-xs text-gray-400">
          {{ t('admin.channels.form.loadingModelSuggestions', '正在加载可用模型...') }}
        </div>
        <template v-else-if="filteredSuggestions.length > 0">
          <button
            v-for="model in filteredSuggestions"
            :key="model"
            type="button"
            class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
            @mousedown.prevent="selectSuggestion(model)"
          >
            <span class="break-all rounded px-1.5 py-0.5" :class="getPlatformTagClass(props.platform || '')">{{ model }}</span>
          </button>
        </template>
        <div v-else class="px-3 py-2 text-xs text-gray-400">
          {{ t('admin.channels.form.noModelSuggestions', '没有匹配模型，可继续手动输入后按回车添加。') }}
        </div>
      </div>
    </div>
    <p class="mt-1 text-xs text-gray-400">
      {{ t('admin.channels.form.modelInputHint', '可从下拉选择，也可按回车添加，支持粘贴批量导入。') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getPlatformTagClass } from './types'

const { t } = useI18n()

const props = defineProps<{
  models: string[]
  placeholder?: string
  platform?: string
  suggestions?: string[]
  loadingSuggestions?: boolean
}>()

const emit = defineEmits<{
  'update:models': [models: string[]]
}>()

const inputValue = ref('')
const inputRef = ref<HTMLInputElement>()
const showSuggestions = ref(false)

const filteredSuggestions = computed(() => {
  const q = inputValue.value.trim().toLowerCase()
  const selected = new Set(props.models)
  return [...new Set(props.suggestions || [])]
    .filter(model => model && !selected.has(model))
    .filter(model => !q || model.toLowerCase().includes(q))
    .slice(0, 30)
})

const loadingSuggestions = computed(() => props.loadingSuggestions === true)
const showSuggestionDropdown = computed(() => {
  return showSuggestions.value && (loadingSuggestions.value || filteredSuggestions.value.length > 0 || inputValue.value.trim() !== '')
})

function focusInput() {
  inputRef.value?.focus()
}

function addModel() {
  const val = inputValue.value.trim()
  if (!val) return
  addModelValue(val)
  inputValue.value = ''
  showSuggestions.value = false
}

function addModelValue(val: string) {
  if (!props.models.includes(val)) {
    emit('update:models', [...props.models, val])
  }
}

function selectSuggestion(model: string) {
  addModelValue(model)
  inputValue.value = ''
  showSuggestions.value = false
  focusInput()
}

function handleBlur() {
  window.setTimeout(() => {
    addModel()
    showSuggestions.value = false
  }, 120)
}

function removeModel(idx: number) {
  const newModels = [...props.models]
  newModels.splice(idx, 1)
  emit('update:models', newModels)
}

function handleBackspace() {
  if (inputValue.value === '' && props.models.length > 0) {
    removeModel(props.models.length - 1)
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text') || ''
  const items = text.split(/[,\n;]+/).map(s => s.trim()).filter(Boolean)
  if (items.length === 0) return
  const unique = [...new Set([...props.models, ...items])]
  emit('update:models', unique)
  inputValue.value = ''
  showSuggestions.value = false
}
</script>
