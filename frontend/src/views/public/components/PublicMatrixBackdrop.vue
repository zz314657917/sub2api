<template>
  <div class="public-matrix-rain pointer-events-none absolute inset-0" aria-hidden="true">
    <span
      v-for="column in columns"
      :key="column.left"
      class="public-matrix-column"
      :style="{
        left: column.left,
        animationDelay: column.delay,
        animationDuration: column.duration
      }"
    >
      {{ column.text }}
    </span>
  </div>
  <div class="public-blur-field pointer-events-none absolute inset-0"></div>
  <div class="public-noise pointer-events-none absolute inset-0"></div>
</template>

<script setup lang="ts">
interface MatrixColumn {
  left: string
  delay: string
  duration: string
  text: string
}

const matrixSeeds = [
  'OPENAIACCESS10101',
  'MODELROUTER0101',
  'TEAMKEYCODE110',
  'STREAMJSON01011',
  'CODEXCHATGPT010',
  'LATENCYLOW00110',
]

const columns: MatrixColumn[] = Array.from({ length: 42 }, (_, index) => ({
  left: `${(index * 2.52) % 101}%`,
  delay: `${-(index % 13) * 0.72}s`,
  duration: `${9 + (index % 8) * 0.95}s`,
  text: matrixSeeds[index % matrixSeeds.length],
}))
</script>

<style scoped>
.public-matrix-rain {
  overflow: hidden;
  opacity: 0.56;
  mix-blend-mode: screen;
  mask-image: linear-gradient(to bottom, transparent 0%, rgba(0, 0, 0, 0.9) 12%, rgba(0, 0, 0, 0.38) 82%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, transparent 0%, rgba(0, 0, 0, 0.9) 12%, rgba(0, 0, 0, 0.38) 82%, transparent 100%);
}

.public-matrix-column {
  position: absolute;
  top: -80vh;
  display: block;
  width: 1ch;
  background: linear-gradient(to bottom, rgba(210, 255, 220, 0.9), rgba(45, 255, 85, 0.72), rgba(10, 130, 40, 0.14));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 0.78rem;
  font-weight: 800;
  line-height: 1.05;
  text-shadow: 0 0 8px rgba(89, 255, 146, 0.52);
  white-space: normal;
  word-break: break-all;
  animation: public-matrix-rain-fall linear infinite;
}

.public-blur-field {
  background:
    radial-gradient(ellipse at 50% 22%, rgba(52, 255, 128, 0.16), transparent 30%),
    radial-gradient(ellipse at 25% 30%, rgba(80, 87, 230, 0.15), transparent 32%),
    radial-gradient(ellipse at 72% 44%, rgba(45, 178, 105, 0.16), transparent 34%);
  filter: blur(54px);
  opacity: 0.85;
}

.public-noise {
  background-image:
    linear-gradient(rgba(102, 255, 161, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(102, 255, 161, 0.025) 1px, transparent 1px);
  background-size: 54px 54px;
  opacity: 0.34;
}

@keyframes public-matrix-rain-fall {
  0% {
    transform: translate3d(0, -10vh, 0);
  }
  100% {
    transform: translate3d(0, 190vh, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .public-matrix-column {
    animation: none;
    transform: translate3d(0, 20vh, 0);
  }
}
</style>
