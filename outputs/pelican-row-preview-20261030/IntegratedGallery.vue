<template>
  <div class="integrated-gallery" v-html="shell" />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import shell from './integrated-gallery.html?raw'
import './integrated-gallery.css'

onMounted(async () => {
  await import('./integrated-gallery.js')
})
</script>
