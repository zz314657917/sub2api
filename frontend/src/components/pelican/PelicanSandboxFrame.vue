<template><iframe ref="frame" :key="documentIdentity.token" :class="['pelican-frame', { 'pelican-frame--interactive': interactive }]" :style="interactive ? undefined : { height: `${frameHeight}px` }" :sandbox="interactive ? '' : 'allow-scripts'" :scrolling="interactive ? 'auto' : 'no'" :tabindex="interactive ? undefined : -1" :aria-hidden="interactive ? undefined : 'true'" :srcdoc="safeDocument" title="鹈鹕测试作品" /></template>

<script setup lang="ts">
import { computed, onBeforeMount, onBeforeUnmount, ref, watch } from 'vue'
import { toPelicanSrcdoc } from './sanitizePelicanHtml'

const props = withDefaults(defineProps<{ html: string; replayKey: number; interactive?: boolean; minHeight?: number }>(), { interactive: false, minHeight: 220 })
const frame = ref<HTMLIFrameElement>()
const frameHeight = ref(props.minHeight)
const freshSecret = () => Array.from(crypto.getRandomValues(new Uint8Array(24)), byte => byte.toString(16).padStart(2, '0')).join('')
const outerNonce = () => {
  const script = document.querySelector('script[nonce]') as HTMLScriptElement | null
  const nonce = script?.nonce || script?.getAttribute('nonce') || ''
  return /^[A-Za-z0-9+/_=-]{16,256}$/.test(nonce) ? nonce : freshSecret()
}
const documentIdentity = ref({ nonce: outerNonce(), token: freshSecret() })
let lastSequence = 0
watch(() => [props.html, props.replayKey, props.interactive], () => {
  documentIdentity.value = { nonce: outerNonce(), token: freshSecret() }
  lastSequence = 0
  frameHeight.value = props.minHeight
}, { flush: 'sync' })
const safeDocument = computed(() => toPelicanSrcdoc(props.html, !props.interactive, documentIdentity.value))
function receiveHeight(event: MessageEvent) {
  const data = event.data
  if (props.interactive || !frame.value || event.source !== frame.value.contentWindow ||
      !data || typeof data !== 'object' || data.type !== 'pelican-preview-height' ||
      data.token !== documentIdentity.value.token ||
      typeof data.height !== 'number' || !Number.isFinite(data.height) || data.height <= 0 || data.height > 100000 ||
      !Number.isSafeInteger(data.sequence) || data.sequence <= lastSequence) return
  lastSequence = data.sequence
  frameHeight.value = Math.max(1, Math.min(580, Math.ceil(data.height)))
}
// Register before srcdoc can execute; only the current opaque frame may report.
onBeforeMount(() => window.addEventListener('message', receiveHeight))
onBeforeUnmount(() => window.removeEventListener('message', receiveHeight))
</script>

<style scoped>
.pelican-frame { display:block; width:100%; max-height:580px; border:0; border-radius:8px; background:#eef5f6; }
.pelican-frame--interactive { height: clamp(360px, 72dvh, 760px); min-height:220px; max-height: none; }
</style>
