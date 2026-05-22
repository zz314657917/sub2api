import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export interface MatrixColumn {
  id: number
  left: string
  delay: string
  duration: string
  text: string
  opacity: number
  fontSize: string
  lineHeight: string
  mutateEvery: number
  mutateOffset: number
}

const matrixSeeds = [
  'OPENAIACCESS',
  'MODELROUTER',
  'TEAMKEYCODE',
  'STREAMJSON',
  'MAPLEAIHUB',
  'LATENCYLOW',
  'QUOTAUSAGE',
  'APIKEYTOKEN',
  'PROXYDIRECT',
  'RUNTIMESYNC'
]

const matrixGlyphs = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ'
const matrixFontSizeBands = [
  [0.58, 0.7],
  [0.72, 0.88],
  [0.9, 1.04]
] as const
const matrixLaneJitterRatio = 0.18

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function randomFloat(min: number, max: number): number {
  return Math.random() * (max - min) + min
}

function pickRandomSeed(): string {
  return matrixSeeds[randomInt(0, matrixSeeds.length - 1)] || 'CODE'
}

function randomMatrixText(minLength = 14, maxLength = 28): string {
  const targetLength = randomInt(minLength, maxLength)
  let text = ''

  while (text.length < targetLength) {
    const seed = pickRandomSeed()
    const suffixLength = randomInt(1, 4)
    let suffix = ''

    for (let index = 0; index < suffixLength; index += 1) {
      suffix += matrixGlyphs[randomInt(0, matrixGlyphs.length - 1)] || '0'
    }

    text += seed + suffix
  }

  return text.slice(0, targetLength)
}

function randomMatrixFontSize(): string {
  const band = matrixFontSizeBands[randomInt(0, matrixFontSizeBands.length - 1)] || matrixFontSizeBands[1]
  return `${randomFloat(band[0], band[1]).toFixed(2)}rem`
}

function mutateMatrixText(source: string): string {
  const chars = source.split('')
  const mutationCount = Math.max(2, Math.floor(chars.length * randomFloat(0.2, 0.48)))

  for (let index = 0; index < mutationCount; index += 1) {
    const targetIndex = randomInt(0, chars.length - 1)
    chars[targetIndex] = matrixGlyphs[randomInt(0, matrixGlyphs.length - 1)] || '0'
  }

  return chars.join('')
}

export function createMatrixColumns(columnCount: number): MatrixColumn[] {
  const laneWidth = 100 / Math.max(1, columnCount - 1)
  const laneJitter = Math.min(0.45, laneWidth * matrixLaneJitterRatio)

  return Array.from({ length: columnCount }, (_, index) => {
    const mutateEvery = randomInt(2, 6)
    return {
      id: index,
      left: `${Math.min(98.8, Math.max(0.4, index * laneWidth + randomFloat(-laneJitter, laneJitter)))}%`,
      delay: `${(-randomFloat(0, 8.8)).toFixed(2)}s`,
      duration: `${randomFloat(7.6, 13.8).toFixed(2)}s`,
      text: randomMatrixText(),
      opacity: randomFloat(0.48, 0.96),
      fontSize: randomMatrixFontSize(),
      lineHeight: randomFloat(1.02, 1.18).toFixed(2),
      mutateEvery,
      mutateOffset: randomInt(0, mutateEvery - 1)
    }
  })
}

function refreshMatrixColumns(columns: MatrixColumn[], tick: number): MatrixColumn[] {
  return columns.map((column, index) => {
    if ((tick + index + column.mutateOffset) % column.mutateEvery !== 0) {
      return column
    }

    return {
      ...column,
      text: mutateMatrixText(column.text)
    }
  })
}

export function useMatrixRain(columnCount: number, refreshEveryMs = 560): { columns: Ref<MatrixColumn[]> } {
  const columns = ref<MatrixColumn[]>(createMatrixColumns(columnCount))
  let tick = 0
  let timer: number | null = null
  let reducedMotionQuery: MediaQueryList | null = null

  function stopTimer() {
    if (timer !== null) {
      window.clearInterval(timer)
      timer = null
    }
  }

  function startTimer() {
    stopTimer()
    tick = 0
    columns.value = createMatrixColumns(columnCount)

    if (reducedMotionQuery?.matches) {
      return
    }

    timer = window.setInterval(() => {
      tick += 1
      columns.value = refreshMatrixColumns(columns.value, tick)
    }, refreshEveryMs)
  }

  function handleReducedMotionChange() {
    startTimer()
  }

  onMounted(() => {
    if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
      reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
      reducedMotionQuery.addEventListener('change', handleReducedMotionChange)
    }

    startTimer()
  })

  onBeforeUnmount(() => {
    stopTimer()
    reducedMotionQuery?.removeEventListener('change', handleReducedMotionChange)
  })

  return { columns }
}
