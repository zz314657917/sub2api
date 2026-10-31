<template>
  <div ref="host" class="lazy-artwork" :style="{ minHeight: `${height}px` }">
    <button v-if="html && visible && documentVisible" class="artwork-preview" type="button" aria-label="放大查看作品" @click="$emit('enlarge', $event)">
      <PelicanSandboxFrame :html="html" :replay-key="replayKey" :interactive="interactive" :min-height="minHeight" />
      <span v-if="!interactive" class="preview-overlay" />
    </button>
    <div v-else class="artwork-placeholder" role="status">
      {{ failed ? '作品加载失败' : loading ? '正在加载作品…' : '作品等待进入可视区域' }}
      <button v-if="failed" type="button" aria-label="重新加载作品" @click="retry">重新加载</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import PelicanSandboxFrame from './PelicanSandboxFrame.vue'
import { usePelicanArtwork } from './usePelicanArtwork'

const props = withDefaults(defineProps<{ resultId: number; replayKey?: number; root?: HTMLElement | null; interactive?: boolean; minHeight?: number }>(), { replayKey: 0, root: null, interactive: false, minHeight: 220 })
defineEmits<{ enlarge: [event: MouseEvent] }>()
const artwork = usePelicanArtwork()
const host = ref<HTMLElement>()
const visible = ref(false)
const near = ref(false)
const documentVisible = ref(!document.hidden)
const html = ref('')
const loading = ref(false)
const failed = ref(false)
const height = ref(props.minHeight)
let controller: AbortController | undefined
let requestVersion = 0
let displayObserver: IntersectionObserver | undefined
let prefetchObserver: IntersectionObserver | undefined
let resizeObserver: ResizeObserver | undefined
let disposed = false
let listeningRoot: HTMLElement | null = null

function stop() { requestVersion++; controller?.abort(); controller = undefined; loading.value = false }
async function load() {
  if (disposed || loading.value || failed.value || html.value || !near.value || !documentVisible.value || !artwork.allowed.value) return
  const cached = artwork.peek(props.resultId)
  if (cached) { if (visible.value) html.value = cached.result.html || ''; height.value = cached.height ?? height.value; return }
  const version = ++requestVersion
  controller = new AbortController()
  loading.value = true
  try {
    const result = await artwork.request(props.resultId, controller.signal)
    if (version === requestVersion && !disposed && visible.value && documentVisible.value) html.value = result.html || ''
  } catch (reason) {
    if (version === requestVersion && (reason as { name?: string })?.name !== 'AbortError') failed.value = true
  } finally { if (version === requestVersion) loading.value = false }
}
function update() {
  if (!host.value) return
  const rect = host.value.getBoundingClientRect()
  const root = props.root?.getBoundingClientRect()
  let top = Math.max(0, root?.top ?? 0)
  let bottom = Math.min(window.innerHeight, root?.bottom ?? window.innerHeight)
  let left = Math.max(0, root?.left ?? 0)
  let right = Math.min(window.innerWidth, root?.right ?? window.innerWidth)
  for (let parent = host.value.parentElement; parent; parent = parent.parentElement) {
    const style = getComputedStyle(parent)
    if (/(auto|scroll|hidden|clip)/.test(`${style.overflow} ${style.overflowY} ${style.overflowX}`)) {
      const bounds = parent.getBoundingClientRect()
      top = Math.max(top, bounds.top); bottom = Math.min(bottom, bounds.bottom)
      left = Math.max(left, bounds.left); right = Math.min(right, bounds.right)
    }
  }
  const intersects = (margin: number) => bottom > top && right > left && rect.bottom > top - margin && rect.top < bottom + margin && rect.right > left && rect.left < right
  visible.value = intersects(0)
  near.value = intersects(240)
  if (!visible.value || !documentVisible.value) html.value = ''
  if (!near.value || !documentVisible.value) stop()
  else void load()
}
function retry() { failed.value = false; void load() }
function visibilityChanged() { documentVisible.value = !document.hidden; update() }
function measure() {
  // Read only the sandbox's validated host dimensions; never trust window messages here.
  const frame = host.value?.querySelector('iframe')
  if (frame && !props.interactive) {
    const measured = frame.getBoundingClientRect().height
    if (measured > 0) { height.value = Math.min(580, Math.ceil(measured)); artwork.setHeight(props.resultId, height.value) }
  }
  update()
}
function observe() {
  displayObserver?.disconnect(); prefetchObserver?.disconnect()
  listeningRoot?.removeEventListener('scroll', update)
  listeningRoot = props.root
  listeningRoot?.addEventListener('scroll', update, { passive: true })
  if (typeof IntersectionObserver !== 'undefined' && host.value) {
    displayObserver = new IntersectionObserver(update, { root: props.root, rootMargin: '0px' })
    prefetchObserver = new IntersectionObserver(update, { root: props.root, rootMargin: '240px 0px' })
    displayObserver.observe(host.value); prefetchObserver.observe(host.value)
  }
  update()
}
watch(() => props.root, () => observe())
watch(() => [props.resultId, artwork.revision.value, artwork.allowed.value], () => {
  stop(); html.value = ''
  height.value = artwork.height(props.resultId) ?? props.minHeight
  void nextTick(update)
}, { flush: 'sync' })
watch(() => props.resultId, () => { failed.value = false })
watch(html, () => void nextTick(() => {
  resizeObserver?.disconnect()
  const frame = host.value?.querySelector('iframe')
  if (frame) resizeObserver?.observe(frame)
  measure()
}))
onMounted(() => {
  if (typeof ResizeObserver !== 'undefined') resizeObserver = new ResizeObserver(measure)
  document.addEventListener('visibilitychange', visibilityChanged)
  // Capture scroll also covers dialog scroll ancestors; geometry fallback works without IO.
  window.addEventListener('scroll', update, true)
  window.addEventListener('resize', update)
  observe()
})
onBeforeUnmount(() => {
  disposed = true; stop(); html.value = ''
  displayObserver?.disconnect(); prefetchObserver?.disconnect(); resizeObserver?.disconnect()
  listeningRoot?.removeEventListener('scroll', update)
  document.removeEventListener('visibilitychange', visibilityChanged)
  window.removeEventListener('scroll', update, true)
  window.removeEventListener('resize', update)
})
</script>

<style scoped>
.lazy-artwork { width: 100%; min-width: 0; }
.artwork-preview { position: relative; display: block; width: 100%; overflow: hidden; border: 0; border-radius: 8px; padding: 0; cursor: zoom-in; background: var(--pelican-accent-soft); }
.artwork-preview :deep(.pelican-frame:not(.pelican-frame--interactive)) { pointer-events: none; }
.preview-overlay { position: absolute; inset: 0; }
.artwork-placeholder { display: flex; min-height: inherit; align-items: center; justify-content: center; gap: 10px; border-radius: 8px; color: var(--pelican-muted); background: var(--pelican-accent-soft); }
.artwork-placeholder button { color: var(--pelican-accent); background: transparent; border: 0; cursor: pointer; }
</style>
