import type { PublicSettings } from '@/types'

export const DEFAULT_GROUP_BUY_PRODUCT_NAME = 'Token拼拼拼'

export function resolveGroupBuyProductName(settings?: Pick<PublicSettings, 'group_buy_product_name'> | null): string {
  const name = settings?.group_buy_product_name?.trim()
  return name || DEFAULT_GROUP_BUY_PRODUCT_NAME
}
