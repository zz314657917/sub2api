<template>
  <div class="history-preview">
    <button v-if="html" class="preview" type="button" aria-label="放大历史作品" @click="$emit('enlarge', $event)">
      <PelicanSandboxFrame :html="html" :replay-key="replayKey" />
      <span class="overlay" />
    </button>
    <div v-else class="placeholder" role="status">
      {{ loading ? '动画加载中…' : '作品加载失败' }}
      <button v-if="!loading" type="button" @click="load">重新加载</button>
    </div>
    <div v-if="html" class="actions">
      <button type="button" @click="$emit('enlarge', $event)">放大播放</button>
      <button type="button" @click="replayKey++">重新播放</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { getPelicanResult } from '@/api/pelicanTests'
import PelicanSandboxFrame from './PelicanSandboxFrame.vue'

const props = defineProps<{ resultId: number }>()
defineEmits<{ enlarge: [event: MouseEvent] }>()
const html = ref('')
const loading = ref(false)
const replayKey = ref(0)
let controller: AbortController | undefined
async function load(): Promise<void> {
  controller?.abort()
  const request = new AbortController()
  controller = request
  html.value = ''; loading.value = true
  try {
    const result = await getPelicanResult(props.resultId, request.signal)
    if (!request.signal.aborted && result.status === 'success') html.value = result.html || ''
  } catch { /* Keep failures local to this preview; allow an explicit retry. */ }
  finally { if (!request.signal.aborted) loading.value = false }
}
watch(() => props.resultId, () => void load(), { immediate: true })
onBeforeUnmount(() => controller?.abort())
</script>

<style scoped>
.preview, .placeholder { width: 100%; height: 240px; border: 1px solid var(--pelican-border); border-radius: 10px; overflow: hidden; background: var(--pelican-input); }
.preview { position: relative; display: block; padding: 0; cursor: zoom-in; }
.overlay { position: absolute; inset: 0; }
.preview :deep(iframe) { height: 100%; aspect-ratio: auto; }
.placeholder { display: flex; align-items: center; justify-content: center; gap: 12px; color: var(--pelican-muted); }
.actions { display: flex; gap: 14px; margin-top: 8px; }
.actions button, .placeholder button { background: none; border: 0; color: var(--pelican-accent); cursor: pointer; font: inherit; font-size: 12px; padding: 4px 0; }
.preview:focus-visible, button:focus-visible { outline: 2px solid var(--pelican-accent); outline-offset: 2px; }
</style>
