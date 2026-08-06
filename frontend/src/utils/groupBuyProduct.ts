import type { PublicSettings } from '@/types'

export const DEFAULT_GROUP_BUY_PRODUCT_NAME = 'Token拼拼拼'
export const DEFAULT_PIXEL_CAFE_TITLE = '像素网吧'

export function resolveGroupBuyProductName(settings?: Pick<PublicSettings, 'group_buy_product_name'> | null): string {
  const name = settings?.group_buy_product_name?.trim()
  return name || DEFAULT_GROUP_BUY_PRODUCT_NAME
}

export function resolvePixelCafeTitle(settings?: Pick<PublicSettings, 'pixel_cafe_title'> | null): string {
  const title = settings?.pixel_cafe_title?.trim()
  return title || DEFAULT_PIXEL_CAFE_TITLE
}
