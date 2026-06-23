export const DEFAULT_PAYMENT_CURRENCY = 'CNY'

export function normalizePaymentCurrency(currency?: string | null): string {
  const normalized = String(currency || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PAYMENT_CURRENCY
}

function paymentCurrencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function trimFractionTrailingZeros(value: string): string {
  if (!value.includes('.')) return value
  return value.replace(/(\.\d*?[1-9])0+$/, '$1').replace(/\.0+$/, '')
}

export function formatPaymentAmount(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  const value = Number.isFinite(amount) ? amount : 0
  try {
    return new Intl.NumberFormat(locale || undefined, {
      style: 'currency',
      currency: normalized,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(value)
  } catch {
    return `${normalized} ${value.toFixed(fractionDigits)}`
  }
}

export function formatPaymentAmountCompact(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  const value = Number.isFinite(amount) ? amount : 0
  try {
    return new Intl.NumberFormat(locale || undefined, {
      style: 'currency',
      currency: normalized,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: 0,
      maximumFractionDigits: fractionDigits,
    }).format(value)
  } catch {
    return `${normalized} ${trimFractionTrailingZeros(value.toFixed(fractionDigits))}`
  }
}
