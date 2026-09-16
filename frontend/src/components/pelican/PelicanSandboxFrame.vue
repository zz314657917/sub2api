<template><iframe :key="replayKey" class="pelican-frame" sandbox="" :srcdoc="safeDocument" title="鹈鹕测试作品" /></template>

<script setup lang="ts">
import { computed } from 'vue'
import { toPelicanSrcdoc } from './sanitizePelicanHtml'

const props = defineProps<{ html: string; replayKey: number }>()
const safeDocument = computed(() => toPelicanSrcdoc(props.html))
</script>

<style scoped>.pelican-frame { display:block; width:100%; aspect-ratio: 4 / 3; border:0; border-radius:8px; background:#eef5f6; }</style>
