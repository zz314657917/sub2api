export const CREDIT_SYMBOL = '✪'

function toFiniteAmount(value: number | null | undefined): number {
  const amount = Number(value)
  return Number.isFinite(amount) ? amount : 0
}

export function formatCreditAmount(
  value: number | null | undefined,
  options?: {
    minimumFractionDigits?: number
    maximumFractionDigits?: number
    signDisplay?: Intl.NumberFormatOptions['signDisplay']
  }
): string {
  const amount = toFiniteAmount(value)
  const formatter = new Intl.NumberFormat(undefined, {
    minimumFractionDigits: options?.minimumFractionDigits ?? 2,
    maximumFractionDigits: options?.maximumFractionDigits ?? 4,
    signDisplay: options?.signDisplay ?? 'auto',
  })

  return `${CREDIT_SYMBOL} ${formatter.format(amount)}`
}

export function formatCreditCompact(value: number | null | undefined): string {
  const amount = toFiniteAmount(value)
  const abs = Math.abs(amount)

  if (abs >= 100) {
    return formatCreditAmount(amount, { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }
  if (abs >= 1) {
    return formatCreditAmount(amount, { minimumFractionDigits: 4, maximumFractionDigits: 4 })
  }
  return formatCreditAmount(amount, { minimumFractionDigits: 6, maximumFractionDigits: 6 })
}

export function formatCreditExact(value: number | null | undefined): string {
  return formatCreditAmount(value, { minimumFractionDigits: 8, maximumFractionDigits: 8 })
}
